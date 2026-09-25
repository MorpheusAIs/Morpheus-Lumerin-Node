// Package registry is the offline model registry: a generated snapshot of
// models.dev and LiteLLM's model_prices_and_context_window.json (both MIT)
// keyed by normalised model id, and a Lookup that turns one entry into
// apispec.RegistryFacts — the lowest-precedence capability source, which the
// composer lets fill only fields nothing else set.
//
// Regenerate with `make registry` (or `go generate ./internal/apispec/registry/...`);
// the generator accepts -date=YYYY-MM-DD so a re-run on the same day is
// byte-identical.
package registry

//go:generate go run ./gen
