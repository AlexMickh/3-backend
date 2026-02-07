package create_category

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

type CategoryCreater interface {
	CreateCategory(ctx context.Context, req dtos.CreateCategoryRequest) (int64, error)
}

// New godoc
//
//	@Summary		create new category
//	@Description	create new category
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			name	body		string	true	"category name"
//	@Success		201		{object}	dtos.CreateCategoryResponse
//	@Failure		400		{object}	response.ErrorResponse
//	@Failure		401		{object}	response.ErrorResponse
//	@Failure		409		{object}	response.ErrorResponse
//	@Failure		500		{object}	response.ErrorResponse
//	@Security		AdminAuth
//	@Router			/admin/category [post]
func New(validator *validator.Validate, categoryCreater CategoryCreater) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.category.create.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		var req dtos.CreateCategoryRequest
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			log.Error("failed to decode request", logger.Err(err))
			return response.Error("failed to decode request", http.StatusBadRequest)
		}
		defer r.Body.Close()

		if err = validator.Struct(&req); err != nil {
			log.Error("failed to validate request", logger.Err(err))
			return response.Error("failed to validate request", http.StatusBadRequest)
		}

		id, err := categoryCreater.CreateCategory(ctx, req)
		if err != nil {
			if errors.Is(err, errs.ErrCategoryAlreadyExists) {
				log.Error(errs.ErrCategoryAlreadyExists.Error())
				return response.Error(errs.ErrCategoryAlreadyExists.Error(), http.StatusConflict)
			}

			log.Error("failed to create category", logger.Err(err))
			return response.Error("failed to create category", http.StatusInternalServerError)
		}

		render.Status(r, http.StatusCreated)
		render.JSON(w, r, dtos.CreateCategoryResponse{
			ID: id,
		})

		return nil
	}
}
