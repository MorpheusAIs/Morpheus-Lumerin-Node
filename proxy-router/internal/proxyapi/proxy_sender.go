package proxyapi

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	gcs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/chatstorage/genericchatstorage"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	msgs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	sessionrepo "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/session"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin/binding"
	"github.com/sashabaranov/go-openai"
)

var (
	ErrMissingPrKey     = fmt.Errorf("missing private key")
	ErrCreateReq        = fmt.Errorf("failed to create request")
	ErrProvider         = fmt.Errorf("provider request failed")
	ErrInvalidSig       = fmt.Errorf("received invalid signature from provider")
	ErrFailedStore      = fmt.Errorf("failed store user")
	ErrInvalidResponse  = fmt.Errorf("invalid response")
	ErrResponseErr      = fmt.Errorf("response error")
	ErrDecrFailed       = fmt.Errorf("failed to decrypt ai response chunk")
	ErrMasrshalFailed   = fmt.Errorf("failed to marshal response")
	ErrDecode           = fmt.Errorf("failed to decode response")
	ErrSessionNotFound  = fmt.Errorf("session not found")
	ErrSessionExpired   = fmt.Errorf("session expired")
	ErrProviderNotFound = fmt.Errorf("provider not found")
	ErrEmpty            = fmt.Errorf("empty result and no error")
	ErrConnectProvider  = fmt.Errorf("failed to connect to provider")
	ErrWriteProvider    = fmt.Errorf("failed to write to provider")
)

const (
	TimeoutPingDefault = 5 * time.Second

	SENDER_AUDIO_STREAM_CHUNK_SIZE = 1 * 1024 * 1024 // 1mb chunks for audio streaming
)

type ProxyServiceSender struct {
	chainID        *big.Int
	privateKey     interfaces.PrKeyProvider
	logStorage     *lib.Collection[*interfaces.LogStorage]
	sessionStorage *storages.SessionStorage
	sessionRepo    *sessionrepo.SessionRepositoryCached
	morRPC         *msgs.MORRPCMessage
	sessionService SessionService
	log            lib.ILogger
}

func NewProxySender(chainID *big.Int, privateKey interfaces.PrKeyProvider, logStorage *lib.Collection[*interfaces.LogStorage], sessionStorage *storages.SessionStorage, sessionRepo *sessionrepo.SessionRepositoryCached, log lib.ILogger) *ProxyServiceSender {
	return &ProxyServiceSender{
		chainID:        chainID,
		privateKey:     privateKey,
		logStorage:     logStorage,
		sessionStorage: sessionStorage,
		sessionRepo:    sessionRepo,
		morRPC:         msgs.NewMorRpc(),
		log:            log,
	}
}

func (p *ProxyServiceSender) SetSessionService(service SessionService) {
	p.sessionService = service
}

func (p *ProxyServiceSender) Ping(ctx context.Context, providerURL string, providerAddr common.Address) (time.Duration, error) {
	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return 0, ErrMissingPrKey
	}

	// check if context has timeout set
	if _, ok := ctx.Deadline(); !ok {
		subCtx, cancel := context.WithTimeout(ctx, TimeoutPingDefault)
		defer cancel()
		ctx = subCtx
	}

	nonce := make([]byte, 8)
	_, err = rand.Read(nonce)
	if err != nil {
		return 0, lib.WrapError(ErrCreateReq, err)
	}

	msg, err := p.morRPC.PingRequest("0", prKey, nonce)
	if err != nil {
		return 0, lib.WrapError(ErrCreateReq, err)
	}

	reqStartTime := time.Now()
	res, code, err := p.rpcRequest(providerURL, msg)
	if err != nil {
		return 0, lib.WrapError(ErrProvider, fmt.Errorf("code: %d, msg: %v, error: %s", code, res, err))
	}
	pingDuration := time.Since(reqStartTime)

	var typedMsg *msgs.PongRes
	err = json.Unmarshal(*res.Result, &typedMsg)
	if err != nil {
		return pingDuration, lib.WrapError(ErrInvalidResponse, fmt.Errorf("expected PongRes, got %s", res.Result))
	}

	err = binding.Validator.ValidateStruct(typedMsg)
	if err != nil {
		return pingDuration, lib.WrapError(ErrInvalidResponse, err)
	}

	signature := typedMsg.Signature
	typedMsg.Signature = lib.HexString{}

	if !p.morRPC.VerifySignatureAddr(typedMsg, signature, providerAddr, p.log) {
		return pingDuration, ErrInvalidSig
	}

	return pingDuration, nil
}

func (p *ProxyServiceSender) InitiateSession(ctx context.Context, user common.Address, provider common.Address, spend *big.Int, bidID common.Hash, providerURL string) (*msgs.SessionRes, error) {
	requestID := "1"

	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return nil, ErrMissingPrKey
	}

	initiateSessionRequest, err := p.morRPC.InitiateSessionRequest(user, provider, spend, bidID, prKey, requestID)
	if err != nil {
		return nil, lib.WrapError(ErrCreateReq, err)
	}

	msg, code, err := p.rpcRequest(providerURL, initiateSessionRequest)
	if err != nil {
		return nil, lib.WrapError(ErrProvider, fmt.Errorf("code: %d, msg: %v, error: %s", code, msg, err))
	}

	if msg.Error != nil {
		// TODO: verify signature
		return nil, lib.WrapError(ErrResponseErr, fmt.Errorf("error: %v, result: %v", msg.Error.Message, msg.Error.Data))
	}
	if msg.Result == nil {
		return nil, lib.WrapError(ErrInvalidResponse, ErrEmpty)
	}

	var typedMsg *msgs.SessionRes
	err = json.Unmarshal(*msg.Result, &typedMsg)
	if err != nil {
		return nil, lib.WrapError(ErrInvalidResponse, fmt.Errorf("expected InitiateSessionResponse, got %s", msg.Result))
	}

	err = binding.Validator.ValidateStruct(typedMsg)
	if err != nil {
		return nil, lib.WrapError(ErrInvalidResponse, err)
	}

	signature := typedMsg.Signature
	typedMsg.Signature = lib.HexString{}

	providerPubKey := typedMsg.PubKey
	if !p.validateMsgSignature(typedMsg, signature, typedMsg.PubKey) {
		return nil, ErrInvalidSig
	}

	err = p.sessionStorage.AddUser(&storages.User{
		Addr:   provider.Hex(),
		PubKey: providerPubKey.String(),
		Url:    providerURL,
	})
	if err != nil {
		return nil, lib.WrapError(ErrFailedStore, err)
	}

	return typedMsg, nil
}

