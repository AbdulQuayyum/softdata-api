package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	fileRepo "github.com/AbdulQuayyum/softdata-api/internal/repository/file"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type emergencyAppJSONRepository struct{}

func (emergencyAppJSONRepository) Decode(context.Context, string, any) error { return nil }

type emergencyAppRepository struct{}

func (emergencyAppRepository) ListEmergencyServiceContacts(context.Context, interfaces.EmergencyServiceContactQuery) (interfaces.EmergencyServiceContactListResult, error) {
	return interfaces.EmergencyServiceContactListResult{}, nil
}

func (emergencyAppRepository) GetEmergencyServiceContact(context.Context, string) (models.EmergencyServiceContact, error) {
	return models.EmergencyServiceContact{}, nil
}

func TestEmergencyServiceContactDependencyConstructionUsesDatasetPath(t *testing.T) {
	var gotPath string
	service, err := buildEmergencyServiceContactServiceFromJSONRepository(context.Background(), emergencyAppJSONRepository{},
		func(repository interfaces.JSONFileRepository, recordsPath string) (interfaces.EmergencyServiceContactRepository, error) {
			if repository == nil {
				t.Fatal("json repository was not passed")
			}
			gotPath = recordsPath
			return emergencyAppRepository{}, nil
		},
		func(repository interfaces.EmergencyServiceContactRepository) (emergencyServiceContactService, error) {
			if repository == nil {
				t.Fatal("emergency repository was not passed")
			}
			return services.NewEmergencyServiceContactService(repository)
		},
	)
	if err != nil {
		t.Fatalf("build emergency service contact dependency: %v", err)
	}
	if service == nil {
		t.Fatal("expected service")
	}
	if gotPath != emergencyServiceContactsRelativePath {
		t.Fatalf("unexpected dataset path: got %q want %q", gotPath, emergencyServiceContactsRelativePath)
	}
}

func TestEmergencyServiceContactDependencyConstructionPreservesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := buildEmergencyServiceContactServiceFromJSONRepository(ctx, emergencyAppJSONRepository{},
		func(interfaces.JSONFileRepository, string) (interfaces.EmergencyServiceContactRepository, error) {
			t.Fatal("repository constructor should not be called")
			return nil, nil
		},
		func(interfaces.EmergencyServiceContactRepository) (emergencyServiceContactService, error) {
			t.Fatal("service constructor should not be called")
			return nil, nil
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestEmergencyServiceContactStartupVerification(t *testing.T) {
	if err := verifyEmergencyServiceContacts(context.Background(), newStartupEmergencyServiceContactStub()); err != nil {
		t.Fatalf("verifyEmergencyServiceContacts() error = %v", err)
	}
}

func TestEmergencyServiceContactStartupVerificationFailures(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*startupEmergencyServiceContactStub)
	}{
		{"incorrect total", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["default"]
			page.Total = 4
			s.pages["default"] = page
		}},
		{"incorrect total pages", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["default"]
			page.TotalPages = 4
			s.pages["default"] = page
		}},
		{"empty first page", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["default"]
			page.Records = []models.EmergencyServiceContact{}
			s.pages["default"] = page
		}},
		{"nil first page", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["default"]
			page.Records = nil
			s.pages["default"] = page
		}},
		{"incorrect first anchor", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["default"]
			page.Records[0].ID = "wrong-anchor"
			s.pages["default"] = page
		}},
		{"incorrect last anchor", func(s *startupEmergencyServiceContactStub) {
			record := s.details[emergencyServiceContactLastID]
			record.ID = "wrong-last-anchor"
			s.details[emergencyServiceContactLastID] = record
		}},
		{"missing detail anchor", func(s *startupEmergencyServiceContactStub) { delete(s.details, emergencyServiceContactLastID) }},
		{"invalid record contract", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["default"]
			page.Records[0].CountryCode = "US"
			s.pages["default"] = page
		}},
		{"incorrect 112 total", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["contact_value:112"]
			page.Total = 3
			s.pages["contact_value:112"] = page
		}},
		{"only one 112 result", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["contact_value:112"]
			page.Records = page.Records[:1]
			s.pages["contact_value:112"] = page
		}},
		{"duplicate 112 ids", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["contact_value:112"]
			page.Records[1].ID = page.Records[0].ID
			s.pages["contact_value:112"] = page
		}},
		{"unexpected 112 contact value", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["contact_value:112"]
			page.Records[1].ContactValue = "999"
			s.pages["contact_value:112"] = page
		}},
		{"incorrect service-type total", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["service_type:fire"]
			page.Total = 1
			s.pages["service_type:fire"] = page
		}},
		{"incorrect national-coverage total", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["coverage_type:national"]
			page.Total = 4
			s.pages["coverage_type:national"] = page
		}},
		{"incorrect contact-type total", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["contact_type:short_code"]
			page.Total = 2
			s.pages["contact_type:short_code"] = page
		}},
		{"non-empty beyond-final page", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["beyond"]
			page.Records = []models.EmergencyServiceContact{emergencyStartupFirstContact()}
			s.pages["beyond"] = page
		}},
		{"nil beyond-final slice", func(s *startupEmergencyServiceContactStub) {
			page := s.pages["beyond"]
			page.Records = nil
			s.pages["beyond"] = page
		}},
		{"list failure", func(s *startupEmergencyServiceContactStub) { s.listErrKey = "default" }},
		{"detail failure", func(s *startupEmergencyServiceContactStub) { s.detailErrID = emergencyServiceContactLastID }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := newStartupEmergencyServiceContactStub()
			tc.mutate(stub)
			err := verifyEmergencyServiceContacts(context.Background(), stub)
			if err == nil || !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("verifyEmergencyServiceContacts() error = %v", err)
			}
		})
	}
}

