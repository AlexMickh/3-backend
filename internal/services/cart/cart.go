package cart_service

import (
	"context"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/AlexMickh/shop-backend/pkg/utils/slice"
)

type CartRepository interface {
	AddProduct(ctx context.Context, userId, productId int64) (int64, error)
	Cart(ctx context.Context, userId int64) ([]*models.CartItem, error)
	DeleteItem(ctx context.Context, userId, productId int64) error
	Clear(ctx context.Context, userId int64) error
}

type CartService struct {
	cartRepository CartRepository
}

func New(cartRepository CartRepository) *CartService {
	return &CartService{
		cartRepository: cartRepository,
	}
}

func (c *CartService) AddToCart(ctx context.Context, userId, productId int64) (int64, error) {
	const op = "services.cart.AddToCart"

	id, err := c.cartRepository.AddProduct(ctx, userId, productId)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (c *CartService) Cart(ctx context.Context, userId int64) (models.Cart, error) {
	const op = "services.cart.Cart"

	cartItems, err := c.cartRepository.Cart(ctx, userId)
	if err != nil {
		return models.Cart{}, fmt.Errorf("%s: %w", op, err)
	}

	len := len(cartItems)
	var previousId int64 = 0
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

func (c *CartService) DeleteItem(ctx context.Context, userId, productId int64) error {
	const op = "services.cart.DeleteItem"

	err := c.cartRepository.DeleteItem(ctx, userId, productId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *CartService) Clear(ctx context.Context, userId int64) error {
	const op = "services.cart.Clear"

	err := c.cartRepository.Clear(ctx, userId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
