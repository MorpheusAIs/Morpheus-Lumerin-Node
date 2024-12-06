package rating

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/exp/maps"
)

type RatingConfig struct {
	Algorithm         string           `json:"algorithm"`
	Params            json.RawMessage  `json:"params"`
	ProviderAllowList []common.Address `json:"providerAllowlist"`
}

func NewRatingFromConfig(config json.RawMessage, log lib.ILogger) (*Rating, error) {
	var cfg RatingConfig
	err := json.Unmarshal(config, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal rating config: %w", err)
	}

	log.Infof("rating algorithm: %s, params %s", cfg.Algorithm, string(cfg.Params))

	scorer, err := factory(cfg.Algorithm, cfg.Params)
	if err != nil {
		return nil, err
	}

	return NewRating(scorer, cfg.ProviderAllowList, log), nil
}

func NewRating(scorer Scorer, providerAllowList []common.Address, log lib.ILogger) *Rating {
	allowList := map[common.Address]struct{}{}

	if providerAllowList != nil {
		for _, addr := range providerAllowList {
			allowList[addr] = struct{}{}
		}
	}

	providerAllowListLegacy := os.Getenv("PROVIDER_ALLOW_LIST")

	if providerAllowListLegacy != "" {
		log.Warnf("PROVIDER_ALLOW_LIST is deprecated, please use providerAllowList in rating config")
		addresses := strings.Split(providerAllowListLegacy, ",")
		for _, address := range addresses {
			addr := common.HexToAddress(strings.TrimSpace(address))
			allowList[addr] = struct{}{}
		}
		log.Warnf("added %d addresses from PROVIDER_ALLOW_LIST", len(addresses))
	}

	if len(allowList) == 0 {
		log.Infof("providerAllowList is disabled, all providers are allowed")
	} else {
		keys := maps.Keys(allowList)
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].Hex() < keys[j].Hex()
		})
		log.Infof("providerAllowList: %v", keys)
	}

	return &Rating{
		scorer:            scorer,
		providerAllowList: allowList,
	}
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
