// pkg/auth/usecase/usecase.go
package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/your-username/podcast-platform/pkg/auth/models"
	"github.com/your-username/podcast-platform/pkg/auth/repository/postgres"
	"github.com/your-username/podcast-platform/pkg/common/config"
	"golang.org/x/crypto/bcrypt"
)

// Usecase defines the methods for the auth usecase
type Usecase interface {
	Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error)
	Login(ctx context.Context, req *models.LoginRequest) (*models.TokenResponse, error)
	SocialLogin(ctx context.Context, req *models.SocialLoginRequest) (*models.TokenResponse, error)
	RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (*models.TokenResponse, error)
	VerifyToken(ctx context.Context, token string) (*models.IDTokenPayload, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req *models.ChangePasswordRequest) error
	ForgotPassword(ctx context.Context, req *models.ForgotPasswordRequest) error
	ResetPassword(ctx context.Context, req *models.ResetPasswordRequest) error
	VerifyEmail(ctx context.Context, req *models.VerifyEmailRequest) error
	UpdateProfile(ctx context.Context, userID uuid.UUID, req *models.UpdateProfileRequest) (*models.User, error)
}

type usecase struct {
	repo           postgres.Repository
	cfg            *config.Config
	contextTimeout time.Duration
}

// NewUsecase creates a new auth usecase
func NewUsecase(repo postgres.Repository, cfg *config.Config, timeout time.Duration) Usecase {
	return &usecase{
		repo:           repo,
		cfg:            cfg,
		contextTimeout: timeout,
	}
}

// Register registers a new user
func (u *usecase) Register(ctx context.Context, req *models.RegisterRequest) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Check if email already exists
	existingUser, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// Check if username already exists
	existingUser, err = u.repo.GetUserByUsername(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		Email:            req.Email,
		Username:         req.Username,
		PasswordHash:     string(passwordHash),
		FullName:         req.FullName,
		UserType:         req.UserType,
		AuthProvider:     "email",
		IsVerified:       false,
		PreferredLanguage: "ar-sd", // Default to Sudanese Arabic
	}

	if err := u.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	// TODO: Send verification email

	return user, nil
}

// Login logs in a user
func (u *usecase) Login(ctx context.Context, req *models.LoginRequest) (*models.TokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Get user by email
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is using email as auth provider
	if user.AuthProvider != "email" {
		return nil, fmt.Errorf("please login with your %s account", user.AuthProvider)
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Update last login
	if err := u.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, err
	}

	// Generate tokens
	tokenResponse, err := u.generateTokens(user)
	if err != nil {
		return nil, err
	}

	return tokenResponse, nil
}

// SocialLogin performs a social login
func (u *usecase) SocialLogin(ctx context.Context, req *models.SocialLoginRequest) (*models.TokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// TODO: Verify token with Google/Apple

	// For now, we'll mock this with a simple token validation
	var email, providerID, fullName string

	// This would normally be extracted from the verified token
	// In a real implementation, you would use the provider's SDK to verify the token
	if req.Provider == "google" {
		// Mock Google verification
		email = "user@example.com"
		providerID = "google-user-123"
		fullName = "Google User"
	} else if req.Provider == "apple" {
		// Mock Apple verification
		email = "user@example.com"
		providerID = "apple-user-123"
		fullName = "Apple User"
	} else {
		return nil, errors.New("unsupported provider")
	}

	// Check if user exists
	user, err := u.repo.GetUserByAuthProvider(ctx, req.Provider, providerID)
	if err != nil {
		// User doesn't exist, create new user
		username := fmt.Sprintf("%s-%s", req.Provider, uuid.New().String()[:8])

		user = &models.User{
			Email:            email,
			Username:         username,
			FullName:         fullName,
			UserType:         "listener", // Default to listener for social logins
			AuthProvider:     req.Provider,
			AuthProviderID:   providerID,
			IsVerified:       true, // Social logins are automatically verified
			PreferredLanguage: "ar-sd",
		}

		if err := u.repo.CreateUser(ctx, user); err != nil {
			return nil, err
		}
	}

	// Update last login
	if err := u.repo.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, err
	}

	// Generate tokens
	tokenResponse, err := u.generateTokens(user)
	if err != nil {
		return nil, err
	}

	return tokenResponse, nil
}

