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

type ProjectUsecaseItf interface {
	CreateProject(ctx context.Context, entity *entity.Project) error
	GetProject(ctx context.Context, id string) (*entity.Project, error)
	ListProjects(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*entity.Project, int64, error)
	UpdateProject(ctx context.Context, id string, entity *entity.Project) error
	DeleteProject(ctx context.Context, id string) error
	CountProjects(ctx context.Context, filters map[string]interface{}) (int64, error)
	CreateProjectBatch(ctx context.Context, entities []*entity.Project) error
}

// ProjectUsecase handles business logic and transaction orchestration for Project
type projectUsecase struct {
	projectDomain domain.ProjectDomainItf
	txMgr   transaction.TransactionManager
	tracer trace.Tracer
	Config *config.Config
}

// NewProjectUsecase creates a new ProjectUsecase instance
func NewProjectUsecase(
	projectDomain domain.ProjectDomainItf,
	txMgr transaction.TransactionManager,
	conf *config.Config,
	tracer trace.Tracer,
) ProjectUsecaseItf {
	return &projectUsecase{
		projectDomain: projectDomain,
		txMgr:   txMgr,
		Config: conf,
		tracer: tracer,
	}
}

// CreateProject creates a new Project within a transaction
func (uc *projectUsecase) CreateProject(ctx context.Context, entity *entity.Project) error {
	ctx, span := uc.tracer.Start(ctx, "usecase.CreateProject")
	defer span.End()
	
	
	// Execute within transaction
	return uc.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		// Domain layer handles: DB operations, Redis cache, Queue events, HTTP calls
		// All operations will be aware of the transaction context
		return uc.projectDomain.Create(txCtx, entity)
	})
}

// GetProject retrieves a Project by ID (no transaction needed for read)
func (uc *projectUsecase) GetProject(ctx context.Context, id string) (*entity.Project, error) {
	ctx, span := uc.tracer.Start(ctx, "usecase.GetProject")
	defer span.End()
	
	
	return uc.projectDomain.GetByID(ctx, id)
}

// ListProjects retrieves paginated Projects (no transaction needed for read)
func (uc *projectUsecase) ListProjects(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*entity.Project, int64, error) {
	ctx, span := uc.tracer.Start(ctx, "usecase.ListProjects")
	defer span.End()
	
	
	return uc.projectDomain.List(ctx, filters, page, pageSize)
}

// UpdateProject updates an existing Project within a transaction
func (uc *projectUsecase) UpdateProject(ctx context.Context, id string, entity *entity.Project) error {
	ctx, span := uc.tracer.Start(ctx, "usecase.UpdateProject")
	defer span.End()
	
	
	// Execute within transaction
	return uc.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		return uc.projectDomain.Update(txCtx, id, entity)
	})
}

// DeleteProject deletes a Project within a transaction
func (uc *projectUsecase) DeleteProject(ctx context.Context, id string) error {
	ctx, span := uc.tracer.Start(ctx, "usecase.DeleteProject")
	defer span.End()
	
	
	// Execute within transaction
	return uc.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		return uc.projectDomain.Delete(txCtx, id)
	})
}

// CountProjects returns the total count of Projects (no transaction needed)
func (uc *projectUsecase) CountProjects(ctx context.Context, filters map[string]interface{}) (int64, error) {
	ctx, span := uc.tracer.Start(ctx, "usecase.CountProjects")
	defer span.End()
	
	
	return uc.projectDomain.Count(ctx, filters)
}

// CreateProjectBatch creates multiple Projects in a single transaction
// This is an example of cross-operation transaction orchestration
func (uc *projectUsecase) CreateProjectBatch(ctx context.Context, entities []*entity.Project) error {
	ctx, span := uc.tracer.Start(ctx, "usecase.CreateProjectBatch")
	defer span.End()
	
	
	if len(entities) == 0 {
		return errors.BadRequest("no entities to create")
	}
	
	// All creates in single transaction
	return uc.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		for _, entity := range entities {
			if err := uc.projectDomain.Create(txCtx, entity); err != nil {
				// Transaction will rollback automatically on error
				return errors.Internal(err)
			}
		}
		return nil
	})
}
