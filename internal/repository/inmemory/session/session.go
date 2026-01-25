package session_repository

import (
	"fmt"

	"github.com/AlexMickh/shop-backend/internal/errs"
	"github.com/AlexMickh/shop-backend/internal/models"
	"github.com/AlexMickh/shop-backend/pkg/cash"
)

type Cash[K comparable, V cash.Expireser] interface {
	Put(key K, value V)
	GetWithDelete(key K) (V, error)
}

type SessionRepository struct {
	cash Cash[string, models.Session]
}

func New(cash Cash[string, models.Session]) *SessionRepository {
	return &SessionRepository{
		cash: cash,
	}
}

func (s *SessionRepository) SaveSession(session models.Session) {
	s.cash.Put(session.Token, session)
}

func (s *SessionRepository) SessionByToken(token string) (models.Session, error) {
	const op = "repository.inmemory.SessionByToken"

	session, err := s.cash.GetWithDelete(token)
	if err != nil {
		return models.Session{}, fmt.Errorf("%s: %w", op, errs.ErrSessionNotFound)
	}

	return session, nil
}
