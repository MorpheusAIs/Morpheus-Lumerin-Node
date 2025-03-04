package authapi

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authConfig  *system.HTTPAuthConfig
	log         lib.ILogger
	environment string
}

func NewAuthController(authConfig *system.HTTPAuthConfig, environment string, log lib.ILogger) *AuthController {
	a := &AuthController{
		authConfig:  authConfig,
		log:         log,
		environment: environment,
	}

	return a
}

func (s *AuthController) RegisterRoutes(r interfaces.Router) {
	r.POST("/auth/users", s.authConfig.CheckAuth("add_user"), s.AddUser)
	r.DELETE("/auth/users", s.authConfig.CheckAuth("remove_user"), s.DeleteUser)

	r.POST("/auth/users/request", s.RequestAgentUser)
	r.POST("/auth/users/confirm", s.authConfig.CheckAuth("agent_requests"), s.ConfirmAgentRequest)
	r.GET("/auth/users", s.authConfig.CheckAuth("agent_requests"), s.GetAgentUsers)

	r.POST("/auth/allowance/requests", s.authConfig.CheckAuth("request_allowance"), s.RequestAllowance)
	r.POST("/auth/allowance/confirm", s.authConfig.CheckAuth("agent_requests"), s.ConfirmAllowance)
	r.GET("/auth/allowance/requests", s.authConfig.CheckAuth("agent_requests"), s.GetAllowanceRequests)
	r.POST("/auth/allowance/revoke", s.authConfig.CheckAuth("agent_requests"), s.RevokeAllowance)

	r.GET("/auth/users/:username/txs", s.authConfig.CheckAuth("agent_requests"), s.GetAgentTxs)
	r.GET("/auth/cookie/path", s.GetPathToCookieFile)
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

	if req.Username == "admin" {
		err = a.authConfig.UpdateCookieContent(fmt.Sprintf("admin:%s\n", req.Password))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// RemoveUser godoc
//
//	@Summary		Remove User from Proxy API
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

	err := a.authConfig.RemoveAgentUser(req.Username)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// RequestAgentUser godoc
//
//	@Summary	Request New User for Agent
//	@Tags		auth
//	@Produce	json
//	@Param		requestAgentUserReq	body		authapi.RequestAgentUserReq	true	"Request Agent User Request"
//	@Success	200					{object}	authapi.AuthRes
//	@Router		/auth/users/request [post]
func (a *AuthController) RequestAgentUser(ctx *gin.Context) {
	var req *RequestAgentUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := a.authConfig.RequestAgentUser(req.Username, req.Password, req.Perms, req.Allowances)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// GetAgentUsers godoc
//
//	@Summary		Get Agent Users
//	@Description	Permission: agent_requests
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	authapi.AgentUsersRes
//	@Security		BasicAuth
//	@Router			/auth/users [get]
func (a *AuthController) GetAgentUsers(ctx *gin.Context) {
	requests, err := a.authConfig.GetAgentUsers()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var res AgentUsersRes
	res.Agents = make([]*AgentUser, len(requests))

	for i, request := range requests {
		res.Agents[i] = &AgentUser{
			Username:    request.Username,
			Perms:       request.Perms,
			IsConfirmed: request.IsConfirmed,
			Allowances:  request.Allowances,
		}
	}
	ctx.JSON(http.StatusOK, res)
}

// ConfirmAgentRequest godoc
//
//	@Summary		Confirm or Decline Agent User
//	@Description	Permission: agent_requests
//	@Tags			auth
//	@Produce		json
//	@Success		200					{object}	authapi.AuthRes
//	@Param			confirmAgentUserReq	body		authapi.ConfirmAgentReq	true	"Confirm Agent User Request"
//	@Security		BasicAuth
//	@Router			/auth/users/confirm [post]
func (a *AuthController) ConfirmAgentRequest(ctx *gin.Context) {
	var req *ConfirmAgentReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Confirm {
		err := a.authConfig.ConfirmAgentUser(req.Username)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		err := a.authConfig.DeclineAgentUser(req.Username)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// RequestAllowance godoc
//
//	@Summary		Request Allowance for Agent
//	@Description	Permission: request_allowance
//	@Tags			auth
//	@Produce		json
//	@Param			requestAllowanceReq	body		authapi.RequestAllowanceReq	true	"Request Allowance Request with token and amount"
//	@Success		200					{object}	authapi.AuthRes
//	@Security		BasicAuth
//	@Router			/auth/allowance/requests [post]
func (a *AuthController) RequestAllowance(ctx *gin.Context) {
	var req *RequestAllowanceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := a.authConfig.RequestAllowance(req.Username, req.Token, req.Allowance)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// ConfirmAllowance godoc
//
//	@Summary		Confirm or Decline Token Allowance Request
//	@Description	Permission: agent_requests
//	@Tags			auth
//	@Produce		json
//	@Param			confirmAllowanceReq	body		authapi.ConfirmAllowanceReq	true	"Confirm Token Allowance Request"
//	@Success		200					{object}	authapi.AuthRes
//	@Security		BasicAuth
//	@Router			/auth/allowance/confirm [post]
func (a *AuthController) ConfirmAllowance(ctx *gin.Context) {
	var req *ConfirmAllowanceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := a.authConfig.ConfirmOrDeclineAllowanceRequest(req.Username, req.Token, req.Confirm)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// GetAllowanceRequests godoc
//
//	@Summary		Get All Token Allowance Requests
//	@Description	Permission: agent_requests
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	authapi.AllowanceRequestsRes
//	@Security		BasicAuth
//	@Router			/auth/allowance/requests [get]
func (a *AuthController) GetAllowanceRequests(ctx *gin.Context) {
	requests, err := a.authConfig.GetAllowanceRequests()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var res AllowanceRequestsRes
	res.Requests = make([]AllowanceRequest, len(requests))

	for i, request := range requests {
		res.Requests[i] = AllowanceRequest{
			Username:  request.Username,
			Token:     request.Token,
			Allowance: request.Allowance,
		}
	}

	ctx.JSON(http.StatusOK, res)
}

// RevokeAllowance godoc
//
//	@Summary		Revoke Token Allowance for Agent
//	@Description	Permission: agent_requests
//	@Tags			auth
//	@Produce		json
//	@Param			revokeAllowanceReq	body		authapi.RevokeAllowanceReq	true	"Revoke Token Allowance Request"
//	@Success		200					{object}	authapi.AuthRes
//	@Security		BasicAuth
//	@Router			/auth/allowance/revoke [post]
func (a *AuthController) RevokeAllowance(ctx *gin.Context) {
	var req *RevokeAllowanceReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := a.authConfig.RevokeAllowance(req.Username, req.Token)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"result": true})
}

// GetAgentTxs godoc
//
//	@Summary		Get Agent Transactions
//	@Description	Permission: agent_requests
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	authapi.AgentTxsRes
//	@Security		BasicAuth
//	@Router			/auth/users/{username}/txs [get]
func (a *AuthController) GetAgentTxs(ctx *gin.Context) {
	var req AgentTxReqURI
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var query CursorQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	txs, newCursor, err := a.authConfig.GetAgentTxs(req.Username, query.Cursor, query.Limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, AgentTxsRes{TxHashes: txs, NextCursor: newCursor})
}

// GetPathToCookieFile godoc
//
//	@Summary		Get Path to Cookie File
//	@Description	Get the path to the cookie file
//	@Tags			auth
//	@Produce		json
//	@Router			/auth/cookie/path [get]
func (a *AuthController) GetPathToCookieFile(ctx *gin.Context) {
	cookieFilePath := a.authConfig.CookieFilePath
	if a.environment == "development" {
		workingDir, err := os.Getwd()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		cookieFilePath = filepath.Join(workingDir, cookieFilePath)
	} else {
		executablePath, err := os.Executable()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		executableDir := filepath.Dir(executablePath)
		cookieFilePath = filepath.Join(executableDir, cookieFilePath)
	}

	ctx.JSON(http.StatusOK, gin.H{"path": cookieFilePath})
}
