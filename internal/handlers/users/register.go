package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	appErrors "github.com/olad5/file-fort/pkg/errors"

	"github.com/olad5/file-fort/internal/usecases/users"
	response "github.com/olad5/file-fort/pkg/utils"
)

func (u UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Body == nil {
		response.ErrorResponse(w, appErrors.ErrMissingBody, http.StatusBadRequest)
		return
	}
	type requestDTO struct {
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Password  string `json:"password"`
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

	if request.FirstName == "" {
		response.ErrorResponse(w, "first_name required", http.StatusBadRequest)
		return
	}
	if request.LastName == "" {
		response.ErrorResponse(w, "last_name required", http.StatusBadRequest)
		return
	}

	newUser, err := u.userService.CreateUser(ctx, request.FirstName, request.LastName, request.Email, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, users.ErrUserAlreadyExists):
			response.ErrorResponse(w, err.Error(), http.StatusBadRequest)
			return
		default:
			response.ErrorResponse(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError)
			return
		}
	}
	response.SuccessResponse(w, "user created successfully",
		map[string]interface{}{
			"id":         newUser.ID.String(),
			"email":      newUser.Email,
			"first_name": newUser.FirstName,
			"last_name":  newUser.LastName,
		})
}
