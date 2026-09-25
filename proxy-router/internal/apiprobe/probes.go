package apiprobe

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/sashabaranov/go-openai"
)

// Request builders for the inference-traffic probes (design contract §5):
// pure functions over gcs.OpenAICompletionRequestExtra, which the OpenAI
// adapter marshals as the typed go-openai fields plus Extra. Nothing here
// sends a request.

// ProbeMaxTokens bounds every probe completion: the verdict is read from the
// status, the error body and the presence of reasoning content, never from
// the answer, so one generated token is enough.
const ProbeMaxTokens = 1

// DefaultPrompt is the health checker's arithmetic prompt shape with fixed
// operands; a probe does not check the answer.
const DefaultPrompt = "What is 2+3? Reply with only the number."

// TinyPNGDataURL is a 1x1 transparent PNG (74 bytes decoded): the smallest
// image an OpenAI-compatible backend will accept as an image_url part.
const TinyPNGDataURL = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAEUlEQVR42mJiYGBgAAQAAP//AA8AA4NgdNEAAAAASUVORK5CYII="

// DummyToolName is the function tool the tool-call probe declares; the
// prompt gives the model no reason to call it.
const DummyToolName = "noop"

var (
	ErrUnsupportedBinding  = errors.New("binding kind cannot be applied to a chat-completions request")
	ErrBindingWithoutParam = errors.New("binding names no parameter")
	ErrBindingWithoutValue = errors.New("binding carries no value; the caller supplies it")
	ErrToolEntryNotObject  = errors.New("tool_entry value is not an object")
)

// NewProbeRequest is a one-turn, non-streaming, max_tokens 1 request with
// prompt as the user message; every probe starts from it.
func NewProbeRequest(prompt string) *gcs.OpenAICompletionRequestExtra {
	if prompt == "" {
		prompt = DefaultPrompt
	}
	req := &gcs.OpenAICompletionRequestExtra{
		ChatCompletionRequest: openai.ChatCompletionRequest{
			Messages: []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: prompt}},
			Stream:   false,
		},
	}
	WithMaxTokens(req, ProbeMaxTokens)
	return req
}

func WithMaxTokens(req *gcs.OpenAICompletionRequestExtra, n int) *gcs.OpenAICompletionRequestExtra {
	req.MaxTokens = n
	return req
}

// WithDummyToolAuto declares one function tool with an empty object schema
// and tool_choice "auto": the shape vLLM refuses without
// --enable-auto-tool-choice (apiprobe.IsToolChoiceAutoUnsupported) and every
// tool-capable backend accepts.
func WithDummyToolAuto(req *gcs.OpenAICompletionRequestExtra) *gcs.OpenAICompletionRequestExtra {
	req.Tools = append(req.Tools, openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        DummyToolName,
			Description: "Does nothing.",
			Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		},
	})
	req.ToolChoice = "auto"
	return req
}

// WithJSONSchema asks for response_format json_schema with a trivial object
// schema (https://platform.openai.com/docs/guides/structured-outputs).
func WithJSONSchema(req *gcs.OpenAICompletionRequestExtra) *gcs.OpenAICompletionRequestExtra {
	req.ResponseFormat = &openai.ChatCompletionResponseFormat{
		Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
		JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
			Name:   "probe",
			Schema: json.RawMessage(`{"type":"object","properties":{"answer":{"type":"string"}},"required":["answer"],"additionalProperties":false}`),
		},
	}
	return req
}

