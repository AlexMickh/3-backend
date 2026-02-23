package dtos

import "time"

type GetProductsRequest struct {
	Page       int
	Popularity bool
	Price      int
	CategoryID int64
	Search     string
}

type GetProductsResponse struct {
	Products []Product `json:"products"`
}

type Product struct {
	ID                int64      `json:"id"`
	Name              string     `json:"name"`
	Price             int        `json:"price"`
	ImageUrl          string     `json:"image_url"`
	Discount          int        `json:"discount,omitempty"`
	DiscountExpiresAt *time.Time `json:"discount_expires_at,omitempty"`
}
