package health

import (
	"bytes"
	_ "embed"
	"html/template"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

type dashboardTemplateWebsite struct {
	Name   string
	URL    string
	Result Result
}

type dashboardTemplateData struct {
	Title      string
	Status     Status
	Generated  string
	Websites   []dashboardTemplateWebsite
	TotalCount int
}

//go:embed *.gohtml
var bootstrapDashboardTemplateFile string

var bootstrapDashboardTemplate = template.Must(template.New("bootstrap-dashboard").Funcs(template.FuncMap{
	"statusClass": statusClass,
}).Parse(bootstrapDashboardTemplateFile))

func statusClass(status Status) string {
	switch status {
	case StatusUp:
		return "success"
	case StatusDown:
		return "danger"
	case StatusDegraded:
		return "warning"
	default:
		return "secondary"
	}
}

func GenerateBootstrapDashboardHTML(title string, response DashboardResponse) (string, error) {
	if title == "" {
		title = "Health Dashboard"
	}

	names := make([]string, 0, len(response.Websites))
	for name := range response.Websites {
		names = append(names, name)
	}
	sort.Strings(names)

	websites := make([]dashboardTemplateWebsite, 0, len(names))
	for _, name := range names {
		item := response.Websites[name]
		websites = append(websites, dashboardTemplateWebsite{
			Name:   item.Website.Name,
			URL:    item.Website.URL,
			Result: item.Result,
		})
	}

	data := dashboardTemplateData{
		Title:      title,
		Status:     response.Status,
		Generated:  time.Now().UTC().Format(time.RFC3339),
		Websites:   websites,
		TotalCount: len(websites),
	}

	var buf bytes.Buffer
	if err := bootstrapDashboardTemplate.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (d *Dashboard) BootstrapHandler(title string) gin.HandlerFunc {
	return func(c *gin.Context) {
		res := d.Run(c.Request.Context())

		html, err := GenerateBootstrapDashboardHTML(title, res)
		if err != nil {
			c.String(http.StatusInternalServerError, "failed to render dashboard template")
			return
		}

		statusCode := http.StatusOK
		if res.Status == StatusDown {
			statusCode = http.StatusServiceUnavailable
		}

		c.Status(statusCode)
		c.Header("Content-Type", "text/html; charset=utf-8")
		_, _ = c.Writer.WriteString(html)
	}
}
