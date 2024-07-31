package config

import (
	"encoding/json"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

type ModelConfigLoader struct {
	log *lib.Logger
}

type ModelConfig struct {
	ModelName string `json:"modelName"`
	ApiType   string `json:"apiType"`
	ApiURL    string `json:"apiUrl"`
	ApiKey    string `json:"apiKey"`
}

type ModelConfigs map[string]ModelConfig

func NewModelConfigLoader(log *lib.Logger) *ModelConfigLoader {
	return &ModelConfigLoader{
		log: log,
	}
}

func (e *ModelConfigLoader) ModelConfigFromID(ID string) *ModelConfig {
	if ID == "" {
		return &ModelConfig{}
	}

	filePath := "models-config.json"
	modelsConfig, err := lib.ReadJSONFile(filePath)
	if err != nil {
		e.log.Errorf("failed to read models config file: %s", err)

		e.log.Warn("trying to load models config from persistent storage")
		// TODO: load models config from persistent storage

		return &ModelConfig{}
	}

	var modelConfigs ModelConfigs
	err = json.Unmarshal([]byte(modelsConfig), &modelConfigs)
	if err != nil {
		e.log.Errorf("failed to unmarshal models config: %s", err)
		return &ModelConfig{}
	}

	modelConfig := modelConfigs[ID]
	if modelConfig == (ModelConfig{}) {
		e.log.Errorf("model config not found for ID: %s", ID)
		return &ModelConfig{}
	}

	return &modelConfig
}
