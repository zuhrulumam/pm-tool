package postgres

import (
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"

	"github.com/zuhrulumam/pm-tool/pkg/db"
)

// Config holds PostgreSQL configuration
type Config struct {
	Host     string
	Port     int
	Database string
	User string
	Password string
	SSLMode  string
	MaxOpenConns  int
	MaxIdleConns  int
	MaxLifetime  time.Duration
	ServiceName string
	
}

// DefaultConfig returns default PostgreSQL configuration
func DefaultConfig() Config {
	return Config{
		Host:     "localhost",
		Port:     5432,
		Database: "mydb",
		User: "postgres",
		Password: "postgres",
		SSLMode:  "disable",
		MaxOpenConns:  25,
		MaxIdleConns:  5,
		MaxLifetime:  300,
		ServiceName: "postgres",
		
	}
}

// NewConnection creates a new PostgreSQL database connection
func NewConnection(cfg Config) (*db.DB, error) {
	// Build connection string
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)

	// Connect to database
	sqlxDB, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Configure connection pool
	sqlxDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlxDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlxDB.SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)

	// Verify connection
	if err := sqlxDB.Ping(); err != nil {
		sqlxDB.Close()
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	// Wrap in unified DB with optional telemetry
	
	return db.NewDB(sqlxDB, true, cfg.ServiceName), nil
	
}

// MustConnect connects to PostgreSQL and panics on error
func MustConnect(cfg Config) *db.DB {
	database, err := NewConnection(cfg)
	if err != nil {
		panic(err)
	}
	return database
}
