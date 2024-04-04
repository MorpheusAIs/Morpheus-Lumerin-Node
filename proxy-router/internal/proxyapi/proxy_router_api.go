package proxyapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/lib"
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
	config        Sanitizable
	derivedConfig *config.DerivedConfig
	appStartTime  time.Time
	logStorage    *lib.Collection[*interfaces.LogStorage]
	log           interfaces.ILogger
}

func NewProxyRouterApi(sysConfig *system.SystemConfigurator, publicUrl *url.URL, pubKey string, config Sanitizable, derivedConfig *config.DerivedConfig, appStartTime time.Time, logStorage *lib.Collection[*interfaces.LogStorage], log interfaces.ILogger) *ProxyRouterApi {
	return &ProxyRouterApi{
		sysConfig:     sysConfig,
		publicUrl:     publicUrl,
		pubKey:        pubKey,
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
	return 200, gin.H{}
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
