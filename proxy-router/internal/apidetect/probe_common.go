package apidetect

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
)

const maxProbeBody = 2 << 20

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

func (d *Detector) getJSON(ctx context.Context, urlStr, apiKey string) map[string]any {
	return d.requestJSON(ctx, http.MethodGet, urlStr, apiKey, nil, maxProbeBody)
}

// For endpoints the engine documents as open: the key is sent only after the
// backend demanded one (401/403), so a bearer never reaches a path that did
// not ask for it. Any other failure is final.
func (d *Detector) getJSONKeyFree(ctx context.Context, urlStr, apiKey string) map[string]any {
	return d.requestJSONKeyFree(ctx, http.MethodGet, urlStr, apiKey, nil)
}

// requestJSONKeyFree is getJSONKeyFree for any method and body (Ollama's
// /api/show is a POST on an API that has no auth of its own).
func (d *Detector) requestJSONKeyFree(ctx context.Context, method, urlStr, apiKey string, body []byte) map[string]any {
	out, status := d.requestJSONStatus(ctx, method, urlStr, "", body, maxProbeBody)
	if out != nil || apiKey == "" || (status != http.StatusUnauthorized && status != http.StatusForbidden) {
		return out
	}
	tracef(ctx, "%s %s -> HTTP %d without a key; retrying with the configured key", method, urlStr, status)
	out, _ = d.requestJSONStatus(ctx, method, urlStr, apiKey, body, maxProbeBody)
	return out
}

func (d *Detector) requestJSON(ctx context.Context, method, urlStr, apiKey string, body []byte, limit int64) map[string]any {
	out, _ := d.requestJSONStatus(ctx, method, urlStr, apiKey, body, limit)
	return out
}

// The status is 0 when no response arrived.
func (d *Detector) requestJSONStatus(ctx context.Context, method, urlStr, apiKey string, body []byte, limit int64) (map[string]any, int) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, urlStr, reader)
	if err != nil {
		return nil, 0
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := d.clientFor(ctx).Do(req)
	if err != nil {
		tracef(ctx, "%s %s -> %v", method, urlStr, err)
		return nil, 0
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		tracef(ctx, "%s %s -> HTTP %d", method, urlStr, resp.StatusCode)
		return nil, resp.StatusCode
	}

	limited := &countingReader{r: io.LimitReader(resp.Body, limit)}
	var out map[string]any
	if err := json.NewDecoder(limited).Decode(&out); err != nil {
		if limited.n >= limit {
			tracef(ctx, "%s %s -> HTTP %d, but the body exceeds the %d-byte limit; ignored", method, urlStr, resp.StatusCode, limit)
		} else {
			tracef(ctx, "%s %s -> HTTP %d, but not a JSON object: %v", method, urlStr, resp.StatusCode, err)
		}
		return nil, resp.StatusCode
	}
	tracef(ctx, "%s %s -> HTTP %d, JSON object with keys %s", method, urlStr, resp.StatusCode, tracedKeys(mapKeys(out)))
	return out, resp.StatusCode
}

type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

func mapKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

const maxTracedKeys = 20

// Keys are quoted: a third-party key may carry control characters or
// formatting verbs.
func tracedKeys(keys []string) string {
	truncated := len(keys) > maxTracedKeys
	if truncated {
		keys = keys[:maxTracedKeys]
	}
	quoted := make([]string, len(keys))
	for i, k := range keys {
		quoted[i] = strconv.Quote(k)
	}
	joined := strings.Join(quoted, " ")
	if truncated {
		joined += " …"
	}
	return "[" + joined + "]"
}

func (d *Detector) fingerprintEngines(ctx context.Context, bases []string, modelName, apiKey string, ev *apispec.Evidence, introspectLiteLLM bool) {
	for _, base := range bases {
		if d.probeOllama(ctx, base, modelName, apiKey, ev) ||
			d.probeLlamaCpp(ctx, base, apiKey, ev) ||
			d.probeSGLang(ctx, base, apiKey, ev) ||
			d.probeLMStudio(ctx, base, modelName, apiKey, ev) ||
			d.probeKoboldCpp(ctx, base, apiKey, ev) ||
			d.probeTGI(ctx, base, apiKey, ev) ||
			d.probeLiteLLM(ctx, base, modelName, apiKey, ev, introspectLiteLLM) ||
			d.probeVLLM(ctx, base, modelName, apiKey, ev) {
			return
		}
	}
}

