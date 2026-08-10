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

// Read-only public RPCs answer reads happily but reject eth_sendRawTransaction.
// https://mainnet.base.org is first in the public endpoint list for chain 8453
// and returns exactly "method is not allowed on this endpoint". That phrasing
// was not matched, so the client never rotated to an endpoint that would accept
// the broadcast — making it impossible to open a session out of the box without
// setting ETH_NODE_ADDRESS.
func TestShouldRetryRPCError_ReadOnlyEndpoint(t *testing.T) {
	retryable := []struct {
		name string
		msg  string
	}{
		{"base mainnet public rpc", "method is not allowed on this endpoint"},
		{"capitalised", "Method Is Not Allowed On This Endpoint"},
		{"wrapped by proxy-router", "failed to send transaction: open session failed: failed to send transaction: method is not allowed on this endpoint"},
		{"method not allowed", "method not allowed"},
		{"method not found", "the method eth_sendRawTransaction does not exist/is not available: method not found"},
		{"unsupported method", "unsupported method: eth_sendRawTransaction"},
		{"method is not available", "method is not available"},
	}

	for _, tt := range retryable {
		t.Run(tt.name, func(t *testing.T) {
			if !shouldRetryRPCError(fmt.Errorf("%s", tt.msg)) {
				t.Errorf("shouldRetryRPCError(%q) = false, want true (should rotate to the next endpoint)", tt.msg)
			}
		})
	}

	// A revert is a decision by the contract, not an endpoint problem. Rotating
	// would just replay a doomed transaction against every RPC in the list.
	notRetryable := []struct {
		name string
		msg  string
	}{
		{"execution reverted", "execution reverted: insufficient allowance"},
		{"insufficient funds", "insufficient funds for gas * price + value: have 100 want 200"},
	}

	for _, tt := range notRetryable {
		t.Run(tt.name, func(t *testing.T) {
			if shouldRetryRPCError(fmt.Errorf("%s", tt.msg)) {
				t.Errorf("shouldRetryRPCError(%q) = true, want false (rotating cannot help)", tt.msg)
			}
		})
	}
}
