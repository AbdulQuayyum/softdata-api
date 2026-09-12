package file

import (
	"context"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func BenchmarkNHIAAccreditedHMORepository(b *testing.B) {
	repository := newRealNHIAAccreditedHMORepository(b)
	first := loadNHIAAccreditedHMOFixture(b)[0]
	if _, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}); err != nil {
		b.Fatal(err)
	}
	benchmarks := []struct {
		name  string
		query interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery
	}{
		{"default_page", interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}},
		{"hmo_id_filter", interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{HMOID: first.HMOID}},
		{"status_filter", interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{AccreditationStatus: first.AccreditationStatus}},
		{"name_search", interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Search: first.Name[:3]}},
		{"combined_filter", interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{HMOID: first.HMOID, AccreditationStatus: first.AccreditationStatus, Search: first.Name[:3]}},
	}
	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), benchmark.query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
	b.Run("cached_detail", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := repository.GetNHIAAccreditedHealthMaintenanceOrganisation(context.Background(), first.ID); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkNHIAAccreditedHMOFirstLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		repository := newRealNHIAAccreditedHMORepository(b)
		b.StartTimer()
		if _, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}); err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
	}
}
