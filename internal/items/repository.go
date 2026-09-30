package items

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) CreateItem (
	ctx context.Context,
	userID uuid.UUID,
	title *string,
	description *string,
	status *string,
	rating *float64,
	notes *string,
) (Item, error) {
	const query = `
		INSERT INTO items (
			user_id,
			title,
			description,
			status,
			rating,
			notes
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, description, status, rating, notes, created_at, updated_at
	`

	var item Item

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		title,
		description,
		status,
		rating,
		notes,
	).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Status,
		&item.Rating,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		return Item{}, fmt.Errorf(
			"create item: %w",
			err,
		)
	}

	return item, nil
}

func (r *Repository) ListItems (
	ctx context.Context,
	userID uuid.UUID,
) ([]Item, error) {
	const query = `
		SELECT id, title, description, status, rating, notes,
			created_at, updated_at
		FROM items
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list items: %w", err)
	}
	defer rows.Close()

	var items []Item

	for rows.Next() {
		var item Item

		err := rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.Status,
			&item.Rating,
			&item.Notes,
			&item.CreatedAt,
			&item.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scan item: %w", err)
		}

		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate items: %w", err)
	}

	return items, nil
}

func (r *Repository) GetItemByID (
	ctx context.Context,
	itemID uuid.UUID,
	userID uuid.UUID,
) (Item, error) {
	const query = `
		SELECT id, title, description, status, rating, notes,
			created_at, updated_at
		FROM items
		WHERE id = $1 AND user_id = $2
	`

	var item Item

	err := r.db.QueryRow(
		ctx,
		query,
		itemID,
		userID,
	).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Status,
		&item.Rating,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		return Item{}, fmt.Errorf(
			"get item: %w",
			err,
		)
	}

	return item, nil
}

func (r *Repository) UpdateItemByID (
	ctx context.Context,
	itemID uuid.UUID,
	userID uuid.UUID,
	title *string,
	description *string,
	status *string,
	rating *float64,
	notes *string,
) (Item, error) {
	const query = `
		UPDATE items 
		SET
			title = $3,
			description = $4,
			status = $5,
			rating = $6,
			notes = $7,
			updated_at = now()
		WHERE id = $1 AND user_id = $2
		RETURNING id, title, description, status, rating, notes, created_at, updated_at
	`

	var item Item

	err := r.db.QueryRow(
		ctx,
		query,
		itemID,
		userID,
		title,
		description,
		status,
		rating,
		notes,
	).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Status,
		&item.Rating,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		return Item{}, fmt.Errorf(
			"update item: %w",
			err,
		)
	}

	return item, nil
}

func (r *Repository) DeleteItemByID (
	ctx context.Context,
	itemID uuid.UUID,
	userID uuid.UUID,
) (Item, error) {
	const query = `
		DELETE FROM items
		WHERE id = $1 AND user_id = $2
		RETURNING id, title, description, status, rating, notes`

	var item Item

	err := r.db.QueryRow(
		ctx,
		query,
		itemID,
		userID,
	).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Status,
		&item.Rating,
		&item.Notes,
	)

	if err != nil {
		return Item{}, fmt.Errorf(
			"delete item: %w",
			err,
		)
	}

	return item, nil
}