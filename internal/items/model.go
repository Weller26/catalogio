package items

import (
	"time"
	"github.com/google/uuid"
)

type Item struct {
	ID uuid.UUID `json:"id"`
	
	Title *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	
	Status *string `json:"status,omitempty"`
	Rating *float64 `json:"rating,omitempty"`
	Notes *string `json:"notes,omitempty"`
	
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateItemRequest struct {
	Title *string `json:"title"`
	Description *string `json:"description"`
	Status *string `json:"status"`
	Rating *float64 `json:"rating"`
	Notes *string `json:"notes"`
}

type UpdateItemRequest struct {
	Title *string `json:"title"`
	Description *string `json:"description"`
	Status *string `json:"status"`
	Rating *float64 `json:"rating"`
	Notes *string `json:"notes"`
}