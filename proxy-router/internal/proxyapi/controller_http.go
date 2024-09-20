package proxyapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
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
	fileMutexes map[string]*sync.Mutex // Map to store mutexes for each file
}

func NewProxyController(service *ProxyServiceSender, aiEngine *aiengine.AiEngine) *ProxyController {
	c := &ProxyController{
		service:     service,
		aiEngine:    aiEngine,
		fileMutexes: make(map[string]*sync.Mutex),
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

		promptJson, _ := json.Marshal(&prompt)
		fmt.Println("Prompt: ", string(promptJson))
		responseAt := time.Now()

		if t == "openai" {
			res, _ := c.aiEngine.PromptCb(ctx, &body)
			responses = res.([]interface{})
			// for _, response := range responses {
			// 	str := fmt.Sprintf("%v", response)
			// 	fmt.Println("Response: ", str)
			// }
			if err := c.storePromptResponseToFile(modelId, false, prompt, responses, promptAt, responseAt); err != nil {
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
				fmt.Println("Response: ", string(marshalledResponse))
				ctx.Writer.Flush()
				// Collect the response
				prodiaResponses = append(prodiaResponses, completion)
				if err := c.storePromptResponseToFile(modelId, false, prompt, prodiaResponses, promptAt, responseAt); err != nil {
					fmt.Println("Error storing prompt and responses:", err)
				}
				return nil
			})
		}
		return
	}

	res, err := c.service.SendPrompt(ctx, ctx.Writer, &body, head.SessionID.Hash)
	if err != nil {
		fmt.Println("Error sending prompt:", err)
		fmt.Printf("Error: %v\n", err)
		fmt.Println(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	responses = res.([]interface{})
	for _, response := range responses {
		str := fmt.Sprintf("%v", response)
		fmt.Println("Response: ", str)
	}

	responseAt := time.Now()
	sessionIdStr := head.SessionID.Hex()
	if err := c.storePromptResponseToFile(sessionIdStr, true, body, responses, promptAt, responseAt); err != nil {
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

func (c *ProxyController) storePromptResponseToFile(identifier string, isSession bool, prompt interface{}, responses []interface{}, promptAt, responseAt time.Time) error {
	var dir string
	if isSession {
		dir = "sessions"
	} else {
		dir = "models"
	}

	// Ensure the directory exists
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	// Create the file path
	filePath := filepath.Join(dir, identifier+".json")

	// Initialize a mutex for the file if not already present
	c.initFileMutex(filePath)

	// Lock the file mutex
	c.fileMutexes[filePath].Lock()
	defer c.fileMutexes[filePath].Unlock()

	// Read existing data from the file
	var data []map[string]interface{}
	if _, err := os.Stat(filePath); err == nil {
		// File exists, read the content
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(fileContent, &data); err != nil {
			return err
		}
	}

	response := ""
	for _, r := range responses {
		fmt.Println("Response: ", r)
		llmResponse, ok := r.(openai.ChatCompletionStreamResponse)
		if ok {
			response += fmt.Sprintf("%v", llmResponse.Choices[0].Delta.Content)
		} else {
			imageResponse, ok := r.(aiengine.ProdiaGenerationResult)
			if ok {
				response += fmt.Sprintf("%v", imageResponse.ImageUrl)
			} else {
				return fmt.Errorf("unknown response type")
			}
		}
	}

	// Create the new entry
	newEntry := map[string]interface{}{
		"prompt":   prompt,
		"response": response,
		// "chunks":     responses,
		"promptAt":   promptAt.UnixMilli(),
		"responseAt": responseAt.UnixMilli(),
	}

	// Append the new entry to the data
	data = append(data, newEntry)

	// Marshal the updated data
	updatedContent, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write back to the file
	if err := os.WriteFile(filePath, updatedContent, 0644); err != nil {
		return err
	}

	return nil
}

// Initialize a mutex for the file if not already present
func (c *ProxyController) initFileMutex(filePath string) {
	if _, exists := c.fileMutexes[filePath]; !exists {
		c.fileMutexes[filePath] = &sync.Mutex{}
	}
}
