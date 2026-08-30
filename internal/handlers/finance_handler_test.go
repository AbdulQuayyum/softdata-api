package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

type financeHandlerStub struct {
	listAllFn    func(context.Context) ([]models.PaymentServiceProvider, error)
	listByTypeFn func(context.Context, string) ([]models.PaymentServiceProvider, error)
	getFn        func(context.Context, string) (models.PaymentServiceProvider, error)

	listAllCalls    int
	listByTypeCalls int
	getCalls        int
	lastType        string
	lastID          string
}

func (s *financeHandlerStub) ListPaymentServiceProviders(ctx context.Context) ([]models.PaymentServiceProvider, error) {
	s.listAllCalls++
	if s.listAllFn != nil {
		return s.listAllFn(ctx)
	}
	return nil, nil
}

func (s *financeHandlerStub) ListPaymentServiceProvidersByType(ctx context.Context, institutionType string) ([]models.PaymentServiceProvider, error) {
	s.listByTypeCalls++
	s.lastType = institutionType
	if s.listByTypeFn != nil {
		return s.listByTypeFn(ctx, institutionType)
	}
	return nil, nil
}

func (s *financeHandlerStub) GetPaymentServiceProvider(ctx context.Context, providerID string) (models.PaymentServiceProvider, error) {
	s.getCalls++
	s.lastID = providerID
	if s.getFn != nil {
		return s.getFn(ctx, providerID)
	}
	return models.PaymentServiceProvider{}, nil
}

func TestNewFinanceHandlerRejectsNilService(t *testing.T) {
	if _, err := NewFinanceHandler(nil); err == nil {
		t.Fatal("NewFinanceHandler(nil) error = nil, want error")
	}
}

func TestFinanceHandlerListPaymentServiceProviders(t *testing.T) {
	t.Run("list all", func(t *testing.T) {
		stub := &financeHandlerStub{
			listAllFn: func(ctx context.Context) ([]models.PaymentServiceProvider, error) {
				return []models.PaymentServiceProvider{{
					ID:              "mobile-money-operator-abeg-technologies-limited",
					Name:            "Abeg Technologies Limited",
					InstitutionType: "mobile_money_operator",
					CountryCode:     "NG",
				}}, nil
			},
		}
		h, err := NewFinanceHandler(stub)
		if err != nil {
			t.Fatalf("NewFinanceHandler() error = %v", err)
		}

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListPaymentServiceProviders(w, r)
		}, rr)

		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.listAllCalls != 1 || stub.listByTypeCalls != 0 {
			t.Fatalf("unexpected call counts: all=%d type=%d", stub.listAllCalls, stub.listByTypeCalls)
		}

		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		data := body["data"].([]any)
		if len(data) != 1 {
			t.Fatalf("unexpected list payload: %#v", data)
		}
		item := data[0].(map[string]any)
		if len(item) != 4 {
			t.Fatalf("unexpected provider field count: %#v", item)
		}
		if item["id"] != "mobile-money-operator-abeg-technologies-limited" || item["country_code"] != "NG" {
			t.Fatalf("unexpected provider payload: %#v", item)
		}
	})

	t.Run("normalize empty slice", func(t *testing.T) {
		stub := &financeHandlerStub{listAllFn: func(context.Context) ([]models.PaymentServiceProvider, error) { return nil, nil }}
		h, err := NewFinanceHandler(stub)
		if err != nil {
			t.Fatalf("NewFinanceHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListPaymentServiceProviders(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		var body map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatalf("json.Unmarshal: %v", err)
		}
		data := body["data"].([]any)
		if len(data) != 0 {
			t.Fatalf("expected empty array, got %#v", data)
		}
	})

	t.Run("filter by type", func(t *testing.T) {
		stub := &financeHandlerStub{
			listByTypeFn: func(ctx context.Context, institutionType string) ([]models.PaymentServiceProvider, error) {
				if institutionType != "super_agent" {
					t.Fatalf("unexpected institution type: %q", institutionType)
				}
				return []models.PaymentServiceProvider{{
					ID:              "super-agent-fairmoney",
					Name:            "FairMoney Microfinance Bank Limited",
					InstitutionType: "super_agent",
					CountryCode:     "NG",
				}}, nil
			},
		}
		h, err := NewFinanceHandler(stub)
		if err != nil {
			t.Fatalf("NewFinanceHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers?institution_type=super_agent", nil)
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.ListPaymentServiceProviders(w, r)
		}, rr)
		if rr.Code != http.StatusOK {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.listAllCalls != 0 || stub.listByTypeCalls != 1 {
			t.Fatalf("unexpected call counts: all=%d type=%d", stub.listAllCalls, stub.listByTypeCalls)
		}
		if stub.lastType != "super_agent" {
			t.Fatalf("unexpected institution type: %q", stub.lastType)
		}
	})

	t.Run("reject invalid query", func(t *testing.T) {
		stub := &financeHandlerStub{}
		h, err := NewFinanceHandler(stub)
		if err != nil {
			t.Fatalf("NewFinanceHandler() error = %v", err)
		}
		for _, rawQuery := range []string{"institution_type=", "institution_type=Mobile%20Money%20Operators", "institution_type=mobile_money_operator&institution_type=super_agent"} {
			t.Run(rawQuery, func(t *testing.T) {
				rr := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers?"+rawQuery, nil)
				invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
					h.ListPaymentServiceProviders(w, r)
				}, rr)
				if rr.Code != http.StatusUnprocessableEntity {
					t.Fatalf("unexpected status: %d", rr.Code)
				}
				if stub.listAllCalls != 0 || stub.listByTypeCalls != 0 {
					t.Fatalf("service should not be called for invalid query")
				}
				if !strings.Contains(rr.Body.String(), "req_test") {
					t.Fatalf("request id missing from validation response: %s", rr.Body.String())
				}
			})
		}
	})

	t.Run("method guard", func(t *testing.T) {
		h, err := NewFinanceHandler(&financeHandlerStub{})
		if err != nil {
			t.Fatalf("NewFinanceHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/finance/payment-service-providers", nil)
		h.ListPaymentServiceProviders(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if got := rr.Header().Get("Allow"); got != http.MethodGet {
			t.Fatalf("unexpected allow header: %q", got)
		}
	})
}

func TestFinanceHandlerGetPaymentServiceProvider(t *testing.T) {
	stub := &financeHandlerStub{
		getFn: func(ctx context.Context, providerID string) (models.PaymentServiceProvider, error) {
			if providerID != "switching-and-processing-company-unified-payment-services-limited" {
				t.Fatalf("unexpected provider id: %q", providerID)
			}
			return models.PaymentServiceProvider{
				ID:              providerID,
				Name:            "Unified Payment Services Limited",
				InstitutionType: "switching_and_processing_company",
				CountryCode:     "NG",
			}, nil
		},
	}
	h, err := NewFinanceHandler(stub)
	if err != nil {
		t.Fatalf("NewFinanceHandler() error = %v", err)
	}

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers/switching-and-processing-company-unified-payment-services-limited", nil)
	req.SetPathValue("provider_id", " switching-and-processing-company-unified-payment-services-limited ")
	invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
		h.GetPaymentServiceProvider(w, r)
	}, rr)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if stub.getCalls != 1 {
		t.Fatalf("unexpected get call count: %d", stub.getCalls)
	}
	if stub.lastID != "switching-and-processing-company-unified-payment-services-limited" {
		t.Fatalf("unexpected provider id: %q", stub.lastID)
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	data := body["data"].(map[string]any)
	if len(data) != 4 {
		t.Fatalf("unexpected provider payload: %#v", data)
	}
	if data["name"] != "Unified Payment Services Limited" || data["institution_type"] != "switching_and_processing_company" {
		t.Fatalf("unexpected provider response: %#v", data)
	}
}

