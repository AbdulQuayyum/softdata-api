package file

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/datasets"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const emergencyServiceContactTestPath = "emergency/emergency_service_contacts.json"

func TestEmergencyServiceContactRepositoryListGetAndFilters(t *testing.T) {
	repository := mustNewEmergencyServiceContactRepository(t, &emergencyServiceContactJSONStub{records: validEmergencyServiceContactRecords()})
	result, err := repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Records) != 5 || result.Page != 1 || result.PageSize != 50 || result.Total != 5 || result.TotalPages != 1 {
		t.Fatalf("default page mismatch: %#v", result)
	}
	if result.Records == nil {
		t.Fatal("nil records")
	}
	for _, tc := range []struct {
		name  string
		query interfaces.EmergencyServiceContactQuery
		total int
		first string
	}{
		{"maximum page size", interfaces.EmergencyServiceContactQuery{PageSize: 100}, 5, "federal-fire-service-fire-112-national"},
		{"beyond final", interfaces.EmergencyServiceContactQuery{Page: 2, PageSize: 5}, 5, ""},
		{"service type", interfaces.EmergencyServiceContactQuery{ServiceType: "fire"}, 2, "federal-fire-service-fire-112-national"},
		{"contact type", interfaces.EmergencyServiceContactQuery{ContactType: "telephone"}, 2, "federal-fire-service-fire-2348032003557-national"},
		{"coverage type", interfaces.EmergencyServiceContactQuery{CoverageType: "national"}, 5, "federal-fire-service-fire-112-national"},
		{"contact value repeated", interfaces.EmergencyServiceContactQuery{ContactValue: " 112 "}, 2, "federal-fire-service-fire-112-national"},
		{"contact value telephone", interfaces.EmergencyServiceContactQuery{ContactValue: " +2348032003557 "}, 1, "federal-fire-service-fire-2348032003557-national"},
		{"search service", interfaces.EmergencyServiceContactQuery{Search: " ambulance "}, 1, "federal-road-safety-corps-road-emergency-122-national"},
		{"search agency case insensitive", interfaces.EmergencyServiceContactQuery{Search: "nigerian communications"}, 1, "nigerian-communications-commission-general-emergency-112-national"},
		{"combined filters", interfaces.EmergencyServiceContactQuery{ServiceType: "fire", ContactType: "short_code", CoverageType: "national", ContactValue: "112", Search: "federal fire"}, 1, "federal-fire-service-fire-112-national"},
		{"combined no result", interfaces.EmergencyServiceContactQuery{ServiceType: "fire", ContactValue: "122"}, 0, ""},
		{"filter before pagination", interfaces.EmergencyServiceContactQuery{ContactValue: "112", Page: 2, PageSize: 1}, 2, "nigerian-communications-commission-general-emergency-112-national"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repository.ListEmergencyServiceContacts(context.Background(), tc.query)
			if err != nil {
				t.Fatal(err)
			}
			if got.Total != tc.total {
				t.Fatalf("total=%d want %d", got.Total, tc.total)
			}
			if tc.first == "" {
				if got.Records == nil || len(got.Records) != 0 {
					t.Fatalf("expected non-nil empty page: %#v", got.Records)
				}
				return
			}
			if len(got.Records) == 0 || got.Records[0].ID != tc.first {
				t.Fatalf("first=%#v want %s", got.Records, tc.first)
			}
		})
	}
	repeated, err := repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{ContactValue: "112"})
	if err != nil || repeated.Total != 2 || len(repeated.Records) != 2 {
		t.Fatalf("repeated 112 result=%#v err=%v", repeated, err)
	}
	if repeated.Records[0].ID == repeated.Records[1].ID {
		t.Fatal("contact value duplicate was collapsed")
	}
	detail, err := repository.GetEmergencyServiceContact(context.Background(), "federal-road-safety-corps-road-emergency-122-national")
	if err != nil || detail.ContactValue != "122" {
		t.Fatalf("detail=%#v err=%v", detail, err)
	}
	if _, err := repository.GetEmergencyServiceContact(context.Background(), "unknown-valid-id"); !errors.Is(err, interfaces.ErrEmergencyServiceContactNotFound) {
		t.Fatalf("unknown err=%v", err)
	}
	if _, err := repository.GetEmergencyServiceContact(context.Background(), "bad_id"); !errors.Is(err, interfaces.ErrInvalidEmergencyServiceContactQuery) {
		t.Fatalf("malformed err=%v", err)
	}
}

