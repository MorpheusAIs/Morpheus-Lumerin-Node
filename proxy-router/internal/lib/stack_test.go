package lib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStackPop(t *testing.T) {
	st := NewStack[int]()
	st.Push(1)
	st.Push(2)
	st.Push(3)

	item, ok := st.Pop()
	require.True(t, ok)
	require.Equal(t, 3, item)

	item, ok = st.Pop()
	require.True(t, ok)
	require.Equal(t, 2, item)

	item, ok = st.Pop()
	require.True(t, ok)
	require.Equal(t, 1, item)

	_, ok = st.Pop()
	require.False(t, ok)
}

func TestStackLoop(t *testing.T) {
	st := NewStack[int]()
	st.Push(1)
	st.Push(2)
	st.Push(3)

	calledCount := 0

	for {
		_, ok := st.Peek()
		if !ok {
			break
		}
		calledCount++
		_, _ = st.Pop()
		if calledCount > 10 {
			panic("infinite loop")
		}
	}

	require.Equal(t, 3, calledCount)
}
