package models

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

type Cart struct {
	Products []*CartItem
	Price    int
}
