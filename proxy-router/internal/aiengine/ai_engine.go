package aiengine

import (
	"context"
	"encoding/json"
	"errors"

	"fmt"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/ethereum/go-ethereum/common"
	mcpClient "github.com/mark3labs/mcp-go/client"
	mcp "github.com/mark3labs/mcp-go/mcp"
)

type AiEngine struct {
	modelsConfigLoader *config.ModelConfigLoader
	agentsConfigLoader *config.AgentConfigLoader
	service            ProxyService
	storage            gcs.ChatStorageInterface
	log                lib.ILogger
}

type AgentCallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Meta      *struct {
		ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
	} `json:"_meta,omitempty"`
}

var (
	ErrChatCompletion                = errors.New("chat completion error")
	ErrImageGenerationInvalidRequest = errors.New("invalid prodia image generation request")
	ErrImageGenerationRequest        = errors.New("image generation error")
	ErrJobCheckRequest               = errors.New("job status check error")
	ErrJobFailed                     = errors.New("job failed")
)

func NewAiEngine(service ProxyService, storage gcs.ChatStorageInterface, modelsConfigLoader *config.ModelConfigLoader, agentsConfigLoader *config.AgentConfigLoader, log lib.ILogger) *AiEngine {
	return &AiEngine{
		modelsConfigLoader: modelsConfigLoader,
		agentsConfigLoader: agentsConfigLoader,
		service:            service,
		storage:            storage,
		log:                log,
	}
}

func (a *AiEngine) GetAdapter(ctx context.Context, chatID, modelID, sessionID common.Hash, storeChatContext, forwardChatContext bool) (AIEngineStream, error) {
	var engine AIEngineStream
	if sessionID == (common.Hash{}) {
		// local model
		modelConfig := a.modelsConfigLoader.ModelConfigFromID(modelID.Hex())
		if modelConfig == nil {
			return nil, fmt.Errorf("model not found: %s", modelID.Hex())
		}
		var ok bool
		engine, ok = ApiAdapterFactory(modelConfig.ApiType, modelConfig.ModelName, modelConfig.ApiURL, modelConfig.ApiKey, modelConfig.Parameters, a.log)
		if !ok {
			return nil, fmt.Errorf("api adapter not found: %s", modelConfig.ApiType)
		}
	} else {
		// remote model
		engine = &RemoteModel{sessionID: sessionID, service: a.service}
	}

	if storeChatContext {
		var actualModelID common.Hash
		if modelID == (common.Hash{}) {
			modelID, err := a.service.GetModelIdSession(ctx, sessionID)
			if err != nil {
				return nil, err
			}
			actualModelID = modelID
		}
		engine = NewHistory(engine, a.storage, chatID, actualModelID, forwardChatContext, a.log)
	}

	return engine, nil
}

func (a *AiEngine) GetLocalModels() ([]LocalModel, error) {
	models := []LocalModel{}

	IDs, modelsFromConfig := a.modelsConfigLoader.GetAll()
	for i, model := range modelsFromConfig {
		models = append(models, LocalModel{
			Id:             IDs[i].Hex(),
			Name:           model.ModelName,
			Model:          model.ModelName,
			ApiType:        model.ApiType,
			ApiUrl:         model.ApiURL,
			Slots:          model.ConcurrentSlots,
			CapacityPolicy: model.CapacityPolicy,
		})
	}

	return models, nil
}

func (a *AiEngine) GetLocalAgents() ([]LocalAgent, error) {
	agents := []LocalAgent{}

	IDs, agentsFromConfig := a.agentsConfigLoader.GetAll()
	for i, agent := range agentsFromConfig {
		agents = append(agents, LocalAgent{
			Id:              IDs[i].Hex(),
			Name:            agent.AgentName,
			Command:         agent.Command,
			Args:            agent.Args,
			ConcurrentSlots: agent.ConcurrentSlots,
			CapacityPolicy:  agent.CapacityPolicy,
		})
	}

	return agents, nil
}

func (a *AiEngine) GetAgentTools(ctx context.Context, sessionID common.Hash, agentID common.Hash) ([]AgentTool, error) {
	tools := []AgentTool{}

	if sessionID != (common.Hash{}) {
		result, err := a.service.GetAgentTools(ctx, sessionID)
		if err != nil {
			return nil, err
		}

		var tools []AgentTool
		err = json.Unmarshal([]byte(result), &tools)
		if err != nil {
			return nil, err
		}
		return tools, nil
	}

	if agentID == (common.Hash{}) {
		return nil, fmt.Errorf("agent ID is required")
	}

	agentConfig := a.agentsConfigLoader.AgentConfigFromID(agentID.Hex())
	if agentConfig == nil {
		return nil, fmt.Errorf("agent not found: %s", agentID.Hex())
	}

	envs := []string{}
	for key, value := range agentConfig.Env {
		envs = append(envs, fmt.Sprintf("%s=%s", key, value))
	}

	mcpStdioClient, err := mcpClient.NewStdioMCPClient(agentConfig.Command, envs, agentConfig.Args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %s", err)
	}

	_, err = mcpStdioClient.Initialize(context.Background(), mcp.InitializeRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MCP client: %s", err)
	}

	toolsResponse, err := mcpStdioClient.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %s", err)
	}

	for _, tool := range toolsResponse.Tools {
		tools = append(tools, AgentTool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: ToolInputSchema{
				Type:       tool.InputSchema.Type,
				Properties: tool.InputSchema.Properties,
				Required:   tool.InputSchema.Required,
			},
		})
	}

	mcpStdioClient.Close()

	return tools, nil
}

func (a *AiEngine) CallAgentTool(ctx context.Context, sessionID common.Hash, agentID common.Hash, toolName string, input map[string]interface{}) (interface{}, error) {
	if sessionID != (common.Hash{}) {
		result, err := a.service.CallAgentTool(ctx, sessionID, toolName, input)
		if err != nil {
			return nil, err
		}

		fmt.Println("result", result)

		var toolResult *interface{}
		err = json.Unmarshal([]byte(result), &toolResult)
		if err != nil {
			return nil, err
		}
		return toolResult, nil
	}

	if agentID == (common.Hash{}) {
		return nil, fmt.Errorf("agent ID is required")
	}

	agentConfig := a.agentsConfigLoader.AgentConfigFromID(agentID.Hex())
	if agentConfig == nil {
		return nil, fmt.Errorf("agent not found: %s", agentID.Hex())
	}

	envs := []string{}
	for key, value := range agentConfig.Env {
		envs = append(envs, fmt.Sprintf("%s=%s", key, value))
	}

	mcpStdioClient, err := mcpClient.NewStdioMCPClient(agentConfig.Command, envs, agentConfig.Args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %s", err)
	}

	_, err = mcpStdioClient.Initialize(context.Background(), mcp.InitializeRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MCP client: %s", err)
	}

	callToolParams := AgentCallToolParams{
		Name:      toolName,
		Arguments: input,
		Meta:      nil,
	}
	callToolResponse, err := mcpStdioClient.CallTool(context.Background(), mcp.CallToolRequest{
		Params: callToolParams,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to call tool: %s", err)
	}

	mcpStdioClient.Close()

	return callToolResponse, nil
}
