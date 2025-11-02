package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	
)

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
}

// DB wraps sqlx.DB with unified interface and optional telemetry
type DB struct {
	db *sqlx.DB
	driverName  string
	metrics     *DBMetrics
	tracer      trace.Tracer
	serviceName string
	
}

// Config holds database configuration
type Config struct {
	Driver   string
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string
	MaxOpen  int
	MaxIdle  int
	MaxLife  int
	ServiceName string
	
}

// NewDB creates a new DB wrapper
func NewDB(sqlxDB *sqlx.DB, hasTelemetry bool, serviceName string) *DB {
	db := &DB{
		db: sqlxDB,
		driverName: sqlxDB.DriverName(),
	}

	
	if hasTelemetry {
		if serviceName == "" {
			serviceName = "database"
		}
		db.metrics = NewDBMetrics(serviceName)
		db.tracer = otel.Tracer("database")
		db.serviceName = serviceName
	}
	

	return db
}

// GetDB returns the underlying *sqlx.DB for advanced operations
func (d *DB) GetDB() *sqlx.DB {
	return d.db
}

// Close closes the database connection
func (d *DB) Close() error {
	return d.db.Close()
}

// Ping verifies connection to database
func (d *DB) Ping() error {
	return d.db.Ping()
}

// PingContext verifies connection to database with context
func (d *DB) PingContext(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

// ExecContext executes a query without returning any rows
func (d *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	query = d.Rebind(query)
	
	if d.metrics != nil {
		return d.execWithTelemetry(ctx, query, args...)
	}
	
	return d.db.ExecContext(ctx, query, args...)
}

// QueryContext executes a query that returns rows
func (d *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	query = d.Rebind(query)
	
	if d.metrics != nil {
		return d.queryWithTelemetry(ctx, query, args...)
	}
	
	return d.db.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that returns at most one row
func (d *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	query = d.Rebind(query)
	
	if d.metrics != nil {
		return d.queryRowWithTelemetry(ctx, query, args...)
	}
	
	return d.db.QueryRowContext(ctx, query, args...)
}

// GetContext gets a single record and scans it into dest
func (d *DB) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	query = d.Rebind(query)
	
	if d.metrics != nil {
		return d.getWithTelemetry(ctx, dest, query, args...)
	}
	
	return d.db.GetContext(ctx, dest, query, args...)
}

// SelectContext gets multiple records and scans them into dest
func (d *DB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	query = d.Rebind(query)
	
	if d.metrics != nil {
		return d.selectWithTelemetry(ctx, dest, query, args...)
	}
	
	return d.db.SelectContext(ctx, dest, query, args...)
}

// NamedExecContext executes a named query without returning rows
func (d *DB) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	query = d.Rebind(query)
	
	if d.metrics != nil {
		return d.namedExecWithTelemetry(ctx, query, arg)
	}
	
	return d.db.NamedExecContext(ctx, query, arg)
}

// NamedQueryContext executes a named query that returns rows
func (d *DB) NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	query = d.Rebind(query)
	
	if d.metrics != nil {
		return d.namedQueryWithTelemetry(ctx, query, arg)
	}
	
	return d.db.NamedQueryContext(ctx, query, arg)
}

// BeginTxx begins a transaction with options
func (d *DB) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error) {
	
	if d.metrics != nil {
		start := time.Now()
		tx, err := d.db.BeginTxx(ctx, opts)
		duration := time.Since(start).Seconds()
		
		if err != nil {
			d.metrics.RecordQuery("begin_tx", "error", duration)
		} else {
			d.metrics.RecordQuery("begin_tx", "success", duration)
		}
		
		return tx, err
	}
	
	return d.db.BeginTxx(ctx, opts)
}

// Preparex prepares a statement
func (d *DB) Preparex(query string) (*sqlx.Stmt, error) {
	return d.db.Preparex(query)
}

// PreparexContext prepares a statement with context
func (d *DB) PreparexContext(ctx context.Context, query string) (*sqlx.Stmt, error) {
	return d.db.PreparexContext(ctx, query)
}

// PrepareNamed prepares a named statement
func (d *DB) PrepareNamed(query string) (*sqlx.NamedStmt, error) {
	return d.db.PrepareNamed(query)
}

// PrepareNamedContext prepares a named statement with context
func (d *DB) PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error) {
	return d.db.PrepareNamedContext(ctx, query)
}


// execWithTelemetry executes query with telemetry
func (d *DB) execWithTelemetry(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	ctx, span := d.tracer.Start(ctx, "db.Exec")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", d.serviceName),
		attribute.String("db.statement", sanitizeQuery(query)),
	)

	start := time.Now()
	result, err := d.db.ExecContext(ctx, query, args...)
	duration := time.Since(start).Seconds()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		d.metrics.RecordQuery("exec", "error", duration)
		d.metrics.RecordError("exec", categorizeDBError(err))
		return result, err
	}

	span.SetStatus(codes.Ok, "")
	d.metrics.RecordQuery("exec", "success", duration)

	return result, nil
}

