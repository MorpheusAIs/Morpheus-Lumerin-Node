package apidetect

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/apispec"
)

// Facts come from the SGLang server's model-info document (v0.5.20
// http_server.py: /model_info, with /get_model_info as the deprecated alias
// of the same response) and its OpenAI model list (protocol.py: max_model_len
// = context_len). Both are guarded by --api-key, so the key is sent as today.
// Evidence names the endpoint that actually answered, because the alias is
// the only route on servers that predate /model_info.
// Sources: https://github.com/sgl-project/sglang/blob/v0.5.20/python/sglang/srt/entrypoints/http_server.py
// and https://github.com/sgl-project/sglang/blob/v0.5.20/python/sglang/srt/entrypoints/openai/protocol.py

// Every SGLang server supports json_schema, EBNF grammar and regex constrained
// decoding through its built-in grammar backends
// (https://docs.sglang.ai/advanced_features/structured_outputs.html);
// json_object is accepted as ResponseFormat.type and mapped to an object
// json_schema in protocol.py
// (https://github.com/sgl-project/sglang/blob/v0.5.20/python/sglang/srt/entrypoints/openai/protocol.py).
var sglangStructured = []string{"json_object", "json_schema", "grammar", "regex"}

// is_generation is false for embedding, reward and sequence-classification
// models alike (model_config.py is_generation_model). When the server resolved
// an embedding spec it also reports embedding.task (embed | classify | none;
// configs/embedding_model_spec.py); only "embed" serves /v1/embeddings.
func sglangServesEmbeddings(info map[string]any) (bool, string) {
	spec, ok := info["embedding"].(map[string]any)
	if !ok {
		return true, ""
	}
	task, ok := spec["task"].(string)
	if !ok || task == "" {
		return true, ""
	}
	return task == "embed", task
}

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
	src := "sglang:" + endpoint
	ev.Stack = "sglang"
	ev.ServedModelID = path
	tracef(ctx, "identified sglang by %s; served model path %q", endpoint, path)
	served, _ := info["served_model_name"].(string)
	if path == "" && served != "" {
		ev.ServedModelID = served
		tracef(ctx, "%s.served_model_name stands in for the empty model_path", endpoint)
	}

	if parser, ok := info["reasoning_parser"].(string); ok && parser != "" {
		ev.Caps.SetReasoning(true, src+".reasoning_parser")
		tracef(ctx, "%s.reasoning_parser %q -> reasoning", endpoint, parser)
	}
	if parser, ok := info["tool_call_parser"].(string); ok && parser != "" {
		ev.Caps.SetTools(true, src+".tool_call_parser")
		tracef(ctx, "%s.tool_call_parser %q -> tools", endpoint, parser)
	}
	if v, ok := info["has_image_understanding"].(bool); ok && v {
		ev.Caps.AddInputModality("image", src+".has_image_understanding")
		tracef(ctx, "%s.has_image_understanding -> modalities.input image", endpoint)
	}
	if v, ok := info["has_audio_understanding"].(bool); ok && v {
		ev.Caps.AddInputModality("audio", src+".has_audio_understanding")
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
		if embeds, task := sglangServesEmbeddings(info); embeds {
			ev.Caps.SetEmbeddings(true, src+".is_generation")
			ev.Caps.SetOutputModalities([]string{"embeddings"}, src+".is_generation")
			tracef(ctx, "%s.is_generation false -> embeddings, modalities.output embeddings", endpoint)
		} else {
			tracef(ctx, "%s.is_generation false but embedding.task %q -> no embeddings fact", endpoint, task)
		}
	}
	ev.Caps.SetStructured(sglangStructured, src)
	tracef(ctx, "%s -> structured %v (built into every SGLang server)", endpoint, sglangStructured)

	if list := d.getJSON(ctx, base+"/v1/models", apiKey); list != nil {
		if data, ok := list["data"].([]any); ok {
			// The base card's id is served_model_name; LoRA adapters are listed
			// beside it with max_model_len null (http_server.py available_models).
			var entry map[string]any
			if served != "" {
				entry = matchModelEntry(data, served)
			}
			if entry == nil {
				entry = matchModelEntry(data, ev.ModelName)
			}
			if entry != nil {
				if n, ok := jsonWholeNumber(entry["max_model_len"]); ok {
					if ev.Caps.SetContext(n, "sglang:/v1/models.max_model_len") {
						tracef(ctx, "/v1/models.max_model_len %d -> context", n)
					} else {
						tracef(ctx, "/v1/models.max_model_len %d is out of range; ignored", n)
					}
				}
			}
		}
	}
	return true
}
