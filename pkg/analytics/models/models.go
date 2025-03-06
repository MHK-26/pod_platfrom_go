// pkg/analytics/models/models.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// ListenEvent represents a podcast listening event
type ListenEvent struct {
	ID          uuid.UUID `json:"id" db:"id"`
	ListenerID  uuid.UUID `json:"listener_id" db:"listener_id"`
	EpisodeID   uuid.UUID `json:"episode_id" db:"episode_id"`
	Source      string    `json:"source" db:"source"`
	StartedAt   time.Time `json:"started_at" db:"started_at"`
	Duration    int       `json:"duration" db:"duration"`
	Completed   bool      `json:"completed" db:"completed"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	UserAgent   string    `json:"user_agent" db:"user_agent"`
	CountryCode string    `json:"country_code" db:"country_code"`
	City        string    `json:"city" db:"city"`
}

// TrackListenRequest represents a request to track a listen event
type TrackListenRequest struct {
	ListenerID  uuid.UUID `json:"listener_id" validate:"required"`
	EpisodeID   uuid.UUID `json:"episode_id" validate:"required"`
	Source      string    `json:"source" validate:"required,oneof=mobile web embed"`
	Duration    int       `json:"duration" validate:"required,min=1"`
	Completed   bool      `json:"completed"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	CountryCode string    `json:"country_code"`
	City        string    `json:"city"`
}

// ListenStats represents listening statistics
type ListenStats struct {
	TotalListens         int     `json:"total_listens"`
	UniqueListeners      int     `json:"unique_listeners"`
	AverageListenDuration float64 `json:"average_listen_duration"`
	CompletionRate       float64 `json:"completion_rate"`
}

// EpisodeAnalytics represents analytics for an episode
type EpisodeAnalytics struct {
	EpisodeID           uuid.UUID   `json:"episode_id"`
	Title               string      `json:"title"`
	ListenStats         ListenStats `json:"listen_stats"`
	ListensByDay        []TimePoint `json:"listens_by_day"`
	ListensBySource     []SourceStat `json:"listens_by_source"`
	ListensByCountry    []GeoStat   `json:"listens_by_country"`
	ListensByCity       []GeoStat   `json:"listens_by_city"`
	RetentionGraph      []TimePoint `json:"retention_graph"`
}

// PodcastAnalytics represents analytics for a podcast
type PodcastAnalytics struct {
	PodcastID          uuid.UUID     `json:"podcast_id"`
	Title              string        `json:"title"`
	ListenStats        ListenStats   `json:"listen_stats"`
	ListensByDay       []TimePoint   `json:"listens_by_day"`
	ListensByEpisode   []EpisodeStat `json:"listens_by_episode"`
	ListensBySource    []SourceStat  `json:"listens_by_source"`
	ListensByCountry   []GeoStat     `json:"listens_by_country"`
	SubscribersByDay   []TimePoint   `json:"subscribers_by_day"`
	CurrentSubscribers int           `json:"current_subscribers"`
}

// PodcasterAnalytics represents analytics for a podcaster
type PodcasterAnalytics struct {
	PodcasterID        uuid.UUID      `json:"podcaster_id"`
	TotalListens       int            `json:"total_listens"`
	UniqueListeners    int            `json:"unique_listeners"`
	TotalSubscribers   int            `json:"total_subscribers"`
	ListensByDay       []TimePoint    `json:"listens_by_day"`
	ListensByPodcast   []PodcastStat  `json:"listens_by_podcast"`
	SubscribersByDay   []TimePoint    `json:"subscribers_by_day"`
	ListensByCountry   []GeoStat      `json:"listens_by_country"`
	ListensByDevice    []DeviceStat   `json:"listens_by_device"`
}

// TimePoint represents a data point with a timestamp
type TimePoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     int       `json:"value"`
}

// EpisodeStat represents statistics for an episode
type EpisodeStat struct {
	EpisodeID           uuid.UUID `json:"episode_id"`
	Title               string    `json:"title"`
	Listens             int       `json:"listens"`
	UniqueListeners     int       `json:"unique_listeners"`
	AverageListenDuration float64  `json:"average_listen_duration"`
	CompletionRate      float64   `json:"completion_rate"`
}

// PodcastStat represents statistics for a podcast
type PodcastStat struct {
	PodcastID       uuid.UUID `json:"podcast_id"`
	Title           string    `json:"title"`
	Listens         int       `json:"listens"`
	UniqueListeners int       `json:"unique_listeners"`
	Subscribers     int       `json:"subscribers"`
}

// SourceStat represents statistics for a source
type SourceStat struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
}

// GeoStat represents statistics for a geographic location
type GeoStat struct {
	Code  string `json:"code"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// DeviceStat represents statistics for a device type
type DeviceStat struct {
	DeviceType string `json:"device_type"`
	Count      int    `json:"count"`
}

// ListeningHistoryItem represents an item in the listening history
type ListeningHistoryItem struct {
	EpisodeID      uuid.UUID `json:"episode_id" db:"episode_id"`
	EpisodeTitle   string    `json:"episode_title" db:"episode_title"`
	PodcastID      uuid.UUID `json:"podcast_id" db:"podcast_id"`
	PodcastTitle   string    `json:"podcast_title" db:"podcast_title"`
	ListenedAt     time.Time `json:"listened_at" db:"listened_at"`
	Duration       int       `json:"duration" db:"duration"`
	Completed      bool      `json:"completed" db:"completed"`
	CoverImageURL  string    `json:"cover_image_url" db:"cover_image_url"`
}

// AnalyticsParams represents parameters for analytics queries
type AnalyticsParams struct {
	StartDate   time.Time `json:"start_date" form:"start_date"`
	EndDate     time.Time `json:"end_date" form:"end_date"`
	Interval    string    `json:"interval" form:"interval" validate:"omitempty,oneof=day week month"`
	GroupBy     string    `json:"group_by" form:"group_by" validate:"omitempty,oneof=source country device"`
	CountryCode string    `json:"country_code" form:"country_code"`
}