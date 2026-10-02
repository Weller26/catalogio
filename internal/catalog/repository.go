package catalog

import (
	"context"
	"fmt"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5"
)

var ErrNotFound = errors.New("item not found")

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *PostgresRepository{
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CreateItem(
	ctx context.Context,
	userID uuid.UUID,
	title string,
	description *string,
	statusID *uuid.UUID,
	typeID *uuid.UUID,
	rating *float64,
	notes *string,
) (Item, error) {
	const query = `
		INSERT INTO items (
			user_id,
			title,
			description,
			status_id,
			type_id,
			rating,
			notes
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING 
			id, title, description, 
			status_id, type_id, rating, 
			notes, created_at, updated_at
	`

	var item Item

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		title,
		description,
		statusID,
		typeID,
		rating,
		notes,
	).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.StatusID,
		&item.TypeID,
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

func (r *PostgresRepository) ListItems(
	ctx context.Context,
	userID uuid.UUID,
) ([]Item, error) {
	const query = `
		SELECT id, title, description, 
			status_id, type_id, rating, 
			notes, created_at, updated_at
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
			&item.StatusID,
			&item.TypeID,
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

func (r *PostgresRepository) GetItemByID(
	ctx context.Context,
	itemID uuid.UUID,
	userID uuid.UUID,
) (Item, error) {
	const query = `
		SELECT id, title, description,
			status_id, type_id, rating, 
			notes, created_at, updated_at
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
		&item.StatusID,
		&item.TypeID,
		&item.Rating,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrNotFound
	}

	if err != nil {
		return Item{}, fmt.Errorf(
			"get item: %w",
			err,
		)
	}

	return item, nil
}

func (r *PostgresRepository) UpdateItemByID(
	ctx context.Context,
	itemID uuid.UUID,
	userID uuid.UUID,
	title string,
	description *string,
	statusID *uuid.UUID,
	typeID *uuid.UUID,
	rating *float64,
	notes *string,
) (Item, error) {
	const query = `
		UPDATE items 
		SET
			title = $3,
			description = $4,
			status_id = $5,
			type_id = $6,
			rating = $7,
			notes = $8,
			updated_at = now()
		WHERE id = $1 AND user_id = $2
		RETURNING 
			id, title, description, 
			status_id, type_id, rating, 
			notes, created_at, updated_at
	`

	var item Item

	err := r.db.QueryRow(
		ctx,
		query,
		itemID,
		userID,
		title,
		description,
		statusID,
		typeID,
		rating,
		notes,
	).Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.StatusID,
		&item.TypeID,
		&item.Rating,
		&item.Notes,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrNotFound
	}

	if err != nil {
		return Item{}, fmt.Errorf(
			"update item: %w",
			err,
		)
	}

	return item, nil
}

func (r *PostgresRepository) DeleteItemByID(
	ctx context.Context,
	itemID uuid.UUID,
	userID uuid.UUID,
) (Item, error) {
	const query = `
		DELETE FROM items
		WHERE id = $1 AND user_id = $2
		RETURNING id, title, description, status_id, type_id, rating, notes`

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
		&item.StatusID,
		&item.TypeID,
		&item.Rating,
		&item.Notes,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Item{}, ErrNotFound
	}

	if err != nil {
		return Item{}, fmt.Errorf(
			"delete item: %w",
			err,
		)
	}

	return item, nil
}

func (r *PostgresRepository) CreateItemStatus(
	ctx context.Context,
	userID uuid.UUID,
	name string,
) (ItemStatus, error) {
	const query = `
		INSERT INTO item_statuses (
			user_id,
			name
		)
		VALUES ($1, $2)
		RETURNING id, name
	`

	var itemStatus ItemStatus

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		name,
	).Scan(
		&itemStatus.ID,
		&itemStatus.Name,
	)

	if err != nil {
		return ItemStatus{}, fmt.Errorf(
			"create item status: %w",
			err,
		)
	}

	return itemStatus, nil
}

func (r *PostgresRepository) ListItemStatuses(
	ctx context.Context,
	userID uuid.UUID,
) ([]ItemStatus, error) {
	const query = `
		SELECT id, name, user_id, created_at
		FROM item_statuses
		WHERE user_id IS NULL OR user_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list item statuses: %w", err)
	}
	defer rows.Close()

	var statuses []ItemStatus
	for rows.Next() {
		var s ItemStatus

		if err := rows.Scan(&s.ID, &s.Name, &s.UserID, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan item status: %w", err)
		}

		statuses = append(statuses, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate item statuses: %w", err)
	}

	return statuses, nil
}

func (r *PostgresRepository) DeleteItemStatusByID(
	ctx context.Context,
	statusID uuid.UUID,
	userID uuid.UUID,
) (ItemStatus, error) {
	const query = `
		DELETE FROM item_statuses
		WHERE id = $1 AND user_id = $2
		RETURNING id, name
	`

	var itemStatus ItemStatus

	err := r.db.QueryRow(
		ctx,
		query,
		statusID,
		userID,
	).Scan(
		&itemStatus.ID,
		&itemStatus.Name,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return ItemStatus{}, ErrNotFound
	}

	if err != nil {
		return ItemStatus{}, fmt.Errorf(
			"delete item status: %w",
			err,
		)
	}

	return itemStatus, nil
}

func (r *PostgresRepository) CreateItemType(
	ctx context.Context,
	userID uuid.UUID,
	name string,
) (ItemType, error) {
	const query = `
		INSERT INTO item_types (
			user_id,
			name
		)
		VALUES ($1, $2)
		RETURNING id, name
	`

	var itemType ItemType

	err := r.db.QueryRow(
		ctx,
		query,
		userID,
		name,
	).Scan(
		&itemType.ID,
		&itemType.Name,
	)

	if err != nil {
		return ItemType{}, fmt.Errorf(
			"create item type: %w",
			err,
		)
	}

	return itemType, nil
}

func (r *PostgresRepository) ListItemTypes(
	ctx context.Context,
	userID uuid.UUID,
) ([]ItemType, error) {
	const query = `
		SELECT id, name, user_id, created_at
		FROM item_types
		WHERE user_id IS NULL OR user_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list item types: %w", err)
	}
	defer rows.Close()

	var types []ItemType
	for rows.Next() {
		var t ItemType

		if err := rows.Scan(&t.ID, &t.Name, &t.UserID, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan item type: %w", err)
		}

		types = append(types, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate item types: %w", err)
	}

	return types, nil
}

func (r *PostgresRepository) DeleteItemTypeByID(
	ctx context.Context,
	typeID uuid.UUID,
	userID uuid.UUID,
) (ItemType, error) {
	const query = `
		DELETE FROM item_types
		WHERE id = $1 AND user_id = $2
		RETURNING id, name
	`

	var itemType ItemType

	err := r.db.QueryRow(
		ctx,
		query,
		typeID,
		userID,
	).Scan(
		&itemType.ID,
		&itemType.Name,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return ItemType{}, ErrNotFound
	}

	if err != nil {
		return ItemType{}, fmt.Errorf(
			"delete item type: %w",
			err,
		)
	}

	return itemType, nil
}