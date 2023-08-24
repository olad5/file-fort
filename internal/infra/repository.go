package infra

import (
	"context"

	"github.com/olad5/go-cloud-backup-system/internal/domain"
)

var ErrRecordNotFound = "No Record found"

type UserRepository interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
}
