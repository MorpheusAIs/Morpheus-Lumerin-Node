package blockchainapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/blockchainapi/structs"
	i "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/interfaces"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/lib"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/proxyapi"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/rating"
	m "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/marketplace"
	pr "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/providerregistry"
	s "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/sessionrouter"
	sr "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/contracts/bindings/sessionrouter"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/multicall"
	r "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/registries"
	sessionrepo "github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/repositories/session"
	"github.com/MorpheusAIs/Morpheus-Lumerin-Node/proxy-router/internal/system"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// basefeeWiggleMultiplier is a multiplier for the basefee to set the maxFeePerGas
const basefeeWiggleMultiplier = 2

type BlockchainService struct {
	ethClient          i.EthClient
	providerRegistry   *r.ProviderRegistry
	modelRegistry      *r.ModelRegistry
	marketplace        *r.Marketplace
	sessionRouter      *r.SessionRouter
	morToken           *r.MorToken
	morTokenAddr       common.Address
	explorerClient     ExplorerClientInterface
	sessionRepo        *sessionrepo.SessionRepositoryCached
	proxyService       *proxyapi.ProxyServiceSender
	authConfig         *system.HTTPAuthConfig
	diamonContractAddr common.Address
	rating             *rating.Rating
	minStake           *big.Int

	legacyTx    bool
	privateKey  i.PrKeyProvider
	log         lib.ILogger
	txEscalator *lib.TransactionEscalator
}

type ExplorerClientInterface interface {
	GetLastTransactions(ctx context.Context, address common.Address) ([]structs.MappedTransaction, error)
}

var (
	ErrPrKey              = errors.New("cannot get private key")
	ErrTxOpts             = errors.New("failed to get transactOpts")
	ErrNonce              = errors.New("failed to get nonce")
	ErrEstimateGas        = errors.New("failed to estimate gas")
	ErrSignTx             = errors.New("failed to sign transaction")
	ErrSendTx             = errors.New("failed to send transaction")
	ErrWaitMined          = errors.New("failed to wait for transaction to be mined")
	ErrSessionStore       = errors.New("failed to store session")
	ErrSessionReport      = errors.New("failed to get session report from provider")
	ErrSessionUserReport  = errors.New("failed to get session report from user")
	ErrAgentUserAllowance = errors.New("low agent user allowance")

	ErrBid         = errors.New("failed to get bid")
	ErrProvider    = errors.New("failed to get provider")
	ErrTokenSupply = errors.New("failed to parse token supply")
	ErrBudget      = errors.New("failed to parse token budget")
	ErrMyAddress   = errors.New("failed to get my address")
	ErrInitSession = errors.New("failed to initiate session")
	ErrApprove     = errors.New("failed to approve funds")
	ErrMarshal     = errors.New("failed to marshal open session payload")
	ErrOpenOwnBid  = errors.New("cannot open session with own bid")

	ErrNoBid = errors.New("no bids available")
	ErrModel = errors.New("can't get model")
)

func NewBlockchainService(
	ethClient i.EthClient,
	mc multicall.MulticallBackend,
	diamonContractAddr common.Address,
	morTokenAddr common.Address,
	explorer ExplorerClientInterface,
	privateKey i.PrKeyProvider,
	proxyService *proxyapi.ProxyServiceSender,
	sessionRepo *sessionrepo.SessionRepositoryCached,
	scorerAlgo *rating.Rating,
	authConfig *system.HTTPAuthConfig,
	log lib.ILogger,
	logEthRpc lib.ILogger,
	legacyTx bool,
) *BlockchainService {
	providerRegistry := r.NewProviderRegistry(diamonContractAddr, ethClient, mc, logEthRpc)
	modelRegistry := r.NewModelRegistry(diamonContractAddr, ethClient, mc, logEthRpc)
	marketplace := r.NewMarketplace(diamonContractAddr, ethClient, mc, logEthRpc)
	sessionRouter := r.NewSessionRouter(diamonContractAddr, ethClient, mc, logEthRpc)
	morToken := r.NewMorToken(morTokenAddr, ethClient, logEthRpc)

	// Create transaction escalator for RBF support
	txEscalator := lib.NewTransactionEscalator(ethClient, log, lib.DefaultEscalationConfig())

	return &BlockchainService{
		ethClient:          ethClient,
		providerRegistry:   providerRegistry,
		modelRegistry:      modelRegistry,
		marketplace:        marketplace,
		sessionRouter:      sessionRouter,
		legacyTx:           legacyTx,
		privateKey:         privateKey,
		morToken:           morToken,
		explorerClient:     explorer,
		proxyService:       proxyService,
		diamonContractAddr: diamonContractAddr,
		morTokenAddr:       morTokenAddr,
		sessionRepo:        sessionRepo,
		rating:             scorerAlgo,
		log:                log,
		authConfig:         authConfig,
		txEscalator:        txEscalator,
	}
}

func (s *BlockchainService) GetLatestBlock(ctx context.Context) (uint64, error) {
	return s.ethClient.BlockNumber(ctx)
}

