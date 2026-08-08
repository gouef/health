package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestDashboardRegisterWebsiteValidation(t *testing.T) {
	d := NewDashboard(100 * time.Millisecond)

	if err := d.RegisterWebsite("", "https://example.com", nil); !errors.Is(err, ErrWebsiteNameRequired) {
		t.Fatalf("expected ErrWebsiteNameRequired, got %v", err)
	}

	if err := d.RegisterWebsite("site", "", nil); !errors.Is(err, ErrWebsiteURLRequired) {
		t.Fatalf("expected ErrWebsiteURLRequired, got %v", err)
	}
}

func TestDashboardRegisterWebsiteDuplicate(t *testing.T) {
	d := NewDashboard(100 * time.Millisecond)

	if err := d.RegisterWebsite("site", "https://example.com", nil); err != nil {
		t.Fatalf("expected first register to succeed, got %v", err)
	}

	if err := d.RegisterWebsite("site", "https://example.com", nil); !errors.Is(err, ErrWebsiteAlreadyRegistered) {
		t.Fatalf("expected ErrWebsiteAlreadyRegistered, got %v", err)
	}
}

func TestDashboardRun(t *testing.T) {
	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer okServer.Close()

	downServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer downServer.Close()

	d := NewDashboard(100 * time.Millisecond)
	if err := d.RegisterWebsite("ok", okServer.URL, okServer.Client()); err != nil {
		t.Fatalf("expected ok website registration to succeed, got %v", err)
	}
	if err := d.RegisterWebsite("down", downServer.URL, downServer.Client()); err != nil {
		t.Fatalf("expected down website registration to succeed, got %v", err)
	}

	res := d.Run(context.Background())

	if res.Status != StatusDegraded {
		t.Fatalf("expected degraded status, got %s", res.Status)
	}
	if res.Websites["ok"].Website.URL != okServer.URL {
		t.Fatalf("expected ok website URL %q, got %q", okServer.URL, res.Websites["ok"].Website.URL)
	}
	if res.Websites["ok"].Result.Status != StatusUp {
		t.Fatalf("expected ok website status up, got %s", res.Websites["ok"].Result.Status)
	}
	if res.Websites["down"].Result.Status != StatusDown {
		t.Fatalf("expected down website status down, got %s", res.Websites["down"].Result.Status)
	}
}

func TestDashboardWebsitesSorted(t *testing.T) {
	d := NewDashboard(100 * time.Millisecond)
	if err := d.RegisterWebsite("zeta", "https://zeta.example", nil); err != nil {
		t.Fatalf("register zeta failed: %v", err)
	}
	if err := d.RegisterWebsite("alpha", "https://alpha.example", nil); err != nil {
		t.Fatalf("register alpha failed: %v", err)
	}

	list := d.Websites()
	if len(list) != 2 {
		t.Fatalf("expected 2 websites, got %d", len(list))
	}
	if list[0].Name != "alpha" || list[1].Name != "zeta" {
		t.Fatalf("expected sorted website names [alpha zeta], got [%s %s]", list[0].Name, list[1].Name)
	}
}

func TestDashboardHandlerReturnsServiceUnavailableWhenAllDown(t *testing.T) {
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

	app.GET("/dashboard", d.Handler())

	req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 status, got %d", w.Code)
	}
}
