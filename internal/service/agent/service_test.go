package agent

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Shemistan/agent/internal/storage"
)

// MockHealthStorage implements storage.HealthStorage interface
type MockHealthStorage struct {
	calls int
}

func (m *MockHealthStorage) SaveHealthCall(ctx context.Context, calledAt time.Time) error {
	m.calls++
	return nil
}

// MockManagerCheckStorage implements storage.ManagerCheckStorage interface
type MockManagerCheckStorage struct {
	savedChecks []storage.ManagerCheck
}

func (m *MockManagerCheckStorage) SaveManagerCheck(ctx context.Context, check storage.ManagerCheck) error {
	m.savedChecks = append(m.savedChecks, check)
	return nil
}

func TestHealthService_HandleHealth(t *testing.T) {
	mockStorage := &MockHealthStorage{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewHealthService(mockStorage, logger)

	err := service.HandleHealth(context.Background())
	if err != nil {
		t.Fatalf("HandleHealth failed: %v", err)
	}

	if mockStorage.calls != 1 {
		t.Fatalf("Expected 1 call, got %d", mockStorage.calls)
	}
}

func TestManagerCheckService_CheckManager_Success(t *testing.T) {
	// Create a test server that returns success
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("Expected /health, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"success"}`))
	}))
	defer server.Close()

	mockStorage := &MockManagerCheckStorage{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewManagerCheckService(
		server.Client(),
		mockStorage,
		[]string{server.URL},
		logger,
	)

	results, err := service.CheckManager(context.Background())
	if err != nil {
		t.Fatalf("CheckManager failed: %v", err)
	}

	if len(results.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results.Results))
	}

	result := results.Results[0]
	if result.Status != "success" {
		t.Fatalf("Expected success, got %s", result.Status)
	}

	if result.HTTPStatus != http.StatusOK {
		t.Fatalf("Expected HTTP 200, got %d", result.HTTPStatus)
	}

	if len(mockStorage.savedChecks) != 1 {
		t.Fatalf("Expected 1 saved check, got %d", len(mockStorage.savedChecks))
	}

	if mockStorage.savedChecks[0].Status != "success" {
		t.Fatalf("Expected saved status success, got %s", mockStorage.savedChecks[0].Status)
	}
}

func TestManagerCheckService_CheckManager_Failure_BadStatus(t *testing.T) {
	// Create a test server that returns error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"error"}`))
	}))
	defer server.Close()

	mockStorage := &MockManagerCheckStorage{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewManagerCheckService(
		server.Client(),
		mockStorage,
		[]string{server.URL},
		logger,
	)

	results, err := service.CheckManager(context.Background())
	if err != nil {
		t.Fatalf("CheckManager failed: %v", err)
	}

	if len(results.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results.Results))
	}

	result := results.Results[0]
	if result.Status != "error" {
		t.Fatalf("Expected error, got %s", result.Status)
	}

	if len(mockStorage.savedChecks) != 1 {
		t.Fatalf("Expected 1 saved check, got %d", len(mockStorage.savedChecks))
	}
}

func TestManagerCheckService_CheckManager_Failure_NonOKStatus(t *testing.T) {
	// Create a test server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	mockStorage := &MockManagerCheckStorage{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	service := NewManagerCheckService(
		server.Client(),
		mockStorage,
		[]string{server.URL},
		logger,
	)

	results, err := service.CheckManager(context.Background())
	if err != nil {
		t.Fatalf("CheckManager failed: %v", err)
	}

	if len(results.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results.Results))
	}

	result := results.Results[0]
	if result.Status != "error" {
		t.Fatalf("Expected error, got %s", result.Status)
	}

	if result.HTTPStatus != http.StatusInternalServerError {
		t.Fatalf("Expected HTTP 500, got %d", result.HTTPStatus)
	}

	if len(mockStorage.savedChecks) != 1 {
		t.Fatalf("Expected 1 saved check, got %d", len(mockStorage.savedChecks))
	}
}

func TestManagerCheckService_CheckManager_Failure_ConnectionError(t *testing.T) {
	mockStorage := &MockManagerCheckStorage{}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	client := &http.Client{
		Timeout: 1 * time.Millisecond,
	}
	service := NewManagerCheckService(
		client,
		mockStorage,
		[]string{"http://invalid-host-that-does-not-exist.example.com"},
		logger,
	)

	results, err := service.CheckManager(context.Background())
	if err != nil {
		t.Fatalf("CheckManager failed: %v", err)
	}

	if len(results.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results.Results))
	}

	result := results.Results[0]
	if result.Status != "error" {
		t.Fatalf("Expected error, got %s", result.Status)
	}

	if len(mockStorage.savedChecks) != 1 {
		t.Fatalf("Expected 1 saved check, got %d", len(mockStorage.savedChecks))
	}
}
