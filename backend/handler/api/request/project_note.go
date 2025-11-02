package request

import (
	"github.com/zuhrulumam/pm-tool/business/entity"
)

// CreateNoteRequest represents the request to create a Note
type CreateNoteRequest struct {
	ProjectId string `json:"project_id" validate:"required"`
	Title string `json:"title" validate:"required,max=255"`
	Content string `json:"content,omitempty"`
	ContentType string `json:"content_type" validate:"required,oneof=text audio_transcription"`
	AudioUrl string `json:"audio_url,omitempty" validate:"max=512"`
	IsPinned bool `json:"is_pinned" validate:"required"`
}

// ToEntity converts the request to an entity
func (r *CreateNoteRequest) ToEntity() *entity.Note {
	return &entity.Note{
		ProjectId: r.ProjectId,
		Title: r.Title,
		Content: &r.Content,
		ContentType: r.ContentType,
		AudioUrl: &r.AudioUrl,
		IsPinned: r.IsPinned,
	}
}

// UpdateNoteRequest represents the request to update a Note
type UpdateNoteRequest struct {
	ProjectId *string `json:"project_id"`
	Title *string `json:"title"`
	Content *string `json:"content,omitempty"`
	ContentType *string `json:"content_type"`
	AudioUrl *string `json:"audio_url,omitempty"`
	IsPinned *bool `json:"is_pinned"`
}

// ApplyTo applies the updates to an existing entity
func (r *UpdateNoteRequest) ApplyTo(entity *entity.Note) {
	if r.ProjectId != nil {
		entity.ProjectId = *r.ProjectId
	}
	if r.Title != nil {
		entity.Title = *r.Title
	}
	if r.Content != nil {
		entity.Content = r.Content
	}
	if r.ContentType != nil {
		entity.ContentType = *r.ContentType
	}
	if r.AudioUrl != nil {
		entity.AudioUrl = r.AudioUrl
	}
	if r.IsPinned != nil {
		entity.IsPinned = *r.IsPinned
	}
}

// ListNoteRequest represents filters for listing Notes
type ListNoteRequest struct {
	ProjectId *string `form:"project_id" json:"project_id,omitempty"`
	ProjectIdLike *string `form:"project_id_like" json:"project_id_like,omitempty"`
	Title *string `form:"title" json:"title,omitempty"`
	TitleLike *string `form:"title_like" json:"title_like,omitempty"`
	Content *string `form:"content,omitempty" json:"content,omitempty,omitempty"`
	ContentLike *string `form:"content,omitempty_like" json:"content,omitempty_like,omitempty"`
	ContentType *string `form:"content_type" json:"content_type,omitempty"`
	ContentTypeLike *string `form:"content_type_like" json:"content_type_like,omitempty"`
	AudioUrl *string `form:"audio_url,omitempty" json:"audio_url,omitempty,omitempty"`
	AudioUrlLike *string `form:"audio_url,omitempty_like" json:"audio_url,omitempty_like,omitempty"`
	IsPinned *bool `form:"is_pinned" json:"is_pinned,omitempty"`
	// Pagination
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

// ToFilters converts the request to a filters map for the repository
func (r *ListNoteRequest) ToFilters() map[string]interface{} {
	filters := make(map[string]interface{})
	
	if r.ProjectId != nil {
		filters["project_id"] = *r.ProjectId
	}
	if r.ProjectIdLike != nil {
		filters["project_id_like"] = *r.ProjectIdLike
	}
	if r.Title != nil {
		filters["title"] = *r.Title
	}
	if r.TitleLike != nil {
		filters["title_like"] = *r.TitleLike
	}
	if r.Content != nil {
		filters["content"] = *r.Content
	}
	if r.ContentLike != nil {
		filters["content_like"] = *r.ContentLike
	}
	if r.ContentType != nil {
		filters["content_type"] = *r.ContentType
	}
	if r.ContentTypeLike != nil {
		filters["content_type_like"] = *r.ContentTypeLike
	}
	if r.AudioUrl != nil {
		filters["audio_url"] = *r.AudioUrl
	}
	if r.AudioUrlLike != nil {
		filters["audio_url_like"] = *r.AudioUrlLike
	}
	if r.IsPinned != nil {
		filters["is_pinned"] = *r.IsPinned
	}

	return filters
}

// GetPagination returns page and pageSize with defaults
func (r *ListNoteRequest) GetPagination() (int, int) {
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
