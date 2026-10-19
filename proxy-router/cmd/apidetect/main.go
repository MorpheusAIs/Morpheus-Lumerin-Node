// Command apidetect runs backend API detection for every model in a
// models-config.json and prints, per model, what was detected and the probe
// trace that led to it.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/cmd/internal/modelsfile"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apidetect"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apiprobe"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

// now is the clock of the verify verdicts; tests pin it.
var now = time.Now

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
	// The -verify fields: Runtime is the folded evidence MergeRuntime read,
	// VerifyTrace its trace, VerifyRequests every request the probes sent,
	// VerifyVerdicts what each probe concluded and VerifyNotSent how many
	// probes an account or rate-limit answer left unsent.
	Runtime        *apispec.Runtime `json:"runtime,omitempty"`
	VerifyTrace    []string         `json:"verifyTrace,omitempty"`
	VerifyRequests []probeRequest   `json:"verifyRequests,omitempty"`
	VerifyVerdicts []probeVerdict   `json:"verifyVerdicts,omitempty"`
	VerifyNotSent  int              `json:"verifyNotSent,omitempty"`
}

// probeRequest is one request a verify probe sent: the method and path
// only, because the host may carry credentials in the query or userinfo
// and the key travels in a header. Status is 0 when no response arrived.
type probeRequest struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Status int    `json:"status"`
	Failed bool   `json:"failed,omitempty"`
}

func (r probeRequest) String() string {
	if r.Failed {
		return r.Method + " " + r.Path + " -> no response"
	}
	return fmt.Sprintf("%s %s -> HTTP %d", r.Method, r.Path, r.Status)
}

// probeVerdict is the printable subset of apiprobe.Verdict: the transport
// error stays out because its text may quote the URL.
type probeVerdict struct {
	Probe          string   `json:"probe"`
	Outcome        string   `json:"outcome"`
	Status         int      `json:"status"`
	Reason         string   `json:"reason"`
	RejectedParams []string `json:"rejectedParams,omitempty"`
}

func (v probeVerdict) String() string {
	return fmt.Sprintf("%s: %s (HTTP %d): %s", v.Probe, v.Outcome, v.Status, v.Reason)
}

// recordingTransport notes every request that leaves through the client
// the adapter and the render check share, so the verify section reports
// what was sent from the wire rather than from the probes' bookkeeping.
type recordingTransport struct {
	next http.RoundTripper

	mu       sync.Mutex
	requests []probeRequest
}

func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.next.RoundTrip(req)
	entry := probeRequest{Method: req.Method, Path: req.URL.Path}
	if err != nil {
		entry.Failed = true
	} else {
		entry.Status = resp.StatusCode
	}
	t.mu.Lock()
	t.requests = append(t.requests, entry)
	t.mu.Unlock()
	return resp, err
}

func (t *recordingTransport) sent() []probeRequest {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]probeRequest(nil), t.requests...)
}

// verifyResult is what one -verify run produced for the model. NotSent is
// how many probes an account or rate-limit answer left unsent.
type verifyResult struct {
	Api      *system.ModelApiSpec
	Runtime  apispec.Runtime
	Trace    []string
	Requests []probeRequest
	Verdicts []probeVerdict
	NotSent  int
}

