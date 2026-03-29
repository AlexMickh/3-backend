package dtos

import "time"

type ProductByIdResponse struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Description       string     `json:"description"`
	Price             int        `json:"price"`
	Quantity          int        `json:"quantity"`
	ExistingSizes     []string   `json:"existing_sizes"`
	ImageUrl          string     `json:"image_url"`
	Discount          int        `json:"discount,omitempty"`
	DiscountExpiresAt *time.Time `json:"discount_expires_at,omitempty"`
	Category          struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
}
