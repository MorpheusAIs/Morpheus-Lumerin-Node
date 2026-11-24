package blockchainapi

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"math/big"
	"net/http"
	"strings"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/registries"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

func gatewaySessionIDPtr(progress *lib.GatewayProgress) *common.Hash {
	if progress == nil {
		return nil
	}
	if progress.SessionID == (common.Hash{}) {
		return nil
	}
	id := progress.SessionID
	return &id
}

type BlockchainController struct {
	service    *BlockchainService
	userStakes userStakesService
	pool       poolService
	payouts    providerPayoutsService
	log        lib.ILogger
	authConf   system.HTTPAuthConfig
}

type userStakesService interface {
	GetUserStakesOnHold(ctx context.Context, iterations uint8) (available, hold *big.Int, err error)
	WithdrawUserStakes(ctx context.Context, iterations uint8) (common.Hash, error)
}

// poolService covers the delegate-staking pool of a hot wallet, which is the
// node wallet when hot is the zero address. On a diamond without the Stage 1
// cut every call fails with registries.ErrStage1NotDeployed.
type poolService interface {
	GetPool(ctx context.Context, hot common.Address) (*structs.PoolRes, error)
	GetPoolFunders(ctx context.Context, hot common.Address, offset *big.Int, limit uint8) (*structs.PoolFundersRes, error)
	GetPoolWithdrawals(ctx context.Context, hot common.Address, cursor common.Address, limit uint8) (*structs.PoolWithdrawalsRes, error)
	GetPoolLocks(ctx context.Context, hot common.Address, offset *big.Int, limit uint8) (*structs.PoolLocksRes, error)
	ApprovePoolFunder(ctx context.Context, funder common.Address) (common.Hash, error)
	RevokePoolFunder(ctx context.Context, funder common.Address) (common.Hash, error)
	ServePoolWithdrawals(ctx context.Context, hot common.Address, maxEntries uint8) (common.Hash, error)
	ReleasePoolLocks(ctx context.Context, hot common.Address, maxBuckets uint8) (common.Hash, error)
	SettleExpiredSession(ctx context.Context, sessionID common.Hash) (common.Hash, error)
}

type providerPayoutsService interface {
	GetProviderPayouts(ctx context.Context, provider common.Address) (*structs.ProviderPayoutsRes, error)
	SetPayoutTarget(ctx context.Context, provider common.Address, target common.Address) (common.Hash, error)
}

func NewBlockchainController(service *BlockchainService, authConf system.HTTPAuthConfig, log lib.ILogger) *BlockchainController {
	c := &BlockchainController{
		service:    service,
		userStakes: service,
		pool:       service,
		payouts:    service,
		log:        log,
		authConf:   authConf,
	}

	return c
}

func (c *BlockchainController) RegisterRoutes(r interfaces.Router) {
	// transactions
	r.GET("/blockchain/balance", c.authConf.CheckAuth("get_balance"), c.getBalance)
	r.GET("/blockchain/transactions", c.authConf.CheckAuth("get_transactions"), c.getTransactions)
	r.GET("/blockchain/allowance", c.authConf.CheckAuth("get_allowance"), c.getAllowance)
	r.GET("/blockchain/latestBlock", c.authConf.CheckAuth("get_latest_block"), c.getLatestBlock)
	r.POST("/blockchain/approve", c.authConf.CheckAuth("approve"), c.approve)
	r.POST("/blockchain/send/eth", c.authConf.CheckAuth("send_eth"), c.sendETH)
	r.POST("/blockchain/send/mor", c.authConf.CheckAuth("send_mor"), c.sendMOR)
	r.GET("/blockchain/stakes/onhold", c.authConf.CheckAuth("get_stakes_on_hold"), c.getUserStakesOnHold)
	r.POST("/blockchain/stakes/withdraw", c.authConf.CheckAuth("withdraw_user_stakes"), c.withdrawUserStakes)

	// providers
	r.GET("/blockchain/providers", c.authConf.CheckAuth("get_providers"), c.getAllProviders)
	r.POST("/blockchain/providers", c.authConf.CheckAuth("create_provider"), c.createProvider)
	r.DELETE("/blockchain/providers/:id", c.authConf.CheckAuth("delete_provider"), c.deregisterProvider)
	r.GET("/blockchain/providers/:id/payouts", c.authConf.CheckAuth("get_providers"), c.getProviderPayouts)
	r.POST("/blockchain/providers/:id/payouts/target", c.authConf.CheckAuth("set_payout_target"), c.setPayoutTarget)

	// models
	r.GET("/blockchain/models", c.authConf.CheckAuth("get_models"), c.getAllModels)
	// Registered before the /:id routes for the same reason /blockchain/sessions/user
	// is: "prices" is a literal segment, not a model id.
	r.GET("/blockchain/models/prices", c.authConf.CheckAuth("get_bids"), c.getModelPrices)
	r.POST("/blockchain/models", c.authConf.CheckAuth("create_model"), c.createNewModel)
	r.DELETE("/blockchain/models/:id", c.authConf.CheckAuth("delete_model"), c.deregisterModel)

	// bids
	r.POST("/blockchain/bids", c.authConf.CheckAuth("create_bid"), c.createNewBid)
	r.GET("/blockchain/bids/:id", c.authConf.CheckAuth("get_bids"), c.getBidByID)
	r.DELETE("/blockchain/bids/:id", c.authConf.CheckAuth("delete_bids"), c.deleteBid)
	r.GET("/blockchain/models/:id/bids", c.authConf.CheckAuth("get_bids"), c.getBidsByModelAgent)
	r.GET("/blockchain/models/:id/bids/rated", c.authConf.CheckAuth("get_bids"), c.getRatedBids)
	r.GET("/blockchain/models/:id/bids/active", c.authConf.CheckAuth("get_bids"), c.getActiveBidsByModel)
	r.GET("/blockchain/providers/:id/bids", c.authConf.CheckAuth("get_bids"), c.getBidsByProvider)
	r.GET("/blockchain/providers/:id/bids/active", c.authConf.CheckAuth("get_bids"), c.getActiveBidsByProvider)

	// sessions
	r.GET("/proxy/sessions/:id/providerClaimableBalance", c.authConf.CheckAuth("get_sessions"), c.getProviderClaimableBalance)
	r.POST("/proxy/sessions/:id/providerClaim", c.authConf.CheckAuth("session_provider_claim"), c.claimProviderBalance)
	r.GET("/blockchain/sessions/user", c.authConf.CheckAuth("get_sessions"), c.getSessionsForUser)
	r.GET("/blockchain/sessions/user/ids", c.authConf.CheckAuth("get_sessions"), c.getSessionsIdsForUser)
	r.GET("/blockchain/sessions/provider", c.authConf.CheckAuth("get_sessions"), c.getSessionsForProvider)
	r.GET("/blockchain/sessions/:id", c.authConf.CheckAuth("get_sessions"), c.getSession)
	r.POST("/blockchain/sessions", c.authConf.CheckAuth("open_session"), c.openSession)
	r.POST("/blockchain/bids/:id/session", c.authConf.CheckAuth("open_session"), c.openSessionByBid)
	r.POST("/blockchain/models/:id/session", c.authConf.CheckAuth("open_session"), c.openSessionByModelId)
	r.GET("/blockchain/models/:id/session/estimate", c.authConf.CheckAuth("get_bids"), c.estimateOpenSession)
	r.POST("/blockchain/sessions/:id/close", c.authConf.CheckAuth("close_session"), c.closeSession)
	r.POST("/blockchain/sessions/:id/settle", c.authConf.CheckAuth("pool_maintenance"), c.settleExpiredSession)
	r.GET("/blockchain/sessions/duration", c.authConf.CheckAuth("get_sessions"), c.getSessionDurationBounds)
	r.GET("/blockchain/sessions/budget", c.authConf.CheckAuth("get_budget"), c.getBudget)
	r.GET("/blockchain/token/supply", c.authConf.CheckAuth("get_supply"), c.getSupply)

	// delegate-staking pool of a hot wallet, the proxy-router wallet unless the
	// request names another one
	r.GET("/blockchain/pool", c.authConf.CheckAuth("get_pool"), c.getPool)
	r.GET("/blockchain/pool/funders", c.authConf.CheckAuth("get_pool"), c.getPoolFunders)
	r.POST("/blockchain/pool/funders", c.authConf.CheckAuth("manage_pool_funders"), c.approvePoolFunder)
	r.DELETE("/blockchain/pool/funders/:id", c.authConf.CheckAuth("manage_pool_funders"), c.revokePoolFunder)
	r.GET("/blockchain/pool/withdrawals", c.authConf.CheckAuth("get_pool"), c.getPoolWithdrawals)
	r.POST("/blockchain/pool/withdrawals/serve", c.authConf.CheckAuth("pool_maintenance"), c.servePoolWithdrawals)
	r.GET("/blockchain/pool/locks", c.authConf.CheckAuth("get_pool"), c.getPoolLocks)
	r.POST("/blockchain/pool/locks/release", c.authConf.CheckAuth("pool_maintenance"), c.releasePoolLocks)
}

