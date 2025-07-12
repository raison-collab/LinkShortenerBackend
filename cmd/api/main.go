package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerDocs "github.com/raison-collab/LinkShorternetBackend/docs" // импорт swagger документации https://github.com/swaggo/swag
	"github.com/raison-collab/LinkShorternetBackend/internal/delivery/http/router"
	"github.com/raison-collab/LinkShorternetBackend/internal/infrastructure/config"
	"github.com/raison-collab/LinkShorternetBackend/internal/infrastructure/database"
	"github.com/raison-collab/LinkShorternetBackend/pkg/logger"
)

// @title Link Shortener API
// @version 1.0
// @description A modern URL shortening service with analytics

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Update Swagger host information dynamically
	swaggerDocs.SwaggerInfo.Host = cfg.URL.APIHost

	// logger
	log := logger.NewWithConfig(logger.Config{
		Level:    cfg.Log.Level,
		Format:   cfg.Log.Format,
		Output:   logger.LogOutput(cfg.Log.Output),
		FilePath: cfg.Log.FilePath,
	})

	// Connect to PostgreSQL
	db, err := database.NewPostgresDB(cfg.Database.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Connect to Redis
	redisClient, err := database.NewRedisClient(cfg.Redis.GetRedisAddr(), cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize router
	r := router.NewRouter(db, redisClient, cfg, log)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.App.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Infof("Starting server on port %s", cfg.App.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}
