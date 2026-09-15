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

const nhiaHCPTestPath = "healthcare/nhia_active_accredited_healthcare_providers.json"

func TestNHIAActiveAccreditedHealthcareProviderRepositoryListAndGet(t *testing.T) {
	repository := mustNewNHIAHCPRepository(t, &nhiaHCPJSONStub{records: validNHIAHCPRecords()})
	result, err := repository.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), interfaces.NHIAActiveAccreditedHealthcareProviderQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Records) != 50 || result.Page != 1 || result.PageSize != 50 || result.Total != 6536 || result.TotalPages != 131 {
		t.Fatalf("default page mismatch: %#v", result)
	}
	if result.Records == nil {
		t.Fatal("nil records")
	}

	for _, tc := range []struct {
		name  string
		query interfaces.NHIAActiveAccreditedHealthcareProviderQuery
		total int
		first string
	}{
		{"custom page", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 2, PageSize: 25}, 6536, "ab-0032-p"},
		{"maximum page size", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{PageSize: 100}, 6536, "ab-0001-p"},
		{"beyond final", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 132, PageSize: 50}, 6536, ""},
		{"provider code", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{ProviderCode: "AB/0001/P"}, 1, "ab-0001-p"},
		{"provider code not found", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{ProviderCode: "AB/9999/P"}, 0, ""},
		{"facility primary", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{FacilityType: "primary", PageSize: 1}, 4001, "ab-0002-p"},
		{"facility primary and secondary", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{FacilityType: "primary_and_secondary", PageSize: 1}, 2535, "ab-0001-p"},
		{"listing status", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{ListingStatus: "active_accredited", PageSize: 1}, 6536, "ab-0001-p"},
		{"search", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Search: "wildot"}, 1, "fct-0001-p"},
		{"trimmed search", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Search: "  WILDOT  "}, 1, "fct-0001-p"},
		{"combined indexed filters", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{ProviderCode: "FCT/0001/P", FacilityType: "primary", ListingStatus: "active_accredited", Search: "clinic"}, 1, "fct-0001-p"},
		{"combined no result", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{ProviderCode: "FCT/0001/P", FacilityType: "primary_and_secondary"}, 0, ""},
		{"filter before pagination", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{ProviderCode: "FCT/0001/P", Page: 2, PageSize: 1}, 1, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repository.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), tc.query)
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

	detail, err := repository.GetNHIAActiveAccreditedHealthcareProvider(context.Background(), "fct-0001-p")
	if err != nil || detail.ProviderCode != "FCT/0001/P" {
		t.Fatalf("detail=%#v err=%v", detail, err)
	}
	if _, err := repository.GetNHIAActiveAccreditedHealthcareProvider(context.Background(), "unknown-valid-id"); !errors.Is(err, interfaces.ErrNHIAActiveAccreditedHealthcareProviderNotFound) {
		t.Fatalf("unknown err=%v", err)
	}
	if _, err := repository.GetNHIAActiveAccreditedHealthcareProvider(context.Background(), "bad_id"); !errors.Is(err, interfaces.ErrInvalidNHIAActiveAccreditedHealthcareProviderQuery) {
		t.Fatalf("malformed err=%v", err)
	}
}

func TestNHIAActiveAccreditedHealthcareProviderRepositoryValidationAndSanitization(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func([]models.NHIAActiveAccreditedHealthcareProvider) []models.NHIAActiveAccreditedHealthcareProvider
	}{
		{"invalid count", func(r []models.NHIAActiveAccreditedHealthcareProvider) []models.NHIAActiveAccreditedHealthcareProvider {
			return r[:6535]
		}},
		{"duplicate id", func(r []models.NHIAActiveAccreditedHealthcareProvider) []models.NHIAActiveAccreditedHealthcareProvider {
			r[1].ID = r[0].ID
			return r
		}},
		{"duplicate code", func(r []models.NHIAActiveAccreditedHealthcareProvider) []models.NHIAActiveAccreditedHealthcareProvider {
			r[1].ProviderCode = r[0].ProviderCode
			return r
		}},
		{"invalid country", func(r []models.NHIAActiveAccreditedHealthcareProvider) []models.NHIAActiveAccreditedHealthcareProvider {
			r[0].CountryCode = "GH"
			return r
		}},
		{"invalid facility", func(r []models.NHIAActiveAccreditedHealthcareProvider) []models.NHIAActiveAccreditedHealthcareProvider {
			r[0].FacilityType = "secondary"
			return r
		}},
		{"invalid status", func(r []models.NHIAActiveAccreditedHealthcareProvider) []models.NHIAActiveAccreditedHealthcareProvider {
			r[0].ListingStatus = "operational"
			return r
		}},
		{"empty name", func(r []models.NHIAActiveAccreditedHealthcareProvider) []models.NHIAActiveAccreditedHealthcareProvider {
			r[0].Name = " "
			return r
		}},
		{"empty code", func(r []models.NHIAActiveAccreditedHealthcareProvider) []models.NHIAActiveAccreditedHealthcareProvider {
			r[0].ProviderCode = ""
			return r
		}},
		{"bad ordering", func(r []models.NHIAActiveAccreditedHealthcareProvider) []models.NHIAActiveAccreditedHealthcareProvider {
			r[0], r[1] = r[1], r[0]
			return r
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repository := mustNewNHIAHCPRepository(t, &nhiaHCPJSONStub{records: tc.mutate(validNHIAHCPRecords())})
			if _, err := repository.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("err=%v", err)
			}
		})
	}
	for _, payload := range []string{`[{"id":"ab-0001-p","name":"Name","country_code":"NG","provider_code":"AB/0001/P","facility_type":"primary","listing_status":"active_accredited","address":"hidden"}]`, `[] {}`} {
		repository := mustNewNHIAHCPRepository(t, &nhiaHCPJSONStub{raw: payload})
		if _, err := repository.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) || strings.Contains(err.Error(), "address") {
			t.Fatalf("strict decode err=%v", err)
		}
	}
	repository := mustNewNHIAHCPRepository(t, &nhiaHCPJSONStub{err: errors.New("/secret/path decoder address")})
	if _, err := repository.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "address") {
		t.Fatalf("unsanitized err=%v", err)
	}
}

