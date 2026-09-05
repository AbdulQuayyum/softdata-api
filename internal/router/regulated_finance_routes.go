package router

import (
	"fmt"
	"net/http"
)

func registerRegulatedFinanceRoutes(mux *http.ServeMux, catalog *routeCatalog, h Handlers, mw Middleware) error {
	routes := []struct {
		listPattern   string
		detailPattern string
		listHandler   http.Handler
		detailHandler http.Handler
	}{
		{"/v1/finance/non-interest-financial-institutions", "/v1/finance/non-interest-financial-institutions/{institution_id}", http.HandlerFunc(h.Finance.ListNonInterestFinancialInstitutions), http.HandlerFunc(h.Finance.GetNonInterestFinancialInstitution)},
		{"/v1/finance/merchant-banks", "/v1/finance/merchant-banks/{bank_id}", http.HandlerFunc(h.Finance.ListMerchantBanks), http.HandlerFunc(h.Finance.GetMerchantBank)},
		{"/v1/finance/payment-service-banks", "/v1/finance/payment-service-banks/{bank_id}", http.HandlerFunc(h.Finance.ListPaymentServiceBanks), http.HandlerFunc(h.Finance.GetPaymentServiceBank)},
		{"/v1/finance/financial-holding-companies", "/v1/finance/financial-holding-companies/{company_id}", http.HandlerFunc(h.Finance.ListFinancialHoldingCompanies), http.HandlerFunc(h.Finance.GetFinancialHoldingCompany)},
		{"/v1/finance/development-finance-institutions", "/v1/finance/development-finance-institutions/{institution_id}", http.HandlerFunc(h.Finance.ListDevelopmentFinanceInstitutions), http.HandlerFunc(h.Finance.GetDevelopmentFinanceInstitution)},
		{"/v1/finance/primary-mortgage-institutions", "/v1/finance/primary-mortgage-institutions/{institution_id}", http.HandlerFunc(h.Finance.ListPrimaryMortgageInstitutions), http.HandlerFunc(h.Finance.GetPrimaryMortgageInstitution)},
	}
	for _, route := range routes {
		if err := registerFinanceRoute(mux, catalog, mw, route.listPattern, route.listHandler); err != nil {
			return err
		}
		if err := registerFinanceRoute(mux, catalog, mw, route.detailPattern, route.detailHandler); err != nil {
			return err
		}
	}

	assetPattern := "/v1/assets/financial-institutions/ng/{category}/{institution_asset}"
	assetMiddleware, err := buildRouteMiddlewares(mw, assetPattern, "finance", routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	})
	if err != nil {
		return fmt.Errorf("build finance regulated-institution logo middleware: %w", err)
	}
	// Register the path without a method so the handler can return the API's
	// JSON 405 response instead of ServeMux's plain-text method response.
	mux.Handle(assetPattern, compose(http.HandlerFunc(serveFinancialInstitutionLogo), assetMiddleware...))
	if err := catalog.add("GET " + assetPattern); err != nil {
		return err
	}
	return nil
}

func registerFinanceRoute(mux *http.ServeMux, catalog *routeCatalog, mw Middleware, pattern string, handler http.Handler) error {
	middleware, err := buildRouteMiddlewares(mw, pattern, "finance", routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	})
	if err != nil {
		return fmt.Errorf("build finance route middleware for %s: %w", pattern, err)
	}
	mux.Handle("GET "+pattern, compose(handler, middleware...))
	if err := catalog.add("GET " + pattern); err != nil {
		return err
	}
	return nil
}
