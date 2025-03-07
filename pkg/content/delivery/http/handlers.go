// pkg/content/delivery/http/handlers.go
package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-username/podcast-platform/pkg/content/models"
	"github.com/your-username/podcast-platform/pkg/content/usecase"
	"github.com/your-username/podcast-platform/pkg/common/utils"
)

// Handler is the HTTP handler for the content service
type Handler struct {
	usecase usecase.Usecase
}

// NewHandler creates a new content handler
func NewHandler(usecase usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

// GetPodcast godoc
// @Summary Get podcast details
// @Description Get detailed information about a podcast
// @Tags podcasts
// @Accept json
// @Produce json
// @Param id path string true "Podcast ID"
// @Success 200 {object} models.PodcastResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /podcasts/{id} [get]
func (h *Handler) GetPodcast(c *gin.Context) {
	idStr, ok := utils.ExtractIDParam(c, "id")
	if !ok {
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid podcast ID")
		return
	}

	podcast, err := h.usecase.GetPodcastByID(c.Request.Context(), id)
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Podcast not found")
		return
	}

	utils.RespondWithSuccess(c, podcast)
}

// ListPodcasts godoc
// @Summary List podcasts
// @Description Get a paginated list of podcasts with optional filtering
// @Tags podcasts
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20)"
// @Param query query string false "Search query"
// @Param category query string false "Category ID"
// @Param sort_by query string false "Sort field (created_at, title, listens)"
// @Param sort_order query string false "Sort order (asc, desc)"
// @Success 200 {object} utils.PaginatedResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /podcasts [get]
func (h *Handler) ListPodcasts(c *gin.Context) {
	params := models.PodcastSearchParams{
		Query:      c.Query("query"),
		Category:   c.Query("category"),
		Language:   c.Query("language"),
		SortBy:     c.Query("sort_by"),
		SortOrder:  c.Query("sort_order"),
		Page:       utils.GetIntQueryParam(c, "page", 1),
		PageSize:   utils.GetIntQueryParam(c, "page_size", 20),
	}

	podcasts, totalCount, err := h.usecase.ListPodcasts(c.Request.Context(), params)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch podcasts")
		return
	}

	utils.RespondWithPagination(c, podcasts, totalCount, params.Page, params.PageSize)
}

// GetPodcastsByUser godoc
// @Summary Get user's podcasts
// @Description Get podcasts created by a specific user
// @Tags podcasts
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20)"
// @Success 200 {object} utils.PaginatedResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /users/{user_id}/podcasts [get]
func (h *Handler) GetPodcastsByUser(c *gin.Context) {
	userIDStr, ok := utils.ExtractIDParam(c, "user_id")
	if !ok {
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	page := utils.GetIntQueryParam(c, "page", 1)
	pageSize := utils.GetIntQueryParam(c, "page_size", 20)

	podcasts, totalCount, err := h.usecase.GetPodcastsByPodcasterID(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch podcasts")
		return
	}

	utils.RespondWithPagination(c, podcasts, totalCount, page, pageSize)
}

// CreatePodcast godoc
// @Summary Create a podcast
// @Description Create a new podcast from RSS feed
// @Tags podcasts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreatePodcastRequest true "Create Podcast Request"
// @Success 201 {object} models.Podcast
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /podcasts [post]
func (h *Handler) CreatePodcast(c *gin.Context) {
	var req models.CreatePodcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

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

	// Check if user is a podcaster
	userType, exists := c.Get("user_type")
	if !exists || userType.(string) != "podcaster" {
		utils.RespondWithError(c, http.StatusForbidden, "Only podcasters can create podcasts")
		return
	}

	// First parse the RSS feed to get podcast details
	feed, err := h.usecase.ParseRSSFeed(c.Request.Context(), req.RSSUrl)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Failed to parse RSS feed: "+err.Error())
		return
	}

	// Prepare podcast data from RSS feed
	podcast, err := h.usecase.CreatePodcast(c.Request.Context(), userIDParsed, &req, feed)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to create podcast: "+err.Error())
		return
	}

	// Trigger RSS feed sync in the background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		h.usecase.SyncPodcastFromRSS(ctx, podcast.ID)
	}()

	utils.RespondWithCreated(c, podcast)
}

