package product_service

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/models"
)

type ProductRepository interface {
	SaveProduct(ctx context.Context, product *models.Product) (int64, error)
	ProductById(ctx context.Context, id int64) (*models.Product, error)
	ProductCards(
		ctx context.Context,
		page int,
		popularity bool,
		price int,
		categoryId int64,
		search string,
	) ([]models.ProductCard, error)
	UpdateProduct(ctx context.Context, productToUpdate *models.Product) error
	SaveImage(ctx context.Context, id int64, imageUrl string) error
	DeleteProduct(ctx context.Context, id int64) error
}

type FileStorage interface {
	SaveImage(id int64, image []byte) (string, error)
	DeleteImage(id int64) error
}

type ProductService struct {
	productRepository ProductRepository
	fileStorage       FileStorage
}

func New(productRepository ProductRepository, fileStorage FileStorage) *ProductService {
	return &ProductService{
		productRepository: productRepository,
		fileStorage:       fileStorage,
	}
}

func (p *ProductService) CreateProduct(ctx context.Context, req dtos.CreateProductRequest) (int64, error) {
	const op = "services.product.CreateProduct"

	product := models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category: models.Category{
			ID: req.CategoryID,
		},
		Quantity:      req.Quantity,
		ExistingSizes: convertSizes(req.ExistingSizes),
	}

	id, err := p.productRepository.SaveProduct(ctx, &product)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	buf := bytes.NewBuffer(nil)

	if _, err = io.Copy(buf, req.Image); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	product.ImageUrl, err = p.fileStorage.SaveImage(id, buf.Bytes())
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	err = p.productRepository.SaveImage(ctx, id, product.ImageUrl)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (p *ProductService) ProductById(ctx context.Context, id int64) (*models.Product, error) {
	const op = "services.product.ProductById"

	product, err := p.productRepository.ProductById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return product, nil
}

func (p *ProductService) ProductCards(ctx context.Context, req dtos.GetProductsRequest) ([]models.ProductCard, error) {
	const op = "services.product.ProductCards"

	products, err := p.productRepository.ProductCards(
		ctx,
		req.Page,
		req.Popularity,
		req.Price,
		req.CategoryID,
		req.Search,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return products, nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, req *dtos.UpdateProductRequest) error {
	const op = "services.product.UpdateProduct"

	productToUpdate := &models.Product{
		ID:                req.ID,
		Name:              req.Name,
		Description:       req.Description,
		Price:             req.Price,
		Quantity:          req.Quantity,
		ExistingSizes:     convertSizes(req.ExistingSizes),
		Discount:          req.Discount,
		DiscountExpiresAt: req.DiscountExpiresAt,
	}
	var err error

	if req.Image != nil {
		buf := bytes.NewBuffer(nil)

		if _, err = io.Copy(buf, req.Image); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		productToUpdate.ImageUrl, err = p.fileStorage.SaveImage(req.ID, buf.Bytes())
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	err = p.productRepository.UpdateProduct(ctx, productToUpdate)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (p *ProductService) DeleteProduct(ctx context.Context, id int64) error {
	const op = "services.product.DeleteProduct"

	err := p.productRepository.DeleteProduct(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = p.fileStorage.DeleteImage(id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func convertSizes(sizes []string) []models.ProductSize {
	modelSizes := make([]models.ProductSize, 0, len(sizes))

	for _, v := range sizes {
		modelSizes = append(modelSizes, models.ProductSize(v))
	}

	return modelSizes
}
