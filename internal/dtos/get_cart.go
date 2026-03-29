package dtos

import "time"

type CartItem struct {
	ID                string     `json:"id"`
	Name              string     `json:"name"`
	Price             int        `json:"price"`
	ImageUrl          string     `json:"image_url"`
	Discount          int        `json:"discount"`
	DiscountExpiresAt *time.Time `json:"discount_expires_at"`
	Quantity          int        `json:"quantity"`
}

type GetCartResponse struct {
	Products []*CartItem `json:"products"`
	Price    int         `json:"price"`
}