// WithImagePart turns the first user message into a content array of one
// text part (its text) and one image_url part carrying TinyPNGDataURL; a
// request without a user message gets one with DefaultPrompt.
func WithImagePart(req *gcs.OpenAICompletionRequestExtra) *gcs.OpenAICompletionRequestExtra {
	idx := -1
	for i, m := range req.Messages {
		if m.Role == openai.ChatMessageRoleUser {
			idx = i
			break
		}
	}
	if idx < 0 {
		req.Messages = append(req.Messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: DefaultPrompt})
		idx = len(req.Messages) - 1
	}
	text := req.Messages[idx].Content
	if text == "" {
		for _, part := range req.Messages[idx].MultiContent {
			if part.Type == openai.ChatMessagePartTypeText && part.Text != "" {
				text = part.Text
				break
			}
		}
	}
	if text == "" {
		text = DefaultPrompt
	}
	req.Messages[idx].Content = ""
	req.Messages[idx].MultiContent = []openai.ChatMessagePart{
		{Type: openai.ChatMessagePartTypeText, Text: text},
		{Type: openai.ChatMessagePartTypeImageURL, ImageURL: &openai.ChatMessageImageURL{URL: TinyPNGDataURL}},
	}
	return req
}

// ApplyBinding writes a binding's fixed value into the request the way a
// consumer would: body_param and template_kwarg set the dotted path (the
// typed go-openai field when the root is a known chat-completions field,
// Extra otherwise; an array value extends an existing array, an object
// value merges into an existing object), tool_entry appends the value to
// tools[]. system_prompt and native_body_param are not request parameters
// and are refused with ErrUnsupportedBinding.
func ApplyBinding(req *gcs.OpenAICompletionRequestExtra, b *system.ParamBinding) error {
	if req == nil || b == nil {
		return errors.New("nil request or binding")
	}
	switch b.Kind {
	case system.BindingKindToolEntry:
		if b.Value == nil {
			return ErrBindingWithoutValue
		}
		value, err := normalize(b.Value)
		if err != nil {
			return err
		}
		return appendToolEntry(req, value)
	case system.BindingKindBodyParam, system.BindingKindTemplateKwarg:
		if b.Param == "" {
			return ErrBindingWithoutParam
		}
		if b.Value == nil {
			return ErrBindingWithoutValue
		}
		value, err := normalize(b.Value)
		if err != nil {
			return err
		}
		return SetBodyPath(req, b.Param, value)
	}
	return fmt.Errorf("%w: %q", ErrUnsupportedBinding, b.Kind)
}

// SetBodyPath sets the dotted path in the request body to value with
// ApplyBinding's rules; value is any JSON-shaped Go value.
func SetBodyPath(req *gcs.OpenAICompletionRequestExtra, path string, value any) error {
	value, err := normalize(value)
	if err != nil {
		return err
	}
	parts := strings.Split(path, ".")
	for _, p := range parts {
		if p == "" {
			return fmt.Errorf("malformed parameter path %q", path)
		}
	}
	root, rest := parts[0], parts[1:]

	if field, ok := knownField(reflect.ValueOf(&req.ChatCompletionRequest).Elem(), root); ok {
		var tree any
		if !field.IsZero() {
			raw, err := json.Marshal(field.Interface())
			if err != nil {
				return err
			}
			if err := json.Unmarshal(raw, &tree); err != nil {
				return err
			}
		}
		raw, err := json.Marshal(setPath(tree, rest, value))
		if err != nil {
			return err
		}
		fresh := reflect.New(field.Type())
		if err := json.Unmarshal(raw, fresh.Interface()); err != nil {
			return fmt.Errorf("parameter %s: %w", path, err)
		}
		field.Set(fresh.Elem())
		return nil
	}

	var tree any
	if raw, ok := req.Extra[root]; ok && len(raw) > 0 {
		if err := json.Unmarshal(raw, &tree); err != nil {
			return err
		}
	}
	raw, err := json.Marshal(setPath(tree, rest, value))
	if err != nil {
		return err
	}
	if req.Extra == nil {
		req.Extra = map[string]json.RawMessage{}
	}
	req.Extra[root] = raw
	return nil
}

