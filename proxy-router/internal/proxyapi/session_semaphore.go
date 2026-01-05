package proxyapi

import (
	"context"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sync/semaphore"
)

// SessionSemaphore manages per-session concurrency limits using weighted semaphores.
// It ensures only one request can be processed per session at a time,
// with additional requests waiting in a queue.
type SessionSemaphore struct {
	semaphores sync.Map // map[common.Hash]*semaphore.Weighted
}

// NewSessionSemaphore creates a new SessionSemaphore manager
func NewSessionSemaphore() *SessionSemaphore {
	return &SessionSemaphore{}
}

// Acquire attempts to acquire the semaphore for the given session.
// If another request is already processing for this session, this will block
// until that request completes or the context is cancelled.
// Returns nil on success, or an error if the context was cancelled/timed out.
func (s *SessionSemaphore) Acquire(ctx context.Context, sessionID common.Hash) error {
	sem := s.getOrCreateSemaphore(sessionID)
	return sem.Acquire(ctx, 1)
}

// TryAcquire attempts to acquire the semaphore without blocking.
// Returns true if acquired, false if another request is already in progress.
func (s *SessionSemaphore) TryAcquire(sessionID common.Hash) bool {
	sem := s.getOrCreateSemaphore(sessionID)
	return sem.TryAcquire(1)
}

// Release releases the semaphore for the given session.
// Must be called after Acquire returns successfully.
func (s *SessionSemaphore) Release(sessionID common.Hash) {
	if sem, ok := s.semaphores.Load(sessionID); ok {
		sem.(*semaphore.Weighted).Release(1)
	}
}

// getOrCreateSemaphore returns the semaphore for a session, creating it if needed.
// Each semaphore has weight 1, meaning only 1 concurrent request per session.
func (s *SessionSemaphore) getOrCreateSemaphore(sessionID common.Hash) *semaphore.Weighted {
	sem, _ := s.semaphores.LoadOrStore(sessionID, semaphore.NewWeighted(1))
	return sem.(*semaphore.Weighted)
}

// Cleanup removes the semaphore for a session (e.g., when session expires).
// Should only be called when no requests are in progress for this session.
func (s *SessionSemaphore) Cleanup(sessionID common.Hash) {
	s.semaphores.Delete(sessionID)
}
