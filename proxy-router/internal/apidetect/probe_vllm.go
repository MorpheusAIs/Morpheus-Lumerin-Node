package apidetect

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

// Structured outputs are a built-in of the OpenAI-compatible server for every
// model it serves (json_object, json_schema, grammar, regex, choice):
// https://docs.vllm.ai/en/stable/features/structured_outputs.html
var vllmStructuredKinds = []string{"json_object", "json_schema", "grammar", "regex", "choice"}

// /version and /tokenize stay open when the server runs with --api-key; /v1/*
// is guarded: https://docs.vllm.ai/en/stable/serving/openai_compatible_server.html
// Tool calling is not discoverable outside VLLM_SERVER_DEV_MODE, so no tools
// fact is set here.
func (d *Detector) probeVLLM(ctx context.Context, base, modelName, apiKey string, ev *apispec.Evidence) bool {
	ver := d.getJSONKeyFree(ctx, base+"/version", apiKey)
	if ver == nil {
		return false
	}
	if _, ok := ver["version"]; !ok {
		return false
	}
	ev.Stack = "vllm"
	tracef(ctx, "identified vllm by /version")

	if ev.Caps.SetStructured(vllmStructuredKinds, "vllm:/version") {
		tracef(ctx, "vllm: /version -> capabilities.structured = %v (built-in structured outputs)", vllmStructuredKinds)
	}

	contextKnown := false
	if list := d.getJSON(ctx, base+"/v1/models", apiKey); list != nil {
		if data, ok := list["data"].([]any); ok {
			if entry := matchModelEntry(data, modelName); entry != nil {
				if id, ok := entry["id"].(string); ok {
					ev.ServedModelID = id
				}
				if n, ok := jsonPositiveInt(entry["max_model_len"]); ok && ev.Caps.SetContext(n, "vllm:/v1/models.max_model_len") {
					tracef(ctx, "vllm: /v1/models data[].max_model_len = %d -> capabilities.context", n)
					contextKnown = true
				}
			}
		}
	}
	if !contextKnown {
		d.probeVLLMTokenize(ctx, base, modelName, apiKey, ev)
	}
	return true
}

// POST /tokenize answers with the served window (max_model_len) and no
// generation: https://github.com/vllm-project/vllm/blob/v0.30.0/vllm/entrypoints/serve/tokenize/protocol.py
// The served id, when /v1/models gave one, is what the server will accept as
// "model"; the configured name is the fallback.
func (d *Detector) probeVLLMTokenize(ctx context.Context, base, modelName, apiKey string, ev *apispec.Evidence) {
	model := modelName
	if ev.ServedModelID != "" {
		model = ev.ServedModelID
	}
	body, err := json.Marshal(map[string]string{"model": model, "prompt": "x"})
	if err != nil {
		return
	}
	urlStr := base + "/tokenize"
	out, status := d.requestJSONStatus(ctx, http.MethodPost, urlStr, "", body, maxProbeBody)
	if out == nil && apiKey != "" && (status == http.StatusUnauthorized || status == http.StatusForbidden) {
		tracef(ctx, "POST %s -> HTTP %d without a key; retrying with the configured key", urlStr, status)
		out, _ = d.requestJSONStatus(ctx, http.MethodPost, urlStr, apiKey, body, maxProbeBody)
	}
	if out == nil {
		return
	}
	if n, ok := jsonPositiveInt(out["max_model_len"]); ok && ev.Caps.SetContext(n, "vllm:/tokenize.max_model_len") {
		tracef(ctx, "vllm: /tokenize max_model_len = %d -> capabilities.context", n)
	}
}
