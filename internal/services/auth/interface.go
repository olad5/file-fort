package auth

import (
	"context"

	"github.com/olad5/file-fort/internal/domain"
)

type AuthService interface {
	IsUserAdmin(ctx context.Context, authHeader string) (bool, error)
	DecodeJWT(ctx context.Context, tokenString string) (string, error)
	GenerateJWT(ctx context.Context, user domain.User) (string, error)
	IsUserLoggedIn(ctx context.Context, authHeader, userId string) bool
}
