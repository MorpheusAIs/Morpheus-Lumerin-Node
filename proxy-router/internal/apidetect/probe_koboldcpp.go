package apidetect

import (
	"context"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

func (d *Detector) probeKoboldCpp(ctx context.Context, base, apiKey string, ev *apispec.Evidence) bool {
	ver := d.getJSON(ctx, base+"/api/extra/version", apiKey)
	if ver == nil {
		return false
	}
	if result, ok := ver["result"].(string); !ok || !strings.EqualFold(result, "koboldcpp") {
		return false
	}
	ev.Stack = "koboldcpp"
	tracef(ctx, "identified koboldcpp by /api/extra/version")
	return true
}
