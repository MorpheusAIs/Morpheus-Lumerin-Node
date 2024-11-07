package proxyapi

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"time"

	constants "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/aiengine"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	msgs "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
	sessionrepo "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/session"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
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
)

type ProxyServiceSender struct {
	chainID        *big.Int
	publicUrl      *url.URL
	privateKey     interfaces.PrKeyProvider
	logStorage     *lib.Collection[*interfaces.LogStorage]
	sessionStorage *storages.SessionStorage
	sessionRepo    *sessionrepo.SessionRepositoryCached
	morRPC         *msgs.MORRPCMessage
	sessionService SessionService
	log            lib.ILogger
}

func NewProxySender(chainID *big.Int, publicUrl *url.URL, privateKey interfaces.PrKeyProvider, logStorage *lib.Collection[*interfaces.LogStorage], sessionStorage *storages.SessionStorage, sessionRepo *sessionrepo.SessionRepositoryCached, log lib.ILogger) *ProxyServiceSender {
	return &ProxyServiceSender{
		chainID:        chainID,
		publicUrl:      publicUrl,
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

	msg, code, ginErr := p.rpcRequest(providerURL, initiateSessionRequest)
	if ginErr != nil {
		return nil, lib.WrapError(ErrProvider, fmt.Errorf("code: %d, msg: %v, error: %s", code, msg, ginErr))
	}

	if msg.Error != nil {
		// TODO: verify signature
		return nil, lib.WrapError(ErrResponseErr, fmt.Errorf("error: %v, result: %v", msg.Error.Message, msg.Error.Data))
	}
	if msg.Result == nil {
		return nil, lib.WrapError(ErrInvalidResponse, fmt.Errorf("empty result and no error"))
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

	session, ok := p.sessionStorage.GetSession(sessionID.Hex())
	if !ok {
		return nil, ErrSessionNotFound
	}
	provider, ok := p.sessionStorage.GetUser(session.ProviderAddr)
	if !ok {
		return nil, ErrProviderNotFound
	}

	getSessionReportRequest, err := p.morRPC.SessionReportRequest(sessionID, prKey, requestID)
	if err != nil {
		return nil, lib.WrapError(ErrCreateReq, err)
	}

	msg, code, ginErr := p.rpcRequest(provider.Url, getSessionReportRequest)
	if ginErr != nil {
		return nil, lib.WrapError(ErrProvider, fmt.Errorf("code: %d, msg: %v, error: %s", code, msg, ginErr))
	}

	if msg.Error != nil {
		// TODO: verify signature
		return nil, lib.WrapError(ErrResponseErr, fmt.Errorf("error: %v, result: %v", msg.Error.Message, msg.Error.Data))
	}
	if msg.Result == nil {
		return nil, lib.WrapError(ErrInvalidResponse, fmt.Errorf("empty result and no error"))
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
	session, ok := p.sessionStorage.GetSession(sessionID.Hex())
	if !ok {
		return nil, nil, ErrSessionNotFound
	}

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

func (p *ProxyServiceSender) SendPrompt(ctx context.Context, resWriter ResponderFlusher, prompt *openai.ChatCompletionRequest, sessionID common.Hash) (interface{}, error) {
	session, err := p.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	isExpired := session.EndsAt().Int64()-time.Now().Unix() < 0
	if isExpired {
		return nil, ErrSessionExpired
	}

	provider, ok := p.sessionStorage.GetUser(session.ProviderAddr().Hex())
	if !ok {
		return nil, ErrProviderNotFound
	}

	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return nil, ErrMissingPrKey
	}

	requestID := "1"
	pubKey, err := lib.StringToHexString(provider.PubKey)
	if err != nil {
		return nil, lib.WrapError(ErrCreateReq, err)
	}
	promptRequest, err := p.morRPC.SessionPromptRequest(sessionID, prompt, pubKey, prKey, requestID)
	if err != nil {
		return nil, lib.WrapError(ErrCreateReq, err)
	}

	now := time.Now().Unix()
	result, ttftMs, totalTokens, err := p.rpcRequestStream(ctx, resWriter, provider.Url, promptRequest, pubKey)
	if err != nil {
		if !session.FailoverEnabled() {
			return nil, lib.WrapError(ErrProvider, err)
		}

		_, err := p.sessionService.CloseSession(ctx, sessionID)
		if err != nil {
			return nil, err
		}

		resWriter.Header().Set(constants.HEADER_CONTENT_TYPE, constants.CONTENT_TYPE_EVENT_STREAM)
		_, err = resWriter.Write([]byte(fmt.Sprintf("data: %s\n\n", "{\"message\": \"provider failed, failover enabled\"}")))
		if err != nil {
			return nil, err
		}
		resWriter.Flush()

		duration := session.EndsAt().Int64() - time.Now().Unix()

		newSessionID, err := p.sessionService.OpenSessionByModelId(
			ctx,
			session.ModelID(),
			big.NewInt(duration),
			session.FailoverEnabled(),
			session.ProviderAddr(),
		)
		if err != nil {
			return nil, err
		}

		_, err = resWriter.Write([]byte(fmt.Sprintf("data: %s\n\n", "{\"message\": \"new session opened\"}")))
		if err != nil {
			return nil, err
		}
		resWriter.Flush()

		time.Sleep(1 * time.Second) //  sleep for a bit to allow the new session to be created
		return p.SendPrompt(ctx, resWriter, prompt, newSessionID)
	}

	requestDuration := int(time.Now().Unix() - now)
	if requestDuration == 0 {
		requestDuration = 1
	}
	session.AddStats(totalTokens*1000/requestDuration, ttftMs)

	err = p.sessionRepo.SaveSession(ctx, session)
	if err != nil {
		p.log.Error(`failed to update session report stats`, err)
	}

	return result, nil
}

func (p *ProxyServiceSender) rpcRequest(url string, rpcMessage *msgs.RPCMessage) (*msgs.RpcResponse, int, gin.H) {
	TIMEOUT_TO_ESTABLISH_CONNECTION := time.Second * 3
	dialer := net.Dialer{Timeout: TIMEOUT_TO_ESTABLISH_CONNECTION}

	conn, err := dialer.Dial("tcp", url)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to connect to provider"), err)
		p.log.Errorf("%s", err)
		return nil, http.StatusInternalServerError, gin.H{"error": err.Error()}
	}
	defer conn.Close()

	msgJSON, err := json.Marshal(rpcMessage)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to marshal request"), err)
		p.log.Errorf("%s", err)
		return nil, http.StatusInternalServerError, gin.H{"error": err.Error()}
	}
	_, err = conn.Write(msgJSON)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to write request"), err)
		p.log.Errorf("%s", err)
		return nil, http.StatusInternalServerError, gin.H{"error": err.Error()}
	}

	// read response
	reader := bufio.NewReader(conn)
	d := json.NewDecoder(reader)

	var msg *msgs.RpcResponse
	err = d.Decode(&msg)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to decode response"), err)
		p.log.Errorf("%s", err)
		return nil, http.StatusBadRequest, gin.H{"error": err.Error()}
	}
	return msg, 0, nil
}

