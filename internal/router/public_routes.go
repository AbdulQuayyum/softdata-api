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

	return nil
}
