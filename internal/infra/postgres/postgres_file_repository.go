package postgres

import (
	"context"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/olad5/go-cloud-backup-system/internal/domain"
	"github.com/olad5/go-cloud-backup-system/internal/infra"
)

type PostgresFileRepository struct {
	connection *sqlx.DB
}

func NewPostgresFileRepo(ctx context.Context, connection *sqlx.DB) (*PostgresFileRepository, error) {
	if connection == nil {
		return &PostgresFileRepository{}, fmt.Errorf("Failed to create PostgresFileRepository:  connection is nil")
	}

	const fileSchema = `
  CREATE TABLE IF NOT EXISTS files(
      id UUID PRIMARY KEY,
      file_name varchar(255) NOT NULL,
      owner_id UUID NOT NULL REFERENCES users(id), 
      folder_id UUID NOT NULL REFERENCES folders(id), 
      file_store_key varchar(255) NOT NULL,
      file_size INTEGER NOT NULL,
      "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
      "updated_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP
  );

`
	connection.MustExec(fileSchema)
	return &PostgresFileRepository{connection: connection}, nil
}

func (p *PostgresFileRepository) SaveFile(ctx context.Context, file domain.File) error {
	const query = `
    INSERT INTO files 
      (id, file_name, owner_id, folder_id, file_store_key, file_size) 
    VALUES 
      (:id, :file_name, :owner_id, :folder_id, :file_store_key, :file_size)
  `

	_, err := p.connection.NamedExec(query, toSqlxFile(file))
	if err != nil {
		return fmt.Errorf("error saving file in the db: %w", err)
	}
	return nil
}

func (p *PostgresFileRepository) GetFileByFileId(ctx context.Context, fileId uuid.UUID) (domain.File, error) {
	var file SqlxFile
	err := p.connection.Get(&file, "SELECT * FROM files WHERE id=$1", fileId)
	if err != nil {
		if err == ErrRecordNotFound {
			return domain.File{}, infra.ErrFileNotFound
		}
		return domain.File{}, fmt.Errorf("error getting file :%w", err)
	}

	return toDomainFile(file), nil
}

func (p *PostgresFileRepository) GetFilesByFolderId(ctx context.Context, folderId string) ([]domain.File, error) {
	var files []SqlxFile

	err := p.connection.Select(&files, "SELECT * FROM files WHERE folder_id=$1", folderId)
	if err != nil {
		return []domain.File{}, fmt.Errorf("error getting files :%w", err)
	}

	result := []domain.File{}
	for _, element := range files {
		result = append(result, toDomainFile(element))
	}

	return result, nil
}

func (p *PostgresFileRepository) Ping(ctx context.Context) error {
	err := p.connection.Ping()
	if err != nil {
		return fmt.Errorf("Failed to Ping PostgresFileRepository:  %w", err)
	}

	return nil
}

type SqlxFile struct {
	ID           uuid.UUID `db:"id"`
	FileName     string    `db:"file_name"`
	OwnerId      string    `db:"owner_id"`
	FolderId     string    `db:"folder_id"`
	FileStoreKey string    `db:"file_store_key"`
	FileSize     int64     `db:"file_size"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

func toDomainFile(f SqlxFile) domain.File {
	return domain.File{
		ID:           f.ID,
		FileName:     f.FileName,
		OwnerId:      f.OwnerId,
		FolderId:     f.FolderId,
		FileStoreKey: f.FileStoreKey,
		FileSize:     f.FileSize,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
	}
}

func toSqlxFile(f domain.File) SqlxFile {
	return SqlxFile{
		ID:           f.ID,
		FileName:     f.FileName,
		OwnerId:      f.OwnerId,
		FolderId:     f.FolderId,
		FileStoreKey: f.FileStoreKey,
		FileSize:     f.FileSize,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
	}
}