func (s *BlockchainService) GetAllProviders(ctx context.Context) ([]*structs.Provider, error) {
	addrs, providers, err := s.providerRegistry.GetAllProviders(ctx)
	if err != nil {
		return nil, err
	}

	return mapProviders(addrs, providers), nil
}

func (s *BlockchainService) GetProviders(ctx context.Context, offset *big.Int, limit uint8, order r.Order) ([]*structs.Provider, error) {
	addrs, providers, err := s.providerRegistry.GetProviders(ctx, offset, limit, order)
	if err != nil {
		return nil, err
	}

	return mapProviders(addrs, providers), nil
}

// GetMyAddress returns the provider by its wallet address, returns nil if provider is not registered
func (s *BlockchainService) GetProvider(ctx context.Context, providerAddr common.Address) (*structs.Provider, error) {
	provider, err := s.providerRegistry.GetProviderById(ctx, providerAddr)
	if err != nil {
		return nil, err
	}

	if provider.IsDeleted {
		return nil, nil
	}

	if provider.CreatedAt.Cmp(big.NewInt(0)) == 0 {
		return nil, nil
	}

	return mapProvider(providerAddr, *provider), nil
}

func (s *BlockchainService) GetAllModels(ctx context.Context) ([]*structs.Model, error) {
	ids, models, err := s.modelRegistry.GetAllModels(ctx)
	if err != nil {
		return nil, err
	}

	return mapModels(ids, models), nil
}

func (s *BlockchainService) GetModels(ctx context.Context, offset *big.Int, limit uint8, order r.Order) ([]*structs.Model, error) {
	ids, models, err := s.modelRegistry.GetModels(ctx, offset, limit, order)
	if err != nil {
		return nil, err
	}

	return mapModels(ids, models), nil
}

func (s *BlockchainService) GetBidsByProvider(ctx context.Context, providerAddr common.Address, offset *big.Int, limit uint8, order r.Order) ([]*structs.Bid, error) {
	ids, bids, err := s.marketplace.GetBidsByProvider(ctx, providerAddr, offset, limit, order)
	if err != nil {
		return nil, err
	}

	return mapBids(ids, bids), nil
}

func (s *BlockchainService) GetBidsByModelAgent(ctx context.Context, modelId [32]byte, offset *big.Int, limit uint8, order r.Order) ([]*structs.Bid, error) {
	ids, bids, err := s.marketplace.GetBidsByModelAgent(ctx, modelId, offset, limit, order)
	if err != nil {
		return nil, err
	}

	return mapBids(ids, bids), nil
}

func (s *BlockchainService) GetActiveBidsByModel(ctx context.Context, modelId common.Hash, offset *big.Int, limit uint8, order r.Order) ([]*structs.Bid, error) {
	ids, bids, err := s.marketplace.GetActiveBidsByModel(ctx, modelId, offset, limit, order)
	if err != nil {
		return nil, err
	}

	return mapBids(ids, bids), nil
}

func (s *BlockchainService) GetActiveBidsByProviderCount(ctx context.Context, provider common.Address) (*big.Int, error) {
	return s.marketplace.GetActiveBidsByProviderCount(ctx, provider)
}

func (s *BlockchainService) GetActiveBidsByProvider(ctx context.Context, provider common.Address, offset *big.Int, limit uint8, order r.Order) ([]*structs.Bid, error) {
	ids, bids, err := s.marketplace.GetActiveBidsByProvider(ctx, provider, offset, limit, order)
	if err != nil {
		return nil, err
	}

	return mapBids(ids, bids), nil
}

func (s *BlockchainService) GetBidByID(ctx context.Context, ID common.Hash) (*structs.Bid, error) {
	bid, err := s.marketplace.GetBidById(ctx, ID)
	if err != nil {
		return nil, err
	}

	return mapBid(ID, *bid), nil
}

func (s *BlockchainService) GetRatedBids(ctx context.Context, modelID common.Hash) ([]structs.ScoredBid, error) {
	modelStats, err := s.sessionRouter.GetModelStats(ctx, modelID)
	if err != nil {
		return nil, err
	}

	bidIDs, bids, providerModelStats, provider, err := s.GetAllBidsWithRating(ctx, modelID)
	if err != nil {
		return nil, err
	}
	minStake, err := s.getMinStakeCached(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get min stake: %w", err)
	}

	ratedBids := s.rateBids(bidIDs, bids, providerModelStats, provider, modelStats, minStake, s.log)

	return ratedBids, nil
}

