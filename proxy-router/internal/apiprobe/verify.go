package apiprobe

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

// The verifier and the capability probes of design contract §5: each method
// sends one probe request (the render check two) through the model's own
// adapter and reads a verdict from the status, the error body and the
// response metadata. Fold turns the verdicts into apispec.Runtime, which
// MergeRuntime publishes; the caller owns pacing, the once-per-TTL state and
// the hosted-vendor gate.

// Probe names: the verify:<what>@<at> evidence sources (contract §1) except
// ProbeReasoningDisable and ProbeRender, whose verdicts land on a binding
// and take the binding's own source.
const (
	ProbeReasoningDisable = "reasoning.disable"
	ProbeTools            = apispec.CapFieldTools
	ProbeJSONSchema       = "json_schema"
	ProbeImage            = "image"
	ProbeRender           = "render"
)

// toolChoiceAutoSource names the observation behind a refused tool probe:
// vLLM's "auto" tool choice error is a server setting
// (--enable-auto-tool-choice), not a model fact, so the evidence entry
// (verify:tool_choice_auto@<at>) says which it saw.
const toolChoiceAutoSource = "tool_choice_auto"

// thinkingDefaultOn is system.ThinkingSpec.Default's value for a template
// that thinks unless told otherwise.
const thinkingDefaultOn = "on"

// Bounds of the render request, the same as apidetect's endpoint probes:
// 2 s per request, 2 MiB of body.
const (
	RenderTimeout = 2 * time.Second
	MaxRenderBody = 2 << 20
)

type Outcome int

const (
	// OutcomeUnknown: no usable answer (transport error, 5xx, a 4xx naming
	// nothing this package recognises, an adapter without response metadata).
	OutcomeUnknown Outcome = iota
	// OutcomeAccepted: the backend served the request (2xx) and the response
	// supports the claim.
	OutcomeAccepted
	// OutcomeRefuted: 2xx, but the response contradicts the claim (reasoning
	// content with reasoning disabled; a rendered prompt the kwarg left
	// unchanged).
	OutcomeRefuted
	// OutcomeRefused: a documented refusal of the feature (vLLM's tool_choice
	// "auto" error) or a 4xx of a feature request that names no parameter.
	OutcomeRefused
	// OutcomeRejected: a 4xx naming request parameters (RejectedParams).
	OutcomeRejected
)

func (o Outcome) String() string {
	switch o {
	case OutcomeAccepted:
		return "accepted"
	case OutcomeRefuted:
		return "refuted"
	case OutcomeRefused:
		return "refused"
	case OutcomeRejected:
		return "rejected"
	}
	return "unknown"
}

// Verdict is one probe's result. Reason is a trace note and the slug of the
// evidence entry for a refuted binding; it never carries a URL, key or
// served id. Err is the transport error, for a caller that redacts it.
type Verdict struct {
	Probe          string
	Outcome        Outcome
	Status         int
	Reason         string
	RejectedParams []string
	// Requests is how many requests reached the wire (0 when pacing was
	// cancelled before the first).
	Requests int
	At       int64
	Err      error
	Meta     *aiengine.ResponseMeta
}

// Verifier sends the probes for one model. Adapter is the model's own
// engine (the OpenAI adapter fills Meta; another adapter yields acceptance
// verdicts only). ApiURL and ApiKey are the model's; the render URL is
// derived from ApiURL, so the key only ever reaches the model's host. Pace,
// when set, runs before every request and returns false to abandon the
// probe (the caller's context was cancelled).
type Verifier struct {
	Adapter   aiengine.AIEngineStream
	Stack     string
	ModelName string
	ApiURL    string
	ApiKey    string
	Prompt    string
	Pace      func(ctx context.Context) bool
	// Client serves the render requests; nil means NewRenderClient().
	Client *http.Client
	Now    func() time.Time
}

// NewRenderClient bounds a render request the way apidetect bounds its
// probes: 2 s, and no redirect to another host or from https to http, since
// the request carries the key.
func NewRenderClient() *http.Client {
	return &http.Client{Timeout: RenderTimeout, CheckRedirect: refuseCrossHostRedirect}
}

func refuseCrossHostRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= 10 {
		return errors.New("stopped after 10 redirects")
	}
	prev := via[len(via)-1].URL
	if !strings.EqualFold(req.URL.Host, prev.Host) || (prev.Scheme == "https" && req.URL.Scheme != "https") {
		return http.ErrUseLastResponse
	}
	return nil
}

func (v *Verifier) now() int64 {
	if v.Now != nil {
		return v.Now().Unix()
	}
	return time.Now().Unix()
}

func (v *Verifier) pace(ctx context.Context) bool {
	if ctx.Err() != nil {
		return false
	}
	return v.Pace == nil || v.Pace(ctx)
}

func (v *Verifier) prompt() string {
	if v.Prompt == "" {
		return DefaultPrompt
	}
	return v.Prompt
}

// ReasoningDisableBinding is the binding VerifyReasoningDisable can apply:
// reasoning.disable of kind body_param or template_kwarg with a fixed
// value. Nil otherwise.
func ReasoningDisableBinding(spec *system.ModelApiSpec) *system.ParamBinding {
	if spec == nil {
		return nil
	}
	b := spec.Bindings[system.IntentReasoningDisable]
	if b == nil || b.Value == nil || b.Param == "" {
		return nil
	}
	if b.Kind != system.BindingKindBodyParam && b.Kind != system.BindingKindTemplateKwarg {
		return nil
	}
	return b
}

// TemplateToggle is the chat-template kwarg RenderThinkingToggle can test:
// the template_kwarg reasoning.disable binding's key under
// chat_template_kwargs and its fixed value. ok is false for any other
// binding.
func TemplateToggle(spec *system.ModelApiSpec) (key string, value any, ok bool) {
	b := ReasoningDisableBinding(spec)
	if b == nil || b.Kind != system.BindingKindTemplateKwarg {
		return "", nil, false
	}
	key, ok = strings.CutPrefix(b.Param, "chat_template_kwargs.")
	if !ok || key == "" || strings.Contains(key, ".") {
		return "", nil, false
	}
	return key, b.Value, true
}

// templateEnableValue is the value the toggle key takes with reasoning on:
// the reasoning.enable binding's when it is a template_kwarg on the same
// key, the negation of a boolean disable value otherwise (the family table
// pairs them that way). nil when no second value is known (thinking_budget
// 0 has no enable counterpart) or the two would be equal.
func templateEnableValue(spec *system.ModelApiSpec, key string, disable any) any {
	var enable any
	if b := spec.Bindings[system.IntentReasoningEnable]; b != nil && b.Kind == system.BindingKindTemplateKwarg && b.Param == "chat_template_kwargs."+key {
		enable = b.Value
	} else if off, ok := disable.(bool); ok {
		enable = !off
	}
	if enable == nil || reflect.DeepEqual(enable, disable) {
		return nil
	}
	return enable
}

func thinksByDefault(spec *system.ModelApiSpec) bool {
	return spec != nil && spec.Thinking != nil && spec.Thinking.Default == thinkingDefaultOn
}

// send makes one adapter request and reports the status it saw: the
// response metadata's when the adapter fills it, the error callback's
// otherwise, and 200 for a plain success of an adapter that reports
// neither.
func (v *Verifier) send(ctx context.Context, req *gcs.OpenAICompletionRequestExtra) (status int, meta *aiengine.ResponseMeta, err error) {
	meta = &aiengine.ResponseMeta{}
	err = v.Adapter.Prompt(aiengine.WithResponseMeta(ctx, meta), req, func(ctx context.Context, chunk gcs.Chunk, aiErr *gcs.AiEngineErrorResponse) error {
		if aiErr != nil && aiErr.StatusCode != 0 {
			status = aiErr.StatusCode
		}
		return nil
	})
	if meta.StatusCode != 0 {
		status = meta.StatusCode
	} else if err == nil && status == 0 {
		status = http.StatusOK
	}
	return status, meta, err
}

func is2xx(status int) bool { return status >= 200 && status < 300 }
func is4xx(status int) bool { return status >= 400 && status < 500 }

// httpReason labels a status this package draws no verdict from (a 3xx the
// redirect policy stopped, a 5xx) so the trace names what the backend sent.
func httpReason(status int) string { return "HTTP " + strconv.Itoa(status) }

