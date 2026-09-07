package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicRoutesServeRegulatedFinanceLists(t *testing.T) {
	router := newTestRouter(t, &routerRecorder{})
	paths := []string{
		"/v1/finance/non-interest-financial-institutions",
		"/v1/finance/merchant-banks",
		"/v1/finance/payment-service-banks",
		"/v1/finance/financial-holding-companies",
		"/v1/finance/development-finance-institutions",
		"/v1/finance/primary-mortgage-institutions",
	}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
			if rr.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d: %s", rr.Code, http.StatusOK, rr.Body.String())
			}
			if got := rr.Header().Get("Content-Type"); got != "application/json" {
				t.Fatalf("content type = %q", got)
			}
			if !strings.Contains(rr.Body.String(), `"success":true`) {
				t.Fatalf("unexpected response: %s", rr.Body.String())
			}
		})
	}
}

func TestPublicRoutesServeRegulatedFinanceLogo(t *testing.T) {
	router := newTestRouter(t, &routerRecorder{})
	for _, path := range []string{
		"/v1/assets/financial-institutions/ng/non-interest/mint-microfinance-bank.png",
		"/v1/assets/financial-institutions/ng/payment-service-providers/mobile-money-operator-chams-mobile.png",
	} {
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusOK || rr.Header().Get("Content-Type") != "image/png" || len(rr.Body.Bytes()) == 0 {
			t.Fatalf("unexpected logo response for %s: status=%d content-type=%q bytes=%d", path, rr.Code, rr.Header().Get("Content-Type"), len(rr.Body.Bytes()))
		}
		if rr.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
			t.Fatalf("unexpected cache policy for %s: %q", path, rr.Header().Get("Cache-Control"))
		}
	}
}

func TestPublicRoutesRejectRegulatedFinanceLogoWithoutAsset(t *testing.T) {
	router := newTestRouter(t, &routerRecorder{})
	for _, path := range []string{
		"/v1/assets/financial-institutions/ng/primary-mortgage/missing-logo.png",
		"/v1/assets/financial-institutions/ng/unknown/access-bank.png",
		"/v1/assets/financial-institutions/ng/non-interest/../access-bank.png",
	} {
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusNotFound && rr.Code != http.StatusUnprocessableEntity {
			t.Fatalf("path %q status = %d", path, rr.Code)
		}
	}
}

func TestPublicRoutesRejectRegulatedFinanceLogoWithJSON405(t *testing.T) {
	router := newTestRouter(t, &routerRecorder{})
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/assets/financial-institutions/ng/non-interest/mint-microfinance-bank.png", nil))
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusMethodNotAllowed)
	}
	if got := rr.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("Allow = %q, want %q", got, http.MethodGet)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content type = %q, want application/json", got)
	}
	if !strings.Contains(rr.Body.String(), `"code":"INVALID_REQUEST"`) {
		t.Fatalf("unexpected response: %s", rr.Body.String())
	}
}