func (s *BlockchainService) rateBids(bidIds [][32]byte, bids []m.IBidStorageBid, pmStats []s.IStatsStorageProviderModelStats, provider []pr.IProviderStorageProvider, mStats *s.IStatsStorageModelStats, minStake *big.Int, log lib.ILogger) []structs.ScoredBid {
	ratingInputs := make([]rating.RatingInput, len(bids))
	bidIDIndexMap := make(map[common.Hash]int)

	for i := range bids {
		ratingInputs[i] = rating.RatingInput{
			ScoreInput: rating.ScoreInput{
				ProviderModel:  &pmStats[i],
				Model:          mStats,
				ProviderStake:  provider[i].Stake,
				PricePerSecond: bids[i].PricePerSecond,
				MinStake:       minStake,
			},
			BidID:      bidIds[i],
			ModelID:    bids[i].ModelId,
			ProviderID: bids[i].Provider,
		}
		bidIDIndexMap[bidIds[i]] = i
	}

	result := s.rating.RateBids(ratingInputs, log)
	scoredBids := make([]structs.ScoredBid, len(result))

	for i, score := range result {
		inputBidIndex := bidIDIndexMap[score.BidID]
		scoredBid := structs.ScoredBid{
			Bid: structs.Bid{
				Id:             bidIds[inputBidIndex],
				Provider:       bids[inputBidIndex].Provider,
				ModelAgentId:   bids[inputBidIndex].ModelId,
				PricePerSecond: &lib.BigInt{Int: *(bids[inputBidIndex].PricePerSecond)},
				Nonce:          &lib.BigInt{Int: *(bids[inputBidIndex].Nonce)},
				CreatedAt:      &lib.BigInt{Int: *(bids[inputBidIndex].CreatedAt)},
				DeletedAt:      &lib.BigInt{Int: *(bids[inputBidIndex].DeletedAt)},
			},
			Score: score.Score,
		}
		scoredBids[i] = scoredBid
	}

	sort.Slice(scoredBids, func(i, j int) bool {
		return scoredBids[i].Score > scoredBids[j].Score
	})

	return scoredBids
}

func (s *BlockchainService) OpenSession(ctx context.Context, approval, approvalSig []byte, stake *big.Int, directPayment bool, agentUsername string) (common.Hash, error) {
	isAgent, err := s.authConfig.IsAllowanceEnough(agentUsername, s.morTokenAddr.Hex(), stake)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrAgentUserAllowance, err)
	}

	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrPrKey, err)
	}

	addr, err := lib.PrivKeyBytesToAddr(prKey)
	if err != nil {
		return common.Hash{}, err
	}

	// Step 1: Approve MOR tokens with escalation
	approveBaseOpts, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrTxOpts, err)
	}

	approveReceipt, err := s.txEscalator.SendWithEscalation(
		ctx,
		approveBaseOpts,
		func(opts *bind.TransactOpts) (*types.Transaction, error) {
			return s.morToken.IncreaseAllowanceTx(opts, s.diamonContractAddr, stake)
		},
		s.legacyTx,
	)
	if err != nil {
		s.handleTxError(ctx, addr, err)
		return common.Hash{}, lib.WrapError(ErrSendTx, fmt.Errorf("approve failed: %w", err))
	}

	// Check approval succeeded (escalator already waited for mining)
	if approveReceipt.Status != 1 {
		return common.Hash{}, lib.WrapError(ErrSendTx, fmt.Errorf("approve tx failed with status %d", approveReceipt.Status))
	}

	if isAgent {
		err = s.authConfig.AuthStorage.SetAgentTx(approveReceipt.TxHash.Hex(), agentUsername, approveReceipt.BlockNumber)
		if err != nil {
			s.log.Errorf("failed to set agent tx: %s", err)
		}
	}

	// Step 2: Open session with escalation
	sessionBaseOpts, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrTxOpts, err)
	}

	sessionReceipt, err := s.txEscalator.SendWithEscalation(
		ctx,
		sessionBaseOpts,
		func(opts *bind.TransactOpts) (*types.Transaction, error) {
			return s.sessionRouter.OpenSessionTx(opts, approval, approvalSig, stake, directPayment)
		},
		s.legacyTx,
	)
	if err != nil {
		s.handleTxError(ctx, addr, err)
		return common.Hash{}, lib.WrapError(ErrSendTx, fmt.Errorf("open session failed: %w", err))
	}

	// Parse session info from receipt
	sessionID, _, _, err := s.sessionRouter.ParseOpenSessionReceipt(ctx, sessionReceipt)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrSendTx, fmt.Errorf("parse session receipt failed: %w", err))
	}

	if isAgent {
		amountBigInt := lib.BigInt{Int: *stake}
		err = s.authConfig.DecreaseAllowance(agentUsername, s.morTokenAddr.Hex(), amountBigInt)
		if err != nil {
			s.log.Errorf("failed to decrease allowance: %s", err)
			return common.Hash{}, err
		}
		err = s.authConfig.AuthStorage.SetAgentTx(sessionReceipt.TxHash.Hex(), agentUsername, sessionReceipt.BlockNumber)
		if err != nil {
			s.log.Errorf("failed to set agent tx: %s", err)
		}
	}

	session, err := s.sessionRepo.GetSession(ctx, sessionID)
	if err != nil {
		return sessionID, fmt.Errorf("failed to get session: %s", err.Error())
	}

	session.SetAgentUsername(agentUsername)

	err = s.sessionRepo.SaveSession(ctx, session)
	if err != nil {
		return sessionID, fmt.Errorf("failed to store session: %s", err.Error())
	}

	return sessionID, err
}