func TestEmergencyServiceContactStartupVerificationPreservesContextErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := verifyEmergencyServiceContacts(ctx, newStartupEmergencyServiceContactStub()); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled context error = %v", err)
	}

	stub := newStartupEmergencyServiceContactStub()
	stub.listErrKey = "default"
	stub.err = context.DeadlineExceeded
	if err := verifyEmergencyServiceContacts(context.Background(), stub); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline error = %v", err)
	}
}

func TestEmergencyServiceContactStartupVerificationSanitizesUnexpectedErrors(t *testing.T) {
	stub := newStartupEmergencyServiceContactStub()
	stub.listErrKey = "default"
	stub.err = errors.New("open /tmp/private/emergency_service_contacts.json: raw decoder detail")
	err := verifyEmergencyServiceContacts(context.Background(), stub)
	if err == nil || !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("verifyEmergencyServiceContacts() error = %v", err)
	}
	if strings.Contains(err.Error(), "/tmp/private") || strings.Contains(err.Error(), "raw decoder") {
		t.Fatalf("unexpected detail leaked: %v", err)
	}
}

func TestEmergencyServiceContactStartupAndHTTPShareRepositoryCache(t *testing.T) {
	jsonRepo := &countingEmergencyJSONRepository{records: emergencyStartupDataset()}
	repository, err := fileRepo.NewEmergencyServiceContactRepository(jsonRepo, emergencyServiceContactsRelativePath)
	if err != nil {
		t.Fatalf("NewEmergencyServiceContactRepository() error = %v", err)
	}
	service, err := services.NewEmergencyServiceContactService(repository)
	if err != nil {
		t.Fatalf("NewEmergencyServiceContactService() error = %v", err)
	}
	if err := verifyEmergencyServiceContacts(context.Background(), service); err != nil {
		t.Fatalf("verifyEmergencyServiceContacts() error = %v", err)
	}
	if got := jsonRepo.calls(); got != 1 {
		t.Fatalf("decode calls after startup = %d, want 1", got)
	}
	if _, err := service.GetEmergencyServiceContact(context.Background(), emergencyServiceContactFirstID); err != nil {
		t.Fatalf("GetEmergencyServiceContact() error = %v", err)
	}
	if _, err := service.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{Page: 1, PageSize: 2, ContactValue: "112"}); err != nil {
		t.Fatalf("ListEmergencyServiceContacts() error = %v", err)
	}
	if got := jsonRepo.calls(); got != 1 {
		t.Fatalf("decode calls after later access = %d, want 1", got)
	}
	if strings.Contains(strings.Join(jsonRepo.paths(), ","), "reconciliation") {
		t.Fatalf("reconciliation metadata was opened: %v", jsonRepo.paths())
	}
}

