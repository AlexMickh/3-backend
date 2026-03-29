package cart_router

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
	httppath "github.com/AlexMickh/shop-backend/pkg/utils/http-path"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
)

type CartService interface {
	AddToCart(ctx context.Context, req dtos.AddToCartRequest) (uuid.UUID, error)
	Cart(ctx context.Context, userId string) (models.Cart, error)
	DeleteItem(ctx context.Context, userId, productId string) error
	Clear(ctx context.Context, userId string) error
	Buy(ctx context.Context, userId string) (string, error)
}

type TokenValidator interface {
	ValidateJwt(token string) (int64, error)
}

type CartRouter struct {
	cartService    CartService
	tokenValidator TokenValidator
}

func New(cartService CartService, tokenValidator TokenValidator) *CartRouter {
	return &CartRouter{
		cartService:    cartService,
		tokenValidator: tokenValidator,
	}
}

func (c *CartRouter) RegisterRoute(r *chi.Mux) {
	r.Route("/carts", func(r chi.Router) {
		r.Use(middlewares.Login(c.tokenValidator))

		r.Post("/add/{product_id}", response.ErrorWrapper(c.Add))
		r.Get("/", response.ErrorWrapper(c.Get))
		r.Delete("/", response.ErrorWrapper(c.Clear))
		r.Delete("/{item_id}", response.ErrorWrapper(c.DeleteItem))
		r.Post("/buy", response.ErrorWrapper(c.Buy))
	})
}

// Add godoc
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
func (c *CartRouter) Add(w http.ResponseWriter, r *http.Request) error {
	const op = "routers.cart.Add"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	userId, ok := ctx.Value(middlewares.UserIdKey).(string)
	if !ok {
		log.Error("user id not found")
		return response.Error("user id not found", http.StatusUnauthorized)
	}

	req := dtos.AddToCartRequest{
		UserId: userId,
	}
	err := httppath.Decode(r, &req)
	if err != nil {
		log.Error("failed to decode path value", logger.Err(err))
		return response.Error("failed to decode path", http.StatusBadRequest)
	}

	// productId, err := strconv.ParseInt(r.PathValue("product_id"), 10, 64)
	// if err != nil || productId < 1 {
	// 	log.Error("product id is empty")
	// 	return response.Error("product id is required", http.StatusBadRequest)
	// }

	itemId, err := c.cartService.AddToCart(ctx, req)
	if err != nil {
		log.Error("failed to add to cart", logger.Err(err))
		return response.Error("failed to add to cart", http.StatusInternalServerError)
	}

	render.Status(r, 201)
	render.JSON(w, r, dtos.AddToCartResponse{
		ID: itemId.String(),
	})

	return nil
}

// Get godoc
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
func (c *CartRouter) Get(w http.ResponseWriter, r *http.Request) error {
	const op = "routers.cart.Get"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	userId, ok := ctx.Value(middlewares.UserIdKey).(string)
	if !ok {
		log.Error("user id not found")
		return response.Error("user id not found", http.StatusUnauthorized)
	}

	cart, err := c.cartService.Cart(ctx, userId)
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
			ID:                v.ID.String(),
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

// DeleteItem godoc
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
//	@Failure		500	{object}	response.ErrorResponse
//	@Security		UserAuth
//	@Router			/carts/{item_id} [delete]
func (c *CartRouter) DeleteItem(w http.ResponseWriter, r *http.Request) error {
	const op = "routers.cart.DeleteItem"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	userId, ok := ctx.Value(middlewares.UserIdKey).(string)
	if !ok {
		log.Error("user id not found")
		return response.Error("user id not found", http.StatusUnauthorized)
	}

	// itemId, err := strconv.ParseInt(r.PathValue("item_id"), 10, 64)
	// if err != nil || itemId < 1 {
	// 	log.Error("item id id is empty")
	// 	return response.Error("item id is required", http.StatusBadRequest)
	// }

	itemId := r.PathValue("item_id")

	err := c.cartService.DeleteItem(ctx, userId, itemId)
	if err != nil {
		log.Error("failed to delete item", logger.Err(err))
		return response.Error("failed to delete item", http.StatusInternalServerError)
	}

	render.NoContent(w, r)

	return nil
}

// Clear godoc
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
func (c *CartRouter) Clear(w http.ResponseWriter, r *http.Request) error {
	const op = "routers.cart.Clear"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	userId, ok := ctx.Value(middlewares.UserIdKey).(string)
	if !ok {
		log.Error("user id not found")
		return response.Error("user id not found", http.StatusUnauthorized)
	}

	err := c.cartService.Clear(ctx, userId)
	if err != nil {
		if errors.Is(err, errs.ErrCartEmpty) {
			log.Error("cart is empty")
			return response.Error(errs.ErrCartEmpty.Error(), http.StatusNotFound)
		}

		log.Error("failed to clear cart", logger.Err(err))
		return response.Error("failed to clear cart", http.StatusInternalServerError)
	}

	render.NoContent(w, r)

	return nil
}

// Buy godoc
//
//	@Summary		return link to pay
//	@Description	return link to pay
//	@Tags			carts
//	@Accept			json
//	@Produce		json
//	@Success		204	{object}	dtos.BuyResponse
//	@Success		401	{object}	response.ErrorResponse
//	@Success		404	{object}	response.ErrorResponse
//	@Failure		500	{object}	response.ErrorResponse
//	@Security		UserAuth
//	@Router			/carts/buy [post]
func (c *CartRouter) Buy(w http.ResponseWriter, r *http.Request) error {
	const op = "routers.cart.Buy"
	ctx := r.Context()
	log := logger.FromCtx(ctx).With(slog.String("op", op))

	userId, ok := ctx.Value("user_id").(string)
	if !ok {
		log.Error("failed to get user id")
		return response.Error("failed to get user id", http.StatusUnauthorized)
	}

	redirectUrl, err := c.cartService.Buy(ctx, userId)
	if err != nil {
		if errors.Is(err, errs.ErrCartEmpty) {
			log.Error(errs.ErrCartEmpty.Error())
			return response.Error(errs.ErrCartEmpty.Error(), http.StatusNotFound)
		}

		log.Error("failed to buy", logger.Err(err))
		return response.Error("failed to buy", http.StatusInternalServerError)
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, dtos.BuyResponse{
		RedirectUrl: redirectUrl,
	})

	return nil
}
