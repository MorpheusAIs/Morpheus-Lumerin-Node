package httphandlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/http/pprof"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/internal/apibus"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-contrib/cors"
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

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
	}))

	r.GET("/healthcheck", (func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, apiBus.HealthCheck(ctx))
	}))
	r.GET("/config", (func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, apiBus.GetConfig(ctx))
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

		req.Stream = ctx.GetHeader("Accept") == "application/json"

		var response interface{}

		if req.Stream {
			response, err = apiBus.PromptStream(ctx, req, func(response *openai.ChatCompletionStreamResponse) error {

				marshalledResponse, err := json.Marshal(response)

				if err != nil {
					return err
				}

				ctx.Writer.Header().Set("Content-Type", "text/event-stream")
				_, err = ctx.Writer.Write([]byte(fmt.Sprintf("data: %s\n\n", marshalledResponse)))

				if err != nil {
					return err
				}

				return nil
			})
		} else {
			response, err = apiBus.Prompt(ctx, req)
		}

		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx.JSON(http.StatusOK, response)
	}))

	r.POST("/proxy/sessions/initiate", (func(ctx *gin.Context) {
		status, response := apiBus.InitiateSession(ctx)
		ctx.JSON(status, response)
	}))

	r.POST("/proxy/sessions/:id/prompt", (func(ctx *gin.Context) {
		ok, status, response := apiBus.SendPrompt(ctx)
		if !ok {
			ctx.JSON(status, response)
			return
		}
		return
	}))

	r.GET("/proxy/sessions/:id/providerClaimableBalance", (func(ctx *gin.Context) {
		status, response := apiBus.GetProviderClaimableBalance(ctx)
		ctx.JSON(status, response)
	}))

	r.GET("/blockchain/providers", (func(ctx *gin.Context) {
		status, providers := apiBus.GetAllProviders(ctx)
		ctx.JSON(status, providers)
	}))

	r.POST("/blockchain/send/eth", (func(ctx *gin.Context) {
		status, response := apiBus.SendEth(ctx)
		ctx.JSON(status, response)
	}))

	r.POST("/blockchain/send/mor", (func(ctx *gin.Context) {
		status, response := apiBus.SendMor(ctx)
		ctx.JSON(status, response)
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

		id := common.FromHex(modelAgentId)

		status, models := apiBus.GetBidsByModelAgent(ctx, ([32]byte)(id), offset, limit)
		ctx.JSON(status, models)
	}))

	r.GET("/blockchain/balance", (func(ctx *gin.Context) {
		status, balance := apiBus.GetBalance(ctx)
		ctx.JSON(status, balance)
	}))

	r.GET("/blockchain/transactions", (func(ctx *gin.Context) {
		status, transactions := apiBus.GetTransactions(ctx)
		ctx.JSON(status, transactions)
	}))

	r.GET("/blockchain/allowance", (func(ctx *gin.Context) {
		status, balance := apiBus.GetAllowance(ctx)
		ctx.JSON(status, balance)
	}))

	r.POST("/blockchain/sessions", (func(ctx *gin.Context) {
		status, response := apiBus.OpenSession(ctx)
		ctx.JSON(status, response)
	}))

	r.GET("/blockchain/sessions/budget", (func(ctx *gin.Context) {
		status, response := apiBus.GetTodaysBudget(ctx)
		ctx.JSON(status, response)
	}))

	r.GET("/blockchain/token/supply", (func(ctx *gin.Context) {
		status, response := apiBus.GetTokenSupply(ctx)
		ctx.JSON(status, response)
	}))

	r.POST("/blockchain/sessions/:id/close", (func(ctx *gin.Context) {
		status, response := apiBus.CloseSession(ctx)
		ctx.JSON(status, response)
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
		limitStr = "100"
	}

	offset, ok := new(big.Int).SetString(offsetStr, 10)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid offset"})
		return nil, 0
	}

	var limit uint8
	_, err := fmt.Sscanf(limitStr, "%d", &limit)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit"})
		return nil, 0
	}
	return offset, limit
}
