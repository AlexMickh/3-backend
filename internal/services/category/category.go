package category_service

import (
	"context"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/models"
)

type CategoryRepository interface {
	SaveCategory(ctx context.Context, name string) (int64, error)
	DeleteCategory(ctx context.Context, id int64) error
	AllCategories(ctx context.Context) ([]models.Category, error)
}

type CategoryService struct {
	categoryRepository CategoryRepository
}

func New(categoryRepository CategoryRepository) *CategoryService {
	return &CategoryService{
		categoryRepository: categoryRepository,
	}
}

func (c *CategoryService) CreateCategory(ctx context.Context, req dtos.CreateCategoryRequest) (int64, error) {
	const op = "services.category.CreateCategory"

	id, err := c.categoryRepository.SaveCategory(ctx, req.Name)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (c *CategoryService) DeleteCategory(ctx context.Context, id int64) error {
	const op = "services.category.DeleteCategory"

	err := c.categoryRepository.DeleteCategory(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *CategoryService) AllCategories(ctx context.Context) ([]models.Category, error) {
	const op = "services.category.AllCategories"

	categories, err := c.categoryRepository.AllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return categories, nil
}
