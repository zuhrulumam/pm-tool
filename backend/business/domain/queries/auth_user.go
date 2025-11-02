package queries

// User SQL queries
const (
	// CreateUser inserts a new user
	CreateUserQuery = `INSERT INTO %s (
		id, email, password_hash, name, google_id, avatar_url, email_verified, is_active, last_login_at
	) VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9
	)`

	// GetByIDUser retrieves user by ID
	GetByIDUserQuery = `SELECT 
		id, id, email, password_hash, name, google_id, avatar_url, email_verified, is_active, last_login_at, created_at, updated_at 
	FROM %s 
	WHERE id = $1 `

	// UpdateUser updates user fields
	UpdateUserQuery = `UPDATE %s SET 
		id = $1, email = $2, password_hash = $3, name = $4, google_id = $5, avatar_url = $6, email_verified = $7, is_active = $8, last_login_at = $9,
		updated_at = NOW()
	WHERE id = $10`

	// DeleteUser soft deletes user
	DeleteUserQuery = `DELETE FROM %s 
	WHERE id = $1`

	// ListUsers retrieves paginated users
	ListUsersQuery = `SELECT 
		id, id, email, password_hash, name, google_id, avatar_url, email_verified, is_active, last_login_at, created_at, updated_at 
	FROM %s`

	// CountUsers counts total users
	CountUsersQuery = `SELECT COUNT(*) FROM %s`
	
	// CheckUniqueEmail checks if email already exists
	CheckUniqueEmailQuery = `SELECT COUNT(*) FROM %s WHERE email = $1 `// CheckUniqueGoogleId checks if google_id already exists
	CheckUniqueGoogleIdQuery = `SELECT COUNT(*) FROM %s WHERE google_id = $1 `
)
	
