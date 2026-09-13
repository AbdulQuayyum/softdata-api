package file

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const nhiaAccreditedHMOTestPath = "healthcare/nhia_accredited_health_maintenance_organisations.json"

func TestNHIAAccreditedHMORepositoryRealDataset(t *testing.T) {
	repository := newRealNHIAAccreditedHMORepository(t)
	result, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{})
	if err != nil {
		t.Fatalf("ListNHIAAccreditedHealthMaintenanceOrganisations() error = %v", err)
	}
	if result.Page != 1 || result.PageSize != 50 || result.Total != 94 || result.TotalPages != 2 || len(result.Records) != 50 || result.Records == nil {
		t.Fatalf("unexpected default result: %#v", result)
	}

	second, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Page: 2, PageSize: 50})
	if err != nil || second.Page != 2 || second.PageSize != 50 || second.Total != 94 || second.TotalPages != 2 || len(second.Records) != 44 {
		t.Fatalf("unexpected second page: %#v err=%v", second, err)
	}
	custom, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Page: 2, PageSize: 10})
	if err != nil || custom.Page != 2 || custom.PageSize != 10 || custom.Total != 94 || custom.TotalPages != 10 || len(custom.Records) != 10 {
		t.Fatalf("unexpected custom result: %#v err=%v", custom, err)
	}
	beyond, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Page: 11, PageSize: 10})
	if err != nil || beyond.Records == nil || len(beyond.Records) != 0 || beyond.Total != 94 || beyond.TotalPages != 10 || beyond.Page != 11 || beyond.PageSize != 10 {
		t.Fatalf("unexpected beyond result: %#v err=%v", beyond, err)
	}

	status, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{AccreditationStatus: " accredited ", PageSize: 100})
	if err != nil || status.Total != 94 || len(status.Records) != 94 {
		t.Fatalf("status filter failed: %#v err=%v", status, err)
	}
	hmoID, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{HMOID: " 102 "})
	if err != nil || hmoID.Total != 1 || len(hmoID.Records) != 1 || hmoID.Records[0].HMOID != "102" {
		t.Fatalf("hmo_id filter failed: %#v err=%v", hmoID, err)
	}
	search, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Search: "  healthcare  ", PageSize: 100})
	if err != nil || search.Total == 0 {
		t.Fatalf("search failed: %#v err=%v", search, err)
	}
	combined, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{AccreditationStatus: "accredited", HMOID: "102", Search: "a&m"})
	if err != nil || combined.Total != 1 || len(combined.Records) != 1 || combined.Records[0].ID != "a-and-m-healthcare-trust-limited-102" {
		t.Fatalf("combined filter failed: %#v err=%v", combined, err)
	}
	filteredPage, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Search: "health", Page: 2, PageSize: 1})
	if err != nil || filteredPage.Total < 2 || len(filteredPage.Records) != 1 {
		t.Fatalf("filter before pagination failed: %#v err=%v", filteredPage, err)
	}
	empty, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{HMOID: "999"})
	if err != nil || empty.Records == nil || len(empty.Records) != 0 || empty.Total != 0 || empty.TotalPages != 0 {
		t.Fatalf("empty result failed: %#v err=%v", empty, err)
	}

	records := loadNHIAAccreditedHMOFixture(t)
	longest := records[0]
	for _, record := range records[1:] {
		if len(record.ID) > len(longest.ID) {
			longest = record
		}
	}
	for _, want := range []models.NHIAAccreditedHealthMaintenanceOrganisation{records[0], longest} {
		got, err := repository.GetNHIAAccreditedHealthMaintenanceOrganisation(context.Background(), want.ID)
		if err != nil || got.ID != want.ID {
			t.Fatalf("GetNHIAAccreditedHealthMaintenanceOrganisation(%q) = %#v, %v", want.ID, got, err)
		}
	}
	if _, err := repository.GetNHIAAccreditedHealthMaintenanceOrganisation(context.Background(), "valid-but-unknown"); !errors.Is(err, interfaces.ErrNHIAAccreditedHealthMaintenanceOrganisationNotFound) {
		t.Fatalf("unknown lookup error = %v", err)
	}

	result.Records[0].Name = "mutated"
	again, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{})
	if err != nil || again.Records[0].Name == "mutated" {
		t.Fatalf("result mutation affected cache: %#v err=%v", again, err)
	}
}

func TestNHIAAccreditedHMORepositoryDatasetMetadataAndSchemaCounts(t *testing.T) {
	records := loadNHIAAccreditedHMOFixture(t)
	var metadata struct {
		RecordCount int `json:"record_count"`
		SourceRows  int `json:"source_rows"`
	}
	decodeTestJSON(t, "../../../datasets/metadata/healthcare/nhia_accredited_health_maintenance_organisations.json", &metadata)
	var schema struct {
		MinItems int `json:"minItems"`
		MaxItems int `json:"maxItems"`
	}
	decodeTestJSON(t, "../../../datasets/schemas/healthcare/nhia_accredited_health_maintenance_organisations.schema.json", &schema)
	if len(records) != 94 || metadata.RecordCount != 94 || metadata.SourceRows != 94 || schema.MinItems != 94 || schema.MaxItems != 94 {
		t.Fatalf("count metadata mismatch: records=%d metadata=%#v schema=%#v", len(records), metadata, schema)
	}
}