// GetUserStakesOnHold godoc
//
//	@Summary		Get consumer stakes on hold
//	@Description	Get the releasable and still time-locked MOR stake for the proxy-router wallet
//	@Tags			transactions
//	@Produce		json
//	@Param			iterations	query		int	false	"Maximum on-hold entries to inspect" default(255) minimum(1) maximum(255)
//	@Success		200			{object}	structs.UserStakesOnHoldRes
//	@Security		BasicAuth
//	@Router			/blockchain/stakes/onhold [get]
func (c *BlockchainController) getUserStakesOnHold(ctx *gin.Context) {
	var query structs.QueryUserStakeIterations
	if err := ctx.ShouldBindQuery(&query); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	available, hold, err := c.userStakes.GetUserStakesOnHold(ctx, structs.UserStakeIterationsOrDefault(query.Iterations))
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.UserStakesOnHoldRes{
		Available: &lib.BigInt{Int: *available},
		Hold:      &lib.BigInt{Int: *hold},
	})
}

// WithdrawUserStakes godoc
//
//	@Summary		Withdraw consumer stakes on hold
//	@Description	Claim releasable MOR stake for the proxy-router wallet
//	@Tags			transactions
//	@Accept			json
//	@Produce		json
//	@Param			request	body		object{iterations=integer}	false	"Withdrawal options; iterations defaults to 255 and accepts 1 through 255"
//	@Success		200		{object}	structs.TxRes
//	@Security		BasicAuth
//	@Router			/blockchain/stakes/withdraw [post]
func (c *BlockchainController) withdrawUserStakes(ctx *gin.Context) {
	progress, progressErr := lib.BeginGatewayOperation(ctx.GetHeader("X-Gateway-Operation"))
	if progressErr != nil {
		ctx.JSON(http.StatusConflict, structs.ErrRes{Error: "operation exists or journal unavailable"})
		return
	}
	ctx.Set(lib.GatewayProgressKey, progress)
	defer func() { progress.Finish(ctx.Writer.Status()) }()

	var request structs.UserStakeWithdrawalRequest
	if err := ctx.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	txHash, err := c.userStakes.WithdrawUserStakes(ctx, structs.UserStakeIterationsOrDefault(request.Iterations))
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: txHash, Progress: progress})
}

// GetProviderClaimableBalance godoc
//
//	@Summary		Get Provider Claimable Balance
//	@Description	Get provider claimable balance from session
//	@Tags			sessions
//	@Produce		json
//	@Param			id	path		string	true	"Session ID"
//	@Success		200	{object}	structs.BalanceRes
//	@Security		BasicAuth
//	@Router			/proxy/sessions/{id}/providerClaimableBalance [get]
func (c *BlockchainController) getProviderClaimableBalance(ctx *gin.Context) {
	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	balance, err := c.service.GetProviderClaimableBalance(ctx, params.ID.Hash)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.BalanceRes{Balance: &lib.BigInt{Int: *balance}})
	return
}

// ClaimProviderBalance godoc
//
//	@Summary		Claim Provider Balance
//	@Description	Claim provider balance from session. On a diamond with the Stage 1 cut anyone may send it, and the reward and any debt owed to the provider go to the provider's payout target, never to the sender.
//	@Tags			sessions
//	@Produce		json
//	@Param			id	path		string	true	"Session ID"
//	@Success		200	{object}	structs.TxRes
//	@Security		BasicAuth
//	@Router			/proxy/sessions/{id}/providerClaim [post]
func (c *BlockchainController) claimProviderBalance(ctx *gin.Context) {
	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	txHash, err := c.service.ClaimProviderBalance(ctx, params.ID.Hash)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: txHash})
	return
}

// GetProviders godoc
//
//	@Summary		Get providers list
//	@Description	Get providers list from blokchain
//	@Tags			providers
//	@Produce		json
//	@Param			request	query		structs.QueryOffsetLimitOrderNoDefault	true	"Query Params"
//	@Success		200		{object}	structs.ProvidersRes
//	@Security		BasicAuth
//	@Router			/blockchain/providers [get]
func (c *BlockchainController) getAllProviders(ctx *gin.Context) {
	offset, limit, order, err := getOffsetLimitOrderNoDefault(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	var providers []*structs.Provider

	if limit == 0 {
		// if pagination is not used return all providers
		// TODO: deprecate this
		providers, err = c.service.GetAllProviders(ctx)
	} else {
		// if pagination is used return providers with offset and limit
		providers, err = c.service.GetProviders(ctx, offset, limit, order)
	}
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.ProvidersRes{Providers: providers})
	return
}

// SendEth godoc
//
//	@Summary		Send Eth
//	@Description	Send Eth to address
//	@Tags			transactions
//	@Security		BasicAuth
//	@Produce		json
//	@Param			sendeth	body		structs.SendRequest	true	"Send Eth"
//	@Success		200		{object}	structs.TxRes
//	@Router			/blockchain/send/eth [post]
func (c *BlockchainController) sendETH(ctx *gin.Context) {
	to, amount, err := c.getSendParams(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	username, ok := ctx.Get("username")
	if !ok {
		c.log.Error("username not found in context")
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: "username not found in context"})
		return
	}

	usernameStr := username.(string)
	txHash, err := c.service.SendETH(ctx, to, amount, usernameStr)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: txHash})
	return
}

// SendMor godoc
//
//	@Summary		Send Mor
//	@Description	Send Mor to address
//	@Tags			transactions
//	@Produce		json
//	@Param			sendmor	body		structs.SendRequest	true	"Send Mor"
//	@Success		200		{object}	structs.TxRes
//	@Router			/blockchain/send/mor [post]
//	@Security		BasicAuth
func (c *BlockchainController) sendMOR(ctx *gin.Context) {
	to, amount, err := c.getSendParams(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	username, ok := ctx.Get("username")
	if !ok {
		c.log.Error("username not found in context")
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: "username not found in context"})
		return
	}
	usernameStr := username.(string)

	txhash, err := c.service.SendMOR(ctx, to, amount, usernameStr)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: txhash})
	return
}

// GetBidsByProvider godoc
//
//	@Summary		Get Bids by Provider
//	@Description	Get bids from blockchain by provider
//	@Tags			bids
//	@Produce		json
//	@Param			id		path		string							true	"Provider ID"
//	@Param			request	query		structs.QueryOffsetLimitOrder	true	"Query Params"
//	@Success		200		{object}	structs.BidsRes
//	@Security		BasicAuth
//	@Router			/blockchain/providers/{id}/bids [get]
func (c *BlockchainController) getBidsByProvider(ctx *gin.Context) {
	var params structs.PathEthAddrID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	offset, limit, order, err := getOffsetLimitOrder(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	bids, err := c.service.GetBidsByProvider(ctx, params.ID.Address, offset, limit, order)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.BidsRes{Bids: bids})
	return
}

