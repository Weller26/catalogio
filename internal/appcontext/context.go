package appcontext

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const UserIDContextKey contextKey = "user_id"

func UserIDFromContext(
	ctx context.Context,
) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(uuid.UUID)
	return userID, ok
}