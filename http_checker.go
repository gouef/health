package health

import (
	"context"
	"net/http"
	"time"
)

type HTTPChecker struct {
	name   string
	url    string
	client *http.Client
}

func NewHTTPChecker(name, url string, client *http.Client) *HTTPChecker {
	if client == nil {
		client = &http.Client{Timeout: 3 * time.Second}
	}
	return &HTTPChecker{name: name, url: url, client: client}
}

func (c *HTTPChecker) Name() string {
	return c.name
}

func (c *HTTPChecker) Check(ctx context.Context) Result {
	start := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return Result{
			Status:       StatusDown,
			Type:         "http",
			ResponseTime: time.Since(start).Milliseconds(),
			Error:        err.Error(),
		}
	}

	resp, err := c.client.Do(req)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		return Result{
			Status:       StatusDown,
			Type:         "http",
			ResponseTime: duration,
			Error:        err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return Result{
			Status:       StatusDown,
			Type:         "http",
			ResponseTime: duration,
			Error:        http.StatusText(resp.StatusCode),
		}
	}

	return Result{
		Status:       StatusUp,
		Type:         "http",
		ResponseTime: duration,
	}
}
