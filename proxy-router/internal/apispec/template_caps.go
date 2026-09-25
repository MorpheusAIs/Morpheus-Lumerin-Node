package apispec

import (
	"regexp"
	"strings"
)

// templateFacts is what the chat-template scan contributes beyond reasoning
// bindings: capability facts with their sources (tools variable or
// <tool_call> -> tools, vision tokens -> modalities.input image) and the
// reasoning default ({%- set enable_thinking = true|false %}, set thinking =
// ...). They rank below explicit backend facts and above the registry, and
// apply only when the stack is known (unknown wire vocabulary otherwise).
type templateFacts struct {
	caps                  CapabilityFacts
	thinkingDefault       string
	thinkingDefaultSource string
}

const (
	templateSourcePrefix     = "template:chat_template."
	templateSourceTools      = templateSourcePrefix + "tools"
	templateSourceImageToken = templateSourcePrefix + "image_token"
)

// The default a template assigns to its own reasoning switch when the caller
// passed none, e.g. `{%- if not thinking is defined %}{% set thinking = false %}`
// (DeepSeek-V3.1, Hermes 4), `{%- set enable_thinking = false %}` (Cogito v2:
// https://huggingface.co/deepcogito/cogito-v2-preview-llama-70B/raw/main/chat_template.jinja)
// or the one-statement conditional form
// `{%- set enable_thinking = enable_thinking if enable_thinking is defined else True %}`
// (Nemotron 3: https://huggingface.co/nvidia/NVIDIA-Nemotron-3-Nano-30B-A3B-BF16/raw/main/chat_template.jinja).
// RE2 has no backreferences, so the conditional form's three variable names are
// captured separately and required to agree in capsFromTemplate.
// Qwen3 only tests the variable and never assigns it, so it yields no default:
// https://huggingface.co/Qwen/Qwen3-8B/raw/main/tokenizer_config.json
var templateThinkingDefaultRe = regexp.MustCompile(`^\{%-?\s*set\s+(enable_thinking|thinking)\s*=\s*(?:(enable_thinking|thinking)\s+if\s+(enable_thinking|thinking)\s+is\s+defined\s+else\s+)?(true|false)\s*-?%\}$`)

var templateToolsVarRe = regexp.MustCompile(`(^|[^a-zA-Z0-9_.])tools($|[^a-zA-Z0-9_])`)

// A template that names tools only to reject them
// (`{{ raise_exception('This model does not support tools') }}`) proves the
// opposite. Only a raise worded as a rejection counts: tool-capable templates
// also raise on tool misuse, e.g. Llama 3.1 "Cannot put tools in the first
// user message when there's no first user message!" and "This model only
// supports single tool-calls at once!"
// (https://huggingface.co/meta-llama/Llama-3.1-8B-Instruct/raw/main/tokenizer_config.json),
// Mistral Nemo "Tool call IDs should be alphanumeric strings with length 9!"
// (https://huggingface.co/mistralai/Mistral-Nemo-Instruct-2407/raw/main/tokenizer_config.json).
var templateRaiseRe = regexp.MustCompile(`raise_exception\s*\(\s*(?:"[^"]*"|'[^']*')`)

var templateToolsRejectedRe = regexp.MustCompile(`(?:not|n't|un)\s*support`)

var templateToolWordRe = regexp.MustCompile(`\btools?\b`)

// Image placeholder tokens the template emits for image content parts:
// <|vision_start|><|image_pad|> (Qwen2/2.5-VL,
// https://huggingface.co/Qwen/Qwen2.5-VL-7B-Instruct/raw/main/chat_template.json),
// <image> (LLaVA, https://huggingface.co/llava-hf/llava-1.5-7b-hf/raw/main/chat_template.json),
// <|image|> (Llama 3.2 Vision,
// https://www.llama.com/docs/model-cards-and-prompt-formats/llama3_2/),
// <start_of_image> (Gemma 3,
// https://ai.google.dev/gemma/docs/core/prompt-structure).
var templateImageTokens = []string{"<|vision_start|>", "<|image_pad|>", "<image>", "<|image|>", "<start_of_image>"}

func capsFromTemplate(tpl string) templateFacts {
	var f templateFacts
	if tpl == "" {
		return f
	}
	stmts := jinjaStmtRe.FindAllString(tpl, -1)
	for _, stmt := range stmts {
		m := templateThinkingDefaultRe.FindStringSubmatch(strings.ToLower(stmt))
		if m == nil || (m[2] != "" && (m[2] != m[1] || m[3] != m[1])) {
			continue
		}
		f.thinkingDefault = "off"
		if m[4] == "true" {
			f.thinkingDefault = "on"
		}
		f.thinkingDefaultSource = templateSourcePrefix + m[1]
		break
	}
	if templateSupportsTools(tpl, stmts) {
		f.caps.SetTools(true, templateSourceTools)
	}
	for _, token := range templateImageTokens {
		if strings.Contains(tpl, token) {
			f.caps.AddInputModality("image", templateSourceImageToken)
			break
		}
	}
	return f
}

func templateSupportsTools(tpl string, stmts []string) bool {
	if strings.Contains(tpl, "<tool_call>") {
		return true
	}
	for _, raised := range templateRaiseRe.FindAllString(tpl, -1) {
		lower := strings.ToLower(raised)
		if templateToolWordRe.MatchString(lower) && templateToolsRejectedRe.MatchString(lower) {
			return false
		}
	}
	for _, stmt := range stmts {
		if templateToolsVarRe.MatchString(stmt) {
			return true
		}
	}
	return false
}
