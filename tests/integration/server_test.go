//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/olad5/file-fort/config/data"
	"github.com/olad5/file-fort/internal/infra/aws"

	fileHandlers "github.com/olad5/file-fort/internal/handlers/files"
	userHandlers "github.com/olad5/file-fort/internal/handlers/users"
	fileServices "github.com/olad5/file-fort/internal/usecases/files"

	"github.com/olad5/file-fort/config"
	"github.com/olad5/file-fort/internal/app/router"
	"github.com/olad5/file-fort/internal/infra/postgres"
	"github.com/olad5/file-fort/internal/infra/redis"
	"github.com/olad5/file-fort/internal/services/auth"
	"github.com/olad5/file-fort/internal/usecases/users"
	"github.com/olad5/file-fort/pkg/app/server"
	"github.com/olad5/file-fort/tests"
)

var (
	svr            *server.Server
	userHandler    *userHandlers.UserHandler
	configurations *config.Configurations
	authService    auth.AuthService
)

var (
	userEmail    = "will@gmail.com"
	userPassword = "some-random-password"
)

func TestMain(m *testing.M) {
	configurations = config.GetConfig("../config/.test.env")
	ctx := context.Background()

	postgresConnection := data.StartPostgres(configurations.DatabaseUrl)
	userRepo, err := postgres.NewPostgresUserRepo(ctx, postgresConnection)
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

	userService, err := users.NewUserService(userRepo, authService)
	if err != nil {
		log.Fatal("Error dnitializing UserService")
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
		log.Fatal("Error Initializing AWS File store\n", err)
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
	svr = server.CreateNewServer(appRouter)

	exitVal := m.Run()
	os.Exit(exitVal)
}

func TestRegister(t *testing.T) {
	route := "/users"
	t.Run("test for invalid json request body",
		func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, route, nil)
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
			req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
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
			req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
			_ = tests.ExecuteRequest(req, svr)

			secondRequestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "first_name": "will",
      "last_name": "hansen",
      "password": "passcode"
      }`, email))
			req, _ = http.NewRequest(http.MethodPost, route, bytes.NewBuffer(secondRequestBody))

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
			req, _ := http.NewRequest(http.MethodPost, route, nil)
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
			requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "password": "%s"
      }`, userEmail, userPassword))
			req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
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

	t.Run(`Given a user attempts to log in with invalid password,
    when they make a POST request to the login endpoint with their username and password,
    then they should receive a 401 unauthorized response,
    and the response should contain an invalidd credentials message.`,
		func(t *testing.T) {
			email := "will@gmail.com"
			requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "password": "invalid-password"
      }`, email))
			req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
			response := tests.ExecuteRequest(req, svr)

			tests.AssertStatusCode(t, http.StatusUnauthorized, response.Code)
			responseBody := tests.ParseResponse(response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "invalid credentials")
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
			req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
			response := tests.ExecuteRequest(req, svr)

			tests.AssertStatusCode(t, http.StatusNotFound, response.Code)
			message := tests.ParseResponse(response)["message"].(string)
			tests.AssertResponseMessage(t, message, "user does not exist")
		},
	)
}

func TestFileUpload(t *testing.T) {
	route := "/file"
	fieldName := "file"

	t.Run(`Given an authenticated user
        When they upload a file exceeding the size limit of 200mb
        Then the API should return a validation error
        And the file should not be stored on the server
      `,
		func(t *testing.T) {
			fileSizeOf200mb := int64(1024 * 1024 * 200)
			tempFile, closeFile := createTempFile(t, "someLargeFile", fileSizeOf200mb)

			defer closeFile()
			var requestBody bytes.Buffer
			writer := multipart.NewWriter(&requestBody)
			createFormFile(t, writer, tempFile, fieldName)
			writer.Close()

			req, _ := http.NewRequest(http.MethodPost, route, &requestBody)

			token := logUserIn(userEmail, userPassword)

			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			response := ExecuteRequestMultiPart(req, svr)
			tests.AssertStatusCode(t, http.StatusBadRequest, response.Code)
			responseBody := tests.ParseResponse(response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "The file you are trying to upload exceeds the maximum allowed size of 200MB.")
		},
	)
	t.Run(`Given an authenticated user
        When they upload a file with a folder Id and the folder does not exist
        Then the API should return a not found error
        And the file should not be stored on the server
      `,
		func(t *testing.T) {
			fileSize := int64(1024)
			tempFile, fileCleanUp := createTempFile(t, "someFile", fileSize)
			defer fileCleanUp()

			var requestBody bytes.Buffer
			writer := multipart.NewWriter(&requestBody)
			createFormFile(t, writer, tempFile, fieldName)
			createFormField(t, writer, "folder_id", "475e0728-9d29-4e72-93a5-0f94ec430977")
			writer.Close()

			req, _ := http.NewRequest(http.MethodPost, route, &requestBody)
			token := logUserIn(userEmail, userPassword)

			req.Header.Set("Authorization", "Bearer "+token)

			req.Header.Set("Content-Type", writer.FormDataContentType())
			response := ExecuteRequestMultiPart(req, svr)
			tests.AssertStatusCode(t, http.StatusNotFound, response.Code)
			responseBody := tests.ParseResponse(response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "folder does not exist")
		},
	)

	t.Run(`Given an authenticated user
        When they upload a file without a folder Id
        Then the API should save the file successfully in the user's default folder
      `,
		func(t *testing.T) {
			fileSize := int64(1024)
			tempFile, fileCleanUp := openImageFile(t, "someFile", fileSize)

			defer fileCleanUp()
			var requestBody bytes.Buffer
			writer := multipart.NewWriter(&requestBody)

			createFormFile(t, writer, tempFile, fieldName)
			writer.Close()

			req, _ := http.NewRequest(http.MethodPost, route, &requestBody)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			token := logUserIn(userEmail, userPassword)

			userId := getCurrentUser(t, token)["id"].(string)

			req.Header.Set("Authorization", "Bearer "+token)

			response := ExecuteRequestMultiPart(req, svr)
			tests.AssertStatusCode(t, http.StatusOK, response.Code)
			responseBody := tests.ParseResponse(response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "file uploaded successfully")

			data := responseBody["data"].(map[string]interface{})
			filename := strings.Split(tempFile.Name(), "/")[2]
			tests.AssertResponseMessage(t, data["file_name"].(string), filename)
			tests.AssertResponseMessage(t, data["folder_id"].(string), userId)
			tests.AssertResponseMessage(t, data["owner_id"].(string), userId)
		},
	)
}

func TestFileDownload(t *testing.T) {
	route := "/file"
	t.Run(`Given an authenticated user
        When they try to download a file using a valid fileId
        Then the API should return a 200 OK response
        And the download url should be sent in the response body
      `,
		func(t *testing.T) {
			fileSize := int64(1024)
			token := logUserIn(userEmail, userPassword)
			fileId := uploadFile(t, fileSize, "someFile", "", token)

			req, _ := http.NewRequest(http.MethodGet, route+"/"+fileId, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			response := tests.ExecuteRequest(req, svr)
			tests.AssertStatusCode(t, http.StatusOK, response.Code)
			responseBody := tests.ParseResponse(response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "download url generated successfully")
			data := responseBody["data"].(map[string]interface{})
			downloadUrl := data["download_url"].(string)
			_, err := url.ParseRequestURI(downloadUrl)
			if err != nil {
				t.Errorf("downloadUrl is not a url, got url: %s ", downloadUrl)
			}
		},
	)

	t.Run(` Given an unauthenticated user
      When the user attempts to download a file
      Then the server should respond with an "Unauthorized" status code (401)
      And the response message should indicate the user needs to log in
      `,
		func(t *testing.T) {
			fileSize := int64(1024)
			token := logUserIn(userEmail, userPassword)
			fileId := uploadFile(t, fileSize, "someFile", "", token)

			req, _ := http.NewRequest(http.MethodGet, route+"/"+fileId, nil)
			emptyBearerToken := ""
			req.Header.Set("Authorization", "Bearer "+emptyBearerToken)
			response := tests.ExecuteRequest(req, svr)
			tests.AssertStatusCode(t, http.StatusUnauthorized, response.Code)
			responseBody := tests.ParseResponse(response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "unauthorized")
		},
	)
	t.Run(` Given an authenticated user with valid credentials
      When the user attempts to download a file they don't own
      Then the server should respond with a "Forbidden" status code (403)
      And the response message should indicate access is denied
      `,
		func(t *testing.T) {
			fileSize := int64(1024)
			fileOwnerToken := logUserIn(userEmail, userPassword)
			fileId := uploadFile(t, fileSize, "someFile", "", fileOwnerToken)

			req, _ := http.NewRequest(http.MethodGet, route+"/"+fileId, nil)

			newUserEmail := "mikesmith" + fmt.Sprint(tests.GenerateUniqueId()) + "@gmail.com"
			newUserPassword := "some-password"
			_ = createUser(t, "mike", "smith", newUserEmail, newUserPassword)
			currentUserToken := logUserIn(newUserEmail, newUserPassword)

			req.Header.Set("Authorization", "Bearer "+currentUserToken)
			response := tests.ExecuteRequest(req, svr)
			tests.AssertStatusCode(t, http.StatusForbidden, response.Code)
			responseBody := tests.ParseResponse(response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "unauthorized to view this file")
		},
	)
	t.Run(` Given an authenticated user with valid credentials
          When the user requests to download a non-existent file
          Then the server should respond with a "Not Found" status code (404)
          And the response message should indicate the file was not found
        `,
		func(t *testing.T) {
			token := logUserIn(userEmail, userPassword)

			fileId := "6372ad3a-f9b4-4ff9-b7a3-c3f0a42be194"
			req, _ := http.NewRequest(http.MethodGet, route+"/"+fileId, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			response := tests.ExecuteRequest(req, svr)
			tests.AssertStatusCode(t, http.StatusNotFound, response.Code)
			responseBody := tests.ParseResponse(response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "file does not exist")
		},
	)
}

func TestCreateFolder(t *testing.T) {
	route := "/folder"
	t.Run(`Given an authenticated user with valid credentials
        When the user sends a request to create a folder with a name
        Then the server should respond with an OK status code 
        And the response body should contain the created folder information
        And the folder should be stored in the database with the correct owner
      `,
		func(t *testing.T) {
			email := "mikesmith" + fmt.Sprint(tests.GenerateUniqueId()) + "@gmail.com"
			password := "some-password"

			userId := createUser(t, "mike", "smith", email, password)
			token := logUserIn(email, password)
			folderName := "some-new-folder"

			requestBody := []byte(fmt.Sprintf(`{
      "folder_name": "%s"
      }`, folderName))
			req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))

			req.Header.Set("Authorization", "Bearer "+token)
			response := tests.ExecuteRequest(req, svr)
			tests.AssertStatusCode(t, http.StatusOK, response.Code)
			responseBody := tests.ParseResponse(response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "folder created successfully")
			data := responseBody["data"].(map[string]interface{})
			tests.AssertResponseMessage(t, data["folder_name"].(string), folderName)
			tests.AssertResponseMessage(t, data["owner_id"].(string), userId)
		},
	)
}

func TestGetFilesInFolder(t *testing.T) {
	t.Run(`Given a user is authenticated,
      When they request to access files in a specific folder,
      Then the API should return a list of files in the folder with appropriate 
      metadata.
      `,
		func(t *testing.T) {
			email := "mikesmith" + fmt.Sprint(tests.GenerateUniqueId()) + "@gmail.com"
			password := "some-password"

			_ = createUser(t, "mike", "smith", email, password)
			token := logUserIn(email, password)
			userId := getCurrentUser(t, token)["id"].(string)
			folderName := "some-new-folder"

			folderId := createFolder(t, folderName, token)
			fileSize := int64(1024)
			fileIds := []string{}
			numberOfFiles := 3
			for i := 0; i < numberOfFiles; i++ {
				fileId := uploadFile(t, fileSize, "someFile"+fmt.Sprint(tests.GenerateUniqueId()), folderId, token)
				fileIds = append(fileIds, fileId)

			}
			route := "/folder/" + folderId + "/files" + "?page=1&rows=20"
			requestBody := []byte(fmt.Sprintf(`{
      "folder_name": "%s"
      }`, folderName))
			req, _ := http.NewRequest(http.MethodGet, route, bytes.NewBuffer(requestBody))

			req.Header.Set("Authorization", "Bearer "+token)
			response := tests.ExecuteRequest(req, svr)
			tests.AssertStatusCode(t, http.StatusOK, response.Code)
			responseBody := tests.ParseResponse(response)
			message := responseBody["message"].(string)
			tests.AssertResponseMessage(t, message, "files retreived successfully")
			data := responseBody["data"].(map[string]interface{})
			files := data["files"].([]interface{})
			tests.AssertResponseMessage(t, data["folder_id"].(string), folderId)
			tests.AssertResponseMessage(t, data["owner_id"].(string), userId)
			if len(files) != numberOfFiles {
				t.Errorf("got files length: %d expected: %d", len(files), numberOfFiles)
			}
		},
	)
}

func createFolder(t testing.TB, folderName, accessToken string) string {
	t.Helper()
	route := "/folder"

	requestBody := []byte(fmt.Sprintf(`{
      "folder_name": "%s"
      }`, folderName))
	req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))

	req.Header.Set("Authorization", "Bearer "+accessToken)
	response := tests.ExecuteRequest(req, svr)
	responseBody := tests.ParseResponse(response)
	data := responseBody["data"].(map[string]interface{})
	folderId := data["id"].(string)
	return folderId
}

func uploadFile(t testing.TB, fileSize int64, fileName, folderId, accessToken string) string {
	t.Helper()
	route := "/file"
	fieldName := "file"

	tempFile, fileCleanUp := openImageFile(t, fileName, fileSize)

	defer fileCleanUp()
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	createFormFile(t, writer, tempFile, fieldName)
	createFormField(t, writer, "folder_id", folderId)
	writer.Close()

	req, _ := http.NewRequest(http.MethodPost, route, &requestBody)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	req.Header.Set("Authorization", "Bearer "+accessToken)

	response := ExecuteRequestMultiPart(req, svr)
	responseBody := tests.ParseResponse(response)
	data := responseBody["data"].(map[string]interface{})
	fileId := data["id"].(string)

	return fileId
}

func createUser(t testing.TB, firstName, lastName, email, password string) string {
	t.Helper()
	route := "/users"
	requestBody := []byte(fmt.Sprintf(`{
      "first_name": "%s",
      "last_name": "%s",
      "email": "%s",
      "password": "%s"
      }`, firstName, lastName, email, password))
	req, _ := http.NewRequest(http.MethodPost, route, bytes.NewBuffer(requestBody))
	response := tests.ExecuteRequest(req, svr)
	data := tests.ParseResponse(response)["data"].(map[string]interface{})
	userId := data["id"].(string)

	return userId
}

func getCurrentUser(t testing.TB, token string) map[string]interface{} {
	t.Helper()
	req, _ := http.NewRequest(http.MethodGet, "/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	response := tests.ExecuteRequest(req, svr)
	responseBody := tests.ParseResponse(response)
	data := responseBody["data"].(map[string]interface{})
	return data
}

func logUserIn(email, password string) string {
	requestBody := []byte(fmt.Sprintf(`{
      "email": "%s",
      "password": "%s"
      }`, email, password))
	loginReq, _ := http.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(requestBody))
	loginResponse := tests.ExecuteRequest(loginReq, svr)
	loginResponseBody := tests.ParseResponse(loginResponse)
	loginData := loginResponseBody["data"].(map[string]interface{})
	accessToken := loginData["access_token"]
	token := accessToken.(string)
	return token
}

func openImageFile(t testing.TB, fileName string, fileSize int64) (*os.File, func()) {
	t.Helper()

	tempFile, err := os.Open("../data/wall.jpg")
	if err != nil {
		t.Fatal("Error opening file:", err)
	}

	removeFile := func() {
		tempFile.Close()
	}

	return tempFile, removeFile
}

func createTempFile(t testing.TB, fileName string, fileSize int64) (*os.File, func()) {
	t.Helper()

	tempFile, err := os.CreateTemp("", fileName)
	if err != nil {
		t.Fatal("Error creating temporary file:", err)
	}

	tempFileName := tempFile.Name()

	err = os.Truncate(tempFileName, fileSize)
	if err != nil {
		t.Fatal("Error Truncating tempFile to desiredSize:", err)
	}
	removeFile := func() {
		tempFile.Close()
		os.Remove(tempFileName)
	}

	return tempFile, removeFile
}

func createFormFile(t testing.TB, writer *multipart.Writer, tempFile *os.File, fieldName string) {
	t.Helper()
	fileWriter, err := writer.CreateFormFile(fieldName, tempFile.Name())
	if err != nil {
		t.Fatal("Error creating form file:", err)
	}

	_, err = io.Copy(fileWriter, tempFile)
	if err != nil {
		t.Fatal("Error copying file data to request:", err)
	}
}

func createFormField(t testing.TB, writer *multipart.Writer, key, value string) {
	t.Helper()
	formField, err := writer.CreateFormField(key)
	if err != nil {
		t.Fatal("Error creating form field :", err)
	}
	_, err = formField.Write([]byte(value))
	if err != nil {
		t.Fatal("Error writing value to form field:", err)
	}
}

func ExecuteRequestMultiPart(req *http.Request, s *server.Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	return rr
}
