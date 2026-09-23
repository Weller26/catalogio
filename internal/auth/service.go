package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New(
	"invalid credentials",
)

type Service struct {
	repo *Repository
	sessionTTL time.Duration
}

func NewService(
	repo *Repository,
	sessionTTL time.Duration,
) *Service {
	return &Service{
		repo: repo,
		sessionTTL: sessionTTL,
	}
}

func (s *Service) Register(
	ctx context.Context,
	req RegisterRequest,
) (User, error) {
	email := normalizeEmail(req.Email)

	if err := validateEmail(email); err != nil {
		return User{}, err
	}

	if err := validatePassword(req.Password); err != nil {
		return User{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(req.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return User{}, err
	}

	return s.repo.CreateUser(
		ctx,
		email,
		string(passwordHash),
	)
}

func (s *Service) Login(
	ctx context.Context,
	req LoginRequest,
) (User, string, error) {
	email := normalizeEmail(req.Email)

	user, passwordHash, err := s.repo.GetUserByEmail(ctx, email)

	if errors.Is(err, ErrNotFound) {
		return  User{}, "", ErrInvalidCredentials
	}

	if err != nil {
		return User{}, "", err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(req.Password),
	)

	if err != nil {
		return User{}, "", ErrInvalidCredentials
	}

	token, err := generateSessionToken()
	if err != nil {
		return User{}, "", err
	}

	tokenHash := hashSessionToken(token)

	expiresAt := time.Now().Add(s.sessionTTL)

	err = s.repo.CreateSession(
		ctx,
		tokenHash,
		user.ID,
		expiresAt,
	)

	if err != nil {
		return User{}, "", err
	}

	return user, token, nil
}

func (s *Service) Authenticate(
	ctx context.Context,
	token string,
) (User, error) {
	if token == "" {
		return User{}, ErrInvalidCredentials
	}

	tokenHash := hashSessionToken(token)

	return s.repo.GetUserBySession(ctx, tokenHash)
}

func (s *Service) Logout(
	ctx context.Context,
	token string,
) error {
	if token == "" {
		return nil
	}

	tokenHash := hashSessionToken(token)

	return s.repo.DeleteSession(
		ctx,
		tokenHash,
	)
}

func normalizeEmail(email string) string {
	return strings.ToLower(
		strings.TrimSpace(email),
	)
}

func validateEmail(email string) error {
	parsed, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("invalid email")
	}

	if parsed.Address != email {
		return errors.New("invalid email")
	}

	return nil
}

func validatePassword(password string) error {
	if len([]byte(password)) < 8 {
		return errors.New(
			"password must contain at least 8 bytes",
		)
	}

	if len([]byte(password)) > 72 {
		return errors.New(
			"password must not exceed 72 bytes",
		)
	}

	return nil
}

func generateSessionToken() (string, error) {
	data := make([]byte, 32)

	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func hashSessionToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))

	return hash[:]
}