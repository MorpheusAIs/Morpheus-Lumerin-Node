package apidetect

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

// Field names: https://github.com/ollama/ollama/blob/v0.34.4/docs/api.md
// (Show Model Information, List Running Models),
// https://github.com/ollama/ollama/blob/v0.34.4/api/types.go,
// https://github.com/ollama/ollama/blob/v0.34.4/types/model/capability.go.
const (
	srcOllamaCapabilities = "ollama:/api/show.capabilities"
	srcOllamaThinkValues  = "ollama:/api/show.thinking.values"
	srcOllamaThinkDefault = "ollama:/api/show.thinking.default"
	srcOllamaDetailsCtx   = "ollama:/api/show.details.context_length"
	srcOllamaNumCtx       = "ollama:/api/show.parameters.num_ctx"
	srcOllamaQuantization = "ollama:/api/show.details.quantization_level"
	srcOllamaLoadedCtx    = "ollama:/api/ps.models.context_length"
)

// GGUF architecture names (llama, qwen3, gptoss, nomic-bert); anything else
// is not echoed into a source string.
var ollamaArchRe = regexp.MustCompile(`^[a-z0-9_-]{1,32}$`)

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
	if show := d.requestJSON(ctx, http.MethodPost, base+"/api/show", apiKey, body, maxProbeBody); show != nil {
		ollamaShowFacts(ctx, show, ev)
	}
	d.probeOllamaLoaded(ctx, base, modelName, apiKey, ev)
	return true
}

func ollamaShowFacts(ctx context.Context, show map[string]any, ev *apispec.Evidence) {
	if tpl, ok := show["template"].(string); ok {
		ev.ChatTemplate = tpl
	}

	// completion says nothing; insert and image are not capability fields.
	thinkingCap := false
	if caps, ok := show["capabilities"].([]any); ok {
		for _, c := range caps {
			switch c {
			case "tools":
				ev.Caps.SetTools(true, srcOllamaCapabilities)
				tracef(ctx, "ollama /api/show capabilities lists tools -> tools")
			case "vision":
				ev.Caps.AddInputModality("image", srcOllamaCapabilities)
				tracef(ctx, "ollama /api/show capabilities lists vision -> modalities.input image")
			case "audio":
				ev.Caps.AddInputModality("audio", srcOllamaCapabilities)
				tracef(ctx, "ollama /api/show capabilities lists audio -> modalities.input audio")
			case "embedding":
				ev.Caps.SetEmbeddings(true, srcOllamaCapabilities)
				tracef(ctx, "ollama /api/show capabilities lists embedding -> embeddings")
			case "thinking":
				thinkingCap = true
			}
		}
	}
	thinking, _ := show["thinking"].(map[string]any)
	if !ollamaThinkingFacts(ctx, thinking, ev) && thinkingCap {
		ev.OllamaThinking = true
	}

	if info, ok := show["model_info"].(map[string]any); ok {
		if arch, ok := info["general.architecture"].(string); ok {
			ev.Architecture = arch
			tracef(ctx, "ollama /api/show reports architecture %q", arch)
			if key := arch + ".context_length"; ollamaArchRe.MatchString(arch) {
				if n, ok := jsonPositiveInt(info[key]); ok && ev.Caps.SetContextTrained(n, "ollama:/api/show.model_info."+key) {
					tracef(ctx, "ollama /api/show model_info.%s = %d -> contextTrained", key, n)
				}
			}
		}
	}
	if details, ok := show["details"].(map[string]any); ok {
		if n, ok := jsonPositiveInt(details["context_length"]); ok && ev.Caps.SetContext(n, srcOllamaDetailsCtx) {
			tracef(ctx, "ollama /api/show details.context_length = %d -> context", n)
		}
		if q, ok := details["quantization_level"].(string); ok && ev.Caps.SetQuantization(q, srcOllamaQuantization) {
			tracef(ctx, "ollama /api/show details.quantization_level = %s -> quantization", q)
		}
	}
	if params, ok := show["parameters"].(string); ok {
		if n, ok := ollamaNumCtx(params); ok && ev.Caps.SetContext(n, srcOllamaNumCtx) {
			tracef(ctx, "ollama /api/show parameters num_ctx = %d -> context (configured beats details.context_length)", n)
		}
	}
}

