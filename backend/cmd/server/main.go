// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/0xsj/alya.io/backend/internal/api"
	"github.com/0xsj/alya.io/backend/internal/api/handler"
	"github.com/0xsj/alya.io/backend/internal/api/middleware"
	"github.com/0xsj/alya.io/backend/internal/config"
	"github.com/0xsj/alya.io/backend/internal/repository/postgres"
	"github.com/0xsj/alya.io/backend/internal/service"
	"github.com/0xsj/alya.io/backend/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.New(logger.Config{
		Level:        logger.InfoLevel,
		EnableJSON:   false,
		EnableTime:   true,
		EnableCaller: true,
		CallerSkip:   1,
		CallerDepth:  10,
		Writer:       os.Stdout,
	})

	log.Info("Starting Alya.io backend service with transcript support")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}
	
	// Connect to database
	db, err := postgres.NewDB(cfg, log)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	
	// Initialize repositories
	videoRepo := postgres.NewVideoRepository(db, log)
	transcriptRepo := postgres.NewTranscriptRepository(db, log)
	
	// Initialize external services
	youtubeScraper := service.NewYouTubeScraper(log)
	
	// Initialize services
	transcriptService := service.NewTranscriptService(transcriptRepo, youtubeScraper, log)
	videoService := service.NewVideoService(videoRepo, transcriptService, log)
	
	// Initialize middlewares
	authMiddleware := middleware.NewAuthMiddleware(log)
	
	// Initialize handlers
	videoHandler := handler.NewVideoHandler(videoService, log)
	
	// Set up router
	router := api.NewRouter(videoHandler, authMiddleware, log)
	
	// Create a logger middleware for all requests
	loggedRouter := logger.HTTPMiddleware(log)(router)
	
	// Set up HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      loggedRouter,
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout:  2 * cfg.Server.Timeout,
	}
	
	// Start server in a goroutine
	go func() {
		log.Infof("Server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()
	
	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	// Wait for shutdown signal
	<-quit
	log.Info("Shutting down server...")
	
	// Create deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	
	log.Info("Server exited gracefully")
}