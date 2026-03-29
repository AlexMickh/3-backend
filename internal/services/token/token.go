package token_service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/google/uuid"
)

type Repository interface {
	SaveToken(ctx context.Context, token models.Token) error
	Token(ctx context.Context, token string, tokenType models.TokenType) (models.Token, error)
}

type TokenService struct {
	repository          Repository
	verifyEmailTokenTtl time.Duration
}

func New(repository Repository, verifyEmailTokenTtl time.Duration) *TokenService {
	return &TokenService{
		repository:          repository,
		verifyEmailTokenTtl: verifyEmailTokenTtl,
	}
}

func (t *TokenService) CreateToken(ctx context.Context, userID uuid.UUID, tokenType models.TokenType) (models.Token, error) {
	const op = "services.token.CreateToken"

	var token models.Token
	var err error
	switch tokenType {
	case models.TokenTypeValidateEmail:
		token.Token, err = generateRandomString(32)
		if err != nil {
			return models.Token{}, fmt.Errorf("%s: %w", op, err)
		}

		token.ExpiresAt = time.Now().Add(t.verifyEmailTokenTtl)
		fmt.Println(token.ExpiresAt)
	default:
		return models.Token{}, fmt.Errorf("%s: unsupported token type", op)
	}

	token.UserID = userID
	token.Type = tokenType

	err = t.repository.SaveToken(ctx, token)
	if err != nil {
		return models.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (t *TokenService) UserIdByToken(ctx context.Context, token string, tokenType models.TokenType) (uuid.UUID, error) {
	const op = "services.token.UserIdByToken"

	tok, err := t.repository.Token(ctx, token, tokenType)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, err)
	}

	if tok.ExpiresAt.Compare(time.Now()) == -1 {
		return uuid.UUID{}, fmt.Errorf("%s: %w", op, errs.ErrTokenExpired)
	}

	return tok.UserID, nil
}

func generateRandomString(len int) (string, error) {
	const op = "services.token.generateRandomString"

	b := make([]byte, len)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
