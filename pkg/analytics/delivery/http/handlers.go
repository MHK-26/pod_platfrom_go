// pkg/analytics/delivery/http/handlers.go
package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-username/podcast-platform/pkg/analytics/models"
	"github.com/your-username/podcast-platform/pkg/analytics/usecase"
	"github.com/your-username/podcast-platform/pkg/common/utils"
)

// Handler struct
type Handler struct {
	usecase usecase.Usecase
}

// NewHandler creates a new analytics handler
func NewHandler(usecase usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

// TrackListen godoc
// @Summary Track a listen event
// @Description Record a podcast listen event
// @Tags analytics
// @Accept json
// @Produce json
// @Param request body models.TrackListenRequest true "Track Listen Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /analytics/track-listen [post]
func (h *Handler) TrackListen(c *gin.Context) {
	var req models.TrackListenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Get authenticated user ID if available
	userID, exists := c.Get("user_id")
	if exists {
		userIDParsed, err := uuid.Parse(userID.(string))
		if err == nil {
			req.ListenerID = userIDParsed
		}
	}

	// Track listen event
	event, err := h.usecase.TrackListen(c.Request.Context(), &req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to track listen event")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "success",
		"listen_id": event.ID,
	})
}

// GetEpisodeAnalytics godoc
// @Summary Get episode analytics
// @Description Get analytics for a specific episode
// @Tags analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param episode_id path string true "Episode ID"
// @Param start_date query string false "Start Date (YYYY-MM-DD)"
// @Param end_date query string false "End Date (YYYY-MM-DD)"
// @Param interval query string false "Interval (day, week, month)"
// @Success 200 {object} models.EpisodeAnalytics
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /analytics/episodes/{episode_id} [get]
func (h *Handler) GetEpisodeAnalytics(c *gin.Context) {
	// Get episode ID from path
	episodeIDStr := c.Param("episode_id")
	episodeID, err := uuid.Parse(episodeIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid episode ID")
		return
	}

	// Parse query parameters
	startDateStr := c.DefaultQuery("start_date", "")
	endDateStr := c.DefaultQuery("end_date", "")
	interval := c.DefaultQuery("interval", "day")

	var startDate, endDate time.Time
	var parseErr error

	if startDateStr != "" {
		startDate, parseErr = time.Parse("2006-01-02", startDateStr)
		if parseErr != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "Invalid start date format")
			return
		}
	} else {
		// Default to 30 days ago
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, parseErr = time.Parse("2006-01-02", endDateStr)
		if parseErr != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "Invalid end date format")
			return
		}
	} else {
		// Default to now
		endDate = time.Now()
	}

	// Prepare analytics params
	params := models.AnalyticsParams{
		StartDate: startDate,
		EndDate:   endDate,
		Interval:  interval,
	}

	// Get episode analytics
	analytics, err := h.usecase.GetEpisodeAnalytics(c.Request.Context(), episodeID, params)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get episode analytics")
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetPodcastAnalytics godoc
// @Summary Get podcast analytics
// @Description Get analytics for a specific podcast
// @Tags analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param podcast_id path string true "Podcast ID"
// @Param start_date query string false "Start Date (YYYY-MM-DD)"
// @Param end_date query string false "End Date (YYYY-MM-DD)"
// @Param interval query string false "Interval (day, week, month)"
// @Success 200 {object} models.PodcastAnalytics
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /analytics/podcasts/{podcast_id} [get]
func (h *Handler) GetPodcastAnalytics(c *gin.Context) {
	// Get podcast ID from path
	podcastIDStr := c.Param("podcast_id")
	podcastID, err := uuid.Parse(podcastIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid podcast ID")
		return
	}

	// Parse query parameters
	startDateStr := c.DefaultQuery("start_date", "")
	endDateStr := c.DefaultQuery("end_date", "")
	interval := c.DefaultQuery("interval", "day")

	var startDate, endDate time.Time
	var parseErr error

	if startDateStr != "" {
		startDate, parseErr = time.Parse("2006-01-02", startDateStr)
		if parseErr != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "Invalid start date format")
			return
		}
	} else {
		// Default to 30 days ago
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, parseErr = time.Parse("2006-01-02", endDateStr)
		if parseErr != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "Invalid end date format")
			return
		}
	} else {
		// Default to now
		endDate = time.Now()
	}

	// Prepare analytics params
	params := models.AnalyticsParams{
		StartDate: startDate,
		EndDate:   endDate,
		Interval:  interval,
	}

	// Get podcast analytics
	analytics, err := h.usecase.GetPodcastAnalytics(c.Request.Context(), podcastID, params)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get podcast analytics")
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetPodcasterAnalytics godoc
// @Summary Get podcaster analytics
// @Description Get analytics for a podcaster
// @Tags analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start Date (YYYY-MM-DD)"
// @Param end_date query string false "End Date (YYYY-MM-DD)"
// @Success 200 {object} models.PodcasterAnalytics
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /analytics/podcaster [get]
func (h *Handler) GetPodcasterAnalytics(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if user is a podcaster
	userType, exists := c.Get("user_type")
	if !exists || userType.(string) != "podcaster" {
		utils.RespondWithError(c, http.StatusForbidden, "Only podcasters can access this information")
		return
	}

	userIDParsed, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	// Parse query parameters
	startDateStr := c.DefaultQuery("start_date", "")
	endDateStr := c.DefaultQuery("end_date", "")

	var startDate, endDate time.Time
	var parseErr error

	if startDateStr != "" {
		startDate, parseErr = time.Parse("2006-01-02", startDateStr)
		if parseErr != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "Invalid start date format")
			return
		}
	} else {
		// Default to 30 days ago
		startDate = time.Now().AddDate(0, 0, -30)
	}

	if endDateStr != "" {
		endDate, parseErr = time.Parse("2006-01-02", endDateStr)
		if parseErr != nil {
			utils.RespondWithError(c, http.StatusBadRequest, "Invalid end date format")
			return
		}
	} else {
		// Default to now
		endDate = time.Now()
	}

	// Prepare analytics params
	params := models.AnalyticsParams{
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Get podcaster analytics
	analytics, err := h.usecase.GetPodcasterAnalytics(c.Request.Context(), userIDParsed, params)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get podcaster analytics")
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetListeningHistory godoc
// @Summary Get listening history
// @Description Get listening history for the authenticated user
// @Tags analytics
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20)"
// @Success 200 {object} utils.PaginationResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /analytics/history [get]
func (h *Handler) GetListeningHistory(c *gin.Context) {
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

	// Get pagination parameters
	params := utils.GetPaginationParams(c)

	// Get listening history
	history, totalCount, err := h.usecase.GetListeningHistory(c.Request.Context(), userIDParsed, params.Page, params.PageSize)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get listening history")
		return
	}

	utils.RespondWithPagination(c, history, totalCount, params.Page, params.PageSize)
}

// RegisterRoutes registers all the analytics routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	analytics := router.Group("/analytics")
	{
		// Public routes
		analytics.POST("/track-listen", h.TrackListen)

		// Protected routes
		protected := analytics.Group("")
		protected.Use(authMiddleware)
		{
			protected.GET("/episodes/:episode_id", h.GetEpisodeAnalytics)
			protected.GET("/podcasts/:podcast_id", h.GetPodcastAnalytics)
			protected.GET("/podcaster", h.GetPodcasterAnalytics)
			protected.GET("/history", h.GetListeningHistory)
		}
	}
}