func TestNHIAActiveAccreditedHealthcareProviderRepositoryCacheOwnershipContextAndConcurrency(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repository := mustNewNHIAHCPRepository(t, &nhiaHCPJSONStub{records: validNHIAHCPRecords()})
	if _, err := repository.ListNHIAActiveAccreditedHealthcareProviders(ctx, interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled err=%v", err)
	}
	deadline, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()
	if _, err := repository.ListNHIAActiveAccreditedHealthcareProviders(deadline, interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline err=%v", err)
	}
	stub := &nhiaHCPJSONStub{records: validNHIAHCPRecords(), errFirst: errors.New("temporary")}
	repository = mustNewNHIAHCPRepository(t, stub)
	if _, err := repository.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("first err=%v", err)
	}
	if _, err := repository.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), interfaces.NHIAActiveAccreditedHealthcareProviderQuery{}); err != nil {
		t.Fatalf("retry err=%v", err)
	}
	if stub.calls != 2 {
		t.Fatalf("decode calls=%d", stub.calls)
	}
	result, _ := repository.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), interfaces.NHIAActiveAccreditedHealthcareProviderQuery{})
	result.Records[0].Name = "mutated"
	again, _ := repository.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), interfaces.NHIAActiveAccreditedHealthcareProviderQuery{})
	if again.Records[0].Name == "mutated" {
		t.Fatal("caller mutation affected cache")
	}
	stub = &nhiaHCPJSONStub{records: validNHIAHCPRecords()}
	repository = mustNewNHIAHCPRepository(t, stub)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = repository.GetNHIAActiveAccreditedHealthcareProvider(context.Background(), "fct-0001-p")
		}()
	}
	wg.Wait()
	if stub.calls != 1 {
		t.Fatalf("decode calls=%d", stub.calls)
	}
}

func TestNHIAActiveAccreditedHealthcareProviderRepositoryEmbeddedDataset(t *testing.T) {
	jsonRepository, err := NewEmbeddedJSONRepository(datasets.Files(), 64<<20)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := NewNHIAActiveAccreditedHealthcareProviderRepository(jsonRepository, nhiaHCPTestPath)
	if err != nil {
		t.Fatal(err)
	}
	result, err := repository.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), interfaces.NHIAActiveAccreditedHealthcareProviderQuery{PageSize: 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Records) != 100 || result.Total != 6536 {
		t.Fatalf("embedded result mismatch: %#v", result)
	}
}

func mustNewNHIAHCPRepository(t testing.TB, stub *nhiaHCPJSONStub) *NHIAActiveAccreditedHealthcareProviderFileRepository {
	t.Helper()
	repository, err := NewNHIAActiveAccreditedHealthcareProviderRepository(stub, nhiaHCPTestPath)
	if err != nil {
		t.Fatal(err)
	}
	return repository
}

type nhiaHCPJSONStub struct {
	records  []models.NHIAActiveAccreditedHealthcareProvider
	raw      string
	err      error
	errFirst error
	calls    int
	mu       sync.Mutex
}

func (s *nhiaHCPJSONStub) Decode(ctx context.Context, _ string, dst any) error {
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

func validNHIAHCPRecords() []models.NHIAActiveAccreditedHealthcareProvider {
	var records []models.NHIAActiveAccreditedHealthcareProvider
	if err := json.Unmarshal(readNHIAHCPDataset(), &records); err != nil {
		panic(err)
	}
	return records
}

func readNHIAHCPDataset() []byte {
	data, err := fs.ReadFile(datasets.Files(), nhiaHCPTestPath)
	if err != nil {
		panic(err)
	}
	return data
}
