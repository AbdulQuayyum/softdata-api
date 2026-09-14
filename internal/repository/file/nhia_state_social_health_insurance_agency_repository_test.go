package file

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/datasets"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const nhiaSSHIATestPath = "healthcare/nhia_state_social_health_insurance_agencies.json"

func TestNHIAStateSocialHealthInsuranceAgencyRepositoryListAndGet(t *testing.T) {
	repository := mustNewNHIASSHIARepository(t, &nhiaSSHIAJSONStub{records: validNHIASSHIARecords()})
	result, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Records) != 37 || result.Page != 1 || result.PageSize != 50 || result.Total != 37 || result.TotalPages != 1 {
		t.Fatalf("default page mismatch: %#v", result)
	}
	if result.Records == nil {
		t.Fatal("nil records")
	}

	for _, tc := range []struct {
		name  string
		query interfaces.NHIAStateSocialHealthInsuranceAgencyQuery
		total int
		first string
	}{
		{"custom page size", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 1, PageSize: 10}, 37, "abia-state-health-insurance-agency-abshia"},
		{"beyond final", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Page: 2, PageSize: 100}, 37, ""},
		{"state akwa ibom", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{StateID: "akwa-ibom"}, 1, "akwa-ibom-state-health-insurance-agency"},
		{"state cross river", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{StateID: "cross-river"}, 1, "cross-river-state-health-insurance-agency-crshia"},
		{"state fct", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{StateID: "fct"}, 1, "fct-health-insurance-scheme-fhis"},
		{"state rivers", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{StateID: "rivers"}, 1, "rivchpp-rsmoh"},
		{"name search", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Search: "Insurance Scheme"}, 4, "bayelsa-health-insurance-scheme-bhis"},
		{"case-insensitive trimmed search", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{Search: "  lashma  "}, 1, "lashma"},
		{"combined filter", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{StateID: " lagos ", Search: " LASHMA "}, 1, "lashma"},
		{"filter before pagination", interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{StateID: "lagos", Page: 2, PageSize: 1}, 1, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), tc.query)
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

	detail, err := repository.GetNHIAStateSocialHealthInsuranceAgency(context.Background(), "abia-state-health-insurance-agency-abshia")
	if err != nil || detail.StateID != "abia" {
		t.Fatalf("detail=%#v err=%v", detail, err)
	}
	if _, err := repository.GetNHIAStateSocialHealthInsuranceAgency(context.Background(), "unknown-valid-id"); !errors.Is(err, interfaces.ErrNHIAStateSocialHealthInsuranceAgencyNotFound) {
		t.Fatalf("unknown err=%v", err)
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyRepositoryValidationAndSanitization(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func([]models.NHIAStateSocialHealthInsuranceAgency) []models.NHIAStateSocialHealthInsuranceAgency
	}{
		{"invalid country", func(r []models.NHIAStateSocialHealthInsuranceAgency) []models.NHIAStateSocialHealthInsuranceAgency {
			r[0].CountryCode = "GH"
			return r
		}},
		{"invalid type", func(r []models.NHIAStateSocialHealthInsuranceAgency) []models.NHIAStateSocialHealthInsuranceAgency {
			r[0].OrganisationType = "licensed_agency"
			return r
		}},
		{"invalid state", func(r []models.NHIAStateSocialHealthInsuranceAgency) []models.NHIAStateSocialHealthInsuranceAgency {
			r[0].StateID = "aks"
			return r
		}},
		{"duplicate id", func(r []models.NHIAStateSocialHealthInsuranceAgency) []models.NHIAStateSocialHealthInsuranceAgency {
			r[1].ID = r[0].ID
			return r
		}},
		{"duplicate state", func(r []models.NHIAStateSocialHealthInsuranceAgency) []models.NHIAStateSocialHealthInsuranceAgency {
			r[1].StateID = r[0].StateID
			return r
		}},
		{"missing state", func(r []models.NHIAStateSocialHealthInsuranceAgency) []models.NHIAStateSocialHealthInsuranceAgency {
			return r[:36]
		}},
		{"prohibited field text", func(r []models.NHIAStateSocialHealthInsuranceAgency) []models.NHIAStateSocialHealthInsuranceAgency {
			r[0].Name = "Director Contact"
			return r
		}},
		{"bad ordering", func(r []models.NHIAStateSocialHealthInsuranceAgency) []models.NHIAStateSocialHealthInsuranceAgency {
			r[0], r[1] = r[1], r[0]
			return r
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			records := validNHIASSHIARecords()
			repository := mustNewNHIASSHIARepository(t, &nhiaSSHIAJSONStub{records: tc.mutate(records)})
			if _, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("err=%v", err)
			}
		})
	}

	for _, payload := range []string{`[{"id":"extra","name":"Extra","state_id":"abia","country_code":"NG","organisation_type":"state_social_health_insurance_agency","director":"hidden"}]`, `[] {}`} {
		repository := mustNewNHIASSHIARepository(t, &nhiaSSHIAJSONStub{raw: payload})
		if _, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) || strings.Contains(err.Error(), "director") {
			t.Fatalf("strict decode err=%v", err)
		}
	}
	repository := mustNewNHIASSHIARepository(t, &nhiaSSHIAJSONStub{err: errors.New("/secret/path decoder director")})
	if _, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "director") {
		t.Fatalf("unsanitized err=%v", err)
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyRepositoryCacheOwnershipContextAndConcurrency(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repository := mustNewNHIASSHIARepository(t, &nhiaSSHIAJSONStub{records: validNHIASSHIARecords()})
	if _, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(ctx, interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled err=%v", err)
	}
	deadline, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()
	if _, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(deadline, interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline err=%v", err)
	}

	stub := &nhiaSSHIAJSONStub{records: validNHIASSHIARecords(), errFirst: errors.New("temporary")}
	repository = mustNewNHIASSHIARepository(t, stub)
	if _, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("first err=%v", err)
	}
	if _, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{}); err != nil {
		t.Fatalf("retry err=%v", err)
	}
	if stub.calls != 2 {
		t.Fatalf("decode calls=%d", stub.calls)
	}

	result, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{})
	if err != nil {
		t.Fatal(err)
	}
	result.Records[0].Name = "mutated"
	again, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if again.Records[0].Name == "mutated" {
		t.Fatal("caller mutation affected cache")
	}

	stub = &nhiaSSHIAJSONStub{records: validNHIASSHIARecords()}
	repository = mustNewNHIASSHIARepository(t, stub)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = repository.GetNHIAStateSocialHealthInsuranceAgency(context.Background(), "lashma")
		}()
	}
	wg.Wait()
	if stub.calls != 1 {
		t.Fatalf("decode calls=%d", stub.calls)
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyRepositoryEmbeddedDataset(t *testing.T) {
	jsonRepository, err := NewEmbeddedJSONRepository(datasets.Files(), 64<<20)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := NewNHIAStateSocialHealthInsuranceAgencyRepository(jsonRepository, nhiaSSHIATestPath)
	if err != nil {
		t.Fatal(err)
	}
	result, err := repository.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), interfaces.NHIAStateSocialHealthInsuranceAgencyQuery{PageSize: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Records) != 37 || result.Total != 37 {
		t.Fatalf("embedded result mismatch: %#v", result)
	}
}

