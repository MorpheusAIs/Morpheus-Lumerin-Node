package proxyapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	m "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	msg "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	sessionrepo "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/session"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sashabaranov/go-openai"
)

type ProxyReceiver struct {
	privateKeyHex     lib.HexString
	publicKeyHex      lib.HexString
	chainID           *big.Int
	morRpc            *m.MORRPCMessage
	sessionStorage    *storages.SessionStorage
	aiEngine          *aiengine.AiEngine
	modelConfigLoader *config.ModelConfigLoader
	service           BidGetter
	sessionRepo       *sessionrepo.SessionRepositoryCached
	sendResponse      SendResponse
}

func NewProxyReceiver(privateKeyHex, publicKeyHex lib.HexString, sessionStorage *storages.SessionStorage, aiEngine *aiengine.AiEngine, chainID *big.Int, modelConfigLoader *config.ModelConfigLoader, blockchainService BidGetter, sessionRepo *sessionrepo.SessionRepositoryCached) *ProxyReceiver {
	return &ProxyReceiver{
		privateKeyHex:     privateKeyHex,
		publicKeyHex:      publicKeyHex,
		morRpc:            m.NewMorRpc(),
		aiEngine:          aiEngine,
		chainID:           chainID,
		modelConfigLoader: modelConfigLoader,
		service:           blockchainService,
		sessionStorage:    sessionStorage,
		sessionRepo:       sessionRepo,
	}
}

// handleSessionError is a helper function to log errors and return consistent output
func handleError(err error, message string, sourceLog lib.ILogger) (int, int, error) {
	wrappedErr := lib.WrapError(fmt.Errorf(message), err)
	sourceLog.Error(wrappedErr)
	return 0, 0, wrappedErr
}

// processAudioTranscription handles audio transcription request processing
func processAudioTranscription(message []byte, sourceLog lib.ILogger) (*genericchatstorage.AudioTranscriptionRequest, error) {
	var unknownReq map[string]interface{}
	if err := json.Unmarshal(message, &unknownReq); err != nil {
		return nil, lib.WrapError(fmt.Errorf("failed to unmarshal request"), err)
	}

	if unknownReq["type"] != "audio_transcription" {
		return nil, nil // Not an audio transcription request
	}

	var audioRequest *genericchatstorage.AudioTranscriptionRequest
	if err := json.Unmarshal(message, &audioRequest); err != nil {
		return nil, lib.WrapError(fmt.Errorf("failed to unmarshal audio request"), err)
	}

	// Create a temporary file for audio
	tempFilePath, err := createAudioTempFile(unknownReq["base64Audio"].(string), sourceLog)
	if err != nil {
		return nil, err
	}

	audioRequest.FilePath = tempFilePath
	return audioRequest, nil
}

// createAudioTempFile creates a temporary file with decoded audio data
func createAudioTempFile(base64Audio string, sourceLog lib.ILogger) (string, error) {
	// Decode and write audio
	audioBytes, err := base64.StdEncoding.DecodeString(base64Audio)
	if err != nil {
		return "", lib.WrapError(fmt.Errorf("failed to decode base64 audio"), err)
	}

	tempDir := os.TempDir()
	contentType := http.DetectContentType(audioBytes)
	
	audioExtensions := map[string]string{
		"audio/mpeg":       ".mp3",
		"audio/mp3":        ".mp3",
		"audio/wav":        ".wav",
		"audio/wave":       ".wav",
		"audio/x-wav":      ".wav",
		"audio/vnd.wave":   ".wav",
		"audio/ogg":        ".ogg",
		"audio/flac":       ".flac",
		"audio/aac":        ".aac",
		"audio/mp4":        ".m4a",
		"audio/x-m4a":      ".m4a",
		"audio/webm":       ".webm",
		"audio/opus":       ".opus",
		"audio/x-ms-wma":   ".wma",
		"audio/amr":        ".amr",
		"audio/3gpp":       ".3gp",
		"audio/x-aiff":     ".aiff",
		"audio/aiff":       ".aiff",
	}
	
	extension, exists := audioExtensions[contentType]
	if !exists {
		extension = detectAudioExtensionBySignature(audioBytes)
		fmt.Println("Detected extension by signature:", extension)
		if extension == "" {
			extension = ".mp3"
		}
	}
	
	fmt.Println("Detected content type:", contentType)
	fmt.Println("Using extension:", extension)
	
	tempFilePath := filepath.Join(tempDir, fmt.Sprintf("%d%s", time.Now().UnixNano(), extension))
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return "", lib.WrapError(fmt.Errorf("failed to create temp file"), err)
	}
	defer tempFile.Close()

	if _, err = tempFile.Write(audioBytes); err != nil {
		return "", lib.WrapError(fmt.Errorf("failed to write audio to temp file"), err)
	}

	return tempFilePath, nil
}

