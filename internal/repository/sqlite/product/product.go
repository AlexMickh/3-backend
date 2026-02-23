package product_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

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
	db          DB
	builderPool sync.Pool
}

func New(db DB) *ProductRepository {
	return &ProductRepository{
		db: db,
		builderPool: sync.Pool{
			New: func() any {
				return new(strings.Builder)
			},
		},
	}
}

func (p *ProductRepository) SaveProduct(ctx context.Context, product *models.Product) (int64, error) {
	const op = "repository.sqlite.product.SaveProduct"

	query := `INSERT INTO products
			  (name, description, price, category_id, quantity, existing_sizes)
			  VALUES (?, ?, ?, ?, ?, ?, ?)
			  RETURNING id`

	existingSizes := new(strings.Builder)
	for i, v := range product.ExistingSizes {
		if i == len(product.ExistingSizes)-1 {
			existingSizes.Write([]byte(v))
		} else {
			existingSizes.Write([]byte(v + " "))
		}
	}

	var id int64
	err := p.db.QueryRowContext(
		ctx,
		query,
		product.Name,
		product.Description,
		product.Price,
		product.Category.ID,
		product.Quantity,
		existingSizes.String(),
	).Scan(&id)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, errs.ErrProductAlreadyExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (p *ProductRepository) SaveImage(ctx context.Context, id int64, imageUrl string) error {
	const op = "repository.sqlite.product.SaveImage"

	query := "UPDATE products SET image_url = ? WHERE id = ?"
	_, err := p.db.ExecContext(ctx, query, imageUrl, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *ProductRepository) ProductById(ctx context.Context, id int64) (*models.Product, error) {
	const op = "repository.sqlite.product.ProductById"

	query := `SELECT p.name, p.description, p.price, p.quantity, p.existing_sizes, 
			  		 p.image_url, p.discount, p.discount_expires_at, c.id, c.name
			  FROM products AS p
			  JOIN categories AS c ON c.id = p.category_id
			  WHERE p.id = ?`
	product := new(models.Product)
	var sizes string
	var discountExpiresAt *string
	err := p.db.QueryRowContext(ctx, query, id).Scan(
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Quantity,
		&sizes,
		&product.ImageUrl,
		&product.Discount,
		&discountExpiresAt,
		&product.Category.ID,
		&product.Category.Name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, errs.ErrProductNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	for v := range strings.SplitSeq(sizes, " ") {
		product.ExistingSizes = append(product.ExistingSizes, models.ProductSize(v))
	}

	if discountExpiresAt != nil && *discountExpiresAt != "" {
		t, err := time.Parse("2006-01-02 15:04:05", *discountExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		product.DiscountExpiresAt = &t
	} else {
		product.DiscountExpiresAt = nil
	}

	return product, nil
}

func (p *ProductRepository) ProductCards(
	ctx context.Context,
	page int,
	popularity bool,
	price int,
	categoryId int64,
	search string,
) ([]models.ProductCard, error) {
	const op = "repository.sqlite.product.ProductCards"

	query := p.builderPool.Get().(*strings.Builder)
	query.WriteString("SELECT p.id, p.name, p.price, p.image_url, p.discount, p.discount_expires_at FROM products AS p")
	isFirst := true
	isFirstWhere := true

	args := make([]any, 0, 2)

	if categoryId != -1 {
		query.WriteString(" WHERE category_id = ?")
		args = append(args, categoryId)
		isFirstWhere = false
	}

	if search != "" {
		if isFirstWhere {
			query.WriteString(" WHERE name LIKE ? OR description LIKE ?")
			args = append(args, search)
			args = append(args, search)
		} else {
			query.WriteString(" AND name LIKE ? OR description LIKE ?")
			args = append(args, search)
			args = append(args, search)
		}
	}

	if popularity {
		query.WriteString(" ORDER BY p.pieces_sold DESC")
		isFirst = false
	}

	switch price {
	case 1:
		if isFirst {
			query.WriteString(" ORDER BY (p.price - p.price / 100 * p.discount) DESC")
			isFirst = false
		} else {
			query.WriteString(", (p.price - p.price / 100 * p.discount) DESC")
		}
	case 0:
		if isFirst {
			query.WriteString(" ORDER BY (p.price - p.price / 100 * p.discount)")
			isFirst = false
		} else {
			query.WriteString(", (p.price - p.price / 100 * p.discount)")
		}
	}

	query.WriteString(" LIMIT 10 OFFSET ?")
	args = append(args, page*10)
	s := query.String()
	fmt.Println(s)
	cards := make([]models.ProductCard, 0)
	rows, err := p.db.QueryContext(ctx, s, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	query.Reset()
	p.builderPool.Put(query)

	for rows.Next() {
		var card models.ProductCard
		var discountExpiresAt *string

		err = rows.Scan(
			&card.ID,
			&card.Name,
			&card.Price,
			&card.ImageUrl,
			&card.Discount,
			&discountExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		if discountExpiresAt != nil && *discountExpiresAt != "" {
			t, err := time.Parse("2006-01-02 15:04:05", *discountExpiresAt)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", op, err)
			}
			card.DiscountExpiresAt = &t
		} else {
			card.DiscountExpiresAt = nil
		}

		cards = append(cards, card)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(cards) == 0 {
		return nil, fmt.Errorf("%s: %w", op, errs.ErrProductNotFound)
	}

	return cards, nil
}

func (p *ProductRepository) UpdateProduct(ctx context.Context, productToUpdate *models.Product) error {
	const op = "repository.sqlite.product.UpdateProduct"

	query := new(strings.Builder)
	args := make([]any, 0)

	query.WriteString("UPDATE products SET")

	if productToUpdate.Name != "" {
		query.WriteString(" name = ?,")
		args = append(args, productToUpdate.Name)
	}

	if productToUpdate.Description != "" {
		query.WriteString(" description = ?,")
		args = append(args, productToUpdate.Description)
	}

	if productToUpdate.Price != -1 {
		query.WriteString(" price = ?,")
		args = append(args, productToUpdate.Price)
	}

	if productToUpdate.Quantity != -1 {
		query.WriteString(" quantity = ?,")
		args = append(args, productToUpdate.Quantity)
	}

	if len(productToUpdate.ExistingSizes) != 0 {
		existingSizes := new(strings.Builder)
		for i, v := range productToUpdate.ExistingSizes {
			if i == len(productToUpdate.ExistingSizes)-1 {
				existingSizes.Write([]byte(v))
			} else {
				existingSizes.Write([]byte(v + " "))
			}
		}
		query.WriteString(" existing_sizes = ?,")
		args = append(args, existingSizes.String())
	}

	if productToUpdate.ImageUrl != "" {
		query.WriteString(" image_url = ?,")
		args = append(args, productToUpdate.ImageUrl)
	}

	if productToUpdate.Discount != -1 {
		query.WriteString(" discount = ?,")
		args = append(args, productToUpdate.Discount)
	}

	if productToUpdate.DiscountExpiresAt != nil {
		query.WriteString(" discount_expires_at = ?,")
		args = append(args, *productToUpdate.DiscountExpiresAt)
	}

	query.WriteString(" updated_at = datetime('now) WHERE id = ?")
	args = append(args, productToUpdate.ID)

	result, err := p.db.ExecContext(ctx, query.String(), args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if n, err := result.RowsAffected(); err != nil || n == 0 {
		return fmt.Errorf("%s: %w", op, errs.ErrProductNotFound)
	}

	return nil
}

func (p *ProductRepository) DeleteProduct(ctx context.Context, id int64) error {
	const op = "repository.sqlite.product.DeleteProduct"

	query := "DELETE FROM products WHERE id = ?"
	result, err := p.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if n, err := result.RowsAffected(); err != nil || n == 0 {
		return fmt.Errorf("%s: %w", op, errs.ErrProductNotFound)
	}

	return nil
}