func TestFinanceHandlerGetPaymentServiceProviderErrors(t *testing.T) {
	t.Run("invalid id", func(t *testing.T) {
		stub := &financeHandlerStub{}
		h, err := NewFinanceHandler(stub)
		if err != nil {
			t.Fatalf("NewFinanceHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers/invalid", nil)
		req.SetPathValue("provider_id", "Paystack")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetPaymentServiceProvider(w, r)
		}, rr)
		if rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if stub.getCalls != 0 {
			t.Fatalf("service should not be called for invalid provider ids")
		}
	})

	t.Run("missing provider", func(t *testing.T) {
		stub := &financeHandlerStub{
			getFn: func(context.Context, string) (models.PaymentServiceProvider, error) {
				return models.PaymentServiceProvider{}, services.ErrPaymentServiceProviderNotFound
			},
		}
		h, err := NewFinanceHandler(stub)
		if err != nil {
			t.Fatalf("NewFinanceHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers/super-agent-missing", nil)
		req.SetPathValue("provider_id", "super-agent-missing")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetPaymentServiceProvider(w, r)
		}, rr)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
	})

	t.Run("wrapped missing provider", func(t *testing.T) {
		stub := &financeHandlerStub{
			getFn: func(context.Context, string) (models.PaymentServiceProvider, error) {
				return models.PaymentServiceProvider{}, fmtWrappedErr(services.ErrPaymentServiceProviderNotFound)
			},
		}
		h, err := NewFinanceHandler(stub)
		if err != nil {
			t.Fatalf("NewFinanceHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers/super-agent-missing", nil)
		req.SetPathValue("provider_id", "super-agent-missing")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetPaymentServiceProvider(w, r)
		}, rr)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		stub := &financeHandlerStub{
			getFn: func(context.Context, string) (models.PaymentServiceProvider, error) {
				return models.PaymentServiceProvider{}, errors.New("database down")
			},
		}
		h, err := NewFinanceHandler(stub)
		if err != nil {
			t.Fatalf("NewFinanceHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers/super-agent-fairmoney", nil)
		req.SetPathValue("provider_id", "super-agent-fairmoney")
		invokeWithRequestID(t, req, func(w http.ResponseWriter, r *http.Request) {
			h.GetPaymentServiceProvider(w, r)
		}, rr)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if strings.Contains(rr.Body.String(), "database down") {
			t.Fatalf("internal error details leaked: %s", rr.Body.String())
		}
	})

	t.Run("method guard", func(t *testing.T) {
		h, err := NewFinanceHandler(&financeHandlerStub{})
		if err != nil {
			t.Fatalf("NewFinanceHandler() error = %v", err)
		}
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/finance/payment-service-providers/super-agent-fairmoney", nil)
		req.SetPathValue("provider_id", "super-agent-fairmoney")
		h.GetPaymentServiceProvider(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Fatalf("unexpected status: %d", rr.Code)
		}
		if got := rr.Header().Get("Allow"); got != http.MethodGet {
			t.Fatalf("unexpected allow header: %q", got)
		}
	})
}
