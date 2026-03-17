package httphandlers

import (
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/gin-gonic/gin"
)

// RequestLogger is a middleware for logging HTTP requests
func RequestLogger(logger lib.ILogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		start := time.Now()
		logger.Infof("[HTTP-REQ] %s %s",
			c.Request.Method,
			path,
		)

		// Process request
		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		// Log details
		status := c.Writer.Status()
		latency := time.Since(start).Round(time.Millisecond)
		requestID, _ := c.Get("request_id")
		logger.Infof("[HTTP-RES] %s %s [%d] %v request_id=%v",
			c.Request.Method,
			path,
			status,
			latency,
			requestID,
		)
	}
}
