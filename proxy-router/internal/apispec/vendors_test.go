package apispec

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStackFromHost(t *testing.T) {
	cases := map[string]string{
		"https://api.openai.com/v1":                        "openai",
		"https://api.anthropic.com/v1/messages":            "anthropic",
		"https://openrouter.ai/api/v1":                     "openrouter",
		"https://api.venice.ai/api/v1":                     "venice",
		"https://api.together.xyz/v1":                      "together",
		"https://api.fireworks.ai/inference/v1":            "fireworks",
		"https://api.groq.com/openai/v1":                   "groq",
		"https://api.deepinfra.com/v1/openai":              "deepinfra",
		"https://api.hyperbolic.xyz/v1":                    "hyperbolic",
		"https://api.mistral.ai/v1":                        "mistral",
		"https://generativelanguage.googleapis.com/v1beta": "gemini",
		"https://api.x.ai/v1":                              "xai",
		"https://api.deepseek.com/v1":                      "deepseek",
		"https://api.moonshot.ai/v1":                       "moonshot",
		"https://integrate.api.nvidia.com/v1":              "nvidia-nim",
		"https://api.cerebras.ai/v1":                       "cerebras",
		"https://api.sambanova.ai/v1":                      "sambanova",
		"http://192.168.1.10:8000/v1/chat/completions":     "",
		"http://localhost:11434/v1":                        "",
	}
	for url, want := range cases {
		require.Equal(t, want, StackFromHost(url), "url %s", url)
	}
}

func TestGatedStack(t *testing.T) {
	for _, s := range []string{"venice", "openrouter", "litellm", "anthropic", "openai", "groq", "together", "gemini", "moonshot", "tgi", "lmstudio", "koboldcpp"} {
		require.True(t, gatedStack(s), s)
	}
	for _, s := range []string{"vllm", "sglang", "llamacpp", "ollama", ""} {
		require.False(t, gatedStack(s), s)
	}
}
