package system

type FD struct {
	ID   string
	Path string
}

type SetEthNodeURLReq struct {
	URLs []string `json:"urls" binding:"required" validate:"required,url"`
}

type ConfigResponse struct {
	Version       string
	Commit        string
	DerivedConfig interface{}
	Config        interface{}
}

type HealthCheckResponse struct {
	Status     string              `json:"status"`
	Version    string              `json:"version"`
	Uptime     string              `json:"uptime"`
	Components map[string]string   `json:"components,omitempty"`
	Models     []ModelHealthReport `json:"models,omitempty"`
}

const (
	ModelHealthStatusHealthy   = "healthy"
	ModelHealthStatusUnhealthy = "unhealthy"
	ModelHealthStatusNoBid     = "no_bid"
	ModelHealthStatusNoModel   = "no_model_configured"
	ModelHealthStatusSkipped   = "skipped"
	// ModelHealthStatusTeeUnverified marks a TEE-tagged model whose backend
	// TEE self-attestation failed (or never succeeded) on this provider.
	// Session requests for the model would be rejected, so it is reported
	// separately from a plain backend probe failure.
	ModelHealthStatusTeeUnverified = "tee_unverified"

	ModelHealthErrorTimeout     = "timeout"
	ModelHealthErrorConnection  = "connection"
	ModelHealthErrorBadResponse = "bad_response"
	ModelHealthErrorRateLimited = "rate_limited"
	ModelHealthErrorTypeLookup  = "model_type_lookup"
	// ModelHealthErrorSessionErrors marks a model flipped to unhealthy by
	// consecutive real-session prompt failures rather than a scheduled probe.
	ModelHealthErrorSessionErrors = "session_errors"
	// ModelHealthErrorTeeAttestation accompanies ModelHealthStatusTeeUnverified.
	ModelHealthErrorTeeAttestation = "tee_attestation"
)

// ModelHealthReport is the sanitized per-model self-report exposed on the
// public /healthcheck endpoint. It covers the union of configured models and
// this provider's active bids, and intentionally carries only on-chain data
// and derived status — never the private modelName, apiUrl or apiKey from
// models-config.json. The ModelName here is the public name registered
// on-chain for the model ID, not the private backend model string.
type ModelHealthReport struct {
	ModelID       string `json:"modelId"`
	ModelName     string `json:"modelName,omitempty"`
	ModelType     string `json:"modelType,omitempty"`
	HasActiveBid  bool   `json:"hasActiveBid"`
	BidID         string `json:"bidId,omitempty"`
	Status        string `json:"status"`
	LastHealthy   int64  `json:"lastHealthy,omitempty"`
	LastChecked   int64  `json:"lastChecked"`
	LatencyMs     int64  `json:"latencyMs,omitempty"`
	PromptCorrect *bool  `json:"promptCorrect,omitempty"`
	ErrorKind     string `json:"errorKind,omitempty"`
	// HttpStatus is the HTTP status code returned by the upstream backend
	// when the probe failed with a non-200 response (e.g. 402, 429).
	// Zero when the probe succeeded or never got an HTTP response.
	HttpStatus int `json:"httpStatus,omitempty"`
}

type StatusRes struct {
	Status string `json:"status"`
}

func OkRes() StatusRes {
	return StatusRes{Status: "ok"}
}
