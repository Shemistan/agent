package service

import "context"

// HealthService defines the interface for health check operations
type HealthService interface {
	HandleHealth(ctx context.Context) error
}

// ManagerCheckResult represents the result of a manager health check
type ManagerCheckResult struct {
	ManagerURL   string
	Status       string // "success" или "error"
	HTTPStatus   int
	ErrorMessage string
}

// ManagerCheckService defines the interface for manager check operations
type ManagerCheckService interface {
	CheckManager(ctx context.Context) (ManagerCheckResult, error)
}
