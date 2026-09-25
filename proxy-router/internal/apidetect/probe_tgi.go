package apidetect

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

// TGI leaves /info open even with an API key (the key covers generate / chat
// only), so it is requested key-free first (§0). max_total_tokens is the
// served window (prompt + completion); the completion budget is what remains
// above max_input_tokens.
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

	total, ok := jsonPositiveInt(info["max_total_tokens"])
	if !ok || !ev.Caps.SetContext(total, "tgi:/info.max_total_tokens") {
		return true
	}
	tracef(ctx, "tgi /info max_total_tokens %d -> context", total)
	// max_total - max_input is the completion budget left at the maximum
	// prompt length (router/src/validation.rs caps max_new_tokens at
	// max_total_tokens - prompt_len), i.e. the guaranteed floor, not the
	// largest completion a short prompt may request. With TGI's default
	// max_input_tokens = max_total_tokens - 1 this is 1.
	if input, ok := jsonPositiveInt(info["max_input_tokens"]); ok && input < total {
		if ev.Caps.SetMaxOutput(total-input, "tgi:/info.max_total_tokens-max_input_tokens") {
			tracef(ctx, "tgi /info max_total_tokens - max_input_tokens = %d -> maxOutput", total-input)
		}
	}
	return true
}
