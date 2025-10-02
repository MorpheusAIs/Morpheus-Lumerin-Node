package proxyapi

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
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
	delete(audioRequest.Extra, "type")

	return audioRequest, nil
}

func processAudioSpeech(message []byte, sourceLog lib.ILogger) (*genericchatstorage.AudioSpeechRequest, error) {
	var unknownReq map[string]interface{}
	if err := json.Unmarshal(message, &unknownReq); err != nil {
		return nil, lib.WrapError(fmt.Errorf("failed to unmarshal request"), err)
	}

	if unknownReq["type"] != "audio_speech" {
		return nil, nil
	}

	var audioRequest *genericchatstorage.AudioSpeechRequest
	if err := json.Unmarshal(message, &audioRequest); err != nil {
		return nil, lib.WrapError(fmt.Errorf("failed to unmarshal audio speech request"), err)
	}
	delete(audioRequest.Extra, "type")

	return audioRequest, nil
}

func processEmbeddings(message []byte, sourceLog lib.ILogger) (*genericchatstorage.EmbeddingsRequest, error) {
	var unknownReq map[string]interface{}
	if err := json.Unmarshal(message, &unknownReq); err != nil {
		return nil, lib.WrapError(fmt.Errorf("failed to unmarshal request"), err)
	}

	if unknownReq["type"] != "embeddings" {
		return nil, nil // Not an embeddings request
	}

	var embedRequest *genericchatstorage.EmbeddingsRequest
	if err := json.Unmarshal(message, &embedRequest); err != nil {
		return nil, lib.WrapError(fmt.Errorf("failed to unmarshal embeddings request"), err)
	}
	delete(embedRequest.Extra, "type")

	return embedRequest, nil
}

// processChatRequest handles chat completion request processing
func processChatRequest(message []byte, sourceLog lib.ILogger) (*genericchatstorage.OpenAICompletionRequestExtra, error) {
	var chatRequest *genericchatstorage.OpenAICompletionRequestExtra
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

func (s *ProxyReceiver) SessionPrompt(ctx context.Context, requestID string, userPubKey string, payload []byte, sessionID common.Hash, sendResponse SendResponse, sourceLog lib.ILogger) (int, int, error) {
	// Store sendResponse function for later use in callback
	s.sendResponse = sendResponse

	// Get session
	session, err := s.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return handleError(err, "failed to get session", sourceLog)
	}

	// Process request based on type
	var audioTranscriptionReq *genericchatstorage.AudioTranscriptionRequest
	var chatReq *genericchatstorage.OpenAICompletionRequestExtra
	var audioSpeechReq *genericchatstorage.AudioSpeechRequest
	var embeddingsReq *genericchatstorage.EmbeddingsRequest

	// Try to process as audio transcription first
	audioTranscriptionReq, err = processAudioTranscription(payload, sourceLog)
	if err != nil {
		return handleError(err, "failed to process audio transcription", sourceLog)
	}

	// If not audio transcription, try audio speech
	if audioTranscriptionReq == nil {
		audioSpeechReq, err = processAudioSpeech(payload, sourceLog)
		if err != nil {
			return handleError(err, "failed to process audio speech", sourceLog)
		}
	}

	// If not audio nor speech, try embeddings
	if audioTranscriptionReq == nil && audioSpeechReq == nil {
		embeddingsReq, err = processEmbeddings(payload, sourceLog)
		if err != nil {
			return handleError(err, "failed to process embeddings", sourceLog)
		}
	}

	// If not audio, process as chat completion
	if audioTranscriptionReq == nil && audioSpeechReq == nil && embeddingsReq == nil {
		chatReq, err = processChatRequest(payload, sourceLog)
		if err != nil {
			return handleError(err, "failed to process chat request", sourceLog)
		}
	} else if audioTranscriptionReq != nil && audioTranscriptionReq.FilePath != "" {
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
	if audioTranscriptionReq != nil {
		sourceLog.Debugf("Processing audio transcription request")
		err = adapter.AudioTranscription(ctx, audioTranscriptionReq, cb)
	} else if audioSpeechReq != nil {
		sourceLog.Debugf("Processing audio speech request")
		err = adapter.AudioSpeech(ctx, audioSpeechReq, cb)
	} else if embeddingsReq != nil {
		sourceLog.Debugf("Processing embeddings request")
		err = adapter.Embeddings(ctx, embeddingsReq, cb)
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
