package auth_service

import (
	"context"
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, email, phone, password string) (uuid.UUID, error)
	UserByEmail(ctx context.Context, email string) (models.User, error)
}

type TokenService interface {
	CreateToken(ctx context.Context, userID uuid.UUID, tokenType models.TokenType) (models.Token, error)
}

type SessionService interface {
	CreateSession(userID uuid.UUID, role models.UserRole) (string, string, error)
}

type AuthService struct {
	userService    UserService
	tokenService   TokenService
	emailQueue     chan [2]string
	sessionService SessionService
}

func New(
	userService UserService,
	tokenService TokenService,
	emailQueue chan [2]string,
	sessionService SessionService,
) *AuthService {
	return &AuthService{
		userService:    userService,
		tokenService:   tokenService,
		emailQueue:     emailQueue,
		sessionService: sessionService,
	}
}

func (a *AuthService) Register(ctx context.Context, req dtos.RegisterDto) (string, error) {
	const op = "services.auth.Register"

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	userID, err := a.userService.CreateUser(ctx, req.Email, req.Phone, string(hashPassword))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	token, err := a.tokenService.CreateToken(ctx, userID, models.TokenTypeValidateEmail)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	a.emailQueue <- [2]string{req.Email, token.Token}

	return userID.String(), nil
}

func (a *AuthService) Login(ctx context.Context, req dtos.LoginRequest) (string, string, error) {
	const op = "services.auth.Login"

	user, err := a.userService.UserByEmail(ctx, req.Email)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, errs.ErrUserNotFound)
	}

	accessToken, refreshToken, err := a.sessionService.CreateSession(user.ID, models.UserRoleUser)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, refreshToken, nil
}
