package file

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

type institutionDatasetSpec[T any] struct {
	name      string
	path      string
	notFound  error
	list      func(*EducationFileRepository, context.Context) ([]T, error)
	get       func(*EducationFileRepository, context.Context, string) (T, error)
	extraPath string
}

func TestEducationRepositoryNewSmallDatasets(t *testing.T) {
	t.Parallel()

	t.Run("polytechnics", func(t *testing.T) {
		t.Parallel()
		fixture := loadEducationSlice[models.Polytechnic](t, "../../../datasets/education/polytechnics.json")
		runInstitutionDatasetSpec(t, "polytechnics", "education/polytechnics.json", interfaces.ErrPolytechnicNotFound, fixture)
	})
	t.Run("monotechnics", func(t *testing.T) {
		t.Parallel()
		fixture := loadEducationSlice[models.Monotechnic](t, "../../../datasets/education/monotechnics.json")
		runInstitutionDatasetSpec(t, "monotechnics", "education/monotechnics.json", interfaces.ErrMonotechnicNotFound, fixture)
	})
	t.Run("agriculture", func(t *testing.T) {
		t.Parallel()
		fixture := loadEducationSlice[models.CollegeOfAgriculture](t, "../../../datasets/education/colleges_of_agriculture.json")
		runInstitutionDatasetSpec(t, "agriculture", "education/colleges_of_agriculture.json", interfaces.ErrCollegeOfAgricultureNotFound, fixture)
	})
	t.Run("health", func(t *testing.T) {
		t.Parallel()
		fixture := loadEducationSlice[models.CollegeOfHealthSciencesAndTechnology](t, "../../../datasets/education/colleges_of_health_sciences_and_technology.json")
		runInstitutionDatasetSpec(t, "health", "education/colleges_of_health_sciences_and_technology.json", interfaces.ErrCollegeOfHealthSciencesAndTechnologyNotFound, fixture)
	})
	t.Run("nursing", func(t *testing.T) {
		t.Parallel()
		fixture := loadEducationSlice[models.CollegeOfNursingAndMidwifery](t, "../../../datasets/education/colleges_of_nursing_and_midwifery.json")
		runInstitutionDatasetSpec(t, "nursing", "education/colleges_of_nursing_and_midwifery.json", interfaces.ErrCollegeOfNursingAndMidwiferyNotFound, fixture)
	})
	t.Run("technical", func(t *testing.T) {
		t.Parallel()
		fixture := loadEducationSlice[models.TechnicalCollege](t, "../../../datasets/education/technical_colleges.json")
		runInstitutionDatasetSpec(t, "technical", "education/technical_colleges.json", interfaces.ErrTechnicalCollegeNotFound, fixture)
	})
	t.Run("vei", func(t *testing.T) {
		t.Parallel()
		fixture := loadEducationSlice[models.VocationalEnterpriseInstitution](t, "../../../datasets/education/vocational_enterprise_institutions.json")
		runInstitutionDatasetSpec(t, "vei", "education/vocational_enterprise_institutions.json", interfaces.ErrVocationalEnterpriseInstitutionNotFound, fixture)
	})
}