// detectAudioExtensionBySignature detects audio file extension by examining file signature (magic bytes)
func detectAudioExtensionBySignature(data []byte) string {
	if len(data) < 12 {
		return ""
	}
	
	// Check for various audio file signatures
	switch {
	case len(data) >= 4 && string(data[0:3]) == "ID3": // MP3 with ID3 tag
		return ".mp3"
	case len(data) >= 4 && data[0] == 0xFF && (data[1]&0xE0) == 0xE0: // MP3 frame header
		return ".mp3"
	case len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WAVE": // WAV
		return ".wav"
	case len(data) >= 4 && string(data[0:4]) == "OggS": // OGG
		return ".ogg"
	case len(data) >= 4 && string(data[0:4]) == "fLaC": // FLAC
		return ".flac"
	case len(data) >= 8 && string(data[4:8]) == "ftyp": // MP4/M4A container
		if len(data) >= 12 {
			subtype := string(data[8:12])
			if subtype == "M4A " || subtype == "mp41" || subtype == "mp42" {
				return ".m4a"
			}
		}
		return ".mp4"
	case len(data) >= 12 && string(data[0:4]) == "FORM" && string(data[8:12]) == "AIFF": // AIFF
		return ".aiff"
	case len(data) >= 6 && string(data[0:6]) == "#!AMR\n": // AMR
		return ".amr"
	}
	
	return ""
}

// processChatRequest handles chat completion request processing
func processChatRequest(message []byte, sourceLog lib.ILogger) (*openai.ChatCompletionRequest, error) {
	var chatRequest *openai.ChatCompletionRequest
	if err := json.Unmarshal(message, &chatRequest); err != nil {
		return nil, lib.WrapError(fmt.Errorf("failed to unmarshal chat request"), err)
	}
	return chatRequest, nil
}

// createCompletionCallback creates a callback function for handling completion chunks
func (s *ProxyReceiver) createCompletionCallback(
	ctx context.Context,
	startTime int64,
	userPubKey string,
	requestID string,
	sourceLog lib.ILogger,
	ttftMs *int,
	totalTokens *int,
) genericchatstorage.CompletionCallback {
	return func(ctx context.Context, completion genericchatstorage.Chunk, aiEngineErrorResponse *genericchatstorage.AiEngineErrorResponse) error {
		if aiEngineErrorResponse != nil {
			marshalledResponse, err := json.Marshal(aiEngineErrorResponse)
			if err != nil {
				return err
			}
			encryptedResponse, err := lib.EncryptString(string(marshalledResponse), lib.RemoveHexPrefix(userPubKey))
			if err != nil {
				return err
			}

			r, err := s.morRpc.ResponseError(
				encryptedResponse,
				s.privateKeyHex,
				requestID,
			)
			if err != nil {
				err := lib.WrapError(fmt.Errorf("failed to create response"), err)
				sourceLog.Error(err)
				return err
			}
			return s.sendResponse(r)
		}

		*totalTokens += completion.Tokens()

		if *ttftMs == 0 {
			*ttftMs = int(time.Now().UnixMilli() - startTime)
		}

		marshalledResponse, err := json.Marshal(completion.Data())
		if err != nil {
			return err
		}

		encryptedResponse, err := lib.EncryptString(string(marshalledResponse), lib.RemoveHexPrefix(userPubKey))
		if err != nil {
			return err
		}

		// Send response
		r, err := s.morRpc.SessionPromptResponse(
			encryptedResponse,
			s.privateKeyHex,
			requestID,
		)
		if err != nil {
			wrappedErr := lib.WrapError(fmt.Errorf("failed to create response"), err)
			sourceLog.Error(wrappedErr)
			return wrappedErr
		}

		return s.sendResponse(r)
	}
}

