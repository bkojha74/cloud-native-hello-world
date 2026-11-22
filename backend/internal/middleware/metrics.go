package middleware

import (
	"strconv"
	"time"

	"backend/internal/metrics"

	"github.com/gin-gonic/gin"
)

func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		status := strconv.Itoa(c.Writer.Status())
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}
		method := c.Request.Method

		metrics.HTTPRequestsTotal.WithLabelValues(method, status, endpoint).Inc()
		metrics.HTTPRequestDurationSeconds.WithLabelValues(method, endpoint).Observe(duration)
	}
}
