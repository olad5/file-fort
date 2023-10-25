package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/olad5/file-fort/internal/infra"
	appErrors "github.com/olad5/file-fort/pkg/errors"

	response "github.com/olad5/file-fort/pkg/utils"
)

func (f FileHandler) GetFilesByFolderId(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	if id == "" {
		response.ErrorResponse(w, "folder id required", http.StatusBadRequest)
		return
	}

	folderId, err := uuid.Parse(id)
	if err != nil {
		response.ErrorResponse(w, appErrors.ErrInvalidID.Error(), http.StatusBadRequest)
		return
	}

	pageQuery := r.URL.Query().Get("page")
	if pageQuery == "" {
		response.ErrorResponse(w, "page required", http.StatusBadRequest)
		return

	}

	pageNumber, err := strconv.Atoi(pageQuery)
	if err != nil {
		response.ErrorResponse(w, "page required", http.StatusBadRequest)
		return
	}
	rowQuery := r.URL.Query().Get("rows")
	if pageQuery == "" {
		response.ErrorResponse(w, "rows required", http.StatusBadRequest)
		return

	}
	rowsPerPage, err := strconv.Atoi(rowQuery)
	if err != nil {
		response.ErrorResponse(w, "rows required", http.StatusBadRequest)
		return
	}

	if pageNumber < 1 {
		pageNumber = 1
	}

	if rowsPerPage < 1 {
		rowsPerPage = 20
	} else if rowsPerPage > 20 {
		rowsPerPage = 20
	}

	files, err := f.fileService.GetFilesByFolderId(ctx, folderId, pageNumber, rowsPerPage)
	if err != nil {
		switch {
		case errors.Is(err, infra.ErrFolderNotFound):
			response.ErrorResponse(w, "file does not exist", http.StatusNotFound)
			return
		case errors.Is(err, infra.ErrUserNotAuthorized):
			response.ErrorResponse(w, "unauthorized to view this folder", http.StatusForbidden)
			return
		default:
			response.ErrorResponse(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError)
			return
		}
	}

	var results []map[string]interface{}
	for _, file := range files {
		fileData := ToResponseFile(file)
		results = append(results, fileData)
	}

	response.SuccessResponse(w, "files retreived successfully",
		map[string]interface{}{
			"owner_id":      files[0].OwnerId,
			"folder_id":     folderId,
			"files":         results,
			"page":          pageNumber,
			"rows_per_page": rowsPerPage,
		})
}
