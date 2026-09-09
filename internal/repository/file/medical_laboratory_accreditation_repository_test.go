package file

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

const (
	medicalLabAccreditationTestPath = "healthcare/medical_laboratory_accreditations.json"
	medicalLabStatesTestPath        = "geography/states.json"
)

func TestMedicalLaboratoryAccreditationRepositoryRealDataset(t *testing.T) {
	repository := newRealMedicalLaboratoryAccreditationRepository(t)
	result, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{})
	if err != nil {
		t.Fatalf("ListMedicalLaboratoryAccreditations() error = %v", err)
	}
	if result.Page != 1 || result.PageSize != 50 || result.Total != 30 || result.TotalPages != 1 || len(result.Records) != 30 || result.Records == nil {
		t.Fatalf("unexpected default result: %#v", result)
	}

	custom, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{Page: 2, PageSize: 10})
	if err != nil || custom.Page != 2 || custom.PageSize != 10 || custom.Total != 30 || custom.TotalPages != 3 || len(custom.Records) != 10 {
		t.Fatalf("unexpected custom result: %#v err=%v", custom, err)
	}
	beyond, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{Page: 4, PageSize: 10})
	if err != nil || beyond.Records == nil || len(beyond.Records) != 0 || beyond.Total != 30 || beyond.TotalPages != 3 || beyond.Page != 4 || beyond.PageSize != 10 {
		t.Fatalf("unexpected beyond result: %#v err=%v", beyond, err)
	}

	lagos, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{StateID: "lagos", PageSize: 100})
	if err != nil || lagos.Total != 10 || len(lagos.Records) != 10 {
		t.Fatalf("lagos filter failed: %#v err=%v", lagos, err)
	}
	accredited, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{AccreditationStatus: "accredited", PageSize: 100})
	if err != nil || accredited.Total != 26 || len(accredited.Records) != 26 {
		t.Fatalf("accredited filter failed: %#v err=%v", accredited, err)
	}
	expired, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{AccreditationStatus: "expired", PageSize: 100})
	if err != nil || expired.Total != 4 || len(expired.Records) != 4 {
		t.Fatalf("expired filter failed: %#v err=%v", expired, err)
	}
	search, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{Search: "  eVeRiGhT  ", PageSize: 100})
	if err != nil || search.Total != 2 || len(search.Records) != 2 {
		t.Fatalf("search failed: %#v err=%v", search, err)
	}
	combined, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{StateID: "fct", AccreditationStatus: "expired", Search: "life", PageSize: 100})
	if err != nil || combined.Total != 1 || len(combined.Records) != 1 || combined.Records[0].AccreditationNumber != "ML0010" {
		t.Fatalf("combined filter failed: %#v err=%v", combined, err)
	}
	empty, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{StateID: "oyo", AccreditationStatus: "accredited", Search: "nope"})
	if err != nil || empty.Records == nil || len(empty.Records) != 0 || empty.Total != 0 || empty.TotalPages != 0 {
		t.Fatalf("empty result failed: %#v err=%v", empty, err)
	}

	records := loadMedicalLaboratoryAccreditationFixture(t)
	longest := records[0]
	for _, record := range records[1:] {
		if len(record.ID) > len(longest.ID) {
			longest = record
		}
	}
	for _, want := range []models.MedicalLaboratoryAccreditation{records[0], longest} {
		got, err := repository.GetMedicalLaboratoryAccreditation(context.Background(), want.ID)
		if err != nil || got.ID != want.ID {
			t.Fatalf("GetMedicalLaboratoryAccreditation(%q) = %#v, %v", want.ID, got, err)
		}
	}
	if _, err := repository.GetMedicalLaboratoryAccreditation(context.Background(), "valid-but-unknown"); !errors.Is(err, interfaces.ErrMedicalLaboratoryAccreditationNotFound) {
		t.Fatalf("unknown lookup error = %v", err)
	}

	result.Records[0].Name = "mutated"
	again, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{})
	if err != nil || again.Records[0].Name == "mutated" {
		t.Fatalf("result mutation affected cache: %#v err=%v", again, err)
	}
}

