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
func adjustPagination(order Order, length *big.Int, offset *big.Int, limit uint8) (newOffset *big.Int, newLimit *big.Int) {
	newOffset, newLimit = new(big.Int), new(big.Int)

	// if offset is larger than the length of the array,
	// just return empty array
	if offset.Cmp(length) >= 0 {
		return newOffset, newLimit
	}

	newOffset.Set(offset)
	newLimit.SetUint64(uint64(limit))

	// calculate the remaining elements at the end of the array
	remainingAtEnd := new(big.Int).Add(offset, newLimit)
	remainingAtEnd.Sub(length, remainingAtEnd)

	if remainingAtEnd.Sign() < 0 {
		// means that offset+limit is out of bounds,
		// so limit has to be reduced by the amount of overflow
		newLimit.Add(newLimit, remainingAtEnd)
	}

	// if the order is DESC, the offset has to be adjusted
	if order == OrderDESC {
		newOffset.Set(max(remainingAtEnd, big.NewInt(0)))
	}

	return newOffset, newLimit
}

// adjustOrder in-plase reverses the order of the array if the order is DESC
func adjustOrder[T any](order Order, arr []T) {
	if order == OrderASC {
		return
	}
	slices.Reverse(arr)
}

func max(a, b *big.Int) *big.Int {
	if a.Cmp(b) > 0 {
		return a
	}
	return b
}
