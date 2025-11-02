package usecase

import (
	"github.com/zuhrulumam/pm-tool/business/domain"
	"github.com/zuhrulumam/pm-tool/pkg/transaction"
	"go.opentelemetry.io/otel/trace"
	"github.com/zuhrulumam/pm-tool/config"
)

// Usecases holds all use case instances
type Usecases struct {
	User UserUsecaseItf
	Project ProjectUsecaseItf
	Note NoteUsecaseItf
}

// UsecaseDependencies contains dependencies needed to initialize use cases
type UsecaseDependencies struct {
	Domains *domain.Domains
	TxMgr   transaction.TransactionManager
	Config   *config.Config
	Tracer  trace.Tracer
}

// NewUsecases creates and initializes all use cases
func NewUsecases(deps UsecaseDependencies) *Usecases {
	return &Usecases{
		User: NewUserUsecase(deps.Domains.User, deps.TxMgr, deps.Config, deps.Tracer),
		Project: NewProjectUsecase(deps.Domains.Project, deps.TxMgr, deps.Config, deps.Tracer),
		Note: NewNoteUsecase(deps.Domains.Note, deps.TxMgr, deps.Config, deps.Tracer),
	}
}