type startupEmergencyServiceContactStub struct {
	pages       map[string]interfaces.EmergencyServiceContactListResult
	details     map[string]models.EmergencyServiceContact
	listErrKey  string
	detailErrID string
	err         error
}

func newStartupEmergencyServiceContactStub() *startupEmergencyServiceContactStub {
	first := emergencyStartupFirstContact()
	firePhone := models.EmergencyServiceContact{ID: "federal-fire-service-fire-2348032003557-national", ServiceName: "Federal Fire Service Emergency Response", AgencyName: "Federal Fire Service", ServiceType: "fire", ContactType: "telephone", ContactValue: "+2348032003557", CoverageType: "national", CountryCode: "NG"}
	road := models.EmergencyServiceContact{ID: "federal-road-safety-corps-road-emergency-122-national", ServiceName: "Emergency Ambulance Service Scheme (EASS) Zebra", AgencyName: "Federal Road Safety Corps", ServiceType: "road_emergency", ContactType: "short_code", ContactValue: "122", CoverageType: "national", CountryCode: "NG"}
	disaster := models.EmergencyServiceContact{ID: "national-emergency-management-agency-disaster-management-080022556362-national", ServiceName: "NEMA Emergency Contact Line", AgencyName: "National Emergency Management Agency", ServiceType: "disaster_management", ContactType: "telephone", ContactValue: "080022556362", CoverageType: "national", CountryCode: "NG"}
	last := emergencyStartupLastContact()
	return &startupEmergencyServiceContactStub{
		pages: map[string]interfaces.EmergencyServiceContactListResult{
			"default":                          {Records: []models.EmergencyServiceContact{first}, Page: 1, PageSize: 1, Total: 5, TotalPages: 5},
			"contact_value:112":                {Records: []models.EmergencyServiceContact{first, last}, Page: 1, PageSize: 2, Total: 2, TotalPages: 1},
			"service_type:fire":                {Records: []models.EmergencyServiceContact{first}, Page: 1, PageSize: 1, Total: 2, TotalPages: 2},
			"service_type:general_emergency":   {Records: []models.EmergencyServiceContact{last}, Page: 1, PageSize: 1, Total: 1, TotalPages: 1},
			"service_type:road_emergency":      {Records: []models.EmergencyServiceContact{road}, Page: 1, PageSize: 1, Total: 1, TotalPages: 1},
			"service_type:disaster_management": {Records: []models.EmergencyServiceContact{disaster}, Page: 1, PageSize: 1, Total: 1, TotalPages: 1},
			"coverage_type:national":           {Records: []models.EmergencyServiceContact{first}, Page: 1, PageSize: 1, Total: 5, TotalPages: 5},
			"contact_type:short_code":          {Records: []models.EmergencyServiceContact{first}, Page: 1, PageSize: 1, Total: 3, TotalPages: 3},
			"contact_type:telephone":           {Records: []models.EmergencyServiceContact{firePhone}, Page: 1, PageSize: 1, Total: 2, TotalPages: 2},
			"beyond":                           {Records: []models.EmergencyServiceContact{}, Page: 6, PageSize: 1, Total: 5, TotalPages: 5},
		},
		details: map[string]models.EmergencyServiceContact{
			first.ID: first,
			last.ID:  last,
		},
		err: errors.New("boom"),
	}
}

