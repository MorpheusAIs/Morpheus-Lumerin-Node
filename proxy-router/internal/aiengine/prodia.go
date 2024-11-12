package aiengine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	constants "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/sashabaranov/go-openai"
)

type Prodia struct {
	modelName string
	apiURL    string
	apiKey    string

	log lib.ILogger
}

const HEADER_PRODIA_KEY = "X-Prodia-Key"

func NewProdiaEngine(modelName, apiURL, apiKey string, log lib.ILogger) *Prodia {
	return &Prodia{
		modelName: modelName,
		apiURL:    apiURL,
		apiKey:    apiKey,
		log:       log,
	}
}

func (s *Prodia) Prompt(ctx context.Context, prompt *openai.ChatCompletionRequest, cb gcs.CompletionCallback) error {
	body := map[string]string{
		"model":  s.modelName,
		"prompt": prompt.Messages[0].Content,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		err = lib.WrapError(ErrImageGenerationInvalidRequest, err)
		s.log.Error(err)
		return err
	}

	s.log.Debugf("payload: %s", payload)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/sd/generate", s.apiURL), bytes.NewReader(payload))
	if err != nil {
		err = lib.WrapError(ErrImageGenerationRequest, err)
		s.log.Error(err)
	}

	req.Header.Add(constants.HEADER_ACCEPT, constants.CONTENT_TYPE_JSON)
	req.Header.Add(constants.HEADER_CONTENT_TYPE, constants.CONTENT_TYPE_JSON)
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

	s.log.Debugf("response: %s", response)

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

	job, err := s.waitJobResult(ctx, result.Job)
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

	return cb(ctx, chunk)
}

func (s *Prodia) waitJobResult(ctx context.Context, jobID string) (*ProdiaGenerationResult, error) {
	url := fmt.Sprintf("%s/job/%s", s.apiURL, jobID)

	s.log.Debugf("waiting for job result: %s", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		err = lib.WrapError(ErrJobCheckRequest, err)
		s.log.Error(err)
		return nil, err
	}

	req.Header.Add("accept", constants.CONTENT_TYPE_JSON)
	req.Header.Add("X-Prodia-Key", s.apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		err = lib.WrapError(ErrJobCheckRequest, err)
		s.log.Error(err)
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		err = lib.WrapError(ErrJobCheckRequest, err)
		s.log.Error(err)
		return nil, err
	}

	var result ProdiaGenerationResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		err = lib.WrapError(ErrJobCheckRequest, err)
		s.log.Error(err)
		return nil, err
	}

	s.log.Debugf("job result: %v", result)

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

	return s.waitJobResult(ctx, jobID)
}

func (s *Prodia) ApiType() string {
	return "prodia"
}

var _ AIEngineStream = &Prodia{}

type ProdiaGenerationResult struct {
	Job      string `json:"job"`
	Status   string `json:"status"`
	ImageUrl string `json:"imageUrl" binding:"omitempty"`
}
