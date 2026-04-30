package mobile

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// activeModelsResponse matches the JSON shape at ActiveModelsURL.
type activeModelsResponse struct {
	Models []struct {
		ID        string   `json:"Id"`
		IpfsCID   string   `json:"IpfsCID"`
		Fee       int64    `json:"Fee"`
		Stake     int64    `json:"Stake"`
		Owner     string   `json:"Owner"`
		Name      string   `json:"Name"`
		Tags      []string `json:"Tags"`
		CreatedAt int64    `json:"CreatedAt"`
		IsDeleted bool     `json:"IsDeleted"`
		ModelType string   `json:"ModelType"`
	} `json:"models"`
	LastUpdated int64 `json:"last_updated"`
}

// refreshModelsCache fetches models from the HTTP endpoint and updates the cache.
func (s *SDK) refreshModelsCache() error {
	if s.cfg.ActiveModelsURL == "" {
		return fmt.Errorf("ActiveModelsURL not configured")
	}

	resp, err := s.httpClient.Get(s.cfg.ActiveModelsURL)
	if err != nil {
		return fmt.Errorf("fetch active models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("active models returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	hash := sha256.Sum256(body)
	hashStr := hex.EncodeToString(hash[:])

	s.modelsMu.RLock()
	same := hashStr == s.modelsHash
	s.modelsMu.RUnlock()
	if same {
		s.modelsMu.Lock()
		s.modelsCacheAt = time.Now()
		s.modelsMu.Unlock()
		s.log.Debug("active models unchanged, extending cache")
		return nil
	}

	var data activeModelsResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return fmt.Errorf("parse active models: %w", err)
	}

	models := make([]Model, 0, len(data.Models))
	byID := make(map[string]*Model, len(data.Models))
	byName := make(map[string]*Model, len(data.Models))

	for _, m := range data.Models {
		if m.IsDeleted {
			continue
		}
		model := Model{
			ID:        m.ID,
			Name:      m.Name,
			Tags:      m.Tags,
			Fee:       fmt.Sprintf("%d", m.Fee),
			Stake:     fmt.Sprintf("%d", m.Stake),
			Owner:     m.Owner,
			ModelType: m.ModelType,
			CreatedAt: m.CreatedAt,
		}
		models = append(models, model)
		stored := models[len(models)-1:]
		byID[m.ID] = &stored[0]
		byName[strings.ToLower(m.Name)] = &stored[0]
	}

	s.modelsMu.Lock()
	s.modelsCache = models
	s.modelsByID = byID
	s.modelsByName = byName
	s.modelsCacheAt = time.Now()
	s.modelsHash = hashStr
	s.modelsMu.Unlock()

	s.log.Infof("cached %d active models from HTTP endpoint", len(models))
	return nil
}

// GetAllModels returns active models, preferring the HTTP endpoint with blockchain fallback.
func (s *SDK) GetAllModels(ctx context.Context) ([]Model, error) {
	// Try HTTP cache first
	s.modelsMu.RLock()
	cached := s.modelsCache
	age := time.Since(s.modelsCacheAt)
	s.modelsMu.RUnlock()

	if cached != nil && age < DefaultModelCacheTTL {
		return cached, nil
	}

	// Refresh from HTTP endpoint
	if s.cfg.ActiveModelsURL != "" {
		if err := s.refreshModelsCache(); err != nil {
			s.log.Warnf("HTTP model fetch failed, trying blockchain: %v", err)
		} else {
			s.modelsMu.RLock()
			result := s.modelsCache
			s.modelsMu.RUnlock()
			return result, nil
		}
	}

	// Fallback: blockchain via Multicall
	models, err := s.blockchain.GetAllModels(ctx)
	if err != nil {
		if cached != nil {
			s.log.Warn("blockchain fetch also failed, returning stale cache")
			return cached, nil
		}
		return nil, err
	}
	out := make([]Model, 0, len(models))
	for _, m := range models {
		if m.IsDeleted {
			continue
		}
		out = append(out, Model{
			ID:    m.Id.Hex(),
			Name:  m.Name,
			Tags:  m.Tags,
			Fee:   bigStr(m.Fee),
			Stake: bigStr(m.Stake),
			Owner: m.Owner.Hex(),
		})
	}
	return out, nil
}

// ResolveModelID resolves a model name or blockchain ID to the canonical blockchain ID.
func (s *SDK) ResolveModelID(ctx context.Context, nameOrID string) (string, error) {
	if _, err := s.GetAllModels(ctx); err != nil {
		return "", err
	}
	s.modelsMu.RLock()
	defer s.modelsMu.RUnlock()

	if m, ok := s.modelsByID[nameOrID]; ok {
		return m.ID, nil
	}
	if m, ok := s.modelsByName[strings.ToLower(nameOrID)]; ok {
		return m.ID, nil
	}
	return "", fmt.Errorf("model not found: %s", nameOrID)
}

// GetAllModelsJSON returns all models as a JSON string (for FFI).
func (s *SDK) GetAllModelsJSON(ctx context.Context) (string, error) {
	models, err := s.GetAllModels(ctx)
	if err != nil {
		return "", err
	}
	return toJSON(models)
}

// GetRatedBids returns bids for a model, scored and sorted by quality.
func (s *SDK) GetRatedBids(ctx context.Context, modelID string) ([]ScoredBid, error) {
	id := common.HexToHash(modelID)
	bids, err := s.blockchain.GetRatedBids(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]ScoredBid, 0, len(bids))
	for _, b := range bids {
		out = append(out, ScoredBid{
			ID:             b.ID.Hex(),
			Provider:       b.Bid.Provider.Hex(),
			ModelAgentID:   b.Bid.ModelAgentId.Hex(),
			PricePerSecond: bigIntStr(b.Bid.PricePerSecond),
			Score:          b.Score,
		})
	}
	return out, nil
}

// GetRatedBidsJSON returns rated bids as a JSON string (for FFI).
func (s *SDK) GetRatedBidsJSON(ctx context.Context, modelID string) (string, error) {
	bids, err := s.GetRatedBids(ctx, modelID)
	if err != nil {
		return "", err
	}
	return toJSON(bids)
}

// EstimateOpenSessionStakeJSON returns stake + formula inputs as JSON (for FFI / UI).
// Uses the top-scored bid (same as the first provider tried when opening a session).
func (s *SDK) EstimateOpenSessionStakeJSON(ctx context.Context, modelID string, durationSec int64, directPayment bool) (string, error) {
	id := common.HexToHash(modelID)
	dur := big.NewInt(durationSec)
	est, err := s.blockchain.EstimateOpenSessionStake(ctx, id, dur, directPayment)
	if err != nil {
		return "", err
	}
	b, err := json.Marshal(est)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
