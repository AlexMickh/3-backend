package token_repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/models"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type TokenRepository struct {
	db DB
}

func New(db DB) *TokenRepository {
	return &TokenRepository{
		db: db,
	}
}

func (t *TokenRepository) SaveToken(ctx context.Context, token models.Token) error {
	const op = "repository.postgres.token.SaveToken"

	query := "INSERT INTO tokens (token, user_id, type, expires_at) VALUES (?, ?, ?, ?)"
	_, err := t.db.ExecContext(ctx, query, token.Token, token.UserID, string(token.Type), token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
