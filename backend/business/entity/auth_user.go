package entity

import (
	"time"
)

// User represents the users entity
type User struct {
	Id            string     `db:"id" json:"id"`
	Email         string     `db:"email" json:"email" validate:"required,email,max=255"`
	PasswordHash  *string    `db:"password_hash" json:"password_hash,omitempty" validate:"max=255"`
	Name          *string    `db:"name" json:"name,omitempty" validate:"max=100"`
	GoogleId      *string    `db:"google_id" json:"google_id,omitempty" validate:"max=255"`
	AvatarUrl     *string    `db:"avatar_url" json:"avatar_url,omitempty" validate:"max=512"`
	EmailVerified bool       `db:"email_verified" json:"email_verified" validate:"required"`
	IsActive      bool       `db:"is_active" json:"is_active" validate:"required"`
	LastLoginAt   *time.Time `db:"last_login_at" json:"last_login_at,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at" validate:"required"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at" validate:"required"`
}

// TableName returns the table name for User
func (e *User) TableName() string {
	return "users"
}

// Validate validates the User entity
func (e *User) Validate() error {
	// Validation is handled by the validator package
	// This method exists for interface compatibility
	return nil
}
