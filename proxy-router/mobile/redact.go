package mobile

import "regexp"

// Provider-identifying address redaction at the SDK boundary.
//
// Errors that bubble up from the proxy-router routinely include the
// upstream provider's endpoint URL or raw IPv4 + port (e.g.
// `dial tcp 216.81.245.17:18788: connect: connection refused`). Surfacing
// that to external consumers — local OpenAI-compatible gateways, chat UIs,
// MCP servers, anyone embedding this SDK — leaks provider infrastructure
// for no upside; the failure mode (`connection refused`, `i/o timeout`,
// HTTP status) is what callers actually need.
//
// This file applies the redaction to errors that cross the public SDK
// boundary outward. Internal proxy-router logging continues to see the
// raw addresses so operators can still debug. The patterns mirror
// nodeneo/lib/utils/error_redaction.dart and
// nodeneo/go/internal/gateway/redact.go — keep all three in lockstep.

const (
	providerPlaceholder = "<provider endpoint>"
	shortPlaceholder    = "<provider>"
)

var (
	httpURLPattern = regexp.MustCompile(
		`(?i)https?://(?:\[[^\]\s]+\]|[A-Za-z0-9._\-]+)(?::\d+)?(?:/[^\s"\)\],;]*)?`,
	)
	hostPortPattern = regexp.MustCompile(
		`([^A-Za-z0-9./@\-]|^)((?:[A-Za-z0-9\-]+\.)+[A-Za-z0-9\-]+:\d{2,5})([^A-Za-z0-9]|$)`,
	)
	bareIPPattern = regexp.MustCompile(
		`([^\d.]|^)(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})([^\d.]|$)`,
	)
)

// redactProviderEndpointsString replaces provider URLs / host:port pairs /
// bare IPv4s in s with neutral placeholders. Order matters: full URLs are
// stripped first so the host-level passes can't eat fragments of an
// already-cleaned URL. RE2 has no lookbehind, so the host:port and bare-IP
// patterns capture surrounding boundary characters and re-emit them via
// ReplaceAllStringFunc.
func redactProviderEndpointsString(s string) string {
	if s == "" {
		return s
	}
	out := httpURLPattern.ReplaceAllString(s, providerPlaceholder)
	out = hostPortPattern.ReplaceAllStringFunc(out, func(m string) string {
		groups := hostPortPattern.FindStringSubmatch(m)
		if len(groups) < 4 {
			return m
		}
		return groups[1] + shortPlaceholder + groups[3]
	})
	out = bareIPPattern.ReplaceAllStringFunc(out, func(m string) string {
		groups := bareIPPattern.FindStringSubmatch(m)
		if len(groups) < 4 {
			return m
		}
		return groups[1] + shortPlaceholder + groups[3]
	})
	return out
}

// redactedError is a small error wrapper that returns a redacted message
// while preserving the underlying error chain for errors.Is / errors.As.
// Wrapping (rather than allocating a new errors.New) means consumers that
// switch on sentinel errors (e.g. `errors.Is(err, ErrProvider)`) still
// match — only the human-readable string changes.
type redactedError struct {
	original error
	message  string
}

func (e *redactedError) Error() string { return e.message }
func (e *redactedError) Unwrap() error { return e.original }

// redactError returns nil for nil input; otherwise returns an error whose
// message has provider-identifying addresses scrubbed. The original error
// chain is preserved via Unwrap so `errors.Is` / `errors.As` still match.
func redactError(err error) error {
	if err == nil {
		return nil
	}
	cleaned := redactProviderEndpointsString(err.Error())
	if cleaned == err.Error() {
		return err
	}
	return &redactedError{original: err, message: cleaned}
}

// Compile-time interface assertion so future error-wrapper helpers in this
// package don't accidentally regress the Unwrap contract.
var _ interface{ Unwrap() error } = (*redactedError)(nil)
