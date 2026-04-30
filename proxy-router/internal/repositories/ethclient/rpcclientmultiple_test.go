package ethclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
)

func TestShouldRetryRPCError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "context canceled",
			err:      context.Canceled,
			expected: false,
		},
		{
			name:     "context deadline",
			err:      context.DeadlineExceeded,
			expected: false,
		},
		{
			name:     "rate limit text",
			err:      fmt.Errorf("429 too many requests"),
			expected: true,
		},
		{
			name:     "rate limit code",
			err:      fmt.Errorf("-32005"),
			expected: true,
		},
		{
			name:     "execution reverted",
			err:      fmt.Errorf("execution reverted"),
			expected: false,
		},
		{
			name:     "revert",
			err:      fmt.Errorf("revert"),
			expected: false,
		},
		{
			name:     "timeout",
			err:      fmt.Errorf("connection timeout"),
			expected: true,
		},
		{
			name:     "connection refused",
			err:      fmt.Errorf("connection refused"),
			expected: true,
		},
		{
			name:     "connection reset",
			err:      fmt.Errorf("connection reset"),
			expected: true,
		},
		{
			name:     "eof",
			err:      fmt.Errorf("eof"),
			expected: true,
		},
		{
			name:     "cloudflare block",
			err:      fmt.Errorf("just a moment"),
			expected: true,
		},
		{
			name:     "quota exceeded",
			err:      fmt.Errorf("quota exceeded"),
			expected: true,
		},
		{
			name:     "insufficient funds",
			err:      fmt.Errorf("insufficient funds for gas"),
			expected: false,
		},
		{
			name:     "usage limit",
			err:      fmt.Errorf("usage limit"),
			expected: true,
		},
		{
			name:     "method not found code",
			err:      fmt.Errorf("-32601"),
			expected: true,
		},
		{
			name:     "json syntax error",
			err:      &json.SyntaxError{},
			expected: false,
		},
		{
			name:     "url error",
			err:      &url.Error{Op: "Get", URL: "http://example.com", Err: fmt.Errorf("dial failed")},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldRetryRPCError(tt.err)
			if got != tt.expected {
				t.Errorf("shouldRetryRPCError(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}
