package session_service

import (
	"fmt"
	"time"

	"github.com/AlexMickh/shop-backend/internal/dtos"
	"github.com/AlexMickh/shop-backend/internal/models"
)

type SessionRepository interface {
	SaveSession(session models.Session)
	SessionByToken(token string) (models.Session, error)
}

type JwtManager interface {
	NewJwt(userID int64, role models.UserRole) (string, error)
	NewRefresh() (string, error)
	Validate(token string) (int64, error)
}

type SessionService struct {
	repository SessionRepository
	jwtManager JwtManager
	sessionTtl time.Duration
}

func New(repository SessionRepository, jwtManager JwtManager, sessionTtl time.Duration) *SessionService {
	return &SessionService{
		repository: repository,
		jwtManager: jwtManager,
		sessionTtl: sessionTtl,
	}
}

func (s *SessionService) CreateSession(userID int64, role models.UserRole) (string, string, error) {
	const op = "services.session.CreateSession"

	accessToken, err := s.jwtManager.NewJwt(userID, role)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := s.jwtManager.NewRefresh()
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	session := models.Session{
		Token:          refreshToken,
		UserID:         userID,
		Role:           role,
		ExpiresAtField: time.Now().Add(s.sessionTtl),
	}
	s.repository.SaveSession(session)

	return accessToken, refreshToken, nil
}

func (s *SessionService) Refresh(req dtos.RefreshRequest) (string, string, error) {
	const op = "services.session.Refresh"

	session, err := s.repository.SessionByToken(req.RefreshToken)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	accessToken, err := s.jwtManager.NewJwt(session.UserID, session.Role)
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
		Role:           session.Role,
		ExpiresAtField: time.Now().Add(s.sessionTtl),
	}
	s.repository.SaveSession(newSession)

	return accessToken, refreshToken, nil
}

func (s *SessionService) ValidateJwt(token string) (int64, error) {
	const op = "services.session.ValidateJwt"

	userID, err := s.jwtManager.Validate(token)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}