func (s *BlockchainService) CreateNewProvider(ctx context.Context, stake *lib.BigInt, endpoint string) (*structs.Provider, error) {
	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return nil, lib.WrapError(ErrPrKey, err)
	}

	transactOpt, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return nil, lib.WrapError(ErrTxOpts, err)
	}

	// First increase allowance for the stake (safe for concurrent use)
	_, err = s.increaseAllowance(ctx, s.diamonContractAddr, &stake.Int)
	if err != nil {
		return nil, lib.WrapError(ErrApprove, err)
	}

	err = s.providerRegistry.CreateNewProvider(transactOpt, stake, endpoint)
	if err != nil {
		s.handleTxError(ctx, transactOpt.From, err)
		return nil, lib.WrapError(ErrSendTx, err)
	}

	provider, err := s.providerRegistry.GetProviderById(ctx, transactOpt.From)
	if err != nil {
		return nil, lib.WrapError(ErrProvider, err)
	}

	return &structs.Provider{
		Address:   transactOpt.From,
		Endpoint:  provider.Endpoint,
		Stake:     &lib.BigInt{Int: *provider.Stake},
		IsDeleted: provider.IsDeleted,
		CreatedAt: &lib.BigInt{Int: *provider.CreatedAt},
	}, nil
}

func (s *BlockchainService) CreateNewModel(ctx context.Context, modelID common.Hash, ipfsID common.Hash, fee *lib.BigInt, stake *lib.BigInt, name string, tags []string) (*structs.Model, error) {
	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return nil, lib.WrapError(ErrPrKey, err)
	}

	transactOpt, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return nil, lib.WrapError(ErrTxOpts, err)
	}

	// First increase allowance for the stake (safe for concurrent use)
	_, err = s.increaseAllowance(ctx, s.diamonContractAddr, &stake.Int)
	if err != nil {
		return nil, lib.WrapError(ErrApprove, err)
	}

	err = s.modelRegistry.CreateNewModel(transactOpt, modelID, ipfsID, fee, stake, name, tags)
	if err != nil {
		return nil, lib.WrapError(ErrSendTx, err)
	}

	ID, err := s.modelRegistry.GetModelId(ctx, transactOpt.From, modelID)
	if err != nil {
		return nil, lib.WrapError(ErrModel, err)
	}

	model, err := s.modelRegistry.GetModelById(ctx, ID)
	if err != nil {
		return nil, lib.WrapError(ErrModel, err)
	}

	modelType := DetectModelType(model.Tags)

	return &structs.Model{
		Id:        ID,
		IpfsCID:   model.IpfsCID,
		Fee:       model.Fee,
		Stake:     model.Stake,
		Owner:     model.Owner,
		Name:      model.Name,
		Tags:      model.Tags,
		CreatedAt: model.CreatedAt,
		IsDeleted: model.IsDeleted,
		ModelType: modelType,
	}, nil
}

func (s *BlockchainService) DeregisterModel(ctx context.Context, modelId common.Hash) (common.Hash, error) {
	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrPrKey, err)
	}

	transactOpt, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrTxOpts, err)
	}

	tx, err := s.modelRegistry.DeregisterModel(transactOpt, modelId)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrSendTx, err)
	}

	return tx, nil
}

func (s *BlockchainService) ModelExists(ctx context.Context, modelID common.Hash) (bool, error) {
	m, err := s.modelRegistry.GetModelById(ctx, modelID)

	// cannot pull blockchain data
	if err != nil {
		return false, err
	}

	// model never existed
	if m.CreatedAt.Cmp(big.NewInt(0)) == 0 {
		return false, nil
	}

	// model was deleted
	if m.IsDeleted {
		return false, nil
	}

	return true, nil
}

func (s *BlockchainService) CreateNewBid(ctx context.Context, modelID common.Hash, pricePerSecond *lib.BigInt) (*structs.Bid, error) {
	fee, err := s.marketplace.GetBidFee(ctx)
	if err != nil {
		return nil, err
	}

	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return nil, lib.WrapError(ErrPrKey, err)
	}

	transactOpt, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return nil, lib.WrapError(ErrTxOpts, err)
	}

	// Increase allowance for the fee (safe for concurrent use)
	_, err = s.increaseAllowance(ctx, s.diamonContractAddr, fee)
	if err != nil {
		return nil, lib.WrapError(ErrApprove, err)
	}

	newBidId, err := s.marketplace.PostModelBid(transactOpt, modelID, &pricePerSecond.Int)
	if err != nil {
		return nil, lib.WrapError(ErrSendTx, err)
	}

	bid, err := s.GetBidByID(ctx, newBidId)
	if err != nil {
		return nil, lib.WrapError(ErrBid, err)
	}

	return bid, nil
}

func (s *BlockchainService) DeleteBid(ctx context.Context, bidId common.Hash) (common.Hash, error) {
	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrPrKey, err)
	}

	transactOpt, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrTxOpts, err)
	}

	tx, err := s.marketplace.DeleteBid(transactOpt, bidId)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrSendTx, err)
	}

	return tx, nil
}

func (s *BlockchainService) DeregisterProdiver(ctx context.Context) (common.Hash, error) {
	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrPrKey, err)
	}

	transactOpt, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrTxOpts, err)
	}

	tx, err := s.providerRegistry.DeregisterProvider(transactOpt)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrSendTx, err)
	}

	return tx, nil
}

