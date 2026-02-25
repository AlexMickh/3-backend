package cart_repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
)

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type CartRepository struct {
	db DB
}

func New(db DB) *CartRepository {
	return &CartRepository{
		db: db,
	}
}

func (c *CartRepository) AddProduct(ctx context.Context, userId, productId int64) (int64, error) {
	const op = "repository.sqlite.cart.AddProduct"

	query := "INSERT INTO carts (user_id, product_id) VALUES (?, ?) RETURNING id"
	var id int64
	err := c.db.QueryRowContext(ctx, query, userId, productId).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (c *CartRepository) Cart(ctx context.Context, userId int64) ([]*models.CartItem, error) {
	const op = "repository.sqlite.cart.Cart"

	query := `SELECT p.id, p.name, p.price, p.image_url, p.discount, p.discount_expires_at
			  FROM cart AS c
			  JOIN product AS p ON c.product_id = p.id
			  WHERE c.user_id = ?
			  ORDER BY p.id`
	rows, err := c.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	cartItems := make([]*models.CartItem, 0)
	for rows.Next() {
		var discountExpiresAt *string
		cartItem := new(models.CartItem)

		err = rows.Scan(
			&cartItem.ID,
			&cartItem.Name,
			&cartItem.Price,
			&cartItem.Discount,
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
			cartItem.DiscountExpiresAt = &t
		} else {
			cartItem.DiscountExpiresAt = nil
		}

		cartItems = append(cartItems, cartItem)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(cartItems) == 0 {
		return nil, fmt.Errorf("%s: %w", op, errs.ErrCartEmpty)
	}

	return cartItems, nil
}

func (c *CartRepository) DeleteItem(ctx context.Context, userId, productId int64) error {
	const op = "repository.sqlite.cart.DeleteItem"

	query := "DELETE FROM carts WHERE user_id = ? AND product_id = ?"
	result, err := c.db.ExecContext(ctx, query, userId, productId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if n, err := result.RowsAffected(); err != nil || n == 0 {
		return fmt.Errorf("%s: %w", op, errs.ErrProductNotFound)
	}

	return nil
}

func (c *CartRepository) Clear(ctx context.Context, userId int64) error {
	const op = "repository.sqlite.cart.Clear"

	query := "DELETE FROM carts WHERE user_id = ?"
	result, err := c.db.ExecContext(ctx, query, userId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if n, err := result.RowsAffected(); err != nil || n == 0 {
		return fmt.Errorf("%s: %w", op, errs.ErrCartEmpty)
	}

	return nil
}
