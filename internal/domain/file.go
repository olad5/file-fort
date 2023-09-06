package domain

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID           uuid.UUID
	FileName     string
	OwnerId      string
	FolderId     string
	FileStoreKey string
	FileSize     int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
