package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser  Role = "regular"
	RoleAdmin Role = "admin"
)

type User struct {
	ID        uuid.UUID
	Email     string
	FirstName string
	LastName  string
	Password  string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}