func (s *BlockchainService) CloseSession(ctx context.Context, sessionID common.Hash) (common.Hash, error) {
	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrPrKey, err)
	}

	addr, err := lib.PrivKeyBytesToAddr(prKey)
	if err != nil {
		return common.Hash{}, err
	}

	var reportMessage []byte
	var signedReport []byte

	report, err := s.proxyService.GetSessionReportFromProvider(ctx, sessionID)
	if err != nil {
		s.log.Warnf("failed to get provider's report: %s", err)

		s.log.Info("using user report")
		reportMessage, signedReport, err = s.proxyService.GetSessionReportFromUser(ctx, sessionID)
		if err != nil {
			return common.Hash{}, lib.WrapError(ErrSessionUserReport, err)
		}
	} else {
		reportMessage = report.Message
		signedReport = report.SignedReport
	}

	baseOpts, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrTxOpts, err)
	}

	receipt, err := s.txEscalator.SendWithEscalation(
		ctx,
		baseOpts,
		func(opts *bind.TransactOpts) (*types.Transaction, error) {
			return s.sessionRouter.CloseSessionTx(opts, reportMessage, signedReport)
		},
		s.legacyTx,
	)
	if err != nil {
		s.handleTxError(ctx, addr, err)
		return common.Hash{}, lib.WrapError(ErrSendTx, fmt.Errorf("close session failed: %w", err))
	}

	if receipt.Status != 1 {
		return receipt.TxHash, lib.WrapError(ErrSendTx, fmt.Errorf("close session tx failed with status %d", receipt.Status))
	}

	return receipt.TxHash, nil
}

func (s *BlockchainService) GetSession(ctx context.Context, sessionID common.Hash) (*structs.Session, error) {
	ses, err := s.sessionRouter.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	bid, err := s.marketplace.GetBidById(ctx, ses.BidId)
	if err != nil {
		return nil, err
	}
	return mapSession(sessionID, *ses, *bid), nil
}

func (s *BlockchainService) GetProviderClaimableBalance(ctx context.Context, sessionID common.Hash) (*big.Int, error) {
	return s.sessionRouter.GetProviderClaimableBalance(ctx, sessionID)
}

func (s *BlockchainService) GetBalance(ctx context.Context) (eth *big.Int, mor *big.Int, err error) {
	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return nil, nil, lib.WrapError(ErrPrKey, err)
	}

	transactOpt, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return nil, nil, lib.WrapError(ErrTxOpts, err)
	}

	ethBalance, err := s.ethClient.BalanceAt(ctx, transactOpt.From, nil)
	if err != nil {
		return nil, nil, err
	}

	morBalance, err := s.morToken.GetBalance(ctx, transactOpt.From)
	if err != nil {
		return nil, nil, err
	}

	return ethBalance, morBalance, nil
}

func (s *BlockchainService) SendETH(ctx context.Context, to common.Address, amount *big.Int, agentUsername string) (common.Hash, error) {
	shouldDecrease, err := s.authConfig.IsAllowanceEnough(agentUsername, "eth", amount)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrAgentUserAllowance, err)
	}

	signedTx, err := s.createSignedTransaction(ctx, &types.DynamicFeeTx{
		To:    &to,
		Value: amount,
	})
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrSendTx, err)
	}

	err = s.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrSendTx, err)
	}

	// Wait for tx to be mined with timeout
	receipt, err := lib.WaitMinedWithTimeout(ctx, s.ethClient, signedTx, lib.DefaultTxMineTimeout)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrWaitMined, err)
	}

	if shouldDecrease {
		amountBigInt := lib.BigInt{Int: *amount}
		err = s.authConfig.DecreaseAllowance(agentUsername, "eth", amountBigInt)
		if err != nil {
			s.log.Errorf("failed to decrease allowance: %s", err)
			return common.Hash{}, err
		}
		s.authConfig.AuthStorage.SetAgentTx(signedTx.Hash().Hex(), agentUsername, receipt.BlockNumber)
	}

	return signedTx.Hash(), nil
}

func (s *BlockchainService) createSignedTransaction(ctx context.Context, txdata *types.DynamicFeeTx) (*types.Transaction, error) {
	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return nil, lib.WrapError(ErrPrKey, err)
	}
	addr, err := lib.PrivKeyBytesToAddr(prKey)
	if err != nil {
		return nil, err
	}

	gasTipCap, err := s.ethClient.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, err
	}

	head, err := s.ethClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Get the pending nonce from the chain
	nonce, err := s.ethClient.PendingNonceAt(ctx, addr)
	if err != nil {
		return nil, lib.WrapError(ErrNonce, err)
	}

	gasFeeCap := new(big.Int).Add(
		gasTipCap,
		new(big.Int).Mul(head.BaseFee, big.NewInt(basefeeWiggleMultiplier)),
	)

	gas, err := s.ethClient.EstimateGas(ctx, ethereum.CallMsg{
		From:  addr,
		To:    txdata.To,
		Value: txdata.Value,
	})

	chainID, err := s.ethClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gas,
		To:        txdata.To,
		Value:     txdata.Value,
	})

	privateKey, err := crypto.ToECDSA(prKey)
	if err != nil {
		return nil, err
	}

	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(tx.ChainId()), privateKey)
	if err != nil {
		return nil, lib.WrapError(ErrSignTx, err)
	}

	return signedTx, nil
}

