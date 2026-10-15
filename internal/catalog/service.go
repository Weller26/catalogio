package catalog

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type Repository interface {
	CreateItem(
		ctx context.Context,
		userID uuid.UUID, 
		title string, description *string,
		statusID, typeID *uuid.UUID,
		rating *float64, notes *string,
	) (Item, error)

	ListItems(ctx context.Context, userID uuid.UUID) ([]Item, error)

	GetItemByID(ctx context.Context, itemID, userID uuid.UUID) (Item, error)

	UpdateItemByID(
		ctx context.Context, itemID, userID uuid.UUID,
		title string, description *string,
		statusID, typeID *uuid.UUID,
		rating *float64, notes *string,
	) (Item, error)

	DeleteItemByID(ctx context.Context, itemID, userID uuid.UUID) (Item, error)

	CreateItemStatus(ctx context.Context, userID uuid.UUID, name string) (ItemStatus, error)

	ListItemStatuses(ctx context.Context, userID uuid.UUID) ([]ItemStatus, error)

	DeleteItemStatusByID(ctx context.Context, statusID, userID uuid.UUID) (ItemStatus, error)

	CreateItemType(ctx context.Context, userID uuid.UUID, name string) (ItemType, error)

	ListItemTypes(ctx context.Context, userID uuid.UUID) ([]ItemType, error)

	DeleteItemTypeByID(ctx context.Context, typeID, userID uuid.UUID) (ItemType, error)
}

type Service struct {
	repo Repository	
}

func NewService(repo Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateItem(
	ctx context.Context,
	userID uuid.UUID,
	req CreateItemRequest,
) (Item, error) {
	// TODO: create validate function
	if err := validateItemRequest(req.Title, req.Rating); err != nil {
		return Item{}, err
	}

	return s.repo.CreateItem(
		ctx,
		userID,
		req.Title,
		req.Description,
		req.StatusID,
		req.TypeID,
		req.Rating,
		req.Notes,
	)
}

func (s *Service) ListItems(
	ctx context.Context,
	userID uuid.UUID,
) ([]Item, error) {
	return s.repo.ListItems(ctx, userID)
}

func (s *Service) GetItemByID(
	ctx context.Context,
	itemID uuid.UUID,
	userID uuid.UUID,
) (Item, error) {
	return s.repo.GetItemByID(ctx, itemID, userID)
}

func (s *Service) UpdateItemByID(
	ctx context.Context,
	itemID uuid.UUID,
	userID uuid.UUID,
	req UpdateItemRequest,
) (Item, error) {
	if err := validateItemRequest(req.Title, req.Rating); err != nil {
		return Item{}, err
	}

	return s.repo.UpdateItemByID(
		ctx,
		itemID,
		userID,
		req.Title,
		req.Description,
		req.StatusID,
		req.TypeID,
		req.Rating,
		req.Notes,
	)
}

func (s *Service) DeleteItemByID(
	ctx context.Context,
	itemID uuid.UUID,
	userID uuid.UUID,
) (Item, error) {
	return s.repo.DeleteItemByID(ctx, itemID, userID)
}

func (s *Service) CreateItemStatus(
	ctx context.Context,
	userID uuid.UUID,
	name string,
) (ItemStatus, error) {
	if err := validateRequestWithName(name); err != nil {
		return ItemStatus{}, err
	}

	return s.repo.CreateItemStatus(
		ctx,
		userID,
		name,
	)
}

func (s *Service) ListItemStatuses(
	ctx context.Context,
	userID uuid.UUID,
) ([]ItemStatus, error) {
	return s.repo.ListItemStatuses(ctx, userID)
}

func (s *Service) DeleteItemStatusByID(
	ctx context.Context,
	statusID uuid.UUID,
	userID uuid.UUID,
) (ItemStatus, error) {
	return s.repo.DeleteItemStatusByID(ctx, statusID, userID)
}

func (s *Service) CreateItemType(
	ctx context.Context,
	userID uuid.UUID,
	name string,
) (ItemType, error) {
	if err := validateRequestWithName(name); err != nil {
		return ItemType{}, err
	}

	return s.repo.CreateItemType(
		ctx,
		userID,
		name,
	)
}

func (s *Service) ListItemTypes(
	ctx context.Context,
	userID uuid.UUID,
) ([]ItemType, error) {
	return s.repo.ListItemTypes(ctx, userID)
}

func (s *Service) DeleteItemTypeByID(
	ctx context.Context,
	typeID uuid.UUID,
	userID uuid.UUID,
) (ItemType, error) {
	return s.repo.DeleteItemTypeByID(ctx, typeID, userID)
}

func validateItemRequest(title string, rating *float64) error {
	if strings.TrimSpace(title) == "" {
		return NewValidationError("title cannot be empty")
	}

	if rating != nil && (*rating < 0 || *rating > 10) {
		return NewValidationError(
			"rating must be between 0 and 10, got %.1f", *rating,
		)
	}

	return nil
}

func validateRequestWithName(name string) error {
	if strings.TrimSpace(name) == "" {
		return NewValidationError("name cannot be empty")
	}

	return nil
}