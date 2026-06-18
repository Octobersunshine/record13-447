package main

import (
	"log"
	"os"
	"path/filepath"
	"session-management/internal/handler"
	"session-management/internal/router"
	"session-management/internal/service"
	"session-management/internal/store"
	"time"
)

func main() {
	dataDir := getEnv("DATA_DIR", "./data")
	sessionFile := filepath.Join(dataDir, "sessions.json")
	cacheTTL := 5 * time.Minute

	log.Printf("Initializing persistent storage: %s", sessionFile)
	persistentStore, err := store.NewFileSessionStore(sessionFile)
	if err != nil {
		log.Fatalf("Failed to create persistent store: %v", err)
	}

	log.Println("Initializing cache layer...")
	cachedStore := store.NewCachedSessionStore(persistentStore, cacheTTL)

	if err := cachedStore.RefreshCache(); err != nil {
		log.Printf("Warning: Failed to refresh cache from persistent store: %v", err)
	}

	sessionService := service.NewSessionService(cachedStore)
	sessionHandler := handler.NewSessionHandler(sessionService)

	r := router.SetupRouter(sessionHandler)

	log.Println("Server starting on port 8080...")
	log.Println("Storage mode: Persistent (JSON file) + In-memory cache")
	log.Println("Write strategy: Write-through (persistent first, then cache)")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