// verify sends every probe the C2 verifier offers for this model through
// the model's own adapter (the checker's path: aiengine.ApiAdapterFactory on
// the config entry), folds the verdicts and merges them onto a clone of the
// detected spec. Unlike the checker it does not consult the env flags, the
// once-per-TTL state or whether a field is already known:
// the operator asked for the round trips, and a verdict on a known field
// is the confirmation they came for. The render check runs only while the
// completion probe left reasoning.disable undecided (the checker's rule: a
// render verdict is weaker than the completion's on the same intent). Like
// the checker it stops at an account or rate-limit answer, which the probes
// left would meet too; NotSent counts them.
func verify(ctx context.Context, cfg config.ModelConfig, api *system.ModelApiSpec, timeout time.Duration, log lib.ILogger) (verifyResult, error) {
	rec := &recordingTransport{next: http.DefaultTransport}
	// The adapter bounds each request by its context; the key-carrying
	// completion probes get the render client's redirect policy.
	client := &http.Client{Transport: rec, CheckRedirect: apiprobe.RefuseCrossHostRedirect}
	adapter, ok := aiengine.ApiAdapterFactory(cfg.ApiType, cfg.ModelName, cfg.ApiURL, cfg.ApiKey, cfg.Parameters, timeout, log, client)
	if !ok {
		return verifyResult{}, fmt.Errorf("no adapter for apiType %q", cfg.ApiType)
	}
	renderClient := apiprobe.NewRenderClient()
	renderClient.Transport = rec

	spec := &system.ModelApiSpec{}
	if api != nil {
		spec = api.Clone()
	}
	v := &apiprobe.Verifier{
		Adapter: adapter, Stack: spec.Stack, ModelName: cfg.ModelName, ApiURL: cfg.ApiURL, ApiKey: cfg.ApiKey,
		Prompt: apiprobe.DefaultPrompt,
		Client: renderClient,
		Now:    now,
	}

	var verdicts []apiprobe.Verdict
	limited, notSent := false, 0
	run := func(probe func(context.Context) apiprobe.Verdict) bool {
		if limited {
			notSent++
			return false
		}
		probeCtx, cancel := context.WithTimeout(ctx, timeout)
		verdict := probe(probeCtx)
		cancel()
		verdicts = append(verdicts, verdict)
		limited = verdict.AccountLimited()
		return true
	}
	disableDecided := false
	if apiprobe.ReasoningDisableBinding(spec) != nil && run(func(ctx context.Context) apiprobe.Verdict { return v.VerifyReasoningDisable(ctx, spec) }) {
		disableDecided = verdicts[len(verdicts)-1].Outcome != apiprobe.OutcomeUnknown
	}
	run(v.ProbeTools)
	run(v.ProbeJSONSchema)
	run(v.ProbeImage)
	if _, _, ok := apiprobe.TemplateToggle(spec); ok && spec.Stack == "vllm" && !disableDecided {
		run(func(ctx context.Context) apiprobe.Verdict { return v.RenderThinkingToggle(ctx, spec) })
	}

	out := verifyResult{Requests: rec.sent(), NotSent: notSent}
	for _, verdict := range verdicts {
		out.Verdicts = append(out.Verdicts, probeVerdict{
			Probe: verdict.Probe, Outcome: verdict.Outcome.String(), Status: verdict.Status,
			Reason: verdict.Reason, RejectedParams: verdict.RejectedParams,
		})
	}
	apiprobe.Fold(&out.Runtime, verdicts...)
	out.Runtime.At = now().Unix()
	out.Trace = apispec.MergeRuntime(spec, out.Runtime)
	nameRefutedReasons(spec, out.Runtime)
	switch {
	case api != nil:
	case specEmpty(spec):
		// No detection and no probe proved anything: the report says so
		// rather than carrying an api block with no source (the checker
		// publishes nil here, modelhealth.Checker.publishedApi).
		spec = nil
	default:
		// Probes proved facts about a backend detection could not place;
		// what they wrote is verify:* evidence, and the block needs one of
		// the wire contract's source values.
		spec.Source = system.ApiSpecSourceDetected
	}
	out.Api = spec
	return out, nil
}

// nameRefutedReasons is the checker's convention for a refuted binding
// (modelhealth.Checker.publishedApi): MergeRuntime writes the intent's own
// name as the binding's evidence whether the verdict verified or refuted it,
// so the entry is rewritten to name what the probe saw
// (verify:reasoning_content_still_present@<at>) and the [unverified: ...]
// line shows the reason. -verify prints the same api block for a verdict
// as the checker publishes.
func nameRefutedReasons(spec *system.ModelApiSpec, rt apispec.Runtime) {
	for intent, verdict := range rt.Verified {
		if verdict.Verified || verdict.Reason == "" || spec.Bindings[intent] == nil {
			continue
		}
		slug := apiprobe.ReasonSlug(verdict.Reason)
		if slug == "" {
			continue
		}
		if spec.Evidence == nil {
			spec.Evidence = map[string]string{}
		}
		spec.Evidence[system.EvidenceKeyBinding(intent)] = apiprobe.VerifySource(slug, verdict.At)
	}
}

// specEmpty reports a spec MergeRuntime left with no fact in any field,
// which is what an all-unknown verify run leaves of &ModelApiSpec{}.
func specEmpty(spec *system.ModelApiSpec) bool {
	return spec.Stack == "" && spec.Via == "" && spec.ModelFamily == "" &&
		spec.Thinking == nil && spec.Capabilities == nil &&
		len(spec.Bindings) == 0 && len(spec.Parameters) == 0 && len(spec.Evidence) == 0
}

func formatVerify(res verifyResult) string {
	var b strings.Builder
	b.WriteString("  verify:\n")
	b.WriteString("    requests:\n")
	if len(res.Requests) == 0 {
		b.WriteString("      (none sent)\n")
	}
	for _, r := range res.Requests {
		fmt.Fprintf(&b, "      - %s\n", r)
	}
	b.WriteString("    verdicts:\n")
	if len(res.Verdicts) == 0 {
		b.WriteString("      (no probe applies)\n")
	}
	for _, v := range res.Verdicts {
		fmt.Fprintf(&b, "      - %s\n", v)
	}
	if res.NotSent > 0 {
		noun := "probes"
		if res.NotSent == 1 {
			noun = "probe"
		}
		fmt.Fprintf(&b, "      (%d more %s not sent: the backend answered with an account or rate-limit status)\n", res.NotSent, noun)
	}
	b.WriteString("    merge:\n")
	if len(res.Trace) == 0 {
		b.WriteString("      (nothing changed)\n")
	}
	for _, line := range res.Trace {
		fmt.Fprintf(&b, "      - %s\n", line)
	}
	return b.String()
}