func TestNHIAAccreditedHMORepositoryValidationAndContext(t *testing.T) {
	repository := newRealNHIAAccreditedHMORepository(t)
	for _, test := range []struct {
		name  string
		query interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery
		want  error
	}{
		{"bad page", interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Page: -1}, interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationQuery},
		{"bad size", interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{PageSize: 101}, interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationQuery},
		{"blank status", interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{AccreditationStatus: "  "}, interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationQuery},
		{"bad status", interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{AccreditationStatus: "expired"}, interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationStatusFilter},
		{"blank hmo id", interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{HMOID: "  "}, interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationHMOIDFilter},
		{"bad hmo id", interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{HMOID: "12\n0"}, interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationHMOIDFilter},
		{"blank search", interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Search: "  "}, interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationSearch},
		{"bad search", interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{Search: strings.Repeat("x", 101)}, interfaces.ErrInvalidNHIAAccreditedHealthMaintenanceOrganisationSearch},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), test.query); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(ctx, interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled list error = %v", err)
	}
	if _, err := repository.GetNHIAAccreditedHealthMaintenanceOrganisation(ctx, loadNHIAAccreditedHMOFixture(t)[0].ID); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled get error = %v", err)
	}
	deadlineCtx, deadlineCancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer deadlineCancel()
	if _, err := newRealNHIAAccreditedHMORepository(t).ListNHIAAccreditedHealthMaintenanceOrganisations(deadlineCtx, interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline list error = %v", err)
	}
}

func TestNHIAAccreditedHMORepositoryConcurrentFirstLoad(t *testing.T) {
	stub := &nhiaAccreditedHMOJSONStub{records: loadNHIAAccreditedHMOFixture(t)}
	repository := mustNewNHIAAccreditedHMORepository(t, stub)
	const workers = 8
	errs := make(chan error, workers)
	var wait sync.WaitGroup
	for i := 0; i < workers; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{PageSize: 100})
			if err == nil && len(result.Records) != 94 {
				err = fmt.Errorf("got %d records", len(result.Records))
			}
			errs <- err
		}()
	}
	wait.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if stub.calls[nhiaAccreditedHMOTestPath] != 1 {
		t.Fatalf("decode calls = %d, want 1", stub.calls[nhiaAccreditedHMOTestPath])
	}
}

func TestNHIAAccreditedHMORepositoryFailedLoadCanRetry(t *testing.T) {
	stub := &nhiaAccreditedHMOJSONStub{records: loadNHIAAccreditedHMOFixture(t), failRecordsOnce: true}
	repository := mustNewNHIAAccreditedHMORepository(t, stub)
	if _, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("first load error = %v", err)
	}
	if _, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}); err != nil {
		t.Fatalf("retry error = %v", err)
	}
	if stub.calls[nhiaAccreditedHMOTestPath] != 2 {
		t.Fatalf("decode calls = %d, want 2", stub.calls[nhiaAccreditedHMOTestPath])
	}
}

func TestNHIAAccreditedHMORepositorySanitizesUnexpectedErrors(t *testing.T) {
	secret := errors.New("/machine/private/nhia.json: decoder secret")
	stub := &nhiaAccreditedHMOJSONStub{recordsErr: secret}
	repository := mustNewNHIAAccreditedHMORepository(t, stub)
	_, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{})
	if err == nil || errors.Is(err, secret) || stringsContainsAny(err.Error(), []string{"/machine", "decoder secret"}) {
		t.Fatalf("unsanitized error: %v", err)
	}
}

func TestNHIAAccreditedHMORepositoryRejectsMalformedTrailingAndUnknownJSON(t *testing.T) {
	for _, test := range []struct{ name, body string }{
		{"malformed", "["},
		{"trailing", "[] []"},
		{"unknown field", `[{"id":"example","name":"Example","country_code":"NG","organisation_type":"health_maintenance_organisation","accreditation_status":"accredited","hmo_id":"7","website_url":"https://example.test"}]`},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeNHIAAccreditedHMOTestFile(t, root, nhiaAccreditedHMOTestPath, []byte(test.body))
			jsonRepository, err := NewJSONRepository(root, 1<<20)
			if err != nil {
				t.Fatal(err)
			}
			repository, err := NewNHIAAccreditedHealthMaintenanceOrganisationRepository(jsonRepository, nhiaAccreditedHMOTestPath)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("error = %v, want invalid dataset", err)
			}
		})
	}
}

