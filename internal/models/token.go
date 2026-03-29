package models

import (
	"time"

	"github.com/google/uuid"
)

type TokenType string

const (
	TokenTypeValidateEmail  TokenType = "validate-email"
	TokenTypeChangePassword TokenType = "change-password"
)

type Token struct {
	Token     string
	UserID    uuid.UUID
	Type      TokenType
	ExpiresAt time.Time
}
