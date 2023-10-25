package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/olad5/file-fort/internal/infra"
	appErrors "github.com/olad5/file-fort/pkg/errors"

	response "github.com/olad5/file-fort/pkg/utils"
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
	response.SuccessResponse(w, "file uploaded successfully", ToResponseFile(uploadedFile))
}
