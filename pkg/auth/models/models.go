// pkg/auth/models/models.go
package models

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Email          string     `json:"email" db:"email"`
	Username       string     `json:"username" db:"username"`
	PasswordHash   string     `json:"-" db:"password_hash"`
	FullName       string     `json:"full_name" db:"full_name"`
	Bio            string     `json:"bio" db:"bio"`
	ProfileImageURL string    `json:"profile_image_url" db:"profile_image_url"`
	UserType       string     `json:"user_type" db:"user_type"`
	AuthProvider   string     `json:"auth_provider" db:"auth_provider"`
	AuthProviderID string     `json:"auth_provider_id" db:"auth_provider_id"`
	IsVerified     bool       `json:"is_verified" db:"is_verified"`
	PreferredLanguage string  `json:"preferred_language" db:"preferred_language"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
	LastLoginAt    *time.Time `json:"last_login_at" db:"last_login_at"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email           string `json:"email" validate:"required,email"`
	Username        string `json:"username" validate:"required,min=3,max=50"`
	Password        string `json:"password" validate:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
	FullName        string `json:"full_name"`
	UserType        string `json:"user_type" validate:"required,oneof=listener podcaster"`
}

// SocialLoginRequest represents a social login request
type SocialLoginRequest struct {
	Provider string `json:"provider" validate:"required,oneof=google apple"`
	Token    string `json:"token" validate:"required"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiredAt    time.Time `json:"expired_at"`
	UserID       uuid.UUID `json:"user_id"`
	UserType     string    `json:"user_type"`
}

// IDTokenPayload represents the payload of the ID token
type IDTokenPayload struct {
	UserID   uuid.UUID `json:"user_id"`
	Email    string    `json:"email"`
	UserType string    `json:"user_type"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// ChangePasswordRequest represents a change password request
type ChangePasswordRequest struct {
	OldPassword    string `json:"old_password" validate:"required"`
	NewPassword    string `json:"new_password" validate:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

// ForgotPasswordRequest represents a forgot password request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents a reset password request
type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}

// VerifyEmailRequest represents a verify email request
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required"`
}

// UpdateProfileRequest represents an update profile request
type UpdateProfileRequest struct {
	FullName         string `json:"full_name"`
	Bio              string `json:"bio"`
	PreferredLanguage string `json:"preferred_language"`
}