// pkg/common/config/config.go
package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the application configuration
type Config struct {
	Server   ServerConfig
	DB       DBConfig
	JWT      JWTConfig
	Storage  StorageConfig
	MediaURL string
}

// ServerConfig represents the server configuration
type ServerConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Mode         string
}

// DBConfig represents the database configuration
type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int
	MaxIdle  int
	Timeout  time.Duration
}

// JWTConfig represents the JWT configuration
type JWTConfig struct {
	AccessSecret        string
	RefreshSecret       string
	AccessExpiryMinutes int
	RefreshExpiryDays   int
}

// StorageConfig represents the file storage configuration
type StorageConfig struct {
	BasePath string // Base path for storing files
	MaxSize  int64  // Maximum file size in bytes
}

// LoadConfig loads the application configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	// Server config
	serverPort := getEnv("SERVER_PORT", "8080")
	serverMode := getEnv("SERVER_MODE", "release")
	readTimeout, _ := strconv.Atoi(getEnv("SERVER_READ_TIMEOUT", "5"))
	writeTimeout, _ := strconv.Atoi(getEnv("SERVER_WRITE_TIMEOUT", "5"))

	// Database config
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "podcast_platform")
	dbSSLMode := getEnv("DB_SSL_MODE", "disable")
	dbMaxConns, _ := strconv.Atoi(getEnv("DB_MAX_CONNS", "20"))
	dbMaxIdle, _ := strconv.Atoi(getEnv("DB_MAX_IDLE", "5"))
	dbTimeout, _ := strconv.Atoi(getEnv("DB_TIMEOUT", "5"))

	// JWT config
	jwtAccessSecret := getEnv("JWT_ACCESS_SECRET", "access_secret")
	jwtRefreshSecret := getEnv("JWT_REFRESH_SECRET", "refresh_secret")
	jwtAccessExpiryMinutes, _ := strconv.Atoi(getEnv("JWT_ACCESS_EXPIRY_MINUTES", "15"))
	jwtRefreshExpiryDays, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXPIRY_DAYS", "7"))

	// File storage config
	storagePath := getEnv("STORAGE_PATH", "./storage")
	maxFileSize, _ := strconv.ParseInt(getEnv("MAX_FILE_SIZE", "52428800"), 10, 64) // 50MB default

	// Media URL for public access
	mediaURL := getEnv("MEDIA_URL", "http://localhost:8080/media")

	return &Config{
		Server: ServerConfig{
			Port:         serverPort,
			Mode:         serverMode,
			ReadTimeout:  time.Duration(readTimeout) * time.Second,
			WriteTimeout: time.Duration(writeTimeout) * time.Second,
		},
		DB: DBConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			DBName:   dbName,
			SSLMode:  dbSSLMode,
			MaxConns: dbMaxConns,
			MaxIdle:  dbMaxIdle,
			Timeout:  time.Duration(dbTimeout) * time.Second,
		},
		JWT: JWTConfig{
			AccessSecret:        jwtAccessSecret,
			RefreshSecret:       jwtRefreshSecret,
			AccessExpiryMinutes: jwtAccessExpiryMinutes,
			RefreshExpiryDays:   jwtRefreshExpiryDays,
		},
		Storage: StorageConfig{
			BasePath: storagePath,
			MaxSize:  maxFileSize,
		},
		MediaURL: mediaURL,
	}, nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}