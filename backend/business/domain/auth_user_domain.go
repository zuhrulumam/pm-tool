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
	"github.com/zuhrulumam/pm-tool/pkg/middleware"
	oauthhelper "github.com/zuhrulumam/pm-tool/pkg/oauth"
	qbu "github.com/zuhrulumam/pm-tool/pkg/query_builder"
	"github.com/zuhrulumam/pm-tool/pkg/transaction"
)

type UserDomainItf interface {
	Create(ctx context.Context, user *entity.User) error
	GetByID(ctx context.Context, id string) (*entity.User, error)
	List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*entity.User, int64, error)
	Update(ctx context.Context, id string, user *entity.User) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context, filters map[string]interface{}) (int64, error)
	OAuthLoginWithCredential(ctx context.Context, credential string) (*entity.User, string, error)

	GetGoogleLoginURL(ctx context.Context, state string) string
}

// UserDomain handles business logic for User
// Schema: auth.
// Table: auth.users
type userDomain struct {
	db    *db.DB // Leader for writes, Follower for reads
	redis *redis.Client

	http         *httpclient.Client
	tracer       trace.Tracer
	schemaPrefix string
	cfg          *config.Config
	oauth        oauthhelper.Oauth
}

// NewUserDomain creates a new UserDomain instance with all dependencies
func NewUserDomain(
	db *db.DB,
	conf *config.Config,
	redisClient *redis.Client,

	httpClient *httpclient.Client,
	tracer trace.Tracer,
	schemaPrefix string,
	oauthelp oauthhelper.Oauth,
) UserDomainItf {
	return &userDomain{
		db:    db,
		redis: redisClient,

		http:         httpClient,
		tracer:       tracer,
		schemaPrefix: schemaPrefix,
		cfg:          conf,
		oauth:        oauthelp,
	}
}

// tableName returns full table name with schema
func (d *userDomain) tableName() string {
	return "auth.users"
}

// getExecutor returns appropriate executor based on context
// If transaction exists in context, use that
// Otherwise, use db directly
func (d *userDomain) getExecutor(ctx context.Context) db.Executor {
	if tx := transaction.GetTxFromContext(ctx); tx != nil {
		return tx
	}
	return d.db
}

// Create creates a new User with full business logic
func (d *userDomain) Create(ctx context.Context, user *entity.User) error {
	ctx, span := d.tracer.Start(ctx, "domain.User.Create")
	defer span.End()

	// 1. Business validation
	if err := d.validateCreate(ctx, user); err != nil {
		return apperr.Propagate(err, apperr.CodeValidationError, "validation failed", 400)
	}

	// 2. Check uniqueness via cache
	if err := d.checkUniqueEmail(ctx, user.Email); err != nil {
		return err // Already wrapped with proper error
	}
	if err := d.checkUniqueGoogleId(ctx, *user.GoogleId); err != nil {
		return err // Already wrapped with proper error
	}

	// 3. Get executor (transaction from context or db)
	exec := d.getExecutor(ctx)

	// 4. Execute insert
	query := fmt.Sprintf(queries.CreateUserQuery, d.tableName())
	user.Id = uuid.NewString()
	_, err := exec.ExecContext(ctx, query,
		user.Id,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.GoogleId,
		user.AvatarUrl,
		user.EmailVerified,
		user.IsActive,
		user.LastLoginAt,
	)
	if err != nil {
		return apperr.DatabaseError(err, "create user")
	}

	// 5. Update cache
	d.cacheSet(ctx, user)
	d.cacheSetUniqueEmail(ctx, user.Email, user.Id)
	d.cacheSetUniqueGoogleId(ctx, *user.GoogleId, user.Id)

	return nil
}

// GetByID retrieves a User by ID with cache support
func (d *userDomain) GetByID(ctx context.Context, id string) (*entity.User, error) {
	ctx, span := d.tracer.Start(ctx, "domain.User.GetByID")
	defer span.End()

	// 1. Try cache first
	if user, err := d.cacheGet(ctx, id); err == nil && user != nil {
		return user, nil
	}

	// 2. Get executor (transaction from context or db)
	exec := d.getExecutor(ctx)

	// 3. Query database
	query := fmt.Sprintf(queries.GetByIDUserQuery, d.tableName())
	var user entity.User
	err := exec.GetContext(ctx, &user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperr.NotFoundError("user")
		}
		return nil, apperr.DatabaseError(err, "get user")
	}

	// 4. Update cache
	d.cacheSet(ctx, &user)

	return &user, nil
}

