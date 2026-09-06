package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
	"github.com/AbdulQuayyum/softdata-api/internal/services"
)

func TestFinanceHandlerMicrofinanceBankEndpoints(t *testing.T) {
	stub := &financeHandlerStub{
		listMicrofinanceBanksFn: func(context.Context) ([]models.MicrofinanceBank, error) { return nil, nil },
		getMicrofinanceBankFn: func(_ context.Context, id string) (models.MicrofinanceBank, error) {
			return models.MicrofinanceBank{ID: id, Name: "Example Microfinance Bank", CountryCode: "NG"}, nil
		},
	}
	h, err := NewFinanceHandler(stub)
	if err != nil {
		t.Fatal(err)
	}

	listResponse := httptest.NewRecorder()
	h.ListMicrofinanceBanks(listResponse, httptest.NewRequest(http.MethodGet, "/v1/finance/microfinance-banks", nil))
	if listResponse.Code != http.StatusOK || listResponse.Body.String() != `{"success":true,"data":[]}
` {
		t.Fatalf("unexpected list response: %d %s", listResponse.Code, listResponse.Body.String())
	}

	detailRequest := httptest.NewRequest(http.MethodGet, "/v1/finance/microfinance-banks/bway-microfinance-bank-limited", nil)
	detailRequest.SetPathValue("bank_id", "bway-microfinance-bank-limited")
	detailResponse := httptest.NewRecorder()
	h.GetMicrofinanceBank(detailResponse, detailRequest)
	if detailResponse.Code != http.StatusOK {
		t.Fatalf("unexpected detail response: %d %s", detailResponse.Code, detailResponse.Body.String())
	}
	var detail map[string]any
	if err := json.Unmarshal(detailResponse.Body.Bytes(), &detail); err != nil || detail["success"] != true {
		t.Fatalf("invalid detail envelope: %v %#v", err, detail)
	}

	invalidRequest := httptest.NewRequest(http.MethodGet, "/v1/finance/microfinance-banks/BWAY", nil)
	invalidRequest.SetPathValue("bank_id", "BWAY")
	invalidResponse := httptest.NewRecorder()
	h.GetMicrofinanceBank(invalidResponse, invalidRequest)
	if invalidResponse.Code != http.StatusUnprocessableEntity || stub.getMicrofinanceCalls != 1 {
		t.Fatalf("invalid ID response/calls = %d/%d", invalidResponse.Code, stub.getMicrofinanceCalls)
	}

	notFoundStub := &financeHandlerStub{getMicrofinanceBankFn: func(context.Context, string) (models.MicrofinanceBank, error) {
		return models.MicrofinanceBank{}, services.ErrMicrofinanceBankNotFound
	}}
	notFoundHandler, err := NewFinanceHandler(notFoundStub)
	if err != nil {
		t.Fatal(err)
	}
	notFoundRequest := httptest.NewRequest(http.MethodGet, "/v1/finance/microfinance-banks/missing", nil)
	notFoundRequest.SetPathValue("bank_id", "missing")
	notFoundResponse := httptest.NewRecorder()
	notFoundHandler.GetMicrofinanceBank(notFoundResponse, notFoundRequest)
	if notFoundResponse.Code != http.StatusNotFound {
		t.Fatalf("not-found response = %d", notFoundResponse.Code)
	}

	methodResponse := httptest.NewRecorder()
	h.ListMicrofinanceBanks(methodResponse, httptest.NewRequest(http.MethodPost, "/v1/finance/microfinance-banks", nil))
	if methodResponse.Code != http.StatusMethodNotAllowed || methodResponse.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("unexpected method response: %d allow=%q", methodResponse.Code, methodResponse.Header().Get("Allow"))
	}

}
