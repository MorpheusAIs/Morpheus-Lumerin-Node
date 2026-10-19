package apiprobe

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

// scriptedBackend answers /v1/chat/completions and its /render sibling from
// a per-test function and records every request body in order.
type scriptedBackend struct {
	mu       sync.Mutex
	requests []recorded
	respond  func(path string, body map[string]any) (int, string)
	server   *httptest.Server
}

type recorded struct {
	path string
	auth string
	body map[string]any
}

func newScriptedBackend(t *testing.T, respond func(path string, body map[string]any) (int, string)) *scriptedBackend {
	t.Helper()
	b := &scriptedBackend{respond: respond}
	b.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		b.mu.Lock()
		b.requests = append(b.requests, recorded{path: r.URL.Path, auth: r.Header.Get("Authorization"), body: body})
		respond := b.respond
		b.mu.Unlock()
		status, resp := respond(r.URL.Path, body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(resp))
	}))
	t.Cleanup(b.server.Close)
	return b
}

func (b *scriptedBackend) recorded() []recorded {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]recorded(nil), b.requests...)
}

func (b *scriptedBackend) verifier(stack string) *Verifier {
	chatURL := b.server.URL + "/v1/chat/completions"
	return &Verifier{
		Adapter:   aiengine.NewOpenAIEngine("qwen3-8b", chatURL, "sk-test", time.Minute, &lib.LoggerMock{}, nil),
		Stack:     stack,
		ModelName: "qwen3-8b",
		ApiURL:    chatURL,
		ApiKey:    "sk-test",
		Now:       func() time.Time { return time.Unix(1_700_000_000, 0) },
	}
}

const (
	okAnswer        = `{"id":"1","object":"chat.completion","model":"qwen3-8b","choices":[{"index":0,"message":{"role":"assistant","content":"5"},"finish_reason":"stop"}]}`
	reasoningAnswer = `{"id":"1","object":"chat.completion","model":"qwen3-8b","choices":[{"index":0,"message":{"role":"assistant","content":"5","reasoning_content":"2 and 3 make 5"},"finish_reason":"stop"}]}`
	kwargRejected   = `{"error":{"message":"[{'type': 'extra_forbidden', 'loc': ('body', 'chat_template_kwargs'), 'msg': 'Extra inputs are not permitted', 'input': {'enable_thinking': False}}]","type":"BadRequestError","param":null,"code":400}}`
	toolChoiceError = `{"error":{"message":"\"auto\" tool choice requires --enable-auto-tool-choice and --tool-call-parser to be set","type":"BadRequestError","param":null,"code":400}}`
	plainError      = `{"error":{"message":"image input is not supported by this model","type":"BadRequestError","code":400}}`
)

func always(status int, body string) func(string, map[string]any) (int, string) {
	return func(string, map[string]any) (int, string) { return status, body }
}

