package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zuhrulumam/pm-tool/infra/redis"
	"go.opentelemetry.io/otel/trace"

	"github.com/zuhrulumam/pm-tool/business/domain/queries"
	"github.com/zuhrulumam/pm-tool/business/entity"

	"github.com/zuhrulumam/pm-tool/config"
	"github.com/zuhrulumam/pm-tool/pkg/db"
	apperr "github.com/zuhrulumam/pm-tool/pkg/errors"
	"github.com/zuhrulumam/pm-tool/pkg/httpclient"
	qbu "github.com/zuhrulumam/pm-tool/pkg/query_builder"
	"github.com/zuhrulumam/pm-tool/pkg/transaction"
)

type ProjectDomainItf interface {
	Create(ctx context.Context, project *entity.Project) error
	GetByID(ctx context.Context, id string) (*entity.Project, error)
	List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*entity.Project, int64, error)
	Update(ctx context.Context, id string, project *entity.Project) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
	// Relation loaders
	GetByIDWithUser(ctx context.Context, id string) (*entity.Project, error)
	LoadUser(ctx context.Context, project *entity.Project) error
}

// ProjectDomain handles business logic for Project
// Schema: project.
// Table: project.projects
type projectDomain struct {
	db    *db.DB // Leader for writes, Follower for reads
	redis *redis.Client

	http         *httpclient.Client
	tracer       trace.Tracer
	schemaPrefix string
	cfg          *config.Config
}

// NewProjectDomain creates a new ProjectDomain instance with all dependencies
func NewProjectDomain(
	db *db.DB,
	conf *config.Config,
	redisClient *redis.Client,

	httpClient *httpclient.Client,
	tracer trace.Tracer,
	schemaPrefix string,
) ProjectDomainItf {
	return &projectDomain{
		db:    db,
		redis: redisClient,

		http:         httpClient,
		tracer:       tracer,
		schemaPrefix: schemaPrefix,
		cfg:          conf,
	}
}

// tableName returns full table name with schema
func (d *projectDomain) tableName() string {

	return "project.projects"
}

// getExecutor returns appropriate executor based on context
// If transaction exists in context, use that
// Otherwise, use db directly
func (d *projectDomain) getExecutor(ctx context.Context) db.Executor {
	if tx := transaction.GetTxFromContext(ctx); tx != nil {
		return tx
	}
	return d.db
}

// Create creates a new Project with full business logic
func (d *projectDomain) Create(ctx context.Context, project *entity.Project) error {
	ctx, span := d.tracer.Start(ctx, "domain.Project.Create")
	defer span.End()

	// 1. Business validation
	if err := d.validateCreate(ctx, project); err != nil {
		return apperr.Propagate(err, apperr.CodeValidationError, "validation failed", 400)
	}

	// 2. Check uniqueness via cache

	// 3. Get executor (transaction from context or db)
	exec := d.getExecutor(ctx)

	// 4. Execute insert
	query := fmt.Sprintf(queries.CreateProjectQuery, d.tableName())
	project.Id = uuid.NewString()
	_, err := exec.ExecContext(ctx, query,
		project.Id,
		project.UserId,
		project.Name,
		project.Description,
		project.Color,
		project.IsArchived,
	)
	if err != nil {
		return apperr.DatabaseError(err, "create project")
	}

	// 5. Update cache
	d.cacheSet(ctx, project)

	return nil
}

// GetByID retrieves a Project by ID with cache support
func (d *projectDomain) GetByID(ctx context.Context, id string) (*entity.Project, error) {
	ctx, span := d.tracer.Start(ctx, "domain.Project.GetByID")
	defer span.End()

	// 1. Try cache first
	if project, err := d.cacheGet(ctx, id); err == nil && project != nil {
		return project, nil
	}

	// 2. Get executor (transaction from context or db)
	exec := d.getExecutor(ctx)

	// 3. Query database
	query := fmt.Sprintf(queries.GetByIDProjectQuery, d.tableName())
	var project entity.Project
	err := exec.GetContext(ctx, &project, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFoundError("project")
		}
		return nil, apperr.DatabaseError(err, "get project")
	}

	// 4. Update cache
	d.cacheSet(ctx, &project)

	return &project, nil
}