// recordActivity records the session activity
func (s *ProxyReceiver) recordActivity(ctx context.Context, session *sessionrepo.SessionModel, startTime int64, sourceLog lib.ILogger) {
	activity := storages.PromptActivity{
		SessionID: session.ID().Hex(),
		StartTime: startTime,
		EndTime:   time.Now().Unix(),
	}

	if err := s.sessionStorage.AddActivity(session.ModelID().Hex(), &activity); err != nil {
		sourceLog.Warnf("failed to store activity: %s", err)
	}
}

func (s *ProxyReceiver) SessionPrompt(ctx context.Context, requestID string, userPubKey string, rq *m.SessionPromptReq, sendResponse SendResponse, sourceLog lib.ILogger) (int, int, error) {
	// Store sendResponse function for later use in callback
	s.sendResponse = sendResponse

	// Get session
	session, err := s.sessionRepo.GetSession(ctx, rq.SessionID)
	if err != nil {
		return handleError(err, "failed to get session", sourceLog)
	}

	// Process request based on type
	var audioTranscriptionReq *genericchatstorage.AudioTranscriptionRequest
	var chatReq *openai.ChatCompletionRequest

	// Try to process as audio transcription first
	audioTranscriptionReq, err = processAudioTranscription([]byte(rq.Message), sourceLog)
	if err != nil {
		return handleError(err, "failed to process audio transcription", sourceLog)
	}

	// If not audio, process as chat completion
	if audioTranscriptionReq == nil {
		chatReq, err = processChatRequest([]byte(rq.Message), sourceLog)
		if err != nil {
			return handleError(err, "failed to process chat request", sourceLog)
		}
	} else {
		defer os.Remove(audioTranscriptionReq.FilePath)
	}

	// Start timing and get adapter
	startTime := time.Now().UnixMilli()
	ttftMs, totalTokens := 0, 0

	adapter, err := s.aiEngine.GetAdapter(ctx, common.Hash{}, session.ModelID(), common.Hash{}, false, false)
	if err != nil {
		return handleError(err, "failed to get adapter", sourceLog)
	}

	// Create completion callback
	cb := s.createCompletionCallback(ctx, startTime, userPubKey, requestID, sourceLog, &ttftMs, &totalTokens)

	// Process request with appropriate adapter method
	if audioTranscriptionReq != nil && audioTranscriptionReq.FilePath != "" {
		sourceLog.Debugf("Processing audio transcription request")
		err = adapter.AudioTranscription(ctx, audioTranscriptionReq, "", cb)
	} else {
		sourceLog.Debugf("Processing chat completion request")
		err = adapter.Prompt(ctx, chatReq, cb)
	}

	if err != nil {
		return handleError(err, "failed to prompt", sourceLog)
	}

	// Record activity
	s.recordActivity(ctx, session, startTime, sourceLog)

	return ttftMs, totalTokens, nil
}

func (s *ProxyReceiver) SessionRequest(ctx context.Context, msgID string, reqID string, req *m.SessionReq, log lib.ILogger) (*msg.RpcResponse, error) {
	log.Debugf("Received session request from %s, timestamp: %s", req.User, req.Timestamp)

	bid, err := s.service.GetBidByID(ctx, req.BidID)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to get bid"), err)
		log.Error(err)
		return nil, err
	}

	modelID := bid.ModelAgentId.String()
	modelConfig := s.modelConfigLoader.ModelConfigFromID(modelID)
	capacityManager := CreateCapacityManager(modelConfig, s.sessionStorage, log)

	hasCapacity := capacityManager.HasCapacity(modelID)
	if !hasCapacity {
		err := fmt.Errorf("no capacity")
		log.Error(err)
		return nil, err
	}

	// Send response
	response, err := s.morRpc.InitiateSessionResponse(
		s.publicKeyHex,
		req.User,
		req.BidID,
		s.privateKeyHex,
		reqID,
		s.chainID,
	)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to create response"), err)
		log.Error(err)
		return nil, err
	}

	user := storages.User{
		Addr:   req.User.Hex(),
		PubKey: req.Key.Hex(),
	}

	err = s.sessionStorage.AddUser(&user)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed store user"), err)
		log.Error(err)
		return nil, err
	}

	return response, nil
}