// RefreshToken refreshes an access token
func (u *usecase) RefreshToken(ctx context.Context, req *models.RefreshTokenRequest) (*models.TokenResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Verify refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(u.cfg.JWT.RefreshSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Extract user ID
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Get user
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Generate new tokens
	tokenResponse, err := u.generateTokens(user)
	if err != nil {
		return nil, err
	}

	return tokenResponse, nil
}

// VerifyToken verifies a token
func (u *usecase) VerifyToken(ctx context.Context, tokenStr string) (*models.IDTokenPayload, error) {
	// Parse token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(u.cfg.JWT.AccessSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Extract payload
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("invalid email in token")
	}

	userType, ok := claims["user_type"].(string)
	if !ok {
		return nil, errors.New("invalid user type in token")
	}

	payload := &models.IDTokenPayload{
		UserID:   userID,
		Email:    email,
		UserType: userType,
	}

	return payload, nil
}

// GetUserByID gets a user by ID
func (u *usecase) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	return u.repo.GetUserByID(ctx, id)
}

// ChangePassword changes a user's password
func (u *usecase) ChangePassword(ctx context.Context, userID uuid.UUID, req *models.ChangePasswordRequest) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Get user
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	// Check if user is using email as auth provider
	if user.AuthProvider != "email" {
		return fmt.Errorf("password change not available for %s accounts", user.AuthProvider)
	}

	// Verify old password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword))
	if err != nil {
		return errors.New("incorrect old password")
	}

	// Hash new password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	if err := u.repo.UpdatePassword(ctx, userID, string(passwordHash)); err != nil {
		return err
	}

	return nil
}

// ForgotPassword initiates the forgot password process
func (u *usecase) ForgotPassword(ctx context.Context, req *models.ForgotPasswordRequest) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Get user by email
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists for security reasons
		return nil
	}

	// Check if user is using email as auth provider
	if user.AuthProvider != "email" {
		// Don't reveal if it's a social account for security reasons
		return nil
	}

	// TODO: Generate reset token and send email

	return nil
}

// ResetPassword resets a user's password
func (u *usecase) ResetPassword(ctx context.Context, req *models.ResetPasswordRequest) error {
	// TODO: Verify reset token
	// Get user ID from token
	// Hash password
	// Update password
	return errors.New("not implemented")
}

// VerifyEmail verifies a user's email
func (u *usecase) VerifyEmail(ctx context.Context, req *models.VerifyEmailRequest) error {
	// TODO: Verify email token
	// Get user ID from token
	// Update user verification status
	return errors.New("not implemented")
}

// UpdateProfile updates a user's profile
func (u *usecase) UpdateProfile(ctx context.Context, userID uuid.UUID, req *models.UpdateProfileRequest) (*models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// Get user
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.PreferredLanguage != "" {
		user.PreferredLanguage = req.PreferredLanguage
	}

	// Update user
	if err := u.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// generateTokens generates access and refresh tokens
func (u *usecase) generateTokens(user *models.User) (*models.TokenResponse, error) {
	// Access token expiry
	accessExpiry := time.Now().Add(time.Duration(u.cfg.JWT.AccessExpiryMinutes) * time.Minute)

	// Create access token claims
	accessClaims := jwt.MapClaims{
		"user_id":   user.ID.String(),
		"email":     user.Email,
		"user_type": user.UserType,
		"exp":       accessExpiry.Unix(),
	}

	// Create access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(u.cfg.JWT.AccessSecret))
	if err != nil {
		return nil, err
	}

	// Refresh token expiry
	refreshExpiry := time.Now().Add(time.Duration(u.cfg.JWT.RefreshExpiryDays) * 24 * time.Hour)

	// Create refresh token claims
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     refreshExpiry.Unix(),
	}

	// Create refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(u.cfg.JWT.RefreshSecret))
	if err != nil {
		return nil, err
	}

	// Create token response
	tokenResponse := &models.TokenResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiredAt:    accessExpiry,
		UserID:       user.ID,
		UserType:     user.UserType,
	}

	return tokenResponse, nil
}