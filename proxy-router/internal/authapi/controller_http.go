package authapi

import (
	"net/http"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authConfig *system.HTTPAuthConfig
	log        lib.ILogger
}

func NewAuthController(authConfig *system.HTTPAuthConfig, log lib.ILogger) *AuthController {
	a := &AuthController{
		authConfig: authConfig,
		log:        log,
	}

	return a
}

func (s *AuthController) RegisterRoutes(r interfaces.Router) {
	r.POST("/auth/users", s.authConfig.CheckAuth("add_user"), s.AddUser)
	r.DELETE("/auth/users", s.authConfig.CheckAuth("remove_user"), s.DeleteUser)
}

// AddUser godoc
//
//	@Summary		Add/Update User in Proxy Conf
//	@Description	Permission: add_user
//	@Tags			auth
//	@Produce		json
//	@Param			addUserReq	body		authapi.AddUserReq	true	"Add User Request"
//	@Success		200			{object}	authapi.AuthRes
//	@Security		BasicAuth
//	@Router			/auth/users [post]
func (a *AuthController) AddUser(ctx *gin.Context) {
	var req *AddUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := a.authConfig.AddUser(req.Username, req.Password, req.Perms)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// RemoveUser godoc
//
//	@Summary		Add User to Proxy API
//	@Description	Permission: remove_user
//	@Tags			auth
//	@Produce		json
//	@Param			removeUserReq	body		authapi.RemoveUserReq	true	"Remove User Request"
//	@Success		200				{object}	authapi.AuthRes
//	@Security		BasicAuth
//	@Router			/auth/users [delete]
func (a *AuthController) DeleteUser(ctx *gin.Context) {
	var req *RemoveUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := a.authConfig.RemoveUser(req.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"result": true})
}
