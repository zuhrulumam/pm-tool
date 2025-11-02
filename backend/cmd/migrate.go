package cmd

import (
    "context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/zuhrulumam/pm-tool/config"
	"github.com/zuhrulumam/pm-tool/infra/postgres"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migration commands",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Run all pending migrations",
	Run:   runMigrateUp,
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the last migration",
	Run:   runMigrateDown,
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Run:   runMigrateStatus,
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
}

func runMigrateUp(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	dbCfg := postgres.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := postgres.NewConnection(dbCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Get migration files
	files, err := getMigrationFiles("up")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get migration files")
	}

	if len(files) == 0 {
		log.Info().Msg("No migrations to run")
		return
	}

	for _, file := range files {
		log.Info().Str("file", filepath.Base(file)).Msg("Running migration")
		sql, err := os.ReadFile(file)
		if err != nil {
			log.Fatal().Err(err).Str("file", file).Msg("Failed to read migration file")
		}

		if _, err := db.ExecContext(context.Background(),string(sql)); err != nil {
			log.Fatal().Err(err).Str("file", file).Msg("Failed to execute migration")
		}
	}

	log.Info().Int("count", len(files)).Msg("All migrations completed successfully")
}

func runMigrateDown(cmd *cobra.Command, args []string) {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	dbCfg := postgres.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	}

	db, err := postgres.NewConnection(dbCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	// Get migration files
	files, err := getMigrationFiles("down")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get migration files")
	}

	if len(files) == 0 {
		log.Info().Msg("No migrations to rollback")
		return
	}

	// Run the last migration
	file := files[len(files)-1]
	log.Info().Str("file", filepath.Base(file)).Msg("Rolling back migration")
	sql, err := os.ReadFile(file)
	if err != nil {
		log.Fatal().Err(err).Str("file", file).Msg("Failed to read migration file")
	}

	if _, err := db.ExecContext(context.Background(), string(sql)); err != nil {
		log.Fatal().Err(err).Str("file", file).Msg("Failed to execute migration")
	}

	log.Info().Msg("Migration rolled back successfully")
}

func runMigrateStatus(cmd *cobra.Command, args []string) {
	upFiles, err := getMigrationFiles("up")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get migration files")
	}

	downFiles, err := getMigrationFiles("down")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get migration files")
	}

	log.Info().
		Int("pending_up", len(upFiles)).
		Int("pending_down", len(downFiles)).
		Msg("Migration status")

	if len(upFiles) > 0 {
		fmt.Println("\nPending UP migrations:")
		for _, file := range upFiles {
			fmt.Printf("  - %s\n", filepath.Base(file))
		}
	}

	if len(downFiles) > 0 {
		fmt.Println("\nAvailable DOWN migrations:")
		for _, file := range downFiles {
			fmt.Printf("  - %s\n", filepath.Base(file))
		}
	}
}

func getMigrationFiles(direction string) ([]string, error) {
	migrationDir := "migration"
	var files []string

	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), fmt.Sprintf(".%s.sql", direction)) {
			files = append(files, filepath.Join(migrationDir, entry.Name()))
		}
	}

	sort.Strings(files)
	return files, nil
}
