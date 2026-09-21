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

func TestNEMAZonalTerritorialOperationOfficeServiceListValidationAndFilters(t *testing.T) {
	stub := &nemaOfficeServiceRepositoryStub{}
	service, err := NewNEMAZonalTerritorialOperationOfficeService(stub)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.ListNEMAZonalTerritorialOperationOffices(context.Background(), NEMAZonalTerritorialOperationOfficeQuery{
		StateID:    " lagos ",
		OfficeType: " zonal_territorial_operation_office ",
		Search:     " Lagos ",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := interfaces.NEMAZonalTerritorialOperationOfficeQuery{Page: 1, PageSize: 50, StateID: "lagos", OfficeType: "zonal_territorial_operation_office", Search: "Lagos"}
	if !reflect.DeepEqual(stub.queries[0], want) {
		t.Fatalf("query=%#v want %#v", stub.queries[0], want)
	}
	if result.Records == nil {
		t.Fatal("nil records")
	}
	result.Records[0].Name = "mutated"
	again, err := service.ListNEMAZonalTerritorialOperationOffices(context.Background(), NEMAZonalTerritorialOperationOfficeQuery{StateID: "lagos"})
	if err != nil {
		t.Fatal(err)
	}
	if len(again.Records) != 1 || again.Records[0].Name == "mutated" {
		t.Fatalf("unexpected caller-owned result: %#v", again.Records)
	}
	valid := NEMAZonalTerritorialOperationOfficeQuery{Page: 1, PageSize: 100, StateID: "fct", OfficeType: "zonal_territorial_operation_office", Search: strings.Repeat("a", NEMAZonalTerritorialOperationOfficeMaxSearchLength)}
	if _, err := service.ListNEMAZonalTerritorialOperationOffices(context.Background(), valid); err != nil {
		t.Fatalf("valid boundary err=%v", err)
	}
}

func TestNEMAZonalTerritorialOperationOfficeServiceRejectsInvalidInputBeforeRepository(t *testing.T) {
	for _, tc := range []struct {
		name  string
		query NEMAZonalTerritorialOperationOfficeQuery
		want  error
	}{
		{"bad page", NEMAZonalTerritorialOperationOfficeQuery{Page: -1}, ErrInvalidNEMAZonalTerritorialOperationOfficePagination},
		{"bad size", NEMAZonalTerritorialOperationOfficeQuery{PageSize: 101}, ErrInvalidNEMAZonalTerritorialOperationOfficePagination},
		{"empty state", NEMAZonalTerritorialOperationOfficeQuery{StateID: " "}, ErrInvalidNEMAZonalTerritorialOperationOfficeStateID},
		{"bad state", NEMAZonalTerritorialOperationOfficeQuery{StateID: "not-a-state"}, ErrInvalidNEMAZonalTerritorialOperationOfficeStateID},
		{"empty office type", NEMAZonalTerritorialOperationOfficeQuery{OfficeType: " "}, ErrInvalidNEMAZonalTerritorialOperationOfficeType},
		{"bad office type", NEMAZonalTerritorialOperationOfficeQuery{OfficeType: "regional"}, ErrInvalidNEMAZonalTerritorialOperationOfficeType},
		{"empty search", NEMAZonalTerritorialOperationOfficeQuery{Search: " "}, ErrInvalidNEMAZonalTerritorialOperationOfficeSearch},
		{"long search", NEMAZonalTerritorialOperationOfficeQuery{Search: strings.Repeat("a", NEMAZonalTerritorialOperationOfficeMaxSearchLength+1)}, ErrInvalidNEMAZonalTerritorialOperationOfficeSearch},
		{"newline search", NEMAZonalTerritorialOperationOfficeQuery{Search: "lag\nos"}, ErrInvalidNEMAZonalTerritorialOperationOfficeSearch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &nemaOfficeServiceRepositoryStub{}
			service, _ := NewNEMAZonalTerritorialOperationOfficeService(stub)
			if _, err := service.ListNEMAZonalTerritorialOperationOffices(context.Background(), tc.query); !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want %v", err, tc.want)
			}
			if len(stub.queries) != 0 {
				t.Fatal("repository was called for invalid input")
			}
		})
	}
}

