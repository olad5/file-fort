package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/olad5/go-cloud-backup-system/config"
	"github.com/olad5/go-cloud-backup-system/internal/domain"
	"github.com/olad5/go-cloud-backup-system/internal/infra"
)

type RedisAuthService struct {
	Cache     infra.Cache
	SecretKey string
}

var ErrInvalidToken = errors.New("invalid token")

const (
	JWT_HASH_NAME       = "activeJwtClients"
	SessionTTLInMinutes = 10
)

func NewRedisAuthService(ctx context.Context, cache infra.Cache, configurations *config.Configurations) (*RedisAuthService, error) {
	if cache == nil {
		return nil, fmt.Errorf("failed to initialize auth service, cache is nil")
	}

	if err := cache.Ping(ctx); err != nil {
		return nil, err
	}

	return &RedisAuthService{cache, configurations.JwtSecretKey}, nil
}

func (r *RedisAuthService) GenerateJWT(ctx context.Context, user domain.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(time.Minute * SessionTTLInMinutes).Unix(),
	})
	tokenString, err := token.SignedString([]byte(r.SecretKey))
	if err != nil {
		return "", errors.New("Error generating JWT token")
	}

	err = r.Cache.SetOne(ctx, constructUserIdKey(user.ID.String()), tokenString)
	if err != nil {
		return "", errors.New("Error generating JWT token")
	}
	return tokenString, nil
}

func (r *RedisAuthService) DecodeJWT(ctx context.Context, tokenString string) (string, error) {
	if tokenString == "" {
		return "", ErrInvalidToken
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["sub"])
		}
		return []byte(r.SecretKey), nil
	})
	if err != nil {
		return "", errors.New("error decoding jwt")
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			return "", errors.New("expired token")
		}
		userId, ok := claims["sub"]
		if ok == true && userId != nil {
			return userId.(string), nil
		}
		return "", ErrInvalidToken
	} else {
		return "", ErrInvalidToken
	}
}

func (r *RedisAuthService) IsUserLoggedIn(ctx context.Context, authHeader, userId string) bool {
	token := strings.Split(authHeader, " ")[1]
	cachedToken, err := r.Cache.GetOne(ctx, constructUserIdKey(userId))
	if err != nil || cachedToken != token {
		return false
	}
	return true
}

func (r *RedisAuthService) IsUserAdmin(ctx context.Context, authHeader string) (bool, error) {
	return false, nil
}

func constructUserIdKey(key string) string {
	return JWT_HASH_NAME + key
}
