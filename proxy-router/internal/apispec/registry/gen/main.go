// Command gen writes registry_data.go, the offline model registry snapshot, from
// models.dev and LiteLLM's model_prices_and_context_window.json. Output is
// deterministic for a given input and -date, so a re-run on the same day is
// byte-identical.
//
// Sources (both MIT):
//   - https://models.dev/api.json — https://github.com/anomalyco/models.dev
//   - https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json — https://github.com/BerriAI/litellm
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	modelsDevURL = "https://models.dev/api.json"
	liteLLMURL   = "https://raw.githubusercontent.com/BerriAI/litellm/main/model_prices_and_context_window.json"

	// maxBody is the detector's registry-body bound (16 MiB).
	maxBody = 16 << 20
	// maxSource caps the generated Go file (spec §6: ≤ 1.5 MB of source).
	maxSource = 1_500_000

	srcModelsDev = "srcModelsDev"
	srcLiteLLM   = "srcLiteLLM"
	srcBoth      = "srcBoth"
)

// Vocabulary of config.ModalityInputs / config.ModalityOutputs; anything else
// is dropped here so the snapshot never carries a value the composer rejects.
var (
	inputModalities  = map[string]bool{"text": true, "image": true, "audio": true, "video": true, "pdf": true}
	outputModalities = map[string]bool{"text": true, "image": true, "audio": true, "embeddings": true}
)

type entry struct {
	ID          string
	Context     int64
	Output      int64
	Reasoning   *bool
	ToolCall    *bool
	Structured  *bool
	WebSearch   *bool
	PromptCache *bool
	Attachment  *bool
	In, Out     []string
	Src         string
	// prefix is the provider prefix normalize stripped from the source id;
	// vendor is the provider that listed it. Both drive conflict resolution
	// only.
	prefix string
	vendor string
}

func main() {
	date := flag.String("date", time.Now().UTC().Format("2006-01-02"), "snapshot date written to the header (YYYY-MM-DD)")
	// Not data.go: proxy-router/.gitignore ignores data* (the badger dir).
	out := flag.String("out", "registry_data.go", "output file, relative to the package directory")
	modelsDevPath := flag.String("modelsdev", "", "read models.dev api.json from this file instead of the network")
	liteLLMPath := flag.String("litellm", "", "read the LiteLLM json from this file instead of the network")
	flag.Parse()
	if _, err := time.Parse("2006-01-02", *date); err != nil {
		log.Fatalf("-date must be YYYY-MM-DD: %v", err)
	}

	md, err := load(*modelsDevPath, modelsDevURL)
	if err != nil {
		log.Fatalf("models.dev: %v", err)
	}
	ll, err := load(*liteLLMPath, liteLLMURL)
	if err != nil {
		log.Fatalf("litellm: %v", err)
	}

	merged, err := merge(md, ll)
	if err != nil {
		log.Fatal(err)
	}
	src, dropped, err := render(merged, *date, maxSource)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(*out, src, 0o644); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote %s: %d entries, %d bytes%s", *out, len(merged)-dropped.count(), len(src), dropped.suffix())
}

func load(path, url string) ([]byte, error) {
	if path != "" {
		return os.ReadFile(path)
	}
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		if attempt > 1 {
			time.Sleep(2 * time.Second)
		}
		body, err := fetch(url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		log.Printf("attempt %d for %s failed: %v", attempt, url, err)
	}
	return nil, lastErr
}

func fetch(url string) ([]byte, error) {
	client := &http.Client{Timeout: 90 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "morpheus-proxy-router registry generator")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
	if err != nil {
		return nil, err
	}
	if len(body) > maxBody {
		return nil, fmt.Errorf("body exceeds %d bytes", maxBody)
	}
	return body, nil
}

// models.dev: {"<provider>": {"id", "models": {"<id>": {...}}}}.
type mdProvider struct {
	ID     string             `json:"id"`
	Models map[string]mdModel `json:"models"`
}

type mdModel struct {
	Attachment       *bool `json:"attachment"`
	Reasoning        *bool `json:"reasoning"`
	ToolCall         *bool `json:"tool_call"`
	StructuredOutput *bool `json:"structured_output"`
	Modalities       struct {
		Input  []string `json:"input"`
		Output []string `json:"output"`
	} `json:"modalities"`
	Limit struct {
		Context json.Number `json:"context"`
		Output  json.Number `json:"output"`
	} `json:"limit"`
}

