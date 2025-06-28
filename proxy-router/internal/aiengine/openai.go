package aiengine

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	c "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/sashabaranov/go-openai"
)

const API_TYPE_OPENAI = "openai"

type OpenAI struct {
	baseURL   string
	apiKey    string
	modelName string
	client    *http.Client
	log       lib.ILogger
}

func NewOpenAIEngine(modelName, baseURL, apiKey string, log lib.ILogger) *OpenAI {
	if baseURL != "" {
		baseURL = strings.TrimSuffix(baseURL, "/")
	}
	return &OpenAI{
		baseURL:   baseURL,
		modelName: modelName,
		apiKey:    apiKey,
		client:    &http.Client{},
		log:       log,
	}
}

func (a *OpenAI) Prompt(ctx context.Context, compl *openai.ChatCompletionRequest, cb gcs.CompletionCallback) error {
	compl.Model = a.modelName
	requestBody, err := json.Marshal(compl)
	if err != nil {
		return fmt.Errorf("failed to encode request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/chat/completions", bytes.NewReader(requestBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	if a.apiKey != "" {
		req.Header.Set(c.HEADER_AUTHORIZATION, fmt.Sprintf("%s %s", c.BEARER, a.apiKey))
	}
	req.Header.Set(c.HEADER_CONTENT_TYPE, c.CONTENT_TYPE_JSON)
	req.Header.Set(c.HEADER_CONNECTION, c.CONNECTION_KEEP_ALIVE)
	if compl.Stream {
		req.Header.Set(c.HEADER_ACCEPT, c.CONTENT_TYPE_EVENT_STREAM)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	a.log.Debugf("AI Model responded with status code: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		a.log.Warnf("AI Model responded with error: %s", resp.StatusCode)
		return a.readError(ctx, resp.Body, cb)
	}

	if isContentTypeStream(resp.Header) {
		return a.readStream(ctx, resp.Body, cb)
	}

	return a.readResponse(ctx, resp.Body, cb)
}

func (a *OpenAI) readResponse(ctx context.Context, body io.Reader, cb gcs.CompletionCallback) error {
	var compl openai.ChatCompletionResponse
	if err := json.NewDecoder(body).Decode(&compl); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	chunk := gcs.NewChunkText(&compl)
	err := cb(ctx, chunk, nil)
	if err != nil {
		return fmt.Errorf("callback failed: %v", err)
	}

	return nil
}

func (a *OpenAI) readError(ctx context.Context, body io.Reader, cb gcs.CompletionCallback) error {
	var aiEngineErrorResponse interface{}
	if err := json.NewDecoder(body).Decode(&aiEngineErrorResponse); err != nil {
		return fmt.Errorf("failed to decode error response: %v", err)
	}

	err := cb(ctx, nil, gcs.NewAiEngineErrorResponse(aiEngineErrorResponse))
	if err != nil {
		return fmt.Errorf("callback failed: %v", err)
	}
	return nil
}

func (a *OpenAI) readStream(ctx context.Context, body io.Reader, cb gcs.CompletionCallback) error {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, StreamDataPrefix) {
			data := line[len(StreamDataPrefix):] // Skip the "data: " prefix
			var compl openai.ChatCompletionStreamResponse
			if err := json.Unmarshal([]byte(data), &compl); err != nil {
				if isStreamFinished(data) {
					return nil
				} else {
					return fmt.Errorf("error decoding response: %s\n%s", err, line)
				}
			}
			// Call the callback function with the unmarshalled completion
			chunk := gcs.NewChunkStreaming(&compl)
			err := cb(ctx, chunk, nil)
			if err != nil {
				return fmt.Errorf("callback failed: %v", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %v", err)
	}

	return nil
}

// readTranscriptionStream handles streaming audio transcription responses
func (a *OpenAI) readTranscriptionStream(ctx context.Context, body io.Reader, cb gcs.CompletionCallback) error {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, StreamDataPrefix) {
			data := line[len(StreamDataPrefix):] // Skip the "data: " prefix

			if isStreamFinished(data) {
				return nil
			}

			var event map[string]interface{}
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				return fmt.Errorf("error decoding transcription stream response: %s\n%s", err, line)
			}

			eventType, ok := event["type"].(string)
			if !ok {
				continue
			}
			switch eventType {
			case "transcript.text.delta":
				dataStruct := gcs.AudioTranscriptionDelta{
					Delta: event["delta"].(string),
					Type:  eventType,
				}
				chunk := gcs.NewChunkAudioTranscriptionDelta(dataStruct)
				if err := cb(ctx, chunk, nil); err != nil {
					return fmt.Errorf("callback failed: %v", err)
				}

			case "transcript.text.done":
				dataStruct := gcs.AudioTranscriptionDelta{
					Type: eventType,
					Text: event["text"].(string),
				}
				chunk := gcs.NewChunkAudioTranscriptionDelta(dataStruct)
				if err := cb(ctx, chunk, nil); err != nil {
					return fmt.Errorf("callback failed: %v", err)
				}
				return nil

			case "error":
				errorMsg, _ := event["error"].(map[string]interface{})
				return fmt.Errorf("transcription error: %v", errorMsg)

			default:
				a.log.Debugf("Received transcription event: %s", eventType)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading transcription stream: %v", err)
	}

	return nil
}

func (a *OpenAI) AudioTranscription(ctx context.Context, audioRequest *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) error {
	audioRequest.Model = a.modelName

	// Prepare the request
	req, err := a.prepareTranscriptionRequest(ctx, audioRequest)
	if err != nil {
		return fmt.Errorf("failed to prepare transcription request: %w", err)
	}

	if audioRequest.Stream {
		req.Header.Set(c.HEADER_ACCEPT, c.CONTENT_TYPE_EVENT_STREAM)
	}

	// Send the request
	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send transcription request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return a.readError(ctx, resp.Body, cb)
	}

	// Check if response is streaming
	if audioRequest.Stream && isContentTypeStream(resp.Header) {
		return a.readTranscriptionStream(ctx, resp.Body, cb)
	}

	// Process the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	return a.processTranscriptionResponse(ctx, responseBody, audioRequest.Format, cb)
}

func (a *OpenAI) prepareTranscriptionRequest(ctx context.Context, audioRequest *gcs.AudioTranscriptionRequest) (*http.Request, error) {
	// Open the audio file
	file, err := os.Open(audioRequest.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}

	// Create multipart form data
	pr, contentType, err := a.createMultipartForm(file, audioRequest)
	if err != nil {
		return nil, err
	}

	// Create the request
	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL, pr)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", contentType)
	req.Header.Set(c.HEADER_CONNECTION, c.CONNECTION_KEEP_ALIVE)
	if a.apiKey != "" {
		req.Header.Set(c.HEADER_AUTHORIZATION, fmt.Sprintf("%s %s", c.BEARER, a.apiKey))
	}

	return req, nil
}

func (a *OpenAI) createMultipartForm(
	file *os.File,
	audioReq *gcs.AudioTranscriptionRequest,
) (*io.PipeReader, string, error) {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer mw.Close()
		defer file.Close()

		_ = mw.WriteField("model", audioReq.Model)
		_ = a.addOptionalFields(mw, audioReq)

		// file part
		part, _ := mw.CreateFormFile("file", filepath.Base(file.Name()))
		_, err := io.Copy(part, file) // streamed, no big buffer
		if err != nil {
			pw.CloseWithError(err) // propagate errors
		}
	}()

	return pr, mw.FormDataContentType(), nil
}

func (a *OpenAI) addOptionalFields(writer *multipart.Writer, audioRequest *gcs.AudioTranscriptionRequest) error {
	// Add optional parameters if provided
	if audioRequest.Language != "" {
		if err := writer.WriteField("language", audioRequest.Language); err != nil {
			return fmt.Errorf("failed to add language field: %w", err)
		}
	}
	if audioRequest.Prompt != "" {
		if err := writer.WriteField("prompt", audioRequest.Prompt); err != nil {
			return fmt.Errorf("failed to add prompt field: %w", err)
		}
	}
	if audioRequest.Format != "" {
		if err := writer.WriteField("response_format", string(audioRequest.Format)); err != nil {
			return fmt.Errorf("failed to add response_format field: %w", err)
		}
	}
	if audioRequest.Temperature != 0 {
		if err := writer.WriteField("temperature", fmt.Sprintf("%f", audioRequest.Temperature)); err != nil {
			return fmt.Errorf("failed to add temperature field: %w", err)
		}
	}
	if audioRequest.TimestampGranularity != "" {
		fmt.Println("Adding timestamp_granularity:", audioRequest.TimestampGranularity)
		if err := writer.WriteField("timestamp_granularity", string(audioRequest.TimestampGranularity)); err != nil {
			return fmt.Errorf("failed to add timestamp_granularity field: %w", err)
		}
	}
	if len(audioRequest.TimestampGranularities) > 0 {
		for _, granularity := range audioRequest.TimestampGranularities {
			if err := writer.WriteField("timestamp_granularities[]", string(granularity)); err != nil {
				return fmt.Errorf("failed to add timestamp_granularity field: %w", err)
			}
		}
	}
	if audioRequest.Stream {
		if err := writer.WriteField("stream", "true"); err != nil {
			return fmt.Errorf("failed to add stream field: %w", err)
		}
	}
	return nil
}

func (a *OpenAI) addFilePart(writer *multipart.Writer, file *os.File) error {
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	filePart, err := writer.CreateFormFile("file", fileInfo.Name())
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	if _, err := io.Copy(filePart, file); err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}

	return nil
}

