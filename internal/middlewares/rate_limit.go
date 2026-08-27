package middlewares

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AbdulQuayyum/softdata-api/internal/repository/interfaces"
	"github.com/AbdulQuayyum/softdata-api/internal/response"
	"github.com/AbdulQuayyum/softdata-api/internal/security"
)

const (
	rateLimitHeaderLimit      = "RateLimit-Limit"
	rateLimitHeaderRemaining  = "RateLimit-Remaining"
	rateLimitHeaderReset      = "RateLimit-Reset"
	rateLimitHeaderRetryAfter = "Retry-After"
	rateLimitErrorCode        = "RATE_LIMIT_EXCEEDED"
	rateLimitErrorMessage     = "The request limit has been exceeded."
)

// RateLimitPolicy configures the short-window limits for public traffic.
type RateLimitPolicy struct {
	AnonymousLimit int64
	APIKeyLimit    int64
	DownloadLimit  int64
	Window         time.Duration
	FailOpen       bool
}

// AnonymousIdentifier produces the opaque daily anonymous identifier used for anonymous traffic.
type AnonymousIdentifier interface {
	Identify(r *http.Request) (string, error)
}

type securityAnonymousIdentifier struct {
	secret string
}

// NewSecurityAnonymousIdentifier returns an anonymous-ID generator backed by the existing security helper.
func NewSecurityAnonymousIdentifier(secret string) (AnonymousIdentifier, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return nil, fmt.Errorf("anonymous id secret is required")
	}
	return &securityAnonymousIdentifier{secret: secret}, nil
}

func (i *securityAnonymousIdentifier) Identify(r *http.Request) (string, error) {
	if i == nil {
		return "", fmt.Errorf("anonymous identifier is required")
	}
	if r == nil {
		return "", fmt.Errorf("request is required")
	}
	if err := r.Context().Err(); err != nil {
		return "", err
	}

	ip, err := normalizedRemoteIP(r.RemoteAddr)
	if err != nil {
		return "", err
	}

	return security.DeriveAnonymousID(i.secret, ip, strings.TrimSpace(r.UserAgent()), time.Now().UTC())
}

type rateLimitStage int

const (
	rateLimitStageOrdinary rateLimitStage = iota
	rateLimitStageDownload
)

type rateLimitMiddleware struct {
	repository          interfaces.RateLimitRepository
	anonymousIdentifier AnonymousIdentifier
	policy              RateLimitPolicy
	stage               rateLimitStage
}

// RateLimit returns middleware for ordinary public requests.
func RateLimit(repository interfaces.RateLimitRepository, anonymousIdentifier AnonymousIdentifier, policy RateLimitPolicy) (Middleware, error) {
	return newRateLimitMiddleware(repository, anonymousIdentifier, policy, rateLimitStageOrdinary)
}

// DownloadRateLimit returns middleware for dataset-download requests.
func DownloadRateLimit(repository interfaces.RateLimitRepository, anonymousIdentifier AnonymousIdentifier, policy RateLimitPolicy) (Middleware, error) {
	return newRateLimitMiddleware(repository, anonymousIdentifier, policy, rateLimitStageDownload)
}

func newRateLimitMiddleware(repository interfaces.RateLimitRepository, anonymousIdentifier AnonymousIdentifier, policy RateLimitPolicy, stage rateLimitStage) (Middleware, error) {
	if repository == nil {
		return nil, fmt.Errorf("rate limit repository is required")
	}
	if anonymousIdentifier == nil {
		return nil, fmt.Errorf("anonymous identifier is required")
	}
	if policy.AnonymousLimit <= 0 || policy.APIKeyLimit <= 0 || policy.DownloadLimit <= 0 {
		return nil, fmt.Errorf("rate limit limits must be positive")
	}
	if policy.Window <= 0 {
		return nil, fmt.Errorf("rate limit window must be positive")
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID, _ := RequestIDFromContext(r.Context())
			if err := r.Context().Err(); err != nil {
				_ = response.Error(w, err, requestID)
				return
			}

			req, err := buildRateLimitRequest(r, anonymousIdentifier, policy, stage)
			if err != nil {
				_ = response.Error(w, err, requestID)
				return
			}

			result, err := repository.Allow(r.Context(), req)
			if err != nil {
				switch {
				case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
					_ = response.Error(w, err, requestID)
				case errors.Is(err, interfaces.ErrRateLimitUnavailable):
					next.ServeHTTP(w, r)
				default:
					_ = response.Error(w, err, requestID)
				}
				return
			}

			if !result.Allowed {
				if err := writeRateLimitExceeded(w, requestID, result); err != nil {
					_ = response.Error(w, err, requestID)
					return
				}
				return
			}

			if err := writeRateLimitHeaders(w, result); err != nil {
				_ = response.Error(w, err, requestID)
				return
			}

			next.ServeHTTP(w, r)
		})
	}, nil
}