func TestMedicalLaboratoryAccreditationRepositoryValidationAndContext(t *testing.T) {
	repository := newRealMedicalLaboratoryAccreditationRepository(t)
	for _, test := range []struct {
		name  string
		query interfaces.MedicalLaboratoryAccreditationQuery
		want  error
	}{
		{"bad page", interfaces.MedicalLaboratoryAccreditationQuery{Page: -1}, interfaces.ErrInvalidMedicalLaboratoryAccreditationQuery},
		{"bad size", interfaces.MedicalLaboratoryAccreditationQuery{PageSize: 101}, interfaces.ErrInvalidMedicalLaboratoryAccreditationQuery},
		{"bad state", interfaces.MedicalLaboratoryAccreditationQuery{StateID: "not-a-state"}, interfaces.ErrInvalidMedicalLaboratoryAccreditationStateFilter},
		{"bad status", interfaces.MedicalLaboratoryAccreditationQuery{AccreditationStatus: "licensed"}, interfaces.ErrInvalidMedicalLaboratoryAccreditationStatusFilter},
		{"bad search", interfaces.MedicalLaboratoryAccreditationQuery{Search: strings.Repeat("x", 101)}, interfaces.ErrInvalidMedicalLaboratoryAccreditationSearch},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), test.query); !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := repository.ListMedicalLaboratoryAccreditations(ctx, interfaces.MedicalLaboratoryAccreditationQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled list error = %v", err)
	}
	if _, err := repository.GetMedicalLaboratoryAccreditation(ctx, loadMedicalLaboratoryAccreditationFixture(t)[0].ID); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled get error = %v", err)
	}
}

func TestMedicalLaboratoryAccreditationRepositoryConcurrentFirstLoad(t *testing.T) {
	repository := newRealMedicalLaboratoryAccreditationRepository(t)
	const workers = 8
	errs := make(chan error, workers)
	var wait sync.WaitGroup
	for i := 0; i < workers; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{PageSize: 100})
			if err == nil && len(result.Records) != 30 {
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
}

func TestMedicalLaboratoryAccreditationRepositoryFailedLoadCanRetry(t *testing.T) {
	stub := &medicalLabAccreditationJSONStub{records: loadMedicalLaboratoryAccreditationFixture(t), states: loadMedicalLaboratoryAccreditationStatesFixture(t), failRecordsOnce: true}
	repository := mustNewMedicalLaboratoryAccreditationRepository(t, stub)
	if _, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
		t.Fatalf("first load error = %v", err)
	}
	if _, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{}); err != nil {
		t.Fatalf("retry error = %v", err)
	}
	if stub.calls[medicalLabAccreditationTestPath] != 2 {
		t.Fatalf("decode calls = %d, want 2", stub.calls[medicalLabAccreditationTestPath])
	}
}

func TestMedicalLaboratoryAccreditationRepositorySanitizesUnexpectedErrors(t *testing.T) {
	secret := errors.New("/machine/private/medical_laboratory_accreditations.json: decoder secret")
	stub := &medicalLabAccreditationJSONStub{recordsErr: secret}
	repository := mustNewMedicalLaboratoryAccreditationRepository(t, stub)
	_, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{})
	if err == nil || errors.Is(err, secret) || stringsContainsAny(err.Error(), []string{"/machine", "decoder secret"}) {
		t.Fatalf("unsanitized error: %v", err)
	}
}

func TestMedicalLaboratoryAccreditationRepositoryRejectsMalformedTrailingAndUnknownJSON(t *testing.T) {
	for _, test := range []struct{ name, body string }{
		{"malformed", "["},
		{"trailing", "[] []"},
		{"unknown field", `[{"id":"example","name":"Example","state_id":"lagos","country_code":"NG","accreditation_status":"accredited","registration_status":"active"}]`},
	} {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeMedicalLabAccreditationTestFile(t, root, medicalLabAccreditationTestPath, []byte(test.body))
			writeJSONTestFile(t, root, medicalLabStatesTestPath, loadMedicalLaboratoryAccreditationStatesFixture(t))
			jsonRepository, err := NewJSONRepository(root, 1<<20)
			if err != nil {
				t.Fatal(err)
			}
			repository, err := NewMedicalLaboratoryAccreditationRepository(jsonRepository, medicalLabAccreditationTestPath, medicalLabStatesTestPath)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("error = %v, want invalid dataset", err)
			}
		})
	}
}

