package category_router

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type CategoryService interface {
	AllCategories(ctx context.Context) ([]models.Category, error)
}

type CategoryRouter struct {
	categoryService CategoryService
}

func New(categoryService CategoryService) *CategoryRouter {
	return &CategoryRouter{
		categoryService: categoryService,
	}
}

func (c *CategoryRouter) RegisterRoute(r *chi.Mux) {
	r.Get("/categories", response.ErrorWrapper(c.All))
}

// All godoc
//
//	@Summary		get all categories
//	@Description	get all categories
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dtos.GetCategoriesResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/categories [get]
func (c *CategoryRouter) All(w http.ResponseWriter, r *http.Request) error {
	const op = "routers.categories.All"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	categories, err := c.categoryService.AllCategories(ctx)
	if err != nil {
		log.Error("failed to get categories")
		return response.Error("failed to get categories", http.StatusInternalServerError)
	}

	render.JSON(w, r, dtos.ToGetCategoriesResponse(categories))

	return nil
}
