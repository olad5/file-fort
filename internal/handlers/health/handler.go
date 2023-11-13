package handlers

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/olad5/file-fort/internal/infra"
)

type HealthHandler struct {
	db    *sqlx.DB
	cache infra.Cache
}

func NewHealthHandler(ctx context.Context, db *sqlx.DB, cache infra.Cache) (*HealthHandler, error) {
	return &HealthHandler{db, cache}, nil
}