// GetActiveBidsByProvider godoc
//
//	@Summary		Get Bids by Provider
//	@Description	Get bids from blockchain by provider
//	@Tags			bids
//	@Produce		json
//	@Param			id		path		string							true	"Provider ID"
//	@Param			request	query		structs.QueryOffsetLimitOrder	true	"Query Params"
//	@Success		200		{object}	structs.BidsRes
//	@Security		BasicAuth
//	@Router			/blockchain/providers/{id}/bids/active [get]
func (c *BlockchainController) getActiveBidsByProvider(ctx *gin.Context) {
	var params structs.PathEthAddrID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	offset, limit, order, err := getOffsetLimitOrder(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	bids, err := c.service.GetActiveBidsByProvider(ctx, params.ID.Address, offset, limit, order)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.BidsRes{Bids: bids})
	return
}

// GetModels godoc
//
//	@Summary		Get models list
//	@Description	Get models list from blokchain
//	@Tags			models
//	@Produce		json
//	@Param			request	query		structs.QueryOffsetLimitOrderNoDefault	true	"Query Params"
//	@Success		200		{object}	structs.ModelsRes
//	@Security		BasicAuth
//	@Router			/blockchain/models [get]
func (c *BlockchainController) getAllModels(ctx *gin.Context) {
	offset, limit, order, err := getOffsetLimitOrderNoDefault(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	var models []*structs.Model
	if limit == 0 {
		// if pagination is not used return all models
		// TODO: deprecate this
		models, err = c.service.GetAllModels(ctx)
	} else {
		// if pagination is used return models with offset and limit
		models, err = c.service.GetModels(ctx, offset, limit, order)
	}
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.ModelsRes{Models: models})
	return
}

// GetModelPrices godoc
//
//	@Summary		Get price index for every model
//	@Description	Get the live price range per second of compute for every registered model, so a client can rank models by cost without one request per model
//	@Tags			models
//	@Produce		json
//	@Success		200	{object}	structs.ModelPricesRes
//	@Failure		500	{object}	structs.ErrRes
//	@Security		BasicAuth
//	@Router			/blockchain/models/prices [get]
func (c *BlockchainController) getModelPrices(ctx *gin.Context) {
	prices, err := c.service.GetModelPrices(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	// Deliberately not sorted here. A model with no live provider has no price to
	// rank, and whether those rows belong at the bottom of a cheap-first list or
	// hidden entirely is the client's decision, not the router's.
	ctx.JSON(http.StatusOK, prices)
}

// GetBidsByModelAgent godoc
//
//	@Summary		Get Bids by	Model Agent
//	@Description	Get bids from blockchain by model agent
//	@Tags			bids
//	@Produce		json
//	@Param			id		path		string							true	"ModelAgent ID"
//	@Param			request	query		structs.QueryOffsetLimitOrder	true	"Query Params"
//	@Success		200		{object}	structs.BidsRes
//	@Security		BasicAuth
//	@Router			/blockchain/models/{id}/bids [get]
func (c *BlockchainController) getBidsByModelAgent(ctx *gin.Context) {
	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	offset, limit, order, err := getOffsetLimitOrder(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	bids, err := c.service.GetBidsByModelAgent(ctx, params.ID.Hash, offset, limit, order)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.BidsRes{Bids: bids})
	return
}

// GetActiveBidsByModel godoc
//
//	@Summary		Get Active Bids by Model
//	@Description	Get bids from blockchain by model agent
//	@Tags			bids
//	@Produce		json
//	@Param			id		path		string							true	"ModelAgent ID"
//	@Param			request	query		structs.QueryOffsetLimitOrder	true	"Query Params"
//	@Success		200		{object}	structs.BidsRes
//	@Security		BasicAuth
//	@Router			/blockchain/models/{id}/bids/active [get]
func (c *BlockchainController) getActiveBidsByModel(ctx *gin.Context) {
	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	offset, limit, order, err := getOffsetLimitOrder(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	bids, err := c.service.GetActiveBidsByModel(ctx, params.ID.Hash, offset, limit, order)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.BidsRes{Bids: bids})
	return
}

// GetBalance godoc
//
//	@Summary		Get ETH and MOR balance
//	@Description	Get ETH and MOR balance of the user
//	@Tags			transactions
//	@Produce		json
//	@Success		200	{object}	structs.TokenBalanceRes
//	@Router			/blockchain/balance [get]
//	@Security		BasicAuth
func (s *BlockchainController) getBalance(ctx *gin.Context) {
	ethBalance, morBalance, err := s.service.GetBalance(ctx)
	if err != nil {
		s.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.TokenBalanceRes{
		ETH: &lib.BigInt{Int: *ethBalance},
		MOR: &lib.BigInt{Int: *morBalance},
	})
	return
}

// GetTransactions godoc
//
//	@Summary		Get Transactions
//	@Description	Get MOR and ETH transactions
//	@Tags			transactions
//	@Produce		json
//	@Param			page	query	string	false	"Page"
//	@Param			limit	query	string	false	"Limit"
//	@Security		BasicAuth
//	@Success		200	{object}	structs.TransactionsRes
//	@Router			/blockchain/transactions [get]
//	@Security		BasicAuth
func (c *BlockchainController) getTransactions(ctx *gin.Context) {
	page, limit, err := getPageLimit(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	txs, err := c.service.GetTransactions(ctx, page, limit)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.TransactionsRes{Transactions: txs})
}

// GetAllowance godoc
//
//	@Summary		Get Allowance for MOR
//	@Description	Get MOR allowance for spender
//	@Tags			transactions
//	@Produce		json
//	@Param			spender	query		string	true	"Spender address"
//	@Success		200		{object}	structs.AllowanceRes
//	@Router			/blockchain/allowance [get]
//	@Security		BasicAuth
func (c *BlockchainController) getAllowance(ctx *gin.Context) {
	var query structs.QuerySpender
	err := ctx.ShouldBindQuery(&query)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	allowance, err := c.service.GetAllowance(ctx, query.Spender.Address)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.AllowanceRes{Allowance: &lib.BigInt{Int: *allowance}})
	return
}

// Approve godoc
//
//	@Summary		Approve MOR allowance
//	@Description	Approve MOR allowance for spender
//	@Tags			transactions
//	@Produce		json
//	@Param			spender	query		string	true	"Spender address"
//	@Param			amount	query		string	true	"Amount"
//	@Success		200		{object}	structs.TxRes
//	@Router			/blockchain/approve [post]
//	@Security		BasicAuth
func (c *BlockchainController) approve(ctx *gin.Context) {
	var query structs.QueryApprove
	err := ctx.ShouldBindQuery(&query)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	tx, err := c.service.Approve(ctx, query.Spender.Address, query.Amount.Unpack())
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: tx})
	return
}

// OpenSession godoc
//
//	@Summary		Open Session with Provider in blockchain
//	@Description	Sends transaction in blockchain to open a session. With fromPool the stake is drawn from the delegate-staking pool of the proxy-router wallet: 400 together with directPayment, 409 when the pool cannot cover the stake, 501 on a diamond without the Stage 1 cut.
//	@Tags			sessions
//	@Produce		json
//	@Accept			json
//	@Param			opensession	body		structs.OpenSessionRequest	true	"Open session"
//	@Success		200			{object}	structs.OpenSessionRes
//	@Failure		400			{object}	structs.ErrRes
//	@Failure		409			{object}	structs.ErrRes
//	@Failure		500			{object}	structs.ErrRes
//	@Failure		501			{object}	structs.ErrRes
//	@Router			/blockchain/sessions [post]
//	@Security		BasicAuth
func (c *BlockchainController) openSession(ctx *gin.Context) {
	var reqPayload structs.OpenSessionRequest
	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	username, ok := ctx.Get("username")
	if !ok {
		c.log.Error("username not found in context")
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: "username not found in context"})
		return
	}
	usernameStr := username.(string)

	sessionId, err := c.service.OpenSession(ctx, reqPayload.Approval, reqPayload.ApprovalSig, reqPayload.Stake.Unpack(), reqPayload.DirectPayment, reqPayload.FromPool, usernameStr, false)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.OpenSessionRes{SessionID: sessionId})
	return
}

