package dtos

type RegisterDto struct {
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required,e164"`
	Password string `json:"password" validate:"required,min=4"`
}

type RegisterResponse struct {
	ID int64 `json:"id"`
}
