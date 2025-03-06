// pkg/analytics/usecase/usecase.go
package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/your-username/podcast-platform/pkg/analytics/models"
	"github.com/your-username/podcast-platform/pkg/analytics/repository/postgres"
	"github.com/your-username/podcast-platform/pkg/common/config"
)

// Usecase defines the methods for the analytics usecase
type Usecase interface {
	TrackListen(ctx context.Context, req *models.TrackListenRequest) (*models.ListenEvent, error)
	GetEpisodeAnalytics(ctx context.Context, episodeID uuid.UUID, params models.AnalyticsParams) (*models.EpisodeAnalytics, error)
	GetPodcastAnalytics(ctx context.Context, podcastID uuid.UUID, params models.AnalyticsParams) (*models.PodcastAnalytics, error)
	GetPodcasterAnalytics(ctx context.Context, podcasterID uuid.UUID, params models.AnalyticsParams) (*models.PodcasterAnalytics, error)
	GetListeningHistory(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.ListeningHistoryItem, int, error)
}

type usecase struct {
	repo           postgres.Repository
	cfg            *config.Config
	contextTimeout time.Duration
}

// NewUsecase creates a new analytics usecase
func NewUsecase(repo postgres.Repository, cfg *config.Config, timeout time.Duration) Usecase {
	return &usecase{
		repo:           repo,
		cfg:            cfg,
		contextTimeout: timeout,
	}
}

// TrackListen tracks a listen event
func (u *usecase) TrackListen(ctx context.Context, req *models.TrackListenRequest) (*models.ListenEvent, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	event := &models.ListenEvent{
		ListenerID:  req.ListenerID,
		EpisodeID:   req.EpisodeID,
		Source:      req.Source,
		Duration:    req.Duration,
		Completed:   req.Completed,
		IPAddress:   req.IPAddress,
		UserAgent:   req.UserAgent,
		CountryCode: req.CountryCode,
		City:        req.City,
		StartedAt:   time.Now(),
	}

	err := u.repo.TrackListen(ctx, event)
	if err != nil {
		return nil, err
	}

	return event, nil
}

// GetEpisodeAnalytics gets analytics for an episode
func (u *usecase) GetEpisodeAnalytics(ctx context.Context, episodeID uuid.UUID, params models.AnalyticsParams) (*models.EpisodeAnalytics, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Get listen stats and timeseries
	stats, timePoints, err := u.repo.GetEpisodeListens(ctx, episodeID, params)
	if err != nil {
		return nil, err
	}

	// TODO: Get episode details from content service
	// For now, we'll create a placeholder
	analytics := &models.EpisodeAnalytics{
		EpisodeID:      episodeID,
		Title:          "Episode Title", // Should be fetched from content service
		ListenStats:    *stats,
		ListensByDay:   timePoints,
	}

	return analytics, nil
}

// GetPodcastAnalytics gets analytics for a podcast
func (u *usecase) GetPodcastAnalytics(ctx context.Context, podcastID uuid.UUID, params models.AnalyticsParams) (*models.PodcastAnalytics, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Get listen stats, timeseries, and episode stats
	stats, timePoints, episodeStats, err := u.repo.GetPodcastListens(ctx, podcastID, params)
	if err != nil {
		return nil, err
	}

	// TODO: Get podcast details from content service
	// For now, we'll create a placeholder
	analytics := &models.PodcastAnalytics{
		PodcastID:       podcastID,
		Title:           "Podcast Title", // Should be fetched from content service
		ListenStats:     *stats,
		ListensByDay:    timePoints,
		ListensByEpisode: episodeStats,
	}

	return analytics, nil
}

// GetPodcasterAnalytics gets analytics for a podcaster
func (u *usecase) GetPodcasterAnalytics(ctx context.Context, podcasterID uuid.UUID, params models.AnalyticsParams) (*models.PodcasterAnalytics, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Get podcaster analytics
	analytics, err := u.repo.GetPodcasterListens(ctx, podcasterID, params)
	if err != nil {
		return nil, err
	}

	return analytics, nil
}

// GetListeningHistory gets the listening history for a user
func (u *usecase) GetListeningHistory(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.ListeningHistoryItem, int, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	return u.repo.GetListeningHistory(ctx, listenerID, page, pageSize)
}