package file

import (
	"context"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/datasets"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func BenchmarkNHIAActiveAccreditedHealthcareProviderRepositoryFirstLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		jsonRepository, _ := NewEmbeddedJSONRepository(datasets.Files(), 64<<20)
		repository, _ := NewNHIAActiveAccreditedHealthcareProviderRepository(jsonRepository, nhiaHCPTestPath)
		_, _ = repository.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), interfaces.NHIAActiveAccreditedHealthcareProviderQuery{})
	}
}

func BenchmarkNHIAActiveAccreditedHealthcareProviderRepositoryCached(b *testing.B) {
	jsonRepository, _ := NewEmbeddedJSONRepository(datasets.Files(), 64<<20)
	repository, _ := NewNHIAActiveAccreditedHealthcareProviderRepository(jsonRepository, nhiaHCPTestPath)
	_, _ = repository.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), interfaces.NHIAActiveAccreditedHealthcareProviderQuery{})
	benchmarks := []struct {
		name  string
		query interfaces.NHIAActiveAccreditedHealthcareProviderQuery
	}{
		{"default page", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}},
		{"provider code", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{ProviderCode: "FCT/0001/P"}},
		{"facility type", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{FacilityType: "primary", PageSize: 100}},
		{"listing status", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{ListingStatus: "active_accredited", PageSize: 100}},
		{"search", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Search: "clinic"}},
		{"combined", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{ProviderCode: "FCT/0001/P", FacilityType: "primary", ListingStatus: "active_accredited", Search: "clinic"}},
	}
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, _ = repository.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), bm.query)
			}
		})
	}
	b.Run("detail", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = repository.GetNHIAActiveAccreditedHealthcareProvider(context.Background(), "fct-0001-p")
		}
	})
}
