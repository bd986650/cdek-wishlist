package model

import "time"

type Wishlist struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	EventDate   string    `json:"event_date" db:"event_date"`
	Token       string    `json:"token" db:"token"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	Items []Item `json:"items,omitempty"`
}

type CreateWishlistRequest struct {
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description" validate:"max=1000"`
	EventDate   string `json:"event_date" validate:"omitempty"`
}

type UpdateWishlistRequest struct {
	Title       *string `json:"title" validate:"omitempty,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	EventDate   *string `json:"event_date" validate:"omitempty"`
}
