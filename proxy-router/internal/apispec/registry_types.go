package apispec

import "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"

// RegistryFacts is one model's entry from the generated offline registry
// snapshot (models.dev, LiteLLM model_prices_and_context_window.json). The
// lowest-precedence source: it fills only fields nothing else set, with
// evidence "registry:<name>@<date>". Unset fields stay nil.
type RegistryFacts struct {
	system.ModelCapabilities
	// Name identifies the snapshot (models.dev, litellm, or a merged label).
	Name string
	// Date is the snapshot's generation date, YYYY-MM-DD.
	Date string
}

func (r *RegistryFacts) Source() string {
	if r == nil {
		return ""
	}
	return RegistrySource(r.Name, r.Date)
}

func RegistrySource(name, date string) string { return "registry:" + name + "@" + date }
