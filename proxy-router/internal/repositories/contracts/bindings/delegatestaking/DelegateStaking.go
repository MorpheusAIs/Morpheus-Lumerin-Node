// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package delegatestaking

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// IBidStorageBid is an auto generated low-level Go binding around an user-defined struct.
type IBidStorageBid struct {
	Provider       common.Address
	ModelId        [32]byte
	PricePerSecond *big.Int
	Nonce          *big.Int
	CreatedAt      *big.Int
	DeletedAt      *big.Int
}

// IProviderStorageProvider is an auto generated low-level Go binding around an user-defined struct.
type IProviderStorageProvider struct {
	Endpoint          string
	Stake             *big.Int
	CreatedAt         *big.Int
	LimitPeriodEnd    *big.Int
	LimitPeriodEarned *big.Int
	IsDeleted         bool
}

// ISessionStoragePool is an auto generated low-level Go binding around an user-defined struct.
type ISessionStoragePool struct {
	InitialReward    *big.Int
	RewardDecrease   *big.Int
	PayoutStart      *big.Int
	DecreaseInterval *big.Int
}

// ISessionStorageSession is an auto generated low-level Go binding around an user-defined struct.
type ISessionStorageSession struct {
	User                    common.Address
	BidId                   [32]byte
	Stake                   *big.Int
	CloseoutReceipt         []byte
	CloseoutType            *big.Int
	ProviderWithdrawnAmount *big.Int
	OpenedAt                *big.Int
	EndsAt                  *big.Int
	ClosedAt                *big.Int
	IsActive                bool
	IsDirectPaymentFromUser bool
}

// IStatsStorageModelStats is an auto generated low-level Go binding around an user-defined struct.
type IStatsStorageModelStats struct {
	TpsScaled1000 LibSDSD
	TtftMs        LibSDSD
	TotalDuration LibSDSD
	Count         uint32
}

// IStatsStorageProviderModelStats is an auto generated low-level Go binding around an user-defined struct.
type IStatsStorageProviderModelStats struct {
	TpsScaled1000 LibSDSD
	TtftMs        LibSDSD
	TotalDuration uint32
	SuccessCount  uint32
	TotalCount    uint32
}

// LibSDSD is an auto generated low-level Go binding around an user-defined struct.
type LibSDSD struct {
	Mean  int64
	SqSum int64
}

// DelegateStakingMetaData contains all meta data concerning the DelegateStaking contract.
var DelegateStakingMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"DelegateStakingActiveFundersLimitExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingAllowanceNotActive\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingCapBelowLifetimeFunding\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingFundingCapExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInsufficientCapacity\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInsufficientLiquidity\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInsufficientPrincipal\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidExpiry\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidPoolSettlement\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidWithdrawalCursor\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingLimitExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPoolFunderNotApproved\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPoolSessionAlreadyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPoolSessionNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPrincipalTooLow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingSessionSlotNotEmpty\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingSettlementNotReady\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingWithdrawalNotFound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"delegatee\",\"type\":\"address\"}],\"name\":\"InsufficientRightsForOperation\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionAlreadyClosed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionApproveExpired\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionApprovedForAnotherChainId\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionApprovedForAnotherUser\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionBidNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionDuplicateApproval\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionProviderSignatureMismatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionReceiptExpired\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionReceiptForAnotherChainId\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionStakeTooLow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionTooShort\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingAllowanceFunded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"cumulativeFundingCap\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"expiry\",\"type\":\"uint128\"}],\"name\":\"DelegateStakingAllowanceGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"}],\"name\":\"DelegateStakingAllowanceRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMaxActiveFundersUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMaxAutoReleaseDaysUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMaxAutoServiceUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMinPrincipalUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolDrawn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"}],\"name\":\"DelegateStakingPoolFunderApprovalRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"}],\"name\":\"DelegateStakingPoolFunderApproved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint128\",\"name\":\"releaseAt\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolLockCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint128\",\"name\":\"releaseAt\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolLockReleased\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolReturned\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"DelegateStakingPoolSessionSettled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingSettlementGraceUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingWithdrawalCancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingWithdrawalPaid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"paid\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"pending\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingWithdrawalRequested\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"payoutTarget\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ProviderDebtPaid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"payoutTarget\",\"type\":\"address\"}],\"name\":\"ProviderPayoutTargetUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"accrued\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"paid\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"owed\",\"type\":\"uint256\"}],\"name\":\"ProviderRewardAccrued\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerId\",\"type\":\"address\"}],\"name\":\"SessionClosed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerId\",\"type\":\"address\"}],\"name\":\"SessionClosed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerId\",\"type\":\"address\"}],\"name\":\"SessionOpened\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"tpsScaled1000\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"ttftMs\",\"type\":\"uint32\"}],\"name\":\"SessionRating\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BIDS_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"COMPUTE_POOL_INDEX\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_MARKETPLACE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_MODEL\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_PROVIDER\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_SESSION\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_SESSION_DURATION\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PROVIDERS_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SESSIONS_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SIGNATURE_TTL\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"STATS_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"receiptEncoded_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature_\",\"type\":\"bytes\"}],\"name\":\"closePoolSession\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getActiveProviders\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId_\",\"type\":\"bytes32\"}],\"name\":\"getBid\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"deletedAt\",\"type\":\"uint128\"}],\"internalType\":\"structIBidStorage.Bid\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getFundingAccount\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getIsProviderActive\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"approval_\",\"type\":\"bytes\"}],\"name\":\"getIsProviderApprovalUsed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMaxSessionDuration\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getModelActiveBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getModelBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getModelSessions\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"}],\"name\":\"getModelStats\",\"outputs\":[{\"components\":[{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"tpsScaled1000\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"ttftMs\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"totalDuration\",\"type\":\"tuple\"},{\"internalType\":\"uint32\",\"name\":\"count\",\"type\":\"uint32\"}],\"internalType\":\"structIStatsStorage.ModelStats\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index_\",\"type\":\"uint256\"}],\"name\":\"getPool\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"initialReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardDecrease\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"payoutStart\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"decreaseInterval\",\"type\":\"uint128\"}],\"internalType\":\"structISessionStorage.Pool\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPools\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"initialReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardDecrease\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"payoutStart\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"decreaseInterval\",\"type\":\"uint128\"}],\"internalType\":\"structISessionStorage.Pool[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getProvider\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"endpoint\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"limitPeriodEnd\",\"type\":\"uint128\"},{\"internalType\":\"uint256\",\"name\":\"limitPeriodEarned\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"internalType\":\"structIProviderStorage.Provider\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getProviderActiveBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getProviderBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getProviderMinimumStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getProviderModelStats\",\"outputs\":[{\"components\":[{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"tpsScaled1000\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"ttftMs\",\"type\":\"tuple\"},{\"internalType\":\"uint32\",\"name\":\"totalDuration\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"successCount\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"totalCount\",\"type\":\"uint32\"}],\"internalType\":\"structIStatsStorage.ProviderModelStats\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getProviderSessions\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getProvidersTotalClaimed\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRegistry\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId_\",\"type\":\"bytes32\"}],\"name\":\"getSession\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"closeoutReceipt\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerWithdrawnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"openedAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"endsAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"closedAt\",\"type\":\"uint128\"},{\"internalType\":\"bool\",\"name\":\"isActive\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"isDirectPaymentFromUser\",\"type\":\"bool\"}],\"internalType\":\"structISessionStorage.Session\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getTotalSessions\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getUserSessions\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId_\",\"type\":\"bytes32\"}],\"name\":\"isBidActive\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegatee_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"delegator_\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"rights_\",\"type\":\"bytes32\"}],\"name\":\"isRightsDelegated\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"approvalEncoded_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature_\",\"type\":\"bytes\"}],\"name\":\"openSessionFromPool\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId_\",\"type\":\"bytes32\"}],\"name\":\"settleExpiredSession\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// DelegateStakingABI is the input ABI used to generate the binding from.
// Deprecated: Use DelegateStakingMetaData.ABI instead.
var DelegateStakingABI = DelegateStakingMetaData.ABI

// DelegateStaking is an auto generated Go binding around an Ethereum contract.
type DelegateStaking struct {
	DelegateStakingCaller     // Read-only binding to the contract
	DelegateStakingTransactor // Write-only binding to the contract
	DelegateStakingFilterer   // Log filterer for contract events
}

// DelegateStakingCaller is an auto generated read-only Go binding around an Ethereum contract.
type DelegateStakingCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DelegateStakingTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DelegateStakingTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DelegateStakingFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DelegateStakingFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DelegateStakingSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DelegateStakingSession struct {
	Contract     *DelegateStaking  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DelegateStakingCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DelegateStakingCallerSession struct {
	Contract *DelegateStakingCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// DelegateStakingTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DelegateStakingTransactorSession struct {
	Contract     *DelegateStakingTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// DelegateStakingRaw is an auto generated low-level Go binding around an Ethereum contract.
type DelegateStakingRaw struct {
	Contract *DelegateStaking // Generic contract binding to access the raw methods on
}

// DelegateStakingCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DelegateStakingCallerRaw struct {
	Contract *DelegateStakingCaller // Generic read-only contract binding to access the raw methods on
}

// DelegateStakingTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DelegateStakingTransactorRaw struct {
	Contract *DelegateStakingTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDelegateStaking creates a new instance of DelegateStaking, bound to a specific deployed contract.
func NewDelegateStaking(address common.Address, backend bind.ContractBackend) (*DelegateStaking, error) {
	contract, err := bindDelegateStaking(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &DelegateStaking{DelegateStakingCaller: DelegateStakingCaller{contract: contract}, DelegateStakingTransactor: DelegateStakingTransactor{contract: contract}, DelegateStakingFilterer: DelegateStakingFilterer{contract: contract}}, nil
}

// NewDelegateStakingCaller creates a new read-only instance of DelegateStaking, bound to a specific deployed contract.
func NewDelegateStakingCaller(address common.Address, caller bind.ContractCaller) (*DelegateStakingCaller, error) {
	contract, err := bindDelegateStaking(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingCaller{contract: contract}, nil
}

// NewDelegateStakingTransactor creates a new write-only instance of DelegateStaking, bound to a specific deployed contract.
func NewDelegateStakingTransactor(address common.Address, transactor bind.ContractTransactor) (*DelegateStakingTransactor, error) {
	contract, err := bindDelegateStaking(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingTransactor{contract: contract}, nil
}

// NewDelegateStakingFilterer creates a new log filterer instance of DelegateStaking, bound to a specific deployed contract.
func NewDelegateStakingFilterer(address common.Address, filterer bind.ContractFilterer) (*DelegateStakingFilterer, error) {
	contract, err := bindDelegateStaking(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFilterer{contract: contract}, nil
}

// bindDelegateStaking binds a generic wrapper to an already deployed contract.
func bindDelegateStaking(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := DelegateStakingMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DelegateStaking *DelegateStakingRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DelegateStaking.Contract.DelegateStakingCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DelegateStaking *DelegateStakingRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DelegateStaking.Contract.DelegateStakingTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DelegateStaking *DelegateStakingRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DelegateStaking.Contract.DelegateStakingTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DelegateStaking *DelegateStakingCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DelegateStaking.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DelegateStaking *DelegateStakingTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DelegateStaking.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DelegateStaking *DelegateStakingTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DelegateStaking.Contract.contract.Transact(opts, method, params...)
}

// BIDSSTORAGESLOT is a free data retrieval call binding the contract method 0x266ccff0.
//
// Solidity: function BIDS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCaller) BIDSSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "BIDS_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BIDSSTORAGESLOT is a free data retrieval call binding the contract method 0x266ccff0.
//
// Solidity: function BIDS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingSession) BIDSSTORAGESLOT() ([32]byte, error) {
	return _DelegateStaking.Contract.BIDSSTORAGESLOT(&_DelegateStaking.CallOpts)
}

// BIDSSTORAGESLOT is a free data retrieval call binding the contract method 0x266ccff0.
//
// Solidity: function BIDS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCallerSession) BIDSSTORAGESLOT() ([32]byte, error) {
	return _DelegateStaking.Contract.BIDSSTORAGESLOT(&_DelegateStaking.CallOpts)
}

// COMPUTEPOOLINDEX is a free data retrieval call binding the contract method 0xc56d09a0.
//
// Solidity: function COMPUTE_POOL_INDEX() view returns(uint256)
func (_DelegateStaking *DelegateStakingCaller) COMPUTEPOOLINDEX(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "COMPUTE_POOL_INDEX")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// COMPUTEPOOLINDEX is a free data retrieval call binding the contract method 0xc56d09a0.
//
// Solidity: function COMPUTE_POOL_INDEX() view returns(uint256)
func (_DelegateStaking *DelegateStakingSession) COMPUTEPOOLINDEX() (*big.Int, error) {
	return _DelegateStaking.Contract.COMPUTEPOOLINDEX(&_DelegateStaking.CallOpts)
}

// COMPUTEPOOLINDEX is a free data retrieval call binding the contract method 0xc56d09a0.
//
// Solidity: function COMPUTE_POOL_INDEX() view returns(uint256)
func (_DelegateStaking *DelegateStakingCallerSession) COMPUTEPOOLINDEX() (*big.Int, error) {
	return _DelegateStaking.Contract.COMPUTEPOOLINDEX(&_DelegateStaking.CallOpts)
}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCaller) DELEGATIONRULESMARKETPLACE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "DELEGATION_RULES_MARKETPLACE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_DelegateStaking *DelegateStakingSession) DELEGATIONRULESMARKETPLACE() ([32]byte, error) {
	return _DelegateStaking.Contract.DELEGATIONRULESMARKETPLACE(&_DelegateStaking.CallOpts)
}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCallerSession) DELEGATIONRULESMARKETPLACE() ([32]byte, error) {
	return _DelegateStaking.Contract.DELEGATIONRULESMARKETPLACE(&_DelegateStaking.CallOpts)
}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCaller) DELEGATIONRULESMODEL(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "DELEGATION_RULES_MODEL")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_DelegateStaking *DelegateStakingSession) DELEGATIONRULESMODEL() ([32]byte, error) {
	return _DelegateStaking.Contract.DELEGATIONRULESMODEL(&_DelegateStaking.CallOpts)
}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCallerSession) DELEGATIONRULESMODEL() ([32]byte, error) {
	return _DelegateStaking.Contract.DELEGATIONRULESMODEL(&_DelegateStaking.CallOpts)
}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCaller) DELEGATIONRULESPROVIDER(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "DELEGATION_RULES_PROVIDER")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_DelegateStaking *DelegateStakingSession) DELEGATIONRULESPROVIDER() ([32]byte, error) {
	return _DelegateStaking.Contract.DELEGATIONRULESPROVIDER(&_DelegateStaking.CallOpts)
}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCallerSession) DELEGATIONRULESPROVIDER() ([32]byte, error) {
	return _DelegateStaking.Contract.DELEGATIONRULESPROVIDER(&_DelegateStaking.CallOpts)
}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCaller) DELEGATIONRULESSESSION(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "DELEGATION_RULES_SESSION")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_DelegateStaking *DelegateStakingSession) DELEGATIONRULESSESSION() ([32]byte, error) {
	return _DelegateStaking.Contract.DELEGATIONRULESSESSION(&_DelegateStaking.CallOpts)
}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCallerSession) DELEGATIONRULESSESSION() ([32]byte, error) {
	return _DelegateStaking.Contract.DELEGATIONRULESSESSION(&_DelegateStaking.CallOpts)
}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCaller) DELEGATIONSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "DELEGATION_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingSession) DELEGATIONSTORAGESLOT() ([32]byte, error) {
	return _DelegateStaking.Contract.DELEGATIONSTORAGESLOT(&_DelegateStaking.CallOpts)
}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCallerSession) DELEGATIONSTORAGESLOT() ([32]byte, error) {
	return _DelegateStaking.Contract.DELEGATIONSTORAGESLOT(&_DelegateStaking.CallOpts)
}

