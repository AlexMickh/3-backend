package session_service

import (
	"fmt"
	"time"

	"github.com/AlexMickh/shop-backend/internal/models"
)

type SessionRepository interface {
	SaveSession(session models.Session)
}

type JwtManager interface {
	NewJwt(userID int64, role models.UserRole) (string, error)
	NewRefresh() (string, error)
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
		ExpiresAtField: time.Now().Add(s.sessionTtl),
	}
	s.repository.SaveSession(session)

	return accessToken, refreshToken, nil
}
