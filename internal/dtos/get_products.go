package dtos

import (
	"time"

	"github.com/google/uuid"
)

type GetProductsRequest struct {
	Page       int
	Popularity bool
	Price      int
	CategoryID string
	Search     string
}

type GetProductsResponse struct {
	Products []Product `json:"products"`
}

type Product struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	Price             int        `json:"price"`
	ImageUrl          string     `json:"image_url"`
	Discount          int        `json:"discount,omitempty"`
	DiscountExpiresAt *time.Time `json:"discount_expires_at,omitempty"`
}