// List retrieves paginated Users with filters
func (d *userDomain) List(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*entity.User, int64, error) {
	ctx, span := d.tracer.Start(ctx, "domain.User.List")
	defer span.End()

	// Build WHERE clause using query builder
	qb := qbu.NewQueryBuilder()
	qb.AddFilters(filters)
	whereClause, args := qb.Build()

	// Get executor
	exec := d.getExecutor(ctx)

	// Count total
	countQuery := fmt.Sprintf(queries.CountUsersQuery, d.tableName())
	if whereClause != "" {
		countQuery = fmt.Sprintf("%s %s", countQuery, whereClause)
	}

	var total int64
	if err := exec.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, apperr.DatabaseError(err, "count users")
	}

	listQuery := fmt.Sprintf(queries.ListUsersQuery, d.tableName())
	if whereClause != "" {
		listQuery = fmt.Sprintf("%s %s", listQuery, whereClause)
	}

	// Add pagination
	offset := (page - 1) * pageSize
	query := fmt.Sprintf("%s ORDER BY id DESC LIMIT ? OFFSET ?", listQuery)
	args = append(args, pageSize, offset)

	var entities []*entity.User
	if err := exec.SelectContext(ctx, &entities, query, args...); err != nil {
		fmt.Println("sini gan")
		return nil, 0, apperr.DatabaseError(err, "list users")
	}

	return entities, total, nil
}

// Update updates an existing User with full business logic
func (d *userDomain) Update(ctx context.Context, id string, user *entity.User) error {
	ctx, span := d.tracer.Start(ctx, "domain.User.Update")
	defer span.End()

	// 1. Get existing entity
	existing, err := d.GetByID(ctx, id)
	if err != nil {
		return err // Already wrapped
	}

	// 2. Business validation
	if err := d.validateUpdate(ctx, existing, user); err != nil {
		return apperr.Propagate(err, apperr.CodeValidationError, "validation failed", 400)
	}

	// 3. Check if unique fields changed
	if existing.Email != user.Email {
		if err := d.checkUniqueEmail(ctx, user.Email); err != nil {
			return err // Already wrapped
		}
	}

	// 4. Get executor (transaction from context or db)
	exec := d.getExecutor(ctx)

	// 5. Execute update
	query := fmt.Sprintf(queries.UpdateUserQuery, d.tableName())
	_, err = exec.ExecContext(ctx, query,
		user.Id,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.GoogleId,
		user.AvatarUrl,
		user.EmailVerified,
		user.IsActive,
		user.LastLoginAt,
		id,
	)
	if err != nil {
		return apperr.DatabaseError(err, "update user")
	}

	// 6. Invalidate cache
	d.cacheDelete(ctx, id)
	if existing.Email != user.Email {
		d.cacheDeleteUniqueEmail(ctx, existing.Email)
		d.cacheSetUniqueEmail(ctx, user.Email, id)
	}
	if existing.GoogleId != user.GoogleId {
		d.cacheDeleteUniqueGoogleId(ctx, *existing.GoogleId)
		d.cacheSetUniqueGoogleId(ctx, *user.GoogleId, id)
	}

	return nil
}

// Delete deletes a User with business logic validation (soft delete)
func (d *userDomain) Delete(ctx context.Context, id string) error {
	ctx, span := d.tracer.Start(ctx, "domain.User.Delete")
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
	query := fmt.Sprintf(queries.DeleteUserQuery, d.tableName())
	_, err = exec.ExecContext(ctx, query, id)
	if err != nil {
		return apperr.DatabaseError(err, "delete user")
	}

	// 5. Invalidate all related cache
	d.cacheDelete(ctx, id)
	d.cacheDeleteUniqueEmail(ctx, existing.Email)
	d.cacheDeleteUniqueGoogleId(ctx, *existing.GoogleId)

	return nil
}

// Count returns the total count of Users
func (d *userDomain) Count(ctx context.Context, filters map[string]interface{}) (int64, error) {
	ctx, span := d.tracer.Start(ctx, "domain.User.Count")
	defer span.End()

	// Build WHERE clause using query builder
	qb := qbu.NewQueryBuilder()
	qb.AddFilters(filters)
	whereClause, args := qb.Build()

	exec := d.getExecutor(ctx)
	query := fmt.Sprintf(queries.CountUsersQuery, d.tableName())
	if whereClause != "" {
		query = fmt.Sprintf("%s %s", query, whereClause)
	}

	var count int64
	if err := exec.GetContext(ctx, &count, query, args...); err != nil {
		return 0, apperr.DatabaseError(err, "count users")
	}

	return count, nil
}

// ========== Business Validation Methods ==========

func (d *userDomain) validateCreate(ctx context.Context, user *entity.User) error {
	if user == nil {
		return apperr.InvalidInputError("user cannot be nil")
	}

	// Add your business validation rules here
	// Example:
	// if user.Name == "" {
	//     return apperr.InvalidInputError("name is required")
	// }

	return nil
}

func (d *userDomain) validateUpdate(ctx context.Context, existing, updated *entity.User) error {
	if updated == nil {
		return apperr.InvalidInputError("updated user cannot be nil")
	}

	// Add your business validation rules for updates
	// Example:
	// if existing.Status == "completed" && updated.Status != "completed" {
	//     return apperr.InvalidInputError("cannot modify completed user")
	// }

	return nil
}

