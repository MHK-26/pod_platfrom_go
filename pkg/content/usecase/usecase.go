// pkg/content/usecase/usecase.go
package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/your-username/podcast-platform/pkg/common/config"
	"github.com/your-username/podcast-platform/pkg/content/models"
	"github.com/your-username/podcast-platform/pkg/content/repository/postgres"
	"github.com/your-username/podcast-platform/pkg/content/rss"
	"github.com/your-username/podcast-platform/pkg/content/sync"
)

// Usecase defines the methods for the content usecase
type Usecase interface {
	// Podcast methods
	CreatePodcast(ctx context.Context, podcasterID uuid.UUID, req *models.CreatePodcastRequest, feed *models.RSSFeed) (*models.Podcast, error)
	GetPodcastByID(ctx context.Context, id uuid.UUID) (*models.PodcastResponse, error)
	GetPodcastsByPodcasterID(ctx context.Context, podcasterID uuid.UUID, page, pageSize int) ([]*models.PodcastResponse, int, error)
	UpdatePodcast(ctx context.Context, id, podcasterID uuid.UUID, req *models.UpdatePodcastRequest) (*models.Podcast, error)
	DeletePodcast(ctx context.Context, id, podcasterID uuid.UUID) error
	ListPodcasts(ctx context.Context, params models.PodcastSearchParams) ([]*models.PodcastResponse, int, error)
	IsUserAuthorizedForPodcast(ctx context.Context, podcastID, userID uuid.UUID) (bool, error)
	
	// RSS feed methods
	ParseRSSFeed(ctx context.Context, url string) (*models.RSSFeed, error)
	SyncPodcastFromRSS(ctx context.Context, podcastID uuid.UUID) (*models.RSSFeedSyncResult, error)
	SyncAllPodcasts(ctx context.Context) ([]models.RSSFeedSyncResult, error)
	GetLatestSyncLog(ctx context.Context, podcastID uuid.UUID) (*models.RSSFeedSyncLog, error)
	GetSyncLogs(ctx context.Context, podcastID uuid.UUID, page, pageSize int) ([]*models.RSSFeedSyncLog, int, error)
	
	// Episode methods
	GetEpisodeByID(ctx context.Context, id uuid.UUID) (*models.EpisodeResponse, error)
	GetEpisodesByPodcastID(ctx context.Context, podcastID uuid.UUID, page, pageSize int) ([]*models.EpisodeResponse, int, error)
	
	// Category methods
	GetCategories(ctx context.Context) ([]*models.Category, error)
	
	// Subscription methods
	SubscribeToPodcast(ctx context.Context, listenerID, podcastID uuid.UUID) error
	UnsubscribeFromPodcast(ctx context.Context, listenerID, podcastID uuid.UUID) error
	GetSubscribedPodcasts(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.PodcastResponse, int, error)
	IsSubscribed(ctx context.Context, listenerID, podcastID uuid.UUID) (bool, error)
	
	// Playback history methods
	SavePlaybackPosition(ctx context.Context, listenerID, episodeID uuid.UUID, position int, completed bool) error
	GetPlaybackPosition(ctx context.Context, listenerID, episodeID uuid.UUID) (int, bool, error)
	GetListeningHistory(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.PlaybackHistory, int, error)
	
	// Like methods
	LikeEpisode(ctx context.Context, listenerID, episodeID uuid.UUID) error
	UnlikeEpisode(ctx context.Context, listenerID, episodeID uuid.UUID) error
	IsEpisodeLiked(ctx context.Context, listenerID, episodeID uuid.UUID) (bool, error)
	GetLikedEpisodes(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.EpisodeResponse, int, error)
}

type usecase struct {
	repo           postgres.Repository
	rssParser      rss.Parser
	syncService    sync.Service
	cfg            *config.Config
	contextTimeout time.Duration
}

// NewUsecase creates a new content usecase
func NewUsecase(repo postgres.Repository, syncService sync.Service, cfg *config.Config, timeout time.Duration) Usecase {
	return &usecase{
		repo:           repo,
		syncService:    syncService,
		cfg:            cfg,
		contextTimeout: timeout,
	}
}

