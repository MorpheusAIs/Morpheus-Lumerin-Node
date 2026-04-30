package blockchainapi

import (
	"math/big"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

func TestComputeSessionTokenAmount(t *testing.T) {
	makeBid := func(pricePerSecond int64) *structs.Bid {
		return &structs.Bid{
			PricePerSecond: &lib.BigInt{Int: *big.NewInt(pricePerSecond)},
		}
	}

	tests := []struct {
		name          string
		bid           *structs.Bid
		duration      *big.Int
		supply        *big.Int
		budget        *big.Int
		directPayment bool
		want          *big.Int
		wantErr       bool
	}{
		{
			name:          "direct payment",
			bid:           makeBid(100),
			duration:      big.NewInt(3600),
			supply:        big.NewInt(1_000_000),
			budget:        big.NewInt(50_000),
			directPayment: true,
			want:          big.NewInt(360_000),
		},
		{
			name:          "staked",
			bid:           makeBid(100),
			duration:      big.NewInt(3600),
			supply:        big.NewInt(1_000_000),
			budget:        big.NewInt(50_000),
			directPayment: false,
			want:          big.NewInt(7_200_000),
		},
		{
			name:          "nil bid",
			bid:           nil,
			duration:      big.NewInt(3600),
			supply:        big.NewInt(1_000_000),
			budget:        big.NewInt(50_000),
			directPayment: false,
			wantErr:       true,
		},
		{
			name:          "zero budget",
			bid:           makeBid(100),
			duration:      big.NewInt(3600),
			supply:        big.NewInt(1_000_000),
			budget:        big.NewInt(0),
			directPayment: false,
			wantErr:       true,
		},
		{
			name:          "nil budget",
			bid:           makeBid(100),
			duration:      big.NewInt(3600),
			supply:        big.NewInt(1_000_000),
			budget:        nil,
			directPayment: false,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := computeSessionTokenAmount(tt.bid, tt.duration, tt.supply, tt.budget, tt.directPayment)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got result %s", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Cmp(tt.want) != 0 {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
