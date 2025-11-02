package transaction

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/zuhrulumam/pm-tool/pkg/db"
)

// Manager implements TransactionManager for sqlx database
type Manager struct {
	db             *db.DB
	defaultTimeout time.Duration
}

// NewManager creates a new transaction manager
func NewManager(db *db.DB) TransactionManager {
	return &Manager{
		db:             db,
		defaultTimeout: 30 * time.Second, // Default transaction timeout
	}
}

// WithDefaultTimeout sets the default transaction timeout
func (m *Manager) WithDefaultTimeout(timeout time.Duration) *Manager {
	m.defaultTimeout = timeout
	return m
}

// WithTransaction executes a function within a database transaction with default timeout
func (m *Manager) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return m.WithTransactionTimeout(ctx, m.defaultTimeout, fn)
}

// WithTransactionTimeout executes a function within a database transaction with custom timeout
func (m *Manager) WithTransactionTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled before transaction start")
	}

	// Create transaction context with timeout
	txCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Start transaction with context
	tx, err := m.db.BeginTxx(txCtx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}

	// Set transaction in context
	txCtx = SetTxToContext(txCtx, tx)

	// Track if transaction was committed
	var committed bool
	
	// Defer rollback with proper error handling
	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic
			if rbErr := tx.Rollback(); rbErr != nil {
				// Log rollback error but re-panic with original panic
				// In production, you should log this properly
				fmt.Printf("failed to rollback transaction after panic: %v\n", rbErr)
			}
			panic(p) // Re-throw panic after rollback
		}
		
		// Only rollback if not committed and transaction is still open
		if !committed {
			if rbErr := tx.Rollback(); rbErr != nil && rbErr != sql.ErrTxDone {
				// Log rollback error
				fmt.Printf("failed to rollback transaction: %v\n", rbErr)
			}
		}
	}()

	// Execute function
	if err := fn(txCtx); err != nil {
		// Function returned error, rollback will happen in defer
		return errors.Wrap(err, "transaction function failed")
	}

	// Check if context was cancelled during execution
	if err := txCtx.Err(); err != nil {
		return errors.Wrap(err, "transaction cancelled during execution")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		// Commit failed, rollback will happen in defer
		return errors.Wrap(err, "failed to commit transaction")
	}

	// Mark as committed to prevent rollback in defer
	committed = true
	return nil
}

// InTransaction returns true if a transaction is active in the context
func (m *Manager) InTransaction(ctx context.Context) bool {
	return GetTxFromContext(ctx) != nil
}