// OpenSessionByBidId godoc
//
//	@Summary		Open Session by bidId in blockchain
//	@Description	Full flow to open a session by bidId. 409 carries existingSessionID when rejectExisting finds a live session. With fromPool the stake is drawn from the delegate-staking pool of the proxy-router wallet: 400 together with directPayment, 409 with an error when the pool cannot cover the stake, 501 on a diamond without the Stage 1 cut.
//	@Tags			sessions
//	@Produce		json
//	@Accept			json
//	@Param			opensession	body		structs.OpenSessionWithDurationRequest	true	"Open session"
//	@Param			id			path		string									true	"Bid ID"
//	@Success		200			{object}	structs.OpenSessionRes
//	@Failure		400			{object}	structs.ErrRes
//	@Failure		409			{object}	structs.ExistingSessionRes
//	@Failure		500			{object}	structs.ErrRes
//	@Failure		501			{object}	structs.ErrRes
//	@Router			/blockchain/bids/{id}/session [post]
//	@Security		BasicAuth
func (s *BlockchainController) openSessionByBid(ctx *gin.Context) {
	progress, progressErr := lib.BeginGatewayOperation(ctx.GetHeader("X-Gateway-Operation"))
	if progressErr != nil {
		ctx.JSON(http.StatusConflict, structs.ErrRes{Error: "operation exists or journal unavailable"})
		return
	}
	ctx.Set(lib.GatewayProgressKey, progress)
	defer func() { progress.Finish(ctx.Writer.Status()) }()

	var reqPayload structs.OpenSessionWithDurationRequest
	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error(), SessionID: gatewaySessionIDPtr(progress), Progress: progress})
		return
	}

	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error(), SessionID: gatewaySessionIDPtr(progress), Progress: progress})
		return
	}

	username, ok := ctx.Get("username")
	if !ok {
		s.log.Error("username not found in context")
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: "username not found in context", SessionID: gatewaySessionIDPtr(progress), Progress: progress})
		return
	}
	usernameStr := username.(string)

	if reqPayload.MaxStakeWei != nil {
		ctx.Set(lib.GatewayMaxStakeKey, reqPayload.MaxStakeWei.Unpack())
	}

	var sessionId common.Hash
	if reqPayload.RejectExisting {
		sessionId, err = s.service.OpenSessionByBidIdRejectExisting(ctx, params.ID.Hash, reqPayload.SessionDuration.Unpack(), reqPayload.DirectPayment, reqPayload.FromPool, usernameStr)
	} else {
		sessionId, err = s.service.openSessionByBid(ctx, params.ID.Hash, reqPayload.SessionDuration.Unpack(), reqPayload.DirectPayment, reqPayload.FromPool, usernameStr, nil)
	}
	if err != nil {
		s.log.Error(err)
		var existing *ExistingSessionError
		if errors.As(err, &existing) {
			ctx.JSON(http.StatusConflict, structs.ExistingSessionRes{ExistingSessionID: existing.SessionID})
			return
		}
		ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error(), SessionID: gatewaySessionIDPtr(progress), Progress: progress})
		return
	}

	ctx.JSON(http.StatusOK, structs.OpenSessionRes{SessionID: sessionId, Progress: progress})
	return
}

// EstimateOpenSession godoc
//
//	@Summary		Estimate what opening a session costs
//	@Description	Returns the MOR amount an open would move and every input it was derived from, for the top-scored bid or for one named bid
//	@Tags			sessions
//	@Produce		json
//	@Param			id				path		string	true	"Model ID"
//	@Param			sessionDuration	query		int		true	"Session length in seconds"
//	@Param			directPayment	query		bool	false	"Pay the provider from the escrowed amount instead of staking"
//	@Param			bidId			query		string	false	"Quote this bid instead of the top-scored one"
//	@Param			fromPool		query		bool	false	"Quote an open from the delegate-staking pool of the proxy-router wallet"
//	@Success		200				{object}	structs.OpenSessionStakeEstimate
//	@Failure		400				{object}	structs.ErrRes
//	@Failure		500				{object}	structs.ErrRes
//	@Failure		501				{object}	structs.ErrRes
//	@Router			/blockchain/models/{id}/session/estimate [get]
//	@Security		BasicAuth
func (c *BlockchainController) estimateOpenSession(ctx *gin.Context) {
	var params structs.PathHex32ID
	if err := ctx.ShouldBindUri(&params); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	var query structs.QueryOpenSessionEstimate
	if err := ctx.ShouldBindQuery(&query); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	// HexToHash silently left-pads anything shorter, which would quote the
	// wrong bid rather than fail, so the length is checked first.
	var bidID common.Hash
	if query.BidID != "" {
		if len(common.FromHex(query.BidID)) != common.HashLength {
			ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: "bidId must be a 32 byte hex value"})
			return
		}
		bidID = common.HexToHash(query.BidID)
	}

	estimate, err := c.service.EstimateOpenSessionStake(ctx, params.ID.Hash, bidID, query.SessionDuration.Unpack(), query.DirectPayment, query.FromPool)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, estimate)
}

// GetSessionDurationBounds godoc
//
//	@Summary		Get the session length range the contract accepts
//	@Description	Returns MIN_SESSION_DURATION and getMaxSessionDuration so consumers stop hardcoding them
//	@Tags			sessions
//	@Produce		json
//	@Success		200	{object}	structs.SessionDurationBounds
//	@Router			/blockchain/sessions/duration [get]
//	@Security		BasicAuth
func (c *BlockchainController) getSessionDurationBounds(ctx *gin.Context) {
	bounds, err := c.service.GetSessionDurationBounds(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, bounds)
}

// OpenSessionByModelId godoc
//
//	@Summary		Open Session by ModelID in blockchain
//	@Description	Full flow to open a session by modelId. 409 carries existingSessionID when rejectExisting finds a live session. With fromPool the stake is drawn from the delegate-staking pool of the proxy-router wallet: 400 together with directPayment, 409 with an error when the pool cannot cover the stake of any provider tried, 501 on a diamond without the Stage 1 cut.
//	@Tags			sessions
//	@Produce		json
//	@Accept			json
//	@Param			opensession	body		structs.OpenSessionWithFailover	true	"Open session"
//	@Param			id			path		string							true	"Model ID"
//	@Success		200			{object}	structs.OpenSessionRes
//	@Failure		400			{object}	structs.ErrRes
//	@Failure		409			{object}	structs.ExistingSessionRes
//	@Failure		500			{object}	structs.ErrRes
//	@Failure		501			{object}	structs.ErrRes
//	@Router			/blockchain/models/{id}/session [post]
//	@Security		BasicAuth
func (s *BlockchainController) openSessionByModelId(ctx *gin.Context) {
	var reqPayload structs.OpenSessionWithFailover
	if err := ctx.ShouldBindJSON(&reqPayload); err != nil {
		s.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		s.log.Error(err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, ok := ctx.Get("username")
	if !ok {
		s.log.Error("username not found in context")
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: "username not found in context"})
		return
	}
	usernameStr := username.(string)

	isFailoverEnabled := reqPayload.Failover
	var sessionId common.Hash
	if reqPayload.RejectExisting {
		sessionId, err = s.service.OpenSessionByModelIdRejectExisting(ctx, params.ID.Hash, reqPayload.SessionDuration.Unpack(), reqPayload.DirectPayment, reqPayload.FromPool, isFailoverEnabled, reqPayload.OmitProvider.Address, usernameStr)
	} else {
		sessionId, err = s.service.OpenSessionByModelId(ctx, params.ID.Hash, reqPayload.SessionDuration.Unpack(), reqPayload.DirectPayment, reqPayload.FromPool, isFailoverEnabled, reqPayload.OmitProvider.Address, usernameStr)
	}
	if err != nil {
		s.log.Error(err)
		writeOpenSessionError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, structs.OpenSessionRes{SessionID: sessionId})
	return
}

