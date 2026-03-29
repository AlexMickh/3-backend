package cart_service

import (
	"context"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/AlexMickh/shop-backend/pkg/utils/slice"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type CartRepository interface {
	AddProduct(ctx context.Context, userId, productId uuid.UUID) (uuid.UUID, error)
	Cart(ctx context.Context, userId uuid.UUID) ([]*models.CartItem, error)
	DeleteItem(ctx context.Context, userId, productId uuid.UUID) error
	Clear(ctx context.Context, userId uuid.UUID) error
}

type UserService interface {
	CanBuy(ctx context.Context, userId uuid.UUID) error
}

type PaymentService interface {
	CreatePayment(userId uuid.UUID, price float32) (string, error)
}

type CartService struct {
	cartRepository CartRepository
	userService    UserService
	paymentService PaymentService
	validator      *validator.Validate
}

func New(cartRepository CartRepository /*userService UserService, paymentService PaymentService*/) *CartService {
	return &CartService{
		cartRepository: cartRepository,
		// userService:    userService,
		// paymentService: paymentService,
	}
}

func (c *CartService) AddToCart(ctx context.Context, req dtos.AddToCartRequest) (uuid.UUID, error) {
	const op = "services.cart.AddToCart"

	if err := c.validator.Struct(&req); err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	id, err := c.cartRepository.AddProduct(ctx, uuid.MustParse(req.UserId), uuid.MustParse(req.ProductId))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (c *CartService) Cart(ctx context.Context, userId string) (models.Cart, error) {
	const op = "services.cart.Cart"

	// if userId < 1 {
	// 	return models.Cart{}, fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	// }
	id, err := uuid.Parse(userId)
	if err != nil {
		return models.Cart{}, fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	cartItems, err := c.cartRepository.Cart(ctx, id)
	if err != nil {
		return models.Cart{}, fmt.Errorf("%s: %w", op, err)
	}

	len := len(cartItems)
	previousId := uuid.UUID{}
	var cart models.Cart

	for i := range len {
		if cartItems[i].ID == previousId {
			cartItems[i].Quantity += cartItems[i-1].Quantity
			slice.RemoveElement(cartItems, i-1)
			len--
			i--
		}
		previousId = cartItems[i].ID
		cart.Price += (cartItems[i].Price - cartItems[i].Price/100*cartItems[i].Discount)
	}

	return cart, nil
}

func (c *CartService) DeleteItem(ctx context.Context, userId, productId string) error {
	const op = "services.cart.DeleteItem"

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	productUUID, err := uuid.Parse(productId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	err = c.cartRepository.DeleteItem(ctx, userUUID, productUUID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *CartService) Clear(ctx context.Context, userId string) error {
	const op = "services.cart.Clear"

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	err = c.cartRepository.Clear(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *CartService) Buy(ctx context.Context, userId string) (string, error) {
	const op = "services.cart.Buy"

	userUUID, err := uuid.Parse(userId)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	err = c.userService.CanBuy(ctx, userUUID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	cart, err := c.Cart(ctx, userId)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	cartPrice := float32(cart.Price) / 100

	redirectUrl, err := c.paymentService.CreatePayment(userUUID, cartPrice)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return redirectUrl, nil
}
