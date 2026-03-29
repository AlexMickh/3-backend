package models

import (
	"time"

	"github.com/google/uuid"
)

type ProductSize string

const (
	SizeXS ProductSize = "xs"
	SizeS  ProductSize = "s"
	SizeM  ProductSize = "m"
	SizeL  ProductSize = "l"
	SizeXL ProductSize = "xl"
	Size52 ProductSize = "52"
	Size54 ProductSize = "54"
)

type Product struct {
	ID                uuid.UUID
	Category          Category
	Name              string
	Description       string
	Price             int
	Quantity          int
	ExistingSizes     []ProductSize
	ImageUrl          string
	PeicesSold        int
	Discount          int
	DiscountExpiresAt *time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type ProductCard struct {
	ID                uuid.UUID
	Name              string
	Price             int
	ImageUrl          string
	Discount          int
	DiscountExpiresAt *time.Time
}
