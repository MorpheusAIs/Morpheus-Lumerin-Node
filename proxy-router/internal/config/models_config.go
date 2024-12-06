package config

import (
	"encoding/json"
	"fmt"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

type ModelConfigLoader struct {
	log             *lib.Logger
	modelConfigsMap ModelConfigs
	path            string
}

type ModelConfig struct {
	ModelName       string `json:"modelName"`
	ApiType         string `json:"apiType"`
	ApiURL          string `json:"apiUrl"`
	ApiKey          string `json:"apiKey"`
	ConcurrentSlots int    `json:"concurrentSlots"`
	CapacityPolicy  string `json:"capacityPolicy"`
}

type ModelConfigs map[string]ModelConfig
type ModelConfigsV2 struct {
	Models []struct {
		ID string `json:"modelId"`
		ModelConfig
	} `json:"models"`
}

func NewModelConfigLoader(path string, log *lib.Logger) *ModelConfigLoader {
	return &ModelConfigLoader{
		log:             log,
		modelConfigsMap: ModelConfigs{},
		path:            path,
	}
}

func (e *ModelConfigLoader) Init() error {
	filePath := "models-config.json"
	if e.path != "" {
		filePath = e.path
	}

	modelsConfig, err := lib.ReadJSONFile(filePath)
	if err != nil {
		e.log.Errorf("failed to read models config file: %s", err)

		e.log.Warn("trying to load models config from persistent storage")
		// TODO: load models config from persistent storage

		return err
	}
	e.log.Infof("models config loaded from file: %s", filePath)

	// check config format
	var cfgMap map[string]json.RawMessage
	err = json.Unmarshal([]byte(modelsConfig), &cfgMap)
	if err != nil {
		return fmt.Errorf("invalid models config format: %s", err)
	}
	if cfgMap["models"] != nil {
		var modelConfigsV2 ModelConfigsV2
		err = json.Unmarshal([]byte(modelsConfig), &modelConfigsV2)
		if err != nil {
			return fmt.Errorf("invalid models config V2 format: %s", err)
		}
		for _, v := range modelConfigsV2.Models {
			e.modelConfigsMap[v.ID] = v.ModelConfig
			e.log.Infof("local model: %s", v.ModelName)
		}
		return nil
	}

	e.log.Warnf("failed to unmarshal to new models config, trying legacy")

	// try old config format
	var modelConfigs ModelConfigs
	err = json.Unmarshal([]byte(modelsConfig), &modelConfigs)
	if err != nil {
		e.log.Errorf("failed to unmarshal models config: %s", err)
		return err
	}

	e.modelConfigsMap = modelConfigs
	return nil
}

func (e *ModelConfigLoader) ModelConfigFromID(ID string) *ModelConfig {
	if ID == "" {
		return &ModelConfig{}
	}

	modelConfig := e.modelConfigsMap[ID]
	if modelConfig == (ModelConfig{}) {
		e.log.Warnf("model config not found for ID: %s", ID)
		return &ModelConfig{}
	}

	return &modelConfig
}

func (e *ModelConfigLoader) GetAll() ([]string, []ModelConfig) {
	var modelConfigs []ModelConfig
	var modelIDs []string
	for ID, v := range e.modelConfigsMap {
		modelConfigs = append(modelConfigs, v)
		modelIDs = append(modelIDs, ID)
	}

	return modelIDs, modelConfigs
}
