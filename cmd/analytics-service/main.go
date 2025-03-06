// cmd/analytics-service/main.go
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
	"github.com/your-username/podcast-platform/pkg/common/config"
	"github.com/your-username/podcast-platform/pkg/common/database"
	"github.com/your-username/podcast-platform/pkg/common/logger"
	"github.com/your-username/podcast-platform/pkg/common/middleware"
	analyticsRepo "github.com/your-username/podcast-platform/pkg/analytics/repository/postgres"
	analyticsUsecase "github.com/your-username/podcast-platform/pkg/analytics/usecase"
	analyticsHttp "github.com/your-username/podcast-platform/pkg/analytics/delivery/http"
	authUsecase "github.com/your-username/podcast-platform/pkg/auth/usecase"
)

func main() {
	// Initialize logger
	logger.Initialize("analytics-service", "info")
	defer logger.Close()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", logger.Field("error", err))
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Connect to database
	db, err := database.NewPostgresDB(&cfg.DB)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.Field("error", err))
	}
	defer database.CloseDB(db)

	// Initialize repositories
	analyticsRepository := analyticsRepo.NewRepository(db)

	// Initialize usecases
	analyticsUC := analyticsUsecase.NewUsecase(analyticsRepository, cfg, 10*time.Second)
	authUC := authUsecase.NewUsecase(nil, cfg, 10*time.Second) // We only need token verification

	// Initialize router
	router := gin.New()

	// Middlewares
	router.Use(middleware.LoggingMiddleware())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Auth middleware
	authMiddleware := middleware.AuthMiddleware(authUC)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		err := database.PostgresHealthCheck(db)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "error",
				"message": "Database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "analytics-service",
		})
	})

	// Initialize HTTP handlers
	analyticsHandler := analyticsHttp.NewHandler(analyticsUC)

	// Register routes
	v1 := router.Group("/api/v1")
	analyticsHandler.RegisterRoutes(v1, authMiddleware)

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Start the server in a goroutine
	go func() {
		logger.Info("Analytics service listening", logger.Field("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", logger.Field("error", err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Create a deadline for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shut down the server
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", logger.Field("error", err))
	}

	logger.Info("Server exiting")
}