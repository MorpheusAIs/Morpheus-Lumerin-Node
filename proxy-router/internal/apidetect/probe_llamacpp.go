package apidetect

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

func (d *Detector) probeLlamaCpp(ctx context.Context, base, apiKey string, ev *apispec.Evidence) bool {
	props := d.getJSON(ctx, base+"/props", apiKey)
	if props == nil {
		return false
	}
	_, hasTemplate := props["chat_template"]
	_, hasSlots := props["total_slots"]
	if !hasTemplate && !hasSlots {
		return false
	}
	ev.Stack = "llamacpp"
	if tpl, ok := props["chat_template"].(string); ok {
		ev.ChatTemplate = tpl
		tracef(ctx, "identified llamacpp by /props; chat template source available (%d chars)", len(tpl))
	} else {
		tracef(ctx, "identified llamacpp by /props; no chat template exposed")
	}
	return true
}
