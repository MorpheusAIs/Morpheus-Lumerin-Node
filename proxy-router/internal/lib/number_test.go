package lib

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRat(t *testing.T) {
	expInt := 10
	numInt := 20.0
	denInt := 4.0

	exp := big.NewInt(0).Exp(big.NewInt(10), big.NewInt(int64(expInt)), nil)
	num := big.NewInt(0).Mul(big.NewInt(int64(numInt)), exp)
	den := big.NewInt(0).Mul(big.NewInt(int64(denInt)), exp)

	rat, ok := NewRat(num, den).Float64()

	require.True(t, ok)
	require.Equal(t, numInt/denInt, rat)
}
