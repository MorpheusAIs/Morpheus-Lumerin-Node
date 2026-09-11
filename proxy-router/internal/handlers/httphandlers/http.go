package httphandlers

import (
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	ginSwagger "github.com/swaggo/gin-swagger"

	// gin-swagger middleware
	swaggerFiles "github.com/swaggo/files"

	_ "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/docs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/config"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
)

type Registrable interface {
	RegisterRoutes(r interfaces.Router)
}

//	@title						Morpheus Lumerin Node API
//	@description				API for Morpheus Lumerin Node
//	@termsOfService				http://swagger.io/terms/
//	@SecurityDefinitions.basic	BasicAuth

//	@BasePath	/

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func CreateHTTPServer(log lib.ILogger, authConfig system.HTTPAuthConfig, controllers ...Registrable) *gin.Engine {
	ginValidatorInstance := binding.Validator.Engine().(*validator.Validate)
	err := config.RegisterHex32(ginValidatorInstance)
	if err != nil {
		panic(err)
	}
	err = config.RegisterDuration(ginValidatorInstance)
	if err != nil {
		panic(err)
	}
	err = config.RegisterEthAddr(ginValidatorInstance)
	if err != nil {
		panic(err)
	}
	err = config.RegisterHexadecimal(ginValidatorInstance)
	if err != nil {
		panic(err)
	}

	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(RequestLogger(log))

	// CORS was previously `AllowOrigins: ["*"]` on an API that exposes wallet
	// operations including /blockchain/send/mor. Basic Auth still gated every
	// route, so this was not directly exploitable by a drive-by page, but a
	// wildcard is the wrong default for an admin surface that is documented as
	// localhost-only (see AGENTS.md: the :8082 port should not be public).
	//
	// Default: any loopback origin, plus Electron's null/file origins. Override
	// with PROXY_CORS_ALLOWED_ORIGINS (comma-separated) if you front the API
	// with something else; set it to "*" to restore the old behaviour.
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: newCORSOriginChecker(log),
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"session_id", "model_id", "chat_id", "x-morpheus-history", "Authorization", "content-type"},
	}))

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	// r.Any("/debug/pprof/*action", gin.WrapF(pprof.Index))

	// r.Use(func(ctx *gin.Context) {
	// 	basicAuth := ctx.GetHeader("Authorization")
	// 	if basicAuth == "" {
	// 		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no basic auth provided"})
	// 		return
	// 	}

	// 	username, password := authConfig.ParseBasicAuthHeader(basicAuth)
	// 	if username == "" || password == "" {
	// 		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid basic auth provided"})
	// 		return
	// 	}

	// 	result := authConfig.IsMethodAllowed(username, "add_user")
	// 	if !result {
	// 		ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "method not allowed"})
	// 		return
	// 	}
	// })

	for _, c := range controllers {
		c.RegisterRoutes(r)
	}

	if err := r.SetTrustedProxies(nil); err != nil {
		panic(err)
	}

	return r
}

// newCORSOriginChecker builds the AllowOriginFunc used above.
//
// Loopback origins are always permitted (the desktop app, local agents and the
// Swagger UI all live there). PROXY_CORS_ALLOWED_ORIGINS adds explicit extra
// origins, and the literal "*" restores the previous allow-everything
// behaviour for anyone who depends on it.
func newCORSOriginChecker(log lib.ILogger) func(origin string) bool {
	raw := strings.TrimSpace(os.Getenv("PROXY_CORS_ALLOWED_ORIGINS"))

	allowAll := false
	extra := map[string]struct{}{}
	for _, o := range strings.Split(raw, ",") {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if o == "*" {
			allowAll = true
			continue
		}
		extra[strings.ToLower(strings.TrimSuffix(o, "/"))] = struct{}{}
	}

	if allowAll {
		log.Warnf("PROXY_CORS_ALLOWED_ORIGINS=* — the API accepts cross-origin requests from any site. Do not use this on a public interface.")
		return func(string) bool { return true }
	}

	if len(extra) > 0 {
		log.Infof("CORS: allowing loopback origins plus %d configured origin(s)", len(extra))
	}

	return func(origin string) bool {
		o := strings.ToLower(strings.TrimSuffix(strings.TrimSpace(origin), "/"))

		// Electron and file:// pages send "null" or no Origin at all.
		if o == "" || o == "null" || strings.HasPrefix(o, "file://") {
			return true
		}

		if _, ok := extra[o]; ok {
			return true
		}

		u, err := url.Parse(o)
		if err != nil {
			return false
		}
		return isLoopbackHost(u.Hostname())
	}
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}
