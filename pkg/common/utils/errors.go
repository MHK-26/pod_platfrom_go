// pkg/common/utils/errors.go
package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// RespondWithError sends an error response
func RespondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{
		Status:  code,
		Message: message,
	})
}

// RespondWithValidationError sends a validation error response
func RespondWithValidationError(c *gin.Context, errors map[string]string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"status":  http.StatusBadRequest,
		"message": "Validation failed",
		"errors":  errors,
	})
}