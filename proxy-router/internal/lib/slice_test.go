package lib

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopyEqual(t *testing.T) {
	sl := []int{1, 2, 3, 4, 5}
	sl2 := CopySlice(sl)
	require.ElementsMatch(t, sl, sl2)
}

func TestCopyMutation(t *testing.T) {
	sl := []int{1, 2, 3, 4, 5}
	sl2 := sl
	sl2[0] = 10
	require.Equal(t, sl[0], sl2[0])

	sl2 = CopySlice(sl)
	sl2[0] = 20
	require.NotEqual(t, sl[0], sl2[0])
}
