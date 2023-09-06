package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/olad5/go-cloud-backup-system/internal/infra"
	appErrors "github.com/olad5/go-cloud-backup-system/pkg/errors"

	response "github.com/olad5/go-cloud-backup-system/pkg/utils"
)

const MAX_UPLOAD_SIZE = 1024 * 1024 * 200 // 1MB * 200

func (f FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE); err != nil {
		response.ErrorResponse(w, "The file you are trying to upload exceeds the maximum allowed size of 200MB.", http.StatusBadRequest)
		return
	}

	folderId := r.FormValue("folder_id")

	file, handler, err := r.FormFile("file")
	if err != nil {
		response.ErrorResponse(w, "Error retrieving file, please try again", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ctx := context.WithValue(r.Context(), "folderId", folderId)

	uploadedFile, err := f.fileService.UploadFile(ctx, file, handler)
	if err != nil {
		switch {
		case errors.Is(err, infra.ErrFolderNotFound):
			response.ErrorResponse(w, "folder does not exist", http.StatusNotFound)
			return
		default:
			response.ErrorResponse(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError)
			return
		}
	}
	response.SuccessResponse(w, "file uploaded successfully",
		map[string]interface{}{
			"id":              uploadedFile.ID,
			"file_name":       uploadedFile.FileName,
			"file_size":       uploadedFile.FileSize,
			"file_store_link": uploadedFile.FileStoreKey,
			"owner_id":        uploadedFile.OwnerId,
			"folder_id":       uploadedFile.FolderId,
			"created_at":      uploadedFile.CreatedAt,
			"updated_at":      uploadedFile.UpdatedAt,
		})
}