func (s *BlockchainService) SendMOR(ctx context.Context, to common.Address, amount *big.Int, agentUsername string) (common.Hash, error) {
	shouldDecrease, err := s.authConfig.IsAllowanceEnough(agentUsername, s.morTokenAddr.Hex(), amount)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrAgentUserAllowance, err)
	}

	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrPrKey, err)
	}

	transactOpt, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrTxOpts, err)
	}

	tx, receipt, err := s.morToken.Transfer(transactOpt, to, amount)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrSendTx, err)
	}

	if shouldDecrease {
		amountBigInt := lib.BigInt{Int: *amount}
		err = s.authConfig.DecreaseAllowance(agentUsername, s.morTokenAddr.Hex(), amountBigInt)
		if err != nil {
			s.log.Errorf("failed to decrease allowance: %s", err)
			return common.Hash{}, err
		}
		err = s.authConfig.AuthStorage.SetAgentTx(tx.Hash().Hex(), agentUsername, receipt.BlockNumber)
		if err != nil {
			s.log.Errorf("failed to set agent tx: %s", err)
		}
	}

	return tx.Hash(), nil
}

func (s *BlockchainService) GetAllowance(ctx context.Context, spender common.Address) (*big.Int, error) {
	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return nil, lib.WrapError(ErrPrKey, err)
	}

	transactOpt, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return nil, lib.WrapError(ErrTxOpts, err)
	}

	allowance, err := s.morToken.GetAllowance(ctx, transactOpt.From, spender)
	if err != nil {
		return nil, err
	}

	return allowance, nil
}

// Approve sets the allowance for a spender to the given amount.
// For public API use only. Internal operations should use increaseAllowance.
func (s *BlockchainService) Approve(ctx context.Context, spender common.Address, amount *big.Int) (common.Hash, error) {
	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrPrKey, err)
	}

	transactOpt, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrTxOpts, err)
	}

	tx, _, err := s.morToken.Approve(transactOpt, spender, amount)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrSendTx, err)
	}

	return tx.Hash(), nil
}

// increaseAllowance adds to the allowance for a spender.
// Safe for concurrent use - each call adds to existing allowance.
func (s *BlockchainService) increaseAllowance(ctx context.Context, spender common.Address, amount *big.Int) (common.Hash, error) {
	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrPrKey, err)
	}

	transactOpt, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrTxOpts, err)
	}

	tx, _, err := s.morToken.IncreaseAllowance(transactOpt, spender, amount)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrSendTx, err)
	}

	return tx.Hash(), nil
}

func (s *BlockchainService) ClaimProviderBalance(ctx context.Context, sessionID [32]byte) (common.Hash, error) {
	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrPrKey, err)
	}

	transactOpt, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrTxOpts, err)
	}

	txHash, err := s.sessionRouter.ClaimProviderBalance(transactOpt, sessionID)
	if err != nil {
		return common.Hash{}, err
	}

	return txHash, nil
}

func (s *BlockchainService) GetTokenSupply(ctx context.Context) (*big.Int, error) {
	return s.sessionRouter.GetTotalMORSupply(ctx, big.NewInt(time.Now().Unix()))
}

func (s *BlockchainService) GetTodaysBudget(ctx context.Context) (*big.Int, error) {
	return s.sessionRouter.GetTodaysBudget(ctx, big.NewInt(time.Now().Unix()))
}

func (s *BlockchainService) GetSessions(ctx context.Context, user, provider common.Address, offset *big.Int, limit uint8, order r.Order) ([]*structs.Session, error) {
	var (
		ids      [][32]byte
		sessions []sr.ISessionStorageSession
		err      error
	)
	if (user != common.Address{}) {
		ids, sessions, err = s.sessionRouter.GetSessionsByUser(ctx, user, offset, limit, order)
	} else {
		// hasProvider
		ids, sessions, err = s.sessionRouter.GetSessionsByProvider(ctx, provider, offset, limit, order)
	}
	if err != nil {
		return nil, err
	}

	bidIDs := make([][32]byte, len(sessions))
	for i := 0; i < len(sessions); i++ {
		bidIDs[i] = sessions[i].BidId
	}

	_, bids, err := s.marketplace.GetMultipleBids(ctx, bidIDs)
	if err != nil {
		return nil, err
	}

	return mapSessions(ids, sessions, bids), nil
}

func (s *BlockchainService) GetSessionsIds(ctx context.Context, user, provider common.Address, offset *big.Int, limit uint8, order r.Order) ([]common.Hash, error) {
	ids, err := s.sessionRouter.GetSessionsIdsByUser(ctx, user, offset, limit, order)

	if err != nil {
		return nil, err
	}

	bidIDs := make([]common.Hash, len(ids))
	for i := 0; i < len(ids); i++ {
		bidIDs[i] = ids[i]
	}

	return bidIDs, nil
}

