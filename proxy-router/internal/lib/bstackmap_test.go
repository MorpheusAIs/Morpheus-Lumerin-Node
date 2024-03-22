package lib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBstackMapCount(t *testing.T) {
	bsm := makeSampleBSM()
	require.Equal(t, 3, bsm.Count())
}

func TestBstackMapCapacity(t *testing.T) {
	bsm := makeSampleBSM()
	bsm.Push("fourth", 4)
	require.Equal(t, 3, bsm.Count())
	require.Equal(t, 3, bsm.Capacity())
}

func TestBstackMapOverwrite(t *testing.T) {
	bsm := makeSampleBSM()
	bsm.Push("fourth", 4)
	_, ok := bsm.Get("first")
	require.False(t, ok)
}

func TestBstackMapAt(t *testing.T) {
	bsm := makeSampleBSM()
	item, _ := bsm.At(0)
	require.Equal(t, 1, item)

	item, _ = bsm.At(-1)
	require.Equal(t, 3, item)
}

func TestBstackMapAtNegative(t *testing.T) {
	bsm := makeSampleBSM()
	item, _ := bsm.At(-1)
	require.Equal(t, 3, item)
}

func TestBstackMapAtOutOfBounds(t *testing.T) {
	bsm := makeSampleBSM()
	_, ok := bsm.At(bsm.Count())
	require.False(t, ok)
}

func TestBstackMapAtOutOfBoundsNegative(t *testing.T) {
	bsm := makeSampleBSM()
	_, ok := bsm.At(-bsm.Count() - 1)
	require.False(t, ok)
}

func TestBstackMapClear(t *testing.T) {
	bsm := makeSampleBSM()
	bsm.Clear()
	require.Equal(t, 0, bsm.Count())
	require.Equal(t, 3, bsm.Capacity())

	_, ok := bsm.Get("second")
	require.False(t, ok)

	_, ok = bsm.At(0)
	require.False(t, ok)
}

func TestBstackMapRange(t *testing.T) {
	bsm := makeSampleBSM()
	var sum int
	bsm.Range(func(key string, value int) bool {
		sum += value
		return true
	})
	require.Equal(t, 6, sum)
}

func TestBstackMapRangeBreak(t *testing.T) {
	bsm := makeSampleBSM()
	loops := 0
	bsm.Range(func(key string, value int) bool {
		loops++
		return false
	})
	require.Equal(t, 1, loops)
}

func TestBstackMapFilter(t *testing.T) {
	bsm := makeSampleBSM()
	bsm.Filter(func(key string, value int) bool {
		return value%2 != 0
	})
	require.Equal(t, 2, bsm.Count())
	at0, _ := bsm.At(0)
	at1, _ := bsm.At(1)
	require.Equal(t, 1, at0)
	require.Equal(t, 3, at1)
}

func makeSampleBSM() *BoundStackMap[int] {
	bsm := NewBoundStackMap[int](3)
	bsm.Push("first", 1)
	bsm.Push("second", 2)
	bsm.Push("third", 3)
	return bsm
}
