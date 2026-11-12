package catalog

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

type validationRepository struct {
	Repository
	calls  int
	userID uuid.UUID
	title  string
	rating *float64
	name   string
}

func (r *validationRepository) CreateItem(_ context.Context, userID uuid.UUID, title string, description *string, statusID, typeID *uuid.UUID, rating *float64, notes *string) (Item, error) {
	r.calls++
	r.userID, r.title, r.rating = userID, title, rating
	return Item{ID: uuid.New(), Title: title, Rating: rating}, nil
}

func (r *validationRepository) CreateItemStatus(_ context.Context, userID uuid.UUID, name string) (ItemStatus, error) {
	r.calls++
	r.userID, r.name = userID, name
	return ItemStatus{ID: uuid.New(), Name: name, UserID: &userID}, nil
}

func TestCreateItemValidatesTitleAndRating(t *testing.T) {
	zero, ten, negative, tooHigh := 0.0, 10.0, -0.1, 10.1
	tests := []struct {
		name    string
		title   string
		rating  *float64
		wantErr bool
	}{
		{name: "no rating", title: "Dune"},
		{name: "minimum rating", title: "Dune", rating: &zero},
		{name: "maximum rating", title: "Dune", rating: &ten},
		{name: "empty title", wantErr: true},
		{name: "whitespace title", title: " \t\n", wantErr: true},
		{name: "negative rating", title: "Dune", rating: &negative, wantErr: true},
		{name: "rating above maximum", title: "Dune", rating: &tooHigh, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &validationRepository{}
			userID := uuid.New()
			item, err := NewService(repo).CreateItem(context.Background(), userID, CreateItemRequest{Title: tt.title, Rating: tt.rating})
			if tt.wantErr {
				var validationErr *ValidationError
				if !errors.As(err, &validationErr) {
					t.Fatalf("expected ValidationError, got %v", err)
				}
				if repo.calls != 0 {
					t.Fatal("invalid item must not reach the repository")
				}
				return
			}
			if err != nil {
				t.Fatalf("CreateItem: %v", err)
			}
			if repo.calls != 1 || repo.userID != userID || repo.title != tt.title || repo.rating != tt.rating {
				t.Fatal("valid item was not passed to the repository with its owner and rating")
			}
			if item.ID == uuid.Nil || item.Title != tt.title || item.Rating != tt.rating {
				t.Fatalf("unexpected returned item: %+v", item)
			}
		})
	}
}

func TestCreateItemStatusValidatesName(t *testing.T) {
	for _, name := range []string{"", " \t\n", "In progress"} {
		t.Run(name, func(t *testing.T) {
			repo := &validationRepository{}
			userID := uuid.New()
			status, err := NewService(repo).CreateItemStatus(context.Background(), userID, name)
			if name != "In progress" {
				var validationErr *ValidationError
				if !errors.As(err, &validationErr) || repo.calls != 0 {
					t.Fatalf("invalid name: error = %v, repository calls = %d", err, repo.calls)
				}
				return
			}
			if err != nil || repo.calls != 1 || repo.userID != userID || repo.name != name {
				t.Fatalf("CreateItemStatus: error = %v, repository = %+v", err, repo)
			}
			if status.Name != name || status.UserID == nil || *status.UserID != userID {
				t.Fatalf("unexpected returned status: %+v", status)
			}
		})
	}
}
