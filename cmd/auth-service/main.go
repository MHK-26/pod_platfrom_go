// cmd/auth-service/main.go
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
	"github.com/your-username/podcast-platform/pkg/auth/delivery/http/handlers"
	"github.com/your-username/podcast-platform/pkg/auth/repository/postgres"
	"github.com/your-username/podcast-platform/pkg/auth/usecase"
	"github.com/your-username/podcast-platform/pkg/common/config"
	"github.com/your-username/podcast-platform/pkg/common/database"
	"github.com/your-username/podcast-platform/pkg/common/middleware"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	// Connect to database
	db, err := database.NewPostgresDB(&cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDB(db)

	// Initialize repository
	repo := postgres.NewRepository(db)

	// Initialize usecase
	usecase := usecase.NewUsecase(repo, cfg, 10*time.Second)

	// Initialize router
	router := gin.Default()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	// Auth middleware
	authMiddleware := middleware.AuthMiddleware(usecase)

	// Initialize handlers
	handler := handlers.NewHandler(usecase)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		err := database.PostgresHealthCheck(db)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "error",
				"message": "Database connection failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "auth-service",
		})
	})

	// Register routes
	v1 := router.Group("/api/v1")
	handler.RegisterRoutes(v1, authMiddleware)

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
		fmt.Printf("Auth service listening on port %s\n", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a deadline for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shut down the server
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}