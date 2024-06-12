package proxyapi

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"time"

	constants "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/morrpc"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/storages"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/gin-gonic/gin"
)

type Sanitizable interface {
	GetSanitized() any
}

type ConfigResponse struct {
	Version       string
	Commit        string
	DerivedConfig interface{}
	Config        interface{}
}

type HealthCheckResponse struct {
	Status  string
	Version string
	Uptime  string
}

type PromptMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PromptRequest struct {
	Messages []PromptMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Model    string          `json:"model"`
}

type RemotePromptRequest struct {
	ProviderPublicKey string
	Prompt            PromptRequest
	ProviderUrl       string
}

type ProxyRouterApi struct {
	sysConfig      *system.SystemConfigurator
	publicUrl      *url.URL
	privateKey     interfaces.PrKeyProvider
	config         Sanitizable
	derivedConfig  *config.DerivedConfig
	appStartTime   time.Time
	logStorage     *lib.Collection[*interfaces.LogStorage]
	sessionStorage *storages.SessionStorage
	log            interfaces.ILogger
}

func NewProxyRouterApi(sysConfig *system.SystemConfigurator, publicUrl *url.URL, privateKey interfaces.PrKeyProvider, config Sanitizable, derivedConfig *config.DerivedConfig, appStartTime time.Time, logStorage *lib.Collection[*interfaces.LogStorage], sessionStorage *storages.SessionStorage, log interfaces.ILogger) *ProxyRouterApi {
	return &ProxyRouterApi{
		sysConfig:      sysConfig,
		publicUrl:      publicUrl,
		privateKey:     privateKey,
		config:         config,
		derivedConfig:  derivedConfig,
		appStartTime:   appStartTime,
		logStorage:     logStorage,
		sessionStorage: sessionStorage,
		log:            log,
	}
}

func (p *ProxyRouterApi) GetConfig(ctx context.Context) ConfigResponse {
	return ConfigResponse{
		Version:       config.BuildVersion,
		Commit:        config.Commit,
		Config:        p.config.GetSanitized(),
		DerivedConfig: p.derivedConfig,
	}
}

func (p *ProxyRouterApi) HealthCheck(ctx context.Context) HealthCheckResponse {
	return HealthCheckResponse{
		Status:  "healthy",
		Version: config.BuildVersion,
		Uptime:  time.Since(p.appStartTime).Round(time.Second).String(),
	}
}

func (p *ProxyRouterApi) InitiateSession(ctx *gin.Context) (int, gin.H) {
	var reqPayload map[string]interface{}
	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}

	user, ok := reqPayload["user"].(string)
	if !ok {
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": "user is required"}
	}

	provider, ok := reqPayload["provider"].(string)
	if !ok {
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": "provider is required"}
	}

	spend, ok := reqPayload["spend"].(float64)
	if !ok {
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": "spend is required"}
	}

	providerUrl, ok := reqPayload["providerUrl"].(string)
	if !ok {
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": "providerUrl is required"}
	}

	bidId, ok := reqPayload["bidId"].(string)
	if !ok {
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": "bidId is required"}
	}

	requestID := "1"

	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return constants.HTTP_INTERNAL_SERVER_ERROR, gin.H{"error": err.Error()}
	}

	initiateSessionRequest, err := morrpc.NewMorRpc().InitiateSessionRequest(user, provider, spend, bidId, prKey, requestID)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to create initiate session request"), err)
		p.log.Errorf("%s", err)
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}

	msg, code, ginErr := p.rpcRequest(providerUrl, initiateSessionRequest)
	if ginErr != nil {
		return code, ginErr
	}

	providerPubKey := fmt.Sprintf("%v", msg.Result["message"])
	if !p.validateMsgSignature(msg, providerPubKey) {
		err = fmt.Errorf("received invalid signature from provider")
		p.log.Errorf("%s", err)
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}

	userObj := storages.User{
		Addr:   provider,
		PubKey: providerPubKey,
		Url:    providerUrl,
	}
	err = p.sessionStorage.AddUser(&userObj)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed store user"), err)
		p.log.Errorf("%s", err)
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}

	return constants.HTTP_STATUS_OK, gin.H{
		"response": msg,
	}
}

func (p *ProxyRouterApi) SendPrompt(ctx *gin.Context) (bool, int, gin.H) {
	var prompt map[string]interface{}
	if err := ctx.ShouldBindJSON(&prompt); err != nil {
		return true, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}

	sessionId := ctx.GetHeader("session_id")
	if sessionId == "" {
		return true, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": "sessionId is required"}
	}

	session, ok := p.sessionStorage.GetSession(sessionId)
	if !ok {
		return true, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": "session not found"}
	}
	provider, ok := p.sessionStorage.GetUser(session.ProviderAddr)
	if !ok {
		return true, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": "provider not found"}
	}

	providerUrl := provider.Url
	providerPublicKey := provider.PubKey
	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return false, constants.HTTP_INTERNAL_SERVER_ERROR, gin.H{"error": err.Error()}
	}

	requestID := "1"
	promptRequest, err := morrpc.NewMorRpc().SessionPromptRequest(sessionId, prompt, providerPublicKey, prKey, requestID)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to create session prompt request"), err)
		p.log.Errorf("%s", err)
		return true, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}

	return p.rpcRequestStream(ctx, providerUrl, promptRequest, providerPublicKey)
}

