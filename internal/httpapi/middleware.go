package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/Weller26/catalogio/internal/appcontext"
	"github.com/Weller26/catalogio/internal/auth"
	"github.com/Weller26/catalogio/internal/httpresponse"
)

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
					w,
					http.StatusUnauthorized,
					"authorization required",
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
					w,
					http.StatusUnauthorized,
					"invalid authorization header",
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
                    w,
                    http.StatusUnauthorized,
                    "unauthorized",
                )
                return
            }

            if err != nil {
                httpresponse.WriteError(
                    w,
                    http.StatusInternalServerError,
                    "internal server error",
                )
                return
            }

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