package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/olad5/file-fort/internal/infra"
	appErrors "github.com/olad5/file-fort/pkg/errors"
	response "github.com/olad5/file-fort/pkg/utils"
)

func (f FileHandler) MarkFileAsUnSafe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if id == "" {
		response.ErrorResponse(w, "file id required", http.StatusBadRequest)
		return
	}

	fileId, err := uuid.Parse(id)
	if err != nil {
		response.ErrorResponse(w, appErrors.ErrInvalidID.Error(), http.StatusBadRequest)
		return
	}

	err = f.fileService.MarkFileAsUnsafe(ctx, fileId)
	if err != nil {
		switch {
		case errors.Is(err, infra.ErrFileNotFound):
			response.ErrorResponse(w, "file does not exist", http.StatusNotFound)
			return
		case errors.Is(err, infra.ErrUserNotAuthorized):
			response.ErrorResponse(w, appErrors.ErrUserNotAdmin, http.StatusUnauthorized)
			return
		default:
			response.ErrorResponse(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError)
			return
		}
	}

	response.SuccessResponse(w, "file marked unsafe successfully",
		map[string]interface{}{
			"file_id": fileId,
		})
}
