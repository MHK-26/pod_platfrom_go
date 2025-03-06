// pkg/auth/delivery/http/handlers.go
package http

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-username/podcast-platform/pkg/auth/models"
	"github.com/your-username/podcast-platform/pkg/auth/usecase"
	"github.com/your-username/podcast-platform/pkg/common/utils"
)

// Handler struct
type Handler struct {
	usecase usecase.Usecase
}

// NewHandler creates a new auth handler
func NewHandler(usecase usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Register Request"
// @Success 201 {object} models.User
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := h.usecase.Register(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			utils.RespondWithError(c, http.StatusBadRequest, err.Error())
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to register user")
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login godoc
// @Summary Login user
// @Description Login user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login Request"
// @Success 200 {object} models.TokenResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	tokenResponse, err := h.usecase.Login(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid credentials") {
			utils.RespondWithError(c, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to login")
		return
	}

	c.JSON(http.StatusOK, tokenResponse)
}

// SocialLogin godoc
// @Summary Login with social provider
// @Description Login with Google or Apple
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.SocialLoginRequest true "Social Login Request"
// @Success 200 {object} models.TokenResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/social-login [post]
func (h *Handler) SocialLogin(c *gin.Context) {
	var req models.SocialLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	tokenResponse, err := h.usecase.SocialLogin(c.Request.Context(), &req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to login with social provider")
		return
	}

	c.JSON(http.StatusOK, tokenResponse)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} models.TokenResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/refresh-token [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	tokenResponse, err := h.usecase.RefreshToken(c.Request.Context(), &req)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "Invalid or expired refresh token")
		return
	}

	c.JSON(http.StatusOK, tokenResponse)
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get authenticated user profile
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 401 {object} utils.ErrorResponse
// @Failure 404 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/profile [get]
func (h *Handler) GetProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	uuid, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	user, err := h.usecase.GetUserByID(c.Request.Context(), uuid)
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "User not found")
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update authenticated user profile
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UpdateProfileRequest true "Update Profile Request"
// @Success 200 {object} models.User
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/profile [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	uuid, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	user, err := h.usecase.UpdateProfile(c.Request.Context(), uuid, &req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	c.JSON(http.StatusOK, user)
}

// ChangePassword godoc
// @Summary Change password
// @Description Change authenticated user password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.ChangePasswordRequest true "Change Password Request"
// @Success 204 "No Content"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/change-password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	uuid, err := uuid.Parse(userID.(string))
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Invalid user ID")
		return
	}

	var req models.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err = h.usecase.ChangePassword(c.Request.Context(), uuid, &req)
	if err != nil {
		if strings.Contains(err.Error(), "incorrect old password") {
			utils.RespondWithError(c, http.StatusBadRequest, "Incorrect old password")
			return
		}
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to change password")
		return
	}

	c.Status(http.StatusNoContent)
}

// ForgotPassword godoc
// @Summary Forgot password
// @Description Request password reset email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.ForgotPasswordRequest true "Forgot Password Request"
// @Success 204 "No Content"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.usecase.ForgotPassword(c.Request.Context(), &req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to process request")
		return
	}

	// Always return success for security reasons, even if email doesn't exist
	c.Status(http.StatusNoContent)
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset password with token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.ResetPasswordRequest true "Reset Password Request"
// @Success 204 "No Content"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.usecase.ResetPassword(c.Request.Context(), &req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to reset password")
		return
	}

	c.Status(http.StatusNoContent)
}

// VerifyEmail godoc
// @Summary Verify email
// @Description Verify email with token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.VerifyEmailRequest true "Verify Email Request"
// @Success 204 "No Content"
// @Failure 400 {object} utils.ErrorResponse
// @Failure 500 {object} utils.ErrorResponse
// @Router /auth/verify-email [post]
func (h *Handler) VerifyEmail(c *gin.Context) {
	var req models.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "Invalid request payload")
		return
	}

	err := h.usecase.VerifyEmail(c.Request.Context(), &req)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "Failed to verify email")
		return
	}

	c.Status(http.StatusNoContent)
}

// RegisterRoutes registers all the auth routes
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	auth := router.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/social-login", h.SocialLogin)
		auth.POST("/refresh-token", h.RefreshToken)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/reset-password", h.ResetPassword)
		auth.POST("/verify-email", h.VerifyEmail)

		// Protected routes
		protected := auth.Group("")
		protected.Use(authMiddleware)
		{
			protected.GET("/profile", h.GetProfile)
			protected.PUT("/profile", h.UpdateProfile)
			protected.POST("/change-password", h.ChangePassword)
		}
	}
}