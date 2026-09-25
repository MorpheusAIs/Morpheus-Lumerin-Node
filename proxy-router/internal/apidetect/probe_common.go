package apidetect

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
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

	resp, err := d.clientFor(ctx).Do(req)
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
	tracef(ctx, "%s %s -> HTTP %d, JSON object with keys %s", method, urlStr, resp.StatusCode, tracedKeys(mapKeys(out)))
	return out
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

func matchModelEntry(data []any, modelName string) map[string]any {
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
			if id, ok := m["id"].(string); ok && strings.EqualFold(id, name) {
				return m
			}
		}
	}
	if len(entries) == 1 {
		return entries[0]
	}
	for _, name := range candidates {
		for _, m := range entries {
			if id, ok := m["id"].(string); ok && strings.HasSuffix(strings.ToLower(id), "/"+strings.ToLower(name)) {
				return m
			}
		}
	}
	return nil
}
