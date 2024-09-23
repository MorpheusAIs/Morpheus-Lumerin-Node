package proxyapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	constants "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

type ProxyController struct {
	service     *ProxyServiceSender
	aiEngine    *aiengine.AiEngine
	chatStorage *ChatStorage
}

func NewProxyController(service *ProxyServiceSender, aiEngine *aiengine.AiEngine, chatStorage *ChatStorage) *ProxyController {
	c := &ProxyController{
		service:     service,
		aiEngine:    aiEngine,
		chatStorage: chatStorage,
	}

	return c
}

func (s *ProxyController) RegisterRoutes(r interfaces.Router) {
	r.POST("/proxy/sessions/initiate", s.InitiateSession)
	r.POST("/v1/chat/completions", s.Prompt)
	r.GET("/v1/models", s.Models)
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
//	@Param			session_id	header		string								false	"Session ID" format(hex32)
//	@Param 			model_id header string false "Model ID" format(hex32)
//	@Param			prompt		body		proxyapi.OpenAiCompletitionRequest	true	"Prompt"
//	@Success		200			{object}	proxyapi.ChatCompletionResponse
//	@Router			/v1/chat/completions [post]
func (c *ProxyController) Prompt(ctx *gin.Context) {
	var (
		body openai.ChatCompletionRequest
		head PromptHead
	)
	var responses []interface{}

	if err := ctx.ShouldBindHeader(&head); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Record the prompt time
	promptAt := time.Now()

	if (head.SessionID == lib.Hash{}) {
		body.Stream = ctx.GetHeader(constants.HEADER_ACCEPT) == constants.CONTENT_TYPE_JSON
		modelId := head.ModelID.Hex()

		prompt, t := c.GetBodyForLocalPrompt(modelId, &body)
		responseAt := time.Now()

		if t == "openai" {
			res, _ := c.aiEngine.PromptCb(ctx, &body)
			responses = res.([]interface{})
			if err := c.chatStorage.StorePromptResponseToFile(modelId, false, prompt, responses, promptAt, responseAt); err != nil {
				fmt.Println("Error storing prompt and responses:", err)
			}
		}
		if t == "prodia" {
			var prodiaResponses []interface{}
			c.aiEngine.PromptProdiaImage(ctx, prompt.(*aiengine.ProdiaGenerationRequest), func(completion interface{}) error {
				ctx.Writer.Header().Set(constants.HEADER_CONTENT_TYPE, constants.CONTENT_TYPE_EVENT_STREAM)
				marshalledResponse, err := json.Marshal(completion)
				if err != nil {
					return err
				}
				_, err = ctx.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", marshalledResponse)))
				if err != nil {
					fmt.Println("Error writing response:", err)
					return err
				}
				ctx.Writer.Flush()

				prodiaResponses = append(prodiaResponses, completion)
				if err := c.chatStorage.StorePromptResponseToFile(modelId, false, prompt, prodiaResponses, promptAt, responseAt); err != nil {
					fmt.Println("Error storing prompt and responses:", err)
				}
				return nil
			})
		}
		return
	}

	res, err := c.service.SendPrompt(ctx, ctx.Writer, &body, head.SessionID.Hash)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	responses = res.([]interface{})
	responseAt := time.Now()
	sessionIdStr := head.SessionID.Hex()
	if err := c.chatStorage.StorePromptResponseToFile(sessionIdStr, true, body, responses, promptAt, responseAt); err != nil {
		fmt.Println("Error storing prompt and responses:", err)
	}
	return
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

func (c *ProxyController) GetBodyForLocalPrompt(modelId string, req *openai.ChatCompletionRequest) (interface{}, string) {
	if modelId == "" {
		req.Model = "llama2"
		return req, "openai"
	}

	ids, models := c.aiEngine.GetModelsConfig()

	for i, model := range models {
		if ids[i] == modelId {
			if model.ApiType == "openai" {
				req.Model = model.ModelName
				return req, model.ApiType
			}

			if model.ApiType == "prodia" {
				prompt := &aiengine.ProdiaGenerationRequest{
					Model:  model.ModelName,
					Prompt: req.Messages[0].Content,
					ApiUrl: model.ApiURL,
					ApiKey: model.ApiKey,
				}
				return prompt, model.ApiType
			}

			return req, "openai"
		}
	}

	req.Model = "llama2"
	return req, "openai"
}
