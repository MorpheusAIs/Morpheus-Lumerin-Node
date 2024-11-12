package aiengine

import "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"

func ApiAdapterFactory(apiType string, modelName string, url string, apikey string, log lib.ILogger) (AIEngineStream, bool) {
	switch apiType {
	case "openai":
		return NewOpenAIEngine(modelName, url, apikey, log), true
	case "prodia":
		return NewProdiaEngine(modelName, url, apikey, log), true
	}
	return nil, false
}
