// Command apidetect runs backend API detection (internal/apidetect) for every
// model in a models-config.json and prints, per model, what was detected and
// the exact steps that led to it — every probe attempted, what each returned,
// and which evidence source decided the serving stack, model family and
// thinking knob. Diagnostics tool for providers; sends no inference requests.
//
// Usage:
//
//	go run ./cmd/apidetect [-config models-config.json] [-model <substring>] [-ignore-host] [-two-hop=false] [-timeout 15s] [-json]
//
// -model runs only models whose name or ID contains the substring;
// -ignore-host skips hostname recognition so hosted vendors must be
// identified by their endpoints and listing shape alone (what an unknown
// custom domain would get); -two-hop=false disables the LiteLLM second hop
// (GET /model/info and the anonymous read of its upstream, which — when it
// identifies the upstream — becomes the reported stack, printed with a
// "via: litellm" line).
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

// formatResult renders one model's detection outcome and trace as
// indented text.
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
		}
		fmt.Fprintf(&b, "    thinking:    %s\n", thinking)
		if len(api.Bindings) > 0 {
			b.WriteString("    bindings:\n")
			intents := make([]string, 0, len(api.Bindings))
			for intent := range api.Bindings {
				intents = append(intents, intent)
			}
			sort.Strings(intents)
			for _, intent := range intents {
				fmt.Fprintf(&b, "      %-18s %s\n", intent+":", apispec.DescribeBinding(api.Bindings[intent]))
			}
		}
		if len(api.Parameters) > 0 {
			fmt.Fprintf(&b, "    parameters:  %s\n", strings.Join(api.Parameters, ", "))
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

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// jsonReport is the machine-readable per-model output for -json mode.
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