func (s *ProxyReceiver) SessionReport(ctx context.Context, msgID string, reqID string, session *storages.Session, sourceLog lib.ILogger) (*msg.RpcResponse, error) {
	sourceLog.Debugf("received session report request for %s", session.Id)

	tps := 0
	ttft := 0
	for _, tpsVal := range session.TPSScaled1000Arr {
		tps += tpsVal
	}
	for _, ttftVal := range session.TTFTMsArr {
		ttft += ttftVal
	}

	if len(session.TPSScaled1000Arr) != 0 {
		tps /= len(session.TPSScaled1000Arr)
	}
	if len(session.TTFTMsArr) != 0 {
		ttft /= len(session.TTFTMsArr)
	}

	response, err := s.morRpc.SessionReportResponse(
		uint32(tps),
		uint32(ttft),
		common.HexToHash(session.Id),
		s.privateKeyHex,
		reqID,
		s.chainID,
	)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to create response"), err)
		sourceLog.Error(err)
		return nil, err
	}

	return response, nil
}

func (s *ProxyReceiver) CallAgentTool(ctx context.Context, msgID string, reqID string, userPubKey string, req *m.CallAgentToolReq, sourceLog lib.ILogger) (*msg.RpcResponse, error) {
	sourceLog.Debugf("received call agent tool request for %s", req.SessionID)

	var input map[string]interface{}

	err := json.Unmarshal([]byte(req.Message), &input)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to unmarshal prompt"), err)
		sourceLog.Error(err)
		return nil, err
	}

	session, err := s.sessionRepo.GetSession(ctx, req.SessionID)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to get session"), err)
		sourceLog.Error(err)
		return nil, err
	}

	result, err := s.aiEngine.CallAgentTool(ctx, common.Hash{}, session.ModelID(), req.ToolName, input)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to call agent tool"), err)
		sourceLog.Error(err)
		return nil, err
	}

	marshalledResult, err := json.Marshal(result)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to marshal result"), err)
		sourceLog.Error(err)
		return nil, err
	}

	encryptedResponse, err := lib.EncryptString(string(marshalledResult), lib.RemoveHexPrefix(userPubKey))
	if err != nil {
		return nil, err
	}

	response, err := s.morRpc.CallAgentToolResponse(
		string(encryptedResponse),
		s.privateKeyHex,
		reqID,
	)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to create response"), err)
		sourceLog.Error(err)
		return nil, err
	}

	return response, nil
}

func (s *ProxyReceiver) GetAgentTools(ctx context.Context, msgID string, reqID string, userPubKey string, req *m.GetAgentToolsReq, sourceLog lib.ILogger) (*msg.RpcResponse, error) {
	sourceLog.Debugf("received get agent tools request for %s", req.SessionID)

	session, err := s.sessionRepo.GetSession(ctx, req.SessionID)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to get session"), err)
		sourceLog.Error(err)
		return nil, err
	}

	tools, err := s.aiEngine.GetAgentTools(ctx, common.Hash{}, session.ModelID())
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to get agent tools"), err)
		sourceLog.Error(err)
		return nil, err
	}

	marshalledTools, err := json.Marshal(tools)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to marshal tools"), err)
		sourceLog.Error(err)
		return nil, err
	}

	encryptedResponse, err := lib.EncryptString(string(marshalledTools), lib.RemoveHexPrefix(userPubKey))
	if err != nil {
		return nil, err
	}

	response, err := s.morRpc.GetAgentToolsResponse(
		encryptedResponse,
		s.privateKeyHex,
		reqID,
	)
	if err != nil {
		err := lib.WrapError(fmt.Errorf("failed to create response"), err)
		sourceLog.Error(err)
		return nil, err
	}

	return response, nil
}
