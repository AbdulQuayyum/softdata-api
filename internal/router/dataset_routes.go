package router

import (
	"fmt"
	"net/http"
)

func registerDatasetRoutes(mux *http.ServeMux, catalog *routeCatalog, h Handlers, mw Middleware) error {
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

	datasetOpts := routeOptions{
		useOptionalAPIKey: true,
		useRateLimit:      true,
		useUsageTracking:  true,
	}

	if err := register("GET /v1/datasets", "/v1/datasets", http.HandlerFunc(h.Dataset.ListDatasets), datasetOpts); err != nil {
		return fmt.Errorf("register dataset list route: %w", err)
	}
	if err := register("GET /v1/datasets/{dataset_id}", "/v1/datasets/{dataset_id}", http.HandlerFunc(h.Dataset.GetDataset), datasetOpts); err != nil {
		return fmt.Errorf("register dataset detail route: %w", err)
	}
	if err := register("GET /v1/datasets/{dataset_id}/sources", "/v1/datasets/{dataset_id}/sources", http.HandlerFunc(h.Dataset.ListDatasetSources), datasetOpts); err != nil {
		return fmt.Errorf("register dataset sources route: %w", err)
	}
	if err := register("GET /v1/datasets/{dataset_id}/versions", "/v1/datasets/{dataset_id}/versions", http.HandlerFunc(h.Dataset.ListDatasetVersions), datasetOpts); err != nil {
		return fmt.Errorf("register dataset versions route: %w", err)
	}

	return nil
}
