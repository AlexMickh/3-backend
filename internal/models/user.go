package models

import "github.com/google/uuid"

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type User struct {
	ID              uuid.UUID
	Email           string
	Phone           string
	Password        string
	Role            UserRole
	IsEmailVerified bool
}
