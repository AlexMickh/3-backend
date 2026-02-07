package product_repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/mattn/go-sqlite3"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type ProductRepository struct {
	db DB
}

func New(db DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

func (p *ProductRepository) SaveProduct(ctx context.Context, product *models.Product) (int64, error) {
	const op = "repository.sqlite.product.SaveProduct"

	query := `INSERT INTO products
			  (name, description, price, category_id, quantity, existing_sizes, image_ulr)
			  VALUES (?, ?, ?, ?, ?, ?, ?)
			  RETURNING id`

	existingSizes := new(strings.Builder)
	for _, v := range product.ExistingSizes {
		existingSizes.Write([]byte(v + " "))
	}

	var id int64
	err := p.db.QueryRowContext(
		ctx,
		query,
		product.Name,
		product.Description,
		product.Price,
		product.CategoryID,
		product.Quantity,
		existingSizes.String(),
		product.ImageUrl,
	).Scan(&id)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, errs.ErrProductAlreadyExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