// writeOpenSessionError is shared by both open routes: either can now be asked
// to reject an existing session, so either can need to answer 409.
func writeOpenSessionError(ctx *gin.Context, err error) {
	var existing *ExistingSessionError
	if errors.As(err, &existing) {
		ctx.JSON(http.StatusConflict, structs.ExistingSessionRes{
			ExistingSessionID: existing.SessionID,
		})
		return
	}
	ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error()})
}

// stage1ErrorStatus is the status of a failed call that can reach the Stage 1
// facets: 400 for a request no chain state would accept, 409 for one the
// chain does not accept now, and 501 on a diamond without the Stage 1 cut, so
// a client can tell an older diamond from a failed chain call.
func stage1ErrorStatus(err error) int {
	switch {
	case errors.Is(err, ErrPoolStakeOnly), errors.Is(err, ErrPoolSelfFunder):
		return http.StatusBadRequest
	case errors.Is(err, ErrNotSettleable), errors.Is(err, ErrPoolCapacity), isStage1PreconditionRevert(err):
		return http.StatusConflict
	case errors.Is(err, registries.ErrStage1NotDeployed):
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

// isStage1PreconditionRevert reports a Stage 1 call the contract rejected for
// the chain state or the sender's rights at that moment, such as a pool that
// cannot cover a stake, a session closed in the meantime, or a payout target
// the contract refuses.
func isStage1PreconditionRevert(err error) bool {
	name := contractErrorName(err)
	switch {
	case strings.HasPrefix(name, "DelegateStaking"):
		return true
	case name == "ProviderPayoutTargetInvalid", name == "InsufficientRightsForOperation", name == "SessionAlreadyClosed":
		return true
	default:
		return false
	}
}

// rejectZeroAddress answers 400 when a path address is the zero address,
// which no funder or provider can be, and reports whether it did.
func rejectZeroAddress(ctx *gin.Context, addr common.Address, what string, progress *lib.GatewayProgress) bool {
	if addr != (common.Address{}) {
		return false
	}
	ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: what + " must not be the zero address", Progress: progress})
	return true
}

// CloseSession godoc
//
//	@Summary		Close Session with Provider
//	@Description	Sends transaction in blockchain to close a session
//	@Tags			sessions
//	@Produce		json
//	@Param			id	path		string	true	"Session ID"
//	@Success		200	{object}	structs.TxRes
//	@Router			/blockchain/sessions/{id}/close [post]
//	@Security		BasicAuth
func (c *BlockchainController) closeSession(ctx *gin.Context) {
	progress, progressErr := lib.BeginGatewayOperation(ctx.GetHeader("X-Gateway-Operation"))
	if progressErr != nil {
		ctx.JSON(http.StatusConflict, structs.ErrRes{Error: "operation exists or journal unavailable"})
		return
	}
	ctx.Set(lib.GatewayProgressKey, progress)
	defer func() { progress.Finish(ctx.Writer.Status()) }()

	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	txHash, err := c.service.CloseSession(ctx, params.ID.Hash)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: txHash, Progress: progress})
	return
}

// GetSession godoc
//
//	@Summary		Get session
//	@Description	Returns session by ID
//	@Tags			sessions
//	@Produce		json
//	@Param			id	path		string	true	"Session ID"
//	@Success		200	{object}	structs.SessionRes
//	@Router			/blockchain/sessions/{id} [get]
//	@Security		BasicAuth
func (c *BlockchainController) getSession(ctx *gin.Context) {
	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	session, err := c.service.GetSession(ctx, params.ID.Hash)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.SessionRes{Session: session})
	return
}

// GetSessions godoc
//
//	@Summary		Get Sessions for User
//	@Description	Get sessions from blockchain by user
//	@Tags			sessions
//	@Produce		json
//	@Param			user	query		string							true	"User address"
//	@Param			request	query		structs.QueryOffsetLimitOrder	true	"Query Params"
//	@Success		200		{object}	structs.SessionsRes
//	@Router			/blockchain/sessions/user [get]
//	@Security		BasicAuth
func (c *BlockchainController) getSessionsForUser(ctx *gin.Context) {
	offset, limit, order, err := getOffsetLimitOrder(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	var req structs.QueryUser
	err = ctx.ShouldBindQuery(&req)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	sessions, err := c.service.GetSessions(ctx, req.User.Address, common.Address{}, offset, limit, order)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.SessionsRes{Sessions: sessions})
	return
}

// GetSessionsIds godoc
//
//	@Summary		Get Sessions for User
//	@Description	Get sessions from blockchain by user
//	@Tags			sessions
//	@Produce		json
//	@Param			user	query		string							true	"User address"
//	@Param			request	query		structs.QueryOffsetLimitOrder	true	"Query Params"
//	@Success		200		{object}	structs.SessionsRes
//	@Router			/blockchain/sessions/user/ids [get]
//	@Security		BasicAuth
func (c *BlockchainController) getSessionsIdsForUser(ctx *gin.Context) {
	offset, limit, order, err := getOffsetLimitOrder(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	var req structs.QueryUser
	err = ctx.ShouldBindQuery(&req)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	sessionsIds, err := c.service.GetSessionsIds(ctx, req.User.Address, common.Address{}, offset, limit, order)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, sessionsIds)
	return
}

// GetSessions godoc
//
//	@Summary		Get Sessions for Provider
//	@Description	Get sessions from blockchain by provider
//	@Tags			sessions
//	@Produce		json
//	@Param			request		query		structs.QueryOffsetLimitOrder	true	"Query Params"
//	@Param			provider	query		string							true	"Provider address"
//	@Success		200			{object}	structs.SessionsRes
//	@Router			/blockchain/sessions/provider [get]
//	@Security		BasicAuth
func (c *BlockchainController) getSessionsForProvider(ctx *gin.Context) {
	offset, limit, order, err := getOffsetLimitOrder(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	var req structs.QueryProvider
	err = ctx.ShouldBindQuery(&req)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	sessions, err := c.service.GetSessions(ctx, common.Address{}, req.Provider.Address, offset, limit, order)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.SessionsRes{Sessions: sessions})
	return
}

// GetTodaysBudget godoc
//
//	@Summary		Get Todays Budget
//	@Description	Get todays budget from blockchain
//	@Tags			sessions
//	@Produce		json
//	@Success		200	{object}	structs.BudgetRes
//	@Router			/blockchain/sessions/budget [get]
//	@Security		BasicAuth
func (s *BlockchainController) getBudget(ctx *gin.Context) {
	budget, err := s.service.GetTodaysBudget(ctx)
	if err != nil {
		s.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.BudgetRes{Budget: &lib.BigInt{Int: *budget}})
	return
}

// GetTokenSupply godoc
//
//	@Summary		Get Token Supply
//	@Description	Get MOR token supply from blockchain
//	@Tags			sessions
//	@Produce		json
//	@Success		200	{object}	structs.SupplyRes
//	@Router			/blockchain/token/supply [get]
//	@Security		BasicAuth
func (s *BlockchainController) getSupply(ctx *gin.Context) {
	supply, err := s.service.GetTokenSupply(ctx)
	if err != nil {
		s.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.SupplyRes{Supply: &lib.BigInt{Int: *supply}})
	return
}

