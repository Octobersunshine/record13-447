package store

import (
	"session-management/internal/model"
	"sync"
	"time"
)

type CacheEntry struct {
	session    *model.Session
	expiryTime time.Time
}

type CachedSessionStore struct {
	persistentStore SessionStore
	cacheStore      *InMemorySessionStore
	cacheTTL        time.Duration
	userCache       map[string]time.Time
	mu              sync.RWMutex
}

func NewCachedSessionStore(persistentStore SessionStore, cacheTTL time.Duration) *CachedSessionStore {
	return &CachedSessionStore{
		persistentStore: persistentStore,
		cacheStore:      NewInMemorySessionStore(),
		cacheTTL:        cacheTTL,
		userCache:       make(map[string]time.Time),
	}
}

func (s *CachedSessionStore) Create(session *model.Session) error {
	if err := s.persistentStore.Create(session); err != nil {
		return err
	}

	s.cacheStore.Create(session)

	s.invalidateUserCache(session.UserID)

	return nil
}

func (s *CachedSessionStore) GetByID(id string) (*model.Session, error) {
	session, err := s.cacheStore.GetByID(id)
	if err == nil {
		return session, nil
	}

	session, err = s.persistentStore.GetByID(id)
	if err != nil {
		return nil, err
	}

	s.cacheStore.Create(session)

	return session, nil
}

func (s *CachedSessionStore) GetByUserID(userID string) ([]*model.Session, error) {
	s.mu.RLock()
	cacheExpiry, exists := s.userCache[userID]
	s.mu.RUnlock()

	now := time.Now()
	if exists && cacheExpiry.After(now) {
		sessions, err := s.cacheStore.GetByUserID(userID)
		if err == nil {
			return sessions, nil
		}
	}

	sessions, err := s.persistentStore.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	for _, session := range sessions {
		if _, err := s.cacheStore.GetByID(session.ID); err != nil {
			s.cacheStore.Create(session)
		}
	}

	s.mu.Lock()
	s.userCache[userID] = now.Add(s.cacheTTL)
	s.mu.Unlock()

	return sessions, nil
}

func (s *CachedSessionStore) GetByToken(token string) (*model.Session, error) {
	session, err := s.cacheStore.GetByToken(token)
	if err == nil {
		return session, nil
	}

	session, err = s.persistentStore.GetByToken(token)
	if err != nil {
		return nil, err
	}

	s.cacheStore.Create(session)

	return session, nil
}

func (s *CachedSessionStore) Update(session *model.Session) error {
	if err := s.persistentStore.Update(session); err != nil {
		return err
	}

	s.cacheStore.Update(session)

	s.invalidateUserCache(session.UserID)

	return nil
}

func (s *CachedSessionStore) Delete(id string) error {
	session, err := s.persistentStore.GetByID(id)
	if err == nil {
		s.invalidateUserCache(session.UserID)
	}

	if err := s.persistentStore.Delete(id); err != nil {
		return err
	}

	s.cacheStore.Delete(id)

	return nil
}

func (s *CachedSessionStore) FreezeByUserID(userID string, reason string) ([]*model.Session, error) {
	updatedSessions, err := s.persistentStore.FreezeByUserID(userID, reason)
	if err != nil {
		return nil, err
	}

	for _, session := range updatedSessions {
		if _, err := s.cacheStore.GetByID(session.ID); err == nil {
			s.cacheStore.Update(session)
		} else {
			s.cacheStore.Create(session)
		}
	}

	s.invalidateUserCache(userID)

	s.invalidateAllSessionsCache()

	return updatedSessions, nil
}

func (s *CachedSessionStore) UnfreezeByUserID(userID string, reason string) ([]*model.Session, error) {
	updatedSessions, err := s.persistentStore.UnfreezeByUserID(userID, reason)
	if err != nil {
		return nil, err
	}

	for _, session := range updatedSessions {
		if _, err := s.cacheStore.GetByID(session.ID); err == nil {
			s.cacheStore.Update(session)
		} else {
			s.cacheStore.Create(session)
		}
	}

	s.invalidateUserCache(userID)

	s.invalidateAllSessionsCache()

	return updatedSessions, nil
}

func (s *CachedSessionStore) FreezeByUserIDs(userIDs []string, reason string) map[string][]*model.Session {
	result := s.persistentStore.FreezeByUserIDs(userIDs, reason)

	for _, sessions := range result {
		for _, session := range sessions {
			if _, err := s.cacheStore.GetByID(session.ID); err == nil {
				s.cacheStore.Update(session)
			} else {
				s.cacheStore.Create(session)
			}
		}
	}

	for _, userID := range userIDs {
		s.invalidateUserCache(userID)
	}

	s.invalidateAllSessionsCache()

	return result
}

func (s *CachedSessionStore) UnfreezeByUserIDs(userIDs []string, reason string) map[string][]*model.Session {
	result := s.persistentStore.UnfreezeByUserIDs(userIDs, reason)

	for _, sessions := range result {
		for _, session := range sessions {
			if _, err := s.cacheStore.GetByID(session.ID); err == nil {
				s.cacheStore.Update(session)
			} else {
				s.cacheStore.Create(session)
			}
		}
	}

	for _, userID := range userIDs {
		s.invalidateUserCache(userID)
	}

	s.invalidateAllSessionsCache()

	return result
}

func (s *CachedSessionStore) List() ([]*model.Session, error) {
	return s.persistentStore.List()
}

func (s *CachedSessionStore) invalidateUserCache(userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.userCache, userID)
}

func (s *CachedSessionStore) invalidateAllSessionsCache() {
	s.cacheStore = NewInMemorySessionStore()

	s.mu.Lock()
	defer s.mu.Unlock()
	s.userCache = make(map[string]time.Time)
}

func (s *CachedSessionStore) RefreshCache() error {
	sessions, err := s.persistentStore.List()
	if err != nil {
		return err
	}

	newCacheStore := NewInMemorySessionStore()
	for _, session := range sessions {
		newCacheStore.Create(session)
	}

	s.cacheStore = newCacheStore

	s.mu.Lock()
	defer s.mu.Unlock()
	s.userCache = make(map[string]time.Time)

	return nil
}
