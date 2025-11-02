package transaction

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type contextKey string

const txKey contextKey = "tx"

// SetTxToContext sets a transaction to the context
func SetTxToContext(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// GetTxFromContext retrieves a transaction from the context
// Returns nil if no transaction is found
func GetTxFromContext(ctx context.Context) *sqlx.Tx {
	if tx, ok := ctx.Value(txKey).(*sqlx.Tx); ok {
		return tx
	}
	return nil
}

// Executor returns either the transaction from context or the provided db
// This is useful for domain layer to get the right executor
func Executor(ctx context.Context, db *sqlx.DB) sqlx.ExtContext {
	if tx := GetTxFromContext(ctx); tx != nil {
		return tx
	}
	return db
}

// MustGetTxFromContext retrieves a transaction from the context
// Panics if no transaction is found - use this when you REQUIRE a transaction
func MustGetTxFromContext(ctx context.Context) *sqlx.Tx {
	tx := GetTxFromContext(ctx)
	if tx == nil {
		panic("transaction required but not found in context")
	}
	return tx
}
