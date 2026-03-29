package user_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type UserRepository struct {
	db           DB
	queryBuilder goqu.DialectWrapper
}

func New(db DB) *UserRepository {
	return &UserRepository{
		db:           db,
		queryBuilder: goqu.Dialect("postgres"),
	}
}

func (u *UserRepository) SaveUser(ctx context.Context, email, password string) (uuid.UUID, error) {
	const op = "repository.postgres.user.CreateUser"

	query, args, err := u.queryBuilder.Insert("users").
		Rows(goqu.Record{"email": email, "password": password}).
		Returning("id").
		ToSQL()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	var id uuid.UUID
	err = u.db.QueryRow(ctx, query, args...).Scan(&id)
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

	query, args, err := u.queryBuilder.From("users").
		Select("id", "password", "is_email_verified").
		Where(goqu.Ex{"email": email}).
		ToSQL()
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	var user models.User
	err = u.db.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Password,
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

func (u *UserRepository) VerifyEmail(ctx context.Context, id uuid.UUID) error {
	const op = "repository.postgres.user.VerifyEmail"

	query, args, err := u.queryBuilder.Update("users").
		Set(goqu.Record{"is_email_verified": true, "updated_at": time.Now()}).
		Where(goqu.Ex{"id": id}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	result, err := u.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, errs.ErrUserNotFound)
	}

	return nil
}

func (u *UserRepository) CanBuy(ctx context.Context, userId uuid.UUID) error {
	const op = "repository.postgres.user.CanBuy"

	query := "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND phone IS NOT NULL AND delivery_address IS NOT NULL)"
	var canBuy bool
	err := u.db.QueryRow(ctx, query, userId).Scan(&canBuy)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if !canBuy {
		return fmt.Errorf("%s: %w", op, errs.ErrUserCantBuy)
	}

	return nil
}
