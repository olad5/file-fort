package auth

import (
	"context"
	"net/http"

	"github.com/olad5/file-fort/internal/services/auth"
	appErrors "github.com/olad5/file-fort/pkg/errors"
	response "github.com/olad5/file-fort/pkg/utils"
)

func AuthMiddleware(authService auth.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			authHeader := r.Header.Get("Authorization")

			userId, err := authService.DecodeJWT(ctx, authHeader)
			if err != nil {
				response.ErrorResponse(w, appErrors.ErrUnauthorized, http.StatusUnauthorized)
				return
			}

			if isUserLoggedIn := authService.IsUserLoggedIn(ctx, authHeader, userId); isUserLoggedIn != true {
				response.ErrorResponse(w, appErrors.ErrUnauthorized, http.StatusUnauthorized)
				return
			}

			ctx = context.WithValue(r.Context(), "userId", userId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
