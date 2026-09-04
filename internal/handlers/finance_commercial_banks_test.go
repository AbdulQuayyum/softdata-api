package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AbdulQuayyum/softdata-api/internal/models"
)

func TestFinanceHandlerCommercialBankEndpoints(t *testing.T) {
	stub := &financeHandlerStub{
		listBanksFn: func(context.Context) ([]models.CommercialBank, error) { return nil, nil },
		getBankFn: func(_ context.Context, id string) (models.CommercialBank, error) {
			if id == "nova-bank" {
				return models.CommercialBank{ID: id, Name: "Nova Commercial Bank Limited", OfficialWebsiteURL: "https://www.novabank.ng/", LogoURL: "/v1/assets/banks/ng/nova-bank.png", CountryCode: "NG"}, nil
			}
			return models.CommercialBank{ID: id, Name: "Access Bank Plc", CBNCode: "044", NIPCode: "000014", OfficialWebsiteURL: "https://www.accessbankplc.com/", LogoURL: "/v1/assets/banks/ng/access-bank.png", CountryCode: "NG"}, nil
		},
	}
	h, err := NewFinanceHandler(stub)
	if err != nil {
		t.Fatal(err)
	}

	listResponse := httptest.NewRecorder()
	h.ListCommercialBanks(listResponse, httptest.NewRequest(http.MethodGet, "/v1/finance/commercial-banks", nil))
	if listResponse.Code != http.StatusOK || !strings.Contains(listResponse.Body.String(), `"data":[]`) {
		t.Fatalf("unexpected list response: %d %s", listResponse.Code, listResponse.Body.String())
	}

	detailRequest := httptest.NewRequest(http.MethodGet, "/v1/finance/commercial-banks/nova-bank", nil)
	detailRequest.SetPathValue("bank_id", "nova-bank")
	detailResponse := httptest.NewRecorder()
	h.GetCommercialBank(detailResponse, detailRequest)
	if detailResponse.Code != http.StatusOK || strings.Contains(detailResponse.Body.String(), "cbn_code") || strings.Contains(detailResponse.Body.String(), "nip_code") {
		t.Fatalf("unexpected NOVA response: %d %s", detailResponse.Code, detailResponse.Body.String())
	}

	invalidRequest := httptest.NewRequest(http.MethodGet, "/v1/finance/commercial-banks/bad", nil)
	invalidRequest.SetPathValue("bank_id", "bad")
	invalidResponse := httptest.NewRecorder()
	h.GetCommercialBank(invalidResponse, invalidRequest)
	if invalidResponse.Code != http.StatusUnprocessableEntity || stub.getBankCalls != 1 {
		t.Fatalf("invalid ID response/calls = %d/%d", invalidResponse.Code, stub.getBankCalls)
	}

	methodResponse := httptest.NewRecorder()
	h.ListCommercialBanks(methodResponse, httptest.NewRequest(http.MethodPost, "/v1/finance/commercial-banks", nil))
	if methodResponse.Code != http.StatusMethodNotAllowed || methodResponse.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("unexpected method response: %d allow=%q", methodResponse.Code, methodResponse.Header().Get("Allow"))
	}

	var envelope map[string]any
	if err := json.Unmarshal(detailResponse.Body.Bytes(), &envelope); err != nil || envelope["success"] != true {
		t.Fatalf("invalid detail envelope: %v %#v", err, envelope)
	}
}
