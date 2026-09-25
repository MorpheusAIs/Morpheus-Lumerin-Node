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
// (DeepSeek-V3.1, Hermes 4) or `{%- set enable_thinking = false %}` (Cogito v2:
// https://huggingface.co/deepcogito/cogito-v2-preview-llama-70B/raw/main/chat_template.jinja).
// Qwen3 only tests the variable and never assigns it, so it yields no default:
// https://huggingface.co/Qwen/Qwen3-8B/raw/main/tokenizer_config.json
var templateThinkingDefaultRe = regexp.MustCompile(`^\{%-?\s*set\s+(enable_thinking|thinking)\s*=\s*(true|false)\s*-?%\}$`)

var templateToolsVarRe = regexp.MustCompile(`(^|[^a-zA-Z0-9_.])tools($|[^a-zA-Z0-9_])`)

// A template that names tools only to reject them
// (`{{ raise_exception('... tools ...') }}`) proves the opposite.
var templateRaiseRe = regexp.MustCompile(`raise_exception\s*\(\s*(?:"[^"]*"|'[^']*')`)

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
		if m == nil {
			continue
		}
		f.thinkingDefault = "off"
		if m[2] == "true" {
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
	for _, raised := range templateRaiseRe.FindAllString(tpl, -1) {
		if strings.Contains(strings.ToLower(raised), "tool") {
			return false
		}
	}
	if strings.Contains(tpl, "<tool_call>") {
		return true
	}
	for _, stmt := range stmts {
		if templateToolsVarRe.MatchString(stmt) {
			return true
		}
	}
	return false
}