func runInstitutionDatasetSpec[T any](t *testing.T, name, path string, notFound error, fixture []T) {
	t.Helper()
	repo := mustNewEducationRepositoryWithDatasetStub(t, path, fixture)

	listed1, err := callInstitutionList[T](repo, context.Background(), name)
	if err != nil {
		t.Fatalf("%s list error = %v", name, err)
	}
	if listed1 == nil {
		t.Fatalf("%s list returned nil slice", name)
	}
	if got := len(listed1); got != len(fixture) {
		t.Fatalf("%s list count = %d, want %d", name, got, len(fixture))
	}

	if len(listed1) > 0 {
		firstID := institutionFieldString(listed1[0], "ID")
		if firstID == "" {
			t.Fatalf("%s first ID was empty", name)
		}
		got, err := callInstitutionGet[T](repo, context.Background(), firstID, name)
		if err != nil {
			t.Fatalf("%s get error = %v", name, err)
		}
		if institutionFieldString(got, "ID") != firstID {
			t.Fatalf("%s get ID mismatch", name)
		}
	}

	listed1 = append(listed1[:0:0], listed1...)
	if len(listed1) > 0 {
		institutionSetStringField(&listed1[0], "Name", "Changed")
	}
	listed2, err := callInstitutionList[T](repo, context.Background(), name)
	if err != nil {
		t.Fatalf("%s second list error = %v", name, err)
	}
	if len(listed2) > 0 && institutionFieldString(listed2[0], "Name") == "Changed" {
		t.Fatalf("%s list exposed cached slice state", name)
	}

	if _, err := callInstitutionGet[T](repo, context.Background(), "missing-"+slugifyEducationInstitutionName(name), name); !errors.Is(err, notFound) {
		t.Fatalf("%s missing lookup = %v, want %v", name, err, notFound)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := callInstitutionList[T](repo, ctx, name); !errors.Is(err, context.Canceled) {
		t.Fatalf("%s canceled list = %v, want context.Canceled", name, err)
	}

	stub := &educationJSONRepoStub{pathCalls: map[string]int{}}
	stub.decodeFn = func(ctx context.Context, relativePath string, destination any) error {
		if relativePath != path {
			t.Fatalf("%s decode path = %q, want %q", name, relativePath, path)
		}
		switch dest := destination.(type) {
		case *[]T:
			*dest = cloneSlice(fixture)
			return nil
		default:
			return fmt.Errorf("unexpected destination %T", destination)
		}
	}
	repo, err = NewEducationRepository(stub, "education/universities.json", "education/colleges_of_education.json",
		"education/polytechnics.json", "education/monotechnics.json", "education/colleges_of_agriculture.json",
		"education/colleges_of_health_sciences_and_technology.json", "education/colleges_of_nursing_and_midwifery.json",
		"education/technical_colleges.json", "education/vocational_enterprise_institutions.json",
		"education/primary_and_secondary_schools.json")
	if err != nil {
		t.Fatalf("NewEducationRepository() error = %v", err)
	}

	if _, err := callInstitutionList[T](repo, context.Background(), name); err != nil {
		t.Fatalf("%s first cached list error = %v", name, err)
	}
	if _, err := callInstitutionList[T](repo, context.Background(), name); err != nil {
		t.Fatalf("%s second cached list error = %v", name, err)
	}
	if stub.pathCalls[path] != 1 {
		t.Fatalf("%s decode count = %d, want 1", name, stub.pathCalls[path])
	}

	t.Run(name+" mutations", func(t *testing.T) {
		mutations := []struct {
			name string
			mut  func([]T) []T
		}{
			{name: "nil slice", mut: func([]T) []T { return nil }},
			{name: "empty slice", mut: func([]T) []T { return make([]T, 0) }},
			{name: "short slice", mut: func(items []T) []T { return cloneSlice(items[:len(items)-1]) }},
			{name: "duplicate id", mut: func(items []T) []T {
				out := cloneSlice(items)
				institutionSetStringField(&out[1], "ID", institutionFieldString(out[0], "ID"))
				return out
			}},
			{name: "duplicate name", mut: func(items []T) []T {
				out := cloneSlice(items)
				institutionSetStringField(&out[1], "Name", institutionFieldString(out[0], "Name"))
				institutionSetStringField(&out[1], "ID", slugifyEducationInstitutionName(institutionFieldString(out[0], "Name")))
				return out
			}},
			{name: "invalid ordering", mut: func(items []T) []T {
				out := cloneSlice(items)
				if len(out) > 1 {
					out[0], out[1] = out[1], out[0]
				}
				return out
			}},
			{name: "invalid ownership", mut: func(items []T) []T {
				out := cloneSlice(items)
				institutionSetStringField(&out[0], "OwnershipType", "unknown")
				return out
			}},
			{name: "invalid state", mut: func(items []T) []T {
				out := cloneSlice(items)
				institutionSetStringField(&out[0], "StateID", "invalid")
				return out
			}},
			{name: "invalid country", mut: func(items []T) []T {
				out := cloneSlice(items)
				institutionSetStringField(&out[0], "CountryCode", "GH")
				return out
			}},
		}

		for _, mutation := range mutations {
			mutation := mutation
			t.Run(mutation.name, func(t *testing.T) {
				badRepo := mustNewEducationRepositoryWithDatasetStub(t, path, mutation.mut(fixture))
				_, err := callInstitutionList[T](badRepo, context.Background(), name)
				if err == nil {
					t.Fatalf("%s mutation %s unexpectedly succeeded", name, mutation.name)
				}
				if !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
					t.Fatalf("%s mutation %s error = %v, want ErrInvalidDatasetFile", name, mutation.name, err)
				}
			})
		}
	})

	t.Run(name+" unknown field", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Clean("../../../datasets/" + path))
		if err != nil {
			t.Fatalf("ReadFile() error = %v", err)
		}
		updated := bytes.Replace(data, []byte(`"id":"`), []byte(`"unexpected":"x","id":"`), 1)
		if bytes.Equal(updated, data) {
			updated = bytes.Replace(data, []byte(`"id": `), []byte(`"unexpected":"x","id": `), 1)
		}
		data = updated
		root := t.TempDir()
		fixturePath := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(fixturePath), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(fixturePath, data, 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}
		jsonRepo, err := NewJSONRepository(root, 100<<20)
		if err != nil {
			t.Fatalf("NewJSONRepository() error = %v", err)
		}
		repo, err := NewEducationRepository(jsonRepo, "education/universities.json", "education/colleges_of_education.json",
			"education/polytechnics.json", "education/monotechnics.json", "education/colleges_of_agriculture.json",
			"education/colleges_of_health_sciences_and_technology.json", "education/colleges_of_nursing_and_midwifery.json",
			"education/technical_colleges.json", "education/vocational_enterprise_institutions.json",
			"education/primary_and_secondary_schools.json")
		if err != nil {
			t.Fatalf("NewEducationRepository() error = %v", err)
		}
		_, err = callInstitutionList[T](repo, context.Background(), name)
		if err == nil {
			t.Fatal("expected unknown-field decode to fail")
		}
		if !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
			t.Fatalf("unknown-field error = %v, want ErrInvalidDatasetFile", err)
		}
	})
}

