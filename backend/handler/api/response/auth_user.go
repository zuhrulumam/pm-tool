package response

import (
	"github.com/zuhrulumam/pm-tool/business/entity"
)

// UserResponse represents the response for a User
type UserResponse struct {
	*entity.User
}

// NewUserResponse creates a new UserResponse from an entity
func NewUserResponse(entity *entity.User) *UserResponse {
	return &UserResponse{
		User: entity,
	}
}

// UserListResponse represents a paginated list of Users
type UserListResponse struct {
	Items      []*UserResponse `json:"items"`
	Total      int64                      `json:"total"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
	TotalPages int                        `json:"total_pages"`
}

// NewUserListResponse creates a new UserListResponse
func NewUserListResponse(entities []*entity.User, total int64, page, pageSize int) *UserListResponse {
	items := make([]*UserResponse, len(entities))
	for i, e := range entities {
		items[i] = NewUserResponse(e)
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize != 0 {
		totalPages++
	}

	return &UserListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
