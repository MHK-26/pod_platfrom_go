package storage

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-username/podcast-platform/pkg/common/config"
)

// Service defines the interface for storage operations
type Service interface {
	// SaveFile saves a file and returns the path to the file
	SaveFile(file *multipart.FileHeader, directory string) (string, error)
	
	// GetFileURL returns the URL to the file
	GetFileURL(filePath string) string
	
	// DeleteFile deletes a file
	DeleteFile(filePath string) error
}

type localService struct {
	cfg *config.Config
}

// NewLocalService creates a new local storage service
func NewLocalService(cfg *config.Config) Service {
	// Ensure base directory exists
	os.MkdirAll(cfg.Storage.BasePath, os.ModePerm)
	
	return &localService{
		cfg: cfg,
	}
}

// SaveFile saves a file to the local filesystem
func (s *localService) SaveFile(file *multipart.FileHeader, directory string) (string, error) {
	// Check file size
	if file.Size > s.cfg.Storage.MaxSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of %d bytes", s.cfg.Storage.MaxSize)
	}
	
	// Get file extension
	ext := filepath.Ext(file.Filename)
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, // Images
		".mp3": true, ".m4a": true, ".wav": true, ".ogg": true, // Audio
	}
	
	if !allowedExts[strings.ToLower(ext)] {
		return "", errors.New("file type not allowed")
	}
	
	// Create directory if it doesn't exist
	dirPath := filepath.Join(s.cfg.Storage.BasePath, directory)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return "", err
	}
	
	// Generate a unique filename
	filename := uuid.New().String() + ext
	filePath := filepath.Join(dirPath, filename)
	
	// Open the source file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()
	
	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	
	// Copy the file
	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}
	
	// Return the relative path
	relativePath := filepath.Join(directory, filename)
	return relativePath, nil
}

// GetFileURL returns the URL to the file
func (s *localService) GetFileURL(filePath string) string {
	// Return the URL based on the media URL in the config
	return fmt.Sprintf("%s/%s", s.cfg.MediaURL, filePath)
}

// DeleteFile deletes a file
func (s *localService) DeleteFile(filePath string) error {
	// Get the absolute file path
	absPath := filepath.Join(s.cfg.Storage.BasePath, filePath)
	
	// Check if file exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return errors.New("file not found")
	}
	
	// Delete the file
	return os.Remove(absPath)
}

// SetupMediaRoute sets up a route for serving media files
func SetupMediaRoute(r *gin.Engine, storagePath string) {
	r.Static("/media", storagePath)
}