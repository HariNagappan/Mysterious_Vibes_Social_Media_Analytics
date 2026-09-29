package auth

import (
	"errors"
	"time"
)

// Domain errors surfaced to handlers.
var (
	// ErrEmailTaken is returned when registration hits a duplicate email.
	ErrEmailTaken = errors.New("auth: email already registered")
	// ErrInvalidCredentials covers wrong password and unknown email alike.
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
)

// User is the domain view of an authenticated principal (the password hash
// never leaves the service layer).
type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// RegisterRequest is the POST /auth/register payload.
type RegisterRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=120"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=128"`
	Role     string `json:"role" binding:"omitempty,oneof=admin analyst viewer"`
}

// LoginRequest is the POST /auth/login payload.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is returned by register and login.
type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      User      `json:"user"`
}
