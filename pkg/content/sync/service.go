// pkg/content/sync/service.go
package sync

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/your-username/podcast-platform/pkg/content/models"
	"github.com/your-username/podcast-platform/pkg/content/repository/postgres"
	"github.com/your-username/podcast-platform/pkg/content/rss"
)

// Service defines the interface for the RSS sync service
type Service interface {
	// SyncPodcast synchronizes a podcast feed by ID
	SyncPodcast(ctx context.Context, podcastID uuid.UUID) (*models.RSSFeedSyncResult, error)
	
	// SyncAllPodcasts synchronizes all active podcasts
	SyncAllPodcasts(ctx context.Context) ([]models.RSSFeedSyncResult, error)
	
	// GetSyncStatus gets the latest sync status for a podcast
	GetSyncStatus(ctx context.Context, podcastID uuid.UUID) (*models.RSSFeedSyncLog, error)
	
	// ParseFeed parses an RSS feed from a URL
	ParseFeed(ctx context.Context, url string) (*models.RSSFeed, error)
}

type service struct {
	repo       postgres.Repository
	parser     rss.Parser
	db         *sqlx.DB
	syncMutex  *sync.Map // To prevent concurrent syncs for the same podcast
}

// NewService creates a new RSS sync service
func NewService(repo postgres.Repository, parser rss.Parser, db *sqlx.DB) Service {
	return &service{
		repo:      repo,
		parser:    parser,
		db:        db,
		syncMutex: &sync.Map{},
	}
}

// ParseFeed parses an RSS feed from a URL
func (s *service) ParseFeed(ctx context.Context, url string) (*models.RSSFeed, error) {
	return s.parser.ParseFeed(ctx, url)
}