// MINSESSIONDURATION is a free data retrieval call binding the contract method 0x7d980286.
//
// Solidity: function MIN_SESSION_DURATION() view returns(uint32)
func (_DelegateStaking *DelegateStakingCaller) MINSESSIONDURATION(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "MIN_SESSION_DURATION")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// MINSESSIONDURATION is a free data retrieval call binding the contract method 0x7d980286.
//
// Solidity: function MIN_SESSION_DURATION() view returns(uint32)
func (_DelegateStaking *DelegateStakingSession) MINSESSIONDURATION() (uint32, error) {
	return _DelegateStaking.Contract.MINSESSIONDURATION(&_DelegateStaking.CallOpts)
}

// MINSESSIONDURATION is a free data retrieval call binding the contract method 0x7d980286.
//
// Solidity: function MIN_SESSION_DURATION() view returns(uint32)
func (_DelegateStaking *DelegateStakingCallerSession) MINSESSIONDURATION() (uint32, error) {
	return _DelegateStaking.Contract.MINSESSIONDURATION(&_DelegateStaking.CallOpts)
}

// PROVIDERSSTORAGESLOT is a free data retrieval call binding the contract method 0xc51830f6.
//
// Solidity: function PROVIDERS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCaller) PROVIDERSSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "PROVIDERS_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PROVIDERSSTORAGESLOT is a free data retrieval call binding the contract method 0xc51830f6.
//
// Solidity: function PROVIDERS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingSession) PROVIDERSSTORAGESLOT() ([32]byte, error) {
	return _DelegateStaking.Contract.PROVIDERSSTORAGESLOT(&_DelegateStaking.CallOpts)
}

// PROVIDERSSTORAGESLOT is a free data retrieval call binding the contract method 0xc51830f6.
//
// Solidity: function PROVIDERS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCallerSession) PROVIDERSSTORAGESLOT() ([32]byte, error) {
	return _DelegateStaking.Contract.PROVIDERSSTORAGESLOT(&_DelegateStaking.CallOpts)
}

// SESSIONSSTORAGESLOT is a free data retrieval call binding the contract method 0xb392636e.
//
// Solidity: function SESSIONS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCaller) SESSIONSSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "SESSIONS_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SESSIONSSTORAGESLOT is a free data retrieval call binding the contract method 0xb392636e.
//
// Solidity: function SESSIONS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingSession) SESSIONSSTORAGESLOT() ([32]byte, error) {
	return _DelegateStaking.Contract.SESSIONSSTORAGESLOT(&_DelegateStaking.CallOpts)
}

// SESSIONSSTORAGESLOT is a free data retrieval call binding the contract method 0xb392636e.
//
// Solidity: function SESSIONS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCallerSession) SESSIONSSTORAGESLOT() ([32]byte, error) {
	return _DelegateStaking.Contract.SESSIONSSTORAGESLOT(&_DelegateStaking.CallOpts)
}

// SIGNATURETTL is a free data retrieval call binding the contract method 0xe7d791d0.
//
// Solidity: function SIGNATURE_TTL() view returns(uint32)
func (_DelegateStaking *DelegateStakingCaller) SIGNATURETTL(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "SIGNATURE_TTL")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// SIGNATURETTL is a free data retrieval call binding the contract method 0xe7d791d0.
//
// Solidity: function SIGNATURE_TTL() view returns(uint32)
func (_DelegateStaking *DelegateStakingSession) SIGNATURETTL() (uint32, error) {
	return _DelegateStaking.Contract.SIGNATURETTL(&_DelegateStaking.CallOpts)
}

// SIGNATURETTL is a free data retrieval call binding the contract method 0xe7d791d0.
//
// Solidity: function SIGNATURE_TTL() view returns(uint32)
func (_DelegateStaking *DelegateStakingCallerSession) SIGNATURETTL() (uint32, error) {
	return _DelegateStaking.Contract.SIGNATURETTL(&_DelegateStaking.CallOpts)
}

// STATSSTORAGESLOT is a free data retrieval call binding the contract method 0x87040789.
//
// Solidity: function STATS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCaller) STATSSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "STATS_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// STATSSTORAGESLOT is a free data retrieval call binding the contract method 0x87040789.
//
// Solidity: function STATS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingSession) STATSSTORAGESLOT() ([32]byte, error) {
	return _DelegateStaking.Contract.STATSSTORAGESLOT(&_DelegateStaking.CallOpts)
}

// STATSSTORAGESLOT is a free data retrieval call binding the contract method 0x87040789.
//
// Solidity: function STATS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStaking *DelegateStakingCallerSession) STATSSTORAGESLOT() ([32]byte, error) {
	return _DelegateStaking.Contract.STATSSTORAGESLOT(&_DelegateStaking.CallOpts)
}

// GetActiveProviders is a free data retrieval call binding the contract method 0xd5472642.
//
// Solidity: function getActiveProviders(uint256 offset_, uint256 limit_) view returns(address[], uint256)
func (_DelegateStaking *DelegateStakingCaller) GetActiveProviders(opts *bind.CallOpts, offset_ *big.Int, limit_ *big.Int) ([]common.Address, *big.Int, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getActiveProviders", offset_, limit_)

	if err != nil {
		return *new([]common.Address), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([]common.Address)).(*[]common.Address)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// GetActiveProviders is a free data retrieval call binding the contract method 0xd5472642.
//
// Solidity: function getActiveProviders(uint256 offset_, uint256 limit_) view returns(address[], uint256)
func (_DelegateStaking *DelegateStakingSession) GetActiveProviders(offset_ *big.Int, limit_ *big.Int) ([]common.Address, *big.Int, error) {
	return _DelegateStaking.Contract.GetActiveProviders(&_DelegateStaking.CallOpts, offset_, limit_)
}

// GetActiveProviders is a free data retrieval call binding the contract method 0xd5472642.
//
// Solidity: function getActiveProviders(uint256 offset_, uint256 limit_) view returns(address[], uint256)
func (_DelegateStaking *DelegateStakingCallerSession) GetActiveProviders(offset_ *big.Int, limit_ *big.Int) ([]common.Address, *big.Int, error) {
	return _DelegateStaking.Contract.GetActiveProviders(&_DelegateStaking.CallOpts, offset_, limit_)
}

// GetBid is a free data retrieval call binding the contract method 0x91704e1e.
//
// Solidity: function getBid(bytes32 bidId_) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_DelegateStaking *DelegateStakingCaller) GetBid(opts *bind.CallOpts, bidId_ [32]byte) (IBidStorageBid, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getBid", bidId_)

	if err != nil {
		return *new(IBidStorageBid), err
	}

	out0 := *abi.ConvertType(out[0], new(IBidStorageBid)).(*IBidStorageBid)

	return out0, err

}

// GetBid is a free data retrieval call binding the contract method 0x91704e1e.
//
// Solidity: function getBid(bytes32 bidId_) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_DelegateStaking *DelegateStakingSession) GetBid(bidId_ [32]byte) (IBidStorageBid, error) {
	return _DelegateStaking.Contract.GetBid(&_DelegateStaking.CallOpts, bidId_)
}

// GetBid is a free data retrieval call binding the contract method 0x91704e1e.
//
// Solidity: function getBid(bytes32 bidId_) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_DelegateStaking *DelegateStakingCallerSession) GetBid(bidId_ [32]byte) (IBidStorageBid, error) {
	return _DelegateStaking.Contract.GetBid(&_DelegateStaking.CallOpts, bidId_)
}

// GetFundingAccount is a free data retrieval call binding the contract method 0x775c3727.
//
// Solidity: function getFundingAccount() view returns(address)
func (_DelegateStaking *DelegateStakingCaller) GetFundingAccount(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getFundingAccount")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetFundingAccount is a free data retrieval call binding the contract method 0x775c3727.
//
// Solidity: function getFundingAccount() view returns(address)
func (_DelegateStaking *DelegateStakingSession) GetFundingAccount() (common.Address, error) {
	return _DelegateStaking.Contract.GetFundingAccount(&_DelegateStaking.CallOpts)
}

// GetFundingAccount is a free data retrieval call binding the contract method 0x775c3727.
//
// Solidity: function getFundingAccount() view returns(address)
func (_DelegateStaking *DelegateStakingCallerSession) GetFundingAccount() (common.Address, error) {
	return _DelegateStaking.Contract.GetFundingAccount(&_DelegateStaking.CallOpts)
}

// GetIsProviderActive is a free data retrieval call binding the contract method 0x63ef175d.
//
// Solidity: function getIsProviderActive(address provider_) view returns(bool)
func (_DelegateStaking *DelegateStakingCaller) GetIsProviderActive(opts *bind.CallOpts, provider_ common.Address) (bool, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getIsProviderActive", provider_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// GetIsProviderActive is a free data retrieval call binding the contract method 0x63ef175d.
//
// Solidity: function getIsProviderActive(address provider_) view returns(bool)
func (_DelegateStaking *DelegateStakingSession) GetIsProviderActive(provider_ common.Address) (bool, error) {
	return _DelegateStaking.Contract.GetIsProviderActive(&_DelegateStaking.CallOpts, provider_)
}

// GetIsProviderActive is a free data retrieval call binding the contract method 0x63ef175d.
//
// Solidity: function getIsProviderActive(address provider_) view returns(bool)
func (_DelegateStaking *DelegateStakingCallerSession) GetIsProviderActive(provider_ common.Address) (bool, error) {
	return _DelegateStaking.Contract.GetIsProviderActive(&_DelegateStaking.CallOpts, provider_)
}

// GetIsProviderApprovalUsed is a free data retrieval call binding the contract method 0xdb1cf1e2.
//
// Solidity: function getIsProviderApprovalUsed(bytes approval_) view returns(bool)
func (_DelegateStaking *DelegateStakingCaller) GetIsProviderApprovalUsed(opts *bind.CallOpts, approval_ []byte) (bool, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getIsProviderApprovalUsed", approval_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// GetIsProviderApprovalUsed is a free data retrieval call binding the contract method 0xdb1cf1e2.
//
// Solidity: function getIsProviderApprovalUsed(bytes approval_) view returns(bool)
func (_DelegateStaking *DelegateStakingSession) GetIsProviderApprovalUsed(approval_ []byte) (bool, error) {
	return _DelegateStaking.Contract.GetIsProviderApprovalUsed(&_DelegateStaking.CallOpts, approval_)
}

// GetIsProviderApprovalUsed is a free data retrieval call binding the contract method 0xdb1cf1e2.
//
// Solidity: function getIsProviderApprovalUsed(bytes approval_) view returns(bool)
func (_DelegateStaking *DelegateStakingCallerSession) GetIsProviderApprovalUsed(approval_ []byte) (bool, error) {
	return _DelegateStaking.Contract.GetIsProviderApprovalUsed(&_DelegateStaking.CallOpts, approval_)
}

// GetMaxSessionDuration is a free data retrieval call binding the contract method 0xa9756858.
//
// Solidity: function getMaxSessionDuration() view returns(uint128)
func (_DelegateStaking *DelegateStakingCaller) GetMaxSessionDuration(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getMaxSessionDuration")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMaxSessionDuration is a free data retrieval call binding the contract method 0xa9756858.
//
// Solidity: function getMaxSessionDuration() view returns(uint128)
func (_DelegateStaking *DelegateStakingSession) GetMaxSessionDuration() (*big.Int, error) {
	return _DelegateStaking.Contract.GetMaxSessionDuration(&_DelegateStaking.CallOpts)
}

// GetMaxSessionDuration is a free data retrieval call binding the contract method 0xa9756858.
//
// Solidity: function getMaxSessionDuration() view returns(uint128)
func (_DelegateStaking *DelegateStakingCallerSession) GetMaxSessionDuration() (*big.Int, error) {
	return _DelegateStaking.Contract.GetMaxSessionDuration(&_DelegateStaking.CallOpts)
}

// GetModelActiveBids is a free data retrieval call binding the contract method 0x8a683b6e.
//
// Solidity: function getModelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingCaller) GetModelActiveBids(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getModelActiveBids", modelId_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// GetModelActiveBids is a free data retrieval call binding the contract method 0x8a683b6e.
//
// Solidity: function getModelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingSession) GetModelActiveBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStaking.Contract.GetModelActiveBids(&_DelegateStaking.CallOpts, modelId_, offset_, limit_)
}

// GetModelActiveBids is a free data retrieval call binding the contract method 0x8a683b6e.
//
// Solidity: function getModelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingCallerSession) GetModelActiveBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStaking.Contract.GetModelActiveBids(&_DelegateStaking.CallOpts, modelId_, offset_, limit_)
}

// GetModelBids is a free data retrieval call binding the contract method 0xfade17b1.
//
// Solidity: function getModelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingCaller) GetModelBids(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getModelBids", modelId_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// GetModelBids is a free data retrieval call binding the contract method 0xfade17b1.
//
// Solidity: function getModelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingSession) GetModelBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStaking.Contract.GetModelBids(&_DelegateStaking.CallOpts, modelId_, offset_, limit_)
}

// GetModelBids is a free data retrieval call binding the contract method 0xfade17b1.
//
// Solidity: function getModelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingCallerSession) GetModelBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStaking.Contract.GetModelBids(&_DelegateStaking.CallOpts, modelId_, offset_, limit_)
}

