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

type timeZoneMetadata struct {
	DatasetKey             string              `json:"dataset_key"`
	Title                  string              `json:"title"`
	Description            string              `json:"description"`
	CountryCode            string              `json:"country_code"`
	DatasetGroup           string              `json:"dataset_group"`
	Format                 string              `json:"format"`
	RelativePath           string              `json:"relative_path"`
	SchemaPath             string              `json:"schema_path"`
	RecordCount            int                 `json:"record_count"`
	Version                string              `json:"version"`
	VerifiedAt             string              `json:"verified_at"`
	LicenseID              string              `json:"license_id"`
	LicenseURL             string              `json:"license_url"`
	SourceTitle            string              `json:"source_title"`
	SourceURL              string              `json:"source_url"`
	ArchiveSHA256          string              `json:"archive_sha256"`
	ArchiveSizeBytes       int                 `json:"archive_size_bytes"`
	ArchiveLastModified    string              `json:"archive_last_modified"`
	ArchiveETag            string              `json:"archive_etag"`
	RetrievalTimestamps    []string            `json:"retrieval_timestamps"`
	SourceFileHashes       map[string]string   `json:"source_file_hashes"`
	Boundary               timeZoneBoundary    `json:"boundary"`
	CrosswalkCounts        timeZoneCrosswalk   `json:"crosswalk_counts"`
	DeterministicSorting   string              `json:"deterministic_sorting"`
	MappingAnchors         map[string][]string `json:"mapping_anchors"`
	MultiCountryZones      []string            `json:"multi_country_zones"`
	UpdatePolicy           string              `json:"update_policy"`
	VersioningPolicy       string              `json:"versioning_policy"`
	LicensingBoundary      string              `json:"licensing_boundary"`
	RecommendedAttribution string              `json:"recommended_attribution"`
}

type timeZoneBoundary struct {
	CanonicalSource        string     `json:"canonical_source"`
	CanonicalRecords       int        `json:"canonical_records"`
	ZoneTabExcluded        bool       `json:"zone_tab_excluded"`
	BackzoneExcluded       bool       `json:"backzone_excluded"`
	AliasPolicy            string     `json:"alias_policy"`
	AliasAudit             aliasAudit `json:"alias_audit"`
	CoordinateExclusion    bool       `json:"coordinate_exclusion"`
	CommentExclusion       bool       `json:"comment_exclusion"`
	OffsetAndDSTExclusion  bool       `json:"offset_and_dst_exclusion"`
	AsiaTaipeiEmptyMapping bool       `json:"asia_taipei_empty_mapping"`
	UnmatchedIANACode      string     `json:"unmatched_iana_code"`
	ZeroZoneM49IDs         []string   `json:"zero_zone_m49_ids"`
	ZeroZoneCount          int        `json:"zero_zone_count"`
}

type aliasAudit struct {
	BackwardAliases  int `json:"backward_aliases"`
	CanonicalTargets int `json:"canonical_targets"`
	ExcludedTargets  int `json:"excluded_targets"`
}

type timeZoneCrosswalk struct {
	TotalRelationships         int `json:"total_relationships"`
	UniqueMappedCountryAreaIDs int `json:"unique_mapped_country_area_ids"`
	ForwardZero                int `json:"forward_zero"`
	ForwardOne                 int `json:"forward_one"`
	ForwardMultiple            int `json:"forward_multiple"`
	ReverseZero                int `json:"reverse_zero"`
	ReverseOne                 int `json:"reverse_one"`
	ReverseMultiple            int `json:"reverse_multiple"`
}

