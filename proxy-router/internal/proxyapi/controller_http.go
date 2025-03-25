package proxyapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	constants "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

type AIEngine interface {
	GetLocalModels() ([]aiengine.LocalModel, error)
	GetAdapter(ctx context.Context, chatID, modelID, sessionID common.Hash, storeContext, forwardContext bool) (aiengine.AIEngineStream, error)
}

type ProxyController struct {
	service            *ProxyServiceSender
	aiEngine           AIEngine
	chatStorage        genericchatstorage.ChatStorageInterface
	storeChatContext   bool
	forwardChatContext bool
	log                lib.ILogger
	authConfig         system.HTTPAuthConfig
	ipfsManager        *IpfsManager
}

func NewProxyController(service *ProxyServiceSender, aiEngine AIEngine, chatStorage genericchatstorage.ChatStorageInterface, storeChatContext, forwardChatContext bool, authConfig system.HTTPAuthConfig, log lib.ILogger) *ProxyController {
	ipfsManager := NewIpfsManager(log)
	c := &ProxyController{
		service:            service,
		aiEngine:           aiEngine,
		chatStorage:        chatStorage,
		storeChatContext:   storeChatContext,
		forwardChatContext: forwardChatContext,
		log:                log,
		authConfig:         authConfig,
		ipfsManager:        ipfsManager,
	}

	return c
}

func (s *ProxyController) RegisterRoutes(r interfaces.Router) {
	r.POST("/proxy/provider/ping", s.Ping)
	r.POST("/proxy/sessions/initiate", s.authConfig.CheckAuth("initiate_session"), s.InitiateSession)
	r.POST("/v1/chat/completions", s.authConfig.CheckAuth("chat"), s.Prompt)
	r.GET("/v1/models", s.authConfig.CheckAuth("get_local_models"), s.Models)
	r.GET("/v1/chats", s.authConfig.CheckAuth("get_chat_history"), s.GetChats)
	r.GET("/v1/chats/:id", s.authConfig.CheckAuth("get_chat_history"), s.GetChat)
	r.DELETE("/v1/chats/:id", s.authConfig.CheckAuth("edit_chat_history"), s.DeleteChat)
	r.POST("/v1/chats/:id", s.authConfig.CheckAuth("edit_chat_history"), s.UpdateChatTitle)

	r.POST("/ipfs/pin", s.authConfig.CheckAuth("ipfs_pin"), s.Pin)
	r.POST("/ipfs/unpin", s.authConfig.CheckAuth("ipfs_unpin"), s.Unpin)
	r.GET("/ipfs/download/:cidHash", s.authConfig.CheckAuth("ipfs_get"), s.DownloadFile)
	r.GET("/ipfs/download/stream/:cidHash", s.authConfig.CheckAuth("ipfs_get"), s.StreamDownloadFile)
	r.POST("/ipfs/add", s.authConfig.CheckAuth("ipfs_add"), s.AddFile)
	r.GET("/ipfs/version", s.authConfig.CheckAuth("ipfs_version"), s.GetIpfsVersion)
	r.GET("/ipfs/pin", s.authConfig.CheckAuth("ipfs_pinned"), s.GetPinnedFiles)
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

	err = adapter.Prompt(ctx, &body, func(cbctx context.Context, completion genericchatstorage.Chunk) error {
		marshalledResponse, err := json.Marshal(completion.Data())
		if err != nil {
			return err
		}

		_, err = ctx.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", marshalledResponse)))
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
