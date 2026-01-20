package models

import (
	"time"
)

type TokenType string

const (
	TokenTypeValidateEmail  TokenType = "validate-email"
	TokenTypeChangePassword TokenType = "change-password"
)

type Token struct {
	Token     string
	UserID    int64
	Type      TokenType
	ExpiresAt time.Time
}
