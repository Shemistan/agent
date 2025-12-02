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
	managerURLs         []string
	logger              *slog.Logger
}

// NewManagerCheckService creates a new ManagerCheckService instance
func NewManagerCheckService(
	httpClient *http.Client,
	managerCheckStorage storage.ManagerCheckStorage,
	managerURLs []string,
	logger *slog.Logger,
) *ManagerCheckService {
	return &ManagerCheckService{
		httpClient:          httpClient,
		managerCheckStorage: managerCheckStorage,
		managerURLs:         managerURLs,
		logger:              logger,
	}
}

// healthResponse represents the expected response from manager /health
type healthResponse struct {
	Status string `json:"status"`
}

// CheckManager checks all configured manager services and records results
func (s *ManagerCheckService) CheckManager(ctx context.Context) (service.ManagerCheckResults, error) {
	results := service.ManagerCheckResults{
		Results: make([]service.ManagerCheckResult, 0, len(s.managerURLs)),
	}

	// Check each manager URL
	for _, managerURL := range s.managerURLs {
		result := s.checkSingleManager(ctx, managerURL)
		results.Results = append(results.Results, result)
	}

	return results, nil
}

// checkSingleManager checks a single manager service health
func (s *ManagerCheckService) checkSingleManager(ctx context.Context, managerURL string) service.ManagerCheckResult {
	result := service.ManagerCheckResult{
		ManagerURL: managerURL,
		Status:     "error",
	}

	url := fmt.Sprintf("%s/health", managerURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		errMsg := fmt.Sprintf("failed to create request: %v", err)
		result.ErrorMessage = errMsg
		s.logger.Error("manager check: request creation failed", slog.String("url", managerURL), slog.String("error", errMsg))
		s.saveResult(ctx, result)
		return result
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		errMsg := fmt.Sprintf("HTTP request failed: %v", err)
		result.ErrorMessage = errMsg
		s.logger.Error("manager check: HTTP request failed", slog.String("url", managerURL), slog.String("error", errMsg))
		s.saveResult(ctx, result)
		return result
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			s.logger.Warn("manager check: failed to close response body", slog.String("url", managerURL), slog.String("error", cerr.Error()))
		}
	}()

	result.HTTPStatus = resp.StatusCode

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("unexpected HTTP status: %d", resp.StatusCode)
		result.ErrorMessage = errMsg
		s.logger.Error("manager check: unexpected status", slog.String("url", managerURL), slog.Int("status", resp.StatusCode))
		s.saveResult(ctx, result)
		return result
	}

	// Parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errMsg := fmt.Sprintf("failed to read response body: %v", err)
		result.ErrorMessage = errMsg
		s.logger.Error("manager check: failed to read body", slog.String("url", managerURL), slog.String("error", errMsg))
		s.saveResult(ctx, result)
		return result
	}

	var healthResp healthResponse
	if err := json.Unmarshal(body, &healthResp); err != nil {
		errMsg := fmt.Sprintf("failed to parse response: %v", err)
		result.ErrorMessage = errMsg
		s.logger.Error("manager check: failed to parse response", slog.String("url", managerURL), slog.String("error", errMsg))
		s.saveResult(ctx, result)
		return result
	}

	// Check if status is "success"
	if healthResp.Status != "success" {
		errMsg := fmt.Sprintf("manager returned status: %s", healthResp.Status)
		result.ErrorMessage = errMsg
		result.Status = "error"
		s.logger.Error("manager check: manager returned error status", slog.String("url", managerURL), slog.String("status", healthResp.Status))
		s.saveResult(ctx, result)
		return result
	}

	// Success case
	result.Status = "success"
	result.ErrorMessage = ""
	s.logger.Info("manager check: success", slog.String("url", managerURL))
	s.saveResult(ctx, result)
	return result
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
