package app

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/datasets"
	"github.com/AbdulQuayyum/softdata-api/internal/handlers"
	"github.com/AbdulQuayyum/softdata-api/internal/models"
	fileRepo "github.com/AbdulQuayyum/softdata-api/internal/repository/file"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type nemaOfficeAppJSONRepository struct{}

func (nemaOfficeAppJSONRepository) Decode(context.Context, string, any) error { return nil }

type nemaOfficeAppRepository struct{}

func (nemaOfficeAppRepository) ListNEMAZonalTerritorialOperationOffices(context.Context, interfaces.NEMAZonalTerritorialOperationOfficeQuery) (interfaces.NEMAZonalTerritorialOperationOfficeListResult, error) {
	return interfaces.NEMAZonalTerritorialOperationOfficeListResult{}, nil
}

func (nemaOfficeAppRepository) GetNEMAZonalTerritorialOperationOffice(context.Context, string) (models.NEMAZonalTerritorialOperationOffice, error) {
	return models.NEMAZonalTerritorialOperationOffice{}, nil
}

func TestNEMAZonalTerritorialOperationOfficeDependencyConstructionUsesDatasetPath(t *testing.T) {
	var gotPath string
	service, err := buildNEMAZonalTerritorialOperationOfficeServiceFromJSONRepository(context.Background(), nemaOfficeAppJSONRepository{},
		func(repository interfaces.JSONFileRepository, recordsPath string) (interfaces.NEMAZonalTerritorialOperationOfficeRepository, error) {
			if repository == nil {
				t.Fatal("nil JSON repository")
			}
			gotPath = recordsPath
			return nemaOfficeAppRepository{}, nil
		},
		func(repository interfaces.NEMAZonalTerritorialOperationOfficeRepository) (nemaZonalTerritorialOperationOfficeService, error) {
			return services.NewNEMAZonalTerritorialOperationOfficeService(repository)
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if service == nil || gotPath != emergencyNEMAZonalTerritorialOperationOfficesRelativePath {
		t.Fatalf("service=%v path=%q", service, gotPath)
	}
}

func TestNEMAZonalTerritorialOperationOfficeHandlerConstructionVerifiesAndSharesService(t *testing.T) {
	var gotPath string
	var handlerService nemaZonalTerritorialOperationOfficeService
	service, handler, err := buildNEMAZonalTerritorialOperationOfficeHandler(context.Background(), nemaOfficeAppJSONRepository{},
		func(repository interfaces.JSONFileRepository, recordsPath string) (interfaces.NEMAZonalTerritorialOperationOfficeRepository, error) {
			if repository == nil {
				t.Fatal("nil JSON repository")
			}
			gotPath = recordsPath
			return nemaOfficeStartupRepository{}, nil
		},
		func(repository interfaces.NEMAZonalTerritorialOperationOfficeRepository) (nemaZonalTerritorialOperationOfficeService, error) {
			return services.NewNEMAZonalTerritorialOperationOfficeService(repository)
		},
		func(service nemaZonalTerritorialOperationOfficeService) (*handlers.NEMAZonalTerritorialOperationOfficeHandler, error) {
			handlerService = service
			return handlers.NewNEMAZonalTerritorialOperationOfficeHandler(service)
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if service == nil || handler == nil || handlerService != service || gotPath != emergencyNEMAZonalTerritorialOperationOfficesRelativePath {
		t.Fatalf("service=%v handler=%v handlerService=%v path=%q", service, handler, handlerService, gotPath)
	}
}

func TestNEMAZonalTerritorialOperationOfficeDependencyConstructionPreservesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := buildNEMAZonalTerritorialOperationOfficeServiceFromJSONRepository(ctx, nemaOfficeAppJSONRepository{},
		func(interfaces.JSONFileRepository, string) (interfaces.NEMAZonalTerritorialOperationOfficeRepository, error) {
			t.Fatal("repository constructor should not be called")
			return nil, nil
		},
		func(interfaces.NEMAZonalTerritorialOperationOfficeRepository) (nemaZonalTerritorialOperationOfficeService, error) {
			return nil, nil
		},
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}

func TestNEMAZonalTerritorialOperationOfficeStartupVerification(t *testing.T) {
	if err := verifyNEMAZonalTerritorialOperationOffices(context.Background(), nemaOfficeStartupService{}); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		svc  nemaZonalTerritorialOperationOfficeService
		want string
	}{
		{"nil service", nil, "service is required"},
		{"wrong total", nemaOfficeStartupService{wrongTotal: true}, "invalid verification page"},
		{"empty first page", nemaOfficeStartupService{emptyFirstPage: true}, "invalid verification page"},
		{"nil first page", nemaOfficeStartupService{nilFirstPage: true}, "invalid verification page"},
		{"wrong first anchor", nemaOfficeStartupService{wrongFirstAnchor: true}, "anchor mismatch"},
		{"wrong last anchor", nemaOfficeStartupService{wrongLastAnchor: true}, "anchor mismatch"},
		{"missing detail", nemaOfficeStartupService{missingDetail: true}, "retrieve anchor failed"},
		{"invalid record", nemaOfficeStartupService{invalidRecord: true}, "invalid record contract"},
		{"wrong office type total", nemaOfficeStartupService{wrongOfficeTypeTotal: true}, "invalid office-type verification page"},
		{"wrong state result", nemaOfficeStartupService{wrongStateResult: true}, "invalid state verification page"},
		{"non-empty beyond final", nemaOfficeStartupService{nonEmptyBeyondFinal: true}, "invalid beyond-final page"},
		{"nil beyond final", nemaOfficeStartupService{nilBeyondFinal: true}, "invalid beyond-final page"},
		{"list failure", nemaOfficeStartupService{listErr: errors.New("/tmp/raw decoder failure")}, "load verification page failed"},
		{"detail failure", nemaOfficeStartupService{detailErr: errors.New("/tmp/raw decoder failure")}, "retrieve anchor failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := verifyNEMAZonalTerritorialOperationOffices(context.Background(), tc.svc)
			if err == nil || !strings.Contains(err.Error(), tc.want) || strings.Contains(err.Error(), "/tmp/raw") {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestNEMAZonalTerritorialOperationOfficeStartupVerificationPreservesContextErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := verifyNEMAZonalTerritorialOperationOffices(ctx, nemaOfficeStartupService{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled err=%v", err)
	}
	if err := verifyNEMAZonalTerritorialOperationOffices(context.Background(), nemaOfficeStartupService{listErr: context.DeadlineExceeded}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("deadline err=%v", err)
	}
}

func TestNEMAZonalTerritorialOperationOfficeStartupAndHTTPShareRepositoryCache(t *testing.T) {
	jsonRepository := &nemaOfficeCountingJSONRepository{}
	repository, err := fileRepo.NewNEMAZonalTerritorialOperationOfficeRepository(jsonRepository, emergencyNEMAZonalTerritorialOperationOfficesRelativePath)
	if err != nil {
		t.Fatal(err)
	}
	service, err := services.NewNEMAZonalTerritorialOperationOfficeService(repository)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyNEMAZonalTerritorialOperationOffices(context.Background(), service); err != nil {
		t.Fatal(err)
	}
	if jsonRepository.count.Load() != 1 {
		t.Fatalf("startup decodes = %d", jsonRepository.count.Load())
	}
	if result, err := service.ListNEMAZonalTerritorialOperationOffices(context.Background(), services.NEMAZonalTerritorialOperationOfficeQuery{}); err != nil || result.Total != 17 {
		t.Fatalf("later access result=%#v err=%v", result, err)
	}
	if jsonRepository.count.Load() != 1 {
		t.Fatalf("later access decoded again: %d", jsonRepository.count.Load())
	}
	for _, path := range jsonRepository.paths {
		if strings.Contains(path, "metadata") || strings.Contains(path, "reconciliation") {
			t.Fatalf("runtime opened metadata/reconciliation path %q", path)
		}
	}
}

func TestNEMAZonalTerritorialOperationOfficeEmbeddedDatasetFallback(t *testing.T) {
	jsonRepository, err := fileRepo.NewEmbeddedJSONRepository(datasets.Files(), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	repository, err := fileRepo.NewNEMAZonalTerritorialOperationOfficeRepository(jsonRepository, emergencyNEMAZonalTerritorialOperationOfficesRelativePath)
	if err != nil {
		t.Fatal(err)
	}
	service, err := services.NewNEMAZonalTerritorialOperationOfficeService(repository)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.ListNEMAZonalTerritorialOperationOffices(context.Background(), services.NEMAZonalTerritorialOperationOfficeQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 17 || len(result.Records) != 17 {
		t.Fatalf("embedded result mismatch: %#v", result)
	}
	if _, err := fs.ReadFile(datasets.Files(), "metadata/emergency/nema_zonal_territorial_operation_offices_reconciliation/index.json"); err == nil {
		t.Fatal("reconciliation metadata must not be embedded for runtime")
	}
}

type nemaOfficeStartupRepository struct{}

func (nemaOfficeStartupRepository) ListNEMAZonalTerritorialOperationOffices(ctx context.Context, query interfaces.NEMAZonalTerritorialOperationOfficeQuery) (interfaces.NEMAZonalTerritorialOperationOfficeListResult, error) {
	return nemaOfficeStartupService{}.ListNEMAZonalTerritorialOperationOffices(ctx, query)
}

func (nemaOfficeStartupRepository) GetNEMAZonalTerritorialOperationOffice(ctx context.Context, id string) (models.NEMAZonalTerritorialOperationOffice, error) {
	return nemaOfficeStartupService{}.GetNEMAZonalTerritorialOperationOffice(ctx, id)
}

type nemaOfficeStartupService struct {
	wrongTotal           bool
	emptyFirstPage       bool
	nilFirstPage         bool
	wrongFirstAnchor     bool
	wrongLastAnchor      bool
	missingDetail        bool
	invalidRecord        bool
	wrongOfficeTypeTotal bool
	wrongStateResult     bool
	nonEmptyBeyondFinal  bool
	nilBeyondFinal       bool
	listErr              error
	detailErr            error
}

func (s nemaOfficeStartupService) ListNEMAZonalTerritorialOperationOffices(_ context.Context, query interfaces.NEMAZonalTerritorialOperationOfficeQuery) (interfaces.NEMAZonalTerritorialOperationOfficeListResult, error) {
	if s.listErr != nil {
		return interfaces.NEMAZonalTerritorialOperationOfficeListResult{}, s.listErr
	}
	total := 17
	if s.wrongTotal {
		total = 16
	}
	totalPages := total
	if query.PageSize > 1 && total > 0 {
		totalPages = (total + query.PageSize - 1) / query.PageSize
	}
	record := nemaOfficeStartupFirstRecord()
	if s.invalidRecord {
		record.CountryCode = "ZZ"
	}
	switch {
	case query.OfficeType != "":
		if s.wrongOfficeTypeTotal {
			total = 16
			totalPages = 16
		}
	case query.StateID != "":
		total = 1
		totalPages = 1
		if s.wrongStateResult {
			total = 2
			totalPages = 2
		}
	case query.Page == 18:
		records := []models.NEMAZonalTerritorialOperationOffice{}
		if s.nonEmptyBeyondFinal {
			records = []models.NEMAZonalTerritorialOperationOffice{record}
		}
		if s.nilBeyondFinal {
			records = nil
		}
		return interfaces.NEMAZonalTerritorialOperationOfficeListResult{Records: records, Page: query.Page, PageSize: query.PageSize, Total: 17, TotalPages: 17}, nil
	}
	records := []models.NEMAZonalTerritorialOperationOffice{record}
	if s.emptyFirstPage {
		records = []models.NEMAZonalTerritorialOperationOffice{}
	}
	if s.nilFirstPage {
		records = nil
	}
	return interfaces.NEMAZonalTerritorialOperationOfficeListResult{Records: records, Page: query.Page, PageSize: query.PageSize, Total: total, TotalPages: totalPages}, nil
}

func (s nemaOfficeStartupService) GetNEMAZonalTerritorialOperationOffice(_ context.Context, id string) (models.NEMAZonalTerritorialOperationOffice, error) {
	if s.detailErr != nil {
		return models.NEMAZonalTerritorialOperationOffice{}, s.detailErr
	}
	if s.missingDetail {
		return models.NEMAZonalTerritorialOperationOffice{}, services.ErrNEMAZonalTerritorialOperationOfficeNotFound
	}
	switch id {
	case nemaZonalTerritorialOperationOfficeFirstID:
		record := nemaOfficeStartupFirstRecord()
		if s.wrongFirstAnchor {
			record.Name = "Wrong Office"
		}
		return record, nil
	case nemaZonalTerritorialOperationOfficeLastID:
		record := models.NEMAZonalTerritorialOperationOffice{ID: nemaZonalTerritorialOperationOfficeLastID, Name: nemaZonalTerritorialOperationOfficeLastName, OfficeType: nemaZonalTerritorialOperationOfficeType, StateID: nemaZonalTerritorialOperationOfficeLastState, CountryCode: "NG"}
		if s.wrongLastAnchor {
			record.StateID = "lagos"
		}
		return record, nil
	default:
		return models.NEMAZonalTerritorialOperationOffice{}, services.ErrNEMAZonalTerritorialOperationOfficeNotFound
	}
}

func nemaOfficeStartupFirstRecord() models.NEMAZonalTerritorialOperationOffice {
	return models.NEMAZonalTerritorialOperationOffice{ID: nemaZonalTerritorialOperationOfficeFirstID, Name: nemaZonalTerritorialOperationOfficeFirstName, OfficeType: nemaZonalTerritorialOperationOfficeType, StateID: nemaZonalTerritorialOperationOfficeFirstState, CountryCode: "NG"}
}

type nemaOfficeCountingJSONRepository struct {
	count atomic.Int64
	paths []string
}

func (r *nemaOfficeCountingJSONRepository) Decode(ctx context.Context, path string, target any) error {
	r.count.Add(1)
	r.paths = append(r.paths, path)
	repository, err := fileRepo.NewEmbeddedJSONRepository(datasets.Files(), 1<<20)
	if err != nil {
		return err
	}
	return repository.Decode(ctx, path, target)
}