// GetModelSessions is a free data retrieval call binding the contract method 0x1d78a872.
//
// Solidity: function getModelSessions(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingCaller) GetModelSessions(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getModelSessions", modelId_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// GetModelSessions is a free data retrieval call binding the contract method 0x1d78a872.
//
// Solidity: function getModelSessions(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingSession) GetModelSessions(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStaking.Contract.GetModelSessions(&_DelegateStaking.CallOpts, modelId_, offset_, limit_)
}

// GetModelSessions is a free data retrieval call binding the contract method 0x1d78a872.
//
// Solidity: function getModelSessions(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingCallerSession) GetModelSessions(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStaking.Contract.GetModelSessions(&_DelegateStaking.CallOpts, modelId_, offset_, limit_)
}

// GetModelStats is a free data retrieval call binding the contract method 0xce535723.
//
// Solidity: function getModelStats(bytes32 modelId_) view returns(((int64,int64),(int64,int64),(int64,int64),uint32))
func (_DelegateStaking *DelegateStakingCaller) GetModelStats(opts *bind.CallOpts, modelId_ [32]byte) (IStatsStorageModelStats, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getModelStats", modelId_)

	if err != nil {
		return *new(IStatsStorageModelStats), err
	}

	out0 := *abi.ConvertType(out[0], new(IStatsStorageModelStats)).(*IStatsStorageModelStats)

	return out0, err

}

// GetModelStats is a free data retrieval call binding the contract method 0xce535723.
//
// Solidity: function getModelStats(bytes32 modelId_) view returns(((int64,int64),(int64,int64),(int64,int64),uint32))
func (_DelegateStaking *DelegateStakingSession) GetModelStats(modelId_ [32]byte) (IStatsStorageModelStats, error) {
	return _DelegateStaking.Contract.GetModelStats(&_DelegateStaking.CallOpts, modelId_)
}

// GetModelStats is a free data retrieval call binding the contract method 0xce535723.
//
// Solidity: function getModelStats(bytes32 modelId_) view returns(((int64,int64),(int64,int64),(int64,int64),uint32))
func (_DelegateStaking *DelegateStakingCallerSession) GetModelStats(modelId_ [32]byte) (IStatsStorageModelStats, error) {
	return _DelegateStaking.Contract.GetModelStats(&_DelegateStaking.CallOpts, modelId_)
}

// GetPool is a free data retrieval call binding the contract method 0x068bcd8d.
//
// Solidity: function getPool(uint256 index_) view returns((uint256,uint256,uint128,uint128))
func (_DelegateStaking *DelegateStakingCaller) GetPool(opts *bind.CallOpts, index_ *big.Int) (ISessionStoragePool, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getPool", index_)

	if err != nil {
		return *new(ISessionStoragePool), err
	}

	out0 := *abi.ConvertType(out[0], new(ISessionStoragePool)).(*ISessionStoragePool)

	return out0, err

}

// GetPool is a free data retrieval call binding the contract method 0x068bcd8d.
//
// Solidity: function getPool(uint256 index_) view returns((uint256,uint256,uint128,uint128))
func (_DelegateStaking *DelegateStakingSession) GetPool(index_ *big.Int) (ISessionStoragePool, error) {
	return _DelegateStaking.Contract.GetPool(&_DelegateStaking.CallOpts, index_)
}

// GetPool is a free data retrieval call binding the contract method 0x068bcd8d.
//
// Solidity: function getPool(uint256 index_) view returns((uint256,uint256,uint128,uint128))
func (_DelegateStaking *DelegateStakingCallerSession) GetPool(index_ *big.Int) (ISessionStoragePool, error) {
	return _DelegateStaking.Contract.GetPool(&_DelegateStaking.CallOpts, index_)
}

// GetPools is a free data retrieval call binding the contract method 0x673a2a1f.
//
// Solidity: function getPools() view returns((uint256,uint256,uint128,uint128)[])
func (_DelegateStaking *DelegateStakingCaller) GetPools(opts *bind.CallOpts) ([]ISessionStoragePool, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getPools")

	if err != nil {
		return *new([]ISessionStoragePool), err
	}

	out0 := *abi.ConvertType(out[0], new([]ISessionStoragePool)).(*[]ISessionStoragePool)

	return out0, err

}

// GetPools is a free data retrieval call binding the contract method 0x673a2a1f.
//
// Solidity: function getPools() view returns((uint256,uint256,uint128,uint128)[])
func (_DelegateStaking *DelegateStakingSession) GetPools() ([]ISessionStoragePool, error) {
	return _DelegateStaking.Contract.GetPools(&_DelegateStaking.CallOpts)
}

// GetPools is a free data retrieval call binding the contract method 0x673a2a1f.
//
// Solidity: function getPools() view returns((uint256,uint256,uint128,uint128)[])
func (_DelegateStaking *DelegateStakingCallerSession) GetPools() ([]ISessionStoragePool, error) {
	return _DelegateStaking.Contract.GetPools(&_DelegateStaking.CallOpts)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider_) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_DelegateStaking *DelegateStakingCaller) GetProvider(opts *bind.CallOpts, provider_ common.Address) (IProviderStorageProvider, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getProvider", provider_)

	if err != nil {
		return *new(IProviderStorageProvider), err
	}

	out0 := *abi.ConvertType(out[0], new(IProviderStorageProvider)).(*IProviderStorageProvider)

	return out0, err

}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider_) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_DelegateStaking *DelegateStakingSession) GetProvider(provider_ common.Address) (IProviderStorageProvider, error) {
	return _DelegateStaking.Contract.GetProvider(&_DelegateStaking.CallOpts, provider_)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider_) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_DelegateStaking *DelegateStakingCallerSession) GetProvider(provider_ common.Address) (IProviderStorageProvider, error) {
	return _DelegateStaking.Contract.GetProvider(&_DelegateStaking.CallOpts, provider_)
}

// GetProviderActiveBids is a free data retrieval call binding the contract method 0xaf5b77ca.
//
// Solidity: function getProviderActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingCaller) GetProviderActiveBids(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getProviderActiveBids", provider_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// GetProviderActiveBids is a free data retrieval call binding the contract method 0xaf5b77ca.
//
// Solidity: function getProviderActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingSession) GetProviderActiveBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStaking.Contract.GetProviderActiveBids(&_DelegateStaking.CallOpts, provider_, offset_, limit_)
}

// GetProviderActiveBids is a free data retrieval call binding the contract method 0xaf5b77ca.
//
// Solidity: function getProviderActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingCallerSession) GetProviderActiveBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStaking.Contract.GetProviderActiveBids(&_DelegateStaking.CallOpts, provider_, offset_, limit_)
}

// GetProviderBids is a free data retrieval call binding the contract method 0x59d435c4.
//
// Solidity: function getProviderBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingCaller) GetProviderBids(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getProviderBids", provider_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// GetProviderBids is a free data retrieval call binding the contract method 0x59d435c4.
//
// Solidity: function getProviderBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingSession) GetProviderBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStaking.Contract.GetProviderBids(&_DelegateStaking.CallOpts, provider_, offset_, limit_)
}

// GetProviderBids is a free data retrieval call binding the contract method 0x59d435c4.
//
// Solidity: function getProviderBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingCallerSession) GetProviderBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStaking.Contract.GetProviderBids(&_DelegateStaking.CallOpts, provider_, offset_, limit_)
}

// GetProviderMinimumStake is a free data retrieval call binding the contract method 0x53c029f6.
//
// Solidity: function getProviderMinimumStake() view returns(uint256)
func (_DelegateStaking *DelegateStakingCaller) GetProviderMinimumStake(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getProviderMinimumStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProviderMinimumStake is a free data retrieval call binding the contract method 0x53c029f6.
//
// Solidity: function getProviderMinimumStake() view returns(uint256)
func (_DelegateStaking *DelegateStakingSession) GetProviderMinimumStake() (*big.Int, error) {
	return _DelegateStaking.Contract.GetProviderMinimumStake(&_DelegateStaking.CallOpts)
}

// GetProviderMinimumStake is a free data retrieval call binding the contract method 0x53c029f6.
//
// Solidity: function getProviderMinimumStake() view returns(uint256)
func (_DelegateStaking *DelegateStakingCallerSession) GetProviderMinimumStake() (*big.Int, error) {
	return _DelegateStaking.Contract.GetProviderMinimumStake(&_DelegateStaking.CallOpts)
}

// GetProviderModelStats is a free data retrieval call binding the contract method 0x1b26c116.
//
// Solidity: function getProviderModelStats(bytes32 modelId_, address provider_) view returns(((int64,int64),(int64,int64),uint32,uint32,uint32))
func (_DelegateStaking *DelegateStakingCaller) GetProviderModelStats(opts *bind.CallOpts, modelId_ [32]byte, provider_ common.Address) (IStatsStorageProviderModelStats, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getProviderModelStats", modelId_, provider_)

	if err != nil {
		return *new(IStatsStorageProviderModelStats), err
	}

	out0 := *abi.ConvertType(out[0], new(IStatsStorageProviderModelStats)).(*IStatsStorageProviderModelStats)

	return out0, err

}

// GetProviderModelStats is a free data retrieval call binding the contract method 0x1b26c116.
//
// Solidity: function getProviderModelStats(bytes32 modelId_, address provider_) view returns(((int64,int64),(int64,int64),uint32,uint32,uint32))
func (_DelegateStaking *DelegateStakingSession) GetProviderModelStats(modelId_ [32]byte, provider_ common.Address) (IStatsStorageProviderModelStats, error) {
	return _DelegateStaking.Contract.GetProviderModelStats(&_DelegateStaking.CallOpts, modelId_, provider_)
}

// GetProviderModelStats is a free data retrieval call binding the contract method 0x1b26c116.
//
// Solidity: function getProviderModelStats(bytes32 modelId_, address provider_) view returns(((int64,int64),(int64,int64),uint32,uint32,uint32))
func (_DelegateStaking *DelegateStakingCallerSession) GetProviderModelStats(modelId_ [32]byte, provider_ common.Address) (IStatsStorageProviderModelStats, error) {
	return _DelegateStaking.Contract.GetProviderModelStats(&_DelegateStaking.CallOpts, modelId_, provider_)
}

// GetProviderSessions is a free data retrieval call binding the contract method 0x87bced7d.
//
// Solidity: function getProviderSessions(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingCaller) GetProviderSessions(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getProviderSessions", provider_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// GetProviderSessions is a free data retrieval call binding the contract method 0x87bced7d.
//
// Solidity: function getProviderSessions(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingSession) GetProviderSessions(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStaking.Contract.GetProviderSessions(&_DelegateStaking.CallOpts, provider_, offset_, limit_)
}

// GetProviderSessions is a free data retrieval call binding the contract method 0x87bced7d.
//
// Solidity: function getProviderSessions(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingCallerSession) GetProviderSessions(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStaking.Contract.GetProviderSessions(&_DelegateStaking.CallOpts, provider_, offset_, limit_)
}