func (s *BlockchainService) GetTransactions(ctx context.Context, page uint64, limit uint8) ([]structs.MappedTransaction, error) {
	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return nil, lib.WrapError(ErrPrKey, err)
	}

	transactOpt, err := s.getTransactOpts(ctx, prKey)
	if err != nil {
		return nil, lib.WrapError(ErrTxOpts, err)
	}
	address := transactOpt.From

	allTrxs, err := s.explorerClient.GetLastTransactions(ctx, address)
	if err != nil {
		s.log.Errorf("failed to get  transactions: %s", err.Error())
		return nil, err
	}

	return allTrxs, nil
}

func (s *BlockchainService) openSessionByBid(ctx context.Context, bidID common.Hash, duration *big.Int, agentUsername string) (common.Hash, error) {
	supply, err := s.GetTokenSupply(ctx)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrTokenSupply, err)
	}

	budget, err := s.GetTodaysBudget(ctx)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrBudget, err)
	}

	bid, err := s.GetBidByID(ctx, bidID)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrBid, err)
	}

	userAddr, err := s.GetMyAddress(ctx)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrMyAddress, err)
	}

	if bid.Provider == userAddr {
		return common.Hash{}, ErrOpenOwnBid
	}

	hash, _, err := s.tryOpenSession(ctx, bid, duration, supply, budget, userAddr, false, false, agentUsername)
	return hash, err
}

func (s *BlockchainService) OpenSessionByModelId(ctx context.Context, modelID common.Hash, duration *big.Int, directPayment bool, isFailoverEnabled bool, omitProvider common.Address, agentUsername string) (common.Hash, error) {
	supply, err := s.GetTokenSupply(ctx)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrTokenSupply, err)
	}

	budget, err := s.GetTodaysBudget(ctx)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrBudget, err)
	}

	modelStats, err := s.sessionRouter.GetModelStats(ctx, modelID)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrModel, err)
	}

	bidIDs, bids, providerStats, providers, err := s.GetAllBidsWithRating(ctx, modelID)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrBid, err)
	}

	if len(bids) == 0 {
		return common.Hash{}, ErrNoBid
	}

	userAddr, err := s.GetMyAddress(ctx)
	if err != nil {
		return common.Hash{}, lib.WrapError(ErrMyAddress, err)
	}

	minStake, err := s.getMinStakeCached(ctx)
	if err != nil {
		return common.Hash{}, fmt.Errorf("failed to get min stake: %w", err)
	}

	scoredBids := s.rateBids(bidIDs, bids, providerStats, providers, modelStats, minStake, s.log)
	for i, bid := range scoredBids {
		providerAddr := bid.Bid.Provider
		if providerAddr == omitProvider {
			s.log.Infof("skipping provider #%d %s", i, providerAddr.String())
			continue
		}

		if providerAddr == userAddr {
			s.log.Infof("skipping own bid #%d %s", i, bid.Bid.Id)
			continue
		}

		s.log.Infof("trying to open session with provider #%d %s", i, bid.Bid.Provider.String())
		durationCopy := new(big.Int).Set(duration)

		hash, tryNext, err := s.tryOpenSession(ctx, &bid.Bid, durationCopy, supply, budget, userAddr, directPayment, isFailoverEnabled, agentUsername)
		if err != nil {
			s.log.Errorf("failed to open session with provider %s: %s", bid.Bid.Provider.String(), err.Error())
			if tryNext {
				continue
			} else {
				return common.Hash{}, err
			}
		}

		return hash, nil
	}

	return common.Hash{}, fmt.Errorf("no provider accepting session")
}

func (s *BlockchainService) GetAllBidsWithRating(ctx context.Context, modelAgentID [32]byte) ([][32]byte, []m.IBidStorageBid, []sr.IStatsStorageProviderModelStats, []pr.IProviderStorageProvider, error) {
	batchSize := uint8(255)
	offset := big.NewInt(0)
	bids := make([]m.IBidStorageBid, 0)
	ids := make([][32]byte, 0)
	providerModelStats := make([]sr.IStatsStorageProviderModelStats, 0)
	providers := make([]pr.IProviderStorageProvider, 0)

	for {
		if ctx.Err() != nil {
			return nil, nil, nil, nil, ctx.Err()
		}

		idsBatch, bidsBatch, err := s.marketplace.GetActiveBidsByModel(ctx, modelAgentID, offset, batchSize, r.OrderASC)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		ids = append(ids, idsBatch...)
		bids = append(bids, bidsBatch...)

		for _, bid := range bidsBatch {
			//TODO: replace with multicall
			providerModelStat, err := s.sessionRouter.GetProviderModelStats(ctx, modelAgentID, bid.Provider)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			provider, err := s.providerRegistry.GetProviderById(ctx, bid.Provider)
			if err != nil {
				return nil, nil, nil, nil, err
			}
			providerModelStats = append(providerModelStats, *providerModelStat)
			providers = append(providers, *provider)
		}

		if len(bidsBatch) < int(batchSize) {
			break
		}

		offset.Add(offset, big.NewInt(int64(batchSize)))
	}

	return ids, bids, providerModelStats, providers, nil
}

