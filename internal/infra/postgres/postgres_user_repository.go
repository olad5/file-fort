package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/olad5/file-fort/internal/domain"
	"github.com/olad5/file-fort/internal/infra"
)

type PostgresUserRepository struct {
	connection *sqlx.DB
}

func NewPostgresUserRepo(ctx context.Context, connection *sqlx.DB) (*PostgresUserRepository, error) {
	if connection == nil {
		return &PostgresUserRepository{}, fmt.Errorf("Failed to create PostgresFileRepository: connection is nil")
	}

	return &PostgresUserRepository{connection: connection}, nil
}

func (p *PostgresUserRepository) CreateUser(ctx context.Context, user domain.User) error {
	const query = `
    INSERT INTO users
      (id, first_name, last_name, email, password, role) 
    VALUES 
      (:id, :first_name, :last_name, :email, :password, :role)
  `

	_, err := p.connection.NamedExec(query, toSqlxUser(user))
	if err != nil {
		return fmt.Errorf("error creating user in the db: %w", err)
	}
	return nil
}

func (p *PostgresUserRepository) GetUserByEmail(ctx context.Context, userEmail string) (domain.User, error) {
	var user SqlxUser

	err := p.connection.Get(&user, "SELECT * FROM users WHERE email = $1", userEmail)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return domain.User{}, infra.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("error getting user by email: %w", err)
	}
	return toUser(user), nil
}

func (p *PostgresUserRepository) GetUserByUserId(ctx context.Context, userId uuid.UUID) (domain.User, error) {
	var user SqlxUser

	err := p.connection.Get(&user, "SELECT * FROM users WHERE id = $1", userId)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return domain.User{}, infra.ErrUserNotFound
		}
		return domain.User{}, fmt.Errorf("error getting user by userId: %w", err)
	}
	return toUser(user), nil
}

type SqlxUser struct {
	ID        uuid.UUID   `db:"id"`
	Email     string      `db:"email"`
	FirstName string      `db:"first_name"`
	LastName  string      `db:"last_name"`
	Password  string      `db:"password"`
	Role      domain.Role `db:"role"`
	CreatedAt time.Time   `db:"created_at"`
	UpdatedAt time.Time   `db:"updated_at"`
}

func toUser(u SqlxUser) domain.User {
	return domain.User{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Password:  u.Password,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toSqlxUser(u domain.User) SqlxUser {
	return SqlxUser{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Password:  u.Password,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
