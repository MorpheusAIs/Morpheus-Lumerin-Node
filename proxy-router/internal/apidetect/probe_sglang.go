package apidetect

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

// Facts come from the SGLang server's model-info document (v0.5.20
// http_server.py: /model_info, with /get_model_info as the deprecated alias
// of the same response) and its OpenAI model list (protocol.py: max_model_len
// = context_len). Both are guarded by --api-key, so the key is sent as today.
// Sources: https://github.com/sgl-project/sglang/blob/v0.5.20/python/sglang/srt/entrypoints/http_server.py
// and https://github.com/sgl-project/sglang/blob/v0.5.20/python/sglang/srt/entrypoints/openai/protocol.py
const sglangModelInfoSource = "sglang:/model_info"

// Every SGLang server supports json_object, json_schema, EBNF grammar and
// regex constrained decoding through its built-in grammar backends:
// https://docs.sglang.ai/advanced_features/structured_outputs.html
var sglangStructured = []string{"json_object", "json_schema", "grammar", "regex"}

func (d *Detector) probeSGLang(ctx context.Context, base, apiKey string, ev *apispec.Evidence) bool {
	endpoint := "/model_info"
	info := d.getJSON(ctx, base+endpoint, apiKey)
	if info == nil {
		endpoint = "/get_model_info"
		info = d.getJSON(ctx, base+endpoint, apiKey)
	}
	if info == nil {
		return false
	}
	path, ok := info["model_path"].(string)
	if !ok {
		return false
	}
	ev.Stack = "sglang"
	ev.ServedModelID = path
	tracef(ctx, "identified sglang by %s; served model path %q", endpoint, path)
	if path == "" {
		if served, ok := info["served_model_name"].(string); ok && served != "" {
			ev.ServedModelID = served
			tracef(ctx, "%s.served_model_name stands in for the empty model_path", endpoint)
		}
	}

	if parser, ok := info["reasoning_parser"].(string); ok && parser != "" {
		ev.Caps.SetReasoning(true, sglangModelInfoSource+".reasoning_parser")
		tracef(ctx, "%s.reasoning_parser %q -> reasoning", endpoint, parser)
	}
	if parser, ok := info["tool_call_parser"].(string); ok && parser != "" {
		ev.Caps.SetTools(true, sglangModelInfoSource+".tool_call_parser")
		tracef(ctx, "%s.tool_call_parser %q -> tools", endpoint, parser)
	}
	if v, ok := info["has_image_understanding"].(bool); ok && v {
		ev.Caps.AddInputModality("image", sglangModelInfoSource+".has_image_understanding")
		tracef(ctx, "%s.has_image_understanding -> modalities.input image", endpoint)
	}
	if v, ok := info["has_audio_understanding"].(bool); ok && v {
		ev.Caps.AddInputModality("audio", sglangModelInfoSource+".has_audio_understanding")
		tracef(ctx, "%s.has_audio_understanding -> modalities.input audio", endpoint)
	}
	if ev.Architecture == "" {
		if archs, ok := info["architectures"].([]any); ok && len(archs) > 0 {
			if arch, ok := archs[0].(string); ok && arch != "" {
				ev.Architecture = arch
				tracef(ctx, "%s.architectures[0] %q -> architecture", endpoint, arch)
			}
		}
	}
	if gen, ok := info["is_generation"].(bool); ok && !gen {
		ev.Caps.SetEmbeddings(true, sglangModelInfoSource+".is_generation")
		ev.Caps.SetOutputModalities([]string{"embeddings"}, sglangModelInfoSource+".is_generation")
		tracef(ctx, "%s.is_generation false -> embeddings, modalities.output embeddings", endpoint)
	}
	ev.Caps.SetStructured(sglangStructured, sglangModelInfoSource)
	tracef(ctx, "%s -> structured %v (built into every SGLang server)", endpoint, sglangStructured)

	if list := d.getJSON(ctx, base+"/v1/models", apiKey); list != nil {
		if data, ok := list["data"].([]any); ok {
			if entry := matchModelEntry(data, ev.ModelName); entry != nil {
				if n, ok := entry["max_model_len"].(float64); ok {
					if ev.Caps.SetContext(int64(n), "sglang:/v1/models.max_model_len") {
						tracef(ctx, "/v1/models.max_model_len %d -> context", int64(n))
					} else {
						tracef(ctx, "/v1/models.max_model_len %v is out of range; ignored", n)
					}
				}
			}
		}
	}
	return true
}