func TestEmergencyServiceContactRepositoryValidationAndStrictDecode(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func([]models.EmergencyServiceContact) []models.EmergencyServiceContact
	}{
		{"invalid count", func(r []models.EmergencyServiceContact) []models.EmergencyServiceContact { return r[:4] }},
		{"duplicate id", func(r []models.EmergencyServiceContact) []models.EmergencyServiceContact { r[1].ID = r[0].ID; return r }},
		{"invalid country", func(r []models.EmergencyServiceContact) []models.EmergencyServiceContact {
			r[0].CountryCode = "GH"
			return r
		}},
		{"empty service", func(r []models.EmergencyServiceContact) []models.EmergencyServiceContact {
			r[0].ServiceName = ""
			return r
		}},
		{"empty agency", func(r []models.EmergencyServiceContact) []models.EmergencyServiceContact {
			r[0].AgencyName = " "
			return r
		}},
		{"unsupported service type", func(r []models.EmergencyServiceContact) []models.EmergencyServiceContact {
			r[0].ServiceType = "office"
			return r
		}},
		{"unsupported contact type", func(r []models.EmergencyServiceContact) []models.EmergencyServiceContact {
			r[0].ContactType = "email"
			return r
		}},
		{"unsupported coverage", func(r []models.EmergencyServiceContact) []models.EmergencyServiceContact {
			r[0].CoverageType = "lga"
			return r
		}},
		{"state coverage rejected", func(r []models.EmergencyServiceContact) []models.EmergencyServiceContact {
			r[0].CoverageType = "state"
			return r
		}},
		{"national state id rejected", func(r []models.EmergencyServiceContact) []models.EmergencyServiceContact {
			r[0].StateID = "lagos"
			return r
		}},
		{"invalid availability", func(r []models.EmergencyServiceContact) []models.EmergencyServiceContact {
			r[0].Availability = "daily"
			return r
		}},
		{"invalid call cost", func(r []models.EmergencyServiceContact) []models.EmergencyServiceContact {
			r[0].CallCost = "free"
			return r
		}},
		{"invalid short code", func(r []models.EmergencyServiceContact) []models.EmergencyServiceContact {
			r[0].ContactValue = "91"
			return r
		}},
		{"bad ordering", func(r []models.EmergencyServiceContact) []models.EmergencyServiceContact {
			r[0], r[1] = r[1], r[0]
			return r
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repository := mustNewEmergencyServiceContactRepository(t, &emergencyServiceContactJSONStub{records: tc.mutate(validEmergencyServiceContactRecords())})
			if _, err := repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("err=%v", err)
			}
		})
	}
	for _, payload := range []string{`[{"id":"synthetic-id","service_name":"Synthetic","agency_name":"Synthetic Agency","service_type":"fire","contact_type":"short_code","contact_value":"123","coverage_type":"national","country_code":"NG","email":"hidden"}]`, `[] {}`} {
		repository := mustNewEmergencyServiceContactRepository(t, &emergencyServiceContactJSONStub{raw: payload})
		if _, err := repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) || strings.Contains(err.Error(), "email") {
			t.Fatalf("strict decode err=%v", err)
		}
	}
	repository := mustNewEmergencyServiceContactRepository(t, &emergencyServiceContactJSONStub{err: errors.New("/secret/path decoder email")})
	if _, err := repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "email") {
		t.Fatalf("unsanitized err=%v", err)
	}
}

