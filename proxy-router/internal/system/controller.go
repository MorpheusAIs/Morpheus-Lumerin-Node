package system

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	i "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/ethclient"
	"github.com/gin-gonic/gin"
)

// StorageHealthChecker provides health check and metrics for storage backends.
type StorageHealthChecker interface {
	HealthCheck() error
	DBSize() (lsmSize int64, vlogSize int64)
}

// ModelHealthReporter provides cached per-model health self-reports and
// accepts manual re-check triggers.
type ModelHealthReporter interface {
	GetReports() []ModelHealthReport
	// TriggerNow queues an immediate re-check sweep; false means one is
	// already queued. The sweep runs asynchronously.
	TriggerNow() bool
}

// ModelConfigReloader re-reads the models config file without a restart.
type ModelConfigReloader interface {
	Reload() (added []string, removed []string, err error)
}

type SystemController struct {
	config                 *config.Config
	wallet                 i.Wallet
	ethRPC                 i.RPCEndpoints
	sysConfig              *SystemConfigurator
	appStartTime           time.Time
	chainID                *big.Int
	log                    lib.ILogger
	ethConnectionValidator IEthConnectionValidator
	authConfig             HTTPAuthConfig
	storage                StorageHealthChecker
	modelHealth            ModelHealthReporter
	modelConfigs           ModelConfigReloader
}

// SetModelConfigReloader enables POST /config/models/reload. Kept as a setter
// so the constructor signature stays as it is.
func (s *SystemController) SetModelConfigReloader(r ModelConfigReloader) {
	s.modelConfigs = r
}

func NewSystemController(config *config.Config, wallet i.Wallet, ethRPC i.RPCEndpoints, sysConfig *SystemConfigurator, appStartTime time.Time, chainID *big.Int, log lib.ILogger, ethConnectionValidator IEthConnectionValidator, authConfig HTTPAuthConfig, storage StorageHealthChecker, modelHealth ModelHealthReporter) *SystemController {
	c := &SystemController{
		config:                 config,
		wallet:                 wallet,
		ethRPC:                 ethRPC,
		sysConfig:              sysConfig,
		appStartTime:           appStartTime,
		chainID:                chainID,
		log:                    log,
		ethConnectionValidator: ethConnectionValidator,
		authConfig:             authConfig,
		storage:                storage,
		modelHealth:            modelHealth,
	}

	return c
}

func (s *SystemController) RegisterRoutes(r i.Router) {
	r.GET("/healthcheck", s.HealthCheck)
	r.POST("/healthcheck/models/refresh", s.authConfig.CheckAuth("model_health_refresh"), s.RefreshModelHealth)
	r.POST("/config/models/reload", s.authConfig.CheckAuth("system_config"), s.ReloadModelsConfig)
	r.GET("/config", s.authConfig.CheckAuth("system_config"), s.GetConfig)
	r.GET("/files", s.authConfig.CheckAuth("system_config"), s.GetFiles)

	r.POST("/config/ethNode", s.authConfig.CheckAuth("system_config"), s.SetEthNode)
	r.DELETE("/config/ethNode", s.authConfig.CheckAuth("system_config"), s.RemoveEthNode)
}

// HealthCheck godoc
//
//	@Summary		Healthcheck example
//	@Description	do ping
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	HealthCheckResponse
//	@Router			/healthcheck [get]
func (s *SystemController) HealthCheck(ctx *gin.Context) {
	status := "healthy"
	components := make(map[string]string)

	if s.storage != nil {
		if err := s.storage.HealthCheck(); err != nil {
			status = "degraded"
			components["badgerdb"] = fmt.Sprintf("unhealthy: %s", err)
			s.log.Errorf("health check: badgerdb is unhealthy: %s", err)
		} else {
			lsmSize, vlogSize := s.storage.DBSize()
			totalMB := float64(lsmSize+vlogSize) / (1024 * 1024)
			components["badgerdb"] = fmt.Sprintf("healthy, size=%.1fMB", totalMB)
		}
	}

	// model probe failures don't flip the node to 503: bid presence and
	// backend health are marketplace concerns, not process liveness
	var models []ModelHealthReport
	if s.modelHealth != nil {
		models = s.modelHealth.GetReports()
	}

	httpStatus := http.StatusOK
	if status != "healthy" {
		httpStatus = http.StatusServiceUnavailable
	}

	ctx.JSON(httpStatus, HealthCheckResponse{
		Status:     status,
		Version:    config.BuildVersion,
		Uptime:     time.Since(s.appStartTime).Round(time.Second).String(),
		Components: components,
		Models:     models,
	})
}

