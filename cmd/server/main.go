package main

import (
	"log"
	"session-management/internal/handler"
	"session-management/internal/router"
	"session-management/internal/service"
	"session-management/internal/store"
)

func main() {
	sessionStore := store.NewInMemorySessionStore()
	sessionService := service.NewSessionService(sessionStore)
	sessionHandler := handler.NewSessionHandler(sessionService)

	r := router.SetupRouter(sessionHandler)

	log.Println("Server starting on port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
