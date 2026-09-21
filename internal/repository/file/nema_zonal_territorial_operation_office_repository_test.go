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

const nemaOfficeTestPath = "emergency/nema_zonal_territorial_operation_offices.json"

func TestNEMAZonalTerritorialOperationOfficeRepositoryListGetAndFilters(t *testing.T) {
	repository := mustNewNEMAOfficeRepository(t, &nemaOfficeJSONStub{records: validNEMAOfficeRecords()})
	result, err := repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), interfaces.NEMAZonalTerritorialOperationOfficeQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Records) != 17 || result.Page != 1 || result.PageSize != 50 || result.Total != 17 || result.TotalPages != 1 {
		t.Fatalf("default page mismatch: %#v", result)
	}
	if result.Records == nil {
		t.Fatal("nil records")
	}
	for _, tc := range []struct {
		name  string
		query interfaces.NEMAZonalTerritorialOperationOfficeQuery
		total int
		first string
	}{
		{"maximum page size", interfaces.NEMAZonalTerritorialOperationOfficeQuery{PageSize: 100}, 17, "nema-abuja-zonal-territorial-operation-office"},
		{"beyond final", interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: 2, PageSize: 17}, 17, ""},
		{"state", interfaces.NEMAZonalTerritorialOperationOfficeQuery{StateID: " lagos "}, 1, "nema-lagos-zonal-territorial-operation-office"},
		{"office type", interfaces.NEMAZonalTerritorialOperationOfficeQuery{OfficeType: " zonal_territorial_operation_office "}, 17, "nema-abuja-zonal-territorial-operation-office"},
		{"search name", interfaces.NEMAZonalTerritorialOperationOfficeQuery{Search: " port HARCOURT "}, 1, "nema-port-harcourt-zonal-territorial-operation-office"},
		{"combined filters", interfaces.NEMAZonalTerritorialOperationOfficeQuery{StateID: "rivers", OfficeType: "zonal_territorial_operation_office", Search: "harcourt"}, 1, "nema-port-harcourt-zonal-territorial-operation-office"},
		{"combined no result", interfaces.NEMAZonalTerritorialOperationOfficeQuery{StateID: "rivers", Search: "lagos"}, 0, ""},
		{"filter before pagination", interfaces.NEMAZonalTerritorialOperationOfficeQuery{OfficeType: "zonal_territorial_operation_office", Page: 2, PageSize: 1}, 17, "nema-edo-zonal-territorial-operation-office"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), tc.query)
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
	detail, err := repository.GetNEMAZonalTerritorialOperationOffice(context.Background(), "nema-lagos-zonal-territorial-operation-office")
	if err != nil || detail.StateID != "lagos" {
		t.Fatalf("detail=%#v err=%v", detail, err)
	}
	if _, err := repository.GetNEMAZonalTerritorialOperationOffice(context.Background(), "unknown-valid-id"); !errors.Is(err, interfaces.ErrNEMAZonalTerritorialOperationOfficeNotFound) {
		t.Fatalf("unknown err=%v", err)
	}
	if _, err := repository.GetNEMAZonalTerritorialOperationOffice(context.Background(), "bad_id"); !errors.Is(err, interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeQuery) {
		t.Fatalf("malformed err=%v", err)
	}
}

func TestNEMAZonalTerritorialOperationOfficeRepositoryValidationAndStrictDecode(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func([]models.NEMAZonalTerritorialOperationOffice) []models.NEMAZonalTerritorialOperationOffice
	}{
		{"invalid count", func(r []models.NEMAZonalTerritorialOperationOffice) []models.NEMAZonalTerritorialOperationOffice {
			return r[:16]
		}},
		{"duplicate id", func(r []models.NEMAZonalTerritorialOperationOffice) []models.NEMAZonalTerritorialOperationOffice {
			r[1].ID = r[0].ID
			return r
		}},
		{"duplicate state", func(r []models.NEMAZonalTerritorialOperationOffice) []models.NEMAZonalTerritorialOperationOffice {
			r[1].StateID = r[0].StateID
			return r
		}},
		{"invalid country", func(r []models.NEMAZonalTerritorialOperationOffice) []models.NEMAZonalTerritorialOperationOffice {
			r[0].CountryCode = "GH"
			return r
		}},
		{"empty name", func(r []models.NEMAZonalTerritorialOperationOffice) []models.NEMAZonalTerritorialOperationOffice {
			r[0].Name = " "
			return r
		}},
		{"invalid state", func(r []models.NEMAZonalTerritorialOperationOffice) []models.NEMAZonalTerritorialOperationOffice {
			r[0].StateID = "invalid-state"
			return r
		}},
		{"invalid office type", func(r []models.NEMAZonalTerritorialOperationOffice) []models.NEMAZonalTerritorialOperationOffice {
			r[0].OfficeType = "regional_office"
			return r
		}},
		{"bad ordering", func(r []models.NEMAZonalTerritorialOperationOffice) []models.NEMAZonalTerritorialOperationOffice {
			r[0], r[1] = r[1], r[0]
			return r
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repository := mustNewNEMAOfficeRepository(t, &nemaOfficeJSONStub{records: tc.mutate(validNEMAOfficeRecords())})
			if _, err := repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), interfaces.NEMAZonalTerritorialOperationOfficeQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("err=%v", err)
			}
		})
	}
	for _, payload := range []string{`[{"id":"synthetic-id","name":"Synthetic","office_type":"zonal_territorial_operation_office","state_id":"lagos","country_code":"NG","email":"hidden"}]`, `[] {`, `[] {}`} {
		repository := mustNewNEMAOfficeRepository(t, &nemaOfficeJSONStub{raw: payload})
		if _, err := repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), interfaces.NEMAZonalTerritorialOperationOfficeQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) || strings.Contains(err.Error(), "email") {
			t.Fatalf("strict decode err=%v", err)
		}
	}
	repository := mustNewNEMAOfficeRepository(t, &nemaOfficeJSONStub{err: errors.New("/internal/source/path decoder detail")})
	if _, err := repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), interfaces.NEMAZonalTerritorialOperationOfficeQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) || strings.Contains(err.Error(), "internal/source") {
		t.Fatalf("unsanitized err=%v", err)
	}
}