func (p *ProxyServiceSender) GetSessionReportFromProvider(ctx context.Context, sessionID common.Hash) (*msgs.SessionReportRes, error) {
	requestID := "1"

	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return nil, ErrMissingPrKey
	}

	session, err := p.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}
	provider, ok := p.sessionStorage.GetUser(session.ProviderAddr().Hex())
	if !ok {
		return nil, ErrProviderNotFound
	}

	getSessionReportRequest, err := p.morRPC.SessionReportRequest(sessionID, prKey, requestID)
	if err != nil {
		return nil, lib.WrapError(ErrCreateReq, err)
	}

	msg, code, err := p.rpcRequest(provider.Url, getSessionReportRequest)
	if err != nil {
		return nil, lib.WrapError(ErrProvider, fmt.Errorf("code: %d, msg: %v, error: %s", code, msg, err))
	}

	if msg.Error != nil {
		// TODO: verify signature
		return nil, lib.WrapError(ErrResponseErr, fmt.Errorf("error: %v, result: %v", msg.Error.Message, msg.Error.Data))
	}
	if msg.Result == nil {
		return nil, lib.WrapError(ErrInvalidResponse, ErrEmpty)
	}

	var typedMsg *msgs.SessionReportRes
	err = json.Unmarshal(*msg.Result, &typedMsg)
	if err != nil {
		return nil, lib.WrapError(ErrInvalidResponse, fmt.Errorf("expected SessionReportRespose, got %s", msg.Result))
	}

	err = binding.Validator.ValidateStruct(typedMsg)
	if err != nil {
		return nil, lib.WrapError(ErrInvalidResponse, err)
	}

	signature := typedMsg.Signature
	typedMsg.Signature = lib.HexString{}

	hexPubKey, err := lib.StringToHexString(provider.PubKey)
	if err != nil {
		return nil, lib.WrapError(ErrInvalidResponse, err)
	}

	if !p.validateMsgSignature(typedMsg, signature, hexPubKey) {
		return nil, ErrInvalidSig
	}

	return typedMsg, nil
}

func (p *ProxyServiceSender) GetSessionReportFromUser(ctx context.Context, sessionID common.Hash) (lib.HexString, lib.HexString, error) {
	session, err := p.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, nil, ErrSessionNotFound
	}

	TPSScaled1000Arr, TTFTMsArr := session.GetStats()

	tps := 0
	ttft := 0
	for _, tpsVal := range TPSScaled1000Arr {
		tps += tpsVal
	}
	for _, ttftVal := range TTFTMsArr {
		ttft += ttftVal
	}

	if len(TPSScaled1000Arr) != 0 {
		tps /= len(TPSScaled1000Arr)
	}
	if len(TTFTMsArr) != 0 {
		ttft /= len(TTFTMsArr)
	}

	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return nil, nil, ErrMissingPrKey
	}

	response, err := p.morRPC.SessionReportResponse(
		uint32(tps),
		uint32(ttft),
		sessionID,
		prKey,
		"1",
		p.chainID,
	)

	if err != nil {
		return nil, nil, lib.WrapError(ErrGenerateReport, err)
	}

	var typedMsg *msgs.SessionReportRes
	err = json.Unmarshal(*response.Result, &typedMsg)
	if err != nil {
		return nil, nil, lib.WrapError(ErrInvalidResponse, fmt.Errorf("expected SessionReportRespose, got %s", response.Result))
	}

	return typedMsg.Message, typedMsg.SignedReport, nil
}

func (p *ProxyServiceSender) CallAgentTool(ctx context.Context, sessionID common.Hash, toolName string, input map[string]interface{}) (string, error) {
	requestID := "1"

	session, err := p.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return "", ErrSessionNotFound
	}

	isExpired := session.EndsAt().Int64()-time.Now().Unix() < 0
	if isExpired {
		return "", ErrSessionExpired
	}

	provider, ok := p.sessionStorage.GetUser(session.ProviderAddr().Hex())
	if !ok {
		return "", ErrProviderNotFound
	}

	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return "", ErrMissingPrKey
	}

	callAgentToolRequest, err := p.morRPC.CallAgentToolRequest(sessionID, toolName, input, prKey, requestID)
	if err != nil {
		return "", lib.WrapError(ErrCreateReq, err)
	}

	msg, code, err := p.rpcRequest(provider.Url, callAgentToolRequest)
	if err != nil {
		return "", lib.WrapError(ErrProvider, fmt.Errorf("code: %d, msg: %v, error: %s", code, msg, err))
	}

	if msg.Error != nil {
		return "", lib.WrapError(ErrResponseErr, fmt.Errorf("error: %v, result: %v", msg.Error.Message, msg.Error.Data))
	}
	if msg.Result == nil {
		return "", lib.WrapError(ErrInvalidResponse, ErrEmpty)
	}

	var typedMsg *msgs.CallAgentToolRes
	err = json.Unmarshal(*msg.Result, &typedMsg)
	if err != nil {
		return "", lib.WrapError(ErrInvalidResponse, fmt.Errorf("expected CallAgentToolRespose, got %s", msg.Result))
	}

	signature := typedMsg.Signature
	typedMsg.Signature = lib.HexString{}

	hexPubKey, err := lib.StringToHexString(provider.PubKey)
	if !p.validateMsgSignature(typedMsg, signature, hexPubKey) {
		return "", ErrInvalidSig
	}

	decryptedMessage, err := lib.DecryptString(typedMsg.Message, prKey.Hex())
	if err != nil {
		return "", lib.WrapError(ErrDecrFailed, err)
	}

	return string(decryptedMessage), nil
}

