package main

import (
	"context"
	"log"
	"net/http"

	"github.com/olad5/file-fort/config/data"

	"github.com/olad5/file-fort/config"
	"github.com/olad5/file-fort/internal/app/router"
	fileHandlers "github.com/olad5/file-fort/internal/handlers/files"
	userHandlers "github.com/olad5/file-fort/internal/handlers/users"
	"github.com/olad5/file-fort/internal/infra/aws"
	"github.com/olad5/file-fort/internal/infra/postgres"
	"github.com/olad5/file-fort/internal/infra/redis"
	"github.com/olad5/file-fort/internal/services/auth"
	fileServices "github.com/olad5/file-fort/internal/usecases/files"
	"github.com/olad5/file-fort/internal/usecases/users"
	"github.com/olad5/file-fort/pkg/app/server"
)

func main() {
	configurations := config.GetConfig(".env")
	ctx := context.Background()

	port := configurations.Port

	postgresConnection := data.StartPostgres(configurations.DatabaseUrl)
	if err := postgres.Migrate(ctx, postgresConnection); err != nil {
		log.Fatal("Error Migrating postgres", err)
	}

	userRepo, err := postgres.NewPostgresUserRepo(ctx, postgresConnection)
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

	userService, err := users.NewUserService(userRepo, authService)
	if err != nil {
		log.Fatal("Error Initializing UserService")
	}

	userHandler, err := userHandlers.NewUserHandler(*userService, authService)
	if err != nil {
		log.Fatal("failed to create the User handler: ", err)
	}

	folderRepo, err := postgres.NewPostgresFolderRepo(ctx, postgresConnection)
	if err != nil {
		log.Fatal("Error Initializing Folder Repo", err)
	}

	fileRepo, err := postgres.NewPostgresFileRepo(ctx, postgresConnection)
	if err != nil {
		log.Fatal("Error Initializing File Repo", err)
	}

	fileStore, err := aws.NewAwsFileStore(ctx, configurations)
	if err != nil {
		log.Fatal("Error Initializing AWS File store", err)
	}

	filesService, err := fileServices.NewFileService(fileRepo, folderRepo, fileStore)
	if err != nil {
		log.Fatal("Error Initializing UserService")
	}

	fileHandler, err := fileHandlers.NewFileHandler(*filesService)
	if err != nil {
		log.Fatal("failed to create the fileHandler: ", err)
	}

	appRouter := router.NewHttpRouter(*userHandler, *fileHandler, authService)

	svr := server.CreateNewServer(appRouter)

	log.Printf("Server Listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, svr.Router))
}
