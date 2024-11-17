package proxyapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	constants "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
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
}

func NewProxyController(service *ProxyServiceSender, aiEngine AIEngine, chatStorage genericchatstorage.ChatStorageInterface, storeChatContext, forwardChatContext bool, log lib.ILogger) *ProxyController {
	c := &ProxyController{
		service:            service,
		aiEngine:           aiEngine,
		chatStorage:        chatStorage,
		storeChatContext:   storeChatContext,
		forwardChatContext: forwardChatContext,
		log:                log,
	}

	return c
}

func (s *ProxyController) RegisterRoutes(r interfaces.Router) {
	r.POST("/proxy/sessions/initiate", s.InitiateSession)
	r.POST("/v1/chat/completions", s.Prompt)
	r.GET("/v1/models", s.Models)
	r.GET("/v1/chats", s.GetChats)
	r.GET("/v1/chats/:id", s.GetChat)
	r.DELETE("/v1/chats/:id", s.DeleteChat)
	r.POST("/v1/chats/:id", s.UpdateChatTitle)
}

// InitiateSession godoc
//
//	@Summary		Initiate Session with Provider
//	@Description	sends a handshake to the provider
//	@Tags			chat
//	@Produce		json
//	@Param			initiateSession	body		proxyapi.InitiateSessionReq	true	"Initiate Session"
//	@Success		200				{object}	morrpcmesssage.SessionRes
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
//	@Param			session_id	header		string	false	"Session ID"	format(hex32)
//	@Param			model_id	header		string	false	"Model ID"		format(hex32)
//	@Param			chat_id		header		string	false	"Chat ID"		format(hex32)
//	@Param			prompt		body		string	true	"Prompt"
//	@Success		200			{object}	string
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
//	@Tags		chat
//	@Produce	json
//	@Success	200	{object}	[]aiengine.LocalModel
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
