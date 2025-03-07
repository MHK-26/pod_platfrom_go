// pkg/recommendation/delivery/http/handlers.go
package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/MHK-26/pod_platfrom_go/pkg/common/utils"
	"github.com/MHK-26/pod_platfrom_go/pkg/recommendation/models"
	"github.com/MHK-26/pod_platfrom_go/pkg/recommendation/usecase"
)

// Handler is the HTTP handler for the recommendation service
type Handler struct {
	usecase usecase.Usecase
}

// NewHandler creates a new recommendation handler
func NewHandler(usecase usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

// GetPersonalizedRecommendations godoc
// @Summary Get personalized recommendations
// @Description Get podcast and episode recommendations based on user's history
// @Tags recommendations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of recommendations to return (default 10, max 50)"
// @Param excluded_ids query []string false "IDs to exclude from recommendations"
// @Success 200 {object} models.RecommendationResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /recommendations/personalized [get]
func (h *Handler) GetPersonalizedRecommendations(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	userIDParsed, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	// Parse query parameters
	limit := utils.GetIntQueryParam(c, "limit", 10)
	excludedIDsStr := c.QueryArray("excluded_ids")
	
	// Convert excluded IDs from strings to UUIDs
	var excludedIDs []uuid.UUID
	for _, idStr := range excludedIDsStr {
		id, err := uuid.Parse(idStr)
		if err == nil { // Skip invalid UUIDs
			excludedIDs = append(excludedIDs, id)
		}
	}

	// Prepare request
	req := &models.RecommendationRequest{
		UserID:      userIDParsed,
		Limit:       limit,
		ExcludedIDs: excludedIDs,
	}

	// Get recommendations
	response, err := h.usecase.GetPersonalizedRecommendations(c.Request.Context(), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get recommendations")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetSimilarPodcasts godoc
// @Summary Get similar podcasts
// @Description Get podcasts similar to the specified podcast
// @Tags recommendations
// @Accept json
// @Produce json
// @Param podcast_id path string true "Podcast ID"
// @Param limit query int false "Number of recommendations to return (default 10, max 50)"
// @Param excluded_ids query []string false "IDs to exclude from recommendations"
// @Success 200 {object} models.RecommendationResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /recommendations/similar/podcasts/{podcast_id} [get]
func (h *Handler) GetSimilarPodcasts(c *gin.Context) {
	// Get podcast ID from path
	podcastIDStr, ok := utils.ExtractIDParam(c, "podcast_id")
	if !ok {
		return
	}

	podcastID, err := uuid.Parse(podcastIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid podcast ID")
		return
	}

	// Parse query parameters
	limit := utils.GetIntQueryParam(c, "limit", 10)
	excludedIDsStr := c.QueryArray("excluded_ids")
	
	// Convert excluded IDs from strings to UUIDs
	var excludedIDs []uuid.UUID
	for _, idStr := range excludedIDsStr {
		id, err := uuid.Parse(idStr)
		if err == nil { // Skip invalid UUIDs
			excludedIDs = append(excludedIDs, id)
		}
	}

	// Prepare request
	req := &models.SimilarContentRequest{
		ContentID:   podcastID,
		ContentType: "podcast",
		Limit:       limit,
		ExcludedIDs: excludedIDs,
	}

	// Get similar podcasts
	response, err := h.usecase.GetSimilarPodcasts(c.Request.Context(), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get similar podcasts")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetSimilarEpisodes godoc
// @Summary Get similar episodes
// @Description Get episodes similar to the specified episode
// @Tags recommendations
// @Accept json
// @Produce json
// @Param episode_id path string true "Episode ID"
// @Param limit query int false "Number of recommendations to return (default 10, max 50)"
// @Param excluded_ids query []string false "IDs to exclude from recommendations"
// @Success 200 {object} models.RecommendationResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /recommendations/similar/episodes/{episode_id} [get]
func (h *Handler) GetSimilarEpisodes(c *gin.Context) {
	// Get episode ID from path
	episodeIDStr, ok := utils.ExtractIDParam(c, "episode_id")
	if !ok {
		return
	}

	episodeID, err := uuid.Parse(episodeIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid episode ID")
		return
	}

	// Parse query parameters
	limit := utils.GetIntQueryParam(c, "limit", 10)
	excludedIDsStr := c.QueryArray("excluded_ids")
	
	// Convert excluded IDs from strings to UUIDs
	var excludedIDs []uuid.UUID
	for _, idStr := range excludedIDsStr {
		id, err := uuid.Parse(idStr)
		if err == nil { // Skip invalid UUIDs
			excludedIDs = append(excludedIDs, id)
		}
	}

	// Prepare request
	req := &models.SimilarContentRequest{
		ContentID:   episodeID,
		ContentType: "episode",
		Limit:       limit,
		ExcludedIDs: excludedIDs,
	}

	// Get similar episodes
	response, err := h.usecase.GetSimilarEpisodes(c.Request.Context(), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get similar episodes")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetTrendingPodcasts godoc
// @Summary Get trending podcasts
// @Description Get trending podcasts for a specific time range
// @Tags recommendations
// @Accept json
// @Produce json
// @Param time_range query string false "Time range (daily, weekly, monthly) (default: weekly)"
// @Param limit query int false "Number of podcasts to return (default 10, max 50)"
// @Param excluded_ids query []string false "IDs to exclude from recommendations"
// @Success 200 {object} models.RecommendationResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /recommendations/trending [get]
func (h *Handler) GetTrendingPodcasts(c *gin.Context) {
	// Parse query parameters
	timeRange := c.DefaultQuery("time_range", "weekly")
	if timeRange != "daily" && timeRange != "weekly" && timeRange != "monthly" {
		timeRange = "weekly"
	}

	limit := utils.GetIntQueryParam(c, "limit", 10)
	excludedIDsStr := c.QueryArray("excluded_ids")
	
	// Convert excluded IDs from strings to UUIDs
	var excludedIDs []uuid.UUID
	for _, idStr := range excludedIDsStr {
		id, err := uuid.Parse(idStr)
		if err == nil { // Skip invalid UUIDs
			excludedIDs = append(excludedIDs, id)
		}
	}

	// Prepare request
	req := &models.TrendingRequest{
		TimeRange:   timeRange,
		Limit:       limit,
		ExcludedIDs: excludedIDs,
	}

	// Get trending podcasts
	response, err := h.usecase.GetTrendingPodcasts(c.Request.Context(), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get trending podcasts")
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPopularInCategory godoc
// @Summary Get popular podcasts in category
// @Description Get popular podcasts for a specific category
// @Tags recommendations
// @Accept json
// @Produce json
// @Param category_id path string true "Category ID"
// @Param limit query int false "Number of podcasts to return (default 10, max 50)"
// @Param excluded_ids query []string false "IDs to exclude from recommendations"
// @Success 200 {object} models.RecommendationResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /recommendations/categories/{category_id}/popular [get]
func (h *Handler) GetPopularInCategory(c *gin.Context) {
	// Get category ID from path
	categoryIDStr, ok := utils.ExtractIDParam(c, "category_id")
	if !ok {
		return
	}

	categoryID, err := uuid.Parse(categoryIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	// Parse query parameters
	limit := utils.GetIntQueryParam(c, "limit", 10)
	excludedIDsStr := c.QueryArray("excluded_ids")
	
	// Convert excluded IDs from strings to UUIDs
	var excludedIDs []uuid.UUID
	for _, idStr := range excludedIDsStr {
		id, err := uuid.Parse(idStr)
		if err == nil { // Skip invalid UUIDs
			excludedIDs = append(excludedIDs, id)
		}
	}

	// Prepare request
	req := &models.CategoryPopularRequest{
		CategoryID:  categoryID,
		Limit:       limit,
		ExcludedIDs: excludedIDs,
	}

	// Get popular podcasts in category
	response, err := h.usecase.GetPopularInCategory(c.Request.Context(), req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get popular podcasts in category")
		return
	}

	c.JSON(http.StatusOK, response)
}

// RegisterRoutes registers all the recommendation routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	recommendations := router.Group("/recommendations")
	{
		// Public routes
		recommendations.GET("/similar/podcasts/:podcast_id", h.GetSimilarPodcasts)
		recommendations.GET("/similar/episodes/:episode_id", h.GetSimilarEpisodes)
		recommendations.GET("/trending", h.GetTrendingPodcasts)
		recommendations.GET("/categories/:category_id/popular", h.GetPopularInCategory)
		
		// Protected routes
		protected := recommendations.Group("")
		protected.Use(authMiddleware)
		{
			protected.GET("/personalized", h.GetPersonalizedRecommendations)
		}
	}
}