package apidetect

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

func (d *Detector) probeTGI(ctx context.Context, base, apiKey string, ev *apispec.Evidence) bool {
	info := d.getJSON(ctx, base+"/info", apiKey)
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
	return true
}