// GetLatestBlock godoc
//
//	@Summary		Get Latest Block
//	@Description	Get latest block number from blockchain
//	@Tags			transactions
//	@Produce		json
//	@Success		200	{object}	structs.BlockRes
//	@Router			/blockchain/latestBlock [get]
//	@Security		BasicAuth
func (c *BlockchainController) getLatestBlock(ctx *gin.Context) {
	block, err := c.service.GetLatestBlock(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, structs.BlockRes{Block: block})
	return
}

// GetBidByID godoc
//
//	@Summary		Get Bid by ID
//	@Description	Get bid from blockchain by ID
//	@Tags			bids
//	@Produce		json
//	@Param			id	path		string	true	"Bid ID"
//	@Success		200	{object}	structs.BidRes
//	@Router			/blockchain/bids/{id} [get]
//	@Security		BasicAuth
func (c *BlockchainController) getBidByID(ctx *gin.Context) {
	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	bid, err := c.service.GetBidByID(ctx, params.ID.Hash)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.BidRes{Bid: bid})
	return
}

// GetRatedBids godoc
//
//	@Summary		Get Rated Bids
//	@Description	Get rated bids from blockchain by model
//	@Tags			bids
//	@Produce		json
//	@Param			id	path		string	true	"Model ID"
//	@Success		200	{object}	structs.ScoredBidsRes
//	@Router			/blockchain/models/{id}/bids/rated [get]
//	@Security		BasicAuth
func (c *BlockchainController) getRatedBids(ctx *gin.Context) {
	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	bids, err := c.service.GetRatedBids(ctx, params.ID.Hash)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.ScoredBidsRes{Bids: bids})
	return
}

// СreateNewProvider godoc
//
//	@Summary	Creates or updates provider in blockchain
//	@Tags		providers
//	@Produce	json
//	@Accept		json
//	@Param		provider	body		structs.CreateProviderRequest	true	"Provider"
//	@Success	200			{object}	structs.ProviderRes
//	@Router		/blockchain/providers [post]
//	@Security	BasicAuth
func (c *BlockchainController) createProvider(ctx *gin.Context) {
	var provider structs.CreateProviderRequest
	if err := ctx.ShouldBindJSON(&provider); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	result, err := c.service.CreateNewProvider(ctx, provider.Stake, provider.Endpoint)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.ProviderRes{Provider: result})
	return
}

// DeregisterProvider godoc
//
//	@Summary	Deregister Provider
//	@Tags		providers
//	@Produce	json
//	@Param		id	path		string	true	"Provider ID"
//	@Success	200	{object}	structs.TxRes
//	@Router		/blockchain/providers/{id} [delete]
//	@Security	BasicAuth
func (c *BlockchainController) deregisterProvider(ctx *gin.Context) {
	txHash, err := c.service.DeregisterProdiver(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: txHash})
	return
}

// CreateNewModel godoc
//
//	@Summary		Creates model in blockchain
//	@Description	If you provide ID in request it will be used as "Base Id" for generation of new model ID. So actual ID will be generated from it, and you will get it in response.
//	@Tags			models
//	@Produce		json
//	@Accept			json
//	@Param			model	body		structs.CreateModelRequest	true	"Model"
//	@Success		200		{object}	structs.ModelRes
//	@Router			/blockchain/models [post]
//	@Security		BasicAuth
func (c *BlockchainController) createNewModel(ctx *gin.Context) {
	var model structs.CreateModelRequest
	if err := ctx.ShouldBindJSON(&model); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	modelType := DetectModelType(model.Tags)
	if modelType == structs.ModelTypeUnknown {
		c.log.Error("Model tags must include a supported type tag (llm, embedding, tts, stt, decisions, video)")
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: "Model tags must include a supported type tag (llm, embedding, tts, stt, decisions, video)"})
		return
	}

	var modelId common.Hash
	if model.ID == "" {
		var hash common.Hash
		_, err := rand.Read(hash[:])
		if err != nil {
			c.log.Error(err)
			ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
			return
		}
		modelId = hash
	} else {
		modelId = common.HexToHash(model.ID)
	}
	ipsfHash := common.HexToHash(model.IpfsID)

	result, err := c.service.CreateNewModel(ctx, modelId, ipsfHash, model.Fee, model.Stake, model.Name, model.Tags)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.ModelRes{Model: result})
	return

}

// DeregisterModel godoc
//
//	@Summary	Deregister Model
//	@Tags		models
//	@Produce	json
//	@Param		id	path		string	true	"Model ID"
//	@Success	200	{object}	structs.TxRes
//	@Router		/blockchain/models/{id} [delete]
//	@Security	BasicAuth
func (c *BlockchainController) deregisterModel(ctx *gin.Context) {
	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	txHash, err := c.service.DeregisterModel(ctx, params.ID.Hash)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: txHash})
	return
}

// CreateBidRequest godoc
//
//	@Summary	Creates bid in blockchain
//	@Tags		bids
//	@Produce	json
//	@Accept		json
//	@Param		bid	body		structs.CreateBidRequest	true	"Bid"
//	@Success	200	{object}	structs.BidRes
//	@Router		/blockchain/bids [post]
//	@Security	BasicAuth
func (c *BlockchainController) createNewBid(ctx *gin.Context) {
	var bid structs.CreateBidRequest
	if err := ctx.ShouldBindJSON(&bid); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	modelId := common.HexToHash(bid.ModelID)
	result, err := c.service.CreateNewBid(ctx, modelId, bid.PricePerSecond)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.BidRes{Bid: result})
	return
}

// DeleteBid godoc
//
//	@Summary	Delete Bid
//	@Tags		bids
//	@Produce	json
//	@Param		id	path		string	true	"Bid ID"
//	@Success	200	{object}	structs.TxRes
//	@Router		/blockchain/bids/{id} [delete]
//	@Security	BasicAuth
func (c *BlockchainController) deleteBid(ctx *gin.Context) {
	var params structs.PathHex32ID
	err := ctx.ShouldBindUri(&params)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	txHash, err := c.service.DeleteBid(ctx, params.ID.Hash)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: txHash})
	return
}

// SettleExpiredSession godoc
//
//	@Summary		Settle an expired pool session
//	@Description	Permission: pool_maintenance. Settles a pool session nobody closed once endsAt + settlementGrace has passed on chain: the stake returns to the pool and the provider reward becomes debt. Anyone may send it. 409 when the session is not a pool session, is already closed or settled, or the grace has not passed. Accepts X-Gateway-Operation.
//	@Tags			sessions
//	@Produce		json
//	@Param			id	path		string	true	"Session ID"
//	@Success		200	{object}	structs.TxRes
//	@Failure		400	{object}	structs.ErrRes
//	@Failure		409	{object}	structs.ErrRes
//	@Failure		500	{object}	structs.ErrRes
//	@Failure		501	{object}	structs.ErrRes
//	@Security		BasicAuth
//	@Router			/blockchain/sessions/{id}/settle [post]
func (c *BlockchainController) settleExpiredSession(ctx *gin.Context) {
	progress, progressErr := lib.BeginGatewayOperation(ctx.GetHeader("X-Gateway-Operation"))
	if progressErr != nil {
		ctx.JSON(http.StatusConflict, structs.ErrRes{Error: "operation exists or journal unavailable"})
		return
	}
	ctx.Set(lib.GatewayProgressKey, progress)
	defer func() { progress.Finish(ctx.Writer.Status()) }()

	var params structs.PathHex32ID
	if err := ctx.ShouldBindUri(&params); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	txHash, err := c.pool.SettleExpiredSession(ctx, params.ID.Hash)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: txHash, Progress: progress})
}

