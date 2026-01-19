package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	Token          string
	UserID         uuid.UUID
	ExpiresAtField time.Time
}

func (s Session) ExpiresAt() time.Time {
	return s.ExpiresAtField
}
