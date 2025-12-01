package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/Shemistan/agent/internal/service"
	"github.com/Shemistan/agent/internal/storage"
)

// HealthService implements the health check service
type HealthService struct {
	healthStorage storage.HealthStorage
	logger        *slog.Logger
}

// NewHealthService creates a new HealthService instance
func NewHealthService(healthStorage storage.HealthStorage, logger *slog.Logger) *HealthService {
	return &HealthService{
		healthStorage: healthStorage,
		logger:        logger,
	}
}

// HandleHealth records a health check call
func (s *HealthService) HandleHealth(ctx context.Context) error {
	return s.healthStorage.SaveHealthCall(ctx, time.Now())
}

// ManagerCheckService implements the manager check service
type ManagerCheckService struct {
	httpClient          *http.Client
	managerCheckStorage storage.ManagerCheckStorage
	managerURL          string
	logger              *slog.Logger
}

// NewManagerCheckService creates a new ManagerCheckService instance
func NewManagerCheckService(
	httpClient *http.Client,
	managerCheckStorage storage.ManagerCheckStorage,
	managerURL string,
	logger *slog.Logger,
) *ManagerCheckService {
	return &ManagerCheckService{
		httpClient:          httpClient,
		managerCheckStorage: managerCheckStorage,
		managerURL:          managerURL,
		logger:              logger,
	}
}

// healthResponse represents the expected response from manager /health
type healthResponse struct {
	Status string `json:"status"`
}

// CheckManager checks the manager service health and records the result
func (s *ManagerCheckService) CheckManager(ctx context.Context) (service.ManagerCheckResult, error) {
	result := service.ManagerCheckResult{
		ManagerURL: s.managerURL,
		Status:     "error",
	}

	url := fmt.Sprintf("%s/health", s.managerURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create request: %v", err)
		result.ErrorMessage = errMsg
		s.logger.Error("manager check: request creation failed", slog.String("error", errMsg))
		s.saveResult(ctx, result)
		return result, nil
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("HTTP request failed: %v", err)
		result.ErrorMessage = errMsg
		s.logger.Error("manager check: HTTP request failed", slog.String("error", errMsg))
		s.saveResult(ctx, result)
		return result, nil
	}
	defer resp.Body.Close()

	result.HTTPStatus = resp.StatusCode

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("unexpected HTTP status: %d", resp.StatusCode)
		result.ErrorMessage = errMsg
		s.logger.Error("manager check: unexpected status", slog.Int("status", resp.StatusCode))
		s.saveResult(ctx, result)
		return result, nil
	}

	// Parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errMsg := fmt.Sprintf("failed to read response body: %v", err)
		result.ErrorMessage = errMsg
		s.logger.Error("manager check: failed to read body", slog.String("error", errMsg))
		s.saveResult(ctx, result)
		return result, nil
	}

	var healthResp healthResponse
	if err := json.Unmarshal(body, &healthResp); err != nil {
		errMsg := fmt.Sprintf("failed to parse response: %v", err)
		result.ErrorMessage = errMsg
		s.logger.Error("manager check: failed to parse response", slog.String("error", errMsg))
		s.saveResult(ctx, result)
		return result, nil
	}

	// Check if status is "success"
	if healthResp.Status != "success" {
		errMsg := fmt.Sprintf("manager returned status: %s", healthResp.Status)
		result.ErrorMessage = errMsg
		result.Status = "error"
		s.logger.Error("manager check: manager returned error status", slog.String("status", healthResp.Status))
		s.saveResult(ctx, result)
		return result, nil
	}

	// Success case
	result.Status = "success"
	result.ErrorMessage = ""
	s.logger.Info("manager check: success")
	s.saveResult(ctx, result)
	return result, nil
}

// saveResult saves the check result to the database, logging errors without failing
func (s *ManagerCheckService) saveResult(ctx context.Context, result service.ManagerCheckResult) {
	check := storage.ManagerCheck{
		CheckedAt:    time.Now(),
		ManagerURL:   result.ManagerURL,
		Status:       result.Status,
		HTTPStatus:   nil,
		ErrorMessage: nil,
	}

	if result.HTTPStatus != 0 {
		check.HTTPStatus = &result.HTTPStatus
	}

	if result.ErrorMessage != "" {
		check.ErrorMessage = &result.ErrorMessage
	}

	if err := s.managerCheckStorage.SaveManagerCheck(ctx, check); err != nil {
		s.logger.Error("failed to save manager check result", slog.String("error", err.Error()))
	}
}