func (v *Verifier) unknown(probe, reason string, err error) Verdict {
	return Verdict{Probe: probe, Outcome: OutcomeUnknown, Reason: reason, At: v.now(), Err: err}
}

// VerifyReasoningDisable resends the probe with the reasoning.disable
// binding applied (contract §5 a): 2xx without reasoning content verifies
// the binding, 2xx with reasoning content refutes it, a 4xx naming the
// parameter rejects it.
func (v *Verifier) VerifyReasoningDisable(ctx context.Context, spec *system.ModelApiSpec) Verdict {
	b := ReasoningDisableBinding(spec)
	if b == nil {
		return v.unknown(ProbeReasoningDisable, "no reasoning.disable body binding with a fixed value", nil)
	}
	req := NewProbeRequest(v.prompt())
	if err := ApplyBinding(req, b); err != nil {
		return v.unknown(ProbeReasoningDisable, "binding cannot be applied: "+err.Error(), err)
	}
	if !v.pace(ctx) {
		return v.unknown(ProbeReasoningDisable, "cancelled", ctx.Err())
	}
	status, meta, err := v.send(ctx, req)
	out := Verdict{Probe: ProbeReasoningDisable, Status: status, Requests: 1, At: v.now(), Meta: meta, Err: err}
	switch {
	case err != nil:
		out.Reason = "no response"
	case meta.StatusCode == 0:
		out.Reason = "adapter reports no response metadata"
	case is2xx(status) && !meta.ReasoningSeen:
		out.Outcome = OutcomeAccepted
		out.Reason = "no reasoning content with the binding applied"
	case is2xx(status):
		out.Outcome = OutcomeRefuted
		out.Reason = "reasoning content still present"
	case is4xx(status):
		if params, ok := ParseRejectedParams(meta.ErrorBody); ok {
			out.Outcome = OutcomeRejected
			out.RejectedParams = params
			out.Reason = "backend rejected " + strings.Join(params, ", ")
		} else {
			out.Reason = "HTTP 4xx naming no parameter"
		}
	default:
		out.Reason = httpReason(status)
	}
	return out
}

// ProbeTools sends the probe with a dummy tool and tool_choice "auto"
// (contract §5 b): 2xx means tools are accepted, vLLM's documented refusal
// means they are not, any other 4xx says nothing.
func (v *Verifier) ProbeTools(ctx context.Context) Verdict {
	return v.acceptance(ctx, ProbeTools, WithDummyToolAuto(NewProbeRequest(v.prompt())), func(meta *aiengine.ResponseMeta) (Outcome, string) {
		if IsToolChoiceAutoUnsupported(meta.ErrorBody) {
			return OutcomeRefused, "tool_choice auto refused by the server"
		}
		return OutcomeUnknown, "HTTP 4xx without the documented tool-choice refusal"
	})
}

// ProbeJSONSchema sends the probe with response_format json_schema: 2xx
// means the backend accepts the format. A 4xx is a refusal that publishes
// nothing: structured is a list, and a probe proves only what it saw.
func (v *Verifier) ProbeJSONSchema(ctx context.Context) Verdict {
	return v.acceptance(ctx, ProbeJSONSchema, WithJSONSchema(NewProbeRequest(v.prompt())), refusal)
}

// ProbeImage sends the probe with one 1x1 PNG image part: 2xx means the
// backend accepts image input. A 4xx publishes nothing, as for
// ProbeJSONSchema.
func (v *Verifier) ProbeImage(ctx context.Context) Verdict {
	return v.acceptance(ctx, ProbeImage, WithImagePart(NewProbeRequest(v.prompt())), refusal)
}

func refusal(meta *aiengine.ResponseMeta) (Outcome, string) {
	if params, ok := ParseRejectedParams(meta.ErrorBody); ok {
		return OutcomeRejected, "backend rejected " + strings.Join(params, ", ")
	}
	return OutcomeRefused, "HTTP 4xx"
}