func mustNewEducationRepositoryWithDatasetStub[T any](t *testing.T, path string, fixture []T) *EducationFileRepository {
	t.Helper()

	stub := &educationJSONRepoStub{pathCalls: map[string]int{}}
	stub.decodeFn = func(ctx context.Context, relativePath string, destination any) error {
		stub.pathCalls[relativePath]++
		switch dest := destination.(type) {
		case *[]T:
			if relativePath != path {
				return fmt.Errorf("unexpected path %s", relativePath)
			}
			*dest = cloneSlice(fixture)
			return nil
		default:
			return fmt.Errorf("unexpected destination %T", destination)
		}
	}

	repo, err := NewEducationRepository(stub, "education/universities.json", "education/colleges_of_education.json",
		"education/polytechnics.json", "education/monotechnics.json", "education/colleges_of_agriculture.json",
		"education/colleges_of_health_sciences_and_technology.json", "education/colleges_of_nursing_and_midwifery.json",
		"education/technical_colleges.json", "education/vocational_enterprise_institutions.json",
		"education/primary_and_secondary_schools.json")
	if err != nil {
		t.Fatalf("NewEducationRepository() error = %v", err)
	}
	return repo
}

func callInstitutionList[T any](repo *EducationFileRepository, ctx context.Context, name string) ([]T, error) {
	switch name {
	case "polytechnics":
		rows, err := repo.ListPolytechnics(ctx)
		return any(rows).([]T), err
	case "monotechnics":
		rows, err := repo.ListMonotechnics(ctx)
		return any(rows).([]T), err
	case "agriculture":
		rows, err := repo.ListCollegesOfAgriculture(ctx)
		return any(rows).([]T), err
	case "health":
		rows, err := repo.ListCollegesOfHealthSciencesAndTechnology(ctx)
		return any(rows).([]T), err
	case "nursing":
		rows, err := repo.ListCollegesOfNursingAndMidwifery(ctx)
		return any(rows).([]T), err
	case "technical":
		rows, err := repo.ListTechnicalColleges(ctx)
		return any(rows).([]T), err
	case "vei":
		rows, err := repo.ListVocationalEnterpriseInstitutions(ctx)
		return any(rows).([]T), err
	default:
		panic("unknown institution dataset")
	}
}

