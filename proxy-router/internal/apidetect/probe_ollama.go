package apidetect

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

func (d *Detector) probeOllama(ctx context.Context, base, modelName, apiKey string, ev *apispec.Evidence) bool {
	tags := d.getJSON(ctx, base+"/api/tags", apiKey)
	if tags == nil {
		return false
	}
	if _, ok := tags["models"]; !ok {
		return false
	}
	ev.Stack = "ollama"
	tracef(ctx, "identified ollama by /api/tags shape")

	body, _ := json.Marshal(map[string]string{"model": modelName, "name": modelName})
	show := d.requestJSON(ctx, http.MethodPost, base+"/api/show", apiKey, body, maxProbeBody)
	if show == nil {
		return true
	}
	if tpl, ok := show["template"].(string); ok {
		ev.ChatTemplate = tpl
	}
	if caps, ok := show["capabilities"].([]any); ok {
		for _, c := range caps {
			if c == "thinking" {
				ev.OllamaThinking = true
			}
		}
	}
	if info, ok := show["model_info"].(map[string]any); ok {
		if arch, ok := info["general.architecture"].(string); ok {
			ev.Architecture = arch
			tracef(ctx, "ollama /api/show reports architecture %q", arch)
		}
	}
	return true
}
