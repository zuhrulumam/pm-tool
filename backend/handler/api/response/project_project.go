package response

import (
	"github.com/zuhrulumam/pm-tool/business/entity"
)

// ProjectResponse represents the response for a Project
type ProjectResponse struct {
	*entity.Project
}

// NewProjectResponse creates a new ProjectResponse from an entity
func NewProjectResponse(entity *entity.Project) *ProjectResponse {
	return &ProjectResponse{
		Project: entity,
	}
}

// ProjectListResponse represents a paginated list of Projects
type ProjectListResponse struct {
	Items      []*ProjectResponse `json:"items"`
	Total      int64                      `json:"total"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
	TotalPages int                        `json:"total_pages"`
}

// NewProjectListResponse creates a new ProjectListResponse
func NewProjectListResponse(entities []*entity.Project, total int64, page, pageSize int) *ProjectListResponse {
	items := make([]*ProjectResponse, len(entities))
	for i, e := range entities {
		items[i] = NewProjectResponse(e)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &ProjectListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
