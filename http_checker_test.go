package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestHTTPCheckerName(t *testing.T) {
	checker := NewHTTPChecker("demo", "http://example.com", nil)
	if checker.Name() != "demo" {
		t.Fatalf("expected name demo, got %s", checker.Name())
	}
}

func TestHTTPCheckerCheckSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	checker := NewHTTPChecker("demo", server.URL, server.Client())
	res := checker.Check(context.Background())

	if res.Status != StatusUp {
		t.Fatalf("expected up status, got %s", res.Status)
	}
	if res.Type != "http" {
		t.Fatalf("expected http type, got %s", res.Type)
	}
	if res.ResponseTime < 0 {
		t.Fatalf("expected non-negative response time, got %d", res.ResponseTime)
	}
}

func TestHTTPCheckerCheckFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	checker := NewHTTPChecker("demo", server.URL, server.Client())
	res := checker.Check(context.Background())

	if res.Status != StatusDown {
		t.Fatalf("expected down status, got %s", res.Status)
	}
	if res.Type != "http" {
		t.Fatalf("expected http type, got %s", res.Type)
	}
	if res.Error != http.StatusText(http.StatusInternalServerError) {
		t.Fatalf("expected error %q, got %q", http.StatusText(http.StatusInternalServerError), res.Error)
	}
}

func TestHTTPCheckerCheckInvalidURL(t *testing.T) {
	checker := NewHTTPChecker("demo", "://invalid-url", nil)
	res := checker.Check(context.Background())

	if res.Status != StatusDown {
		t.Fatalf("expected down status, got %s", res.Status)
	}
	if res.Type != "http" {
		t.Fatalf("expected http type, got %s", res.Type)
	}
	if res.Error == "" {
		t.Fatal("expected non-empty error for invalid URL")
	}
}

func TestHTTPCheckerCheckClientError(t *testing.T) {
	client := &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		return nil, errors.New("transport failed")
	})}
	checker := NewHTTPChecker("demo", "http://example.com", client)
	res := checker.Check(context.Background())

	if res.Status != StatusDown {
		t.Fatalf("expected down status, got %s", res.Status)
	}
	if res.Type != "http" {
		t.Fatalf("expected http type, got %s", res.Type)
	}
	if !strings.Contains(res.Error, "transport failed") {
		t.Fatalf("expected transport failed error, got %q", res.Error)
	}
}
