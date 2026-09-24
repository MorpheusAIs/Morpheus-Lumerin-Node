package structs

import (
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type AllowanceRes struct {
	Allowance *lib.BigInt `json:"allowance" example:"100000000" swaggertype:"integer"`
}

type TxRes struct {
	Tx       common.Hash          `json:"tx" example:"0x1234"`
	Progress *lib.GatewayProgress `json:"progress,omitempty"`
}

type ErrRes struct {
	Error     string               `json:"error" example:"error message"`
	SessionID *common.Hash         `json:"sessionID,omitempty"`
	Progress  *lib.GatewayProgress `json:"progress,omitempty"`
}

type OpenSessionRes struct {
	SessionID common.Hash          `json:"sessionID" example:"0x1234"`
	Progress  *lib.GatewayProgress `json:"progress,omitempty"`
}

type ExistingSessionRes struct {
	ExistingSessionID string `json:"existingSessionID" example:"0x1234"`
}

type BalanceRes struct {
	Balance *lib.BigInt `json:"balance" swaggertype:"string"`
}

type ProviderRes struct {
	Provider *Provider `json:"provider"`
}

type ProvidersRes struct {
	Providers []*Provider `json:"providers"`
}

type BidRes struct {
	Bid *Bid `json:"bid"`
}

type BidsRes struct {
	Bids []*Bid `json:"bids"`
}

type ScoredBidsRes struct {
	Bids []ScoredBid `json:"bids"`
}

type ModelRes struct {
	Model *Model `json:"model"`
}

type ModelsRes struct {
	Models []*Model `json:"models"`
}

type TokenBalanceRes struct {
	MOR *lib.BigInt `json:"mor" example:"100000000" swaggertype:"integer"`
	ETH *lib.BigInt `json:"eth" example:"100000000" swaggertype:"integer"`
}

type TransactionsRes struct {
	Transactions []MappedTransaction `json:"transactions"`
}

type SessionRes struct {
	Session *Session `json:"session"`
}

type SessionsRes struct {
	Sessions []*Session `json:"sessions"`
}

type BudgetRes struct {
	Budget *lib.BigInt `json:"budget" example:"100000000" swaggertype:"integer"`
}

type SupplyRes struct {
	Supply *lib.BigInt `json:"supply" example:"100000000" swaggertype:"integer"`
}

type BlockRes struct {
	Block uint64 `json:"block" example:"1234"`
}

// OpenSessionStakeEstimate is the MOR amount and the inputs it was derived
// from, for the bid that was quoted. Unless a bid is named explicitly that is
// the top-scored bid, the one an unattended open tries first (same ordering as
// GetRatedBids).
//
// StakeWei is what the diamond pulls from the wallet. SessionCostWei is the
// compute that amount buys, price_per_second × duration. The two differ by
// roughly the supply-to-budget ratio and confusing them is what made the
// desktop app look like it was charging hundreds of times the real price.
type OpenSessionStakeEstimate struct {
	StakeWei           string  `json:"stake_wei"`
	SessionCostWei     string  `json:"session_cost_wei"`
	MorSupplyWei       string  `json:"mor_supply_wei"`
	EmissionsBudgetWei string  `json:"emissions_budget_wei"`
	PricePerSecondWei  string  `json:"price_per_second_wei"`
	DurationSeconds    string  `json:"duration_seconds"`
	DirectPayment      bool    `json:"direct_payment"`
	BidID              string  `json:"bid_id"`
	TopBidProvider     string  `json:"top_bid_provider"`
	TopBidScore        float64 `json:"top_bid_score"`
	Explanation        string  `json:"explanation"`
}

// SessionDurationBounds is the session length range the deployed contract
// accepts, so consumers stop hardcoding it.
type SessionDurationBounds struct {
	MinSeconds string `json:"min_seconds" example:"300"`
	MaxSeconds string `json:"max_seconds" example:"86400"`
}

// ModelPrice is what one model currently costs across its live providers.
//
// The per-model bid routes answer the same question one model at a time, which
// is right when a model has already been chosen and useless when the question
// is "which of these is cheap": a client would have to make one request per
// registered model, and there are hundreds. This is that sweep done once, in
// the process that already holds the RPC connection and a model cache.
//
// Prices are decimal wei strings. They are per second of compute, not the MOR
// an open moves; the emissions conversion is several hundred times larger and
// belongs to the estimate route, not here.
type ModelPrice struct {
	ModelID string `json:"model_id"`
	// MinPricePerSecondWei and MaxPricePerSecondWei are equal when a model has
	// exactly one live bid, and both are empty when it has none. An empty pair
	// means "no provider", which is not the same as "free", so clients must not
	// coerce it to zero.
	MinPricePerSecondWei string `json:"min_price_per_second_wei"`
	MaxPricePerSecondWei string `json:"max_price_per_second_wei"`
	// BidCount counts the live bids the prices were taken from, so a client can
	// say "cheapest of 4" rather than implying a single quoted price.
	BidCount int `json:"bid_count"`
}

// ModelPricesRes carries a partial answer deliberately. One model's bid read
// failing is not a reason to withhold the prices of every other model, so the
// failures are named and the rest are returned; a client showing a sorted list
// can mark those rows unknown instead of showing nothing at all.
type ModelPricesRes struct {
	Prices []ModelPrice `json:"prices"`
	// FailedModelIDs lists models whose bids could not be read this time.
	FailedModelIDs []string `json:"failed_model_ids"`
}
