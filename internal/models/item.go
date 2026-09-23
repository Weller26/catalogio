package models

import "time"

type Item struct {
	ID int
	Title string
	Description string
	CreatedAt time.Time
	UpdatedAt time.Time
}