package apispec

import (
	"net/url"
	"strings"
)

// hostVendors maps well-known hosted API hostnames to a stack identifier.
// Matched exactly or as a domain suffix, so api.eu.example vanity subdomains
// of the same vendor still resolve. Hostname recognition is knowledge, not
// I/O: the runtime detector (internal/apidetect) calls StackFromHost and
// Compose gates family defaults on the result.
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

// StackFromHost recognizes hosted vendors by hostname; returns "" for
// self-hosted / unknown hosts (those are fingerprinted by probing instead).
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

// isHostedVendor reports whether stack names a hosted API vendor (as opposed
// to a self-hosted serving engine or an unrecognized host).
func isHostedVendor(stack string) bool {
	for _, vendor := range hostVendors {
		if vendor == stack {
			return true
		}
	}
	return false
}

// gatedStack reports whether family defaults must not be advertised for
// stack unless the family's native vendor is that stack: gateways translate
// requests for upstream providers and hosted vendors are not HF-template
// engines, so template kwargs / vendor-native params would be wrong shapes.
func gatedStack(stack string) bool {
	return gatewayStacks[stack] || isHostedVendor(stack)
}
