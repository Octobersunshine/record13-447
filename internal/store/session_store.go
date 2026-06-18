package store

import (
	"errors"
	"session-management/internal/model"
	"sync"
	"time"
)

var (
	ErrSessionNotFound = errors.New("session not found")
	ErrUserNotFound    = errors.New("user has no active sessions")
)

type SessionStore interface {
	Create(session *model.Session) error
	GetByID(id string) (*model.Session, error)
	GetByUserID(userID string) ([]*model.Session, error)
	GetByToken(token string) (*model.Session, error)
	Update(session *model.Session) error
	Delete(id string) error
	FreezeByUserID(userID string, reason string) ([]*model.Session, error)
	UnfreezeByUserID(userID string, reason string) ([]*model.Session, error)
	List() ([]*model.Session, error)
}

type InMemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*model.Session
}

func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]*model.Session),
	}
}

func (s *InMemorySessionStore) Create(session *model.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	s.sessions[session.ID] = session
	return nil
}

func (s *InMemorySessionStore) GetByID(id string) (*model.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[id]
	if !exists {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (s *InMemorySessionStore) GetByUserID(userID string) ([]*model.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.Session
	for _, session := range s.sessions {
		if session.UserID == userID {
			result = append(result, session)
		}
	}

	if len(result) == 0 {
		return nil, ErrUserNotFound
	}

	return result, nil
}

func (s *InMemorySessionStore) GetByToken(token string) (*model.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, session := range s.sessions {
		if session.Token == token {
			return session, nil
		}
	}
	return nil, ErrSessionNotFound
}

func (s *InMemorySessionStore) Update(session *model.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.sessions[session.ID]
	if !exists {
		return ErrSessionNotFound
	}

	session.UpdatedAt = time.Now()
	s.sessions[session.ID] = session
	return nil
}

func (s *InMemorySessionStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
	return nil
}

func (s *InMemorySessionStore) FreezeByUserID(userID string, reason string) ([]*model.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var updatedSessions []*model.Session
	for _, session := range s.sessions {
		if session.UserID == userID && session.Status == model.SessionStatusActive {
			session.Status = model.SessionStatusFrozen
			session.UpdatedAt = time.Now()
			updatedSessions = append(updatedSessions, session)
		}
	}

	if len(updatedSessions) == 0 {
		return nil, ErrUserNotFound
	}

	return updatedSessions, nil
}

func (s *InMemorySessionStore) UnfreezeByUserID(userID string, reason string) ([]*model.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var updatedSessions []*model.Session
	for _, session := range s.sessions {
		if session.UserID == userID && session.Status == model.SessionStatusFrozen {
			session.Status = model.SessionStatusActive
			session.UpdatedAt = time.Now()
			updatedSessions = append(updatedSessions, session)
		}
	}

	if len(updatedSessions) == 0 {
		return nil, ErrUserNotFound
	}

	return updatedSessions, nil
}

func (s *InMemorySessionStore) List() ([]*model.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.Session
	for _, session := range s.sessions {
		result = append(result, session)
	}
	return result, nil
}
