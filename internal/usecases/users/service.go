package users

import (
	"context"
	"errors"
	"strings"

	"github.com/olad5/go-cloud-backup-system/internal/services/auth"

	"github.com/google/uuid"
	"github.com/olad5/go-cloud-backup-system/config"
	"github.com/olad5/go-cloud-backup-system/internal/domain"
	"github.com/olad5/go-cloud-backup-system/internal/infra"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo       infra.UserRepository
	configurations *config.Configurations
	authService    auth.AuthService
}

var (
	ErrUserAlreadyExists = "email already exist"
	ErrPasswordIncorrect = "incorrect credentials"
	ErrUserNotFound      = "user not found"
	ErrInvalidToken      = errors.New("invalid token")
)

func NewUserService(userRepo infra.UserRepository, authService auth.AuthService, configurations *config.Configurations) (*UserService, error) {
	if userRepo == nil {
		return &UserService{}, errors.New("UserService failed to initialize, userRepo is nil")
	}
	if authService == nil {
		return &UserService{}, errors.New("UserService failed to initialize, authService is nil")
	}
	return &UserService{userRepo, configurations, authService}, nil
}

func (u *UserService) CreateUser(ctx context.Context, firstName, lastName, email, password string) (domain.User, error) {
	existingUser, err := u.userRepo.GetUserByEmail(ctx, email)
	if err == nil && existingUser.Email == email {
		return domain.User{}, errors.New(ErrUserAlreadyExists)
	}
	if err != nil {
		return domain.User{}, err
	}

	hashedPassword, err := hashAndSalt([]byte(password))
	newUser := domain.User{
		ID:        uuid.New(),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  hashedPassword,
		Role:      domain.RegularUserRole,
	}
	if err != nil {
		return domain.User{}, err
	}

	err = u.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		return domain.User{}, err
	}
	return newUser, nil
}

func (u *UserService) LogUserIn(ctx context.Context, email, password string) (string, error) {
	existingUser, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil && err.Error() == infra.ErrRecordNotFound {
		return "", errors.New(ErrUserNotFound)
	}
	if isPasswordCorrect := comparePasswords(existingUser.Password, []byte(password)); isPasswordCorrect == false {
		return "", errors.New(ErrPasswordIncorrect)
	}

	accessToken, err := u.authService.GenerateJWT(ctx, existingUser)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

func (u *UserService) VerifyUser(ctx context.Context, authHeader string) (string, error) {
	const Bearer = "Bearer "
	if authHeader != "" && strings.HasPrefix(authHeader, Bearer) {
		token := strings.TrimPrefix(authHeader, Bearer)
		return u.authService.DecodeJWT(ctx, token)
	}

	return "", ErrInvalidToken
}

func hashAndSalt(plainPassword []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(plainPassword, bcrypt.MinCost)
	if err != nil {
		return "", errors.New("error hashing password")
	}
	return string(hash), nil
}

func comparePasswords(hashedPassword string, plainPassword []byte) bool {
	byteHash := []byte(hashedPassword)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPassword)
	if err != nil {
		return false
	}

	return true
}
