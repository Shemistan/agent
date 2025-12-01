package service

import "context"

// HealthService defines the interface for health check operations
type HealthService interface {
	HandleHealth(ctx context.Context) error
}

// ManagerCheckResult represents the result of a single manager health check
type ManagerCheckResult struct {
	ManagerURL   string
	Status       string // "success" or "error"
	HTTPStatus   int
	ErrorMessage string
}

// ManagerCheckResults represents results from checking multiple managers
type ManagerCheckResults struct {
	Results []ManagerCheckResult
}

// ManagerCheckService defines the interface for manager check operations
type ManagerCheckService interface {
	CheckManager(ctx context.Context) (ManagerCheckResults, error)
}
