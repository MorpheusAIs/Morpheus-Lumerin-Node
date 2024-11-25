package rating

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

type RatingConfig struct {
	Algorithm         string           `json:"algorithm"`
	Params            json.RawMessage  `json:"params"`
	ProviderAllowList []common.Address `json:"providerAllowlist"`
}

func NewRatingFromConfig(config json.RawMessage) (*Rating, error) {
	var cfg RatingConfig
	err := json.Unmarshal(config, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal rating config: %w", err)
	}

	scorer, err := factory(cfg.Algorithm, cfg.Params)
	if err != nil {
		return nil, err
	}

	allowList := map[common.Address]struct{}{}
	for _, addr := range cfg.ProviderAllowList {
		allowList[addr] = struct{}{}
	}

	providerAllowListLegacy := os.Getenv("PROVIDER_ALLOW_LIST")

	if providerAllowListLegacy != "" {
		addresses := strings.Split(providerAllowListLegacy, ",")
		for _, address := range addresses {
			addr := strings.TrimSpace(address)
			allowList[common.HexToAddress(addr)] = struct{}{}
		}
	}

	return &Rating{
		scorer:            scorer,
		providerAllowList: allowList,
	}, nil
}

var (
	ErrUnknownAlgorithm = errors.New("unknown rating algorithm")
	ErrAlgorithmParams  = errors.New("invalid algorithm params")
)

func factory(algo string, params json.RawMessage) (a Scorer, err error) {
	switch algo {
	case ScorerNameDefault:
		a, err = NewScorerDefaultFromJSON(params)
		break
		// place here the new rating algorithms
		//
		// case "new_algorithm":
		// 	a, err = NewScorerNewAlgorithmFromJSON(params)
		// 	break
		//
	}
	if err != nil {
		return nil, lib.WrapError(ErrAlgorithmParams, fmt.Errorf("algorithm %s, error: %w, config: %s", algo, err, params))
	}
	if a == nil {
		return nil, lib.WrapError(ErrUnknownAlgorithm, fmt.Errorf("%s", algo))
	}

	return a, nil
}
