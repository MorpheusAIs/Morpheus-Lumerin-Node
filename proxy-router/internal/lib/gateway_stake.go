package lib

import (
	"fmt"
	"math/big"
)

// GatewaySessionStake rounds up before the contract rounds stipend and duration
// down. One basis point of bounded headroom covers ordinary supply drift while
// approval/open mine. The existing stake guard applies to this final amount.
// Governance changes can still invalidate a quote; those failures must not loop.
func GatewaySessionStake(price, duration, supply, budget *big.Int) (*big.Int, error) {
	for _, value := range []*big.Int{price, duration, supply, budget} {
		if value == nil || value.Sign() <= 0 {
			return nil, fmt.Errorf("invalid session stake inputs")
		}
	}
	if duration.Cmp(big.NewInt(300)) < 0 {
		return nil, fmt.Errorf("gateway session duration below contract minimum")
	}
	numerator := new(big.Int).Mul(supply, price)
	numerator.Mul(numerator, duration)
	numerator.Mul(numerator, big.NewInt(10001))
	denominator := new(big.Int).Mul(budget, big.NewInt(10000))
	numerator.Add(numerator, new(big.Int).Sub(denominator, big.NewInt(1)))
	return numerator.Div(numerator, denominator), nil
}