func (a *OpenAI) processTranscriptionResponse(ctx context.Context, responseBody []byte, format openai.AudioResponseFormat, cb gcs.CompletionCallback) error {
	if format == openai.AudioResponseFormatJSON || format == openai.AudioResponseFormatVerboseJSON {
		// Create a transcription response wrapper since we don't have a direct openai.AudioResponse struct
		var transcriptionResponse openai.AudioResponse
		if err := json.Unmarshal(responseBody, &transcriptionResponse); err != nil {
			return fmt.Errorf("failed to parse transcription response: %w", err)
		}

		// Create a proper response chunk
		chunk := gcs.NewChunkAudioTranscriptionJson(transcriptionResponse)

		// Call the callback with the transcription result
		if err := cb(ctx, chunk, nil); err != nil {
			return fmt.Errorf("callback failed: %w", err)
		}
		return nil
	} else {
		chunk := gcs.NewChunkAudioTranscriptionText(string(responseBody))
		if err := cb(ctx, chunk, nil); err != nil {
			return fmt.Errorf("callback failed: %w", err)
		}

		return nil
	}
}

func (a *OpenAI) ApiType() string {
	return API_TYPE_OPENAI
}

func isStreamFinished(data string) bool {
	return strings.Index(data, StreamDone) != -1
}

func isContentTypeStream(header http.Header) bool {
	contentType := header.Get(c.HEADER_CONTENT_TYPE)
	cTypeParams := strings.Split(contentType, ";")
	cType := strings.TrimSpace(cTypeParams[0])
	return cType == c.CONTENT_TYPE_EVENT_STREAM
}

const (
	StreamDone       = "[DONE]"
	StreamDataPrefix = "data: "
)

var _ AIEngineStream = &OpenAI{}
