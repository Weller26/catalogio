package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Only registration uses this repository; no database is needed.
type registrationRepository struct {
	Repository
	calls        int
	email        string
	passwordHash string
	user         User
}

func (r *registrationRepository) CreateUser(_ context.Context, email, passwordHash string) (User, error) {
	r.calls++
	r.email = email
	r.passwordHash = passwordHash
	return r.user, nil
}

func TestRegisterValidatesCredentials(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{name: "normalized email and minimum password", email: "  Alice@Example.COM  ", password: "12345678"},
		{name: "maximum password", email: "alice@example.com", password: strings.Repeat("a", 72)},
		{name: "empty email", password: "12345678", wantErr: true},
		{name: "invalid email", email: "not-an-email", password: "12345678", wantErr: true},
		{name: "display name is not an email", email: "Alice <alice@example.com>", password: "12345678", wantErr: true},
		{name: "short password", email: "alice@example.com", password: "1234567", wantErr: true},
		{name: "too long password", email: "alice@example.com", password: strings.Repeat("a", 73), wantErr: true},
		{name: "password limit counts bytes", email: "alice@example.com", password: strings.Repeat("я", 37), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &registrationRepository{user: User{ID: uuid.New(), Email: "alice@example.com"}}
			service := NewService(repo, "test-secret")
			user, err := service.Register(context.Background(), RegisterRequest{Email: tt.email, Password: tt.password})
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected invalid credentials to be rejected")
				}
				if repo.calls != 0 {
					t.Fatal("invalid registration must not reach the repository")
				}
				return
			}
			if err != nil {
				t.Fatalf("Register: %v", err)
			}
			if repo.calls != 1 || repo.email != "alice@example.com" {
				t.Fatalf("repository calls = %d, email = %q", repo.calls, repo.email)
			}
			if user != repo.user {
				t.Fatalf("returned user = %+v, want %+v", user, repo.user)
			}
			if repo.passwordHash == tt.password {
				t.Fatal("repository received a plaintext password")
			}
			if err := bcrypt.CompareHashAndPassword([]byte(repo.passwordHash), []byte(tt.password)); err != nil {
				t.Fatalf("stored hash does not match the password: %v", err)
			}
		})
	}
}

func TestAuthenticateValidatesToken(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name      string
		secret    string
		method    jwt.SigningMethod
		subject   string
		expiresAt time.Time
		wantErr   bool
	}{
		{name: "valid", secret: "test-secret", method: jwt.SigningMethodHS256, subject: userID.String(), expiresAt: time.Now().Add(time.Hour)},
		{name: "expired", secret: "test-secret", method: jwt.SigningMethodHS256, subject: userID.String(), expiresAt: time.Now().Add(-time.Hour), wantErr: true},
		{name: "wrong secret", secret: "another-secret", method: jwt.SigningMethodHS256, subject: userID.String(), expiresAt: time.Now().Add(time.Hour), wantErr: true},
		{name: "wrong signing method", secret: "test-secret", method: jwt.SigningMethodHS384, subject: userID.String(), expiresAt: time.Now().Add(time.Hour), wantErr: true},
		{name: "invalid user ID", secret: "test-secret", method: jwt.SigningMethodHS256, subject: "not-a-uuid", expiresAt: time.Now().Add(time.Hour), wantErr: true},
	}
	service := NewService(nil, "test-secret")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwt.NewWithClaims(tt.method, jwt.MapClaims{
				"sub": tt.subject,
				"exp": tt.expiresAt.Unix(),
			}).SignedString([]byte(tt.secret))
			if err != nil {
				t.Fatalf("sign test token: %v", err)
			}
			got, err := service.Authenticate(context.Background(), token)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidCredentials) || got != uuid.Nil {
					t.Fatalf("Authenticate = (%v, %v), want invalid credentials", got, err)
				}
				return
			}
			if err != nil || got != userID {
				t.Fatalf("Authenticate = (%v, %v), want (%v, nil)", got, err, userID)
			}
		})
	}
	for _, token := range []string{"", "not-a-jwt"} {
		if _, err := service.Authenticate(context.Background(), token); !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("Authenticate(%q): expected invalid credentials, got %v", token, err)
		}
	}
}