// probeCount is the most requests a verify run sends: one completion per
// probe and two render calls.
const probeCount = 6

// hostedWarning is the cost line printed before probing a backend that
// bills per request, or one detection could not place (a hosted vendor
// behind an unrecognised host is the likely case).
func hostedWarning(api *system.ModelApiSpec) string {
	switch {
	case api == nil || api.Stack == "":
		return fmt.Sprintf("warning: stack not identified; -verify sends up to %d inference requests, which a hosted vendor bills", probeCount)
	case api.Via != "":
		return fmt.Sprintf("warning: %s is reached through %s; -verify sends up to %d inference requests, which the gateway bills", api.Stack, api.Via, probeCount)
	case apispec.IsHostedOrGatewayStack(api.Stack):
		return fmt.Sprintf("warning: %s is a hosted vendor or gateway; -verify sends up to %d inference requests, which it bills", api.Stack, probeCount)
	}
	return ""
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// Exit codes: 0 for -h (flag.ExitOnError's behaviour, which scripts may
// rely on), 1 for an unusable config or report, 2 for a flag the command
// cannot parse or a combination it cannot act on.
func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("apidetect", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "models-config.json", "path to models-config.json (V2 or legacy format)")
	timeout := fs.Duration("timeout", 15*time.Second, "detection budget per model; with -verify also the budget of each probe")
	asJSON := fs.Bool("json", false, "print machine-readable JSON instead of text")
	only := fs.String("model", "", "only models whose name or ID contains this substring")
	ignoreHost := fs.Bool("ignore-host", false, "skip hostname recognition; identify by endpoints and listing shape only")
	twoHop := fs.Bool("two-hop", true, "follow a LiteLLM proxy to its upstream deployment (GET /model/info) for family/reasoning evidence; set =false to disable")
	doVerify := fs.Bool("verify", false, "after detection send the binding verifier and capability probes (inference requests) to the one model -model selects, ignoring MODEL_API_VERIFY_BINDINGS / MODEL_API_PROBE_CAPABILITIES")
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}

	if *doVerify && *only == "" {
		fmt.Fprintln(stderr, "-verify needs -model naming exactly one entry: it sends inference requests to that model")
		return 2
	}

	entries, err := modelsfile.Load(*configPath)
	if err != nil {
		fmt.Fprintf(stderr, "cannot load %s: %v\n", *configPath, err)
		return 1
	}
	entries = modelsfile.Filter(entries, *only)
	if len(entries) == 0 {
		fmt.Fprintf(stderr, "no models found in %s (filter %q)\n", *configPath, *only)
		return 1
	}
	if *doVerify && len(entries) != 1 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Cfg.ModelName+" ("+e.ID+")")
		}
		fmt.Fprintf(stderr, "-verify needs -model naming exactly one entry; %q matches %d: %s\n", *only, len(entries), strings.Join(names, ", "))
		return 2
	}

	logger, err := lib.NewLogger("error", false, false, false, "")
	if err != nil {
		fmt.Fprintf(stderr, "cannot create logger: %v\n", err)
		return 1
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

		report := jsonReport{
			ModelID:   e.ID,
			ModelName: e.Cfg.ModelName,
			ApiType:   e.Cfg.ApiType,
			Api:       api,
			Trace:     trace,
		}
		var verified *verifyResult
		if *doVerify {
			if warning := hostedWarning(api); warning != "" {
				fmt.Fprintln(stderr, warning)
			}
			res, err := verify(context.Background(), e.Cfg, api, *timeout, logger)
			if err != nil {
				fmt.Fprintf(stderr, "cannot verify %s: %v\n", e.Cfg.ModelName, err)
				return 1
			}
			verified = &res
			report.Api = res.Api
			report.Runtime = &res.Runtime
			report.VerifyTrace = res.Trace
			report.VerifyRequests = res.Requests
			report.VerifyVerdicts = res.Verdicts
			report.VerifyNotSent = res.NotSent
		}

		if *asJSON {
			reports = append(reports, report)
			continue
		}

		fmt.Fprintf(stdout, "=== %s (modelID %s, apiType %s)\n", e.Cfg.ModelName, e.ID, e.Cfg.ApiType)
		fmt.Fprintf(stdout, "  url: %s\n", apidetect.RedactURL(e.Cfg.ApiURL))
		fmt.Fprint(stdout, formatResult(report.Api, trace))
		if verified != nil {
			fmt.Fprint(stdout, formatVerify(*verified))
		}
		fmt.Fprintln(stdout)
	}

	if *asJSON {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(reports); err != nil {
			fmt.Fprintf(stderr, "cannot encode report: %v\n", err)
			return 1
		}
	}
	return 0
}
