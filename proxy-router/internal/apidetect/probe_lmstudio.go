package apidetect

import (
	"context"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

// LM Studio serves two listings. /api/v1/models (0.4.0+) carries the served
// context of each loaded instance and the trained capabilities; /api/v0/models
// is the legacy listing kept as the identification fallback. Authentication
// is off by default; the key still goes with both, as with the rest of
// detection, so an auth proxy in front of LM Studio lets the probe in.
// https://lmstudio.ai/docs/developer/rest/list
// https://lmstudio.ai/docs/developer/rest/endpoints
func (d *Detector) probeLMStudio(ctx context.Context, base, modelName, apiKey string, ev *apispec.Evidence) bool {
	if list := d.getJSON(ctx, base+"/api/v1/models", apiKey); list != nil {
		if models, ok := list["models"].([]any); ok {
			ev.Stack = "lmstudio"
			tracef(ctx, "identified lmstudio by /api/v1/models")
			if entry, bySpelling := findModelEntryBy(models, modelName, "key"); entry != nil {
				if bySpelling {
					key, _ := entry["key"].(string)
					traceSpelledEntry(ctx, "lmstudio /api/v1/models", modelName, key)
				}
				readLMStudioV1(ctx, entry, ev)
			} else {
				tracef(ctx, "lmstudio /api/v1/models lists no entry for the configured model; no capability facts")
			}
			return true
		}
	}

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
	if entry, bySpelling := findModelEntry(data, modelName); entry != nil {
		if bySpelling {
			id, _ := entry["id"].(string)
			traceSpelledEntry(ctx, "lmstudio /api/v0/models", modelName, id)
		}
		readLMStudioV0(ctx, entry, ev)
	}
	return true
}

func readLMStudioV1(ctx context.Context, entry map[string]any, ev *apispec.Evidence) {
	const ep = "lmstudio:/api/v1/models"
	if key, ok := entry["key"].(string); ok && ev.ServedModelID == "" {
		ev.ServedModelID = key
	}
	if arch, ok := entry["architecture"].(string); ok && arch != "" {
		ev.Architecture = arch
		tracef(ctx, "lmstudio /api/v1/models reports architecture %q", arch)
	}
	if typ, ok := entry["type"].(string); ok && typ == "embedding" {
		ev.Caps.SetEmbeddings(true, ep+".type")
		tracef(ctx, "lmstudio /api/v1/models type embedding -> embeddings")
	}
	if q, ok := entry["quantization"].(map[string]any); ok {
		if name, ok := q["name"].(string); ok && ev.Caps.SetQuantization(name, ep+".quantization.name") {
			tracef(ctx, "lmstudio /api/v1/models quantization.name %q", name)
		}
	}
	if v, ok := jsonPositiveInt(entry["max_context_length"]); ok && ev.Caps.SetContextTrained(v, ep+".max_context_length") {
		tracef(ctx, "lmstudio /api/v1/models max_context_length %d -> contextTrained", v)
	}
	// Requests may land on any loaded instance, so the smallest configured
	// window is the served limit to plan against.
	if instances, ok := entry["loaded_instances"].([]any); ok {
		var served int64
		for _, inst := range instances {
			m, ok := inst.(map[string]any)
			if !ok {
				continue
			}
			cfg, ok := m["config"].(map[string]any)
			if !ok {
				continue
			}
			if v, ok := jsonPositiveInt(cfg["context_length"]); ok && (served == 0 || v < served) {
				served = v
			}
		}
		if served > 0 && ev.Caps.SetContext(served, ep+".loaded_instances[].config.context_length") {
			tracef(ctx, "lmstudio /api/v1/models loaded_instances[].config.context_length %d -> context", served)
		}
	}
	caps, ok := entry["capabilities"].(map[string]any)
	if !ok {
		return
	}
	if vision, ok := caps["vision"].(bool); ok && vision && ev.Caps.AddInputModality("image", ep+".capabilities.vision") {
		tracef(ctx, "lmstudio /api/v1/models capabilities.vision -> modalities.input image")
	}
	// trained_for_tool_use describes training, not acceptance: LM Studio
	// accepts tools[] for every model, so false is not a tools: false report.
	// https://lmstudio.ai/docs/developer/openai-compat/tools
	if tools, ok := caps["trained_for_tool_use"].(bool); ok && tools {
		ev.Caps.SetTools(true, ep+".capabilities.trained_for_tool_use")
		tracef(ctx, "lmstudio /api/v1/models capabilities.trained_for_tool_use true -> tools")
	}
	reasoning, ok := caps["reasoning"].(map[string]any)
	if !ok {
		return
	}
	if options, ok := reasoning["allowed_options"].([]any); ok {
		// "off" / "on" describe a toggle; anything else is a level. Only "off"
		// is not a documented shape, so it sets no reasoning fact (rule 3 lists
		// the explicit false reports; LM Studio is not among them).
		var levels []string
		hasOff, hasOn := false, false
		for _, o := range options {
			s, ok := o.(string)
			if !ok {
				continue
			}
			switch strings.ToLower(s) {
			case "off":
				hasOff = true
			case "on":
				hasOn = true
			default:
				levels = append(levels, s)
			}
		}
		if hasOn || len(levels) > 0 {
			ev.Caps.SetReasoning(true, ep+".capabilities.reasoning.allowed_options")
			tracef(ctx, "lmstudio /api/v1/models capabilities.reasoning.allowed_options -> reasoning")
		}
		if len(levels) > 0 {
			if hasOff {
				levels = append([]string{"none"}, levels...)
			}
			ev.ReasoningEfforts = levels
			ev.EffortsSource = ep + ".capabilities.reasoning.allowed_options"
			tracef(ctx, "lmstudio /api/v1/models capabilities.reasoning.allowed_options lists levels %v", levels)
		}
	}
	if def, ok := reasoning["default"].(string); ok && def != "" {
		ev.ThinkingDefault = def
		ev.ThinkingDefaultSource = ep + ".capabilities.reasoning.default"
		tracef(ctx, "lmstudio /api/v1/models capabilities.reasoning.default %q -> thinking.default", def)
	}
}

func readLMStudioV0(ctx context.Context, entry map[string]any, ev *apispec.Evidence) {
	const ep = "lmstudio:/api/v0/models"
	if arch, ok := entry["arch"].(string); ok {
		ev.Architecture = arch
	}
	if id, ok := entry["id"].(string); ok && ev.ServedModelID == "" {
		ev.ServedModelID = id
	}
	if v, ok := jsonPositiveInt(entry["max_context_length"]); ok && ev.Caps.SetContext(v, ep+".max_context_length") {
		tracef(ctx, "lmstudio /api/v0/models max_context_length %d -> context", v)
	}
	switch typ, _ := entry["type"].(string); typ {
	case "vlm":
		if ev.Caps.AddInputModality("image", ep+".type") {
			tracef(ctx, "lmstudio /api/v0/models type vlm -> modalities.input image")
		}
	case "embeddings":
		ev.Caps.SetEmbeddings(true, ep+".type")
		tracef(ctx, "lmstudio /api/v0/models type embeddings -> embeddings")
	}
	if q, ok := entry["quantization"].(string); ok && ev.Caps.SetQuantization(q, ep+".quantization") {
		tracef(ctx, "lmstudio /api/v0/models quantization %q", q)
	}
	if caps, ok := entry["capabilities"].([]any); ok {
		for _, c := range caps {
			if c == "tool_use" {
				ev.Caps.SetTools(true, ep+".capabilities")
				tracef(ctx, "lmstudio /api/v0/models capabilities tool_use -> tools")
				break
			}
		}
	}
}
