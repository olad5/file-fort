package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	appErrors "github.com/olad5/go-cloud-backup-system/pkg/errors"

	"github.com/olad5/go-cloud-backup-system/internal/infra"
	"github.com/olad5/go-cloud-backup-system/internal/usecases/users"
	response "github.com/olad5/go-cloud-backup-system/pkg/utils"
)

func (u UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Body == nil {
		response.ErrorResponse(w, appErrors.ErrMissingBody, http.StatusBadRequest)
		return
	}
	type requestDTO struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var request requestDTO
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		response.ErrorResponse(w, appErrors.ErrInvalidJson, http.StatusBadRequest)
		return
	}
	if request.Email == "" {
		response.ErrorResponse(w, "email required", http.StatusBadRequest)
		return
	}
	if request.Password == "" {
		response.ErrorResponse(w, "password required", http.StatusBadRequest)
		return
	}

	accessToken, err := u.userService.LogUserIn(ctx, request.Email, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, infra.ErrUserNotFound):
			response.ErrorResponse(w, "user does not exist", http.StatusNotFound)
			return
		case errors.Is(err, users.ErrPasswordIncorrect):
			response.ErrorResponse(w, "invalid credentials", http.StatusUnauthorized)
			return
		default:
			response.ErrorResponse(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError)
			return

		}
	}

	response.SuccessResponse(w, "user logged in successfully",
		map[string]interface{}{
			"access_token": accessToken,
		})
}