func (p *ProxyServiceSender) GetAgentTools(ctx context.Context, sessionID common.Hash) (string, error) {
	requestID := "1"

	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return "", ErrMissingPrKey
	}

	session, err := p.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return "", ErrSessionNotFound
	}

	isExpired := session.EndsAt().Int64()-time.Now().Unix() < 0
	if isExpired {
		return "", ErrSessionExpired
	}

	provider, ok := p.sessionStorage.GetUser(session.ProviderAddr().Hex())
	if !ok {
		return "", ErrProviderNotFound
	}

	getAgentToolsRequest, err := p.morRPC.CallGetAgentToolsRequest(sessionID, prKey, requestID)
	if err != nil {
		return "", lib.WrapError(ErrCreateReq, err)
	}

	msg, code, err := p.rpcRequest(provider.Url, getAgentToolsRequest)
	if err != nil {
		return "", lib.WrapError(ErrProvider, fmt.Errorf("code: %d, msg: %v, error: %s", code, msg, err))
	}

	if msg.Error != nil {
		return "", lib.WrapError(ErrResponseErr, fmt.Errorf("error: %v, result: %v", msg.Error.Message, msg.Error.Data))
	}
	if msg.Result == nil {
		return "", lib.WrapError(ErrInvalidResponse, ErrEmpty)
	}

	var typedMsg *msgs.GetAgentToolsRes
	err = json.Unmarshal(*msg.Result, &typedMsg)
	if err != nil {
		return "", lib.WrapError(ErrInvalidResponse, fmt.Errorf("expected GetAgentToolsRespose, got %s", msg.Result))
	}

	signature := typedMsg.Signature
	typedMsg.Signature = lib.HexString{}

	hexPubKey, err := lib.StringToHexString(provider.PubKey)
	if !p.validateMsgSignature(typedMsg, signature, hexPubKey) {
		return "", ErrInvalidSig
	}

	decryptedResponse, err := lib.DecryptString(typedMsg.Message, prKey.Hex())
	if err != nil {
		return "", lib.WrapError(ErrDecrFailed, err)
	}

	return string(decryptedResponse), nil
}

func (p *ProxyServiceSender) rpcRequest(url string, rpcMessage *msgs.RPCMessage) (*msgs.RpcResponse, int, error) {
	// TODO: enable request-response matching by using requestID
	// TODO: add context cancellation

	TIMEOUT_TO_ESTABLISH_CONNECTION := time.Second * 3
	dialer := net.Dialer{Timeout: TIMEOUT_TO_ESTABLISH_CONNECTION}

	conn, err := dialer.Dial("tcp", url)
	if err != nil {
		err = lib.WrapError(ErrConnectProvider, err)
		p.log.Warnf(err.Error())
		return nil, http.StatusInternalServerError, err
	}
	defer conn.Close()

	msgJSON, err := json.Marshal(rpcMessage)
	if err != nil {
		err = lib.WrapError(ErrMasrshalFailed, err)
		p.log.Errorf("%s", err)
		return nil, http.StatusInternalServerError, err
	}
	_, err = conn.Write(msgJSON)
	if err != nil {
		err = lib.WrapError(ErrWriteProvider, err)
		p.log.Errorf("%s", err)
		return nil, http.StatusInternalServerError, err
	}

	// read response
	reader := bufio.NewReader(conn)
	d := json.NewDecoder(reader)

	var msg *msgs.RpcResponse
	err = d.Decode(&msg)
	if err != nil {
		err = lib.WrapError(ErrDecode, err)
		p.log.Errorf("%s", err)
		return nil, http.StatusBadRequest, err
	}
	return msg, 0, nil
}

func (p *ProxyServiceSender) validateMsgSignature(result any, signature lib.HexString, providerPubicKey lib.HexString) bool {
	return p.morRPC.VerifySignature(result, signature, providerPubicKey, p.log)
}

func (p *ProxyServiceSender) GetModelIdSession(ctx context.Context, sessionID common.Hash) (common.Hash, error) {
	session, err := p.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return common.Hash{}, ErrSessionNotFound
	}
	return session.ModelID(), nil
}

func (p *ProxyServiceSender) validateMsgSignatureAddr(result any, signature lib.HexString, providerAddr common.Address) bool {
	return p.morRPC.VerifySignatureAddr(result, signature, providerAddr, p.log)
}

// validateSession checks if a session is valid and returns session and provider information
func (p *ProxyServiceSender) validateSession(ctx context.Context, sessionID common.Hash) (*sessionrepo.SessionModel, *storages.User, error) {
	// Get session and verify it exists
	session, err := p.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, nil, ErrSessionNotFound
	}

	// Check if session is expired
	if session.EndsAt().Int64() < time.Now().Unix() {
		return nil, nil, ErrSessionExpired
	}

	// Get provider information
	provider, ok := p.sessionStorage.GetUser(session.ProviderAddr().Hex())
	if !ok {
		return nil, nil, ErrProviderNotFound
	}

	return session, provider, nil
}

