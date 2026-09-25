package apidetect

import (
	"context"
	"fmt"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

// Field names follow tools/server/server-context.cpp (get_res_props,
// get_res_models, get_res_model_info) and common/jinja/caps.cpp:
// https://github.com/ggml-org/llama.cpp/blob/master/tools/server/server-context.cpp
// https://github.com/ggml-org/llama.cpp/blob/master/common/jinja/caps.cpp
const (
	llamaCppSrcProps         = "llamacpp:/props"
	llamaCppSrcNCtx          = "llamacpp:/props.default_generation_settings.n_ctx"
	llamaCppSrcModalities    = "llamacpp:/props.modalities"
	llamaCppSrcToolCalls     = "llamacpp:/props.chat_template_caps.supports_tool_calls"
	llamaCppSrcParallelCalls = "llamacpp:/props.chat_template_caps.supports_parallel_tool_calls"
	llamaCppSrcMetaNCtx      = "llamacpp:/v1/models.meta.n_ctx"
	llamaCppSrcMetaNCtxTrain = "llamacpp:/v1/models.meta.n_ctx_train"
	llamaCppSrcMetaFtype     = "llamacpp:/v1/models.meta.ftype"
	llamaCppSrcModelsCaps    = "llamacpp:/v1/models.models.capabilities"
)

// Both endpoints are guarded by --api-key (only /health is open), so the key
// goes with every request as before.
func (d *Detector) probeLlamaCpp(ctx context.Context, base, apiKey string, ev *apispec.Evidence) bool {
	props := d.getJSON(ctx, base+"/props", apiKey)
	if props == nil {
		return false
	}
	_, hasTemplate := props["chat_template"]
	_, hasSlots := props["total_slots"]
	if !hasTemplate && !hasSlots {
		return false
	}
	ev.Stack = "llamacpp"
	if tpl, ok := props["chat_template"].(string); ok {
		ev.ChatTemplate = tpl
		tracef(ctx, "identified llamacpp by /props; chat template source available (%d chars)", len(tpl))
	} else {
		tracef(ctx, "identified llamacpp by /props; no chat template exposed")
	}

	hasModalities := d.llamaCppPropsFacts(ctx, props, ev)
	d.llamaCppModelsFacts(ctx, base, apiKey, ev, hasModalities)
	return true
}

// llamaCppPropsFacts reads the capability facts of /props and reports
// whether a modalities object was present (then /v1/models' coarser
// "multimodal" flag adds nothing).
func (d *Detector) llamaCppPropsFacts(ctx context.Context, props map[string]any, ev *apispec.Evidence) bool {
	if settings, ok := props["default_generation_settings"].(map[string]any); ok {
		if n, ok := jsonWholeNumber(settings["n_ctx"]); ok {
			if ev.Caps.SetContext(n, llamaCppSrcNCtx) {
				tracef(ctx, "/props default_generation_settings.n_ctx = %d: served context", n)
			} else {
				tracef(ctx, "/props default_generation_settings.n_ctx = %d out of range; ignored", n)
			}
		}
	}

	hasModalities := false
	if mods, ok := props["modalities"].(map[string]any); ok {
		hasModalities = true
		for _, m := range [][2]string{{"vision", "image"}, {"video", "video"}, {"audio", "audio"}} {
			if on, ok := mods[m[0]].(bool); ok && on {
				ev.Caps.AddInputModality(m[1], llamaCppSrcModalities+"."+m[0])
				tracef(ctx, "/props modalities.%s = true: input modality %s", m[0], m[1])
			}
		}
	}

	if caps, ok := props["chat_template_caps"].(map[string]any); ok {
		if v, ok := caps["supports_tool_calls"].(bool); ok {
			ev.Caps.SetTools(v, llamaCppSrcToolCalls)
			tracef(ctx, "/props chat_template_caps.supports_tool_calls = %t: tools", v)
		}
		if v, ok := caps["supports_parallel_tool_calls"].(bool); ok {
			ev.Caps.SetToolsParallel(v, llamaCppSrcParallelCalls)
			tracef(ctx, "/props chat_template_caps.supports_parallel_tool_calls = %t: toolsParallel", v)
		}
		if v, ok := caps["supports_reasoning_effort"].(bool); ok && v {
			tracef(ctx, "/props chat_template_caps.supports_reasoning_effort = true: the template honours reasoning_effort (binding from the stack table)")
		}
	}

	// Grammar-based sampling is built into the server: response_format
	// json_object / json_schema and a raw GBNF grammar per request.
	// https://github.com/ggml-org/llama.cpp/blob/master/tools/server/README.md
	ev.Caps.SetStructured([]string{"json_object", "json_schema", "grammar"}, llamaCppSrcProps)
	tracef(ctx, "/props: llamacpp constrains output with its grammar engine; structured = [json_object json_schema grammar]")
	return hasModalities
}

func (d *Detector) llamaCppModelsFacts(ctx context.Context, base, apiKey string, ev *apispec.Evidence, hasModalities bool) {
	list := d.getJSON(ctx, base+"/v1/models", apiKey)
	if list == nil {
		return
	}

	if data, ok := list["data"].([]any); ok {
		if entry := matchModelEntry(data, ev.ModelName); entry != nil {
			// meta is null while the model is still loading.
			if meta, ok := entry["meta"].(map[string]any); ok {
				d.llamaCppMetaFacts(ctx, meta, ev)
			} else {
				tracef(ctx, "/v1/models meta absent or null (model still loading?); no facts")
			}
		}
	}

	if hasModalities {
		return
	}
	if entry := llamaCppModelsEntry(list["models"], ev.ModelName); entry != nil {
		if caps, ok := entry["capabilities"].([]any); ok {
			for _, c := range caps {
				if s, ok := c.(string); ok && s == "multimodal" {
					ev.Caps.AddInputModality("image", llamaCppSrcModelsCaps)
					tracef(ctx, "/v1/models models[].capabilities has multimodal and /props listed no modalities: input modality image")
					break
				}
			}
		}
	}
}

func (d *Detector) llamaCppMetaFacts(ctx context.Context, meta map[string]any, ev *apispec.Evidence) {
	if n, ok := jsonWholeNumber(meta["n_ctx_train"]); ok {
		if ev.Caps.SetContextTrained(n, llamaCppSrcMetaNCtxTrain) {
			tracef(ctx, "/v1/models meta.n_ctx_train = %d: trained context", n)
		} else {
			tracef(ctx, "/v1/models meta.n_ctx_train = %d out of range; ignored", n)
		}
	}
	if n, ok := jsonWholeNumber(meta["n_ctx"]); ok {
		if _, served := ev.Caps.Sources[apispec.CapFieldContext]; served {
			tracef(ctx, "/v1/models meta.n_ctx = %d: /props already gave the served context", n)
		} else if ev.Caps.SetContext(n, llamaCppSrcMetaNCtx) {
			tracef(ctx, "/v1/models meta.n_ctx = %d: served context (/props had none)", n)
		} else {
			tracef(ctx, "/v1/models meta.n_ctx = %d out of range; ignored", n)
		}
	}
	if raw, present := meta["ftype"]; present {
		if label, ok := llamaCppFtypeLabel(raw); ok {
			if ev.Caps.SetQuantization(label, llamaCppSrcMetaFtype) {
				tracef(ctx, "/v1/models meta.ftype %q: quantization %s", fmt.Sprint(raw), label)
			}
		} else {
			tracef(ctx, "/v1/models meta.ftype %q has no known GGUF type name; ignored", fmt.Sprint(raw))
		}
	}
}

// The Ollama-style models[] list carries one entry per served model; match on
// name / model, else take a sole entry.
func llamaCppModelsEntry(raw any, modelName string) map[string]any {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	var entries []map[string]any
	for _, it := range items {
		if m, ok := it.(map[string]any); ok {
			entries = append(entries, m)
		}
	}
	for _, key := range []string{"name", "model"} {
		for _, m := range entries {
			if id, ok := m[key].(string); ok && modelName != "" && strings.EqualFold(id, modelName) {
				return m
			}
		}
	}
	if len(entries) == 1 {
		return entries[0]
	}
	return nil
}

// llamaCppFtype pairs a llama_ftype enum value with the GGUF file-type name
// and the label llama_ftype_name() prints for it; the server reports the
// latter as meta.ftype, older builds the raw integer.
// https://github.com/ggml-org/llama.cpp/blob/master/include/llama.h (enum llama_ftype)
// https://github.com/ggml-org/llama.cpp/blob/master/src/llama-model-loader.cpp (llama_ftype_name)
type llamaCppFtype struct {
	value int64
	name  string
	label string
}

var llamaCppFtypes = []llamaCppFtype{
	{0, "F32", "all F32"},
	{1, "F16", "F16"},
	{2, "Q4_0", "Q4_0"},
	{3, "Q4_1", "Q4_1"},
	{7, "Q8_0", "Q8_0"},
	{8, "Q5_0", "Q5_0"},
	{9, "Q5_1", "Q5_1"},
	{10, "Q2_K", "Q2_K - Medium"},
	{11, "Q3_K_S", "Q3_K - Small"},
	{12, "Q3_K_M", "Q3_K - Medium"},
	{13, "Q3_K_L", "Q3_K - Large"},
	{14, "Q4_K_S", "Q4_K - Small"},
	{15, "Q4_K_M", "Q4_K - Medium"},
	{16, "Q5_K_S", "Q5_K - Small"},
	{17, "Q5_K_M", "Q5_K - Medium"},
	{18, "Q6_K", "Q6_K"},
	{19, "IQ2_XXS", "IQ2_XXS - 2.0625 bpw"},
	{20, "IQ2_XS", "IQ2_XS - 2.3125 bpw"},
	{21, "Q2_K_S", "Q2_K - Small"},
	{22, "IQ3_XS", "IQ3_XS - 3.3 bpw"},
	{23, "IQ3_XXS", "IQ3_XXS - 3.0625 bpw"},
	{24, "IQ1_S", "IQ1_S - 1.5625 bpw"},
	{25, "IQ4_NL", "IQ4_NL - 4.5 bpw"},
	{26, "IQ3_S", "IQ3_S - 3.4375 bpw"},
	{27, "IQ3_M", "IQ3_S mix - 3.66 bpw"},
	{28, "IQ2_S", "IQ2_S - 2.5 bpw"},
	{29, "IQ2_M", "IQ2_M - 2.7 bpw"},
	{30, "IQ4_XS", "IQ4_XS - 4.25 bpw"},
	{31, "IQ1_M", "IQ1_M - 1.75 bpw"},
	{32, "BF16", "BF16"},
	{36, "TQ1_0", "TQ1_0 - 1.69 bpw ternary"},
	{37, "TQ2_0", "TQ2_0 - 2.06 bpw ternary"},
	{38, "MXFP4_MOE", "MXFP4 MoE"},
	{39, "NVFP4", "NVFP4"},
	{40, "Q1_0", "Q1_0"},
	{41, "Q2_0", "Q2_0"},
}

// LLAMA_FTYPE_GUESSED is or-ed into the value (and printed as a prefix) when
// the file does not state its type; the type itself is still the same.
const llamaCppFtypeGuessed = 1024

const llamaCppGuessedPrefix = "(guessed) "

func llamaCppFtypeLabel(raw any) (string, bool) {
	switch v := raw.(type) {
	case float64:
		n, ok := jsonWholeNumber(v)
		if !ok {
			return "", false
		}
		n &^= llamaCppFtypeGuessed
		for _, ft := range llamaCppFtypes {
			if ft.value == n {
				return ft.name, true
			}
		}
	case string:
		label := strings.TrimPrefix(strings.TrimSpace(v), llamaCppGuessedPrefix)
		for _, ft := range llamaCppFtypes {
			if ft.label == label {
				return ft.name, true
			}
		}
	}
	return "", false
}
