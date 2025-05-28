package httphandlers

import (
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

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{"session_id", "model_id", "chat_id", "Authorization", "content-type"},
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
