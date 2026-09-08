package file

import (
	"context"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func BenchmarkHealthFacilityRepository(b *testing.B) {
	repository := newRealHealthFacilityRepository(b)
	first, _, _ := allFacilityRecords(b)
	if _, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{}); err != nil {
		b.Fatal(err)
	}
	benchmarks := []struct {
		name  string
		query interfaces.HealthFacilityQuery
	}{
		{"default_page", interfaces.HealthFacilityQuery{}},
		{"state_page", interfaces.HealthFacilityQuery{StateID: first.StateID}},
		{"lga_page", interfaces.HealthFacilityQuery{LGAID: first.LGAID}},
		{"type_page", interfaces.HealthFacilityQuery{FacilityType: first.FacilityType}},
		{"combined_indexed_filters", interfaces.HealthFacilityQuery{StateID: first.StateID, LGAID: first.LGAID, FacilityType: first.FacilityType}},
		{"search_all", interfaces.HealthFacilityQuery{Search: first.Name[:3]}},
		{"search_after_state_lga", interfaces.HealthFacilityQuery{StateID: first.StateID, LGAID: first.LGAID, Search: first.Name[:3]}},
	}
	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := repository.ListHealthFacilities(context.Background(), benchmark.query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	b.Run("cached_detail", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := repository.GetHealthFacility(context.Background(), first.ID); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkHealthFacilityFirstLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		repository := newRealHealthFacilityRepository(b)
		b.StartTimer()
		if _, err := repository.ListHealthFacilities(context.Background(), interfaces.HealthFacilityQuery{}); err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
	}
}
