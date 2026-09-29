package apispec

import (
	"net/url"
	"strings"
)

var hostVendors = map[string]string{
	"api.openai.com":                    "openai",
	"api.anthropic.com":                 "anthropic",
	"openrouter.ai":                     "openrouter",
	"venice.ai":                         "venice",
	"api.together.xyz":                  "together",
	"api.together.ai":                   "together",
	"api.fireworks.ai":                  "fireworks",
	"api.groq.com":                      "groq",
	"api.deepinfra.com":                 "deepinfra",
	"api.hyperbolic.xyz":                "hyperbolic",
	"api.mistral.ai":                    "mistral",
	"generativelanguage.googleapis.com": "gemini",
	"api.x.ai":                          "xai",
	"api.deepseek.com":                  "deepseek",
	"api.moonshot.ai":                   "moonshot",
	"api.moonshot.cn":                   "moonshot",
	"integrate.api.nvidia.com":          "nvidia-nim",
	"api.cerebras.ai":                   "cerebras",
	"api.sambanova.ai":                  "sambanova",
}

func StackFromHost(apiURL string) string {
	u, err := url.Parse(apiURL)
	if err != nil {
		return ""
	}
	host := strings.ToLower(u.Hostname())
	for domain, vendor := range hostVendors {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return vendor
		}
	}
	return ""
}

func isHostedVendor(stack string) bool {
	for _, vendor := range hostVendors {
		if vendor == stack {
			return true
		}
	}
	return false
}

// Engines detection can name but no preset documents: nothing establishes
// that they honour chat-template kwargs.
var detectedOnlyEngines = map[string]bool{"tgi": true, "lmstudio": true, "koboldcpp": true}

// Family defaults are HF chat-template kwargs or a vendor's native params:
// wrong or unverified on gateways, other hosted vendors and undocumented
// engines.
func gatedStack(stack string) bool {
	return gatewayStacks[stack] || isHostedVendor(stack) || detectedOnlyEngines[stack]
}
