package health

import (
	"context"
	"testing"
)

func TestFuncCheckerNameAndCheck(t *testing.T) {
	var called bool
	checker := NewFuncChecker("custom", func(ctx context.Context) Result {
		called = true
		return Result{Status: StatusUp, Type: "custom"}
	})

	if checker.Name() != "custom" {
		t.Fatalf("expected name custom, got %s", checker.Name())
	}

	res := checker.Check(context.Background())
	if !called {
		t.Fatal("expected checker function to be called")
	}
	if res.Status != StatusUp {
		t.Fatalf("expected up status, got %s", res.Status)
	}
	if res.Type != "custom" {
		t.Fatalf("expected custom type, got %s", res.Type)
	}
}
