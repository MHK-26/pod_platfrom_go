// pkg/content/models/models.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// Podcast represents a podcast
type Podcast struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	PodcasterID  uuid.UUID  `json:"podcaster_id" db:"podcaster_id"`
	Title        string     `json:"title" db:"title"`
	Description  string     `json:"description" db:"description"`
	CoverImageURL string    `json:"cover_image_url" db:"cover_image_url"`
	RSSUrl       string     `json:"rss_url" db:"rss_url"`
	WebsiteURL   string     `json:"website_url" db:"website_url"`
	Language     string     `json:"language" db:"language"`
	Author       string     `json:"author" db:"author"`
	Category     string     `json:"category" db:"category"`
	Subcategory  string     `json:"subcategory" db:"subcategory"`
	Explicit     bool       `json:"explicit" db:"explicit"`
	Status       string     `json:"status" db:"status"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
	LastSyncedAt *time.Time `json:"last_synced_at" db:"last_synced_at"`
	EpisodeCount int        `json:"episode_count,omitempty"`
	Categories   []*Category `json:"categories,omitempty"`
}

// Episode represents a podcast episode
type Episode struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	PodcastID       uuid.UUID  `json:"podcast_id" db:"podcast_id"`
	Title           string     `json:"title" db:"title"`
	Description     string     `json:"description" db:"description"`
	AudioURL        string     `json:"audio_url" db:"audio_url"`
	Duration        int        `json:"duration" db:"duration"`
	CoverImageURL   string     `json:"cover_image_url" db:"cover_image_url"`
	PublicationDate time.Time  `json:"publication_date" db:"publication_date"`
	GUID            string     `json:"guid" db:"guid"`
	EpisodeNumber   *int       `json:"episode_number" db:"episode_number"`
	SeasonNumber    *int       `json:"season_number" db:"season_number"`
	Transcript      string     `json:"transcript" db:"transcript"`
	Status          string     `json:"status" db:"status"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// Category represents a podcast category
type Category struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	IconURL     string    `json:"icon_url" db:"icon_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// PlaybackHistory represents a user's listening history for an episode
type PlaybackHistory struct {
	ID         uuid.UUID `json:"id" db:"id"`
	ListenerID uuid.UUID `json:"listener_id" db:"listener_id"`
	EpisodeID  uuid.UUID `json:"episode_id" db:"episode_id"`
	Position   int       `json:"position" db:"position"`
	Completed  bool      `json:"completed" db:"completed"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	
	// Joined data
	EpisodeTitle   string    `json:"episode_title" db:"episode_title"`
	PodcastID      uuid.UUID `json:"podcast_id" db:"podcast_id"`
	PodcastTitle   string    `json:"podcast_title" db:"podcast_title"`
	CoverImageURL  string    `json:"cover_image_url" db:"cover_image_url"`
}

// Comment represents a user comment on an episode
type Comment struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	EpisodeID uuid.UUID `json:"episode_id" db:"episode_id"`
	Content   string    `json:"content" db:"content"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	
	// Joined data
	Username       string `json:"username" db:"username"`
	UserFullName   string `json:"user_full_name" db:"user_full_name"`
	UserProfileURL string `json:"user_profile_url" db:"user_profile_url"`
}

// Playlist represents a user's playlist
type Playlist struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	IsPublic    bool      `json:"is_public" db:"is_public"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	
	// Metadata
	EpisodeCount int `json:"episode_count,omitempty"`
}

// PlaylistItem represents an episode in a playlist
type PlaylistItem struct {
	PlaylistID uuid.UUID `json:"playlist_id" db:"playlist_id"`
	EpisodeID  uuid.UUID `json:"episode_id" db:"episode_id"`
	Position   int       `json:"position" db:"position"`
	AddedAt    time.Time `json:"added_at" db:"added_at"`
	
	// Joined data
	EpisodeTitle   string    `json:"episode_title" db:"episode_title"`
	PodcastID      uuid.UUID `json:"podcast_id" db:"podcast_id"`
	PodcastTitle   string    `json:"podcast_title" db:"podcast_title"`
	Duration       int       `json:"duration" db:"duration"`
	CoverImageURL  string    `json:"cover_image_url" db:"cover_image_url"`
}

// Request/Response structures

