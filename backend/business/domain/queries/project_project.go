package queries

// Project SQL queries
const (
	// CreateProject inserts a new project
	CreateProjectQuery = `INSERT INTO %s (
		id, user_id, name, description, color, is_archived
	) VALUES (
		$1, $2, $3, $4, $5, $6
	)`

	// GetByIDProject retrieves project by ID
	GetByIDProjectQuery = `SELECT 
		id, id, user_id, name, description, color, is_archived, created_at, updated_at 
	FROM %s 
	WHERE id = $1 `

	// UpdateProject updates project fields
	UpdateProjectQuery = `UPDATE %s SET 
		id = $1, user_id = $2, name = $3, description = $4, color = $5, is_archived = $6,
		updated_at = NOW()
	WHERE id = $7`

	// DeleteProject soft deletes project
	DeleteProjectQuery = `DELETE FROM %s 
	WHERE id = $1`

	// ListProjects retrieves paginated projects
	ListProjectsQuery = `SELECT 
		id, id, user_id, name, description, color, is_archived, created_at, updated_at 
	FROM %s`

	// CountProjects counts total projects
	CountProjectsQuery = `SELECT COUNT(*) FROM %s`
	
	
	
	// ========== Relation Queries ==========
	
	// Project with user joined
	GetByIDProjectWithUserQuery = `SELECT
        p.id,
        p.user_id,
        p.name,
        p.description,
        p.color,
        p.is_archived,
        p.created_at,
        p.updated_at,
        u.id as "user.id",
        u.email as "user.email",
        u.password_hash as "user.password_hash",
        u.name as "user.name",
        u.google_id as "user.google_id",
        u.avatar_url as "user.avatar_url",
        u.email_verified as "user.email_verified",
        u.is_active as "user.is_active",
        u.last_login_at as "user.last_login_at",
        u.created_at as "user.created_at",
        u.updated_at as "user.updated_at"
    FROM %s p
    LEFT JOIN %s u ON p.user_id = u.id
    WHERE p.id = $1`
	
)
	