// UpdatePodcast godoc
// @Summary Update a podcast
// @Description Update an existing podcast
// @Tags podcasts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Podcast ID"
// @Param request body models.UpdatePodcastRequest true "Update Podcast Request"
// @Success 200 {object} models.Podcast
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /podcasts/{id} [put]
func (h *Handler) UpdatePodcast(c *gin.Context) {
	idStr, ok := utils.ExtractIDParam(c, "id")
	if !ok {
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid podcast ID")
		return
	}

	var req models.UpdatePodcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

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

	// Check if RSS URL changed
	var needsSync bool
	if req.RSSUrl != "" {
		// Get current podcast to check if URL changed
		currentPodcast, err := h.usecase.GetPodcastByID(c.Request.Context(), id)
		if err != nil {
			utils.RespondWithError(c, http.StatusNotFound, "Podcast not found")
			return
		}
		
		if currentPodcast.Podcast.RSSUrl != req.RSSUrl {
			needsSync = true
			
			// Validate and parse the new RSS feed
			_, err := h.usecase.ParseRSSFeed(c.Request.Context(), req.RSSUrl)
			if err != nil {
				utils.RespondWithError(c, http.StatusBadRequest, "Failed to parse RSS feed: "+err.Error())
				return
			}
		}
	}

	podcast, err := h.usecase.UpdatePodcast(c.Request.Context(), id, userIDParsed, &req)
	if err != nil {
		if err.Error() == "podcast not found" {
			utils.RespondWithError(c, http.StatusNotFound, "Podcast not found")
			return
		}
		if err.Error() == "not authorized" {
			utils.RespondWithError(c, http.StatusForbidden, "Not authorized to update this podcast")
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update podcast")
		return
	}

	// If RSS URL was changed, trigger a sync in the background
	if needsSync {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()
			h.usecase.SyncPodcastFromRSS(ctx, podcast.ID)
		}()
	}

	utils.RespondWithSuccess(c, podcast)
}

// SyncPodcast godoc
// @Summary Synchronize podcast from RSS feed
// @Description Manually trigger a synchronization of podcast content from its RSS feed
// @Tags podcasts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Podcast ID"
// @Success 202 {object} utils.Message
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /podcasts/{id}/sync [post]
func (h *Handler) SyncPodcast(c *gin.Context) {
	idStr, ok := utils.ExtractIDParam(c, "id")
	if !ok {
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid podcast ID")
		return
	}

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

	// Check if user is authorized to sync this podcast
	isAuthorized, err := h.usecase.IsUserAuthorizedForPodcast(c.Request.Context(), id, userIDParsed)
	if err != nil {
		if err.Error() == "podcast not found" {
			utils.RespondWithError(c, http.StatusNotFound, "Podcast not found")
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to check authorization")
		return
	}

	if !isAuthorized {
		utils.RespondWithError(c, http.StatusForbidden, "Not authorized to sync this podcast")
		return
	}

	// Trigger the sync in the background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		h.usecase.SyncPodcastFromRSS(ctx, id)
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Podcast synchronization started",
	})
}

// DeletePodcast godoc
// @Summary Delete a podcast
// @Description Delete an existing podcast
// @Tags podcasts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Podcast ID"
// @Success 204 "No Content"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /podcasts/{id} [delete]
func (h *Handler) DeletePodcast(c *gin.Context) {
	idStr, ok := utils.ExtractIDParam(c, "id")
	if !ok {
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid podcast ID")
		return
	}

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

	err = h.usecase.DeletePodcast(c.Request.Context(), id, userIDParsed)
	if err != nil {
		if err.Error() == "podcast not found" {
			utils.RespondWithError(c, http.StatusNotFound, "Podcast not found")
			return
		}
		if err.Error() == "not authorized" {
			utils.RespondWithError(c, http.StatusForbidden, "Not authorized to delete this podcast")
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to delete podcast")
		return
	}

	utils.RespondWithNoContent(c)
}

// GetEpisode godoc
// @Summary Get episode details
// @Description Get detailed information about an episode
// @Tags episodes
// @Accept json
// @Produce json
// @Param id path string true "Episode ID"
// @Success 200 {object} models.EpisodeResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /episodes/{id} [get]
func (h *Handler) GetEpisode(c *gin.Context) {
	idStr, ok := utils.ExtractIDParam(c, "id")
	if !ok {
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid episode ID")
		return
	}

	episode, err := h.usecase.GetEpisodeByID(c.Request.Context(), id)
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "Episode not found")
		return
	}

	utils.RespondWithSuccess(c, episode)
}

// GetEpisodesByPodcast godoc
// @Summary Get podcast episodes
// @Description Get episodes for a specific podcast
// @Tags episodes
// @Accept json
// @Produce json
// @Param podcast_id path string true "Podcast ID"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 20)"
// @Success 200 {object} utils.PaginatedResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /podcasts/{podcast_id}/episodes [get]
func (h *Handler) GetEpisodesByPodcast(c *gin.Context) {
	podcastIDStr, ok := utils.ExtractIDParam(c, "podcast_id")
	if !ok {
		return
	}

	podcastID, err := uuid.Parse(podcastIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid podcast ID")
		return
	}

	page := utils.GetIntQueryParam(c, "page", 1)
	pageSize := utils.GetIntQueryParam(c, "page_size", 20)

	episodes, totalCount, err := h.usecase.GetEpisodesByPodcastID(c.Request.Context(), podcastID, page, pageSize)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch episodes")
		return
	}

	utils.RespondWithPagination(c, episodes, totalCount, page, pageSize)
}

