package dtos

import "time"

type ProductByIdResponse struct {
	ID                int64      `json:"id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	Price             int        `json:"price"`
	Quantity          int        `json:"quantity"`
	ExistingSizes     []string   `json:"existing_sizes"`
	ImageUrl          string     `json:"image_url"`
	Discount          int        `json:"discount,omitempty"`
	DiscountExpiresAt *time.Time `json:"discount_expires_at,omitempty"`
	Category          struct {
		ID   int64
		Name string
	} `json:"category"`
}
