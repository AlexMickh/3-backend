package dtos

type RegisterDto struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=4"`
}

type RegisterResponse struct {
	ID int64 `json:"id"`
}
