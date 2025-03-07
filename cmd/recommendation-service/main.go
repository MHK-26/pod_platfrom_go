// cmd/recommendation-service/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"net"
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
	recommendationRepo "github.com/your-username/podcast-platform/pkg/recommendation/repository/postgres"
	recommendationUsecase "github.com/your-username/podcast-platform/pkg/recommendation/usecase"
	recommendationHttp "github.com/your-username/podcast-platform/pkg/recommendation/delivery/http"
	recommendationGrpc "github.com/your-username/podcast-platform/pkg/recommendation/delivery/grpc"
	pb "github.com/your-username/podcast-platform/api/proto/recommendation"
	"google.golang.org/grpc"
)

func main() {
	// Initialize logger
	logger.Initialize("recommendation-service", "info")
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
	recommendationRepository := recommendationRepo.NewRepository(db)

	// Initialize usecases
	recommendationUC := recommendationUsecase.NewUsecase(recommendationRepository, cfg, 10*time.Second)
	authUC := authUsecase.NewUsecase(nil, cfg, 10*time.Second) // We only need token verification

	// Setup HTTP server
	router := gin.New()
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
			"service": "recommendation-service",
		})
	})

	// Initialize HTTP handlers
	recommendationHandler := recommendationHttp.NewHandler(recommendationUC)

	// Register HTTP routes
	v1 := router.Group("/api/v1")
	recommendationHandler.RegisterRoutes(v1, authMiddleware)

	// Start HTTP server
	httpSrv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Start the HTTP server in a goroutine
	go func() {
		logger.Info("Recommendation HTTP service listening", logger.Field("port", cfg.Server.Port))
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", logger.Field("error", err))
		}
	}()

	// Setup gRPC server
	grpcPort := cfg.Server.Port + "1" // Use port+1 for gRPC
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		logger.Fatal("Failed to listen for gRPC", logger.Field("error", err))
	}

	grpcServer := grpc.NewServer()
	grpcHandler := recommendationGrpc.NewHandler(recommendationUC)
	pb.RegisterRecommendationServiceServer(grpcServer, grpcHandler)

	// Start the gRPC server in a goroutine
	go func() {
		logger.Info("Recommendation gRPC service listening", logger.Field("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("Failed to start gRPC server", logger.Field("error", err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down servers...")

	// Create a deadline for the shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shut down the HTTP server
	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Fatal("HTTP Server forced to shutdown", logger.Field("error", err))
	}

	// Shut down the gRPC server
	grpcServer.GracefulStop()

	logger.Info("Servers exited")
}