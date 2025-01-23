package aiengine

import "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"

func ApiAdapterFactory(apiType string, modelName string, url string, apikey string, parameters ModelParameters, log lib.ILogger) (AIEngineStream, bool) {
	switch apiType {
	case API_TYPE_OPENAI:
		return NewOpenAIEngine(modelName, url, apikey, log), true
	case API_TYPE_PRODIA_SD:
		return NewProdiaSDEngine(modelName, url, apikey, log), true
	case API_TYPE_PRODIA_SDXL:
		return NewProdiaSDXLEngine(modelName, url, apikey, log), true
	case API_TYPE_PRODIA_V2:
		return NewProdiaV2Engine(modelName, url, apikey, log), true
	case API_TYPE_HYPERBOLIC_SD:
		return NewHyperbolicSDEngine(modelName, url, apikey, parameters, log), true
	case API_TYPE_CLAUDEAI:
		return NewClaudeAIEngine(modelName, url, apikey, log), true
	}
	return nil, false
}
