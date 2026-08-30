package models

import (
	"bytes"
	"encoding/json"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

type collegeOfEducationMetadata struct {
	DatasetKey      string          `json:"dataset_key"`
	Title           string          `json:"title"`
	Description     string          `json:"description"`
	CountryCode     string          `json:"country_code"`
	DatasetGroup    string          `json:"dataset_group"`
	Format          string          `json:"format"`
	RelativePath    string          `json:"relative_path"`
	SchemaPath      string          `json:"schema_path"`
	RecordCount     int             `json:"record_count"`
	OwnershipCounts map[string]int  `json:"ownership_counts"`
	Version         string          `json:"version"`
	LicenseID       string          `json:"license_id"`
	LicenseURL      string          `json:"license_url"`
	Methodology     string          `json:"methodology"`
	Sources         []datasetSource `json:"sources"`
	Excluded        []string        `json:"excluded_institutions"`
	ExclusionNotes  []string        `json:"exclusion_notes"`
	VerifiedAt      string          `json:"verified_at"`
}

type collegeOfEducationSchema struct {
	Schema      string          `json:"$schema"`
	ID          string          `json:"$id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Type        string          `json:"type"`
	MinItems    int             `json:"minItems"`
	MaxItems    int             `json:"maxItems"`
	UniqueItems bool            `json:"uniqueItems"`
	Items       json.RawMessage `json:"items"`
}

func TestNigeriaCollegesOfEducationDatasetMatchesApprovedRegister(t *testing.T) {
	states := loadStateDataset(t)
	universities := loadUniversityDataset(t)
	colleges := loadCollegeOfEducationDataset(t)

	if colleges == nil {
		t.Fatal("decoded dataset is nil")
	}
	if got := len(colleges); got != 244 {
		t.Fatalf("unexpected record count: got %d want 244", got)
	}

	stateByID := make(map[string]State, len(states))
	for _, state := range states {
		stateByID[state.ID] = state
	}
	universityByName := make(map[string]University, len(universities))
	for _, university := range universities {
		universityByName[university.Name] = university
	}

	wantCounts := map[string]int{
		"abia": 5, "adamawa": 2, "akwa-ibom": 2, "anambra": 6, "bauchi": 16,
		"bayelsa": 1, "benue": 15, "borno": 4, "cross-river": 3, "delta": 5,
		"ebonyi": 3, "edo": 3, "ekiti": 3, "enugu": 8, "fct": 5, "gombe": 8,
		"imo": 4, "jigawa": 3, "kaduna": 6, "kano": 16, "katsina": 4,
		"kebbi": 4, "kogi": 9, "kwara": 21, "lagos": 12, "nasarawa": 6,
		"niger": 2, "ogun": 10, "ondo": 9, "osun": 13, "oyo": 12, "plateau": 9,
		"rivers": 2, "sokoto": 3, "taraba": 4, "yobe": 4, "zamfara": 2,
	}
	wantOwnershipCounts := map[string]int{
		"federal": 28,
		"state":   48,
		"private": 168,
	}

	seenIDs := make(map[string]struct{}, len(colleges))
	seenNamesByState := make(map[string]map[string]struct{}, len(states))
	stateCounts := make(map[string]int, len(states))
	ownershipCounts := make(map[string]int, len(wantOwnershipCounts))
	sorted := append([]CollegeOfEducation(nil), colleges...)
	sort.SliceStable(sorted, func(i, j int) bool {
		left := strings.ToLower(sorted[i].Name)
		right := strings.ToLower(sorted[j].Name)
		if left == right {
			return sorted[i].ID < sorted[j].ID
		}
		return left < right
	})
	if !reflect.DeepEqual(colleges, sorted) {
		t.Fatal("dataset ordering is not deterministic by name then id")
	}

	for i, college := range colleges {
		if college.ID == "" || college.Name == "" || college.OwnershipType == "" || college.StateID == "" || college.CountryCode == "" {
			t.Fatalf("record %d has empty required field: %#v", i, college)
		}
		if college.CountryCode != "NG" {
			t.Fatalf("record %d has unexpected country code %q", i, college.CountryCode)
		}
		if _, ok := stateByID[college.StateID]; !ok {
			t.Fatalf("record %d references unknown state %q", i, college.StateID)
		}
		if !regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(college.ID) {
			t.Fatalf("record %d has invalid id %q", i, college.ID)
		}
		if wantID := slugifyCollegeName(college.Name); college.ID != wantID {
			t.Fatalf("record %d id mismatch: got %q want %q", i, college.ID, wantID)
		}
		if _, ok := seenIDs[college.ID]; ok {
			t.Fatalf("duplicate id found: %q", college.ID)
		}
		seenIDs[college.ID] = struct{}{}
		if _, ok := wantCounts[college.StateID]; !ok {
			t.Fatalf("record %d has unsupported state %q", i, college.StateID)
		}
		stateCounts[college.StateID]++
		ownershipCounts[college.OwnershipType]++

		stateNames, ok := seenNamesByState[college.StateID]
		if !ok {
			stateNames = make(map[string]struct{})
			seenNamesByState[college.StateID] = stateNames
		}
		if _, ok := stateNames[college.Name]; ok {
			t.Fatalf("duplicate name within state %q: %q", college.StateID, college.Name)
		}
		stateNames[college.Name] = struct{}{}

		if strings.Contains(strings.ToUpper(college.Name), "OPEN") {
			t.Fatalf("directory marker leaked into public name: %q", college.Name)
		}

		marshaled, err := json.Marshal(college)
		if err != nil {
			t.Fatalf("marshal college %d: %v", i, err)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(marshaled, &raw); err != nil {
			t.Fatalf("unmarshal marshaled college %d: %v", i, err)
		}
		if len(raw) != 5 {
			t.Fatalf("record %d marshaled with unexpected field count: got %d want 5", i, len(raw))
		}
		for _, field := range []string{"id", "name", "ownership_type", "state_id", "country_code"} {
			if _, ok := raw[field]; !ok {
				t.Fatalf("record %d missing serialized field %q", i, field)
			}
		}
	}

	if len(seenNamesByState) != len(states) {
		t.Fatalf("unexpected parent coverage: got %d states with children want %d", len(seenNamesByState), len(states))
	}
	for _, state := range states {
		if got := stateCounts[state.ID]; got != wantCounts[state.ID] {
			t.Fatalf("unexpected count for %s: got %d want %d", state.ID, got, wantCounts[state.ID])
		}
	}
	if !reflect.DeepEqual(ownershipCounts, wantOwnershipCounts) {
		t.Fatalf("unexpected ownership totals: got %#v want %#v", ownershipCounts, wantOwnershipCounts)
	}

	for _, wantName := range []string{
		"FCT College of Education, Zuba",
		"Federal College of Education (Special), Oyo",
		"Federal College of Education (T), Umunze",
		"Federal College of Education Ofeme-Ohuhu",
		"Isaac Jasper Boro COE, Sagbama",
		"Yusuf Maitama Sule College of Education & Advanced Studies, Ghari",
		"Al-Ibadan COE",
		"St. Paul's College of Education NNewi",
		"A.D. Rufa’i College of Education, Legal and General Studies",
	} {
		if _, ok := findCollegeByName(colleges, wantName); !ok {
			t.Fatalf("missing expected college %q", wantName)
		}
	}

	wantSpecific := map[string]struct {
		stateID string
		own     string
	}{
		"FCT College of Education, Zuba":                                    {stateID: "fct", own: "federal"},
		"Federal College of Education (Special), Oyo":                       {stateID: "oyo", own: "federal"},
		"Federal College of Education (T), Umunze":                          {stateID: "anambra", own: "federal"},
		"Federal College of Education Ofeme-Ohuhu":                          {stateID: "abia", own: "federal"},
		"Isaac Jasper Boro COE, Sagbama":                                    {stateID: "bayelsa", own: "state"},
		"Yusuf Maitama Sule College of Education & Advanced Studies, Ghari": {stateID: "kano", own: "state"},
	}
	for name, want := range wantSpecific {
		college, ok := findCollegeByName(colleges, name)
		if !ok {
			t.Fatalf("missing expected college %q", name)
		}
		if college.StateID != want.stateID {
			t.Fatalf("unexpected state for %q: got %q want %q", name, college.StateID, want.stateID)
		}
		if college.OwnershipType != want.own {
			t.Fatalf("unexpected ownership for %q: got %q want %q", name, college.OwnershipType, want.own)
		}
	}

	for _, excluded := range []string{
		"Abubakar Tatari Polytechnic",
		"Aminu Kano College of Islamic and Legal Studies",
		"Bauchi Institute for Arabic and Islamic Studies",
		"Hassan Usman Katsina Polytechnic",
		"Institute of Ecumenical Education (Thinkers Corner)",
		"Jigawa State Polytechnic",
		"Kaduna Polytechnics",
		"Kano State Polytechnic",
		"Kebbi State Polytechnic",
		"Muhammad Goni College of Legal and Islamic Studies (MOGOLIS)",
		"National Institute for Nigerian Languages",
		"National Teachers Institute (NTI)",
		"Nigerian Army College of Education (NACOE), Ilorin",
		"Nuhu Bamalli Polytechnic",
		"Plateau State Polytechnic",
		"Ramat Polytechnic",
		"The Polytechnic Iree, Osun State",
		"Waziri Umaru Federal Polytechnic",
		"Zaria Institute of Information Technology",
		"Cross River State Coll. of Education, Akampa",
	} {
		if _, ok := findCollegeByName(colleges, excluded); ok {
			t.Fatalf("excluded institution unexpectedly present: %q", excluded)
		}
	}

	if university, ok := universityByName["Cross River University of Education and Entrepreneurship, Akamkpa, Cross River State"]; !ok {
		t.Fatal("successor university is missing from the universities dataset")
	} else if university.StateID != "cross-river" {
		t.Fatalf("unexpected successor university state: %q", university.StateID)
	}
}

func TestNigeriaCollegesOfEducationMetadataSchemaAndNotices(t *testing.T) {
	metadata := loadCollegeOfEducationMetadata(t)
	schema := loadCollegeOfEducationSchema(t)

	if metadata.DatasetKey != "ng-colleges-of-education" {
		t.Fatalf("unexpected dataset key: %q", metadata.DatasetKey)
	}
	if metadata.Title != "Nigeria Colleges of Education" {
		t.Fatalf("unexpected title: %q", metadata.Title)
	}
	if metadata.Description == "" || !strings.Contains(metadata.Description, "current National Commission for Colleges of Education register") {
		t.Fatalf("unexpected description: %q", metadata.Description)
	}
	if metadata.CountryCode != "NG" || metadata.DatasetGroup != "education" || metadata.Format != "json" {
		t.Fatalf("unexpected metadata classification: %#v", metadata)
	}
	if metadata.RelativePath != "education/colleges_of_education.json" {
		t.Fatalf("unexpected relative path: %q", metadata.RelativePath)
	}
	if metadata.SchemaPath != "schemas/education/colleges_of_education.schema.json" {
		t.Fatalf("unexpected schema path: %q", metadata.SchemaPath)
	}
	if metadata.RecordCount != 244 {
		t.Fatalf("unexpected record count: %d", metadata.RecordCount)
	}
	if metadata.Version != "1.0.0" {
		t.Fatalf("unexpected version: %q", metadata.Version)
	}
	if metadata.LicenseID != "CC-BY-4.0" || metadata.LicenseURL != "https://creativecommons.org/licenses/by/4.0/" {
		t.Fatalf("unexpected license metadata: %#v", metadata)
	}
	if metadata.VerifiedAt != "2026-08-30" {
		t.Fatalf("unexpected verified_at: %q", metadata.VerifiedAt)
	}
	verifiedAt, err := time.Parse("2006-01-02", metadata.VerifiedAt)
	if err != nil {
		t.Fatalf("verified_at is not a valid date: %v", err)
	}
	if verifiedAt.After(time.Now().UTC()) {
		t.Fatalf("verified_at is in the future: %s", metadata.VerifiedAt)
	}
	if !reflect.DeepEqual(metadata.OwnershipCounts, map[string]int{"federal": 28, "state": 48, "private": 168}) {
		t.Fatalf("unexpected ownership counts: %#v", metadata.OwnershipCounts)
	}
	wantExcluded := []string{
		"Abubakar Tatari Polytechnic",
		"Aminu Kano College of Islamic and Legal Studies",
		"Bauchi Institute for Arabic and Islamic Studies",
		"Hassan Usman Katsina Polytechnic",
		"Institute of Ecumenical Education (Thinkers Corner)",
		"Jigawa State Polytechnic",
		"Kaduna Polytechnics",
		"Kano State Polytechnic",
		"Kebbi State Polytechnic",
		"Muhammad Goni College of Legal and Islamic Studies (MOGOLIS)",
		"National Institute for Nigerian Languages",
		"National Teachers Institute (NTI)",
		"Nigerian Army College of Education (NACOE), Ilorin",
		"Nuhu Bamalli Polytechnic",
		"Plateau State Polytechnic",
		"Ramat Polytechnic",
		"The Polytechnic Iree, Osun State",
		"Waziri Umaru Federal Polytechnic",
		"Zaria Institute of Information Technology",
	}
	if !reflect.DeepEqual(metadata.Excluded, wantExcluded) {
		t.Fatalf("unexpected excluded institution list:\n got: %#v\nwant: %#v", metadata.Excluded, wantExcluded)
	}
	if len(metadata.ExclusionNotes) != 1 || !strings.Contains(metadata.ExclusionNotes[0], "Cross River State Coll. of Education, Akampa") {
		t.Fatalf("unexpected exclusion note: %#v", metadata.ExclusionNotes)
	}
	if len(metadata.Sources) != 11 {
		t.Fatalf("unexpected source count: %d", len(metadata.Sources))
	}
	for _, source := range metadata.Sources {
		if source.Organization == "" || source.Title == "" || source.URL == "" || source.Purpose == "" || source.AccessedAt == "" {
			t.Fatalf("incomplete provenance source: %#v", source)
		}
		if !strings.HasPrefix(source.URL, "https://") {
			t.Fatalf("non-https provenance url: %q", source.URL)
		}
	}
	for _, snippet := range []string{
		"retrieved three times with consistent 245-row results",
		"245-college NCCE boundary",
		"19 non-COE NCE-awarding institutions",
		"directory-only OPEN marker",
		"normalized defective state labels",
		"literal value 1",
		"Al-Ibadan COE",
		"St. Paul's College of Education NNewi",
		"A.D. Rufa’i College of Education, Legal and General Studies",
		"Yusuf Maitama Sule College of Education & Advanced Studies, Ghari",
		"legacy archive was treated as a candidate inventory only",
		"official publications retain their own rights",
	} {
		if !strings.Contains(metadata.Methodology, snippet) {
			t.Fatalf("methodology missing %q: %q", snippet, metadata.Methodology)
		}
	}

	if schema.Schema != "https://json-schema.org/draft/2020-12/schema" {
		t.Fatalf("unexpected schema dialect: %q", schema.Schema)
	}
	if schema.ID != "https://softdata-api.local/schemas/education/colleges_of_education.schema.json" {
		t.Fatalf("unexpected schema id: %q", schema.ID)
	}
	if schema.Title != "Nigeria Colleges of Education" {
		t.Fatalf("unexpected schema title: %q", schema.Title)
	}
	if schema.Type != "array" || schema.MinItems != 244 || schema.MaxItems != 244 || !schema.UniqueItems {
		t.Fatalf("unexpected schema constraints: %#v", schema)
	}

	var items map[string]any
	if err := json.Unmarshal(schema.Items, &items); err != nil {
		t.Fatalf("decode schema items: %v", err)
	}
	if got := items["type"]; got != "object" {
		t.Fatalf("unexpected schema item type: %#v", got)
	}
	if got := items["additionalProperties"]; got != false {
		t.Fatalf("unexpected additionalProperties: %#v", got)
	}
	required := asStringSlice(t, items["required"])
	if !reflect.DeepEqual(required, []string{"id", "name", "ownership_type", "state_id", "country_code"}) {
		t.Fatalf("unexpected required fields: %#v", required)
	}
	props := asMap(t, items["properties"])
	ownershipProp := asMap(t, props["ownership_type"])
	ownershipEnum := asStringSlice(t, ownershipProp["enum"])
	if !reflect.DeepEqual(ownershipEnum, []string{"federal", "state", "private"}) {
		t.Fatalf("unexpected ownership_type enum: %#v", ownershipEnum)
	}
	stateProp := asMap(t, props["state_id"])
	stateEnum := asStringSlice(t, stateProp["enum"])
	if len(stateEnum) != 37 || !containsString(stateEnum, "fct") || !containsString(stateEnum, "taraba") || !containsString(stateEnum, "abia") {
		t.Fatalf("unexpected state_id enum: %#v", stateEnum)
	}
	countryProp := asMap(t, props["country_code"])
	if got := countryProp["const"]; got != "NG" {
		t.Fatalf("unexpected country_code const: %#v", got)
	}
	if schema.Description == "" || !strings.Contains(schema.Description, "ng-colleges-of-education") {
		t.Fatalf("unexpected schema description: %q", schema.Description)
	}
}

func loadCollegeOfEducationDataset(t *testing.T) []CollegeOfEducation {
	t.Helper()

	var colleges []CollegeOfEducation
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("education/colleges_of_education.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&colleges); err != nil {
		t.Fatalf("decode colleges of education dataset: %v", err)
	}
	if err := dec.Decode(new(any)); err == nil {
		t.Fatal("colleges of education dataset contains trailing JSON")
	}
	return colleges
}

func loadCollegeOfEducationMetadata(t *testing.T) collegeOfEducationMetadata {
	t.Helper()

	var metadata collegeOfEducationMetadata
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("metadata/education/colleges_of_education.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&metadata); err != nil {
		t.Fatalf("decode colleges of education metadata: %v", err)
	}
	return metadata
}

func loadCollegeOfEducationSchema(t *testing.T) collegeOfEducationSchema {
	t.Helper()

	var schema collegeOfEducationSchema
	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("schemas/education/colleges_of_education.schema.json"))))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&schema); err != nil {
		t.Fatalf("decode colleges of education schema: %v", err)
	}
	return schema
}

func slugifyCollegeName(name string) string {
	var b strings.Builder
	b.Grow(len(name))
	lastHyphen := false
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		default:
			if !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	slug := strings.Trim(b.String(), "-")
	slug = strings.ReplaceAll(slug, "--", "-")
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return slug
}

func findCollegeByName(colleges []CollegeOfEducation, want string) (CollegeOfEducation, bool) {
	for _, college := range colleges {
		if college.Name == want {
			return college, true
		}
	}
	return CollegeOfEducation{}, false
}
