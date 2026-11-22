package repository

import (
	"context"
)

type PostgresMessageRepository struct{}

func NewPostgresMessageRepository() *PostgresMessageRepository {
	return &PostgresMessageRepository{}
}

func (r *PostgresMessageRepository) GetMessage(ctx context.Context) (string, error) {
	return "", ErrNotImplemented
}
