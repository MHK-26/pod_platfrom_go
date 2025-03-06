// pkg/content/repository/postgres/repository.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/your-username/podcast-platform/pkg/content/models"
)

// Repository defines the methods for the content repository
type Repository interface {
	// Podcast methods
	CreatePodcast(ctx context.Context, podcast *models.Podcast) error
	GetPodcastByID(ctx context.Context, id uuid.UUID) (*models.Podcast, error)
	GetPodcastsByPodcasterID(ctx context.Context, podcasterID uuid.UUID, page, pageSize int) ([]*models.Podcast, int, error)
	UpdatePodcast(ctx context.Context, podcast *models.Podcast) error
	DeletePodcast(ctx context.Context, id uuid.UUID) error
	ListPodcasts(ctx context.Context, params models.PodcastSearchParams) ([]*models.Podcast, int, error)
	
	// Episode methods
	CreateEpisode(ctx context.Context, episode *models.Episode) error
	GetEpisodeByID(ctx context.Context, id uuid.UUID) (*models.Episode, error)
	GetEpisodesByPodcastID(ctx context.Context, podcastID uuid.UUID, page, pageSize int) ([]*models.Episode, int, error)
	UpdateEpisode(ctx context.Context, episode *models.Episode) error
	DeleteEpisode(ctx context.Context, id uuid.UUID) error
	ListEpisodes(ctx context.Context, params models.EpisodeSearchParams) ([]*models.Episode, int, error)
	
	// Category methods
	GetCategories(ctx context.Context) ([]*models.Category, error)
	AssociatePodcastWithCategories(ctx context.Context, podcastID uuid.UUID, categoryIDs []uuid.UUID) error
	GetCategoriesByPodcastID(ctx context.Context, podcastID uuid.UUID) ([]*models.Category, error)
	
	// Subscription methods
	SubscribeToPodcast(ctx context.Context, listenerID, podcastID uuid.UUID) error
	UnsubscribeFromPodcast(ctx context.Context, listenerID, podcastID uuid.UUID) error
	GetSubscribedPodcasts(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.Podcast, int, error)
	IsSubscribed(ctx context.Context, listenerID, podcastID uuid.UUID) (bool, error)
	
	// Playback history methods
	SavePlaybackPosition(ctx context.Context, listenerID, episodeID uuid.UUID, position int, completed bool) error
	GetPlaybackPosition(ctx context.Context, listenerID, episodeID uuid.UUID) (int, bool, error)
	GetListeningHistory(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.PlaybackHistory, int, error)
	
	// Like methods
	LikeEpisode(ctx context.Context, listenerID, episodeID uuid.UUID) error
	UnlikeEpisode(ctx context.Context, listenerID, episodeID uuid.UUID) error
	IsEpisodeLiked(ctx context.Context, listenerID, episodeID uuid.UUID) (bool, error)
	GetLikedEpisodes(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.Episode, int, error)
	
	// Comments methods
	AddComment(ctx context.Context, comment *models.Comment) error
	GetCommentsByEpisodeID(ctx context.Context, episodeID uuid.UUID, page, pageSize int) ([]*models.Comment, int, error)
	DeleteComment(ctx context.Context, commentID, userID uuid.UUID) error
	
	// Playlist methods
	CreatePlaylist(ctx context.Context, playlist *models.Playlist) error
	GetPlaylistByID(ctx context.Context, id, userID uuid.UUID) (*models.Playlist, error)
	GetUserPlaylists(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*models.Playlist, int, error)
	UpdatePlaylist(ctx context.Context, playlist *models.Playlist) error
	DeletePlaylist(ctx context.Context, id, userID uuid.UUID) error
	AddToPlaylist(ctx context.Context, playlistID, episodeID uuid.UUID, position int) error
	RemoveFromPlaylist(ctx context.Context, playlistID, episodeID uuid.UUID) error
	GetPlaylistItems(ctx context.Context, playlistID uuid.UUID, page, pageSize int) ([]*models.PlaylistItem, int, error)
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new content repository
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

// CreatePodcast creates a new podcast
func (r *repository) CreatePodcast(ctx context.Context, podcast *models.Podcast) error {
	query := `
		INSERT INTO podcasts (
			id, podcaster_id, title, description, cover_image_url, rss_url, website_url,
			language, author, category, subcategory, explicit, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id
	`

	if podcast.ID == uuid.Nil {
		podcast.ID = uuid.New()
	}

	now := time.Now()
	podcast.CreatedAt = now
	podcast.UpdatedAt = now

	err := r.db.QueryRowContext(
		ctx,
		query,
		podcast.ID,
		podcast.PodcasterID,
		podcast.Title,
		podcast.Description,
		podcast.CoverImageURL,
		podcast.RSSUrl,
		podcast.WebsiteURL,
		podcast.Language,
		podcast.Author,
		podcast.Category,
		podcast.Subcategory,
		podcast.Explicit,
		podcast.Status,
		podcast.CreatedAt,
		podcast.UpdatedAt,
	).Scan(&podcast.ID)

	return err
}

// GetPodcastByID gets a podcast by ID
func (r *repository) GetPodcastByID(ctx context.Context, id uuid.UUID) (*models.Podcast, error) {
	var podcast models.Podcast
	query := `
		SELECT
			id, podcaster_id, title, description, cover_image_url, rss_url, website_url,
			language, author, category, subcategory, explicit, status, created_at, updated_at,
			last_synced_at
		FROM podcasts
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &podcast, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("podcast not found")
		}
		return nil, err
	}

	// Get episode count
	countQuery := `SELECT COUNT(*) FROM episodes WHERE podcast_id = $1 AND status = 'active'`
	var count int
	err = r.db.GetContext(ctx, &count, countQuery, id)
	if err != nil {
		return nil, err
	}
	
	podcast.EpisodeCount = count

	// Get categories
	categories, err := r.GetCategoriesByPodcastID(ctx, id)
	if err != nil {
		return nil, err
	}
	podcast.Categories = categories

	return &podcast, nil
}

// Implementation for the rest of the methods would go here...
// For brevity, I've included just a couple of methods as examples.

// GetPodcastsByPodcasterID gets podcasts by podcaster ID
func (r *repository) GetPodcastsByPodcasterID(ctx context.Context, podcasterID uuid.UUID, page, pageSize int) ([]*models.Podcast, int, error) {
	query := `
		SELECT
			id, podcaster_id, title, description, cover_image_url, rss_url, website_url,
			language, author, category, subcategory, explicit, status, created_at, updated_at,
			last_synced_at
		FROM podcasts
		WHERE podcaster_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var podcasts []*models.Podcast
	offset := (page - 1) * pageSize
	err := r.db.SelectContext(ctx, &podcasts, query, podcasterID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM podcasts WHERE podcaster_id = $1`
	var count int
	err = r.db.GetContext(ctx, &count, countQuery, podcasterID)
	if err != nil {
		return nil, 0, err
	}

	// Get categories for each podcast
	for _, podcast := range podcasts {
		categories, err := r.GetCategoriesByPodcastID(ctx, podcast.ID)
		if err != nil {
			return nil, 0, err
		}
		podcast.Categories = categories
		
		// Get episode count
		episodeCountQuery := `SELECT COUNT(*) FROM episodes WHERE podcast_id = $1 AND status = 'active'`
		var episodeCount int
		err = r.db.GetContext(ctx, &episodeCount, episodeCountQuery, podcast.ID)
		if err != nil {
			return nil, 0, err
		}
		podcast.EpisodeCount = episodeCount
	}

	return podcasts, count, nil
}

// GetCategoriesByPodcastID gets categories by podcast ID
func (r *repository) GetCategoriesByPodcastID(ctx context.Context, podcastID uuid.UUID) ([]*models.Category, error) {
	query := `
		SELECT c.id, c.name, c.description, c.icon_url
		FROM categories c
		JOIN podcast_categories pc ON c.id = pc.category_id
		WHERE pc.podcast_id = $1
	`

	var categories []*models.Category
	err := r.db.SelectContext(ctx, &categories, query, podcastID)
	if err != nil {
		return nil, err
	}

	return categories, nil
}