func TestNHIAAccreditedHMORepositoryRejectsInvalidDatasetValues(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func([]models.NHIAAccreditedHealthMaintenanceOrganisation) []models.NHIAAccreditedHealthMaintenanceOrganisation
	}{
		{"duplicate id", func(records []models.NHIAAccreditedHealthMaintenanceOrganisation) []models.NHIAAccreditedHealthMaintenanceOrganisation {
			records[1].ID = records[0].ID
			return records
		}},
		{"duplicate hmo id", func(records []models.NHIAAccreditedHealthMaintenanceOrganisation) []models.NHIAAccreditedHealthMaintenanceOrganisation {
			records[1].HMOID = records[0].HMOID
			return records
		}},
		{"invalid country", func(records []models.NHIAAccreditedHealthMaintenanceOrganisation) []models.NHIAAccreditedHealthMaintenanceOrganisation {
			records[0].CountryCode = "US"
			return records
		}},
		{"invalid organisation type", func(records []models.NHIAAccreditedHealthMaintenanceOrganisation) []models.NHIAAccreditedHealthMaintenanceOrganisation {
			records[0].OrganisationType = "third_party_administrator"
			return records
		}},
		{"invalid status", func(records []models.NHIAAccreditedHealthMaintenanceOrganisation) []models.NHIAAccreditedHealthMaintenanceOrganisation {
			records[0].AccreditationStatus = "expired"
			return records
		}},
		{"missing required", func(records []models.NHIAAccreditedHealthMaintenanceOrganisation) []models.NHIAAccreditedHealthMaintenanceOrganisation {
			records[0].Name = ""
			return records
		}},
		{"empty optional-like string", func(records []models.NHIAAccreditedHealthMaintenanceOrganisation) []models.NHIAAccreditedHealthMaintenanceOrganisation {
			records[0].HMOID = " "
			return records
		}},
		{"invalid ordering", func(records []models.NHIAAccreditedHealthMaintenanceOrganisation) []models.NHIAAccreditedHealthMaintenanceOrganisation {
			records[0], records[1] = records[1], records[0]
			return records
		}},
		{"wrong count", func(records []models.NHIAAccreditedHealthMaintenanceOrganisation) []models.NHIAAccreditedHealthMaintenanceOrganisation {
			return records[:93]
		}},
		{"forbidden marker", func(records []models.NHIAAccreditedHealthMaintenanceOrganisation) []models.NHIAAccreditedHealthMaintenanceOrganisation {
			records[0].Name = "Director Example HMO"
			return records
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			stub := &nhiaAccreditedHMOJSONStub{records: test.mutate(loadNHIAAccreditedHMOFixture(t))}
			repository := mustNewNHIAAccreditedHMORepository(t, stub)
			if _, err := repository.ListNHIAAccreditedHealthMaintenanceOrganisations(context.Background(), interfaces.NHIAAccreditedHealthMaintenanceOrganisationQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("error = %v, want invalid dataset", err)
			}
		})
	}
}

func newRealNHIAAccreditedHMORepository(t testing.TB) *NHIAAccreditedHealthMaintenanceOrganisationFileRepository {
	t.Helper()
	jsonRepository, err := NewJSONRepository("../../../datasets", 64<<20)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := NewNHIAAccreditedHealthMaintenanceOrganisationRepository(jsonRepository, nhiaAccreditedHMOTestPath)
	if err != nil {
		t.Fatal(err)
	}
	return repository
}

func mustNewNHIAAccreditedHMORepository(t testing.TB, stub *nhiaAccreditedHMOJSONStub) *NHIAAccreditedHealthMaintenanceOrganisationFileRepository {
	t.Helper()
	repository, err := NewNHIAAccreditedHealthMaintenanceOrganisationRepository(stub, nhiaAccreditedHMOTestPath)
	if err != nil {
		t.Fatal(err)
	}
	return repository
}

func loadNHIAAccreditedHMOFixture(t testing.TB) []models.NHIAAccreditedHealthMaintenanceOrganisation {
	t.Helper()
	var items []models.NHIAAccreditedHealthMaintenanceOrganisation
	decodeTestJSON(t, "../../../datasets/healthcare/nhia_accredited_health_maintenance_organisations.json", &items)
	return items
}

type nhiaAccreditedHMOJSONStub struct {
	records         []models.NHIAAccreditedHealthMaintenanceOrganisation
	recordsErr      error
	failRecordsOnce bool
	calls           map[string]int
}

func (s *nhiaAccreditedHMOJSONStub) Decode(_ context.Context, path string, destination any) error {
	if s.calls == nil {
		s.calls = map[string]int{}
	}
	s.calls[path]++
	if path != nhiaAccreditedHMOTestPath {
		return fmt.Errorf("unexpected path %s", path)
	}
	if s.failRecordsOnce {
		s.failRecordsOnce = false
		return interfaces.ErrInvalidDatasetFile
	}
	if s.recordsErr != nil {
		return s.recordsErr
	}
	*(destination.(*[]models.NHIAAccreditedHealthMaintenanceOrganisation)) = append([]models.NHIAAccreditedHealthMaintenanceOrganisation(nil), s.records...)
	return nil
}

func writeNHIAAccreditedHMOTestFile(t testing.TB, root, relativePath string, data []byte) {
	t.Helper()
	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}
