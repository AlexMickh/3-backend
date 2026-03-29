package product_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type ProductRepository struct {
	db           DB
	queryBuilder goqu.DialectWrapper
}

func New(db DB) *ProductRepository {
	return &ProductRepository{
		db:           db,
		queryBuilder: goqu.Dialect("postgres"),
	}
}

func (p *ProductRepository) SaveProduct(ctx context.Context, product *models.Product) error {
	const op = "repository.postgres.product.SaveProduct"

	query := `INSERT INTO products 
			  (id, category_id, name, description, price, quantity, existing_sizes, image_url)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := p.db.Exec(
		ctx,
		query,
		product.ID,
		product.Category.ID,
		product.Name,
		product.Description,
		product.Price,
		product.Quantity,
		product.ExistingSizes,
		product.ImageUrl,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *ProductRepository) ProductById(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	const op = "repository.postgres.product.ProductById"

	query, args, err := p.queryBuilder.From("products").
		Select(
			"products.name", "products.description", "products.price", "products.quantity",
			"products.existing_sizes", "products.image_url", "products.discount",
			"products.discount_expires_at", "categories.id", "categories.name",
		).
		Join(
			goqu.T("categories"),
			goqu.On(goqu.Ex{"products.category_id": goqu.I("categories.id")}),
		).
		Where(goqu.Ex{"products.id": id}).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	fmt.Println(query)

	product := new(models.Product)
	err = p.db.QueryRow(ctx, query, args...).Scan(
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Quantity,
		&product.ExistingSizes,
		&product.ImageUrl,
		&product.Discount,
		&product.DiscountExpiresAt,
		&product.Category.ID,
		&product.Category.Name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, errs.ErrProductNotFound)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	product.ID = id

	return product, nil
}

func (p *ProductRepository) ProductCards(
	ctx context.Context,
	page int,
	popularity bool,
	price int,
	categoryId uuid.UUID,
	search string,
) ([]models.ProductCard, error) {
	const op = "repository.postgres.product.ProductCards"

	filter := goqu.Ex{}
	groupBy := make([]exp.OrderedExpression, 0)

	if categoryId != uuid.Max {
		filter["category_id"] = categoryId
	}

	if search != "" {
		filter["name"] = goqu.Op{"like": search}
	}

	if popularity {
		groupBy = append(groupBy, goqu.C("pieces_sold").Desc())
	}

	switch price {
	case 1:
		groupBy = append(groupBy, goqu.C("price - price / 100 * discount").Desc())
	case 0:
		groupBy = append(groupBy, goqu.C("price - price / 100 * discount").Asc())
	}

	query, args, err := p.queryBuilder.From("products").
		Select("id", "name", "price", "image_url", "discount", "discount_expires_at").
		Where(filter).
		Limit(10).
		Offset(uint(page * 10)).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := p.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	products := make([]models.ProductCard, 0)
	for rows.Next() {
		var product models.ProductCard

		err = rows.Scan(
			&product.ID,
			&product.Name,
			&product.Price,
			&product.ImageUrl,
			&product.Discount,
			&product.DiscountExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		products = append(products, product)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(products) == 0 {
		return nil, fmt.Errorf("%s: %w", op, errs.ErrProductNotFound)
	}

	return products, nil
}

func (p *ProductRepository) UpdateProduct(ctx context.Context, productToUpdate *models.Product) error {
	const op = "repository.postgres.product.UpdateProduct"

	// query, args, err := p.queryBuilder.Update("products").
	// 	Set(goqu.Record{
	// 		"name":                productToUpdate.Name,
	// 		"description":         productToUpdate.Description,
	// 		"price":               productToUpdate.Price,
	// 		"quantity":            productToUpdate.Quantity,
	// 		"existing_sizes":      productToUpdate.ExistingSizes,
	// 		"image_url":           productToUpdate.ImageUrl,
	// 		"discount":            productToUpdate.Discount,
	// 		"discount_expires_at": productToUpdate.DiscountExpiresAt,
	// 		"updated_at":          time.Now(),
	// 	}).
	// 	Where(goqu.Ex{"id": productToUpdate.ID}).
	// 	ToSQL()
	// if err != nil {
	// 	return fmt.Errorf("%s: %w", op, err)
	// }

	query := `UPDATE products
			  SET name = $1, 
			  	  description = $2, 
				  price = $3, 
				  quantity = $4, 
				  existing_sizes = $5, 
				  image_url = $6, 
				  discount = $7, 
				  discount_expires_at = $8
			  WHERE id = $9`

	result, err := p.db.Exec(
		ctx,
		query,
		productToUpdate.Name,
		productToUpdate.Description,
		productToUpdate.Price,
		productToUpdate.Quantity,
		productToUpdate.ExistingSizes,
		productToUpdate.ImageUrl,
		productToUpdate.Discount,
		productToUpdate.DiscountExpiresAt,
		productToUpdate.ID,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, errs.ErrProductNotFound)
	}

	return nil
}

func (p *ProductRepository) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	const op = "repository.postgres.product.DeleteProduct"

	query, args, err := p.queryBuilder.Delete("products").
		Where(goqu.Ex{"id": id}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = p.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
