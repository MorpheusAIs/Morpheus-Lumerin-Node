package config

import (
	"encoding/json"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

type ModelConfigLoader struct {
	log          *lib.Logger
	modelConfigs ModelConfigs
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
		log:          log,
		modelConfigs: ModelConfigs{},
	}
}

func (e *ModelConfigLoader) Init() error {
	filePath := "models-config.json"
	modelsConfig, err := lib.ReadJSONFile(filePath)
	if err != nil {
		e.log.Errorf("failed to read models config file: %s", err)

		e.log.Warn("trying to load models config from persistent storage")
		// TODO: load models config from persistent storage

		return err
	}

	var modelConfigs ModelConfigs
	err = json.Unmarshal([]byte(modelsConfig), &modelConfigs)
	if err != nil {
		e.log.Errorf("failed to unmarshal models config: %s", err)
		return err
	}
	e.modelConfigs = modelConfigs
	return nil
}

func (e *ModelConfigLoader) ModelConfigFromID(ID string) *ModelConfig {
	if ID == "" {
		return &ModelConfig{}
	}

	modelConfig := e.modelConfigs[ID]
	if modelConfig == (ModelConfig{}) {
		e.log.Errorf("model config not found for ID: %s", ID)
		return &ModelConfig{}
	}

	return &modelConfig
}
