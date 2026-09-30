package items

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type Service struct {
	repo *Repository	
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateItem (
	ctx context.Context,
	userID uuid.UUID,
	req CreateItemRequest,
) (Item, error) {
	if req.Rating != nil {
		if *req.Rating < 0 || *req.Rating > 10 {
			return Item{}, errors.New("rating must be between 0 and 10")
		}
	}

	return s.repo.CreateItem(
		ctx,
		userID,
		req.Title,
		req.Description,
		req.Status,
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
	return s.repo.UpdateItemByID(
		ctx,
		itemID,
		userID,
		req.Title,
		req.Description,
		req.Status,
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