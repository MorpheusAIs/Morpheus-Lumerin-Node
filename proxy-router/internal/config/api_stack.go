package config

import (
	"fmt"
	"strings"
)

// Lives in config, not apispec, because apispec imports config and the
// loader validates against it. Adapter names are literals because config
// cannot import aiengine (aiengine imports config); they mirror
// aiengine.API_TYPE_OPENAI and aiengine.API_TYPE_CLAUDEAI.
var StackTransport = map[string]string{
	"openai":     "openai",
	"vllm":       "openai",
	"sglang":     "openai",
	"llamacpp":   "openai",
	"ollama":     "openai",
	"venice":     "openai",
	"openrouter": "openai",
	"litellm":    "openai",
	"anthropic":  "claudeai",
}

// Derived from StackTransport so a new chat-capable stack cannot drift out
// of sync.
func IsChatTransport(apiType string) bool {
	for _, transport := range StackTransport {
		if transport == apiType {
			return true
		}
	}
	return false
}

func NormalizeApiStack(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

func ValidateApiStack(modelID string, cfg ModelConfig) error {
	stack := NormalizeApiStack(cfg.ApiStack)
	if stack == "" {
		return nil
	}
	adapter, ok := StackTransport[stack]
	if !ok {
		return fmt.Errorf("model %s: unknown apiStack %q", modelID, stack)
	}
	if cfg.ApiType != adapter {
		return fmt.Errorf("model %s: apiStack %q requires apiType %q, got %q", modelID, stack, adapter, cfg.ApiType)
	}
	return nil
}