func (p *ProxyServiceSender) rpcRequestStream(ctx context.Context, resWriter ResponderFlusher, url string, rpcMessage *msgs.RPCMessage, providerPublicKey lib.HexString) (interface{}, int, int, error) {
	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return nil, 0, 0, ErrMissingPrKey
	}

	conn, err := net.Dial("tcp", url)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to connect to provider"), err)
		p.log.Errorf("%s", err)
		return nil, 0, 0, err
	}
	defer conn.Close()

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

	// read response
	reader := bufio.NewReader(conn)
	d := json.NewDecoder(reader)
	resWriter.Header().Set(constants.HEADER_CONTENT_TYPE, constants.CONTENT_TYPE_EVENT_STREAM)

	responses := make([]interface{}, 0)

	for {
		if ctx.Err() != nil {
			return nil, ttftMs, totalTokens, ctx.Err()
		}
		var msg *msgs.RpcResponse
		err = d.Decode(&msg)
		p.log.Debugf("Received stream msg:", msg)
		if err != nil {
			p.log.Warnf("Failed to decode response: %v", err)
			return responses, ttftMs, totalTokens, nil
		}

		if msg.Error != nil {
			return nil, ttftMs, totalTokens, lib.WrapError(ErrResponseErr, fmt.Errorf("error: %v, data: %v", msg.Error.Message, msg.Error.Data))
		}

		if msg.Result == nil {
			return nil, ttftMs, totalTokens, lib.WrapError(ErrInvalidResponse, fmt.Errorf("empty result and no error"))
		}

		if ttftMs == 0 {
			ttftMs = int(time.Now().UnixMilli() - now)
		}

		var inferenceRes InferenceRes
		err := json.Unmarshal(*msg.Result, &inferenceRes)
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

		var payload ChatCompletionResponse
		err = json.Unmarshal(aiResponse, &payload)
		var stop = true
		if err == nil && len(payload.Choices) > 0 {
			stop = false
			choices := payload.Choices
			for _, choice := range choices {
				if choice.FinishReason == FinishReasonStop {
					stop = true
				}
			}
			totalTokens += len(choices)
			responses = append(responses, payload)
		} else {
			var prodiaPayload aiengine.ProdiaGenerationResult
			err = json.Unmarshal(aiResponse, &prodiaPayload)
			if err != nil {
				return nil, ttftMs, totalTokens, lib.WrapError(ErrInvalidResponse, err)
			}
			totalTokens += 1
			responses = append(responses, prodiaPayload)
		}

		if ctx.Err() != nil {
			return nil, ttftMs, totalTokens, ctx.Err()
		}
		_, err = resWriter.Write([]byte(fmt.Sprintf("data: %s\n\n", aiResponse)))
		if err != nil {
			return nil, ttftMs, totalTokens, err
		}
		resWriter.Flush()
		if stop {
			break
		}
	}

	return responses, ttftMs, totalTokens, nil
}

func (p *ProxyServiceSender) validateMsgSignature(result any, signature lib.HexString, providerPubicKey lib.HexString) bool {
	return p.morRPC.VerifySignature(result, signature, providerPubicKey, p.log)
}
