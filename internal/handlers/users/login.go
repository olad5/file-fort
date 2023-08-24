package handlers

import (
	"encoding/json"
	"net/http"

	appErrors "github.com/olad5/go-cloud-backup-system/pkg/errors"

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
	if err != nil && err.Error() == users.ErrUserNotFound {
		response.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil && err.Error() == users.ErrPasswordIncorrect {
		response.ErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		response.ErrorResponse(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError)
		return
	}

	response.SuccessResponse(w, "user logged in successfully",
		map[string]interface{}{
			"access_token": accessToken,
		})
}
