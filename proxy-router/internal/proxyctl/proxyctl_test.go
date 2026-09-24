package proxyctl

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

// rateLimitErr is the shape the eth node returns when it throttles a call.
var rateLimitErr = errors.New(`429 Too Many Requests: {"jsonrpc":"2.0","error":{"code":-32016,"message":"over rate limit"},"id":37}`)

func TestRetryWithBackoff(t *testing.T) {
	tests := []struct {
		name        string
		attempts    int
		failures    int // how many leading calls fail before success
		alwaysFails bool
		wantCalls   int
		wantErr     bool
		wantErrIsFn bool // the returned error should wrap the error fn produced
	}{
		{
			name:      "succeeds on first attempt",
			attempts:  5,
			failures:  0,
			wantCalls: 1,
		},
		{
			name:      "succeeds after transient rate limit",
			attempts:  5,
			failures:  1,
			wantCalls: 2,
		},
		{
			name:      "succeeds on the final allowed attempt",
			attempts:  3,
			failures:  2,
			wantCalls: 3,
		},
		{
			name:        "gives up after attempts are exhausted",
			attempts:    3,
			alwaysFails: true,
			wantCalls:   3,
			wantErr:     true,
			wantErrIsFn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			fn := func() error {
				calls++
				if tt.alwaysFails || calls <= tt.failures {
					return rateLimitErr
				}
				return nil
			}

			err := retryWithBackoff(context.Background(), tt.attempts, time.Millisecond, lib.NewTestLogger(), "balance read", fn)

			if calls != tt.wantCalls {
				t.Errorf("fn called %d times, want %d", calls, tt.wantCalls)
			}
			if tt.wantErr && err == nil {
				t.Fatal("expected an error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if tt.wantErrIsFn && !errors.Is(err, rateLimitErr) {
				t.Errorf("error %v does not wrap the error returned by fn", err)
			}
		})
	}
}

// A 429 must not become a fatal error while retries remain. This is the
// regression the startup restart loop came from.
func TestRetryWithBackoffToleratesRateLimit(t *testing.T) {
	calls := 0
	fn := func() error {
		calls++
		if calls < 3 {
			return rateLimitErr
		}
		return nil
	}

	if err := retryWithBackoff(context.Background(), 5, time.Millisecond, lib.NewTestLogger(), "balance read", fn); err != nil {
		t.Fatalf("rate limit should have been retried, got %v", err)
	}
	if calls != 3 {
		t.Errorf("fn called %d times, want 3", calls)
	}
}

func TestRetryWithBackoffStopsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	fn := func() error {
		calls++
		cancel() // the proxy is shutting down mid-flight
		return rateLimitErr
	}

	err := retryWithBackoff(ctx, 5, time.Hour, lib.NewTestLogger(), "balance read", fn)

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if calls != 1 {
		t.Errorf("fn called %d times after cancellation, want 1", calls)
	}
}

// The retry must not outlive a context deadline, so a shutdown is not held up
// by a long backoff delay.
func TestRetryWithBackoffRespectsDeadlineDuringDelay(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	fn := func() error { return rateLimitErr }

	start := time.Now()
	err := retryWithBackoff(ctx, 5, time.Hour, lib.NewTestLogger(), "balance read", fn)
	elapsed := time.Since(start)

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
	if elapsed > time.Second {
		t.Errorf("retry waited %s, should have returned at the deadline", elapsed)
	}
}

func TestRetryWithBackoffDelayGrows(t *testing.T) {
	var timestamps []time.Time
	fn := func() error {
		timestamps = append(timestamps, time.Now())
		return rateLimitErr
	}

	_ = retryWithBackoff(context.Background(), 4, 10*time.Millisecond, lib.NewTestLogger(), "balance read", fn)

	if len(timestamps) != 4 {
		t.Fatalf("fn called %d times, want 4", len(timestamps))
	}
	first := timestamps[1].Sub(timestamps[0])
	last := timestamps[3].Sub(timestamps[2])
	if last <= first {
		t.Errorf("delay did not grow: first gap %s, last gap %s", first, last)
	}
}