// List retrieves paginated Projects with filters
func (d *projectDomain) List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*entity.Project, int64, error) {
	ctx, span := d.tracer.Start(ctx, "domain.Project.List")
	defer span.End()

	// Build WHERE clause using query builder
	qb := qbu.NewQueryBuilder()
	qb.AddFilters(filters)
	whereClause, args := qb.Build()

	// Get executor
	exec := d.getExecutor(ctx)

	// Count total
	countQuery := fmt.Sprintf(queries.CountProjectsQuery, d.tableName())
	if whereClause != "" {
		countQuery = fmt.Sprintf("%s %s", countQuery, whereClause)
	}

	var total int64
	if err := exec.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, apperr.DatabaseError(err, "count projects")
	}

	listQuery := fmt.Sprintf(queries.ListProjectsQuery, d.tableName())
	if whereClause != "" {
		listQuery = fmt.Sprintf("%s %s", listQuery, whereClause)
	}
	// Add pagination
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("%s ORDER BY id DESC LIMIT ? OFFSET ?", listQuery)
	args = append(args, pageSize, offset)

	var entities []*entity.Project
	if err := exec.SelectContext(ctx, &entities, query, args...); err != nil {
		return nil, 0, apperr.DatabaseError(err, "list projects")
	}

	return entities, total, nil
}

// Update updates an existing Project with full business logic
func (d *projectDomain) Update(ctx context.Context, id string, project *entity.Project) error {
	ctx, span := d.tracer.Start(ctx, "domain.Project.Update")
	defer span.End()

	// 1. Get existing entity
	existing, err := d.GetByID(ctx, id)
	if err != nil {
		return err // Already wrapped
	}

	// 2. Business validation
	if err := d.validateUpdate(ctx, existing, project); err != nil {
		return apperr.Propagate(err, apperr.CodeValidationError, "validation failed", 400)
	}

	// 4. Get executor (transaction from context or db)
	exec := d.getExecutor(ctx)

	// 5. Execute update
	query := fmt.Sprintf(queries.UpdateProjectQuery, d.tableName())
	_, err = exec.ExecContext(ctx, query,
		project.Id,
		project.UserId,
		project.Name,
		project.Description,
		project.Color,
		project.IsArchived,
		id,
	)
	if err != nil {
		return apperr.DatabaseError(err, "update project")
	}

	// 6. Invalidate cache
	d.cacheDelete(ctx, id)

	return nil
}

// Delete deletes a Project with business logic validation (soft delete)
func (d *projectDomain) Delete(ctx context.Context, id string) error {
	ctx, span := d.tracer.Start(ctx, "domain.Project.Delete")
	defer span.End()

	// 1. Get existing for validation and cache cleanup
	existing, err := d.GetByID(ctx, id)
	if err != nil {
		return err // Already wrapped
	}

	// 2. Business validation
	if err := d.validateDelete(ctx, existing); err != nil {
		return apperr.Propagate(err, apperr.CodeValidationError, "validation failed", 400)
	}

	// 3. Get executor (transaction from context or db)
	exec := d.getExecutor(ctx)

	// 4. Execute soft delete
	query := fmt.Sprintf(queries.DeleteProjectQuery, d.tableName())
	_, err = exec.ExecContext(ctx, query, id)
	if err != nil {
		return apperr.DatabaseError(err, "delete project")
	}

	// 5. Invalidate all related cache
	d.cacheDelete(ctx, id)

	return nil
}

// Count returns the total count of Projects
func (d *projectDomain) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	ctx, span := d.tracer.Start(ctx, "domain.Project.Count")
	defer span.End()

	// Build WHERE clause using query builder
	qb := qbu.NewQueryBuilder()
	qb.AddFilters(filters)
	whereClause, args := qb.Build()

	exec := d.getExecutor(ctx)
	query := fmt.Sprintf(queries.CountProjectsQuery, d.tableName())
	if whereClause != "" {
		query = fmt.Sprintf("%s %s", query, whereClause)
	}

	var count int64
	if err := exec.GetContext(ctx, &count, query, args...); err != nil {
		return 0, apperr.DatabaseError(err, "count projects")
	}

	return count, nil
}

