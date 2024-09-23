package aiengine

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"fmt"

	c "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	constants "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	api "github.com/sashabaranov/go-openai"
)

type AiEngine struct {
	client             *api.Client
	baseURL            string
	apiKey             string
	modelsConfigLoader *config.ModelConfigLoader
	log                lib.ILogger
}

var (
	ErrChatCompletion                = errors.New("chat completion error")
	ErrImageGenerationInvalidRequest = errors.New("invalid prodia image generation request")
	ErrImageGenerationRequest        = errors.New("image generation error")
	ErrJobCheckRequest               = errors.New("job status check error")
	ErrJobFailed                     = errors.New("job failed")
)

func NewAiEngine(apiBaseURL, apiKey string, modelsConfigLoader *config.ModelConfigLoader, log lib.ILogger) *AiEngine {
	return &AiEngine{
		modelsConfigLoader: modelsConfigLoader,
		baseURL:            apiBaseURL,
		apiKey:             apiKey,
		client: api.NewClientWithConfig(api.ClientConfig{
			BaseURL:    apiBaseURL,
			APIType:    api.APITypeOpenAI,
			HTTPClient: &http.Client{},
		}),
		log: log,
	}
}

func (a *AiEngine) Prompt(ctx context.Context, request *api.ChatCompletionRequest) (*api.ChatCompletionResponse, error) {
	response, err := a.client.CreateChatCompletion(
		ctx,
		*request,
	)

	if err != nil {
		err = lib.WrapError(ErrChatCompletion, err)
		a.log.Error(err)
		return nil, err
	}
	return &response, nil
}

func (a *AiEngine) PromptProdiaImage(ctx context.Context, request *ProdiaGenerationRequest, chunkCallback CompletionCallback) error {
	url := request.ApiUrl
	apiKey := request.ApiKey

	body := map[string]string{
		"model":  request.Model,
		"prompt": request.Prompt,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		err = lib.WrapError(ErrImageGenerationInvalidRequest, err)
		a.log.Error(err)
		return err
	}

	req, _ := http.NewRequest("POST", url, bytes.NewReader(payload))

	req.Header.Add("accept", constants.CONTENT_TYPE_JSON)
	req.Header.Add("content-type", constants.CONTENT_TYPE_JSON)
	req.Header.Add("X-Prodia-Key", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		err = lib.WrapError(ErrImageGenerationRequest, err)
		a.log.Error(err)
		return err
	}

	defer res.Body.Close()
	response, _ := io.ReadAll(res.Body)

	bodyStr := string(response)
	if strings.Contains(bodyStr, "Invalid Generation Parameters") {
		return ErrImageGenerationInvalidRequest
	}

	result := ProdiaGenerationResult{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		err = lib.WrapError(ErrImageGenerationRequest, err)
		a.log.Error(err)
		return err
	}

	job, err := a.waitJobResult(ctx, result.Job, apiKey)
	if err != nil {
		err = lib.WrapError(ErrImageGenerationRequest, err)
		a.log.Error(err)
		return err
	}

	return chunkCallback(job)
}

func (a *AiEngine) waitJobResult(ctx context.Context, jobID string, apiKey string) (*ProdiaGenerationResult, error) {
	url := fmt.Sprintf("https://api.prodia.com/v1/job/%s", jobID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		err = lib.WrapError(ErrJobCheckRequest, err)
		a.log.Error(err)
		return nil, err
	}

	req.Header.Add("accept", constants.CONTENT_TYPE_JSON)
	req.Header.Add("X-Prodia-Key", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		err = lib.WrapError(ErrJobCheckRequest, err)
		a.log.Error(err)
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		err = lib.WrapError(ErrJobCheckRequest, err)
		a.log.Error(err)
		return nil, err
	}

	var result ProdiaGenerationResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = lib.WrapError(ErrJobCheckRequest, err)
		a.log.Error(err)
		return nil, err
	}

	if result.Status == "succeeded" {
		return &result, nil
	}

	if result.Status == "failed" {
		return nil, ErrJobFailed
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(1 * time.Second):
	}
	return a.waitJobResult(ctx, jobID, apiKey)
}

func (a *AiEngine) PromptStream(ctx context.Context, request *api.ChatCompletionRequest, chunkCallback CompletionCallback) (*api.ChatCompletionStreamResponse, error) {
	resp, err := a.requestChatCompletionStream(ctx, request, chunkCallback)

	if err != nil {
		err = lib.WrapError(ErrChatCompletion, err)
		a.log.Error(err)
		return nil, err
	}

	return resp, nil
}

func (a *AiEngine) PromptCb(ctx *gin.Context, body *openai.ChatCompletionRequest) (interface{}, error) {
	resp := []interface{}{}
	if body.Stream {
		response, err := a.PromptStream(ctx, body, func(response interface{}) error {
			marshalledResponse, err := json.Marshal(response)
			if err != nil {
				return err
			}

			ctx.Writer.Header().Set(c.HEADER_CONTENT_TYPE, c.CONTENT_TYPE_EVENT_STREAM)

			_, err = ctx.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", marshalledResponse)))
			if err != nil {
				return err
			}

			resp = append(resp, response)

			ctx.Writer.Flush()
			return nil
		})

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return "", err
		}

		ctx.JSON(http.StatusOK, response)
		return resp, nil
	} else {
		response, err := a.Prompt(ctx, body)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return "", err
		}
		resp = append(resp, response)
		ctx.JSON(http.StatusOK, response)
		return resp, nil
	}
}

func (a *AiEngine) requestChatCompletionStream(ctx context.Context, request *api.ChatCompletionRequest, callback CompletionCallback) (*api.ChatCompletionStreamResponse, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to encode request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/chat/completions", bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	if a.apiKey != "" {
		req.Header.Set(c.HEADER_AUTHORIZATION, fmt.Sprintf("%s %s", c.BEARER, a.apiKey))
	}
	req.Header.Set(c.HEADER_CONTENT_TYPE, c.CONTENT_TYPE_JSON)
	req.Header.Set(c.HEADER_ACCEPT, c.CONTENT_TYPE_EVENT_STREAM)
	req.Header.Set(c.HEADER_CONNECTION, c.CONNECTION_KEEP_ALIVE)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "data: ") {
			data := line[6:] // Skip the "data: " prefix
			var completion api.ChatCompletionStreamResponse
			if err := json.Unmarshal([]byte(data), &completion); err != nil {
				if strings.Index(data, "[DONE]") != -1 {
					a.log.Debugf("reached end of the response")
				} else {
					a.log.Errorf("error decoding response: %s\n%s", err, line)
				}
				continue
			}
			// Call the callback function with the unmarshalled completion
			err := callback(&completion)
			if err != nil {
				return nil, fmt.Errorf("callback failed: %v", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading stream: %v", err)
	}

	return nil, err
}

func (a *AiEngine) GetModelsConfig() ([]string, []config.ModelConfig) {
	return a.modelsConfigLoader.GetAll()
}

func (a *AiEngine) GetLocalModels() ([]LocalModel, error) {
	models := []LocalModel{}

	IDs, modelsFromConfig := a.modelsConfigLoader.GetAll()
	for i, model := range modelsFromConfig {
		models = append(models, LocalModel{
			Id:      IDs[i],
			Name:    model.ModelName,
			Model:   model.ModelName,
			ApiType: model.ApiType,
		})
	}

	return models, nil
}
