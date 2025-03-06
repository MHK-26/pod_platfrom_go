// pkg/common/logger/logger.go
package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the global logger
var Logger *zap.Logger

// Initialize initializes the logger
func Initialize(serviceName, level string) {
	// Configure logger
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	
	// Set level
	switch level {
	case "debug":
		config.Level.SetLevel(zapcore.DebugLevel)
	case "info":
		config.Level.SetLevel(zapcore.InfoLevel)
	case "warn":
		config.Level.SetLevel(zapcore.WarnLevel)
	case "error":
		config.Level.SetLevel(zapcore.ErrorLevel)
	default:
		config.Level.SetLevel(zapcore.InfoLevel)
	}

	// Add service name to all logs
	config.InitialFields = map[string]interface{}{
		"service": serviceName,
	}

	// Create logger
	var err error
	Logger, err = config.Build()
	if err != nil {
		panic(err)
	}
}

// Debug logs a debug message
func Debug(msg string, fields ...zapcore.Field) {
	Logger.Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zapcore.Field) {
	Logger.Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zapcore.Field) {
	Logger.Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zapcore.Field) {
	Logger.Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zapcore.Field) {
	Logger.Fatal(msg, fields...)
	os.Exit(1)
}

// With returns a logger with additional fields
func With(fields ...zapcore.Field) *zap.Logger {
	return Logger.With(fields...)
}

// Field creates a field for the logger
func Field(key string, value interface{}) zapcore.Field {
	return zap.Any(key, value)
}

// Close flushes the logger buffer
func Close() {
	Logger.Sync()
}