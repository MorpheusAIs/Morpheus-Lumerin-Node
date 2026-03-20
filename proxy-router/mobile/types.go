package mobile

import (
	"math/big"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
)

// Balance holds ETH and MOR balances as decimal strings (in wei).
type Balance struct {
	ETH string `json:"eth"`
	MOR string `json:"mor"`
}

// Model represents a registered AI model on the Morpheus network.
type Model struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Tags      []string `json:"tags"`
	Fee       string   `json:"fee"`
	Stake     string   `json:"stake"`
	Owner     string   `json:"owner"`
	ModelType string   `json:"model_type,omitempty"` // LLM, TTS, STT, EMBEDDING
	CreatedAt int64    `json:"created_at,omitempty"`
}

// ScoredBid represents a provider bid with its quality score.
type ScoredBid struct {
	ID             string  `json:"id"`
	Provider       string  `json:"provider"`
	ModelAgentID   string  `json:"model_agent_id"`
	PricePerSecond string  `json:"price_per_second"`
	Score          float64 `json:"score"`
}

// Session represents an active or closed inference session.
type Session struct {
	ID             string `json:"id"`
	User           string `json:"user"`
	Provider       string `json:"provider"`
	ModelAgentID   string `json:"model_agent_id"`
	BidID          string `json:"bid_id"`
	PricePerSecond string `json:"price_per_second"`
	OpenedAt       string `json:"opened_at"`
	EndsAt         string `json:"ends_at"`
	ClosedAt       string `json:"closed_at"`
}

func sessionFromInternal(s *structs.Session) *Session {
	if s == nil {
		return nil
	}
	return &Session{
		ID:             s.Id,
		User:           s.User.Hex(),
		Provider:       s.Provider.Hex(),
		ModelAgentID:   s.ModelAgentId,
		BidID:          s.BidID,
		PricePerSecond: bigOptStr(s.PricePerSecond),
		OpenedAt:       bigOptStr(s.OpenedAt),
		EndsAt:         bigOptStr(s.EndsAt),
		ClosedAt:       bigOptStr(s.ClosedAt),
	}
}

func bigOptStr(b *big.Int) string {
	if b == nil {
		return ""
	}
	return b.String()
}
