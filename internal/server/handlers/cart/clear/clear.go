package clear_cart

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/server/middlewares"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
)

type Clearer interface {
	Clear(ctx context.Context, userId int64) error
}

// New godoc
//
//	@Summary		delete all item from cart
//	@Description	delete all item from cart
//	@Tags			carts
//	@Accept			json
//	@Produce		json
//	@Success		204
//	@Success		401	{object}	response.ErrorResponse
//	@Success		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Security		UserAuth
//	@Router			/carts [delete]
func New(clearer Clearer) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.cart.delete_item.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		userId, ok := ctx.Value(middlewares.UserIdKey).(int64)
		if !ok {
			log.Error("user id not found")
			return response.Error("user id not found", http.StatusUnauthorized)
		}

		err := clearer.Clear(ctx, userId)
		if err != nil {
			if errors.Is(err, errs.ErrCartEmpty) {
				log.Error("cart is empty")
				return response.Error(errs.ErrCartEmpty.Error(), http.StatusNotFound)
			}

			log.Error("failed to clear cart", logger.Err(err))
			return response.Error("failed to clear cart", http.StatusInternalServerError)
		}

		w.WriteHeader(http.StatusNoContent)

		return nil
	}
}