func callInstitutionGet[T any](repo *EducationFileRepository, ctx context.Context, id string, name string) (T, error) {
	switch name {
	case "polytechnics":
		rows, err := repo.GetPolytechnic(ctx, id)
		return any(rows).(T), err
	case "monotechnics":
		rows, err := repo.GetMonotechnic(ctx, id)
		return any(rows).(T), err
	case "agriculture":
		rows, err := repo.GetCollegeOfAgriculture(ctx, id)
		return any(rows).(T), err
	case "health":
		rows, err := repo.GetCollegeOfHealthSciencesAndTechnology(ctx, id)
		return any(rows).(T), err
	case "nursing":
		rows, err := repo.GetCollegeOfNursingAndMidwifery(ctx, id)
		return any(rows).(T), err
	case "technical":
		rows, err := repo.GetTechnicalCollege(ctx, id)
		return any(rows).(T), err
	case "vei":
		rows, err := repo.GetVocationalEnterpriseInstitution(ctx, id)
		return any(rows).(T), err
	default:
		panic("unknown institution dataset")
	}
}

func loadEducationSlice[T any](t *testing.T, path string) []T {
	t.Helper()

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	var rows []T
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&rows); err != nil {
		t.Fatalf("decode %s error = %v", path, err)
	}
	return rows
}

func institutionFieldString(item any, field string) string {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	f := v.FieldByName(field)
	if !f.IsValid() || f.Kind() != reflect.String {
		return ""
	}
	return f.String()
}

func institutionSetStringField(item any, field, value string) {
	v := reflect.ValueOf(item)
	if v.Kind() != reflect.Pointer {
		panic("institutionSetStringField requires a pointer")
	}
	v = v.Elem()
	f := v.FieldByName(field)
	if !f.IsValid() || !f.CanSet() || f.Kind() != reflect.String {
		panic("unable to set institution string field " + field)
	}
	f.SetString(value)
}

