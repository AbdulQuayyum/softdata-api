package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublicRoutesServeCompleteEducationContract(t *testing.T) {
	tests := []struct {
		name           string
		listPath       string
		detailPath     string
		detailTemplate string
	}{
		{"polytechnics", "/v1/education/polytechnics", "/v1/education/polytechnics/sample-polytechnic", "/v1/education/polytechnics/{institution_id}"},
		{"monotechnics", "/v1/education/monotechnics", "/v1/education/monotechnics/sample-monotechnic", "/v1/education/monotechnics/{institution_id}"},
		{"agriculture", "/v1/education/colleges-of-agriculture", "/v1/education/colleges-of-agriculture/sample-agriculture", "/v1/education/colleges-of-agriculture/{institution_id}"},
		{"health", "/v1/education/colleges-of-health-sciences-and-technology", "/v1/education/colleges-of-health-sciences-and-technology/sample-health-college", "/v1/education/colleges-of-health-sciences-and-technology/{institution_id}"},
		{"nursing", "/v1/education/colleges-of-nursing-and-midwifery", "/v1/education/colleges-of-nursing-and-midwifery/sample-nursing-college", "/v1/education/colleges-of-nursing-and-midwifery/{institution_id}"},
		{"vei", "/v1/education/vocational-enterprise-institutions", "/v1/education/vocational-enterprise-institutions/sample-vei", "/v1/education/vocational-enterprise-institutions/{institution_id}"},
		{"technical", "/v1/education/technical-colleges", "/v1/education/technical-colleges/sample-technical-college", "/v1/education/technical-colleges/{institution_id}"},
		{"schools", "/v1/education/primary-and-secondary-schools", "/v1/education/primary-and-secondary-schools/sample-school", "/v1/education/primary-and-secondary-schools/{school_id}"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := &routerRecorder{}
			r := newTestRouter(t, rec)
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, tc.listPath+"?page_size=100", nil))
			if rr.Code != http.StatusOK || !strings.Contains(rr.Header().Get("Content-Type"), "application/json") {
				t.Fatalf("list response: %d %s", rr.Code, rr.Body.String())
			}
			if !strings.Contains(strings.Join(rec.snapshot(), ","), "usage:"+tc.listPath+"|education") {
				t.Fatalf("list usage template missing: %v", rec.snapshot())
			}
			if strings.Contains(strings.Join(rec.snapshot(), ","), "sample-") {
				t.Fatalf("list usage should not contain an ID: %v", rec.snapshot())
			}

			rec = &routerRecorder{}
			r = newTestRouter(t, rec)
			rr = httptest.NewRecorder()
			r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, tc.detailPath, nil))
			if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"success":true`) {
				t.Fatalf("detail response: %d %s", rr.Code, rr.Body.String())
			}
			if !strings.Contains(strings.Join(rec.snapshot(), ","), "usage:"+tc.detailTemplate+"|education") {
				t.Fatalf("detail usage template missing: %v", rec.snapshot())
			}
		})
	}
}

func TestPublicRoutesEducationMethodProtectionAndPaginationSafety(t *testing.T) {
	r := newTestRouter(t, &routerRecorder{})
	for _, tc := range []struct{ method, path string }{
		{http.MethodPost, "/v1/education/polytechnics"},
		{http.MethodPut, "/v1/education/polytechnics/sample-polytechnic"},
		{http.MethodPatch, "/v1/education/technical-colleges"},
		{http.MethodDelete, "/v1/education/primary-and-secondary-schools/sample-school"},
	} {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, httptest.NewRequest(tc.method, tc.path, nil))
		if rr.Code != http.StatusMethodNotAllowed || rr.Header().Get("Allow") != http.MethodGet || !strings.Contains(rr.Header().Get("Content-Type"), "application/json") {
			t.Fatalf("%s %s: status=%d allow=%q content-type=%q body=%s", tc.method, tc.path, rr.Code, rr.Header().Get("Allow"), rr.Header().Get("Content-Type"), rr.Body.String())
		}
		if strings.Contains(rr.Body.String(), "Method Not Allowed") || strings.Contains(rr.Body.String(), "405 method") {
			t.Fatalf("plain method error: %s", rr.Body.String())
		}
	}

	for _, path := range []string{
		"/v1/education/polytechnics/one/two",
		"/v1/education/primary-and-secondary-schools/one/two",
	} {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, path, nil))
		if rr.Code != http.StatusNotFound {
			t.Fatalf("extra path %s status=%d", path, rr.Code)
		}
	}

	for _, query := range []string{"", "?page_size=100"} {
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/education/primary-and-secondary-schools"+query, nil))
		var body struct {
			Data []json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		limit := 50
		if query != "" {
			limit = 100
		}
		if len(body.Data) > limit {
			t.Fatalf("query %q returned %d records", query, len(body.Data))
		}
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/education/primary-and-secondary-schools?page_size=101", nil))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("page_size=101 status=%d", rr.Code)
	}
}

func TestPublicRoutesServeHealthAndDiscovery(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type: %q", got)
	}
	if !strings.Contains(rr.Body.String(), `"status":"ok"`) {
		t.Fatalf("unexpected health response: %s", rr.Body.String())
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"success":true`) {
		t.Fatalf("unexpected discovery response: %s", rr.Body.String())
	}
}

