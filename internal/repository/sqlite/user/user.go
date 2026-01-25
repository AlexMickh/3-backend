package user_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/mattn/go-sqlite3"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type UserRepository struct {
	db DB
}

func New(db DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (u *UserRepository) SaveUser(ctx context.Context, email, phone, password string) (int64, error) {
	const op = "repository.sqlite.user.CreateUser"

	query := "INSERT INTO users (email, phone, password) VALUES (?, ?, ?) RETURNING id"
	var id int64
	err := u.db.QueryRowContext(ctx, query, email, phone, password).Scan(&id)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, errs.ErrUserAlreadyExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (u *UserRepository) UserByEmail(ctx context.Context, email string) (models.User, error) {
	const op = "repository.sqlite.user.UserByEmail"

	query := "SELECT id, phone, password, role, is_email_verified FROM users WHERE email = ?"
	var user models.User
	err := u.db.QueryRowContext(ctx, query, email).Scan(
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

func (u *UserRepository) VerifyEmail(ctx context.Context, token string) error {
	const op = "repository.sqlite.user.VerifyEmail"

	query := `UPDATE users SET is_email_verified = 1, updated_at = DATETIME('now') WHERE id = (
    		      SELECT user_id FROM tokens WHERE token = ? AND expires_at > DATETIME('now')
			  )`
	result, err := u.db.ExecContext(ctx, query, token)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if n == 0 {
		return fmt.Errorf("%s: %w", op, errs.ErrUserNotFound)
	}

	return nil
}