func (s *startupEmergencyServiceContactStub) ListEmergencyServiceContacts(_ context.Context, q services.EmergencyServiceContactQuery) (services.EmergencyServiceContactListResult, error) {
	key := "default"
	switch {
	case q.ContactValue != "":
		key = "contact_value:" + q.ContactValue
	case q.ServiceType != "":
		key = "service_type:" + q.ServiceType
	case q.CoverageType != "":
		key = "coverage_type:" + q.CoverageType
	case q.ContactType != "":
		key = "contact_type:" + q.ContactType
	case q.Page == 6:
		key = "beyond"
	}
	if s.listErrKey == key {
		return services.EmergencyServiceContactListResult{}, s.err
	}
	return s.pages[key], nil
}

func (s *startupEmergencyServiceContactStub) GetEmergencyServiceContact(_ context.Context, id string) (models.EmergencyServiceContact, error) {
	if s.detailErrID == id {
		return models.EmergencyServiceContact{}, s.err
	}
	record, ok := s.details[id]
	if !ok {
		return models.EmergencyServiceContact{}, services.ErrEmergencyServiceContactNotFound
	}
	return record, nil
}

type countingEmergencyJSONRepository struct {
	mu      sync.Mutex
	records []models.EmergencyServiceContact
	n       int
	opened  []string
}

func (r *countingEmergencyJSONRepository) Decode(ctx context.Context, path string, dst any) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if strings.Contains(path, "reconciliation") {
		return errors.New("reconciliation should not be opened")
	}
	r.mu.Lock()
	r.n++
	r.opened = append(r.opened, path)
	r.mu.Unlock()
	payload, err := json.Marshal(r.records)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(payload))
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func (r *countingEmergencyJSONRepository) calls() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.n
}

func (r *countingEmergencyJSONRepository) paths() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	paths := make([]string, len(r.opened))
	copy(paths, r.opened)
	return paths
}

func emergencyStartupDataset() []models.EmergencyServiceContact {
	return []models.EmergencyServiceContact{
		emergencyStartupFirstContact(),
		{ID: "federal-fire-service-fire-2348032003557-national", ServiceName: "Federal Fire Service Emergency Response", AgencyName: "Federal Fire Service", ServiceType: "fire", ContactType: "telephone", ContactValue: "+2348032003557", CoverageType: "national", CountryCode: "NG"},
		{ID: "federal-road-safety-corps-road-emergency-122-national", ServiceName: "Emergency Ambulance Service Scheme (EASS) Zebra", AgencyName: "Federal Road Safety Corps", ServiceType: "road_emergency", ContactType: "short_code", ContactValue: "122", CoverageType: "national", CountryCode: "NG"},
		{ID: "national-emergency-management-agency-disaster-management-080022556362-national", ServiceName: "NEMA Emergency Contact Line", AgencyName: "National Emergency Management Agency", ServiceType: "disaster_management", ContactType: "telephone", ContactValue: "080022556362", CoverageType: "national", CountryCode: "NG"},
		emergencyStartupLastContact(),
	}
}

func emergencyStartupFirstContact() models.EmergencyServiceContact {
	return models.EmergencyServiceContact{ID: emergencyServiceContactFirstID, ServiceName: emergencyServiceContactFirstService, AgencyName: emergencyServiceContactFirstAgency, ServiceType: emergencyServiceContactFirstType, ContactType: emergencyServiceContactFirstContact, ContactValue: emergencyServiceContactFirstValue, CoverageType: "national", CountryCode: "NG"}
}

func emergencyStartupLastContact() models.EmergencyServiceContact {
	return models.EmergencyServiceContact{ID: emergencyServiceContactLastID, ServiceName: emergencyServiceContactLastService, AgencyName: emergencyServiceContactLastAgency, ServiceType: emergencyServiceContactLastType, ContactType: emergencyServiceContactLastContact, ContactValue: emergencyServiceContactLastValue, CoverageType: "national", CountryCode: "NG"}
}
