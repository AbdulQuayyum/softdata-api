package router

import (
	"fmt"
	"net/http"
)

func registerAccountRoutes(mux *http.ServeMux, catalog *routeCatalog, h Handlers, mw Middleware) error {
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

	accountOpts := routeOptions{
		useOptionalAPIKey: true,
		useAuthentication: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	}

	if err := register("GET /v1/account", "/v1/account", http.HandlerFunc(h.Account.GetCurrent), accountOpts); err != nil {
		return fmt.Errorf("register account get route: %w", err)
	}
	if err := register("PATCH /v1/account", "/v1/account", http.HandlerFunc(h.Account.UpdateCurrent), accountOpts); err != nil {
		return fmt.Errorf("register account patch route: %w", err)
	}
	if err := register("DELETE /v1/account", "/v1/account", http.HandlerFunc(h.Account.DeleteCurrent), accountOpts); err != nil {
		return fmt.Errorf("register account delete route: %w", err)
	}

	apiKeyOpts := accountOpts
	if err := register("GET /v1/account/api-keys", "/v1/account/api-keys", http.HandlerFunc(h.APIKey.ListKeys), apiKeyOpts); err != nil {
		return fmt.Errorf("register api key list route: %w", err)
	}
	if err := register("POST /v1/account/api-keys", "/v1/account/api-keys", http.HandlerFunc(h.APIKey.CreateKey), apiKeyOpts); err != nil {
		return fmt.Errorf("register api key create route: %w", err)
	}
	if err := register("DELETE /v1/account/api-keys/{key_id}", "/v1/account/api-keys/{key_id}", http.HandlerFunc(h.APIKey.RevokeKey), apiKeyOpts); err != nil {
		return fmt.Errorf("register api key revoke route: %w", err)
	}
	if err := register("POST /v1/account/api-keys/{key_id}/rotate", "/v1/account/api-keys/{key_id}/rotate", http.HandlerFunc(h.APIKey.RotateKey), apiKeyOpts); err != nil {
		return fmt.Errorf("register api key rotate route: %w", err)
	}

	usageOpts := accountOpts
	if err := register("GET /v1/account/usage", "/v1/account/usage", http.HandlerFunc(h.Usage.UsageSummary), usageOpts); err != nil {
		return fmt.Errorf("register usage summary route: %w", err)
	}
	if err := register("GET /v1/account/usage/history", "/v1/account/usage/history", http.HandlerFunc(h.Usage.UsageHistory), usageOpts); err != nil {
		return fmt.Errorf("register usage history route: %w", err)
	}
	if err := register("GET /v1/account/usage/endpoints", "/v1/account/usage/endpoints", http.HandlerFunc(h.Usage.EndpointUsage), usageOpts); err != nil {
		return fmt.Errorf("register usage endpoints route: %w", err)
	}
	if err := register("GET /v1/account/usage/dataset-groups", "/v1/account/usage/dataset-groups", http.HandlerFunc(h.Usage.DatasetGroupUsage), usageOpts); err != nil {
		return fmt.Errorf("register usage dataset groups route: %w", err)
	}

	return nil
}
