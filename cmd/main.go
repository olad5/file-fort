package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/olad5/file-fort/config/data"

	"github.com/olad5/file-fort/config"
	"github.com/olad5/file-fort/internal/app/router"
	fileHandlers "github.com/olad5/file-fort/internal/handlers/files"
	healthHandlers "github.com/olad5/file-fort/internal/handlers/health"
	userHandlers "github.com/olad5/file-fort/internal/handlers/users"
	"github.com/olad5/file-fort/internal/infra/aws"
	"github.com/olad5/file-fort/internal/infra/postgres"
	"github.com/olad5/file-fort/internal/infra/redis"
	"github.com/olad5/file-fort/internal/services/auth"
	fileServices "github.com/olad5/file-fort/internal/usecases/files"
	"github.com/olad5/file-fort/internal/usecases/users"
)

func main() {
	configurations := config.GetConfig(".env")
	ctx := context.Background()

	port := configurations.Port

	postgresConnection := data.StartPostgres(configurations.DatabaseUrl)
	if err := postgres.Migrate(ctx, postgresConnection); err != nil {
		log.Fatal("Error Migrating postgres", err)
	}

	defer postgresConnection.Close()

	userRepo, err := postgres.NewPostgresUserRepo(ctx, postgresConnection)
	if err != nil {
		log.Fatal("Error Initializing User Repo", err)
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

	healthHandler, err := healthHandlers.NewHealthHandler(ctx, postgresConnection, redisCache)
	if err != nil {
		log.Fatal("failed to create the healthHandler: ", err)
	}

	appRouter := router.NewHttpRouter(*userHandler, *fileHandler, *healthHandler, authService)

	server := &http.Server{Addr: ":" + port, Handler: appRouter}
	go func() {
		fmt.Println("Server is running....")
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Printf("HTTP server ListenAndServe: %v", err)
		}
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	<-signals

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exiting gracefully")
}
