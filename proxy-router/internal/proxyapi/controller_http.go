package proxyapi

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	constants "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	gsc "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

const (
	// Memory optimization constants for audio processing
	BASE64_ENCODING_CHUNK_SIZE  = 1024 * 1024 // 1MB chunks for base64 encoding
	BASE64_DECODING_CHUNK_SIZE  = 64 * 1024   // 64KB chunks for base64 decoding
	CONTENT_TYPE_DETECTION_SIZE = 512         // 512 bytes for content type detection
)

type AIEngine interface {
	GetLocalModels() ([]aiengine.LocalModel, error)
	GetLocalAgents() ([]aiengine.LocalAgent, error)
	CallAgentTool(ctx context.Context, sessionID, agentID common.Hash, toolName string, input map[string]interface{}) (interface{}, error)
	GetAgentTools(ctx context.Context, sessionID, agentID common.Hash) ([]aiengine.AgentTool, error)
	GetAdapter(ctx context.Context, chatID, modelID, sessionID common.Hash, storeContext, forwardContext bool) (aiengine.AIEngineStream, error)
}

type ProxyController struct {
	service            *ProxyServiceSender
	aiEngine           AIEngine
	chatStorage        gsc.ChatStorageInterface
	storeChatContext   bool
	forwardChatContext bool
	log                lib.ILogger
	authConfig         system.HTTPAuthConfig
	ipfsManager        *IpfsManager
	dockerManager      *DockerManager
}

func NewProxyController(service *ProxyServiceSender, aiEngine AIEngine, chatStorage gsc.ChatStorageInterface, storeChatContext, forwardChatContext bool, authConfig system.HTTPAuthConfig, ipfsManager *IpfsManager, log lib.ILogger) *ProxyController {
	dockerManager := NewDockerManager(log)

	c := &ProxyController{
		service:            service,
		aiEngine:           aiEngine,
		chatStorage:        chatStorage,
		storeChatContext:   storeChatContext,
		forwardChatContext: forwardChatContext,
		log:                log,
		authConfig:         authConfig,
		ipfsManager:        ipfsManager,
		dockerManager:      dockerManager,
	}

	return c
}

func (s *ProxyController) RegisterRoutes(r interfaces.Router) {
	r.POST("/proxy/provider/ping", s.Ping)
	r.POST("/proxy/sessions/initiate", s.authConfig.CheckAuth("initiate_session"), s.InitiateSession)
	r.POST("/v1/chat/completions", s.authConfig.CheckAuth("chat"), s.Prompt)
	r.GET("/v1/models", s.authConfig.CheckAuth("get_local_models"), s.Models)
	r.GET("/v1/agents", s.authConfig.CheckAuth("get_local_agents"), s.Agents)
	r.GET("/v1/agents/tools", s.authConfig.CheckAuth("get_agent_tools"), s.GetAgentTools)
	r.POST("/v1/agents/tools", s.authConfig.CheckAuth("call_agent_tool"), s.CallAgentTool)
	r.GET("/v1/chats", s.authConfig.CheckAuth("get_chat_history"), s.GetChats)
	r.GET("/v1/chats/:id", s.authConfig.CheckAuth("get_chat_history"), s.GetChat)
	r.DELETE("/v1/chats/:id", s.authConfig.CheckAuth("edit_chat_history"), s.DeleteChat)
	r.POST("/v1/chats/:id", s.authConfig.CheckAuth("edit_chat_history"), s.UpdateChatTitle)
	r.POST("/v1/audio/transcriptions", s.authConfig.CheckAuth("audio_transcription"), s.AudioTranscription)
	r.POST("/v1/audio/speech", s.authConfig.CheckAuth("audio_speech"), s.AudioSpeech)

	r.POST("/ipfs/pin", s.authConfig.CheckAuth("ipfs_pin"), s.Pin)
	r.POST("/ipfs/unpin", s.authConfig.CheckAuth("ipfs_unpin"), s.Unpin)
	r.GET("/ipfs/download/:cidHash", s.authConfig.CheckAuth("ipfs_get"), s.DownloadFile)
	r.GET("/ipfs/download/stream/:cidHash", s.authConfig.CheckAuth("ipfs_get"), s.StreamDownloadFile)
	r.POST("/ipfs/add", s.authConfig.CheckAuth("ipfs_add"), s.AddFile)
	r.GET("/ipfs/version", s.authConfig.CheckAuth("ipfs_version"), s.GetIpfsVersion)
	r.GET("/ipfs/pin", s.authConfig.CheckAuth("ipfs_pinned"), s.GetPinnedFiles)

	// Docker management routes
	r.POST("/docker/build", s.authConfig.CheckAuth("docker_build"), s.BuildDockerImage)
	r.POST("/docker/build/stream", s.authConfig.CheckAuth("docker_build"), s.StreamBuildDockerImage)
	r.POST("/docker/container/start", s.authConfig.CheckAuth("docker_manage"), s.StartContainer)
	r.POST("/docker/container/stop", s.authConfig.CheckAuth("docker_manage"), s.StopContainer)
	r.POST("/docker/container/remove", s.authConfig.CheckAuth("docker_manage"), s.RemoveContainer)
	r.GET("/docker/container/:id", s.authConfig.CheckAuth("docker_manage"), s.GetContainer)
	r.POST("/docker/containers", s.authConfig.CheckAuth("docker_manage"), s.ListContainers)
	r.GET("/docker/container/:id/logs", s.authConfig.CheckAuth("docker_manage"), s.GetContainerLogs)
	r.GET("/docker/container/:id/logs/stream", s.authConfig.CheckAuth("docker_manage"), s.StreamContainerLogs)
	r.GET("/docker/version", s.authConfig.CheckAuth("docker_version"), s.GetDockerVersion)
	r.POST("/docker/prune/images", s.authConfig.CheckAuth("docker_manage"), s.PruneImages)
	r.POST("/docker/prune/containers", s.authConfig.CheckAuth("docker_manage"), s.PruneContainers)
}