func TestSchoolRepositoryPaginationAndFilters(t *testing.T) {
	t.Parallel()

	fixture := loadEducationSlice[models.PrimaryAndSecondarySchool](t, "../../../datasets/education/primary_and_secondary_schools.json")
	repo := mustNewSchoolRepositoryWithFixture(t, fixture)

	result, err := repo.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{})
	if err != nil {
		t.Fatalf("ListPrimaryAndSecondarySchools() error = %v", err)
	}
	if result.Page != 1 || result.PageSize != 50 {
		t.Fatalf("unexpected defaults: %#v", result)
	}
	if result.Total != len(fixture) {
		t.Fatalf("unexpected total: got %d want %d", result.Total, len(fixture))
	}
	if result.TotalPages != 3333 {
		t.Fatalf("unexpected total pages: got %d want 3333", result.TotalPages)
	}
	if len(result.Schools) != 50 {
		t.Fatalf("unexpected default page length: %d", len(result.Schools))
	}
	if result.Schools[0].ID != fixture[0].ID {
		t.Fatalf("unexpected first record: got %s want %s", result.Schools[0].ID, fixture[0].ID)
	}

	result.Schools[0].Name = "Changed"
	result.Schools[0].EducationLevels[0] = "primary"
	again, err := repo.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{})
	if err != nil {
		t.Fatalf("second list error = %v", err)
	}
	if again.Schools[0].Name == "Changed" {
		t.Fatal("school list exposed shared slice state")
	}
	if len(again.Schools[0].EducationLevels) > 0 && again.Schools[0].EducationLevels[0] == "primary" && fixture[0].EducationLevels[0] != "primary" {
		t.Fatal("school list exposed shared education level backing slice")
	}

	lagosQuery := interfaces.PrimaryAndSecondarySchoolQuery{StateID: "lagos"}
	lagos, err := repo.ListPrimaryAndSecondarySchools(context.Background(), lagosQuery)
	if err != nil {
		t.Fatalf("state filter error = %v", err)
	}
	if len(lagos.Schools) == 0 {
		t.Fatal("expected lagos filter to return records")
	}
	for _, school := range lagos.Schools {
		if school.StateID != "lagos" {
			t.Fatalf("state filter returned wrong record: %#v", school)
		}
	}

	lgaQuery := interfaces.PrimaryAndSecondarySchoolQuery{LGAID: fixture[0].LGAID}
	lga, err := repo.ListPrimaryAndSecondarySchools(context.Background(), lgaQuery)
	if err != nil {
		t.Fatalf("lga filter error = %v", err)
	}
	for _, school := range lga.Schools {
		if school.LGAID != fixture[0].LGAID {
			t.Fatalf("lga filter returned wrong record: %#v", school)
		}
	}

	badQuery := interfaces.PrimaryAndSecondarySchoolQuery{StateID: "lagos", LGAID: fixture[0].LGAID}
	if _, err := repo.ListPrimaryAndSecondarySchools(context.Background(), badQuery); !errors.Is(err, interfaces.ErrInvalidPrimaryAndSecondarySchoolQuery) {
		t.Fatalf("state/lga mismatch error = %v, want ErrInvalidPrimaryAndSecondarySchoolQuery", err)
	}

	level := fixture[0].EducationLevels[0]
	levelQuery := interfaces.PrimaryAndSecondarySchoolQuery{EducationLevel: level}
	levelResult, err := repo.ListPrimaryAndSecondarySchools(context.Background(), levelQuery)
	if err != nil {
		t.Fatalf("education level filter error = %v", err)
	}
	for _, school := range levelResult.Schools {
		if !containsString(school.EducationLevels, level) {
			t.Fatalf("education level filter returned wrong record: %#v", school)
		}
	}

	ownershipQuery := interfaces.PrimaryAndSecondarySchoolQuery{OwnershipType: fixture[0].OwnershipType}
	ownershipResult, err := repo.ListPrimaryAndSecondarySchools(context.Background(), ownershipQuery)
	if err != nil {
		t.Fatalf("ownership filter error = %v", err)
	}
	for _, school := range ownershipResult.Schools {
		if school.OwnershipType != fixture[0].OwnershipType {
			t.Fatalf("ownership filter returned wrong record: %#v", school)
		}
	}

	searchText := strings.Fields(fixture[0].Name)[0]
	searchResult, err := repo.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{Search: searchText})
	if err != nil {
		t.Fatalf("search error = %v", err)
	}
	if len(searchResult.Schools) == 0 {
		t.Fatal("expected search to return records")
	}

	combined, err := repo.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{
		StateID:        fixture[0].StateID,
		LGAID:          fixture[0].LGAID,
		EducationLevel: fixture[0].EducationLevels[0],
		OwnershipType:  fixture[0].OwnershipType,
		Search:         searchText,
		PageSize:       100,
	})
	if err != nil {
		t.Fatalf("combined filter error = %v", err)
	}
	if len(combined.Schools) == 0 {
		t.Fatal("combined filter returned no results")
	}
	if combined.PageSize != 100 {
		t.Fatalf("unexpected page size: %d", combined.PageSize)
	}

	pastEnd, err := repo.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{Page: 999999, PageSize: 50})
	if err != nil {
		t.Fatalf("past-end query error = %v", err)
	}
	if len(pastEnd.Schools) != 0 {
		t.Fatalf("past-end query returned %d schools, want 0", len(pastEnd.Schools))
	}
	if pastEnd.Schools == nil {
		t.Fatal("past-end query returned nil slice")
	}

	if _, err := repo.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{Page: -1, PageSize: 50}); !errors.Is(err, interfaces.ErrInvalidPrimaryAndSecondarySchoolQuery) {
		t.Fatalf("invalid page error = %v", err)
	}
	if _, err := repo.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{Page: 1, PageSize: 101}); !errors.Is(err, interfaces.ErrInvalidPrimaryAndSecondarySchoolQuery) {
		t.Fatalf("invalid page size error = %v", err)
	}

	if _, err := repo.GetPrimaryAndSecondarySchool(context.Background(), fixture[0].ID); err != nil {
		t.Fatalf("GetPrimaryAndSecondarySchool() error = %v", err)
	}
	if _, err := repo.GetPrimaryAndSecondarySchool(context.Background(), "missing-school"); !errors.Is(err, interfaces.ErrPrimaryAndSecondarySchoolNotFound) {
		t.Fatalf("missing school error = %v", err)
	}
}

