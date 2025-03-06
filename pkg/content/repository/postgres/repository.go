// pkg/content/repository/postgres/repository.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
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

// UpdatePodcast updates a podcast
func (r *repository) UpdatePodcast(ctx context.Context, podcast *models.Podcast) error {
	query := `
		UPDATE podcasts SET
			title = $1,
			description = $2,
			cover_image_url = $3,
			website_url = $4,
			language = $5,
			author = $6,
			category = $7,
			subcategory = $8,
			explicit = $9,
			status = $10,
			updated_at = $11
		WHERE id = $12
	`
	
	podcast.UpdatedAt = time.Now()
	
	_, err := r.db.ExecContext(
		ctx,
		query,
		podcast.Title,
		podcast.Description,
		podcast.CoverImageURL,
		podcast.WebsiteURL,
		podcast.Language,
		podcast.Author,
		podcast.Category,
		podcast.Subcategory,
		podcast.Explicit,
		podcast.Status,
		podcast.UpdatedAt,
		podcast.ID,
	)
	
	return err
}

// DeletePodcast deletes a podcast
func (r *repository) DeletePodcast(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM podcasts WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListPodcasts lists podcasts with optional filtering
func (r *repository) ListPodcasts(ctx context.Context, params models.PodcastSearchParams) ([]*models.Podcast, int, error) {
	// Base query
	baseQuery := `
		SELECT
			id, podcaster_id, title, description, cover_image_url, rss_url, website_url,
			language, author, category, subcategory, explicit, status, created_at, updated_at,
			last_synced_at
		FROM podcasts
		WHERE status = 'active'
	`
	
	// Count query
	countQuery := `
		SELECT COUNT(*)
		FROM podcasts
		WHERE status = 'active'
	`
	
	// Add filters
	var filters []string
	var args []interface{}
	argIndex := 1
	
	if params.Query != "" {
		filters = append(filters, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+params.Query+"%")
		argIndex++
	}
	
	if params.Category != "" {
		filters = append(filters, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, params.Category)
		argIndex++
	}
	
	if params.Language != "" {
		filters = append(filters, fmt.Sprintf("language = $%d", argIndex))
		args = append(args, params.Language)
		argIndex++
	}
	
	// Apply filters to queries
	if len(filters) > 0 {
		filterStr := strings.Join(filters, " AND ")
		baseQuery += " AND " + filterStr
		countQuery += " AND " + filterStr
	}
	
	// Add sorting
	if params.SortBy != "" {
		sortOrder := "ASC"
		if params.SortOrder == "desc" {
			sortOrder = "DESC"
		}
		
		// Validate sort field
		validSortFields := map[string]string{
			"title":      "title",
			"created_at": "created_at",
			"updated_at": "updated_at",
		}
		
		if sortField, ok := validSortFields[params.SortBy]; ok {
			baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortField, sortOrder)
		} else {
			baseQuery += " ORDER BY created_at DESC"
		}
	} else {
		baseQuery += " ORDER BY created_at DESC"
	}
	
	// Add pagination
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, params.PageSize, (params.Page-1)*params.PageSize)
	
	// Get total count
	var totalCount int
	err := r.db.GetContext(ctx, &totalCount, countQuery, args[:argIndex-1]...)
	if err != nil {
		return nil, 0, err
	}
	
	// Get podcasts
	var podcasts []*models.Podcast
	err = r.db.SelectContext(ctx, &podcasts, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	
	// Get categories and episode counts for each podcast
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
	
	return podcasts, totalCount, nil
}

// CreateEpisode creates a new episode
func (r *repository) CreateEpisode(ctx context.Context, episode *models.Episode) error {
	query := `
		INSERT INTO episodes (
			id, podcast_id, title, description, audio_url, duration, cover_image_url,
			publication_date, guid, episode_number, season_number, transcript, status,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		) RETURNING id
	`

	if episode.ID == uuid.Nil {
		episode.ID = uuid.New()
	}

	now := time.Now()
	episode.CreatedAt = now
	episode.UpdatedAt = now
	
	if episode.PublicationDate.IsZero() {
		episode.PublicationDate = now
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		episode.ID,
		episode.PodcastID,
		episode.Title,
		episode.Description,
		episode.AudioURL,
		episode.Duration,
		episode.CoverImageURL,
		episode.PublicationDate,
		episode.GUID,
		episode.EpisodeNumber,
		episode.SeasonNumber,
		episode.Transcript,
		episode.Status,
		episode.CreatedAt,
		episode.UpdatedAt,
	).Scan(&episode.ID)

	return err
}

