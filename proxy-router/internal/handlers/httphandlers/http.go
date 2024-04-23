package httphandlers

import (
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
