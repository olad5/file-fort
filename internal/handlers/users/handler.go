package handlers

import (
	"errors"

	"github.com/olad5/go-cloud-backup-system/internal/services/auth"
	"github.com/olad5/go-cloud-backup-system/internal/usecases/users"
)

type UserHandler struct {
	userService users.UserService
	authService auth.AuthService
}

func NewHandler(userService users.UserService, authService auth.AuthService) (*UserHandler, error) {
	if userService == (users.UserService{}) {
		return nil, errors.New("user service cannot be empty")
	}
	if authService == nil {
		return nil, errors.New("auth service cannot be empty")
	}

	return &UserHandler{userService, authService}, nil
}
