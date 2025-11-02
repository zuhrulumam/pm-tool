package entity

import (
	"time"
)

// Note represents the notes entity
type Note struct {
	
	Id string `db:"id" json:"id"`
	ProjectId string `db:"project_id" json:"project_id" validate:"required"` // Foreign key reference to Project
	Title string `db:"title" json:"title" validate:"required,max=255"`
	Content *string `db:"content" json:"content,omitempty"`
	ContentType string `db:"content_type" json:"content_type" validate:"required,oneof=text audio_transcription"` // Possible values: text, audio_transcription
	AudioUrl *string `db:"audio_url" json:"audio_url,omitempty" validate:"max=512"`
	IsPinned bool `db:"is_pinned" json:"is_pinned" validate:"required"`
	CreatedAt time.Time `db:"created_at" json:"created_at" validate:"required"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at" validate:"required"`

	// Relationships (not persisted, loaded separately)
	Project *Project `db:"-" json:"project,omitempty"` // belongs_to relation

}

// TableName returns the table name for Note
func (e *Note) TableName() string {
	return "notes"
}

// Validate validates the Note entity
func (e *Note) Validate() error {
	// Validation is handled by the validator package
	// This method exists for interface compatibility
	return nil
}

