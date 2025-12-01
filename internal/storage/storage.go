package storage

import (
	"context"
	"time"
)

// HealthStorage defines the interface for health call storage operations
type HealthStorage interface {
	SaveHealthCall(ctx context.Context, calledAt time.Time) error
}

// ManagerCheck represents a manager health check record
type ManagerCheck struct {
	ID           int64
	CheckedAt    time.Time
	ManagerURL   string
	Status       string // "success" или "error"
	HTTPStatus   *int
	ErrorMessage *string
}

// ManagerCheckStorage defines the interface for manager check storage operations
type ManagerCheckStorage interface {
	SaveManagerCheck(ctx context.Context, check ManagerCheck) error
}
