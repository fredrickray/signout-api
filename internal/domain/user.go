package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrValidation          = errors.New("validation failed")
	ErrRefreshTokenInvalid = errors.New("refresh token invalid")
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (t RefreshToken) IsActive(now time.Time) bool {
	return t.RevokedAt == nil && t.ExpiresAt.After(now)
}
