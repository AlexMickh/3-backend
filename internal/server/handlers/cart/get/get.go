package get_cart

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/AlexMickh/shop-backend/internal/server/middlewares"
	"github.com/AlexMickh/shop-backend/pkg/logger"
	"github.com/AlexMickh/shop-backend/pkg/response"
	"github.com/go-chi/render"
)

type CartProvider interface {
	Cart(ctx context.Context, userId int64) (models.Cart, error)
}

// New godoc
//
//	@Summary		get users cart
//	@Description	get users cart
//	@Tags			carts
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	dtos.GetCartResponse
//	@Success		204
//	@Success		401	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Security		UserAuth
//	@Router			/carts [get]
func New(cartProvider CartProvider) response.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		const op = "handlers.cart.get.New"
		ctx := r.Context()
		log := logger.FromCtx(ctx).With(slog.String("op", op))

		userId, ok := ctx.Value(middlewares.UserIdKey).(int64)
		if !ok {
			log.Error("user id not found")
			return response.Error("user id not found", http.StatusUnauthorized)
		}

		cart, err := cartProvider.Cart(ctx, userId)
		if err != nil {
			if errors.Is(err, errs.ErrCartEmpty) {
				render.Status(r, http.StatusNoContent)
				return nil
			}

			log.Error("failed to get cart", logger.Err(err))
			return response.Error("failed to get cart", http.StatusInternalServerError)
		}

		cartItems := make([]*dtos.CartItem, 0, len(cart.Products))
		finalPrice := 0
		for _, v := range cart.Products {
			cartItem := &dtos.CartItem{
				ID:                v.ID,
				Name:              v.Name,
				Price:             v.Price,
				ImageUrl:          v.ImageUrl,
				Discount:          v.Discount,
				DiscountExpiresAt: v.DiscountExpiresAt,
				Quantity:          v.Quantity,
			}
			cartItems = append(cartItems, cartItem)
		}

		render.JSON(w, r, dtos.GetCartResponse{
			Products: cartItems,
			Price:    finalPrice,
		})

		return nil
	}
}
