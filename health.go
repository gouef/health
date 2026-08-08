package health

import (
	"context"
	"sync"
	"time"
)

type Status string

const (
	StatusUp       Status = "UP"
	StatusDown     Status = "DOWN"
	StatusDegraded Status = "DEGRADED"
)

type Result struct {
	Status       Status                 `json:"status"`
	Type         string                 `json:"type,omitempty"`
	ResponseTime int64                  `json:"response_time_ms"`
	Error        string                 `json:"error,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

type Response struct {
	Status   Status            `json:"status"`
	Services map[string]Result `json:"services"`
}

type Health struct {
	mu       sync.RWMutex
	checkers []Checker
	timeout  time.Duration
}

func New(timeout time.Duration) *Health {
	if timeout == 0 {
		timeout = 3 * time.Second
	}
	return &Health{
		checkers: make([]Checker, 0),
		timeout:  timeout,
	}
}

func (h *Health) Register(checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers = append(h.checkers, checker)
}

func (h *Health) RunChecks(ctx context.Context) Response {
	h.mu.RLock()
	defer h.mu.RUnlock()

	ctx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()

	results := make(map[string]Result)
	var mu sync.Mutex
	var wg sync.WaitGroup

	overallStatus := StatusUp

	for _, c := range h.checkers {
		wg.Add(1)
		go func(checker Checker) {
			defer wg.Done()

			res := checker.Check(ctx)

			mu.Lock()
			results[checker.Name()] = res

			if res.Status == StatusDown {
				overallStatus = StatusDegraded
			}
			mu.Unlock()
		}(c)
	}

	wg.Wait()

	if len(results) > 0 {
		allDown := true
		for _, res := range results {
			if res.Status == StatusUp {
				allDown = false
				break
			}
		}
		if allDown {
			overallStatus = StatusDown
		}
	}

	return Response{
		Status:   overallStatus,
		Services: results,
	}
}
