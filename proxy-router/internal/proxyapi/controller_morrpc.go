package proxyapi

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	m "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	msg "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	sessionrepo "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/session"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-playground/validator/v10"
)

type MORRPCController struct {
	service        *ProxyReceiver
	sessionRepo    *sessionrepo.SessionRepositoryCached
	validator      *validator.Validate
	sessionStorage *storages.SessionStorage
	morRpc         *m.MORRPCMessage
	prKey          lib.HexString
	streamManager  *StreamingSessionManager
	sessionSema    *SessionSemaphore // Limits to 1 concurrent request per session
}

type SendResponse func(*msg.RpcResponse) error

var (
	ErrUnknownMethod = fmt.Errorf("unknown method")
)

func NewMORRPCController(service *ProxyReceiver, validator *validator.Validate, sessionRepo *sessionrepo.SessionRepositoryCached, sessionStorage *storages.SessionStorage, prKey lib.HexString) *MORRPCController {
	c := &MORRPCController{
		service:        service,
		validator:      validator,
		sessionStorage: sessionStorage,
		sessionRepo:    sessionRepo,
		morRpc:         m.NewMorRpc(),
		prKey:          prKey,
		streamManager:  NewStreamingSessionManager(),
		sessionSema:    NewSessionSemaphore(),
	}

	return c
}

func (s *MORRPCController) Handle(ctx context.Context, msg m.RPCMessage, sourceLog lib.ILogger, sendResponse SendResponse) error {
	sourceLog.Debugf("received TCP message with method %s", msg.Method)
	switch msg.Method {
	case "network.ping":
		return s.networkPing(ctx, msg, sendResponse, sourceLog)
	case "session.request":
		return s.sessionRequest(ctx, msg, sendResponse, sourceLog)
	case "session.prompt":
		return s.sessionPrompt(ctx, msg, sendResponse, sourceLog)
	case "session.report":
		return s.sessionReport(ctx, msg, sendResponse, sourceLog)
	case "agent.call_tool":
		return s.callAgentTool(ctx, msg, sendResponse, sourceLog)
	case "agent.get_tools":
		return s.getAgentTools(ctx, msg, sendResponse, sourceLog)
	case "session.prompt.stream.start":
		return s.sessionPromptStreamStart(ctx, msg, sendResponse, sourceLog)
	case "session.prompt.stream.chunk":
		return s.sessionPromptStreamChunk(ctx, msg, sendResponse, sourceLog)
	case "session.prompt.stream.end":
		return s.sessionPromptStreamEnd(ctx, msg, sendResponse, sourceLog)
	default:
		return lib.WrapError(ErrUnknownMethod, fmt.Errorf("unknown method: %s", msg.Method))
	}
}

var (
	ErrValidation     = fmt.Errorf("request validation failed")
	ErrUnmarshal      = fmt.Errorf("failed to unmarshal request")
	ErrGenerateReport = fmt.Errorf("failed to generate report")
)

func (s *MORRPCController) networkPing(_ context.Context, msg m.RPCMessage, sendResponse SendResponse, sourceLog lib.ILogger) error {
	var req m.PingReq
	err := json.Unmarshal(msg.Params, &req)
	if err != nil {
		return lib.WrapError(ErrUnmarshal, err)
	}

	if err := s.validator.Struct(req); err != nil {
		return lib.WrapError(ErrValidation, err)
	}

	res, err := s.morRpc.PongResponce(msg.ID, s.prKey, req.Nonce)
	if err != nil {
		sourceLog.Error(err)
		return err
	}

	return sendResponse(res)
}

func (s *MORRPCController) sessionRequest(ctx context.Context, msg m.RPCMessage, sendResponse SendResponse, sourceLog lib.ILogger) error {
	var req m.SessionReq
	err := json.Unmarshal(msg.Params, &req)
	if err != nil {
		return lib.WrapError(ErrUnmarshal, err)
	}

	if err := s.validator.Struct(req); err != nil {
		return lib.WrapError(ErrValidation, err)
	}
	sig := req.Signature
	req.Signature = lib.HexString{}
	isValid := s.morRpc.VerifySignature(req, sig, req.Key, sourceLog)
	if !isValid {
		err := ErrInvalidSig
		sourceLog.Error(err)
		return err
	}

	res, err := s.service.SessionRequest(ctx, msg.ID, msg.ID, &req, sourceLog)
	if err != nil {
		sourceLog.Error(err)
		return err
	}

	return sendResponse(res)
}