func (p *ProxyRouterApi) GetFiles(ctx *gin.Context) (int, gin.H) {
	files, err := p.sysConfig.GetFileDescriptors(ctx, os.Getpid())
	if err != nil {
		return 500, gin.H{"error": err.Error()}
	}

	systemCfg, err := p.sysConfig.GetConfig()
	if err != nil {
		fmt.Fprintf(ctx.Writer, "failed to get system config: %s\n", err)
	} else {
		json, err := json.Marshal(systemCfg)
		if err != nil {
			fmt.Fprintf(ctx.Writer, "failed to marshal system config: %s\n", err)
		} else {
			fmt.Fprintf(ctx.Writer, "system config: %s\n", json)
		}
	}
	fmt.Fprintf(ctx.Writer, "\n")

	err = writeFiles(ctx.Writer, files)
	if err != nil {
		p.log.Errorf("failed to write files: %s", err)
		_ = ctx.Error(err)
		ctx.Abort()
	}
	return constants.HTTP_STATUS_OK, gin.H{}
}

func (p *ProxyRouterApi) rpcRequest(url string, rpcMessage *morrpc.RpcMessage) (*morrpc.RpcResponse, int, gin.H) {
	conn, err := net.Dial("tcp", url)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to connect to provider"), err)
		p.log.Errorf("%s", err)
		return nil, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}
	defer conn.Close()

	msgJSON, err := json.Marshal(rpcMessage)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to marshal request"), err)
		p.log.Errorf("%s", err)
		return nil, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}
	conn.Write([]byte(msgJSON))

	// read response
	reader := bufio.NewReader(conn)
	d := json.NewDecoder(reader)
	var msg *morrpc.RpcResponse
	err = d.Decode(&msg)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to decode response"), err)
		p.log.Errorf("%s", err)
		return nil, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}
	return msg, 0, nil
}

func (p *ProxyRouterApi) rpcRequestStream(ctx *gin.Context, url string, rpcMessage *morrpc.RpcMessage, providerPublicKey string) (bool, int, gin.H) {
	prKey, err := p.privateKey.GetPrivateKey()
	if err != nil {
		return false, constants.HTTP_INTERNAL_SERVER_ERROR, gin.H{"error": err.Error()}
	}

	conn, err := net.Dial("tcp", url)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to connect to provider"), err)
		p.log.Errorf("%s", err)
		return true, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}
	defer conn.Close()

	msgJSON, err := json.Marshal(rpcMessage)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to marshal request"), err)
		p.log.Errorf("%s", err)
		return true, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}
	conn.Write([]byte(msgJSON))

	// read response
	reader := bufio.NewReader(conn)
	d := json.NewDecoder(reader)
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")

	for {
		var msg *morrpc.RpcResponse
		err = d.Decode(&msg)
		p.log.Debugf("Received stream msg:", msg)
		if err != nil {
			err = lib.WrapError(fmt.Errorf("failed to decode response"), err)
			p.log.Errorf("%s", err)
			return true, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
		}

		if !p.validateMsgSignature(msg, providerPublicKey) {
			err = fmt.Errorf("received invalid signature from provider")
			p.log.Errorf("%s", err)
			return true, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
		}

		aiResponseEncrypted := msg.Result["message"].(string)
		aiResponse, err := lib.DecryptString(aiResponseEncrypted, prKey)
		if err != nil {
			err = lib.WrapError(fmt.Errorf("failed to decrypt ai response chunk"), err)
			p.log.Errorf("%s", err)
			return true, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
		}

		var payload map[string]interface{}
		err = json.Unmarshal([]byte(aiResponse), &payload)
		if err != nil {
			err = lib.WrapError(fmt.Errorf("failed to unmarshal response"), err)
			p.log.Errorf("%s", err)
			return true, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
		}

		var stop = false
		choices := payload["choices"].([]interface{})
		for _, choice := range choices {
			choiceMap := choice.(map[string]interface{})
			finishReason, ok := choiceMap["finish_reason"].(string)
			if ok && finishReason == "stop" {
				stop = true
			}
		}

		msgJSON, err := json.Marshal(payload)
		if err != nil {
			err = lib.WrapError(fmt.Errorf("failed to marshal response"), err)
			p.log.Errorf("%s", err)
			return true, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
		}
		_, err = ctx.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", msgJSON)))
		if err != nil {
			return true, constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
		}
		ctx.Writer.Flush()
		if stop {
			break
		}
	}

	return false, constants.HTTP_STATUS_OK, gin.H{}
}

func (p *ProxyRouterApi) validateMsgSignature(msg *morrpc.RpcResponse, providerPubicKey string) bool {
	signature := fmt.Sprintf("%v", msg.Result["signature"])

	isValidSignature := morrpc.NewMorRpc().VerifySignature(msg.Result, signature, providerPubicKey, p.log)
	p.log.Debugf("Is valid signature: %t", isValidSignature)
	return isValidSignature
}

func writeFiles(writer io.Writer, files []system.FD) error {
	text := fmt.Sprintf("Total: %d\n", len(files))
	text += "\n"
	text += "fd\tpath\n"

	if _, err := fmt.Fprint(writer, text); err != nil {
		return err
	}

	for _, f := range files {
		if _, err := fmt.Fprintf(writer, "%s\t%s\n", f.ID, f.Path); err != nil {
			return err
		}
	}

	return nil
}