// CreatePodcast creates a new podcast
func (u *usecase) CreatePodcast(ctx context.Context, podcasterID uuid.UUID, req *models.CreatePodcastRequest, feed *models.RSSFeed) (*models.Podcast, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	// Create podcast model using the feed data
	podcast := &models.Podcast{
		PodcasterID:    podcasterID,
		RSSUrl:         req.RSSUrl,
		Status:         "active",
	}
	
	// If feed is provided, use its data
	if feed != nil {
		podcast.Title = feed.Title
		podcast.Description = feed.Description 
		podcast.CoverImageURL = feed.CoverImageURL
		podcast.WebsiteURL = feed.WebsiteURL
		podcast.Language = feed.Language
		podcast.Author = feed.Author
		podcast.Category = feed.Category
		podcast.Subcategory = feed.Subcategory
		podcast.Explicit = feed.Explicit
	} else {
		// Otherwise use data from request
		podcast.Description = req.Description
		podcast.Category = req.Category
		podcast.Subcategory = req.Subcategory
	}
	
	// Create podcast in database
	err := u.repo.CreatePodcast(ctx, podcast)
	if err != nil {
		return nil, err
	}
	
	return podcast, nil
}

// GetPodcastByID gets a podcast by ID
func (u *usecase) GetPodcastByID(ctx context.Context, id uuid.UUID) (*models.PodcastResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	podcast, err := u.repo.GetPodcastByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Get latest episodes
	episodes, _, err := u.repo.GetEpisodesByPodcastID(ctx, id, 1, 5)
	if err != nil {
		return nil, err
	}
	
	// Convert episodes to episode responses
	latestEpisodes := make([]models.EpisodeResponse, 0, len(episodes))
	for _, episode := range episodes {
		latestEpisodes = append(latestEpisodes, models.EpisodeResponse{
			Episode:          *episode,
			PodcastTitle:     podcast.Title,
			PodcastAuthor:    podcast.Author,
			PodcastImageURL:  podcast.CoverImageURL,
		})
	}
	
	// Create podcast response
	podcastResponse := &models.PodcastResponse{
		Podcast:        *podcast,
		EpisodeCount:   podcast.EpisodeCount,
		LatestEpisodes: latestEpisodes,
	}
	
	return podcastResponse, nil
}

// GetPodcastsByPodcasterID gets podcasts by podcaster ID
func (u *usecase) GetPodcastsByPodcasterID(ctx context.Context, podcasterID uuid.UUID, page, pageSize int) ([]*models.PodcastResponse, int, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	podcasts, totalCount, err := u.repo.GetPodcastsByPodcasterID(ctx, podcasterID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	
	// Convert podcasts to podcast responses
	podcastResponses := make([]*models.PodcastResponse, 0, len(podcasts))
	for _, podcast := range podcasts {
		podcastResponse := &models.PodcastResponse{
			Podcast:      *podcast,
			EpisodeCount: podcast.EpisodeCount,
		}
		podcastResponses = append(podcastResponses, podcastResponse)
	}
	
	return podcastResponses, totalCount, nil
}

// UpdatePodcast updates a podcast
func (u *usecase) UpdatePodcast(ctx context.Context, id, podcasterID uuid.UUID, req *models.UpdatePodcastRequest) (*models.Podcast, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	// Get podcast
	podcast, err := u.repo.GetPodcastByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Check if user is authorized to update podcast
	if podcast.PodcasterID != podcasterID {
		return nil, errors.New("not authorized")
	}
	
	// Update fields
	if req.Description != "" {
		podcast.Description = req.Description
	}
	if req.RSSUrl != "" {
		podcast.RSSUrl = req.RSSUrl
	}
	if req.Category != "" {
		podcast.Category = req.Category
	}
	if req.Subcategory != "" {
		podcast.Subcategory = req.Subcategory
	}
	podcast.UpdatedAt = time.Now()
	
	// Update podcast in database
	err = u.repo.UpdatePodcast(ctx, podcast)
	if err != nil {
		return nil, err
	}
	
	return podcast, nil
}

// DeletePodcast deletes a podcast
func (u *usecase) DeletePodcast(ctx context.Context, id, podcasterID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	// Get podcast
	podcast, err := u.repo.GetPodcastByID(ctx, id)
	if err != nil {
		return err
	}
	
	// Check if user is authorized to delete podcast
	if podcast.PodcasterID != podcasterID {
		return errors.New("not authorized")
	}
	
	// Delete podcast from database
	return u.repo.DeletePodcast(ctx, id)
}

// ListPodcasts lists podcasts with optional filtering
func (u *usecase) ListPodcasts(ctx context.Context, params models.PodcastSearchParams) ([]*models.PodcastResponse, int, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	podcasts, totalCount, err := u.repo.ListPodcasts(ctx, params)
	if err != nil {
		return nil, 0, err
	}
	
	// Convert podcasts to podcast responses
	podcastResponses := make([]*models.PodcastResponse, 0, len(podcasts))
	for _, podcast := range podcasts {
		podcastResponse := &models.PodcastResponse{
			Podcast:      *podcast,
			EpisodeCount: podcast.EpisodeCount,
		}
		podcastResponses = append(podcastResponses, podcastResponse)
	}
	
	return podcastResponses, totalCount, nil
}

// IsUserAuthorizedForPodcast checks if a user is authorized for a podcast
func (u *usecase) IsUserAuthorizedForPodcast(ctx context.Context, podcastID, userID uuid.UUID) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	return u.repo.IsUserAuthorizedForPodcast(ctx, podcastID, userID)
}

