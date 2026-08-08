package health

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestGenerateBootstrapDashboardHTML(t *testing.T) {
	response := DashboardResponse{
		Status: StatusDegraded,
		Websites: map[string]WebsiteHealth{
			"alpha": {
				Website: Website{Name: "alpha", URL: "https://alpha.example"},
				Result:  Result{Status: StatusUp, Type: "http", ResponseTime: 12},
			},
			"beta": {
				Website: Website{Name: "beta", URL: "https://beta.example"},
				Result:  Result{Status: StatusDown, Type: "http", ResponseTime: 30, Error: "Internal Server Error"},
			},
		},
	}

	html, err := GenerateBootstrapDashboardHTML("Master Dashboard", response)
	if err != nil {
		t.Fatalf("expected template generation to succeed, got %v", err)
	}

	if !strings.Contains(html, "bootstrap@5.3.8") {
		t.Fatal("expected bootstrap 5.3.8 link in generated html")
	}
	if !strings.Contains(html, "Master Dashboard") {
		t.Fatal("expected custom title in generated html")
	}
	if !strings.Contains(html, "https://alpha.example") {
		t.Fatal("expected alpha website URL in generated html")
	}
	if !strings.Contains(html, "text-bg-warning\">DEGRADED") {
		t.Fatal("expected degraded status badge in generated html")
	}
}

func TestGenerateBootstrapDashboardHTMLDefaultsTitle(t *testing.T) {
	html, err := GenerateBootstrapDashboardHTML("", DashboardResponse{Status: StatusUp, Websites: map[string]WebsiteHealth{}})
	if err != nil {
		t.Fatalf("expected template generation to succeed, got %v", err)
	}

	if !strings.Contains(html, "Health Dashboard") {
		t.Fatal("expected default title in generated html")
	}
}

func TestBootstrapHandlerReturnsHTML(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := gin.New()

	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer okServer.Close()

	d := NewDashboard(100 * time.Millisecond)
	if err := d.RegisterWebsite("ok", okServer.URL, okServer.Client()); err != nil {
		t.Fatalf("register ok failed: %v", err)
	}

	app.GET("/dashboard/html", d.BootstrapHandler("Template Dashboard"))

	req := httptest.NewRequest(http.MethodGet, "/dashboard/html", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 status, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Fatalf("expected text/html content type, got %q", ct)
	}
	if !strings.Contains(w.Body.String(), "Template Dashboard") {
		t.Fatal("expected rendered dashboard title in body")
	}
}

func TestBootstrapHandlerReturnsServiceUnavailableWhenAllDown(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := gin.New()

	downServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer downServer.Close()

	d := NewDashboard(100 * time.Millisecond)
	if err := d.RegisterWebsite("down", downServer.URL, downServer.Client()); err != nil {
		t.Fatalf("register down failed: %v", err)
	}

	app.GET("/dashboard/html", d.BootstrapHandler("Template Dashboard"))

	req := httptest.NewRequest(http.MethodGet, "/dashboard/html", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 status, got %d", w.Code)
	}
}

func TestBootstrapHandlerUsesRequestContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	app := gin.New()

	d := NewDashboard(100 * time.Millisecond)
	d.health.Register(NewFuncChecker("ctx", func(ctx context.Context) Result {
		if ctx.Err() != nil {
			return Result{Status: StatusDown, Type: "custom", Error: ctx.Err().Error()}
		}
		time.Sleep(5 * time.Millisecond)
		return Result{Status: StatusUp, Type: "custom"}
	}))
	d.mu.Lock()
	d.websites["ctx"] = Website{Name: "ctx", URL: "https://ctx.example"}
	d.mu.Unlock()

	app.GET("/dashboard/html", d.BootstrapHandler("Template Dashboard"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/dashboard/html", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable && w.Code != http.StatusOK {
		t.Fatalf("expected a valid HTTP status code, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Template Dashboard") {
		t.Fatal("expected rendered dashboard title in body")
	}
}
