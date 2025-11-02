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

type NoteUsecaseItf interface {
	CreateNote(ctx context.Context, entity *entity.Note) error
	GetNote(ctx context.Context, id string) (*entity.Note, error)
	ListNotes(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*entity.Note, int64, error)
	UpdateNote(ctx context.Context, id string, entity *entity.Note) error
	DeleteNote(ctx context.Context, id string) error
	CountNotes(ctx context.Context, filters map[string]interface{}) (int64, error)
	CreateNoteBatch(ctx context.Context, entities []*entity.Note) error
}

// NoteUsecase handles business logic and transaction orchestration for Note
type noteUsecase struct {
	noteDomain domain.NoteDomainItf
	txMgr   transaction.TransactionManager
	tracer trace.Tracer
	Config *config.Config
}

// NewNoteUsecase creates a new NoteUsecase instance
func NewNoteUsecase(
	noteDomain domain.NoteDomainItf,
	txMgr transaction.TransactionManager,
	conf *config.Config,
	tracer trace.Tracer,
) NoteUsecaseItf {
	return &noteUsecase{
		noteDomain: noteDomain,
		txMgr:   txMgr,
		Config: conf,
		tracer: tracer,
	}
}

// CreateNote creates a new Note within a transaction
func (uc *noteUsecase) CreateNote(ctx context.Context, entity *entity.Note) error {
	ctx, span := uc.tracer.Start(ctx, "usecase.CreateNote")
	defer span.End()
	
	
	// Execute within transaction
	return uc.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		// Domain layer handles: DB operations, Redis cache, Queue events, HTTP calls
		// All operations will be aware of the transaction context
		return uc.noteDomain.Create(txCtx, entity)
	})
}

// GetNote retrieves a Note by ID (no transaction needed for read)
func (uc *noteUsecase) GetNote(ctx context.Context, id string) (*entity.Note, error) {
	ctx, span := uc.tracer.Start(ctx, "usecase.GetNote")
	defer span.End()
	
	
	return uc.noteDomain.GetByID(ctx, id)
}

// ListNotes retrieves paginated Notes (no transaction needed for read)
func (uc *noteUsecase) ListNotes(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*entity.Note, int64, error) {
	ctx, span := uc.tracer.Start(ctx, "usecase.ListNotes")
	defer span.End()
	
	
	return uc.noteDomain.List(ctx, filters, page, pageSize)
}

// UpdateNote updates an existing Note within a transaction
func (uc *noteUsecase) UpdateNote(ctx context.Context, id string, entity *entity.Note) error {
	ctx, span := uc.tracer.Start(ctx, "usecase.UpdateNote")
	defer span.End()
	
	
	// Execute within transaction
	return uc.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		return uc.noteDomain.Update(txCtx, id, entity)
	})
}

// DeleteNote deletes a Note within a transaction
func (uc *noteUsecase) DeleteNote(ctx context.Context, id string) error {
	ctx, span := uc.tracer.Start(ctx, "usecase.DeleteNote")
	defer span.End()
	
	
	// Execute within transaction
	return uc.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		return uc.noteDomain.Delete(txCtx, id)
	})
}

// CountNotes returns the total count of Notes (no transaction needed)
func (uc *noteUsecase) CountNotes(ctx context.Context, filters map[string]interface{}) (int64, error) {
	ctx, span := uc.tracer.Start(ctx, "usecase.CountNotes")
	defer span.End()
	
	
	return uc.noteDomain.Count(ctx, filters)
}

// CreateNoteBatch creates multiple Notes in a single transaction
// This is an example of cross-operation transaction orchestration
func (uc *noteUsecase) CreateNoteBatch(ctx context.Context, entities []*entity.Note) error {
	ctx, span := uc.tracer.Start(ctx, "usecase.CreateNoteBatch")
	defer span.End()
	
	
	if len(entities) == 0 {
		return errors.BadRequest("no entities to create")
	}
	
	// All creates in single transaction
	return uc.txMgr.WithTransaction(ctx, func(txCtx context.Context) error {
		for _, entity := range entities {
			if err := uc.noteDomain.Create(txCtx, entity); err != nil {
				// Transaction will rollback automatically on error
				return errors.Internal(err)
			}
		}
		return nil
	})
}
