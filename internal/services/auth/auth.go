package auth_service

import (
	"context"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, email, password string) (uuid.UUID, error)
	UserByEmail(ctx context.Context, email string) (models.User, error)
}

type TokenService interface {
	CreateToken(ctx context.Context, userID uuid.UUID, tokenType models.TokenType) (models.Token, error)
}

type SessionService interface {
	CreateSession(userID uuid.UUID) (string, string, error)
}

type AuthService struct {
	userService    UserService
	tokenService   TokenService
	emailQueue     chan [2]string
	sessionService SessionService
	validator      *validator.Validate
}

func New(
	userService UserService,
	tokenService TokenService,
	emailQueue chan [2]string,
	sessionService SessionService,
	validator *validator.Validate,
) *AuthService {
	return &AuthService{
		userService:    userService,
		tokenService:   tokenService,
		emailQueue:     emailQueue,
		sessionService: sessionService,
		validator:      validator,
	}
}

func (a *AuthService) Register(ctx context.Context, req dtos.RegisterDto) (uuid.UUID, error) {
	const op = "services.auth.Register"

	if err := a.validator.Struct(&req); err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	userID, err := a.userService.CreateUser(ctx, req.Email, string(hashPassword))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	token, err := a.tokenService.CreateToken(ctx, userID, models.TokenTypeValidateEmail)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	a.emailQueue <- [2]string{req.Email, token.Token}

	return userID, nil
}

func (a *AuthService) Login(ctx context.Context, req dtos.LoginRequest) (string, string, error) {
	const op = "services.auth.Login"

	if err := a.validator.Struct(&req); err != nil {
		return "", "", fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	user, err := a.userService.UserByEmail(ctx, req.Email)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, errs.ErrUserNotFound)
	}

	accessToken, refreshToken, err := a.sessionService.CreateSession(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, refreshToken, nil
}
