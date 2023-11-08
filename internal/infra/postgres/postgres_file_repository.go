package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/olad5/file-fort/internal/domain"
	"github.com/olad5/file-fort/internal/infra"
)

type PostgresFileRepository struct {
	connection *sqlx.DB
}

func NewPostgresFileRepo(ctx context.Context, connection *sqlx.DB) (*PostgresFileRepository, error) {
	if connection == nil {
		return &PostgresFileRepository{}, fmt.Errorf("Failed to create PostgresFileRepository:  connection is nil")
	}

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
	err := p.connection.Get(&file, "SELECT * FROM files WHERE id=$1 AND is_unsafe=false", fileId)
	if err != nil {
		if err == ErrRecordNotFound {
			return domain.File{}, infra.ErrFileNotFound
		}
		return domain.File{}, fmt.Errorf("error getting file :%w", err)
	}

	return toDomainFile(file), nil
}

func (p *PostgresFileRepository) GetFilesByFolderId(ctx context.Context, folderId uuid.UUID, pageNumber, rowsPerPage int) ([]domain.File, error) {
	offset := (pageNumber - 1) * rowsPerPage

	var files []SqlxFile

	query := fmt.Sprintf(`
    SELECT * FROM files WHERE folder_id =$1 AND is_unsafe=false
    OFFSET %d ROWS FETCH NEXT %d ROWS ONLY
	`, offset, rowsPerPage)

	err := p.connection.Select(&files, query, folderId)
	if err != nil {
		return []domain.File{}, fmt.Errorf("error getting files :%w", err)
	}

	result := []domain.File{}
	for _, element := range files {
		result = append(result, toDomainFile(element))
	}

	return result, nil
}

func (p *PostgresFileRepository) MarkFileAsUnsafe(ctx context.Context, file domain.File) error {
	file.UpdatedAt = time.Now()
	file.IsUnsafe = true

	const query = `UPDATE files SET is_unsafe=:is_unsafe, updated_at=:updated_at WHERE id=:id`
	_, err := p.connection.NamedExec(query, toSqlxFile(file))
	if err != nil {
		return fmt.Errorf("error marking file as unsafe in the db: %w", err)
	}
	return nil
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
	OwnerId      uuid.UUID `db:"owner_id"`
	FolderId     uuid.UUID `db:"folder_id"`
	FileStoreKey string    `db:"file_store_key"`
	FileSize     int64     `db:"file_size"`
	IsUnsafe     bool      `db:"is_unsafe"`
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
		IsUnsafe:     f.IsUnsafe,
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
		IsUnsafe:     f.IsUnsafe,
		CreatedAt:    f.CreatedAt,
		UpdatedAt:    f.UpdatedAt,
	}
}