func (s *BlockchainService) tryOpenSession(ctx context.Context, bid *structs.Bid, duration, supply, budget *big.Int, userAddr common.Address, directPayment bool, failoverEnabled bool, agentUsername string) (common.Hash, bool, error) {
	provider, err := s.providerRegistry.GetProviderById(ctx, bid.Provider)
	if err != nil {
		return common.Hash{}, false, lib.WrapError(ErrProvider, err)
	}
	sessionCost := (&big.Int{}).Mul(&bid.PricePerSecond.Int, duration)

	var amountTransferred = new(big.Int)
	if directPayment {
		// amount transferred is the session cost
		amountTransferred = sessionCost
	} else {
		// amount transferred is the stake
		stake := (&big.Int{}).Div((&big.Int{}).Mul(supply, sessionCost), budget)
		amountTransferred = stake
	}

	s.log.Infof("attempting to initiate session %s", map[string]string{
		"provider":          bid.Provider.String(),
		"directPayment":     strconv.FormatBool(directPayment),
		"duration":          duration.String(),
		"bid":               bid.Id.String(),
		"endpoint":          provider.Endpoint,
		"amountTransferred": amountTransferred.String(),
	})

	initRes, err := s.proxyService.InitiateSession(ctx, userAddr, bid.Provider, amountTransferred, bid.Id, provider.Endpoint)
	if err != nil {
		return common.Hash{}, true, lib.WrapError(ErrInitSession, err)
	}

	hash, err := s.OpenSession(ctx, initRes.Approval, initRes.ApprovalSig, amountTransferred, directPayment, agentUsername)
	if err != nil {
		return common.Hash{}, false, err
	}

	return hash, false, nil
}

func (s *BlockchainService) GetMyAddress(ctx context.Context) (common.Address, error) {
	prKey, err := s.privateKey.GetPrivateKey()
	if err != nil {
		return common.Address{}, lib.WrapError(ErrPrKey, err)
	}

	return lib.PrivKeyBytesToAddr(prKey)
}

func (s *BlockchainService) CheckConnectivity(ctx context.Context, url string, addr common.Address) (time.Duration, error) {
	return s.proxyService.Ping(ctx, url, addr)
}

func (s *BlockchainService) CheckPortOpen(ctx context.Context, urlStr string) (bool, error) {
	host, port, err := net.SplitHostPort(urlStr)
	if err != nil {
		return false, err
	}
	portInt, err := strconv.ParseInt(port, 10, 0)
	if err != nil {
		return false, err
	}

	body, _ := json.Marshal(struct {
		Host  string  `json:"host"`
		Ports []int64 `json:"ports"`
	}{
		Host:  host,
		Ports: []int64{portInt},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", "https://portchecker.io/api/query", bytes.NewBuffer(body))
	if err != nil {
		return false, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return false, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	var response struct {
		Error bool
		Msg   string
		Check []struct {
			Port   int
			Status bool
		}
		Host string
	}

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return false, err
	}

	if response.Error {
		return false, fmt.Errorf("portchecker.io error: %s", response.Msg)
	}

	if len(response.Check) != 1 {
		return false, fmt.Errorf("unexpected response from portchecker.io")
	}

	return response.Check[0].Status, nil
}

func (s *BlockchainService) getTransactOpts(ctx context.Context, privKey lib.HexString) (*bind.TransactOpts, error) {
	privateKey, err := crypto.ToECDSA(privKey)
	if err != nil {
		return nil, err
	}

	chainId, err := s.ethClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	transactOpts, err := bind.NewKeyedTransactorWithChainID(privateKey, chainId)
	if err != nil {
		return nil, err
	}

	// NOTE: We intentionally don't set Nonce here. When Nonce is nil,
	// go-ethereum's bind library will automatically fetch the pending nonce
	// at transaction submission time, which handles most cases correctly.

	if s.legacyTx {
		gasPrice, err := s.ethClient.SuggestGasPrice(ctx)
		if err != nil {
			return nil, err
		}
		transactOpts.GasPrice = gasPrice
	}

	transactOpts.Value = big.NewInt(0)
	transactOpts.Context = ctx

	return transactOpts, nil
}

func (s *BlockchainService) signTx(ctx context.Context, tx *types.Transaction, privKey lib.HexString) (*types.Transaction, error) {
	privateKey, err := crypto.ToECDSA(privKey)
	if err != nil {
		return nil, err
	}

	chainId, err := s.ethClient.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	return types.SignTx(tx, types.NewEIP155Signer(chainId), privateKey)
}

func (s *BlockchainService) getMinStakeCached(ctx context.Context) (*big.Int, error) {
	if s.minStake != nil {
		return s.minStake, nil
	}

	minStake, err := s.providerRegistry.GetMinStake(ctx)
	if err != nil {
		return nil, err
	}
	s.minStake = minStake
	return minStake, nil
}

// handleTxError logs transaction errors for debugging.
func (s *BlockchainService) handleTxError(ctx context.Context, addr common.Address, err error) {
	if lib.IsNonceError(err) {
		s.log.Warnf("Nonce error for %s: %v", addr.Hex(), err)
	}
}
