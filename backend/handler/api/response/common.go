package response

import "time"

// Response represents a single item response
type Response struct {
	Data interface{} `json:"data"`
}

// ListResponse represents a list response with pagination
type ListResponse struct {
	Data []interface{}   `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	CurrentPage int   `json:"current_page"`
	PageSize    int   `json:"page_size"`
	TotalPages  int   `json:"total_pages"`
	TotalItems  int64 `json:"total_items"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
}

// NewPaginationMeta creates pagination metadata
func NewPaginationMeta(currentPage, pageSize int, totalItems int64) PaginationMeta {
	totalPages := int(totalItems) / pageSize
	if int(totalItems)%pageSize > 0 {
		totalPages++
	}

	return PaginationMeta{
		CurrentPage: currentPage,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		TotalItems:  totalItems,
		HasNext:     currentPage < totalPages,
		HasPrevious: currentPage > 1,
	}
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
}
