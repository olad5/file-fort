package domain

import (
	"time"

	"github.com/google/uuid"
)

type Folder struct {
	ID         uuid.UUID
	FolderName string
	OwnerId    uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
