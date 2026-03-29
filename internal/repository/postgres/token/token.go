package token_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type TokenRepository struct {
	db           DB
	queryBuilder goqu.DialectWrapper
}

func New(db DB) *TokenRepository {
	return &TokenRepository{
		db:           db,
		queryBuilder: goqu.Dialect("postgres"),
	}
}

func (t *TokenRepository) SaveToken(ctx context.Context, token models.Token) error {
	const op = "repository.postgres.token.SaveToken"

	query, args, err := t.queryBuilder.Insert("tokens").
		Rows(goqu.Record{"token": token.Token, "user_id": token.UserID, "type": token.Type, "expires_at": token.ExpiresAt}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = t.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (t *TokenRepository) Token(ctx context.Context, token string, tokenType models.TokenType) (models.Token, error) {
	const op = "repository.postgres.token.Token"

	query, args, err := t.queryBuilder.From("tokens").
		Select("user_id", "expires_at").
		Where(goqu.Ex{"token": token, "type": tokenType}).
		ToSQL()
	if err != nil {
		return models.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	var tok models.Token
	err = t.db.QueryRow(ctx, query, args...).Scan(
		&tok.UserID,
		&tok.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Token{}, fmt.Errorf("%s: %w", op, errs.ErrTokenNotFound)
		}

		return models.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	tok.Token = token
	tok.Type = tokenType

	return tok, nil
}
