package httphandlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"encoding/hex"
	"fmt"
	"math/big"
	"net/http/pprof"

	constants "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/apibus"
	"github.com/gin-gonic/gin"

	openai "github.com/sashabaranov/go-openai"
)

const (
	SUCCESS_STATUS = 200
	ERROR_STATUS   = 500
)

type HTTPHandler struct{}

func NewHTTPHandler(apiBus *apibus.ApiBus) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.GET("/healthcheck", (func(ctx *gin.Context) {
		ctx.JSON(SUCCESS_STATUS, apiBus.HealthCheck(ctx))
	}))
	r.GET("/config", (func(ctx *gin.Context) {
		ctx.JSON(SUCCESS_STATUS, apiBus.GetConfig(ctx))
	}))
	r.GET("/files", (func(ctx *gin.Context) {
		status, files := apiBus.GetFiles(ctx)
		ctx.JSON(status, files)
	}))
	r.POST("/v1/chat/completions", (func(ctx *gin.Context) {

		var req *openai.ChatCompletionRequest

		err := ctx.ShouldBindJSON(&req)
		switch {
		case errors.Is(err, io.EOF):
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing request body"})
			return
		case err != nil:
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := apiBus.Prompt(ctx, req)
fmt.Println("apibus prompt response: ",response)
		if err != nil {
			ctx.AbortWithError(ERROR_STATUS, err)
			return
		}

		ctx.JSON(SUCCESS_STATUS, response)
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

		status, bids := apiBus.GetBidsByProvider(ctx, providerId, offset, limit)
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
