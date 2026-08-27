package router

import (
	"fmt"
	"net/http"
)

func registerAuthRoutes(mux *http.ServeMux, catalog *routeCatalog, h Handlers, mw Middleware) error {
	register := func(pattern, endpoint string, handler http.Handler, opts routeOptions) error {
		chain, err := buildRouteMiddlewares(mw, endpoint, "", opts)
		if err != nil {
			return err
		}
		mux.Handle(pattern, compose(handler, chain...))
		if err := catalog.add(pattern); err != nil {
			return err
		}
		return nil
	}

	if err := register("POST /v1/auth/register", "/v1/auth/register", http.HandlerFunc(h.Auth.Register), routeOptions{
		useRateLimit:     true,
		useUsageTracking: true,
	}); err != nil {
		return fmt.Errorf("register auth register route: %w", err)
	}
	if err := register("POST /v1/auth/login", "/v1/auth/login", http.HandlerFunc(h.Auth.Login), routeOptions{
		useRateLimit:     true,
		useUsageTracking: true,
	}); err != nil {
		return fmt.Errorf("register auth login route: %w", err)
	}
	if err := register("POST /v1/auth/refresh", "/v1/auth/refresh", http.HandlerFunc(h.Auth.Refresh), routeOptions{
		useRateLimit:     true,
		useUsageTracking: true,
	}); err != nil {
		return fmt.Errorf("register auth refresh route: %w", err)
	}
	if err := register("POST /v1/auth/logout", "/v1/auth/logout", http.HandlerFunc(h.Auth.Logout), routeOptions{
		useAuthentication: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	}); err != nil {
		return fmt.Errorf("register auth logout route: %w", err)
	}

	return nil
}
