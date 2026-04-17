package model

import "time"

type Item struct {
	ID         int64     `json:"id" db:"id"`
	WishlistID int64     `json:"wishlist_id" db:"wishlist_id"`
	Name       string    `json:"name" db:"name"`
	Description string   `json:"description" db:"description"`
	URL        string    `json:"url" db:"url"`
	Priority   int       `json:"priority" db:"priority"`
	IsReserved bool      `json:"is_reserved" db:"is_reserved"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type CreateItemRequest struct {
	Name        string `json:"name" validate:"required,max=255"`
	Description string `json:"description" validate:"max=1000"`
	URL         string `json:"url" validate:"omitempty,url,max=2048"`
	Priority    int    `json:"priority" validate:"min=0,max=10"`
}

type UpdateItemRequest struct {
	Name        *string `json:"name" validate:"omitempty,max=255"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	URL         *string `json:"url" validate:"omitempty,url,max=2048"`
	Priority    *int    `json:"priority" validate:"omitempty,min=0,max=10"`
}
