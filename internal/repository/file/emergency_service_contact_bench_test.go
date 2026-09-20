package file

import (
	"context"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func BenchmarkEmergencyServiceContactRepositoryFirstLoad(b *testing.B) {
	records := validEmergencyServiceContactRecords()
	for i := 0; i < b.N; i++ {
		repository := mustNewEmergencyServiceContactRepository(b, &emergencyServiceContactJSONStub{records: records})
		if _, err := repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEmergencyServiceContactRepositoryCachedDefaultPage(b *testing.B) {
	repository := mustNewEmergencyServiceContactRepository(b, &emergencyServiceContactJSONStub{records: validEmergencyServiceContactRecords()})
	_, _ = repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEmergencyServiceContactRepositoryCachedDetail(b *testing.B) {
	repository := mustNewEmergencyServiceContactRepository(b, &emergencyServiceContactJSONStub{records: validEmergencyServiceContactRecords()})
	_, _ = repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := repository.GetEmergencyServiceContact(context.Background(), "federal-road-safety-corps-road-emergency-122-national"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEmergencyServiceContactRepositoryFilters(b *testing.B) {
	repository := mustNewEmergencyServiceContactRepository(b, &emergencyServiceContactJSONStub{records: validEmergencyServiceContactRecords()})
	queries := map[string]interfaces.EmergencyServiceContactQuery{
		"service_type":  {ServiceType: "fire"},
		"contact_type":  {ContactType: "short_code"},
		"contact_value": {ContactValue: "112"},
		"combined":      {ServiceType: "fire", ContactType: "short_code", CoverageType: "national", ContactValue: "112"},
		"name_search":   {Search: "federal fire"},
	}
	_, _ = repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{})
	for name, query := range queries {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := repository.ListEmergencyServiceContacts(context.Background(), query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
