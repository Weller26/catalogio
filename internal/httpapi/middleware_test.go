package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Weller26/catalogio/internal/appcontext"
	"github.com/Weller26/catalogio/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestRequireAuthProtectsHandler(t *testing.T) {
	userID := uuid.New()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}
	tests := []struct {
		name       string
		header     string
		wantStatus int
	}{
		{name: "missing header", wantStatus: http.StatusUnauthorized},
		{name: "wrong scheme", header: "Basic " + token, wantStatus: http.StatusUnauthorized},
		{name: "missing token", header: "Bearer ", wantStatus: http.StatusUnauthorized},
		{name: "invalid token", header: "Bearer invalid", wantStatus: http.StatusUnauthorized},
		{name: "valid token", header: "Bearer " + token, wantStatus: http.StatusNoContent},
	}
	middleware := NewAuthMiddleware(auth.NewService(nil, "test-secret"))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := middleware.requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				got, ok := appcontext.UserIDFromContext(r.Context())
				if !ok || got != userID {
					t.Errorf("handler user ID = %v, present = %v", got, ok)
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			req := httptest.NewRequest(http.MethodGet, "/api/v1/items", nil)
			req.Header.Set("Authorization", tt.header)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if called != (tt.wantStatus == http.StatusNoContent) {
				t.Fatalf("protected handler called = %v", called)
			}
		})
	}
}