func TestPublicRoutesServeCommercialBanks(t *testing.T) {
	rec := &routerRecorder{}
	r := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/finance/commercial-banks?ignored=true", nil))
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"success":true`) {
		t.Fatalf("unexpected commercial-bank list response: %d %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "finance.commercial-banks.list") {
		t.Fatalf("commercial-bank list was not dispatched: %v", rec.snapshot())
	}

	rec = &routerRecorder{}
	r = newTestRouter(t, rec)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/finance/commercial-banks/access-bank", nil))
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"cbn_code":"044"`) {
		t.Fatalf("unexpected commercial-bank detail response: %d %s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "finance.commercial-banks.get:access-bank") {
		t.Fatalf("commercial-bank detail was not dispatched: %v", rec.snapshot())
	}
}

func TestPublicRoutesServeMicrofinanceBanks(t *testing.T) {
	rec := &routerRecorder{}
	r := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/finance/microfinance-banks?ignored=true", nil))
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"success":true`) {
		t.Fatalf("unexpected microfinance-bank list response: %d %s", rr.Code, rr.Body.String())
	}
	joined := strings.Join(rec.snapshot(), ",")
	if !strings.Contains(joined, "finance.microfinance-banks.list") ||
		!strings.Contains(joined, "usage:/v1/finance/microfinance-banks|finance") {
		t.Fatalf("microfinance-bank list dispatch/usage missing: %v", rec.snapshot())
	}

	rec = &routerRecorder{}
	r = newTestRouter(t, rec)
	rr = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/finance/microfinance-banks/bway-microfinance-bank-limited", nil)
	r.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !strings.Contains(rr.Body.String(), `"country_code":"NG"`) {
		t.Fatalf("unexpected microfinance-bank detail response: %d %s", rr.Code, rr.Body.String())
	}
	joined = strings.Join(rec.snapshot(), ",")
	if !strings.Contains(joined, "finance.microfinance-banks.get:bway-microfinance-bank-limited") ||
		!strings.Contains(joined, "usage:/v1/finance/microfinance-banks/{bank_id}|finance") {
		t.Fatalf("microfinance-bank detail dispatch/usage missing: %v", rec.snapshot())
	}

	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/v1/finance/microfinance-banks", nil))
	if rr.Code != http.StatusMethodNotAllowed || rr.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("unexpected method response: %d allow=%q", rr.Code, rr.Header().Get("Allow"))
	}

	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/v1/finance/microfinance-banks/bway-microfinance-bank-limited/extra", nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("nested microfinance-bank route status = %d, want 404", rr.Code)
	}
}

func TestPublicRoutesServeGeographyZones(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "geography.zone.list") {
		t.Fatalf("expected zone list handler to run: %v", rec.snapshot())
	}

	rec = &routerRecorder{}
	router = newTestRouter(t, rec)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/geography/geopolitical-zones/north-central", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "geography.zone.get:north-central") {
		t.Fatalf("expected zone detail handler to run: %v", rec.snapshot())
	}
}

func TestPublicRoutesServeGeographyLanguages(t *testing.T) {
	rec := &routerRecorder{}
	geography := &routerGeographyStub{rec: rec}
	router, err := New(testHandlersWithGeography(t, rec, geography), testMiddleware(rec))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	for _, tc := range []struct {
		path string
		want string
	}{
		{path: "/v1/geography/languages", want: "geography.language.list"},
		{path: "/v1/geography/languages/en", want: "geography.language.get:en"},
		{path: "/v1/geography/country-languages?country_area_id=ng&language_id=yo&status=official", want: "geography.country-language.list"},
	} {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		if tc.want == "geography.language.get:en" {
			req.SetPathValue("language_id", "en")
		}
		router.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("%s status = %d, body=%s", tc.path, rr.Code, rr.Body.String())
		}
		if !strings.Contains(strings.Join(rec.snapshot(), ","), tc.want) {
			t.Fatalf("expected %q in route calls: %v", tc.want, rec.snapshot())
		}
	}
	if got := geography.lastCountryLanguageInput; got.CountryAreaID != "ng" || got.LanguageID != "yo" || got.Status != "official" {
		t.Fatalf("unexpected country-language filter: %#v", got)
	}
	if !geography.lastHadAPIKey {
		t.Fatal("expected optional API-key identification middleware")
	}
}

func TestPublicRoutesServeGeographyLGAs(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/lgas", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "geography.lga.list") {
		t.Fatalf("expected lga list handler to run: %v", rec.snapshot())
	}

	rec = &routerRecorder{}
	router = newTestRouter(t, rec)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/geography/lgas/lagos-ikeja", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "geography.lga.get:lagos-ikeja") {
		t.Fatalf("expected lga detail handler to run: %v", rec.snapshot())
	}

	rec = &routerRecorder{}
	router = newTestRouter(t, rec)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/geography/lgas?state_id=fct", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "geography.lga.list-by:fct") {
		t.Fatalf("expected lga list-by-state handler to run: %v", rec.snapshot())
	}
}

func TestPublicRoutesServeGeographyCountriesAndAreas(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "usage:/v1/geography/countries|geography") {
		t.Fatalf("expected country list usage middleware to run: %v", rec.snapshot())
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "geography.country.list") {
		t.Fatalf("expected country list handler to run: %v", rec.snapshot())
	}

	rec = &routerRecorder{}
	router = newTestRouter(t, rec)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/geography/countries/ng", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "usage:/v1/geography/countries/{country_id}|geography") {
		t.Fatalf("expected country detail usage middleware to run: %v", rec.snapshot())
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "geography.country.get:ng") {
		t.Fatalf("expected country detail handler to run: %v", rec.snapshot())
	}
}

func TestPublicRoutesServeGeographyCountryProfiles(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/geography/countries/ng/profile?state_id=lagos", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "usage:/v1/geography/countries/{country_id}/profile|geography") {
		t.Fatalf("expected country profile usage middleware to run: %v", rec.snapshot())
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "geography.country.profile:ng") {
		t.Fatalf("expected country profile handler to run: %v", rec.snapshot())
	}
}

func TestPublicRoutesServeCountryFlagAssets(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/assets/flags/ng.svg", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "image/svg+xml" {
		t.Fatalf("unexpected content type: %q", got)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "usage:/v1/assets/flags/{country_id}.svg|geography") {
		t.Fatalf("expected flag asset usage middleware to run: %v", rec.snapshot())
	}
}

func TestPublicRoutesServeEducationUniversities(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/education/universities", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "education.list") {
		t.Fatalf("expected education list handler to run: %v", rec.snapshot())
	}

	rec = &routerRecorder{}
	router = newTestRouter(t, rec)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/education/universities/ahmadu-bello-university-zaria", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "education.get:ahmadu-bello-university-zaria") {
		t.Fatalf("expected education detail handler to run: %v", rec.snapshot())
	}
}

func TestPublicRoutesServeEducationCollegesOfEducation(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/education/colleges-of-education", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "education.college.list") {
		t.Fatalf("expected college list handler to run: %v", rec.snapshot())
	}

	rec = &routerRecorder{}
	router = newTestRouter(t, rec)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/education/colleges-of-education/federal-college-of-education-zaria", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "education.college.get:federal-college-of-education-zaria") {
		t.Fatalf("expected college detail handler to run: %v", rec.snapshot())
	}

	rec = &routerRecorder{}
	router = newTestRouter(t, rec)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/education/colleges-of-education?state_id=lagos&ownership_type=private", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "education.college.list") {
		t.Fatalf("expected college filtered handler to run: %v", rec.snapshot())
	}
}

func TestPublicRoutesServeFinancePaymentServiceProviders(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "finance.list") {
		t.Fatalf("expected finance list handler to run: %v", rec.snapshot())
	}

	rec = &routerRecorder{}
	router = newTestRouter(t, rec)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers?institution_type=super_agent", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "finance.list-by:super_agent") {
		t.Fatalf("expected finance filtered handler to run: %v", rec.snapshot())
	}

	rec = &routerRecorder{}
	router = newTestRouter(t, rec)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/finance/payment-service-providers/super-agent-fairmoney", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "finance.get:super-agent-fairmoney") {
		t.Fatalf("expected finance detail handler to run: %v", rec.snapshot())
	}
}

func TestPublicRoutesServeFinanceInternationalMoneyTransferOperators(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v1/finance/international-money-transfer-operators", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "finance.imto.list") {
		t.Fatalf("expected imto list handler to run: %v", rec.snapshot())
	}

	rec = &routerRecorder{}
	router = newTestRouter(t, rec)

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/v1/finance/international-money-transfer-operators/olive-monies-express-limited", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if !strings.Contains(strings.Join(rec.snapshot(), ","), "finance.imto.get:olive-monies-express-limited") {
		t.Fatalf("expected imto detail handler to run: %v", rec.snapshot())
	}
}

func TestPublicRoutesRejectUnsupportedMethods(t *testing.T) {
	rec := &routerRecorder{}
	router := newTestRouter(t, rec)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Allow"); got == "" {
		t.Fatal("expected Allow header")
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodHead, "/v1/geography/geopolitical-zones/north-central", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("unexpected allow header: %q", got)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodHead, "/v1/geography/countries/ng", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("unexpected allow header: %q", got)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodHead, "/v1/finance/payment-service-providers/super-agent-fairmoney", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("unexpected allow header: %q", got)
	}

	rr = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodHead, "/v1/assets/flags/ng.svg", nil)
	router.ServeHTTP(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("unexpected status: %d", rr.Code)
	}
	if got := rr.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("unexpected allow header: %q", got)
	}
}
