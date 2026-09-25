package apidetect

import (
	"context"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

// KoboldCpp guards no GET, so every request here goes key-free first (§0).
// The vision / audio flags are documented as "not always correct" upstream;
// they are still the backend's own report and are published as such.
// https://github.com/LostRuins/koboldcpp/blob/v1.121/koboldcpp.py
func (d *Detector) probeKoboldCpp(ctx context.Context, base, apiKey string, ev *apispec.Evidence) bool {
	ver := d.getJSONKeyFree(ctx, base+"/api/extra/version", apiKey)
	if ver == nil {
		return false
	}
	if result, ok := ver["result"].(string); !ok || !strings.EqualFold(result, "koboldcpp") {
		return false
	}
	ev.Stack = "koboldcpp"
	tracef(ctx, "identified koboldcpp by /api/extra/version")

	const ep = "koboldcpp:/api/extra/version"
	if v, ok := ver["vision"].(bool); ok && v && ev.Caps.AddInputModality("image", ep+".vision") {
		tracef(ctx, "koboldcpp /api/extra/version vision -> modalities.input image")
	}
	if v, ok := ver["audio"].(bool); ok && v && ev.Caps.AddInputModality("audio", ep+".audio") {
		tracef(ctx, "koboldcpp /api/extra/version audio -> modalities.input audio")
	}
	if v, ok := ver["websearch"].(bool); ok {
		ev.Caps.SetWebSearch(v, ep+".websearch")
		tracef(ctx, "koboldcpp /api/extra/version websearch %v -> webSearch", v)
	}
	if v, ok := ver["embeddings"].(bool); ok {
		ev.Caps.SetEmbeddings(v, ep+".embeddings")
		tracef(ctx, "koboldcpp /api/extra/version embeddings %v -> embeddings", v)
	}

	if maxCtx := d.getJSONKeyFree(ctx, base+"/api/extra/true_max_context_length", apiKey); maxCtx != nil {
		if v, ok := capInt64(maxCtx["value"]); ok && ev.Caps.SetContext(v, "koboldcpp:/api/extra/true_max_context_length.value") {
			tracef(ctx, "koboldcpp /api/extra/true_max_context_length value %d -> context", v)
		}
	}
	return true
}
