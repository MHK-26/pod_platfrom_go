// pkg/auth/repository/postgres/repository.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/MHK-26/pod_platfrom_go/pkg/auth/models"
)

// Repository defines the methods for the auth repository
type Repository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByAuthProvider(ctx context.Context, provider, providerID string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type repository struct {
	db *sqlx.DB
}

// NewRepository creates a new auth repository
func NewRepository(db *sqlx.DB) Repository {
	return &repository{db: db}
}

// CreateUser creates a new user
func (r *repository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (
			id, email, username, password_hash, full_name, bio, profile_image_url,
			user_type, auth_provider, auth_provider_id, is_verified, preferred_language,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		) RETURNING id
	`

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRowContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.FullName,
		user.Bio,
		user.ProfileImageURL,
		user.UserType,
		user.AuthProvider,
		user.AuthProviderID,
		user.IsVerified,
		user.PreferredLanguage,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID)

	return err
}

// GetUserByID gets a user by ID
func (r *repository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `
		SELECT
			id, email, username, password_hash, full_name, bio, profile_image_url,
			user_type, auth_provider, auth_provider_id, is_verified, preferred_language,
			created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail gets a user by email
func (r *repository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `
		SELECT
			id, email, username, password_hash, full_name, bio, profile_image_url,
			user_type, auth_provider, auth_provider_id, is_verified, preferred_language,
			created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1
	`

	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByUsername gets a user by username
func (r *repository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	query := `
		SELECT
			id, email, username, password_hash, full_name, bio, profile_image_url,
			user_type, auth_provider, auth_provider_id, is_verified, preferred_language,
			created_at, updated_at, last_login_at
		FROM users
		WHERE username = $1
	`

	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByAuthProvider gets a user by auth provider
func (r *repository) GetUserByAuthProvider(ctx context.Context, provider, providerID string) (*models.User, error) {
	var user models.User
	query := `
		SELECT
			id, email, username, password_hash, full_name, bio, profile_image_url,
			user_type, auth_provider, auth_provider_id, is_verified, preferred_language,
			created_at, updated_at, last_login_at
		FROM users
		WHERE auth_provider = $1 AND auth_provider_id = $2
	`

	err := r.db.GetContext(ctx, &user, query, provider, providerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates a user
func (r *repository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET
			email = $2,
			username = $3,
			full_name = $4,
			bio = $5,
			profile_image_url = $6,
			user_type = $7,
			auth_provider = $8,
			auth_provider_id = $9,
			is_verified = $10,
			preferred_language = $11,
			updated_at = $12
		WHERE id = $1
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		user.ID,
		user.Email,
		user.Username,
		user.FullName,
		user.Bio,
		user.ProfileImageURL,
		user.UserType,
		user.AuthProvider,
		user.AuthProviderID,
		user.IsVerified,
		user.PreferredLanguage,
		user.UpdatedAt,
	)

	return err
}

// UpdateLastLogin updates the last login timestamp
func (r *repository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET last_login_at = $2
		WHERE id = $1
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, now)
	return err
}

// UpdatePassword updates a user's password
func (r *repository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $2, updated_at = $3
		WHERE id = $1
	`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query, userID, passwordHash, now)
	return err
}

// DeleteUser deletes a user
func (r *repository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}