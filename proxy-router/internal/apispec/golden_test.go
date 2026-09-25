package apispec

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/stretchr/testify/require"
)

var updateGolden = flag.Bool("update", false, "rewrite testdata/compose_golden.json from the current composition")

// Every fixture carries no capability evidence: the golden file pins the api
// block that consumers (gateway PR #359) read today, so capability work
// cannot change it byte for byte.
func goldenFixtures() (names []string, evidence map[string]Evidence) {
	evidence = map[string]Evidence{}
	add := func(name string, ev Evidence) {
		names = append(names, name)
		evidence[name] = ev
	}
	stacks := make([]string, 0, len(config.StackTransport))
	for stack := range config.StackTransport {
		stacks = append(stacks, stack)
	}
	sort.Strings(stacks)
	models := []string{"Qwen/Qwen3-235B-A22B", "deepseek-r1", "gpt-oss:120b", "claude-opus-4-6", "claude-sonnet-4-5", "claude-fable-5-1", "o4-mini", "meta-llama/Llama-3.3-70B-Instruct", "seed-oss:36b", "MiniMaxAI/MiniMax-M3", "google/gemma-4-31B-it", "moonshotai/Kimi-K2.6", "some-embedder"}
	for _, stack := range stacks {
		for _, name := range models {
			for _, family := range []string{"", "deepseek-v3.1", " Llama "} {
				add(fmt.Sprintf("declared/%s/%s/%q", stack, name, family), Evidence{Stack: stack, ModelName: name, ModelFamily: family, Declared: true})
			}
			add(fmt.Sprintf("detected/%s/%s", stack, name), Evidence{Stack: stack, ModelName: name})
		}
	}
	registry := bindingSet{system.IntentReasoningDisable: {Kind: system.BindingKindBodyParam, Param: "registry.off", ParamType: "boolean", Value: true}}
	add("registry bindings win", Evidence{Stack: "vllm", ModelName: "qwen3-32b", ChatTemplate: qwen3Template, OllamaThinking: true, RegistryBindings: registry})
	add("ollama capability beats template", Evidence{Stack: "ollama", ModelName: "qwen3:32b", ChatTemplate: qwen3Template, OllamaThinking: true})
	add("ollama capability gpt-oss", Evidence{Stack: "ollama", ModelName: "gpt-oss:120b", OllamaThinking: true})
	add("ollama capability deepseek-r1", Evidence{Stack: "ollama", ModelName: "deepseek-r1:70b", OllamaThinking: true})
	add("template beats family default", Evidence{Stack: "llamacpp", ModelName: "qwen3-32b", ChatTemplate: deepseekV31Template})
	add("family default", Evidence{Stack: "vllm", ModelName: "qwen3-32b"})
	add("undetermined stack", Evidence{ModelName: "qwen3-32b"})
	add("undetermined stack always-on", Evidence{ModelName: "deepseek-r1-distill"})
	add("undetermined stack with template", Evidence{Stack: "", ModelName: "Qwen3-8B", ChatTemplate: qwen3Template})
	add("explicit family never overridden", Evidence{Stack: "vllm", ModelName: "alias", ModelFamily: "llama", Architecture: "qwen3", ServedModelID: "deepseek-ai/DeepSeek-R1"})
	add("family chain architecture", Evidence{Stack: "lmstudio", ModelName: "deepseek-r1-distill", Architecture: "deepseek2"})
	add("family chain served id", Evidence{Stack: "sglang", ModelName: "whatever-local-alias", ServedModelID: "deepseek-ai/DeepSeek-R1"})
	add("family chain thinking served id", Evidence{Stack: "vllm", ModelName: "alias", ServedModelID: "Qwen/Qwen3-235B-A22B-Thinking-2507"})
	add("detected litellm no reasoning evidence", Evidence{Stack: "litellm", ModelName: "qwen3-32b"})
	add("detected litellm gateway reasoning", Evidence{Stack: "litellm", ModelName: "qwen3-32b", GatewayReasoning: true})
	add("groq gates family", Evidence{Stack: "groq", ModelName: "qwen/qwen3-32b"})
	add("together gates family", Evidence{Stack: "together", ModelName: "claude-sonnet-4-5"})
	for _, stack := range []string{"tgi", "lmstudio", "koboldcpp"} {
		add("detected-only engine "+stack, Evidence{Stack: stack, ModelName: "qwen3-32b"})
		add("detected-only engine always-on "+stack, Evidence{Stack: stack, ModelName: "deepseek-r1-distill"})
	}
	add("imported parameters", Evidence{Stack: "openrouter", ModelName: "x/y", GatewayReasoning: true, Parameters: []string{"tools", "reasoning"}})
	add("litellm efforts low medium high", Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true, ReasoningEfforts: []string{"low", "medium", "high"}})
	add("litellm efforts with none", Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true, ReasoningEfforts: []string{"none", "low", "high"}})
	add("litellm efforts only none", Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true, ReasoningEfforts: []string{"none"}})
	add("litellm efforts no reasoning evidence", Evidence{Stack: "litellm", ModelName: "llama-3.3-70b", ReasoningEfforts: []string{"low"}})
	add("litellm junk efforts", Evidence{Stack: "litellm", ModelName: "claude-sonnet-4-5", GatewayReasoning: true, ReasoningEfforts: []string{"low", "none\n", "high", "low"}})
	supported := []string{"temperature", "tools", "response_format", "stream_options", "max_tokens"}
	add("venice via litellm", Evidence{Stack: "venice", Via: "litellm", ModelName: "my-chat", ServedModelID: "deepseek-v4-pro", GatewayReasoning: true, LiteLLMSeen: true, LiteLLMSupportedParams: supported})
	add("venice via litellm reasoning_effort supported", Evidence{Stack: "venice", Via: "litellm", ModelName: "deepseek-v4-pro", GatewayReasoning: true, LiteLLMSeen: true, LiteLLMSupportedParams: append([]string{"reasoning_effort"}, supported...)})
	add("vllm via litellm", Evidence{Stack: "vllm", Via: "litellm", ModelName: "my-chat", ServedModelID: "Qwen/Qwen3-32B", LiteLLMSeen: true, LiteLLMSupportedParams: supported})
	add("vllm via litellm narrow", Evidence{Stack: "vllm", Via: "litellm", ModelName: "Qwen/Qwen3-32B", LiteLLMSeen: true, LiteLLMSupportedParams: []string{"temperature", "tools"}})
	add("plain litellm filtered", Evidence{Stack: "litellm", ModelName: "qwen3-32b", GatewayReasoning: true, LiteLLMSeen: true, LiteLLMSupportedParams: supported})
	add("plain litellm llama filtered", Evidence{Stack: "litellm", ModelName: "llama-3.3-70b", LiteLLMSeen: true, LiteLLMSupportedParams: supported})
	add("venice via litellm supported unknown", Evidence{Stack: "venice", Via: "litellm", ModelName: "deepseek-v4-pro", GatewayReasoning: true, LiteLLMSeen: true})
	add("plain litellm supported unknown", Evidence{Stack: "litellm", ModelName: "qwen3-32b", GatewayReasoning: true, LiteLLMSeen: true})
	add("venice via litellm empty supported", Evidence{Stack: "venice", Via: "litellm", ModelName: "deepseek-v4-pro", GatewayReasoning: true, LiteLLMSeen: true, LiteLLMSupportedParams: []string{}})
	add("venice no litellm on path", Evidence{Stack: "venice", ModelName: "deepseek-v4-pro", GatewayReasoning: true, LiteLLMSupportedParams: []string{"temperature"}})
	add("anthropic via litellm", Evidence{Stack: "anthropic", Via: "litellm", ModelName: "claude-opus-4-6", LiteLLMSeen: true, LiteLLMSupportedParams: supported})
	add("anthropic via litellm supported unknown", Evidence{Stack: "anthropic", Via: "litellm", ModelName: "claude-opus-4-6", LiteLLMSeen: true})
	add("venice via litellm junk supported", Evidence{Stack: "venice", Via: "litellm", ModelName: "deepseek-v4-pro", GatewayReasoning: true, LiteLLMSeen: true, LiteLLMSupportedParams: []string{"temperature", "junk name\n", "tools", "tools", "response_format", ""}})
	add("venice via litellm stamped", Evidence{Stack: "venice", Via: "litellm", ModelName: "deepseek-v4-pro", LiteLLMSeen: true})
	return names, evidence
}

func composeGolden(t *testing.T) []byte {
	t.Helper()
	names, evidence := goldenFixtures()
	var buf bytes.Buffer
	for _, name := range names {
		api := Compose(evidence[name])
		raw, err := json.Marshal(api)
		require.NoError(t, err, name)
		buf.WriteString(name)
		buf.WriteString("\t")
		buf.Write(raw)
		buf.WriteString("\n")
	}
	return buf.Bytes()
}

func TestComposeWithoutCapabilityEvidenceIsByteIdentical(t *testing.T) {
	path := filepath.Join("testdata", "compose_golden.txt")
	got := composeGolden(t)
	if *updateGolden {
		require.NoError(t, os.WriteFile(path, got, 0o644))
		return
	}
	want, err := os.ReadFile(path)
	require.NoError(t, err, "run: go test ./internal/apispec -run TestComposeWithoutCapabilityEvidenceIsByteIdentical -update")
	require.Equal(t, string(want), string(got))
}
