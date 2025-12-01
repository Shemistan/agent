package migrator

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/Shemistan/agent/internal/config"
)

// Run runs the database migrations
func Run(configPath, migrationDir string) error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	logger := initLogger(cfg.ServiceEnv)
	logger.Info("Starting migrator", slog.String("service_name", cfg.ServiceName))

	// Connect to PostgreSQL
	dsn := cfg.GetDSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection with retries
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to database")

	// Run migrations
	if err := runMigrations(db, migrationDir, logger); err != nil {
		return err
	}

	logger.Info("Migrations completed successfully")
	return nil
}

// runMigrations executes all SQL migrations in order
func runMigrations(db *sql.DB, migrationDir string, logger *slog.Logger) error {
	// Read migration directory
	files, err := ioutil.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	// Filter and sort .sql files
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}

	if len(sqlFiles) == 0 {
		logger.Info("No migration files found")
		return nil
	}

	sort.Strings(sqlFiles)

	// Execute each migration file
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, sqlFile := range sqlFiles {
		filePath := filepath.Join(migrationDir, sqlFile)
		logger.Info("Running migration", slog.String("file", sqlFile))

		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", sqlFile, err)
		}

		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", sqlFile, err)
		}

		logger.Info("Migration completed", slog.String("file", sqlFile))
	}

	return nil
}

// initLogger initializes the logger based on the environment
func initLogger(env string) *slog.Logger {
	var level slog.Level
	switch env {
	case "prod":
		level = slog.LevelInfo
	default:
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}