type timeZoneSchema struct {
	Schema      string `json:"$schema"`
	ID          string `json:"$id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Type        string `json:"type"`
	MinItems    int    `json:"minItems"`
	MaxItems    int    `json:"maxItems"`
	UniqueItems bool   `json:"uniqueItems"`
	Items       struct {
		Type                 string   `json:"type"`
		AdditionalProperties bool     `json:"additionalProperties"`
		Required             []string `json:"required"`
		Properties           map[string]struct {
			Type        string `json:"type"`
			Pattern     string `json:"pattern,omitempty"`
			UniqueItems bool   `json:"uniqueItems,omitempty"`
			Items       *struct {
				Type string   `json:"type"`
				Enum []string `json:"enum,omitempty"`
			} `json:"items,omitempty"`
		} `json:"properties"`
	} `json:"items"`
}

func TestWorldTimeZonesDatasetMatchesApprovedManifest(t *testing.T) {
	rows := loadTimeZoneDataset(t)
	rawRows := loadTimeZoneRawDataset(t)
	approvedIDs := approvedCountryAreaIDs(t)
	approvedSet := make(map[string]struct{}, len(approvedIDs))
	for _, id := range approvedIDs {
		approvedSet[id] = struct{}{}
	}

	if rows == nil {
		t.Fatal("decoded dataset is nil")
	}
	if got := len(rows); got != 312 {
		t.Fatalf("unexpected record count: got %d want 312", got)
	}
	if len(rawRows) != 312 {
		t.Fatalf("unexpected raw record count: got %d want 312", len(rawRows))
	}

	seenIDs := make(map[string]struct{}, len(rows))
	totalRelationships := 0
	uniqueCountryAreaIDs := make(map[string]struct{}, len(approvedSet))
	zeroMappingZone := ""
	zeroZoneCount := 0
	multiCountryZones := make([]string, 0, 34)
	forwardHist := map[int]int{}
	reverseHist := make(map[string]int, len(approvedSet))

	for i, row := range rows {
		raw := rawRows[i]
		if row.ID == "" {
			t.Fatalf("record %d has empty id", i)
		}
		if row.CountryAreaIDs == nil {
			t.Fatalf("record %d has nil country_area_ids slice", i)
		}
		if !canonicalTimeZoneIDPattern.MatchString(row.ID) {
			t.Fatalf("record %d has invalid id %q", i, row.ID)
		}
		assertCanonicalTimeZoneID(t, row.ID)
		if _, ok := seenIDs[row.ID]; ok {
			t.Fatalf("duplicate id found: %q", row.ID)
		}
		seenIDs[row.ID] = struct{}{}
		if i > 0 && rows[i-1].ID > row.ID {
			t.Fatalf("records are not sorted by canonical id: %q before %q", rows[i-1].ID, row.ID)
		}

		for _, key := range rawFieldKeys(raw) {
			if key != "id" && key != "country_area_ids" {
				t.Fatalf("record %d contains unexpected field %q", i, key)
			}
		}

		if !reflect.DeepEqual(row.CountryAreaIDs, sortedUniqueStrings(row.CountryAreaIDs)) {
			t.Fatalf("record %q country_area_ids are not sorted and unique: %#v", row.ID, row.CountryAreaIDs)
		}

		for _, id := range row.CountryAreaIDs {
			if _, ok := approvedSet[id]; !ok {
				t.Fatalf("record %q contains unknown country_area_id %q", row.ID, id)
			}
			uniqueCountryAreaIDs[id] = struct{}{}
			reverseHist[id]++
		}

		switch len(row.CountryAreaIDs) {
		case 0:
			zeroMappingZone = row.ID
			zeroZoneCount++
		case 1:
		default:
			multiCountryZones = append(multiCountryZones, row.ID)
		}
		forwardHist[len(row.CountryAreaIDs)]++
		totalRelationships += len(row.CountryAreaIDs)

		switch row.ID {
		case "Africa/Lagos":
			assertStringSliceEqual(t, row.CountryAreaIDs, []string{"ao", "bj", "cd", "cf", "cg", "cm", "ga", "gq", "ne", "ng"})
		case "Europe/Simferopol":
			assertStringSliceEqual(t, row.CountryAreaIDs, []string{"ru", "ua"})
		case "Asia/Gaza", "Asia/Hebron":
			assertStringSliceEqual(t, row.CountryAreaIDs, []string{"ps"})
		case "Asia/Hong_Kong":
			assertStringSliceEqual(t, row.CountryAreaIDs, []string{"hk"})
		case "Asia/Macau":
			assertStringSliceEqual(t, row.CountryAreaIDs, []string{"mo"})
		case "Europe/London":
			assertStringSliceEqual(t, row.CountryAreaIDs, []string{"gb", "gg", "im", "je"})
		case "Atlantic/Faroe":
			assertStringSliceEqual(t, row.CountryAreaIDs, []string{"fo"})
		case "Asia/Taipei":
			assertStringSliceEqual(t, row.CountryAreaIDs, []string{})
		}

		if strings.HasPrefix(row.ID, "Etc/") || strings.HasPrefix(row.ID, "posix/") || strings.HasPrefix(row.ID, "right/") || row.ID == "Factory" {
			t.Fatalf("record %d contains excluded alias or special identifier %q", i, row.ID)
		}
		if strings.Contains(row.ID, " ") || strings.Contains(row.ID, "\\") || strings.Contains(row.ID, "%") {
			t.Fatalf("record %d contains unsafe characters in id %q", i, row.ID)
		}
	}

	if zeroMappingZone != "Asia/Taipei" {
		t.Fatalf("unexpected zero-mapping zone: got %q want Asia/Taipei", zeroMappingZone)
	}
	if zeroZoneCount != 1 {
		t.Fatalf("unexpected zero-mapping count: got %d want 1", zeroZoneCount)
	}

	expectedMulti := []string{
		"Africa/Abidjan",
		"Africa/Johannesburg",
		"Africa/Lagos",
		"Africa/Maputo",
		"Africa/Nairobi",
		"America/Panama",
		"America/Phoenix",
		"America/Puerto_Rico",
		"America/Toronto",
		"Asia/Bangkok",
		"Asia/Dubai",
		"Asia/Kuching",
		"Asia/Qatar",
		"Asia/Riyadh",
		"Asia/Singapore",
		"Asia/Tokyo",
		"Asia/Yangon",
		"Europe/Belgrade",
		"Europe/Berlin",
		"Europe/Brussels",
		"Europe/Helsinki",
		"Europe/London",
		"Europe/Paris",
		"Europe/Prague",
		"Europe/Rome",
		"Europe/Simferopol",
		"Europe/Zurich",
		"Indian/Maldives",
		"Pacific/Auckland",
		"Pacific/Guadalcanal",
		"Pacific/Guam",
		"Pacific/Pago_Pago",
		"Pacific/Port_Moresby",
		"Pacific/Tarawa",
	}
	sort.Strings(multiCountryZones)
	sort.Strings(expectedMulti)
	if !reflect.DeepEqual(multiCountryZones, expectedMulti) {
		t.Fatalf("unexpected multi-country zone set:\n got: %#v\nwant: %#v", multiCountryZones, expectedMulti)
	}
	if len(multiCountryZones) != 34 {
		t.Fatalf("unexpected multi-country zone count: got %d want 34", len(multiCountryZones))
	}

	if totalRelationships != 422 {
		t.Fatalf("unexpected total relationships: got %d want 422", totalRelationships)
	}
	if len(uniqueCountryAreaIDs) != 246 {
		t.Fatalf("unexpected unique mapped country/area id count: got %d want 246", len(uniqueCountryAreaIDs))
	}
	if forwardHist[0] != 1 || forwardHist[1] != 277 || len(multiCountryZones) != 34 {
		t.Fatalf("unexpected forward histogram: %#v", forwardHist)
	}

	reverseCounts := map[int]int{}
	for _, count := range reverseHist {
		reverseCounts[count]++
	}
	if reverseCounts[0] != 0 {
		t.Fatalf("unexpected reverse zero bucket in mapped ids: %#v", reverseCounts)
	}
	if reverseCounts[1] != 213 || reverseCounts[2] != 16 || reverseCounts[3] != 5 || reverseCounts[4] != 3 || reverseCounts[7] != 1 || reverseCounts[11] != 1 || reverseCounts[12] != 2 || reverseCounts[13] != 1 || reverseCounts[16] != 1 || reverseCounts[23] != 1 || reverseCounts[27] != 1 || reverseCounts[29] != 1 {
		t.Fatalf("unexpected reverse histogram: %#v", reverseCounts)
	}

	if _, ok := approvedSet["tw"]; ok {
		t.Fatal("approved world-countries-and-areas dataset unexpectedly includes tw")
	}
	if _, ok := reverseHist["tw"]; ok {
		t.Fatal("time zone dataset unexpectedly maps tw")
	}
}

func TestWorldTimeZonesMetadataSchemaAndNotices(t *testing.T) {
	metadata := loadTimeZoneMetadata(t)
	schema := loadTimeZoneSchema(t)
	approvedIDs := approvedCountryAreaIDs(t)

	if metadata.DatasetKey != "world-time-zones" {
		t.Fatalf("unexpected dataset key: %q", metadata.DatasetKey)
	}
	if metadata.Title != "World Time Zones" {
		t.Fatalf("unexpected title: %q", metadata.Title)
	}
	if metadata.CountryCode != "001" {
		t.Fatalf("unexpected country code: %q", metadata.CountryCode)
	}
	if metadata.DatasetGroup != "geography" {
		t.Fatalf("unexpected dataset group: %q", metadata.DatasetGroup)
	}
	if metadata.Format != "json" {
		t.Fatalf("unexpected format: %q", metadata.Format)
	}
	if metadata.RelativePath != "geography/time_zones.json" {
		t.Fatalf("unexpected relative path: %q", metadata.RelativePath)
	}
	if metadata.SchemaPath != "schemas/geography/time_zones.schema.json" {
		t.Fatalf("unexpected schema path: %q", metadata.SchemaPath)
	}
	if metadata.RecordCount != 312 {
		t.Fatalf("unexpected record count: %d", metadata.RecordCount)
	}
	if metadata.Version != "1.0.0" {
		t.Fatalf("unexpected version: %q", metadata.Version)
	}
	if metadata.VerifiedAt != "2026-09-01" {
		t.Fatalf("unexpected verification date: %q", metadata.VerifiedAt)
	}
	verifiedAt, err := time.Parse("2006-01-02", metadata.VerifiedAt)
	if err != nil {
		t.Fatalf("verified_at did not parse: %v", err)
	}
	today := time.Now().UTC()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	if verifiedAt.After(today) {
		t.Fatalf("verified_at is future-dated: %s > %s", verifiedAt.Format("2006-01-02"), today.Format("2006-01-02"))
	}
	if metadata.LicenseID != "CC-BY-4.0" {
		t.Fatalf("unexpected license id: %q", metadata.LicenseID)
	}
	if metadata.LicenseURL != "https://creativecommons.org/licenses/by/4.0/" {
		t.Fatalf("unexpected license url: %q", metadata.LicenseURL)
	}
	if metadata.SourceTitle != "Internet Assigned Numbers Authority, IANA Time Zone Database, release 2026c" {
		t.Fatalf("unexpected source title: %q", metadata.SourceTitle)
	}
	if metadata.SourceURL != "https://data.iana.org/time-zones/releases/tzdb-2026c.tar.lz" {
		t.Fatalf("unexpected source url: %q", metadata.SourceURL)
	}
	if metadata.ArchiveSHA256 != "427a11b1c5f2ebccad18f11650221c4f0465b4f1bb7f44dd02ff192d2808d944" {
		t.Fatalf("unexpected archive sha256: %q", metadata.ArchiveSHA256)
	}
	if metadata.ArchiveSizeBytes != 563235 {
		t.Fatalf("unexpected archive size: %d", metadata.ArchiveSizeBytes)
	}
	if metadata.ArchiveLastModified != "Wed, 08 Jul 2026 18:02:48 GMT" {
		t.Fatalf("unexpected archive last-modified: %q", metadata.ArchiveLastModified)
	}
	if metadata.ArchiveETag != "\"89823-6561d50afc200\"" {
		t.Fatalf("unexpected archive etag: %q", metadata.ArchiveETag)
	}
	assertStringSliceEqual(t, metadata.RetrievalTimestamps, []string{
		"2026-09-01T12:43:48Z",
		"2026-09-01T12:43:53Z",
		"2026-09-01T12:43:58Z",
	})
	if got := metadata.SourceFileHashes["zone1970.tab"]; got != "77b5e45415fa684fcc42de3421a6b0f15cc9b2c137f258083850346e8f76eea8" {
		t.Fatalf("unexpected zone1970.tab hash: %q", got)
	}
	if got := metadata.SourceFileHashes["zone.tab"]; got != "7cc78ea166261b3dedf951cdd721051460851e6fcd96c12b8e3194cf25677f21" {
		t.Fatalf("unexpected zone.tab hash: %q", got)
	}
	if got := metadata.SourceFileHashes["iso3166.tab"]; got != "837c80785080c8433fd9d4ea87e78f161ac7a40389301c5153d4f90198baeb2a" {
		t.Fatalf("unexpected iso3166.tab hash: %q", got)
	}
	if got := metadata.SourceFileHashes["backward"]; got != "d2f4c8953f204982ddf4dc0c2debf41b2464de376dad7d546d0fc70f889fa706" {
		t.Fatalf("unexpected backward hash: %q", got)
	}
	if got := metadata.SourceFileHashes["backzone"]; got != "63fb39adae0b0d8b2179629725a9dfb694c7a386b99750b636a017d896d28dfa" {
		t.Fatalf("unexpected backzone hash: %q", got)
	}
	if got := metadata.SourceFileHashes["LICENSE"]; got != "0613408568889f5739e5ae252b722a2659c02002839ad970a63dc5e9174b27cf" {
		t.Fatalf("unexpected LICENSE hash: %q", got)
	}
	if got := metadata.SourceFileHashes["README"]; got != "f6d96b82996a6ccac80027816704183c63a754d5b1eb2e7b25858e32164e2707" {
		t.Fatalf("unexpected README hash: %q", got)
	}
	if got := metadata.SourceFileHashes["theory.html"]; got != "88fb142cca79196eb804c3eb3b7511f6f366fef36d3e53bd2640f3c24d1d127e" {
		t.Fatalf("unexpected theory.html hash: %q", got)
	}

	if metadata.Boundary.CanonicalSource != "zone1970.tab" {
		t.Fatalf("unexpected canonical source: %q", metadata.Boundary.CanonicalSource)
	}
	if metadata.Boundary.CanonicalRecords != 312 || !metadata.Boundary.ZoneTabExcluded || !metadata.Boundary.BackzoneExcluded {
		t.Fatalf("unexpected boundary flags: %#v", metadata.Boundary)
	}
	if metadata.Boundary.AliasAudit.BackwardAliases != 256 || metadata.Boundary.AliasAudit.CanonicalTargets != 241 || metadata.Boundary.AliasAudit.ExcludedTargets != 15 {
		t.Fatalf("unexpected alias audit: %#v", metadata.Boundary.AliasAudit)
	}
	if !metadata.Boundary.CoordinateExclusion || !metadata.Boundary.CommentExclusion || !metadata.Boundary.OffsetAndDSTExclusion || !metadata.Boundary.AsiaTaipeiEmptyMapping {
		t.Fatalf("unexpected boundary exclusions: %#v", metadata.Boundary)
	}
	if metadata.Boundary.UnmatchedIANACode != "TW" {
		t.Fatalf("unexpected unmatched IANA code: %q", metadata.Boundary.UnmatchedIANACode)
	}
	assertStringSliceEqual(t, metadata.Boundary.ZeroZoneM49IDs, []string{"bv", "hm"})
	if metadata.Boundary.ZeroZoneCount != 2 {
		t.Fatalf("unexpected zero-zone count: %d", metadata.Boundary.ZeroZoneCount)
	}
	if metadata.CrosswalkCounts.TotalRelationships != 422 || metadata.CrosswalkCounts.UniqueMappedCountryAreaIDs != 246 || metadata.CrosswalkCounts.ForwardZero != 1 || metadata.CrosswalkCounts.ForwardOne != 277 || metadata.CrosswalkCounts.ForwardMultiple != 34 || metadata.CrosswalkCounts.ReverseZero != 2 || metadata.CrosswalkCounts.ReverseOne != 213 || metadata.CrosswalkCounts.ReverseMultiple != 33 {
		t.Fatalf("unexpected crosswalk counts: %#v", metadata.CrosswalkCounts)
	}
	if !strings.Contains(metadata.DeterministicSorting, "bytewise") || !strings.Contains(metadata.DeterministicSorting, "country_area_ids") {
		t.Fatalf("unexpected deterministic sorting note: %q", metadata.DeterministicSorting)
	}
	if !strings.Contains(metadata.LicensingBoundary, "public domain") || !strings.Contains(metadata.LicensingBoundary, "CC BY 4.0") || !strings.Contains(metadata.LicensingBoundary, "not relicensed") {
		t.Fatalf("unexpected licensing boundary: %q", metadata.LicensingBoundary)
	}
	if !strings.Contains(metadata.RecommendedAttribution, "Internet Assigned Numbers Authority") || !strings.Contains(metadata.RecommendedAttribution, "release 2026c") || !strings.Contains(metadata.RecommendedAttribution, "accessed 2026-09-01") {
		t.Fatalf("unexpected attribution: %q", metadata.RecommendedAttribution)
	}

	if schema.Schema == "" || schema.ID == "" || schema.Title == "" || schema.Description == "" {
		t.Fatalf("schema missing required metadata: %#v", schema)
	}
	if schema.Type != "array" || schema.MinItems != 312 || schema.MaxItems != 312 || !schema.UniqueItems {
		t.Fatalf("unexpected schema top-level constraints: %#v", schema)
	}
	if schema.Items.Type != "object" || schema.Items.AdditionalProperties {
		t.Fatalf("unexpected schema item constraints: %#v", schema.Items)
	}
	if !reflect.DeepEqual(schema.Items.Required, []string{"id", "country_area_ids"}) {
		t.Fatalf("unexpected schema required fields: %#v", schema.Items.Required)
	}
	idProp := schema.Items.Properties["id"]
	if idProp.Type != "string" || idProp.Pattern != "^[A-Za-z][A-Za-z0-9._+-]*(?:/[A-Za-z0-9._+-]+)+$" {
		t.Fatalf("unexpected id schema: %#v", idProp)
	}
	countryAreaIDsProp := schema.Items.Properties["country_area_ids"]
	if countryAreaIDsProp.Type != "array" || !countryAreaIDsProp.UniqueItems || countryAreaIDsProp.Items == nil {
		t.Fatalf("unexpected country_area_ids schema: %#v", countryAreaIDsProp)
	}
	if !reflect.DeepEqual(countryAreaIDsProp.Items.Enum, approvedIDs) {
		t.Fatalf("unexpected country_area_ids enum:\n got: %#v\nwant: %#v", countryAreaIDsProp.Items.Enum, approvedIDs)
	}

	if len(metadata.MappingAnchors) != 9 {
		t.Fatalf("unexpected mapping anchor count: %d", len(metadata.MappingAnchors))
	}
	if !reflect.DeepEqual(metadata.MappingAnchors["Asia/Taipei"], []string{}) {
		t.Fatalf("unexpected Asia/Taipei mapping anchor: %#v", metadata.MappingAnchors["Asia/Taipei"])
	}
}

func TestWorldTimeZonesDatasetRejectsUnexpectedFormats(t *testing.T) {
	rows := loadTimeZoneDataset(t)
	if len(rows) != 312 {
		t.Fatalf("unexpected record count: %d", len(rows))
	}
	for _, row := range rows {
		if strings.HasPrefix(row.ID, "Etc/") || strings.HasPrefix(row.ID, "posix/") || strings.HasPrefix(row.ID, "right/") || row.ID == "Factory" {
			t.Fatalf("unexpected excluded identifier: %q", row.ID)
		}
		if strings.Contains(row.ID, " ") || strings.Contains(row.ID, "\\") || strings.Contains(row.ID, "%") {
			t.Fatalf("unexpected unsafe identifier content: %q", row.ID)
		}
		for _, part := range strings.Split(row.ID, "/") {
			if part == "" || part == "." || part == ".." {
				t.Fatalf("unexpected unsafe path component %q in %q", part, row.ID)
			}
		}
	}
}

func loadTimeZoneDataset(t *testing.T) []TimeZone {
	t.Helper()

	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("geography/time_zones.json"))))
	var rows []TimeZone
	if err := dec.Decode(&rows); err != nil {
		t.Fatalf("decode time zone dataset: %v", err)
	}
	return rows
}

func loadTimeZoneRawDataset(t *testing.T) []map[string]any {
	t.Helper()

	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("geography/time_zones.json"))))
	var rows []map[string]any
	if err := dec.Decode(&rows); err != nil {
		t.Fatalf("decode raw time zone dataset: %v", err)
	}
	return rows
}

func loadTimeZoneMetadata(t *testing.T) timeZoneMetadata {
	t.Helper()

	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("metadata/geography/time_zones.json"))))
	var metadata timeZoneMetadata
	if err := dec.Decode(&metadata); err != nil {
		t.Fatalf("decode time zone metadata: %v", err)
	}
	return metadata
}

func loadTimeZoneSchema(t *testing.T) timeZoneSchema {
	t.Helper()

	dec := json.NewDecoder(bytes.NewReader(readTextBytes(t, datasetPath("schemas/geography/time_zones.schema.json"))))
	var schema timeZoneSchema
	if err := dec.Decode(&schema); err != nil {
		t.Fatalf("decode time zone schema: %v", err)
	}
	return schema
}

func approvedCountryAreaIDs(t *testing.T) []string {
	t.Helper()

	countries := loadCountryOrAreaDataset(t)
	ids := make([]string, 0, len(countries))
	for _, country := range countries {
		ids = append(ids, country.ID)
	}
	sort.Strings(ids)
	return ids
}

func assertCanonicalTimeZoneID(t *testing.T, id string) {
	t.Helper()

	if strings.HasPrefix(id, "/") || strings.HasSuffix(id, "/") {
		t.Fatalf("identifier has leading or trailing slash: %q", id)
	}
	parts := strings.Split(id, "/")
	if len(parts) < 2 {
		t.Fatalf("identifier does not contain a slash: %q", id)
	}
	first := regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._+-]*$`)
	rest := regexp.MustCompile(`^[A-Za-z0-9._+-]+$`)
	for i, part := range parts {
		if part == "" || part == "." || part == ".." {
			t.Fatalf("identifier contains unsafe path component %q: %q", part, id)
		}
		if i == 0 {
			if !first.MatchString(part) {
				t.Fatalf("identifier has invalid first component %q: %q", part, id)
			}
		} else if !rest.MatchString(part) {
			t.Fatalf("identifier has invalid component %q: %q", part, id)
		}
	}
	if strings.ContainsAny(id, " \t\n\r\\%?#") {
		t.Fatalf("identifier contains unsafe characters: %q", id)
	}
}

func rawFieldKeys(raw map[string]any) []string {
	keys := make([]string, 0, len(raw))
	for key := range raw {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedUniqueStrings(values []string) []string {
	out := append([]string(nil), values...)
	sort.Strings(out)
	if len(out) == 0 {
		return []string{}
	}
	dedup := out[:1]
	for i := 1; i < len(out); i++ {
		if out[i] != out[i-1] {
			dedup = append(dedup, out[i])
		}
	}
	return dedup
}

func assertStringSliceEqual(t *testing.T, got, want []string) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected slice:\n got: %#v\nwant: %#v", got, want)
	}
}

var canonicalTimeZoneIDPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._+-]*(?:/[A-Za-z0-9._+-]+)+$`)
