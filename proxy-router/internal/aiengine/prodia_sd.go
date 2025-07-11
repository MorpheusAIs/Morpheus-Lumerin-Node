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

const API_TYPE_PRODIA_SD = "prodia-sd"

type ProdiaSD struct {
	modelName string
	apiURL    string
	apiKey    string

	log lib.ILogger
}

func NewProdiaSDEngine(modelName, apiURL, apiKey string, log lib.ILogger) *ProdiaSD {
	if apiURL == "" {
		apiURL = PRODIA_DEFAULT_BASE_URL
	}
	return &ProdiaSD{
		modelName: modelName,
		apiURL:    apiURL,
		apiKey:    apiKey,
		log:       log,
	}
}

func (s *ProdiaSD) Prompt(ctx context.Context, prompt *openai.ChatCompletionRequest, cb gcs.CompletionCallback) error {
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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/sd/generate", s.apiURL), bytes.NewReader(payload))
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

func (s *ProdiaSD) AudioTranscription(ctx context.Context, prompt *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) error {
	return fmt.Errorf("audio transcription not supported")
}

func (s *ProdiaSD) AudioSpeech(ctx context.Context, prompt *gcs.AudioSpeechRequest, cb gcs.CompletionCallback) error {
	return fmt.Errorf("audio speech not supported")
}

func (s *ProdiaSD) ApiType() string {
	return API_TYPE_PRODIA_SD
}

var _ AIEngineStream = &ProdiaSD{}
