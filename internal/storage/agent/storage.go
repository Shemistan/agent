package agent

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/Shemistan/agent/internal/storage"
)

// Storage implements HealthStorage and ManagerCheckStorage interfaces
type Storage struct {
	db     *sql.DB
	logger *slog.Logger
}

// NewStorage creates a new Storage instance
func NewStorage(db *sql.DB, logger *slog.Logger) *Storage {
	return &Storage{
		db:     db,
		logger: logger,
	}
}

// SaveHealthCall saves a health check call to the database
func (s *Storage) SaveHealthCall(ctx context.Context, calledAt time.Time) error {
	query := `
		INSERT INTO health_calls (called_at)
		VALUES ($1)
	`
	_, err := s.db.ExecContext(ctx, query, calledAt)
	if err != nil {
		s.logger.Error("failed to save health call", slog.String("error", err.Error()))
		return fmt.Errorf("save health call: %w", err)
	}
	return nil
}

// SaveManagerCheck saves a manager health check to the database
func (s *Storage) SaveManagerCheck(ctx context.Context, check storage.ManagerCheck) error {
	query := `
		INSERT INTO manager_checks (checked_at, manager_url, status, http_status, error_message)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	var id int64
	err := s.db.QueryRowContext(
		ctx, query,
		check.CheckedAt, check.ManagerURL, check.Status, check.HTTPStatus, check.ErrorMessage,
	).Scan(&id)
	if err != nil {
		s.logger.Error("failed to save manager check", slog.String("error", err.Error()))
		return fmt.Errorf("save manager check: %w", err)
	}
	return nil
}
