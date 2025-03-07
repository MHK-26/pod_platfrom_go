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

// GetListeningHistory gets the listening history for a user
func (r *repository) GetListeningHistory(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.PlaybackHistory, int, error) {
	query := `
		SELECT 
			ph.id, ph.listener_id, ph.episode_id, ph.position, ph.completed, 
			ph.created_at, ph.updated_at,
			e.title as episode_title, e.podcast_id,
			p.title as podcast_title,
			COALESCE(e.cover_image_url, p.cover_image_url) as cover_image_url
		FROM playback_history ph
		JOIN episodes e ON ph.episode_id = e.id
		JOIN podcasts p ON e.podcast_id = p.id
		WHERE ph.listener_id = $1
		ORDER BY ph.updated_at DESC
		LIMIT $2 OFFSET $3
	`

	var history []*models.PlaybackHistory
	offset := (page - 1) * pageSize
	err := r.db.SelectContext(ctx, &history, query, listenerID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM playback_history WHERE listener_id = $1`
	var totalCount int
	err = r.db.GetContext(ctx, &totalCount, countQuery, listenerID)
	if err != nil {
		return nil, 0, err
	}

	return history, totalCount, nil
}

// LikeEpisode adds a like to an episode
func (r *repository) LikeEpisode(ctx context.Context, listenerID, episodeID uuid.UUID) error {
	query := `
		INSERT INTO likes (listener_id, episode_id)
		VALUES ($1, $2)
		ON CONFLICT (listener_id, episode_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, listenerID, episodeID)
	return err
}

// UnlikeEpisode removes a like from an episode
func (r *repository) UnlikeEpisode(ctx context.Context, listenerID, episodeID uuid.UUID) error {
	query := `DELETE FROM likes WHERE listener_id = $1 AND episode_id = $2`
	_, err := r.db.ExecContext(ctx, query, listenerID, episodeID)
	return err
}

// IsEpisodeLiked checks if a listener has liked an episode
func (r *repository) IsEpisodeLiked(ctx context.Context, listenerID, episodeID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM likes
			WHERE listener_id = $1 AND episode_id = $2
		)
	`

	var liked bool
	err := r.db.GetContext(ctx, &liked, query, listenerID, episodeID)
	return liked, err
}

// GetLikedEpisodes gets a list of episodes liked by a listener
func (r *repository) GetLikedEpisodes(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.Episode, int, error) {
	query := `
		SELECT e.id, e.podcast_id, e.title, e.description, e.audio_url, e.duration, 
			e.cover_image_url, e.publication_date, e.guid, e.episode_number, e.season_number, 
			e.transcript, e.status, e.created_at, e.updated_at
		FROM episodes e
		JOIN likes l ON e.id = l.episode_id
		WHERE l.listener_id = $1 AND e.status = 'active'
		ORDER BY l.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var episodes []*models.Episode
	offset := (page - 1) * pageSize
	err := r.db.SelectContext(ctx, &episodes, query, listenerID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM likes l
		JOIN episodes e ON l.episode_id = e.id
		WHERE l.listener_id = $1 AND e.status = 'active'
	`

	var totalCount int
	err = r.db.GetContext(ctx, &totalCount, countQuery, listenerID)
	if err != nil {
		return nil, 0, err
	}

	return episodes, totalCount, nil
}

// AddComment adds a comment to an episode
func (r *repository) AddComment(ctx context.Context, comment *models.Comment) error {
	query := `
		INSERT INTO comments (
			id, user_id, episode_id, content, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id
	`

	if comment.ID == uuid.Nil {
		comment.ID = uuid.New()
	}

	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now

	if comment.Status == "" {
		comment.Status = "active"
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		comment.ID,
		comment.UserID,
		comment.EpisodeID,
		comment.Content,
		comment.Status,
		comment.CreatedAt,
		comment.UpdatedAt,
	).Scan(&comment.ID)

	return err
}

// GetCommentsByEpisodeID gets comments for an episode
func (r *repository) GetCommentsByEpisodeID(ctx context.Context, episodeID uuid.UUID, page, pageSize int) ([]*models.Comment, int, error) {
	query := `
		SELECT 
			c.id, c.user_id, c.episode_id, c.content, c.status, c.created_at, c.updated_at,
			u.username, u.full_name, u.profile_image_url as user_profile_url
		FROM comments c
		JOIN users u ON c.user_id = u.id
		WHERE c.episode_id = $1 AND c.status = 'active'
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var comments []*models.Comment
	offset := (page - 1) * pageSize
	err := r.db.SelectContext(ctx, &comments, query, episodeID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM comments
		WHERE episode_id = $1 AND status = 'active'
	`

	var totalCount int
	err = r.db.GetContext(ctx, &totalCount, countQuery, episodeID)
	if err != nil {
		return nil, 0, err
	}

	return comments, totalCount, nil
}

// DeleteComment deletes a comment
func (r *repository) DeleteComment(ctx context.Context, commentID, userID uuid.UUID) error {
	// First check if user owns the comment
	checkQuery := `SELECT user_id FROM comments WHERE id = $1`
	var commentUserID uuid.UUID
	err := r.db.GetContext(ctx, &commentUserID, checkQuery, commentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("comment not found")
		}
		return err
	}

	// Only allow deletion if the user is the comment author
	if commentUserID != userID {
		return errors.New("not authorized to delete this comment")
	}

	// Delete the comment
	deleteQuery := `DELETE FROM comments WHERE id = $1`
	_, err = r.db.ExecContext(ctx, deleteQuery, commentID)
	return err
}

// CreatePlaylist creates a new playlist
func (r *repository) CreatePlaylist(ctx context.Context, playlist *models.Playlist) error {
	query := `
		INSERT INTO playlists (
			id, user_id, name, description, is_public, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id
	`

	if playlist.ID == uuid.Nil {
		playlist.ID = uuid.New()
	}

	now := time.Now()
	playlist.CreatedAt = now
	playlist.UpdatedAt = now

	err := r.db.QueryRowContext(
		ctx,
		query,
		playlist.ID,
		playlist.UserID,
		playlist.Name,
		playlist.Description,
		playlist.IsPublic,
		playlist.CreatedAt,
		playlist.UpdatedAt,
	).Scan(&playlist.ID)

	return err
}

// GetPlaylistByID gets a playlist by ID
func (r *repository) GetPlaylistByID(ctx context.Context, id, userID uuid.UUID) (*models.Playlist, error) {
	var playlist models.Playlist
	query := `
		SELECT id, user_id, name, description, is_public, created_at, updated_at
		FROM playlists
		WHERE id = $1 AND (user_id = $2 OR is_public = true)
	`

	err := r.db.GetContext(ctx, &playlist, query, id, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("playlist not found or not accessible")
		}
		return nil, err
	}

	// Get episode count
	countQuery := `SELECT COUNT(*) FROM playlist_items WHERE playlist_id = $1`
	var episodeCount int
	err = r.db.GetContext(ctx, &episodeCount, countQuery, id)
	if err != nil {
		return nil, err
	}

	playlist.EpisodeCount = episodeCount

	return &playlist, nil
}

// GetUserPlaylists gets playlists for a user
func (r *repository) GetUserPlaylists(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*models.Playlist, int, error) {
	query := `
		SELECT id, user_id, name, description, is_public, created_at, updated_at
		FROM playlists
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var playlists []*models.Playlist
	offset := (page - 1) * pageSize
	err := r.db.SelectContext(ctx, &playlists, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM playlists WHERE user_id = $1`
	var totalCount int
	err = r.db.GetContext(ctx, &totalCount, countQuery, userID)
	if err != nil {
		return nil, 0, err
	}

	// Get episode counts for each playlist
	for _, playlist := range playlists {
		episodeCountQuery := `SELECT COUNT(*) FROM playlist_items WHERE playlist_id = $1`
		var episodeCount int
		err = r.db.GetContext(ctx, &episodeCount, episodeCountQuery, playlist.ID)
		if err != nil {
			return nil, 0, err
		}
		playlist.EpisodeCount = episodeCount
	}

	return playlists, totalCount, nil
}

// UpdatePlaylist updates a playlist
func (r *repository) UpdatePlaylist(ctx context.Context, playlist *models.Playlist) error {
	// First check if user owns the playlist
	checkQuery := `SELECT user_id FROM playlists WHERE id = $1`
	var playlistUserID uuid.UUID
	err := r.db.GetContext(ctx, &playlistUserID, checkQuery, playlist.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("playlist not found")
		}
		return err
	}

	// Only allow updates if the user is the playlist owner
	if playlistUserID != playlist.UserID {
		return errors.New("not authorized to update this playlist")
	}

	// Update the playlist
	query := `
		UPDATE playlists SET
			name = $1,
			description = $2,
			is_public = $3,
			updated_at = $4
		WHERE id = $5
	`

	playlist.UpdatedAt = time.Now()

	_, err = r.db.ExecContext(
		ctx,
		query,
		playlist.Name,
		playlist.Description,
		playlist.IsPublic,
		playlist.UpdatedAt,
		playlist.ID,
	)

	return err
}

// DeletePlaylist deletes a playlist
func (r *repository) DeletePlaylist(ctx context.Context, id, userID uuid.UUID) error {
	// First check if user owns the playlist
	checkQuery := `SELECT user_id FROM playlists WHERE id = $1`
	var playlistUserID uuid.UUID
	err := r.db.GetContext(ctx, &playlistUserID, checkQuery, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("playlist not found")
		}
		return err
	}

	// Only allow deletion if the user is the playlist owner
	if playlistUserID != userID {
		return errors.New("not authorized to delete this playlist")
	}

	// Delete the playlist
	deleteQuery := `DELETE FROM playlists WHERE id = $1`
	_, err = r.db.ExecContext(ctx, deleteQuery, id)
	return err
}

// AddToPlaylist adds an episode to a playlist
func (r *repository) AddToPlaylist(ctx context.Context, playlistID, episodeID uuid.UUID, position int) error {
	// Check if the episode exists
	episodeQuery := `SELECT id FROM episodes WHERE id = $1 AND status = 'active'`
	var episode uuid.UUID
	err := r.db.GetContext(ctx, &episode, episodeQuery, episodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("episode not found")
		}
		return err
	}

	// If position is not specified, get the next position
	if position <= 0 {
		positionQuery := `
			SELECT COALESCE(MAX(position), 0) + 1
			FROM playlist_items
			WHERE playlist_id = $1
		`
		err = r.db.GetContext(ctx, &position, positionQuery, playlistID)
		if err != nil {
			return err
		}
	}

	// Add the episode to the playlist
	query := `
		INSERT INTO playlist_items (playlist_id, episode_id, position, added_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (playlist_id, episode_id) DO UPDATE
		SET position = $3, added_at = $4
	`

	_, err = r.db.ExecContext(ctx, query, playlistID, episodeID, position, time.Now())
	return err
}

// RemoveFromPlaylist removes an episode from a playlist
func (r *repository) RemoveFromPlaylist(ctx context.Context, playlistID, episodeID uuid.UUID) error {
	query := `DELETE FROM playlist_items WHERE playlist_id = $1 AND episode_id = $2`
	_, err := r.db.ExecContext(ctx, query, playlistID, episodeID)
	return err
}

// GetPlaylistItems gets episodes in a playlist
func (r *repository) GetPlaylistItems(ctx context.Context, playlistID uuid.UUID, page, pageSize int) ([]*models.PlaylistItem, int, error) {
	query := `
		SELECT 
			pi.playlist_id, pi.episode_id, pi.position, pi.added_at,
			e.title AS episode_title, e.podcast_id, e.duration,
			p.title AS podcast_title,
			COALESCE(e.cover_image_url, p.cover_image_url) AS cover_image_url
		FROM playlist_items pi
		JOIN episodes e ON pi.episode_id = e.id
		JOIN podcasts p ON e.podcast_id = p.id
		WHERE pi.playlist_id = $1
		ORDER BY pi.position
		LIMIT $2 OFFSET $3
	`

	var items []*models.PlaylistItem
	offset := (page - 1) * pageSize
	err := r.db.SelectContext(ctx, &items, query, playlistID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM playlist_items WHERE playlist_id = $1`
	var totalCount int
	err = r.db.GetContext(ctx, &totalCount, countQuery, playlistID)
	if err != nil {
		return nil, 0, err
	}

	return items, totalCount, nil
}