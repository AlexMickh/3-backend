package errs

import "errors"

var (
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrUserNotFound          = errors.New("user with this email or password not found")
	ErrEmailNotVerified      = errors.New("email not verify")
	ErrTokenNotFound         = errors.New("token not found")
	ErrTokenExpired          = errors.New("token expired")
	ErrSessionNotFound       = errors.New("session not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrCategoryNotFound      = errors.New("category not found")
	ErrUnsupportedImageType  = errors.New("unsupported image type (only png)")
	ErrProductAlreadyExists  = errors.New("producct already exists")
)