func (v *Verifier) acceptance(ctx context.Context, probe string, req *gcs.OpenAICompletionRequestExtra, on4xx func(*aiengine.ResponseMeta) (Outcome, string)) Verdict {
	if !v.pace(ctx) {
		return v.unknown(probe, "cancelled", ctx.Err())
	}
	status, meta, err := v.send(ctx, req)
	out := Verdict{Probe: probe, Status: status, Requests: 1, At: v.now(), Meta: meta, Err: err}
	switch {
	case err != nil:
		out.Reason = "no response"
	case is2xx(status):
		out.Outcome = OutcomeAccepted
		out.Reason = "accepted"
	case is4xx(status):
		out.Outcome, out.Reason = on4xx(meta)
		if out.Outcome == OutcomeRejected {
			out.RejectedParams, _ = ParseRejectedParams(meta.ErrorBody)
		}
	default:
		out.Reason = httpReason(status)
	}
	return out
}

// RenderThinkingToggle asks vLLM to render the chat template (POST
// /v1/chat/completions/render, no generation) with the template
// reasoning.disable kwarg at its off value and at its on value. Different
// token_ids prove the template honours the kwarg: the reasoning.disable and
// reasoning.enable template bindings are verified. Identical token_ids
// refute them: a template that reads the kwarg renders thinking on and off
// differently. The two values are compared with each other, not against a
// render without the kwarg, because a template whose default already is
// the off value (Gemma 4: enable_thinking off unless set) renders the plain
// and the disabled prompt identically although the kwarg is honoured. With
// only the off value known the plain render stands in for the on value, and
// identical token_ids then refute the binding only when the template is
// known to think by default (Thinking.Default "on"); otherwise they prove
// nothing and no verdict is published. Only the vllm stack serves the
// endpoint.
func (v *Verifier) RenderThinkingToggle(ctx context.Context, spec *system.ModelApiSpec) Verdict {
	if v.Stack != "vllm" {
		return v.unknown(ProbeRender, "render endpoint is vLLM's", nil)
	}
	key, off, ok := TemplateToggle(spec)
	if !ok {
		return v.unknown(ProbeRender, "no template_kwarg reasoning.disable binding", nil)
	}
	urlStr := RenderURL(v.ApiURL)
	if urlStr == "" {
		return v.unknown(ProbeRender, "apiUrl names no host", nil)
	}
	model := v.ModelName
	disabled, err := RenderRequestBody(model, v.prompt(), map[string]any{key: off})
	if err != nil {
		return v.unknown(ProbeRender, "cannot build request", err)
	}
	var onKwargs map[string]any
	on := templateEnableValue(spec, key, off)
	if on != nil {
		onKwargs = map[string]any{key: on}
	}
	enabled, err := RenderRequestBody(model, v.prompt(), onKwargs)
	if err != nil {
		return v.unknown(ProbeRender, "cannot build request", err)
	}

	out := Verdict{Probe: ProbeRender, At: v.now()}
	var tokens [2][]int64
	for i, body := range [][]byte{disabled, enabled} {
		if !v.pace(ctx) {
			out.Reason = "cancelled"
			out.Err = ctx.Err()
			return out
		}
		status, resp, err := v.render(ctx, urlStr, body)
		out.Requests++
		out.Status = status
		switch {
		case err != nil:
			out.Reason = "no response"
			out.Err = err
			return out
		case is2xx(status):
			ids, ok := RenderedTokens(resp)
			if !ok {
				out.Reason = "render response carries no token_ids"
				return out
			}
			tokens[i] = ids
		case is4xx(status):
			if params, ok := ParseRejectedParams(resp); ok {
				out.Outcome = OutcomeRejected
				out.RejectedParams = params
				out.Reason = "backend rejected " + strings.Join(params, ", ")
				return out
			}
			out.Reason = "HTTP 4xx naming no parameter"
			return out
		default:
			out.Reason = httpReason(status)
			return out
		}
	}
	if equalTokens(tokens[0], tokens[1]) {
		if on == nil && !thinksByDefault(spec) {
			out.Reason = "rendered prompt unchanged; the template's default is not known to think"
			return out
		}
		out.Outcome = OutcomeRefuted
		out.Reason = "rendered prompt unchanged"
		return out
	}
	out.Outcome = OutcomeAccepted
	out.Reason = "rendered prompt differs with the kwarg"
	return out
}

func (v *Verifier) render(ctx context.Context, urlStr string, body []byte) (int, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlStr, bytes.NewReader(body))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if v.ApiKey != "" {
		req.Header.Set("Authorization", "Bearer "+v.ApiKey)
	}
	client := v.Client
	if client == nil {
		client = NewRenderClient()
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, MaxRenderBody))
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, raw, nil
}

