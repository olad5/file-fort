package handlers

import (
	"github.com/olad5/file-fort/internal/domain"
)

func ToResponseFile(file domain.File) map[string]interface{} {
	return map[string]interface{}{
		"id":              file.ID,
		"file_name":       file.FileName,
		"file_size":       file.FileSize,
		"file_store_link": file.FileStoreKey,
		"owner_id":        file.OwnerId,
		"folder_id":       file.FolderId,
		"created_at":      file.CreatedAt,
		"updated_at":      file.UpdatedAt,
	}
}
