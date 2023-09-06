package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/olad5/go-cloud-backup-system/internal/infra"
	appErrors "github.com/olad5/go-cloud-backup-system/pkg/errors"

	response "github.com/olad5/go-cloud-backup-system/pkg/utils"
)

func (f FileHandler) Download(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileId := chi.URLParam(r, "id")
	if fileId == "" {
		response.ErrorResponse(w, "file id required", http.StatusBadRequest)
		return
	}

	fileIdAsUUID, err := uuid.Parse(fileId)
	if err != nil {
		response.ErrorResponse(w, appErrors.ErrInvalidID, http.StatusBadRequest)
		return
	}

	downloadUrl, err := f.fileService.DownloadFile(ctx, fileIdAsUUID)
	if err != nil {
		switch {
		case errors.Is(err, infra.ErrFileNotFound):
			response.ErrorResponse(w, "file does not exist", http.StatusNotFound)
			return
		case errors.Is(err, infra.ErrUserNotAuthorized):
			response.ErrorResponse(w, "unauthorized to view this file", http.StatusForbidden)
			return
		default:
			response.ErrorResponse(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError)
			return
		}
	}

	response.SuccessResponse(w, "download url generated successfully",
		map[string]interface{}{
			"download_url": downloadUrl,
		})
}
