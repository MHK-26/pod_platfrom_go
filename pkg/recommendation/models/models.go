// pkg/recommendation/models/models.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// RecommendedItem represents a recommended podcast or episode
type RecommendedItem struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Type        string    `json:"type" db:"type"` // podcast or episode
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	ImageURL    string    `json:"image_url" db:"image_url"`
	PodcastID   uuid.UUID `json:"podcast_id,omitempty" db:"podcast_id"`
	PodcastTitle string   `json:"podcast_title,omitempty" db:"podcast_title"`
	Score       float64   `json:"score" db:"score"`
}

// UserPreference represents a user's content preference
type UserPreference struct {
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	CategoryID  uuid.UUID `json:"category_id" db:"category_id"`
	Weight      float64   `json:"weight" db:"weight"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
}

// SimilarityScore represents similarity between content items
type SimilarityScore struct {
	ItemID1    uuid.UUID `json:"item_id1" db:"item_id1"`
	ItemID2    uuid.UUID `json:"item_id2" db:"item_id2"`
	ItemType   string    `json:"item_type" db:"item_type"` // podcast or episode
	Score      float64   `json:"score" db:"score"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
}

// TrendingItem represents a trending podcast or episode
type TrendingItem struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Type        string    `json:"type" db:"type"` // podcast or episode
	Score       float64   `json:"score" db:"score"`
	TimeRange   string    `json:"time_range" db:"time_range"` // daily, weekly, monthly
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
}

// RecommendationRequest represents a request for recommendations
type RecommendationRequest struct {
	UserID      uuid.UUID   `json:"user_id" validate:"required"`
	Limit       int         `json:"limit" validate:"min=1,max=50"`
	ExcludedIDs []uuid.UUID `json:"excluded_ids"`
}

// SimilarContentRequest represents a request for similar content
type SimilarContentRequest struct {
	ContentID   uuid.UUID   `json:"content_id" validate:"required"`
	ContentType string      `json:"content_type" validate:"required,oneof=podcast episode"`
	Limit       int         `json:"limit" validate:"min=1,max=50"`
	ExcludedIDs []uuid.UUID `json:"excluded_ids"`
}

// TrendingRequest represents a request for trending content
type TrendingRequest struct {
	TimeRange   string      `json:"time_range" validate:"required,oneof=daily weekly monthly"`
	Limit       int         `json:"limit" validate:"min=1,max=50"`
	ExcludedIDs []uuid.UUID `json:"excluded_ids"`
}

// CategoryPopularRequest represents a request for popular content in a category
type CategoryPopularRequest struct {
	CategoryID  uuid.UUID   `json:"category_id" validate:"required"`
	Limit       int         `json:"limit" validate:"min=1,max=50"`
	ExcludedIDs []uuid.UUID `json:"excluded_ids"`
}

// RecommendationResponse represents a response with recommended items
type RecommendationResponse struct {
	Items []RecommendedItem `json:"items"`
}