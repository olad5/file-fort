package main

import (
	"context"
	"log"
	"net/http"

	"github.com/olad5/go-cloud-backup-system/config"
	"github.com/olad5/go-cloud-backup-system/internal/app/router"
	handlers "github.com/olad5/go-cloud-backup-system/internal/handlers/users"
	"github.com/olad5/go-cloud-backup-system/internal/infra/postgres"
	"github.com/olad5/go-cloud-backup-system/internal/infra/redis"
	"github.com/olad5/go-cloud-backup-system/internal/services/auth"
	"github.com/olad5/go-cloud-backup-system/internal/usecases/users"
	"github.com/olad5/go-cloud-backup-system/pkg/app/server"
)

func main() {
	configurations := config.GetConfig(".env")
	ctx := context.Background()

	port := configurations.Port
	userRepo, err := postgres.NewPostgresRepo(ctx, configurations.DatabaseUrl)
	if err != nil {
		log.Fatal("Error Initializing User Repo", err)
	}

	err = userRepo.Ping(ctx)
	if err != nil {
		log.Fatal("Failed to ping UserRepo", err)
	}

	redisCache, err := redis.New(ctx, configurations)
	if err != nil {
		log.Fatal("Error Initializing redisCache", err)
	}

	authService, err := auth.NewRedisAuthService(ctx, redisCache, configurations)
	if err != nil {
		log.Fatal("Error Initializing Auth Service", err)
	}

	userService, err := users.NewUserService(userRepo, authService, configurations)
	if err != nil {
		log.Fatal("Error Initializing UserService")
	}

	userHandler, err := handlers.NewHandler(*userService, authService)
	if err != nil {
		log.Fatal("failed to create the User handler: ", err)
	}

	appRouter := router.NewHttpRouter(*userHandler, authService)

	svr := server.CreateNewServer(appRouter)

	log.Printf("Server Listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, svr.Router))
}
