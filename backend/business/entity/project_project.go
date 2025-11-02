package entity

import (
	"time"
)

// Project represents the projects entity
type Project struct {
	
	Id string `db:"id" json:"id"`
	UserId string `db:"user_id" json:"user_id" validate:"required"` // Foreign key reference to User
	Name string `db:"name" json:"name" validate:"required,max=255"`
	Description *string `db:"description" json:"description,omitempty"`
	Color *string `db:"color" json:"color,omitempty" validate:"max=7"`
	IsArchived bool `db:"is_archived" json:"is_archived" validate:"required"`
	CreatedAt time.Time `db:"created_at" json:"created_at" validate:"required"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at" validate:"required"`

	// Relationships (not persisted, loaded separately)
	User *User `db:"-" json:"user,omitempty"` // belongs_to relation

}

// TableName returns the table name for Project
func (e *Project) TableName() string {
	return "projects"
}

// Validate validates the Project entity
func (e *Project) Validate() error {
	// Validation is handled by the validator package
	// This method exists for interface compatibility
	return nil
}

