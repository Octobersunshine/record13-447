package model

import "time"

type SessionStatus string

const (
	SessionStatusActive   SessionStatus = "active"
	SessionStatusFrozen   SessionStatus = "frozen"
	SessionStatusExpired  SessionStatus = "expired"
)

type Session struct {
	ID        string        `json:"id"`
	UserID    string        `json:"user_id"`
	Token     string        `json:"token"`
	Status    SessionStatus `json:"status"`
	IP        string        `json:"ip"`
	UserAgent string        `json:"user_agent"`
	Device    string        `json:"device"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	ExpiresAt time.Time     `json:"expires_at"`
}

type FreezeSessionRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Reason string `json:"reason"`
}

type UnfreezeSessionRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Reason string `json:"reason"`
}

type BatchFreezeRequest struct {
	UserIDs []string `json:"user_ids" binding:"required,min=1"`
	Reason  string   `json:"reason"`
}

type BatchUnfreezeRequest struct {
	UserIDs []string `json:"user_ids" binding:"required,min=1"`
	Reason  string   `json:"reason"`
}

type UserFreezeResult struct {
	UserID       string    `json:"user_id"`
	FrozenCount  int       `json:"frozen_count"`
	Sessions     []*Session `json:"sessions"`
	Error        string    `json:"error,omitempty"`
}

type UserUnfreezeResult struct {
	UserID         string    `json:"user_id"`
	UnfrozenCount  int       `json:"unfrozen_count"`
	Sessions       []*Session `json:"sessions"`
	Error          string    `json:"error,omitempty"`
}

type BatchFreezeResponse struct {
	Results     []UserFreezeResult `json:"results"`
	TotalUsers  int                `json:"total_users"`
	SuccessCount int               `json:"success_count"`
	FailCount    int               `json:"fail_count"`
	TotalFrozen  int               `json:"total_frozen"`
	Message      string            `json:"message"`
}

type BatchUnfreezeResponse struct {
	Results       []UserUnfreezeResult `json:"results"`
	TotalUsers    int                  `json:"total_users"`
	SuccessCount  int                  `json:"success_count"`
	FailCount     int                  `json:"fail_count"`
	TotalUnfrozen int                  `json:"total_unfrozen"`
	Message       string               `json:"message"`
}

type SessionResponse struct {
	Session *Session `json:"session,omitempty"`
	Message string   `json:"message"`
}

type BatchSessionResponse struct {
	Sessions    []*Session `json:"sessions"`
	TotalCount  int        `json:"total_count"`
	UpdateCount int        `json:"update_count"`
	Message     string     `json:"message"`
}