func TestNEMAZonalTerritorialOperationOfficeRepositoryInvalidQueriesContextCacheAndConcurrency(t *testing.T) {
	repository := mustNewNEMAOfficeRepository(t, &nemaOfficeJSONStub{records: validNEMAOfficeRecords()})
	for _, query := range []interfaces.NEMAZonalTerritorialOperationOfficeQuery{
		{Page: -1}, {PageSize: -1}, {PageSize: 101}, {StateID: "not-a-state"}, {OfficeType: "regional"}, {Search: strings.Repeat("a", nemaZonalTerritorialOperationOfficeMaxSearch+1)}, {Search: "\n"},
	} {
		if _, err := repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), query); !errors.Is(err, interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeQuery) &&
			!errors.Is(err, interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeStateFilter) &&
			!errors.Is(err, interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeTypeFilter) &&
			!errors.Is(err, interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeSearch) {
			t.Fatalf("query %#v err=%v", query, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repository = mustNewNEMAOfficeRepository(t, &nemaOfficeJSONStub{records: validNEMAOfficeRecords()})
	if _, err := repository.ListNEMAZonalTerritorialOperationOffices(ctx, interfaces.NEMAZonalTerritorialOperationOfficeQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled err=%v", err)
	}
	deadline, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()
	if _, err := repository.ListNEMAZonalTerritorialOperationOffices(deadline, interfaces.NEMAZonalTerritorialOperationOfficeQuery{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline err=%v", err)
	}
	stub := &nemaOfficeJSONStub{records: validNEMAOfficeRecords(), errFirst: errors.New("temporary")}
	repository = mustNewNEMAOfficeRepository(t, stub)
	if _, err := repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), interfaces.NEMAZonalTerritorialOperationOfficeQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("first err=%v", err)
	}
	if _, err := repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), interfaces.NEMAZonalTerritorialOperationOfficeQuery{}); err != nil {
		t.Fatalf("retry err=%v", err)
	}
	if stub.calls != 2 {
		t.Fatalf("decode calls=%d", stub.calls)
	}
	result, _ := repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), interfaces.NEMAZonalTerritorialOperationOfficeQuery{})
	result.Records[0].Name = "mutated"
	again, _ := repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), interfaces.NEMAZonalTerritorialOperationOfficeQuery{})
	if again.Records[0].Name == "mutated" {
		t.Fatal("caller mutation affected cache")
	}
	detail, _ := repository.GetNEMAZonalTerritorialOperationOffice(context.Background(), result.Records[0].ID)
	detail.Name = "mutated detail"
	detailAgain, _ := repository.GetNEMAZonalTerritorialOperationOffice(context.Background(), result.Records[0].ID)
	if detailAgain.Name == "mutated detail" {
		t.Fatal("caller detail mutation affected cache")
	}
	stub = &nemaOfficeJSONStub{records: validNEMAOfficeRecords()}
	repository = mustNewNEMAOfficeRepository(t, stub)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = repository.GetNEMAZonalTerritorialOperationOffice(context.Background(), "nema-lagos-zonal-territorial-operation-office")
		}()
	}
	wg.Wait()
	if stub.calls != 1 {
		t.Fatalf("decode calls=%d", stub.calls)
	}
}

