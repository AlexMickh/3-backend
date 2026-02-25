package dtos

import "time"

type CartItem struct {
	ID                int64
	Name              string
	Price             int
	ImageUrl          string
	Discount          int
	DiscountExpiresAt *time.Time
	Quantity          int
}

type GetCartResponse struct {
	Products []*CartItem
	Price    int
}