func (s *MORRPCController) sessionPrompt(ctx context.Context, msg m.RPCMessage, sendResponse SendResponse, sourceLog lib.ILogger) error {
	var req m.SessionPromptReq
	err := json.Unmarshal(msg.Params, &req)
	if err != nil {
		return lib.WrapError(ErrUnmarshal, err)
	}

	if err := s.validator.Struct(req); err != nil {
		return lib.WrapError(ErrValidation, err)
	}

	sourceLog.Debugf("received prompt from session %s, timestamp: %d", req.SessionID, req.Timestamp)

	// Validate session exists and is not expired
	session, err := s.isSessionValid(ctx, req.SessionID, req.Timestamp)
	if err != nil {
		return err
	}

	user, err := s.sessionStorage.GetUser(session.UserAddr().Hex())
	if err != nil {
		return fmt.Errorf("error reading user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	pubKeyHex, err := lib.StringToHexString(user.PubKey)
	if err != nil {
		return fmt.Errorf("invalid pubkey %s", err)
	}

	sig := req.Signature
	req.Signature = lib.HexString{}

	isValid := s.morRpc.VerifySignature(req, sig, pubKeyHex, sourceLog)
	if !isValid {
		err := fmt.Errorf("invalid signature")
		sourceLog.Error(err)
		return err
	}

	// Acquire session semaphore to ensure only 1 concurrent request per session
	// This will block if another request is already being processed for this session
	sourceLog.Debugf("acquiring session semaphore for session %s", req.SessionID.Hex())
	if err := s.sessionSema.Acquire(ctx, req.SessionID); err != nil {
		return fmt.Errorf("request cancelled while waiting in queue: %w", err)
	}
	defer s.sessionSema.Release(req.SessionID)
	sourceLog.Debugf("acquired session semaphore for session %s", req.SessionID.Hex())

	now := time.Now().Unix()
	ttftMs, inputTokens, outputTokens, err := s.service.SessionPrompt(ctx, msg.ID, user.PubKey, []byte(req.Message), req.SessionID, sendResponse, sourceLog)
	if err != nil {
		sourceLog.Error(err)
		return err
	}

	requestDuration := int(time.Now().Unix() - now)
	if requestDuration == 0 {
		requestDuration = 1
	}

	tpsScaled1000 := outputTokens * 1000 / requestDuration
	session.AddStats(tpsScaled1000, ttftMs, inputTokens, outputTokens)

	err = s.sessionRepo.SaveSession(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to save session %s", err)
	}
	return nil
}

func (s *MORRPCController) sessionReport(ctx context.Context, msg m.RPCMessage, sendResponse SendResponse, sourceLog lib.ILogger) error {
	var req m.SessionReportReq
	err := json.Unmarshal(msg.Params, &req)
	if err != nil {
		err := lib.WrapError(ErrUnmarshal, err)
		sourceLog.Error(err)
		return err
	}

	if err := s.validator.Struct(req); err != nil {
		err := lib.WrapError(ErrValidation, err)
		sourceLog.Error(err)
		return err
	}

	sessionID := req.Message
	sourceLog.Debugf("Requested report from session %s, timestamp: %s", sessionID, req.Timestamp)
	session, err := s.sessionStorage.GetSession(sessionID)
	if err != nil {
		sourceLog.Errorf("error reading session: %s", err)
		return fmt.Errorf("error reading session: %w", err)
	}
	if session == nil {
		err := fmt.Errorf("session not found")
		sourceLog.Error(err)
		return err
	}

	user, err := s.sessionStorage.GetUser(session.UserAddr)
	if err != nil {
		sourceLog.Errorf("error reading user: %s", err)
		return fmt.Errorf("error reading user: %w", err)
	}
	if user == nil {
		err := fmt.Errorf("user not found")
		sourceLog.Error(err)
		return err
	}
	pubKeyHex, err := lib.StringToHexString(user.PubKey)
	if err != nil {
		sourceLog.Error(err)
		return err
	}

	sig := req.Signature
	req.Signature = lib.HexString{}

	isValid := s.morRpc.VerifySignature(req, sig, pubKeyHex, sourceLog)
	if !isValid {
		err := fmt.Errorf("invalid signature")
		sourceLog.Error(err)
		return err
	}

	res, err := s.service.SessionReport(ctx, msg.ID, msg.ID, session, sourceLog)
	if err != nil {
		sourceLog.Error(err)
		return ErrGenerateReport
	}

	return sendResponse(res)
}

func (s *MORRPCController) callAgentTool(ctx context.Context, msg m.RPCMessage, sendResponse SendResponse, sourceLog lib.ILogger) error {
	var req m.CallAgentToolReq
	err := json.Unmarshal(msg.Params, &req)
	if err != nil {
		return lib.WrapError(ErrUnmarshal, err)
	}

	if err := s.validator.Struct(req); err != nil {
		return lib.WrapError(ErrValidation, err)
	}

	// Validate session exists and is not expired
	session, err := s.isSessionValid(ctx, req.SessionID, req.Timestamp)
	if err != nil {
		return err
	}

	user, err := s.sessionStorage.GetUser(session.UserAddr().Hex())
	if err != nil {
		return fmt.Errorf("error reading user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	pubKeyHex, err := lib.StringToHexString(user.PubKey)
	if err != nil {
		return fmt.Errorf("invalid pubkey %s", err)
	}

	sig := req.Signature
	req.Signature = lib.HexString{}

	isValid := s.morRpc.VerifySignature(req, sig, pubKeyHex, sourceLog)
	if !isValid {
		err := fmt.Errorf("invalid signature")
		sourceLog.Error(err)
		return err
	}

	res, err := s.service.CallAgentTool(ctx, msg.ID, msg.ID, user.PubKey, &req, sourceLog)
	if err != nil {
		sourceLog.Error(err)
		return err
	}

	return sendResponse(res)
}

func (s *MORRPCController) getAgentTools(ctx context.Context, msg m.RPCMessage, sendResponse SendResponse, sourceLog lib.ILogger) error {
	var req m.GetAgentToolsReq
	err := json.Unmarshal(msg.Params, &req)
	if err != nil {
		return lib.WrapError(ErrUnmarshal, err)
	}

	if err := s.validator.Struct(req); err != nil {
		return lib.WrapError(ErrValidation, err)
	}

	// Validate session exists and is not expired
	session, err := s.isSessionValid(ctx, req.SessionID, req.Timestamp)
	if err != nil {
		return err
	}

	user, err := s.sessionStorage.GetUser(session.UserAddr().Hex())
	if err != nil {
		return fmt.Errorf("error reading user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	pubKeyHex, err := lib.StringToHexString(user.PubKey)
	if err != nil {
		return fmt.Errorf("invalid pubkey %s", err)
	}

	sig := req.Signature
	req.Signature = lib.HexString{}

	isValid := s.morRpc.VerifySignature(req, sig, pubKeyHex, sourceLog)
	if !isValid {
		err := fmt.Errorf("invalid signature")
		sourceLog.Error(err)
		return err
	}

	res, err := s.service.GetAgentTools(ctx, msg.ID, msg.ID, user.PubKey, &req, sourceLog)
	if err != nil {
		sourceLog.Error(err)
		return err
	}

	return sendResponse(res)
}

func (s *MORRPCController) sessionPromptStreamStart(ctx context.Context, msg m.RPCMessage, sendResponse SendResponse, sourceLog lib.ILogger) error {
	var req m.SessionPromptStreamStartReq
	err := json.Unmarshal(msg.Params, &req)
	if err != nil {
		return lib.WrapError(ErrUnmarshal, err)
	}

	if err := s.validator.Struct(req); err != nil {
		return lib.WrapError(ErrValidation, err)
	}

	// Validate session exists and is not expired
	session, err := s.isSessionValid(ctx, req.SessionID, req.Timestamp)
	if err != nil {
		return err
	}

	// Verify user signature
	user, err := s.sessionStorage.GetUser(session.UserAddr().Hex())
	if err != nil {
		return fmt.Errorf("error reading user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	pubKeyHex, err := lib.StringToHexString(user.PubKey)
	if err != nil {
		return fmt.Errorf("invalid pubkey %s", err)
	}

	sig := req.Signature
	req.Signature = lib.HexString{}

	isValid := s.morRpc.VerifySignature(req, sig, pubKeyHex, sourceLog)
	if !isValid {
		err := fmt.Errorf("invalid signature")
		sourceLog.Error(err)
		return err
	}

	// Clean up expired sessions
	s.streamManager.CleanupExpiredSessions()

	// Check if stream already exists
	if _, exists := s.streamManager.GetSession(req.StreamID); exists {
		return fmt.Errorf("stream with ID %s already exists", req.StreamID)
	}

	// Create streaming session
	_, err = s.streamManager.CreateSession(req.StreamID, req.SessionID.Hex(), req.TotalChunks, req.FileSize, req.ContentType)
	if err != nil {
		return fmt.Errorf("failed to create streaming session: %s", err)
	}

	sourceLog.Debugf("started audio streaming session %s for session %s, total chunks: %d, file size: %d",
		req.StreamID, req.SessionID.Hex(), req.TotalChunks, req.FileSize)

	// Send success response
	res, err := s.morRpc.SessionPromptStreamStartResponse(req.StreamID, "started", s.prKey, msg.ID)
	if err != nil {
		sourceLog.Error(err)
		return err
	}

	return sendResponse(res)
}

func (s *MORRPCController) sessionPromptStreamChunk(ctx context.Context, msg m.RPCMessage, sendResponse SendResponse, sourceLog lib.ILogger) error {
	var req m.SessionPromptStreamChunkReq
	err := json.Unmarshal(msg.Params, &req)
	if err != nil {
		return lib.WrapError(ErrUnmarshal, err)
	}

	if err := s.validator.Struct(req); err != nil {
		return lib.WrapError(ErrValidation, err)
	}

	// Validate session exists and is not expired
	session, err := s.isSessionValid(ctx, req.SessionID, req.Timestamp)
	if err != nil {
		return err
	}

	// Verify user signature
	user, err := s.sessionStorage.GetUser(session.UserAddr().Hex())
	if err != nil {
		return fmt.Errorf("error reading user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	pubKeyHex, err := lib.StringToHexString(user.PubKey)
	if err != nil {
		return fmt.Errorf("invalid pubkey %s", err)
	}

	sig := req.Signature
	req.Signature = lib.HexString{}

	isValid := s.morRpc.VerifySignature(req, sig, pubKeyHex, sourceLog)
	if !isValid {
		err := fmt.Errorf("invalid signature")
		sourceLog.Error(err)
		return err
	}

	// Get streaming session
	streamSession, exists := s.streamManager.GetSession(req.StreamID)
	if !exists {
		return fmt.Errorf("streaming session %s not found", req.StreamID)
	}

	// Validate chunk index
	if req.ChunkIndex != streamSession.ChunkCount {
		return fmt.Errorf("expected chunk index %d, got %d", streamSession.ChunkCount, req.ChunkIndex)
	}

	// Decode and append chunk data
	chunkData, err := base64.StdEncoding.DecodeString(req.ChunkData)
	if err != nil {
		return fmt.Errorf("failed to decode chunk data: %s", err)
	}

	// Append chunk data to temp file
	file, err := os.OpenFile(streamSession.TempFilePath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open temp file for writing: %s", err)
	}
	defer file.Close()

	_, err = file.Write(chunkData)
	if err != nil {
		return fmt.Errorf("failed to write chunk data to temp file: %s", err)
	}

	streamSession.ChunkCount++
	streamSession.LastActivity = time.Now()

	sourceLog.Debugf("received chunk %d/%d for stream %s, size: %d bytes",
		req.ChunkIndex+1, streamSession.TotalChunks, req.StreamID, len(chunkData))

	// Send success response
	res, err := s.morRpc.SessionPromptStreamChunkResponse(req.StreamID, req.ChunkIndex, "received", s.prKey, msg.ID)
	if err != nil {
		sourceLog.Error(err)
		return err
	}

	return sendResponse(res)
}

func (s *MORRPCController) sessionPromptStreamEnd(ctx context.Context, msg m.RPCMessage, sendResponse SendResponse, sourceLog lib.ILogger) error {
	var req m.SessionPromptStreamEndReq
	err := json.Unmarshal(msg.Params, &req)
	if err != nil {
		return lib.WrapError(ErrUnmarshal, err)
	}

	if err := s.validator.Struct(req); err != nil {
		return lib.WrapError(ErrValidation, err)
	}

	// Validate session exists and is not expired
	session, err := s.isSessionValid(ctx, req.SessionID, req.Timestamp)
	if err != nil {
		return err
	}

	// Verify user signature
	user, err := s.sessionStorage.GetUser(session.UserAddr().Hex())
	if err != nil {
		return fmt.Errorf("error reading user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	pubKeyHex, err := lib.StringToHexString(user.PubKey)
	if err != nil {
		return fmt.Errorf("invalid pubkey %s", err)
	}

	sig := req.Signature
	req.Signature = lib.HexString{}

	isValid := s.morRpc.VerifySignature(req, sig, pubKeyHex, sourceLog)
	if !isValid {
		err := fmt.Errorf("invalid signature")
		sourceLog.Error(err)
		return err
	}

	// Get streaming session
	streamSession, exists := s.streamManager.GetSession(req.StreamID)
	if !exists {
		return fmt.Errorf("streaming session %s not found", req.StreamID)
	}

	// Validate all chunks received
	if streamSession.ChunkCount != streamSession.TotalChunks {
		return fmt.Errorf("incomplete stream: received %d chunks, expected %d", streamSession.ChunkCount, streamSession.TotalChunks)
	}

	sourceLog.Debugf("completed audio streaming session %s, temp file: %s", req.StreamID, streamSession.TempFilePath)

	// Acquire session semaphore to ensure only 1 concurrent request per session
	sourceLog.Debugf("acquiring session semaphore for session %s (stream end)", req.SessionID.Hex())
	if err := s.sessionSema.Acquire(ctx, req.SessionID); err != nil {
		return fmt.Errorf("request cancelled while waiting in queue: %w", err)
	}
	defer s.sessionSema.Release(req.SessionID)
	sourceLog.Debugf("acquired session semaphore for session %s (stream end)", req.SessionID.Hex())

	// Process the complete audio file
	err = s.processStreamedAudioFile(ctx, session, streamSession.TempFilePath, req.AudioRequestParam, msg.ID, user.PubKey, req.SessionID, sendResponse, sourceLog)
	if err != nil {
		// Clean up streaming session on error
		s.streamManager.RemoveSession(req.StreamID)
		sourceLog.Errorf("failed to process streamed audio file: %s", err)
		return err
	}

	// Clean up streaming session after successful processing
	s.streamManager.RemoveSession(req.StreamID)

	sourceLog.Debugf("processed streamed audio file for session %s", req.SessionID.Hex())

	// The response is sent by the streaming callback, so we don't send one here
	return nil
}

func (s *MORRPCController) processStreamedAudioFile(ctx context.Context, session *sessionrepo.SessionModel, tempFilePath string, audioRequestParam string, requestID string, userPubKey string, sessionID common.Hash, sendResponse SendResponse, sourceLog lib.ILogger) error {
	// Parse the audio request parameters
	var audioParams map[string]interface{}
	err := json.Unmarshal([]byte(audioRequestParam), &audioParams)
	if err != nil {
		return fmt.Errorf("failed to parse audio request parameters: %s", err)
	}

	// Add the file path to the parameters
	audioParams["FilePath"] = tempFilePath

	payload, err := json.Marshal(audioParams)
	if err != nil {
		return fmt.Errorf("failed to serialize audio request parameters: %s", err)
	}

	now := time.Now().Unix()
	ttftMs, inputTokens, outputTokens, err := s.service.SessionPrompt(ctx, requestID, userPubKey, payload, sessionID, sendResponse, sourceLog)
	if err != nil {
		sourceLog.Error(err)
		return err
	}

	requestDuration := int(time.Now().Unix() - now)
	if requestDuration == 0 {
		requestDuration = 1
	}

	tpsScaled1000 := outputTokens * 1000 / requestDuration
	session.AddStats(tpsScaled1000, ttftMs, inputTokens, outputTokens)

	err = s.sessionRepo.SaveSession(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to save session %s", err)
	}
	return nil
}

// Validate session exists and is not expired
func (s *MORRPCController) isSessionValid(ctx context.Context, sessionID common.Hash, reqTimestamp uint64) (*sessionrepo.SessionModel, error) {
	session, err := s.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("session cannot be loaded %s", err)
	}

	isSessionExpired := session.EndsAt().Uint64()*1000 < uint64(reqTimestamp)
	if isSessionExpired {
		return nil, fmt.Errorf("session expired")
	}
	return session, nil
}
