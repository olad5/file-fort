package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/olad5/go-cloud-backup-system/internal/domain"
	"github.com/olad5/go-cloud-backup-system/internal/infra"
)

type PostgresRepository struct {
	connection *pgx.Conn
}

func NewPostgresRepo(ctx context.Context, DatabaseUrl string) (*PostgresRepository, error) {
	connection, err := pgx.Connect(ctx, DatabaseUrl)
	if err != nil {
		return &PostgresRepository{}, fmt.Errorf("Failed to create PostgresRepository:  %w", err)
	}

	return &PostgresRepository{connection: connection}, nil
}

func (p *PostgresRepository) Migrate(ctx context.Context) error {
	query := `
  CREATE TABLE IF NOT EXISTS users(
      id UUID PRIMARY KEY,
      email TEXT NOT NULL UNIQUE,
      first_name TEXT NOT NULL,
      last_name TEXT NOT NULL,
      password TEXT NOT NULL,
      role TEXT NOT NULL,
      CHECK (role IN ('regular', 'admin'))
  );

`
	_, err := p.connection.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("Failed to Migrate Postgres DB:  %w", err)
	}

	return nil
}

func (p *PostgresRepository) CreateUser(ctx context.Context, user domain.User) error {
	err := p.Migrate(ctx)
	if err != nil {
		return err
	}

	query := `
    INSERT INTO users(id, first_name, last_name, email, password, role) values($1, $2, $3, $4, $5, $6) RETURNING id
  `

	err = p.connection.QueryRow(ctx, query, user.ID, user.FirstName, user.LastName, user.Email, user.Password, user.Role).Scan(nil)
	return err
}

func (p *PostgresRepository) GetUserByEmail(ctx context.Context, userEmail string) (domain.User, error) {
	err := p.Migrate(ctx)
	if err != nil {
		return domain.User{}, err
	}
	row := p.connection.QueryRow(ctx, "SELECT * FROM users WHERE email = $1", userEmail)

	var id uuid.UUID
	var first_name string
	var email string
	var last_name string
	var password string
	var role domain.UserRole

	if err := row.Scan(&id, &email, &first_name, &last_name, &password, &role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, errors.New(infra.ErrRecordNotFound)
		}
	}

	return domain.User{
		ID:        id,
		Email:     email,
		FirstName: first_name,
		LastName:  last_name,
		Password:  password,
		Role:      role,
	}, nil
}

func (p *PostgresRepository) Ping(ctx context.Context) error {
	err := p.connection.Ping(ctx)
	if err != nil {
		return fmt.Errorf("Failed to Ping Postgres DB:  %w", err)
	}

	return nil
}
