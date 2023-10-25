package handlers

import (
	"encoding/json"
	"net/http"

	appErrors "github.com/olad5/file-fort/pkg/errors"

	response "github.com/olad5/file-fort/pkg/utils"
)

func (f FileHandler) CreateFolder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Body == nil {
		response.ErrorResponse(w, appErrors.ErrMissingBody, http.StatusBadRequest)
		return
	}
	type requestDTO struct {
		FolderName string `json:"folder_name"`
	}

	var request requestDTO
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		response.ErrorResponse(w, appErrors.ErrInvalidJson, http.StatusBadRequest)
		return
	}
	if request.FolderName == "" {
		response.ErrorResponse(w, "folder_name required", http.StatusBadRequest)
		return
	}

	newFolder, err := f.fileService.CreateFolder(ctx, request.FolderName)
	if err != nil {
		switch {
		default:
			response.ErrorResponse(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError)
			return
		}
	}

	response.SuccessResponse(w, "folder created successfully",
		map[string]interface{}{
			"id":          newFolder.ID,
			"folder_name": newFolder.FolderName,
			"owner_id":    newFolder.OwnerId,
			"created_at":  newFolder.CreatedAt,
			"updated_at":  newFolder.UpdatedAt,
		})
}