// ========== Business Validation Methods ==========

func (d *projectDomain) validateCreate(ctx context.Context, project *entity.Project) error {
	if project == nil {
		return apperr.InvalidInputError("project cannot be nil")
	}

	// Add your business validation rules here
	// Example:
	// if project.Name == "" {
	//     return apperr.InvalidInputError("name is required")
	// }

	return nil
}

func (d *projectDomain) validateUpdate(ctx context.Context, existing, updated *entity.Project) error {
	if updated == nil {
		return apperr.InvalidInputError("updated project cannot be nil")
	}

	// Add your business validation rules for updates
	// Example:
	// if existing.Status == "completed" && updated.Status != "completed" {
	//     return apperr.InvalidInputError("cannot modify completed project")
	// }

	return nil
}

func (d *projectDomain) validateDelete(ctx context.Context, project *entity.Project) error {
	// Add your business validation rules for deletion
	// Example: Check if has active dependencies
	// query := fmt.Sprintf("SELECT COUNT(*) FROM %srelated_table WHERE project_id = $1", d.schemaPrefix)
	// var count int64
	// exec := d.getExecutor(ctx)
	// exec.GetContext(ctx, &count, query, project.Id)
	// if count > 0 {
	//     return apperr.DependencyError(nil, "cannot delete project with active dependencies")
	// }

	return nil
}

// ========== Cache Methods ==========

func (d *projectDomain) cacheGet(ctx context.Context, id string) (*entity.Project, error) {
	key := d.cacheKey(id)
	data, err := d.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var project entity.Project
	if err := json.Unmarshal([]byte(data), &project); err != nil {
		return nil, err
	}

	return &project, nil
}

func (d *projectDomain) cacheSet(ctx context.Context, project *entity.Project) {
	key := d.cacheKey(project.Id)
	data, _ := json.Marshal(project)
	d.redis.Set(ctx, key, data, 1*time.Hour)
}

func (d *projectDomain) cacheDelete(ctx context.Context, id string) {
	key := d.cacheKey(id)
	d.redis.Del(ctx, key)
}

func (d *projectDomain) cacheKey(id string) string {
	return fmt.Sprintf("project.projects:%v", id)

}

// ========== Related Table Name Helpers ==========

// UserTableName returns full table name for users with schema
func (d *projectDomain) UserTableName() string {
	if d.schemaPrefix != "" {
		return d.schemaPrefix + "users"
	}
	return "users"
}

// ========== Relation Loader Methods ==========

// GetByIDWithUser retrieves Project with user relation loaded via JOIN
func (d *projectDomain) GetByIDWithUser(ctx context.Context, id string) (*entity.Project, error) {

	ctx, span := d.tracer.Start(ctx, "domain.Project.GetByIDWithUser")
	defer span.End()

	exec := d.getExecutor(ctx)

	// Query with JOIN
	query := fmt.Sprintf(queries.GetByIDProjectWithUserQuery, d.tableName(), d.UserTableName())

	// Use nested struct for sqlx scanning
	var result struct {
		entity.Project
		User entity.User `db:"user"`
	}

	err := exec.GetContext(ctx, &result, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFoundError("project")
		}
		return nil, apperr.DatabaseError(err, "get project with user")
	}

	// Attach the relation
	result.Project.User = &result.User

	return &result.Project, nil
}

// LoadUser loads the user relation for an existing entity
func (d *projectDomain) LoadUser(ctx context.Context, project *entity.Project) error {
	if project == nil || project.UserId == "" {
		return nil
	}

	exec := d.getExecutor(ctx)
	query := fmt.Sprintf(queries.GetByIDUserQuery, d.UserTableName())

	var user entity.User
	err := exec.GetContext(ctx, &user, query, project.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // Relation not found, but not an error
		}
		return apperr.DatabaseError(err, "load user")
	}

	project.User = &user
	return nil
}