func TestNEMAZonalTerritorialOperationOfficeRepositoryEmbeddedDataset(t *testing.T) {
	jsonRepository, err := NewEmbeddedJSONRepository(datasets.Files(), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := NewNEMAZonalTerritorialOperationOfficeRepository(jsonRepository, nemaOfficeTestPath)
	if err != nil {
		t.Fatal(err)
	}
	result, err := repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), interfaces.NEMAZonalTerritorialOperationOfficeQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Records) != 17 || result.Total != 17 {
		t.Fatalf("embedded result mismatch: %#v", result)
	}
	if _, err := fs.ReadFile(datasets.Files(), "metadata/emergency/nema_zonal_territorial_operation_offices_reconciliation/index.json"); err == nil {
		t.Fatal("reconciliation metadata must not be embedded for runtime")
	}
}

func BenchmarkNEMAZonalTerritorialOperationOfficeRepositoryFirstLoad(b *testing.B) {
	for i := 0; i < b.N; i++ {
		repository := mustNewNEMAOfficeRepository(b, &nemaOfficeJSONStub{records: validNEMAOfficeRecords()})
		if _, err := repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), interfaces.NEMAZonalTerritorialOperationOfficeQuery{}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNEMAZonalTerritorialOperationOfficeRepositoryCachedDefaultPage(b *testing.B) {
	benchmarkNEMAOfficeRepository(b, interfaces.NEMAZonalTerritorialOperationOfficeQuery{})
}

func BenchmarkNEMAZonalTerritorialOperationOfficeRepositoryCachedDetail(b *testing.B) {
	repository := mustNewNEMAOfficeRepository(b, &nemaOfficeJSONStub{records: validNEMAOfficeRecords()})
	_, _ = repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), interfaces.NEMAZonalTerritorialOperationOfficeQuery{})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := repository.GetNEMAZonalTerritorialOperationOffice(context.Background(), "nema-lagos-zonal-territorial-operation-office"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNEMAZonalTerritorialOperationOfficeRepositoryStateFilter(b *testing.B) {
	benchmarkNEMAOfficeRepository(b, interfaces.NEMAZonalTerritorialOperationOfficeQuery{StateID: "lagos"})
}

func BenchmarkNEMAZonalTerritorialOperationOfficeRepositoryOfficeTypeFilter(b *testing.B) {
	benchmarkNEMAOfficeRepository(b, interfaces.NEMAZonalTerritorialOperationOfficeQuery{OfficeType: "zonal_territorial_operation_office"})
}

func BenchmarkNEMAZonalTerritorialOperationOfficeRepositoryNameSearch(b *testing.B) {
	benchmarkNEMAOfficeRepository(b, interfaces.NEMAZonalTerritorialOperationOfficeQuery{Search: "lagos"})
}

func BenchmarkNEMAZonalTerritorialOperationOfficeRepositoryCombinedFilters(b *testing.B) {
	benchmarkNEMAOfficeRepository(b, interfaces.NEMAZonalTerritorialOperationOfficeQuery{StateID: "lagos", OfficeType: "zonal_territorial_operation_office", Search: "lagos"})
}

func benchmarkNEMAOfficeRepository(b *testing.B, query interfaces.NEMAZonalTerritorialOperationOfficeQuery) {
	repository := mustNewNEMAOfficeRepository(b, &nemaOfficeJSONStub{records: validNEMAOfficeRecords()})
	_, _ = repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), interfaces.NEMAZonalTerritorialOperationOfficeQuery{})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := repository.ListNEMAZonalTerritorialOperationOffices(context.Background(), query); err != nil {
			b.Fatal(err)
		}
	}
}

func mustNewNEMAOfficeRepository(t testing.TB, stub *nemaOfficeJSONStub) *NEMAZonalTerritorialOperationOfficeFileRepository {
	t.Helper()
	repository, err := NewNEMAZonalTerritorialOperationOfficeRepository(stub, nemaOfficeTestPath)
	if err != nil {
		t.Fatal(err)
	}
	return repository
}

type nemaOfficeJSONStub struct {
	records  []models.NEMAZonalTerritorialOperationOffice
	raw      string
	err      error
	errFirst error
	calls    int
	mu       sync.Mutex
}

func (s *nemaOfficeJSONStub) Decode(ctx context.Context, _ string, dst any) error {
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

func validNEMAOfficeRecords() []models.NEMAZonalTerritorialOperationOffice {
	var records []models.NEMAZonalTerritorialOperationOffice
	if err := json.Unmarshal(readNEMAOfficeDataset(), &records); err != nil {
		panic(err)
	}
	return records
}

func readNEMAOfficeDataset() []byte {
	data, err := fs.ReadFile(datasets.Files(), nemaOfficeTestPath)
	if err != nil {
		panic(err)
	}
	return data
}
