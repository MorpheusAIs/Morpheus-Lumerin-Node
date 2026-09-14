package aiengine

import (
	"net/http"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

// ApiAdapterFactory creates the appropriate AI engine adapter for the given API type.
// httpClient is optional; when non-nil it overrides the default http.Client used by
// adapters that support it (currently openai and claudeai). Pass nil for default behavior.
func ApiAdapterFactory(apiType string, modelName string, url string, apikey string, parameters ModelParameters, llmTimeout time.Duration, log lib.ILogger, httpClient *http.Client) (AIEngineStream, bool) {
	switch apiType {
	case API_TYPE_OPENAI:
		return NewOpenAIEngine(modelName, url, apikey, llmTimeout, log, httpClient), true
	case API_TYPE_PRODIA_SD:
		return NewProdiaSDEngine(modelName, url, apikey, log), true
	case API_TYPE_PRODIA_SDXL:
		return NewProdiaSDXLEngine(modelName, url, apikey, log), true
	case API_TYPE_PRODIA_V2:
		return NewProdiaV2Engine(modelName, url, apikey, log), true
	case API_TYPE_HYPERBOLIC_SD:
		return NewHyperbolicSDEngine(modelName, url, apikey, parameters, log), true
	case API_TYPE_CLAUDEAI:
		return NewClaudeAIEngine(modelName, url, apikey, llmTimeout, log, httpClient), true
	}
	return nil, false
}
