package apispec

import (
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

// Shapes lifted from the cited templates (Cogito v2, Hermes 4, Qwen3,
// Qwen2.5-VL, LLaVA); the text itself is trimmed to the statements under test.
const (
	cogitoTemplate = `{%- if not enable_thinking is defined %}
    {%- set enable_thinking = false %}
{%- endif %}
{%- if enable_thinking %}{{- '<think>\n' }}{%- endif %}`

	hermes4Template = `{%- if not thinking is defined %}{% set thinking = false %}{% endif %}
{%- if tools %}{%- for tool in tools %}{{ tool | tojson }}{%- endfor %}{%- endif %}
{%- if thinking %}{%- set system_prompt = thinking_prompt %}{%- endif %}`

	qwen3ToolsTemplate = `{%- if tools %}{{- "# Tools" }}{%- for tool in tools %}{{- tool | tojson }}{%- endfor %}{%- endif %}
{{- '<tool_call>\n{"name": "' }}{{ tool_call.name }}{{- '}\n</tool_call>' }}
{%- if enable_thinking is defined and enable_thinking is false %}<think></think>{%- endif %}`

	qwenVLTemplate = `{%- for content in message['content'] %}{%- if content['type'] == 'image' %}<|vision_start|><|image_pad|><|vision_end|>{%- endif %}{%- endfor %}`

	llavaTemplate = `{% for content in message['content'] | selectattr('type', 'equalto', 'image') %}{{ '<image>\n' }}{% endfor %}`

	toolsRejectedTemplate = `{%- if tools is defined and tools %}{{ raise_exception('This model does not support tools') }}{%- endif %}`
)

func TestCapsFromTemplateThinkingDefault(t *testing.T) {
	cases := []struct {
		name, tpl, want, source string
	}{
		{"cogito off", cogitoTemplate, "off", "template:chat_template.enable_thinking"},
		{"enable_thinking on", `{%- set enable_thinking = true %}`, "on", "template:chat_template.enable_thinking"},
		{"hermes thinking off", hermes4Template, "off", "template:chat_template.thinking"},
		{"deepseek thinking off", deepseekV31Template, "off", "template:chat_template.thinking"},
		{"python-cased literal", `{% set thinking = True %}`, "on", "template:chat_template.thinking"},
		{"qwen3 tests but never assigns", qwen3Template, "", ""},
		{"prose is not a statement", "set enable_thinking = false {{ messages }}", "", ""},
		{"assignment to another variable", `{%- set thinking_prompt = "x" %}`, "", ""},
		{"empty", "", "", ""},
	}
	for _, c := range cases {
		got := capsFromTemplate(c.tpl)
		require.Equal(t, c.want, got.thinkingDefault, c.name)
		require.Equal(t, c.source, got.thinkingDefaultSource, c.name)
	}
}

func TestCapsFromTemplateTools(t *testing.T) {
	for name, tpl := range map[string]string{
		"if tools":          qwen3ToolsTemplate,
		"tools is defined":  `{%- if tools is defined and tools is not none %}x{%- endif %}`,
		"for tool in tools": `{%- for tool in tools %}{{ tool }}{%- endfor %}`,
		"tool_call marker":  `{{ '<tool_call>' }}`,
		"hermes":            hermes4Template,
	} {
		got := capsFromTemplate(tpl)
		require.NotNil(t, got.caps.Tools, name)
		require.True(t, *got.caps.Tools, name)
		require.Equal(t, "template:chat_template.tools", got.caps.Sources[CapFieldTools], name)
	}
	for name, tpl := range map[string]string{
		"no tools":              qwen3Template,
		"tool_calls only":       `{%- if message.tool_calls %}x{%- endif %}`,
		"tools_json identifier": `{%- set tools_json = 1 %}`,
		"attribute":             `{%- if message.tools %}x{%- endif %}`,
		"prose":                 "tools are great {{ messages }}",
		"rejected":              toolsRejectedTemplate,
		"empty":                 "",
	} {
		got := capsFromTemplate(tpl)
		require.Nil(t, got.caps.Tools, name)
		require.Empty(t, got.caps.Sources[CapFieldTools], name)
	}
}

func TestCapsFromTemplateImageTokens(t *testing.T) {
	for name, tpl := range map[string]string{
		"qwen-vl":        qwenVLTemplate,
		"llava":          llavaTemplate,
		"llama vision":   `{{ '<|image|>' }}`,
		"gemma 3":        `{{ '<start_of_image>' }}`,
		"image_pad only": `{{ '<|image_pad|>' }}`,
	} {
		got := capsFromTemplate(tpl)
		require.NotNil(t, got.caps.Modalities, name)
		require.Equal(t, []string{"image"}, got.caps.Modalities.Input, name)
		require.Equal(t, "template:chat_template.image_token", got.caps.Sources[CapFieldModalitiesInput], name)
	}
	for name, tpl := range map[string]string{
		"text only": qwen3ToolsTemplate,
		"empty":     "",
	} {
		require.Nil(t, capsFromTemplate(tpl).caps.Modalities, name)
	}
}

func TestCapsFromTemplateEmptyTemplateIsEmpty(t *testing.T) {
	got := capsFromTemplate("")
	require.True(t, got.caps.Empty())
	require.Empty(t, got.thinkingDefault)
}

func TestComposeTemplateFacts(t *testing.T) {
	tpl := cogitoTemplate + qwen3ToolsTemplate + qwenVLTemplate
	api, trace := ComposeWithTrace(Evidence{Stack: "llamacpp", ModelName: "cogito-v2-preview-llama-70b", ChatTemplate: tpl})
	require.NotNil(t, api)
	require.Equal(t, "cogito", api.ModelFamily)
	require.Equal(t, system.ThinkingModeControllable, api.Thinking.Mode)
	require.Equal(t, "off", api.Thinking.Default)
	require.Equal(t, "template:chat_template.enable_thinking", api.Evidence[system.EvidenceKeyThinkingDefault])
	require.NotNil(t, api.Capabilities)
	require.True(t, *api.Capabilities.Tools)
	require.Equal(t, "template:chat_template.tools", api.Evidence[system.EvidenceKeyCapability(CapFieldTools)])
	require.Equal(t, []string{"image"}, api.Capabilities.Modalities.Input)
	require.Equal(t, "template:chat_template.image_token", api.Evidence[system.EvidenceKeyCapability(CapFieldModalitiesInput)])
	require.True(t, *api.Capabilities.Reasoning)
	require.Equal(t, SourceDerivedBindings, api.Evidence[system.EvidenceKeyCapability(CapFieldReasoning)])
	require.Contains(t, trace, "thinking.default = off (template:chat_template.enable_thinking)")
	require.Contains(t, trace, "capabilities.tools = true (template:chat_template.tools)")
}

func TestComposeTemplateFactsSkippedWithoutStack(t *testing.T) {
	tpl := cogitoTemplate + qwen3ToolsTemplate + qwenVLTemplate
	api := Compose(Evidence{ModelName: "cogito-v2-preview-llama-70b", ChatTemplate: tpl})
	require.NotNil(t, api)
	require.Nil(t, api.Thinking)
	require.Nil(t, api.Capabilities)
	require.Nil(t, api.Evidence)
}

func TestComposeBackendFactsBeatTemplate(t *testing.T) {
	var ev Evidence
	ev.Stack = "llamacpp"
	ev.ModelName = "qwen3-32b"
	ev.ChatTemplate = qwen3ToolsTemplate + qwenVLTemplate
	ev.Caps.SetTools(false, "llamacpp:/props.chat_template_caps.supports_tool_calls")
	ev.Caps.SetInputModalities([]string{"text"}, "llamacpp:/props.modalities")
	api := Compose(ev)
	require.False(t, *api.Capabilities.Tools)
	require.Equal(t, "llamacpp:/props.chat_template_caps.supports_tool_calls", api.Evidence[system.EvidenceKeyCapability(CapFieldTools)])
	require.Equal(t, []string{"text"}, api.Capabilities.Modalities.Input)
	require.Equal(t, "llamacpp:/props.modalities", api.Evidence[system.EvidenceKeyCapability(CapFieldModalitiesInput)])
}

func TestComposeReportedDefaultBeatsTemplate(t *testing.T) {
	api := Compose(Evidence{Stack: "ollama", ModelName: "cogito:70b", ChatTemplate: cogitoTemplate, OllamaThinking: true, ThinkingDefault: "on", ThinkingDefaultSource: "ollama:/api/show.thinking.default"})
	require.Equal(t, "on", api.Thinking.Default)
	require.Equal(t, "ollama:/api/show.thinking.default", api.Evidence[system.EvidenceKeyThinkingDefault])
}
