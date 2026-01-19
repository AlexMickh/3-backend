package session_repository

import (
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
