package services

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
)

func TestNHIAStateSocialHealthInsuranceAgencyServiceListValidationAndPropagation(t *testing.T) {
	stub := &nhiaSSHIAServiceRepositoryStub{result: interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{Records: []models.NHIAStateSocialHealthInsuranceAgency{{ID: "lashma", Name: "LASHMA", StateID: "lagos", CountryCode: "NG", OrganisationType: "state_social_health_insurance_agency"}}, Page: 1, PageSize: 50, Total: 1, TotalPages: 1}}
	service := mustNewNHIASSHIAService(t, stub)
	result, err := service.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), NHIAStateSocialHealthInsuranceAgencyQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if result.PageSize != 50 || stub.query.Page != 1 || stub.query.PageSize != 50 {
		t.Fatalf("default normalization failed result=%#v query=%#v", result, stub.query)
	}
	result.Records[0].Name = "mutated"
	again, err := service.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), NHIAStateSocialHealthInsuranceAgencyQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if again.Records[0].Name == "mutated" {
		t.Fatal("service returned caller-owned cached mutation")
	}

	for _, tc := range []struct {
		name  string
		query NHIAStateSocialHealthInsuranceAgencyQuery
		err   error
	}{
		{"valid pagination and trimmed state/search", NHIAStateSocialHealthInsuranceAgencyQuery{Page: 2, PageSize: 10, StateID: " lagos ", Search: " lashma "}, nil},
		{"invalid page", NHIAStateSocialHealthInsuranceAgencyQuery{Page: -1}, ErrInvalidNHIAStateSocialHealthInsuranceAgencyPagination},
		{"invalid page size", NHIAStateSocialHealthInsuranceAgencyQuery{PageSize: 101}, ErrInvalidNHIAStateSocialHealthInsuranceAgencyPagination},
		{"raw alias rejected", NHIAStateSocialHealthInsuranceAgencyQuery{StateID: "AKS"}, ErrInvalidNHIAStateSocialHealthInsuranceAgencyStateID},
		{"unknown state rejected", NHIAStateSocialHealthInsuranceAgencyQuery{StateID: "unknown-state"}, ErrInvalidNHIAStateSocialHealthInsuranceAgencyStateID},
		{"invalid state syntax", NHIAStateSocialHealthInsuranceAgencyQuery{StateID: "akwa_ibom"}, ErrInvalidNHIAStateSocialHealthInsuranceAgencyStateID},
		{"empty state", NHIAStateSocialHealthInsuranceAgencyQuery{StateID: " "}, ErrInvalidNHIAStateSocialHealthInsuranceAgencyStateID},
		{"empty search", NHIAStateSocialHealthInsuranceAgencyQuery{Search: " "}, ErrInvalidNHIAStateSocialHealthInsuranceAgencySearch},
		{"excess search", NHIAStateSocialHealthInsuranceAgencyQuery{Search: strings.Repeat("a", 101)}, ErrInvalidNHIAStateSocialHealthInsuranceAgencySearch},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stub := &nhiaSSHIAServiceRepositoryStub{result: interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{}}
			service := mustNewNHIASSHIAService(t, stub)
			_, err := service.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), tc.query)
			if tc.err == nil {
				if err != nil {
					t.Fatal(err)
				}
				if stub.query.StateID != "lagos" || stub.query.Search != "lashma" || stub.query.Page != 2 || stub.query.PageSize != 10 {
					t.Fatalf("query not normalized: %#v", stub.query)
				}
			} else if !errors.Is(err, tc.err) {
				t.Fatalf("err=%v want %v", err, tc.err)
			}
		})
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyServiceGetAndErrors(t *testing.T) {
	record := models.NHIAStateSocialHealthInsuranceAgency{ID: "lashma", Name: "LASHMA", StateID: "lagos", CountryCode: "NG", OrganisationType: "state_social_health_insurance_agency"}
	service := mustNewNHIASSHIAService(t, &nhiaSSHIAServiceRepositoryStub{record: record})
	got, err := service.GetNHIAStateSocialHealthInsuranceAgency(context.Background(), " lashma ")
	if err != nil || got.ID != "lashma" {
		t.Fatalf("got=%#v err=%v", got, err)
	}
	if _, err := service.GetNHIAStateSocialHealthInsuranceAgency(context.Background(), strings.Repeat("a", 255)); err != nil {
		t.Fatalf("255-char slug rejected: %v", err)
	}
	for _, id := range []string{"", "bad_id", "has space", strings.Repeat("a", 256)} {
		if _, err := service.GetNHIAStateSocialHealthInsuranceAgency(context.Background(), id); !errors.Is(err, ErrInvalidNHIAStateSocialHealthInsuranceAgencyID) {
			t.Fatalf("id %q err=%v", id, err)
		}
	}
	for _, tc := range []struct {
		name string
		err  error
		want error
	}{
		{"not found", interfaces.ErrNHIAStateSocialHealthInsuranceAgencyNotFound, ErrNHIAStateSocialHealthInsuranceAgencyNotFound},
		{"query", interfaces.ErrInvalidNHIAStateSocialHealthInsuranceAgencyQuery, ErrInvalidNHIAStateSocialHealthInsuranceAgencyPagination},
		{"state", interfaces.ErrInvalidNHIAStateSocialHealthInsuranceAgencyStateFilter, ErrInvalidNHIAStateSocialHealthInsuranceAgencyStateID},
		{"search", interfaces.ErrInvalidNHIAStateSocialHealthInsuranceAgencySearch, ErrInvalidNHIAStateSocialHealthInsuranceAgencySearch},
		{"dataset", interfaces.ErrInvalidDatasetFile, ErrInvalidNHIAStateSocialHealthInsuranceAgencyDataset},
		{"unexpected", errors.New("secret path director"), nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := mustNewNHIASSHIAService(t, &nhiaSSHIAServiceRepositoryStub{err: tc.err})
			_, err := service.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), NHIAStateSocialHealthInsuranceAgencyQuery{})
			if tc.want != nil {
				if !errors.Is(err, tc.want) {
					t.Fatalf("err=%v want %v", err, tc.want)
				}
			} else if err == nil || strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), "director") {
				t.Fatalf("unexpected error not sanitized: %v", err)
			}
		})
	}
}

