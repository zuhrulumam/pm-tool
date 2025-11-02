package request

import (
	"github.com/zuhrulumam/pm-tool/business/entity"
)

// CreateProjectRequest represents the request to create a Project
type CreateProjectRequest struct {
	UserId string `json:"user_id" validate:"required"`
	Name string `json:"name" validate:"required,max=255"`
	Description string `json:"description,omitempty"`
	Color string `json:"color,omitempty" validate:"max=7"`
	IsArchived bool `json:"is_archived" validate:"required"`
}

// ToEntity converts the request to an entity
func (r *CreateProjectRequest) ToEntity() *entity.Project {
	return &entity.Project{
		UserId: r.UserId,
		Name: r.Name,
		Description: &r.Description,
		Color: &r.Color,
		IsArchived: r.IsArchived,
	}
}

// UpdateProjectRequest represents the request to update a Project
type UpdateProjectRequest struct {
	UserId *string `json:"user_id"`
	Name *string `json:"name"`
	Description *string `json:"description,omitempty"`
	Color *string `json:"color,omitempty"`
	IsArchived *bool `json:"is_archived"`
}

// ApplyTo applies the updates to an existing entity
func (r *UpdateProjectRequest) ApplyTo(entity *entity.Project) {
	if r.UserId != nil {
		entity.UserId = *r.UserId
	}
	if r.Name != nil {
		entity.Name = *r.Name
	}
	if r.Description != nil {
		entity.Description = r.Description
	}
	if r.Color != nil {
		entity.Color = r.Color
	}
	if r.IsArchived != nil {
		entity.IsArchived = *r.IsArchived
	}
}

// ListProjectRequest represents filters for listing Projects
type ListProjectRequest struct {
	UserId *string `form:"user_id" json:"user_id,omitempty"`
	UserIdLike *string `form:"user_id_like" json:"user_id_like,omitempty"`
	Name *string `form:"name" json:"name,omitempty"`
	NameLike *string `form:"name_like" json:"name_like,omitempty"`
	Description *string `form:"description,omitempty" json:"description,omitempty,omitempty"`
	DescriptionLike *string `form:"description,omitempty_like" json:"description,omitempty_like,omitempty"`
	Color *string `form:"color,omitempty" json:"color,omitempty,omitempty"`
	ColorLike *string `form:"color,omitempty_like" json:"color,omitempty_like,omitempty"`
	IsArchived *bool `form:"is_archived" json:"is_archived,omitempty"`
	// Pagination
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

// ToFilters converts the request to a filters map for the repository
func (r *ListProjectRequest) ToFilters() map[string]interface{} {
	filters := make(map[string]interface{})
	
	if r.UserId != nil {
		filters["user_id"] = *r.UserId
	}
	if r.UserIdLike != nil {
		filters["user_id_like"] = *r.UserIdLike
	}
	if r.Name != nil {
		filters["name"] = *r.Name
	}
	if r.NameLike != nil {
		filters["name_like"] = *r.NameLike
	}
	if r.Description != nil {
		filters["description"] = *r.Description
	}
	if r.DescriptionLike != nil {
		filters["description_like"] = *r.DescriptionLike
	}
	if r.Color != nil {
		filters["color"] = *r.Color
	}
	if r.ColorLike != nil {
		filters["color_like"] = *r.ColorLike
	}
	if r.IsArchived != nil {
		filters["is_archived"] = *r.IsArchived
	}

	return filters
}

// GetPagination returns page and pageSize with defaults
func (r *ListProjectRequest) GetPagination() (int, int) {
	page := r.Page
	if page < 1 {
		page = 1
	}
	
	pageSize := r.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	
	return page, pageSize
}
