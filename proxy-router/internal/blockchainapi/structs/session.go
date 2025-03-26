package structs

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Session struct {
	Id                      string
	User                    common.Address
	Provider                common.Address
	ModelAgentId            string
	BidID                   string
	Stake                   *big.Int `swaggertype:"integer"`
	PricePerSecond          *big.Int `swaggertype:"integer"`
	CloseoutReceipt         string
	CloseoutType            *big.Int `swaggertype:"integer"`
	ProviderWithdrawnAmount *big.Int `swaggertype:"integer"`
	OpenedAt                *big.Int `swaggertype:"integer"`
	EndsAt                  *big.Int `swaggertype:"integer"`
	ClosedAt                *big.Int `swaggertype:"integer"`
}
