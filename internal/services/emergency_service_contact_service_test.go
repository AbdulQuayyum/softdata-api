package services

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestEmergencyServiceContactServiceListValidationAndFilters(t *testing.T) {
	stub := &emergencyServiceContactServiceRepositoryStub{}
	service, err := NewEmergencyServiceContactService(stub)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.ListEmergencyServiceContacts(context.Background(), EmergencyServiceContactQuery{
		ServiceType:  " fire ",
		ContactType:  " short_code ",
		CoverageType: " national ",
		ContactValue: " 112 ",
		Search:       " Federal Fire ",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 50, ServiceType: "fire", ContactType: "short_code", CoverageType: "national", ContactValue: "112", Search: "Federal Fire"}
	if !reflect.DeepEqual(stub.queries[0], want) {
		t.Fatalf("query=%#v want %#v", stub.queries[0], want)
	}
	if result.Records == nil {
		t.Fatal("nil records")
	}
	result.Records[0].ServiceName = "mutated"
	again, err := service.ListEmergencyServiceContacts(context.Background(), EmergencyServiceContactQuery{ContactValue: "112"})
	if err != nil {
		t.Fatal(err)
	}
	if len(again.Records) != 2 || again.Records[0].ServiceName == "mutated" {
		t.Fatalf("unexpected repeated 112 result: %#v", again.Records)
	}

	valid := EmergencyServiceContactQuery{Page: 1, PageSize: 100, ServiceType: "police", ContactType: "telephone", CoverageType: "state", ContactValue: "080012345678", Search: strings.Repeat("a", EmergencyServiceContactMaxSearchLength)}
	if _, err := service.ListEmergencyServiceContacts(context.Background(), valid); err != nil {
		t.Fatalf("valid boundary err=%v", err)
	}
}

func TestEmergencyServiceContactServiceRejectsInvalidInputBeforeRepository(t *testing.T) {
	for _, tc := range []struct {
		name  string
		query EmergencyServiceContactQuery
		want  error
	}{
		{"bad page", EmergencyServiceContactQuery{Page: -1}, ErrInvalidEmergencyServiceContactPagination},
		{"bad size", EmergencyServiceContactQuery{PageSize: 101}, ErrInvalidEmergencyServiceContactPagination},
		{"empty service type", EmergencyServiceContactQuery{ServiceType: " "}, ErrInvalidEmergencyServiceContactServiceType},
		{"bad service type", EmergencyServiceContactQuery{ServiceType: "office"}, ErrInvalidEmergencyServiceContactServiceType},
		{"empty contact type", EmergencyServiceContactQuery{ContactType: " "}, ErrInvalidEmergencyServiceContactContactType},
		{"bad contact type", EmergencyServiceContactQuery{ContactType: "email"}, ErrInvalidEmergencyServiceContactContactType},
		{"empty coverage type", EmergencyServiceContactQuery{CoverageType: " "}, ErrInvalidEmergencyServiceContactCoverageType},
		{"bad coverage type", EmergencyServiceContactQuery{CoverageType: "lga"}, ErrInvalidEmergencyServiceContactCoverageType},
		{"empty contact value", EmergencyServiceContactQuery{ContactValue: " "}, ErrInvalidEmergencyServiceContactContactValue},
		{"bad contact value", EmergencyServiceContactQuery{ContactValue: "abc"}, ErrInvalidEmergencyServiceContactContactValue},
		{"empty search", EmergencyServiceContactQuery{Search: " "}, ErrInvalidEmergencyServiceContactSearch},
		{"long search", EmergencyServiceContactQuery{Search: strings.Repeat("a", EmergencyServiceContactMaxSearchLength+1)}, ErrInvalidEmergencyServiceContactSearch},
		{"newline search", EmergencyServiceContactQuery{Search: "fi\nre"}, ErrInvalidEmergencyServiceContactSearch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &emergencyServiceContactServiceRepositoryStub{}
			service, _ := NewEmergencyServiceContactService(stub)
			if _, err := service.ListEmergencyServiceContacts(context.Background(), tc.query); !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want %v", err, tc.want)
			}
			if len(stub.queries) != 0 {
				t.Fatal("repository was called for invalid input")
			}
		})
	}
}

