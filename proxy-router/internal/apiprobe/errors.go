// Package apiprobe turns a backend's inference-error body into capability
// facts: the served context window from a context-exceeded error and the
// request parameters a backend refused. Pure functions over bounded input;
// no request is made here.
package apiprobe

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
)

// MaxErrorBody bounds the body a parser looks at; every format below fits in
// its first kilobyte, and a longer body is a bulk echo of the request.
const MaxErrorBody = 64 << 10

// Evidence sources (the "health-probe:<what>" family of the design contract
// §1), one per recognised error format so a consumer can tell which backend
// produced the number.
const (
	// The OpenAI sentence, also emitted by vLLM's chat endpoint:
	// https://github.com/vllm-project/vllm/blob/v0.30.0/vllm/renderers/params.py
	ContextSourceOpenAI = "health-probe:context_length_exceeded"
	// vLLM's second form, "Input length (n) exceeds model's maximum context length (N)":
	// https://github.com/vllm-project/vllm/blob/v0.30.0/vllm/entrypoints/serve/utils/api_utils.py
	ContextSourceVLLM = "health-probe:vllm.context_length_exceeded"
	// llama.cpp server, type exceed_context_size_error with structured n_ctx:
	// https://github.com/ggml-org/llama.cpp/blob/master/tools/server/server.cpp
	ContextSourceLlamaCpp = "health-probe:llamacpp.exceed_context_size_error"
	// TGI ValidationError::MaxTotalTokens, the served window (max_total_tokens):
	// https://github.com/huggingface/text-generation-inference/blob/v3.3.7/router/src/validation.rs#L999
	ContextSourceTGI = "health-probe:tgi.validation_error"
	// TGI ValidationError::InputLength carries max_input_tokens, a lower bound
	// of the window (TGI requires max_total_tokens > max_input_tokens):
	// https://github.com/huggingface/text-generation-inference/blob/v3.3.7/router/src/validation.rs#L1001
	ContextSourceTGIInput = "health-probe:tgi.max_input_tokens"
	// Anthropic 400 invalid_request_error "prompt is too long":
	// https://docs.anthropic.com/en/api/errors
	ContextSourceAnthropic = "health-probe:anthropic.prompt_too_long"
)

