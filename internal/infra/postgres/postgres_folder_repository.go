package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/olad5/go-cloud-backup-system/internal/domain"
	"github.com/olad5/go-cloud-backup-system/internal/infra"
)

type PostgresFolderRepository struct {
	connection *sqlx.DB
}

func NewPostgresFolderRepo(ctx context.Context, connection *sqlx.DB) (*PostgresFolderRepository, error) {
	if connection == nil {
		return &PostgresFolderRepository{}, fmt.Errorf("Failed to create PostgresFolderRepository: connection is nil")
	}
	const folderSchema = `
  CREATE TABLE IF NOT EXISTS folders(
      id UUID PRIMARY KEY,
      folder_name varchar(255) NOT NULL,
      owner_id UUID NOT NULL REFERENCES users(id), 
      "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
      "updated_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
  );

`

	connection.MustExec(folderSchema)
	return &PostgresFolderRepository{connection: connection}, nil
}

func (p *PostgresFolderRepository) CreateFolder(ctx context.Context, folder domain.Folder) error {
	const query = `
    INSERT INTO folders 
      (id, folder_name, owner_id) 
    VALUES 
      (:id, :folder_name, :owner_id)
  `

	_, err := p.connection.NamedExec(query, toSqlxFolder(folder))
	if err != nil {
		return fmt.Errorf("error creating folder in the db: %w", err)
	}
	return nil
}

func (p *PostgresFolderRepository) GetFolderByFolderId(ctx context.Context, folderId uuid.UUID) (domain.Folder, error) {
	var folder SqlxFolder
	err := p.connection.Get(&folder, "SELECT * FROM folders WHERE id=$1", folderId)
	if err != nil {
		if err == ErrRecordNotFound {
			return domain.Folder{}, infra.ErrFolderNotFound
		}
		return domain.Folder{}, fmt.Errorf("error getting folder :%w", err)
	}

	return toDomainFolder(folder), nil
}

func (p *PostgresFolderRepository) Ping(ctx context.Context) error {
	err := p.connection.Ping()
	if err != nil {
		return fmt.Errorf("Failed to Ping PostgresFolderRepository:  %w", err)
	}

	return nil
}

type SqlxFolder struct {
	ID         uuid.UUID `db:"id"`
	FolderName string    `db:"folder_name"`
	OwnerId    uuid.UUID `db:"owner_id"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

func toDomainFolder(f SqlxFolder) domain.Folder {
	return domain.Folder{
		ID:         f.ID,
		FolderName: f.FolderName,
		OwnerId:    f.OwnerId,
		CreatedAt:  f.CreatedAt,
		UpdatedAt:  f.UpdatedAt,
	}
}

func toSqlxFolder(f domain.Folder) SqlxFolder {
	return SqlxFolder{
		ID:         f.ID,
		FolderName: f.FolderName,
		OwnerId:    f.OwnerId,
		CreatedAt:  f.CreatedAt,
		UpdatedAt:  f.UpdatedAt,
	}
}