// queryWithTelemetry queries with telemetry
func (d *DB) queryWithTelemetry(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	ctx, span := d.tracer.Start(ctx, "db.Query")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", d.serviceName),
		attribute.String("db.statement", sanitizeQuery(query)),
	)

	start := time.Now()
	rows, err := d.db.QueryContext(ctx, query, args...)
	duration := time.Since(start).Seconds()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		d.metrics.RecordQuery("query", "error", duration)
		d.metrics.RecordError("query", categorizeDBError(err))
		return rows, err
	}

	span.SetStatus(codes.Ok, "")
	d.metrics.RecordQuery("query", "success", duration)

	return rows, nil
}

// queryRowWithTelemetry queries single row with telemetry
func (d *DB) queryRowWithTelemetry(ctx context.Context, query string, args ...interface{}) *sql.Row {
	ctx, span := d.tracer.Start(ctx, "db.QueryRow")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", d.serviceName),
		attribute.String("db.statement", sanitizeQuery(query)),
	)

	start := time.Now()
	row := d.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start).Seconds()

	d.metrics.RecordQuery("query_row", "success", duration)
	span.SetStatus(codes.Ok, "")

	return row
}

// getWithTelemetry gets single record with telemetry
func (d *DB) getWithTelemetry(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	ctx, span := d.tracer.Start(ctx, "db.Get")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", d.serviceName),
		attribute.String("db.statement", sanitizeQuery(query)),
	)

	start := time.Now()
	err := d.db.GetContext(ctx, dest, query, args...)
	duration := time.Since(start).Seconds()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		d.metrics.RecordQuery("get", "error", duration)
		d.metrics.RecordError("get", categorizeDBError(err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	d.metrics.RecordQuery("get", "success", duration)

	return nil
}

// selectWithTelemetry selects multiple records with telemetry
func (d *DB) selectWithTelemetry(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	ctx, span := d.tracer.Start(ctx, "db.Select")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", d.serviceName),
		attribute.String("db.statement", sanitizeQuery(query)),
	)

	start := time.Now()
	err := d.db.SelectContext(ctx, dest, query, args...)
	duration := time.Since(start).Seconds()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		d.metrics.RecordQuery("select", "error", duration)
		d.metrics.RecordError("select", categorizeDBError(err))
		return err
	}

	span.SetStatus(codes.Ok, "")
	d.metrics.RecordQuery("select", "success", duration)

	return nil
}

// namedExecWithTelemetry executes named query with telemetry
func (d *DB) namedExecWithTelemetry(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	ctx, span := d.tracer.Start(ctx, "db.NamedExec")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", d.serviceName),
		attribute.String("db.statement", sanitizeQuery(query)),
	)

	start := time.Now()
	result, err := d.db.NamedExecContext(ctx, query, arg)
	duration := time.Since(start).Seconds()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		d.metrics.RecordQuery("named_exec", "error", duration)
		d.metrics.RecordError("named_exec", categorizeDBError(err))
		return result, err
	}

	span.SetStatus(codes.Ok, "")
	d.metrics.RecordQuery("named_exec", "success", duration)

	return result, nil
}

// namedQueryWithTelemetry queries with named params and telemetry
func (d *DB) namedQueryWithTelemetry(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error) {
	ctx, span := d.tracer.Start(ctx, "db.NamedQuery")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", d.serviceName),
		attribute.String("db.statement", sanitizeQuery(query)),
	)

	start := time.Now()
	rows, err := d.db.NamedQueryContext(ctx, query, arg)
	duration := time.Since(start).Seconds()

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		d.metrics.RecordQuery("named_query", "error", duration)
		d.metrics.RecordError("named_query", categorizeDBError(err))
		return rows, err
	}

	span.SetStatus(codes.Ok, "")
	d.metrics.RecordQuery("named_query", "success", duration)

	return rows, nil
}

// sanitizeQuery removes sensitive data from query for tracing
func sanitizeQuery(query string) string {
	// Simple truncation - could be more sophisticated
	if len(query) > 100 {
		return query[:100] + "..."
	}
	return query
}

// categorizeDBError categorizes database errors
func categorizeDBError(err error) string {
	if err == nil {
		return "none"
	}

	errStr := err.Error()
	switch {
	case err == sql.ErrNoRows:
		return "no_rows"
	case err == sql.ErrTxDone:
		return "tx_done"
	case err == sql.ErrConnDone:
		return "conn_done"
	case contains(errStr, "connection refused"):
		return "connection_refused"
	case contains(errStr, "connection reset"):
		return "connection_reset"
	case contains(errStr, "timeout"):
		return "timeout"
	case contains(errStr, "duplicate"):
		return "duplicate_key"
	case contains(errStr, "constraint"):
		return "constraint_violation"
	case contains(errStr, "syntax"):
		return "syntax_error"
	default:
		return "unknown_error"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr))
}


func (d *DB) Rebind(query string) string {
	switch d.driverName {
	case "postgres", "pgx":
		return sqlx.Rebind(sqlx.DOLLAR, query)
	case "sqlite3":
		return sqlx.Rebind(sqlx.QUESTION, query)
	case "mysql":
		return sqlx.Rebind(sqlx.QUESTION, query)
	default:
		// Default to question mark if unknown
		return sqlx.Rebind(sqlx.QUESTION, query)
	}
}
