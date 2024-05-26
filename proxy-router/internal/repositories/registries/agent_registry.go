package registries

import (
	"context"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/contracts/agentregistry"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/lib"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type AgentRegistry struct {
	// config
	agentRegistryAddr common.Address

	// state
	nonce uint64
	mutex lib.Mutex
	mrABI *abi.ABI

	// deps
	agentRegistry *agentregistry.AgentRegistry
	client        *ethclient.Client
	log           interfaces.ILogger
}

func NewAgentRegistry(agentRegistryAddr common.Address, client *ethclient.Client, log interfaces.ILogger) *AgentRegistry {
	mr, err := agentregistry.NewAgentRegistry(agentRegistryAddr, client)
	if err != nil {
		panic("invalid model registry ABI")
	}
	mrABI, err := agentregistry.AgentRegistryMetaData.GetAbi()
	if err != nil {
		panic("invalid model registry ABI: " + err.Error())
	}
	return &AgentRegistry{
		agentRegistry:     mr,
		agentRegistryAddr: agentRegistryAddr,
		client:            client,
		mrABI:             mrABI,
		mutex:             lib.NewMutex(),
		log:               log,
	}
}

func (ar *AgentRegistry) GetAllAgents(ctx context.Context) ([]agentregistry.AgentRegistryAgent, error) {
	return ar.agentRegistry.GetAll(&bind.CallOpts{Context: ctx})
}
