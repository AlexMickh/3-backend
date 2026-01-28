package get_categories

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type CategoryProvider interface {
	AllCategories(ctx context.Context) ([]models.Category, error)
}

// New godoc
//
//	@Summary		get all categories
//	@Description	get all categories
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dtos.GetCategoriesResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Router			/categories [get]
func New(validator *validator.Validate, categoryProvider CategoryProvider) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.category.get.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		categories, err := categoryProvider.AllCategories(ctx)
		if err != nil {
			log.Error("failed to get categories")
			return response.Error("failed to get categories", http.StatusInternalServerError)
		}

		render.JSON(w, r, dtos.ToGetCategoriesResponse(categories))

		return nil
	}
}
