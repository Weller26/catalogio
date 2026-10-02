package catalog

import (
	"time"
	"github.com/google/uuid"
)

type Item struct {
	ID uuid.UUID `json:"id"`
	
	Title string `json:"title"`
	Description *string `json:"description,omitempty"`
	
	StatusID *uuid.UUID `json:"status_id,omitempty"`
	TypeID *uuid.UUID `json:"type_id,omitempty"`

	Rating *float64 `json:"rating,omitempty"`
	Notes *string `json:"notes,omitempty"`
	
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateItemRequest struct {
	Title string `json:"title"`
	Description *string `json:"description,omitempty"`
	StatusID *uuid.UUID `json:"status_id,omitempty"`
	TypeID *uuid.UUID `json:"type_id,omitempty"`
	Rating *float64 `json:"rating,omitempty"`
	Notes *string `json:"notes,omitempty"`
}

type UpdateItemRequest struct {
	Title string `json:"title"`
	Description *string `json:"description,omitempty"`
	StatusID *uuid.UUID `json:"status_id,omitempty"`
	TypeID *uuid.UUID `json:"type_id,omitempty"`
	Rating *float64 `json:"rating,omitempty"`
	Notes *string `json:"notes,omitempty"`
}

type ItemStatus struct {
	ID uuid.UUID `json:"id"`
	Name string `json:"name"`
	UserID *uuid.UUID `json:"user_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ItemType struct {
	ID uuid.UUID `json:"id"`
	Name string `json:"name"`
	UserID *uuid.UUID `json:"user_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}