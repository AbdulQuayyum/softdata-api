package middlewares

import (
	"errors"
	"net/http"

	"github.com/AbdulQuayyum/softdata-api/internal/response"
)

// Recovery returns middleware that converts downstream panics into safe 500 responses.
func Recovery() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			recorder := newLoggingResponseWriter(w)
			defer func() {
				if rec := recover(); rec != nil {
					if errors.Is(asError(rec), http.ErrAbortHandler) || rec == http.ErrAbortHandler {
						panic(rec)
					}
					if recorder.status != 0 {
						return
					}
					requestID, _ := RequestIDFromContext(r.Context())
					_ = response.Error(recorder, errors.New("internal server error"), requestID)
				}
			}()
			next.ServeHTTP(recorder, r)
		})
	}
}

func asError(value any) error {
	if err, ok := value.(error); ok {
		return err
	}
	return nil
}