// GetProvidersTotalClaimed is a free data retrieval call binding the contract method 0xbdfbbada.
//
// Solidity: function getProvidersTotalClaimed() view returns(uint256)
func (_DelegateStaking *DelegateStakingCaller) GetProvidersTotalClaimed(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getProvidersTotalClaimed")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProvidersTotalClaimed is a free data retrieval call binding the contract method 0xbdfbbada.
//
// Solidity: function getProvidersTotalClaimed() view returns(uint256)
func (_DelegateStaking *DelegateStakingSession) GetProvidersTotalClaimed() (*big.Int, error) {
	return _DelegateStaking.Contract.GetProvidersTotalClaimed(&_DelegateStaking.CallOpts)
}

// GetProvidersTotalClaimed is a free data retrieval call binding the contract method 0xbdfbbada.
//
// Solidity: function getProvidersTotalClaimed() view returns(uint256)
func (_DelegateStaking *DelegateStakingCallerSession) GetProvidersTotalClaimed() (*big.Int, error) {
	return _DelegateStaking.Contract.GetProvidersTotalClaimed(&_DelegateStaking.CallOpts)
}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_DelegateStaking *DelegateStakingCaller) GetRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_DelegateStaking *DelegateStakingSession) GetRegistry() (common.Address, error) {
	return _DelegateStaking.Contract.GetRegistry(&_DelegateStaking.CallOpts)
}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_DelegateStaking *DelegateStakingCallerSession) GetRegistry() (common.Address, error) {
	return _DelegateStaking.Contract.GetRegistry(&_DelegateStaking.CallOpts)
}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId_) view returns((address,bytes32,uint256,bytes,uint256,uint256,uint128,uint128,uint128,bool,bool))
func (_DelegateStaking *DelegateStakingCaller) GetSession(opts *bind.CallOpts, sessionId_ [32]byte) (ISessionStorageSession, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getSession", sessionId_)

	if err != nil {
		return *new(ISessionStorageSession), err
	}

	out0 := *abi.ConvertType(out[0], new(ISessionStorageSession)).(*ISessionStorageSession)

	return out0, err

}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId_) view returns((address,bytes32,uint256,bytes,uint256,uint256,uint128,uint128,uint128,bool,bool))
func (_DelegateStaking *DelegateStakingSession) GetSession(sessionId_ [32]byte) (ISessionStorageSession, error) {
	return _DelegateStaking.Contract.GetSession(&_DelegateStaking.CallOpts, sessionId_)
}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId_) view returns((address,bytes32,uint256,bytes,uint256,uint256,uint128,uint128,uint128,bool,bool))
func (_DelegateStaking *DelegateStakingCallerSession) GetSession(sessionId_ [32]byte) (ISessionStorageSession, error) {
	return _DelegateStaking.Contract.GetSession(&_DelegateStaking.CallOpts, sessionId_)
}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_DelegateStaking *DelegateStakingCaller) GetToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_DelegateStaking *DelegateStakingSession) GetToken() (common.Address, error) {
	return _DelegateStaking.Contract.GetToken(&_DelegateStaking.CallOpts)
}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_DelegateStaking *DelegateStakingCallerSession) GetToken() (common.Address, error) {
	return _DelegateStaking.Contract.GetToken(&_DelegateStaking.CallOpts)
}

// GetTotalSessions is a free data retrieval call binding the contract method 0x76400935.
//
// Solidity: function getTotalSessions(address provider_) view returns(uint256)
func (_DelegateStaking *DelegateStakingCaller) GetTotalSessions(opts *bind.CallOpts, provider_ common.Address) (*big.Int, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getTotalSessions", provider_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalSessions is a free data retrieval call binding the contract method 0x76400935.
//
// Solidity: function getTotalSessions(address provider_) view returns(uint256)
func (_DelegateStaking *DelegateStakingSession) GetTotalSessions(provider_ common.Address) (*big.Int, error) {
	return _DelegateStaking.Contract.GetTotalSessions(&_DelegateStaking.CallOpts, provider_)
}

// GetTotalSessions is a free data retrieval call binding the contract method 0x76400935.
//
// Solidity: function getTotalSessions(address provider_) view returns(uint256)
func (_DelegateStaking *DelegateStakingCallerSession) GetTotalSessions(provider_ common.Address) (*big.Int, error) {
	return _DelegateStaking.Contract.GetTotalSessions(&_DelegateStaking.CallOpts, provider_)
}

// GetUserSessions is a free data retrieval call binding the contract method 0xeb7764bb.
//
// Solidity: function getUserSessions(address user_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingCaller) GetUserSessions(opts *bind.CallOpts, user_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "getUserSessions", user_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// GetUserSessions is a free data retrieval call binding the contract method 0xeb7764bb.
//
// Solidity: function getUserSessions(address user_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingSession) GetUserSessions(user_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStaking.Contract.GetUserSessions(&_DelegateStaking.CallOpts, user_, offset_, limit_)
}

// GetUserSessions is a free data retrieval call binding the contract method 0xeb7764bb.
//
// Solidity: function getUserSessions(address user_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStaking *DelegateStakingCallerSession) GetUserSessions(user_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStaking.Contract.GetUserSessions(&_DelegateStaking.CallOpts, user_, offset_, limit_)
}

// IsBidActive is a free data retrieval call binding the contract method 0x1345df58.
//
// Solidity: function isBidActive(bytes32 bidId_) view returns(bool)
func (_DelegateStaking *DelegateStakingCaller) IsBidActive(opts *bind.CallOpts, bidId_ [32]byte) (bool, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "isBidActive", bidId_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsBidActive is a free data retrieval call binding the contract method 0x1345df58.
//
// Solidity: function isBidActive(bytes32 bidId_) view returns(bool)
func (_DelegateStaking *DelegateStakingSession) IsBidActive(bidId_ [32]byte) (bool, error) {
	return _DelegateStaking.Contract.IsBidActive(&_DelegateStaking.CallOpts, bidId_)
}

// IsBidActive is a free data retrieval call binding the contract method 0x1345df58.
//
// Solidity: function isBidActive(bytes32 bidId_) view returns(bool)
func (_DelegateStaking *DelegateStakingCallerSession) IsBidActive(bidId_ [32]byte) (bool, error) {
	return _DelegateStaking.Contract.IsBidActive(&_DelegateStaking.CallOpts, bidId_)
}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_DelegateStaking *DelegateStakingCaller) IsRightsDelegated(opts *bind.CallOpts, delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	var out []interface{}
	err := _DelegateStaking.contract.Call(opts, &out, "isRightsDelegated", delegatee_, delegator_, rights_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_DelegateStaking *DelegateStakingSession) IsRightsDelegated(delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	return _DelegateStaking.Contract.IsRightsDelegated(&_DelegateStaking.CallOpts, delegatee_, delegator_, rights_)
}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_DelegateStaking *DelegateStakingCallerSession) IsRightsDelegated(delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	return _DelegateStaking.Contract.IsRightsDelegated(&_DelegateStaking.CallOpts, delegatee_, delegator_, rights_)
}

// ClosePoolSession is a paid mutator transaction binding the contract method 0x8ede1625.
//
// Solidity: function closePoolSession(bytes receiptEncoded_, bytes signature_) returns()
func (_DelegateStaking *DelegateStakingTransactor) ClosePoolSession(opts *bind.TransactOpts, receiptEncoded_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _DelegateStaking.contract.Transact(opts, "closePoolSession", receiptEncoded_, signature_)
}

// ClosePoolSession is a paid mutator transaction binding the contract method 0x8ede1625.
//
// Solidity: function closePoolSession(bytes receiptEncoded_, bytes signature_) returns()
func (_DelegateStaking *DelegateStakingSession) ClosePoolSession(receiptEncoded_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _DelegateStaking.Contract.ClosePoolSession(&_DelegateStaking.TransactOpts, receiptEncoded_, signature_)
}

// ClosePoolSession is a paid mutator transaction binding the contract method 0x8ede1625.
//
// Solidity: function closePoolSession(bytes receiptEncoded_, bytes signature_) returns()
func (_DelegateStaking *DelegateStakingTransactorSession) ClosePoolSession(receiptEncoded_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _DelegateStaking.Contract.ClosePoolSession(&_DelegateStaking.TransactOpts, receiptEncoded_, signature_)
}

// OpenSessionFromPool is a paid mutator transaction binding the contract method 0x15624cf6.
//
// Solidity: function openSessionFromPool(address hot_, uint256 amount_, bytes approvalEncoded_, bytes signature_) returns(bytes32)
func (_DelegateStaking *DelegateStakingTransactor) OpenSessionFromPool(opts *bind.TransactOpts, hot_ common.Address, amount_ *big.Int, approvalEncoded_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _DelegateStaking.contract.Transact(opts, "openSessionFromPool", hot_, amount_, approvalEncoded_, signature_)
}

// OpenSessionFromPool is a paid mutator transaction binding the contract method 0x15624cf6.
//
// Solidity: function openSessionFromPool(address hot_, uint256 amount_, bytes approvalEncoded_, bytes signature_) returns(bytes32)
func (_DelegateStaking *DelegateStakingSession) OpenSessionFromPool(hot_ common.Address, amount_ *big.Int, approvalEncoded_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _DelegateStaking.Contract.OpenSessionFromPool(&_DelegateStaking.TransactOpts, hot_, amount_, approvalEncoded_, signature_)
}

// OpenSessionFromPool is a paid mutator transaction binding the contract method 0x15624cf6.
//
// Solidity: function openSessionFromPool(address hot_, uint256 amount_, bytes approvalEncoded_, bytes signature_) returns(bytes32)
func (_DelegateStaking *DelegateStakingTransactorSession) OpenSessionFromPool(hot_ common.Address, amount_ *big.Int, approvalEncoded_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _DelegateStaking.Contract.OpenSessionFromPool(&_DelegateStaking.TransactOpts, hot_, amount_, approvalEncoded_, signature_)
}

// SettleExpiredSession is a paid mutator transaction binding the contract method 0xc1934a65.
//
// Solidity: function settleExpiredSession(bytes32 sessionId_) returns()
func (_DelegateStaking *DelegateStakingTransactor) SettleExpiredSession(opts *bind.TransactOpts, sessionId_ [32]byte) (*types.Transaction, error) {
	return _DelegateStaking.contract.Transact(opts, "settleExpiredSession", sessionId_)
}

// SettleExpiredSession is a paid mutator transaction binding the contract method 0xc1934a65.
//
// Solidity: function settleExpiredSession(bytes32 sessionId_) returns()
func (_DelegateStaking *DelegateStakingSession) SettleExpiredSession(sessionId_ [32]byte) (*types.Transaction, error) {
	return _DelegateStaking.Contract.SettleExpiredSession(&_DelegateStaking.TransactOpts, sessionId_)
}

// SettleExpiredSession is a paid mutator transaction binding the contract method 0xc1934a65.
//
// Solidity: function settleExpiredSession(bytes32 sessionId_) returns()
func (_DelegateStaking *DelegateStakingTransactorSession) SettleExpiredSession(sessionId_ [32]byte) (*types.Transaction, error) {
	return _DelegateStaking.Contract.SettleExpiredSession(&_DelegateStaking.TransactOpts, sessionId_)
}

