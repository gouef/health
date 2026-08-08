package health

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ErrWebsiteNameRequired      = errors.New("website name is required")
	ErrWebsiteURLRequired       = errors.New("website url is required")
	ErrWebsiteAlreadyRegistered = errors.New("website is already registered")
)

type Website struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type WebsiteHealth struct {
	Website Website `json:"website"`
	Result  Result  `json:"result"`
}

type DashboardResponse struct {
	Status   Status                   `json:"status"`
	Websites map[string]WebsiteHealth `json:"websites"`
}

type Dashboard struct {
	mu       sync.RWMutex
	health   *Health
	websites map[string]Website
}

func NewDashboard(timeout time.Duration) *Dashboard {
	return &Dashboard{
		health:   New(timeout),
		websites: make(map[string]Website),
	}
}

func (d *Dashboard) RegisterWebsite(name, url string, client *http.Client) error {
	name = strings.TrimSpace(name)
	url = strings.TrimSpace(url)

	if name == "" {
		return ErrWebsiteNameRequired
	}
	if url == "" {
		return ErrWebsiteURLRequired
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if _, exists := d.websites[name]; exists {
		return ErrWebsiteAlreadyRegistered
	}

	d.websites[name] = Website{Name: name, URL: url}
	d.health.Register(NewHTTPChecker(name, url, client))

	return nil
}

func (d *Dashboard) Websites() []Website {
	d.mu.RLock()
	defer d.mu.RUnlock()

	list := make([]Website, 0, len(d.websites))
	for _, website := range d.websites {
		list = append(list, website)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Name < list[j].Name
	})

	return list
}

func (d *Dashboard) Run(ctx context.Context) DashboardResponse {
	d.mu.RLock()
	websites := make(map[string]Website, len(d.websites))
	for name, website := range d.websites {
		websites[name] = website
	}
	d.mu.RUnlock()

	healthRes := d.health.RunChecks(ctx)
	result := DashboardResponse{
		Status:   healthRes.Status,
		Websites: make(map[string]WebsiteHealth, len(websites)),
	}

	for name, website := range websites {
		result.Websites[name] = WebsiteHealth{
			Website: website,
			Result:  healthRes.Services[name],
		}
	}

	return result
}

func (d *Dashboard) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		res := d.Run(c.Request.Context())

		statusCode := http.StatusOK
		if res.Status == StatusDown {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, res)
	}
}
