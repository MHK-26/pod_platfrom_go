// pkg/analytics/repository/postgres/repository.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/MHK-26/pod_platfrom_go/pkg/analytics/models"
)

// Repository defines the methods for the analytics repository
type Repository interface {
	TrackListen(ctx context.Context, event *models.ListenEvent) error
	GetEpisodeListens(ctx context.Context, episodeID uuid.UUID, params models.AnalyticsParams) (*models.ListenStats, []models.TimePoint, error)
	GetPodcastListens(ctx context.Context, podcastID uuid.UUID, params models.AnalyticsParams) (*models.ListenStats, []models.TimePoint, []models.EpisodeStat, error)
	GetPodcasterListens(ctx context.Context, podcasterID uuid.UUID, params models.AnalyticsParams) (*models.PodcasterAnalytics, error)
	GetListeningHistory(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.ListeningHistoryItem, int, error)
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new analytics repository
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

// TrackListen adds a new listen event
func (r *repository) TrackListen(ctx context.Context, event *models.ListenEvent) error {
	query := `
		INSERT INTO listen_events (
			id, listener_id, episode_id, source, started_at, duration, completed,
			ip_address, user_agent, country_code, city
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id
	`

	if event.ID == uuid.Nil {
		event.ID = uuid.New()
	}

	if event.StartedAt.IsZero() {
		event.StartedAt = time.Now()
	}

	err := r.db.QueryRowContext(
		ctx,
		query,
		event.ID,
		event.ListenerID,
		event.EpisodeID,
		event.Source,
		event.StartedAt,
		event.Duration,
		event.Completed,
		event.IPAddress,
		event.UserAgent,
		event.CountryCode,
		event.City,
	).Scan(&event.ID)

	// Also update playback history
	if event.ListenerID != uuid.Nil {
		historyQuery := `
			INSERT INTO playback_history (
				listener_id, episode_id, position, completed
			) VALUES (
				$1, $2, $3, $4
			) ON CONFLICT (listener_id, episode_id) DO UPDATE 
			SET position = $3, completed = $4, updated_at = CURRENT_TIMESTAMP
		`

		_, histErr := r.db.ExecContext(
			ctx,
			historyQuery,
			event.ListenerID,
			event.EpisodeID,
			event.Duration,
			event.Completed,
		)

		if histErr != nil {
			// Log error but don't fail the main operation
			fmt.Printf("Error updating playback history: %v\n", histErr)
		}
	}

	return err
}

// GetEpisodeListens gets listen statistics for an episode
func (r *repository) GetEpisodeListens(ctx context.Context, episodeID uuid.UUID, params models.AnalyticsParams) (*models.ListenStats, []models.TimePoint, error) {
	// Get episode stats
	statsQuery := `
		SELECT 
			COUNT(*) as total_listens,
			COUNT(DISTINCT listener_id) as unique_listeners,
			AVG(duration) as average_listen_duration,
			(SUM(CASE WHEN completed THEN 1 ELSE 0 END)::float / COUNT(*)) * 100 as completion_rate
		FROM listen_events
		WHERE episode_id = $1
		AND started_at BETWEEN $2 AND $3
	`

	var stats models.ListenStats
	err := r.db.GetContext(ctx, &stats, statsQuery, episodeID, params.StartDate, params.EndDate)
	if err != nil {
		return nil, nil, err
	}

	// Get timeseries data
	var timeFormat string
	var groupBy string
	
	switch params.Interval {
	case "week":
		timeFormat = "YYYY-IW" // ISO week
		groupBy = "date_trunc('week', started_at)"
	case "month":
		timeFormat = "YYYY-MM"
		groupBy = "date_trunc('month', started_at)"
	default: // day
		timeFormat = "YYYY-MM-DD"
		groupBy = "date_trunc('day', started_at)"
	}

	timeSeriesQuery := `
		SELECT 
			to_char(${groupBy}, '${timeFormat}') as day_str,
			${groupBy} as timestamp,
			COUNT(*) as count
		FROM listen_events
		WHERE episode_id = $1
		AND started_at BETWEEN $2 AND $3
		GROUP BY day_str, ${groupBy}
		ORDER BY ${groupBy}
	`

	// Replace placeholders
	timeSeriesQuery = strings.ReplaceAll(timeSeriesQuery, "${groupBy}", groupBy)
	timeSeriesQuery = strings.ReplaceAll(timeSeriesQuery, "${timeFormat}", timeFormat)
	timeSeriesQuery = sqlx.Rebind(sqlx.DOLLAR, timeSeriesQuery)
	

	rows, err := r.db.QueryxContext(ctx, timeSeriesQuery, episodeID, params.StartDate, params.EndDate)
	if err != nil {
		return &stats, nil, err
	}
	defer rows.Close()

	var timePoints []models.TimePoint
	for rows.Next() {
		var tp struct {
			DayStr    string    `db:"day_str"`
			Timestamp time.Time `db:"timestamp"`
			Count     int       `db:"count"`
		}
		if err := rows.StructScan(&tp); err != nil {
			return &stats, nil, err
		}
		timePoints = append(timePoints, models.TimePoint{
			Timestamp: tp.Timestamp,
			Value:     tp.Count,
		})
	}

	if err := rows.Err(); err != nil {
		return &stats, nil, err
	}

	return &stats, timePoints, nil
}

// GetPodcasterListens gets listen statistics for all podcasts by a podcaster
func (r *repository) GetPodcasterListens(ctx context.Context, podcasterID uuid.UUID, params models.AnalyticsParams) (*models.PodcasterAnalytics, error) {
	// Initialize the result
	result := &models.PodcasterAnalytics{
		PodcasterID: podcasterID,
	}

	// Get total listens and unique listeners
	statsQuery := `
		SELECT 
			COUNT(*) as total_listens,
			COUNT(DISTINCT listener_id) as unique_listeners
		FROM listen_events le
		JOIN episodes e ON le.episode_id = e.id
		JOIN podcasts p ON e.podcast_id = p.id
		WHERE p.podcaster_id = $1
		AND le.started_at BETWEEN $2 AND $3
	`

	err := r.db.QueryRowContext(
		ctx, 
		statsQuery, 
		podcasterID, 
		params.StartDate, 
		params.EndDate,
	).Scan(&result.TotalListens, &result.UniqueListeners)
	
	if err != nil {
		return nil, err
	}

	// Get listens by day
	listensByDayQuery := `
		SELECT 
			date_trunc('day', le.started_at) as timestamp,
			COUNT(*) as value
		FROM listen_events le
		JOIN episodes e ON le.episode_id = e.id
		JOIN podcasts p ON e.podcast_id = p.id
		WHERE p.podcaster_id = $1
		AND le.started_at BETWEEN $2 AND $3
		GROUP BY timestamp
		ORDER BY timestamp
	`

	rows, err := r.db.QueryxContext(ctx, listensByDayQuery, podcasterID, params.StartDate, params.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tp models.TimePoint
		if err := rows.StructScan(&tp); err != nil {
			return nil, err
		}
		result.ListensByDay = append(result.ListensByDay, tp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Get listens by podcast
	listensByPodcastQuery := `
		SELECT 
			p.id as podcast_id,
			p.title,
			COUNT(le.*) as listens,
			COUNT(DISTINCT le.listener_id) as unique_listeners
		FROM podcasts p
		LEFT JOIN episodes e ON p.id = e.podcast_id
		LEFT JOIN listen_events le ON e.id = le.episode_id 
		AND le.started_at BETWEEN $2 AND $3
		WHERE p.podcaster_id = $1
		GROUP BY p.id, p.title
		ORDER BY listens DESC
	`

	rows, err = r.db.QueryxContext(ctx, listensByPodcastQuery, podcasterID, params.StartDate, params.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ps models.PodcastStat
		if err := rows.StructScan(&ps); err != nil {
			return nil, err
		}
		result.ListensByPodcast = append(result.ListensByPodcast, ps)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Get listens by country
	listensByCountryQuery := `
		SELECT 
			le.country_code as code,
			COUNT(*) as count
		FROM listen_events le
		JOIN episodes e ON le.episode_id = e.id
		JOIN podcasts p ON e.podcast_id = p.id
		WHERE p.podcaster_id = $1
		AND le.started_at BETWEEN $2 AND $3
		AND le.country_code IS NOT NULL
		GROUP BY le.country_code
		ORDER BY count DESC
	`

	rows, err = r.db.QueryxContext(ctx, listensByCountryQuery, podcasterID, params.StartDate, params.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var gs models.GeoStat
		if err := rows.StructScan(&gs); err != nil {
			return nil, err
		}
		result.ListensByCountry = append(result.ListensByCountry, gs)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Get listens by device
	listensByDeviceQuery := `
		SELECT 
			CASE 
				WHEN le.user_agent LIKE '%Android%' THEN 'Android'
				WHEN le.user_agent LIKE '%iPhone%' THEN 'iPhone'
				WHEN le.user_agent LIKE '%iPad%' THEN 'iPad'
				WHEN le.user_agent LIKE '%Windows%' THEN 'Windows'
				WHEN le.user_agent LIKE '%Mac%' THEN 'Mac'
				ELSE 'Other'
			END as device_type,
			COUNT(*) as count
		FROM listen_events le
		JOIN episodes e ON le.episode_id = e.id
		JOIN podcasts p ON e.podcast_id = p.id
		WHERE p.podcaster_id = $1
		AND le.started_at BETWEEN $2 AND $3
		GROUP BY device_type
		ORDER BY count DESC
	`

	rows, err = r.db.QueryxContext(ctx, listensByDeviceQuery, podcasterID, params.StartDate, params.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ds models.DeviceStat
		if err := rows.StructScan(&ds); err != nil {
			return nil, err
		}
		result.ListensByDevice = append(result.ListensByDevice, ds)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Get total subscribers
	subscribersQuery := `
		SELECT COUNT(*) 
		FROM subscriptions s
		JOIN podcasts p ON s.podcast_id = p.id
		WHERE p.podcaster_id = $1
	`

	err = r.db.QueryRowContext(ctx, subscribersQuery, podcasterID).Scan(&result.TotalSubscribers)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return result, nil
}

// GetListeningHistory gets the listening history for a user
func (r *repository) GetListeningHistory(ctx context.Context, listenerID uuid.UUID, page, pageSize int) ([]*models.ListeningHistoryItem, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM playback_history ph
		WHERE ph.listener_id = $1
	`

	var totalCount int
	err := r.db.GetContext(ctx, &totalCount, countQuery, listenerID)
	if err != nil {
		return nil, 0, err
	}

	// Get history items with pagination
	offset := (page - 1) * pageSize
	historyQuery := `
		SELECT 
			ph.episode_id,
			e.title as episode_title,
			e.podcast_id,
			p.title as podcast_title,
			ph.updated_at as listened_at,
			ph.position as duration,
			ph.completed,
			COALESCE(e.cover_image_url, p.cover_image_url) as cover_image_url
		FROM playback_history ph
		JOIN episodes e ON ph.episode_id = e.id
		JOIN podcasts p ON e.podcast_id = p.id
		WHERE ph.listener_id = $1
		ORDER BY ph.updated_at DESC
		LIMIT $2 OFFSET $3
	`

	var history []*models.ListeningHistoryItem
	err = r.db.SelectContext(ctx, &history, historyQuery, listenerID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return history, totalCount, nil
}

// GetPodcastListens gets listen statistics for a podcast
func (r *repository) GetPodcastListens(ctx context.Context, podcastID uuid.UUID, params models.AnalyticsParams) (*models.ListenStats, []models.TimePoint, []models.EpisodeStat, error) {
	// Get podcast stats
	statsQuery := `
		SELECT 
			COUNT(*) as total_listens,
			COUNT(DISTINCT listener_id) as unique_listeners,
			AVG(duration) as average_listen_duration,
			(SUM(CASE WHEN completed THEN 1 ELSE 0 END)::float / COUNT(*)) * 100 as completion_rate
		FROM listen_events le
		JOIN episodes e ON le.episode_id = e.id
		WHERE e.podcast_id = $1
		AND le.started_at BETWEEN $2 AND $3
	`

	var stats models.ListenStats
	err := r.db.GetContext(ctx, &stats, statsQuery, podcastID, params.StartDate, params.EndDate)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get timeseries data
	var timeFormat string
	var groupBy string
	
	switch params.Interval {
	case "week":
		timeFormat = "YYYY-IW" // ISO week
		groupBy = "date_trunc('week', le.started_at)"
	case "month":
		timeFormat = "YYYY-MM"
		groupBy = "date_trunc('month', le.started_at)"
	default: // day
		timeFormat = "YYYY-MM-DD"
		groupBy = "date_trunc('day', le.started_at)"
	}

	timeSeriesQuery := fmt.Sprintf(`
		SELECT 
			%s as timestamp,
			COUNT(*) as count
		FROM listen_events le
		JOIN episodes e ON le.episode_id = e.id
		WHERE e.podcast_id = $1
		AND le.started_at BETWEEN $2 AND $3
		GROUP BY timestamp
		ORDER BY timestamp
	`, groupBy)

	rows, err := r.db.QueryxContext(ctx, timeSeriesQuery, podcastID, params.StartDate, params.EndDate)
	if err != nil {
		return &stats, nil, nil, err
	}
	defer rows.Close()

	var timePoints []models.TimePoint
	for rows.Next() {
		var tp struct {
			Timestamp time.Time `db:"timestamp"`
			Count     int       `db:"count"`
		}
		if err := rows.StructScan(&tp); err != nil {
			return &stats, nil, nil, err
		}
		timePoints = append(timePoints, models.TimePoint{
			Timestamp: tp.Timestamp,
			Value:     tp.Count,
		})
	}

	if err := rows.Err(); err != nil {
		return &stats, nil, nil, err
	}

	// Get episode stats
	episodeStatsQuery := `
		SELECT 
			e.id as episode_id,
			e.title,
			COUNT(le.*) as listens,
			AVG(le.duration) as average_listen_duration,
			(SUM(CASE WHEN le.completed THEN 1 ELSE 0 END)::float / COUNT(*)) * 100 as completion_rate
		FROM episodes e
		LEFT JOIN listen_events le ON e.id = le.episode_id
		AND le.started_at BETWEEN $2 AND $3
		WHERE e.podcast_id = $1
		GROUP BY e.id, e.title
		ORDER BY listens DESC
	`

	rows, err = r.db.QueryxContext(ctx, episodeStatsQuery, podcastID, params.StartDate, params.EndDate)
	if err != nil {
		return &stats, timePoints, nil, err
	}
	defer rows.Close()

	var episodeStats []models.EpisodeStat
	for rows.Next() {
		var es models.EpisodeStat
		if err := rows.StructScan(&es); err != nil {
			return &stats, timePoints, nil, err
		}
		episodeStats = append(episodeStats, es)
	}

	if err := rows.Err(); err != nil {
		return &stats, timePoints, nil, err
	}

	return &stats, timePoints, episodeStats, nil
}