package router

import (
	"fmt"
	"net/http"
)

func registerPublicRoutes(mux *http.ServeMux, catalog *routeCatalog, h Handlers, mw Middleware) error {
	health, err := buildRouteMiddlewares(mw, "/health", "", routeOptions{
		useRateLimit: true,
	})
	if err != nil {
		return fmt.Errorf("build health middleware: %w", err)
	}
	mux.Handle("GET /health", compose(http.HandlerFunc(h.Health.ServeHTTP), health...))
	if err := catalog.add("GET /health"); err != nil {
		return err
	}

	discovery, err := buildRouteMiddlewares(mw, "/v1", "", routeOptions{
		useRateLimit: true,
	})
	if err != nil {
		return fmt.Errorf("build discovery middleware: %w", err)
	}
	mux.Handle("GET /v1", compose(http.HandlerFunc(h.Discovery.ServeHTTP), discovery...))
	if err := catalog.add("GET /v1"); err != nil {
		return err
	}

	geographyList, err := buildRouteMiddlewares(mw, "/v1/geography/states", "geography", routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	})
	if err != nil {
		return fmt.Errorf("build geography states middleware: %w", err)
	}
	mux.Handle("GET /v1/geography/states", compose(http.HandlerFunc(h.Geography.ListStates), geographyList...))
	if err := catalog.add("GET /v1/geography/states"); err != nil {
		return err
	}

	geographyDetail, err := buildRouteMiddlewares(mw, "/v1/geography/states/{state_id}", "geography", routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	})
	if err != nil {
		return fmt.Errorf("build geography state middleware: %w", err)
	}
	mux.Handle("GET /v1/geography/states/{state_id}", compose(http.HandlerFunc(h.Geography.GetState), geographyDetail...))
	if err := catalog.add("GET /v1/geography/states/{state_id}"); err != nil {
		return err
	}

	geographyZonesList, err := buildRouteMiddlewares(mw, "/v1/geography/geopolitical-zones", "geography", routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	})
	if err != nil {
		return fmt.Errorf("build geography geopolitical zones middleware: %w", err)
	}
	mux.Handle("GET /v1/geography/geopolitical-zones", compose(http.HandlerFunc(h.Geography.ListGeopoliticalZones), geographyZonesList...))
	if err := catalog.add("GET /v1/geography/geopolitical-zones"); err != nil {
		return err
	}

	geographyZoneDetail, err := buildRouteMiddlewares(mw, "/v1/geography/geopolitical-zones/{zone_id}", "geography", routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	})
	if err != nil {
		return fmt.Errorf("build geography geopolitical zone middleware: %w", err)
	}
	mux.Handle("GET /v1/geography/geopolitical-zones/{zone_id}", compose(http.HandlerFunc(h.Geography.GetGeopoliticalZone), geographyZoneDetail...))
	if err := catalog.add("GET /v1/geography/geopolitical-zones/{zone_id}"); err != nil {
		return err
	}

	geographyLGAsList, err := buildRouteMiddlewares(mw, "/v1/geography/lgas", "geography", routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	})
	if err != nil {
		return fmt.Errorf("build geography lgas middleware: %w", err)
	}
	mux.Handle("GET /v1/geography/lgas", compose(http.HandlerFunc(h.Geography.ListLocalGovernmentUnits), geographyLGAsList...))
	if err := catalog.add("GET /v1/geography/lgas"); err != nil {
		return err
	}

	geographyLGADetail, err := buildRouteMiddlewares(mw, "/v1/geography/lgas/{lga_id}", "geography", routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	})
	if err != nil {
		return fmt.Errorf("build geography lgas detail middleware: %w", err)
	}
	mux.Handle("GET /v1/geography/lgas/{lga_id}", compose(http.HandlerFunc(h.Geography.GetLocalGovernmentUnit), geographyLGADetail...))
	if err := catalog.add("GET /v1/geography/lgas/{lga_id}"); err != nil {
		return err
	}

	educationList, err := buildRouteMiddlewares(mw, "/v1/education/universities", "education", routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	})
	if err != nil {
		return fmt.Errorf("build education universities middleware: %w", err)
	}
	mux.Handle("GET /v1/education/universities", compose(http.HandlerFunc(h.Education.ListUniversities), educationList...))
	if err := catalog.add("GET /v1/education/universities"); err != nil {
		return err
	}

	educationDetail, err := buildRouteMiddlewares(mw, "/v1/education/universities/{university_id}", "education", routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	})
	if err != nil {
		return fmt.Errorf("build education university detail middleware: %w", err)
	}
	mux.Handle("GET /v1/education/universities/{university_id}", compose(http.HandlerFunc(h.Education.GetUniversity), educationDetail...))
	if err := catalog.add("GET /v1/education/universities/{university_id}"); err != nil {
		return err
	}

	financeList, err := buildRouteMiddlewares(mw, "/v1/finance/payment-service-providers", "finance", routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	})
	if err != nil {
		return fmt.Errorf("build finance payment-service-providers middleware: %w", err)
	}
	mux.Handle("GET /v1/finance/payment-service-providers", compose(http.HandlerFunc(h.Finance.ListPaymentServiceProviders), financeList...))
	if err := catalog.add("GET /v1/finance/payment-service-providers"); err != nil {
		return err
	}

	financeDetail, err := buildRouteMiddlewares(mw, "/v1/finance/payment-service-providers/{provider_id}", "finance", routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	})
	if err != nil {
		return fmt.Errorf("build finance payment-service-providers detail middleware: %w", err)
	}
	mux.Handle("GET /v1/finance/payment-service-providers/{provider_id}", compose(http.HandlerFunc(h.Finance.GetPaymentServiceProvider), financeDetail...))
	if err := catalog.add("GET /v1/finance/payment-service-providers/{provider_id}"); err != nil {
		return err
	}

	financeIMTOsList, err := buildRouteMiddlewares(mw, "/v1/finance/international-money-transfer-operators", "finance", routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	})
	if err != nil {
		return fmt.Errorf("build finance international-money-transfer-operators middleware: %w", err)
	}
	mux.Handle("GET /v1/finance/international-money-transfer-operators", compose(http.HandlerFunc(h.Finance.ListInternationalMoneyTransferOperators), financeIMTOsList...))
	if err := catalog.add("GET /v1/finance/international-money-transfer-operators"); err != nil {
		return err
	}

	financeIMTODetail, err := buildRouteMiddlewares(mw, "/v1/finance/international-money-transfer-operators/{operator_id}", "finance", routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	})
	if err != nil {
		return fmt.Errorf("build finance international-money-transfer-operators detail middleware: %w", err)
	}
	mux.Handle("GET /v1/finance/international-money-transfer-operators/{operator_id}", compose(http.HandlerFunc(h.Finance.GetInternationalMoneyTransferOperator), financeIMTODetail...))
	if err := catalog.add("GET /v1/finance/international-money-transfer-operators/{operator_id}"); err != nil {
		return err
	}

	return nil
}
