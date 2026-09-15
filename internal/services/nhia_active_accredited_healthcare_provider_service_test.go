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

func TestNHIAActiveAccreditedHealthcareProviderServiceListValidationAndPropagation(t *testing.T) {
	stub := &nhiaHCPServiceRepositoryStub{}
	service, err := NewNHIAActiveAccreditedHealthcareProviderService(stub)
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), NHIAActiveAccreditedHealthcareProviderQuery{
		ProviderCode:  " FCT/0001/P ",
		FacilityType:  " primary ",
		ListingStatus: " active_accredited ",
		Search:        " Clinic ",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 50, ProviderCode: "FCT/0001/P", FacilityType: "primary", ListingStatus: "active_accredited", Search: "Clinic"}
	if !reflect.DeepEqual(stub.queries[0], want) {
		t.Fatalf("query=%#v want %#v", stub.queries[0], want)
	}
	if result.Records == nil {
		t.Fatal("nil records")
	}

	valid := interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: 1, PageSize: 100, ProviderCode: "AB/0001/P", FacilityType: "primary_and_secondary", ListingStatus: "active_accredited", Search: strings.Repeat("a", NHIAActiveAccreditedHealthcareProviderMaxSearchLength)}
	if _, err := service.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), valid); err != nil {
		t.Fatalf("valid boundary err=%v", err)
	}
	for _, tc := range []struct {
		name  string
		query interfaces.NHIAActiveAccreditedHealthcareProviderQuery
		want  error
	}{
		{"bad page", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Page: -1}, ErrInvalidNHIAActiveAccreditedHealthcareProviderPagination},
		{"bad size", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{PageSize: 101}, ErrInvalidNHIAActiveAccreditedHealthcareProviderPagination},
		{"empty code", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{ProviderCode: " "}, ErrInvalidNHIAActiveAccreditedHealthcareProviderCode},
		{"bad code", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{ProviderCode: "FCT/1/P"}, ErrInvalidNHIAActiveAccreditedHealthcareProviderCode},
		{"bad facility", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{FacilityType: "secondary"}, ErrInvalidNHIAActiveAccreditedHealthcareProviderFacilityType},
		{"bad status", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{ListingStatus: "licensed"}, ErrInvalidNHIAActiveAccreditedHealthcareProviderListingStatus},
		{"empty search", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Search: " "}, ErrInvalidNHIAActiveAccreditedHealthcareProviderSearch},
		{"long search", interfaces.NHIAActiveAccreditedHealthcareProviderQuery{Search: strings.Repeat("a", NHIAActiveAccreditedHealthcareProviderMaxSearchLength+1)}, ErrInvalidNHIAActiveAccreditedHealthcareProviderSearch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := service.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), tc.query); !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want %v", err, tc.want)
			}
		})
	}
}

func TestNHIAActiveAccreditedHealthcareProviderServiceDetailAndErrors(t *testing.T) {
	stub := &nhiaHCPServiceRepositoryStub{}
	service, _ := NewNHIAActiveAccreditedHealthcareProviderService(stub)
	record, err := service.GetNHIAActiveAccreditedHealthcareProvider(context.Background(), " fct-0001-p ")
	if err != nil || record.ID != "fct-0001-p" || stub.ids[0] != "fct-0001-p" {
		t.Fatalf("record=%#v ids=%#v err=%v", record, stub.ids, err)
	}
	if _, err := service.GetNHIAActiveAccreditedHealthcareProvider(context.Background(), "bad_id"); !errors.Is(err, ErrInvalidNHIAActiveAccreditedHealthcareProviderID) {
		t.Fatalf("invalid id err=%v", err)
	}
	if _, err := service.GetNHIAActiveAccreditedHealthcareProvider(context.Background(), strings.Repeat("a", 256)); !errors.Is(err, ErrInvalidNHIAActiveAccreditedHealthcareProviderID) {
		t.Fatalf("long id err=%v", err)
	}

	for _, tc := range []struct {
		name string
		err  error
		want error
	}{
		{"not found", interfaces.ErrNHIAActiveAccreditedHealthcareProviderNotFound, ErrNHIAActiveAccreditedHealthcareProviderNotFound},
		{"dataset", interfaces.ErrInvalidDatasetFile, ErrInvalidNHIAActiveAccreditedHealthcareProviderDataset},
		{"canceled", context.Canceled, context.Canceled},
		{"deadline", context.DeadlineExceeded, context.DeadlineExceeded},
		{"unexpected", errors.New("/secret/path decoder address"), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub.err = tc.err
			_, err := service.ListNHIAActiveAccreditedHealthcareProviders(context.Background(), NHIAActiveAccreditedHealthcareProviderQuery{})
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want %v", err, tc.want)
			}
			if tc.want == nil && (err == nil || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "address")) {
				t.Fatalf("unexpected unsanitized err=%v", err)
			}
		})
	}
}

func TestNHIAActiveAccreditedHealthcareProviderServiceConstructorAndCanceledContext(t *testing.T) {
	if _, err := NewNHIAActiveAccreditedHealthcareProviderService(nil); err == nil {
		t.Fatal("expected nil repository error")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	service, _ := NewNHIAActiveAccreditedHealthcareProviderService(&nhiaHCPServiceRepositoryStub{})
	if _, err := service.ListNHIAActiveAccreditedHealthcareProviders(ctx, NHIAActiveAccreditedHealthcareProviderQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled list err=%v", err)
	}
}

type nhiaHCPServiceRepositoryStub struct {
	queries []interfaces.NHIAActiveAccreditedHealthcareProviderQuery
	ids     []string
	err     error
}

func (s *nhiaHCPServiceRepositoryStub) ListNHIAActiveAccreditedHealthcareProviders(_ context.Context, query interfaces.NHIAActiveAccreditedHealthcareProviderQuery) (interfaces.NHIAActiveAccreditedHealthcareProviderListResult, error) {
	s.queries = append(s.queries, query)
	if s.err != nil {
		return interfaces.NHIAActiveAccreditedHealthcareProviderListResult{}, s.err
	}
	return interfaces.NHIAActiveAccreditedHealthcareProviderListResult{Records: nil, Page: query.Page, PageSize: query.PageSize, Total: 0, TotalPages: 0}, nil
}

func (s *nhiaHCPServiceRepositoryStub) GetNHIAActiveAccreditedHealthcareProvider(_ context.Context, id string) (models.NHIAActiveAccreditedHealthcareProvider, error) {
	s.ids = append(s.ids, id)
	if s.err != nil {
		return models.NHIAActiveAccreditedHealthcareProvider{}, s.err
	}
	return models.NHIAActiveAccreditedHealthcareProvider{ID: id, Name: "WILDOT CLINIC", CountryCode: "NG", ProviderCode: "FCT/0001/P", FacilityType: "primary", ListingStatus: "active_accredited"}, nil
}
