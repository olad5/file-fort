package domain

import "github.com/google/uuid"

// TODO:TODO: user will have role
type UserRole string

// TODO:TODO: i hate these role names
const (
	RegularUserRole UserRole = "regular"
	AdminUserRole   UserRole = "admin"
)

type User struct {
	ID        uuid.UUID
	Email     string
	FirstName string
	LastName  string
	Password  string
	Role      UserRole
}
