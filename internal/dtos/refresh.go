package dtos

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,len=32"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
