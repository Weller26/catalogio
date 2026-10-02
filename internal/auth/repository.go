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

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) CreateUser(
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

func (r *PostgresRepository) GetUserByEmail(
	ctx context.Context,
	email string,
) (User, string, error) {
	const query = `
		SELECT id, email, password_hash
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

func (r *PostgresRepository) GetUserByID(
	ctx context.Context,
	userID uuid.UUID,
) (User, error) {
	const query = `
		SELECT id, email
		FROM users
		WHERE id = $1
	`

	var user User

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.Email,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}

	if err != nil {
		return User{}, fmt.Errorf(
			"get user by id: %w",
			err,
		)
	}

	return user, nil
}

func (r *PostgresRepository) CreateRefreshToken(
	ctx context.Context,
	userID uuid.UUID,
	tokenHash []byte,
	expriresAt time.Time,
) error {
	const query = `
		INSERT INTO refresh_tokens (
			user_id,
			token_hash,
			expires_at
		)
		VALUES ($1, $2, $3)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		userID,
		tokenHash,
		expriresAt,
	)

	if err != nil {
		return fmt.Errorf(
			"create refresh token: %w",
			err,
		)
	}

	return nil
}

func (r *PostgresRepository) GetRefreshToken(
	ctx context.Context,
	tokenHash []byte,
) (uuid.UUID, time.Time, error) {
	const query = `
		SELECT user_id, expires_at
		FROM refresh_tokens
		WHERE token_hash = $1
	`

	var userID uuid.UUID
	var expiresAt time.Time

	err := r.db.QueryRow(
		ctx,
		query,
		tokenHash,
	).Scan(
		&userID,
		&expiresAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, time.Time{}, ErrNotFound
	}

	if err != nil {
		return uuid.Nil, time.Time{}, fmt.Errorf(
			"get refresh token: %w",
			err,
		)
	}

	return userID, expiresAt, nil
}

func (r *PostgresRepository) DeleteRefreshToken(
	ctx context.Context,
	tokenHash []byte,
) error {
	const query = `
		DELETE FROM refresh_tokens
		WHERE token_hash = $1
	`

	_, err := r.db.Exec(
		ctx,
		query,
		tokenHash,
	)

	if err != nil {
		return fmt.Errorf(
			"delete refresh token: %w",
			err,
		)
	}

	return nil
}