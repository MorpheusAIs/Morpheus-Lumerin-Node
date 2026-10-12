package apidetect

import (
	"context"
	"encoding/json"
	"net/http"
	"path"
	"strings"

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
	servedAs := ""
	if list := d.getJSON(ctx, base+"/v1/models", apiKey); list != nil {
		if data, ok := list["data"].([]any); ok {
			if entry, bySpelling := findModelEntry(data, modelName); entry != nil {
				id, _ := entry["id"].(string)
				if bySpelling {
					traceSpelledEntry(ctx, "vllm /v1/models", modelName, id)
				}
				servedAs = id
				root, _ := entry["root"].(string)
				ev.ServedModelID = vllmWeightsName(ctx, id, root)
				if n, ok := jsonPositiveInt(entry["max_model_len"]); ok && ev.Caps.SetContext(n, "vllm:/v1/models.max_model_len") {
					tracef(ctx, "vllm: /v1/models data[].max_model_len = %d -> capabilities.context", n)
					contextKnown = true
				}
			}
		}
	}
	if !contextKnown {
		d.probeVLLMTokenize(ctx, base, modelName, servedAs, apiKey, ev)
	}
	return true
}

// vLLM's id is --served-model-name, which must equal the configured
// modelName, so it is the operator's label; root is the --model it loaded.
// A hub id (org/name) names the published weights and wins. A local path is
// the operator's choice as much as the label, so it wins only when it names a
// known family and says more than the label: /models/qwen3-235b under the
// label qwen3-235b-a22b-thinking-2507 does not. Only root's last segment is
// kept, so no directory reaches the trace.
func vllmWeightsName(ctx context.Context, id, root string) string {
	if root == "" || strings.EqualFold(root, id) {
		return id
	}
	weights := path.Base(strings.TrimRight(root, "/"))
	switch {
	case weights == "." || weights == "/" || apispec.FamilyFromName(weights) == "":
		tracef(ctx, "vllm serves %q from weights whose name maps to no known family; the served name stands", id)
		return id
	case localPath(root) && strings.Contains(apispec.NormalizeModelName(id), apispec.NormalizeModelName(weights)):
		tracef(ctx, "vllm serves %q from local weights %q, which say no more than the served name; the served name stands", id, weights)
		return id
	}
	tracef(ctx, "vllm serves %q from weights %q; the weights name the model", id, weights)
	return weights
}

// A hub id is exactly org/name; anything else vLLM was given is a path.
func localPath(root string) bool {
	return strings.HasPrefix(root, "/") || strings.HasPrefix(root, ".") || strings.HasPrefix(root, "~") || strings.Count(root, "/") != 1
}

// POST /tokenize answers with the served window (max_model_len) and no
// generation: https://github.com/vllm-project/vllm/blob/v0.30.0/vllm/entrypoints/serve/tokenize/protocol.py
// servedAs, the /v1/models id when it gave one, is what the server will
// accept as "model" (ServedModelID may name the weights instead); the
// configured name is the fallback.
func (d *Detector) probeVLLMTokenize(ctx context.Context, base, modelName, servedAs, apiKey string, ev *apispec.Evidence) {
	model := modelName
	if servedAs != "" {
		model = servedAs
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
