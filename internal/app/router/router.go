package router

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/olad5/file-fort/internal/handlers/auth"
	fileHandlers "github.com/olad5/file-fort/internal/handlers/files"
	userHandlers "github.com/olad5/file-fort/internal/handlers/users"
	authService "github.com/olad5/file-fort/internal/services/auth"

	"github.com/go-chi/chi/v5"
)

func NewHttpRouter(userHandler userHandlers.UserHandler, fileHandler fileHandlers.FileHandler, authService authService.AuthService) http.Handler {
	router := chi.NewRouter()

	router.Group(func(r chi.Router) {
		r.Use(
			middleware.AllowContentType("application/json"),
			middleware.SetHeader("Content-Type", "application/json"),
		)
		r.Post("/users/login", userHandler.Login)
		r.Post("/users", userHandler.Register)
	})

	// -------------------------------------------------------------------------

	router.Group(func(r chi.Router) {
		r.Use(
			middleware.AllowContentType("application/json"),
			middleware.SetHeader("Content-Type", "application/json"),
		)
		r.Use(auth.EnsureAuthenticated(authService))

		r.Get("/users/me", userHandler.GetLoggedInUser)
		r.Get("/file/{id}", fileHandler.Download)
		r.Post("/folder", fileHandler.CreateFolder)
		r.Get("/folder/{id}/files", fileHandler.GetFilesByFolderId)
	})

	// -------------------------------------------------------------------------

	router.Group(func(r chi.Router) {
		r.Use(middleware.AllowContentType("multipart/form-data"))
		r.Use(auth.EnsureAuthenticated(authService))

		r.Post("/file", fileHandler.Upload)
	})

	// -------------------------------------------------------------------------

	router.Group(func(r chi.Router) {
		r.Use(
			middleware.AllowContentType("application/json"),
			middleware.SetHeader("Content-Type", "application/json"),
		)
		r.Use(auth.EnsureAuthenticated(authService))
		r.Use(auth.AdminGuard(authService))

		r.Post("/file/{id}/mark-unsafe", fileHandler.MarkFileAsUnSafe)
	})

	return router
}
