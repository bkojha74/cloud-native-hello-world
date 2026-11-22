package middleware

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

const ContextSourceKey = "response_source"

type structuredLog struct {
	Timestamp  string `json:"timestamp"`
	Level      string `json:"level"`
	RequestID  string `json:"request_id"`
	Endpoint   string `json:"endpoint"`
	DurationMS int64  `json:"duration_ms"`
	Status     int    `json:"status"`
	Source     string `json:"source,omitempty"`
}

func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		reqID, _ := c.Get(RequestIDKey)
		source, _ := c.Get(ContextSourceKey)

		status := c.Writer.Status()
		level := "info"
		if status >= 500 {
			level = "error"
		} else if status >= 400 {
			level = "warn"
		}

		logEntry := structuredLog{
			Timestamp:  start.UTC().Format(time.RFC3339),
			Level:      level,
			RequestID:  fmt.Sprintf("%v", reqID),
			Endpoint:   c.FullPath(),
			DurationMS: duration.Milliseconds(),
			Status:     status,
		}
		if s, ok := source.(string); ok && s != "" {
			logEntry.Source = s
		}
		if logEntry.Endpoint == "" {
			logEntry.Endpoint = c.Request.URL.Path
		}

		bytes, _ := json.Marshal(logEntry)
		fmt.Fprintln(os.Stdout, string(bytes))
	}
}
