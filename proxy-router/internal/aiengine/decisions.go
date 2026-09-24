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

	c "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
)

const API_TYPE_DECISIONS = "decisions"

// Decisions is a thin HTTP adapter that POSTs Decisions/System One JSON to the
// configured full apiUrl (embeddings-parity: no path append).
type Decisions struct {
	baseURL    string
	apiKey     string
	modelName  string
	client     *http.Client
	llmTimeout time.Duration
	log        lib.ILogger
}

func NewDecisionsEngine(modelName, baseURL, apiKey string, llmTimeout time.Duration, log lib.ILogger, httpClient *http.Client) *Decisions {
	if baseURL != "" {
		baseURL = strings.TrimSuffix(baseURL, "/")
	}
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &Decisions{
		baseURL:    baseURL,
		modelName:  modelName,
		apiKey:     apiKey,
		client:     httpClient,
		llmTimeout: llmTimeout,
		log:        log,
	}
}

func (a *Decisions) Prompt(ctx context.Context, prompt *gcs.OpenAICompletionRequestExtra, cb gcs.CompletionCallback) error {
	return fmt.Errorf("chat completions not supported on decisions adapter")
}

func (a *Decisions) AudioTranscription(ctx context.Context, prompt *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) error {
	return fmt.Errorf("audio transcription not supported")
}

func (a *Decisions) AudioSpeech(ctx context.Context, prompt *gcs.AudioSpeechRequest, cb gcs.CompletionCallback) error {
	return fmt.Errorf("audio speech not supported")
}

func (a *Decisions) Embeddings(ctx context.Context, prompt *gcs.EmbeddingsRequest, cb gcs.CompletionCallback) error {
	return fmt.Errorf("embeddings not supported")
}

// Decisions POSTs the Decisions body to the full apiUrl.
// Body model is overwritten with models-config.modelName (embeddings parity).
// Unknown request Extra fields are forwarded; Extra["type"] is never sent upstream
// (stripped by the receiver before dispatch).
func (a *Decisions) Decisions(ctx context.Context, req *gcs.DecisionsRequest, cb gcs.CompletionCallback) error {
	if a.llmTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, a.llmTimeout)
		defer cancel()
	}

	log := a.log.With("request_id", lib.RequestIDFromContext(ctx))
	// J2: never log full state/questions at info — PII / case data.
	log.Infof("decisions request: model=%s questions=%d", a.modelName, len(req.Questions))

	req.Model = a.modelName

	requestBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to encode decisions request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", a.baseURL, bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create decisions request: %w", err)
	}

	if a.apiKey != "" {
		httpReq.Header.Set(c.HEADER_AUTHORIZATION, fmt.Sprintf("%s %s", c.BEARER, a.apiKey))
	}
	httpReq.Header.Set(c.HEADER_CONTENT_TYPE, c.CONTENT_TYPE_JSON)
	httpReq.Header.Set(c.HEADER_CONNECTION, c.CONNECTION_KEEP_ALIVE)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send decisions request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return a.readSanitizedError(ctx, resp.StatusCode, resp.Body, cb)
	}

	var decResp gcs.DecisionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&decResp); err != nil {
		return fmt.Errorf("failed to decode decisions response: %w", err)
	}

	chunk := gcs.NewChunkDecisions(decResp)
	return cb(ctx, chunk, nil)
}

// readSanitizedError maps upstream failures to a client-safe error body.
// J6: passthrough of unknown *request* fields does not imply forwarding raw
// upstream error bodies (esp. 422 validation) to untrusted clients.
func (a *Decisions) readSanitizedError(ctx context.Context, statusCode int, body io.Reader, cb gcs.CompletionCallback) error {
	raw, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("failed to read error response body: %w", err)
	}

	log := a.log.With("request_id", lib.RequestIDFromContext(ctx))
	// Log status only — do not dump raw upstream body at info (may contain PII).
	log.Warnf("decisions upstream error status=%d body_len=%d", statusCode, len(raw))

	msg := "upstream decisions request failed"
	errType := "upstream_error"
	if statusCode == http.StatusUnprocessableEntity || statusCode == http.StatusBadRequest {
		msg = "upstream rejected the decisions request"
		errType = "validation_error"
	}

	parsed := map[string]interface{}{
		"error": map[string]interface{}{
			"message": msg,
			"type":    errType,
			"code":    statusCode,
		},
	}
	return cb(ctx, nil, gcs.NewAiEngineErrorResponse(statusCode, parsed))
}

func (a *Decisions) ApiType() string {
	return API_TYPE_DECISIONS
}

var _ AIEngineStream = &Decisions{}
