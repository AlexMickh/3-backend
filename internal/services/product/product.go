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
}

type FileStorage interface {
	SaveImage(image []byte) (string, error)
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
		CategoryID:  req.CategoryID,
		Quantity:    req.Quantity,
	}
	var err error

	for _, v := range req.ExistingSizes {
		switch v {
		case "xs":
			product.ExistingSizes = append(product.ExistingSizes, models.SizeXS)
		case "s":
			product.ExistingSizes = append(product.ExistingSizes, models.SizeS)
		case "m":
			product.ExistingSizes = append(product.ExistingSizes, models.SizeM)
		case "l":
			product.ExistingSizes = append(product.ExistingSizes, models.SizeL)
		case "xl":
			product.ExistingSizes = append(product.ExistingSizes, models.SizeXL)
		case "52":
			product.ExistingSizes = append(product.ExistingSizes, models.Size52)
		case "54":
			product.ExistingSizes = append(product.ExistingSizes, models.Size54)
		default:
			return 0, fmt.Errorf("%s: size not supported", op)
		}
	}

	buf := bytes.NewBuffer(nil)

	if _, err = io.Copy(buf, req.Image); err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	product.ImageUrl, err = p.fileStorage.SaveImage(buf.Bytes())
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := p.productRepository.SaveProduct(ctx, &product)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
