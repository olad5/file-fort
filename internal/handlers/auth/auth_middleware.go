package auth

import (
	"context"
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
			if isUserLoggedIn := authService.IsUserLoggedIn(ctx, authHeader, userId); isUserLoggedIn != true {
				response.ErrorResponse(w, appErrors.ErrUnauthorized, http.StatusUnauthorized)
				return
			}

			ctx = context.WithValue(ctx, "jwtClaims", jwtClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminGuard(authService auth.AuthService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			jwtClaims, ok := ctx.Value("jwtClaims").(auth.JWTClaims)
			if ok == false {
				response.ErrorResponse(w, appErrors.ErrUserNotAdmin, http.StatusUnauthorized)
				return
			}

			if isUserAdmin := authService.IsUserAdmin(ctx, jwtClaims); isUserAdmin != true {
				response.ErrorResponse(w, appErrors.ErrUserNotAdmin, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
