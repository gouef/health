package health

import (
	"context"
	"testing"
	"time"
)

func TestNewDefaultsTimeout(t *testing.T) {
	health := New(0)
	if health.timeout <= 0 {
		t.Fatalf("expected default timeout to be positive, got %v", health.timeout)
	}
}

func TestHealthRegisterAndRunChecks(t *testing.T) {
	h := New(100 * time.Millisecond)
	h.Register(NewFuncChecker("ok", func(ctx context.Context) Result {
		return Result{Status: StatusUp, Type: "test"}
	}))
	h.Register(NewFuncChecker("down", func(ctx context.Context) Result {
		return Result{Status: StatusDown, Type: "test", Error: "boom"}
	}))

	res := h.RunChecks(context.Background())

	if res.Status != StatusDegraded {
		t.Fatalf("expected degraded status, got %s", res.Status)
	}
	if res.Services["ok"].Status != StatusUp {
		t.Fatalf("expected ok service to be up, got %s", res.Services["ok"].Status)
	}
	if res.Services["down"].Status != StatusDown {
		t.Fatalf("expected down service to be down, got %s", res.Services["down"].Status)
	}
}

func TestHealthRunChecksAllDown(t *testing.T) {
	h := New(100 * time.Millisecond)
	h.Register(NewFuncChecker("down-1", func(ctx context.Context) Result {
		return Result{Status: StatusDown, Type: "test", Error: "boom"}
	}))
	h.Register(NewFuncChecker("down-2", func(ctx context.Context) Result {
		return Result{Status: StatusDown, Type: "test", Error: "boom"}
	}))

	res := h.RunChecks(context.Background())

	if res.Status != StatusDown {
		t.Fatalf("expected down status, got %s", res.Status)
	}
	if len(res.Services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(res.Services))
	}
}

func TestHealthRunChecksWithoutCheckers(t *testing.T) {
	h := New(50 * time.Millisecond)

	res := h.RunChecks(context.Background())

	if res.Status != StatusUp {
		t.Fatalf("expected up status with no checkers, got %s", res.Status)
	}
	if len(res.Services) != 0 {
		t.Fatalf("expected no services, got %d", len(res.Services))
	}
}

func TestHealthRunChecksWithCanceledContext(t *testing.T) {
	h := New(50 * time.Millisecond)
	h.Register(NewFuncChecker("cancelled", func(ctx context.Context) Result {
		if ctx.Err() != nil {
			return Result{Status: StatusDown, Type: "test", Error: ctx.Err().Error()}
		}
		return Result{Status: StatusUp, Type: "test"}
	}))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	res := h.RunChecks(ctx)

	if res.Services["cancelled"].Status != StatusDown {
		t.Fatalf("expected cancelled checker to report down, got %s", res.Services["cancelled"].Status)
	}
}
