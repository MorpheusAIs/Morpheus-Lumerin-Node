package lib

import (
	"math/big"
	"testing"
)

func TestGatewayDurationRounding(t *testing.T) {
	for _, seconds := range []int64{300, 301, 600, 1800, 86400} {
		for _, supply := range []int64{13, 100003, 2050000000000000000} {
			for _, budget := range []int64{7, 101, 1000003} {
				price := big.NewInt(19)
				stake, err := GatewaySessionStake(price, big.NewInt(seconds), big.NewInt(supply), big.NewInt(budget))
				if err != nil {
					t.Fatal(err)
				}
				stipend := new(big.Int).Div(new(big.Int).Mul(stake, big.NewInt(budget)), big.NewInt(supply))
				duration := stipend.Div(stipend, price)
				if duration.Cmp(big.NewInt(seconds)) < 0 {
					t.Fatalf("duration %v below requested %v", duration, seconds)
				}
				// Supply growth of one basis point is covered by the explicit allowance.
				adjusted := new(big.Int).Mul(big.NewInt(supply), big.NewInt(10001))
				stipend.Mul(stake, big.NewInt(budget)).Mul(stipend, big.NewInt(10000)).Div(stipend, adjusted).Div(stipend, price)
				if stipend.Cmp(big.NewInt(seconds)) < 0 {
					t.Fatal("supply drift lost requested duration")
				}
			}
		}
	}
}

func TestGatewayDurationMinimumAndInvalidInputs(t *testing.T) {
	for _, duration := range []int64{-1, 0, 299} {
		if _, err := GatewaySessionStake(big.NewInt(1), big.NewInt(duration), big.NewInt(7), big.NewInt(3)); err == nil {
			t.Fatal("accepted invalid duration")
		}
	}
	if _, err := GatewaySessionStake(nil, big.NewInt(300), big.NewInt(7), big.NewInt(3)); err == nil {
		t.Fatal("accepted nil price")
	}
}
