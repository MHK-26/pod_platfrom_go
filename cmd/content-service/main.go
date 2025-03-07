// cmd/content-service/main.go
package main

import (
	"context"
	"flag"
	"fmt"
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
	
	authUsecase "github.com/your-username/podcast-platform/pkg/auth/usecase"
	contentRepo "github.com/your-username/podcast-platform/pkg/content/repository/postgres"
	contentUsecase "github.com/your-username/podcast-platform/pkg/content/usecase"
	contentHttp "github.com/your-username/podcast-platform/pkg/content/delivery/http"
	contentRSS "github.com/your-username/podcast-platform/pkg/content/rss"
	contentSync "github.com/your-username/podcast-platform/pkg/content/sync"
)

func main() {
	// Define command line flags
	syncRSS := flag.Bool("sync-rss", false, "Only perform RSS feed synchronization and exit")
	flag.Parse()

	// Initialize logger
	logger.Initialize("content-service", "info")
	defer logger.Close()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config", logger.Field("error", err))
	}

	// Connect to database
	db, err := database.NewPostgresDB(&cfg.DB)
	if err != nil {
		logger.Fatal("Failed to connect to database", logger.Field("error", err))
	}
	defer database.CloseDB(db)

	// Initialize repositories
	contentRepository := contentRepo.NewRepository(db)

	// Initialize RSS parser
	rssParser := contentRSS.NewParser(30 * time.Second)

	// Initialize sync service
	syncService := contentSync.NewService(contentRepository, rssParser, db)

	// Initialize usecases
	contentUC := contentUsecase.NewUsecase(contentRepository, syncService, cfg, 10*time.Second)
	authUC := authUsecase.NewUsecase(nil, cfg, 10*time.Second) // We only need token verification

	// If sync-rss flag is set, perform sync and exit
	if *syncRSS {
		logger.Info("Starting RSS feed synchronization")
		
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
		defer cancel()
		
		results, err := contentUC.SyncAllPodcasts(ctx)
		if err != nil {
			logger.Fatal("Failed to sync podcasts", logger.Field("error", err))
		}
		
		// Log results
		successCount := 0
		for _, result := range results {
			if result.Success {
				successCount++
				logger.Info("Successfully synced podcast", 
					logger.Field("podcast_id", result.PodcastID),
					logger.Field("episodes_added", result.EpisodesAdded),
					logger.Field("episodes_updated", result.EpisodesUpdated))
			} else {
				logger.Error("Failed to sync podcast", 
					logger.Field("podcast_id", result.PodcastID),
					logger.Field("error", result.ErrorMessage))
			}
		}
		
		logger.Info("RSS feed synchronization completed", 
			logger.Field("total", len(results)),
			logger.Field("success", successCount),
			logger.Field("failed", len(results) - successCount))
		
		return
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

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
			"service": "content-service",
		})
	})

	// Initialize HTTP handlers
	contentHandler := contentHttp.NewHandler(contentUC)

	// Register routes
	v1 := router.Group("/api/v1")
	contentHandler.RegisterRoutes(v1, authMiddleware)

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
		logger.Info("Content service listening", logger.Field("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", logger.Field("error", err))
		}
	}()
	
	// Start a background goroutine to sync RSS feeds periodically
	go func() {
		// Wait for initial delay before starting
		time.Sleep(1 * time.Minute)
		
		// Create a ticker to run every 6 hours
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		
		// Run sync once at startup
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
		logger.Info("Running initial RSS feed sync")
		_, err := contentUC.SyncAllPodcasts(ctx)
		if err != nil {
			logger.Error("Failed to sync podcasts", logger.Field("error", err))
		}
		cancel()
		
		// Run sync at regular intervals
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
			logger.Info("Running scheduled RSS feed sync")
			_, err := contentUC.SyncAllPodcasts(ctx)
			if err != nil {
				logger.Error("Failed to sync podcasts", logger.Field("error", err))
			}
			cancel()
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