// ParseRSSFeed parses an RSS feed from a URL
func (u *usecase) ParseRSSFeed(ctx context.Context, url string) (*models.RSSFeed, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	// Check if a podcast with this RSS URL already exists
	existingPodcast, err := u.repo.GetPodcastByRSSURL(ctx, url)
	if err == nil && existingPodcast != nil {
		return nil, errors.New("a podcast with this RSS feed already exists")
	}
	
	// Parse the feed using the RSS parser from the sync service
	return u.syncService.(sync.Service).ParseFeed(ctx, url)
}

// SyncPodcastFromRSS syncs a podcast from its RSS feed
func (u *usecase) SyncPodcastFromRSS(ctx context.Context, podcastID uuid.UUID) (*models.RSSFeedSyncResult, error) {
	return u.syncService.SyncPodcast(ctx, podcastID)
}

// SyncAllPodcasts syncs all podcasts from their RSS feeds
func (u *usecase) SyncAllPodcasts(ctx context.Context) ([]models.RSSFeedSyncResult, error) {
	return u.syncService.SyncAllPodcasts(ctx)
}

// GetLatestSyncLog gets the latest sync log for a podcast
func (u *usecase) GetLatestSyncLog(ctx context.Context, podcastID uuid.UUID) (*models.RSSFeedSyncLog, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	return u.repo.GetLatestSyncLog(ctx, podcastID)
}

// GetSyncLogs gets the sync logs for a podcast
func (u *usecase) GetSyncLogs(ctx context.Context, podcastID uuid.UUID, page, pageSize int) ([]*models.RSSFeedSyncLog, int, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	return u.repo.GetSyncLogs(ctx, podcastID, page, pageSize)
}

// GetEpisodeByID gets an episode by ID
func (u *usecase) GetEpisodeByID(ctx context.Context, id uuid.UUID) (*models.EpisodeResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	episode, err := u.repo.GetEpisodeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Get podcast details
	podcast, err := u.repo.GetPodcastByID(ctx, episode.PodcastID)
	if err != nil {
		return nil, err
	}
	
	// Create episode response
	episodeResponse := &models.EpisodeResponse{
		Episode:         *episode,
		PodcastTitle:    podcast.Title,
		PodcastAuthor:   podcast.Author,
		PodcastImageURL: podcast.CoverImageURL,
	}
	
	return episodeResponse, nil
}

