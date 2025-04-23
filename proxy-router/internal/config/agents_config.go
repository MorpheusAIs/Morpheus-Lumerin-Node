package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
)

const (
	ConfigAgentsPathDefault = "agents-config.json"
)

var (
	ErrAgentNotFound = errors.New("agent not found in blockchain, local-only")
)

type AgentConfigLoader struct {
	log               lib.ILogger
	agentConfigs      AgentConfigs
	validator         Validator
	blockchainChecker BlockchainChecker
	configPath        string
	configContent     string
}

type AgentConfig struct {
	AgentID         string            `json:"agentId" validate:"required"`
	AgentName       string            `json:"agentName" validate:"required"`
	Command         string            `json:"command" validate:"required"`
	Args            []string          `json:"args" validate:"required"`
	Env             map[string]string `json:"env"`
	ConcurrentSlots int               `json:"concurrentSlots" validate:"number"`
	CapacityPolicy  string            `json:"capacityPolicy"`
}

type AgentConfigs map[string]AgentConfig

type AgentsConfigFile struct {
	Agents []AgentConfig `json:"agents" validate:"required"`
}

func NewAgentConfigLoader(configPath string, configContent string, validator Validator, blockchainChecker BlockchainChecker, log lib.ILogger) *AgentConfigLoader {
	return &AgentConfigLoader{
		log:               log.Named("AGENT_LOADER"),
		agentConfigs:      AgentConfigs{},
		validator:         validator,
		blockchainChecker: blockchainChecker,
		configPath:        configPath,
		configContent:     configContent,
	}
}

func (e *AgentConfigLoader) Init() error {
	filePath := ConfigAgentsPathDefault
	if e.configPath != "" {
		filePath = e.configPath
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if e.configContent != "" {
			err = os.WriteFile(filePath, []byte(e.configContent), 0644)
			if err != nil {
				return fmt.Errorf("failed to write models config content to file: %s", err)
			}
		} else {
			return fmt.Errorf("agents config file not found: %s", filePath)
		}
	}

	agentsConfig, err := lib.ReadJSONFile(filePath)
	if err != nil {
		e.log.Errorf("failed to read agents config file: %s", err)

		// TODO: load models config from persistent storage
		// e.log.Warn("trying to load models config from persistent storage")

		return err
	}
	e.log.Infof("agents config loaded from file: %s", filePath)

	// check config format
	var cfgMap map[string]json.RawMessage
	err = json.Unmarshal([]byte(agentsConfig), &cfgMap)
	if err != nil {
		return fmt.Errorf("invalid agents config format: %s", err)
	}
	if cfgMap["agents"] != nil {
		var agentsConfigV2 AgentsConfigFile
		err = json.Unmarshal([]byte(agentsConfig), &agentsConfigV2)
		if err != nil {
			return fmt.Errorf("invalid agents config V2 format: %s", err)
		}
		for _, v := range agentsConfigV2.Agents {
			e.agentConfigs[v.AgentID] = v
			_ = e.Validate(context.Background(), common.HexToHash(v.AgentID), v)
		}
		return nil
	}

	return fmt.Errorf("invalid agents config format: %s", err)
}

func (e *AgentConfigLoader) AgentConfigFromID(ID string) *AgentConfig {
	if ID == "" {
		return &AgentConfig{}
	}

	agentConfig := e.agentConfigs[ID]
	if agentConfig.AgentName == "" {
		e.log.Warnf("agent config not found for ID: %s", ID)
		return &AgentConfig{}
	}

	return &agentConfig
}

func (e *AgentConfigLoader) GetAll() ([]common.Hash, []AgentConfig) {
	var agentConfigs []AgentConfig
	var agentIDs []common.Hash
	for ID, v := range e.agentConfigs {
		agentConfigs = append(agentConfigs, v)
		agentIDs = append(agentIDs, common.HexToHash(ID))
	}

	return agentIDs, agentConfigs
}

func (e *AgentConfigLoader) Validate(ctx context.Context, agentID common.Hash, cfg AgentConfig) error {
	// check if agent exists
	exists, err := e.blockchainChecker.ModelExists(ctx, agentID)
	if err != nil {
		err = lib.WrapError(ErrValidate, err)
	} else if !exists {
		err = ErrAgentNotFound
	}

	if err != nil {
		e.log.Warnf(e.formatLogPrefix(agentID, cfg)+"%s", err)
	}

	if exists && err == nil {
		e.log.Infof(e.formatLogPrefix(agentID, cfg) + "loaded and validated")
	}

	return nil
}

func (e *AgentConfigLoader) formatLogPrefix(agentID common.Hash, config AgentConfig) string {
	return fmt.Sprintf("agentID %s, name %s: ",
		lib.Short(agentID), config.AgentName)
}