// RefreshModelHealth godoc
//
//	@Summary		Trigger model health re-check
//	@Description	Queue an immediate model health sweep instead of waiting for the next scheduled run (provider nodes only). Returns immediately; the sweep runs in the background — poll GET /healthcheck and watch models[].lastChecked for fresh results. Probes are paced by MODEL_HEALTH_CHECK_PROBE_DELAY, so a full sweep over many models takes minutes.
//	@Tags			system
//	@Produce		json
//	@Success		202	{object}	StatusRes
//	@Security		BasicAuth
//	@Router			/healthcheck/models/refresh [post]
func (s *SystemController) RefreshModelHealth(ctx *gin.Context) {
	if s.modelHealth == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "model health checks are disabled on this node"})
		return
	}

	if !s.modelHealth.TriggerNow() {
		ctx.JSON(http.StatusAccepted, StatusRes{Status: "refresh already pending"})
		return
	}
	ctx.JSON(http.StatusAccepted, StatusRes{Status: "refresh queued"})
}

// ReloadModelsConfig godoc
//
//	@Summary		Reload models config
//	@Description	Re-read models-config.json and swap the in-memory model table without restarting the node. New models are servable immediately and a health sweep is queued so those with a bid are probed now. A file that fails to parse leaves the running table untouched and returns 400.
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	ModelsReloadRes
//	@Failure		400	{object}	ErrorResponse
//	@Security		BasicAuth
//	@Router			/config/models/reload [post]
func (s *SystemController) ReloadModelsConfig(ctx *gin.Context) {
	if s.modelConfigs == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "models config reload is not wired on this node"})
		return
	}
	added, removed, err := s.modelConfigs.Reload()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("models config not reloaded, previous table still in force: %s", err)})
		return
	}
	res := ModelsReloadRes{Added: added, Removed: removed}
	if s.modelHealth != nil {
		res.HealthSweepQueued = s.modelHealth.TriggerNow()
	}
	ctx.JSON(http.StatusOK, res)
}

// GetConfig godoc
//
//	@Summary		Get Config
//	@Description	Return the current config of proxy router
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	ConfigResponse
//	@Security		BasicAuth
//	@Router			/config [get]
func (s *SystemController) GetConfig(ctx *gin.Context) {
	prkey, err := s.wallet.GetPrivateKey()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	addr, err := lib.PrivKeyBytesToAddr(prkey)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, &ConfigResponse{
		Version: config.BuildVersion,
		Commit:  config.Commit,
		Config:  s.config.GetSanitized(),
		GatewayCapabilities: []string{
			"stake-limit-v1",
			"operation-journal-v1",
			"transaction-progress-v1",
		},
		DerivedConfig: config.DerivedConfig{
			WalletAddress: addr,
			ChainID:       s.chainID,
			EthNodeURLs:   s.ethRPC.GetURLs(),
		},
	})
}

// GetFiles godoc
//
//	@Summary		Get files
//	@Description	Returns opened files
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	[]FD
//	@Security		BasicAuth
//	@Router			/files [get]
func (s *SystemController) GetFiles(ctx *gin.Context) {
	files, err := s.sysConfig.GetFileDescriptors(ctx, os.Getpid())
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	systemCfg, err := s.sysConfig.GetConfig()
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
		s.log.Errorf("failed to write files: %s", err)
		_ = ctx.Error(err)
		ctx.Abort()
	}
	ctx.JSON(http.StatusOK, gin.H{})
	return
}

// SetEthNode godoc
//
//	@Summary		Set Eth Node URLs
//	@Description	Set the Eth Node URLs
//	@Tags			system
//	@Accept			json
//	@Produce		json
//	@Param			urls	body		SetEthNodeURLReq	true	"URLs"
//	@Success		200		{object}	StatusRes
//	@Security		BasicAuth
//	@Router			/config/ethNode [post]
func (s *SystemController) SetEthNode(ctx *gin.Context) {
	var req SetEthNodeURLReq
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, url := range req.URLs {
		validationErr := s.ethConnectionValidator.ValidateEthResourse(ctx, url, time.Second*2)
		if validationErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Resource %s is not available. %s", url, validationErr)})
			return
		}
	}

	err := s.ethRPC.SetURLs(req.URLs)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, OkRes())
}

// DeleteEthNode godoc
//
//	@Summary		Delete Eth Node URLs
//	@Description	Delete the Eth Node URLs
//	@Tags			system
//	@Produce		json
//	@Success		200	{object}	StatusRes
//	@Security		BasicAuth
//	@Router			/config/ethNode [delete]
func (c *SystemController) RemoveEthNode(ctx *gin.Context) {
	err := c.ethRPC.RemoveURLs()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	urls, err := ethclient.GetPublicRPCURLs(int(c.chainID.Int64()))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = c.ethRPC.SetURLsNoPersist(urls)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, OkRes())
}

func writeFiles(writer io.Writer, files []FD) error {
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
