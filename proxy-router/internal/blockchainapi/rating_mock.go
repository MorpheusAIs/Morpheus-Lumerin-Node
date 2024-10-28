package blockchainapi

import (
	"math"
	"math/big"

	m "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/contracts/marketplace"
	"github.com/ethereum/go-ethereum/common"
)

func sampleDataTPS() ([][32]byte, []m.Bid, []m.ProviderModelStats, m.ModelStats) {
	modelID := common.HexToHash("0x01")
	bidIds := [][32]byte{
		{0x01},
		{0x02},
		{0x03},
	}

	bids := []m.Bid{
		{
			PricePerSecond: ToDecimal(10, DecimalsMOR),
			Provider:       common.HexToAddress("0x01"),
			ModelAgentId:   modelID,
			Nonce:          common.Big0,
			CreatedAt:      common.Big1,
			DeletedAt:      common.Big0,
		},
		{
			PricePerSecond: ToDecimal(10, DecimalsMOR),
			Provider:       common.HexToAddress("0x02"),
			ModelAgentId:   modelID,
			Nonce:          common.Big0,
			CreatedAt:      common.Big1,
			DeletedAt:      common.Big0,
		},
		{
			PricePerSecond: ToDecimal(10, DecimalsMOR),
			Provider:       common.HexToAddress("0x03"),
			ModelAgentId:   modelID,
			Nonce:          common.Big0,
			CreatedAt:      common.Big1,
			DeletedAt:      common.Big0,
		},
	}
	pmStats := []m.ProviderModelStats{
		{
			TpsScaled1000: m.LibSDSD{Mean: 10, SqSum: 100},
			TtftMs:        m.LibSDSD{Mean: 20, SqSum: 200},
			TotalDuration: 30,
			SuccessCount:  6,
			TotalCount:    10,
		},
		{
			TpsScaled1000: m.LibSDSD{Mean: 20, SqSum: 100},
			TtftMs:        m.LibSDSD{Mean: 20, SqSum: 200},
			TotalDuration: 30,
			SuccessCount:  6,
			TotalCount:    10,
		},
		{
			TpsScaled1000: m.LibSDSD{Mean: 30, SqSum: 100},
			TtftMs:        m.LibSDSD{Mean: 20, SqSum: 200},
			TotalDuration: 30,
			SuccessCount:  6,
			TotalCount:    10,
		},
	}
	mStats := m.ModelStats{
		TpsScaled1000: m.LibSDSD{Mean: 20, SqSum: 100},
		TtftMs:        m.LibSDSD{Mean: 20, SqSum: 200},
		TotalDuration: m.LibSDSD{Mean: 30, SqSum: 300},
		Count:         3,
	}

	return bidIds, bids, pmStats, mStats
}

func FromDecimal(value *big.Int, decimals int) float64 {
	return float64(value.Int64()) / float64(math.Pow10(decimals))
}

func ToDecimal(value float64, decimals int) *big.Int {
	a, _ := big.NewFloat(0).Mul(big.NewFloat(math.Pow10(decimals)), big.NewFloat(value)).Int(nil)
	return a
}

const DecimalsMOR = 18