func TestEmergencyServiceContactRepositoryInvalidQueriesContextCacheAndConcurrency(t *testing.T) {
	repository := mustNewEmergencyServiceContactRepository(t, &emergencyServiceContactJSONStub{records: validEmergencyServiceContactRecords()})
	for _, query := range []interfaces.EmergencyServiceContactQuery{
		{Page: -1}, {PageSize: -1}, {PageSize: 101}, {ServiceType: "office"}, {ContactType: "email"}, {CoverageType: "lga"},
		{ContactValue: "abc"}, {Search: strings.Repeat("a", emergencyServiceContactMaxSearch+1)}, {Search: "\n"},
	} {
		if _, err := repository.ListEmergencyServiceContacts(context.Background(), query); !errors.Is(err, interfaces.ErrInvalidEmergencyServiceContactQuery) &&
			!errors.Is(err, interfaces.ErrInvalidEmergencyServiceContactServiceTypeFilter) &&
			!errors.Is(err, interfaces.ErrInvalidEmergencyServiceContactContactTypeFilter) &&
			!errors.Is(err, interfaces.ErrInvalidEmergencyServiceContactCoverageTypeFilter) &&
			!errors.Is(err, interfaces.ErrInvalidEmergencyServiceContactContactValueFilter) &&
			!errors.Is(err, interfaces.ErrInvalidEmergencyServiceContactSearch) {
			t.Fatalf("query %#v err=%v", query, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repository = mustNewEmergencyServiceContactRepository(t, &emergencyServiceContactJSONStub{records: validEmergencyServiceContactRecords()})
	if _, err := repository.ListEmergencyServiceContacts(ctx, interfaces.EmergencyServiceContactQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled err=%v", err)
	}
	deadline, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()
	if _, err := repository.ListEmergencyServiceContacts(deadline, interfaces.EmergencyServiceContactQuery{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline err=%v", err)
	}
	stub := &emergencyServiceContactJSONStub{records: validEmergencyServiceContactRecords(), errFirst: errors.New("temporary")}
	repository = mustNewEmergencyServiceContactRepository(t, stub)
	if _, err := repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("first err=%v", err)
	}
	if _, err := repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{}); err != nil {
		t.Fatalf("retry err=%v", err)
	}
	if stub.calls != 2 {
		t.Fatalf("decode calls=%d", stub.calls)
	}
	result, _ := repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{})
	result.Records[0].ServiceName = "mutated"
	again, _ := repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{})
	if again.Records[0].ServiceName == "mutated" {
		t.Fatal("caller mutation affected cache")
	}
	if _, err := repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{}); err != nil {
		t.Fatal(err)
	}
	if stub.calls != 2 {
		t.Fatalf("cached decode calls=%d", stub.calls)
	}
	stub = &emergencyServiceContactJSONStub{records: validEmergencyServiceContactRecords()}
	repository = mustNewEmergencyServiceContactRepository(t, stub)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = repository.GetEmergencyServiceContact(context.Background(), "federal-road-safety-corps-road-emergency-122-national")
		}()
	}
	wg.Wait()
	if stub.calls != 1 {
		t.Fatalf("decode calls=%d", stub.calls)
	}
}

func TestEmergencyServiceContactRepositoryEmbeddedDataset(t *testing.T) {
	jsonRepository, err := NewEmbeddedJSONRepository(datasets.Files(), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := NewEmergencyServiceContactRepository(jsonRepository, emergencyServiceContactTestPath)
	if err != nil {
		t.Fatal(err)
	}
	result, err := repository.ListEmergencyServiceContacts(context.Background(), interfaces.EmergencyServiceContactQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Records) != 5 || result.Total != 5 {
		t.Fatalf("embedded result mismatch: %#v", result)
	}
}

func mustNewEmergencyServiceContactRepository(t testing.TB, stub *emergencyServiceContactJSONStub) *EmergencyServiceContactFileRepository {
	t.Helper()
	repository, err := NewEmergencyServiceContactRepository(stub, emergencyServiceContactTestPath)
	if err != nil {
		t.Fatal(err)
	}
	return repository
}

type emergencyServiceContactJSONStub struct {
	records  []models.EmergencyServiceContact
	raw      string
	err      error
	errFirst error
	calls    int
	mu       sync.Mutex
}

func (s *emergencyServiceContactJSONStub) Decode(ctx context.Context, _ string, dst any) error {
	s.mu.Lock()
	s.calls++
	call := s.calls
	s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.errFirst != nil && call == 1 {
		return s.errFirst
	}
	if s.err != nil {
		return s.err
	}
	var data []byte
	if s.raw != "" {
		data = []byte(s.raw)
	} else {
		data, _ = json.Marshal(s.records)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return interfaces.ErrInvalidDatasetFile
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		return interfaces.ErrInvalidDatasetFile
	}
	return nil
}

func validEmergencyServiceContactRecords() []models.EmergencyServiceContact {
	var records []models.EmergencyServiceContact
	if err := json.Unmarshal(readEmergencyServiceContactDataset(), &records); err != nil {
		panic(err)
	}
	return records
}

func readEmergencyServiceContactDataset() []byte {
	data, err := fs.ReadFile(datasets.Files(), emergencyServiceContactTestPath)
	if err != nil {
		panic(err)
	}
	return data
}