func qwen3Spec() *system.ModelApiSpec {
	return &system.ModelApiSpec{Stack: "vllm", ModelFamily: "qwen3", Thinking: &system.ThinkingSpec{Mode: system.ThinkingModeControllable},
		Bindings: map[string]*system.ParamBinding{
			system.IntentReasoningDisable: {Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.enable_thinking", ParamType: "boolean", Value: false},
			system.IntentReasoningEnable:  {Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.enable_thinking", ParamType: "boolean", Value: true},
		}}
}

func TestVerifyReasoningDisable(t *testing.T) {
	t.Run("2xx without reasoning verifies the binding", func(t *testing.T) {
		b := newScriptedBackend(t, always(200, okAnswer))
		v := b.verifier("vllm").VerifyReasoningDisable(context.Background(), qwen3Spec())
		require.Equal(t, OutcomeAccepted, v.Outcome, v.Reason)
		require.Equal(t, 200, v.Status)
		require.Equal(t, 1, v.Requests)
		require.Equal(t, int64(1_700_000_000), v.At)

		reqs := b.recorded()
		require.Len(t, reqs, 1)
		require.Equal(t, "/v1/chat/completions", reqs[0].path)
		require.Equal(t, "Bearer sk-test", reqs[0].auth)
		require.Equal(t, map[string]any{"enable_thinking": false}, reqs[0].body["chat_template_kwargs"])
		require.Equal(t, float64(1), reqs[0].body["max_tokens"])
		require.Contains(t, reqs[0].body["messages"].([]any)[0].(map[string]any)["content"], "What is")
	})

	t.Run("2xx with reasoning content refutes it", func(t *testing.T) {
		b := newScriptedBackend(t, always(200, reasoningAnswer))
		v := b.verifier("vllm").VerifyReasoningDisable(context.Background(), qwen3Spec())
		require.Equal(t, OutcomeRefuted, v.Outcome)
		require.Equal(t, "reasoning content still present", v.Reason)
		require.Equal(t, "reasoning_content_still_present", ReasonSlug(v.Reason))
	})

	t.Run("4xx naming the parameter rejects it", func(t *testing.T) {
		b := newScriptedBackend(t, always(400, kwargRejected))
		v := b.verifier("vllm").VerifyReasoningDisable(context.Background(), qwen3Spec())
		require.Equal(t, OutcomeRejected, v.Outcome)
		require.Equal(t, []string{"chat_template_kwargs"}, v.RejectedParams)
		require.Equal(t, 400, v.Status)
	})

	t.Run("4xx naming nothing is unknown", func(t *testing.T) {
		b := newScriptedBackend(t, always(400, plainError))
		v := b.verifier("vllm").VerifyReasoningDisable(context.Background(), qwen3Spec())
		require.Equal(t, OutcomeUnknown, v.Outcome)
		require.Equal(t, 400, v.Status)
	})

	t.Run("5xx is unknown", func(t *testing.T) {
		b := newScriptedBackend(t, always(503, `{"error":"down"}`))
		v := b.verifier("vllm").VerifyReasoningDisable(context.Background(), qwen3Spec())
		require.Equal(t, OutcomeUnknown, v.Outcome)
		require.Equal(t, 503, v.Status)
	})

	t.Run("no body binding sends nothing", func(t *testing.T) {
		b := newScriptedBackend(t, always(200, okAnswer))
		spec := qwen3Spec()
		spec.Bindings[system.IntentReasoningDisable] = &system.ParamBinding{Kind: system.BindingKindSystemPrompt, Hint: "/no_think"}
		v := b.verifier("vllm").VerifyReasoningDisable(context.Background(), spec)
		require.Equal(t, OutcomeUnknown, v.Outcome)
		require.Zero(t, v.Requests)
		require.Empty(t, b.recorded())

		require.Nil(t, ReasoningDisableBinding(nil))
		require.Nil(t, ReasoningDisableBinding(&system.ModelApiSpec{}))
		spec.Bindings[system.IntentReasoningDisable] = &system.ParamBinding{Kind: system.BindingKindBodyParam, Param: "think"}
		require.Nil(t, ReasoningDisableBinding(spec), "no fixed value")
	})

	t.Run("a body_param binding is applied at its path", func(t *testing.T) {
		b := newScriptedBackend(t, always(200, okAnswer))
		spec := &system.ModelApiSpec{Stack: "openai", Bindings: map[string]*system.ParamBinding{
			system.IntentReasoningDisable: {Kind: system.BindingKindBodyParam, Param: "reasoning_effort", Value: "none"},
		}}
		v := b.verifier("openai").VerifyReasoningDisable(context.Background(), spec)
		require.Equal(t, OutcomeAccepted, v.Outcome)
		require.Equal(t, "none", b.recorded()[0].body["reasoning_effort"])
	})

	t.Run("transport failure is unknown with the error kept aside", func(t *testing.T) {
		b := newScriptedBackend(t, always(200, okAnswer))
		b.server.Close()
		v := b.verifier("vllm").VerifyReasoningDisable(context.Background(), qwen3Spec())
		require.Equal(t, OutcomeUnknown, v.Outcome)
		require.Equal(t, "no response", v.Reason, "the reason never carries the error text (it names the URL)")
		require.Error(t, v.Err)
	})
}

func TestProbeTools(t *testing.T) {
	t.Run("2xx accepts", func(t *testing.T) {
		b := newScriptedBackend(t, always(200, okAnswer))
		v := b.verifier("vllm").ProbeTools(context.Background())
		require.Equal(t, OutcomeAccepted, v.Outcome)
		req := b.recorded()[0].body
		require.Equal(t, "auto", req["tool_choice"])
		require.Equal(t, DummyToolName, req["tools"].([]any)[0].(map[string]any)["function"].(map[string]any)["name"])
		require.Equal(t, float64(1), req["max_tokens"])
	})
	t.Run("the vLLM refusal refuses", func(t *testing.T) {
		b := newScriptedBackend(t, always(400, toolChoiceError))
		v := b.verifier("vllm").ProbeTools(context.Background())
		require.Equal(t, OutcomeRefused, v.Outcome)
	})
	t.Run("another 4xx is unknown", func(t *testing.T) {
		b := newScriptedBackend(t, always(400, plainError))
		v := b.verifier("vllm").ProbeTools(context.Background())
		require.Equal(t, OutcomeUnknown, v.Outcome)
	})
}

func TestProbeJSONSchemaAndImage(t *testing.T) {
	t.Run("json_schema 200", func(t *testing.T) {
		b := newScriptedBackend(t, always(200, okAnswer))
		v := b.verifier("vllm").ProbeJSONSchema(context.Background())
		require.Equal(t, OutcomeAccepted, v.Outcome)
		require.Equal(t, "json_schema", b.recorded()[0].body["response_format"].(map[string]any)["type"])
	})
	t.Run("json_schema 400", func(t *testing.T) {
		b := newScriptedBackend(t, always(400, plainError))
		v := b.verifier("vllm").ProbeJSONSchema(context.Background())
		require.Equal(t, OutcomeRefused, v.Outcome)
	})
	t.Run("json_schema 400 naming the parameter", func(t *testing.T) {
		b := newScriptedBackend(t, always(400, `{"error":{"message":"Unrecognized request argument supplied: response_format"}}`))
		v := b.verifier("openai").ProbeJSONSchema(context.Background())
		require.Equal(t, OutcomeRejected, v.Outcome)
		require.Equal(t, []string{"response_format"}, v.RejectedParams)
	})
	t.Run("image 200", func(t *testing.T) {
		b := newScriptedBackend(t, always(200, okAnswer))
		v := b.verifier("vllm").ProbeImage(context.Background())
		require.Equal(t, OutcomeAccepted, v.Outcome)
		parts := b.recorded()[0].body["messages"].([]any)[0].(map[string]any)["content"].([]any)
		require.Len(t, parts, 2)
		require.Equal(t, "image_url", parts[1].(map[string]any)["type"])
	})
	t.Run("image 400", func(t *testing.T) {
		b := newScriptedBackend(t, always(400, plainError))
		v := b.verifier("vllm").ProbeImage(context.Background())
		require.Equal(t, OutcomeRefused, v.Outcome)
	})
}

// An account or rate-limit status judges the caller, not the feature: every
// probe reads it as unknown before the body (here the one that probe would
// otherwise turn into a rejection or a refusal), and Fold publishes nothing
// from it.
func TestAccountStatusIsUnknown(t *testing.T) {
	const namesParam = `{"error":{"message":"Unrecognized request argument supplied: response_format"}}`
	probes := []struct {
		name string
		body string
		run  func(*Verifier) Verdict
	}{
		{ProbeReasoningDisable, kwargRejected, func(v *Verifier) Verdict { return v.VerifyReasoningDisable(context.Background(), qwen3Spec()) }},
		{ProbeTools, toolChoiceError, func(v *Verifier) Verdict { return v.ProbeTools(context.Background()) }},
		{ProbeJSONSchema, namesParam, func(v *Verifier) Verdict { return v.ProbeJSONSchema(context.Background()) }},
		{ProbeImage, namesParam, func(v *Verifier) Verdict { return v.ProbeImage(context.Background()) }},
		{ProbeRender, kwargRejected, func(v *Verifier) Verdict { return v.RenderThinkingToggle(context.Background(), qwen3Spec()) }},
	}
	for _, status := range []int{http.StatusUnauthorized, http.StatusPaymentRequired, http.StatusForbidden, http.StatusTooManyRequests} {
		for _, p := range probes {
			t.Run(p.name+" "+http.StatusText(status), func(t *testing.T) {
				b := newScriptedBackend(t, always(status, p.body))
				v := p.run(b.verifier("vllm"))
				require.Equal(t, OutcomeUnknown, v.Outcome)
				require.Equal(t, status, v.Status)
				require.Equal(t, "account or rate limit", v.Reason)
				require.Empty(t, v.RejectedParams)
				require.Equal(t, 1, v.Requests, "the render check stops at the first answer")
				require.True(t, v.AccountLimited())

				var rt apispec.Runtime
				Fold(&rt, v)
				require.Equal(t, apispec.Runtime{}, rt)
			})
		}
	}

	t.Run("other statuses keep their verdicts", func(t *testing.T) {
		for status, want := range map[int]Outcome{
			http.StatusBadRequest:         OutcomeRejected,
			http.StatusNotFound:           OutcomeRejected,
			http.StatusServiceUnavailable: OutcomeUnknown,
		} {
			b := newScriptedBackend(t, always(status, namesParam))
			v := b.verifier("vllm").ProbeJSONSchema(context.Background())
			require.Equal(t, want, v.Outcome, status)
			require.False(t, v.AccountLimited(), status)
		}
	})
}

// renderScript renders a template whose default is thinking on: the plain
// prompt and the enable_thinking:true prompt are the same, the
// enable_thinking:false prompt differs when differ is set.
func renderScript(differ bool, status int, errBody string) func(string, map[string]any) (int, string) {
	return renderByKwargs(status, errBody, func(kwargs map[string]any) string {
		if differ && kwargs != nil && kwargs["enable_thinking"] == false {
			return `{"request_id":"r2","token_ids":[1,2,3,9],"sampling_params":{}}`
		}
		return `{"request_id":"r1","token_ids":[1,2,3],"sampling_params":{}}`
	})
}

func renderByKwargs(status int, errBody string, render func(kwargs map[string]any) string) func(string, map[string]any) (int, string) {
	return func(path string, body map[string]any) (int, string) {
		if !strings.HasSuffix(path, "/render") {
			return 404, `{"error":"not the render endpoint"}`
		}
		if status != 200 {
			return status, errBody
		}
		kwargs, _ := body["chat_template_kwargs"].(map[string]any)
		return 200, render(kwargs)
	}
}

const (
	renderedOff = `{"request_id":"off","token_ids":[1,2,3,9],"sampling_params":{}}`
	renderedOn  = `{"request_id":"on","token_ids":[1,2,3],"sampling_params":{}}`
)

// gemma4Spec is a template whose default is thinking off: the plain render
// and the enable_thinking:false render are identical.
func gemma4Spec() *system.ModelApiSpec {
	spec := qwen3Spec()
	spec.ModelFamily = "gemma"
	spec.Thinking.Default = "off"
	return spec
}

// budgetSpec has a disable value with no enable counterpart
// (thinking_budget 0).
func budgetSpec(thinkingDefault string) *system.ModelApiSpec {
	return &system.ModelApiSpec{Stack: "vllm", Thinking: &system.ThinkingSpec{Mode: system.ThinkingModeControllable, Default: thinkingDefault},
		Bindings: map[string]*system.ParamBinding{
			system.IntentReasoningDisable: {Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.thinking_budget", ParamType: "number", Value: 0},
		}}
}

func TestRenderThinkingToggle(t *testing.T) {
	t.Run("differing prompts verify the template bindings", func(t *testing.T) {
		b := newScriptedBackend(t, renderScript(true, 200, ""))
		v := b.verifier("vllm").RenderThinkingToggle(context.Background(), qwen3Spec())
		require.Equal(t, OutcomeAccepted, v.Outcome, v.Reason)
		require.Equal(t, 2, v.Requests)

		reqs := b.recorded()
		require.Len(t, reqs, 2)
		for _, r := range reqs {
			require.Equal(t, "/v1/chat/completions/render", r.path)
			require.Equal(t, "Bearer sk-test", r.auth)
			require.Equal(t, "qwen3-8b", r.body["model"])
			require.Equal(t, float64(1), r.body["max_tokens"])
		}
		require.Equal(t, map[string]any{"enable_thinking": false}, reqs[0].body["chat_template_kwargs"], "the off value first")
		require.Equal(t, map[string]any{"enable_thinking": true}, reqs[1].body["chat_template_kwargs"], "then the on value")
	})

	t.Run("identical prompts refute them", func(t *testing.T) {
		b := newScriptedBackend(t, renderScript(false, 200, ""))
		v := b.verifier("vllm").RenderThinkingToggle(context.Background(), qwen3Spec())
		require.Equal(t, OutcomeRefuted, v.Outcome)
		require.Equal(t, "rendered_prompt_unchanged", ReasonSlug(v.Reason))
	})

	t.Run("a default-off template is verified although plain equals disabled", func(t *testing.T) {
		b := newScriptedBackend(t, renderByKwargs(200, "", func(kwargs map[string]any) string {
			if kwargs != nil && kwargs["enable_thinking"] == true {
				return renderedOn
			}
			return renderedOff
		}))
		v := b.verifier("vllm").RenderThinkingToggle(context.Background(), gemma4Spec())
		require.Equal(t, OutcomeAccepted, v.Outcome, v.Reason)
		reqs := b.recorded()
		require.Len(t, reqs, 2)
		require.NotNil(t, reqs[0].body["chat_template_kwargs"], "no plain render: it equals the disabled one here")
		require.NotNil(t, reqs[1].body["chat_template_kwargs"])
	})

	t.Run("the enable value is the reasoning.enable binding's, not a negation", func(t *testing.T) {
		spec := qwen3Spec()
		spec.Bindings[system.IntentReasoningDisable] = &system.ParamBinding{Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.thinking", ParamType: "string", Value: "off"}
		spec.Bindings[system.IntentReasoningEnable] = &system.ParamBinding{Kind: system.BindingKindTemplateKwarg, Param: "chat_template_kwargs.thinking", ParamType: "string", Value: "on"}
		b := newScriptedBackend(t, renderByKwargs(200, "", func(kwargs map[string]any) string {
			if kwargs["thinking"] == "on" {
				return renderedOn
			}
			return renderedOff
		}))
		v := b.verifier("vllm").RenderThinkingToggle(context.Background(), spec)
		require.Equal(t, OutcomeAccepted, v.Outcome, v.Reason)
		reqs := b.recorded()
		require.Equal(t, map[string]any{"thinking": "off"}, reqs[0].body["chat_template_kwargs"])
		require.Equal(t, map[string]any{"thinking": "on"}, reqs[1].body["chat_template_kwargs"])
	})

	t.Run("one value only: unchanged proves nothing unless the template thinks by default", func(t *testing.T) {
		unchanged := renderByKwargs(200, "", func(map[string]any) string { return renderedOn })
		b := newScriptedBackend(t, unchanged)
		v := b.verifier("vllm").RenderThinkingToggle(context.Background(), budgetSpec(""))
		require.Equal(t, OutcomeUnknown, v.Outcome, v.Reason)
		require.Equal(t, 2, v.Requests)
		reqs := b.recorded()
		require.Equal(t, map[string]any{"thinking_budget": float64(0)}, reqs[0].body["chat_template_kwargs"])
		require.Nil(t, reqs[1].body["chat_template_kwargs"], "the plain render stands in for the on value")

		b = newScriptedBackend(t, unchanged)
		v = b.verifier("vllm").RenderThinkingToggle(context.Background(), budgetSpec("off"))
		require.Equal(t, OutcomeUnknown, v.Outcome, "off by default: the disabled render is expected to match the plain one")

		b = newScriptedBackend(t, unchanged)
		v = b.verifier("vllm").RenderThinkingToggle(context.Background(), budgetSpec("on"))
		require.Equal(t, OutcomeRefuted, v.Outcome, "on by default: the budget should have changed the prompt")

		b = newScriptedBackend(t, renderByKwargs(200, "", func(kwargs map[string]any) string {
			if kwargs == nil {
				return renderedOn
			}
			return renderedOff
		}))
		v = b.verifier("vllm").RenderThinkingToggle(context.Background(), budgetSpec(""))
		require.Equal(t, OutcomeAccepted, v.Outcome, "a differing render verifies whatever the default")
	})

	t.Run("a 4xx naming the parameter rejects", func(t *testing.T) {
		b := newScriptedBackend(t, renderScript(false, 400, kwargRejected))
		v := b.verifier("vllm").RenderThinkingToggle(context.Background(), qwen3Spec())
		require.Equal(t, OutcomeRejected, v.Outcome)
		require.Equal(t, []string{"chat_template_kwargs"}, v.RejectedParams)
		require.Equal(t, 1, v.Requests, "the first request already failed")
	})

	t.Run("a missing endpoint is unknown", func(t *testing.T) {
		b := newScriptedBackend(t, renderScript(false, 404, `{"detail":"Not Found"}`))
		v := b.verifier("vllm").RenderThinkingToggle(context.Background(), qwen3Spec())
		require.Equal(t, OutcomeUnknown, v.Outcome)
		require.Equal(t, 404, v.Status)
	})

	t.Run("only vllm and only a template binding", func(t *testing.T) {
		b := newScriptedBackend(t, renderScript(true, 200, ""))
		v := b.verifier("sglang").RenderThinkingToggle(context.Background(), qwen3Spec())
		require.Equal(t, OutcomeUnknown, v.Outcome)
		require.Empty(t, b.recorded())

		spec := &system.ModelApiSpec{Stack: "vllm", Bindings: map[string]*system.ParamBinding{
			system.IntentReasoningDisable: {Kind: system.BindingKindBodyParam, Param: "reasoning_effort", Value: "none"},
		}}
		v = b.verifier("vllm").RenderThinkingToggle(context.Background(), spec)
		require.Equal(t, OutcomeUnknown, v.Outcome)
		require.Empty(t, b.recorded())
	})

	t.Run("a redirect to another host is not followed", func(t *testing.T) {
		other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("the key reached another host: %s", r.Header.Get("Authorization"))
		}))
		defer other.Close()
		b := newScriptedBackend(t, nil)
		b.respond = func(string, map[string]any) (int, string) { return 200, "" }
		redirecting := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, other.URL+"/v1/chat/completions/render", http.StatusTemporaryRedirect)
		}))
		defer redirecting.Close()
		v := b.verifier("vllm")
		v.ApiURL = redirecting.URL + "/v1/chat/completions"
		verdict := v.RenderThinkingToggle(context.Background(), qwen3Spec())
		require.Equal(t, OutcomeUnknown, verdict.Outcome)
		require.Equal(t, http.StatusTemporaryRedirect, verdict.Status)
		require.Equal(t, "HTTP 307", verdict.Reason, "the trace names the status the redirect policy stopped at")
	})

	t.Run("the body is bounded", func(t *testing.T) {
		huge := `{"token_ids":[` + strings.Repeat("1,", MaxRenderBody) + `1]}`
		b := newScriptedBackend(t, func(string, map[string]any) (int, string) { return 200, huge })
		v := b.verifier("vllm").RenderThinkingToggle(context.Background(), qwen3Spec())
		require.Equal(t, OutcomeUnknown, v.Outcome, "a truncated body parses to nothing")
	})
}

