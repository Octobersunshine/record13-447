package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"session-management/internal/model"
	"sync"
	"time"
)

var (
	ErrPersistentStoreError = errors.New("persistent store error")
)

type PersistentSessionStore interface {
	SessionStore
	Load() error
	Save() error
}

type FileSessionStore struct {
	mu       sync.RWMutex
	filePath string
	sessions map[string]*model.Session
}

func NewFileSessionStore(filePath string) (*FileSessionStore, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	store := &FileSessionStore{
		filePath: filePath,
		sessions: make(map[string]*model.Session),
	}

	if err := store.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	return store, nil
}

func (s *FileSessionStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	var sessions []*model.Session
	if err := json.Unmarshal(data, &sessions); err != nil {
		return err
	}

	s.sessions = make(map[string]*model.Session)
	for _, session := range sessions {
		s.sessions[session.ID] = session
	}

	return nil
}

func (s *FileSessionStore) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var sessions []*model.Session
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}

	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := s.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, s.filePath)
}

func (s *FileSessionStore) Create(session *model.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	s.sessions[session.ID] = session
	return s.saveLocked()
}

func (s *FileSessionStore) GetByID(id string) (*model.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, exists := s.sessions[id]
	if !exists {
		return nil, ErrSessionNotFound
	}
	return session, nil
}

func (s *FileSessionStore) GetByUserID(userID string) ([]*model.Session, error) {
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

func (s *FileSessionStore) GetByToken(token string) (*model.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, session := range s.sessions {
		if session.Token == token {
			return session, nil
		}
	}
	return nil, ErrSessionNotFound
}

func (s *FileSessionStore) Update(session *model.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.sessions[session.ID]
	if !exists {
		return ErrSessionNotFound
	}

	session.UpdatedAt = time.Now()
	s.sessions[session.ID] = session
	return s.saveLocked()
}

func (s *FileSessionStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
	return s.saveLocked()
}

func (s *FileSessionStore) FreezeByUserID(userID string, reason string) ([]*model.Session, error) {
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

	if err := s.saveLocked(); err != nil {
		return nil, err
	}

	return updatedSessions, nil
}

func (s *FileSessionStore) UnfreezeByUserID(userID string, reason string) ([]*model.Session, error) {
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

	if err := s.saveLocked(); err != nil {
		return nil, err
	}

	return updatedSessions, nil
}

func (s *FileSessionStore) List() ([]*model.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.Session
	for _, session := range s.sessions {
		result = append(result, session)
	}
	return result, nil
}

func (s *FileSessionStore) saveLocked() error {
	var sessions []*model.Session
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}

	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := s.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpPath, s.filePath)
}
