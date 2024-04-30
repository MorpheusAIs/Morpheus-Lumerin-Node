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

	constants "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/morrpc"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/system"
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

type ProxyRouterApi struct {
	sysConfig     *system.SystemConfigurator
	publicUrl     *url.URL
	pubKey        string
	privateKey    string
	config        Sanitizable
	derivedConfig *config.DerivedConfig
	appStartTime  time.Time
	logStorage    *lib.Collection[*interfaces.LogStorage]
	log           interfaces.ILogger
}

func NewProxyRouterApi(sysConfig *system.SystemConfigurator, publicUrl *url.URL, pubKey string, privateKey string, config Sanitizable, derivedConfig *config.DerivedConfig, appStartTime time.Time, logStorage *lib.Collection[*interfaces.LogStorage], log interfaces.ILogger) *ProxyRouterApi {
	return &ProxyRouterApi{
		sysConfig:     sysConfig,
		publicUrl:     publicUrl,
		pubKey:        pubKey,
		privateKey:    privateKey,
		config:        config,
		derivedConfig: derivedConfig,
		appStartTime:  appStartTime,
		logStorage:    logStorage,
		log:           log,
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

func (p *ProxyRouterApi) HealthCheck(ctx context.Context) gin.H {
	return gin.H{
		"status":  "healthy",
		"version": config.BuildVersion,
		"uptime":  time.Since(p.appStartTime).Round(time.Second).String(),
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

	requestID := "1"

	initiateSessionRequest, err := morrpc.NewMorRpc().InitiateSessionRequest(user, provider, p.pubKey, spend, p.privateKey, requestID)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to create initiate session request"), err)
		p.log.Errorf("%s", err)
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}

	conn, err := net.Dial("tcp", providerUrl)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to connect to provider"), err)
		p.log.Errorf("%s", err)
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}
	defer conn.Close()

	msgJSON, err := json.Marshal(initiateSessionRequest)
	if err != nil {
		err = lib.WrapError(fmt.Errorf("failed to marshal initiate session request"), err)
		p.log.Errorf("%s", err)
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
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
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}

	signature := fmt.Sprintf("%v", msg.Result["signature"])
	providerPubKey := fmt.Sprintf("%v", msg.Result["message"])
	p.log.Debugf("Signature: %s, Provider Pub Key: %s", signature, providerPubKey)

	isValidSignature := morrpc.NewMorRpc().VerifySignature(msg.Result, signature, providerPubKey, p.log)
	p.log.Debugf("Is valid signature: %t", isValidSignature)
	if !isValidSignature {
		err = fmt.Errorf("invalid signature from provider")
		p.log.Errorf("%s", err)
		return constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": err.Error()}
	}

	return constants.HTTP_STATUS_OK, gin.H{
		"response": msg,
	}
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

func writeFiles(writer io.Writer, files []system.FD) error {
	text := fmt.Sprintf("Total: %d\n", len(files))
	text += "\n"
	text += "fd\tpath\n"

	_, err := fmt.Fprintf(writer, text)
	if err != nil {
		return err
	}

	for _, f := range files {
		_, err := fmt.Fprintf(writer, "%s\t%s\n", f.ID, f.Path)
		if err != nil {
			return err
		}
	}

	return nil
}
