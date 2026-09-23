package auth

import (
	"context"
	"errors"
	"net/http"
)

type contextKey string

const userContextKey contextKey = "user"

func UserFromContext(
	ctx context.Context,
) (User, bool) {
	user, ok := ctx.Value(
		userContextKey,
	).(User)

	return user, ok
}

func (s *Service) RequireAuth(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(
				"catalogio_session",
			)

			if err != nil {
				writeError(
					w,
					http.StatusUnauthorized,
					"unauthorized",
				)
				return
			}

			user, err := s.Authenticate(
				r.Context(),
				cookie.Value,
			)

			if errors.Is(err, ErrNotFound) || errors.Is(err, ErrInvalidCredentials) {
				writeError(
					w,
					http.StatusUnauthorized,
					"unauthorized",
				)
				return
			}

			if err != nil {
				writeError(
					w,
					http.StatusInternalServerError,
					"internal server error",
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				userContextKey,
				user,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}