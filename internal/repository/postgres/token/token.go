package token_repository

import (
	"context"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Postgres interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type TokenRepository struct {
	db Postgres
}

func New(db Postgres) *TokenRepository {
	return &TokenRepository{
		db: db,
	}
}

func (t *TokenRepository) SaveToken(ctx context.Context, token models.Token) error {
	const op = "repository.postgres.token.SaveToken"

	query := "INSERT INTO tokens (token, user_id, type, expires_at) VALUES ($1, $2, $3, $4)"
	_, err := t.db.Exec(ctx, query, token.Token, token.UserID, string(token.Type), token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
