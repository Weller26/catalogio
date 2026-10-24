package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Weller26/catalogio/internal/appcontext"
	"github.com/Weller26/catalogio/internal/auth"
	"github.com/Weller26/catalogio/internal/httpresponse"
	"github.com/Weller26/catalogio/internal/metrics"
	"github.com/google/uuid"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		rw := newResponseWriter(w)

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()

		pattern := r.Pattern
		if pattern == "" {
			pattern = r.URL.Path
		}
		statusStr := strconv.Itoa(rw.statusCode)

		metrics.HTTPRequestsTotal.WithLabelValues(r.Method, pattern, statusStr).Inc()
		metrics.HTTPRequestsDuration.WithLabelValues(r.Method, pattern, statusStr).Observe(duration)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		
		start := time.Now()

		next.ServeHTTP(w, r.WithContext(ctx))

		slog.InfoContext(ctx, "http request",
			"method", r.Method,
			"path", r.URL.Path,
			"request_id", requestID,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}

type AuthMiddleware struct {
	authService *auth.Service
}

func NewAuthMiddleware(
	authService *auth.Service,
) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (m *AuthMiddleware) requireAuth(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				httpresponse.WriteError(
					r.Context(), w, http.StatusUnauthorized,
					"authorization required", nil,
				)
				return
			}

			parts := strings.SplitN(
				authHeader,
				" ",
				2,
			)

			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
				httpresponse.WriteError(
					r.Context(), w, http.StatusUnauthorized,
					"invalid authorization header", nil,
				)
				return
			}

			accessToken := parts[1]

			userID, err := m.authService.Authenticate(
				r.Context(),
				accessToken,
			)

			if errors.Is(err, auth.ErrInvalidCredentials) {
                httpresponse.WriteError(
                    r.Context(), w, http.StatusUnauthorized,
                    "unauthorized", err,
                )
                return
            }

            if err != nil {
                httpresponse.WriteError(
                    r.Context(), w, http.StatusInternalServerError,
                    "internal server error", err,
                )
                return
            }

			slog.DebugContext(
				r.Context(), "access token verified",
				"user_id", userID.String(),
			)

            ctx := context.WithValue(
                r.Context(),
                appcontext.UserIDContextKey,
                userID,
            )

            next.ServeHTTP(
                w,
                r.WithContext(ctx),
            )
		},
	)
}