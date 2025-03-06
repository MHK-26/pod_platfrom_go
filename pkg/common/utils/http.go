// pkg/common/utils/http.go
package utils

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int
	PageSize int
}

// GetPaginationParams gets pagination parameters from the request
func GetPaginationParams(c *gin.Context) PaginationParams {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	// Validate and set default values
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}
}

// CalculateOffset calculates the offset for pagination
func CalculateOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}

// RespondWithPagination sends a paginated response
func RespondWithPagination(c *gin.Context, data interface{}, totalCount, page, pageSize int) {
	// Calculate total pages
	totalPages := totalCount / pageSize
	if totalCount%pageSize != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        data,
		"total_count": totalCount,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// RespondWithSuccess sends a success response
func RespondWithSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// RespondWithCreated sends a created response
func RespondWithCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// RespondWithNoContent sends a no content response
func RespondWithNoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// ExtractIDParam extracts and validates an ID parameter
func ExtractIDParam(c *gin.Context, paramName string) (string, bool) {
	id := c.Param(paramName)
	if id == "" {
		RespondWithError(c, http.StatusBadRequest, paramName+" is required")
		return "", false
	}
	return id, true
}

// GetQueryParam gets a query parameter with a default value
func GetQueryParam(c *gin.Context, key, defaultValue string) string {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetBoolQueryParam gets a boolean query parameter with a default value
func GetBoolQueryParam(c *gin.Context, key string, defaultValue bool) bool {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}

// GetIntQueryParam gets an integer query parameter with a default value
func GetIntQueryParam(c *gin.Context, key string, defaultValue int) int {
	value := c.Query(key)
	if value == "" {
		return defaultValue
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return intValue
}

// RespondWithFile sends a file response
func RespondWithFile(c *gin.Context, fileName, contentType string, data []byte) {
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Data(http.StatusOK, contentType, data)
}