func TestMedicalLaboratoryAccreditationRepositoryRejectsInvalidDatasetValues(t *testing.T) {
	for _, test := range []struct {
		name   string
		mutate func([]models.MedicalLaboratoryAccreditation) []models.MedicalLaboratoryAccreditation
	}{
		{"duplicate id", func(records []models.MedicalLaboratoryAccreditation) []models.MedicalLaboratoryAccreditation {
			records[1].ID = records[0].ID
			return records
		}},
		{"invalid status", func(records []models.MedicalLaboratoryAccreditation) []models.MedicalLaboratoryAccreditation {
			records[0].AccreditationStatus = "licensed"
			return records
		}},
		{"invalid state", func(records []models.MedicalLaboratoryAccreditation) []models.MedicalLaboratoryAccreditation {
			records[0].StateID = "not-a-state"
			return records
		}},
		{"invalid country", func(records []models.MedicalLaboratoryAccreditation) []models.MedicalLaboratoryAccreditation {
			records[0].CountryCode = "US"
			return records
		}},
		{"invalid date", func(records []models.MedicalLaboratoryAccreditation) []models.MedicalLaboratoryAccreditation {
			records[0].ExpiryDate = "2030/01/15"
			return records
		}},
		{"invalid ordering", func(records []models.MedicalLaboratoryAccreditation) []models.MedicalLaboratoryAccreditation {
			records[0], records[1] = records[1], records[0]
			return records
		}},
		{"wrong count", func(records []models.MedicalLaboratoryAccreditation) []models.MedicalLaboratoryAccreditation {
			return records[:29]
		}},
		{"bad arithmetic", func(records []models.MedicalLaboratoryAccreditation) []models.MedicalLaboratoryAccreditation {
			records[0].AccreditationStatus = "expired"
			return records
		}},
		{"personal marker", func(records []models.MedicalLaboratoryAccreditation) []models.MedicalLaboratoryAccreditation {
			records[0].Address = "Superintendent office, Lagos"
			return records
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			stub := &medicalLabAccreditationJSONStub{records: test.mutate(loadMedicalLaboratoryAccreditationFixture(t)), states: loadMedicalLaboratoryAccreditationStatesFixture(t)}
			repository := mustNewMedicalLaboratoryAccreditationRepository(t, stub)
			if _, err := repository.ListMedicalLaboratoryAccreditations(context.Background(), interfaces.MedicalLaboratoryAccreditationQuery{}); !errors.Is(err, interfaces.ErrInvalidDatasetFile) {
				t.Fatalf("error = %v, want invalid dataset", err)
			}
		})
	}
}

func newRealMedicalLaboratoryAccreditationRepository(t testing.TB) *MedicalLaboratoryAccreditationFileRepository {
	t.Helper()
	jsonRepository, err := NewJSONRepository("../../../datasets", 64<<20)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := NewMedicalLaboratoryAccreditationRepository(jsonRepository, medicalLabAccreditationTestPath, medicalLabStatesTestPath)
	if err != nil {
		t.Fatal(err)
	}
	return repository
}

func mustNewMedicalLaboratoryAccreditationRepository(t testing.TB, stub *medicalLabAccreditationJSONStub) *MedicalLaboratoryAccreditationFileRepository {
	t.Helper()
	repository, err := NewMedicalLaboratoryAccreditationRepository(stub, medicalLabAccreditationTestPath, medicalLabStatesTestPath)
	if err != nil {
		t.Fatal(err)
	}
	return repository
}

func loadMedicalLaboratoryAccreditationFixture(t testing.TB) []models.MedicalLaboratoryAccreditation {
	t.Helper()
	var items []models.MedicalLaboratoryAccreditation
	decodeTestJSON(t, "../../../datasets/healthcare/medical_laboratory_accreditations.json", &items)
	return items
}

func loadMedicalLaboratoryAccreditationStatesFixture(t testing.TB) []models.State {
	t.Helper()
	var items []models.State
	decodeTestJSON(t, "../../../datasets/geography/states.json", &items)
	return items
}

type medicalLabAccreditationJSONStub struct {
	records         []models.MedicalLaboratoryAccreditation
	states          []models.State
	recordsErr      error
	failRecordsOnce bool
	calls           map[string]int
}

func (s *medicalLabAccreditationJSONStub) Decode(_ context.Context, path string, destination any) error {
	if s.calls == nil {
		s.calls = map[string]int{}
	}
	s.calls[path]++
	if path == medicalLabAccreditationTestPath {
		if s.failRecordsOnce {
			s.failRecordsOnce = false
			return interfaces.ErrInvalidDatasetFile
		}
		if s.recordsErr != nil {
			return s.recordsErr
		}
		*(destination.(*[]models.MedicalLaboratoryAccreditation)) = append([]models.MedicalLaboratoryAccreditation(nil), s.records...)
		return nil
	}
	if path == medicalLabStatesTestPath {
		*(destination.(*[]models.State)) = append([]models.State(nil), s.states...)
		return nil
	}
	return fmt.Errorf("unexpected path %s", path)
}

func writeMedicalLabAccreditationTestFile(t testing.TB, root, relativePath string, data []byte) {
	t.Helper()
	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeJSONTestFile(t testing.TB, root, relativePath string, value any) {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	writeMedicalLabAccreditationTestFile(t, root, relativePath, data)
}