// GetPool godoc
//
//	@Summary		Get a delegate-staking pool
//	@Description	Permission: get_pool. The balances of the pool, the stake its next session can draw, and the delegate-staking parameters in effect with their hard limits. 501 on a diamond without the Stage 1 cut, as for every pool route.
//	@Tags			pool
//	@Produce		json
//	@Param			hot	query		string	false	"Hot wallet address; the proxy-router wallet when omitted"
//	@Success		200	{object}	structs.PoolRes
//	@Failure		400	{object}	structs.ErrRes
//	@Failure		500	{object}	structs.ErrRes
//	@Failure		501	{object}	structs.ErrRes
//	@Security		BasicAuth
//	@Router			/blockchain/pool [get]
func (c *BlockchainController) getPool(ctx *gin.Context) {
	var query structs.QueryPoolHot
	if err := ctx.ShouldBindQuery(&query); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	pool, err := c.pool.GetPool(ctx, query.Hot.Address)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, pool)
}

// GetPoolFunders godoc
//
//	@Summary		List the funders of a delegate-staking pool
//	@Description	Permission: get_pool. Every funder that has funded the pool, active or not, with its allowance.
//	@Tags			pool
//	@Produce		json
//	@Param			hot		query		string	false	"Hot wallet address; the proxy-router wallet when omitted"
//	@Param			offset	query		int		false	"Funders to skip"	default(0)	minimum(0)
//	@Param			limit	query		int		false	"Page size"			default(10)	minimum(1)	maximum(128)
//	@Success		200		{object}	structs.PoolFundersRes
//	@Failure		400		{object}	structs.ErrRes
//	@Failure		500		{object}	structs.ErrRes
//	@Failure		501		{object}	structs.ErrRes
//	@Security		BasicAuth
//	@Router			/blockchain/pool/funders [get]
func (c *BlockchainController) getPoolFunders(ctx *gin.Context) {
	var query structs.QueryPoolFunders
	if err := ctx.ShouldBindQuery(&query); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	funders, err := c.pool.GetPoolFunders(ctx, query.Hot.Address, query.Offset.Unpack(), query.Limit)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, funders)
}

// ApprovePoolFunder godoc
//
//	@Summary		Approve a funder of the delegate-staking pool
//	@Description	Permission: manage_pool_funders. Lets the funder fund the pool of the proxy-router wallet. 400 for the proxy-router wallet itself, which funds its own pool without an approval. Accepts X-Gateway-Operation.
//	@Tags			pool
//	@Accept			json
//	@Produce		json
//	@Param			request	body		structs.PoolFunderRequest	true	"Funder"
//	@Success		200		{object}	structs.TxRes
//	@Failure		400		{object}	structs.ErrRes
//	@Failure		409		{object}	structs.ErrRes
//	@Failure		500		{object}	structs.ErrRes
//	@Failure		501		{object}	structs.ErrRes
//	@Security		BasicAuth
//	@Router			/blockchain/pool/funders [post]
func (c *BlockchainController) approvePoolFunder(ctx *gin.Context) {
	progress, progressErr := lib.BeginGatewayOperation(ctx.GetHeader("X-Gateway-Operation"))
	if progressErr != nil {
		ctx.JSON(http.StatusConflict, structs.ErrRes{Error: "operation exists or journal unavailable"})
		return
	}
	ctx.Set(lib.GatewayProgressKey, progress)
	defer func() { progress.Finish(ctx.Writer.Status()) }()

	var request structs.PoolFunderRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	txHash, err := c.pool.ApprovePoolFunder(ctx, request.Funder)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: txHash, Progress: progress})
}

// RevokePoolFunder godoc
//
//	@Summary		Revoke a funder of the delegate-staking pool
//	@Description	Permission: manage_pool_funders. Stops draws from and top-ups by the funder in the pool of the proxy-router wallet at once. It never blocks the funder's withdrawal. 400 for the zero address and for the proxy-router wallet itself, whose own escrow no approval controls. Accepts X-Gateway-Operation.
//	@Tags			pool
//	@Produce		json
//	@Param			id	path		string	true	"Funder address"
//	@Success		200	{object}	structs.TxRes
//	@Failure		400	{object}	structs.ErrRes
//	@Failure		409	{object}	structs.ErrRes
//	@Failure		500	{object}	structs.ErrRes
//	@Failure		501	{object}	structs.ErrRes
//	@Security		BasicAuth
//	@Router			/blockchain/pool/funders/{id} [delete]
func (c *BlockchainController) revokePoolFunder(ctx *gin.Context) {
	progress, progressErr := lib.BeginGatewayOperation(ctx.GetHeader("X-Gateway-Operation"))
	if progressErr != nil {
		ctx.JSON(http.StatusConflict, structs.ErrRes{Error: "operation exists or journal unavailable"})
		return
	}
	ctx.Set(lib.GatewayProgressKey, progress)
	defer func() { progress.Finish(ctx.Writer.Status()) }()

	var params structs.PathEthAddrID
	if err := ctx.ShouldBindUri(&params); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}
	if rejectZeroAddress(ctx, params.ID.Address, "funder", progress) {
		return
	}

	txHash, err := c.pool.RevokePoolFunder(ctx, params.ID.Address)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: txHash, Progress: progress})
}

// GetPoolWithdrawals godoc
//
//	@Summary		List the withdrawal queue of a delegate-staking pool
//	@Description	Permission: get_pool. A page of the queued funder withdrawals in FIFO order. nextCursor starts the next page and is the zero address once the tail is reached.
//	@Tags			pool
//	@Produce		json
//	@Param			hot		query		string	false	"Hot wallet address; the proxy-router wallet when omitted"
//	@Param			cursor	query		string	false	"Entry to start from, the nextCursor of the previous page; the head of the queue when omitted"
//	@Param			limit	query		int		false	"Page size"	default(10)	minimum(1)	maximum(128)
//	@Success		200		{object}	structs.PoolWithdrawalsRes
//	@Failure		400		{object}	structs.ErrRes
//	@Failure		500		{object}	structs.ErrRes
//	@Failure		501		{object}	structs.ErrRes
//	@Security		BasicAuth
//	@Router			/blockchain/pool/withdrawals [get]
func (c *BlockchainController) getPoolWithdrawals(ctx *gin.Context) {
	var query structs.QueryPoolWithdrawals
	if err := ctx.ShouldBindQuery(&query); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	withdrawals, err := c.pool.GetPoolWithdrawals(ctx, query.Hot.Address, query.Cursor.Address, query.Limit)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, withdrawals)
}

// ServePoolWithdrawals godoc
//
//	@Summary		Pay queued withdrawals of a delegate-staking pool
//	@Description	Permission: pool_maintenance. Pays up to maxEntries queued withdrawals to their funders; the sender never receives funds. maxEntries defaults to 32 and accepts 1 through 32. Accepts X-Gateway-Operation.
//	@Tags			pool
//	@Accept			json
//	@Produce		json
//	@Param			request	body		structs.ServePoolWithdrawalsRequest	false	"Pool and batch size, both optional"
//	@Success		200		{object}	structs.TxRes
//	@Failure		400		{object}	structs.ErrRes
//	@Failure		409		{object}	structs.ErrRes
//	@Failure		500		{object}	structs.ErrRes
//	@Failure		501		{object}	structs.ErrRes
//	@Security		BasicAuth
//	@Router			/blockchain/pool/withdrawals/serve [post]
func (c *BlockchainController) servePoolWithdrawals(ctx *gin.Context) {
	progress, progressErr := lib.BeginGatewayOperation(ctx.GetHeader("X-Gateway-Operation"))
	if progressErr != nil {
		ctx.JSON(http.StatusConflict, structs.ErrRes{Error: "operation exists or journal unavailable"})
		return
	}
	ctx.Set(lib.GatewayProgressKey, progress)
	defer func() { progress.Finish(ctx.Writer.Status()) }()

	var request structs.ServePoolWithdrawalsRequest
	if err := ctx.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	var hot common.Address
	if request.Hot != nil {
		hot = *request.Hot
	}
	txHash, err := c.pool.ServePoolWithdrawals(ctx, hot, structs.PoolBatchSizeOrDefault(request.MaxEntries))
	if err != nil {
		c.log.Error(err)
		ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: txHash, Progress: progress})
}

