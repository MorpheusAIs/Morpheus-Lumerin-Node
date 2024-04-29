package httphandlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
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
		ctx.JSON(constants.HTTP_STATUS_OK, apiBus.HealthCheck(ctx))
	}))
	r.GET("/config", (func(ctx *gin.Context) {
		ctx.JSON(constants.HTTP_STATUS_OK, apiBus.GetConfig(ctx))
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
			response, err = apiBus.PromptStream(ctx, req)
		} else {
			response, err = apiBus.Prompt(ctx, req)
		}
		
		fmt.Println("apibus prompt response: ", response)
		if err != nil {
			ctx.AbortWithError(ERROR_STATUS, err)
			return
		}

		ctx.JSON(SUCCESS_STATUS, response)
	}))

	r.POST("/sessions/initiate", (func(ctx *gin.Context) {
		status, response := apiBus.InitiateSession(ctx)
		ctx.JSON(status, response)
	}))

	r.Any("/debug/pprof/*action", gin.WrapF(pprof.Index))

	err := r.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}

	return r
}
