// pkg/recommendation/usecase/usecase.go
package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/MHK-26/pod_platfrom_go/pkg/common/config"
	"github.com/MHK-26/pod_platfrom_go/pkg/recommendation/models"
	"github.com/MHK-26/pod_platfrom_go/pkg/recommendation/repository/postgres"
)

// Usecase defines the methods for the recommendation usecase
type Usecase interface {
	// User-based recommendations
	GetPersonalizedRecommendations(ctx context.Context, req *models.RecommendationRequest) (*models.RecommendationResponse, error)
	
	// Similar content recommendations
	GetSimilarPodcasts(ctx context.Context, req *models.SimilarContentRequest) (*models.RecommendationResponse, error)
	GetSimilarEpisodes(ctx context.Context, req *models.SimilarContentRequest) (*models.RecommendationResponse, error)
	
	// Popular content recommendations
	GetTrendingPodcasts(ctx context.Context, req *models.TrendingRequest) (*models.RecommendationResponse, error)
	GetPopularInCategory(ctx context.Context, req *models.CategoryPopularRequest) (*models.RecommendationResponse, error)
	
	// User preferences management
	UpdateUserPreference(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, weight float64) error
	GetUserPreferences(ctx context.Context, userID uuid.UUID) ([]models.UserPreference, error)
}

type usecase struct {
	repo           postgres.Repository
	cfg            *config.Config
	contextTimeout time.Duration
}

// NewUsecase creates a new recommendation usecase
func NewUsecase(repo postgres.Repository, cfg *config.Config, timeout time.Duration) Usecase {
	return &usecase{
		repo:           repo,
		cfg:            cfg,
		contextTimeout: timeout,
	}
}

// GetPersonalizedRecommendations gets personalized recommendations for a user
func (u *usecase) GetPersonalizedRecommendations(ctx context.Context, req *models.RecommendationRequest) (*models.RecommendationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	// Set default limit if not specified
	if req.Limit <= 0 {
		req.Limit = 10
	}
	
	// Cap the limit
	if req.Limit > 50 {
		req.Limit = 50
	}
	
	items, err := u.repo.GetPersonalizedRecommendations(ctx, req.UserID, req.Limit, req.ExcludedIDs)
	if err != nil {
		return nil, err
	}
	
	return &models.RecommendationResponse{Items: items}, nil
}

// GetSimilarPodcasts gets podcasts similar to a specified podcast
func (u *usecase) GetSimilarPodcasts(ctx context.Context, req *models.SimilarContentRequest) (*models.RecommendationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	// Set default limit if not specified
	if req.Limit <= 0 {
		req.Limit = 10
	}
	
	// Cap the limit
	if req.Limit > 50 {
		req.Limit = 50
	}
	
	items, err := u.repo.GetSimilarPodcasts(ctx, req.ContentID, req.Limit, req.ExcludedIDs)
	if err != nil {
		return nil, err
	}
	
	return &models.RecommendationResponse{Items: items}, nil
}

// GetSimilarEpisodes gets episodes similar to a specified episode
func (u *usecase) GetSimilarEpisodes(ctx context.Context, req *models.SimilarContentRequest) (*models.RecommendationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	// Set default limit if not specified
	if req.Limit <= 0 {
		req.Limit = 10
	}
	
	// Cap the limit
	if req.Limit > 50 {
		req.Limit = 50
	}
	
	items, err := u.repo.GetSimilarEpisodes(ctx, req.ContentID, req.Limit, req.ExcludedIDs)
	if err != nil {
		return nil, err
	}
	
	return &models.RecommendationResponse{Items: items}, nil
}

// GetTrendingPodcasts gets trending podcasts
func (u *usecase) GetTrendingPodcasts(ctx context.Context, req *models.TrendingRequest) (*models.RecommendationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	// Set default limit if not specified
	if req.Limit <= 0 {
		req.Limit = 10
	}
	
	// Cap the limit
	if req.Limit > 50 {
		req.Limit = 50
	}
	
	items, err := u.repo.GetTrendingPodcasts(ctx, req.TimeRange, req.Limit, req.ExcludedIDs)
	if err != nil {
		return nil, err
	}
	
	return &models.RecommendationResponse{Items: items}, nil
}

// GetPopularInCategory gets popular content in a category
func (u *usecase) GetPopularInCategory(ctx context.Context, req *models.CategoryPopularRequest) (*models.RecommendationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	// Set default limit if not specified
	if req.Limit <= 0 {
		req.Limit = 10
	}
	
	// Cap the limit
	if req.Limit > 50 {
		req.Limit = 50
	}
	
	items, err := u.repo.GetPopularInCategory(ctx, req.CategoryID, req.Limit, req.ExcludedIDs)
	if err != nil {
		return nil, err
	}
	
	return &models.RecommendationResponse{Items: items}, nil
}

// UpdateUserPreference updates a user's category preference
func (u *usecase) UpdateUserPreference(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, weight float64) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	return u.repo.UpdateUserPreference(ctx, userID, categoryID, weight)
}

// GetUserPreferences gets a user's category preferences
func (u *usecase) GetUserPreferences(ctx context.Context, userID uuid.UUID) ([]models.UserPreference, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	return u.repo.GetUserPreferences(ctx, userID)
}