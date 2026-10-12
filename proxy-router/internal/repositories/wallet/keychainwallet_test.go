package wallet

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// notifyUpdated used to close updatedCh before taking the mutex, so two
// concurrent wallet updates could close the same channel and panic.
func TestNotifyUpdated_ConcurrentCallsDoNotPanic(t *testing.T) {
	w := NewKeychainWallet(nil)
	for iter := 0; iter < 200; iter++ {
		var wg sync.WaitGroup
		for g := 0; g < 8; g++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				w.notifyUpdated()
			}()
		}
		wg.Wait()
	}
}

func TestNotifyUpdated_WakesWaitersAndRearms(t *testing.T) {
	w := NewKeychainWallet(nil)
	ch := w.PrivateKeyUpdated()
	w.notifyUpdated()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("waiter was not woken")
	}

	next := w.PrivateKeyUpdated()
	require.NotEqual(t, ch, next)
	select {
	case <-next:
		t.Fatal("fresh channel must be open")
	default:
	}
}
