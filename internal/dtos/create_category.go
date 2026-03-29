package dtos

type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required,min=4"`
}

type CreateCategoryResponse struct {
	ID string `json:"id"`
}
