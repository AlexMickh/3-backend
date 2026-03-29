package product_service

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ProductRepository interface {
	SaveProduct(ctx context.Context, product *models.Product) error
	ProductById(ctx context.Context, id uuid.UUID) (*models.Product, error)
	ProductCards(
		ctx context.Context,
		page int,
		popularity bool,
		price int,
		categoryId uuid.UUID,
		search string,
	) ([]models.ProductCard, error)
	UpdateProduct(ctx context.Context, productToUpdate *models.Product) error
	DeleteProduct(ctx context.Context, id uuid.UUID) error
}

type FileStorage interface {
	SaveImage(id uuid.UUID, image []byte) (string, error)
	DeleteImage(id uuid.UUID) error
}

type ProductService struct {
	productRepository ProductRepository
	fileStorage       FileStorage
	validator         *validator.Validate
}

func New(productRepository ProductRepository, fileStorage FileStorage, validator *validator.Validate) *ProductService {
	return &ProductService{
		productRepository: productRepository,
		fileStorage:       fileStorage,
		validator:         validator,
	}
}

func (p *ProductService) CreateProduct(ctx context.Context, req dtos.CreateProductRequest) (uuid.UUID, error) {
	const op = "services.product.CreateProduct"

	if err := p.validator.Struct(&req); err != nil {
		fmt.Println(err)
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	categoryId, err := uuid.Parse(req.CategoryID)
	if err != nil {
		fmt.Println(err)
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	userId, err := uuid.NewV7()
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	buf := bytes.NewBuffer(nil)

	if _, err = io.Copy(buf, req.Image); err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	imageUrl, err := p.fileStorage.SaveImage(userId, buf.Bytes())
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	product := models.Product{
		ID:          userId,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category: models.Category{
			ID: categoryId,
		},
		Quantity:      req.Quantity,
		ExistingSizes: convertSizes(req.ExistingSizes),
		ImageUrl:      imageUrl,
	}

	err = p.productRepository.SaveProduct(ctx, &product)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	return userId, nil
}

func (p *ProductService) ProductById(ctx context.Context, id string) (*models.Product, error) {
	const op = "services.product.ProductById"

	productId, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	product, err := p.productRepository.ProductById(ctx, productId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return product, nil
}

func (p *ProductService) ProductCards(ctx context.Context, req dtos.GetProductsRequest) ([]models.ProductCard, error) {
	const op = "services.product.ProductCards"

	err := p.validator.Struct(&req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	categoryId := uuid.Max // magic value for null id, because in this lib we don't have null value
	if req.CategoryID != "" {
		categoryId, err = uuid.Parse(req.CategoryID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
		}
	}

	products, err := p.productRepository.ProductCards(
		ctx,
		req.Page,
		req.Popularity,
		req.Price,
		categoryId,
		req.Search,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return products, nil
}

func (p *ProductService) UpdateProduct(ctx context.Context, req *dtos.UpdateProductRequest) error {
	const op = "services.product.UpdateProduct"

	if err := p.validator.Struct(&req); err != nil {
		return fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	productId, err := uuid.Parse(req.ID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	productToUpdate := &models.Product{
		ID:                productId,
		Name:              req.Name,
		Description:       req.Description,
		Price:             req.Price,
		Quantity:          req.Quantity,
		ExistingSizes:     convertSizes(req.ExistingSizes),
		Discount:          req.Discount,
		DiscountExpiresAt: req.DiscountExpiresAt,
	}

	if req.Image != nil {
		buf := bytes.NewBuffer(nil)

		if _, err = io.Copy(buf, req.Image); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		productToUpdate.ImageUrl, err = p.fileStorage.SaveImage(productId, buf.Bytes())
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

func (p *ProductService) DeleteProduct(ctx context.Context, id string) error {
	const op = "services.product.DeleteProduct"

	productId, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	err = p.productRepository.DeleteProduct(ctx, productId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = p.fileStorage.DeleteImage(productId)
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
