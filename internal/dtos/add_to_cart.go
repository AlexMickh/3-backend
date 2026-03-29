package dtos

type AddToCartRequest struct {
	UserId    string `validate:"required,uuid"`
	ProductId string `path:"product_id" validate:"required,uuid"`
}

type AddToCartResponse struct {
	ID string `json:"id"`
}
