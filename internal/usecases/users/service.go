package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/olad5/file-fort/internal/services/auth"

	"github.com/google/uuid"
	"github.com/olad5/file-fort/internal/domain"
	"github.com/olad5/file-fort/internal/infra"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo    infra.UserRepository
	authService auth.AuthService
}

var (
	ErrUserAlreadyExists = errors.New("email already exist")
	ErrPasswordIncorrect = errors.New("invalid credentials")
	ErrInvalidToken      = errors.New("invalid token")
)

func NewUserService(userRepo infra.UserRepository, authService auth.AuthService) (*UserService, error) {
	if userRepo == nil {
		return &UserService{}, errors.New("UserService failed to initialize, userRepo is nil")
	}
	if authService == nil {
		return &UserService{}, errors.New("UserService failed to initialize, authService is nil")
	}
	return &UserService{userRepo, authService}, nil
}

func (u *UserService) CreateUser(ctx context.Context, firstName, lastName, email, password string) (domain.User, error) {
	existingUser, err := u.userRepo.GetUserByEmail(ctx, email)
	if err == nil && existingUser.Email == email {
		return domain.User{}, ErrUserAlreadyExists
	}

	hashedPassword, err := hashAndSalt([]byte(password))
	if err != nil {
		return domain.User{}, err
	}

	newUser := domain.User{
		ID:        uuid.New(),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  hashedPassword,
		Role:      domain.RoleUser,
	}

	err = u.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		return domain.User{}, err
	}
	return newUser, nil
}

func (u *UserService) LogUserIn(ctx context.Context, email, password string) (string, error) {
	existingUser, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", err
	}

	if isPasswordCorrect := comparePasswords(existingUser.Password, []byte(password)); isPasswordCorrect == false {
		return "", ErrPasswordIncorrect
	}

	accessToken, err := u.authService.GenerateJWT(ctx, existingUser)
	if err != nil {
		return "", err
	}
	return accessToken, nil
}

func (u *UserService) GetLoggedInUser(ctx context.Context) (domain.User, error) {
	userId, err := uuid.Parse(ctx.Value("userId").(string))
	if err != nil {
		return domain.User{}, fmt.Errorf("error parsing userId to UUID:%w", err)
	}

	existingUser, err := u.userRepo.GetUserByUserId(ctx, userId)
	if err != nil {
		return domain.User{}, err
	}
	return existingUser, nil
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