func mustNewNHIASSHIARepository(t testing.TB, stub *nhiaSSHIAJSONStub) *NHIAStateSocialHealthInsuranceAgencyFileRepository {
	t.Helper()
	repository, err := NewNHIAStateSocialHealthInsuranceAgencyRepository(stub, nhiaSSHIATestPath)
	if err != nil {
		t.Fatal(err)
	}
	return repository
}

type nhiaSSHIAJSONStub struct {
	records  []models.NHIAStateSocialHealthInsuranceAgency
	raw      string
	err      error
	errFirst error
	calls    int
	mu       sync.Mutex
}

func (s *nhiaSSHIAJSONStub) Decode(ctx context.Context, _ string, dst any) error {
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

func validNHIASSHIARecords() []models.NHIAStateSocialHealthInsuranceAgency {
	var records []models.NHIAStateSocialHealthInsuranceAgency
	if err := json.Unmarshal(readTextBytesForSSHIA(), &records); err != nil {
		panic(err)
	}
	return records
}

func readTextBytesForSSHIA() []byte {
	data, err := fs.ReadFile(datasets.Files(), "healthcare/nhia_state_social_health_insurance_agencies.json")
	if err != nil {
		panic(err)
	}
	return data
}

func TestNHIAStateSocialHealthInsuranceAgencyRepositoryUniqueCoverage(t *testing.T) {
	records := validNHIASSHIARecords()
	ids := map[string]struct{}{}
	states := map[string]struct{}{}
	for _, record := range records {
		ids[record.ID] = struct{}{}
		states[record.StateID] = struct{}{}
	}
	if len(ids) != 37 || len(states) != 37 || !reflect.DeepEqual(states, nhiaSSHIACanonicalStateIDs) {
		t.Fatalf("coverage mismatch ids=%d states=%d", len(ids), len(states))
	}
}
