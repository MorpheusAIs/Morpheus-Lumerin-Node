package lib

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTaskReturnsNoError(t *testing.T) {
	testFunc := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}

	task := NewTaskFunc(testFunc)
	task.Start(context.Background())
	<-task.Done()
	require.NoError(t, task.Err())
}

func TestTaskReturnsError(t *testing.T) {
	err := fmt.Errorf("test error")
	testFunc := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return err
		}
	}

	task := NewTaskFunc(testFunc)
	task.Start(context.Background())
	<-task.Done()
	require.ErrorIs(t, err, task.Err())
}

func TestTaskStopNoError(t *testing.T) {
	testFunc := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			return nil
		}
	}

	task := NewTaskFunc(testFunc)
	task.Start(context.Background())
	time.Sleep(500 * time.Millisecond)

	select {
	case <-task.Done():
		require.Fail(t, "task should not be done")
	case <-task.Stop():
		require.NoError(t, task.Err())
	case <-time.After(2 * time.Second):
		require.Fail(t, "task should be stopped")
	}
}

func TestTaskRestart(t *testing.T) {
	sleepDuration := 50 * time.Millisecond
	testFunc := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * sleepDuration):
			return nil
		}
	}

	task := NewTaskFunc(testFunc)
	task.Start(context.Background())
	time.Sleep(sleepDuration)
	task.Stop()

	time.Sleep(sleepDuration)

	task.Start(context.Background())
	time.Sleep(sleepDuration)
	task.Stop()

	select {
	case <-task.Done():
		require.Fail(t, "task should not be done")
	case <-time.After(sleepDuration):
	}

	require.NoError(t, task.Err())
}

func TestTaskContextCancel(t *testing.T) {
	testFunc := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(1 * time.Second):
			return nil
		}
	}

	task := NewTaskFunc(testFunc)

	ctx, cancel := context.WithCancel(context.Background())
	task.Start(ctx)

	time.Sleep(500 * time.Millisecond)
	cancel()

	select {
	case <-task.Done():
		require.ErrorIs(t, context.Canceled, task.Err())
	case <-time.After(500 * time.Millisecond):
		require.Fail(t, "task should be cancelled")
	}
}

// TestGlobalDone tests that the global done channel remains the same for all start/stops and closes only when
// main routine is completed
func TestGlobalDone(t *testing.T) {
	testFunc := func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}

	task := NewTaskFunc(testFunc)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		task.Start(ctx)
		time.Sleep(100 * time.Millisecond)
		<-task.Stop()

		time.Sleep(100 * time.Millisecond)

		task.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		cancel()
	}()

	select {
	case <-task.Done():
		require.ErrorIs(t, context.Canceled, task.Err())
	case <-time.After(3000 * time.Millisecond):
		require.Fail(t, "task should be cancelled")
	}
}

func TestWaitDoneBeforeStart(t *testing.T) {
	sleepDuration := 50 * time.Millisecond
	testFunc := func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}

	task := NewTaskFunc(testFunc)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(sleepDuration)
		task.Start(ctx)
		time.Sleep(sleepDuration)
		<-task.Stop()

		time.Sleep(sleepDuration)

		task.Start(ctx)
		time.Sleep(sleepDuration)

		cancel()
	}()

	select {
	case <-task.Done():
		require.ErrorIs(t, context.Canceled, task.Err())
	case <-time.After(30 * sleepDuration):
		require.Fail(t, "task should be cancelled")
	}
}

func TestStartStop(t *testing.T) {
	testFunc := func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}

	task := NewTaskFunc(testFunc)
	for i := 0; i < 1000; i++ {
		task.Start(context.Background())
		<-task.Stop()
	}
}

func TestNoPanic(t *testing.T) {
	testFunc := func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	}

	task := NewTaskFunc(testFunc)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := 0; i < 10000; i++ {
		go task.Start(ctx)
		go task.Stop()
		go cancel()
	}
}

func TestStopAfterEnd(t *testing.T) {
	testFunc := func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("end task")
		return nil
	}

	task := NewTaskFunc(testFunc)
	ctx := context.Background()

	go task.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	<-task.Stop()
	fmt.Println("end")
}

func TestNilStringErr(t *testing.T) {
	var nilErr error
	err2 := WrapError(fmt.Errorf("test error"), nilErr)

	require.NotContains(t, strings.ToLower(err2.Error()), "panic")
}