// GetEpisodesByPodcastID gets episodes by podcast ID
func (u *usecase) GetEpisodesByPodcastID(ctx context.Context, podcastID uuid.UUID, page, pageSize int) ([]*models.EpisodeResponse, int, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	episodes, totalCount, err := u.repo.GetEpisodesByPodcastID(ctx, podcastID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	
	// Get podcast details
	podcast, err := u.repo.GetPodcastByID(ctx, podcastID)
	if err != nil {
		return nil, 0, err
	}
	
	// Convert episodes to episode responses
	episodeResponses := make([]*models.EpisodeResponse, 0, len(episodes))
	for _, episode := range episodes {
		episodeResponse := &models.EpisodeResponse{
			Episode:         *episode,
			PodcastTitle:    podcast.Title,
			PodcastAuthor:   podcast.Author,
			PodcastImageURL: podcast.CoverImageURL,
		}
		episodeResponses = append(episodeResponses, episodeResponse)
	}
	
	return episodeResponses, totalCount, nil
}

// GetCategories gets all categories
func (u *usecase) GetCategories(ctx context.Context) ([]*models.Category, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	return u.repo.GetCategories(ctx)
}

// SubscribeToPodcast subscribes a listener to a podcast
func (u *usecase) SubscribeToPodcast(ctx context.Context, listenerID, podcastID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	// Check if podcast exists
	_, err := u.repo.GetPodcastByID(ctx, podcastID)
	if err != nil {
		return errors.New("podcast not found")
	}
	
	// Subscribe to podcast
	return u.repo.SubscribeToPodcast(ctx, listenerID, podcastID)
}

// UnsubscribeFromPodcast unsubscribes a listener from a podcast
func (u *usecase) UnsubscribeFromPodcast(ctx context.Context, listenerID, podcastID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	return u.repo.UnsubscribeFromPodcast(ctx, listenerID, podcastID)
}

// GetSubscribedPodcasts gets podcasts subscribed by a listener
func (u *usecase) GetSubscribedPodcasts(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.PodcastResponse, int, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	podcasts, totalCount, err := u.repo.GetSubscribedPodcasts(ctx, listenerID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	
	// Convert podcasts to podcast responses
	podcastResponses := make([]*models.PodcastResponse, 0, len(podcasts))
	for _, podcast := range podcasts {
		podcastResponse := &models.PodcastResponse{
			Podcast:      *podcast,
			EpisodeCount: podcast.EpisodeCount,
		}
		podcastResponses = append(podcastResponses, podcastResponse)
	}
	
	return podcastResponses, totalCount, nil
}

// IsSubscribed checks if a listener is subscribed to a podcast
func (u *usecase) IsSubscribed(ctx context.Context, listenerID, podcastID uuid.UUID) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	return u.repo.IsSubscribed(ctx, listenerID, podcastID)
}

// SavePlaybackPosition saves the playback position for an episode
func (u *usecase) SavePlaybackPosition(ctx context.Context, listenerID, episodeID uuid.UUID, position int, completed bool) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	// Check if episode exists
	_, err := u.repo.GetEpisodeByID(ctx, episodeID)
	if err != nil {
		return errors.New("episode not found")
	}
	
	return u.repo.SavePlaybackPosition(ctx, listenerID, episodeID, position, completed)
}

// GetPlaybackPosition gets the playback position for an episode
func (u *usecase) GetPlaybackPosition(ctx context.Context, listenerID, episodeID uuid.UUID) (int, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	return u.repo.GetPlaybackPosition(ctx, listenerID, episodeID)
}

// GetListeningHistory gets the listening history for a listener
func (u *usecase) GetListeningHistory(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.PlaybackHistory, int, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	return u.repo.GetListeningHistory(ctx, listenerID, page, pageSize)
}

// LikeEpisode likes an episode
func (u *usecase) LikeEpisode(ctx context.Context, listenerID, episodeID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	// Check if episode exists
	_, err := u.repo.GetEpisodeByID(ctx, episodeID)
	if err != nil {
		return errors.New("episode not found")
	}
	
	return u.repo.LikeEpisode(ctx, listenerID, episodeID)
}

// UnlikeEpisode unlikes an episode
func (u *usecase) UnlikeEpisode(ctx context.Context, listenerID, episodeID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	return u.repo.UnlikeEpisode(ctx, listenerID, episodeID)
}

// IsEpisodeLiked checks if an episode is liked by a listener
func (u *usecase) IsEpisodeLiked(ctx context.Context, listenerID, episodeID uuid.UUID) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	return u.repo.IsEpisodeLiked(ctx, listenerID, episodeID)
}

// GetLikedEpisodes gets episodes liked by a listener
func (u *usecase) GetLikedEpisodes(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.EpisodeResponse, int, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()
	
	episodes, totalCount, err := u.repo.GetLikedEpisodes(ctx, listenerID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	
	// Convert episodes to episode responses
	episodeResponses := make([]*models.EpisodeResponse, 0, len(episodes))
	for _, episode := range episodes {
		// Get podcast details
		podcast, err := u.repo.GetPodcastByID(ctx, episode.PodcastID)
		if err != nil {
			continue // Skip if podcast not found
		}
		
		episodeResponse := &models.EpisodeResponse{
			Episode:         *episode,
			PodcastTitle:    podcast.Title,
			PodcastAuthor:   podcast.Author,
			PodcastImageURL: podcast.CoverImageURL,
		}
		episodeResponses = append(episodeResponses, episodeResponse)
	}
	
	return episodeResponses, totalCount, nil
}