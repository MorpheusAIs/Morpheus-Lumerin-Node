package registries

import (
	"math/big"
	"slices"
)

type Order = bool

const (
	OrderASC  Order = false
	OrderDESC Order = true
)

// function to "reverse" pagination based on the desired ordering
func adjustPagination(order Order, length *big.Int, offset *big.Int, limit uint8) (_offset *big.Int, _limit *big.Int) {
	bigLimit := big.NewInt(int64(limit))
	if order == OrderASC {
		return offset, bigLimit
	}
	offsetPlusLimit := new(big.Int).Add(offset, bigLimit)
	if offsetPlusLimit.Cmp(length) > 0 {
		offsetPlusLimit.Set(length)
	}
	newOffset := new(big.Int).Sub(length, offsetPlusLimit)
	return newOffset, bigLimit
}

// adjustOrder in-plase reverses the order of the array if the order is DESC
func adjustOrder[T any](order Order, arr []T) {
	if order == OrderASC {
		return
	}
	slices.Reverse(arr)
}