func equalTokens(a, b []int64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Fold writes the verdicts into rt as MergeRuntime reads them: binding
// verdicts into Verified (a refuted reasoning.disable also marks reasoning
// seen with the disable applied, which upgrades thinking to always_on),
// rejections into RejectedParams, the tool verdict into ToolsSupported
// (verify:tools@<at> when accepted, verify:tool_choice_auto@<at> for the
// refusal), and the accepted capability probes into CapabilityFacts under
// verify:<probe>@<at>. Unknown and refused capability verdicts publish
// nothing. The completion with the binding applied is the direct
// observation of reasoning.disable; a render verdict only proves the
// template changed, so it never overrides a completion verdict on that
// intent, whatever the order of the verdicts.
func Fold(rt *apispec.Runtime, verdicts ...Verdict) {
	if rt == nil {
		return
	}
	disableDecided := false
	for _, v := range verdicts {
		if v.Probe == ProbeReasoningDisable && (v.Outcome == OutcomeAccepted || v.Outcome == OutcomeRefuted) {
			disableDecided = true
		}
	}
	verified := func(intent string, ok bool, v Verdict) {
		if rt.Verified == nil {
			rt.Verified = map[string]apispec.VerifiedBinding{}
		}
		rt.Verified[intent] = apispec.VerifiedBinding{Verified: ok, At: v.At, Reason: v.Reason}
	}
	for _, v := range verdicts {
		if v.Outcome == OutcomeRejected {
			for _, p := range v.RejectedParams {
				if !containsString(rt.RejectedParams, p) {
					rt.RejectedParams = append(rt.RejectedParams, p)
				}
			}
			continue
		}
		switch v.Probe {
		case ProbeReasoningDisable:
			switch v.Outcome {
			case OutcomeAccepted:
				verified(system.IntentReasoningDisable, true, v)
			case OutcomeRefuted:
				verified(system.IntentReasoningDisable, false, v)
				rt.ReasoningSeen = true
				rt.ReasoningDisabledInRequest = true
			}
		case ProbeRender:
			if v.Outcome != OutcomeAccepted && v.Outcome != OutcomeRefuted {
				continue
			}
			ok := v.Outcome == OutcomeAccepted
			if !disableDecided {
				verified(system.IntentReasoningDisable, ok, v)
			}
			verified(system.IntentReasoningEnable, ok, v)
		case ProbeTools:
			switch v.Outcome {
			case OutcomeAccepted:
				supported := true
				rt.ToolsSupported = &supported
				rt.ToolsSource = VerifySource(ProbeTools, v.At)
			case OutcomeRefused:
				supported := false
				rt.ToolsSupported = &supported
				rt.ToolsSource = VerifySource(toolChoiceAutoSource, v.At)
			}
		case ProbeJSONSchema:
			if v.Outcome == OutcomeAccepted {
				rt.CapabilityFacts.SetStructured([]string{"json_schema"}, VerifySource(ProbeJSONSchema, v.At))
			}
		case ProbeImage:
			if v.Outcome == OutcomeAccepted {
				rt.CapabilityFacts.SetInputModalities([]string{"image"}, VerifySource(ProbeImage, v.At))
			}
		}
	}
}

// VerifySource is the verify:<what>@<unix> evidence entry of contract §1.
func VerifySource(what string, at int64) string {
	return "verify:" + what + "@" + strconv.FormatInt(at, 10)
}

// ReasonSlug turns a verdict reason into the <what> of a verify:<what>@<at>
// entry: lower case, runs of anything but letters and digits collapsed to
// one underscore ("reasoning content still present" ->
// reasoning_content_still_present).
func ReasonSlug(reason string) string {
	var b strings.Builder
	pendingSep := false
	for _, r := range strings.ToLower(reason) {
		alnum := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		switch {
		case alnum:
			if pendingSep && b.Len() > 0 {
				b.WriteByte('_')
			}
			pendingSep = false
			b.WriteRune(r)
		default:
			pendingSep = true
		}
	}
	return b.String()
}

func containsString(list []string, s string) bool {
	for _, have := range list {
		if have == s {
			return true
		}
	}
	return false
}
