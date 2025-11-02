package pagination

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Params represents pagination parameters
type Params struct {
	Page     int
	PageSize int
	Sort     string
	Order    string
}

// Meta represents pagination metadata
type Meta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
	TotalItems int64 `json:"total_items"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// Default pagination values
const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// FromRequest extracts pagination parameters from request
func FromRequest(c *gin.Context) Params {
	page := parseIntOrDefault(c.Query("page"), DefaultPage)
	pageSize := parseIntOrDefault(c.Query("page_size"), DefaultPageSize)
	sort := c.DefaultQuery("sort", "id")
	order := c.DefaultQuery("order", "asc")

	params := Params{
		Page:     page,
		PageSize: pageSize,
		Sort:     sort,
		Order:    order,
	}

	// Validate and normalize
	ValidateParams(&params)

	return params
}

// ValidateParams validates and normalizes pagination parameters
func ValidateParams(params *Params) {
	// Ensure page is at least 1
	if params.Page < 1 {
		params.Page = DefaultPage
	}

	// Ensure pageSize is within valid range
	if params.PageSize < 1 {
		params.PageSize = DefaultPageSize
	}
	if params.PageSize > MaxPageSize {
		params.PageSize = MaxPageSize
	}

	// Ensure order is valid
	if params.Order != "asc" && params.Order != "desc" {
		params.Order = "asc"
	}
}

// CalculateMeta calculates pagination metadata
func CalculateMeta(page, pageSize int, totalItems int64) Meta {
	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))
	if totalPages < 1 {
		totalPages = 1
	}

	return Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
		TotalItems: totalItems,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// GetOffset calculates the database offset from page and pageSize
func GetOffset(page, pageSize int) int {
	return (page - 1) * pageSize
}

// parseIntOrDefault parses a string to int or returns default value
func parseIntOrDefault(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	
	value, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	
	return value
}
