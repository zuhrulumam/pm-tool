package queries

// Note SQL queries
const (
	// CreateNote inserts a new note
	CreateNoteQuery = `INSERT INTO %s (
		id, project_id, title, content, content_type, audio_url, is_pinned
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7
	)`

	// GetByIDNote retrieves note by ID
	GetByIDNoteQuery = `SELECT 
		id, id, project_id, title, content, content_type, audio_url, is_pinned, created_at, updated_at 
	FROM %s 
	WHERE id = $1 `

	// UpdateNote updates note fields
	UpdateNoteQuery = `UPDATE %s SET 
		id = $1, project_id = $2, title = $3, content = $4, content_type = $5, audio_url = $6, is_pinned = $7,
		updated_at = NOW()
	WHERE id = $8`

	// DeleteNote soft deletes note
	DeleteNoteQuery = `DELETE FROM %s 
	WHERE id = $1`

	// ListNotes retrieves paginated notes
	ListNotesQuery = `SELECT 
		id, id, project_id, title, content, content_type, audio_url, is_pinned, created_at, updated_at 
	FROM %s`

	// CountNotes counts total notes
	CountNotesQuery = `SELECT COUNT(*) FROM %s`
	
	
	
	// ========== Relation Queries ==========
	
	// Note with project joined
	GetByIDNoteWithProjectQuery = `SELECT
        n.id,
        n.project_id,
        n.title,
        n.content,
        n.content_type,
        n.audio_url,
        n.is_pinned,
        n.created_at,
        n.updated_at,
        p.id as "project.id",
        p.user_id as "project.user_id",
        p.name as "project.name",
        p.description as "project.description",
        p.color as "project.color",
        p.is_archived as "project.is_archived",
        p.created_at as "project.created_at",
        p.updated_at as "project.updated_at"
    FROM %s n
    LEFT JOIN %s p ON n.project_id = p.id
    WHERE n.id = $1`
	
)
	
