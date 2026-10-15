package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/Weller26/catalogio/internal/appcontext"
	"github.com/Weller26/catalogio/internal/auth"
	"github.com/Weller26/catalogio/internal/httpresponse"
)

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