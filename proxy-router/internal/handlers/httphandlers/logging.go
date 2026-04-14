package httphandlers

import (
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/gin-gonic/gin"
)

// RequestLogger is a middleware for logging HTTP requests.
// It generates a request_id for every incoming request (unless one was
// already set by a downstream handler) and stores it in both the gin
// context and the stdlib context so that all layers can retrieve it
// via lib.RequestIDFromContext.
func RequestLogger(logger lib.ILogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-Id")
		if requestID == "" {
			requestID = lib.GenerateRequestID()
		}
		c.Set("request_id", requestID)
		c.Request = c.Request.WithContext(lib.ContextWithRequestID(c.Request.Context(), requestID))
		c.Writer.Header().Set("X-Request-Id", requestID)

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		start := time.Now()
		logger.Infof("[HTTP-REQ] %s %s request_id=%s",
			c.Request.Method,
			path,
			requestID,
		)

		c.Next()

		if raw != "" {
			path = path + "?" + raw
		}

		status := c.Writer.Status()
		latency := time.Since(start).Round(time.Millisecond)
		logger.Infof("[HTTP-RES] %s %s [%d] %v request_id=%s",
			c.Request.Method,
			path,
			status,
			latency,
			requestID,
		)
	}
}