// appendToolEntry appends value to the wire tools[]. The typed Tools cannot
// hold a vendor tool object ({"googleSearch":{}}), so the whole array moves
// to Extra["tools"], which the request's MarshalJSON writes over the typed
// field.
func appendToolEntry(req *gcs.OpenAICompletionRequestExtra, value any) error {
	if _, ok := value.(map[string]any); !ok {
		return ErrToolEntryNotObject
	}
	var list []any
	if raw, ok := req.Extra["tools"]; ok && len(raw) > 0 {
		if err := json.Unmarshal(raw, &list); err != nil {
			return err
		}
	} else if len(req.Tools) > 0 {
		raw, err := json.Marshal(req.Tools)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, &list); err != nil {
			return err
		}
	}
	raw, err := json.Marshal(append(list, value))
	if err != nil {
		return err
	}
	if req.Extra == nil {
		req.Extra = map[string]json.RawMessage{}
	}
	req.Extra["tools"] = raw
	req.Tools = nil
	return nil
}

// normalize round-trips value through JSON so the tree below sees only
// map[string]any, []any, float64, string, bool and nil.
func normalize(value any) (any, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func setPath(tree any, path []string, value any) any {
	if len(path) == 0 {
		return mergeValue(tree, value)
	}
	m, _ := tree.(map[string]any)
	if m == nil {
		m = map[string]any{}
	}
	m[path[0]] = setPath(m[path[0]], path[1:], value)
	return m
}

// mergeValue: arrays extend, objects merge key by key, anything else is
// replaced.
func mergeValue(existing, value any) any {
	switch v := value.(type) {
	case []any:
		if have, ok := existing.([]any); ok {
			return append(append([]any(nil), have...), v...)
		}
	case map[string]any:
		if have, ok := existing.(map[string]any); ok {
			merged := make(map[string]any, len(have)+len(v))
			for k, x := range have {
				merged[k] = x
			}
			for k, x := range v {
				merged[k] = mergeValue(have[k], x)
			}
			return merged
		}
	}
	return value
}

// knownField finds the addressable field of a go-openai request struct whose
// json tag is name, descending into embedded structs
// (ChatCompletionRequestExtensions).
func knownField(v reflect.Value, name string) (reflect.Value, bool) {
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			if found, ok := knownField(v.Field(i), name); ok {
				return found, true
			}
			continue
		}
		tag, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if tag == "" || tag == "-" {
			continue
		}
		if tag == name {
			return v.Field(i), true
		}
	}
	return reflect.Value{}, false
}

// RenderURL is vLLM's chat render endpoint next to the model's chat
// endpoint: POST /v1/chat/completions/render
// (vllm/entrypoints/scale_out/render/api_router.py). It is derived from the
// model's own apiUrl, so the key never leaves the model's host. Empty when
// apiURL has no host.
func RenderURL(apiURL string) string {
	u, err := url.Parse(apiURL)
	if err != nil || u.Host == "" {
		return ""
	}
	path := strings.TrimSuffix(u.Path, "/")
	if strings.HasSuffix(path, "/chat/completions") {
		return u.Scheme + "://" + u.Host + path + "/render"
	}
	for _, suffix := range []string{"/completions", "/embeddings", "/messages"} {
		path = strings.TrimSuffix(path, suffix)
	}
	path = strings.TrimSuffix(path, "/")
	if !strings.HasSuffix(path, "/v1") {
		path += "/v1"
	}
	return u.Scheme + "://" + u.Host + path + "/chat/completions/render"
}

// RenderRequestBody is the render endpoint's input: a chat-completions
// request (the endpoint takes ChatCompletionRequest) with max_tokens 1 and,
// when kwargs is non-empty, chat_template_kwargs.
func RenderRequestBody(model, prompt string, kwargs map[string]any) ([]byte, error) {
	req := NewProbeRequest(prompt)
	req.Model = model
	if len(kwargs) > 0 {
		req.ChatTemplateKwargs = kwargs
	}
	return json.Marshal(req)
}

// RenderedTokens reads the rendered prompt out of a render response: the
// token_ids of the GenerateRequest it returns
// (vllm/entrypoints/scale_out/token_in_token_out/protocol.py). ok is false
// when the body carries none.
func RenderedTokens(body []byte) (tokens []int64, ok bool) {
	var out struct {
		TokenIDs []int64 `json:"token_ids"`
	}
	if err := json.Unmarshal(bound(body), &out); err != nil || len(out.TokenIDs) == 0 {
		return nil, false
	}
	return out.TokenIDs, true
}
