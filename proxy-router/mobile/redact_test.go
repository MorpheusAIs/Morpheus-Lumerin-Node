package mobile

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestRedactProviderEndpointsString_KeepInLockstep documents the exact
// shape of the redaction output. The same patterns are mirrored on the
// gateway side (nodeneo/go/internal/gateway/redact.go) and the Flutter UI
// (nodeneo/lib/utils/error_redaction.dart). Updating one without updating
// the others creates inconsistent error rendering.
func TestRedactProviderEndpointsString_KeepInLockstep(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "full URL with IPv4 + port + path",
			in:   `Post "http://216.81.245.17:18788/embeddings": dial tcp 216.81.245.17:18788: connect: connection refused`,
			want: `Post "<provider endpoint>": dial tcp <provider>: connect: connection refused`,
		},
		{
			name: "https with FQDN host",
			in:   "could not connect to https://provider.mor.org:3333/v1 again",
			want: "could not connect to <provider endpoint> again",
		},
		{
			name: "bare host:port",
			in:   "dial tcp provider.example.com:36318: i/o timeout",
			want: "dial tcp <provider>: i/o timeout",
		},
		{
			name: "bare IPv4",
			in:   "no route to host 74.48.78.46 — try again",
			want: "no route to host <provider> — try again",
		},
		{
			name: "non-matching text untouched",
			in:   "insufficient MOR balance — your wallet does not have enough MOR",
			want: "insufficient MOR balance — your wallet does not have enough MOR",
		},
		{
			name: "version numbers and timestamps NOT mistaken for IPs",
			in:   "v1.2.3 released at 12:34:56",
			want: "v1.2.3 released at 12:34:56",
		},
		{
			name: "empty input",
			in:   "",
			want: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := redactProviderEndpointsString(tc.in)
			if got != tc.want {
				t.Errorf("\n  in:   %q\n  got:  %q\n  want: %q", tc.in, got, tc.want)
			}
		})
	}
}

// TestRedactError_PreservesErrorChain ensures the wrapped error still
// matches errors.Is/As against the original sentinel. Callers (notably
// the proxy-router itself) use sentinels like ErrProvider; if redaction
// broke those checks we'd silently change behaviour for retry / rate-limit
// logic that depends on identity comparisons.
func TestRedactError_PreservesErrorChain(t *testing.T) {
	sentinel := errors.New("provider request failed")
	wrapped := fmt.Errorf("%w: dial tcp 216.81.245.17:18788: connect: connection refused", sentinel)

	red := redactError(wrapped)
	if red == nil {
		t.Fatal("redactError returned nil for non-nil input")
	}

	msg := red.Error()
	for _, leaked := range []string{"216.81.245.17", "18788"} {
		if strings.Contains(msg, leaked) {
			t.Errorf("redacted message still leaks %q: %s", leaked, msg)
		}
	}
	if !strings.Contains(msg, "<provider>") && !strings.Contains(msg, "<provider endpoint>") {
		t.Errorf("redacted message missing placeholder: %s", msg)
	}

	if !errors.Is(red, sentinel) {
		t.Errorf("redacted error must still match original sentinel via errors.Is")
	}
}

// TestRedactError_NilInput is a fast guard against the dumb mistake of
// dereferencing a nil error during wrapping.
func TestRedactError_NilInput(t *testing.T) {
	if got := redactError(nil); got != nil {
		t.Errorf("redactError(nil) = %v, want nil", got)
	}
}

// TestRedactError_NoChangeIsPassthrough verifies that messages without any
// provider-identifying tokens are returned unchanged (same pointer, no
// wrapper allocated). This keeps the cost of the SDK-boundary redaction
// effectively zero on the common case where errors never carried provider
// info to begin with (e.g. "insufficient MOR balance").
func TestRedactError_NoChangeIsPassthrough(t *testing.T) {
	original := errors.New("insufficient MOR balance — your wallet does not have enough MOR")
	got := redactError(original)
	if got != original {
		t.Errorf("clean error should be returned as-is, got wrapper %T", got)
	}
}
