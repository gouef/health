package health

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func (h *Health) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		res := h.RunChecks(c.Request.Context())

		statusCode := http.StatusOK
		if res.Status == StatusDown {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, res)
	}
}