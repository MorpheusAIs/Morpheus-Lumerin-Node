package apidetect

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

func (d *Detector) probeVLLM(ctx context.Context, base, modelName, apiKey string, ev *apispec.Evidence) bool {
	ver := d.getJSON(ctx, base+"/version", apiKey)
	if ver == nil {
		return false
	}
	if _, ok := ver["version"]; !ok {
		return false
	}
	ev.Stack = "vllm"
	tracef(ctx, "identified vllm by /version")

	if list := d.getJSON(ctx, base+"/v1/models", apiKey); list != nil {
		if data, ok := list["data"].([]any); ok {
			if entry := matchModelEntry(data, modelName); entry != nil {
				if id, ok := entry["id"].(string); ok {
					ev.ServedModelID = id
				}
			}
		}
	}
	return true
}