func TestPaceStopsAProbe(t *testing.T) {
	b := newScriptedBackend(t, always(200, okAnswer))
	v := b.verifier("vllm")
	paced := 0
	v.Pace = func(ctx context.Context) bool { paced++; return false }
	verdict := v.ProbeTools(context.Background())
	require.Equal(t, OutcomeUnknown, verdict.Outcome)
	require.Zero(t, verdict.Requests)
	require.Equal(t, 1, paced)
	require.Empty(t, b.recorded())

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	verdict = v.RenderThinkingToggle(ctx, qwen3Spec())
	require.Equal(t, "cancelled", verdict.Reason)
	require.Zero(t, verdict.Requests)
}

func TestFold(t *testing.T) {
	const at = int64(1_700_000_000)
	t.Run("verified binding, tools, structured, image", func(t *testing.T) {
		var rt apispec.Runtime
		Fold(&rt,
			Verdict{Probe: ProbeReasoningDisable, Outcome: OutcomeAccepted, Reason: "ok", At: at},
			Verdict{Probe: ProbeTools, Outcome: OutcomeAccepted, At: at},
			Verdict{Probe: ProbeJSONSchema, Outcome: OutcomeAccepted, At: at},
			Verdict{Probe: ProbeImage, Outcome: OutcomeAccepted, At: at},
		)
		require.Equal(t, apispec.VerifiedBinding{Verified: true, At: at, Reason: "ok"}, rt.Verified[system.IntentReasoningDisable])
		require.True(t, *rt.ToolsSupported)
		require.Equal(t, "verify:tools@1700000000", rt.ToolsSource)
		require.Equal(t, []string{"json_schema"}, rt.CapabilityFacts.Structured)
		require.Equal(t, "verify:json_schema@1700000000", rt.CapabilityFacts.Sources[apispec.CapFieldStructured])
		require.Equal(t, []string{"image"}, rt.CapabilityFacts.Modalities.Input)
		require.Equal(t, "verify:image@1700000000", rt.CapabilityFacts.Sources[apispec.CapFieldModalitiesInput])
		require.False(t, rt.ReasoningSeen)
	})

	t.Run("refuted binding marks reasoning seen with the disable applied", func(t *testing.T) {
		var rt apispec.Runtime
		Fold(&rt, Verdict{Probe: ProbeReasoningDisable, Outcome: OutcomeRefuted, Reason: "reasoning content still present", At: at})
		require.False(t, rt.Verified[system.IntentReasoningDisable].Verified)
		require.True(t, rt.ReasoningSeen)
		require.True(t, rt.ReasoningDisabledInRequest)
	})

	t.Run("rejections go to RejectedParams whatever the probe", func(t *testing.T) {
		var rt apispec.Runtime
		Fold(&rt,
			Verdict{Probe: ProbeReasoningDisable, Outcome: OutcomeRejected, RejectedParams: []string{"chat_template_kwargs"}, At: at},
			Verdict{Probe: ProbeJSONSchema, Outcome: OutcomeRejected, RejectedParams: []string{"response_format", "chat_template_kwargs"}, At: at},
		)
		require.Equal(t, []string{"chat_template_kwargs", "response_format"}, rt.RejectedParams)
		require.Empty(t, rt.Verified)
	})

	t.Run("refused tools publish false under the refusal's own source; refused capability probes publish nothing", func(t *testing.T) {
		var rt apispec.Runtime
		Fold(&rt,
			Verdict{Probe: ProbeTools, Outcome: OutcomeRefused, At: at},
			Verdict{Probe: ProbeJSONSchema, Outcome: OutcomeRefused, At: at},
			Verdict{Probe: ProbeImage, Outcome: OutcomeUnknown, At: at},
		)
		require.False(t, *rt.ToolsSupported)
		require.Equal(t, "verify:tool_choice_auto@1700000000", rt.ToolsSource)
		require.True(t, rt.CapabilityFacts.Empty())
	})

	t.Run("a render verdict never overrides the completion's reasoning.disable verdict", func(t *testing.T) {
		var rt apispec.Runtime
		Fold(&rt,
			Verdict{Probe: ProbeReasoningDisable, Outcome: OutcomeRefuted, Reason: "reasoning content still present", At: at},
			Verdict{Probe: ProbeRender, Outcome: OutcomeAccepted, Reason: "differs", At: at},
		)
		require.False(t, rt.Verified[system.IntentReasoningDisable].Verified)
		require.Equal(t, "reasoning content still present", rt.Verified[system.IntentReasoningDisable].Reason)
		require.True(t, rt.Verified[system.IntentReasoningEnable].Verified, "the render still speaks for the enable binding")
		require.True(t, rt.ReasoningDisabledInRequest)

		rt = apispec.Runtime{}
		Fold(&rt,
			Verdict{Probe: ProbeRender, Outcome: OutcomeRefuted, Reason: "rendered prompt unchanged", At: at},
			Verdict{Probe: ProbeReasoningDisable, Outcome: OutcomeAccepted, Reason: "ok", At: at},
		)
		require.True(t, rt.Verified[system.IntentReasoningDisable].Verified, "order does not matter")
		require.False(t, rt.Verified[system.IntentReasoningEnable].Verified)
	})

	t.Run("render verifies both template bindings", func(t *testing.T) {
		var rt apispec.Runtime
		Fold(&rt, Verdict{Probe: ProbeRender, Outcome: OutcomeAccepted, Reason: "differs", At: at})
		require.True(t, rt.Verified[system.IntentReasoningDisable].Verified)
		require.True(t, rt.Verified[system.IntentReasoningEnable].Verified)
		Fold(&rt, Verdict{Probe: ProbeRender, Outcome: OutcomeRefuted, Reason: "rendered prompt unchanged", At: at})
		require.False(t, rt.Verified[system.IntentReasoningDisable].Verified)
		require.False(t, rt.Verified[system.IntentReasoningEnable].Verified)
		require.False(t, rt.ReasoningSeen, "an unchanged template says nothing about reasoning content")
	})

	t.Run("nil runtime is a no-op", func(t *testing.T) {
		Fold(nil, Verdict{Probe: ProbeTools, Outcome: OutcomeAccepted})
	})
}

func TestReasonSlug(t *testing.T) {
	require.Equal(t, "reasoning_content_still_present", ReasonSlug("reasoning content still present"))
	require.Equal(t, "rendered_prompt_unchanged", ReasonSlug("Rendered prompt: unchanged!"))
	require.Equal(t, "", ReasonSlug("---"))
	require.Equal(t, "verify:reasoning_content_still_present@1700000000", VerifySource(ReasonSlug("reasoning content still present"), 1_700_000_000))
}

func TestOutcomeString(t *testing.T) {
	require.Equal(t, "unknown", OutcomeUnknown.String())
	require.Equal(t, "accepted", OutcomeAccepted.String())
	require.Equal(t, "refuted", OutcomeRefuted.String())
	require.Equal(t, "refused", OutcomeRefused.String())
	require.Equal(t, "rejected", OutcomeRejected.String())
}
