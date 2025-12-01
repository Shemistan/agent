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

// ManagerCheckResponse represents the response for the /check-manager endpoint
type ManagerCheckResponse struct {
	Status        string `json:"status"`
	ManagerStatus string `json:"manager_status,omitempty"`
	ManagerURL    string `json:"manager_url,omitempty"`
	HTTPStatus    *int   `json:"http_status,omitempty"`
	Error         string `json:"error,omitempty"`
}

// Health handles GET /health requests
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.healthService.HandleHealth(ctx); err != nil {
		h.logger.Error("health handler: failed to save health call", slog.String("error", err.Error()))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(HealthResponse{Status: "error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{Status: "success"})
}

// CheckManager handles GET /check-manager requests
func (h *Handler) CheckManager(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result, err := h.managerCheckService.CheckManager(ctx)
	if err != nil {
		h.logger.Error("check-manager handler: service error", slog.String("error", err.Error()))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ManagerCheckResponse{
			Status: "error",
			Error:  "internal server error",
		})
		return
	}

	response := ManagerCheckResponse{
		ManagerStatus: result.Status,
		ManagerURL:    result.ManagerURL,
	}

	if result.Status == "success" {
		response.Status = "success"
		if result.HTTPStatus != 0 {
			response.HTTPStatus = &result.HTTPStatus
		}
	} else {
		response.Status = "error"
		if result.HTTPStatus != 0 {
			response.HTTPStatus = &result.HTTPStatus
		}
		response.Error = result.ErrorMessage
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
