package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

// func UserFromContext(
// 	ctx context.Context,
// ) (User, bool) {
// 	user, ok := ctx.Value(
// 		userContextKey,
// 	).(User)

// 	return user, ok
// }

func UserIDFromContext(
	ctx context.Context,
) (uuid.UUID, bool) {
	userID, ok := ctx.Value(
		userIDContextKey,
	).(uuid.UUID)

	return userID, ok
}

func (s *Service) RequireAuth(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				writeError(
					w,
					http.StatusUnauthorized,
					"autorization required",
				)
				return
			}

			parts := strings.SplitN(
				authHeader,
				" ",
				2,
			)

			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
				writeError(
					w,
					http.StatusUnauthorized,
					"invalid authorization header",
				)
				return
			}

			accessToken := parts[1]

			userID, err := s.Authenticate(
				r.Context(),
				accessToken,
			)

			if errors.Is(err, ErrInvalidCredentials) {
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
                userIDContextKey,
                userID,
            )

            next.ServeHTTP(
                w,
                r.WithContext(ctx),
            )
		},
	)
}