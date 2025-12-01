package agent

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	api "github.com/Shemistan/agent/internal/api/agent"
	"github.com/Shemistan/agent/internal/config"
	svc "github.com/Shemistan/agent/internal/service/agent"
	stg "github.com/Shemistan/agent/internal/storage/agent"
)

// Run initializes and starts the agent service
func Run(configPath string) error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	logger := initLogger(cfg.ServiceEnv)
	logger.Info("Starting agent service", slog.String("service_name", cfg.ServiceName))

	// Connect to PostgreSQL
	db, err := connectDB(cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()
	logger.Info("Connected to database")

	// Create HTTP client for manager service
	httpClient := &http.Client{
		Timeout: time.Duration(cfg.GetManagerTimeout()) * time.Second,
	}

	// Initialize storage layer
	storage := stg.NewStorage(db, logger)

	// Initialize service layer
	healthService := svc.NewHealthService(storage, logger)
	managerCheckService := svc.NewManagerCheckService(
		httpClient,
		storage,
		cfg.GetManagerURLs(),
		logger,
	)

	// Initialize HTTP layer
	handler := api.NewHandler(healthService, managerCheckService, logger)
	router := api.NewRouter(handler)

	// Start HTTP server
	addr := fmt.Sprintf(":%d", cfg.HTTPPort)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Info("Starting HTTP server", slog.String("address", addr))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
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

// connectDB establishes a connection to PostgreSQL
func connectDB(cfg *config.Config, logger *slog.Logger) (*sql.DB, error) {
	dsn := cfg.GetDSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}
