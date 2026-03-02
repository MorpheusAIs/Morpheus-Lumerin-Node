package httphandlers

import (
	"net/http"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/gin-gonic/gin"
)

// DefaultMaxConcurrent is the default maximum concurrent requests.
// Can be overridden via PROXY_MAX_CONCURRENT environment variable.
// Tested stable at 500 concurrent on M-series Mac; 100 is conservative default.
const DefaultMaxConcurrent = 100

// ConcurrencyLimiter provides bounded concurrency control for HTTP handlers.
// When the limit is reached, new requests receive 503 Service Unavailable
// instead of being queued indefinitely or crashing the server.
type ConcurrencyLimiter struct {
	maxConcurrent int64
	current       int64
}

// NewConcurrencyLimiter creates a limiter with the specified max concurrent requests.
// Pass 0 to use default (or PROXY_MAX_CONCURRENT env var).
func NewConcurrencyLimiter(maxConcurrent int) *ConcurrencyLimiter {
	limit := maxConcurrent
	if limit <= 0 {
		limit = DefaultMaxConcurrent
		// Check environment variable override
		if envLimit := os.Getenv("PROXY_MAX_CONCURRENT"); envLimit != "" {
			if parsed, err := strconv.Atoi(envLimit); err == nil && parsed > 0 {
				limit = parsed
			}
		}
	}
	return &ConcurrencyLimiter{
		maxConcurrent: int64(limit),
		current:       0,
	}
}

// Middleware returns a gin middleware that enforces concurrency limits.
// Requests that exceed the limit receive 503 with queue depth info.
func (cl *ConcurrencyLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to acquire a slot
		current := atomic.AddInt64(&cl.current, 1)

		if current > cl.maxConcurrent {
			// Over limit - release and reject
			atomic.AddInt64(&cl.current, -1)
			c.Header("X-Concurrency-Limit", strconv.FormatInt(cl.maxConcurrent, 10))
			c.Header("X-Concurrency-Current", strconv.FormatInt(current-1, 10))
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "Server at capacity",
				"code":    "concurrency_limit_exceeded",
				"limit":   cl.maxConcurrent,
				"current": current - 1,
				"message": "Too many concurrent requests. Please retry.",
			})
			return
		}

		// Set headers for observability
		c.Header("X-Concurrency-Current", strconv.FormatInt(current, 10))
		c.Header("X-Concurrency-Limit", strconv.FormatInt(cl.maxConcurrent, 10))

		// Ensure we release the slot when done
		defer atomic.AddInt64(&cl.current, -1)

		c.Next()
	}
}

// Current returns the current number of in-flight requests.
func (cl *ConcurrencyLimiter) Current() int64 {
	return atomic.LoadInt64(&cl.current)
}

// MaxConcurrent returns the configured maximum concurrent requests.
func (cl *ConcurrencyLimiter) MaxConcurrent() int64 {
	return cl.maxConcurrent
}