// prepareRequest creates and prepares an RPC request for the provider
func (p *ProxyServiceSender) prepareRequest(sessionID common.Hash, payload interface{}, providerPubKey string) (*msgs.RPCMessage, lib.HexString, error) {
	// Get private key for encryption
	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return nil, nil, ErrMissingPrKey
	}

	// Convert provider public key to hex string
	pubKey, err := lib.StringToHexString(providerPubKey)
	if err != nil {
		return nil, nil, lib.WrapError(ErrCreateReq, err)
	}

	// Create RPC request
	promptRequest, err := p.morRPC.SessionPromptRequest(sessionID, payload, pubKey, prKey, "1")
	if err != nil {
		return nil, nil, lib.WrapError(ErrCreateReq, err)
	}

	return promptRequest, pubKey, nil
}

// handleFailover manages the failover process when a request fails
func (p *ProxyServiceSender) handleFailover(ctx context.Context, session sessionrepo.SessionModel, cb gcs.CompletionCallback) (common.Hash, error) {
	// Close current session
	_, err := p.sessionService.CloseSession(ctx, session.ID())
	if err != nil {
		return common.Hash{}, err
	}

	if err = cb(ctx, gcs.NewChunkControl("provider failed, failover enabled"), nil); err != nil {
		return common.Hash{}, err
	}

	// Calculate remaining session duration
	duration := session.EndsAt().Int64() - time.Now().Unix()

	// Open new session with same parameters
	newSessionID, err := p.sessionService.OpenSessionByModelId(
		ctx,
		session.ModelID(),
		big.NewInt(duration),
		session.DirectPayment(),
		session.FailoverEnabled(),
		session.ProviderAddr(),
		session.AgentUsername(),
	)
	if err != nil {
		return common.Hash{}, err
	}

	// Notify about new session
	msg := fmt.Sprintf("new session opened: %s", newSessionID.Hex())
	if err = cb(ctx, gcs.NewChunkControl(msg), nil); err != nil {
		return common.Hash{}, err
	}

	return newSessionID, nil
}

// updateSessionStats updates session statistics after request completion
func (p *ProxyServiceSender) updateSessionStats(ctx context.Context, session sessionrepo.SessionModel, startTime int64, ttftMs, totalTokens int) error {
	requestDuration := int(time.Now().Unix() - startTime)
	if requestDuration == 0 {
		requestDuration = 1
	}

	session.AddStats(totalTokens*1000/requestDuration, ttftMs)

	err := p.sessionRepo.SaveSession(ctx, &session)
	if err != nil {
		p.log.Error(`failed to update session report stats`, err)
		return err
	}

	return nil
}

func (p *ProxyServiceSender) SendPromptV2(ctx context.Context, sessionID common.Hash, prompt *gcs.OpenAICompletionRequestExtra, cb gcs.CompletionCallback) (interface{}, error) {
	// Validate session and get provider
	session, provider, err := p.validateSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Prepare request
	promptRequest, pubKey, err := p.prepareRequest(sessionID, prompt, provider.PubKey)
	if err != nil {
		return nil, err
	}

	// Send request and process response
	startTime := time.Now().Unix()
	result, ttftMs, totalTokens, err := p.rpcRequestStreamV2(ctx, cb, provider.Url, promptRequest, pubKey, "chat_completion")

	// Handle errors with failover if enabled
	if err != nil {
		if !session.FailoverEnabled() {
			return nil, lib.WrapError(ErrProvider, err)
		}

		// Handle failover
		newSessionID, failoverErr := p.handleFailover(ctx, *session, cb)
		if failoverErr != nil {
			return nil, failoverErr
		}

		// Retry with new session
		return p.SendPromptV2(ctx, newSessionID, prompt, cb)
	}

	// Update session statistics
	if updateErr := p.updateSessionStats(ctx, *session, startTime, ttftMs, totalTokens); updateErr != nil {
		// Log error but don't fail the request
		p.log.Error("Failed to update session stats", updateErr)
	}

	return result, nil
}