// GetEpisodeByID gets an episode by ID
func (r *repository) GetEpisodeByID(ctx context.Context, id uuid.UUID) (*models.Episode, error) {
	var episode models.Episode
	query := `
		SELECT
			id, podcast_id, title, description, audio_url, duration, cover_image_url,
			publication_date, guid, episode_number, season_number, transcript, status,
			created_at, updated_at
		FROM episodes
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &episode, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("episode not found")
		}
		return nil, err
	}

	return &episode, nil
}

// GetEpisodesByPodcastID gets episodes by podcast ID
func (r *repository) GetEpisodesByPodcastID(ctx context.Context, podcastID uuid.UUID, page, pageSize int) ([]*models.Episode, int, error) {
	query := `
		SELECT
			id, podcast_id, title, description, audio_url, duration, cover_image_url,
			publication_date, guid, episode_number, season_number, transcript, status,
			created_at, updated_at
		FROM episodes
		WHERE podcast_id = $1 AND status = 'active'
		ORDER BY 
			CASE 
				WHEN season_number IS NOT NULL AND episode_number IS NOT NULL 
				THEN season_number * 1000 + episode_number
				ELSE 999999
			END ASC,
			publication_date DESC
		LIMIT $2 OFFSET $3
	`

	var episodes []*models.Episode
	offset := (page - 1) * pageSize
	err := r.db.SelectContext(ctx, &episodes, query, podcastID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM episodes WHERE podcast_id = $1 AND status = 'active'`
	var count int
	err = r.db.GetContext(ctx, &count, countQuery, podcastID)
	if err != nil {
		return nil, 0, err
	}

	return episodes, count, nil
}

// UpdateEpisode updates an episode
func (r *repository) UpdateEpisode(ctx context.Context, episode *models.Episode) error {
	query := `
		UPDATE episodes SET
			title = $1,
			description = $2,
			audio_url = $3,
			duration = $4,
			cover_image_url = $5,
			publication_date = $6,
			guid = $7,
			episode_number = $8,
			season_number = $9,
			transcript = $10,
			status = $11,
			updated_at = $12
		WHERE id = $13
	`

	episode.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		episode.Title,
		episode.Description,
		episode.AudioURL,
		episode.Duration,
		episode.CoverImageURL,
		episode.PublicationDate,
		episode.GUID,
		episode.EpisodeNumber,
		episode.SeasonNumber,
		episode.Transcript,
		episode.Status,
		episode.UpdatedAt,
		episode.ID,
	)

	return err
}