// Ping godoc
//
//	@Summary		Ping Provider
//	@Description	sends a ping to the provider on the RPC level
//	@Tags			chat
//	@Produce		json
//	@Param			pingReq	body		proxyapi.PingReq	true	"Ping Request"
//	@Success		200		{object}	proxyapi.PingRes
//	@Router			/proxy/provider/ping [post]
func (s *ProxyController) Ping(ctx *gin.Context) {
	var req *PingReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ping, err := s.service.Ping(ctx, req.ProviderURL, req.ProviderAddr)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, &PingRes{PingMs: ping.Milliseconds()})
}

// InitiateSession godoc
//
//	@Summary		Initiate Session with Provider
//	@Description	sends a handshake to the provider
//	@Tags			chat
//	@Produce		json
//	@Param			initiateSession	body		proxyapi.InitiateSessionReq	true	"Initiate Session"
//	@Success		200				{object}	morrpcmesssage.SessionRes
//	@Security		BasicAuth
//
//	@Router			/proxy/sessions/initiate [post]
func (s *ProxyController) InitiateSession(ctx *gin.Context) {
	var req *InitiateSessionReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := s.service.InitiateSession(ctx, req.User, req.Provider, req.Spend.Unpack(), req.BidID, req.ProviderUrl)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// SendPrompt godoc
//
//	@Summary		Send Local Or Remote Prompt
//	@Description	Send prompt to a local or remote model based on session id in header
//	@Tags			chat
//	@Produce		text/event-stream
//	@Param			session_id	header		string											false	"Session ID"	format(hex32)
//	@Param			model_id	header		string											false	"Model ID"		format(hex32)
//	@Param			chat_id		header		string											false	"Chat ID"		format(hex32)
//	@Param			prompt		body		proxyapi.ChatCompletionRequestSwaggerExample	true	"Prompt"
//	@Success		200			{object}	string
//	@Security		BasicAuth
//
//	@Router			/v1/chat/completions [post]
func (c *ProxyController) Prompt(ctx *gin.Context) {
	var (
		body openai.ChatCompletionRequest
		head PromptHead
	)

	if err := ctx.ShouldBindHeader(&head); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chatID := head.ChatID
	if chatID == (lib.Hash{}) {
		var err error
		chatID, err = lib.GetRandomHash()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	adapter, err := c.aiEngine.GetAdapter(ctx, chatID.Hash, head.ModelID.Hash, head.SessionID.Hash, c.storeChatContext, c.forwardChatContext)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var contentType string
	if body.Stream {
		contentType = constants.CONTENT_TYPE_EVENT_STREAM
	} else {
		contentType = constants.CONTENT_TYPE_JSON
	}

	ctx.Writer.Header().Set(constants.HEADER_CONTENT_TYPE, contentType)

	err = adapter.Prompt(ctx, &body, func(cbctx context.Context, completion gsc.Chunk, aiResponseError *gsc.AiEngineErrorResponse) error {
		if aiResponseError != nil {
			ctx.Writer.Header().Set(constants.HEADER_CONTENT_TYPE, constants.CONTENT_TYPE_JSON)
			ctx.JSON(http.StatusBadRequest, aiResponseError)
			return nil
		}

		marshalledResponse, err := json.Marshal(completion.Data())
		if err != nil {
			return err
		}

		if body.Stream {
			_, err = ctx.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", marshalledResponse)))
		} else {
			_, err = ctx.Writer.Write(marshalledResponse)
		}

		if err != nil {
			return err
		}

		ctx.Writer.Flush()
		return nil
	})

	if err != nil {
		c.log.Errorf("error sending prompt: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

// GetLocalModels godoc
//
//	@Summary	Get local models
//	@Tags		system
//	@Produce	json
//	@Success	200	{object}	[]aiengine.LocalModel
//	@Security	BasicAuth
//	@Router		/v1/models [get]
func (c *ProxyController) Models(ctx *gin.Context) {
	models, err := c.aiEngine.GetLocalModels()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, models)
}

// GetLocalAgents godoc
//
//	@Summary	Get local agents
//	@Tags		agents
//	@Produce	json
//	@Success	200	{object}	[]aiengine.LocalAgent
//	@Security	BasicAuth
//	@Router		/v1/agents [get]
func (c *ProxyController) Agents(ctx *gin.Context) {
	agents, err := c.aiEngine.GetLocalAgents()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, agents)
}

// GetAgentTools godoc
//
//	@Summary	Get agent tools
//	@Tags		agents
//	@Produce	json
//	@Param		session_id	header		string	false	"Session ID"	format(hex32)
//	@Param		agent_id	header		string	false	"Agent ID"		format(hex32)
//	@Success	200			{object}	[]aiengine.AgentTool
//	@Security	BasicAuth
//	@Router		/v1/agents/tools [get]
func (c *ProxyController) GetAgentTools(ctx *gin.Context) {
	var head AgentPromptHead
	err := ctx.ShouldBindHeader(&head)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	tools, err := c.aiEngine.GetAgentTools(ctx, head.SessionID.Hash, head.AgentId.Hash)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, tools)
	return
}

// GetChats godoc
//
//	@Summary	Get all chats stored in the system
//	@Tags		chat
//	@Produce	json
//	@Success	200	{object}	[]genericchatstorage.Chat
//	@Security	BasicAuth
//	@Router		/v1/chats [get]
func (c *ProxyController) GetChats(ctx *gin.Context) {
	chats := c.chatStorage.GetChats()

	if chats == nil {
		ctx.JSON(http.StatusOK, make([]struct{}, 0))
		return
	}

	ctx.JSON(http.StatusOK, chats)
}

// CallAgentTool godoc
//
//	@Summary	Call agent tool
//	@Tags		agents
//	@Produce	json
//	@Param		session_id	header		string						false	"Session ID"	format(hex32)
//	@Param		agent_id	header		string						false	"Agent ID"		format(hex32)
//	@Param		input		body		proxyapi.CallAgentToolReq	true	"Input"
//	@Success	200			{object}	interface{}
//	@Security	BasicAuth
//	@Router		/v1/agents/tools [post]
func (c *ProxyController) CallAgentTool(ctx *gin.Context) {
	var head AgentPromptHead
	err := ctx.ShouldBindHeader(&head)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	var req CallAgentToolReq
	err = ctx.ShouldBindJSON(&req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	result, err := c.aiEngine.CallAgentTool(ctx, head.SessionID.Hash, head.AgentId.Hash, req.ToolName, req.Input)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// GetChat godoc
//
//	@Summary	Get chat by id
//	@Tags		chat
//	@Produce	json
//	@Param		id	path		string	true	"Chat ID"
//	@Success	200	{object}	genericchatstorage.ChatHistory
//	@Security	BasicAuth
//	@Router		/v1/chats/{id} [get]
func (c *ProxyController) GetChat(ctx *gin.Context) {
	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	chat, err := c.chatStorage.LoadChatFromFile(params.ID.Hex())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, chat)
}

// DeleteChat godoc
//
//	@Summary	Delete chat by id from storage
//	@Tags		chat
//	@Produce	json
//	@Param		id	path		string	true	"Chat ID"
//	@Success	200	{object}	proxyapi.ResultResponse
//	@Security	BasicAuth
//	@Router		/v1/chats/{id} [delete]
func (c *ProxyController) DeleteChat(ctx *gin.Context) {
	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	err = c.chatStorage.DeleteChat(params.ID.Hex())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// UpdateChatTitle godoc
//
//	@Summary	Update chat title by id
//	@Tags		chat
//	@Produce	json
//	@Param		id		path		string						true	"Chat ID"
//	@Param		title	body		proxyapi.UpdateChatTitleReq	true	"Chat Title"
//	@Success	200		{object}	proxyapi.ResultResponse
//	@Security	BasicAuth
//	@Router		/v1/chats/{id} [post]
func (c *ProxyController) UpdateChatTitle(ctx *gin.Context) {
	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	var req UpdateChatTitleReq
	err = ctx.ShouldBindJSON(&req)
	if err != nil {
		c.service.log.Errorf("error binding json: %s", err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	err = c.chatStorage.UpdateChatTitle(params.ID.Hex(), req.Title)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// Pin godoc
//
//	@Summary	Pin a file to IPFS
//	@Tags		ipfs
//	@Produce	json
//	@Param		cidHash	body		proxyapi.CIDReq	true	"cidHash"
//	@Success	200		{object}	proxyapi.ResultResponse
//	@Security	BasicAuth
//	@Router		/ipfs/pin [post]
func (c *ProxyController) Pin(ctx *gin.Context) {
	var req CIDReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	CID, err := lib.ManualBytes32ToCID(req.CID.Bytes())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err = c.ipfsManager.Pin(ctx, CID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// Unpin godoc
//
//	@Summary	Unpin a file from IPFS
//	@Tags		ipfs
//	@Produce	json
//	@Param		cidHash	body		proxyapi.CIDReq	true	"cidHash"
//	@Success	200		{object}	proxyapi.ResultResponse
//	@Security	BasicAuth
//	@Router		/ipfs/unpin [post]
func (c *ProxyController) Unpin(ctx *gin.Context) {
	var req CIDReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	CID, err := lib.ManualBytes32ToCID(req.CID.Bytes())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = c.ipfsManager.Unpin(ctx, CID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// DownloadFile godoc
//
//	@Summary	Download a file from IPFS
//	@Tags		ipfs
//	@Produce	json
//	@Param		cidHash	path		string	true	"cidHash"
//	@Param		dest	query		string	true	"Destination Path"
//	@Success	200		{object}	proxyapi.ResultResponse
//	@Security	BasicAuth
//	@Router		/ipfs/download/{cidHash} [get]
func (c *ProxyController) DownloadFile(ctx *gin.Context) {
	var params struct {
		CID lib.Hash `uri:"cidHash" binding:"required"`
	}

	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	destinationPath := ctx.Query("dest")
	if destinationPath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "destination path is required"})
		return
	}

	CID, err := lib.ManualBytes32ToCID(params.CID.Bytes())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = c.ipfsManager.GetFile(ctx, CID, destinationPath)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// StreamDownloadFile godoc
//
//	@Summary	Download a file from IPFS with progress updates as SSE stream
//	@Tags		ipfs
//	@Produce	text/event-stream
//	@Param		cidHash	path		string	true	"cidHash"
//	@Param		dest	query		string	true	"Destination Path"
//	@Success	200		{object}	proxyapi.DownloadProgressEvent
//	@Security	BasicAuth
//	@Router		/ipfs/download/stream/{cidHash} [get]
func (c *ProxyController) StreamDownloadFile(ctx *gin.Context) {
	var params struct {
		CID lib.Hash `uri:"cidHash" binding:"required"`
	}

	if err := ctx.ShouldBindUri(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	destinationPath := ctx.Query("dest")
	if destinationPath == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "destination path is required"})
		return
	}

	CID, err := lib.ManualBytes32ToCID(params.CID.Bytes())
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if ctx.Request.Method == "OPTIONS" {
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		ctx.Writer.Header().Set("Access-Control-Max-Age", "86400")
		ctx.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	if ctx.Request.Method == "HEAD" {
		ctx.Writer.WriteHeader(http.StatusOK)
		return
	}

	// Set headers for SSE
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no") // For Nginx
	ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	// Create a cancelable context that will be used to stop the download when client disconnects
	downloadCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup client disconnection detection
	// Use a done channel to signal when the client disconnects
	clientGone := ctx.Request.Context().Done()

	// Monitor for client disconnection in a separate goroutine
	go func() {
		select {
		case <-clientGone:
			c.log.Info("Client disconnected, canceling download operation")
			cancel() // Cancel the download context when client disconnects
		case <-downloadCtx.Done():
			// Context canceled elsewhere, just return
		}
	}()

	progressCallback := func(downloaded, total int64) error {
		select {
		case <-downloadCtx.Done():
			return fmt.Errorf("download canceled: client disconnected or context canceled")
		default:
		}

		var percentage float64
		if total > 0 {
			percentage = float64(downloaded) / float64(total) * 100
		}

		event := DownloadProgressEvent{
			Status:      "downloading",
			Downloaded:  downloaded,
			Total:       total,
			Percentage:  percentage,
			TimeUpdated: time.Now().UnixMilli(),
		}

		data, err := json.Marshal(event)
		if err != nil {
			return err
		}

		// Check context again before writing to avoid writing to closed connection
		select {
		case <-downloadCtx.Done():
			return fmt.Errorf("download canceled: client disconnected or context canceled")
		default:
			if _, err = fmt.Fprintf(ctx.Writer, "data: %s\n\n", data); err != nil {
				cancel() // Cancel if we can't write to the client
				return err
			}

			ctx.Writer.Flush()

			// Don't spam too many updates
			if downloaded < total && downloaded%1048576 != 0 { // Send at least every 1MB
				// Skip some updates for better performance
				return nil
			}

			return nil
		}
	}

	// Start the download with progress tracking
	err = c.ipfsManager.GetFileWithProgress(downloadCtx, CID, destinationPath, progressCallback)
	if err != nil {
		// Check if this is a cancellation error, which is expected when client disconnects
		if downloadCtx.Err() != nil {
			c.log.Info("Download canceled: %v", err)
			return // Just return without sending an error event since client is gone
		}

		event := DownloadProgressEvent{
			Status:      "error",
			Error:       err.Error(),
			TimeUpdated: time.Now().UnixMilli(),
		}
		data, _ := json.Marshal(event)

		select {
		case <-downloadCtx.Done():
			return // Client is gone, no need to write
		default:
			fmt.Fprintf(ctx.Writer, "data: %s\n\n", data)
			ctx.Writer.Flush()
		}
		return
	}

	// Only send completion event if the client is still connected
	select {
	case <-downloadCtx.Done():
		return // Client is gone, no need to send completion
	default:
		event := DownloadProgressEvent{
			Status:      "completed",
			Downloaded:  100,
			Total:       100,
			Percentage:  100,
			TimeUpdated: time.Now().UnixMilli(),
		}
		data, _ := json.Marshal(event)
		fmt.Fprintf(ctx.Writer, "data: %s\n\n", data)
		ctx.Writer.Flush()
	}
}

// AddFile godoc
//
//	@Summary	Add a file to IPFS with metadata
//	@Tags		ipfs
//	@Produce	json
//	@Param		request	body		proxyapi.AddFileReq	true	"File Path and Metadata"
//	@Success	200		{object}	proxyapi.AddIpfsFileRes
//	@Security	BasicAuth
//	@Router		/ipfs/add [post]
func (c *ProxyController) AddFile(ctx *gin.Context) {
	var req AddFileReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := c.ipfsManager.AddFile(ctx, req.FilePath, req.Tags, req.ID.Hex(), req.ModelName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	fileCIDHash, err := lib.CIDToBytes32(result.FileCID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fileHash := lib.HexString(fileCIDHash)

	metadataCIDHash, err := lib.CIDToBytes32(result.MetadataCID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	metadataHash := lib.HexString(metadataCIDHash)

	ctx.JSON(http.StatusOK, AddIpfsFileRes{
		FileCID:         result.FileCID,
		MetadataCID:     result.MetadataCID,
		FileCIDHash:     fileHash,
		MetadataCIDHash: metadataHash,
	})
}

// GetIpfsVersion godoc
//
//	@Summary	Get IPFS Version
//	@Tags		ipfs
//	@Produce	json
//	@Success	200	{object}	proxyapi.IpfsVersionRes
//	@Security	BasicAuth
//	@Router		/ipfs/version [get]
func (c *ProxyController) GetIpfsVersion(ctx *gin.Context) {
	version, err := c.ipfsManager.GetVersion(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, IpfsVersionRes{Version: version})
}

// GetPinnedFiles godoc
//
//	@Summary	Get all pinned files metadata
//	@Tags		ipfs
//	@Produce	json
//	@Success	200	{array}	proxyapi.PinnedFileRes
//	@Security	BasicAuth
//	@Router		/ipfs/pin [get]
func (c *ProxyController) GetPinnedFiles(ctx *gin.Context) {
	metadata, err := c.ipfsManager.GetPinnedFiles(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses := make([]PinnedFileRes, 0, len(metadata))
	for _, item := range metadata {
		fileCIDBytes, err := lib.CIDToBytes32(item.FileCID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		fileCIDHash := lib.HexString(fileCIDBytes)

		metadataCIDBytes, err := lib.CIDToBytes32(item.MetadataCID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		metadataCIDHash := lib.HexString(metadataCIDBytes)

		// Create response object
		response := PinnedFileRes{
			FileName:        item.FileName,
			FileSize:        item.FileSize,
			FileCID:         item.FileCID,
			FileCIDHash:     fileCIDHash,
			Tags:            item.Tags,
			ID:              item.ID,
			ModelName:       item.ModelName,
			MetadataCID:     item.MetadataCID,
			MetadataCIDHash: metadataCIDHash,
		}

		responses = append(responses, response)
	}

	ctx.JSON(http.StatusOK, responses)
}

// BuildDockerImage godoc
//
//	@Summary	Build a Docker image
//	@Tags		docker
//	@Produce	json
//	@Param		request	body		proxyapi.DockerBuildReq	true	"Docker build request"
//	@Success	200		{object}	proxyapi.DockerBuildRes
//	@Security	BasicAuth
//	@Router		/docker/build [post]
func (c *ProxyController) BuildDockerImage(ctx *gin.Context) {
	var req DockerBuildReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := c.dockerManager.BuildImage(ctx, req.ContextPath, req.Dockerfile, req.ImageName, req.ImageTag, req.BuildArgs, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, DockerBuildRes{ImageTag: tag})
}

// StreamBuildDockerImage godoc
//
//	@Summary	Build a Docker image with progress updates as SSE stream
//	@Tags		docker
//	@Produce	text/event-stream
//	@Param		request	body		proxyapi.DockerBuildReq	true	"Docker build request"
//	@Success	200		{object}	proxyapi.DockerStreamBuildEvent
//	@Security	BasicAuth
//	@Router		/docker/build/stream [post]
func (c *ProxyController) StreamBuildDockerImage(ctx *gin.Context) {
	var req DockerBuildReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if ctx.Request.Method == "OPTIONS" {
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		ctx.Writer.Header().Set("Access-Control-Max-Age", "86400")
		ctx.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	// Set headers for SSE
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("X-Accel-Buffering", "no") // For Nginx
	ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	// Create a cancelable context that will be used to stop the build when client disconnects
	buildCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup client disconnection detection
	clientGone := ctx.Request.Context().Done()

	// Monitor for client disconnection in a separate goroutine
	go func() {
		select {
		case <-clientGone:
			c.log.Info("Client disconnected, canceling build operation")
			cancel() // Cancel the build context when client disconnects
		case <-buildCtx.Done():
			// Context canceled elsewhere, just return
		}
	}()

	progressCallback := func(progress BuildProgress) error {
		select {
		case <-buildCtx.Done():
			return fmt.Errorf("build canceled: client disconnected or context canceled")
		default:
		}

		data, err := json.Marshal(progress)
		if err != nil {
			return err
		}

		// Check context again before writing to avoid writing to closed connection
		select {
		case <-buildCtx.Done():
			return fmt.Errorf("build canceled: client disconnected or context canceled")
		default:
			if _, err = fmt.Fprintf(ctx.Writer, "data: %s\n\n", data); err != nil {
				cancel() // Cancel if we can't write to the client
				return err
			}

			ctx.Writer.Flush()
			return nil
		}
	}

	tag, err := c.dockerManager.BuildImage(buildCtx, req.ContextPath, req.Dockerfile, req.ImageName, req.ImageTag, req.BuildArgs, progressCallback)
	if err != nil {
		// Check if this is a cancellation error, which is expected when client disconnects
		if buildCtx.Err() != nil {
			c.log.Info("Build canceled: %v", err)
			return // Just return without sending an error event since client is gone
		}

		event := BuildProgress{
			Status:       "error",
			Error:        err.Error(),
			ErrorDetails: err.Error(),
			TimeUpdated:  time.Now().UnixMilli(),
		}
		data, _ := json.Marshal(event)

		select {
		case <-buildCtx.Done():
			return // Client is gone, no need to write
		default:
			fmt.Fprintf(ctx.Writer, "data: %s\n\n", data)
			ctx.Writer.Flush()
		}
		return
	}

	// Only send completion event if the client is still connected
	select {
	case <-buildCtx.Done():
		return // Client is gone, no need to send completion
	default:
		event := BuildProgress{
			Status:      "completed",
			Stream:      fmt.Sprintf("Successfully built image: %s\n", tag),
			TimeUpdated: time.Now().UnixMilli(),
		}
		data, _ := json.Marshal(event)
		fmt.Fprintf(ctx.Writer, "data: %s\n\n", data)
		ctx.Writer.Flush()
	}
}

// StartContainer godoc
//
//	@Summary	Start a Docker container
//	@Tags		docker
//	@Produce	json
//	@Param		request	body		proxyapi.DockerStartContainerReq	true	"Docker start container request"
//	@Success	200		{object}	proxyapi.DockerStartContainerRes
//	@Security	BasicAuth
//	@Router		/docker/container/start [post]
func (c *ProxyController) StartContainer(ctx *gin.Context) {
	var req DockerStartContainerReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	containerID, err := c.dockerManager.StartContainer(ctx, req.ImageName, req.ContainerName, req.Env, req.Ports, req.Volumes, req.NetworkMode)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, DockerStartContainerRes{ContainerID: containerID})
}

// StopContainer godoc
//
//	@Summary	Stop a Docker container
//	@Tags		docker
//	@Produce	json
//	@Param		request	body		proxyapi.DockerContainerActionReq	true	"Docker container stop request"
//	@Success	200		{object}	proxyapi.ResultResponse
//	@Security	BasicAuth
//	@Router		/docker/container/stop [post]
func (c *ProxyController) StopContainer(ctx *gin.Context) {
	var req DockerContainerActionReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.dockerManager.StopContainer(ctx, req.ContainerID, req.Timeout)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// RemoveContainer godoc
//
//	@Summary	Remove a Docker container
//	@Tags		docker
//	@Produce	json
//	@Param		request	body		proxyapi.DockerContainerActionReq	true	"Docker container remove request"
//	@Success	200		{object}	proxyapi.ResultResponse
//	@Security	BasicAuth
//	@Router		/docker/container/remove [post]
func (c *ProxyController) RemoveContainer(ctx *gin.Context) {
	var req DockerContainerActionReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.dockerManager.RemoveContainer(ctx, req.ContainerID, req.Force)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// GetContainer godoc
//
//	@Summary	Get Docker container info
//	@Tags		docker
//	@Produce	json
//	@Param		id	path		string	true	"Container ID"
//	@Success	200	{object}	proxyapi.DockerContainerInfoRes
//	@Security	BasicAuth
//	@Router		/docker/container/{id} [get]
func (c *ProxyController) GetContainer(ctx *gin.Context) {
	containerID := ctx.Param("id")
	if containerID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "container ID is required"})
		return
	}

	info, err := c.dockerManager.GetContainerInfo(ctx, containerID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, DockerContainerInfoRes{ContainerInfo: *info})
}

// ListContainers godoc
//
//	@Summary	List Docker containers
//	@Tags		docker
//	@Produce	json
//	@Param		request	body		proxyapi.DockerListContainersReq	true	"Docker list containers request"
//	@Success	200		{object}	proxyapi.DockerListContainersRes
//	@Security	BasicAuth
//	@Router		/docker/containers [post]
func (c *ProxyController) ListContainers(ctx *gin.Context) {
	var req DockerListContainersReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	containers, err := c.dockerManager.ListContainers(ctx, req.All, req.FilterLabels)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, DockerListContainersRes{Containers: containers})
}

// GetContainerLogs godoc
//
//	@Summary	Get Docker container logs
//	@Tags		docker
//	@Produce	text/plain
//	@Param		id		path		string	true	"Container ID"
//	@Param		tail	query		integer	false	"Number of lines to show from the end"
//	@Param		follow	query		boolean	false	"Follow log output"
//	@Success	200		{string}	string	"Log output"
//	@Security	BasicAuth
//	@Router		/docker/container/{id}/logs [get]
func (c *ProxyController) GetContainerLogs(ctx *gin.Context) {
	containerID := ctx.Param("id")
	if containerID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "container ID is required"})
		return
	}

	tail, _ := ctx.GetQuery("tail")
	tailLines := 100 // default
	if tail != "" {
		var err error
		tailLines, err = strconv.Atoi(tail)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid tail parameter"})
			return
		}
	}

	follow := false
	followStr, _ := ctx.GetQuery("follow")
	if followStr == "true" {
		follow = true
	}

	// For non-streaming logs, set a reasonable timeout
	logsCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	reader, err := c.dockerManager.GetContainerLogs(logsCtx, containerID, tailLines, follow)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer reader.Close()

	// Set content type
	ctx.Header("Content-Type", "text/plain")

	// Copy logs to response
	_, err = io.Copy(ctx.Writer, reader)
	if err != nil {
		c.log.Errorf("Error copying container logs: %v", err)
	}
}

// StreamContainerLogs godoc
//
//	@Summary	Stream Docker container logs as SSE
//	@Tags		docker
//	@Produce	text/event-stream
//	@Param		id		path		string	true	"Container ID"
//	@Param		tail	query		integer	false	"Number of lines to show from the end"
//	@Success	200		{string}	string	"Log output events"
//	@Security	BasicAuth
//	@Router		/docker/container/{id}/logs/stream [get]
func (c *ProxyController) StreamContainerLogs(ctx *gin.Context) {
	containerID := ctx.Param("id")
	if containerID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "container ID is required"})
		return
	}

	tail, _ := ctx.GetQuery("tail")
	tailLines := 100 // default
	if tail != "" {
		var err error
		tailLines, err = strconv.Atoi(tail)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid tail parameter"})
			return
		}
	}

	if ctx.Request.Method == "OPTIONS" {
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		ctx.Writer.Header().Set("Access-Control-Max-Age", "86400")
		ctx.Writer.WriteHeader(http.StatusNoContent)
		return
	}

	// Set headers for SSE
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

	// Create a cancelable context that will be used to stop the logs stream when client disconnects
	logsCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Setup client disconnection detection
	clientGone := ctx.Request.Context().Done()

	// Monitor for client disconnection in a separate goroutine
	go func() {
		select {
		case <-clientGone:
			c.log.Info("Client disconnected, canceling logs stream")
			cancel() // Cancel the logs context when client disconnects
		case <-logsCtx.Done():
			// Context canceled elsewhere, just return
		}
	}()

	// Get container logs with follow mode
	reader, err := c.dockerManager.GetContainerLogs(logsCtx, containerID, tailLines, true)
	if err != nil {
		event := struct {
			Status string `json:"status"`
			Error  string `json:"error"`
		}{
			Status: "error",
			Error:  err.Error(),
		}
		data, _ := json.Marshal(event)
		fmt.Fprintf(ctx.Writer, "data: %s\n\n", data)
		ctx.Writer.Flush()
		return
	}
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		select {
		case <-logsCtx.Done():
			return // Stop if context is canceled
		default:
			line := scanner.Text()
			// Create a log event
			event := struct {
				Status    string `json:"status"`
				Line      string `json:"line"`
				Timestamp int64  `json:"timestamp"`
			}{
				Status:    "log",
				Line:      line,
				Timestamp: time.Now().UnixMilli(),
			}
			data, err := json.Marshal(event)
			if err != nil {
				c.log.Errorf("Error marshaling log event: %v", err)
				continue
			}

			// Send the event
			_, err = fmt.Fprintf(ctx.Writer, "data: %s\n\n", data)
			if err != nil {
				c.log.Errorf("Error writing log event: %v", err)
				return // Stop on write error
			}
			ctx.Writer.Flush()
		}
	}

	if err := scanner.Err(); err != nil {
		if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			c.log.Errorf("Error scanning container logs: %v", err)
		}
	}
}

// GetDockerVersion godoc
//
//	@Summary	Get Docker version
//	@Tags		docker
//	@Produce	json
//	@Success	200	{object}	proxyapi.DockerVersionRes
//	@Security	BasicAuth
//	@Router		/docker/version [get]
func (c *ProxyController) GetDockerVersion(ctx *gin.Context) {
	version, err := c.dockerManager.GetDockerVersion(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, DockerVersionRes{Version: version})
}

// PruneImages godoc
//
//	@Summary	Prune unused Docker images
//	@Tags		docker
//	@Produce	json
//	@Success	200	{object}	proxyapi.DockerPruneRes
//	@Security	BasicAuth
//	@Router		/docker/prune/images [post]
func (c *ProxyController) PruneImages(ctx *gin.Context) {
	spaceReclaimed, err := c.dockerManager.PruneImages(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, DockerPruneRes{SpaceReclaimed: spaceReclaimed})
}

// PruneContainers godoc
//
//	@Summary	Prune stopped Docker containers
//	@Tags		docker
//	@Produce	json
//	@Success	200	{object}	proxyapi.DockerPruneRes
//	@Security	BasicAuth
//	@Router		/docker/prune/containers [post]
func (c *ProxyController) PruneContainers(ctx *gin.Context) {
	spaceReclaimed, err := c.dockerManager.PruneContainers(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, DockerPruneRes{SpaceReclaimed: spaceReclaimed})
}

// AudioTranscription godoc
//
//	@Summary		Transcribe audio file
//	@Description	Transcribes audio files to text using AI models
//	@Tags			audio
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			session_id					header		string	false	"Session ID"	format(hex32)
//	@Param			model_id					header		string	false	"Model ID"		format(hex32)
//	@Param			file						formData	file	true	"Audio file to transcribe"
//	@Param			language					formData	string	false	"Language of the audio"
//	@Param			prompt						formData	string	false	"Optional prompt to guide transcription"
//	@Param			response_format				formData	string	false	"Response format: json, text, srt, verbose_json, vtt"
//	@Param			temperature					formData	number	false	"Temperature for sampling"
//	@Param			timestamp_granularities[]	formData	string	false	"Timestamp granularity: word or segment"
//	@Param			timestamp_granularity		formData	string	false	"Timestamp granularity: word or segment (default is segment)"
//	@Param			stream						formData	boolean	false	"Whether to stream the results or not"
//	@Security		BasicAuth
//	@Router			/v1/audio/transcriptions [post]
func (c *ProxyController) AudioTranscription(ctx *gin.Context) {
	// Parse request
	params, err := c.parseAudioTranscriptionParams(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process audio file
	tempFilePath, err := c.createTempFile(ctx)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "Failed to get file" {
			statusCode = http.StatusBadRequest
		}
		ctx.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}
	defer os.Remove(tempFilePath) // Clean up temp file

	// Get AI adapter
	chatID := params.head.ChatID
	if chatID == (lib.Hash{}) {
		var err error
		chatID, err = lib.GetRandomHash()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	adapter, err := c.aiEngine.GetAdapter(ctx, chatID.Hash, params.head.ModelID.Hash, params.head.SessionID.Hash, c.storeChatContext, c.forwardChatContext)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Prepare transcription request
	transcriptionRequest := &gsc.AudioTranscriptionRequest{
		FilePath:               tempFilePath,
		Language:               params.language,
		Prompt:                 params.prompt,
		Format:                 openai.AudioResponseFormat(params.responseFormat),
		Temperature:            params.temperature,
		TimestampGranularities: params.timestampGranularities,
		TimestampGranularity:   params.timestampGranularity,
		Stream:                 params.stream,
	}

	// Process transcription with callback
	if err := c.executeTranscription(ctx, adapter, transcriptionRequest, params.stream); err != nil {
		c.log.Errorf("error sending transcription request: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

type audioTranscriptionParams struct {
	head                   PromptHead
	language               string
	prompt                 string
	responseFormat         string
	temperature            float32
	timestampGranularities []openai.TranscriptionTimestampGranularity
	timestampGranularity   openai.TranscriptionTimestampGranularity
	stream                 bool
}

func (c *ProxyController) parseAudioTranscriptionParams(ctx *gin.Context) (*audioTranscriptionParams, error) {
	var head PromptHead
	if err := ctx.ShouldBindHeader(&head); err != nil {
		return nil, err
	}

	// Parse form parameters
	params := &audioTranscriptionParams{
		head:           head,
		language:       ctx.Request.FormValue("language"),
		prompt:         ctx.Request.FormValue("prompt"),
		responseFormat: ctx.Request.FormValue("response_format"),
	}

	// Parse temperature
	temperatureStr := ctx.Request.FormValue("temperature")
	if temperatureStr != "" {
		temp, err := strconv.ParseFloat(temperatureStr, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid temperature value: %v", err)
		}
		params.temperature = float32(temp)
	}

	// Parse stream flag
	streamStr := ctx.Request.FormValue("stream")
	if streamStr != "" {
		stream, err := strconv.ParseBool(streamStr)
		if err != nil {
			return nil, fmt.Errorf("invalid stream value: %v", err)
		}
		params.stream = stream
	}

	timestampGranularityStr := ctx.Request.FormValue("timestamp_granularity")
	fmt.Println("timestampGranularityStr:", timestampGranularityStr)
	// Parse timestamp granularity
	if timestampGranularityStr != "" {
		switch timestampGranularityStr {
		case "word":
			fmt.Println("Setting timestamp granularity to word")
			params.timestampGranularity = openai.TranscriptionTimestampGranularityWord
		case "segment", "":
			fmt.Println("Setting timestamp granularity to segment")
			params.timestampGranularity = openai.TranscriptionTimestampGranularitySegment
		default:
			return nil, fmt.Errorf("invalid timestamp granularity: %s", timestampGranularityStr)
		}
	}

	// Parse timestamp granularities
	if ctx.Request.MultipartForm != nil && ctx.Request.MultipartForm.Value != nil {
		timestampGranularitiesRaw := ctx.Request.MultipartForm.Value["timestamp_granularities[]"]
		if len(timestampGranularitiesRaw) > 0 {
			for _, granularity := range timestampGranularitiesRaw {
				switch granularity {
				case "word":
					params.timestampGranularities = append(params.timestampGranularities, openai.TranscriptionTimestampGranularityWord)
				case "segment", "":
					params.timestampGranularities = append(params.timestampGranularities, openai.TranscriptionTimestampGranularitySegment)
				default:
					params.timestampGranularities = append(params.timestampGranularities, openai.TranscriptionTimestampGranularitySegment)
				}
			}
		}
	}

	return params, nil
}

func (c *ProxyController) createTempFile(ctx *gin.Context) (string, error) {
	// Get the file from form data
	file, fileHeader, err := ctx.Request.FormFile("file")
	if err != nil {
		return "", fmt.Errorf("Failed to get file: %v", err)
	}
	defer file.Close()

	// Create a temporary file to save the uploaded audio
	tempDir := os.TempDir()
	tempFilePath := filepath.Join(tempDir, fileHeader.Filename)
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return "", fmt.Errorf("Failed to create temp file: %v", err)
	}
	defer tempFile.Close()

	// Copy the uploaded file to the temporary file
	if _, err = io.Copy(tempFile, file); err != nil {
		return "", fmt.Errorf("Failed to save audio file: %v", err)
	}

	// Close the file before returning
	tempFile.Close()

	return tempFilePath, nil
}

func (c *ProxyController) executeTranscription(ctx *gin.Context, adapter aiengine.AIEngineStream, request *gsc.AudioTranscriptionRequest, stream bool) error {
	return adapter.AudioTranscription(ctx, request, func(cbctx context.Context, completion gsc.Chunk, aiResponseError *gsc.AiEngineErrorResponse) error {
		if aiResponseError != nil {
			ctx.Writer.Header().Set(constants.HEADER_CONTENT_TYPE, constants.CONTENT_TYPE_JSON)
			ctx.JSON(http.StatusBadRequest, aiResponseError)
			return nil
		}

		var response []byte

		// Determine response type and format accordingly
		switch completion.Type() {
		case genericchatstorage.ChunkTypeAudioTranscriptionText:
			ctx.Writer.Header().Set(constants.HEADER_CONTENT_TYPE, constants.CONTENT_TYPE_TEXT_PLAIN)
			response = []byte(completion.Data().(string))
		case genericchatstorage.ChunkTypeAudioTranscriptionDelta:
			ctx.Writer.Header().Set(constants.HEADER_CONTENT_TYPE, constants.CONTENT_TYPE_EVENT_STREAM)
			marshalledResponse, err := json.Marshal(completion.Data())
			if err != nil {
				return err
			}
			response = marshalledResponse
		default:
			ctx.Writer.Header().Set(constants.HEADER_CONTENT_TYPE, constants.CONTENT_TYPE_JSON)
			marshalledResponse, err := json.Marshal(completion.Data())
			if err != nil {
				return err
			}
			response = marshalledResponse
		}

		// Write response based on stream mode
		var err error
		if stream || completion.IsStreaming() {
			_, err = ctx.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", response)))
		} else {
			_, err = ctx.Writer.Write(response)
		}

		if err != nil {
			return err
		}

		ctx.Writer.Flush()
		return nil
	})
}

// AudioSpeech godoc
//
//	@Summary		Generate Audio Speech
//	@Description	Convert text to speech using TTS model
//	@Tags			audio
//	@Accept			json
//	@Produce		audio/mpeg
//	@Param			session_id	header	string					false	"Session ID"	format(hex32)
//	@Param			model_id	header	string					false	"Model ID"		format(hex32)
//	@Param			chat_id		header	string					false	"Chat ID"		format(hex32)
//	@Param			request		body	gsc.AudioSpeechRequest	true	"Audio Speech Request"
//	@Success		200			{file}	binary
//	@Security		BasicAuth
//	@Router			/v1/audio/speech [post]
func (c *ProxyController) AudioSpeech(ctx *gin.Context) {
	// Parse request
	params, err := c.parseAudioSpeechParams(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get AI adapter
	chatID := params.head.ChatID
	if chatID == (lib.Hash{}) {
		var err error
		chatID, err = lib.GetRandomHash()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	adapter, err := c.aiEngine.GetAdapter(ctx, chatID.Hash, params.head.ModelID.Hash, params.head.SessionID.Hash, c.storeChatContext, c.forwardChatContext)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Prepare speech request
	speechRequest := &gsc.AudioSpeechRequest{
		Input:          params.input,
		Voice:          params.voice,
		ResponseFormat: params.responseFormat,
		Speed:          params.speed,
	}

	// Process speech generation with callback
	if err := c.executeSpeechGeneration(ctx, adapter, speechRequest); err != nil {
		c.log.Errorf("error sending speech generation request: %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

type audioSpeechParams struct {
	head           PromptHead
	input          string
	voice          string
	responseFormat string
	speed          float64
}

func (c *ProxyController) parseAudioSpeechParams(ctx *gin.Context) (*audioSpeechParams, error) {
	var head PromptHead
	if err := ctx.ShouldBindHeader(&head); err != nil {
		return nil, err
	}

	var requestBody struct {
		Input          string  `json:"input" binding:"required"`
		Voice          string  `json:"voice" binding:"required"`
		ResponseFormat string  `json:"response_format,omitempty"`
		Speed          float64 `json:"speed,omitempty"`
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		return nil, err
	}

	// Set defaults if not provided
	if requestBody.ResponseFormat == "" {
		requestBody.ResponseFormat = "mp3"
	}
	if requestBody.Speed == 0 {
		requestBody.Speed = 1.0
	}

	params := &audioSpeechParams{
		head:           head,
		input:          requestBody.Input,
		voice:          requestBody.Voice,
		responseFormat: requestBody.ResponseFormat,
		speed:          requestBody.Speed,
	}

	return params, nil
}

func (c *ProxyController) executeSpeechGeneration(ctx *gin.Context, adapter aiengine.AIEngineStream, request *gsc.AudioSpeechRequest) error {
	return adapter.AudioSpeech(ctx, request, func(cbctx context.Context, completion gsc.Chunk, aiResponseError *gsc.AiEngineErrorResponse) error {
		if aiResponseError != nil {
			ctx.Writer.Header().Set(constants.HEADER_CONTENT_TYPE, constants.CONTENT_TYPE_JSON)
			ctx.JSON(http.StatusBadRequest, aiResponseError)
			return nil
		}

		// Set appropriate content type based on response format
		contentType := "audio/mpeg" // default
		switch request.ResponseFormat {
		case "mp3":
			contentType = "audio/mpeg"
		case "opus":
			contentType = "audio/opus"
		case "aac":
			contentType = "audio/aac"
		case "flac":
			contentType = "audio/flac"
		case "wav":
			contentType = "audio/wav"
		case "pcm":
			contentType = "audio/pcm"
		}

		ctx.Writer.Header().Set(constants.HEADER_CONTENT_TYPE, contentType)

		// Handle audio speech response
		switch completion.Type() {
		case genericchatstorage.ChunkTypeAudioSpeech:
			// Write binary audio data directly
			audioData := completion.Data().([]byte)
			_, err := ctx.Writer.Write(audioData)
			if err != nil {
				return err
			}
		default:
			// Fallback for other response types
			marshalledResponse, err := json.Marshal(completion.Data())
			if err != nil {
				return err
			}
			_, err = ctx.Writer.Write(marshalledResponse)
			if err != nil {
				return err
			}
		}

		ctx.Writer.Flush()
		return nil
	})
}
