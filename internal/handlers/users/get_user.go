package handlers

import (
	"errors"
	"net/http"

	"github.com/olad5/go-cloud-backup-system/internal/infra"
	appErrors "github.com/olad5/go-cloud-backup-system/pkg/errors"
	response "github.com/olad5/go-cloud-backup-system/pkg/utils"
)

func (u UserHandler) GetLoggedInUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, err := u.userService.GetLoggedInUser(ctx)
	if err != nil {
		switch {
		case errors.Is(err, infra.ErrUserNotFound):
			response.ErrorResponse(w, err.Error(), http.StatusNotFound)
			return
		default:
			response.ErrorResponse(w, appErrors.ErrSomethingWentWrong, http.StatusInternalServerError)
			return
		}
	}

	response.SuccessResponse(w, "user retrieved successfully",
		map[string]interface{}{
			"id":         user.ID.String(),
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"role":       user.Role,
		})
}
