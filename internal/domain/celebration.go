package domain

import (
	"time"

	"github.com/google/uuid"
)

type ShirtTemplate struct {
	ID           uuid.UUID `json:"id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Category     string    `json:"category"`
	ModelURL     string    `json:"model_url"`
	PreviewImage string    `json:"preview_image"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Celebration struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	Slug            string    `json:"slug"`
	DisplayName     string    `json:"display_name"`
	School          string    `json:"school"`
	Faculty         string    `json:"faculty"`
	ClassOf         string    `json:"class_of"`
	CelebrationType string    `json:"celebration_type"` // graduation | nysc
	ShirtTemplateID uuid.UUID `json:"shirt_template_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CelebrationPublic is returned on the share page with shirt embedded.
type CelebrationPublic struct {
	Celebration
	Shirt *ShirtTemplate `json:"shirt"`
}
