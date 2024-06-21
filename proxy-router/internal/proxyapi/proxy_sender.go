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
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	msg "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi/morrpcmessage"
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
	ErrDecrFailed       = fmt.Errorf("failed to decrypt ai response chunk")
	ErrMasrshalFailed   = fmt.Errorf("failed to marshal response")
	ErrDecode           = fmt.Errorf("failed to decode response")
	ErrSessionNotFound  = fmt.Errorf("session not found")
	ErrProviderNotFound = fmt.Errorf("provider not found")
)

type ProxyServiceSender struct {
	publicUrl      *url.URL
	privateKey     interfaces.PrKeyProvider
	appStartTime   time.Time
	logStorage     *lib.Collection[*interfaces.LogStorage]
	sessionStorage *storages.SessionStorage
	log            lib.ILogger
}

func NewProxySender(publicUrl *url.URL, privateKey interfaces.PrKeyProvider, logStorage *lib.Collection[*interfaces.LogStorage], sessionStorage *storages.SessionStorage, log lib.ILogger) *ProxyServiceSender {
	return &ProxyServiceSender{
		publicUrl:      publicUrl,
		privateKey:     privateKey,
		logStorage:     logStorage,
		sessionStorage: sessionStorage,
		log:            log,
	}
}

func (p *ProxyServiceSender) InitiateSession(ctx context.Context, user common.Address, provider common.Address, spend *big.Int, bidID common.Hash, providerURL string) (*constants.InitiateSessionResponse, error) {
	requestID := "1"

	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return nil, ErrMissingPrKey
	}

	initiateSessionRequest, err := msg.NewMorRpc().InitiateSessionRequest(user, provider, spend, bidID, prKey, requestID)
	if err != nil {
		return nil, lib.WrapError(ErrCreateReq, err)
	}

	msg, code, ginErr := p.rpcRequest(providerURL, initiateSessionRequest)
	if ginErr != nil {
		return nil, lib.WrapError(ErrProvider, fmt.Errorf("code: %s, msg: %s, error", code, msg, ginErr))
	}

	typedMsg, ok := msg.Result.(*constants.InitiateSessionResponse)
	if !ok {
		return nil, lib.WrapError(ErrInvalidResponse, fmt.Errorf("expected InitiateSessionResponse, got %s", msg.Result))
	}

	err = binding.Validator.ValidateStruct(typedMsg)
	if err != nil {
		return nil, lib.WrapError(ErrInvalidResponse, err)
	}

	signature := typedMsg.Signature
	typedMsg.Signature = lib.HexString{}

	providerPubKey := typedMsg.Message
	if !p.validateMsgSignature(msg, signature, typedMsg.Message) {
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

func (p *ProxyServiceSender) SendPrompt(ctx context.Context, resWriter ResponderFlusher, prompt *openai.ChatCompletionRequest, sessionID common.Hash) error {
	session, ok := p.sessionStorage.GetSession(sessionID.Hex())
	if !ok {
		return ErrSessionNotFound
	}

	// TODO: add check for session expiration

	provider, ok := p.sessionStorage.GetUser(session.ProviderAddr)
	if !ok {
		return ErrProviderNotFound
	}

	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return ErrMissingPrKey
	}

	requestID := "1"
	pubKey, err := lib.StringToHexString(provider.PubKey)
	if err != nil {
		return lib.WrapError(ErrCreateReq, err)
	}
	promptRequest, err := msg.NewMorRpc().SessionPromptRequest(sessionID, prompt, pubKey, prKey, requestID)
	if err != nil {
		return lib.WrapError(ErrCreateReq, err)
	}

	return p.rpcRequestStream(ctx, resWriter, provider.Url, promptRequest, pubKey)
}

func (p *ProxyServiceSender) rpcRequest(url string, rpcMessage *msg.RpcMessage) (*msg.RpcResponse, int, gin.H) {
	conn, err := net.Dial("tcp", url)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to connect to provider"), err)
		p.log.Errorf("%s", err)
		return nil, http.StatusBadRequest, gin.H{"error": err.Error()}
	}
	defer conn.Close()

	msgJSON, err := json.Marshal(rpcMessage)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to marshal request"), err)
		p.log.Errorf("%s", err)
		return nil, http.StatusBadRequest, gin.H{"error": err.Error()}
	}
	conn.Write([]byte(msgJSON))

	// read response
	reader := bufio.NewReader(conn)
	d := json.NewDecoder(reader)
	var msg *msg.RpcResponse
	err = d.Decode(&msg)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to decode response"), err)
		p.log.Errorf("%s", err)
		return nil, http.StatusBadRequest, gin.H{"error": err.Error()}
	}
	return msg, 0, nil
}

type ResponderFlusher interface {
	http.ResponseWriter
	http.Flusher
}

func (p *ProxyServiceSender) rpcRequestStream(ctx context.Context, resWriter ResponderFlusher, url string, rpcMessage *msg.RpcMessage, providerPublicKey lib.HexString) error {
	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return ErrMissingPrKey
	}

	conn, err := net.Dial("tcp", url)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to connect to provider"), err)
		p.log.Errorf("%s", err)
		return err
	}
	defer conn.Close()

	msgJSON, err := json.Marshal(rpcMessage)
	if err != nil {
		return lib.WrapError(ErrMasrshalFailed, err)
	}
	_, err = conn.Write(msgJSON)
	if err != nil {
		return err
	}

	// read response
	reader := bufio.NewReader(conn)
	d := json.NewDecoder(reader)
	resWriter.Header().Set(constants.HEADER_CONTENT_TYPE, "text/event-stream")

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		var msg *msg.RpcResponse
		err = d.Decode(&msg)
		p.log.Debugf("Received stream msg:", msg)
		if err != nil {
			return lib.WrapError(ErrDecode, err)
		}

		var inferenceRes InferenceRes
		err := json.Unmarshal([]byte(msg.Result.(string)), &msg)
		if err != nil {
			return lib.WrapError(ErrInvalidResponse, err)
		}
		sig := inferenceRes.Signature
		inferenceRes.Signature = []byte{}

		if !p.validateMsgSignature(inferenceRes, sig, providerPublicKey) {
			return ErrInvalidSig
		}

		aiResponse, err := lib.DecodeBytes(inferenceRes.Message, prKey.String())
		if err != nil {
			return lib.WrapError(ErrDecrFailed, err)
		}

		var payload openai.ChatCompletionResponse
		err = json.Unmarshal(aiResponse, &payload)
		if err != nil {
			return lib.WrapError(ErrInvalidResponse, err)
		}

		var stop = false
		choices := payload.Choices
		for _, choice := range choices {
			if choice.FinishReason == openai.FinishReasonStop {
				stop = true
			}
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
		_, err = resWriter.Write([]byte(fmt.Sprintf("data: %s\n\n", aiResponse)))
		if err != nil {
			return err
		}
		resWriter.Flush()
		if stop {
			break
		}
	}

	return nil
}

func (p *ProxyServiceSender) validateMsgSignature(result any, signature lib.HexString, providerPubicKey lib.HexString) bool {
	isValidSignature := msg.NewMorRpc().VerifySignature(result, signature, providerPubicKey, p.log)
	p.log.Debugf("Is valid signature: %t", isValidSignature)
	return isValidSignature
}
