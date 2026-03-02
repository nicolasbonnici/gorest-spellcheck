package spellcheck

import (
	"time"

	"github.com/google/uuid"
)

type Item struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	Active      bool       `json:"active" db:"active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

type CreateItemRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"max=1000"`
	Active      bool   `json:"active"`
}

type UpdateItemRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
	Active      *bool   `json:"active,omitempty"`
}

type ListItemsResponse struct {
	Items  []Item `json:"items"`
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}
