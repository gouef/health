package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestHandlerReturnsOKWhenHealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := gin.New()
	h := New(100 * time.Millisecond)
	h.Register(NewFuncChecker("ok", func(ctx context.Context) Result {
		return Result{Status: StatusUp, Type: "test"}
	}))
	app.GET("/health", h.Handler())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 status, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Status != StatusUp {
		t.Fatalf("expected up status, got %s", response.Status)
	}
}

func TestHandlerReturnsServiceUnavailableWhenDown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := gin.New()
	h := New(100 * time.Millisecond)
	h.Register(NewFuncChecker("down", func(ctx context.Context) Result {
		return Result{Status: StatusDown, Type: "test", Error: "boom"}
	}))
	app.GET("/health", h.Handler())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 status, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Status != StatusDegraded && response.Status != StatusDown {
		t.Fatalf("expected degraded or down status, got %s", response.Status)
	}
}

func TestHandlerReturnsEmptyResponseWhenNoCheckers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := gin.New()
	h := New(50 * time.Millisecond)
	app.GET("/health", h.Handler())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 status, got %d", w.Code)
	}

	var response Response
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Status != StatusUp {
		t.Fatalf("expected up status, got %s", response.Status)
	}
	if len(response.Services) != 0 {
		t.Fatalf("expected no services, got %d", len(response.Services))
	}
}
