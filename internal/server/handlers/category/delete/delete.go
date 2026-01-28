package delete_category

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/go-playground/validator/v10"
)

type CategoryDeleter interface {
	DeleteCategory(ctx context.Context, id int64) error
}

// New godoc
//
//	@Summary		delete category
//	@Description	delete category
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"category id"
//	@Success		204
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Security		AdminAuth
//	@Router			/admin/delete-category/{id} [delete]
func New(validator *validator.Validate, categoryDeleter CategoryDeleter) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.category.delete.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		idStr := r.PathValue("id")
		if idStr == "" {
			log.Error("id is empty")
			return response.Error("id is required", http.StatusBadRequest)
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			log.Error("failed to parse id", logger.Err(err))
			return response.Error("failed to parse id", http.StatusBadRequest)
		}

		err = categoryDeleter.DeleteCategory(ctx, id)
		if err != nil {
			if errors.Is(err, errs.ErrCategoryNotFound) {
				log.Error(errs.ErrCategoryNotFound.Error())
				return response.Error(errs.ErrCategoryNotFound.Error(), http.StatusNotFound)
			}

			log.Error("failed to delete category", logger.Err(err))
			return response.Error("failed to delete category", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusNoContent)

		return nil
	}
}
