package cart_add

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/server/middlewares"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/go-chi/render"
)

type CartAdder interface {
	AddToCart(ctx context.Context, userId, productId int64) (int64, error)
}

// New godoc
//
//	@Summary		add product to cart
//	@Description	add product to cart
//	@Tags			carts
//	@Accept			json
//	@Produce		json
//	@Param			product_id	path		int	true	"product id"
//	@Success		201			{object}	dtos.AddToCartResponse
//	@Failure		400			{object}	response.ErrorResponse
//	@Failure		401			{object}	response.ErrorResponse
//	@Failure		404			{object}	response.ErrorResponse
//	@Failure		500			{object}	response.ErrorResponse
//	@Security		UserAuth
//	@Router			/carts/add/{product_id} [post]
func New(cartAdder CartAdder) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.cart.add.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		userId, ok := ctx.Value(middlewares.UserIdKey).(int64)
		if !ok {
			log.Error("user id not found")
			return response.Error("user id not found", http.StatusUnauthorized)
		}

		productId, err := strconv.ParseInt(r.PathValue("product_id"), 10, 64)
		if err != nil || productId < 1 {
			log.Error("product id is empty")
			return response.Error("product id is required", http.StatusBadRequest)
		}

		itemId, err := cartAdder.AddToCart(ctx, userId, productId)
		if err != nil {
			log.Error("failed to add to cart", logger.Err(err))
			return response.Error("failed to add to cart", http.StatusInternalServerError)
		}

		render.Status(r, 201)
		render.JSON(w, r, dtos.AddToCartResponse{
			ID: itemId,
		})

		return nil
	}
}
