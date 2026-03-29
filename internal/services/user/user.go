package user_service

import (
	"context"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/google/uuid"
)

type UserRepository interface {
	SaveUser(ctx context.Context, email, password string) (uuid.UUID, error)
	UserByEmail(ctx context.Context, email string) (models.User, error)
	VerifyEmail(ctx context.Context, id uuid.UUID) error
	CanBuy(ctx context.Context, userId uuid.UUID) error
}

type TokenService interface {
	UserIdByToken(ctx context.Context, token string, tokenType models.TokenType) (uuid.UUID, error)
}

type UserService struct {
	userRepository UserRepository
	tokenService   TokenService
}

func New(userRepository UserRepository, tokenService TokenService) *UserService {
	return &UserService{
		userRepository: userRepository,
		tokenService:   tokenService,
	}
}

func (u *UserService) CreateUser(ctx context.Context, email, password string) (uuid.UUID, error) {
	const op = "services.user.CreateUser"

	userID, err := u.userRepository.SaveUser(ctx, email, password)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

func (u *UserService) UserByEmail(ctx context.Context, email string) (models.User, error) {
	const op = "services.user.UserByEmail"

	user, err := u.userRepository.UserByEmail(ctx, email)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	if !user.IsEmailVerified {
		return models.User{}, fmt.Errorf("%s: %w", op, errs.ErrEmailNotVerified)
	}

	return user, nil
}

func (u *UserService) VerifyEmail(ctx context.Context, token string) error {
	const op = "services.user.VerifyEmail"

	id, err := u.tokenService.UserIdByToken(ctx, token, models.TokenTypeValidateEmail)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = u.userRepository.VerifyEmail(ctx, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// check is user has anought creds for delivery
func (u *UserService) CanBuy(ctx context.Context, userId uuid.UUID) error {
	const op = "services.user.CanBuy"

	err := u.userRepository.CanBuy(ctx, userId)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
