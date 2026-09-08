package file

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func benchmarkSchoolRepository(b *testing.B) *EducationFileRepository {
	b.Helper()
	b.ReportAllocs()

	root := filepath.Clean("../../../datasets")
	jsonRepo, err := NewJSONRepository(root, 256<<20)
	if err != nil {
		b.Fatalf("NewJSONRepository() error = %v", err)
	}
	repo, err := NewEducationRepository(
		jsonRepo,
		"education/universities.json",
		"education/colleges_of_education.json",
		"education/polytechnics.json",
		"education/monotechnics.json",
		"education/colleges_of_agriculture.json",
		"education/colleges_of_health_sciences_and_technology.json",
		"education/colleges_of_nursing_and_midwifery.json",
		"education/technical_colleges.json",
		"education/vocational_enterprise_institutions.json",
		"education/primary_and_secondary_schools.json",
	)
	if err != nil {
		b.Fatalf("NewEducationRepository() error = %v", err)
	}
	return repo
}

func BenchmarkSchoolInitialLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		repo := benchmarkSchoolRepository(b)
		if _, err := repo.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{}); err != nil {
			b.Fatalf("ListPrimaryAndSecondarySchools() error = %v", err)
		}
	}
}

func BenchmarkSchoolCachedGet(b *testing.B) {
	repo := benchmarkSchoolRepository(b)
	fixture := loadEducationSliceForBench[models.PrimaryAndSecondarySchool](b, "../../../datasets/education/primary_and_secondary_schools.json")
	if _, err := repo.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{}); err != nil {
		b.Fatalf("warm list error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := repo.GetPrimaryAndSecondarySchool(context.Background(), fixture[0].ID); err != nil {
			b.Fatalf("GetPrimaryAndSecondarySchool() error = %v", err)
		}
	}
}

func BenchmarkSchoolUnfilteredPage(b *testing.B) {
	repo := benchmarkSchoolRepository(b)
	if _, err := repo.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{}); err != nil {
		b.Fatalf("warm list error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := repo.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{Page: 1, PageSize: 50}); err != nil {
			b.Fatalf("ListPrimaryAndSecondarySchools() error = %v", err)
		}
	}
}

func BenchmarkSchoolStateFilteredPage(b *testing.B) {
	repo := benchmarkSchoolRepository(b)
	if _, err := repo.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{}); err != nil {
		b.Fatalf("warm list error = %v", err)
	}

	query := interfaces.PrimaryAndSecondarySchoolQuery{Page: 1, PageSize: 50, StateID: "lagos"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := repo.ListPrimaryAndSecondarySchools(context.Background(), query); err != nil {
			b.Fatalf("ListPrimaryAndSecondarySchools() error = %v", err)
		}
	}
}

func BenchmarkSchoolCombinedFilteredPage(b *testing.B) {
	repo := benchmarkSchoolRepository(b)
	fixture := loadEducationSliceForBench[models.PrimaryAndSecondarySchool](b, "../../../datasets/education/primary_and_secondary_schools.json")
	if _, err := repo.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{}); err != nil {
		b.Fatalf("warm list error = %v", err)
	}

	query := interfaces.PrimaryAndSecondarySchoolQuery{
		Page:           1,
		PageSize:       50,
		StateID:        fixture[0].StateID,
		LGAID:          fixture[0].LGAID,
		EducationLevel: fixture[0].EducationLevels[0],
		OwnershipType:  fixture[0].OwnershipType,
		Search:         fixture[0].Name[:minInt(12, len(fixture[0].Name))],
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := repo.ListPrimaryAndSecondarySchools(context.Background(), query); err != nil {
			b.Fatalf("ListPrimaryAndSecondarySchools() error = %v", err)
		}
	}
}

func BenchmarkSchoolSearchPage(b *testing.B) {
	repo := benchmarkSchoolRepository(b)
	fixture := loadEducationSliceForBench[models.PrimaryAndSecondarySchool](b, "../../../datasets/education/primary_and_secondary_schools.json")
	if _, err := repo.ListPrimaryAndSecondarySchools(context.Background(), interfaces.PrimaryAndSecondarySchoolQuery{}); err != nil {
		b.Fatalf("warm list error = %v", err)
	}

	query := interfaces.PrimaryAndSecondarySchoolQuery{Page: 1, PageSize: 50, Search: minStringPrefix(fixture[0].Name, 16)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := repo.ListPrimaryAndSecondarySchools(context.Background(), query); err != nil {
			b.Fatalf("ListPrimaryAndSecondarySchools() error = %v", err)
		}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func minStringPrefix(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func loadEducationSliceForBench[T any](b testing.TB, path string) []T {
	b.Helper()

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		b.Fatalf("ReadFile(%s) error = %v", path, err)
	}
	var rows []T
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&rows); err != nil {
		b.Fatalf("decode %s error = %v", path, err)
	}
	return rows
}
