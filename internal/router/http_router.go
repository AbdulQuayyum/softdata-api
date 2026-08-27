package router

import (
	"context"
	"net/http"
	"strings"

	"github.com/AbdulQuayyum/softdata-api/internal/middlewares"
	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
)

type httpRouter struct {
	mux    *http.ServeMux
	routes *routeCatalog
}

func newHTTPRouter(mux *http.ServeMux, routes *routeCatalog) http.Handler {
	return &httpRouter{mux: mux, routes: routes}
}

func (h *httpRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.mux == nil || r == nil {
		_ = response.Error(w, interfaces.ErrNotFound, "")
		return
	}

	requestID := requestIDFromContext(r.Context())
	path := requestPath(r)

	if h.routes == nil || path == "" {
		_ = response.Error(w, interfaces.ErrNotFound, requestID)
		return
	}

	if h.routes.supports(r.Method, path) {
		h.mux.ServeHTTP(w, r)
		return
	}

	allow := h.routes.allow(path)
	if len(allow) == 0 {
		_ = response.Error(w, interfaces.ErrNotFound, requestID)
		return
	}

	writeMethodNotAllowed(w, requestID, allow)
}

func requestPath(r *http.Request) string {
	if r == nil || r.URL == nil {
		return ""
	}
	return strings.TrimSpace(r.URL.Path)
}

func writeMethodNotAllowed(w http.ResponseWriter, requestID string, allow []string) {
	if w == nil {
		return
	}
	allow = sanitizeAllowMethods(allow)
	if len(allow) > 0 {
		w.Header().Set("Allow", strings.Join(allow, ", "))
	}
	_ = response.JSON(w, http.StatusMethodNotAllowed, response.ErrorResponse{
		Success: false,
		Error: response.ErrorBody{
			Code:      "INVALID_REQUEST",
			Message:   "The request was invalid.",
			RequestID: requestID,
		},
	})
}

func sanitizeAllowMethods(methods []string) []string {
	if len(methods) == 0 {
		return nil
	}
	sanitized := make([]string, 0, len(methods))
	for _, method := range methods {
		if strings.EqualFold(method, http.MethodHead) {
			continue
		}
		sanitized = append(sanitized, method)
	}
	return sanitized
}

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := middlewares.RequestIDFromContext(ctx)
	return requestID
}
