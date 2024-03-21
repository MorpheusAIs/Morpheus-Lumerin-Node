package httphandlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"net/http/pprof"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/system"
	"github.com/gin-gonic/gin"
)

type Proxy interface {
	SetDest(ctx context.Context, newDestURL *url.URL, onSubmit func(diff float64)) error
}

type Sanitizable interface {
	GetSanitized() any
}

type HTTPHandler struct {
	sysConfig     *system.SystemConfigurator
	publicUrl     *url.URL
	pubKey        string
	config        Sanitizable
	derivedConfig *config.DerivedConfig
	appStartTime  time.Time
	logStorage    *lib.Collection[*interfaces.LogStorage]
	log           interfaces.ILogger
}

func NewHTTPHandler(sysConfig *system.SystemConfigurator, publicUrl *url.URL, config Sanitizable, derivedConfig *config.DerivedConfig, appStartTime time.Time, logStorage *lib.Collection[*interfaces.LogStorage], log interfaces.ILogger) *gin.Engine {
	handl := &HTTPHandler{
		sysConfig:     sysConfig,
		publicUrl:     publicUrl,
		config:        config,
		derivedConfig: derivedConfig,
		appStartTime:  appStartTime,
		logStorage:    logStorage,
		log:           log,
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.GET("/healthcheck", handl.HealthCheck)
	r.GET("/config", handl.GetConfig)
	r.GET("/files", handl.GetFiles)

	r.Any("/debug/pprof/*action", gin.WrapF(pprof.Index))

	err := r.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}

	return r
}

func (h *HTTPHandler) HealthCheck(ctx *gin.Context) {
	ctx.JSON(200, gin.H{
		"status":  "healthy",
		"version": config.BuildVersion,
		"uptime":  time.Since(h.appStartTime).Round(time.Second).String(),
	})
}

func (h *HTTPHandler) GetFiles(ctx *gin.Context) {
	files, err := h.sysConfig.GetFileDescriptors(ctx, os.Getpid())
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(200)

	systemCfg, err := h.sysConfig.GetConfig()
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
		h.log.Errorf("failed to write files: %s", err)
		_ = ctx.Error(err)
		ctx.Abort()
	}
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
