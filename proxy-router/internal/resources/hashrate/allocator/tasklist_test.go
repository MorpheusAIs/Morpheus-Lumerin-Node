package allocator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/TitanInd/proxy/proxy-router-v3/internal/testlib"
)

func TestTasklistAdd(t *testing.T) {
	repeats := 10000

	tl := NewTaskList()
	testlib.RepeatConcurrent(t, repeats, func(t *testing.T) {
		tl.Add("", nil, 0, time.Now(), nil, nil, nil)
	})

	require.Equal(t, tl.Size(), repeats)
}

func TestTasklistRemove(t *testing.T) {
	repeats := 10000

	tl := NewTaskList()
	startCh := make(chan struct{})

	testlib.RepeatConcurrent(t, repeats, func(t *testing.T) {
		tl.Add("", nil, 0, time.Now(), nil, nil, nil)
		select {
		case <-startCh:
		default:
			close(startCh)
		}
	})

	<-startCh
	testlib.Repeat(t, repeats, func(t *testing.T) {
		_, ok := tl.LockNextTask()
		if ok {
			tl.UnlockAndRemove()
		}
	})

	require.Equal(t, tl.Size(), 0)
}
