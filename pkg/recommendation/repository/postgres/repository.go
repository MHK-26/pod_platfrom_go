// pkg/recommendation/repository/postgres/repository.go
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/your-username/podcast-platform/pkg/recommendation/models"
)

// Repository defines the methods for the recommendation repository
type Repository interface {
	// User-based recommendations
	GetPersonalizedRecommendations(ctx context.Context, userID uuid.UUID, limit int, excludedIDs []uuid.UUID) ([]models.RecommendedItem, error)
	
	// Similar content recommendations
	GetSimilarPodcasts(ctx context.Context, podcastID uuid.UUID, limit int, excludedIDs []uuid.UUID) ([]models.RecommendedItem, error)
	GetSimilarEpisodes(ctx context.Context, episodeID uuid.UUID, limit int, excludedIDs []uuid.UUID) ([]models.RecommendedItem, error)
	
	// Popular content recommendations
	GetTrendingPodcasts(ctx context.Context, timeRange string, limit int, excludedIDs []uuid.UUID) ([]models.RecommendedItem, error)
	GetPopularInCategory(ctx context.Context, categoryID uuid.UUID, limit int, excludedIDs []uuid.UUID) ([]models.RecommendedItem, error)
	
	// User preferences management
	UpdateUserPreference(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, weight float64) error
	GetUserPreferences(ctx context.Context, userID uuid.UUID) ([]models.UserPreference, error)
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new recommendation repository
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

// GetPersonalizedRecommendations gets personalized recommendations for a user
func (r *repository) GetPersonalizedRecommendations(ctx context.Context, userID uuid.UUID, limit int, excludedIDs []uuid.UUID) ([]models.RecommendedItem, error) {
	// In a production scenario, this would use a sophisticated recommendation algorithm
	// For now, we'll implement a simpler version based on categories the user has engaged with
	
	// Build the exclusion list for the query
	var excludedIDsParam interface{}
	var excludeCondition string
	if len(excludedIDs) > 0 {
		excludedIDsParam = excludedIDs
		excludeCondition = "AND p.id != ANY($3)"
	} else {
		excludedIDsParam = nil
		excludeCondition = ""
	}
	
	// Base the recommendations on user's listening history categories and subscriptions
	query := fmt.Sprintf(`
		WITH user_categories AS (
			-- Categories from user's listen history
			SELECT DISTINCT c.id AS category_id
			FROM listen_events le
			JOIN episodes e ON le.episode_id = e.id
			JOIN podcasts p ON e.podcast_id = p.id
			JOIN podcast_categories pc ON p.id = pc.podcast_id
			JOIN categories c ON pc.category_id = c.id
			WHERE le.listener_id = $1
			
			UNION
			
			-- Categories from user's subscriptions
			SELECT DISTINCT c.id AS category_id
			FROM subscriptions s
			JOIN podcasts p ON s.podcast_id = p.id
			JOIN podcast_categories pc ON p.id = pc.podcast_id
			JOIN categories c ON pc.category_id = c.id
			WHERE s.listener_id = $1
		)
		
		SELECT 
			p.id,
			'podcast' AS type,
			p.title,
			p.description,
			p.cover_image_url AS image_url,
			p.id AS podcast_id,
			p.title AS podcast_title,
			-- Simple scoring based on number of matching categories and listen counts
			(
				SELECT COUNT(*)::float 
				FROM podcast_categories pc2 
				JOIN user_categories uc ON pc2.category_id = uc.category_id
				WHERE pc2.podcast_id = p.id
			) * 10 +
			(
				SELECT COALESCE(COUNT(le.id), 0)::float
				FROM listen_events le
				JOIN episodes e ON le.episode_id = e.id
				WHERE e.podcast_id = p.id
			) / 100 AS score
		FROM podcasts p
		JOIN podcast_categories pc ON p.id = pc.podcast_id
		JOIN user_categories uc ON pc.category_id = uc.category_id
		-- Exclude podcasts the user is already subscribed to
		WHERE p.id NOT IN (
			SELECT podcast_id FROM subscriptions WHERE listener_id = $1
		)
		-- Exclude specified podcasts
		%s
		AND p.status = 'active'
		GROUP BY p.id, p.title, p.description, p.cover_image_url
		ORDER BY score DESC
		LIMIT $2
	`, excludeCondition)
	
	var items []models.RecommendedItem
	var err error
	
	if len(excludedIDs) > 0 {
		err = r.db.SelectContext(ctx, &items, query, userID, limit, excludedIDsParam)
	} else {
		err = r.db.SelectContext(ctx, &items, query, userID, limit)
	}
	
	if err != nil {
		return nil, err
	}
	
	// If we couldn't find enough recommendations based on user behavior,
	// supplement with trending podcasts
	if len(items) < limit {
		trendings, err := r.GetTrendingPodcasts(ctx, "weekly", limit-len(items), excludedIDs)
		if err != nil {
			return items, nil // Return what we have even if trending query fails
		}
		
		// Add trending items, but avoid duplicates
		existingIDs := make(map[uuid.UUID]bool)
		for _, item := range items {
			existingIDs[item.ID] = true
		}
		
		for _, trending := range trendings {
			if !existingIDs[trending.ID] {
				items = append(items, trending)
				existingIDs[trending.ID] = true
			}
		}
	}
	
	return items, nil
}

// GetSimilarPodcasts gets podcasts similar to a specified podcast
func (r *repository) GetSimilarPodcasts(ctx context.Context, podcastID uuid.UUID, limit int, excludedIDs []uuid.UUID) ([]models.RecommendedItem, error) {
	// Build the exclusion list for the query
	var excludedIDsParam interface{}
	var excludeCondition string
	if len(excludedIDs) > 0 {
		excludedIDsParam = excludedIDs
		excludeCondition = "AND p2.id != ANY($3) AND p2.id != $1"
	} else {
		excludedIDsParam = nil
		excludeCondition = "AND p2.id != $1"
	}
	
	// Find similar podcasts based on category overlap
	query := fmt.Sprintf(`
		WITH podcast_cats AS (
			SELECT category_id
			FROM podcast_categories
			WHERE podcast_id = $1
		)
		
		SELECT 
			p2.id,
			'podcast' AS type,
			p2.title,
			p2.description,
			p2.cover_image_url AS image_url,
			p2.id AS podcast_id,
			p2.title AS podcast_title,
			-- Score based on category overlap
			(
				SELECT COUNT(*)::float 
				FROM podcast_categories pc2 
				JOIN podcast_cats pc ON pc2.category_id = pc.category_id
				WHERE pc2.podcast_id = p2.id
			) / 
			(
				SELECT COUNT(*)::float 
				FROM podcast_categories 
				WHERE podcast_id = p2.id
			) * 100 AS score
		FROM podcasts p2
		WHERE EXISTS (
			SELECT 1 
			FROM podcast_categories pc2 
			JOIN podcast_cats pc ON pc2.category_id = pc.category_id
			WHERE pc2.podcast_id = p2.id
		)
		%s
		AND p2.status = 'active'
		ORDER BY score DESC
		LIMIT $2
	`, excludeCondition)
	
	var items []models.RecommendedItem
	var err error
	
	if len(excludedIDs) > 0 {
		err = r.db.SelectContext(ctx, &items, query, podcastID, limit, excludedIDsParam)
	} else {
		err = r.db.SelectContext(ctx, &items, query, podcastID, limit)
	}
	
	return items, err
}

// GetSimilarEpisodes gets episodes similar to a specified episode
func (r *repository) GetSimilarEpisodes(ctx context.Context, episodeID uuid.UUID, limit int, excludedIDs []uuid.UUID) ([]models.RecommendedItem, error) {
	// Build the exclusion list for the query
	var excludedIDsParam interface{}
	var excludeCondition string
	if len(excludedIDs) > 0 {
		excludedIDsParam = excludedIDs
		excludeCondition = "AND e2.id != ANY($3) AND e2.id != $1"
	} else {
		excludedIDsParam = nil
		excludeCondition = "AND e2.id != $1"
	}
	
	// Get the source episode details first
	sourceEpisodeQuery := `
		SELECT e.podcast_id, e.title
		FROM episodes e
		WHERE e.id = $1
	`
	
	var sourcePodcastID uuid.UUID
	var sourceTitle string
	err := r.db.QueryRowContext(ctx, sourceEpisodeQuery, episodeID).Scan(&sourcePodcastID, &sourceTitle)
	if err != nil {
		return nil, err
	}
	
	// Find similar episodes based on same podcast and title similarity
	query := fmt.Sprintf(`
		SELECT 
			e2.id,
			'episode' AS type,
			e2.title,
			e2.description,
			COALESCE(e2.cover_image_url, p.cover_image_url) AS image_url,
			p.id AS podcast_id,
			p.title AS podcast_title,
			-- Score based on same podcast and title similarity
			CASE
				WHEN e2.podcast_id = $4 THEN 50
				ELSE 10
			END +
			-- Simple text similarity score (placeholder for more sophisticated algorithm)
			(similarity(e2.title, $5) * 50) AS score
		FROM episodes e2
		JOIN podcasts p ON e2.podcast_id = p.id
		WHERE 
			-- Either from the same podcast or contains similar words in title
			(e2.podcast_id = $4 OR similarity(e2.title, $5) > 0.2)
			%s
			AND e2.status = 'active'
		ORDER BY score DESC
		LIMIT $2
	`, excludeCondition)
	
	var items []models.RecommendedItem
	
	if len(excludedIDs) > 0 {
		err = r.db.SelectContext(ctx, &items, query, episodeID, limit, excludedIDsParam, sourcePodcastID, sourceTitle)
	} else {
		err = r.db.SelectContext(ctx, &items, query, episodeID, limit, sourcePodcastID, sourceTitle)
	}
	
	return items, err
}

// GetTrendingPodcasts gets trending podcasts
func (r *repository) GetTrendingPodcasts(ctx context.Context, timeRange string, limit int, excludedIDs []uuid.UUID) ([]models.RecommendedItem, error) {
	// Determine the time filter based on time range
	var timeFilter string
	switch timeRange {
	case "daily":
		timeFilter = "AND le.started_at > CURRENT_TIMESTAMP - INTERVAL '1 day'"
	case "weekly":
		timeFilter = "AND le.started_at > CURRENT_TIMESTAMP - INTERVAL '7 days'"
	case "monthly":
		timeFilter = "AND le.started_at > CURRENT_TIMESTAMP - INTERVAL '30 days'"
	default:
		timeFilter = "AND le.started_at > CURRENT_TIMESTAMP - INTERVAL '7 days'"
	}
	
	// Build the exclusion list for the query
	var excludedIDsParam interface{}
	var excludeCondition string
	if len(excludedIDs) > 0 {
		excludedIDsParam = excludedIDs
		excludeCondition = "AND p.id != ANY($3)"
	} else {
		excludedIDsParam = nil
		excludeCondition = ""
	}
	
	// Query trending podcasts based on listen events
	query := fmt.Sprintf(`
		SELECT 
			p.id,
			'podcast' AS type,
			p.title,
			p.description,
			p.cover_image_url AS image_url,
			p.id AS podcast_id,
			p.title AS podcast_title,
			-- Score based on listen count and recency
			COUNT(le.id) * 
			(1.0 + 0.1 * (
				SELECT COUNT(DISTINCT le2.listener_id) 
				FROM listen_events le2 
				JOIN episodes e2 ON le2.episode_id = e2.id 
				WHERE e2.podcast_id = p.id %s
			)) AS score
		FROM listen_events le
		JOIN episodes e ON le.episode_id = e.id
		JOIN podcasts p ON e.podcast_id = p.id
		WHERE 1=1 %s %s
		AND p.status = 'active'
		GROUP BY p.id, p.title, p.description, p.cover_image_url
		ORDER BY score DESC
		LIMIT $2
	`, timeFilter, timeFilter, excludeCondition)
	
	var items []models.RecommendedItem
	var err error
	
	if len(excludedIDs) > 0 {
		err = r.db.SelectContext(ctx, &items, query, limit, excludedIDsParam)
	} else {
		err = r.db.SelectContext(ctx, &items, query, limit)
	}
	
	if err != nil {
		return nil, err
	}
	
	// If there are not enough trending podcasts, supplement with recent podcasts
	if len(items) < limit {
		recentQuery := fmt.Sprintf(`
			SELECT 
				p.id,
				'podcast' AS type,
				p.title,
				p.description,
				p.cover_image_url AS image_url,
				p.id AS podcast_id,
				p.title AS podcast_title,
				-- Score based on recency
				EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - p.created_at)) / 86400 AS score
			FROM podcasts p
			WHERE p.status = 'active' %s
			ORDER BY p.created_at DESC
			LIMIT $1
		`, excludeCondition)
		
		var recentItems []models.RecommendedItem
		
		if len(excludedIDs) > 0 {
			err = r.db.SelectContext(ctx, &recentItems, recentQuery, limit-len(items), excludedIDsParam)
		} else {
			err = r.db.SelectContext(ctx, &recentItems, recentQuery, limit-len(items))
		}
		
		if err == nil {
			// Add recent items, avoiding duplicates
			existingIDs := make(map[uuid.UUID]bool)
			for _, item := range items {
				existingIDs[item.ID] = true
			}
			
			for _, recent := range recentItems {
				if !existingIDs[recent.ID] {
					items = append(items, recent)
					existingIDs[recent.ID] = true
				}
			}
		}
	}
	
	return items, nil
}

// GetPopularInCategory gets popular content in a category
func (r *repository) GetPopularInCategory(ctx context.Context, categoryID uuid.UUID, limit int, excludedIDs []uuid.UUID) ([]models.RecommendedItem, error) {
	// Build the exclusion list for the query
	var excludedIDsParam interface{}
	var excludeCondition string
	if len(excludedIDs) > 0 {
		excludedIDsParam = excludedIDs
		excludeCondition = "AND p.id != ANY($3)"
	} else {
		excludedIDsParam = nil
		excludeCondition = ""
	}
	
	// Query popular podcasts in the given category
	query := fmt.Sprintf(`
		SELECT 
			p.id,
			'podcast' AS type,
			p.title,
			p.description,
			p.cover_image_url AS image_url,
			p.id AS podcast_id,
			p.title AS podcast_title,
			-- Score based on listen count in the last 30 days
			(
				SELECT COUNT(le.id)
				FROM listen_events le
				JOIN episodes e ON le.episode_id = e.id
				WHERE e.podcast_id = p.id
				AND le.started_at > CURRENT_TIMESTAMP - INTERVAL '30 days'
			) AS score
		FROM podcasts p
		JOIN podcast_categories pc ON p.id = pc.podcast_id
		WHERE pc.category_id = $1 %s
		AND p.status = 'active'
		ORDER BY score DESC
		LIMIT $2
	`, excludeCondition)
	
	var items []models.RecommendedItem
	var err error
	
	if len(excludedIDs) > 0 {
		err = r.db.SelectContext(ctx, &items, query, categoryID, limit, excludedIDsParam)
	} else {
		err = r.db.SelectContext(ctx, &items, query, categoryID, limit)
	}
	
	return items, err
}

// UpdateUserPreference updates a user's category preference
func (r *repository) UpdateUserPreference(ctx context.Context, userID uuid.UUID, categoryID uuid.UUID, weight float64) error {
	query := `
		INSERT INTO user_preferences (user_id, category_id, weight, last_updated)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, category_id) 
		DO UPDATE SET weight = $3, last_updated = $4
	`
	
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, categoryID, weight, now)
	return err
}

// GetUserPreferences gets a user's category preferences
func (r *repository) GetUserPreferences(ctx context.Context, userID uuid.UUID) ([]models.UserPreference, error) {
	query := `
		SELECT user_id, category_id, weight, last_updated
		FROM user_preferences
		WHERE user_id = $1
		ORDER BY weight DESC
	`
	
	var preferences []models.UserPreference
	err := r.db.SelectContext(ctx, &preferences, query, userID)
	return preferences, err
}