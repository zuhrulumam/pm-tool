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

type NoteDomainItf interface {
	Create(ctx context.Context, note *entity.Note) error
	GetByID(ctx context.Context, id string) (*entity.Note, error)
	List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*entity.Note, int64, error)
	Update(ctx context.Context, id string, note *entity.Note) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
	// Relation loaders
	GetByIDWithProject(ctx context.Context, id string) (*entity.Note, error)
	LoadProject(ctx context.Context, note *entity.Note) error
}

// NoteDomain handles business logic for Note
// Schema: project.
// Table: project.notes
type noteDomain struct {
	db    *db.DB // Leader for writes, Follower for reads
	redis *redis.Client

	http         *httpclient.Client
	tracer       trace.Tracer
	schemaPrefix string
	cfg          *config.Config
}

// NewNoteDomain creates a new NoteDomain instance with all dependencies
func NewNoteDomain(
	db *db.DB,
	conf *config.Config,
	redisClient *redis.Client,

	httpClient *httpclient.Client,
	tracer trace.Tracer,
	schemaPrefix string,
) NoteDomainItf {
	return &noteDomain{
		db:    db,
		redis: redisClient,

		http:         httpClient,
		tracer:       tracer,
		schemaPrefix: schemaPrefix,
		cfg:          conf,
	}
}

// tableName returns full table name with schema
func (d *noteDomain) tableName() string {
	return "project.notes"
}

// getExecutor returns appropriate executor based on context
// If transaction exists in context, use that
// Otherwise, use db directly
func (d *noteDomain) getExecutor(ctx context.Context) db.Executor {
	if tx := transaction.GetTxFromContext(ctx); tx != nil {
		return tx
	}
	return d.db
}

// Create creates a new Note with full business logic
func (d *noteDomain) Create(ctx context.Context, note *entity.Note) error {
	ctx, span := d.tracer.Start(ctx, "domain.Note.Create")
	defer span.End()

	// 1. Business validation
	if err := d.validateCreate(ctx, note); err != nil {
		return apperr.Propagate(err, apperr.CodeValidationError, "validation failed", 400)
	}

	// 2. Check uniqueness via cache

	// 3. Get executor (transaction from context or db)
	exec := d.getExecutor(ctx)

	// 4. Execute insert
	query := fmt.Sprintf(queries.CreateNoteQuery, d.tableName())
	note.Id = uuid.NewString()
	_, err := exec.ExecContext(ctx, query,
		note.Id,
		note.ProjectId,
		note.Title,
		note.Content,
		note.ContentType,
		note.AudioUrl,
		note.IsPinned,
	)
	if err != nil {
		return apperr.DatabaseError(err, "create note")
	}

	// 5. Update cache
	d.cacheSet(ctx, note)

	return nil
}

// GetByID retrieves a Note by ID with cache support
func (d *noteDomain) GetByID(ctx context.Context, id string) (*entity.Note, error) {
	ctx, span := d.tracer.Start(ctx, "domain.Note.GetByID")
	defer span.End()

	// 1. Try cache first
	if note, err := d.cacheGet(ctx, id); err == nil && note != nil {
		return note, nil
	}

	// 2. Get executor (transaction from context or db)
	exec := d.getExecutor(ctx)

	// 3. Query database
	query := fmt.Sprintf(queries.GetByIDNoteQuery, d.tableName())
	var note entity.Note
	err := exec.GetContext(ctx, &note, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFoundError("note")
		}
		return nil, apperr.DatabaseError(err, "get note")
	}

	// 4. Update cache
	d.cacheSet(ctx, &note)

	return &note, nil
}

// List retrieves paginated Notes with filters
func (d *noteDomain) List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*entity.Note, int64, error) {
	ctx, span := d.tracer.Start(ctx, "domain.Note.List")
	defer span.End()

	// Build WHERE clause using query builder
	qb := qbu.NewQueryBuilder()
	qb.AddFilters(filters)
	whereClause, args := qb.Build()

	// Get executor
	exec := d.getExecutor(ctx)

	// Count total
	countQuery := fmt.Sprintf(queries.CountNotesQuery, d.tableName())
	if whereClause != "" {
		countQuery = fmt.Sprintf("%s %s", countQuery, whereClause)
	}

	var total int64
	if err := exec.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, apperr.DatabaseError(err, "count notes")
	}

	listQuery := fmt.Sprintf(queries.ListNotesQuery, d.tableName())
	if whereClause != "" {
		listQuery = fmt.Sprintf("%s %s", listQuery, whereClause)
	}
	// Add pagination
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("%s ORDER BY id DESC LIMIT ? OFFSET ?", listQuery)
	args = append(args, pageSize, offset)

	var entities []*entity.Note
	if err := exec.SelectContext(ctx, &entities, query, args...); err != nil {
		return nil, 0, apperr.DatabaseError(err, "list notes")
	}

	return entities, total, nil
}

// Update updates an existing Note with full business logic
func (d *noteDomain) Update(ctx context.Context, id string, note *entity.Note) error {
	ctx, span := d.tracer.Start(ctx, "domain.Note.Update")
	defer span.End()

	// 1. Get existing entity
	existing, err := d.GetByID(ctx, id)
	if err != nil {
		return err // Already wrapped
	}

	// 2. Business validation
	if err := d.validateUpdate(ctx, existing, note); err != nil {
		return apperr.Propagate(err, apperr.CodeValidationError, "validation failed", 400)
	}

	// 4. Get executor (transaction from context or db)
	exec := d.getExecutor(ctx)

	// 5. Execute update
	query := fmt.Sprintf(queries.UpdateNoteQuery, d.tableName())
	_, err = exec.ExecContext(ctx, query,
		note.Id,
		note.ProjectId,
		note.Title,
		note.Content,
		note.ContentType,
		note.AudioUrl,
		note.IsPinned,
		id,
	)
	if err != nil {
		return apperr.DatabaseError(err, "update note")
	}

	// 6. Invalidate cache
	d.cacheDelete(ctx, id)

	return nil
}

