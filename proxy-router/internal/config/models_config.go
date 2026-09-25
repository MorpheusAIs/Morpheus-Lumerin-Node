package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"

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
	mu                sync.RWMutex
	modelConfigs      ModelConfigs
	validator         Validator
	blockchainChecker BlockchainChecker
	connectionChecker ConnectionChecker
	configPath        string
	configContent     string
}

type ModelConfig struct {
	ModelName       string            `json:"modelName" validate:"required"`
	ApiType         string            `json:"apiType" validate:"required"`
	ApiStack        string            `json:"apiStack"`
	ModelFamily     string            `json:"modelFamily"`
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

func NewModelConfigLoader(configPath string, configContent string, validator Validator, blockchainChecker BlockchainChecker, connectionChecker ConnectionChecker, log lib.ILogger) *ModelConfigLoader {
	return &ModelConfigLoader{
		log:               log.Named("MODEL_LOADER"),
		modelConfigs:      ModelConfigs{},
		validator:         validator,
		blockchainChecker: blockchainChecker,
		connectionChecker: connectionChecker,
		configPath:        configPath,
		configContent:     configContent,
	}
}

// Init loads the models config at startup. Reload re-reads the same file
// later without a restart; both go through load.
func (e *ModelConfigLoader) Init() error {
	cfgs, err := e.load()
	if err != nil {
		return err
	}
	e.mu.Lock()
	e.modelConfigs = cfgs
	e.mu.Unlock()
	return nil
}

// Reload re-reads the models config file and swaps the in-memory table for
// the new one atomically. Readers that already fetched a config for an
// in-flight request keep the copy they have. Returns the model IDs that were
// added and removed relative to the previous table, sorted. On any parse or
// validation error the previous table stays in force and the error is
// returned, so a half-edited file can never empty a serving node.
//
// A new model becomes servable as soon as this returns; the health checker
// picks it up on its next sweep (or on a triggered one), and, because the
// checker only probes models that carry a bid, a model added here with no bid
// costs nothing until one is posted.
func (e *ModelConfigLoader) Reload() (added []string, removed []string, err error) {
	cfgs, err := e.load()
	if err != nil {
		return nil, nil, err
	}
	e.mu.Lock()
	prev := e.modelConfigs
	e.modelConfigs = cfgs
	e.mu.Unlock()

	added, removed = []string{}, []string{}
	for id := range cfgs {
		if _, ok := prev[id]; !ok {
			added = append(added, id)
		}
	}
	for id := range prev {
		if _, ok := cfgs[id]; !ok {
			removed = append(removed, id)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	e.log.Infof("models config reloaded: %d models, %d added, %d removed", len(cfgs), len(added), len(removed))
	return added, removed, nil
}

// load parses the models config file into a fresh table without touching the
// one in use.
func (e *ModelConfigLoader) load() (ModelConfigs, error) {
	filePath := ConfigPathDefault
	if e.configPath != "" {
		filePath = e.configPath
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if e.configContent != "" {
			err = os.WriteFile(filePath, []byte(e.configContent), 0644)
			if err != nil {
				return nil, fmt.Errorf("failed to write models config content to file: %s", err)
			}
		} else {
			return nil, fmt.Errorf("models config file not found: %s", filePath)
		}
	}

	modelsConfig, err := lib.ReadJSONFile(filePath)
	if err != nil {
		e.log.Errorf("failed to read models config file: %s", err)

		// TODO: load models config from persistent storage
		// e.log.Warn("trying to load models config from persistent storage")

		return nil, err
	}
	e.log.Infof("models config loaded from file: %s", filePath)

	// check config format
	var cfgMap map[string]json.RawMessage
	err = json.Unmarshal([]byte(modelsConfig), &cfgMap)
	if err != nil {
		return nil, fmt.Errorf("invalid models config format: %s", err)
	}
	if cfgMap["models"] != nil {
		var modelConfigsV2 ModelConfigsV2
		err = json.Unmarshal([]byte(modelsConfig), &modelConfigsV2)
		if err != nil {
			return nil, fmt.Errorf("invalid models config V2 format: %s", err)
		}
		cfgs := make(ModelConfigs, len(modelConfigsV2.Models))
		for _, v := range modelConfigsV2.Models {
			v.ApiStack = e.loadApiStack(v.ID, v.ModelConfig)
			cfgs[v.ID] = v.ModelConfig
			_ = e.Validate(context.Background(), common.HexToHash(v.ID), v.ModelConfig)
		}
		return cfgs, nil
	}

	e.log.Warnf("failed to unmarshal to new models config, trying legacy")

	// try old config format
	var modelConfigs ModelConfigs
	err = json.Unmarshal([]byte(modelsConfig), &modelConfigs)
	if err != nil {
		return nil, fmt.Errorf("invalid models config: %w", err)
	}

	err = e.validator.Struct(modelConfigs)
	if err != nil {
		return nil, fmt.Errorf("invalid models config: %w", err)
	}
	for id, cfg := range modelConfigs {
		cfg.ApiStack = e.loadApiStack(id, cfg)
		modelConfigs[id] = cfg
	}

	return modelConfigs, nil
}

// An invalid apiStack degrades only that model: main.go merely warns on an
// Init error and would otherwise run with zero models.
func (e *ModelConfigLoader) loadApiStack(modelID string, cfg ModelConfig) string {
	if err := ValidateApiStack(modelID, cfg); err != nil {
		e.log.Errorf("%s — apiStack ignored for this model; its backend API will be detected at runtime", err)
		return ""
	}
	return NormalizeApiStack(cfg.ApiStack)
}

func (e *ModelConfigLoader) ModelConfigFromID(ID string) *ModelConfig {
	if ID == "" {
		return &ModelConfig{}
	}

	e.mu.RLock()
	modelConfig := e.modelConfigs[ID]
	e.mu.RUnlock()
	if modelConfig.ModelName == "" {
		e.log.Warnf("model config not found for ID: %s", ID)
		return &ModelConfig{}
	}

	return &modelConfig
}

func (e *ModelConfigLoader) GetAll() ([]common.Hash, []ModelConfig) {
	var modelConfigs []ModelConfig
	var modelIDs []common.Hash
	e.mu.RLock()
	defer e.mu.RUnlock()
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
