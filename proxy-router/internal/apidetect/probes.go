package apidetect

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

// maxProbeBody caps how much of a service-endpoint probe response is read.
const maxProbeBody = 2 << 20 // 2 MB

// maxRegistryBody caps a registry model listing, which grows with the
// vendor's catalogue (OpenRouter's is ~0.7 MB today). A var so tests can
// exercise the truncation path.
var maxRegistryBody int64 = 16 << 20 // 16 MB

// baseCandidates derives base URLs to probe from the configured endpoint URL:
// the endpoint with known inference suffixes stripped, then the bare origin.
func baseCandidates(apiURL string) []string {
	u, err := url.Parse(apiURL)
	if err != nil || u.Host == "" {
		return nil
	}
	origin := u.Scheme + "://" + u.Host

	path := strings.TrimSuffix(u.Path, "/")
	for _, suffix := range []string{"/chat/completions", "/completions", "/embeddings", "/messages"} {
		path = strings.TrimSuffix(path, suffix)
	}
	path = strings.TrimSuffix(path, "/")

	var bases []string
	if path != "" {
		bases = append(bases, origin+path)
	}
	if len(bases) == 0 || bases[0] != origin {
		bases = append(bases, origin)
	}
	return bases
}

// getJSON fetches urlStr and decodes the JSON object response into a map.
// Returns nil on any transport error, non-2xx status or non-object payload.
func (d *Detector) getJSON(ctx context.Context, urlStr, apiKey string) map[string]any {
	return d.requestJSON(ctx, http.MethodGet, urlStr, apiKey, nil, maxProbeBody)
}

func (d *Detector) requestJSON(ctx context.Context, method, urlStr, apiKey string, body []byte, limit int64) map[string]any {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, urlStr, reader)
	if err != nil {
		return nil
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		tracef(ctx, "%s %s -> %v", method, urlStr, err)
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		tracef(ctx, "%s %s -> HTTP %d", method, urlStr, resp.StatusCode)
		return nil
	}

	limited := &countingReader{r: io.LimitReader(resp.Body, limit)}
	var out map[string]any
	if err := json.NewDecoder(limited).Decode(&out); err != nil {
		if limited.n >= limit {
			tracef(ctx, "%s %s -> HTTP %d, but the body exceeds the %d-byte limit; ignored", method, urlStr, resp.StatusCode, limit)
		} else {
			tracef(ctx, "%s %s -> HTTP %d, but not a JSON object: %v", method, urlStr, resp.StatusCode, err)
		}
		return nil
	}
	tracef(ctx, "%s %s -> HTTP %d, JSON object with keys %v", method, urlStr, resp.StatusCode, mapKeys(out))
	return out
}

// countingReader counts bytes read so a decode failure can be attributed
// to truncation rather than malformed JSON.
type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

// mapKeys returns the object's top-level keys, sorted, for trace output.
func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// fingerprintEngines identifies a self-hosted serving stack by probing each
// engine's distinctive service endpoint, most distinctive first. It fills the
// stack and whatever evidence that engine exposes.
func (d *Detector) fingerprintEngines(ctx context.Context, bases []string, modelName, apiKey string, ev *apispec.Evidence) {
	for _, base := range bases {
		if d.probeOllama(ctx, base, modelName, apiKey, ev) ||
			d.probeLlamaCpp(ctx, base, apiKey, ev) ||
			d.probeSGLang(ctx, base, apiKey, ev) ||
			d.probeLMStudio(ctx, base, modelName, apiKey, ev) ||
			d.probeKoboldCpp(ctx, base, apiKey, ev) ||
			d.probeTGI(ctx, base, apiKey, ev) ||
			d.probeLiteLLM(ctx, base, modelName, apiKey, ev) ||
			d.probeVLLM(ctx, base, modelName, apiKey, ev) {
			return
		}
	}
}

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

func (d *Detector) probeSGLang(ctx context.Context, base, apiKey string, ev *apispec.Evidence) bool {
	info := d.getJSON(ctx, base+"/get_model_info", apiKey)
	if info == nil {
		return false
	}
	path, ok := info["model_path"].(string)
	if !ok {
		return false
	}
	ev.Stack = "sglang"
	ev.ServedModelID = path
	tracef(ctx, "identified sglang by /get_model_info; served model path %q", path)
	return true
}

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

func (d *Detector) probeKoboldCpp(ctx context.Context, base, apiKey string, ev *apispec.Evidence) bool {
	ver := d.getJSON(ctx, base+"/api/extra/version", apiKey)
	if ver == nil {
		return false
	}
	if result, ok := ver["result"].(string); !ok || !strings.EqualFold(result, "koboldcpp") {
		return false
	}
	ev.Stack = "koboldcpp"
	tracef(ctx, "identified koboldcpp by /api/extra/version")
	return true
}