// ListCategories godoc
// @Summary List categories
// @Description Get a list of podcast categories
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} models.Category
// @Failure 500 {object} utils.ErrorResponse
// @Router /categories [get]
func (h *Handler) ListCategories(c *gin.Context) {
	categories, err := h.usecase.GetCategories(c.Request.Context())
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	utils.RespondWithSuccess(c, categories)
}

// Subscribe godoc
// @Summary Subscribe to podcast
// @Description Subscribe to a podcast
// @Tags subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param podcast_id path string true "Podcast ID"
// @Success 204 "No Content"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /podcasts/{podcast_id}/subscribe [post]
func (h *Handler) Subscribe(c *gin.Context) {
	podcastIDStr, ok := utils.ExtractIDParam(c, "podcast_id")
	if !ok {
		return
	}

	podcastID, err := uuid.Parse(podcastIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid podcast ID")
		return
	}

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

	err = h.usecase.SubscribeToPodcast(c.Request.Context(), userIDParsed, podcastID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to subscribe")
		return
	}

	utils.RespondWithNoContent(c)
}

// Unsubscribe godoc
// @Summary Unsubscribe from podcast
// @Description Unsubscribe from a podcast
// @Tags subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param podcast_id path string true "Podcast ID"
// @Success 204 "No Content"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /podcasts/{podcast_id}/unsubscribe [post]
func (h *Handler) Unsubscribe(c *gin.Context) {
	podcastIDStr, ok := utils.ExtractIDParam(c, "podcast_id")
	if !ok {
		return
	}

	podcastID, err := uuid.Parse(podcastIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid podcast ID")
		return
	}

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

	err = h.usecase.UnsubscribeFromPodcast(c.Request.Context(), userIDParsed, podcastID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to unsubscribe")
		return
	}

	utils.RespondWithNoContent(c)
}

// SavePlaybackPosition godoc
// @Summary Save playback position
// @Description Save the current playback position for an episode
// @Tags episodes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.SavePlaybackPositionRequest true "Save Playback Position Request"
// @Success 204 "No Content"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /episodes/playback [post]
func (h *Handler) SavePlaybackPosition(c *gin.Context) {
	var req models.SavePlaybackPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

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

	err = h.usecase.SavePlaybackPosition(c.Request.Context(), userIDParsed, req.EpisodeID, req.Position, req.Completed)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to save playback position")
		return
	}

	utils.RespondWithNoContent(c)
}

// GetSyncStatus godoc
// @Summary Get RSS feed sync status
// @Description Get the status of RSS feed synchronization for a podcast
// @Tags podcasts
// @Accept json
// @Produce json
// @Param podcast_id path string true "Podcast ID"
// @Success 200 {object} models.RSSFeedSyncLog
// @Failure 400 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /podcasts/{podcast_id}/sync-status [get]
func (h *Handler) GetSyncStatus(c *gin.Context) {
	podcastIDStr, ok := utils.ExtractIDParam(c, "podcast_id")
	if !ok {
		return
	}

	podcastID, err := uuid.Parse(podcastIDStr)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid podcast ID")
		return
	}

	syncLog, err := h.usecase.GetLatestSyncLog(c.Request.Context(), podcastID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to get sync status")
		return
	}

	if syncLog == nil {
		utils.RespondWithError(c, http.StatusNotFound, "No sync logs found for this podcast")
		return
	}

	utils.RespondWithSuccess(c, syncLog)
}

// RegisterRoutes registers all the content routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	// Public routes
	podcasts := router.Group("/podcasts")
	{
		podcasts.GET("", h.ListPodcasts)
		podcasts.GET("/:id", h.GetPodcast)
		podcasts.GET("/:podcast_id/episodes", h.GetEpisodesByPodcast)
	}

	episodes := router.Group("/episodes")
	{
		episodes.GET("/:id", h.GetEpisode)
	}

	router.GET("/categories", h.ListCategories)
	router.GET("/users/:user_id/podcasts", h.GetPodcastsByUser)

	// Protected routes
	protected := router.Group("")
	protected.Use(authMiddleware)
	{
		protected.POST("/podcasts", h.CreatePodcast)
		protected.PUT("/podcasts/:id", h.UpdatePodcast)
		protected.DELETE("/podcasts/:id", h.DeletePodcast)
		protected.POST("/podcasts/:id/sync", h.SyncPodcast)
		
		protected.POST("/podcasts/:podcast_id/subscribe", h.Subscribe)
		protected.POST("/podcasts/:podcast_id/unsubscribe", h.Unsubscribe)
		
		protected.POST("/episodes/playback", h.SavePlaybackPosition)
	}
}