// DeleteEpisode deletes an episode
func (r *repository) DeleteEpisode(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM episodes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ListEpisodes lists episodes with optional filtering
func (r *repository) ListEpisodes(ctx context.Context, params models.EpisodeSearchParams) ([]*models.Episode, int, error) {
	// Base query
	baseQuery := `
		SELECT 
			id, podcast_id, title, description, audio_url, duration, cover_image_url,
			publication_date, guid, episode_number, season_number, transcript, status,
			created_at, updated_at
		FROM episodes
		WHERE status = 'active'
	`
	
	// Count query
	countQuery := `
		SELECT COUNT(*)
		FROM episodes
		WHERE status = 'active'
	`
	
	// Add filters
	var filters []string
	var args []interface{}
	argIndex := 1
	
	if params.Query != "" {
		filters = append(filters, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+params.Query+"%")
		argIndex++
	}
	
	if params.PodcastID != "" {
		podcastID, err := uuid.Parse(params.PodcastID)
		if err == nil {
			filters = append(filters, fmt.Sprintf("podcast_id = $%d", argIndex))
			args = append(args, podcastID)
			argIndex++
		}
	}
	
	if !params.FromDate.IsZero() {
		filters = append(filters, fmt.Sprintf("publication_date >= $%d", argIndex))
		args = append(args, params.FromDate)
		argIndex++
	}
	
	if !params.ToDate.IsZero() {
		filters = append(filters, fmt.Sprintf("publication_date <= $%d", argIndex))
		args = append(args, params.ToDate)
		argIndex++
	}
	
	// Apply filters to queries
	if len(filters) > 0 {
		filterStr := strings.Join(filters, " AND ")
		baseQuery += " AND " + filterStr
		countQuery += " AND " + filterStr
	}
	
	// Add sorting
	if params.SortBy != "" {
		sortOrder := "ASC"
		if params.SortOrder == "desc" {
			sortOrder = "DESC"
		}
		
		// Validate sort field
		validSortFields := map[string]string{
			"title":            "title",
			"publication_date": "publication_date",
			"created_at":       "created_at",
			"updated_at":       "updated_at",
		}
		
		if sortField, ok := validSortFields[params.SortBy]; ok {
			baseQuery += fmt.Sprintf(" ORDER BY %s %s", sortField, sortOrder)
		} else {
			baseQuery += " ORDER BY publication_date DESC"
		}
	} else {
		baseQuery += " ORDER BY publication_date DESC"
	}
	
	// Add pagination
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, params.PageSize, (params.Page-1)*params.PageSize)
	
	// Get total count
	var totalCount int
	err := r.db.GetContext(ctx, &totalCount, countQuery, args[:argIndex-1]...)
	if err != nil {
		return nil, 0, err
	}
	
	// Get episodes
	var episodes []*models.Episode
	err = r.db.SelectContext(ctx, &episodes, baseQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	
	return episodes, totalCount, nil
}

// GetCategories gets all categories
func (r *repository) GetCategories(ctx context.Context) ([]*models.Category, error) {
	query := `
		SELECT id, name, description, icon_url, created_at, updated_at
		FROM categories
		ORDER BY name
	`
	
	var categories []*models.Category
	err := r.db.SelectContext(ctx, &categories, query)
	return categories, err
}

// AssociatePodcastWithCategories associates a podcast with categories
func (r *repository) AssociatePodcastWithCategories(ctx context.Context, podcastID uuid.UUID, categoryIDs []uuid.UUID) error {
	// First remove existing associations
	deleteQuery := `DELETE FROM podcast_categories WHERE podcast_id = $1`
	_, err := r.db.ExecContext(ctx, deleteQuery, podcastID)
	if err != nil {
		return err
	}
	
	// Add new associations
	for _, categoryID := range categoryIDs {
		insertQuery := `INSERT INTO podcast_categories (podcast_id, category_id) VALUES ($1, $2)`
		_, err := r.db.ExecContext(ctx, insertQuery, podcastID, categoryID)
		if err != nil {
			return err
		}
	}
	
	return nil
}

// SubscribeToPodcast subscribes a listener to a podcast
func (r *repository) SubscribeToPodcast(ctx context.Context, listenerID, podcastID uuid.UUID) error {
	query := `
		INSERT INTO subscriptions (listener_id, podcast_id)
		VALUES ($1, $2)
		ON CONFLICT (listener_id, podcast_id) DO NOTHING
	`
	
	_, err := r.db.ExecContext(ctx, query, listenerID, podcastID)
	return err
}

// UnsubscribeFromPodcast unsubscribes a listener from a podcast
func (r *repository) UnsubscribeFromPodcast(ctx context.Context, listenerID, podcastID uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE listener_id = $1 AND podcast_id = $2`
	_, err := r.db.ExecContext(ctx, query, listenerID, podcastID)
	return err
}

// GetSubscribedPodcasts gets podcasts subscribed by a listener
func (r *repository) GetSubscribedPodcasts(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.Podcast, int, error) {
	query := `
		SELECT p.id, p.podcaster_id, p.title, p.description, p.cover_image_url, p.rss_url, p.website_url,
		       p.language, p.author, p.category, p.subcategory, p.explicit, p.status, p.created_at, p.updated_at,
		       p.last_synced_at
		FROM podcasts p
		JOIN subscriptions s ON p.id = s.podcast_id
		WHERE s.listener_id = $1
		ORDER BY s.created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	var podcasts []*models.Podcast
	offset := (page - 1) * pageSize
	err := r.db.SelectContext(ctx, &podcasts, query, listenerID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM subscriptions
		WHERE listener_id = $1
	`
	
	var totalCount int
	err = r.db.GetContext(ctx, &totalCount, countQuery, listenerID)
	if err != nil {
		return nil, 0, err
	}
	
	// Get categories and episode counts for each podcast
	for _, podcast := range podcasts {
		categories, err := r.GetCategoriesByPodcastID(ctx, podcast.ID)
		if err != nil {
			return nil, 0, err
		}
		podcast.Categories = categories
		
		episodeCountQuery := `SELECT COUNT(*) FROM episodes WHERE podcast_id = $1 AND status = 'active'`
		var episodeCount int
		err = r.db.GetContext(ctx, &episodeCount, episodeCountQuery, podcast.ID)
		if err != nil {
			return nil, 0, err
		}
		podcast.EpisodeCount = episodeCount
	}
	
	return podcasts, totalCount, nil
}

// IsSubscribed checks if a listener is subscribed to a podcast
func (r *repository) IsSubscribed(ctx context.Context, listenerID, podcastID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM subscriptions 
			WHERE listener_id = $1 AND podcast_id = $2
		)
	`
	
	var isSubscribed bool
	err := r.db.GetContext(ctx, &isSubscribed, query, listenerID, podcastID)
	return isSubscribed, err
}

// SavePlaybackPosition saves the playback position for an episode
func (r *repository) SavePlaybackPosition(ctx context.Context, listenerID, episodeID uuid.UUID, position int, completed bool) error {
	query := `
		INSERT INTO playback_history (
			listener_id, episode_id, position, completed, updated_at
		) VALUES (
			$1, $2, $3, $4, $5
		) ON CONFLICT (listener_id, episode_id) DO UPDATE 
		SET position = $3, completed = $4, updated_at = $5
	`
	
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, listenerID, episodeID, position, completed, now)
	return err
}

// GetPlaybackPosition gets the playback position for an episode
func (r *repository) GetPlaybackPosition(ctx context.Context, listenerID, episodeID uuid.UUID) (int, bool, error) {
	query := `
		SELECT position, completed
		FROM playback_history
		WHERE listener_id = $1 AND episode_id = $2
	`
	
	var position int
	var completed bool
	err := r.db.QueryRowContext(ctx, query, listenerID, episodeID).Scan(&position, &completed)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil // Not found, return defaults
		}
		return 0, false, err
	}
	
	return position, completed, nil
}

// GetListeningHistory gets the listening history for a user
func (r *repository) GetListeningHistory(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.PlaybackHistory, int, error) {
	query := `
		SELECT 
			ph.id, ph.listener_id, ph.episode_id, ph.position, ph.completed, 
			ph.created_at, ph.updated_at,
			e.title as episode_title, e.podcast_id,
			p.title as podcast_title,
			COALESCE(e.cover_image_