func (d *userDomain) validateDelete(ctx context.Context, user *entity.User) error {
	// Add your business validation rules for deletion
	// Example: Check if has active dependencies
	// query := fmt.Sprintf("SELECT COUNT(*) FROM %srelated_table WHERE user_id = $1", d.schemaPrefix)
	// var count int64
	// exec := d.getExecutor(ctx)
	// exec.GetContext(ctx, &count, query, user.Id)
	// if count > 0 {
	//     return apperr.DependencyError(nil, "cannot delete user with active dependencies")
	// }

	return nil
}

// ========== Cache Methods ==========

func (d *userDomain) cacheGet(ctx context.Context, id string) (*entity.User, error) {
	key := d.cacheKey(id)
	data, err := d.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var user entity.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (d *userDomain) cacheSet(ctx context.Context, user *entity.User) {
	key := d.cacheKey(user.Id)
	data, _ := json.Marshal(user)
	d.redis.Set(ctx, key, data, 1*time.Hour)
}

func (d *userDomain) cacheDelete(ctx context.Context, id string) {
	key := d.cacheKey(id)
	d.redis.Del(ctx, key)
}

func (d *userDomain) cacheKey(id string) string {
	return fmt.Sprintf("auth.users:%v", id)

}

func (d *userDomain) checkUniqueEmail(ctx context.Context, email string) error {
	key := fmt.Sprintf("auth.users:email:%v", email)

	exists, _ := d.redis.Get(ctx, key).Result()
	if exists != "" {
		return apperr.AlreadyExistsError("email")
	}
	return nil
}

func (d *userDomain) cacheSetUniqueEmail(ctx context.Context, email string, id string) {
	key := fmt.Sprintf("auth.users:email:%v", email)

	d.redis.Set(ctx, key, fmt.Sprintf("%v", id), 1*time.Hour)
}

func (d *userDomain) cacheDeleteUniqueEmail(ctx context.Context, email string) {
	key := fmt.Sprintf("auth.users:email:%v", email)

	d.redis.Del(ctx, key)
}

func (d *userDomain) checkUniqueGoogleId(ctx context.Context, google_id string) error {
	key := fmt.Sprintf("auth.users:google_id:%v", google_id)

	exists, _ := d.redis.Get(ctx, key).Result()
	if exists != "" {
		return apperr.AlreadyExistsError("google_id")
	}
	return nil
}

func (d *userDomain) cacheSetUniqueGoogleId(ctx context.Context, google_id string, id string) {
	key := fmt.Sprintf("auth.users:google_id:%v", google_id)

	d.redis.Set(ctx, key, fmt.Sprintf("%v", id), 1*time.Hour)
}

func (d *userDomain) cacheDeleteUniqueGoogleId(ctx context.Context, google_id string) {
	key := fmt.Sprintf("auth.users:google_id:%v", google_id)

	d.redis.Del(ctx, key)
}

func (d *userDomain) OAuthLoginWithCredential(ctx context.Context, credential string) (*entity.User, string, error) {
	ctx, span := d.tracer.Start(ctx, "domain.User.OAuthLoginWithCredential")
	defer span.End()

	// Use the new VerifyIDToken method
	userInfo, err := d.oauth.VerifyIDToken(ctx, credential)
	if err != nil {
		return nil, "", apperr.Propagate(err, apperr.CodeExternalAPI, "oauth verify token failed", 401)
	}

	// Get or create user
	filters := map[string]interface{}{
		"google_id": userInfo.ID,
	}

	users, total, err := d.List(ctx, filters, 1, 1)
	if err != nil {
		return nil, "", apperr.DatabaseError(err, "get user by google id")
	}

	now := time.Now()
	var user *entity.User

	if total == 0 || len(users) == 0 {
		// Create new user
		user = &entity.User{
			Id:            uuid.NewString(),
			Email:         userInfo.Email,
			Name:          &userInfo.Name,
			AvatarUrl:     &userInfo.Picture,
			GoogleId:      &userInfo.ID,
			EmailVerified: true, // Google tokens are always verified
			IsActive:      true,
			LastLoginAt:   &now,
		}
		if err := d.Create(ctx, user); err != nil {
			return nil, "", err
		}
	} else {
		// Update existing user
		user = users[0]
		user.Name = &userInfo.Name
		user.AvatarUrl = &userInfo.Picture
		user.Email = userInfo.Email
		user.LastLoginAt = &now
		user.EmailVerified = true
		if err := d.Update(ctx, user.Id, user); err != nil {
			fmt.Println("sini gan", err)
			return nil, "", err
		}
	}

	// Generate JWT
	token, err := middleware.GenerateToken(user.Id, user.Email, d.cfg.JWT.Secret, d.cfg.JWT.Expiration)
	if err != nil {
		return nil, "", apperr.InternalError(err, "create jwt failed")
	}

	return user, token, nil
}

func (d *userDomain) GetGoogleLoginURL(ctx context.Context, state string) string {
	return d.oauth.GetLoginUrl(ctx, oauthhelper.GetUserInfoReq{
		State: state,
	})
}