// LiteLLM: {"<key>": {"mode", "litellm_provider", "max_input_tokens", "supports_*"}}.
type llModel struct {
	Mode            string      `json:"mode"`
	Provider        string      `json:"litellm_provider"`
	MaxInput        json.Number `json:"max_input_tokens"`
	MaxOutput       json.Number `json:"max_output_tokens"`
	FunctionCalling *bool       `json:"supports_function_calling"`
	Vision          *bool       `json:"supports_vision"`
	Reasoning       *bool       `json:"supports_reasoning"`
	ResponseSchema  *bool       `json:"supports_response_schema"`
	WebSearch       *bool       `json:"supports_web_search"`
	PromptCaching   *bool       `json:"supports_prompt_caching"`
	PDFInput        *bool       `json:"supports_pdf_input"`
	AudioInput      *bool       `json:"supports_audio_input"`
	AudioOutput     *bool       `json:"supports_audio_output"`
	VideoInput      *bool       `json:"supports_video_input"`
}

func merge(modelsDev, liteLLM []byte) ([]entry, error) {
	var providers map[string]mdProvider
	if err := json.Unmarshal(modelsDev, &providers); err != nil {
		return nil, fmt.Errorf("models.dev: %w", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(liteLLM, &raw); err != nil {
		return nil, fmt.Errorf("litellm: %w", err)
	}
	// The LiteLLM file carries a "sample_spec" documentation key whose
	// values are strings; a key that does not decode is skipped.
	ll := map[string]llModel{}
	for key, msg := range raw {
		var m llModel
		if err := json.Unmarshal(msg, &m); err != nil {
			continue
		}
		if m.Mode == "chat" || m.Mode == "responses" {
			ll[key] = m
		}
	}

	prefixes := map[string]bool{}
	for id := range providers {
		prefixes[canonicalProvider(id)] = true
	}
	for _, m := range ll {
		if m.Provider != "" {
			prefixes[canonicalProvider(m.Provider)] = true
		}
	}

	byID := map[string][]entry{}
	for _, pid := range sortedKeys(providers) {
		p := providers[pid]
		for _, mid := range sortedKeys(p.Models) {
			id, prefix := normalize(mid, prefixes)
			if id == "" {
				continue
			}
			byID[id] = append(byID[id], fromModelsDev(id, pid, prefix, p.Models[mid]))
		}
	}
	llByID := map[string][]entry{}
	for _, key := range sortedKeys(ll) {
		id, prefix := normalize(key, prefixes)
		if id == "" {
			continue
		}
		llByID[id] = append(llByID[id], fromLiteLLM(id, prefix, ll[key]))
	}

	var out []entry
	for _, id := range sortedKeys(byID) {
		e := resolve(byID[id])
		if lls, ok := llByID[id]; ok {
			e = fill(e, resolve(lls))
			e.Src = srcBoth
		}
		out = append(out, e)
	}
	for _, id := range sortedKeys(llByID) {
		if _, ok := byID[id]; ok {
			continue
		}
		out = append(out, resolve(llByID[id]))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	log.Printf("merged %d ids (%d listed by models.dev, %d litellm-only)", len(out), len(byID), len(out)-len(byID))
	return out, nil
}

// canonicalProvider makes "fireworks_ai" and "fireworks-ai" the same prefix.
func canonicalProvider(id string) string {
	return strings.ReplaceAll(strings.ToLower(id), "_", "-")
}

// normalize lowercases, strips one leading "<provider>/" when the prefix is
// a known provider id (a Hugging Face org like "meta-llama/" stays) and drops
// any ":" segment carrying "=" (a parameter; Ollama-style ":q4" and
// OpenRouter ":free" are tags and stay). The second result is the stripped
// prefix, canonical, or "".
func normalize(id string, prefixes map[string]bool) (string, string) {
	id = strings.ToLower(strings.TrimSpace(id))
	prefix := ""
	if i := strings.IndexByte(id, '/'); i > 0 && prefixes[canonicalProvider(id[:i])] {
		prefix = canonicalProvider(id[:i])
		id = id[i+1:]
	}
	id = stripParamSegments(id)
	if id == "" || id == "/" {
		return "", ""
	}
	return id, prefix
}

func stripParamSegments(id string) string {
	if !strings.Contains(id, "=") {
		return id
	}
	parts := strings.Split(id, ":")
	kept := parts[:1]
	for _, p := range parts[1:] {
		if !strings.Contains(p, "=") {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, ":")
}

func fromModelsDev(id, provider, prefix string, m mdModel) entry {
	return entry{
		ID:         id,
		Context:    num(m.Limit.Context),
		Output:     num(m.Limit.Output),
		Reasoning:  m.Reasoning,
		ToolCall:   m.ToolCall,
		Structured: m.StructuredOutput,
		Attachment: m.Attachment,
		In:         keep(m.Modalities.Input, inputModalities),
		Out:        keep(m.Modalities.Output, outputModalities),
		Src:        srcModelsDev,
		prefix:     prefix,
		vendor:     canonicalProvider(provider),
	}
}

func fromLiteLLM(id, prefix string, m llModel) entry {
	in := []string{"text"}
	if isTrue(m.Vision) {
		in = append(in, "image")
	}
	if isTrue(m.AudioInput) {
		in = append(in, "audio")
	}
	if isTrue(m.VideoInput) {
		in = append(in, "video")
	}
	if isTrue(m.PDFInput) {
		in = append(in, "pdf")
	}
	out := []string{"text"}
	if isTrue(m.AudioOutput) {
		out = append(out, "audio")
	}
	return entry{
		ID:          id,
		Context:     num(m.MaxInput),
		Output:      num(m.MaxOutput),
		Reasoning:   m.Reasoning,
		ToolCall:    m.FunctionCalling,
		Structured:  m.ResponseSchema,
		WebSearch:   m.WebSearch,
		PromptCache: m.PromptCaching,
		Attachment:  m.PDFInput,
		In:          in,
		Out:         out,
		Src:         srcLiteLLM,
		prefix:      prefix,
		vendor:      canonicalProvider(m.Provider),
	}
}

// resolve picks one listing per id from one source and lets the others fill
// its blanks. Preference: the vendor of record (the provider that the most
// other listings carried as their prefix — "openai/gpt-4o" on openrouter and
// eleven other aggregators names "openai"; one "azure/gpt-4o" does not
// outrank it), then listings without a prefix, then provider id order: a
// vendor's own limits beat an aggregator's.
func resolve(listings []entry) entry {
	if len(listings) == 1 {
		return listings[0]
	}
	named := map[string]int{}
	for _, e := range listings {
		if e.prefix != "" {
			named[e.prefix]++
		}
	}
	rank := func(e entry) (int, int) {
		switch {
		case e.prefix != "":
			return 2, 0
		case named[e.vendor] > 0:
			return 0, -named[e.vendor]
		}
		return 1, 0
	}
	sort.SliceStable(listings, func(i, j int) bool {
		ri, ni := rank(listings[i])
		rj, nj := rank(listings[j])
		if ri != rj {
			return ri < rj
		}
		if ni != nj {
			return ni < nj
		}
		return listings[i].vendor < listings[j].vendor
	})
	base := listings[0]
	for _, other := range listings[1:] {
		base = fill(base, other)
	}
	return base
}

// fill copies other's fields into base's blanks; a stated value in base is
// never overridden, so the preferred listing (and models.dev over LiteLLM)
// wins on conflicts.
func fill(base, other entry) entry {
	if base.Context == 0 {
		base.Context = other.Context
	}
	if base.Output == 0 {
		base.Output = other.Output
	}
	fillBool := func(dst **bool, v *bool) {
		if *dst == nil {
			*dst = v
		}
	}
	fillBool(&base.Reasoning, other.Reasoning)
	fillBool(&base.ToolCall, other.ToolCall)
	fillBool(&base.Structured, other.Structured)
	fillBool(&base.WebSearch, other.WebSearch)
	fillBool(&base.PromptCache, other.PromptCache)
	fillBool(&base.Attachment, other.Attachment)
	if len(base.In) == 0 {
		base.In = other.In
	}
	if len(base.Out) == 0 {
		base.Out = other.Out
	}
	return base
}

func (e entry) hasFlag() bool {
	return isTrue(e.Reasoning) || isTrue(e.ToolCall) || isTrue(e.Structured) ||
		isTrue(e.WebSearch) || isTrue(e.PromptCache) || isTrue(e.Attachment)
}

type dropReport struct {
	liteLLMOnly int
	flagless    int
}

func (d dropReport) count() int { return d.liteLLMOnly + d.flagless }

func (d dropReport) suffix() string {
	if d.count() == 0 {
		return ""
	}
	return fmt.Sprintf(" (dropped %d litellm-only, %d flagless for the size cap)", d.liteLLMOnly, d.flagless)
}

// render writes the Go source; over cap bytes it drops LiteLLM-only entries
// first, then entries stating no capability, and records that in the header.
func render(entries []entry, date string, cap int) ([]byte, dropReport, error) {
	var dropped dropReport
	src, err := renderEntries(entries, date, dropped)
	if err != nil {
		return nil, dropped, err
	}
	if len(src) > cap {
		entries, dropped.liteLLMOnly = without(entries, func(e entry) bool { return e.Src == srcLiteLLM })
		if src, err = renderEntries(entries, date, dropped); err != nil {
			return nil, dropped, err
		}
	}
	if len(src) > cap {
		entries, dropped.flagless = without(entries, func(e entry) bool { return !e.hasFlag() })
		if src, err = renderEntries(entries, date, dropped); err != nil {
			return nil, dropped, err
		}
	}
	if len(src) > cap {
		return nil, dropped, fmt.Errorf("snapshot is %d bytes after dropping %d entries; cap is %d", len(src), dropped.count(), cap)
	}
	return src, dropped, nil
}

func without(entries []entry, drop func(entry) bool) ([]entry, int) {
	kept := make([]entry, 0, len(entries))
	n := 0
	for _, e := range entries {
		if drop(e) {
			n++
			continue
		}
		kept = append(kept, e)
	}
	return kept, n
}

func renderEntries(entries []entry, date string, dropped dropReport) ([]byte, error) {
	counts := map[string]int{}
	for _, e := range entries {
		counts[e.Src]++
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated by go run ./gen -date=%s; DO NOT EDIT.\n", date)
	b.WriteString("//\n")
	fmt.Fprintf(&b, "// Snapshot %s of\n", date)
	fmt.Fprintf(&b, "//   %s (MIT, https://github.com/anomalyco/models.dev)\n", modelsDevURL)
	fmt.Fprintf(&b, "//   %s (MIT, https://github.com/BerriAI/litellm)\n", liteLLMURL)
	fmt.Fprintf(&b, "// Entries: %d (models.dev only %d, both %d, litellm only %d).\n",
		len(entries), counts[srcModelsDev], counts[srcBoth], counts[srcLiteLLM])
	if dropped.count() == 0 {
		b.WriteString("// Dropped for the size cap: none.\n")
	} else {
		fmt.Fprintf(&b, "// Dropped for the size cap: %d litellm-only entries, %d entries stating no capability.\n",
			dropped.liteLLMOnly, dropped.flagless)
	}
	b.WriteString("\npackage registry\n\n")
	fmt.Fprintf(&b, "const snapshotDate = %q\n\n", date)
	b.WriteString("var entries = []entry{\n")
	for _, e := range entries {
		writeEntry(&b, e)
	}
	b.WriteString("}\n")
	return format.Source(b.Bytes())
}

func writeEntry(b *bytes.Buffer, e entry) {
	fmt.Fprintf(b, "\t{ID: %q", e.ID)
	if e.Context > 0 {
		fmt.Fprintf(b, ", Context: %d", e.Context)
	}
	if e.Output > 0 {
		fmt.Fprintf(b, ", Output: %d", e.Output)
	}
	for _, f := range []struct {
		name string
		v    *bool
	}{
		{"Reasoning", e.Reasoning}, {"ToolCall", e.ToolCall}, {"Structured", e.Structured},
		{"WebSearch", e.WebSearch}, {"PromptCache", e.PromptCache}, {"Attachment", e.Attachment},
	} {
		if isTrue(f.v) {
			fmt.Fprintf(b, ", %s: true", f.name)
		}
	}
	if len(e.In) > 0 {
		fmt.Fprintf(b, ", In: %s", goStrings(e.In))
	}
	if len(e.Out) > 0 {
		fmt.Fprintf(b, ", Out: %s", goStrings(e.Out))
	}
	fmt.Fprintf(b, ", Src: %s},\n", e.Src)
}

func goStrings(list []string) string {
	quoted := make([]string, len(list))
	for i, s := range list {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return "[]string{" + strings.Join(quoted, ", ") + "}"
}

// num accepts only integers in (0, 1<<31), the bound the composer enforces.
func num(n json.Number) int64 {
	v, err := n.Int64()
	if err != nil {
		f, ferr := n.Float64()
		if ferr != nil || f != float64(int64(f)) {
			return 0
		}
		v = int64(f)
	}
	if v <= 0 || v >= 1<<31 {
		return 0
	}
	return v
}

func isTrue(b *bool) bool { return b != nil && *b }

func keep(list []string, vocab map[string]bool) []string {
	var out []string
	for _, v := range list {
		v = strings.ToLower(strings.TrimSpace(v))
		if vocab[v] && !contains(out, v) {
			out = append(out, v)
		}
	}
	return out
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
