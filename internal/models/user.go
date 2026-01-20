package models

type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type User struct {
	ID              int64
	Email           string
	Phone           string
	Password        string
	Role            UserRole
	IsEmailVerified bool
}
