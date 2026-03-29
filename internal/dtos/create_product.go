package dtos

import "mime/multipart"

type CreateProductRequest struct {
	Name          string         `validate:"required,min=3"`
	Description   string         `validate:"required,min=5"`
	Price         int            `validate:"required,gt=0"`
	CategoryID    string         `validate:"required"`
	Quantity      int            `validate:"required,gte=0"`
	ExistingSizes []string       `validate:"required,gt=0,dive,oneof=xs s m l xl 52 54"`
	Image         multipart.File `validate:"required"`
}

type CreateProductResponse struct {
	ID string `json:"id"`
}
