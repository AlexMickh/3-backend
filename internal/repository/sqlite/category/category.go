package category_repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/mattn/go-sqlite3"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type CategoryRepository struct {
	db DB
}

func New(db DB) *CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

func (c *CategoryRepository) SaveCategory(ctx context.Context, name string) (int64, error) {
	const op = "repository.sqlite.category.SaveCategory"

	query := "INSERT INTO categories (name) VALUES (?) RETURNING id"
	var id int64
	err := c.db.QueryRowContext(ctx, query, name).Scan(&id)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, errs.ErrCategoryAlreadyExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (c *CategoryRepository) DeleteCategory(ctx context.Context, id int64) error {
	const op = "repository.sqlite.category.DeleteCategory"

	query := "DELETE FROM categories WHERE id = ?"
	result, err := c.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if n, err := result.RowsAffected(); err != nil || n == 0 {
		return fmt.Errorf("%s: %w", op, errs.ErrCategoryNotFound)
	}

	return nil
}

func (c *CategoryRepository) AllCategories(ctx context.Context) ([]models.Category, error) {
	const op = "repository.sqlite.category.AllCategories"

	query := "SELECT id, name FROM categories"
	var categories []models.Category
	rows, err := c.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var category models.Category

		err = rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		categories = append(categories, category)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return categories, nil
}