// GetPoolLocks godoc
//
//	@Summary		List the locked stake of a delegate-staking pool
//	@Description	Permission: get_pool. The day-locked stake of the pool that is not released yet, one bucket per release time.
//	@Tags			pool
//	@Produce		json
//	@Param			hot		query		string	false	"Hot wallet address; the proxy-router wallet when omitted"
//	@Param			offset	query		int		false	"Buckets to skip"	default(0)	minimum(0)
//	@Param			limit	query		int		false	"Page size"			default(10)	minimum(1)	maximum(32)
//	@Success		200		{object}	structs.PoolLocksRes
//	@Failure		400		{object}	structs.ErrRes
//	@Failure		500		{object}	structs.ErrRes
//	@Failure		501		{object}	structs.ErrRes
//	@Security		BasicAuth
//	@Router			/blockchain/pool/locks [get]
func (c *BlockchainController) getPoolLocks(ctx *gin.Context) {
	var query structs.QueryPoolLocks
	if err := ctx.ShouldBindQuery(&query); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}

	locks, err := c.pool.GetPoolLocks(ctx, query.Hot.Address, query.Offset.Unpack(), query.Limit)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, locks)
}

// ReleasePoolLocks godoc
//
//	@Summary		Release matured locks of a delegate-staking pool
//	@Description	Permission: pool_maintenance. Frees up to maxBuckets lock buckets whose release time has passed, then pays up to maxAutoService queued withdrawals to their funders. maxBuckets defaults to 32 and accepts 1 through 32. Accepts X-Gateway-Operation.
//	@Tags			pool
//	@Accept			json
//	@Produce		json
//	@Param			request	body		structs.ReleasePoolLocksRequest	false	"Pool and batch size, both optional"
//	@Success		200		{object}	structs.TxRes
//	@Failure		400		{object}	structs.ErrRes
//	@Failure		409		{object}	structs.ErrRes
//	@Failure		500		{object}	structs.ErrRes
//	@Failure		501		{object}	structs.ErrRes
//	@Security		BasicAuth
//	@Router			/blockchain/pool/locks/release [post]
func (c *BlockchainController) releasePoolLocks(ctx *gin.Context) {
	progress, progressErr := lib.BeginGatewayOperation(ctx.GetHeader("X-Gateway-Operation"))
	if progressErr != nil {
		ctx.JSON(http.StatusConflict, structs.ErrRes{Error: "operation exists or journal unavailable"})
		return
	}
	ctx.Set(lib.GatewayProgressKey, progress)
	defer func() { progress.Finish(ctx.Writer.Status()) }()

	var request structs.ReleasePoolLocksRequest
	if err := ctx.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	var hot common.Address
	if request.Hot != nil {
		hot = *request.Hot
	}
	txHash, err := c.pool.ReleasePoolLocks(ctx, hot, structs.PoolBatchSizeOrDefault(request.MaxBuckets))
	if err != nil {
		c.log.Error(err)
		ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: txHash, Progress: progress})
}

// GetProviderPayouts godoc
//
//	@Summary		Get where a provider is paid and what it is owed
//	@Description	Permission: get_providers. The payout target of the provider, the reward still owed to it, and what it was paid since the Stage 1 upgrade. pendingProviderClaims is the debt to all providers together. 501 on a diamond without the Stage 1 cut.
//	@Tags			providers
//	@Produce		json
//	@Param			id	path		string	true	"Provider address"
//	@Success		200	{object}	structs.ProviderPayoutsRes
//	@Failure		400	{object}	structs.ErrRes
//	@Failure		500	{object}	structs.ErrRes
//	@Failure		501	{object}	structs.ErrRes
//	@Security		BasicAuth
//	@Router			/blockchain/providers/{id}/payouts [get]
func (c *BlockchainController) getProviderPayouts(ctx *gin.Context) {
	var params structs.PathEthAddrID
	if err := ctx.ShouldBindUri(&params); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error()})
		return
	}
	if rejectZeroAddress(ctx, params.ID.Address, "provider", nil) {
		return
	}

	payouts, err := c.payouts.GetProviderPayouts(ctx, params.ID.Address)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, payouts)
}

// SetPayoutTarget godoc
//
//	@Summary		Set where a provider is paid
//	@Description	Permission: set_payout_target, and the caller must have full access ("*"): the target receives the provider's income, so other callers return 403. The zero address or the provider itself resets the target. The contract accepts it only when the proxy-router wallet is the provider or a delegate holding its payout right. Accepts X-Gateway-Operation.
//	@Tags			providers
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Provider address"
//	@Param			request	body		structs.PayoutTargetRequest	true	"Payout target"
//	@Success		200		{object}	structs.TxRes
//	@Failure		400		{object}	structs.ErrRes
//	@Failure		403		{object}	structs.ErrRes
//	@Failure		409		{object}	structs.ErrRes
//	@Failure		500		{object}	structs.ErrRes
//	@Failure		501		{object}	structs.ErrRes
//	@Security		BasicAuth
//	@Router			/blockchain/providers/{id}/payouts/target [post]
func (c *BlockchainController) setPayoutTarget(ctx *gin.Context) {
	progress, progressErr := lib.BeginGatewayOperation(ctx.GetHeader("X-Gateway-Operation"))
	if progressErr != nil {
		ctx.JSON(http.StatusConflict, structs.ErrRes{Error: "operation exists or journal unavailable"})
		return
	}
	ctx.Set(lib.GatewayProgressKey, progress)
	defer func() { progress.Finish(ctx.Writer.Status()) }()

	var params structs.PathEthAddrID
	if err := ctx.ShouldBindUri(&params); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}
	if rejectZeroAddress(ctx, params.ID.Address, "provider", progress) {
		return
	}

	var request structs.PayoutTargetRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusBadRequest, structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	// The permission alone is not enough: the target receives the provider's
	// income, so only the operator redirects it, as with spending allowances.
	if !c.authConf.HasFullAccess(ctx.GetString("username")) {
		ctx.JSON(http.StatusForbidden, structs.ErrRes{Error: "only a full-access user may change a provider's payout target", Progress: progress})
		return
	}

	txHash, err := c.payouts.SetPayoutTarget(ctx, params.ID.Address, *request.Target)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(stage1ErrorStatus(err), structs.ErrRes{Error: err.Error(), Progress: progress})
		return
	}

	ctx.JSON(http.StatusOK, structs.TxRes{Tx: txHash, Progress: progress})
}

// helpers

func (s *BlockchainController) getSendParams(ctx *gin.Context) (to common.Address, amount *big.Int, err error) {
	var body structs.SendRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return common.Address{}, nil, err
	}
	return body.To, body.Amount.Unpack(), nil
}

func getOffsetLimitOrder(ctx *gin.Context) (offset *big.Int, limit uint8, order registries.Order, err error) {
	var paging structs.QueryOffsetLimitOrder

	err = ctx.ShouldBindQuery(&paging)
	if err != nil {
		return nil, 0, false, err
	}

	return paging.Offset.Unpack(), paging.Limit, mapOrder(paging.Order), nil
}

func getOffsetLimitOrderNoDefault(ctx *gin.Context) (offset *big.Int, limit uint8, order registries.Order, err error) {
	var paging structs.QueryOffsetLimitOrderNoDefault

	err = ctx.ShouldBindQuery(&paging)
	if err != nil {
		return nil, 0, false, err
	}

	return paging.Offset.Unpack(), paging.Limit, mapOrder(paging.Order), nil
}

func getPageLimit(ctx *gin.Context) (page uint64, limit uint8, err error) {
	var paging structs.QueryPageLimit

	err = ctx.ShouldBindQuery(&paging)
	if err != nil {
		return 0, 0, err
	}

	return paging.Page, paging.Limit, nil
}
