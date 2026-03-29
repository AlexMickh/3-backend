package category_repository

import (
	"context"
	"errors"
	"fmt"

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
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type CategoryRepository struct {
	db           DB
	queryBuilder goqu.DialectWrapper
}

func New(db DB) *CategoryRepository {
	return &CategoryRepository{
		db:           db,
		queryBuilder: goqu.Dialect("postgres"),
	}
}

func (c *CategoryRepository) SaveCategory(ctx context.Context, category models.Category) (uuid.UUID, error) {
	const op = "repository.postgres.category.SaveCategory"

	query, args, err := c.queryBuilder.Insert("categories").
		Rows(goqu.Record{"name": category.Name}).
		Returning("id").
		ToSQL()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	var id uuid.UUID
	err = c.db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return uuid.UUID{}, fmt.Errorf("%s: %w", op, errs.ErrCategoryAlreadyExists)
			}
		}

		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (c *CategoryRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	const op = "repository.postgres.category.DeleteCategory"

	query, args, err := c.queryBuilder.Delete("categories").
		Where(goqu.Ex{"id": id}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	result, err := c.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, errs.ErrCategoryNotFound)
	}

	return nil
}

func (c *CategoryRepository) AllCategories(ctx context.Context) ([]models.Category, error) {
	const op = "repository.postgres.category.AllCategories"

	query, _, err := c.queryBuilder.From("categories").
		Select("id", "name").
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var categories []models.Category
	rows, err := c.db.Query(ctx, query)
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
