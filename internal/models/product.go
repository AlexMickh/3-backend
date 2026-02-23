package models

import "time"

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
	ID                int64
	Name              string
	Description       string
	Category          Category
	Price             int
	Quantity          int
	ExistingSizes     []ProductSize
	ImageUrl          string
	PeicesSold        int
	Discount          int
	DiscountExpiresAt *time.Time
	CreatedAt         time.Time
}

type ProductCard struct {
	ID                int64
	Name              string
	Price             int
	ImageUrl          string
	Discount          int
	DiscountExpiresAt *time.Time
}
