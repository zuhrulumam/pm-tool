package response

import (
	"github.com/zuhrulumam/pm-tool/business/entity"
)

// NoteResponse represents the response for a Note
type NoteResponse struct {
	*entity.Note
}

// NewNoteResponse creates a new NoteResponse from an entity
func NewNoteResponse(entity *entity.Note) *NoteResponse {
	return &NoteResponse{
		Note: entity,
	}
}

// NoteListResponse represents a paginated list of Notes
type NoteListResponse struct {
	Items      []*NoteResponse `json:"items"`
	Total      int64                      `json:"total"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
	TotalPages int                        `json:"total_pages"`
}

// NewNoteListResponse creates a new NoteListResponse
func NewNoteListResponse(entities []*entity.Note, total int64, page, pageSize int) *NoteListResponse {
	items := make([]*NoteResponse, len(entities))
	for i, e := range entities {
		items[i] = NewNoteResponse(e)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &NoteListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
