package transaction

import "errors"

var (
	// ErrNoTransaction is returned when a transaction is required but not found
	ErrNoTransaction = errors.New("no transaction in context")
	
	// ErrTransactionCancelled is returned when transaction context is cancelled
	ErrTransactionCancelled = errors.New("transaction cancelled")
	
	// ErrTransactionTimeout is returned when transaction exceeds timeout
	ErrTransactionTimeout = errors.New("transaction timeout")
	
	// ErrCommitFailed is returned when transaction commit fails
	ErrCommitFailed = errors.New("transaction commit failed")
	
	// ErrRollbackFailed is returned when transaction rollback fails
	ErrRollbackFailed = errors.New("transaction rollback failed")
)