func (p *ProxyServiceSender) rpcRequestStreamV2(
	ctx context.Context,
	cb gcs.CompletionCallback,
	url string,
	rpcMessage *msgs.RPCMessage,
	providerPublicKey lib.HexString,
	requestType string,
) (interface{}, int, int, error) {
	const (
		TIMEOUT_TO_ESTABLISH_CONNECTION   = time.Second * 3
		TIMEOUT_TO_RECEIVE_FIRST_RESPONSE = time.Second * 30
	)
	var MAX_RETRIES = 5
	if requestType == "audio_transcription" || requestType == "audio_speech" {
		MAX_RETRIES = 20 // Increase retries for audio transcription
	}

	dialer := net.Dialer{Timeout: TIMEOUT_TO_ESTABLISH_CONNECTION}

	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return nil, 0, 0, ErrMissingPrKey
	}

	conn, err := dialer.Dial("tcp", url)
	if err != nil {
		err = lib.WrapError(ErrConnectProvider, err)
		p.log.Warnf(err.Error())
		return nil, 0, 0, err
	}
	defer conn.Close()

	// Set initial read deadline
	_ = conn.SetReadDeadline(time.Now().Add(TIMEOUT_TO_RECEIVE_FIRST_RESPONSE))

	msgJSON, err := json.Marshal(rpcMessage)
	if err != nil {
		return nil, 0, 0, lib.WrapError(ErrMasrshalFailed, err)
	}

	ttftMs := 0
	totalTokens := 0
	now := time.Now().UnixMilli()

	_, err = conn.Write(msgJSON)
	if err != nil {
		return nil, ttftMs, totalTokens, err
	}

	reader := bufio.NewReader(conn)
	// We need to recreate the decoder if it becomes invalid
	var d *json.Decoder

	responses := make([]interface{}, 0)

	retryCount := 0

	for {
		if ctx.Err() != nil {
			return nil, ttftMs, totalTokens, ctx.Err()
		}

		// Initialize or reset the decoder
		if d == nil {
			d = json.NewDecoder(reader)
		}

		var msg *msgs.RpcResponse
		err = d.Decode(&msg)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				p.log.Warnf("Read operation timed out: %v", err)
				p.log.Infof("Retry count: %d, max retries: %d", retryCount, MAX_RETRIES)
				if retryCount < MAX_RETRIES {
					alive, availErr := checkProviderAvailability(url)
					if availErr != nil {
						p.log.Warnf("Provider availability check failed: %v", availErr)
						return nil, ttftMs, totalTokens, fmt.Errorf("provider availability check failed: %w", availErr)
					}
					if alive {
						retryCount++
						p.log.Infof("Provider is alive, retrying (%d/%d)...", retryCount, MAX_RETRIES)
						// Reset the read deadline
						conn.SetReadDeadline(time.Now().Add(TIMEOUT_TO_RECEIVE_FIRST_RESPONSE))
						// Clear the error state by reading any remaining data
						reader.Discard(reader.Buffered())
						// Reset the decoder
						d = nil
						continue
					} else {
						return nil, ttftMs, totalTokens, fmt.Errorf("provider is not available")
					}
				} else {
					return nil, ttftMs, totalTokens, fmt.Errorf("read timed out after %d retries: %w", retryCount, err)
				}
			} else if err == io.EOF {
				p.log.Debugf("Connection closed by provider")
				break
			} else {
				p.log.Warnf("Failed to decode response: %v", err)
				return nil, ttftMs, totalTokens, lib.WrapError(ErrInvalidResponse, err)
			}
		}

		if msg.Error != nil {
			sig := msg.Error.Data.Signature
			msg.Error.Data.Signature = []byte{}

			if !p.validateMsgSignature(msg.Error, sig, providerPublicKey) {
				return nil, ttftMs, totalTokens, ErrInvalidSig
			}

			errorMessage, err := lib.DecryptString(msg.Error.Message, prKey.Hex())
			if err != nil {
				return nil, ttftMs, totalTokens, lib.WrapError(ErrDecrFailed, err)
			}

			var aiEngineErrorResponse gcs.AiEngineErrorResponse
			err = json.Unmarshal([]byte(errorMessage), &aiEngineErrorResponse)
			if err != nil {
				return nil, ttftMs, totalTokens, lib.WrapError(ErrInvalidResponse, err)
			}

			cb(ctx, nil, &aiEngineErrorResponse)
			return nil, ttftMs, totalTokens, nil
		}

		if msg.Result == nil {
			return nil, ttftMs, totalTokens, lib.WrapError(ErrInvalidResponse, ErrEmpty)
		}

		if ttftMs == 0 {
			ttftMs = int(time.Now().UnixMilli() - now)
			_ = conn.SetReadDeadline(time.Time{}) // Clear read deadline
		}

		var inferenceRes InferenceRes
		err = json.Unmarshal(*msg.Result, &inferenceRes)
		if err != nil {
			return nil, ttftMs, totalTokens, lib.WrapError(ErrInvalidResponse, err)
		}
		sig := inferenceRes.Signature
		inferenceRes.Signature = []byte{}

		if !p.validateMsgSignature(inferenceRes, sig, providerPublicKey) {
			return nil, ttftMs, totalTokens, ErrInvalidSig
		}

		var message lib.HexString
		err = json.Unmarshal(inferenceRes.Message, &message)
		if err != nil {
			return nil, ttftMs, totalTokens, lib.WrapError(ErrInvalidResponse, err)
		}

		aiResponse, err := lib.DecryptBytes(message, prKey)
		if err != nil {
			return nil, ttftMs, totalTokens, lib.WrapError(ErrDecrFailed, err)
		}

		// Process the AI response based on the request type
		result, tokens, shouldStop, err := p.processAIResponse(requestType, aiResponse, responses)
		if err != nil {
			return nil, ttftMs, totalTokens, err
		}

		totalTokens += tokens

		if ctx.Err() != nil {
			return nil, ttftMs, totalTokens, ctx.Err()
		}
		err = cb(ctx, result, nil)
		if err != nil {
			return nil, ttftMs, totalTokens, lib.WrapError(ErrResponseErr, err)
		}

		if shouldStop {
			break
		}
	}

	return responses, ttftMs, totalTokens, nil
}

// processAIResponse handles different response types and returns the appropriate chunk
func (p *ProxyServiceSender) processAIResponse(requestType string, aiResponse []byte, responses []interface{}) (gcs.Chunk, int, bool, error) {
	switch requestType {
	case "audio_transcription":
		return p.handleAudioTranscription(aiResponse, responses)
	case "audio_speech":
		return p.handleAudioSpeech(aiResponse, responses)
	case "chat_completion":
		return p.handleChatCompletion(aiResponse, responses)
	case "embeddings":
		return p.handleEmbeddings(aiResponse, responses)
	default:
		return p.handleMediaGeneration(aiResponse, responses)
	}
}

// handleAudioTranscription processes audio transcription responses
func (p *ProxyServiceSender) handleAudioTranscription(aiResponse []byte, responses []interface{}) (gcs.Chunk, int, bool, error) {
	if aiResponse == nil || len(aiResponse) == 0 {
		return nil, 0, false, lib.WrapError(ErrInvalidResponse, fmt.Errorf("empty audio response"))
	}

	// Check if this is a streaming delta response
	var deltaResponse gcs.AudioTranscriptionDelta
	err := json.Unmarshal(aiResponse, &deltaResponse)
	if err == nil && deltaResponse.Type != "" {
		chunk := gcs.NewChunkAudioTranscriptionDelta(deltaResponse)
		responses = append(responses, deltaResponse)
		return chunk, 0, false, nil // Don't stop for delta responses
	}

	// Try to parse as JSON response first
	var jsonResponse gcs.AudioResponseExtra
	err = json.Unmarshal(aiResponse, &jsonResponse)
	if err == nil {
		chunk := gcs.NewChunkAudioTranscriptionJson(jsonResponse)
		responses = append(responses, jsonResponse)
		return chunk, chunk.Tokens(), true, nil
	}

	// Fall back to string response
	var responseString string
	err = json.Unmarshal(aiResponse, &responseString)
	if err != nil {
		return nil, 0, false, lib.WrapError(ErrInvalidResponse, err)
	}

	chunk := gcs.NewChunkAudioTranscriptionText(responseString)
	responses = append(responses, responseString)
	return chunk, chunk.Tokens(), true, nil
}

