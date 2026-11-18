package blockchainapi

import (
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/videojob"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

// testVideoGuide prices a job at duration × resolutionWeight × 10 seconds:
// the default job (5 s, 720p, weight 2) quotes 100 s.
func testVideoGuide(quoteMax int64) *videojob.Guide {
	return videojob.Spec{
		ModelName:    "test-video",
		MinDuration:  4,
		MaxDuration:  12,
		AspectRatios: []string{"16:9", "9:16"},
		Resolutions:  []string{"480p", "720p"},
		Defaults:     videojob.Defaults{DurationSeconds: 5, AspectRatio: "16:9", Resolution: "720p"},
		Quote: videojob.Quote{
			SecondsPerWorkUnit: 10,
			ResolutionWeights:  map[string]float64{"480p": 1, "720p": 2},
			MaxSessionSeconds:  quoteMax,
		},
	}.Guide()
}

func TestVideoGuideIn(t *testing.T) {
	modelID := common.HexToHash("0xabcdef01")
	otherID := common.HexToHash("0x02")
	guide := testVideoGuide(0)
	otherGuide := testVideoGuide(0)

	t.Run("matches the model id ignoring case", func(t *testing.T) {
		upper := "0x" + strings.ToUpper(strings.TrimPrefix(modelID.Hex(), "0x"))
		require.NotEqual(t, modelID.Hex(), upper)
		reports := []system.ModelHealthReport{
			{ModelID: otherID.Hex(), Video: otherGuide},
			{ModelID: upper, Video: guide},
		}
		require.Same(t, guide, videoGuideIn(reports, modelID))
	})

	t.Run("nil when the model is absent", func(t *testing.T) {
		reports := []system.ModelHealthReport{{ModelID: otherID.Hex(), Video: otherGuide}}
		require.Nil(t, videoGuideIn(reports, modelID))
		require.Nil(t, videoGuideIn(nil, modelID), "pre-upgrade provider: no reports at all")
	})

	t.Run("nil when the model advertises no guide", func(t *testing.T) {
		reports := []system.ModelHealthReport{{ModelID: modelID.Hex(), Status: system.ModelHealthStatusHealthy}}
		require.Nil(t, videoGuideIn(reports, modelID))
	})
}

// A video model is never probed, so its provider reports it "skipped". An
// open by bid, which is how a caller opens the session for a video job,
// refuses only a blocking status, so a skipped video model still opens.
func TestVideoSkippedStatusDoesNotBlockOpen(t *testing.T) {
	modelID := common.HexToHash("0x0a")
	pong := []system.ModelHealthReport{{ModelID: modelID.Hex(), Status: system.ModelHealthStatusSkipped, Video: testVideoGuide(0)}}
	require.Empty(t, blockingModelHealthStatus(pong, modelID))
}
