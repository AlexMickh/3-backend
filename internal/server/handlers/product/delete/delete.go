package delete_product

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
)

type ProductDeleter interface {
	DeleteProduct(ctx context.Context, id int64) error
}

// New godoc
//
//	@Summary		delete product
//	@Description	delete product
//	@Tags			admin
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"product id"
//	@Success		204
//	@Failure		400	{object}	response.ErrorResponse
//	@Failure		401	{object}	response.ErrorResponse
//	@Failure		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Security		AdminAuth
//	@Router			/admin/products/{id} [delete]
func New(productDeleter ProductDeleter) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.product.delete.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			log.Error("failed to parse id", logger.Err(err))
			return response.Error("failed to parse id", http.StatusBadRequest)
		}

		if err = productDeleter.DeleteProduct(ctx, id); err != nil {
			if errors.Is(err, errs.ErrProductNotFound) {
				log.Error(errs.ErrProductNotFound.Error())
				return response.Error(errs.ErrProductNotFound.Error(), http.StatusNotFound)
			}

			log.Error("failed to delete product", logger.Err(err))
			return response.Error("failed to delete product", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusNoContent)

		return nil
	}
}
