package models

import (
	"time"
)

type Session struct {
	Token          string
	UserID         int64
	Role           UserRole
	ExpiresAtField time.Time
}

func (s Session) ExpiresAt() time.Time {
	return s.ExpiresAtField
}