// SyncPodcast synchronizes a podcast feed by ID
func (s *service) SyncPodcast(ctx context.Context, podcastID uuid.UUID) (*models.RSSFeedSyncResult, error) {
	// Check if a sync is already in progress for this podcast
	if _, loaded := s.syncMutex.LoadOrStore(podcastID.String(), true); loaded {
		return nil, fmt.Errorf("sync already in progress for podcast: %s", podcastID)
	}
	defer s.syncMutex.Delete(podcastID.String())

	// Get podcast from database
	podcast, err := s.repo.GetPodcastByID(ctx, podcastID)
	if err != nil {
		return nil, fmt.Errorf("failed to get podcast: %w", err)
	}

	if podcast.RSSUrl == "" {
		return nil, fmt.Errorf("podcast has no RSS URL")
	}

	// Create result object
	result := &models.RSSFeedSyncResult{
		PodcastID: podcastID,
		Success:   false,
	}

	// Parse the feed
	feed, err := s.parser.ParseFeed(ctx, podcast.RSSUrl)
	if err != nil {
		s.logSyncFailure(ctx, podcastID, 0, 0, err.Error())
		result.ErrorMessage = err.Error()
		return result, fmt.Errorf("failed to parse feed: %w", err)
	}

	// Start a transaction
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		s.logSyncFailure(ctx, podcastID, 0, 0, "Failed to start transaction")
		result.ErrorMessage = "Database error"
		return result, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Update podcast metadata if it has changed
	updated := false
	updatedPodcast := *podcast

	if feed.Title != "" && feed.Title != podcast.Title {
		updatedPodcast.Title = feed.Title
		updated = true
	}

	if feed.Description != "" && feed.Description != podcast.Description {
		updatedPodcast.Description = feed.Description
		updated = true
	}

	if feed.Language != "" && feed.Language != podcast.Language {
		updatedPodcast.Language = feed.Language
		updated = true
	}

	if feed.Author != "" && feed.Author != podcast.Author {
		updatedPodcast.Author = feed.Author
		updated = true
	}

	if feed.CoverImageURL != "" && feed.CoverImageURL != podcast.CoverImageURL {
		updatedPodcast.CoverImageURL = feed.CoverImageURL
		updated = true
	}

	if feed.WebsiteURL != "" && feed.WebsiteURL != podcast.WebsiteURL {
		updatedPodcast.WebsiteURL = feed.WebsiteURL
		updated = true
	}

	if feed.Category != "" && feed.Category != podcast.Category {
		updatedPodcast.Category = feed.Category
		updated = true
	}

	if feed.Subcategory != "" && feed.Subcategory != podcast.Subcategory {
		updatedPodcast.Subcategory = feed.Subcategory
		updated = true
	}

	// Update the podcast explicit flag
	if feed.Explicit != podcast.Explicit {
		updatedPodcast.Explicit = feed.Explicit
		updated = true
	}

	// Set the last synced time
	now := time.Now()
	updatedPodcast.LastSyncedAt = &now
	updated = true

	// Update podcast if metadata has changed
	if updated {
		if err := s.repo.UpdatePodcastTx(ctx, tx, &updatedPodcast); err != nil {
			s.logSyncFailure(ctx, podcastID, 0, 0, "Failed to update podcast metadata")
			result.ErrorMessage = "Failed to update podcast metadata"
			return result, fmt.Errorf("failed to update podcast: %w", err)
		}
	}

	// Get existing episodes for this podcast
	existingEpisodes, err := s.repo.GetAllEpisodesByPodcastIDTx(ctx, tx, podcastID)
	if err != nil {
		s.logSyncFailure(ctx, podcastID, 0, 0, "Failed to get existing episodes")
		result.ErrorMessage = "Failed to get existing episodes"
		return result, fmt.Errorf("failed to get existing episodes: %w", err)
	}

	// Create a map of existing episodes for quick lookup
	existingEpisodeMap := make(map[string]*models.Episode)
	for _, episode := range existingEpisodes {
		existingEpisodeMap[episode.GUID] = episode
	}

	// Process episodes in the feed
	episodesAdded := 0
	episodesUpdated := 0

	for _, item := range feed.Items {
		// Skip if GUID is empty
		if item.GUID == "" {
			continue
		}

		// Check if episode already exists
		existingEpisode, exists := existingEpisodeMap[item.GUID]
		if exists {
			// Update episode if needed
			updated := false
			updatedEpisode := *existingEpisode

			if item.Title != "" && item.Title != existingEpisode.Title {
				updatedEpisode.Title = item.Title
				updated = true
			}

			if item.Description != "" && item.Description != existingEpisode.Description {
				updatedEpisode.Description = item.Description
				updated = true
			}

			if item.AudioURL != "" && item.AudioURL != existingEpisode.AudioURL {
				updatedEpisode.AudioURL = item.AudioURL
				updated = true
			}

			if item.Duration > 0 && item.Duration != existingEpisode.Duration {
				updatedEpisode.Duration = item.Duration
				updated = true
			}

			if item.CoverImageURL != "" && item.CoverImageURL != existingEpisode.CoverImageURL {
				updatedEpisode.CoverImageURL = item.CoverImageURL
				updated = true
			}

			if !item.PublicationDate.IsZero() && !item.PublicationDate.Equal(existingEpisode.PublicationDate) {
				updatedEpisode.PublicationDate = item.PublicationDate
				updated = true
			}

			if item.EpisodeNumber != nil && (existingEpisode.EpisodeNumber == nil || *item.EpisodeNumber != *existingEpisode.EpisodeNumber) {
				updatedEpisode.EpisodeNumber = item.EpisodeNumber
				updated = true
			}

			if item.SeasonNumber != nil && (existingEpisode.SeasonNumber == nil || *item.SeasonNumber != *existingEpisode.SeasonNumber) {
				updatedEpisode.SeasonNumber = item.SeasonNumber
				updated = true
			}

			// Update episode if metadata has changed
			if updated {
				updatedEpisode.UpdatedAt = time.Now()
				if err := s.repo.UpdateEpisodeTx(ctx, tx, &updatedEpisode); err != nil {
					log.Printf("Failed to update episode %s: %v", existingEpisode.ID, err)
					continue
				}
				episodesUpdated++
			}
		} else {
			// Create new episode
			newEpisode := &models.Episode{
				ID:              uuid.New(),
				PodcastID:       podcastID,
				Title:           item.Title,
				Description:     item.Description,
				AudioURL:        item.AudioURL,
				Duration:        item.Duration,
				CoverImageURL:   item.CoverImageURL,
				PublicationDate: item.PublicationDate,
				GUID:            item.GUID,
				EpisodeNumber:   item.EpisodeNumber,
				SeasonNumber:    item.SeasonNumber,
				Status:          "active",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			if err := s.repo.CreateEpisodeTx(ctx, tx, newEpisode); err != nil {
				log.Printf("Failed to create episode with GUID %s: %v", item.GUID, err)
				continue
			}
			episodesAdded++
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		s.logSyncFailure(ctx, podcastID, episodesAdded, episodesUpdated, "Failed to commit transaction")
		result.ErrorMessage = "Database error"
		return result, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Log success
	s.logSyncSuccess(ctx, podcastID, episodesAdded, episodesUpdated)

	// Update result
	result.Success = true
	result.EpisodesAdded = episodesAdded
	result.EpisodesUpdated = episodesUpdated

	return result, nil