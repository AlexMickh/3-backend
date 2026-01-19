package user_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Postgres interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type UserRepository struct {
	db Postgres
}

func New(db Postgres) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (u *UserRepository) SaveUser(ctx context.Context, email, phone, password string) (uuid.UUID, error) {
	const op = "repository.postgres.user.SaveUser"

	sql := "INSERT INTO users (email, phone, password) VALUES ($1, $2, $3) RETURNING id"
	var id uuid.UUID
	err := u.db.QueryRow(ctx, sql, email, phone, password).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return uuid.UUID{}, fmt.Errorf("%s: %w", op, errs.ErrUserAlreadyExists)
			}
		}
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (u *UserRepository) UserByEmail(ctx context.Context, email string) (models.User, error) {
	const op = "repository.postgres.user.UserByEmail"

	query := "SELECT id, phone, password, role, is_email_verified FROM users WHERE email = $1"
	var user models.User
	err := u.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Phone,
		&user.Password,
		&user.Role,
		&user.IsEmailVerified,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, errs.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	user.Email = email

	return user, nil
}