func buildRateLimitRequest(r *http.Request, anonymousIdentifier AnonymousIdentifier, policy RateLimitPolicy, stage rateLimitStage) (interfaces.RateLimitRequest, error) {
	if r == nil {
		return interfaces.RateLimitRequest{}, fmt.Errorf("request is required")
	}

	if identity, ok := APIKeyIdentityFromContext(r.Context()); ok {
		switch stage {
		case rateLimitStageDownload:
			return interfaces.RateLimitRequest{
				SubjectKind: interfaces.RateLimitSubjectDownload,
				Subject:     identity.APIKeyID,
				Limit:       policy.DownloadLimit,
				Window:      policy.Window,
			}, nil
		default:
			return interfaces.RateLimitRequest{
				SubjectKind: interfaces.RateLimitSubjectAPIKey,
				Subject:     identity.APIKeyID,
				Limit:       policy.APIKeyLimit,
				Window:      policy.Window,
			}, nil
		}
	}

	subject, err := anonymousIdentifier.Identify(r)
	if err != nil {
		return interfaces.RateLimitRequest{}, err
	}

	subjectKind := interfaces.RateLimitSubjectAnonymous
	limit := policy.AnonymousLimit
	if stage == rateLimitStageDownload {
		subjectKind = interfaces.RateLimitSubjectDownload
		limit = policy.DownloadLimit
	}

	return interfaces.RateLimitRequest{
		SubjectKind: subjectKind,
		Subject:     subject,
		Limit:       limit,
		Window:      policy.Window,
	}, nil
}

func writeRateLimitHeaders(w http.ResponseWriter, result interfaces.RateLimitResult) error {
	if result.Limit <= 0 || result.ResetAt.IsZero() {
		return fmt.Errorf("invalid rate limit result")
	}

	w.Header().Set(rateLimitHeaderLimit, strconv.FormatInt(result.Limit, 10))
	w.Header().Set(rateLimitHeaderRemaining, strconv.FormatInt(maxInt64(result.Remaining, 0), 10))
	w.Header().Set(rateLimitHeaderReset, strconv.FormatInt(result.ResetAt.UTC().Unix(), 10))
	return nil
}

func writeRateLimitExceeded(w http.ResponseWriter, requestID string, result interfaces.RateLimitResult) error {
	if err := writeRateLimitHeaders(w, result); err != nil {
		return err
	}

	retryAfter := retryAfterSeconds(result.ResetAt)
	if retryAfter < 0 {
		retryAfter = 0
	}
	w.Header().Set(rateLimitHeaderRetryAfter, strconv.FormatInt(retryAfter, 10))

	body := rateLimitErrorResponse{
		Success: false,
		Error: rateLimitErrorBody{
			Code:              rateLimitErrorCode,
			Message:           rateLimitErrorMessage,
			RetryAfterSeconds: retryAfter,
			RequestID:         requestID,
		},
	}
	return response.JSON(w, http.StatusTooManyRequests, body)
}

func retryAfterSeconds(resetAt time.Time) int64 {
	if resetAt.IsZero() {
		return 0
	}
	delta := time.Until(resetAt.UTC())
	if delta <= 0 {
		return 0
	}
	seconds := delta / time.Second
	if delta%time.Second != 0 {
		seconds++
	}
	if seconds < 0 {
		return 0
	}
	return int64(seconds)
}

func normalizedRemoteIP(remoteAddr string) (string, error) {
	value := strings.TrimSpace(remoteAddr)
	if value == "" {
		return "", fmt.Errorf("remote addr is required")
	}

	host := value
	if parsedHost, _, err := net.SplitHostPort(value); err == nil {
		host = parsedHost
	}

	host = strings.TrimSpace(host)
	if host == "" || strings.ContainsAny(host, " \t\r\n") {
		return "", fmt.Errorf("remote addr is invalid")
	}
	return host, nil
}

func maxInt64(values ...int64) int64 {
	var max int64
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}

type rateLimitErrorResponse struct {
	Success bool               `json:"success"`
	Error   rateLimitErrorBody `json:"error"`
}

type rateLimitErrorBody struct {
	Code              string `json:"code"`
	Message           string `json:"message"`
	RetryAfterSeconds int64  `json:"retry_after_seconds"`
	RequestID         string `json:"request_id"`
}
