package aiengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	c "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/sashabaranov/go-openai"
)

const API_TYPE_HYPERBOLIC_SD = "hyperbolic-sd"
const HYPERBOLIC_DEFAULT_BASE_URL = "https://api.hyperbolic.xyz/v1"

type HyperbolicSD struct {
	modelName string
	apiURL    string
	apiKey    string

	log lib.ILogger
}

type HyperbolicImageGenerationResult struct {
	Images []Image `json:"images"`
}

type Image struct {
	Image string `json:"image"`
}

func NewHyperbolicSDEngine(modelName, apiURL, apiKey string, log lib.ILogger) *HyperbolicSD {
	if apiURL == "" {
		apiURL = HYPERBOLIC_DEFAULT_BASE_URL
	}
	return &HyperbolicSD{
		modelName: modelName,
		apiURL:    apiURL,
		apiKey:    apiKey,
		log:       log,
	}
}

func (s *HyperbolicSD) Prompt(ctx context.Context, prompt *openai.ChatCompletionRequest, cb gcs.CompletionCallback) error {
	body := map[string]string{
		"model_name": s.modelName,
		"prompt":     prompt.Messages[len(prompt.Messages)-1].Content,
		"height":     "512",
		"width":      "512",
		"backend":    "auto",
	}

	payload, err := json.Marshal(body)
	if err != nil {
		err = lib.WrapError(ErrImageGenerationInvalidRequest, err)
		s.log.Error(err)
		return err
	}

	s.log.Debugf("payload: %s", payload)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/image/generation", s.apiURL), bytes.NewReader(payload))
	if err != nil {
		err = lib.WrapError(ErrImageGenerationRequest, err)
		s.log.Error(err)
	}

	s.log.Debugf("payload: %s", payload)

	req.Header.Add(c.HEADER_ACCEPT, c.CONTENT_TYPE_JSON)
	req.Header.Add(c.HEADER_CONTENT_TYPE, c.CONTENT_TYPE_JSON)
	req.Header.Add(c.HEADER_AUTHORIZATION, fmt.Sprintf("%s %s", c.BEARER, s.apiKey))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		err = lib.WrapError(ErrImageGenerationRequest, err)
		s.log.Error(err)
		return err
	}

	defer res.Body.Close()
	response, err := io.ReadAll(res.Body)
	if err != nil {
		err = lib.WrapError(ErrImageGenerationRequest, err)
		s.log.Error(err)
		return err
	}

	s.log.Debugf("response: %s", response)

	result := HyperbolicImageGenerationResult{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		err = lib.WrapError(ErrImageGenerationRequest, err)
		s.log.Error(err)
		return err
	}

	dataPrefix := "data:image/png;base64,"
	chunk := gcs.NewChunkImageRawContent(&gcs.ImageRawContentResult{
		ImageRawContent: dataPrefix + result.Images[0].Image,
	})

	return cb(ctx, chunk)
}

func (s *HyperbolicSD) ApiType() string {
	return API_TYPE_HYPERBOLIC_SD
}

var _ AIEngineStream = &HyperbolicSD{}
