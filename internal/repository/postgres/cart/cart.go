package cart_repository

import (
	"context"
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

type CartRepository struct {
	db           DB
	queryBuilder goqu.DialectWrapper
}

func New(db DB) *CartRepository {
	return &CartRepository{
		db:           db,
		queryBuilder: goqu.Dialect("postgres"),
	}
}

func (c *CartRepository) AddProduct(ctx context.Context, userId, productId uuid.UUID) (uuid.UUID, error) {
	const op = "repository.postgres.cart.AddProduct"

	query, args, err := c.queryBuilder.Insert("carts").
		Rows(goqu.Record{"user_id": userId, "product_id": productId}).
		Returning("id").
		ToSQL()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	var id uuid.UUID
	err = c.db.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (c *CartRepository) Cart(ctx context.Context, userId uuid.UUID) ([]*models.CartItem, error) {
	const op = "repository.postgres.cart.Cart"

	// TODO: add product_id
	query, args, err := c.queryBuilder.From("carts").
		Select("id", "name", "price", "image_url", "discount", "discount_expires_at").
		Join(
			goqu.T("products"),
			goqu.On(goqu.Ex{"carts.product_id": goqu.I("products.id")}),
		).
		Where(goqu.Ex{"user_id": userId}).
		GroupBy(goqu.C("products.id")).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := c.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	cartItems := make([]*models.CartItem, 0)
	for rows.Next() {
		cartItem := new(models.CartItem)

		err = rows.Scan(
			&cartItem.ID,
			&cartItem.Name,
			&cartItem.Price,
			&cartItem.ImageUrl,
			&cartItem.Discount,
			&cartItem.DiscountExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
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

func (c *CartRepository) DeleteItem(ctx context.Context, userId, productId uuid.UUID) error {
	const op = "repository.postgres.cart.DeleteItem"

	query, args, err := c.queryBuilder.Delete("carts").
		Where(goqu.Ex{"user_id": userId, "product_id": productId}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = c.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *CartRepository) Clear(ctx context.Context, userId uuid.UUID) error {
	const op = "repository.postgres.cart.Clear"

	query, args, err := c.queryBuilder.Delete("carts").
		Where(goqu.Ex{"user_id": userId}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = c.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
