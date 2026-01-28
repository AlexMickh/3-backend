package user_service

import (
	"context"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
)

type UserRepository interface {
	SaveUser(ctx context.Context, email, password string) (int64, error)
	UserByEmail(ctx context.Context, email string) (models.User, error)
	VerifyEmail(ctx context.Context, token string) error
}

type UserService struct {
	userRepository UserRepository
}

func New(userRepository UserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (u *UserService) CreateUser(ctx context.Context, email, password string) (int64, error) {
	const op = "services.user.CreateUser"

	userID, err := u.userRepository.SaveUser(ctx, email, password)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
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

	err := u.userRepository.VerifyEmail(ctx, token)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
