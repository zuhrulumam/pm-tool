package usecase

import (
	"context"
	
	"go.opentelemetry.io/otel/trace"
	"github.com/zuhrulumam/pm-tool/pkg/errors"
	
	"github.com/zuhrulumam/pm-tool/business/domain"
	"github.com/zuhrulumam/pm-tool/config"
	"github.com/zuhrulumam/pm-tool/business/entity"
	"github.com/zuhrulumam/pm-tool/pkg/transaction"
)

type UserUsecaseItf interface {
	CreateUser(ctx context.Context, entity *entity.User) error
	GetUser(ctx context.Context, id string) (*entity.User, error)
	ListUsers(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*entity.User, int64, error)
	UpdateUser(ctx context.Context, id string, entity *entity.User) error
	DeleteUser(ctx context.Context, id string) error
	CountUsers(ctx context.Context, filters map[string]interface{}) (int64, error)
	CreateUserBatch(ctx context.Context, entities []*entity.User) error
}

// UserUsecase handles business logic and transaction orchestration for User
type userUsecase struct {
	userDomain domain.UserDomainItf
	txMgr   transaction.TransactionManager
	tracer trace.Tracer
	Config *config.Config
}

// NewUserUsecase creates a new UserUsecase instance
func NewUserUsecase(
	userDomain domain.UserDomainItf,
	txMgr transaction.TransactionManager,
	conf *config.Config,
	tracer trace.Tracer,
) UserUsecaseItf {
	return &userUsecase{
		userDomain: userDomain,
		txMgr:   txMgr,
		Config: conf,
		tracer: tracer,
	}
}

// CreateUser creates a new User within a transaction
func (uc *userUsecase) CreateUser(ctx context.Context, entity *entity.User) error {
	ctx, span := uc.tracer.Start(ctx, "usecase.CreateUser")
	defer span.End()
	
	
	// Execute within transaction
	return uc.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		// Domain layer handles: DB operations, Redis cache, Queue events, HTTP calls
		// All operations will be aware of the transaction context
		return uc.userDomain.Create(txCtx, entity)
	})
}

// GetUser retrieves a User by ID (no transaction needed for read)
func (uc *userUsecase) GetUser(ctx context.Context, id string) (*entity.User, error) {
	ctx, span := uc.tracer.Start(ctx, "usecase.GetUser")
	defer span.End()
	
	
	return uc.userDomain.GetByID(ctx, id)
}

// ListUsers retrieves paginated Users (no transaction needed for read)
func (uc *userUsecase) ListUsers(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*entity.User, int64, error) {
	ctx, span := uc.tracer.Start(ctx, "usecase.ListUsers")
	defer span.End()
	
	
	return uc.userDomain.List(ctx, filters, page, pageSize)
}

// UpdateUser updates an existing User within a transaction
func (uc *userUsecase) UpdateUser(ctx context.Context, id string, entity *entity.User) error {
	ctx, span := uc.tracer.Start(ctx, "usecase.UpdateUser")
	defer span.End()
	
	
	// Execute within transaction
	return uc.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		return uc.userDomain.Update(txCtx, id, entity)
	})
}

// DeleteUser deletes a User within a transaction
func (uc *userUsecase) DeleteUser(ctx context.Context, id string) error {
	ctx, span := uc.tracer.Start(ctx, "usecase.DeleteUser")
	defer span.End()
	
	
	// Execute within transaction
	return uc.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		return uc.userDomain.Delete(txCtx, id)
	})
}

// CountUsers returns the total count of Users (no transaction needed)
func (uc *userUsecase) CountUsers(ctx context.Context, filters map[string]interface{}) (int64, error) {
	ctx, span := uc.tracer.Start(ctx, "usecase.CountUsers")
	defer span.End()
	
	
	return uc.userDomain.Count(ctx, filters)
}

// CreateUserBatch creates multiple Users in a single transaction
// This is an example of cross-operation transaction orchestration
func (uc *userUsecase) CreateUserBatch(ctx context.Context, entities []*entity.User) error {
	ctx, span := uc.tracer.Start(ctx, "usecase.CreateUserBatch")
	defer span.End()
	
	
	if len(entities) == 0 {
		return errors.BadRequest("no entities to create")
	}
	
	// All creates in single transaction
	return uc.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		for _, entity := range entities {
			if err := uc.userDomain.Create(txCtx, entity); err != nil {
				// Transaction will rollback automatically on error
				return errors.Internal(err)
			}
		}
		return nil
	})
}
