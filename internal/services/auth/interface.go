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

type ctxKey int

const jwtKey ctxKey = 1

func Set(ctx context.Context, jwt JWTClaims) context.Context {
	return context.WithValue(ctx, jwtKey, jwt)
}

func Get(ctx context.Context) (JWTClaims, bool) {
	v, ok := ctx.Value(jwtKey).(JWTClaims)
	return v, ok
}

type AuthService interface {
	IsUserAdmin(ctx context.Context, claims JWTClaims) bool
	DecodeJWT(ctx context.Context, tokenString string) (JWTClaims, error)
	GenerateJWT(ctx context.Context, user domain.User) (string, error)
	IsUserLoggedIn(ctx context.Context, authHeader, userId string) bool
}
