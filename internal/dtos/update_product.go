package dtos

import (
	"mime/multipart"
	"time"
)

type UpdateProductRequest struct {
	ID                int64      `validate:"required"`
	Name              string     `validate:"min=5"`
	Description       string     `validate:"min=5"`
	Price             int        `validate:"gt=0"`
	Quantity          int        `validate:"gte=0"`
	ExistingSizes     []string   `validate:"gt=0,dive,oneof=xs s m l xl 52 54"`
	Discount          int        `validate:"gte=0"`
	DiscountExpiresAt *time.Time `validate:"datetime"`
	Image             multipart.File
}
