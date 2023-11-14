package auth

import (
	"net/http"

	"github.com/olad5/file-fort/internal/services/auth"
	appErrors "github.com/olad5/file-fort/pkg/errors"
	response "github.com/olad5/file-fort/pkg/utils"
)

func EnsureAuthenticated(authService auth.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			authHeader := r.Header.Get("Authorization")

			jwtClaims, err := authService.DecodeJWT(ctx, authHeader)
			if err != nil {
				response.ErrorResponse(w, appErrors.ErrUnauthorized, http.StatusUnauthorized)
				return
			}

			userId := jwtClaims.ID.String()
			if isUserLoggedIn := authService.IsUserLoggedIn(ctx, authHeader, userId); !isUserLoggedIn {
				response.ErrorResponse(w, appErrors.ErrUnauthorized, http.StatusUnauthorized)
				return
			}

			ctx = auth.Set(ctx, jwtClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminGuard(authService auth.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			jwtClaims, ok := auth.Get(ctx)
			if !ok {
				response.ErrorResponse(w, appErrors.ErrUserNotAdmin, http.StatusUnauthorized)
				return
			}

			if isUserAdmin := authService.IsUserAdmin(ctx, jwtClaims); !isUserAdmin {
				response.ErrorResponse(w, appErrors.ErrUserNotAdmin, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