func TestEmergencyServiceContactServiceDetailAndErrorTranslation(t *testing.T) {
	stub := &emergencyServiceContactServiceRepositoryStub{}
	service, _ := NewEmergencyServiceContactService(stub)
	record, err := service.GetEmergencyServiceContact(context.Background(), " federal-road-safety-corps-road-emergency-122-national ")
	if err != nil || record.ID != "federal-road-safety-corps-road-emergency-122-national" || stub.ids[0] != "federal-road-safety-corps-road-emergency-122-national" {
		t.Fatalf("record=%#v ids=%#v err=%v", record, stub.ids, err)
	}
	if _, err := service.GetEmergencyServiceContact(context.Background(), "bad_id"); !errors.Is(err, ErrInvalidEmergencyServiceContactID) {
		t.Fatalf("invalid id err=%v", err)
	}
	if _, err := service.GetEmergencyServiceContact(context.Background(), strings.Repeat("a", 256)); !errors.Is(err, ErrInvalidEmergencyServiceContactID) {
		t.Fatalf("long id err=%v", err)
	}

	for _, tc := range []struct {
		name string
		err  error
		want error
	}{
		{"not found", interfaces.ErrEmergencyServiceContactNotFound, ErrEmergencyServiceContactNotFound},
		{"invalid query", interfaces.ErrInvalidEmergencyServiceContactQuery, ErrInvalidEmergencyServiceContactPagination},
		{"invalid service type", interfaces.ErrInvalidEmergencyServiceContactServiceTypeFilter, ErrInvalidEmergencyServiceContactServiceType},
		{"invalid contact type", interfaces.ErrInvalidEmergencyServiceContactContactTypeFilter, ErrInvalidEmergencyServiceContactContactType},
		{"invalid coverage type", interfaces.ErrInvalidEmergencyServiceContactCoverageTypeFilter, ErrInvalidEmergencyServiceContactCoverageType},
		{"invalid contact value", interfaces.ErrInvalidEmergencyServiceContactContactValueFilter, ErrInvalidEmergencyServiceContactContactValue},
		{"invalid search", interfaces.ErrInvalidEmergencyServiceContactSearch, ErrInvalidEmergencyServiceContactSearch},
		{"dataset", interfaces.ErrInvalidDatasetFile, ErrInvalidEmergencyServiceContactDataset},
		{"unavailable", interfaces.ErrDatasetFileUnavailable, ErrInvalidEmergencyServiceContactDataset},
		{"canceled", context.Canceled, context.Canceled},
		{"deadline", context.DeadlineExceeded, context.DeadlineExceeded},
		{"unexpected", errors.New("/secret/path decoder email"), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub.err = tc.err
			_, err := service.ListEmergencyServiceContacts(context.Background(), EmergencyServiceContactQuery{})
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want %v", err, tc.want)
			}
			if tc.want == nil && (err == nil || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "email")) {
				t.Fatalf("unexpected unsanitized err=%v", err)
			}
		})
	}
}

func TestEmergencyServiceContactServiceConstructorAndContext(t *testing.T) {
	if _, err := NewEmergencyServiceContactService(nil); err == nil {
		t.Fatal("expected nil repository error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service, _ := NewEmergencyServiceContactService(&emergencyServiceContactServiceRepositoryStub{})
	if _, err := service.ListEmergencyServiceContacts(ctx, EmergencyServiceContactQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled list err=%v", err)
	}
	deadline, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.GetEmergencyServiceContact(deadline, "federal-road-safety-corps-road-emergency-122-national"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled detail err=%v", err)
	}
}

type emergencyServiceContactServiceRepositoryStub struct {
	queries []interfaces.EmergencyServiceContactQuery
	ids     []string
	err     error
}

func (s *emergencyServiceContactServiceRepositoryStub) ListEmergencyServiceContacts(_ context.Context, query interfaces.EmergencyServiceContactQuery) (interfaces.EmergencyServiceContactListResult, error) {
	s.queries = append(s.queries, query)
	if s.err != nil {
		return interfaces.EmergencyServiceContactListResult{}, s.err
	}
	return interfaces.EmergencyServiceContactListResult{
		Records: []models.EmergencyServiceContact{
			{ID: "federal-fire-service-fire-112-national", ServiceName: "Federal Fire Service Emergency Response", AgencyName: "Federal Fire Service", ServiceType: "fire", ContactType: "short_code", ContactValue: "112", CoverageType: "national", CountryCode: "NG"},
			{ID: "nigerian-communications-commission-general-emergency-112-national", ServiceName: "112 Emergency Number", AgencyName: "Nigerian Communications Commission", ServiceType: "general_emergency", ContactType: "short_code", ContactValue: "112", CoverageType: "national", CountryCode: "NG"},
		},
		Page:       query.Page,
		PageSize:   query.PageSize,
		Total:      2,
		TotalPages: 1,
	}, nil
}

func (s *emergencyServiceContactServiceRepositoryStub) GetEmergencyServiceContact(_ context.Context, id string) (models.EmergencyServiceContact, error) {
	s.ids = append(s.ids, id)
	if s.err != nil {
		return models.EmergencyServiceContact{}, s.err
	}
	return models.EmergencyServiceContact{ID: id, ServiceName: "Emergency Ambulance Service Scheme (EASS) Zebra", AgencyName: "Federal Road Safety Corps", ServiceType: "road_emergency", ContactType: "short_code", ContactValue: "122", CoverageType: "national", CountryCode: "NG"}, nil
}
