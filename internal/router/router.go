package router

import (
	"session-management/internal/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(sessionHandler *handler.SessionHandler) *gin.Engine {
	r := gin.Default()

	api := r.Group("/api/v1")
	{
		sessions := api.Group("/sessions")
		{
			sessions.POST("", sessionHandler.CreateSession)
			sessions.GET("", sessionHandler.ListSessions)
			sessions.GET("/user/:user_id", sessionHandler.GetUserSessions)
			sessions.POST("/validate", sessionHandler.ValidateSession)

			admin := sessions.Group("/admin")
			{
				admin.POST("/freeze", sessionHandler.FreezeSession)
				admin.POST("/unfreeze", sessionHandler.UnfreezeSession)
				admin.POST("/cache/refresh", sessionHandler.RefreshCache)
			}
		}
	}

	return r
}
