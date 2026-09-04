package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
