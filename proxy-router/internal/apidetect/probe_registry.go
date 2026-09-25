package apidetect

import (
	"context"
	"net/http"
	"sort"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
)

// Registry listings grow with the vendor's catalogue (OpenRouter ~0.7 MB).
// Variable so tests can exercise truncation.
var maxRegistryBody int64 = 16 << 20

var veniceCapabilityNames = map[string]string{
	"supportsReasoning":       "reasoning",
	"supportsFunctionCalling": "tools",
	"supportsVision":          "vision",
	"supportsWebSearch":       "web_search",
	"supportsResponseSchema":  "response_format",
	"supportsLogProbs":        "logprobs",
}

func init() {
	for _, vendor := range []string{"openrouter", "venice"} {
		registerHostedProbe(vendor, probeHostedRegistry)
	}
}

func probeHostedRegistry(ctx context.Context, d *Detector, bases []string, cfg config.ModelConfig, ev *apispec.Evidence) {
	tracef(ctx, "host recognized as %s by hostname; importing its model listing", ev.Stack)
	d.probeRegistryShape(ctx, bases, cfg.ModelName, cfg.ApiKey, ev)
}

func (d *Detector) probeRegistryShape(ctx context.Context, bases []string, modelName, apiKey string, ev *apispec.Evidence) {
	for _, base := range bases {
		list := d.requestJSON(ctx, http.MethodGet, base+"/models", apiKey, nil, maxRegistryBody)
		if list == nil {
			continue
		}
		data, ok := list["data"].([]any)
		if !ok {
			continue
		}
		entry := matchModelEntry(data, modelName)
		if entry == nil {
			tracef(ctx, "model listing at %s/models has %d entries but none matches model %q (config drift?)", base, len(data), modelName)
			continue
		}

		if params, ok := entry["supported_parameters"].([]any); ok {
			if ev.Stack == "" {
				ev.Stack = "openrouter"
			}
			tracef(ctx, "model listing at %s/models has supported_parameters -> openrouter-style registry", base)
			for _, p := range params {
				if s, ok := p.(string); ok {
					ev.Parameters = append(ev.Parameters, s)
				}
			}
			sort.Strings(ev.Parameters)
			for _, p := range ev.Parameters {
				if p == "reasoning" {
					ev.GatewayReasoning = true
					tracef(ctx, "openrouter listing reports reasoning support for %q", modelName)
					break
				}
			}
			return
		}

		if spec, ok := entry["model_spec"].(map[string]any); ok {
			if ev.Stack == "" {
				ev.Stack = "venice"
			}
			tracef(ctx, "model listing at %s/models has model_spec -> venice-style registry", base)
			if caps, ok := spec["capabilities"].(map[string]any); ok {
				for flag, name := range veniceCapabilityNames {
					if enabled, ok := caps[flag].(bool); ok && enabled {
						ev.Parameters = append(ev.Parameters, name)
					}
				}
				sort.Strings(ev.Parameters)
				if reasons, ok := caps["supportsReasoning"].(bool); ok && reasons {
					ev.GatewayReasoning = true
					tracef(ctx, "venice listing reports supportsReasoning")
				}
			}
			return
		}
	}
}
