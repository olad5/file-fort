//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/olad5/go-cloud-backup-system/config"
	"github.com/olad5/go-cloud-backup-system/internal/app/router"
	handlers "github.com/olad5/go-cloud-backup-system/internal/handlers/users"
	"github.com/olad5/go-cloud-backup-system/internal/infra/postgres"
	"github.com/olad5/go-cloud-backup-system/internal/infra/redis"
	"github.com/olad5/go-cloud-backup-system/internal/services/auth"
	"github.com/olad5/go-cloud-backup-system/internal/usecases/users"
	"github.com/olad5/go-cloud-backup-system/pkg/app/server"
	"github.com/olad5/go-cloud-backup-system/tests"
)

var (
	svr            *server.Server
	userHandler    *handlers.UserHandler
	configurations *config.Configurations
	authService    auth.AuthService
)

func TestMain(m *testing.M) {
	configurations = config.GetConfig("../config/.test.env")
	ctx := context.Background()

	userRepo, err := postgres.NewPostgresRepo(ctx, configurations.DatabaseUrl)
	if err != nil {
		log.Fatal("Error Initializing User Repo")
	}

	err = userRepo.Ping(ctx)
	if err != nil {
		log.Fatal(err)
	}

	redisCache, err := redis.New(ctx, configurations)
	if err != nil {
		log.Fatal("Error Initializing redisCache", err)
	}

	authService, err = auth.NewRedisAuthService(ctx, redisCache, configurations)
	if err != nil {
		log.Fatal("Error Initializing Auth Service", err)
	}

	userService, err := users.NewUserService(userRepo, authService, configurations)
	if err != nil {
		log.Fatal("Error dnitializing UserService")
	}

	userHandler, err = handlers.NewHandler(*userService, authService)
	if err != nil {
		log.Fatal("failed to create the User handler: ", err)
	}
	appRouter := router.NewHttpRouter(*userHandler, authService)
	svr = server.CreateNewServer(appRouter)

	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestRegister(t *testing.T) {
	route := "/users"
	t.Run("test for invalid json request body",
		func(t *testing.T) {
			req, _ := http.NewRequest("POST", route, nil)
			response := tests.ExecuteRequest(req, svr)
			tests.AssertStatusCode(t, http.StatusBadRequest, response.Code)
		},
	)
	t.Run(`Given a valid user registration request, when the user submits the request, 
    then the server should respond with a success status code, and the user's account 
    should be created in the database.`,
		func(t *testing.T) {
			email := "will" + fmt.Sprint(tests.GenerateUniqueId()) + "@gmail.com"
			requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "first_name": "will",
      "last_name": "hansen",
      "password": "some-random-password"
      }`, email))
			req, _ := http.NewRequest("POST", route, bytes.NewBuffer(requestBody))
			response := tests.ExecuteRequest(req, svr)
			tests.AssertStatusCode(t, http.StatusOK, response.Code)
			data := tests.ParseResponse(response)["data"].(map[string]interface{})
			tests.AssertResponseMessage(t, data["email"].(string), email)
		},
	)

	t.Run(`Given a user registration request with an email address that already exists,
    when the user submits the request, then the server should respond with an error
    status code, and the response should indicate that the email address is already
    taken. `,
		func(t *testing.T) {
			email := "will@gmail.com"
			requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "first_name": "will",
      "last_name": "hansen",
      "password": "some-random-password"
      }`, email))
			req, _ := http.NewRequest("POST", route, bytes.NewBuffer(requestBody))
			_ = tests.ExecuteRequest(req, svr)

			secondRequestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "first_name": "will",
      "last_name": "hansen",
      "password": "passcode"
      }`, email))
			req, _ = http.NewRequest("POST", route, bytes.NewBuffer(secondRequestBody))

			response := tests.ExecuteRequest(req, svr)
			tests.AssertStatusCode(t, http.StatusBadRequest, response.Code)
			message := tests.ParseResponse(response)["message"].(string)
			tests.AssertResponseMessage(t, message, "email already exist")
		},
	)
}

func TestLogin(t *testing.T) {
	route := "/users/login"
	t.Run("test for invalid json request body",
		func(t *testing.T) {
			req, _ := http.NewRequest("POST", route, nil)
			response := tests.ExecuteRequest(req, svr)
			tests.AssertStatusCode(t, http.StatusBadRequest, response.Code)
		},
	)
	t.Run(`Given a user attempts to log in with valid credentials,
    when they make a POST request to the login endpoint with their username and password,
    then they should receive a 200 OK response,
    and the response should contain a JSON web token (JWT) in the 'token' field,
    and the token should be valid and properly signed.`,
		func(t *testing.T) {
			email := "will@gmail.com"
			requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "password": "some-random-password"
      }`, email))
			req, _ := http.NewRequest("POST", route, bytes.NewBuffer(requestBody))
			response := tests.ExecuteRequest(req, svr)

			tests.AssertStatusCode(t, http.StatusOK, response.Code)
			responseBody := tests.ParseResponse(response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "user logged in successfully")

			data := responseBody["data"].(map[string]interface{})

			accessToken, exists := data["access_token"]
			if !exists {
				t.Error("Missing 'accesstoken' key in the JSON response")
			}

			_, isString := accessToken.(string)
			if !isString {
				t.Error("'accesstoken' value is not a string")
			}
		},
	)

	t.Run(`Given a user tries to log in with an account that does not exist,
    when they make a POST request to the login endpoint with a non-existent email,
    then they should receive a 404 Not Found response,
    and the response should contain an error message indicating that the account 
    does not exist.`,
		func(t *testing.T) {
			email := "emailnoexist@gmail.com"
			requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "first_name": "mike",
      "last_name": "wilson",
      "password": "some-random-password"
      }`, email))
			req, _ := http.NewRequest("POST", route, bytes.NewBuffer(requestBody))
			response := tests.ExecuteRequest(req, svr)

			tests.AssertStatusCode(t, http.StatusBadRequest, response.Code)
			message := tests.ParseResponse(response)["message"].(string)
			tests.AssertResponseMessage(t, message, "incorrect credentials")
		},
	)
}