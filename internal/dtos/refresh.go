package dtos

type RefreshRequest struct {
	RefreshToken string `validate:"required,len=64"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}