func TestNHIAStateSocialHealthInsuranceAgencyServiceContext(t *testing.T) {
	for _, deadline := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		if deadline {
			cancel()
			ctx, cancel = context.WithDeadline(context.Background(), time.Unix(0, 0))
		} else {
			cancel()
		}
		service := mustNewNHIASSHIAService(t, &nhiaSSHIAServiceRepositoryStub{})
		if _, err := service.ListNHIAStateSocialHealthInsuranceAgencies(ctx, NHIAStateSocialHealthInsuranceAgencyQuery{}); !errors.Is(err, ctx.Err()) {
			t.Fatalf("list context err=%v", err)
		}
		if _, err := service.GetNHIAStateSocialHealthInsuranceAgency(ctx, "lashma"); !errors.Is(err, ctx.Err()) {
			t.Fatalf("get context err=%v", err)
		}
		cancel()
	}
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		service := mustNewNHIASSHIAService(t, &nhiaSSHIAServiceRepositoryStub{err: cause})
		if _, err := service.ListNHIAStateSocialHealthInsuranceAgencies(context.Background(), NHIAStateSocialHealthInsuranceAgencyQuery{}); !errors.Is(err, cause) {
			t.Fatalf("repository context err=%v", err)
		}
	}
}

func mustNewNHIASSHIAService(t testing.TB, repo interfaces.NHIAStateSocialHealthInsuranceAgencyRepository) *NHIAStateSocialHealthInsuranceAgencyService {
	t.Helper()
	service, err := NewNHIAStateSocialHealthInsuranceAgencyService(repo)
	if err != nil {
		t.Fatal(err)
	}
	return service
}

type nhiaSSHIAServiceRepositoryStub struct {
	query  interfaces.NHIAStateSocialHealthInsuranceAgencyQuery
	result interfaces.NHIAStateSocialHealthInsuranceAgencyListResult
	record models.NHIAStateSocialHealthInsuranceAgency
	err    error
}

func (s *nhiaSSHIAServiceRepositoryStub) ListNHIAStateSocialHealthInsuranceAgencies(_ context.Context, query interfaces.NHIAStateSocialHealthInsuranceAgencyQuery) (interfaces.NHIAStateSocialHealthInsuranceAgencyListResult, error) {
	s.query = query
	if s.err != nil {
		return interfaces.NHIAStateSocialHealthInsuranceAgencyListResult{}, s.err
	}
	return s.result, nil
}

func (s *nhiaSSHIAServiceRepositoryStub) GetNHIAStateSocialHealthInsuranceAgency(_ context.Context, id string) (models.NHIAStateSocialHealthInsuranceAgency, error) {
	if s.err != nil {
		return models.NHIAStateSocialHealthInsuranceAgency{}, s.err
	}
	if s.record.ID == "" {
		return models.NHIAStateSocialHealthInsuranceAgency{ID: id}, nil
	}
	return s.record, nil
}
