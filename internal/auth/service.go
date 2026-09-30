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

	"github.com/google/uuid"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidCredentials = errors.New(
	"invalid credentials",
)

type Claims struct {
	UserID string `json:"sub"`
	jwt.RegisteredClaims
}

type Service struct {
	repo *Repository
	jwtSecret []byte
}

func NewService(
	repo *Repository,
	jwtSecret string,
) *Service {
	return &Service{
		repo: repo,
		jwtSecret: []byte(jwtSecret),
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
) (User, string, string, error) {
	email := normalizeEmail(req.Email)

	user, passwordHash, err := s.repo.GetUserByEmail(ctx, email)

	if errors.Is(err, ErrNotFound) {
		return  User{}, "", "", ErrInvalidCredentials
	}

	if err != nil {
		return User{}, "", "", err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(req.Password),
	)

	if err != nil {
		return User{}, "", "", ErrInvalidCredentials
	}

	accessToken, err := s.generateAccessToken(user.ID)
	if err != nil {
		return User{}, "", "", err
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return User{}, "", "", err
	}

	refreshTokenHash := hashRefreshToken(refreshToken)

	expiresAt := time.Now().Add(24 * time.Hour)

	err = s.repo.CreateRefreshToken(
		ctx,
		user.ID,
		refreshTokenHash,
		expiresAt,
	)

	if err != nil {
		return User{}, "", "", err
	}

	return user, accessToken, refreshToken, nil
}

func (s *Service) Authenticate(
	ctx context.Context,
	tokenString string,
) (uuid.UUID, error) {
	if tokenString == "" {
		return uuid.Nil, ErrInvalidCredentials
	}

	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}

			return s.jwtSecret, nil
		},
	)

	if err != nil || !token.Valid {
		return uuid.Nil, ErrInvalidCredentials
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, ErrInvalidCredentials
	}

	return userID, nil
}

func (s *Service) Logout(
	ctx context.Context,
	refreshToken string,
) error {
	if refreshToken == "" {
		return nil
	}

	tokenHash := hashRefreshToken(refreshToken)

	return s.repo.DeleteRefreshToken(
		ctx,
		tokenHash,
	)
}

func (s *Service) Refresh(
	ctx context.Context,
	refreshToken string,
) (string, error) {
	if refreshToken == "" {
		return "", ErrInvalidCredentials
	}

	tokenHash := hashRefreshToken(refreshToken)

	userID, expiresAt, err := s.repo.GetRefreshToken(
		ctx,
		tokenHash,
	)

	if errors.Is(err, ErrNotFound) {
        return "", ErrInvalidCredentials
    }

    if err != nil {
        return "", err
    }

    if time.Now().After(expiresAt) {
        return "", ErrInvalidCredentials
    }

    return s.generateAccessToken(userID)
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

func (s *Service) generateAccessToken(
	userID uuid.UUID,
) (string, error) {
	now := time.Now()

	claims := Claims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(
				now.Add(15 * time.Minute),
			),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(s.jwtSecret)
}

func generateRefreshToken() (string, error) {
	data := make([]byte, 32)

	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(data), nil
}

func hashRefreshToken(token string) []byte {
	hash := sha256.Sum256([]byte(token))

	return hash[:]
}