// Delete deletes a Note with business logic validation (soft delete)
func (d *noteDomain) Delete(ctx context.Context, id string) error {
	ctx, span := d.tracer.Start(ctx, "domain.Note.Delete")
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
	query := fmt.Sprintf(queries.DeleteNoteQuery, d.tableName())
	_, err = exec.ExecContext(ctx, query, id)
	if err != nil {
		return apperr.DatabaseError(err, "delete note")
	}

	// 5. Invalidate all related cache
	d.cacheDelete(ctx, id)

	return nil
}

// Count returns the total count of Notes
func (d *noteDomain) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	ctx, span := d.tracer.Start(ctx, "domain.Note.Count")
	defer span.End()

	// Build WHERE clause using query builder
	qb := qbu.NewQueryBuilder()
	qb.AddFilters(filters)
	whereClause, args := qb.Build()

	exec := d.getExecutor(ctx)
	query := fmt.Sprintf(queries.CountNotesQuery, d.tableName())
	if whereClause != "" {
		query = fmt.Sprintf("%s %s", query, whereClause)
	}

	var count int64
	if err := exec.GetContext(ctx, &count, query, args...); err != nil {
		return 0, apperr.DatabaseError(err, "count notes")
	}

	return count, nil
}

// ========== Business Validation Methods ==========

func (d *noteDomain) validateCreate(ctx context.Context, note *entity.Note) error {
	if note == nil {
		return apperr.InvalidInputError("note cannot be nil")
	}

	// Add your business validation rules here
	// Example:
	// if note.Name == "" {
	//     return apperr.InvalidInputError("name is required")
	// }

	return nil
}

func (d *noteDomain) validateUpdate(ctx context.Context, existing, updated *entity.Note) error {
	if updated == nil {
		return apperr.InvalidInputError("updated note cannot be nil")
	}

	// Add your business validation rules for updates
	// Example:
	// if existing.Status == "completed" && updated.Status != "completed" {
	//     return apperr.InvalidInputError("cannot modify completed note")
	// }

	return nil
}

func (d *noteDomain) validateDelete(ctx context.Context, note *entity.Note) error {
	// Add your business validation rules for deletion
	// Example: Check if has active dependencies
	// query := fmt.Sprintf("SELECT COUNT(*) FROM %srelated_table WHERE note_id = $1", d.schemaPrefix)
	// var count int64
	// exec := d.getExecutor(ctx)
	// exec.GetContext(ctx, &count, query, note.Id)
	// if count > 0 {
	//     return apperr.DependencyError(nil, "cannot delete note with active dependencies")
	// }

	return nil
}

// ========== Cache Methods ==========

func (d *noteDomain) cacheGet(ctx context.Context, id string) (*entity.Note, error) {
	key := d.cacheKey(id)
	data, err := d.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var note entity.Note
	if err := json.Unmarshal([]byte(data), &note); err != nil {
		return nil, err
	}

	return &note, nil
}

func (d *noteDomain) cacheSet(ctx context.Context, note *entity.Note) {
	key := d.cacheKey(note.Id)
	data, _ := json.Marshal(note)
	d.redis.Set(ctx, key, data, 1*time.Hour)
}

func (d *noteDomain) cacheDelete(ctx context.Context, id string) {
	key := d.cacheKey(id)
	d.redis.Del(ctx, key)
}

func (d *noteDomain) cacheKey(id string) string {
	return fmt.Sprintf("project.notes:%v", id)

}

// ========== Related Table Name Helpers ==========

// ProjectTableName returns full table name for projects with schema
func (d *noteDomain) ProjectTableName() string {
	if d.schemaPrefix != "" {
		return d.schemaPrefix + "projects"
	}
	return "projects"
}

// ========== Relation Loader Methods ==========

// GetByIDWithProject retrieves Note with project relation loaded via JOIN
func (d *noteDomain) GetByIDWithProject(ctx context.Context, id string) (*entity.Note, error) {

	ctx, span := d.tracer.Start(ctx, "domain.Note.GetByIDWithProject")
	defer span.End()

	exec := d.getExecutor(ctx)

	// Query with JOIN
	query := fmt.Sprintf(queries.GetByIDNoteWithProjectQuery, d.tableName(), d.ProjectTableName())

	// Use nested struct for sqlx scanning
	var result struct {
		entity.Note
		Project entity.Project `db:"project"`
	}

	err := exec.GetContext(ctx, &result, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFoundError("note")
		}
		return nil, apperr.DatabaseError(err, "get note with project")
	}

	// Attach the relation
	result.Note.Project = &result.Project

	return &result.Note, nil
}

// LoadProject loads the project relation for an existing entity
func (d *noteDomain) LoadProject(ctx context.Context, note *entity.Note) error {
	if note == nil || note.ProjectId == "" {
		return nil
	}

	exec := d.getExecutor(ctx)
	query := fmt.Sprintf(queries.GetByIDProjectQuery, d.ProjectTableName())

	var project entity.Project
	err := exec.GetContext(ctx, &project, query, note.ProjectId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // Relation not found, but not an error
		}
		return apperr.DatabaseError(err, "load project")
	}

	note.Project = &project
	return nil
}