// CreatePodcastRequest represents a request to create a podcast
type CreatePodcastRequest struct {
	Title        string   `json:"title" validate:"required"`
	Description  string   `json:"description" validate:"required"`
	CoverImageURL string  `json:"cover_image_url"`
	RSSUrl       string   `json:"rss_url" validate:"required,url"`
	WebsiteURL   string   `json:"website_url"`
	Language     string   `json:"language" validate:"required"`
	Category     string   `json:"category" validate:"required"`
	Subcategory  string   `json:"subcategory"`
	Explicit     bool     `json:"explicit"`
}

// UpdatePodcastRequest represents a request to update a podcast
type UpdatePodcastRequest struct {
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	CoverImageURL string `json:"cover_image_url"`
	WebsiteURL   string  `json:"website_url"`
	Language     string  `json:"language"`
	Category     string  `json:"category"`
	Subcategory  string  `json:"subcategory"`
	Explicit     bool    `json:"explicit"`
}

// SyncPodcastRequest represents a request to sync a podcast
type SyncPodcastRequest struct {
	PodcastID uuid.UUID `json:"podcast_id" validate:"required"`
}

// CreateEpisodeRequest represents a request to create an episode
type CreateEpisodeRequest struct {
	PodcastID       uuid.UUID `json:"podcast_id" validate:"required"`
	Title           string    `json:"title" validate:"required"`
	Description     string    `json:"description" validate:"required"`
	AudioURL        string    `json:"audio_url" validate:"required,url"`
	Duration        int       `json:"duration" validate:"required,min=1"`
	CoverImageURL   string    `json:"cover_image_url"`
	PublicationDate time.Time `json:"publication_date"`
	EpisodeNumber   *int      `json:"episode_number"`
	SeasonNumber    *int      `json:"season_number"`
	Transcript      string    `json:"transcript"`
}

// UpdateEpisodeRequest represents a request to update an episode
type UpdateEpisodeRequest struct {
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	CoverImageURL   string    `json:"cover_image_url"`
	PublicationDate time.Time `json:"publication_date"`
	EpisodeNumber   *int      `json:"episode_number"`
	SeasonNumber    *int      `json:"season_number"`
	Transcript      string    `json:"transcript"`
}

// PodcastResponse represents a podcast response with additional data
type PodcastResponse struct {
	Podcast
	EpisodeCount   int               `json:"episode_count"`
	LatestEpisodes []EpisodeResponse `json:"latest_episodes,omitempty"`
}

// EpisodeResponse represents an episode response with additional data
type EpisodeResponse struct {
	Episode
	PodcastTitle      string `json:"podcast_title"`
	PodcastAuthor     string `json:"podcast_author"`
	PodcastImageURL   string `json:"podcast_image_url"`
	ListenCount       int    `json:"listen_count"`
	AverageCompletion int    `json:"average_completion"` // percentage
}

// CreateCommentRequest represents a request to create a comment
type CreateCommentRequest struct {
	EpisodeID uuid.UUID `json:"episode_id" validate:"required"`
	Content   string    `json:"content" validate:"required"`
}

// CreatePlaylistRequest represents a request to create a playlist
type CreatePlaylistRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// UpdatePlaylistRequest represents a request to update a playlist
type UpdatePlaylistRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// AddToPlaylistRequest represents a request to add an episode to a playlist
type AddToPlaylistRequest struct {
	EpisodeID uuid.UUID `json:"episode_id" validate:"required"`
	Position  int       `json:"position"`
}

// SavePlaybackPositionRequest represents a request to save playback position
type SavePlaybackPositionRequest struct {
	EpisodeID uuid.UUID `json:"episode_id" validate:"required"`
	Position  int       `json:"position" validate:"required,min=0"`
	Completed bool      `json:"completed"`
}

// PodcastSearchParams represents parameters for searching podcasts
type PodcastSearchParams struct {
	Query      string    `form:"query"`
	Category   string    `form:"category"`
	Language   string    `form:"language"`
	SortBy     string    `form:"sort_by"`
	SortOrder  string    `form:"sort_order"`
	Page       int       `form:"page,default=1"`
	PageSize   int       `form:"page_size,default=20"`
}

// EpisodeSearchParams represents parameters for searching episodes
type EpisodeSearchParams struct {
	Query       string    `form:"query"`
	PodcastID   string    `form:"podcast_id"`
	FromDate    time.Time `form:"from_date"`
	ToDate      time.Time `form:"to_date"`
	SortBy      string    `form:"sort_by"`
	SortOrder   string    `form:"sort_order"`
	Page        int       `form:"page,default=1"`
	PageSize    int       `form:"page_size,default=20"`
}