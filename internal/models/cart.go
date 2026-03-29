package models

import (
	"time"

	"github.com/google/uuid"
)

type CartItem struct {
	ID                uuid.UUID
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
