package handler

import (
	"errors"
	"net/http"
	"session-management/internal/model"
	"session-management/internal/service"
	"session-management/internal/store"

	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	service *service.SessionService
}

func NewSessionHandler(service *service.SessionService) *SessionHandler {
	return &SessionHandler{
		service: service,
	}
}

type CreateSessionRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Device    string `json:"device"`
}

func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"message": err.Error(),
		})
		return
	}

	session, err := h.service.CreateSession(req.UserID, req.IP, req.UserAgent, req.Device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create session",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, model.SessionResponse{
		Session: session,
		Message: "session created successfully",
	})
}

func (h *SessionHandler) FreezeSession(c *gin.Context) {
	var req model.FreezeSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"message": err.Error(),
		})
		return
	}

	response, err := h.service.FreezeUserSessions(req.UserID, req.Reason)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user not found",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to freeze sessions",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *SessionHandler) UnfreezeSession(c *gin.Context) {
	var req model.UnfreezeSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"message": err.Error(),
		})
		return
	}

	response, err := h.service.UnfreezeUserSessions(req.UserID, req.Reason)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user not found",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to unfreeze sessions",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *SessionHandler) GetUserSessions(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "missing user_id parameter",
			"message": "user_id is required",
		})
		return
	}

	sessions, err := h.service.GetSessionsByUserID(userID)
	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "user not found",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get sessions",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.BatchSessionResponse{
		Sessions:   sessions,
		TotalCount: len(sessions),
		Message:    "sessions retrieved successfully",
	})
}

func (h *SessionHandler) ListSessions(c *gin.Context) {
	sessions, err := h.service.ListSessions()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to list sessions",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.BatchSessionResponse{
		Sessions:   sessions,
		TotalCount: len(sessions),
		Message:    "sessions retrieved successfully",
	})
}

func (h *SessionHandler) ValidateSession(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "missing authorization header",
			"message": "authorization token is required",
		})
		return
	}

	session, err := h.service.ValidateSession(token)
	if err != nil {
		if errors.Is(err, store.ErrSessionNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "invalid or frozen session",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to validate session",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.SessionResponse{
		Session: session,
		Message: "session is valid",
	})
}
