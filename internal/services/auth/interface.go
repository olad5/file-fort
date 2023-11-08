package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/olad5/file-fort/internal/domain"
)

type JWTClaims struct {
	ID    uuid.UUID
	Role  domain.Role
	Email string
}

type AuthService interface {
	IsUserAdmin(ctx context.Context, claims JWTClaims) bool
	DecodeJWT(ctx context.Context, tokenString string) (JWTClaims, error)
	GenerateJWT(ctx context.Context, user domain.User) (string, error)
	IsUserLoggedIn(ctx context.Context, authHeader, userId string) bool
}
