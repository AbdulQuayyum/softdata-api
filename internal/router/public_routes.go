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

	return nil
}
