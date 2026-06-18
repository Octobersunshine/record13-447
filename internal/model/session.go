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
