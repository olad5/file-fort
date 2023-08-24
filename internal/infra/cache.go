package infra

import (
	"context"
)

type Cache interface {
	SetOne(ctx context.Context, key, value string) error
	GetOne(ctx context.Context, key string) (string, error)
	DeleteOne(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}
