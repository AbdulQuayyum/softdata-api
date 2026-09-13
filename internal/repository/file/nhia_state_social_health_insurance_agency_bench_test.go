package file

import (
	"context"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func BenchmarkNHIAStateSocialHealthInsuranceAgencyRepository(b *testing.B) {
	repository := newRealNHIASSHIARepository(b)
	first := validNHIASSHIARecords()[0]
	if _, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}); err != nil {
		b.Fatal(err)
	}
	benchmarks := []struct {
		name  string
		query interfaces.NHIAStateSocialHealthInsuranceAgencyQuery
	}{
		{"default_page", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}},
		{"state_filter", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{StateID: first.StateID}},
		{"name_search", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Search: first.Name[:3]}},
		{"combined_filter", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{StateID: first.StateID, Search: first.Name[:3]}},
	}
	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), benchmark.query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	b.Run("cached_detail", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := repository.GetNHIAStateSocialHealthInsuranceAgency(context.Background(), first.ID); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkNHIAStateSocialHealthInsuranceAgencyFirstLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		repository := newRealNHIASSHIARepository(b)
		b.StartTimer()
		if _, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}); err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
	}
}

func newRealNHIASSHIARepository(b testing.TB) *NHIAStateSocialHealthInsuranceAgencyFileRepository {
	b.Helper()
	jsonRepository, err := NewJSONRepository("../../../datasets", 64<<20)
	if err != nil {
		b.Fatal(err)
	}
	repository, err := NewNHIAStateSocialHealthInsuranceAgencyRepository(jsonRepository, nhiaSSHIATestPath)
	if err != nil {
		b.Fatal(err)
	}
	return repository
}