// handleAudioSpeech processes audio speech responses
func (p *ProxyServiceSender) handleAudioSpeech(aiResponse []byte, responses []interface{}) (gcs.Chunk, int, bool, error) {
	if aiResponse == nil || len(aiResponse) == 0 {
		return nil, 0, false, lib.WrapError(ErrInvalidResponse, fmt.Errorf("empty audio speech response"))
	}

	var data []byte
	err := json.Unmarshal(aiResponse, &data)
	if err != nil {
		return nil, 0, false, lib.WrapError(ErrInvalidResponse, fmt.Errorf("failed to parse audio speech response: %w", err))
	}

	if len(data) == 0 {
		// Indicate end of speech response
		chunk := gcs.NewChunkAudioSpeech([]byte{})
		return chunk, 0, true, nil
	}

	// Audio speech responses are typically binary audio data
	chunk := gcs.NewChunkAudioSpeech(data)
	responses = append(responses, data)
	return chunk, len(data), false, nil
}

// handleChatCompletion processes chat completion responses
func (p *ProxyServiceSender) handleChatCompletion(aiResponse []byte, responses []interface{}) (gcs.Chunk, int, bool, error) {
	// Try to parse as streaming response
	var streamResponse gcs.ChatCompletionStreamResponseExtra
	err := json.Unmarshal(aiResponse, &streamResponse)
	if err == nil && streamResponse.Usage == nil && len(streamResponse.Choices) > 0 {
		choices := streamResponse.Choices
		shouldStop := false

		for _, choice := range choices {
			if choice.FinishReason == openai.FinishReasonStop {
				shouldStop = true
				break
			}
		}

		chunk := gcs.NewChunkStreaming(&streamResponse)
		responses = append(responses, streamResponse)
		return chunk, len(choices), shouldStop, nil
	}

	// Try to parse as full completion response
	var chatResponse gcs.ChatCompletionResponseExtra
	err = json.Unmarshal(aiResponse, &chatResponse)
	if err == nil && len(chatResponse.Choices) > 0 {
		chunk := gcs.NewChunkText(&chatResponse)
		responses = append(responses, chatResponse)
		return chunk, chatResponse.Usage.TotalTokens, true, nil
	}

	// If not a chat completion, try media generation handlers
	return p.handleMediaGeneration(aiResponse, responses)
}

// handleMediaGeneration processes image and video generation responses
func (p *ProxyServiceSender) handleMediaGeneration(aiResponse []byte, responses []interface{}) (gcs.Chunk, int, bool, error) {
	// Try image URL response
	var imageResult gcs.ImageGenerationResult
	err := json.Unmarshal(aiResponse, &imageResult)
	if err == nil && imageResult.ImageUrl != "" {
		chunk := gcs.NewChunkImage(&imageResult)
		responses = append(responses, imageResult)
		return chunk, 1, true, nil
	}

	// Try video response
	var videoResult gcs.VideoGenerationResult
	err = json.Unmarshal(aiResponse, &videoResult)
	if err == nil && videoResult.VideoRawContent != "" {
		chunk := gcs.NewChunkVideo(&videoResult)
		responses = append(responses, videoResult)
		return chunk, 1, true, nil
	}

	// Try raw image content response
	var rawImageResult gcs.ImageRawContentResult
	err = json.Unmarshal(aiResponse, &rawImageResult)
	if err == nil && rawImageResult.ImageRawContent != "" {
		chunk := gcs.NewChunkImageRawContent(&rawImageResult)
		responses = append(responses, rawImageResult)
		return chunk, 1, true, nil
	}

	// If we got here, we couldn't parse the response
	return nil, 0, false, lib.WrapError(ErrInvalidResponse, fmt.Errorf("unknown response format"))
}

// handleEmbeddings processes embeddings generation responses
func (p *ProxyServiceSender) handleEmbeddings(aiResponse []byte, responses []interface{}) (gcs.Chunk, int, bool, error) {
	if aiResponse == nil || len(aiResponse) == 0 {
		return nil, 0, false, lib.WrapError(ErrInvalidResponse, fmt.Errorf("empty embeddings response"))
	}

	var embeddingResp gcs.EmbeddingsResponse
	if err := json.Unmarshal(aiResponse, &embeddingResp); err == nil && len(embeddingResp.Data) > 0 {
		chunk := gcs.NewChunkEmbedding(embeddingResp)
		responses = append(responses, embeddingResp)
		return chunk, 0, true, nil
	}

	return nil, 0, false, lib.WrapError(ErrInvalidResponse, fmt.Errorf("unknown embeddings response format"))
}

