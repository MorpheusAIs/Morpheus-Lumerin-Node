package aiengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	c "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/sashabaranov/go-openai"
)

const API_TYPE_PRODIA_SDXL = "prodia-sdxl"

type ProdiaSDXL struct {
	modelName string
	apiURL    string
	apiKey    string

	log lib.ILogger
}

func NewProdiaSDXLEngine(modelName, apiURL, apiKey string, log lib.ILogger) *ProdiaSDXL {
	if apiURL == "" {
		apiURL = PRODIA_DEFAULT_BASE_URL
	}
	return &ProdiaSDXL{
		modelName: modelName,
		apiURL:    apiURL,
		apiKey:    apiKey,
		log:       log,
	}
}

func (s *ProdiaSDXL) Prompt(ctx context.Context, prompt *openai.ChatCompletionRequest, cb gcs.CompletionCallback) error {
	body := map[string]string{
		"model":  s.modelName,
		"prompt": prompt.Messages[len(prompt.Messages)-1].Content,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		err = lib.WrapError(ErrImageGenerationInvalidRequest, err)
		s.log.Error(err)
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/sdxl/generate", s.apiURL), bytes.NewReader(payload))
	if err != nil {
		err = lib.WrapError(ErrImageGenerationRequest, err)
		s.log.Error(err)
	}

	req.Header.Add(c.HEADER_ACCEPT, c.CONTENT_TYPE_JSON)
	req.Header.Add(c.HEADER_CONTENT_TYPE, c.CONTENT_TYPE_JSON)
	req.Header.Add(HEADER_PRODIA_KEY, s.apiKey)

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

	bodyStr := string(response)
	if strings.Contains(bodyStr, "Invalid Generation Parameters") {
		return lib.WrapError(ErrImageGenerationRequest, fmt.Errorf(bodyStr))
	}

	result := ProdiaGenerationResult{}
	err = json.Unmarshal(response, &result)
	if err != nil {
		err = lib.WrapError(ErrImageGenerationRequest, err)
		s.log.Error(err)
		return err
	}

	job, err := waitJobResult(ctx, s.apiURL, s.apiKey, result.Job)
	if err != nil {
		err = lib.WrapError(ErrImageGenerationRequest, err)
		s.log.Error(err)
		return err
	}

	chunk := gcs.NewChunkImage(&gcs.ImageGenerationResult{
		Job:      job.Job,
		ImageUrl: job.ImageUrl,
		Status:   job.Status,
	})

	return cb(ctx, chunk, nil)
}

func (s *ProdiaSDXL) AudioTranscription(ctx context.Context, prompt *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) error {
	return fmt.Errorf("audio transcription not supported")
}

func (s *ProdiaSDXL) ApiType() string {
	return API_TYPE_PRODIA_SDXL
}

var _ AIEngineStream = &ProdiaSDXL{}
