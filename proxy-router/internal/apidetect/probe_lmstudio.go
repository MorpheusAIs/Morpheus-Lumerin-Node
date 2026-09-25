package apidetect

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

func (d *Detector) probeLMStudio(ctx context.Context, base, modelName, apiKey string, ev *apispec.Evidence) bool {
	list := d.getJSON(ctx, base+"/api/v0/models", apiKey)
	if list == nil {
		return false
	}
	data, ok := list["data"].([]any)
	if !ok {
		return false
	}
	ev.Stack = "lmstudio"
	tracef(ctx, "identified lmstudio by /api/v0/models")
	if entry := matchModelEntry(data, modelName); entry != nil {
		if arch, ok := entry["arch"].(string); ok {
			ev.Architecture = arch
		}
		if id, ok := entry["id"].(string); ok && ev.ServedModelID == "" {
			ev.ServedModelID = id
		}
	}
	return true
}
