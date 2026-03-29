package session_service

import (
	"fmt"
	"time"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type SessionRepository interface {
	SaveSession(session models.Session)
	SessionByToken(token string) (models.Session, error)
}

type JwtManager interface {
	NewJwt(userID string) (string, error)
	NewRefresh() (string, error)
	Validate(token string) (int64, error)
}

type SessionService struct {
	repository SessionRepository
	jwtManager JwtManager
	sessionTtl time.Duration
	validator  *validator.Validate
}

func New(
	repository SessionRepository,
	jwtManager JwtManager,
	sessionTtl time.Duration,
	validator *validator.Validate,
) *SessionService {
	return &SessionService{
		repository: repository,
		jwtManager: jwtManager,
		sessionTtl: sessionTtl,
		validator:  validator,
	}
}

func (s *SessionService) CreateSession(userID uuid.UUID) (string, string, error) {
	const op = "services.session.CreateSession"

	accessToken, err := s.jwtManager.NewJwt(userID.String())
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := s.jwtManager.NewRefresh()
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	session := models.Session{
		Token:          refreshToken,
		UserID:         userID.String(),
		ExpiresAtField: time.Now().Add(s.sessionTtl),
	}
	s.repository.SaveSession(session)

	return accessToken, refreshToken, nil
}

func (s *SessionService) Refresh(req dtos.RefreshRequest) (string, string, error) {
	const op = "services.session.Refresh"

	if err := s.validator.Struct(&req); err != nil {
		return "", "", fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	session, err := s.repository.SessionByToken(req.RefreshToken)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	accessToken, err := s.jwtManager.NewJwt(session.UserID)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := s.jwtManager.NewRefresh()
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	newSession := models.Session{
		Token:          refreshToken,
		UserID:         session.UserID,
		ExpiresAtField: time.Now().Add(s.sessionTtl),
	}
	s.repository.SaveSession(newSession)

	return accessToken, refreshToken, nil
}

func (s *SessionService) ValidateJwt(token string) (int64, error) {
	const op = "services.session.ValidateJwt"

	if token == "" {
		return 0, fmt.Errorf("%s: %w", op, errs.ErrInvalidRequest)
	}

	userID, err := s.jwtManager.Validate(token)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}
