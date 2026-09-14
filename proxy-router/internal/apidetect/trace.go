package apidetect

import (
	"context"
	"fmt"
	"sync"
)

// detectTrace collects human-readable detection steps when a trace-enabled
// context is used. Production Detect calls carry no trace, so every tracef
// call is a no-op there.
type detectTrace struct {
	mu    sync.Mutex
	lines []string
}

type traceCtxKey struct{}

func withTrace(ctx context.Context) (context.Context, *detectTrace) {
	t := &detectTrace{}
	return context.WithValue(ctx, traceCtxKey{}, t), t
}

func tracef(ctx context.Context, format string, args ...any) {
	t, _ := ctx.Value(traceCtxKey{}).(*detectTrace)
	if t == nil {
		return
	}
	t.mu.Lock()
	t.lines = append(t.lines, fmt.Sprintf(format, args...))
	t.mu.Unlock()
}