// getText fetches urlStr and returns the body as text (bounded), or "" on
// transport error / non-2xx.
func (d *Detector) getText(ctx context.Context, urlStr, apiKey string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return ""
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := d.client.Do(req)
	if err != nil {
		tracef(ctx, "GET %s -> %v", urlStr, err)
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		tracef(ctx, "GET %s -> HTTP %d", urlStr, resp.StatusCode)
		return ""
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return ""
	}
	text := strings.TrimSpace(string(body))
	tracef(ctx, "GET %s -> HTTP %d, %d bytes", urlStr, resp.StatusCode, len(text))
	return text
}

// probeLiteLLM recognizes a LiteLLM proxy by its unauthenticated liveliness
// endpoint, whose body is the JSON string "I'm alive!", then imports the
// model group's declared capabilities (needs the configured key).
// https://docs.litellm.ai/docs/proxy/health
// https://docs.litellm.ai/docs/proxy/model_management
func (d *Detector) probeLiteLLM(ctx context.Context, base, modelName, apiKey string, ev *apispec.Evidence) bool {
	body := d.getText(ctx, base+"/health/liveliness", apiKey)
	if body == "" {
		return false
	}
	alive := strings.Trim(body, `"`)
	if !strings.EqualFold(alive, "I'm alive!") {
		return false
	}
	ev.Stack = "litellm"
	tracef(ctx, "identified litellm by /health/liveliness")

	if modelName == "" {
		return true
	}
	info := d.getJSON(ctx, base+"/model_group/info?model_group="+url.QueryEscape(modelName), apiKey)
	if info == nil {
		return true
	}
	group := info
	if data, ok := info["data"].([]any); ok {
		if len(data) == 0 {
			return true
		}
		if group, ok = data[0].(map[string]any); !ok {
			return true
		}
	}
	if params, ok := group["supported_openai_params"].([]any); ok {
		for _, p := range params {
			if s, ok := p.(string); ok {
				ev.Parameters = append(ev.Parameters, s)
			}
		}
		sort.Strings(ev.Parameters)
		tracef(ctx, "litellm /model_group/info lists %d supported_openai_params", len(ev.Parameters))
	}
	if reasons, ok := group["supports_reasoning"].(bool); ok && reasons {
		ev.GatewayReasoning = true
		tracef(ctx, "litellm /model_group/info reports supports_reasoning")
	}
	if efforts, ok := group["supported_reasoning_efforts"].([]any); ok {
		for _, e := range efforts {
			if s, ok := e.(string); ok {
				ev.ReasoningEfforts = append(ev.ReasoningEfforts, s)
			}
		}
		if len(ev.ReasoningEfforts) > 0 {
			tracef(ctx, "litellm /model_group/info lists supported_reasoning_efforts %v", ev.ReasoningEfforts)
		}
	}
	return true
}

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

// matchModelEntry finds the model list entry whose id equals modelName; when
// there is no exact match, a single-entry list is unambiguous, and otherwise
// an entry whose id ends with "/<modelName>" is accepted.
func matchModelEntry(data []any, modelName string) map[string]any {
	var entries []map[string]any
	for _, item := range data {
		if m, ok := item.(map[string]any); ok {
			entries = append(entries, m)
		}
	}
	for _, m := range entries {
		if id, ok := m["id"].(string); ok && strings.EqualFold(id, modelName) {
			return m
		}
	}
	if len(entries) == 1 {
		return entries[0]
	}
	for _, m := range entries {
		if id, ok := m["id"].(string); ok && strings.HasSuffix(strings.ToLower(id), "/"+strings.ToLower(modelName)) {
			return m
		}
	}
	return nil
}

// veniceCapabilityNames maps Venice capability flags to normalized parameter
// names shared with the OpenRouter import.
var veniceCapabilityNames = map[string]string{
	"supportsReasoning":       "reasoning",
	"supportsFunctionCalling": "tools",
	"supportsVision":          "vision",
	"supportsWebSearch":       "web_search",
	"supportsResponseSchema":  "response_format",
	"supportsLogProbs":        "logprobs",
}

// probeRegistryShape detects registry-style hosted APIs by the shape of their
// model listing: OpenRouter entries carry supported_parameters, Venice entries
// carry model_spec.capabilities. It imports the declared parameter list and,
// for OpenRouter, the unified reasoning knob.
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
					// OpenRouter's unified reasoning object; the knobs are the
					// openrouter stack table, this only says the model reasons.
					// https://openrouter.ai/docs/use-cases/reasoning-tokens
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
					// The knobs themselves come from the venice stack table.
					ev.GatewayReasoning = true
					tracef(ctx, "venice listing reports supportsReasoning")
				}
			}
			return
		}
	}
}
