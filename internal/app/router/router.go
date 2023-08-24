package router

import (
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	handlers "github.com/olad5/go-cloud-backup-system/internal/handlers/users"
	authService "github.com/olad5/go-cloud-backup-system/internal/services/auth"

	"github.com/go-chi/chi/v5"
)

func NewHttpRouter(userHandler handlers.UserHandler, authService authService.AuthService) http.Handler {
	router := chi.NewRouter()

	router.Use(
		middleware.AllowContentType("application/json"),
		middleware.SetHeader("Content-Type", "application/json"),
	)

	router.Post("/users/login", userHandler.Login)
	router.Post("/users", userHandler.Register)
	return router
}
