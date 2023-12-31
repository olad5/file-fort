package infra

import (
	"context"
	"errors"
	"io"

	"github.com/google/uuid"
	"github.com/olad5/file-fort/internal/domain"
)

var (
	ErrFileNotFound      = errors.New("file not found")
	ErrFolderNotFound    = errors.New("folder not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrUserNotAuthorized = errors.New("unauthorized")
)

type UserRepository interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	GetUserByUserId(ctx context.Context, userId uuid.UUID) (domain.User, error)
}

type FileRepository interface {
	SaveFile(ctx context.Context, file domain.File) error
	MarkFileAsUnsafe(ctx context.Context, file domain.File) error
	GetFileByFileId(ctx context.Context, fileId uuid.UUID) (domain.File, error)
	GetFilesByFolderId(ctx context.Context, folderId uuid.UUID, pageNumber, rowsPerPage int) ([]domain.File, error)
}

type FolderRepository interface {
	CreateFolder(ctx context.Context, folder domain.Folder) error
	GetFolderByFolderId(ctx context.Context, folderId uuid.UUID) (domain.Folder, error)
}

type FileStore interface {
	Ping(ctx context.Context) error
	SaveToFileStore(ctx context.Context, filename string, file io.Reader) (string, error)
	GetDownloadUrl(ctx context.Context, key string) (string, error)
	DeleteFile(ctx context.Context, key string) error
}
