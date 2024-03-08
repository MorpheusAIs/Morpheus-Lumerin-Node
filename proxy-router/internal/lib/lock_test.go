package lib

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/testlib"
)

func TestMutexTimeout(t *testing.T) {
	m := NewMutex()
	timeout := time.Millisecond * 40
	smallTimeout := 1 * time.Millisecond

	// test lock
	m.Lock()
	start := time.Now()
	err := m.LockTimeout(timeout)
	elapsed := time.Since(start)
	require.ErrorIsf(t, err, ErrTimeout, "locked mutex should timeout")
	require.GreaterOrEqual(t, elapsed, timeout, "timeout should be at least %s", timeout)

	// test unlock
	m.Unlock()
	err = m.LockTimeout(smallTimeout)
	require.NoErrorf(t, err, "unlocked mutex should not return error")

	// unlock of unlocked
	m.Unlock()
	err = m.LockTimeout(smallTimeout)
	require.NoError(t, err, "unlock of unlocked mutex should not block")
}

func TestMutexTimeoutRepeat(t *testing.T) {
	testlib.Repeat(t, 10, func(t *testing.T) {
		testlib.RepeatConcurrent(t, 500, TestMutexTimeout)
	})
}

func TestMutexCtx(t *testing.T) {
	m := NewMutex()
	timeout := time.Millisecond * 40

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// test lock
	m.Lock()
	start := time.Now()
	err := m.LockCtx(ctx)
	require.ErrorIsf(t, err, context.DeadlineExceeded, "locked mutex should timeout")
	require.InEpsilonf(t, timeout, time.Since(start), 0.2, "timeout should be close to %s", timeout)

	// test unlock
	m.Unlock()
	err = m.LockCtx(context.Background())
	require.NoErrorf(t, err, "unlocked mutex should not return error")

	// unlock of unlocked
	m.Unlock()
	err = m.LockCtx(context.Background())
	require.NoError(t, err, "unlock of unlocked mutex should not block")
}

func TestMutexCtxRepeat(t *testing.T) {
	testlib.Repeat(t, 10, func(t *testing.T) {
		testlib.RepeatConcurrent(t, 500, TestMutexCtx)
	})
}
