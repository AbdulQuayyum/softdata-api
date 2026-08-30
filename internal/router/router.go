package router

import (
	"fmt"
	"net/http"

	"github.com/AbdulQuayyum/softdata-api/internal/handlers"
	"github.com/AbdulQuayyum/softdata-api/internal/middlewares"
)

// MiddlewareFunc is the router-facing middleware signature.
type MiddlewareFunc = middlewares.Middleware

// UsageMiddlewareFactory builds route-specific usage-tracking middleware.
type UsageMiddlewareFactory func(endpoint, datasetGroup string) (MiddlewareFunc, error)

// Handlers bundles the completed handler instances the router wires together.
type Handlers struct {
	Health    *handlers.HealthHandler
	Discovery *handlers.DiscoveryHandler
	Geography *handlers.GeographyHandler
	Finance   *handlers.FinanceHandler
	Auth      *handlers.AuthHandler
	Account   *handlers.AccountHandler
	APIKey    *handlers.APIKeyHandler
	Usage     *handlers.UsageHandler
	Dataset   *handlers.DatasetHandler
}

// Middleware bundles the completed middleware functions the router composes.
type Middleware struct {
	RequestID       MiddlewareFunc
	Recovery        MiddlewareFunc
	Logger          MiddlewareFunc
	SecurityHeaders MiddlewareFunc
	CORS            MiddlewareFunc
	BodyLimit       MiddlewareFunc
	Timeout         MiddlewareFunc
	Authentication  MiddlewareFunc
	OptionalAPIKey  MiddlewareFunc
	StandardLimit   MiddlewareFunc
	UsageTracking   UsageMiddlewareFactory
}

// New builds the full HTTP router using stdlib ServeMux.
func New(handlers Handlers, middleware Middleware) (http.Handler, error) {
	if err := validateDependencies(handlers, middleware); err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	catalog := &routeCatalog{}
	if err := registerPublicRoutes(mux, catalog, handlers, middleware); err != nil {
		return nil, err
	}
	if err := registerAuthRoutes(mux, catalog, handlers, middleware); err != nil {
		return nil, err
	}
	if err := registerAccountRoutes(mux, catalog, handlers, middleware); err != nil {
		return nil, err
	}
	if err := registerDatasetRoutes(mux, catalog, handlers, middleware); err != nil {
		return nil, err
	}

	globalChain := []MiddlewareFunc{
		middleware.RequestID,
		middleware.Recovery,
		middleware.Logger,
		middleware.SecurityHeaders,
		middleware.CORS,
		middleware.BodyLimit,
		middleware.Timeout,
	}

	return compose(newHTTPRouter(mux, catalog), globalChain...), nil
}

func validateDependencies(h Handlers, mw Middleware) error {
	switch {
	case h.Health == nil:
		return fmt.Errorf("health handler is required")
	case h.Discovery == nil:
		return fmt.Errorf("discovery handler is required")
	case h.Geography == nil:
		return fmt.Errorf("geography handler is required")
	case h.Finance == nil:
		return fmt.Errorf("finance handler is required")
	case h.Auth == nil:
		return fmt.Errorf("auth handler is required")
	case h.Account == nil:
		return fmt.Errorf("account handler is required")
	case h.APIKey == nil:
		return fmt.Errorf("api key handler is required")
	case h.Usage == nil:
		return fmt.Errorf("usage handler is required")
	case h.Dataset == nil:
		return fmt.Errorf("dataset handler is required")
	case mw.RequestID == nil:
		return fmt.Errorf("request id middleware is required")
	case mw.Recovery == nil:
		return fmt.Errorf("recovery middleware is required")
	case mw.Logger == nil:
		return fmt.Errorf("logger middleware is required")
	case mw.SecurityHeaders == nil:
		return fmt.Errorf("security headers middleware is required")
	case mw.CORS == nil:
		return fmt.Errorf("cors middleware is required")
	case mw.BodyLimit == nil:
		return fmt.Errorf("body limit middleware is required")
	case mw.Timeout == nil:
		return fmt.Errorf("timeout middleware is required")
	case mw.Authentication == nil:
		return fmt.Errorf("authentication middleware is required")
	case mw.OptionalAPIKey == nil:
		return fmt.Errorf("optional api key middleware is required")
	case mw.StandardLimit == nil:
		return fmt.Errorf("rate limit middleware is required")
	case mw.UsageTracking == nil:
		return fmt.Errorf("usage tracking middleware factory is required")
	}
	return nil
}

func compose(handler http.Handler, middlewares ...MiddlewareFunc) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		if middlewares[i] == nil {
			continue
		}
		handler = middlewares[i](handler)
	}
	return handler
}

func buildRouteMiddlewares(mw Middleware, endpoint, datasetGroup string, opts routeOptions) ([]MiddlewareFunc, error) {
	chain := make([]MiddlewareFunc, 0, 4)
	if opts.useOptionalAPIKey {
		chain = append(chain, mw.OptionalAPIKey)
	}
	if opts.useAuthentication {
		chain = append(chain, mw.Authentication)
	}
	if opts.useRateLimit {
		chain = append(chain, mw.StandardLimit)
	}
	if opts.useUsageTracking {
		usage, err := mw.UsageTracking(endpoint, datasetGroup)
		if err != nil {
			return nil, err
		}
		chain = append(chain, usage)
	}
	return chain, nil
}

type routeOptions struct {
	useAuthentication bool
	useOptionalAPIKey bool
	useRateLimit      bool
	useUsageTracking  bool
}
