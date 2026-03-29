package category_service

import (
	"context"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CategoryRepository interface {
	SaveCategory(ctx context.Context, category models.Category) (uuid.UUID, error)
	DeleteCategory(ctx context.Context, id uuid.UUID) error
	AllCategories(ctx context.Context) ([]models.Category, error)
}

type CategoryService struct {
	categoryRepository CategoryRepository
	validator          *validator.Validate
}

func New(categoryRepository CategoryRepository, validator *validator.Validate) *CategoryService {
	return &CategoryService{
		categoryRepository: categoryRepository,
		validator:          validator,
	}
}

func (c *CategoryService) CreateCategory(ctx context.Context, req dtos.CreateCategoryRequest) (uuid.UUID, error) {
	const op = "services.category.CreateCategory"

	if err := c.validator.Struct(&req); err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	category := models.Category{
		Name: req.Name,
	}

	id, err := c.categoryRepository.SaveCategory(ctx, category)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (c *CategoryService) DeleteCategory(ctx context.Context, id string) error {
	const op = "services.category.DeleteCategory"

	userId, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	err = c.categoryRepository.DeleteCategory(ctx, userId)
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