func TestSchoolRepositoryContextAndCacheIsolation(t *testing.T) {
	t.Parallel()

	fixture := loadEducationSlice[models.PrimaryAndSecondarySchool](t, "../../../datasets/education/primary_and_secondary_schools.json")
	repo := mustNewSchoolRepositoryWithFixture(t, fixture)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repo.ListPrimaryAndSecondarySchools(ctx, interfaces.PrimaryAndSecondarySchoolQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled list error = %v, want context.Canceled", err)
	}

	deadlineCtx, cancelDeadline := context.WithTimeout(context.Background(), time.Nanosecond)
	time.Sleep(time.Nanosecond)
	cancelDeadline()
	if _, err := repo.GetPrimaryAndSecondarySchool(deadlineCtx, fixture[0].ID); !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline get error = %v, want context cancellation/deadline", err)
	}

	stub := &educationJSONRepoStub{pathCalls: map[string]int{}}
	stub.decodeFn = func(ctx context.Context, relativePath string, destination any) error {
		switch dest := destination.(type) {
		case *[]models.PrimaryAndSecondarySchool:
			*dest = clonePrimaryAndSecondarySchoolList(fixture)
			return nil
		default:
			return fmt.Errorf("unexpected destination %T", destination)
		}
	}
	repo, err := NewEducationRepository(stub, "education/universities.json", "education/colleges_of_education.json",
		"education/polytechnics.json", "education/monotechnics.json", "education/colleges_of_agriculture.json",
		"education/colleges_of_health_sciences_and_technology.json", "education/colleges_of_nursing_and_midwifery.json",
		"education/technical_colleges.json", "education/vocational_enterprise_institutions.json",
		"education/primary_and_secondary_schools.json")
	if err != nil {
		t.Fatalf("NewEducationRepository() error = %v", err)
	}
	first, err := repo.GetPrimaryAndSecondarySchool(context.Background(), fixture[0].ID)
	if err != nil {
		t.Fatalf("GetPrimaryAndSecondarySchool() error = %v", err)
	}
	first.EducationLevels[0] = "changed"
	second, err := repo.GetPrimaryAndSecondarySchool(context.Background(), fixture[0].ID)
	if err != nil {
		t.Fatalf("GetPrimaryAndSecondarySchool() second error = %v", err)
	}
	if second.EducationLevels[0] == "changed" {
		t.Fatal("GetPrimaryAndSecondarySchool() exposed shared education level backing slice")
	}
	if stub.pathCalls["education/primary_and_secondary_schools.json"] != 1 {
		t.Fatalf("school decode count = %d, want 1", stub.pathCalls["education/primary_and_secondary_schools.json"])
	}
}

