package agent

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/Shemistan/agent/internal/service"
)

// Handler contains all HTTP handlers for the agent service
type Handler struct {
	healthService       service.HealthService
	managerCheckService service.ManagerCheckService
	logger              *slog.Logger
}

// NewHandler creates a new Handler instance
func NewHandler(
	healthService service.HealthService,
	managerCheckService service.ManagerCheckService,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		healthService:       healthService,
		managerCheckService: managerCheckService,
		logger:              logger,
	}
}

// HealthResponse represents the response for the /health endpoint
type HealthResponse struct {
	Status string `json:"status"`
}

// ManagerCheckItemResponse represents a single manager check result
type ManagerCheckItemResponse struct {
	ManagerURL string `json:"manager_url"`
	Status     string `json:"status"`
	HTTPStatus *int   `json:"http_status,omitempty"`
	Error      string `json:"error,omitempty"`
}

// ManagerCheckResponse represents the response for the /check-manager endpoint
type ManagerCheckResponse struct {
	Status   string                     `json:"status"`
	Managers []ManagerCheckItemResponse `json:"managers"`
}

// Health handles GET /health requests
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.healthService.HandleHealth(ctx); err != nil {
		h.logger.Error("health handler: failed to save health call", slog.String("error", err.Error()))
		h.respondJSON(w, http.StatusInternalServerError, HealthResponse{Status: "error"})
		return
	}

	h.respondJSON(w, http.StatusOK, HealthResponse{Status: "success"})
}

// CheckManager handles GET /check-manager requests
func (h *Handler) CheckManager(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	results, err := h.managerCheckService.CheckManager(ctx)
	if err != nil {
		h.logger.Error("check-manager handler: service error", slog.String("error", err.Error()))
		h.respondJSON(w, http.StatusInternalServerError, ManagerCheckResponse{
			Status:   "error",
			Managers: []ManagerCheckItemResponse{},
		})
		return
	}

	// Build response with all manager check results
	managers := make([]ManagerCheckItemResponse, 0, len(results.Results))
	overallStatus := "success"

	for _, result := range results.Results {
		item := ManagerCheckItemResponse{
			ManagerURL: result.ManagerURL,
			Status:     result.Status,
		}

		if result.HTTPStatus != 0 {
			item.HTTPStatus = &result.HTTPStatus
		}

		if result.Status != "success" {
			item.Error = result.ErrorMessage
			overallStatus = "error"
		}

		managers = append(managers, item)
	}

	response := ManagerCheckResponse{
		Status:   overallStatus,
		Managers: managers,
	}

	h.respondJSON(w, http.StatusOK, response)
}

// respondJSON writes structured JSON responses and logs encoding errors.
func (h *Handler) respondJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		h.logger.Error("failed to encode response", slog.String("error", err.Error()))
	}
}
