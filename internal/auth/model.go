package auth

import "github.com/google/uuid"

type User struct {
	ID uuid.UUID `json:"id"`
	Email string `json:"email"`
}

type RegisterRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email string `json:"email"`
	Password string `json:"password"`
}