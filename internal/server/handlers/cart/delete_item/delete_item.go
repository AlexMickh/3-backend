package delete_cart_item

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/server/middlewares"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
)

type ItemDeleter interface {
	DeleteItem(ctx context.Context, userId, productId int64) error
}

// New godoc
//
//	@Summary		delete item from cart
//	@Description	delete item from cart
//	@Tags			carts
//	@Accept			json
//	@Produce		json
//	@Param			item_id	path	int	true	"item id"
//	@Success		204
//	@Success		400	{object}	response.ErrorResponse
//	@Success		401	{object}	response.ErrorResponse
//	@Success		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Security		UserAuth
//	@Router			/carts/{item_id} [delete]
func New(itemDeleter ItemDeleter) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.cart.delete_item.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		userId, ok := ctx.Value(middlewares.UserIdKey).(int64)
		if !ok {
			log.Error("user id not found")
			return response.Error("user id not found", http.StatusUnauthorized)
		}

		itemId, err := strconv.ParseInt(r.PathValue("item_id"), 10, 64)
		if err != nil || itemId < 1 {
			log.Error("item id id is empty")
			return response.Error("item id is required", http.StatusBadRequest)
		}

		err = itemDeleter.DeleteItem(ctx, userId, itemId)
		if err != nil {
			if errors.Is(err, errs.ErrProductNotFound) {
				log.Error("item not found")
				return response.Error(errs.ErrProductNotFound.Error(), http.StatusNotFound)
			}

			log.Error("failed to delete item", logger.Err(err))
			return response.Error("failed to delete item", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusNoContent)

		return nil
	}
}