func (d *Detector) getText(ctx context.Context, urlStr, apiKey string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return ""
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	resp, err := d.clientFor(ctx).Do(req)
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

// findModelEntry finds the listing entry for modelName by its "id" field.
func findModelEntry(data []any, modelName string) (map[string]any, bool) {
	return findModelEntryBy(data, modelName, "id")
}

// findModelEntryBy is the listing matcher for self-hosted engines and model
// registries: the configured name, then its bare id (inline parameters
// stripped), each exactly (case-insensitive), then a lone entry, then a
// "<prefix>/<name>" suffix match, and last an id that names the model once
// separators, case and inline parameters are ignored, counted only when it is
// the only one (a LiteLLM deployment of google.gemma-4-31b-it reaches Venice,
// which lists google-gemma-4-31b-it). bySpelling reports that last kind.
// field names the identifier key ("id" for OpenAI-shaped listings, "key" for
// LM Studio v1). Hosted listings go through vendorEntry, which keeps only the
// exact-id rule.
func findModelEntryBy(data []any, modelName, field string) (entry map[string]any, bySpelling bool) {
	var entries []map[string]any
	for _, item := range data {
		if m, ok := item.(map[string]any); ok {
			entries = append(entries, m)
		}
	}

	bare := bareModelID(modelName)
	candidates := []string{modelName}
	if bare != modelName {
		candidates = append(candidates, bare)
	}

	for _, name := range candidates {
		for _, m := range entries {
			if id, ok := m[field].(string); ok && strings.EqualFold(id, name) {
				return m, false
			}
		}
	}
	if len(entries) == 1 {
		return entries[0], false
	}
	for _, name := range candidates {
		for _, m := range entries {
			if id, ok := m[field].(string); ok && strings.HasSuffix(strings.ToLower(id), "/"+strings.ToLower(name)) {
				return m, false
			}
		}
	}

	want := apispec.NormalizeModelName(modelName)
	if want == "" {
		return nil, false
	}
	matches := 0
	for _, m := range entries {
		if id, ok := m[field].(string); ok {
			if norm := apispec.NormalizeModelName(id); norm == want || strings.HasSuffix(norm, "/"+want) {
				entry = m
				matches++
			}
		}
	}
	if matches != 1 {
		return nil, false
	}
	return entry, true
}

func traceSpelledEntry(ctx context.Context, listing, modelName, id string) {
	tracef(ctx, "%s lists %q as %q (same name in normalized spelling)", listing, bareModelID(modelName), id)
}

// jsonWholeNumber reads a decoded JSON number (float64) as an integer: only
// a whole value a float64 represents exactly (|v| <= 2^53) converts, so the
// int64 conversion is never implementation-defined. Zero and negatives pass:
// a probe that traces "out of range; ignored" for them gets the value and
// leaves the (0, 1<<31) bound to the capability setter; llama.cpp's ftype
// enum starts at 0.
func jsonWholeNumber(v any) (int64, bool) {
	f, ok := v.(float64)
	if !ok || f != math.Trunc(f) || math.Abs(f) > 1<<53 {
		return 0, false
	}
	return int64(f), true
}

// jsonPositiveInt is jsonWholeNumber within the capability bound
// (0, config.MaxCapabilityInt): the value a token-count field must carry
// before it reaches a subtraction, a minimum or a setter.
func jsonPositiveInt(v any) (int64, bool) {
	n, ok := jsonWholeNumber(v)
	if !ok || !config.ValidCapabilityInt(n) {
		return 0, false
	}
	return n, true
}

// jsonStrings is the one reader of a decoded JSON array's string members:
// non-strings are skipped, and a missing, non-array or all-non-string value
// reads as nil. A probe that must distinguish "empty" from "unreadable"
// (LiteLLM's supported_openai_params) tests the array itself first.
func jsonStrings(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func containsString(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}
