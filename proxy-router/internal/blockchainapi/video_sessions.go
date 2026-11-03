package blockchainapi

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
	"github.com/ethereum/go-ethereum/common"
)

var ErrNoVideoGuide = errors.New("provider advertises no video guide for this model")

// videoGuideIn returns the guide a provider's pong carries for modelID.
func videoGuideIn(models []system.ModelHealthReport, modelID common.Hash) *videojob.Guide {
	for _, report := range models {
		if strings.EqualFold(report.ModelID, modelID.Hex()) {
			return report.Video
		}
	}
	return nil
}

// videoDefaultMinSessionSeconds is the contract's MIN_SESSION_DURATION, for a
// diamond that does not expose the getter.
const videoDefaultMinSessionSeconds = 300

// VideoSessionBounds is the contract's shortest and longest session. A getter
// the diamond does not route (seen on Base Sepolia) leaves the contract
// minimum, or no maximum; any other read error is returned.
func (s *BlockchainService) VideoSessionBounds(ctx context.Context) (int64, int64, error) {
	var minSeconds, maxSeconds int64 = videoDefaultMinSessionSeconds, 0
	if v, err := s.sessionRouter.GetMinSessionDuration(ctx); err == nil {
		minSeconds = v.Int64()
	} else if !selectorNotRegistered(err) {
		return 0, 0, err
	}
	if v, err := s.sessionRouter.GetMaxSessionDuration(ctx); err == nil {
		maxSeconds = v.Int64()
	} else if !selectorNotRegistered(err) {
		return 0, 0, err
	}
	return minSeconds, maxSeconds, nil
}

func selectorNotRegistered(err error) bool {
	return strings.Contains(err.Error(), "selector is not registered")
}

// VideoGuides reads the guide every provider bidding on modelID advertises,
// in bid rating order. A provider that cannot be reached or advertises no
// guide is listed with the reason.
func (s *BlockchainService) VideoGuides(ctx context.Context, modelID common.Hash) ([]structs.VideoProviderGuide, error) {
	bids, err := s.GetRatedBids(ctx, modelID)
	if err != nil {
		return nil, err
	}

	guides := make([]structs.VideoProviderGuide, len(bids))
	var wg sync.WaitGroup
	for i := range bids {
		bid := bids[i].Bid
		guides[i] = structs.VideoProviderGuide{
			Provider:          bid.Provider,
			BidID:             bid.Id,
			PricePerSecondWei: bid.PricePerSecond.String(),
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			provider, err := s.providerRegistry.GetProviderById(ctx, bid.Provider)
			if err != nil {
				guides[i].Error = fmt.Sprintf("failed to get provider: %s", err)
				return
			}
			pingCtx, cancel := context.WithTimeout(ctx, proxyapi.TimeoutPingDefault)
			_, _, models, err := s.proxyService.Ping(pingCtx, provider.Endpoint, bid.Provider)
			cancel()
			if err != nil {
				guides[i].Error = fmt.Sprintf("healthcheck ping failed: %s", err)
				return
			}
			guide := videoGuideIn(models, modelID)
			if guide == nil {
				guides[i].Error = ErrNoVideoGuide.Error()
				return
			}
			guides[i].Guide = guide
		}(i)
	}
	wg.Wait()

	return guides, nil
}