// checkProviderAvailability checks if the provider is alive using portchecker.io API
func checkProviderAvailability(url string) (bool, error) {
	host, port, err := net.SplitHostPort(url)
	if err != nil {
		return false, err
	}

	portInt, err := strconv.Atoi(port)
	if err != nil {
		return false, err
	}

	requestBody, err := json.Marshal(map[string]interface{}{
		"host":  host,
		"ports": []int{portInt},
	})
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", "https://portchecker.io/api/v1/query", bytes.NewBuffer(requestBody))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	var response struct {
		Check []struct {
			Status bool `json:"status"`
			Port   int  `json:"port"`
		} `json:"check"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return false, err
	}

	for _, check := range response.Check {
		if check.Port == portInt {
			return check.Status, nil
		}
	}

	return false, fmt.Errorf("port status not found in response")
}

// SendAudioTranscriptionStreamV2 sends audio transcription using streaming chunks to avoid memory issues with large files
func (p *ProxyServiceSender) SendAudioTranscriptionV2(ctx context.Context, sessionID common.Hash, audioRequest *gcs.AudioTranscriptionRequest, cb gcs.CompletionCallback) (interface{}, error) {
	// Validate session and get provider
	session, provider, err := p.validateSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Get private key for signing
	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return nil, ErrMissingPrKey
	}

	// Open audio file
	audioFilePath := audioRequest.FilePath
	file, err := os.Open(audioFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	fileSize := uint64(fileInfo.Size())
	totalChunks := uint32((fileSize + SENDER_AUDIO_STREAM_CHUNK_SIZE - 1) / SENDER_AUDIO_STREAM_CHUNK_SIZE) // Ceiling division

	// Generate unique stream ID
	streamIDBytes := make([]byte, 16)
	_, err = rand.Read(streamIDBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate stream ID: %w", err)
	}
	streamID := fmt.Sprintf("%x", streamIDBytes)

	// Detect content type
	contentType := detectAudioContentType(audioFilePath)

	p.log.Debugf("Starting audio streaming for file %s, size: %d bytes, chunks: %d", audioFilePath, fileSize, totalChunks)

	// Record start time for session stats
	startTime := time.Now().Unix()

	// Step 1: Start streaming session
	err = p.sendStreamStart(provider, sessionID, streamID, totalChunks, fileSize, contentType, prKey)
	if err != nil {
		return nil, fmt.Errorf("failed to start streaming session: %w", err)
	}

	// Step 2: Send chunks
	err = p.sendStreamChunks(ctx, provider, sessionID, streamID, file, totalChunks, prKey)
	if err != nil {
		return nil, fmt.Errorf("failed to send streaming chunks: %w", err)
	}

	// Step 3: End streaming and get results
	result, ttftMs, totalTokens, err := p.sendStreamEnd(ctx, provider, sessionID, streamID, audioRequest, prKey, cb)
	if err != nil {
		return nil, fmt.Errorf("failed to end streaming session: %w", err)
	}

	// Update session statistics
	if updateErr := p.updateSessionStats(ctx, *session, startTime, ttftMs, totalTokens); updateErr != nil {
		// Log error but don't fail the request
		p.log.Error("Failed to update session stats", updateErr)
	}

	return result, nil
}

// sendStreamStart initiates the streaming session
func (p *ProxyServiceSender) sendStreamStart(provider *storages.User, sessionID common.Hash, streamID string, totalChunks uint32, fileSize uint64, contentType string, prKey lib.HexString) error {
	requestID := "1"
	message, err := p.morRPC.SessionPromptStreamStartRequest(sessionID, streamID, totalChunks, contentType, fileSize, prKey, requestID)
	if err != nil {
		return fmt.Errorf("failed to create stream start request: %w", err)
	}

	response, _, err := p.rpcRequest(provider.Url, message)
	if err != nil {
		return fmt.Errorf("failed to send stream start request: %w", err)
	}

	var typedMsg msgs.SessionPromptStreamStartRes
	err = json.Unmarshal(*response.Result, &typedMsg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stream start response: %w", err)
	}

	signature := typedMsg.Signature
	typedMsg.Signature = lib.HexString{}

	hexPubKey, err := lib.StringToHexString(provider.PubKey)
	if !p.validateMsgSignature(typedMsg, signature, hexPubKey) {
		return fmt.Errorf("invalid signature for stream start response")
	}

	if response.Error != nil {
		return fmt.Errorf("stream start failed: %s", response.Error.Message)
	}

	p.log.Debugf("Stream start successful for stream ID: %s", streamID)
	return nil
}

// sendStreamChunks sends the audio file in chunks
func (p *ProxyServiceSender) sendStreamChunks(ctx context.Context, provider *storages.User, sessionID common.Hash, streamID string, file *os.File, totalChunks uint32, prKey lib.HexString) error {
	buffer := make([]byte, SENDER_AUDIO_STREAM_CHUNK_SIZE)
	var chunkIndex uint32 = 0

	for chunkIndex < totalChunks {
		// Check context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Read chunk
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read chunk %d: %w", chunkIndex, err)
		}

		if n == 0 {
			break
		}

		chunkData := buffer[:n]
		requestID := "1"

		message, err := p.morRPC.SessionPromptStreamChunkRequest(sessionID, streamID, chunkIndex, chunkData, prKey, requestID)
		if err != nil {
			return fmt.Errorf("failed to create chunk request for chunk %d: %w", chunkIndex, err)
		}

		response, _, err := p.rpcRequest(provider.Url, message)
		if err != nil {
			return fmt.Errorf("failed to send chunk %d: %w", chunkIndex, err)
		}

		var typedMsg *msgs.SessionPromptStreamChunkRes
		err = json.Unmarshal(*response.Result, &typedMsg)
		if err != nil {
			return fmt.Errorf("failed to unmarshal stream start response: %w", err)
		}

		signature := typedMsg.Signature
		typedMsg.Signature = lib.HexString{}

		hexPubKey, err := lib.StringToHexString(provider.PubKey)
		if !p.validateMsgSignature(typedMsg, signature, hexPubKey) {
			return fmt.Errorf("invalid signature for stream chunk response")
		}

		if response.Error != nil {
			return fmt.Errorf("chunk %d failed: %s", chunkIndex, response.Error.Message)
		}

		p.log.Debugf("Successfully sent chunk %d/%d for stream %s", chunkIndex+1, totalChunks, streamID)
		chunkIndex++
	}

	return nil
}

// sendStreamEnd finalizes the streaming session and gets the transcription result with session statistics
func (p *ProxyServiceSender) sendStreamEnd(ctx context.Context, provider *storages.User, sessionID common.Hash, streamID string, audioRequest *gcs.AudioTranscriptionRequest, prKey lib.HexString, cb gcs.CompletionCallback) (interface{}, int, int, error) {
	// Marshal audio parameters
	audioRequest.Extra["type"] = json.RawMessage(`"audio_transcription"`)
	audioParamsJSON, err := json.Marshal(audioRequest)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to marshal audio parameters: %w", err)
	}

	pubKey, err := lib.StringToHexString(provider.PubKey)
	if err != nil {
		return nil, 0, 0, lib.WrapError(ErrCreateReq, err)
	}

	requestID := "1"
	message, err := p.morRPC.SessionPromptStreamEndRequest(sessionID, streamID, string(audioParamsJSON), prKey, requestID)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to create stream end request: %w", err)
	}

	// Send request and handle streaming response
	result, ttftMs, totalTokens, err := p.rpcRequestStreamV2(ctx, cb, provider.Url, message, pubKey, "audio_transcription")
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to process stream end request: %w", err)
	}

	p.log.Debugf("Successfully completed streaming session %s with TTFT: %dms, tokens: %d", streamID, ttftMs, totalTokens)
	return result, ttftMs, totalTokens, nil
}

// detectAudioContentType detects the content type of an audio file based on its extension
func detectAudioContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	contentTypes := map[string]string{
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".wave": "audio/wav",
		".ogg":  "audio/ogg",
		".flac": "audio/flac",
		".aac":  "audio/aac",
		".m4a":  "audio/mp4",
		".webm": "audio/webm",
		".opus": "audio/opus",
		".wma":  "audio/x-ms-wma",
		".amr":  "audio/amr",
		".3gp":  "audio/3gpp",
		".aiff": "audio/aiff",
	}

	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}
	return "audio/mpeg" // default
}

// SendAudioSpeech sends audio speech generation request
func (p *ProxyServiceSender) SendAudioSpeech(ctx context.Context, sessionID common.Hash, audioRequest *gcs.AudioSpeechRequest, cb gcs.CompletionCallback) (interface{}, error) {
	// Validate session and get provider
	session, provider, err := p.validateSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Get private key for signing
	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return nil, ErrMissingPrKey
	}

	pubKey, err := lib.StringToHexString(provider.PubKey)
	if err != nil {
		return nil, lib.WrapError(ErrCreateReq, err)
	}

	requestID := "1"
	audioRequest.Extra["type"] = json.RawMessage(`"audio_speech"`)
	message, err := p.morRPC.SessionPromptRequest(sessionID, audioRequest, pubKey, prKey, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to create audio speech request: %w", err)
	}

	// Record start time for session stats
	startTime := time.Now().Unix()

	// Send request and handle response
	result, ttftMs, totalTokens, err := p.rpcRequestStreamV2(ctx, cb, provider.Url, message, pubKey, "audio_speech")

	// Handle errors with failover if enabled
	if err != nil {
		if !session.FailoverEnabled() {
			return nil, lib.WrapError(ErrProvider, err)
		}

		// Handle failover
		newSessionID, failoverErr := p.handleFailover(ctx, *session, cb)
		if failoverErr != nil {
			return nil, failoverErr
		}

		// Retry with new session
		return p.SendAudioSpeech(ctx, newSessionID, audioRequest, cb)
	}

	// Update session statistics
	if updateErr := p.updateSessionStats(ctx, *session, startTime, ttftMs, totalTokens); updateErr != nil {
		// Log error but don't fail the request
		p.log.Error("Failed to update session stats", updateErr)
	}

	p.log.Debugf("Successfully completed audio speech generation for session %s with TTFT: %dms, tokens: %d", sessionID.Hex(), ttftMs, totalTokens)
	return result, nil
}

// SendEmbeddings sends an embeddings generation request
func (p *ProxyServiceSender) SendEmbeddings(ctx context.Context, sessionID common.Hash, embedRequest *gcs.EmbeddingsRequest, cb gcs.CompletionCallback) (interface{}, error) {
	// Validate session and get provider
	session, provider, err := p.validateSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Get private key for signing
	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return nil, ErrMissingPrKey
	}

	// Mark request type for provider side parsing
	if embedRequest.Extra == nil {
		embedRequest.Extra = make(map[string]json.RawMessage)
	}
	embedRequest.Extra["type"] = json.RawMessage(`"embeddings"`)

	pubKey, err := lib.StringToHexString(provider.PubKey)
	if err != nil {
		return nil, lib.WrapError(ErrCreateReq, err)
	}

	requestID := "1"
	message, err := p.morRPC.SessionPromptRequest(sessionID, embedRequest, pubKey, prKey, requestID)
	if err != nil {
		return nil, fmt.Errorf("failed to create embeddings request: %w", err)
	}

	startTime := time.Now().Unix()

	result, ttftMs, totalTokens, err := p.rpcRequestStreamV2(ctx, cb, provider.Url, message, pubKey, "embeddings")

	// Handle errors with failover if enabled
	if err != nil {
		if !session.FailoverEnabled() {
			return nil, lib.WrapError(ErrProvider, err)
		}

		// Handle failover
		newSessionID, failoverErr := p.handleFailover(ctx, *session, cb)
		if failoverErr != nil {
			return nil, failoverErr
		}

		// Retry with new session
		return p.SendEmbeddings(ctx, newSessionID, embedRequest, cb)
	}

	if updateErr := p.updateSessionStats(ctx, *session, startTime, ttftMs, totalTokens); updateErr != nil {
		p.log.Error("Failed to update session stats", updateErr)
	}

	p.log.Debugf("Successfully completed embeddings generation for session %s with TTFT: %dms, tokens: %d", sessionID.Hex(), ttftMs, totalTokens)
	return result, nil
}
