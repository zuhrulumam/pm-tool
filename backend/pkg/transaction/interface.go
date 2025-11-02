package transaction

import (
	"context"
	"time"
)

// TransactionManager defines the interface for managing database transactions
type TransactionManager interface {
	// WithTransaction executes the given function within a database transaction
	// If the function returns an error, the transaction is rolled back
	// If the function returns nil, the transaction is committed
	// Context cancellation will trigger rollback
	WithTransaction(ctx context.Context, fn func(context.Context) error) error
	
	// WithTransactionTimeout executes a transaction with a custom timeout
	WithTransactionTimeout(ctx context.Context, timeout time.Duration, fn func(context.Context) error) error
}
