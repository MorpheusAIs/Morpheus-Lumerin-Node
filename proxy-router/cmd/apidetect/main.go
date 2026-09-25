// Command apidetect runs backend API detection for every model in a
// models-config.json and prints, per model, what was detected and the probe
// trace that led to it.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/cmd/internal/modelsfile"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apidetect"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

func formatResult(api *system.ModelApiSpec, trace []string) string {
	var b strings.Builder

	if api == nil {
		b.WriteString("  result: nothing detected\n")
	} else {
		b.WriteString("  result:\n")
		fmt.Fprintf(&b, "    stack:       %s\n", orDash(api.Stack))
		if api.Via != "" {
			fmt.Fprintf(&b, "    via:         %s\n", api.Via)
		}
		fmt.Fprintf(&b, "    modelFamily: %s\n", orDash(api.ModelFamily))
		fmt.Fprintf(&b, "    source:      %s\n", orDash(api.Source))
		thinking := "unknown"
		if api.Thinking != nil {
			thinking = api.Thinking.Mode
			if api.Thinking.Default != "" {
				thinking += " (default " + api.Thinking.Default + ")"
			}
		}
		fmt.Fprintf(&b, "    thinking:    %s\n", thinking)
		writeCapabilities(&b, api.Capabilities)
		if len(api.Bindings) > 0 {
			b.WriteString("    bindings:\n")
			for _, intent := range sortedKeys(api.Bindings) {
				binding := api.Bindings[intent]
				fmt.Fprintf(&b, "      %-18s %s%s\n", intent+":", apispec.DescribeBinding(binding), describeVerified(binding, api.Evidence[system.EvidenceKeyBinding(intent)]))
			}
		}
		if len(api.Parameters) > 0 {
			fmt.Fprintf(&b, "    parameters:  %s\n", strings.Join(api.Parameters, ", "))
		}
		if len(api.Evidence) > 0 {
			b.WriteString("    evidence:\n")
			for _, key := range sortedKeys(api.Evidence) {
				fmt.Fprintf(&b, "      %s -> %s\n", key, api.Evidence[key])
			}
		}
	}

	b.WriteString("  how:\n")
	if len(trace) == 0 {
		b.WriteString("    (no steps recorded)\n")
	}
	for _, line := range trace {
		fmt.Fprintf(&b, "    - %s\n", line)
	}
	return b.String()
}

// Field order follows the wire contract (system.ModelCapabilities); an
// absent field is unknown and is not printed.
func writeCapabilities(b *strings.Builder, c *system.ModelCapabilities) {
	if c == nil {
		return
	}
	var lines []string
	addInt := func(name string, v *int64) {
		if v != nil {
			lines = append(lines, fmt.Sprintf("%-18s %d", name+":", *v))
		}
	}
	addBool := func(name string, v *bool) {
		if v != nil {
			lines = append(lines, fmt.Sprintf("%-18s %t", name+":", *v))
		}
	}
	addList := func(name string, v []string) {
		if len(v) > 0 {
			lines = append(lines, fmt.Sprintf("%-18s %s", name+":", strings.Join(v, ", ")))
		}
	}
	addInt("context", c.Context)
	addInt("contextTrained", c.ContextTrained)
	addInt("maxOutput", c.MaxOutput)
	if c.Modalities != nil {
		addList("modalities.input", c.Modalities.Input)
		addList("modalities.output", c.Modalities.Output)
	}
	addBool("tools", c.Tools)
	addBool("toolsParallel", c.ToolsParallel)
	addList("structured", c.Structured)
	addBool("reasoning", c.Reasoning)
	addBool("webSearch", c.WebSearch)
	addBool("embeddings", c.Embeddings)
	addBool("promptCache", c.PromptCache)
	if c.Quantization != "" {
		lines = append(lines, fmt.Sprintf("%-18s %s", "quantization:", c.Quantization))
	}
	if len(lines) == 0 {
		return
	}
	b.WriteString("    capabilities:\n")
	for _, line := range lines {
		fmt.Fprintf(b, "      %s\n", line)
	}
}

// An unset Verified means never checked. The verifier records the outcome
// of a failed check as the binding's evidence entry ("verify:<what>@<unix>",
// contract §1), which is the reason shown.
func describeVerified(binding *system.ParamBinding, evidence string) string {
	if binding == nil || binding.Verified == nil {
		return ""
	}
	if *binding.Verified {
		if binding.VerifiedAt == 0 {
			return " [verified]"
		}
		return " [verified " + time.Unix(binding.VerifiedAt, 0).UTC().Format(time.RFC3339) + "]"
	}
	if evidence == "" {
		evidence = "reason not recorded"
	}
	return " [unverified: " + evidence + "]"
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

type jsonReport struct {
	ModelID   string               `json:"modelId"`
	ModelName string               `json:"modelName"`
	ApiType   string               `json:"apiType"`
	Api       *system.ModelApiSpec `json:"api"`
	Trace     []string             `json:"trace"`
}

func main() {
	configPath := flag.String("config", "models-config.json", "path to models-config.json (V2 or legacy format)")
	timeout := flag.Duration("timeout", 15*time.Second, "detection budget per model")
	asJSON := flag.Bool("json", false, "print machine-readable JSON instead of text")
	only := flag.String("model", "", "only models whose name or ID contains this substring")
	ignoreHost := flag.Bool("ignore-host", false, "skip hostname recognition; identify by endpoints and listing shape only")
	twoHop := flag.Bool("two-hop", true, "follow a LiteLLM proxy to its upstream deployment (GET /model/info) for family/reasoning evidence; set =false to disable")
	flag.Parse()

	entries, err := modelsfile.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot load %s: %v\n", *configPath, err)
		os.Exit(1)
	}
	entries = modelsfile.Filter(entries, *only)
	if len(entries) == 0 {
		fmt.Fprintf(os.Stderr, "no models found in %s (filter %q)\n", *configPath, *only)
		os.Exit(1)
	}

	logger, err := lib.NewLogger("error", false, false, false, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot create logger: %v\n", err)
		os.Exit(1)
	}
	opts := apidetect.DefaultOptions()
	opts.IgnoreHostVendors = *ignoreHost
	opts.TwoHop = *twoHop
	detector := apidetect.NewDetector(logger, opts)

	var reports []jsonReport
	for _, e := range entries {
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		api, trace := detector.DetectWithTrace(ctx, e.Cfg)
		cancel()

		if *asJSON {
			reports = append(reports, jsonReport{
				ModelID:   e.ID,
				ModelName: e.Cfg.ModelName,
				ApiType:   e.Cfg.ApiType,
				Api:       api,
				Trace:     trace,
			})
			continue
		}

		fmt.Printf("=== %s (modelID %s, apiType %s)\n", e.Cfg.ModelName, e.ID, e.Cfg.ApiType)
		fmt.Printf("  url: %s\n", apidetect.RedactURL(e.Cfg.ApiURL))
		fmt.Print(formatResult(api, trace))
		fmt.Println()
	}

	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(reports); err != nil {
			fmt.Fprintf(os.Stderr, "cannot encode report: %v\n", err)
			os.Exit(1)
		}
	}
}
