package blockchainapi

import (
	"context"
	"math/big"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

func makeBid(pricePerSecond int64) *structs.Bid {
	return &structs.Bid{
		PricePerSecond: &lib.BigInt{Int: *big.NewInt(pricePerSecond)},
	}
}

func TestComputeSessionTokenAmount(t *testing.T) {
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
			// 100 x (3600 + 1 second of headroom) x 1e6 / 50e3.
			name:     "converts the session cost through the emissions ratio",
			bid:      makeBid(100),
			duration: big.NewInt(3600),
			supply:   big.NewInt(1_000_000),
			budget:   big.NewInt(50_000),
			want:     big.NewInt(7_202_000),
		},
		{
			name:     "rounds up rather than leaving the session a wei short",
			bid:      makeBid(1),
			duration: big.NewInt(1),
			supply:   big.NewInt(10),
			budget:   big.NewInt(3),
			// 1 x 2 x 10 / 3 = 6.67, rounded up.
			want: big.NewInt(7),
		},
		{
			name:     "nil bid",
			bid:      nil,
			duration: big.NewInt(3600),
			supply:   big.NewInt(1_000_000),
			budget:   big.NewInt(50_000),
			wantErr:  true,
		},
		{
			name:     "zero duration",
			bid:      makeBid(100),
			duration: big.NewInt(0),
			supply:   big.NewInt(1_000_000),
			budget:   big.NewInt(50_000),
			wantErr:  true,
		},
		{
			name:     "zero budget",
			bid:      makeBid(100),
			duration: big.NewInt(3600),
			supply:   big.NewInt(1_000_000),
			budget:   big.NewInt(0),
			wantErr:  true,
		},
		{
			name:     "nil budget",
			bid:      makeBid(100),
			duration: big.NewInt(3600),
			supply:   big.NewInt(1_000_000),
			budget:   nil,
			wantErr:  true,
		},
		{
			name:     "nil supply",
			bid:      makeBid(100),
			duration: big.NewInt(3600),
			supply:   nil,
			budget:   big.NewInt(50_000),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := computeSessionTokenAmount(context.Background(), tt.bid, tt.duration, tt.supply, tt.budget, tt.directPayment)
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

// The old direct-payment branch returned price x duration, which the contract
// reads as a stake and converts through stakeToStipend, buying a few hundredth
// of the requested length and reverting as SessionTooShort. The amount must not
// depend on the payment method.
func TestComputeSessionTokenAmountBuysTheRequestedDuration(t *testing.T) {
	price := big.NewInt(190_000_000_000_000) // 0.00019 MOR/s, a real bid price.
	supply, _ := new(big.Int).SetString("42000000000000000000000000", 10)
	budget, _ := new(big.Int).SetString("122800000000000000000000", 10)

	for _, duration := range []*big.Int{
		big.NewInt(900), big.NewInt(3600), big.NewInt(21600), big.NewInt(86400),
	} {
		amount, err := computeSessionTokenAmount(context.Background(), &structs.Bid{PricePerSecond: &lib.BigInt{Int: *price}}, duration, supply, budget, false)
		if err != nil {
			t.Fatalf("duration %s: unexpected error: %v", duration, err)
		}

		// getSessionEnd: stakeToStipend(amount) / pricePerSecond, where
		// stakeToStipend is amount x computeBalance / (supply x 100) and
		// getTodaysBudget is computeBalance / 100.
		stipend := new(big.Int).Div(new(big.Int).Mul(amount, budget), supply)
		bought := new(big.Int).Div(stipend, price)
		if bought.Cmp(duration) < 0 {
			t.Errorf("duration %s: amount %s buys only %s seconds", duration, amount, bought)
		}
	}
}
