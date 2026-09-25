package apidetect

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

// TGI leaves /info open even with an API key (the key covers generate / chat
// only), so it is requested key-free first (§0). max_total_tokens is the
// served window (prompt + completion). max_input_tokens is not published as
// anything: max_total_tokens - max_input_tokens is the completion budget left
// at the longest prompt (router/src/validation.rs caps max_new_tokens at
// max_total_tokens - prompt_len), a guaranteed minimum rather than the
// largest completion a short prompt may request, and with TGI's default
// max_input_tokens = max_total_tokens - 1 it is 1 (contract amendment a).
// https://github.com/huggingface/text-generation-inference/blob/v3.3.7/router/src/lib.rs
func (d *Detector) probeTGI(ctx context.Context, base, apiKey string, ev *apispec.Evidence) bool {
	info := d.getJSONKeyFree(ctx, base+"/info", apiKey)
	if info == nil {
		return false
	}
	modelID, ok := info["model_id"].(string)
	if !ok {
		return false
	}
	ev.Stack = "tgi"
	ev.ServedModelID = modelID
	tracef(ctx, "identified tgi by /info; served model id %q", modelID)

	if total, ok := jsonPositiveInt(info["max_total_tokens"]); ok && ev.Caps.SetContext(total, "tgi:/info.max_total_tokens") {
		tracef(ctx, "tgi /info max_total_tokens %d -> context", total)
	}
	return true
}
