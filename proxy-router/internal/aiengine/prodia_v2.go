package aiengine

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	c "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/sashabaranov/go-openai"
)

const API_TYPE_PRODIA_V2 = "prodia-v2"
const PRODIA_V2_DEFAULT_BASE_URL = "https://inference.prodia.com/v2"

var (
	ErrCapacity               = errors.New("unable to schedule job with current token")
	ErrBadResponse            = errors.New("bad response")
	ErrVideoGenerationRequest = errors.New("video generation error")
)

type ProdiaV2 struct {
	modelName string
	apiURL    string
	apiKey    string

	log lib.ILogger
}

func NewProdiaV2Engine(modelName, apiURL, apiKey string, log lib.ILogger) *ProdiaV2 {
	if apiURL == "" {
		apiURL = PRODIA_V2_DEFAULT_BASE_URL
	}
	return &ProdiaV2{
		modelName: modelName,
		apiURL:    apiURL,
		apiKey:    apiKey,
		log:       log,
	}
}

func (s *ProdiaV2) Prompt(ctx context.Context, prompt *openai.ChatCompletionRequest, cb gcs.CompletionCallback) error {
	body := map[string]interface{}{
		"type": s.modelName,
		"config": map[string]string{
			"prompt": prompt.Messages[len(prompt.Messages)-1].Content,
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		err = lib.WrapError(ErrImageGenerationInvalidRequest, err)
		s.log.Error(err)
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/job", s.apiURL), bytes.NewReader(payload))
	if err != nil {
		err = lib.WrapError(ErrImageGenerationRequest, err)
		s.log.Error(err)
	}

	req.Header.Add(c.HEADER_ACCEPT, c.CONTENT_TYPE_VIDEO_MP4)
	req.Header.Add(c.HEADER_CONTENT_TYPE, c.CONTENT_TYPE_JSON)
	req.Header.Add(c.HEADER_AUTHORIZATION, fmt.Sprintf("Bearer %s", s.apiKey))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		err = lib.WrapError(ErrImageGenerationRequest, err)
		s.log.Error(err)
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusTooManyRequests {
		return ErrCapacity
	} else if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusMultipleChoices {
		return lib.WrapError(ErrBadResponse, fmt.Errorf("status code: %d", res.StatusCode))
	}

	response, err := io.ReadAll(res.Body)
	if err != nil {
		err = lib.WrapError(ErrVideoGenerationRequest, err)
		s.log.Error(err)
		return err
	}

	sEnc := b64.StdEncoding.EncodeToString(response)

	dataPrefix := "data:video/mp4;base64,"
	chunk := gcs.NewChunkVideo(&gcs.VideoGenerationResult{
		VideoRawContent: dataPrefix + sEnc,
	})

	return cb(ctx, chunk)
}

func (s *ProdiaV2) ApiType() string {
	return API_TYPE_PRODIA_V2
}

var _ AIEngineStream = &ProdiaV2{}
