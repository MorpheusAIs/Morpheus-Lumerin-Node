package config

import "fmt"

// StackTransport maps each apiStack preset (the backend API a model
// advertises through the `api` block of its health report) to the transport
// adapter — the apiType — that speaks its protocol. It lives here rather
// than in internal/apispec because apispec imports config, and the loader
// needs the table to validate a model before it is stored.
//
// The adapter names are string literals on purpose: config cannot import
// internal/aiengine (aiengine imports config), so they mirror
// aiengine.API_TYPE_OPENAI (internal/aiengine/openai.go) and
// aiengine.API_TYPE_CLAUDEAI (internal/aiengine/claudeai.go). Legacy adapter
// names (claudeai, prodia-*, hyperbolic-sd) are apiType values only and are
// deliberately absent here.
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

// ValidateApiStack checks the optional apiStack of one model: when set it
// must be a known preset and its transport must equal the model's apiType.
// An empty apiStack is always valid (the api block is opt-in per model).
func ValidateApiStack(modelID string, cfg ModelConfig) error {
	if cfg.ApiStack == "" {
		return nil
	}
	adapter, ok := StackTransport[cfg.ApiStack]
	if !ok {
		return fmt.Errorf("model %s: unknown apiStack %q", modelID, cfg.ApiStack)
	}
	if cfg.ApiType != adapter {
		return fmt.Errorf("model %s: apiStack %q requires apiType %q, got %q", modelID, cfg.ApiStack, adapter, cfg.ApiType)
	}
	return nil
}
