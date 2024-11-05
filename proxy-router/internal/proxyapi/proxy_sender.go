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

type SessionService interface {
	OpenSessionByModelId(ctx context.Context, modelID common.Hash, duration *big.Int, isFailoverEnabled bool, omitProvider common.Address) (common.Hash, error)
	CloseSession(ctx context.Context, sessionID common.Hash) (common.Hash, error)
}

type ProxyServiceSender struct {
	publicUrl      *url.URL
	privateKey     interfaces.PrKeyProvider
	logStorage     *lib.Collection[*interfaces.LogStorage]
	sessionStorage *storages.SessionStorage
	morRPC         *msgs.MORRPCMessage
	sessionService SessionService
	log            lib.ILogger
}

func NewProxySender(publicUrl *url.URL, privateKey interfaces.PrKeyProvider, logStorage *lib.Collection[*interfaces.LogStorage], sessionStorage *storages.SessionStorage, log lib.ILogger) *ProxyServiceSender {
	return &ProxyServiceSender{
		publicUrl:      publicUrl,
		privateKey:     privateKey,
		logStorage:     logStorage,
		sessionStorage: sessionStorage,
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

func (p *ProxyServiceSender) GetSessionReport(ctx context.Context, sessionID common.Hash) (*msgs.SessionReportRes, error) {
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

func (p *ProxyServiceSender) SendPrompt(ctx context.Context, resWriter ResponderFlusher, prompt *openai.ChatCompletionRequest, sessionID common.Hash) (interface{}, error) {
	session, ok := p.sessionStorage.GetSession(sessionID.Hex())
	if !ok {
		return nil, ErrSessionNotFound
	}

	isExpired := session.EndsAt.Int64()-time.Now().Unix() < 0
	if isExpired {
		return nil, ErrSessionExpired
	}

	provider, ok := p.sessionStorage.GetUser(session.ProviderAddr)
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

	result, err := p.rpcRequestStream(ctx, resWriter, provider.Url, promptRequest, pubKey)
	if err != nil {
		if !session.FailoverEnabled {
			return nil, lib.WrapError(ErrProvider, err)
		}

		// _, err := p.sessionService.CloseSession(ctx, sessionID)
		// if err != nil {
		// 	return nil, err
		// }

		resWriter.Header().Set(constants.HEADER_CONTENT_TYPE, constants.CONTENT_TYPE_EVENT_STREAM)
		_, err = resWriter.Write([]byte(fmt.Sprintf("data: %s\n\n", "{\"message\": \"provider failed, failover enabled\"}")))
		if err != nil {
			return nil, err
		}
		resWriter.Flush()

		modelID := common.HexToHash(session.ModelID)
		provider := common.HexToAddress(session.ProviderAddr)
		duration := session.EndsAt.Int64() - time.Now().Unix()
		durationBigInt := big.NewInt(duration)
		newSessionID, err := p.sessionService.OpenSessionByModelId(ctx, modelID, durationBigInt, session.FailoverEnabled, provider)

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

func (p *ProxyServiceSender) rpcRequestStream(ctx context.Context, resWriter ResponderFlusher, url string, rpcMessage *msgs.RPCMessage, providerPublicKey lib.HexString) (interface{}, error) {
	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return nil, ErrMissingPrKey
	}

	conn, err := net.Dial("tcp", url)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to connect to provider"), err)
		p.log.Errorf("%s", err)
		return nil, err
	}
	defer conn.Close()

	msgJSON, err := json.Marshal(rpcMessage)
	if err != nil {
		return nil, lib.WrapError(ErrMasrshalFailed, err)
	}
	_, err = conn.Write(msgJSON)
	if err != nil {
		return nil, err
	}

	// read response
	reader := bufio.NewReader(conn)
	d := json.NewDecoder(reader)
	resWriter.Header().Set(constants.HEADER_CONTENT_TYPE, constants.CONTENT_TYPE_EVENT_STREAM)

	responses := make([]interface{}, 0)

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		var msg *msgs.RpcResponse
		err = d.Decode(&msg)
		p.log.Debugf("Received stream msg:", msg)
		if err != nil {
			p.log.Warnf("Failed to decode response: %v", err)
			return responses, nil
		}

		if msg.Error != nil {
			return nil, lib.WrapError(ErrResponseErr, fmt.Errorf("error: %v, data: %v", msg.Error.Message, msg.Error.Data))
		}

		if msg.Result == nil {
			return nil, lib.WrapError(ErrInvalidResponse, fmt.Errorf("empty result and no error"))
		}

		var inferenceRes InferenceRes
		err := json.Unmarshal(*msg.Result, &inferenceRes)
		if err != nil {
			return nil, lib.WrapError(ErrInvalidResponse, err)
		}
		sig := inferenceRes.Signature
		inferenceRes.Signature = []byte{}

		if !p.validateMsgSignature(inferenceRes, sig, providerPublicKey) {
			return nil, ErrInvalidSig
		}

		var message lib.HexString
		err = json.Unmarshal(inferenceRes.Message, &message)
		if err != nil {
			return nil, lib.WrapError(ErrInvalidResponse, err)
		}

		aiResponse, err := lib.DecryptBytes(message, prKey)
		if err != nil {
			return nil, lib.WrapError(ErrDecrFailed, err)
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
			responses = append(responses, payload)
		} else {
			var prodiaPayload aiengine.ProdiaGenerationResult
			err = json.Unmarshal(aiResponse, &prodiaPayload)
			if err != nil {
				return nil, lib.WrapError(ErrInvalidResponse, err)
			}
			responses = append(responses, prodiaPayload)
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		_, err = resWriter.Write([]byte(fmt.Sprintf("data: %s\n\n", aiResponse)))
		if err != nil {
			return nil, err
		}
		resWriter.Flush()
		if stop {
			break
		}
	}

	return responses, nil
}

func (p *ProxyServiceSender) validateMsgSignature(result any, signature lib.HexString, providerPubicKey lib.HexString) bool {
	return p.morRPC.VerifySignature(result, signature, providerPubicKey, p.log)
}
