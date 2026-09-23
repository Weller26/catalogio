package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")
var ErrEmailTaken = errors.New("email already exists")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository {
		db: db,
	}
}

func (r *Repository) CreateUser(
	ctx context.Context,
	email string,
	passwordHash string,
) (User, error) {
	const query = `
		INSERT INTO users (
			email,
			password_hash
		)
		VALUES ($1, $2)
		RETURNING id, email
	`

	var user User

	err := r.db.QueryRow(
		ctx,
		query,
		email,
		passwordHash,
	).Scan(
		&user.ID,
		&user.Email,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, ErrEmailTaken
		}

		return User{}, fmt.Errorf(
			"create user: %w",
			err,
		)
	}

	return user, nil
}

func (r *Repository) GetUserByEmail(
	ctx context.Context,
	email string,
) (User, string, error) {
	const query = `
		SELECT
			id,
			email,
			password_hash
		FROM users
		WHERE email = $1
	`

	var user User
	var passwordHash string

	err := r.db.QueryRow(
		ctx,
		query,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&passwordHash,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, "", ErrNotFound
	}

	if err != nil {
		return User{}, "", fmt.Errorf(
			"get user by email: %w",
			err,
		)
	}

	return user, passwordHash, nil
}

func (r *Repository) CreateSession(
	ctx context.Context,
	tokenHash []byte,
	userID uuid.UUID,
	expiresAt time.Time,
) error {
	const query = `
		INSERT INTO sessions (
			token_hash,
			user_id,
			expires_at
		)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		tokenHash,
		userID,
		expiresAt,
	)

	if err != nil {
		return fmt.Errorf(
			"create session: %w",
			err,
		)
	}

	return nil
}

func (r *Repository) GetUserBySession(
	ctx context.Context,
	tokenHash []byte,
) (User, error) {
	const query = `
		SELECT
			u.id,
			u.email
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = $1
			AND s.expires_at > now()
	`

	var user User

	err := r.db.QueryRow(
		ctx,
		query,
		tokenHash,
	).Scan(
		&user.ID,
		&user.Email,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}

	if err != nil {
		return User{}, fmt.Errorf(
			"get user by session: %w", 
			err,
		)
	}

	return user, nil
}

func (r *Repository) DeleteSession(
	ctx context.Context,
	tokenHash []byte,
) error {
	const query = `
		DELETE FROM sessions
		WHERE token_hash = $1
	`

	_, err := r.db.Exec(
		ctx,
		query,
		tokenHash,
	)

	if err != nil {
		return fmt.Errorf(
			"delete session: %w",
			err,
		)
	}

	return nil
}