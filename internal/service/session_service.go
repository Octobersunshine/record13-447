package service

import (
	"session-management/internal/model"
	"session-management/internal/store"

	"github.com/google/uuid"
)

type SessionService struct {
	store store.SessionStore
}

func NewSessionService(store store.SessionStore) *SessionService {
	return &SessionService{
		store: store,
	}
}

func (s *SessionService) CreateSession(userID, ip, userAgent, device string) (*model.Session, error) {
	session := &model.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     uuid.New().String(),
		Status:    model.SessionStatusActive,
		IP:        ip,
		UserAgent: userAgent,
		Device:    device,
	}

	err := s.store.Create(session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *SessionService) GetSessionByID(id string) (*model.Session, error) {
	return s.store.GetByID(id)
}

func (s *SessionService) GetSessionsByUserID(userID string) ([]*model.Session, error) {
	return s.store.GetByUserID(userID)
}

func (s *SessionService) FreezeUserSessions(userID, reason string) (*model.BatchSessionResponse, error) {
	allSessions, err := s.store.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	frozenSessions, err := s.store.FreezeByUserID(userID, reason)
	if err != nil {
		return nil, err
	}

	return &model.BatchSessionResponse{
		Sessions:    frozenSessions,
		TotalCount:  len(allSessions),
		UpdateCount: len(frozenSessions),
		Message:     "successfully frozen user sessions",
	}, nil
}

func (s *SessionService) UnfreezeUserSessions(userID, reason string) (*model.BatchSessionResponse, error) {
	allSessions, err := s.store.GetByUserID(userID)
	if err != nil {
		return nil, err
	}

	unfrozenSessions, err := s.store.UnfreezeByUserID(userID, reason)
	if err != nil {
		return nil, err
	}

	return &model.BatchSessionResponse{
		Sessions:    unfrozenSessions,
		TotalCount:  len(allSessions),
		UpdateCount: len(unfrozenSessions),
		Message:     "successfully unfrozen user sessions",
	}, nil
}

func (s *SessionService) ValidateSession(token string) (*model.Session, error) {
	session, err := s.store.GetByToken(token)
	if err != nil {
		return nil, err
	}

	if session.Status != model.SessionStatusActive {
		return nil, store.ErrSessionNotFound
	}

	return session, nil
}

func (s *SessionService) ListSessions() ([]*model.Session, error) {
	return s.store.List()
}

func (s *SessionService) RefreshCache() error {
	if cachedStore, ok := s.store.(interface{ RefreshCache() error }); ok {
		return cachedStore.RefreshCache()
	}
	return nil
}

func (s *SessionService) BatchFreezeUserSessions(userIDs []string, reason string) *model.BatchFreezeResponse {
	response := &model.BatchFreezeResponse{
		Results:     make([]model.UserFreezeResult, 0, len(userIDs)),
		TotalUsers:  len(userIDs),
		Message:     "batch freeze completed",
	}

	results := s.store.FreezeByUserIDs(userIDs, reason)

	for _, userID := range userIDs {
		sessions, ok := results[userID]
		if ok {
			response.Results = append(response.Results, model.UserFreezeResult{
				UserID:      userID,
				FrozenCount: len(sessions),
				Sessions:    sessions,
			})
			response.SuccessCount++
			response.TotalFrozen += len(sessions)
		} else {
			response.Results = append(response.Results, model.UserFreezeResult{
				UserID: userID,
				Error:  "user has no active sessions or user not found",
			})
			response.FailCount++
		}
	}

	return response
}

func (s *SessionService) BatchUnfreezeUserSessions(userIDs []string, reason string) *model.BatchUnfreezeResponse {
	response := &model.BatchUnfreezeResponse{
		Results:     make([]model.UserUnfreezeResult, 0, len(userIDs)),
		TotalUsers:  len(userIDs),
		Message:     "batch unfreeze completed",
	}

	results := s.store.UnfreezeByUserIDs(userIDs, reason)

	for _, userID := range userIDs {
		sessions, ok := results[userID]
		if ok {
			response.Results = append(response.Results, model.UserUnfreezeResult{
				UserID:         userID,
				UnfrozenCount:  len(sessions),
				Sessions:       sessions,
			})
			response.SuccessCount++
			response.TotalUnfrozen += len(sessions)
		} else {
			response.Results = append(response.Results, model.UserUnfreezeResult{
				UserID: userID,
				Error:  "user has no frozen sessions or user not found",
			})
			response.FailCount++
		}
	}

	return response
}
