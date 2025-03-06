package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-username/podcast-platform/pkg/common/logger"
	"go.uber.org/zap"
)

// LoggingMiddleware is a middleware that logs each request
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		// Create request buffer
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Create response buffer
		responseBodyWriter := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = responseBodyWriter

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log request
		fields := []zap.Field{
			logger.Field("request_id", requestID),
			logger.Field("method", c.Request.Method),
			logger.Field("path", c.Request.URL.Path),
			logger.Field("query", c.Request.URL.RawQuery),
			logger.Field("ip", c.ClientIP()),
			logger.Field("user_agent", c.Request.UserAgent()),
			logger.Field("status", c.Writer.Status()),
			logger.Field("duration", duration.String()),
			logger.Field("duration_ms", duration.Milliseconds()),
			logger.Field("size", c.Writer.Size()),
		}

		// Add request body for non-GET methods if it's not too large
		if c.Request.Method != "GET" && len(requestBody) > 0 && len(requestBody) < 10000 {
			fields = append(fields, logger.Field("request_body", string(requestBody)))
		}

		// Add response body if it's not too large
		if responseBodyWriter.body.Len() > 0 && responseBodyWriter.body.Len() < 10000 {
			fields = append(fields, logger.Field("response_body", responseBodyWriter.body.String()))
		}

		// Log based on status code
		if c.Writer.Status() >= 500 {
			logger.Error("Server error", fields...)
		} else if c.Writer.Status() >= 400 {
			logger.Warn("Client error", fields...)
		} else {
			logger.Info("Request processed", fields...)
		}
	}
}

// bodyWriter is a custom response writer that captures the response body
type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write captures the response body and writes it to the original writer
func (w *bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}