var (
	openaiContextRe   = regexp.MustCompile(`(?i)maximum context length is (\d+) tokens`)
	vllmContextRe     = regexp.MustCompile(`(?i)exceeds (?:the )?model's maximum context length \((\d+)\)`)
	llamaCppContextRe = regexp.MustCompile(`(?i)exceeds the available context size \((\d+) tokens\)`)
	// Anchored on the parameter names: the neighbouring MaxNewTokens sentence
	// ("`max_new_tokens` must be <= N") and any "x must be <= N" range error
	// carry no window.
	tgiTotalRe       = regexp.MustCompile("`inputs` tokens \\+ `max_new_tokens` must be <= (\\d+)")
	tgiInputRe       = regexp.MustCompile("`inputs` must have less than (\\d+) tokens")
	anthropicTooLong = regexp.MustCompile(`(?i)prompt is too long`)
	anthropicLimitRe = regexp.MustCompile(`(?i)(\d+) tokens > (\d+) maximum`)

	// OpenAI: "Unrecognized request argument supplied: foo" (plural form with a
	// comma list): https://platform.openai.com/docs/guides/error-codes
	unrecognizedArgRe = regexp.MustCompile(`(?i)Unrecognized request arguments? supplied: ([A-Za-z0-9_.,\- ]+)`)
	// pydantic v2 extra_forbidden in vLLM's stringified form
	// "{'type': 'extra_forbidden', 'loc': ('body', 'foo'), ...}":
	// https://docs.pydantic.dev/latest/errors/validation_errors/#extra_forbidden
	extraForbiddenReprRe = regexp.MustCompile(`extra_forbidden'?,?\s*'loc':\s*\(\s*'body'((?:,\s*'[A-Za-z0-9_.\-]+')+)\s*\)`)
	reprItemRe           = regexp.MustCompile(`'([A-Za-z0-9_.\-]+)'`)
	// LiteLLM: "litellm.UnsupportedParamsError: <provider> does not support
	// parameters: ['a', 'b'], for model=..." or "... parameters: {'a': 1}":
	// https://github.com/BerriAI/litellm/blob/main/litellm/utils.py
	litellmUnsupportedRe = regexp.MustCompile(`(?i)does not support parameters:\s*([\[{][^\]}]*[\]}])`)
	quotedNameRe         = regexp.MustCompile(`['"]([A-Za-z0-9_.\-]+)['"]`)
	dictKeyRe            = regexp.MustCompile(`['"]([A-Za-z0-9_.\-]+)['"]\s*:`)
	// vLLM without --enable-auto-tool-choice:
	// https://docs.vllm.ai/en/latest/features/tool_calling.html
	toolChoiceAutoRe = regexp.MustCompile(`(?i)"auto" tool choice requires\s+--enable-auto-tool-choice`)

	paramNameRe = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.\-]{0,63}$`)
)

// ToolChoiceParam is the sentinel ParseRejectedParams returns for the vLLM
// auto-tool-choice refusal; the request had no such parameter binding to drop,
// the fact is "tools unsupported as configured".
const ToolChoiceParam = "tool_choice"

const maxRejectedParams = 32

// errorEnvelope is the union of the JSON shapes the parsers read. "error" is
// an object for OpenAI, vLLM, llama.cpp and Anthropic, a string for TGI.
type errorEnvelope struct {
	Error   json.RawMessage `json:"error"`
	Message string          `json:"message"`
	Detail  json.RawMessage `json:"detail"`
}

type errorObject struct {
	Message string          `json:"message"`
	Type    string          `json:"type"`
	NCtx    json.RawMessage `json:"n_ctx"`
}

type pydanticItem struct {
	Type string            `json:"type"`
	Loc  []json.RawMessage `json:"loc"`
}

type parsedError struct {
	texts  []string
	object *errorObject
	detail []pydanticItem
}

func bound(body []byte) []byte {
	if len(body) > MaxErrorBody {
		return body[:MaxErrorBody]
	}
	return body
}

// parse collects every human-readable message the body holds (the raw text
// when it is not JSON) plus the structured parts the callers inspect.
func parse(body []byte) parsedError {
	body = bound(body)
	var p parsedError
	var env errorEnvelope
	if err := json.Unmarshal(body, &env); err != nil {
		p.texts = []string{string(body)}
		return p
	}
	if env.Message != "" {
		p.texts = append(p.texts, env.Message)
	}
	if len(env.Error) > 0 {
		var obj errorObject
		var text string
		switch {
		case json.Unmarshal(env.Error, &obj) == nil && (obj.Message != "" || obj.Type != "" || len(obj.NCtx) > 0):
			p.object = &obj
			p.texts = append(p.texts, obj.Message)
		case json.Unmarshal(env.Error, &text) == nil:
			p.texts = append(p.texts, text)
		}
	}
	if len(env.Detail) > 0 {
		var items []pydanticItem
		var text string
		switch {
		case json.Unmarshal(env.Detail, &items) == nil:
			p.detail = items
		case json.Unmarshal(env.Detail, &text) == nil:
			p.texts = append(p.texts, text)
		}
	}
	p.texts = append(p.texts, string(body))
	return p
}

func firstNumber(re *regexp.Regexp, texts []string, group int) (int64, bool) {
	for _, t := range texts {
		m := re.FindStringSubmatch(t)
		if m == nil {
			continue
		}
		n, err := strconv.ParseInt(m[group], 10, 64)
		if err != nil || !config.ValidCapabilityInt(n) {
			continue
		}
		return n, true
	}
	return 0, false
}

func matchesAny(re *regexp.Regexp, texts []string) bool {
	for _, t := range texts {
		if re.MatchString(t) {
			return true
		}
	}
	return false
}

// ParseContextLimit reads the served context window out of a
// context-exceeded error body. status is the HTTP status the body came with
// (0 when unknown); only a 4xx is a request-side rejection. source is the
// evidence entry for the recognised format. A recognised format that names
// no usable number (Anthropic's "prompt is too long" without a total)
// returns ok=false with its source set, so the caller can tell "recognised,
// no value" from "not a context error".
func ParseContextLimit(body []byte, status int) (limit int64, source string, ok bool) {
	if len(body) == 0 || (status != 0 && (status < 400 || status > 499)) {
		return 0, "", false
	}
	p := parse(body)

	if obj := p.object; obj != nil && obj.Type == "exceed_context_size_error" {
		if len(obj.NCtx) > 0 {
			if n, err := strconv.ParseInt(string(obj.NCtx), 10, 64); err == nil && config.ValidCapabilityInt(n) {
				return n, ContextSourceLlamaCpp, true
			}
		}
		if n, found := firstNumber(llamaCppContextRe, p.texts, 1); found {
			return n, ContextSourceLlamaCpp, true
		}
		return 0, ContextSourceLlamaCpp, false
	}
	if n, found := firstNumber(llamaCppContextRe, p.texts, 1); found {
		return n, ContextSourceLlamaCpp, true
	}
	if n, found := firstNumber(vllmContextRe, p.texts, 1); found {
		return n, ContextSourceVLLM, true
	}
	if n, found := firstNumber(openaiContextRe, p.texts, 1); found {
		return n, ContextSourceOpenAI, true
	}
	if n, found := firstNumber(tgiTotalRe, p.texts, 1); found {
		return n, ContextSourceTGI, true
	}
	if n, found := firstNumber(tgiInputRe, p.texts, 1); found {
		return n, ContextSourceTGIInput, true
	}
	if matchesAny(anthropicTooLong, p.texts) {
		if n, found := firstNumber(anthropicLimitRe, p.texts, 2); found {
			return n, ContextSourceAnthropic, true
		}
		return 0, ContextSourceAnthropic, false
	}
	return 0, "", false
}

// ParseRejectedParams lists the request parameters an error body says the
// backend refused: OpenAI "Unrecognized request argument supplied", pydantic
// extra_forbidden entries (FastAPI detail array or vLLM's stringified repr),
// LiteLLM UnsupportedParamsError, and the vLLM auto-tool-choice refusal as
// the sentinel ToolChoiceParam. Names are validated and deduped; ok is false
// when nothing was recognised.
func ParseRejectedParams(body []byte) (params []string, ok bool) {
	if len(body) == 0 {
		return nil, false
	}
	p := parse(body)
	var out []string
	add := func(name string) {
		name = strings.TrimSpace(name)
		if !paramNameRe.MatchString(name) || len(out) >= maxRejectedParams {
			return
		}
		for _, have := range out {
			if have == name {
				return
			}
		}
		out = append(out, name)
	}

	for _, item := range p.detail {
		if item.Type != "extra_forbidden" {
			continue
		}
		if name := locParam(item.Loc); name != "" {
			add(name)
		}
	}
	for _, t := range p.texts {
		if m := unrecognizedArgRe.FindStringSubmatch(t); m != nil {
			for _, name := range strings.Split(m[1], ",") {
				add(name)
			}
		}
		for _, m := range extraForbiddenReprRe.FindAllStringSubmatch(t, -1) {
			var path []string
			for _, item := range reprItemRe.FindAllStringSubmatch(m[1], -1) {
				path = append(path, item[1])
			}
			if len(path) > 0 {
				add(strings.Join(path, "."))
			}
		}
		if m := litellmUnsupportedRe.FindStringSubmatch(t); m != nil {
			list := m[1]
			re := quotedNameRe
			if strings.HasPrefix(list, "{") {
				re = dictKeyRe
			}
			for _, item := range re.FindAllStringSubmatch(list, -1) {
				add(item[1])
			}
		}
	}
	if matchesAny(toolChoiceAutoRe, p.texts) {
		add(ToolChoiceParam)
	}
	return out, len(out) > 0
}

// IsToolChoiceAutoUnsupported reports the vLLM refusal of tool_choice "auto"
// on a server started without --enable-auto-tool-choice / --tool-call-parser:
// tool calling is off for every model it serves.
func IsToolChoiceAutoUnsupported(body []byte) bool {
	if len(body) == 0 {
		return false
	}
	return matchesAny(toolChoiceAutoRe, parse(body).texts)
}

// locParam joins the loc path after the leading "body" with dots; the root
// (text before the first dot) is the request parameter.
func locParam(loc []json.RawMessage) string {
	var parts []string
	for _, raw := range loc {
		var s string
		if json.Unmarshal(raw, &s) != nil {
			var n int64
			if json.Unmarshal(raw, &n) != nil {
				return ""
			}
			continue
		}
		parts = append(parts, s)
	}
	if len(parts) < 2 || parts[0] != "body" {
		return ""
	}
	return strings.Join(parts[1:], ".")
}