// thinking.values is the accepted set of the think request field: booleans
// (a toggle) and level strings, https://github.com/ollama/ollama/blob/v0.34.4/docs/capabilities/thinking.mdx.
// [false] alone is a backend report of no reasoning (rule 3). Levels become
// ReasoningEfforts in the reported order, false among them as "none" (rule 5);
// booleans alone leave the toggle to OllamaThinking. Returns false when there
// is no usable object, so the capabilities list decides.
func ollamaThinkingFacts(ctx context.Context, thinking map[string]any, ev *apispec.Evidence) bool {
	values, _ := thinking["values"].([]any)
	if len(values) == 0 {
		return false
	}
	var on, levels bool
	var efforts []string
	add := func(e string) {
		for _, seen := range efforts {
			if seen == e {
				return
			}
		}
		efforts = append(efforts, e)
	}
	for _, v := range values {
		switch x := v.(type) {
		case bool:
			if x {
				on = true
			} else {
				add("none")
			}
		case string:
			if x != "none" {
				on, levels = true, true
			}
			add(x)
		}
	}
	if !on {
		ev.Caps.SetReasoning(false, srcOllamaThinkValues)
		tracef(ctx, "ollama /api/show thinking.values lists only false -> reasoning false")
		return true
	}

	ev.OllamaThinking = true
	ev.Caps.SetReasoning(true, srcOllamaThinkValues)
	tracef(ctx, "ollama /api/show thinking.values = %v -> reasoning true", values)
	if levels {
		ev.ReasoningEfforts = efforts
		ev.EffortsSource = srcOllamaThinkValues
		tracef(ctx, "ollama /api/show thinking.values levels -> reasoning efforts %s", strings.Join(efforts, "|"))
	}

	switch def := thinking["default"].(type) {
	case bool:
		ev.ThinkingDefault = "off"
		if def {
			ev.ThinkingDefault = "on"
		}
	case string:
		ev.ThinkingDefault = def
	}
	if ev.ThinkingDefault != "" {
		ev.ThinkingDefaultSource = srcOllamaThinkDefault
		tracef(ctx, "ollama /api/show thinking.default -> thinking.default %s", ev.ThinkingDefault)
	}
	return true
}

// The loaded window beats whatever the Modelfile configured: Ollama sizes
// num_ctx at load time from OLLAMA_CONTEXT_LENGTH and free VRAM,
// https://github.com/ollama/ollama/blob/v0.34.4/docs/context-length.mdx.
// The local API has no auth, so the key follows only a 401/403.
func (d *Detector) probeOllamaLoaded(ctx context.Context, base, modelName, apiKey string, ev *apispec.Evidence) {
	ps := d.getJSONKeyFree(ctx, base+"/api/ps", apiKey)
	if ps == nil {
		return
	}
	models, _ := ps["models"].([]any)
	entry := ollamaLoadedEntry(models, modelName)
	if entry == nil {
		tracef(ctx, "ollama /api/ps does not list the configured model as loaded; context stays as configured")
		return
	}
	n, ok := jsonPositiveInt(entry["context_length"])
	if !ok {
		return
	}
	prev := ev.Caps.Sources[apispec.CapFieldContext]
	if !ev.Caps.SetContext(n, srcOllamaLoadedCtx) {
		return
	}
	if prev == "" {
		tracef(ctx, "ollama /api/ps models.context_length = %d -> context (loaded)", n)
		return
	}
	tracef(ctx, "ollama /api/ps models.context_length = %d -> context (loaded beats %s)", n, prev)
}

// A name without a tag is the :latest tag,
// https://github.com/ollama/ollama/blob/v0.34.4/docs/api.md#model-names.
func ollamaLoadedEntry(models []any, modelName string) map[string]any {
	var candidates []string
	for _, name := range []string{modelName, bareModelID(modelName)} {
		if name == "" || containsString(candidates, name) {
			continue
		}
		candidates = append(candidates, name)
		if !strings.Contains(name, ":") {
			candidates = append(candidates, name+":latest")
		}
	}
	for _, item := range models {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		for _, field := range []string{"name", "model"} {
			if id, ok := m[field].(string); ok {
				for _, name := range candidates {
					if strings.EqualFold(id, name) {
						return m
					}
				}
			}
		}
	}
	return nil
}

// The parameters text is one "<key> <value>" line per Modelfile PARAMETER.
func ollamaNumCtx(parameters string) (int64, bool) {
	for _, line := range strings.Split(parameters, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[0] != "num_ctx" {
			continue
		}
		n, err := strconv.ParseInt(fields[1], 10, 64)
		return n, err == nil && n > 0
	}
	return 0, false
}

// Integers arrive as JSON numbers; a string, bool or fraction is not one.
func jsonPositiveInt(v any) (int64, bool) {
	f, ok := v.(float64)
	if !ok || f <= 0 || f != math.Trunc(f) || f >= math.MaxInt64 {
		return 0, false
	}
	return int64(f), true
}

func containsString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
