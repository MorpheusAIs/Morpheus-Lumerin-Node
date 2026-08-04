package blockchainapi

import (
	"strings"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
)

func TestBlockingModelHealthStatus(t *testing.T) {
	modelID := common.HexToHash("0x01")
	otherID := common.HexToHash("0x02")

	tests := []struct {
		name    string
		reports []system.ModelHealthReport
		want    string
	}{
		{"no reports (pre-upgrade provider)", nil, ""},
		{"model not in report", []system.ModelHealthReport{{ModelID: otherID.Hex(), Status: system.ModelHealthStatusUnhealthy}}, ""},
		{"healthy", []system.ModelHealthReport{{ModelID: modelID.Hex(), Status: system.ModelHealthStatusHealthy}}, ""},
		{"skipped (non-probeable type)", []system.ModelHealthReport{{ModelID: modelID.Hex(), Status: system.ModelHealthStatusSkipped}}, ""},
		{"no_bid (stale self-report)", []system.ModelHealthReport{{ModelID: modelID.Hex(), Status: system.ModelHealthStatusNoBid}}, ""},
		{"unhealthy", []system.ModelHealthReport{{ModelID: modelID.Hex(), Status: system.ModelHealthStatusUnhealthy}}, system.ModelHealthStatusUnhealthy},
		{"tee_unverified", []system.ModelHealthReport{{ModelID: modelID.Hex(), Status: system.ModelHealthStatusTeeUnverified}}, system.ModelHealthStatusTeeUnverified},
		{"no_model_configured", []system.ModelHealthReport{{ModelID: modelID.Hex(), Status: system.ModelHealthStatusNoModel}}, system.ModelHealthStatusNoModel},
		{"case-insensitive model ID match", []system.ModelHealthReport{{ModelID: strings.ToUpper(modelID.Hex()), Status: system.ModelHealthStatusUnhealthy}}, system.ModelHealthStatusUnhealthy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, blockingModelHealthStatus(tt.reports, modelID))
		})
	}
}