func TestNursingAndTechnicalCollegeMetadataAndSchemaAlignment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		metadataPath     string
		schemaPath       string
		reconciliation   string
		wantRecordCount  int
		wantTitleSnippet string
	}{
		{
			name:             "nursing",
			metadataPath:     "../../../datasets/metadata/education/colleges_of_nursing_and_midwifery.json",
			schemaPath:       "../../../datasets/schemas/education/colleges_of_nursing_and_midwifery.schema.json",
			reconciliation:   "../../../datasets/metadata/education/colleges_of_nursing_and_midwifery_reconciliation.json",
			wantRecordCount:  152,
			wantTitleSnippet: "NMCN December 2025 Snapshot",
		},
		{
			name:             "technical",
			metadataPath:     "../../../datasets/metadata/education/technical_colleges.json",
			schemaPath:       "../../../datasets/schemas/education/technical_colleges.schema.json",
			reconciliation:   "../../../datasets/metadata/education/technical_colleges_reconciliation.json",
			wantRecordCount:  115,
			wantTitleSnippet: "NBTE Directory Snapshot",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			type metadata struct {
				Status      string `json:"status"`
				Snapshot    bool   `json:"snapshot"`
				Title       string `json:"title"`
				Description string `json:"description"`
				RecordCount int    `json:"record_count"`
			}
			type schema struct {
				MinItems int `json:"minItems"`
				MaxItems int `json:"maxItems"`
			}

			var meta metadata
			if err := decodeJSONFileForTest(tc.metadataPath, &meta); err != nil {
				t.Fatalf("decode metadata: %v", err)
			}
			if meta.Status != "active" || !meta.Snapshot || meta.RecordCount != tc.wantRecordCount {
				t.Fatalf("unexpected metadata: %#v", meta)
			}
			if !strings.Contains(meta.Title, tc.wantTitleSnippet) {
				t.Fatalf("metadata title mismatch: %q", meta.Title)
			}
			if meta.Description == "" {
				t.Fatal("metadata description was empty")
			}

			var sch schema
			if err := decodeJSONFileForTest(tc.schemaPath, &sch); err != nil {
				t.Fatalf("decode schema: %v", err)
			}
			if sch.MinItems != tc.wantRecordCount || sch.MaxItems != tc.wantRecordCount {
				t.Fatalf("unexpected schema bounds: %#v", sch)
			}

			if _, err := os.Stat(filepath.Clean(tc.reconciliation)); err != nil {
				t.Fatalf("reconciliation file missing: %v", err)
			}
		})
	}
}

func mustNewSchoolRepositoryWithFixture(t *testing.T, fixture []models.PrimaryAndSecondarySchool) *EducationFileRepository {
	t.Helper()

	stub := &educationJSONRepoStub{pathCalls: map[string]int{}}
	stub.decodeFn = func(ctx context.Context, relativePath string, destination any) error {
		stub.pathCalls[relativePath]++
		switch dest := destination.(type) {
		case *[]models.PrimaryAndSecondarySchool:
			*dest = clonePrimaryAndSecondarySchoolList(fixture)
			return nil
		default:
			return fmt.Errorf("unexpected destination %T", destination)
		}
	}
	repo, err := NewEducationRepository(stub, "education/universities.json", "education/colleges_of_education.json",
		"education/polytechnics.json", "education/monotechnics.json", "education/colleges_of_agriculture.json",
		"education/colleges_of_health_sciences_and_technology.json", "education/colleges_of_nursing_and_midwifery.json",
		"education/technical_colleges.json", "education/vocational_enterprise_institutions.json",
		"education/primary_and_secondary_schools.json")
	if err != nil {
		t.Fatalf("NewEducationRepository() error = %v", err)
	}
	return repo
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func decodeJSONFileForTest(path string, destination any) error {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(destination); err != nil {
		return err
	}
	return nil
}
