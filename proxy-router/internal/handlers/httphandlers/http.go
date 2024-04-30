package httphandlers

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http/pprof"

	constants "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/apibus"
	"github.com/gin-gonic/gin"
)

type HTTPHandler struct{}

func NewHTTPHandler(apiBus *apibus.ApiBus) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.GET("/healthcheck", (func(ctx *gin.Context) {
		ctx.JSON(constants.HTTP_STATUS_OK, apiBus.HealthCheck(ctx))
	}))
	r.GET("/config", (func(ctx *gin.Context) {
		ctx.JSON(constants.HTTP_STATUS_OK, apiBus.GetConfig(ctx))
	}))
	r.GET("/files", (func(ctx *gin.Context) {
		status, files := apiBus.GetFiles(ctx)
		ctx.JSON(status, files)
	}))

	r.POST("/proxy/sessions/initiate", (func(ctx *gin.Context) {
		status, response := apiBus.InitiateSession(ctx)
		ctx.JSON(status, response)
	}))

	r.POST("/proxy/sessions/:id/prompt", (func(ctx *gin.Context) {
		status, response := apiBus.SendPrompt(ctx)
		ctx.JSON(status, response)
	}))

	r.GET("/blockchain/providers", (func(ctx *gin.Context) {
		status, providers := apiBus.GetAllProviders(ctx)
		ctx.JSON(status, providers)
	}))

	r.GET("/blockchain/providers/:id/bids", (func(ctx *gin.Context) {
		providerId := ctx.Param("id")
		offset, limit := getOffsetLimit(ctx)
		if offset == nil {
			return
		}

		status, bids := apiBus.GetBidsByProdiver(ctx, providerId, offset, limit)
		ctx.JSON(status, bids)
	}))

	r.GET("/blockchain/models", (func(ctx *gin.Context) {
		status, models := apiBus.GetAllModels(ctx)
		ctx.JSON(status, models)
	}))

	r.GET("/blockchain/models/:id/bids", (func(ctx *gin.Context) {
		modelAgentId := ctx.Param("id")

		offset, limit := getOffsetLimit(ctx)
		if offset == nil {
			return
		}

		id, err := hex.DecodeString(modelAgentId)
		if err != nil {
			ctx.JSON(constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": "invalid model agent id"})
			return
		}
		var idBytes [32]byte
		copy(idBytes[:], id)
		status, models := apiBus.GetBidsByModelAgent(ctx, idBytes, offset, limit)
		ctx.JSON(status, models)
	}))

	r.Any("/debug/pprof/*action", gin.WrapF(pprof.Index))

	err := r.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}

	return r
}

func getOffsetLimit(ctx *gin.Context) (*big.Int, uint8) {
	offsetStr := ctx.Query("offset")
	if offsetStr == "" {
		offsetStr = "0"
	}
	limitStr := ctx.Query("limit")
	if limitStr == "" {
		limitStr = "10"
	}

	offset, ok := new(big.Int).SetString(offsetStr, 10)
	if !ok {
		ctx.JSON(constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": "invalid offset"})
		return nil, 0
	}

	var limit uint8
	_, err := fmt.Sscanf(limitStr, "%d", &limit)
	if err != nil {
		ctx.JSON(constants.HTTP_STATUS_BAD_REQUEST, gin.H{"error": "invalid limit"})
		return nil, 0
	}
	return offset, limit
}
