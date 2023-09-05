package infra

import (
	"context"
	"errors"
	"io"

	"github.com/olad5/go-cloud-backup-system/internal/domain"
)

var (
	ErrFileNotFound      = errors.New("file not found")
	ErrFolderNotFound    = errors.New("folder not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserNotAuthorized = errors.New("unauthorized")
)

type UserRepository interface {
	Ping(ctx context.Context) error
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserByUserId(ctx context.Context, userId string) (domain.User, error)
}

type FileRepository interface {
	Ping(ctx context.Context) error
	SaveFile(ctx context.Context, file domain.File) error
	GetFileByFileId(ctx context.Context, fileId string) (domain.File, error)
	GetFilesByFolderId(ctx context.Context, folderId string) ([]domain.File, error)
}

type FolderRepository interface {
	CreateFolder(ctx context.Context, folder domain.Folder) error
	GetFolderByFolderId(ctx context.Context, folderId string) (domain.Folder, error)
}

type FileStore interface {
	Ping(ctx context.Context) error
	SaveToFileStore(ctx context.Context, filename string, file io.Reader) (string, error)
}
