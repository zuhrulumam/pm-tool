package request

import (
	"time"
	"github.com/zuhrulumam/pm-tool/business/entity"
)

// CreateUserRequest represents the request to create a User
type CreateUserRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
	PasswordHash string `json:"password_hash,omitempty" validate:"max=255"`
	Name string `json:"name,omitempty" validate:"max=100"`
	GoogleId string `json:"google_id,omitempty" validate:"max=255"`
	AvatarUrl string `json:"avatar_url,omitempty" validate:"max=512"`
	EmailVerified bool `json:"email_verified" validate:"required"`
	IsActive bool `json:"is_active" validate:"required"`
	LastLoginAt time.Time `json:"last_login_at,omitempty"`
}

// ToEntity converts the request to an entity
func (r *CreateUserRequest) ToEntity() *entity.User {
	return &entity.User{
		Email: r.Email,
		PasswordHash: &r.PasswordHash,
		Name: &r.Name,
		GoogleId: &r.GoogleId,
		AvatarUrl: &r.AvatarUrl,
		EmailVerified: r.EmailVerified,
		IsActive: r.IsActive,
		LastLoginAt: &r.LastLoginAt,
	}
}

// UpdateUserRequest represents the request to update a User
type UpdateUserRequest struct {
	Email *string `json:"email"`
	PasswordHash *string `json:"password_hash,omitempty"`
	Name *string `json:"name,omitempty"`
	GoogleId *string `json:"google_id,omitempty"`
	AvatarUrl *string `json:"avatar_url,omitempty"`
	EmailVerified *bool `json:"email_verified"`
	IsActive *bool `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// ApplyTo applies the updates to an existing entity
func (r *UpdateUserRequest) ApplyTo(entity *entity.User) {
	if r.Email != nil {
		entity.Email = *r.Email
	}
	if r.PasswordHash != nil {
		entity.PasswordHash = r.PasswordHash
	}
	if r.Name != nil {
		entity.Name = r.Name
	}
	if r.GoogleId != nil {
		entity.GoogleId = r.GoogleId
	}
	if r.AvatarUrl != nil {
		entity.AvatarUrl = r.AvatarUrl
	}
	if r.EmailVerified != nil {
		entity.EmailVerified = *r.EmailVerified
	}
	if r.IsActive != nil {
		entity.IsActive = *r.IsActive
	}
	if r.LastLoginAt != nil {
		entity.LastLoginAt = r.LastLoginAt
	}
}

// ListUserRequest represents filters for listing Users
type ListUserRequest struct {
	Email *string `form:"email" json:"email,omitempty"`
	EmailLike *string `form:"email_like" json:"email_like,omitempty"`
	PasswordHash *string `form:"password_hash,omitempty" json:"password_hash,omitempty,omitempty"`
	PasswordHashLike *string `form:"password_hash,omitempty_like" json:"password_hash,omitempty_like,omitempty"`
	Name *string `form:"name,omitempty" json:"name,omitempty,omitempty"`
	NameLike *string `form:"name,omitempty_like" json:"name,omitempty_like,omitempty"`
	GoogleId *string `form:"google_id,omitempty" json:"google_id,omitempty,omitempty"`
	GoogleIdLike *string `form:"google_id,omitempty_like" json:"google_id,omitempty_like,omitempty"`
	AvatarUrl *string `form:"avatar_url,omitempty" json:"avatar_url,omitempty,omitempty"`
	AvatarUrlLike *string `form:"avatar_url,omitempty_like" json:"avatar_url,omitempty_like,omitempty"`
	EmailVerified *bool `form:"email_verified" json:"email_verified,omitempty"`
	IsActive *bool `form:"is_active" json:"is_active,omitempty"`
	LastLoginAt *time.Time `form:"last_login_at,omitempty" json:"last_login_at,omitempty,omitempty"`
	LastLoginAtFrom *time.Time `form:"last_login_at,omitempty_from" json:"last_login_at,omitempty_from,omitempty"`
	LastLoginAtTo   *time.Time `form:"last_login_at,omitempty_to" json:"last_login_at,omitempty_to,omitempty"`
	// Pagination
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

// ToFilters converts the request to a filters map for the repository
func (r *ListUserRequest) ToFilters() map[string]interface{} {
	filters := make(map[string]interface{})
	
	if r.Email != nil {
		filters["email"] = *r.Email
	}
	if r.EmailLike != nil {
		filters["email_like"] = *r.EmailLike
	}
	if r.PasswordHash != nil {
		filters["password_hash"] = *r.PasswordHash
	}
	if r.PasswordHashLike != nil {
		filters["password_hash_like"] = *r.PasswordHashLike
	}
	if r.Name != nil {
		filters["name"] = *r.Name
	}
	if r.NameLike != nil {
		filters["name_like"] = *r.NameLike
	}
	if r.GoogleId != nil {
		filters["google_id"] = *r.GoogleId
	}
	if r.GoogleIdLike != nil {
		filters["google_id_like"] = *r.GoogleIdLike
	}
	if r.AvatarUrl != nil {
		filters["avatar_url"] = *r.AvatarUrl
	}
	if r.AvatarUrlLike != nil {
		filters["avatar_url_like"] = *r.AvatarUrlLike
	}
	if r.EmailVerified != nil {
		filters["email_verified"] = *r.EmailVerified
	}
	if r.IsActive != nil {
		filters["is_active"] = *r.IsActive
	}
	if r.LastLoginAt != nil {
		filters["last_login_at"] = *r.LastLoginAt
	}
	if r.LastLoginAtFrom != nil {
		filters["last_login_at_from"] = *r.LastLoginAtFrom
	}
	if r.LastLoginAtTo != nil {
		filters["last_login_at_to"] = *r.LastLoginAtTo
	}

	return filters
}

// GetPagination returns page and pageSize with defaults
func (r *ListUserRequest) GetPagination() (int, int) {
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