// DelegateStakingDelegateStakingAllowanceFundedIterator is returned from FilterDelegateStakingAllowanceFunded and is used to iterate over the raw logs and unpacked data for DelegateStakingAllowanceFunded events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingAllowanceFundedIterator struct {
	Event *DelegateStakingDelegateStakingAllowanceFunded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingAllowanceFundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingAllowanceFunded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingAllowanceFunded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingAllowanceFundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingAllowanceFundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingAllowanceFunded represents a DelegateStakingAllowanceFunded event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingAllowanceFunded struct {
	Hot    common.Address
	Funder common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingAllowanceFunded is a free log retrieval operation binding the contract event 0xab72e208ad5755afbd9bcdfe0b4304d209b428685d5ee350433f417b117f38f0.
//
// Solidity: event DelegateStakingAllowanceFunded(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingAllowanceFunded(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingDelegateStakingAllowanceFundedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingAllowanceFunded", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingAllowanceFundedIterator{contract: _DelegateStaking.contract, event: "DelegateStakingAllowanceFunded", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingAllowanceFunded is a free log subscription operation binding the contract event 0xab72e208ad5755afbd9bcdfe0b4304d209b428685d5ee350433f417b117f38f0.
//
// Solidity: event DelegateStakingAllowanceFunded(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingAllowanceFunded(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingAllowanceFunded, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingAllowanceFunded", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingAllowanceFunded)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingAllowanceFunded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingAllowanceFunded is a log parse operation binding the contract event 0xab72e208ad5755afbd9bcdfe0b4304d209b428685d5ee350433f417b117f38f0.
//
// Solidity: event DelegateStakingAllowanceFunded(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingAllowanceFunded(log types.Log) (*DelegateStakingDelegateStakingAllowanceFunded, error) {
	event := new(DelegateStakingDelegateStakingAllowanceFunded)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingAllowanceFunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingAllowanceGrantedIterator is returned from FilterDelegateStakingAllowanceGranted and is used to iterate over the raw logs and unpacked data for DelegateStakingAllowanceGranted events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingAllowanceGrantedIterator struct {
	Event *DelegateStakingDelegateStakingAllowanceGranted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingAllowanceGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingAllowanceGranted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingAllowanceGranted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingAllowanceGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingAllowanceGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingAllowanceGranted represents a DelegateStakingAllowanceGranted event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingAllowanceGranted struct {
	Hot                  common.Address
	Funder               common.Address
	CumulativeFundingCap *big.Int
	Expiry               *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingAllowanceGranted is a free log retrieval operation binding the contract event 0x743ecb36e3b0fa6c16a5e8cd3ae70e657b294fd68ffbdfdc73d3a89d04addf46.
//
// Solidity: event DelegateStakingAllowanceGranted(address indexed hot, address indexed funder, uint256 cumulativeFundingCap, uint128 expiry)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingAllowanceGranted(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingDelegateStakingAllowanceGrantedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingAllowanceGranted", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingAllowanceGrantedIterator{contract: _DelegateStaking.contract, event: "DelegateStakingAllowanceGranted", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingAllowanceGranted is a free log subscription operation binding the contract event 0x743ecb36e3b0fa6c16a5e8cd3ae70e657b294fd68ffbdfdc73d3a89d04addf46.
//
// Solidity: event DelegateStakingAllowanceGranted(address indexed hot, address indexed funder, uint256 cumulativeFundingCap, uint128 expiry)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingAllowanceGranted(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingAllowanceGranted, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingAllowanceGranted", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingAllowanceGranted)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingAllowanceGranted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingAllowanceGranted is a log parse operation binding the contract event 0x743ecb36e3b0fa6c16a5e8cd3ae70e657b294fd68ffbdfdc73d3a89d04addf46.
//
// Solidity: event DelegateStakingAllowanceGranted(address indexed hot, address indexed funder, uint256 cumulativeFundingCap, uint128 expiry)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingAllowanceGranted(log types.Log) (*DelegateStakingDelegateStakingAllowanceGranted, error) {
	event := new(DelegateStakingDelegateStakingAllowanceGranted)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingAllowanceGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingAllowanceRevokedIterator is returned from FilterDelegateStakingAllowanceRevoked and is used to iterate over the raw logs and unpacked data for DelegateStakingAllowanceRevoked events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingAllowanceRevokedIterator struct {
	Event *DelegateStakingDelegateStakingAllowanceRevoked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingAllowanceRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingAllowanceRevoked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingAllowanceRevoked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingAllowanceRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingAllowanceRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingAllowanceRevoked represents a DelegateStakingAllowanceRevoked event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingAllowanceRevoked struct {
	Hot    common.Address
	Funder common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingAllowanceRevoked is a free log retrieval operation binding the contract event 0x641afbae43e9aa28104693da9d716e2295e9efaf231c0bfec7dd6f949d3bd726.
//
// Solidity: event DelegateStakingAllowanceRevoked(address indexed hot, address indexed funder)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingAllowanceRevoked(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingDelegateStakingAllowanceRevokedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingAllowanceRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingAllowanceRevokedIterator{contract: _DelegateStaking.contract, event: "DelegateStakingAllowanceRevoked", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingAllowanceRevoked is a free log subscription operation binding the contract event 0x641afbae43e9aa28104693da9d716e2295e9efaf231c0bfec7dd6f949d3bd726.
//
// Solidity: event DelegateStakingAllowanceRevoked(address indexed hot, address indexed funder)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingAllowanceRevoked(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingAllowanceRevoked, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingAllowanceRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingAllowanceRevoked)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingAllowanceRevoked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingAllowanceRevoked is a log parse operation binding the contract event 0x641afbae43e9aa28104693da9d716e2295e9efaf231c0bfec7dd6f949d3bd726.
//
// Solidity: event DelegateStakingAllowanceRevoked(address indexed hot, address indexed funder)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingAllowanceRevoked(log types.Log) (*DelegateStakingDelegateStakingAllowanceRevoked, error) {
	event := new(DelegateStakingDelegateStakingAllowanceRevoked)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingAllowanceRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingMaxActiveFundersUpdatedIterator is returned from FilterDelegateStakingMaxActiveFundersUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMaxActiveFundersUpdated events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingMaxActiveFundersUpdatedIterator struct {
	Event *DelegateStakingDelegateStakingMaxActiveFundersUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingMaxActiveFundersUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingMaxActiveFundersUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingMaxActiveFundersUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingMaxActiveFundersUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingMaxActiveFundersUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingMaxActiveFundersUpdated represents a DelegateStakingMaxActiveFundersUpdated event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingMaxActiveFundersUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMaxActiveFundersUpdated is a free log retrieval operation binding the contract event 0x6018427c349577edf6a6c0b891cf09b86b47cc1834917516e6da404941ef43f5.
//
// Solidity: event DelegateStakingMaxActiveFundersUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingMaxActiveFundersUpdated(opts *bind.FilterOpts) (*DelegateStakingDelegateStakingMaxActiveFundersUpdatedIterator, error) {

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingMaxActiveFundersUpdated")
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingMaxActiveFundersUpdatedIterator{contract: _DelegateStaking.contract, event: "DelegateStakingMaxActiveFundersUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMaxActiveFundersUpdated is a free log subscription operation binding the contract event 0x6018427c349577edf6a6c0b891cf09b86b47cc1834917516e6da404941ef43f5.
//
// Solidity: event DelegateStakingMaxActiveFundersUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingMaxActiveFundersUpdated(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingMaxActiveFundersUpdated) (event.Subscription, error) {

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingMaxActiveFundersUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingMaxActiveFundersUpdated)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingMaxActiveFundersUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingMaxActiveFundersUpdated is a log parse operation binding the contract event 0x6018427c349577edf6a6c0b891cf09b86b47cc1834917516e6da404941ef43f5.
//
// Solidity: event DelegateStakingMaxActiveFundersUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingMaxActiveFundersUpdated(log types.Log) (*DelegateStakingDelegateStakingMaxActiveFundersUpdated, error) {
	event := new(DelegateStakingDelegateStakingMaxActiveFundersUpdated)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingMaxActiveFundersUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdatedIterator is returned from FilterDelegateStakingMaxAutoReleaseDaysUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMaxAutoReleaseDaysUpdated events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdatedIterator struct {
	Event *DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdated represents a DelegateStakingMaxAutoReleaseDaysUpdated event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMaxAutoReleaseDaysUpdated is a free log retrieval operation binding the contract event 0xc9ca6843c5a726f8df173b27882fb647575843c62e31760e9b7f90c94518aa2a.
//
// Solidity: event DelegateStakingMaxAutoReleaseDaysUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingMaxAutoReleaseDaysUpdated(opts *bind.FilterOpts) (*DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdatedIterator, error) {

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingMaxAutoReleaseDaysUpdated")
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdatedIterator{contract: _DelegateStaking.contract, event: "DelegateStakingMaxAutoReleaseDaysUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMaxAutoReleaseDaysUpdated is a free log subscription operation binding the contract event 0xc9ca6843c5a726f8df173b27882fb647575843c62e31760e9b7f90c94518aa2a.
//
// Solidity: event DelegateStakingMaxAutoReleaseDaysUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingMaxAutoReleaseDaysUpdated(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdated) (event.Subscription, error) {

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingMaxAutoReleaseDaysUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdated)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingMaxAutoReleaseDaysUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingMaxAutoReleaseDaysUpdated is a log parse operation binding the contract event 0xc9ca6843c5a726f8df173b27882fb647575843c62e31760e9b7f90c94518aa2a.
//
// Solidity: event DelegateStakingMaxAutoReleaseDaysUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingMaxAutoReleaseDaysUpdated(log types.Log) (*DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdated, error) {
	event := new(DelegateStakingDelegateStakingMaxAutoReleaseDaysUpdated)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingMaxAutoReleaseDaysUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingMaxAutoServiceUpdatedIterator is returned from FilterDelegateStakingMaxAutoServiceUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMaxAutoServiceUpdated events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingMaxAutoServiceUpdatedIterator struct {
	Event *DelegateStakingDelegateStakingMaxAutoServiceUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingMaxAutoServiceUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingMaxAutoServiceUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingMaxAutoServiceUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingMaxAutoServiceUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingMaxAutoServiceUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingMaxAutoServiceUpdated represents a DelegateStakingMaxAutoServiceUpdated event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingMaxAutoServiceUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMaxAutoServiceUpdated is a free log retrieval operation binding the contract event 0x4f927d3e460674eab7e38dac3592136fe961de0c14533b919652a814a7510adc.
//
// Solidity: event DelegateStakingMaxAutoServiceUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingMaxAutoServiceUpdated(opts *bind.FilterOpts) (*DelegateStakingDelegateStakingMaxAutoServiceUpdatedIterator, error) {

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingMaxAutoServiceUpdated")
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingMaxAutoServiceUpdatedIterator{contract: _DelegateStaking.contract, event: "DelegateStakingMaxAutoServiceUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMaxAutoServiceUpdated is a free log subscription operation binding the contract event 0x4f927d3e460674eab7e38dac3592136fe961de0c14533b919652a814a7510adc.
//
// Solidity: event DelegateStakingMaxAutoServiceUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingMaxAutoServiceUpdated(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingMaxAutoServiceUpdated) (event.Subscription, error) {

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingMaxAutoServiceUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingMaxAutoServiceUpdated)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingMaxAutoServiceUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingMaxAutoServiceUpdated is a log parse operation binding the contract event 0x4f927d3e460674eab7e38dac3592136fe961de0c14533b919652a814a7510adc.
//
// Solidity: event DelegateStakingMaxAutoServiceUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingMaxAutoServiceUpdated(log types.Log) (*DelegateStakingDelegateStakingMaxAutoServiceUpdated, error) {
	event := new(DelegateStakingDelegateStakingMaxAutoServiceUpdated)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingMaxAutoServiceUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingMinPrincipalUpdatedIterator is returned from FilterDelegateStakingMinPrincipalUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMinPrincipalUpdated events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingMinPrincipalUpdatedIterator struct {
	Event *DelegateStakingDelegateStakingMinPrincipalUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingMinPrincipalUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingMinPrincipalUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingMinPrincipalUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingMinPrincipalUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingMinPrincipalUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingMinPrincipalUpdated represents a DelegateStakingMinPrincipalUpdated event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingMinPrincipalUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMinPrincipalUpdated is a free log retrieval operation binding the contract event 0x06ae87e705be2a5e7c6ff17a5d3d836834554f7d0aadaef3abb4343a202590a2.
//
// Solidity: event DelegateStakingMinPrincipalUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingMinPrincipalUpdated(opts *bind.FilterOpts) (*DelegateStakingDelegateStakingMinPrincipalUpdatedIterator, error) {

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingMinPrincipalUpdated")
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingMinPrincipalUpdatedIterator{contract: _DelegateStaking.contract, event: "DelegateStakingMinPrincipalUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMinPrincipalUpdated is a free log subscription operation binding the contract event 0x06ae87e705be2a5e7c6ff17a5d3d836834554f7d0aadaef3abb4343a202590a2.
//
// Solidity: event DelegateStakingMinPrincipalUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingMinPrincipalUpdated(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingMinPrincipalUpdated) (event.Subscription, error) {

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingMinPrincipalUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingMinPrincipalUpdated)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingMinPrincipalUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingMinPrincipalUpdated is a log parse operation binding the contract event 0x06ae87e705be2a5e7c6ff17a5d3d836834554f7d0aadaef3abb4343a202590a2.
//
// Solidity: event DelegateStakingMinPrincipalUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingMinPrincipalUpdated(log types.Log) (*DelegateStakingDelegateStakingMinPrincipalUpdated, error) {
	event := new(DelegateStakingDelegateStakingMinPrincipalUpdated)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingMinPrincipalUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingPoolDrawnIterator is returned from FilterDelegateStakingPoolDrawn and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolDrawn events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingPoolDrawnIterator struct {
	Event *DelegateStakingDelegateStakingPoolDrawn // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingPoolDrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingPoolDrawn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingPoolDrawn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingPoolDrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingPoolDrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingPoolDrawn represents a DelegateStakingPoolDrawn event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingPoolDrawn struct {
	SessionId [32]byte
	Hot       common.Address
	Funder    common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolDrawn is a free log retrieval operation binding the contract event 0xd178a2f3e5da2cc00732e1b88d2670a1eaa4f88a0b86e8790237a677f1fa0985.
//
// Solidity: event DelegateStakingPoolDrawn(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingPoolDrawn(opts *bind.FilterOpts, sessionId [][32]byte, hot []common.Address, funder []common.Address) (*DelegateStakingDelegateStakingPoolDrawnIterator, error) {

	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingPoolDrawn", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingPoolDrawnIterator{contract: _DelegateStaking.contract, event: "DelegateStakingPoolDrawn", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolDrawn is a free log subscription operation binding the contract event 0xd178a2f3e5da2cc00732e1b88d2670a1eaa4f88a0b86e8790237a677f1fa0985.
//
// Solidity: event DelegateStakingPoolDrawn(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingPoolDrawn(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingPoolDrawn, sessionId [][32]byte, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingPoolDrawn", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingPoolDrawn)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingPoolDrawn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingPoolDrawn is a log parse operation binding the contract event 0xd178a2f3e5da2cc00732e1b88d2670a1eaa4f88a0b86e8790237a677f1fa0985.
//
// Solidity: event DelegateStakingPoolDrawn(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingPoolDrawn(log types.Log) (*DelegateStakingDelegateStakingPoolDrawn, error) {
	event := new(DelegateStakingDelegateStakingPoolDrawn)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingPoolDrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingPoolFunderApprovalRevokedIterator is returned from FilterDelegateStakingPoolFunderApprovalRevoked and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolFunderApprovalRevoked events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingPoolFunderApprovalRevokedIterator struct {
	Event *DelegateStakingDelegateStakingPoolFunderApprovalRevoked // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingPoolFunderApprovalRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingPoolFunderApprovalRevoked)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingPoolFunderApprovalRevoked)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingPoolFunderApprovalRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingPoolFunderApprovalRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingPoolFunderApprovalRevoked represents a DelegateStakingPoolFunderApprovalRevoked event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingPoolFunderApprovalRevoked struct {
	Hot    common.Address
	Funder common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolFunderApprovalRevoked is a free log retrieval operation binding the contract event 0x5ebb99e7e14d99bff22536b105e92a32d4fbb701f93b0d5b008cc0a0b876aeb2.
//
// Solidity: event DelegateStakingPoolFunderApprovalRevoked(address indexed hot, address indexed funder)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingPoolFunderApprovalRevoked(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingDelegateStakingPoolFunderApprovalRevokedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingPoolFunderApprovalRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingPoolFunderApprovalRevokedIterator{contract: _DelegateStaking.contract, event: "DelegateStakingPoolFunderApprovalRevoked", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolFunderApprovalRevoked is a free log subscription operation binding the contract event 0x5ebb99e7e14d99bff22536b105e92a32d4fbb701f93b0d5b008cc0a0b876aeb2.
//
// Solidity: event DelegateStakingPoolFunderApprovalRevoked(address indexed hot, address indexed funder)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingPoolFunderApprovalRevoked(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingPoolFunderApprovalRevoked, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingPoolFunderApprovalRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingPoolFunderApprovalRevoked)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingPoolFunderApprovalRevoked", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingPoolFunderApprovalRevoked is a log parse operation binding the contract event 0x5ebb99e7e14d99bff22536b105e92a32d4fbb701f93b0d5b008cc0a0b876aeb2.
//
// Solidity: event DelegateStakingPoolFunderApprovalRevoked(address indexed hot, address indexed funder)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingPoolFunderApprovalRevoked(log types.Log) (*DelegateStakingDelegateStakingPoolFunderApprovalRevoked, error) {
	event := new(DelegateStakingDelegateStakingPoolFunderApprovalRevoked)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingPoolFunderApprovalRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingPoolFunderApprovedIterator is returned from FilterDelegateStakingPoolFunderApproved and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolFunderApproved events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingPoolFunderApprovedIterator struct {
	Event *DelegateStakingDelegateStakingPoolFunderApproved // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingPoolFunderApprovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingPoolFunderApproved)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingPoolFunderApproved)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingPoolFunderApprovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingPoolFunderApprovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingPoolFunderApproved represents a DelegateStakingPoolFunderApproved event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingPoolFunderApproved struct {
	Hot    common.Address
	Funder common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolFunderApproved is a free log retrieval operation binding the contract event 0x2a6f85b141a6737d1530bd4b8c916eee97ba403dae709170266f72b6ec58f739.
//
// Solidity: event DelegateStakingPoolFunderApproved(address indexed hot, address indexed funder)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingPoolFunderApproved(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingDelegateStakingPoolFunderApprovedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingPoolFunderApproved", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingPoolFunderApprovedIterator{contract: _DelegateStaking.contract, event: "DelegateStakingPoolFunderApproved", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolFunderApproved is a free log subscription operation binding the contract event 0x2a6f85b141a6737d1530bd4b8c916eee97ba403dae709170266f72b6ec58f739.
//
// Solidity: event DelegateStakingPoolFunderApproved(address indexed hot, address indexed funder)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingPoolFunderApproved(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingPoolFunderApproved, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingPoolFunderApproved", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingPoolFunderApproved)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingPoolFunderApproved", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingPoolFunderApproved is a log parse operation binding the contract event 0x2a6f85b141a6737d1530bd4b8c916eee97ba403dae709170266f72b6ec58f739.
//
// Solidity: event DelegateStakingPoolFunderApproved(address indexed hot, address indexed funder)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingPoolFunderApproved(log types.Log) (*DelegateStakingDelegateStakingPoolFunderApproved, error) {
	event := new(DelegateStakingDelegateStakingPoolFunderApproved)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingPoolFunderApproved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingPoolLockCreatedIterator is returned from FilterDelegateStakingPoolLockCreated and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolLockCreated events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingPoolLockCreatedIterator struct {
	Event *DelegateStakingDelegateStakingPoolLockCreated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingPoolLockCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingPoolLockCreated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingPoolLockCreated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingPoolLockCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingPoolLockCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingPoolLockCreated represents a DelegateStakingPoolLockCreated event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingPoolLockCreated struct {
	Hot       common.Address
	ReleaseAt *big.Int
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolLockCreated is a free log retrieval operation binding the contract event 0x87eb288bde7ca5d84aa5e85a8b81f8d7c31d8e19049823dab5552d8bb461dc0e.
//
// Solidity: event DelegateStakingPoolLockCreated(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingPoolLockCreated(opts *bind.FilterOpts, hot []common.Address, releaseAt []*big.Int) (*DelegateStakingDelegateStakingPoolLockCreatedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingPoolLockCreated", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingPoolLockCreatedIterator{contract: _DelegateStaking.contract, event: "DelegateStakingPoolLockCreated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolLockCreated is a free log subscription operation binding the contract event 0x87eb288bde7ca5d84aa5e85a8b81f8d7c31d8e19049823dab5552d8bb461dc0e.
//
// Solidity: event DelegateStakingPoolLockCreated(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingPoolLockCreated(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingPoolLockCreated, hot []common.Address, releaseAt []*big.Int) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingPoolLockCreated", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingPoolLockCreated)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingPoolLockCreated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingPoolLockCreated is a log parse operation binding the contract event 0x87eb288bde7ca5d84aa5e85a8b81f8d7c31d8e19049823dab5552d8bb461dc0e.
//
// Solidity: event DelegateStakingPoolLockCreated(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingPoolLockCreated(log types.Log) (*DelegateStakingDelegateStakingPoolLockCreated, error) {
	event := new(DelegateStakingDelegateStakingPoolLockCreated)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingPoolLockCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingPoolLockReleasedIterator is returned from FilterDelegateStakingPoolLockReleased and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolLockReleased events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingPoolLockReleasedIterator struct {
	Event *DelegateStakingDelegateStakingPoolLockReleased // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingPoolLockReleasedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingPoolLockReleased)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingPoolLockReleased)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingPoolLockReleasedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingPoolLockReleasedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingPoolLockReleased represents a DelegateStakingPoolLockReleased event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingPoolLockReleased struct {
	Hot       common.Address
	ReleaseAt *big.Int
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolLockReleased is a free log retrieval operation binding the contract event 0xd04817baed5e7c2b01f59b2218c05170e733b0e2a3864693721a2fc89e5f7a4f.
//
// Solidity: event DelegateStakingPoolLockReleased(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingPoolLockReleased(opts *bind.FilterOpts, hot []common.Address, releaseAt []*big.Int) (*DelegateStakingDelegateStakingPoolLockReleasedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingPoolLockReleased", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingPoolLockReleasedIterator{contract: _DelegateStaking.contract, event: "DelegateStakingPoolLockReleased", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolLockReleased is a free log subscription operation binding the contract event 0xd04817baed5e7c2b01f59b2218c05170e733b0e2a3864693721a2fc89e5f7a4f.
//
// Solidity: event DelegateStakingPoolLockReleased(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingPoolLockReleased(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingPoolLockReleased, hot []common.Address, releaseAt []*big.Int) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingPoolLockReleased", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingPoolLockReleased)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingPoolLockReleased", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingPoolLockReleased is a log parse operation binding the contract event 0xd04817baed5e7c2b01f59b2218c05170e733b0e2a3864693721a2fc89e5f7a4f.
//
// Solidity: event DelegateStakingPoolLockReleased(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingPoolLockReleased(log types.Log) (*DelegateStakingDelegateStakingPoolLockReleased, error) {
	event := new(DelegateStakingDelegateStakingPoolLockReleased)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingPoolLockReleased", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingPoolReturnedIterator is returned from FilterDelegateStakingPoolReturned and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolReturned events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingPoolReturnedIterator struct {
	Event *DelegateStakingDelegateStakingPoolReturned // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingPoolReturnedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingPoolReturned)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingPoolReturned)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingPoolReturnedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingPoolReturnedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingPoolReturned represents a DelegateStakingPoolReturned event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingPoolReturned struct {
	SessionId [32]byte
	Hot       common.Address
	Funder    common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolReturned is a free log retrieval operation binding the contract event 0x66524d7951752d8716b27064a7d50819e4db87609ea9eca1f6b1d42f359a061d.
//
// Solidity: event DelegateStakingPoolReturned(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingPoolReturned(opts *bind.FilterOpts, sessionId [][32]byte, hot []common.Address, funder []common.Address) (*DelegateStakingDelegateStakingPoolReturnedIterator, error) {

	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingPoolReturned", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingPoolReturnedIterator{contract: _DelegateStaking.contract, event: "DelegateStakingPoolReturned", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolReturned is a free log subscription operation binding the contract event 0x66524d7951752d8716b27064a7d50819e4db87609ea9eca1f6b1d42f359a061d.
//
// Solidity: event DelegateStakingPoolReturned(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingPoolReturned(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingPoolReturned, sessionId [][32]byte, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingPoolReturned", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingPoolReturned)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingPoolReturned", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingPoolReturned is a log parse operation binding the contract event 0x66524d7951752d8716b27064a7d50819e4db87609ea9eca1f6b1d42f359a061d.
//
// Solidity: event DelegateStakingPoolReturned(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingPoolReturned(log types.Log) (*DelegateStakingDelegateStakingPoolReturned, error) {
	event := new(DelegateStakingDelegateStakingPoolReturned)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingPoolReturned", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingPoolSessionSettledIterator is returned from FilterDelegateStakingPoolSessionSettled and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolSessionSettled events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingPoolSessionSettledIterator struct {
	Event *DelegateStakingDelegateStakingPoolSessionSettled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingPoolSessionSettledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingPoolSessionSettled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingPoolSessionSettled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingPoolSessionSettledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingPoolSessionSettledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingPoolSessionSettled represents a DelegateStakingPoolSessionSettled event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingPoolSessionSettled struct {
	SessionId [32]byte
	Hot       common.Address
	Provider  common.Address
	Caller    common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolSessionSettled is a free log retrieval operation binding the contract event 0x518bf7ca4dc0c0c2144467883693da5fd9b0532de6539122309ecff49a330215.
//
// Solidity: event DelegateStakingPoolSessionSettled(bytes32 indexed sessionId, address indexed hot, address indexed provider, address caller)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingPoolSessionSettled(opts *bind.FilterOpts, sessionId [][32]byte, hot []common.Address, provider []common.Address) (*DelegateStakingDelegateStakingPoolSessionSettledIterator, error) {

	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingPoolSessionSettled", sessionIdRule, hotRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingPoolSessionSettledIterator{contract: _DelegateStaking.contract, event: "DelegateStakingPoolSessionSettled", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolSessionSettled is a free log subscription operation binding the contract event 0x518bf7ca4dc0c0c2144467883693da5fd9b0532de6539122309ecff49a330215.
//
// Solidity: event DelegateStakingPoolSessionSettled(bytes32 indexed sessionId, address indexed hot, address indexed provider, address caller)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingPoolSessionSettled(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingPoolSessionSettled, sessionId [][32]byte, hot []common.Address, provider []common.Address) (event.Subscription, error) {

	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingPoolSessionSettled", sessionIdRule, hotRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingPoolSessionSettled)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingPoolSessionSettled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingPoolSessionSettled is a log parse operation binding the contract event 0x518bf7ca4dc0c0c2144467883693da5fd9b0532de6539122309ecff49a330215.
//
// Solidity: event DelegateStakingPoolSessionSettled(bytes32 indexed sessionId, address indexed hot, address indexed provider, address caller)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingPoolSessionSettled(log types.Log) (*DelegateStakingDelegateStakingPoolSessionSettled, error) {
	event := new(DelegateStakingDelegateStakingPoolSessionSettled)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingPoolSessionSettled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingSettlementGraceUpdatedIterator is returned from FilterDelegateStakingSettlementGraceUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingSettlementGraceUpdated events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingSettlementGraceUpdatedIterator struct {
	Event *DelegateStakingDelegateStakingSettlementGraceUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingSettlementGraceUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingSettlementGraceUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingSettlementGraceUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingSettlementGraceUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingSettlementGraceUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingSettlementGraceUpdated represents a DelegateStakingSettlementGraceUpdated event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingSettlementGraceUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingSettlementGraceUpdated is a free log retrieval operation binding the contract event 0x23d4c1527dfe24f1a043524d5c7c2b730f3fb2841478487080916494bc8487f2.
//
// Solidity: event DelegateStakingSettlementGraceUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingSettlementGraceUpdated(opts *bind.FilterOpts) (*DelegateStakingDelegateStakingSettlementGraceUpdatedIterator, error) {

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingSettlementGraceUpdated")
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingSettlementGraceUpdatedIterator{contract: _DelegateStaking.contract, event: "DelegateStakingSettlementGraceUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingSettlementGraceUpdated is a free log subscription operation binding the contract event 0x23d4c1527dfe24f1a043524d5c7c2b730f3fb2841478487080916494bc8487f2.
//
// Solidity: event DelegateStakingSettlementGraceUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingSettlementGraceUpdated(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingSettlementGraceUpdated) (event.Subscription, error) {

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingSettlementGraceUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingSettlementGraceUpdated)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingSettlementGraceUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingSettlementGraceUpdated is a log parse operation binding the contract event 0x23d4c1527dfe24f1a043524d5c7c2b730f3fb2841478487080916494bc8487f2.
//
// Solidity: event DelegateStakingSettlementGraceUpdated(uint256 value)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingSettlementGraceUpdated(log types.Log) (*DelegateStakingDelegateStakingSettlementGraceUpdated, error) {
	event := new(DelegateStakingDelegateStakingSettlementGraceUpdated)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingSettlementGraceUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingWithdrawalCancelledIterator is returned from FilterDelegateStakingWithdrawalCancelled and is used to iterate over the raw logs and unpacked data for DelegateStakingWithdrawalCancelled events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingWithdrawalCancelledIterator struct {
	Event *DelegateStakingDelegateStakingWithdrawalCancelled // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingWithdrawalCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingWithdrawalCancelled)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingWithdrawalCancelled)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingWithdrawalCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingWithdrawalCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingWithdrawalCancelled represents a DelegateStakingWithdrawalCancelled event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingWithdrawalCancelled struct {
	Hot    common.Address
	Funder common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingWithdrawalCancelled is a free log retrieval operation binding the contract event 0x00041bc92ab137f1f37671a48ce9dc19f99eb0d99b0bf41beb1ffca2dd9cb557.
//
// Solidity: event DelegateStakingWithdrawalCancelled(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingWithdrawalCancelled(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingDelegateStakingWithdrawalCancelledIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingWithdrawalCancelled", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingWithdrawalCancelledIterator{contract: _DelegateStaking.contract, event: "DelegateStakingWithdrawalCancelled", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingWithdrawalCancelled is a free log subscription operation binding the contract event 0x00041bc92ab137f1f37671a48ce9dc19f99eb0d99b0bf41beb1ffca2dd9cb557.
//
// Solidity: event DelegateStakingWithdrawalCancelled(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingWithdrawalCancelled(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingWithdrawalCancelled, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingWithdrawalCancelled", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingWithdrawalCancelled)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingWithdrawalCancelled", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingWithdrawalCancelled is a log parse operation binding the contract event 0x00041bc92ab137f1f37671a48ce9dc19f99eb0d99b0bf41beb1ffca2dd9cb557.
//
// Solidity: event DelegateStakingWithdrawalCancelled(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingWithdrawalCancelled(log types.Log) (*DelegateStakingDelegateStakingWithdrawalCancelled, error) {
	event := new(DelegateStakingDelegateStakingWithdrawalCancelled)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingWithdrawalCancelled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingWithdrawalPaidIterator is returned from FilterDelegateStakingWithdrawalPaid and is used to iterate over the raw logs and unpacked data for DelegateStakingWithdrawalPaid events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingWithdrawalPaidIterator struct {
	Event *DelegateStakingDelegateStakingWithdrawalPaid // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingWithdrawalPaidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingWithdrawalPaid)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingWithdrawalPaid)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingWithdrawalPaidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingWithdrawalPaidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingWithdrawalPaid represents a DelegateStakingWithdrawalPaid event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingWithdrawalPaid struct {
	Hot    common.Address
	Funder common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingWithdrawalPaid is a free log retrieval operation binding the contract event 0x7827e4dd700b7a2a556c261a5a293c5c936068522520173931017b0e9959d935.
//
// Solidity: event DelegateStakingWithdrawalPaid(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingWithdrawalPaid(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingDelegateStakingWithdrawalPaidIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingWithdrawalPaid", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingWithdrawalPaidIterator{contract: _DelegateStaking.contract, event: "DelegateStakingWithdrawalPaid", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingWithdrawalPaid is a free log subscription operation binding the contract event 0x7827e4dd700b7a2a556c261a5a293c5c936068522520173931017b0e9959d935.
//
// Solidity: event DelegateStakingWithdrawalPaid(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingWithdrawalPaid(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingWithdrawalPaid, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingWithdrawalPaid", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingWithdrawalPaid)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingWithdrawalPaid", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingWithdrawalPaid is a log parse operation binding the contract event 0x7827e4dd700b7a2a556c261a5a293c5c936068522520173931017b0e9959d935.
//
// Solidity: event DelegateStakingWithdrawalPaid(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingWithdrawalPaid(log types.Log) (*DelegateStakingDelegateStakingWithdrawalPaid, error) {
	event := new(DelegateStakingDelegateStakingWithdrawalPaid)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingWithdrawalPaid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingDelegateStakingWithdrawalRequestedIterator is returned from FilterDelegateStakingWithdrawalRequested and is used to iterate over the raw logs and unpacked data for DelegateStakingWithdrawalRequested events raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingWithdrawalRequestedIterator struct {
	Event *DelegateStakingDelegateStakingWithdrawalRequested // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingDelegateStakingWithdrawalRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingDelegateStakingWithdrawalRequested)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingDelegateStakingWithdrawalRequested)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingDelegateStakingWithdrawalRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingDelegateStakingWithdrawalRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingDelegateStakingWithdrawalRequested represents a DelegateStakingWithdrawalRequested event raised by the DelegateStaking contract.
type DelegateStakingDelegateStakingWithdrawalRequested struct {
	Hot     common.Address
	Funder  common.Address
	Amount  *big.Int
	Paid    *big.Int
	Pending *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingWithdrawalRequested is a free log retrieval operation binding the contract event 0xab9f026078e090833bf0fd1211137db97a7a96f822c650c7997cdb7bfc859799.
//
// Solidity: event DelegateStakingWithdrawalRequested(address indexed hot, address indexed funder, uint256 amount, uint256 paid, uint256 pending)
func (_DelegateStaking *DelegateStakingFilterer) FilterDelegateStakingWithdrawalRequested(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingDelegateStakingWithdrawalRequestedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "DelegateStakingWithdrawalRequested", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingDelegateStakingWithdrawalRequestedIterator{contract: _DelegateStaking.contract, event: "DelegateStakingWithdrawalRequested", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingWithdrawalRequested is a free log subscription operation binding the contract event 0xab9f026078e090833bf0fd1211137db97a7a96f822c650c7997cdb7bfc859799.
//
// Solidity: event DelegateStakingWithdrawalRequested(address indexed hot, address indexed funder, uint256 amount, uint256 paid, uint256 pending)
func (_DelegateStaking *DelegateStakingFilterer) WatchDelegateStakingWithdrawalRequested(opts *bind.WatchOpts, sink chan<- *DelegateStakingDelegateStakingWithdrawalRequested, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "DelegateStakingWithdrawalRequested", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingDelegateStakingWithdrawalRequested)
				if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingWithdrawalRequested", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDelegateStakingWithdrawalRequested is a log parse operation binding the contract event 0xab9f026078e090833bf0fd1211137db97a7a96f822c650c7997cdb7bfc859799.
//
// Solidity: event DelegateStakingWithdrawalRequested(address indexed hot, address indexed funder, uint256 amount, uint256 paid, uint256 pending)
func (_DelegateStaking *DelegateStakingFilterer) ParseDelegateStakingWithdrawalRequested(log types.Log) (*DelegateStakingDelegateStakingWithdrawalRequested, error) {
	event := new(DelegateStakingDelegateStakingWithdrawalRequested)
	if err := _DelegateStaking.contract.UnpackLog(event, "DelegateStakingWithdrawalRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingProviderDebtPaidIterator is returned from FilterProviderDebtPaid and is used to iterate over the raw logs and unpacked data for ProviderDebtPaid events raised by the DelegateStaking contract.
type DelegateStakingProviderDebtPaidIterator struct {
	Event *DelegateStakingProviderDebtPaid // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingProviderDebtPaidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingProviderDebtPaid)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingProviderDebtPaid)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingProviderDebtPaidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingProviderDebtPaidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingProviderDebtPaid represents a ProviderDebtPaid event raised by the DelegateStaking contract.
type DelegateStakingProviderDebtPaid struct {
	Provider     common.Address
	PayoutTarget common.Address
	Amount       *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterProviderDebtPaid is a free log retrieval operation binding the contract event 0x8d79baa0ce0a07e4a85c9f1c273815f861513a2e9fdb2082bf443519534aa2b0.
//
// Solidity: event ProviderDebtPaid(address indexed provider, address indexed payoutTarget, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) FilterProviderDebtPaid(opts *bind.FilterOpts, provider []common.Address, payoutTarget []common.Address) (*DelegateStakingProviderDebtPaidIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var payoutTargetRule []interface{}
	for _, payoutTargetItem := range payoutTarget {
		payoutTargetRule = append(payoutTargetRule, payoutTargetItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "ProviderDebtPaid", providerRule, payoutTargetRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingProviderDebtPaidIterator{contract: _DelegateStaking.contract, event: "ProviderDebtPaid", logs: logs, sub: sub}, nil
}

// WatchProviderDebtPaid is a free log subscription operation binding the contract event 0x8d79baa0ce0a07e4a85c9f1c273815f861513a2e9fdb2082bf443519534aa2b0.
//
// Solidity: event ProviderDebtPaid(address indexed provider, address indexed payoutTarget, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) WatchProviderDebtPaid(opts *bind.WatchOpts, sink chan<- *DelegateStakingProviderDebtPaid, provider []common.Address, payoutTarget []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var payoutTargetRule []interface{}
	for _, payoutTargetItem := range payoutTarget {
		payoutTargetRule = append(payoutTargetRule, payoutTargetItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "ProviderDebtPaid", providerRule, payoutTargetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingProviderDebtPaid)
				if err := _DelegateStaking.contract.UnpackLog(event, "ProviderDebtPaid", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProviderDebtPaid is a log parse operation binding the contract event 0x8d79baa0ce0a07e4a85c9f1c273815f861513a2e9fdb2082bf443519534aa2b0.
//
// Solidity: event ProviderDebtPaid(address indexed provider, address indexed payoutTarget, uint256 amount)
func (_DelegateStaking *DelegateStakingFilterer) ParseProviderDebtPaid(log types.Log) (*DelegateStakingProviderDebtPaid, error) {
	event := new(DelegateStakingProviderDebtPaid)
	if err := _DelegateStaking.contract.UnpackLog(event, "ProviderDebtPaid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingProviderPayoutTargetUpdatedIterator is returned from FilterProviderPayoutTargetUpdated and is used to iterate over the raw logs and unpacked data for ProviderPayoutTargetUpdated events raised by the DelegateStaking contract.
type DelegateStakingProviderPayoutTargetUpdatedIterator struct {
	Event *DelegateStakingProviderPayoutTargetUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingProviderPayoutTargetUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingProviderPayoutTargetUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingProviderPayoutTargetUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingProviderPayoutTargetUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingProviderPayoutTargetUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingProviderPayoutTargetUpdated represents a ProviderPayoutTargetUpdated event raised by the DelegateStaking contract.
type DelegateStakingProviderPayoutTargetUpdated struct {
	Provider     common.Address
	PayoutTarget common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterProviderPayoutTargetUpdated is a free log retrieval operation binding the contract event 0x1cba50759ff5d1e813b2ed6b4f9ce6cd19137e624f4bbde798eb0572b74a6968.
//
// Solidity: event ProviderPayoutTargetUpdated(address indexed provider, address indexed payoutTarget)
func (_DelegateStaking *DelegateStakingFilterer) FilterProviderPayoutTargetUpdated(opts *bind.FilterOpts, provider []common.Address, payoutTarget []common.Address) (*DelegateStakingProviderPayoutTargetUpdatedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var payoutTargetRule []interface{}
	for _, payoutTargetItem := range payoutTarget {
		payoutTargetRule = append(payoutTargetRule, payoutTargetItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "ProviderPayoutTargetUpdated", providerRule, payoutTargetRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingProviderPayoutTargetUpdatedIterator{contract: _DelegateStaking.contract, event: "ProviderPayoutTargetUpdated", logs: logs, sub: sub}, nil
}

// WatchProviderPayoutTargetUpdated is a free log subscription operation binding the contract event 0x1cba50759ff5d1e813b2ed6b4f9ce6cd19137e624f4bbde798eb0572b74a6968.
//
// Solidity: event ProviderPayoutTargetUpdated(address indexed provider, address indexed payoutTarget)
func (_DelegateStaking *DelegateStakingFilterer) WatchProviderPayoutTargetUpdated(opts *bind.WatchOpts, sink chan<- *DelegateStakingProviderPayoutTargetUpdated, provider []common.Address, payoutTarget []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var payoutTargetRule []interface{}
	for _, payoutTargetItem := range payoutTarget {
		payoutTargetRule = append(payoutTargetRule, payoutTargetItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "ProviderPayoutTargetUpdated", providerRule, payoutTargetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingProviderPayoutTargetUpdated)
				if err := _DelegateStaking.contract.UnpackLog(event, "ProviderPayoutTargetUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProviderPayoutTargetUpdated is a log parse operation binding the contract event 0x1cba50759ff5d1e813b2ed6b4f9ce6cd19137e624f4bbde798eb0572b74a6968.
//
// Solidity: event ProviderPayoutTargetUpdated(address indexed provider, address indexed payoutTarget)
func (_DelegateStaking *DelegateStakingFilterer) ParseProviderPayoutTargetUpdated(log types.Log) (*DelegateStakingProviderPayoutTargetUpdated, error) {
	event := new(DelegateStakingProviderPayoutTargetUpdated)
	if err := _DelegateStaking.contract.UnpackLog(event, "ProviderPayoutTargetUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingProviderRewardAccruedIterator is returned from FilterProviderRewardAccrued and is used to iterate over the raw logs and unpacked data for ProviderRewardAccrued events raised by the DelegateStaking contract.
type DelegateStakingProviderRewardAccruedIterator struct {
	Event *DelegateStakingProviderRewardAccrued // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingProviderRewardAccruedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingProviderRewardAccrued)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingProviderRewardAccrued)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingProviderRewardAccruedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingProviderRewardAccruedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingProviderRewardAccrued represents a ProviderRewardAccrued event raised by the DelegateStaking contract.
type DelegateStakingProviderRewardAccrued struct {
	SessionId [32]byte
	Provider  common.Address
	Accrued   *big.Int
	Paid      *big.Int
	Owed      *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterProviderRewardAccrued is a free log retrieval operation binding the contract event 0x2e9f0aeeab6a0852b0c976f7e0d74566d7b177b6612d645afd583e6ec4c0b2a6.
//
// Solidity: event ProviderRewardAccrued(bytes32 indexed sessionId, address indexed provider, uint256 accrued, uint256 paid, uint256 owed)
func (_DelegateStaking *DelegateStakingFilterer) FilterProviderRewardAccrued(opts *bind.FilterOpts, sessionId [][32]byte, provider []common.Address) (*DelegateStakingProviderRewardAccruedIterator, error) {

	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "ProviderRewardAccrued", sessionIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingProviderRewardAccruedIterator{contract: _DelegateStaking.contract, event: "ProviderRewardAccrued", logs: logs, sub: sub}, nil
}

// WatchProviderRewardAccrued is a free log subscription operation binding the contract event 0x2e9f0aeeab6a0852b0c976f7e0d74566d7b177b6612d645afd583e6ec4c0b2a6.
//
// Solidity: event ProviderRewardAccrued(bytes32 indexed sessionId, address indexed provider, uint256 accrued, uint256 paid, uint256 owed)
func (_DelegateStaking *DelegateStakingFilterer) WatchProviderRewardAccrued(opts *bind.WatchOpts, sink chan<- *DelegateStakingProviderRewardAccrued, sessionId [][32]byte, provider []common.Address) (event.Subscription, error) {

	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "ProviderRewardAccrued", sessionIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingProviderRewardAccrued)
				if err := _DelegateStaking.contract.UnpackLog(event, "ProviderRewardAccrued", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseProviderRewardAccrued is a log parse operation binding the contract event 0x2e9f0aeeab6a0852b0c976f7e0d74566d7b177b6612d645afd583e6ec4c0b2a6.
//
// Solidity: event ProviderRewardAccrued(bytes32 indexed sessionId, address indexed provider, uint256 accrued, uint256 paid, uint256 owed)
func (_DelegateStaking *DelegateStakingFilterer) ParseProviderRewardAccrued(log types.Log) (*DelegateStakingProviderRewardAccrued, error) {
	event := new(DelegateStakingProviderRewardAccrued)
	if err := _DelegateStaking.contract.UnpackLog(event, "ProviderRewardAccrued", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingSessionClosedIterator is returned from FilterSessionClosed and is used to iterate over the raw logs and unpacked data for SessionClosed events raised by the DelegateStaking contract.
type DelegateStakingSessionClosedIterator struct {
	Event *DelegateStakingSessionClosed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingSessionClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingSessionClosed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingSessionClosed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingSessionClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingSessionClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingSessionClosed represents a SessionClosed event raised by the DelegateStaking contract.
type DelegateStakingSessionClosed struct {
	User       common.Address
	SessionId  [32]byte
	ProviderId common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSessionClosed is a free log retrieval operation binding the contract event 0x337fbb0a41a596db800dc836595a57815f967185e3596615c646f2455ac3914a.
//
// Solidity: event SessionClosed(address indexed user, bytes32 indexed sessionId, address indexed providerId)
func (_DelegateStaking *DelegateStakingFilterer) FilterSessionClosed(opts *bind.FilterOpts, user []common.Address, sessionId [][32]byte, providerId []common.Address) (*DelegateStakingSessionClosedIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "SessionClosed", userRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingSessionClosedIterator{contract: _DelegateStaking.contract, event: "SessionClosed", logs: logs, sub: sub}, nil
}

// WatchSessionClosed is a free log subscription operation binding the contract event 0x337fbb0a41a596db800dc836595a57815f967185e3596615c646f2455ac3914a.
//
// Solidity: event SessionClosed(address indexed user, bytes32 indexed sessionId, address indexed providerId)
func (_DelegateStaking *DelegateStakingFilterer) WatchSessionClosed(opts *bind.WatchOpts, sink chan<- *DelegateStakingSessionClosed, user []common.Address, sessionId [][32]byte, providerId []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "SessionClosed", userRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingSessionClosed)
				if err := _DelegateStaking.contract.UnpackLog(event, "SessionClosed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSessionClosed is a log parse operation binding the contract event 0x337fbb0a41a596db800dc836595a57815f967185e3596615c646f2455ac3914a.
//
// Solidity: event SessionClosed(address indexed user, bytes32 indexed sessionId, address indexed providerId)
func (_DelegateStaking *DelegateStakingFilterer) ParseSessionClosed(log types.Log) (*DelegateStakingSessionClosed, error) {
	event := new(DelegateStakingSessionClosed)
	if err := _DelegateStaking.contract.UnpackLog(event, "SessionClosed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingSessionClosed0Iterator is returned from FilterSessionClosed0 and is used to iterate over the raw logs and unpacked data for SessionClosed0 events raised by the DelegateStaking contract.
type DelegateStakingSessionClosed0Iterator struct {
	Event *DelegateStakingSessionClosed0 // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingSessionClosed0Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingSessionClosed0)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingSessionClosed0)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingSessionClosed0Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingSessionClosed0Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingSessionClosed0 represents a SessionClosed0 event raised by the DelegateStaking contract.
type DelegateStakingSessionClosed0 struct {
	User       common.Address
	SessionId  [32]byte
	ProviderId common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSessionClosed0 is a free log retrieval operation binding the contract event 0x337fbb0a41a596db800dc836595a57815f967185e3596615c646f2455ac3914a.
//
// Solidity: event SessionClosed(address indexed user, bytes32 indexed sessionId, address indexed providerId)
func (_DelegateStaking *DelegateStakingFilterer) FilterSessionClosed0(opts *bind.FilterOpts, user []common.Address, sessionId [][32]byte, providerId []common.Address) (*DelegateStakingSessionClosed0Iterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "SessionClosed0", userRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingSessionClosed0Iterator{contract: _DelegateStaking.contract, event: "SessionClosed0", logs: logs, sub: sub}, nil
}

// WatchSessionClosed0 is a free log subscription operation binding the contract event 0x337fbb0a41a596db800dc836595a57815f967185e3596615c646f2455ac3914a.
//
// Solidity: event SessionClosed(address indexed user, bytes32 indexed sessionId, address indexed providerId)
func (_DelegateStaking *DelegateStakingFilterer) WatchSessionClosed0(opts *bind.WatchOpts, sink chan<- *DelegateStakingSessionClosed0, user []common.Address, sessionId [][32]byte, providerId []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "SessionClosed0", userRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingSessionClosed0)
				if err := _DelegateStaking.contract.UnpackLog(event, "SessionClosed0", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSessionClosed0 is a log parse operation binding the contract event 0x337fbb0a41a596db800dc836595a57815f967185e3596615c646f2455ac3914a.
//
// Solidity: event SessionClosed(address indexed user, bytes32 indexed sessionId, address indexed providerId)
func (_DelegateStaking *DelegateStakingFilterer) ParseSessionClosed0(log types.Log) (*DelegateStakingSessionClosed0, error) {
	event := new(DelegateStakingSessionClosed0)
	if err := _DelegateStaking.contract.UnpackLog(event, "SessionClosed0", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingSessionOpenedIterator is returned from FilterSessionOpened and is used to iterate over the raw logs and unpacked data for SessionOpened events raised by the DelegateStaking contract.
type DelegateStakingSessionOpenedIterator struct {
	Event *DelegateStakingSessionOpened // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingSessionOpenedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingSessionOpened)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingSessionOpened)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingSessionOpenedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingSessionOpenedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingSessionOpened represents a SessionOpened event raised by the DelegateStaking contract.
type DelegateStakingSessionOpened struct {
	User       common.Address
	SessionId  [32]byte
	ProviderId common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSessionOpened is a free log retrieval operation binding the contract event 0x2bd7c890baf595977d256a6e784512c873ac58ba612b4895dbb7f784bfbf4839.
//
// Solidity: event SessionOpened(address indexed user, bytes32 indexed sessionId, address indexed providerId)
func (_DelegateStaking *DelegateStakingFilterer) FilterSessionOpened(opts *bind.FilterOpts, user []common.Address, sessionId [][32]byte, providerId []common.Address) (*DelegateStakingSessionOpenedIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "SessionOpened", userRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingSessionOpenedIterator{contract: _DelegateStaking.contract, event: "SessionOpened", logs: logs, sub: sub}, nil
}

// WatchSessionOpened is a free log subscription operation binding the contract event 0x2bd7c890baf595977d256a6e784512c873ac58ba612b4895dbb7f784bfbf4839.
//
// Solidity: event SessionOpened(address indexed user, bytes32 indexed sessionId, address indexed providerId)
func (_DelegateStaking *DelegateStakingFilterer) WatchSessionOpened(opts *bind.WatchOpts, sink chan<- *DelegateStakingSessionOpened, user []common.Address, sessionId [][32]byte, providerId []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}
	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "SessionOpened", userRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingSessionOpened)
				if err := _DelegateStaking.contract.UnpackLog(event, "SessionOpened", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSessionOpened is a log parse operation binding the contract event 0x2bd7c890baf595977d256a6e784512c873ac58ba612b4895dbb7f784bfbf4839.
//
// Solidity: event SessionOpened(address indexed user, bytes32 indexed sessionId, address indexed providerId)
func (_DelegateStaking *DelegateStakingFilterer) ParseSessionOpened(log types.Log) (*DelegateStakingSessionOpened, error) {
	event := new(DelegateStakingSessionOpened)
	if err := _DelegateStaking.contract.UnpackLog(event, "SessionOpened", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingSessionRatingIterator is returned from FilterSessionRating and is used to iterate over the raw logs and unpacked data for SessionRating events raised by the DelegateStaking contract.
type DelegateStakingSessionRatingIterator struct {
	Event *DelegateStakingSessionRating // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *DelegateStakingSessionRatingIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingSessionRating)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(DelegateStakingSessionRating)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *DelegateStakingSessionRatingIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingSessionRatingIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingSessionRating represents a SessionRating event raised by the DelegateStaking contract.
type DelegateStakingSessionRating struct {
	SessionId     [32]byte
	ModelId       [32]byte
	Provider      common.Address
	CloseoutType  *big.Int
	TpsScaled1000 uint32
	TtftMs        uint32
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterSessionRating is a free log retrieval operation binding the contract event 0x62fefdc7afa78415878080df4ed9b3c7dbbad510375c160ca7559f9f4d925ecb.
//
// Solidity: event SessionRating(bytes32 indexed sessionId, bytes32 indexed modelId, address indexed provider, uint256 closeoutType, uint32 tpsScaled1000, uint32 ttftMs)
func (_DelegateStaking *DelegateStakingFilterer) FilterSessionRating(opts *bind.FilterOpts, sessionId [][32]byte, modelId [][32]byte, provider []common.Address) (*DelegateStakingSessionRatingIterator, error) {

	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var modelIdRule []interface{}
	for _, modelIdItem := range modelId {
		modelIdRule = append(modelIdRule, modelIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _DelegateStaking.contract.FilterLogs(opts, "SessionRating", sessionIdRule, modelIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingSessionRatingIterator{contract: _DelegateStaking.contract, event: "SessionRating", logs: logs, sub: sub}, nil
}

// WatchSessionRating is a free log subscription operation binding the contract event 0x62fefdc7afa78415878080df4ed9b3c7dbbad510375c160ca7559f9f4d925ecb.
//
// Solidity: event SessionRating(bytes32 indexed sessionId, bytes32 indexed modelId, address indexed provider, uint256 closeoutType, uint32 tpsScaled1000, uint32 ttftMs)
func (_DelegateStaking *DelegateStakingFilterer) WatchSessionRating(opts *bind.WatchOpts, sink chan<- *DelegateStakingSessionRating, sessionId [][32]byte, modelId [][32]byte, provider []common.Address) (event.Subscription, error) {

	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var modelIdRule []interface{}
	for _, modelIdItem := range modelId {
		modelIdRule = append(modelIdRule, modelIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _DelegateStaking.contract.WatchLogs(opts, "SessionRating", sessionIdRule, modelIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingSessionRating)
				if err := _DelegateStaking.contract.UnpackLog(event, "SessionRating", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSessionRating is a log parse operation binding the contract event 0x62fefdc7afa78415878080df4ed9b3c7dbbad510375c160ca7559f9f4d925ecb.
//
// Solidity: event SessionRating(bytes32 indexed sessionId, bytes32 indexed modelId, address indexed provider, uint256 closeoutType, uint32 tpsScaled1000, uint32 ttftMs)
func (_DelegateStaking *DelegateStakingFilterer) ParseSessionRating(log types.Log) (*DelegateStakingSessionRating, error) {
	event := new(DelegateStakingSessionRating)
	if err := _DelegateStaking.contract.UnpackLog(event, "SessionRating", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
