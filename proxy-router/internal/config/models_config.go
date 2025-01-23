package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

const (
	ConfigPathDefault = "models-config.json"
)

var (
	ErrModelNotFound = errors.New("model not found in blockchain, local-only")
	ErrValidate      = errors.New("cannot perform validation")
	ErrConnect       = errors.New("cannot connect to the model")
)

type BlockchainChecker interface {
	ModelExists(ctx context.Context, ID common.Hash) (bool, error)
}

type ConnectionChecker interface {
	TryConnect(ctx context.Context, url string) error
}

type ModelConfigLoader struct {
	log               lib.ILogger
	modelConfigs      ModelConfigs
	validator         Validator
	blockchainChecker BlockchainChecker
	connectionChecker ConnectionChecker
	configPath        string
}

type ModelConfig struct {
	ModelName       string            `json:"modelName" validate:"required"`
	ApiType         string            `json:"apiType" validate:"required"`
	ApiURL          string            `json:"apiUrl" validate:"required,url"`
	ApiKey          string            `json:"apiKey"`
	ConcurrentSlots int               `json:"concurrentSlots" validate:"number"`
	CapacityPolicy  string            `json:"capacityPolicy"`
	Parameters      map[string]string `json:"parameters"`
}

type ModelConfigs map[string]ModelConfig
type ModelConfigsV2 struct {
	Models []struct {
		ID string `json:"modelId"`
		ModelConfig
	} `json:"models"`
}

func NewModelConfigLoader(configPath string, validator Validator, blockchainChecker BlockchainChecker, connectionChecker ConnectionChecker, log lib.ILogger) *ModelConfigLoader {
	return &ModelConfigLoader{
		log:               log.Named("MODEL_LOADER"),
		modelConfigs:      ModelConfigs{},
		validator:         validator,
		blockchainChecker: blockchainChecker,
		connectionChecker: connectionChecker,
		configPath:        configPath,
	}
}

func (e *ModelConfigLoader) Init() error {
	filePath := ConfigPathDefault
	if e.configPath != "" {
		filePath = e.configPath
	}

	modelsConfig, err := lib.ReadJSONFile(filePath)
	if err != nil {
		e.log.Errorf("failed to read models config file: %s", err)

		// TODO: load models config from persistent storage
		// e.log.Warn("trying to load models config from persistent storage")

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
			e.modelConfigs[v.ID] = v.ModelConfig
			_ = e.Validate(context.Background(), common.HexToHash(v.ID), v.ModelConfig)
		}
		return nil
	}

	e.log.Warnf("failed to unmarshal to new models config, trying legacy")

	// try old config format
	var modelConfigs ModelConfigs
	err = json.Unmarshal([]byte(modelsConfig), &modelConfigs)
	if err != nil {
		return fmt.Errorf("invalid models config: %w", err)
	}

	err = e.validator.Struct(modelConfigs)
	if err != nil {
		return fmt.Errorf("invalid models config: %w", err)
	}

	e.modelConfigs = modelConfigs
	return nil
}

func (e *ModelConfigLoader) ModelConfigFromID(ID string) *ModelConfig {
	if ID == "" {
		return &ModelConfig{}
	}

	modelConfig := e.modelConfigs[ID]
	if modelConfig.ModelName == "" {
		e.log.Warnf("model config not found for ID: %s", ID)
		return &ModelConfig{}
	}

	return &modelConfig
}

func (e *ModelConfigLoader) GetAll() ([]common.Hash, []ModelConfig) {
	var modelConfigs []ModelConfig
	var modelIDs []common.Hash
	for ID, v := range e.modelConfigs {
		modelConfigs = append(modelConfigs, v)
		modelIDs = append(modelIDs, common.HexToHash(ID))
	}

	return modelIDs, modelConfigs
}

func (e *ModelConfigLoader) Validate(ctx context.Context, modelID common.Hash, cfg ModelConfig) error {
	// check if model exists
	exists, err := e.blockchainChecker.ModelExists(ctx, modelID)
	if err != nil {
		err = lib.WrapError(ErrValidate, err)
	} else if !exists {
		err = ErrModelNotFound
	}

	if err != nil {
		e.log.Warnf(e.formatLogPrefix(modelID, cfg)+"%s", err)
	}

	// try to connect to the model
	err = e.connectionChecker.TryConnect(ctx, cfg.ApiURL)
	if err != nil {
		err = lib.WrapError(ErrConnect, err)
		e.log.Warnf(e.formatLogPrefix(modelID, cfg)+"%s", err)
	}

	if exists && err == nil {
		e.log.Infof(e.formatLogPrefix(modelID, cfg) + "loaded and validated")
	}

	return nil
}

func (e *ModelConfigLoader) formatLogPrefix(modelID common.Hash, config ModelConfig) string {
	return fmt.Sprintf("modelID %s, name %s: ",
		lib.Short(modelID), config.ModelName)
}