func TestNEMAZonalTerritorialOperationOfficeServiceDetailAndErrorTranslation(t *testing.T) {
	stub := &nemaOfficeServiceRepositoryStub{}
	service, _ := NewNEMAZonalTerritorialOperationOfficeService(stub)
	record, err := service.GetNEMAZonalTerritorialOperationOffice(context.Background(), " nema-lagos-zonal-territorial-operation-office ")
	if err != nil || record.ID != "nema-lagos-zonal-territorial-operation-office" || stub.ids[0] != "nema-lagos-zonal-territorial-operation-office" {
		t.Fatalf("record=%#v ids=%#v err=%v", record, stub.ids, err)
	}
	if _, err := service.GetNEMAZonalTerritorialOperationOffice(context.Background(), "bad_id"); !errors.Is(err, ErrInvalidNEMAZonalTerritorialOperationOfficeID) {
		t.Fatalf("invalid id err=%v", err)
	}
	if _, err := service.GetNEMAZonalTerritorialOperationOffice(context.Background(), strings.Repeat("a", 256)); !errors.Is(err, ErrInvalidNEMAZonalTerritorialOperationOfficeID) {
		t.Fatalf("long id err=%v", err)
	}
	for _, tc := range []struct {
		name string
		err  error
		want error
	}{
		{"not found", interfaces.ErrNEMAZonalTerritorialOperationOfficeNotFound, ErrNEMAZonalTerritorialOperationOfficeNotFound},
		{"invalid query", interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeQuery, ErrInvalidNEMAZonalTerritorialOperationOfficePagination},
		{"invalid state", interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeStateFilter, ErrInvalidNEMAZonalTerritorialOperationOfficeStateID},
		{"invalid type", interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeTypeFilter, ErrInvalidNEMAZonalTerritorialOperationOfficeType},
		{"invalid search", interfaces.ErrInvalidNEMAZonalTerritorialOperationOfficeSearch, ErrInvalidNEMAZonalTerritorialOperationOfficeSearch},
		{"dataset", interfaces.ErrInvalidDatasetFile, ErrInvalidNEMAZonalTerritorialOperationOfficeDataset},
		{"unavailable", interfaces.ErrDatasetFileUnavailable, ErrInvalidNEMAZonalTerritorialOperationOfficeDataset},
		{"canceled", context.Canceled, context.Canceled},
		{"deadline", context.DeadlineExceeded, context.DeadlineExceeded},
		{"unexpected", errors.New("/internal/source/path decoder detail"), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub.err = tc.err
			_, err := service.ListNEMAZonalTerritorialOperationOffices(context.Background(), NEMAZonalTerritorialOperationOfficeQuery{})
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want %v", err, tc.want)
			}
			if tc.want == nil && (err == nil || strings.Contains(err.Error(), "internal/source") || strings.Contains(err.Error(), "decoder")) {
				t.Fatalf("unexpected unsanitized err=%v", err)
			}
		})
	}
}

func TestNEMAZonalTerritorialOperationOfficeServiceConstructorAndContext(t *testing.T) {
	if _, err := NewNEMAZonalTerritorialOperationOfficeService(nil); err == nil {
		t.Fatal("expected nil repository error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service, _ := NewNEMAZonalTerritorialOperationOfficeService(&nemaOfficeServiceRepositoryStub{})
	if _, err := service.ListNEMAZonalTerritorialOperationOffices(ctx, NEMAZonalTerritorialOperationOfficeQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled list err=%v", err)
	}
	if _, err := service.GetNEMAZonalTerritorialOperationOffice(ctx, "nema-lagos-zonal-territorial-operation-office"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled detail err=%v", err)
	}
}

type nemaOfficeServiceRepositoryStub struct {
	queries []interfaces.NEMAZonalTerritorialOperationOfficeQuery
	ids     []string
	err     error
}

func (s *nemaOfficeServiceRepositoryStub) ListNEMAZonalTerritorialOperationOffices(_ context.Context, query interfaces.NEMAZonalTerritorialOperationOfficeQuery) (interfaces.NEMAZonalTerritorialOperationOfficeListResult, error) {
	s.queries = append(s.queries, query)
	if s.err != nil {
		return interfaces.NEMAZonalTerritorialOperationOfficeListResult{}, s.err
	}
	return interfaces.NEMAZonalTerritorialOperationOfficeListResult{
		Records:    []models.NEMAZonalTerritorialOperationOffice{{ID: "nema-lagos-zonal-territorial-operation-office", Name: "NEMA Lagos Office", OfficeType: "zonal_territorial_operation_office", StateID: "lagos", CountryCode: "NG"}},
		Page:       query.Page,
		PageSize:   query.PageSize,
		Total:      1,
		TotalPages: 1,
	}, nil
}

func (s *nemaOfficeServiceRepositoryStub) GetNEMAZonalTerritorialOperationOffice(_ context.Context, id string) (models.NEMAZonalTerritorialOperationOffice, error) {
	s.ids = append(s.ids, id)
	if s.err != nil {
		return models.NEMAZonalTerritorialOperationOffice{}, s.err
	}
	return models.NEMAZonalTerritorialOperationOffice{ID: id, Name: "NEMA Lagos Office", OfficeType: "zonal_territorial_operation_office", StateID: "lagos", CountryCode: "NG"}, nil
}
