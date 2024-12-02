package registries

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOffsetLimit(t *testing.T) {
	type args struct {
		order  Order
		length *big.Int
		offset *big.Int
		limit  uint8
	}

	type want struct {
		offset *big.Int
		limit  *big.Int
	}

	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "ASC within bounds",
			args: args{order: OrderASC, length: big.NewInt(10), offset: big.NewInt(0), limit: 5},
			want: want{offset: big.NewInt(0), limit: big.NewInt(5)},
		},
		{
			name: "ASC offset out of bounds",
			args: args{order: OrderASC, length: big.NewInt(10), offset: big.NewInt(15), limit: 5},
			want: want{offset: big.NewInt(0), limit: big.NewInt(0)},
		},
		{
			name: "ASC limit out of bounds",
			args: args{order: OrderASC, length: big.NewInt(10), offset: big.NewInt(5), limit: 10},
			want: want{offset: big.NewInt(5), limit: big.NewInt(5)},
		},
		{
			name: "ASC offset and limit out of bounds",
			args: args{order: OrderASC, length: big.NewInt(10), offset: big.NewInt(15), limit: 15},
			want: want{offset: big.NewInt(0), limit: big.NewInt(0)},
		},
		{
			name: "DESC within bounds",
			args: args{order: OrderDESC, length: big.NewInt(10), offset: big.NewInt(0), limit: 5},
			want: want{offset: big.NewInt(5), limit: big.NewInt(5)},
		},
		{
			name: "DESC offset out of bounds",
			args: args{order: OrderDESC, length: big.NewInt(10), offset: big.NewInt(15), limit: 5},
			want: want{offset: big.NewInt(0), limit: big.NewInt(0)},
		},
		{
			name: "DESC limit out of bounds",
			args: args{order: OrderDESC, length: big.NewInt(10), offset: big.NewInt(5), limit: 10},
			want: want{offset: big.NewInt(0), limit: big.NewInt(5)},
		},
		{
			name: "DESC offset and limit out of bounds",
			args: args{order: OrderDESC, length: big.NewInt(10), offset: big.NewInt(15), limit: 15},
			want: want{offset: big.NewInt(0), limit: big.NewInt(0)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newOffset, newLimit := adjustPagination(tt.args.order, tt.args.length, tt.args.offset, tt.args.limit)
			require.Equalf(t, tt.want.offset.Cmp(newOffset), 0, "expected offset %v, got %v", tt.want.offset, newOffset)
			require.Equalf(t, tt.want.limit.Cmp(newLimit), 0, "expected limit %v, got %v", tt.want.limit, newLimit)
		})
	}
}
