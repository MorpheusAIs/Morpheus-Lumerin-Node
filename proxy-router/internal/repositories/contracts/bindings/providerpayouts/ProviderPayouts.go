// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package providerpayouts

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

// ProviderPayoutsMetaData contains all meta data concerning the ProviderPayouts contract.
var ProviderPayoutsMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"DelegateStakingActiveFundersLimitExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingAllowanceNotActive\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingCapBelowLifetimeFunding\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingFundingCapExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInsufficientCapacity\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInsufficientLiquidity\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInsufficientPrincipal\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidExpiry\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidPoolSettlement\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidWithdrawalCursor\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingLimitExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPoolFunderNotApproved\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPoolSessionAlreadyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPoolSessionNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPrincipalTooLow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingSessionSlotNotEmpty\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingSettlementNotReady\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingWithdrawalNotFound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"delegatee\",\"type\":\"address\"}],\"name\":\"InsufficientRightsForOperation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account_\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProviderPayoutTargetInvalid\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingAllowanceFunded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"cumulativeFundingCap\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"expiry\",\"type\":\"uint128\"}],\"name\":\"DelegateStakingAllowanceGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"}],\"name\":\"DelegateStakingAllowanceRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMaxActiveFundersUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMaxAutoReleaseDaysUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMaxAutoServiceUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMinPrincipalUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolDrawn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"}],\"name\":\"DelegateStakingPoolFunderApprovalRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"}],\"name\":\"DelegateStakingPoolFunderApproved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint128\",\"name\":\"releaseAt\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolLockCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint128\",\"name\":\"releaseAt\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolLockReleased\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolReturned\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"DelegateStakingPoolSessionSettled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingSettlementGraceUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingWithdrawalCancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingWithdrawalPaid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"paid\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"pending\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingWithdrawalRequested\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"storageSlot\",\"type\":\"bytes32\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"payoutTarget\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ProviderDebtPaid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"payoutTarget\",\"type\":\"address\"}],\"name\":\"ProviderPayoutTargetUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"accrued\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"paid\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"owed\",\"type\":\"uint256\"}],\"name\":\"ProviderRewardAccrued\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"tpsScaled1000\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"ttftMs\",\"type\":\"uint32\"}],\"name\":\"SessionRating\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BIDS_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"COMPUTE_POOL_INDEX\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_MARKETPLACE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_MODEL\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_PAYOUT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_PROVIDER\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_SESSION\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DIAMOND_OWNABLE_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_SESSION_DURATION\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PROVIDERS_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SESSIONS_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SIGNATURE_TTL\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"STATS_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"__ProviderPayouts_init\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getActiveProviders\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId_\",\"type\":\"bytes32\"}],\"name\":\"getBid\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"deletedAt\",\"type\":\"uint128\"}],\"internalType\":\"structIBidStorage.Bid\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"timestamp_\",\"type\":\"uint128\"}],\"name\":\"getComputeBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getFundingAccount\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getIsProviderActive\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"approval_\",\"type\":\"bytes\"}],\"name\":\"getIsProviderApprovalUsed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMaxSessionDuration\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getModelActiveBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getModelBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getModelSessions\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"}],\"name\":\"getModelStats\",\"outputs\":[{\"components\":[{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"tpsScaled1000\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"ttftMs\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"totalDuration\",\"type\":\"tuple\"},{\"internalType\":\"uint32\",\"name\":\"count\",\"type\":\"uint32\"}],\"internalType\":\"structIStatsStorage.ModelStats\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getPayoutTarget\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPendingProviderClaims\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index_\",\"type\":\"uint256\"}],\"name\":\"getPool\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"initialReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardDecrease\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"payoutStart\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"decreaseInterval\",\"type\":\"uint128\"}],\"internalType\":\"structISessionStorage.Pool\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPools\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"initialReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardDecrease\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"payoutStart\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"decreaseInterval\",\"type\":\"uint128\"}],\"internalType\":\"structISessionStorage.Pool[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getProvider\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"endpoint\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"limitPeriodEnd\",\"type\":\"uint128\"},{\"internalType\":\"uint256\",\"name\":\"limitPeriodEarned\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"internalType\":\"structIProviderStorage.Provider\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getProviderActiveBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getProviderBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getProviderLifetimeClaimed\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getProviderMinimumStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getProviderModelStats\",\"outputs\":[{\"components\":[{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"tpsScaled1000\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"ttftMs\",\"type\":\"tuple\"},{\"internalType\":\"uint32\",\"name\":\"totalDuration\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"successCount\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"totalCount\",\"type\":\"uint32\"}],\"internalType\":\"structIStatsStorage.ProviderModelStats\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getProviderOwed\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getProviderSessions\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getProvidersTotalClaimed\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRegistry\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId_\",\"type\":\"bytes32\"}],\"name\":\"getSession\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"closeoutReceipt\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerWithdrawnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"openedAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"endsAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"closedAt\",\"type\":\"uint128\"},{\"internalType\":\"bool\",\"name\":\"isActive\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"isDirectPaymentFromUser\",\"type\":\"bool\"}],\"internalType\":\"structISessionStorage.Session\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond_\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"openedAt_\",\"type\":\"uint128\"}],\"name\":\"getSessionEnd\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"bidId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"sessionNonce_\",\"type\":\"uint256\"}],\"name\":\"getSessionId\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"timestamp_\",\"type\":\"uint128\"}],\"name\":\"getTodaysBudget\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getTotalSessions\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getUserSessions\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user_\",\"type\":\"address\"},{\"internalType\":\"uint8\",\"name\":\"iterations_\",\"type\":\"uint8\"}],\"name\":\"getUserStakesOnHold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"available_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"hold_\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId_\",\"type\":\"bytes32\"}],\"name\":\"isBidActive\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegatee_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"delegator_\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"rights_\",\"type\":\"bytes32\"}],\"name\":\"isRightsDelegated\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"target_\",\"type\":\"address\"}],\"name\":\"setPayoutTarget\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"timestamp_\",\"type\":\"uint128\"}],\"name\":\"stakeToStipend\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"timestamp_\",\"type\":\"uint128\"}],\"name\":\"startOfTheDay\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"stipend_\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"timestamp_\",\"type\":\"uint128\"}],\"name\":\"stipendToStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"timestamp_\",\"type\":\"uint128\"}],\"name\":\"totalMORSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ProviderPayoutsABI is the input ABI used to generate the binding from.
// Deprecated: Use ProviderPayoutsMetaData.ABI instead.
var ProviderPayoutsABI = ProviderPayoutsMetaData.ABI

// ProviderPayouts is an auto generated Go binding around an Ethereum contract.
type ProviderPayouts struct {
	ProviderPayoutsCaller     // Read-only binding to the contract
	ProviderPayoutsTransactor // Write-only binding to the contract
	ProviderPayoutsFilterer   // Log filterer for contract events
}

// ProviderPayoutsCaller is an auto generated read-only Go binding around an Ethereum contract.
type ProviderPayoutsCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProviderPayoutsTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ProviderPayoutsTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProviderPayoutsFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ProviderPayoutsFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProviderPayoutsSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ProviderPayoutsSession struct {
	Contract     *ProviderPayouts  // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ProviderPayoutsCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ProviderPayoutsCallerSession struct {
	Contract *ProviderPayoutsCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts          // Call options to use throughout this session
}

// ProviderPayoutsTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ProviderPayoutsTransactorSession struct {
	Contract     *ProviderPayoutsTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts          // Transaction auth options to use throughout this session
}

// ProviderPayoutsRaw is an auto generated low-level Go binding around an Ethereum contract.
type ProviderPayoutsRaw struct {
	Contract *ProviderPayouts // Generic contract binding to access the raw methods on
}

// ProviderPayoutsCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ProviderPayoutsCallerRaw struct {
	Contract *ProviderPayoutsCaller // Generic read-only contract binding to access the raw methods on
}

// ProviderPayoutsTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ProviderPayoutsTransactorRaw struct {
	Contract *ProviderPayoutsTransactor // Generic write-only contract binding to access the raw methods on
}

// NewProviderPayouts creates a new instance of ProviderPayouts, bound to a specific deployed contract.
func NewProviderPayouts(address common.Address, backend bind.ContractBackend) (*ProviderPayouts, error) {
	contract, err := bindProviderPayouts(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ProviderPayouts{ProviderPayoutsCaller: ProviderPayoutsCaller{contract: contract}, ProviderPayoutsTransactor: ProviderPayoutsTransactor{contract: contract}, ProviderPayoutsFilterer: ProviderPayoutsFilterer{contract: contract}}, nil
}

// NewProviderPayoutsCaller creates a new read-only instance of ProviderPayouts, bound to a specific deployed contract.
func NewProviderPayoutsCaller(address common.Address, caller bind.ContractCaller) (*ProviderPayoutsCaller, error) {
	contract, err := bindProviderPayouts(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsCaller{contract: contract}, nil
}

// NewProviderPayoutsTransactor creates a new write-only instance of ProviderPayouts, bound to a specific deployed contract.
func NewProviderPayoutsTransactor(address common.Address, transactor bind.ContractTransactor) (*ProviderPayoutsTransactor, error) {
	contract, err := bindProviderPayouts(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsTransactor{contract: contract}, nil
}

// NewProviderPayoutsFilterer creates a new log filterer instance of ProviderPayouts, bound to a specific deployed contract.
func NewProviderPayoutsFilterer(address common.Address, filterer bind.ContractFilterer) (*ProviderPayoutsFilterer, error) {
	contract, err := bindProviderPayouts(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsFilterer{contract: contract}, nil
}

// bindProviderPayouts binds a generic wrapper to an already deployed contract.
func bindProviderPayouts(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ProviderPayoutsMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ProviderPayouts *ProviderPayoutsRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ProviderPayouts.Contract.ProviderPayoutsCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ProviderPayouts *ProviderPayoutsRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProviderPayouts.Contract.ProviderPayoutsTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ProviderPayouts *ProviderPayoutsRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ProviderPayouts.Contract.ProviderPayoutsTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ProviderPayouts *ProviderPayoutsCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ProviderPayouts.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ProviderPayouts *ProviderPayoutsTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProviderPayouts.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ProviderPayouts *ProviderPayoutsTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ProviderPayouts.Contract.contract.Transact(opts, method, params...)
}

// BIDSSTORAGESLOT is a free data retrieval call binding the contract method 0x266ccff0.
//
// Solidity: function BIDS_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCaller) BIDSSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "BIDS_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BIDSSTORAGESLOT is a free data retrieval call binding the contract method 0x266ccff0.
//
// Solidity: function BIDS_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsSession) BIDSSTORAGESLOT() ([32]byte, error) {
	return _ProviderPayouts.Contract.BIDSSTORAGESLOT(&_ProviderPayouts.CallOpts)
}

// BIDSSTORAGESLOT is a free data retrieval call binding the contract method 0x266ccff0.
//
// Solidity: function BIDS_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCallerSession) BIDSSTORAGESLOT() ([32]byte, error) {
	return _ProviderPayouts.Contract.BIDSSTORAGESLOT(&_ProviderPayouts.CallOpts)
}

// COMPUTEPOOLINDEX is a free data retrieval call binding the contract method 0xc56d09a0.
//
// Solidity: function COMPUTE_POOL_INDEX() view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) COMPUTEPOOLINDEX(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "COMPUTE_POOL_INDEX")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// COMPUTEPOOLINDEX is a free data retrieval call binding the contract method 0xc56d09a0.
//
// Solidity: function COMPUTE_POOL_INDEX() view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsSession) COMPUTEPOOLINDEX() (*big.Int, error) {
	return _ProviderPayouts.Contract.COMPUTEPOOLINDEX(&_ProviderPayouts.CallOpts)
}

// COMPUTEPOOLINDEX is a free data retrieval call binding the contract method 0xc56d09a0.
//
// Solidity: function COMPUTE_POOL_INDEX() view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) COMPUTEPOOLINDEX() (*big.Int, error) {
	return _ProviderPayouts.Contract.COMPUTEPOOLINDEX(&_ProviderPayouts.CallOpts)
}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCaller) DELEGATIONRULESMARKETPLACE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "DELEGATION_RULES_MARKETPLACE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsSession) DELEGATIONRULESMARKETPLACE() ([32]byte, error) {
	return _ProviderPayouts.Contract.DELEGATIONRULESMARKETPLACE(&_ProviderPayouts.CallOpts)
}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCallerSession) DELEGATIONRULESMARKETPLACE() ([32]byte, error) {
	return _ProviderPayouts.Contract.DELEGATIONRULESMARKETPLACE(&_ProviderPayouts.CallOpts)
}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCaller) DELEGATIONRULESMODEL(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "DELEGATION_RULES_MODEL")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsSession) DELEGATIONRULESMODEL() ([32]byte, error) {
	return _ProviderPayouts.Contract.DELEGATIONRULESMODEL(&_ProviderPayouts.CallOpts)
}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCallerSession) DELEGATIONRULESMODEL() ([32]byte, error) {
	return _ProviderPayouts.Contract.DELEGATIONRULESMODEL(&_ProviderPayouts.CallOpts)
}

// DELEGATIONRULESPAYOUT is a free data retrieval call binding the contract method 0xb55118e0.
//
// Solidity: function DELEGATION_RULES_PAYOUT() pure returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCaller) DELEGATIONRULESPAYOUT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "DELEGATION_RULES_PAYOUT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESPAYOUT is a free data retrieval call binding the contract method 0xb55118e0.
//
// Solidity: function DELEGATION_RULES_PAYOUT() pure returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsSession) DELEGATIONRULESPAYOUT() ([32]byte, error) {
	return _ProviderPayouts.Contract.DELEGATIONRULESPAYOUT(&_ProviderPayouts.CallOpts)
}

// DELEGATIONRULESPAYOUT is a free data retrieval call binding the contract method 0xb55118e0.
//
// Solidity: function DELEGATION_RULES_PAYOUT() pure returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCallerSession) DELEGATIONRULESPAYOUT() ([32]byte, error) {
	return _ProviderPayouts.Contract.DELEGATIONRULESPAYOUT(&_ProviderPayouts.CallOpts)
}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCaller) DELEGATIONRULESPROVIDER(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "DELEGATION_RULES_PROVIDER")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsSession) DELEGATIONRULESPROVIDER() ([32]byte, error) {
	return _ProviderPayouts.Contract.DELEGATIONRULESPROVIDER(&_ProviderPayouts.CallOpts)
}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCallerSession) DELEGATIONRULESPROVIDER() ([32]byte, error) {
	return _ProviderPayouts.Contract.DELEGATIONRULESPROVIDER(&_ProviderPayouts.CallOpts)
}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCaller) DELEGATIONRULESSESSION(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "DELEGATION_RULES_SESSION")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsSession) DELEGATIONRULESSESSION() ([32]byte, error) {
	return _ProviderPayouts.Contract.DELEGATIONRULESSESSION(&_ProviderPayouts.CallOpts)
}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCallerSession) DELEGATIONRULESSESSION() ([32]byte, error) {
	return _ProviderPayouts.Contract.DELEGATIONRULESSESSION(&_ProviderPayouts.CallOpts)
}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCaller) DELEGATIONSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "DELEGATION_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsSession) DELEGATIONSTORAGESLOT() ([32]byte, error) {
	return _ProviderPayouts.Contract.DELEGATIONSTORAGESLOT(&_ProviderPayouts.CallOpts)
}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCallerSession) DELEGATIONSTORAGESLOT() ([32]byte, error) {
	return _ProviderPayouts.Contract.DELEGATIONSTORAGESLOT(&_ProviderPayouts.CallOpts)
}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCaller) DIAMONDOWNABLESTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "DIAMOND_OWNABLE_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsSession) DIAMONDOWNABLESTORAGESLOT() ([32]byte, error) {
	return _ProviderPayouts.Contract.DIAMONDOWNABLESTORAGESLOT(&_ProviderPayouts.CallOpts)
}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCallerSession) DIAMONDOWNABLESTORAGESLOT() ([32]byte, error) {
	return _ProviderPayouts.Contract.DIAMONDOWNABLESTORAGESLOT(&_ProviderPayouts.CallOpts)
}

// MINSESSIONDURATION is a free data retrieval call binding the contract method 0x7d980286.
//
// Solidity: function MIN_SESSION_DURATION() view returns(uint32)
func (_ProviderPayouts *ProviderPayoutsCaller) MINSESSIONDURATION(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "MIN_SESSION_DURATION")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// MINSESSIONDURATION is a free data retrieval call binding the contract method 0x7d980286.
//
// Solidity: function MIN_SESSION_DURATION() view returns(uint32)
func (_ProviderPayouts *ProviderPayoutsSession) MINSESSIONDURATION() (uint32, error) {
	return _ProviderPayouts.Contract.MINSESSIONDURATION(&_ProviderPayouts.CallOpts)
}

// MINSESSIONDURATION is a free data retrieval call binding the contract method 0x7d980286.
//
// Solidity: function MIN_SESSION_DURATION() view returns(uint32)
func (_ProviderPayouts *ProviderPayoutsCallerSession) MINSESSIONDURATION() (uint32, error) {
	return _ProviderPayouts.Contract.MINSESSIONDURATION(&_ProviderPayouts.CallOpts)
}

// PROVIDERSSTORAGESLOT is a free data retrieval call binding the contract method 0xc51830f6.
//
// Solidity: function PROVIDERS_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCaller) PROVIDERSSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "PROVIDERS_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PROVIDERSSTORAGESLOT is a free data retrieval call binding the contract method 0xc51830f6.
//
// Solidity: function PROVIDERS_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsSession) PROVIDERSSTORAGESLOT() ([32]byte, error) {
	return _ProviderPayouts.Contract.PROVIDERSSTORAGESLOT(&_ProviderPayouts.CallOpts)
}

// PROVIDERSSTORAGESLOT is a free data retrieval call binding the contract method 0xc51830f6.
//
// Solidity: function PROVIDERS_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCallerSession) PROVIDERSSTORAGESLOT() ([32]byte, error) {
	return _ProviderPayouts.Contract.PROVIDERSSTORAGESLOT(&_ProviderPayouts.CallOpts)
}

// SESSIONSSTORAGESLOT is a free data retrieval call binding the contract method 0xb392636e.
//
// Solidity: function SESSIONS_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCaller) SESSIONSSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "SESSIONS_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SESSIONSSTORAGESLOT is a free data retrieval call binding the contract method 0xb392636e.
//
// Solidity: function SESSIONS_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsSession) SESSIONSSTORAGESLOT() ([32]byte, error) {
	return _ProviderPayouts.Contract.SESSIONSSTORAGESLOT(&_ProviderPayouts.CallOpts)
}

// SESSIONSSTORAGESLOT is a free data retrieval call binding the contract method 0xb392636e.
//
// Solidity: function SESSIONS_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCallerSession) SESSIONSSTORAGESLOT() ([32]byte, error) {
	return _ProviderPayouts.Contract.SESSIONSSTORAGESLOT(&_ProviderPayouts.CallOpts)
}

// SIGNATURETTL is a free data retrieval call binding the contract method 0xe7d791d0.
//
// Solidity: function SIGNATURE_TTL() view returns(uint32)
func (_ProviderPayouts *ProviderPayoutsCaller) SIGNATURETTL(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "SIGNATURE_TTL")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// SIGNATURETTL is a free data retrieval call binding the contract method 0xe7d791d0.
//
// Solidity: function SIGNATURE_TTL() view returns(uint32)
func (_ProviderPayouts *ProviderPayoutsSession) SIGNATURETTL() (uint32, error) {
	return _ProviderPayouts.Contract.SIGNATURETTL(&_ProviderPayouts.CallOpts)
}

// SIGNATURETTL is a free data retrieval call binding the contract method 0xe7d791d0.
//
// Solidity: function SIGNATURE_TTL() view returns(uint32)
func (_ProviderPayouts *ProviderPayoutsCallerSession) SIGNATURETTL() (uint32, error) {
	return _ProviderPayouts.Contract.SIGNATURETTL(&_ProviderPayouts.CallOpts)
}

// STATSSTORAGESLOT is a free data retrieval call binding the contract method 0x87040789.
//
// Solidity: function STATS_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCaller) STATSSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "STATS_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// STATSSTORAGESLOT is a free data retrieval call binding the contract method 0x87040789.
//
// Solidity: function STATS_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsSession) STATSSTORAGESLOT() ([32]byte, error) {
	return _ProviderPayouts.Contract.STATSSTORAGESLOT(&_ProviderPayouts.CallOpts)
}

// STATSSTORAGESLOT is a free data retrieval call binding the contract method 0x87040789.
//
// Solidity: function STATS_STORAGE_SLOT() view returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCallerSession) STATSSTORAGESLOT() ([32]byte, error) {
	return _ProviderPayouts.Contract.STATSSTORAGESLOT(&_ProviderPayouts.CallOpts)
}

// GetActiveProviders is a free data retrieval call binding the contract method 0xd5472642.
//
// Solidity: function getActiveProviders(uint256 offset_, uint256 limit_) view returns(address[], uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetActiveProviders(opts *bind.CallOpts, offset_ *big.Int, limit_ *big.Int) ([]common.Address, *big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getActiveProviders", offset_, limit_)

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
func (_ProviderPayouts *ProviderPayoutsSession) GetActiveProviders(offset_ *big.Int, limit_ *big.Int) ([]common.Address, *big.Int, error) {
	return _ProviderPayouts.Contract.GetActiveProviders(&_ProviderPayouts.CallOpts, offset_, limit_)
}

// GetActiveProviders is a free data retrieval call binding the contract method 0xd5472642.
//
// Solidity: function getActiveProviders(uint256 offset_, uint256 limit_) view returns(address[], uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetActiveProviders(offset_ *big.Int, limit_ *big.Int) ([]common.Address, *big.Int, error) {
	return _ProviderPayouts.Contract.GetActiveProviders(&_ProviderPayouts.CallOpts, offset_, limit_)
}

// GetBid is a free data retrieval call binding the contract method 0x91704e1e.
//
// Solidity: function getBid(bytes32 bidId_) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_ProviderPayouts *ProviderPayoutsCaller) GetBid(opts *bind.CallOpts, bidId_ [32]byte) (IBidStorageBid, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getBid", bidId_)

	if err != nil {
		return *new(IBidStorageBid), err
	}

	out0 := *abi.ConvertType(out[0], new(IBidStorageBid)).(*IBidStorageBid)

	return out0, err

}

// GetBid is a free data retrieval call binding the contract method 0x91704e1e.
//
// Solidity: function getBid(bytes32 bidId_) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_ProviderPayouts *ProviderPayoutsSession) GetBid(bidId_ [32]byte) (IBidStorageBid, error) {
	return _ProviderPayouts.Contract.GetBid(&_ProviderPayouts.CallOpts, bidId_)
}

// GetBid is a free data retrieval call binding the contract method 0x91704e1e.
//
// Solidity: function getBid(bytes32 bidId_) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetBid(bidId_ [32]byte) (IBidStorageBid, error) {
	return _ProviderPayouts.Contract.GetBid(&_ProviderPayouts.CallOpts, bidId_)
}

// GetComputeBalance is a free data retrieval call binding the contract method 0x61ce471a.
//
// Solidity: function getComputeBalance(uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetComputeBalance(opts *bind.CallOpts, timestamp_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getComputeBalance", timestamp_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetComputeBalance is a free data retrieval call binding the contract method 0x61ce471a.
//
// Solidity: function getComputeBalance(uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsSession) GetComputeBalance(timestamp_ *big.Int) (*big.Int, error) {
	return _ProviderPayouts.Contract.GetComputeBalance(&_ProviderPayouts.CallOpts, timestamp_)
}

// GetComputeBalance is a free data retrieval call binding the contract method 0x61ce471a.
//
// Solidity: function getComputeBalance(uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetComputeBalance(timestamp_ *big.Int) (*big.Int, error) {
	return _ProviderPayouts.Contract.GetComputeBalance(&_ProviderPayouts.CallOpts, timestamp_)
}

// GetFundingAccount is a free data retrieval call binding the contract method 0x775c3727.
//
// Solidity: function getFundingAccount() view returns(address)
func (_ProviderPayouts *ProviderPayoutsCaller) GetFundingAccount(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getFundingAccount")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetFundingAccount is a free data retrieval call binding the contract method 0x775c3727.
//
// Solidity: function getFundingAccount() view returns(address)
func (_ProviderPayouts *ProviderPayoutsSession) GetFundingAccount() (common.Address, error) {
	return _ProviderPayouts.Contract.GetFundingAccount(&_ProviderPayouts.CallOpts)
}

// GetFundingAccount is a free data retrieval call binding the contract method 0x775c3727.
//
// Solidity: function getFundingAccount() view returns(address)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetFundingAccount() (common.Address, error) {
	return _ProviderPayouts.Contract.GetFundingAccount(&_ProviderPayouts.CallOpts)
}

// GetIsProviderActive is a free data retrieval call binding the contract method 0x63ef175d.
//
// Solidity: function getIsProviderActive(address provider_) view returns(bool)
func (_ProviderPayouts *ProviderPayoutsCaller) GetIsProviderActive(opts *bind.CallOpts, provider_ common.Address) (bool, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getIsProviderActive", provider_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// GetIsProviderActive is a free data retrieval call binding the contract method 0x63ef175d.
//
// Solidity: function getIsProviderActive(address provider_) view returns(bool)
func (_ProviderPayouts *ProviderPayoutsSession) GetIsProviderActive(provider_ common.Address) (bool, error) {
	return _ProviderPayouts.Contract.GetIsProviderActive(&_ProviderPayouts.CallOpts, provider_)
}

// GetIsProviderActive is a free data retrieval call binding the contract method 0x63ef175d.
//
// Solidity: function getIsProviderActive(address provider_) view returns(bool)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetIsProviderActive(provider_ common.Address) (bool, error) {
	return _ProviderPayouts.Contract.GetIsProviderActive(&_ProviderPayouts.CallOpts, provider_)
}

// GetIsProviderApprovalUsed is a free data retrieval call binding the contract method 0xdb1cf1e2.
//
// Solidity: function getIsProviderApprovalUsed(bytes approval_) view returns(bool)
func (_ProviderPayouts *ProviderPayoutsCaller) GetIsProviderApprovalUsed(opts *bind.CallOpts, approval_ []byte) (bool, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getIsProviderApprovalUsed", approval_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// GetIsProviderApprovalUsed is a free data retrieval call binding the contract method 0xdb1cf1e2.
//
// Solidity: function getIsProviderApprovalUsed(bytes approval_) view returns(bool)
func (_ProviderPayouts *ProviderPayoutsSession) GetIsProviderApprovalUsed(approval_ []byte) (bool, error) {
	return _ProviderPayouts.Contract.GetIsProviderApprovalUsed(&_ProviderPayouts.CallOpts, approval_)
}

// GetIsProviderApprovalUsed is a free data retrieval call binding the contract method 0xdb1cf1e2.
//
// Solidity: function getIsProviderApprovalUsed(bytes approval_) view returns(bool)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetIsProviderApprovalUsed(approval_ []byte) (bool, error) {
	return _ProviderPayouts.Contract.GetIsProviderApprovalUsed(&_ProviderPayouts.CallOpts, approval_)
}

// GetMaxSessionDuration is a free data retrieval call binding the contract method 0xa9756858.
//
// Solidity: function getMaxSessionDuration() view returns(uint128)
func (_ProviderPayouts *ProviderPayoutsCaller) GetMaxSessionDuration(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getMaxSessionDuration")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMaxSessionDuration is a free data retrieval call binding the contract method 0xa9756858.
//
// Solidity: function getMaxSessionDuration() view returns(uint128)
func (_ProviderPayouts *ProviderPayoutsSession) GetMaxSessionDuration() (*big.Int, error) {
	return _ProviderPayouts.Contract.GetMaxSessionDuration(&_ProviderPayouts.CallOpts)
}

// GetMaxSessionDuration is a free data retrieval call binding the contract method 0xa9756858.
//
// Solidity: function getMaxSessionDuration() view returns(uint128)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetMaxSessionDuration() (*big.Int, error) {
	return _ProviderPayouts.Contract.GetMaxSessionDuration(&_ProviderPayouts.CallOpts)
}

// GetModelActiveBids is a free data retrieval call binding the contract method 0x8a683b6e.
//
// Solidity: function getModelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetModelActiveBids(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getModelActiveBids", modelId_, offset_, limit_)

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
func (_ProviderPayouts *ProviderPayoutsSession) GetModelActiveBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _ProviderPayouts.Contract.GetModelActiveBids(&_ProviderPayouts.CallOpts, modelId_, offset_, limit_)
}

// GetModelActiveBids is a free data retrieval call binding the contract method 0x8a683b6e.
//
// Solidity: function getModelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetModelActiveBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _ProviderPayouts.Contract.GetModelActiveBids(&_ProviderPayouts.CallOpts, modelId_, offset_, limit_)
}

// GetModelBids is a free data retrieval call binding the contract method 0xfade17b1.
//
// Solidity: function getModelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetModelBids(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getModelBids", modelId_, offset_, limit_)

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
func (_ProviderPayouts *ProviderPayoutsSession) GetModelBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _ProviderPayouts.Contract.GetModelBids(&_ProviderPayouts.CallOpts, modelId_, offset_, limit_)
}

// GetModelBids is a free data retrieval call binding the contract method 0xfade17b1.
//
// Solidity: function getModelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetModelBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _ProviderPayouts.Contract.GetModelBids(&_ProviderPayouts.CallOpts, modelId_, offset_, limit_)
}

// GetModelSessions is a free data retrieval call binding the contract method 0x1d78a872.
//
// Solidity: function getModelSessions(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetModelSessions(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getModelSessions", modelId_, offset_, limit_)

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
func (_ProviderPayouts *ProviderPayoutsSession) GetModelSessions(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _ProviderPayouts.Contract.GetModelSessions(&_ProviderPayouts.CallOpts, modelId_, offset_, limit_)
}

// GetModelSessions is a free data retrieval call binding the contract method 0x1d78a872.
//
// Solidity: function getModelSessions(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetModelSessions(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _ProviderPayouts.Contract.GetModelSessions(&_ProviderPayouts.CallOpts, modelId_, offset_, limit_)
}

// GetModelStats is a free data retrieval call binding the contract method 0xce535723.
//
// Solidity: function getModelStats(bytes32 modelId_) view returns(((int64,int64),(int64,int64),(int64,int64),uint32))
func (_ProviderPayouts *ProviderPayoutsCaller) GetModelStats(opts *bind.CallOpts, modelId_ [32]byte) (IStatsStorageModelStats, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getModelStats", modelId_)

	if err != nil {
		return *new(IStatsStorageModelStats), err
	}

	out0 := *abi.ConvertType(out[0], new(IStatsStorageModelStats)).(*IStatsStorageModelStats)

	return out0, err

}

// GetModelStats is a free data retrieval call binding the contract method 0xce535723.
//
// Solidity: function getModelStats(bytes32 modelId_) view returns(((int64,int64),(int64,int64),(int64,int64),uint32))
func (_ProviderPayouts *ProviderPayoutsSession) GetModelStats(modelId_ [32]byte) (IStatsStorageModelStats, error) {
	return _ProviderPayouts.Contract.GetModelStats(&_ProviderPayouts.CallOpts, modelId_)
}

// GetModelStats is a free data retrieval call binding the contract method 0xce535723.
//
// Solidity: function getModelStats(bytes32 modelId_) view returns(((int64,int64),(int64,int64),(int64,int64),uint32))
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetModelStats(modelId_ [32]byte) (IStatsStorageModelStats, error) {
	return _ProviderPayouts.Contract.GetModelStats(&_ProviderPayouts.CallOpts, modelId_)
}

// GetPayoutTarget is a free data retrieval call binding the contract method 0xfb0c3c1a.
//
// Solidity: function getPayoutTarget(address provider_) view returns(address)
func (_ProviderPayouts *ProviderPayoutsCaller) GetPayoutTarget(opts *bind.CallOpts, provider_ common.Address) (common.Address, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getPayoutTarget", provider_)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetPayoutTarget is a free data retrieval call binding the contract method 0xfb0c3c1a.
//
// Solidity: function getPayoutTarget(address provider_) view returns(address)
func (_ProviderPayouts *ProviderPayoutsSession) GetPayoutTarget(provider_ common.Address) (common.Address, error) {
	return _ProviderPayouts.Contract.GetPayoutTarget(&_ProviderPayouts.CallOpts, provider_)
}

// GetPayoutTarget is a free data retrieval call binding the contract method 0xfb0c3c1a.
//
// Solidity: function getPayoutTarget(address provider_) view returns(address)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetPayoutTarget(provider_ common.Address) (common.Address, error) {
	return _ProviderPayouts.Contract.GetPayoutTarget(&_ProviderPayouts.CallOpts, provider_)
}

// GetPendingProviderClaims is a free data retrieval call binding the contract method 0x92393d3b.
//
// Solidity: function getPendingProviderClaims() view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetPendingProviderClaims(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getPendingProviderClaims")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetPendingProviderClaims is a free data retrieval call binding the contract method 0x92393d3b.
//
// Solidity: function getPendingProviderClaims() view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsSession) GetPendingProviderClaims() (*big.Int, error) {
	return _ProviderPayouts.Contract.GetPendingProviderClaims(&_ProviderPayouts.CallOpts)
}

// GetPendingProviderClaims is a free data retrieval call binding the contract method 0x92393d3b.
//
// Solidity: function getPendingProviderClaims() view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetPendingProviderClaims() (*big.Int, error) {
	return _ProviderPayouts.Contract.GetPendingProviderClaims(&_ProviderPayouts.CallOpts)
}

// GetPool is a free data retrieval call binding the contract method 0x068bcd8d.
//
// Solidity: function getPool(uint256 index_) view returns((uint256,uint256,uint128,uint128))
func (_ProviderPayouts *ProviderPayoutsCaller) GetPool(opts *bind.CallOpts, index_ *big.Int) (ISessionStoragePool, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getPool", index_)

	if err != nil {
		return *new(ISessionStoragePool), err
	}

	out0 := *abi.ConvertType(out[0], new(ISessionStoragePool)).(*ISessionStoragePool)

	return out0, err

}

// GetPool is a free data retrieval call binding the contract method 0x068bcd8d.
//
// Solidity: function getPool(uint256 index_) view returns((uint256,uint256,uint128,uint128))
func (_ProviderPayouts *ProviderPayoutsSession) GetPool(index_ *big.Int) (ISessionStoragePool, error) {
	return _ProviderPayouts.Contract.GetPool(&_ProviderPayouts.CallOpts, index_)
}

// GetPool is a free data retrieval call binding the contract method 0x068bcd8d.
//
// Solidity: function getPool(uint256 index_) view returns((uint256,uint256,uint128,uint128))
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetPool(index_ *big.Int) (ISessionStoragePool, error) {
	return _ProviderPayouts.Contract.GetPool(&_ProviderPayouts.CallOpts, index_)
}

// GetPools is a free data retrieval call binding the contract method 0x673a2a1f.
//
// Solidity: function getPools() view returns((uint256,uint256,uint128,uint128)[])
func (_ProviderPayouts *ProviderPayoutsCaller) GetPools(opts *bind.CallOpts) ([]ISessionStoragePool, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getPools")

	if err != nil {
		return *new([]ISessionStoragePool), err
	}

	out0 := *abi.ConvertType(out[0], new([]ISessionStoragePool)).(*[]ISessionStoragePool)

	return out0, err

}

// GetPools is a free data retrieval call binding the contract method 0x673a2a1f.
//
// Solidity: function getPools() view returns((uint256,uint256,uint128,uint128)[])
func (_ProviderPayouts *ProviderPayoutsSession) GetPools() ([]ISessionStoragePool, error) {
	return _ProviderPayouts.Contract.GetPools(&_ProviderPayouts.CallOpts)
}

// GetPools is a free data retrieval call binding the contract method 0x673a2a1f.
//
// Solidity: function getPools() view returns((uint256,uint256,uint128,uint128)[])
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetPools() ([]ISessionStoragePool, error) {
	return _ProviderPayouts.Contract.GetPools(&_ProviderPayouts.CallOpts)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider_) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_ProviderPayouts *ProviderPayoutsCaller) GetProvider(opts *bind.CallOpts, provider_ common.Address) (IProviderStorageProvider, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getProvider", provider_)

	if err != nil {
		return *new(IProviderStorageProvider), err
	}

	out0 := *abi.ConvertType(out[0], new(IProviderStorageProvider)).(*IProviderStorageProvider)

	return out0, err

}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider_) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_ProviderPayouts *ProviderPayoutsSession) GetProvider(provider_ common.Address) (IProviderStorageProvider, error) {
	return _ProviderPayouts.Contract.GetProvider(&_ProviderPayouts.CallOpts, provider_)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider_) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetProvider(provider_ common.Address) (IProviderStorageProvider, error) {
	return _ProviderPayouts.Contract.GetProvider(&_ProviderPayouts.CallOpts, provider_)
}

// GetProviderActiveBids is a free data retrieval call binding the contract method 0xaf5b77ca.
//
// Solidity: function getProviderActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetProviderActiveBids(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getProviderActiveBids", provider_, offset_, limit_)

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
func (_ProviderPayouts *ProviderPayoutsSession) GetProviderActiveBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _ProviderPayouts.Contract.GetProviderActiveBids(&_ProviderPayouts.CallOpts, provider_, offset_, limit_)
}

// GetProviderActiveBids is a free data retrieval call binding the contract method 0xaf5b77ca.
//
// Solidity: function getProviderActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetProviderActiveBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _ProviderPayouts.Contract.GetProviderActiveBids(&_ProviderPayouts.CallOpts, provider_, offset_, limit_)
}

// GetProviderBids is a free data retrieval call binding the contract method 0x59d435c4.
//
// Solidity: function getProviderBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetProviderBids(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getProviderBids", provider_, offset_, limit_)

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
func (_ProviderPayouts *ProviderPayoutsSession) GetProviderBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _ProviderPayouts.Contract.GetProviderBids(&_ProviderPayouts.CallOpts, provider_, offset_, limit_)
}

// GetProviderBids is a free data retrieval call binding the contract method 0x59d435c4.
//
// Solidity: function getProviderBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetProviderBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _ProviderPayouts.Contract.GetProviderBids(&_ProviderPayouts.CallOpts, provider_, offset_, limit_)
}

// GetProviderLifetimeClaimed is a free data retrieval call binding the contract method 0x432d7a82.
//
// Solidity: function getProviderLifetimeClaimed(address provider_) view returns(uint256, uint128)
func (_ProviderPayouts *ProviderPayoutsCaller) GetProviderLifetimeClaimed(opts *bind.CallOpts, provider_ common.Address) (*big.Int, *big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getProviderLifetimeClaimed", provider_)

	if err != nil {
		return *new(*big.Int), *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	out1 := *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return out0, out1, err

}

// GetProviderLifetimeClaimed is a free data retrieval call binding the contract method 0x432d7a82.
//
// Solidity: function getProviderLifetimeClaimed(address provider_) view returns(uint256, uint128)
func (_ProviderPayouts *ProviderPayoutsSession) GetProviderLifetimeClaimed(provider_ common.Address) (*big.Int, *big.Int, error) {
	return _ProviderPayouts.Contract.GetProviderLifetimeClaimed(&_ProviderPayouts.CallOpts, provider_)
}

// GetProviderLifetimeClaimed is a free data retrieval call binding the contract method 0x432d7a82.
//
// Solidity: function getProviderLifetimeClaimed(address provider_) view returns(uint256, uint128)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetProviderLifetimeClaimed(provider_ common.Address) (*big.Int, *big.Int, error) {
	return _ProviderPayouts.Contract.GetProviderLifetimeClaimed(&_ProviderPayouts.CallOpts, provider_)
}

// GetProviderMinimumStake is a free data retrieval call binding the contract method 0x53c029f6.
//
// Solidity: function getProviderMinimumStake() view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetProviderMinimumStake(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getProviderMinimumStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProviderMinimumStake is a free data retrieval call binding the contract method 0x53c029f6.
//
// Solidity: function getProviderMinimumStake() view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsSession) GetProviderMinimumStake() (*big.Int, error) {
	return _ProviderPayouts.Contract.GetProviderMinimumStake(&_ProviderPayouts.CallOpts)
}

// GetProviderMinimumStake is a free data retrieval call binding the contract method 0x53c029f6.
//
// Solidity: function getProviderMinimumStake() view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetProviderMinimumStake() (*big.Int, error) {
	return _ProviderPayouts.Contract.GetProviderMinimumStake(&_ProviderPayouts.CallOpts)
}

// GetProviderModelStats is a free data retrieval call binding the contract method 0x1b26c116.
//
// Solidity: function getProviderModelStats(bytes32 modelId_, address provider_) view returns(((int64,int64),(int64,int64),uint32,uint32,uint32))
func (_ProviderPayouts *ProviderPayoutsCaller) GetProviderModelStats(opts *bind.CallOpts, modelId_ [32]byte, provider_ common.Address) (IStatsStorageProviderModelStats, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getProviderModelStats", modelId_, provider_)

	if err != nil {
		return *new(IStatsStorageProviderModelStats), err
	}

	out0 := *abi.ConvertType(out[0], new(IStatsStorageProviderModelStats)).(*IStatsStorageProviderModelStats)

	return out0, err

}

// GetProviderModelStats is a free data retrieval call binding the contract method 0x1b26c116.
//
// Solidity: function getProviderModelStats(bytes32 modelId_, address provider_) view returns(((int64,int64),(int64,int64),uint32,uint32,uint32))
func (_ProviderPayouts *ProviderPayoutsSession) GetProviderModelStats(modelId_ [32]byte, provider_ common.Address) (IStatsStorageProviderModelStats, error) {
	return _ProviderPayouts.Contract.GetProviderModelStats(&_ProviderPayouts.CallOpts, modelId_, provider_)
}

// GetProviderModelStats is a free data retrieval call binding the contract method 0x1b26c116.
//
// Solidity: function getProviderModelStats(bytes32 modelId_, address provider_) view returns(((int64,int64),(int64,int64),uint32,uint32,uint32))
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetProviderModelStats(modelId_ [32]byte, provider_ common.Address) (IStatsStorageProviderModelStats, error) {
	return _ProviderPayouts.Contract.GetProviderModelStats(&_ProviderPayouts.CallOpts, modelId_, provider_)
}

// GetProviderOwed is a free data retrieval call binding the contract method 0xb1e3872c.
//
// Solidity: function getProviderOwed(address provider_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetProviderOwed(opts *bind.CallOpts, provider_ common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getProviderOwed", provider_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProviderOwed is a free data retrieval call binding the contract method 0xb1e3872c.
//
// Solidity: function getProviderOwed(address provider_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsSession) GetProviderOwed(provider_ common.Address) (*big.Int, error) {
	return _ProviderPayouts.Contract.GetProviderOwed(&_ProviderPayouts.CallOpts, provider_)
}

// GetProviderOwed is a free data retrieval call binding the contract method 0xb1e3872c.
//
// Solidity: function getProviderOwed(address provider_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetProviderOwed(provider_ common.Address) (*big.Int, error) {
	return _ProviderPayouts.Contract.GetProviderOwed(&_ProviderPayouts.CallOpts, provider_)
}

// GetProviderSessions is a free data retrieval call binding the contract method 0x87bced7d.
//
// Solidity: function getProviderSessions(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetProviderSessions(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getProviderSessions", provider_, offset_, limit_)

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
func (_ProviderPayouts *ProviderPayoutsSession) GetProviderSessions(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _ProviderPayouts.Contract.GetProviderSessions(&_ProviderPayouts.CallOpts, provider_, offset_, limit_)
}

// GetProviderSessions is a free data retrieval call binding the contract method 0x87bced7d.
//
// Solidity: function getProviderSessions(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetProviderSessions(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _ProviderPayouts.Contract.GetProviderSessions(&_ProviderPayouts.CallOpts, provider_, offset_, limit_)
}

// GetProvidersTotalClaimed is a free data retrieval call binding the contract method 0xbdfbbada.
//
// Solidity: function getProvidersTotalClaimed() view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetProvidersTotalClaimed(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getProvidersTotalClaimed")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProvidersTotalClaimed is a free data retrieval call binding the contract method 0xbdfbbada.
//
// Solidity: function getProvidersTotalClaimed() view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsSession) GetProvidersTotalClaimed() (*big.Int, error) {
	return _ProviderPayouts.Contract.GetProvidersTotalClaimed(&_ProviderPayouts.CallOpts)
}

// GetProvidersTotalClaimed is a free data retrieval call binding the contract method 0xbdfbbada.
//
// Solidity: function getProvidersTotalClaimed() view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetProvidersTotalClaimed() (*big.Int, error) {
	return _ProviderPayouts.Contract.GetProvidersTotalClaimed(&_ProviderPayouts.CallOpts)
}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_ProviderPayouts *ProviderPayoutsCaller) GetRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_ProviderPayouts *ProviderPayoutsSession) GetRegistry() (common.Address, error) {
	return _ProviderPayouts.Contract.GetRegistry(&_ProviderPayouts.CallOpts)
}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetRegistry() (common.Address, error) {
	return _ProviderPayouts.Contract.GetRegistry(&_ProviderPayouts.CallOpts)
}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId_) view returns((address,bytes32,uint256,bytes,uint256,uint256,uint128,uint128,uint128,bool,bool))
func (_ProviderPayouts *ProviderPayoutsCaller) GetSession(opts *bind.CallOpts, sessionId_ [32]byte) (ISessionStorageSession, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getSession", sessionId_)

	if err != nil {
		return *new(ISessionStorageSession), err
	}

	out0 := *abi.ConvertType(out[0], new(ISessionStorageSession)).(*ISessionStorageSession)

	return out0, err

}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId_) view returns((address,bytes32,uint256,bytes,uint256,uint256,uint128,uint128,uint128,bool,bool))
func (_ProviderPayouts *ProviderPayoutsSession) GetSession(sessionId_ [32]byte) (ISessionStorageSession, error) {
	return _ProviderPayouts.Contract.GetSession(&_ProviderPayouts.CallOpts, sessionId_)
}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId_) view returns((address,bytes32,uint256,bytes,uint256,uint256,uint128,uint128,uint128,bool,bool))
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetSession(sessionId_ [32]byte) (ISessionStorageSession, error) {
	return _ProviderPayouts.Contract.GetSession(&_ProviderPayouts.CallOpts, sessionId_)
}

// GetSessionEnd is a free data retrieval call binding the contract method 0xb823dc8f.
//
// Solidity: function getSessionEnd(uint256 amount_, uint256 pricePerSecond_, uint128 openedAt_) view returns(uint128)
func (_ProviderPayouts *ProviderPayoutsCaller) GetSessionEnd(opts *bind.CallOpts, amount_ *big.Int, pricePerSecond_ *big.Int, openedAt_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getSessionEnd", amount_, pricePerSecond_, openedAt_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetSessionEnd is a free data retrieval call binding the contract method 0xb823dc8f.
//
// Solidity: function getSessionEnd(uint256 amount_, uint256 pricePerSecond_, uint128 openedAt_) view returns(uint128)
func (_ProviderPayouts *ProviderPayoutsSession) GetSessionEnd(amount_ *big.Int, pricePerSecond_ *big.Int, openedAt_ *big.Int) (*big.Int, error) {
	return _ProviderPayouts.Contract.GetSessionEnd(&_ProviderPayouts.CallOpts, amount_, pricePerSecond_, openedAt_)
}

// GetSessionEnd is a free data retrieval call binding the contract method 0xb823dc8f.
//
// Solidity: function getSessionEnd(uint256 amount_, uint256 pricePerSecond_, uint128 openedAt_) view returns(uint128)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetSessionEnd(amount_ *big.Int, pricePerSecond_ *big.Int, openedAt_ *big.Int) (*big.Int, error) {
	return _ProviderPayouts.Contract.GetSessionEnd(&_ProviderPayouts.CallOpts, amount_, pricePerSecond_, openedAt_)
}

// GetSessionId is a free data retrieval call binding the contract method 0x4d689ffd.
//
// Solidity: function getSessionId(address user_, address provider_, bytes32 bidId_, uint256 sessionNonce_) pure returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCaller) GetSessionId(opts *bind.CallOpts, user_ common.Address, provider_ common.Address, bidId_ [32]byte, sessionNonce_ *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getSessionId", user_, provider_, bidId_, sessionNonce_)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetSessionId is a free data retrieval call binding the contract method 0x4d689ffd.
//
// Solidity: function getSessionId(address user_, address provider_, bytes32 bidId_, uint256 sessionNonce_) pure returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsSession) GetSessionId(user_ common.Address, provider_ common.Address, bidId_ [32]byte, sessionNonce_ *big.Int) ([32]byte, error) {
	return _ProviderPayouts.Contract.GetSessionId(&_ProviderPayouts.CallOpts, user_, provider_, bidId_, sessionNonce_)
}

// GetSessionId is a free data retrieval call binding the contract method 0x4d689ffd.
//
// Solidity: function getSessionId(address user_, address provider_, bytes32 bidId_, uint256 sessionNonce_) pure returns(bytes32)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetSessionId(user_ common.Address, provider_ common.Address, bidId_ [32]byte, sessionNonce_ *big.Int) ([32]byte, error) {
	return _ProviderPayouts.Contract.GetSessionId(&_ProviderPayouts.CallOpts, user_, provider_, bidId_, sessionNonce_)
}

// GetTodaysBudget is a free data retrieval call binding the contract method 0x40005965.
//
// Solidity: function getTodaysBudget(uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetTodaysBudget(opts *bind.CallOpts, timestamp_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getTodaysBudget", timestamp_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTodaysBudget is a free data retrieval call binding the contract method 0x40005965.
//
// Solidity: function getTodaysBudget(uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsSession) GetTodaysBudget(timestamp_ *big.Int) (*big.Int, error) {
	return _ProviderPayouts.Contract.GetTodaysBudget(&_ProviderPayouts.CallOpts, timestamp_)
}

// GetTodaysBudget is a free data retrieval call binding the contract method 0x40005965.
//
// Solidity: function getTodaysBudget(uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetTodaysBudget(timestamp_ *big.Int) (*big.Int, error) {
	return _ProviderPayouts.Contract.GetTodaysBudget(&_ProviderPayouts.CallOpts, timestamp_)
}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_ProviderPayouts *ProviderPayoutsCaller) GetToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_ProviderPayouts *ProviderPayoutsSession) GetToken() (common.Address, error) {
	return _ProviderPayouts.Contract.GetToken(&_ProviderPayouts.CallOpts)
}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetToken() (common.Address, error) {
	return _ProviderPayouts.Contract.GetToken(&_ProviderPayouts.CallOpts)
}

// GetTotalSessions is a free data retrieval call binding the contract method 0x76400935.
//
// Solidity: function getTotalSessions(address provider_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetTotalSessions(opts *bind.CallOpts, provider_ common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getTotalSessions", provider_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalSessions is a free data retrieval call binding the contract method 0x76400935.
//
// Solidity: function getTotalSessions(address provider_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsSession) GetTotalSessions(provider_ common.Address) (*big.Int, error) {
	return _ProviderPayouts.Contract.GetTotalSessions(&_ProviderPayouts.CallOpts, provider_)
}

// GetTotalSessions is a free data retrieval call binding the contract method 0x76400935.
//
// Solidity: function getTotalSessions(address provider_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetTotalSessions(provider_ common.Address) (*big.Int, error) {
	return _ProviderPayouts.Contract.GetTotalSessions(&_ProviderPayouts.CallOpts, provider_)
}

// GetUserSessions is a free data retrieval call binding the contract method 0xeb7764bb.
//
// Solidity: function getUserSessions(address user_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) GetUserSessions(opts *bind.CallOpts, user_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getUserSessions", user_, offset_, limit_)

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
func (_ProviderPayouts *ProviderPayoutsSession) GetUserSessions(user_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _ProviderPayouts.Contract.GetUserSessions(&_ProviderPayouts.CallOpts, user_, offset_, limit_)
}

// GetUserSessions is a free data retrieval call binding the contract method 0xeb7764bb.
//
// Solidity: function getUserSessions(address user_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetUserSessions(user_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _ProviderPayouts.Contract.GetUserSessions(&_ProviderPayouts.CallOpts, user_, offset_, limit_)
}

// GetUserStakesOnHold is a free data retrieval call binding the contract method 0x967885df.
//
// Solidity: function getUserStakesOnHold(address user_, uint8 iterations_) view returns(uint256 available_, uint256 hold_)
func (_ProviderPayouts *ProviderPayoutsCaller) GetUserStakesOnHold(opts *bind.CallOpts, user_ common.Address, iterations_ uint8) (struct {
	Available *big.Int
	Hold      *big.Int
}, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "getUserStakesOnHold", user_, iterations_)

	outstruct := new(struct {
		Available *big.Int
		Hold      *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Available = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Hold = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetUserStakesOnHold is a free data retrieval call binding the contract method 0x967885df.
//
// Solidity: function getUserStakesOnHold(address user_, uint8 iterations_) view returns(uint256 available_, uint256 hold_)
func (_ProviderPayouts *ProviderPayoutsSession) GetUserStakesOnHold(user_ common.Address, iterations_ uint8) (struct {
	Available *big.Int
	Hold      *big.Int
}, error) {
	return _ProviderPayouts.Contract.GetUserStakesOnHold(&_ProviderPayouts.CallOpts, user_, iterations_)
}

// GetUserStakesOnHold is a free data retrieval call binding the contract method 0x967885df.
//
// Solidity: function getUserStakesOnHold(address user_, uint8 iterations_) view returns(uint256 available_, uint256 hold_)
func (_ProviderPayouts *ProviderPayoutsCallerSession) GetUserStakesOnHold(user_ common.Address, iterations_ uint8) (struct {
	Available *big.Int
	Hold      *big.Int
}, error) {
	return _ProviderPayouts.Contract.GetUserStakesOnHold(&_ProviderPayouts.CallOpts, user_, iterations_)
}

// IsBidActive is a free data retrieval call binding the contract method 0x1345df58.
//
// Solidity: function isBidActive(bytes32 bidId_) view returns(bool)
func (_ProviderPayouts *ProviderPayoutsCaller) IsBidActive(opts *bind.CallOpts, bidId_ [32]byte) (bool, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "isBidActive", bidId_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsBidActive is a free data retrieval call binding the contract method 0x1345df58.
//
// Solidity: function isBidActive(bytes32 bidId_) view returns(bool)
func (_ProviderPayouts *ProviderPayoutsSession) IsBidActive(bidId_ [32]byte) (bool, error) {
	return _ProviderPayouts.Contract.IsBidActive(&_ProviderPayouts.CallOpts, bidId_)
}

// IsBidActive is a free data retrieval call binding the contract method 0x1345df58.
//
// Solidity: function isBidActive(bytes32 bidId_) view returns(bool)
func (_ProviderPayouts *ProviderPayoutsCallerSession) IsBidActive(bidId_ [32]byte) (bool, error) {
	return _ProviderPayouts.Contract.IsBidActive(&_ProviderPayouts.CallOpts, bidId_)
}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_ProviderPayouts *ProviderPayoutsCaller) IsRightsDelegated(opts *bind.CallOpts, delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "isRightsDelegated", delegatee_, delegator_, rights_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_ProviderPayouts *ProviderPayoutsSession) IsRightsDelegated(delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	return _ProviderPayouts.Contract.IsRightsDelegated(&_ProviderPayouts.CallOpts, delegatee_, delegator_, rights_)
}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_ProviderPayouts *ProviderPayoutsCallerSession) IsRightsDelegated(delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	return _ProviderPayouts.Contract.IsRightsDelegated(&_ProviderPayouts.CallOpts, delegatee_, delegator_, rights_)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ProviderPayouts *ProviderPayoutsCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ProviderPayouts *ProviderPayoutsSession) Owner() (common.Address, error) {
	return _ProviderPayouts.Contract.Owner(&_ProviderPayouts.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ProviderPayouts *ProviderPayoutsCallerSession) Owner() (common.Address, error) {
	return _ProviderPayouts.Contract.Owner(&_ProviderPayouts.CallOpts)
}

// StakeToStipend is a free data retrieval call binding the contract method 0xb3cb0d0f.
//
// Solidity: function stakeToStipend(uint256 amount_, uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) StakeToStipend(opts *bind.CallOpts, amount_ *big.Int, timestamp_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "stakeToStipend", amount_, timestamp_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StakeToStipend is a free data retrieval call binding the contract method 0xb3cb0d0f.
//
// Solidity: function stakeToStipend(uint256 amount_, uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsSession) StakeToStipend(amount_ *big.Int, timestamp_ *big.Int) (*big.Int, error) {
	return _ProviderPayouts.Contract.StakeToStipend(&_ProviderPayouts.CallOpts, amount_, timestamp_)
}

// StakeToStipend is a free data retrieval call binding the contract method 0xb3cb0d0f.
//
// Solidity: function stakeToStipend(uint256 amount_, uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) StakeToStipend(amount_ *big.Int, timestamp_ *big.Int) (*big.Int, error) {
	return _ProviderPayouts.Contract.StakeToStipend(&_ProviderPayouts.CallOpts, amount_, timestamp_)
}

// StartOfTheDay is a free data retrieval call binding the contract method 0xba26588c.
//
// Solidity: function startOfTheDay(uint128 timestamp_) pure returns(uint128)
func (_ProviderPayouts *ProviderPayoutsCaller) StartOfTheDay(opts *bind.CallOpts, timestamp_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "startOfTheDay", timestamp_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StartOfTheDay is a free data retrieval call binding the contract method 0xba26588c.
//
// Solidity: function startOfTheDay(uint128 timestamp_) pure returns(uint128)
func (_ProviderPayouts *ProviderPayoutsSession) StartOfTheDay(timestamp_ *big.Int) (*big.Int, error) {
	return _ProviderPayouts.Contract.StartOfTheDay(&_ProviderPayouts.CallOpts, timestamp_)
}

// StartOfTheDay is a free data retrieval call binding the contract method 0xba26588c.
//
// Solidity: function startOfTheDay(uint128 timestamp_) pure returns(uint128)
func (_ProviderPayouts *ProviderPayoutsCallerSession) StartOfTheDay(timestamp_ *big.Int) (*big.Int, error) {
	return _ProviderPayouts.Contract.StartOfTheDay(&_ProviderPayouts.CallOpts, timestamp_)
}

// StipendToStake is a free data retrieval call binding the contract method 0xca40d45f.
//
// Solidity: function stipendToStake(uint256 stipend_, uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) StipendToStake(opts *bind.CallOpts, stipend_ *big.Int, timestamp_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "stipendToStake", stipend_, timestamp_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StipendToStake is a free data retrieval call binding the contract method 0xca40d45f.
//
// Solidity: function stipendToStake(uint256 stipend_, uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsSession) StipendToStake(stipend_ *big.Int, timestamp_ *big.Int) (*big.Int, error) {
	return _ProviderPayouts.Contract.StipendToStake(&_ProviderPayouts.CallOpts, stipend_, timestamp_)
}

// StipendToStake is a free data retrieval call binding the contract method 0xca40d45f.
//
// Solidity: function stipendToStake(uint256 stipend_, uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) StipendToStake(stipend_ *big.Int, timestamp_ *big.Int) (*big.Int, error) {
	return _ProviderPayouts.Contract.StipendToStake(&_ProviderPayouts.CallOpts, stipend_, timestamp_)
}

// TotalMORSupply is a free data retrieval call binding the contract method 0x6d0cfe5a.
//
// Solidity: function totalMORSupply(uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCaller) TotalMORSupply(opts *bind.CallOpts, timestamp_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _ProviderPayouts.contract.Call(opts, &out, "totalMORSupply", timestamp_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalMORSupply is a free data retrieval call binding the contract method 0x6d0cfe5a.
//
// Solidity: function totalMORSupply(uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsSession) TotalMORSupply(timestamp_ *big.Int) (*big.Int, error) {
	return _ProviderPayouts.Contract.TotalMORSupply(&_ProviderPayouts.CallOpts, timestamp_)
}

// TotalMORSupply is a free data retrieval call binding the contract method 0x6d0cfe5a.
//
// Solidity: function totalMORSupply(uint128 timestamp_) view returns(uint256)
func (_ProviderPayouts *ProviderPayoutsCallerSession) TotalMORSupply(timestamp_ *big.Int) (*big.Int, error) {
	return _ProviderPayouts.Contract.TotalMORSupply(&_ProviderPayouts.CallOpts, timestamp_)
}

// ProviderPayoutsInit is a paid mutator transaction binding the contract method 0x81db0c1d.
//
// Solidity: function __ProviderPayouts_init() returns()
func (_ProviderPayouts *ProviderPayoutsTransactor) ProviderPayoutsInit(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProviderPayouts.contract.Transact(opts, "__ProviderPayouts_init")
}

// ProviderPayoutsInit is a paid mutator transaction binding the contract method 0x81db0c1d.
//
// Solidity: function __ProviderPayouts_init() returns()
func (_ProviderPayouts *ProviderPayoutsSession) ProviderPayoutsInit() (*types.Transaction, error) {
	return _ProviderPayouts.Contract.ProviderPayoutsInit(&_ProviderPayouts.TransactOpts)
}

// ProviderPayoutsInit is a paid mutator transaction binding the contract method 0x81db0c1d.
//
// Solidity: function __ProviderPayouts_init() returns()
func (_ProviderPayouts *ProviderPayoutsTransactorSession) ProviderPayoutsInit() (*types.Transaction, error) {
	return _ProviderPayouts.Contract.ProviderPayoutsInit(&_ProviderPayouts.TransactOpts)
}

// SetPayoutTarget is a paid mutator transaction binding the contract method 0xfc923e7b.
//
// Solidity: function setPayoutTarget(address provider_, address target_) returns()
func (_ProviderPayouts *ProviderPayoutsTransactor) SetPayoutTarget(opts *bind.TransactOpts, provider_ common.Address, target_ common.Address) (*types.Transaction, error) {
	return _ProviderPayouts.contract.Transact(opts, "setPayoutTarget", provider_, target_)
}

// SetPayoutTarget is a paid mutator transaction binding the contract method 0xfc923e7b.
//
// Solidity: function setPayoutTarget(address provider_, address target_) returns()
func (_ProviderPayouts *ProviderPayoutsSession) SetPayoutTarget(provider_ common.Address, target_ common.Address) (*types.Transaction, error) {
	return _ProviderPayouts.Contract.SetPayoutTarget(&_ProviderPayouts.TransactOpts, provider_, target_)
}

// SetPayoutTarget is a paid mutator transaction binding the contract method 0xfc923e7b.
//
// Solidity: function setPayoutTarget(address provider_, address target_) returns()
func (_ProviderPayouts *ProviderPayoutsTransactorSession) SetPayoutTarget(provider_ common.Address, target_ common.Address) (*types.Transaction, error) {
	return _ProviderPayouts.Contract.SetPayoutTarget(&_ProviderPayouts.TransactOpts, provider_, target_)
}

// ProviderPayoutsDelegateStakingAllowanceFundedIterator is returned from FilterDelegateStakingAllowanceFunded and is used to iterate over the raw logs and unpacked data for DelegateStakingAllowanceFunded events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingAllowanceFundedIterator struct {
	Event *ProviderPayoutsDelegateStakingAllowanceFunded // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingAllowanceFundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingAllowanceFunded)
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
		it.Event = new(ProviderPayoutsDelegateStakingAllowanceFunded)
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
func (it *ProviderPayoutsDelegateStakingAllowanceFundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingAllowanceFundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingAllowanceFunded represents a DelegateStakingAllowanceFunded event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingAllowanceFunded struct {
	Hot    common.Address
	Funder common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingAllowanceFunded is a free log retrieval operation binding the contract event 0xab72e208ad5755afbd9bcdfe0b4304d209b428685d5ee350433f417b117f38f0.
//
// Solidity: event DelegateStakingAllowanceFunded(address indexed hot, address indexed funder, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingAllowanceFunded(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*ProviderPayoutsDelegateStakingAllowanceFundedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingAllowanceFunded", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingAllowanceFundedIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingAllowanceFunded", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingAllowanceFunded is a free log subscription operation binding the contract event 0xab72e208ad5755afbd9bcdfe0b4304d209b428685d5ee350433f417b117f38f0.
//
// Solidity: event DelegateStakingAllowanceFunded(address indexed hot, address indexed funder, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingAllowanceFunded(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingAllowanceFunded, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingAllowanceFunded", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingAllowanceFunded)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingAllowanceFunded", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingAllowanceFunded(log types.Log) (*ProviderPayoutsDelegateStakingAllowanceFunded, error) {
	event := new(ProviderPayoutsDelegateStakingAllowanceFunded)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingAllowanceFunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingAllowanceGrantedIterator is returned from FilterDelegateStakingAllowanceGranted and is used to iterate over the raw logs and unpacked data for DelegateStakingAllowanceGranted events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingAllowanceGrantedIterator struct {
	Event *ProviderPayoutsDelegateStakingAllowanceGranted // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingAllowanceGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingAllowanceGranted)
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
		it.Event = new(ProviderPayoutsDelegateStakingAllowanceGranted)
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
func (it *ProviderPayoutsDelegateStakingAllowanceGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingAllowanceGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingAllowanceGranted represents a DelegateStakingAllowanceGranted event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingAllowanceGranted struct {
	Hot                  common.Address
	Funder               common.Address
	CumulativeFundingCap *big.Int
	Expiry               *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingAllowanceGranted is a free log retrieval operation binding the contract event 0x743ecb36e3b0fa6c16a5e8cd3ae70e657b294fd68ffbdfdc73d3a89d04addf46.
//
// Solidity: event DelegateStakingAllowanceGranted(address indexed hot, address indexed funder, uint256 cumulativeFundingCap, uint128 expiry)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingAllowanceGranted(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*ProviderPayoutsDelegateStakingAllowanceGrantedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingAllowanceGranted", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingAllowanceGrantedIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingAllowanceGranted", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingAllowanceGranted is a free log subscription operation binding the contract event 0x743ecb36e3b0fa6c16a5e8cd3ae70e657b294fd68ffbdfdc73d3a89d04addf46.
//
// Solidity: event DelegateStakingAllowanceGranted(address indexed hot, address indexed funder, uint256 cumulativeFundingCap, uint128 expiry)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingAllowanceGranted(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingAllowanceGranted, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingAllowanceGranted", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingAllowanceGranted)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingAllowanceGranted", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingAllowanceGranted(log types.Log) (*ProviderPayoutsDelegateStakingAllowanceGranted, error) {
	event := new(ProviderPayoutsDelegateStakingAllowanceGranted)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingAllowanceGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingAllowanceRevokedIterator is returned from FilterDelegateStakingAllowanceRevoked and is used to iterate over the raw logs and unpacked data for DelegateStakingAllowanceRevoked events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingAllowanceRevokedIterator struct {
	Event *ProviderPayoutsDelegateStakingAllowanceRevoked // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingAllowanceRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingAllowanceRevoked)
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
		it.Event = new(ProviderPayoutsDelegateStakingAllowanceRevoked)
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
func (it *ProviderPayoutsDelegateStakingAllowanceRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingAllowanceRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingAllowanceRevoked represents a DelegateStakingAllowanceRevoked event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingAllowanceRevoked struct {
	Hot    common.Address
	Funder common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingAllowanceRevoked is a free log retrieval operation binding the contract event 0x641afbae43e9aa28104693da9d716e2295e9efaf231c0bfec7dd6f949d3bd726.
//
// Solidity: event DelegateStakingAllowanceRevoked(address indexed hot, address indexed funder)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingAllowanceRevoked(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*ProviderPayoutsDelegateStakingAllowanceRevokedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingAllowanceRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingAllowanceRevokedIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingAllowanceRevoked", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingAllowanceRevoked is a free log subscription operation binding the contract event 0x641afbae43e9aa28104693da9d716e2295e9efaf231c0bfec7dd6f949d3bd726.
//
// Solidity: event DelegateStakingAllowanceRevoked(address indexed hot, address indexed funder)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingAllowanceRevoked(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingAllowanceRevoked, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingAllowanceRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingAllowanceRevoked)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingAllowanceRevoked", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingAllowanceRevoked(log types.Log) (*ProviderPayoutsDelegateStakingAllowanceRevoked, error) {
	event := new(ProviderPayoutsDelegateStakingAllowanceRevoked)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingAllowanceRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingMaxActiveFundersUpdatedIterator is returned from FilterDelegateStakingMaxActiveFundersUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMaxActiveFundersUpdated events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingMaxActiveFundersUpdatedIterator struct {
	Event *ProviderPayoutsDelegateStakingMaxActiveFundersUpdated // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingMaxActiveFundersUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingMaxActiveFundersUpdated)
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
		it.Event = new(ProviderPayoutsDelegateStakingMaxActiveFundersUpdated)
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
func (it *ProviderPayoutsDelegateStakingMaxActiveFundersUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingMaxActiveFundersUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingMaxActiveFundersUpdated represents a DelegateStakingMaxActiveFundersUpdated event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingMaxActiveFundersUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMaxActiveFundersUpdated is a free log retrieval operation binding the contract event 0x6018427c349577edf6a6c0b891cf09b86b47cc1834917516e6da404941ef43f5.
//
// Solidity: event DelegateStakingMaxActiveFundersUpdated(uint256 value)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingMaxActiveFundersUpdated(opts *bind.FilterOpts) (*ProviderPayoutsDelegateStakingMaxActiveFundersUpdatedIterator, error) {

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingMaxActiveFundersUpdated")
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingMaxActiveFundersUpdatedIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingMaxActiveFundersUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMaxActiveFundersUpdated is a free log subscription operation binding the contract event 0x6018427c349577edf6a6c0b891cf09b86b47cc1834917516e6da404941ef43f5.
//
// Solidity: event DelegateStakingMaxActiveFundersUpdated(uint256 value)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingMaxActiveFundersUpdated(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingMaxActiveFundersUpdated) (event.Subscription, error) {

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingMaxActiveFundersUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingMaxActiveFundersUpdated)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingMaxActiveFundersUpdated", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingMaxActiveFundersUpdated(log types.Log) (*ProviderPayoutsDelegateStakingMaxActiveFundersUpdated, error) {
	event := new(ProviderPayoutsDelegateStakingMaxActiveFundersUpdated)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingMaxActiveFundersUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdatedIterator is returned from FilterDelegateStakingMaxAutoReleaseDaysUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMaxAutoReleaseDaysUpdated events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdatedIterator struct {
	Event *ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdated // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdated)
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
		it.Event = new(ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdated)
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
func (it *ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdated represents a DelegateStakingMaxAutoReleaseDaysUpdated event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMaxAutoReleaseDaysUpdated is a free log retrieval operation binding the contract event 0xc9ca6843c5a726f8df173b27882fb647575843c62e31760e9b7f90c94518aa2a.
//
// Solidity: event DelegateStakingMaxAutoReleaseDaysUpdated(uint256 value)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingMaxAutoReleaseDaysUpdated(opts *bind.FilterOpts) (*ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdatedIterator, error) {

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingMaxAutoReleaseDaysUpdated")
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdatedIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingMaxAutoReleaseDaysUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMaxAutoReleaseDaysUpdated is a free log subscription operation binding the contract event 0xc9ca6843c5a726f8df173b27882fb647575843c62e31760e9b7f90c94518aa2a.
//
// Solidity: event DelegateStakingMaxAutoReleaseDaysUpdated(uint256 value)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingMaxAutoReleaseDaysUpdated(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdated) (event.Subscription, error) {

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingMaxAutoReleaseDaysUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdated)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingMaxAutoReleaseDaysUpdated", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingMaxAutoReleaseDaysUpdated(log types.Log) (*ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdated, error) {
	event := new(ProviderPayoutsDelegateStakingMaxAutoReleaseDaysUpdated)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingMaxAutoReleaseDaysUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingMaxAutoServiceUpdatedIterator is returned from FilterDelegateStakingMaxAutoServiceUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMaxAutoServiceUpdated events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingMaxAutoServiceUpdatedIterator struct {
	Event *ProviderPayoutsDelegateStakingMaxAutoServiceUpdated // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingMaxAutoServiceUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingMaxAutoServiceUpdated)
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
		it.Event = new(ProviderPayoutsDelegateStakingMaxAutoServiceUpdated)
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
func (it *ProviderPayoutsDelegateStakingMaxAutoServiceUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingMaxAutoServiceUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingMaxAutoServiceUpdated represents a DelegateStakingMaxAutoServiceUpdated event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingMaxAutoServiceUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMaxAutoServiceUpdated is a free log retrieval operation binding the contract event 0x4f927d3e460674eab7e38dac3592136fe961de0c14533b919652a814a7510adc.
//
// Solidity: event DelegateStakingMaxAutoServiceUpdated(uint256 value)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingMaxAutoServiceUpdated(opts *bind.FilterOpts) (*ProviderPayoutsDelegateStakingMaxAutoServiceUpdatedIterator, error) {

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingMaxAutoServiceUpdated")
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingMaxAutoServiceUpdatedIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingMaxAutoServiceUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMaxAutoServiceUpdated is a free log subscription operation binding the contract event 0x4f927d3e460674eab7e38dac3592136fe961de0c14533b919652a814a7510adc.
//
// Solidity: event DelegateStakingMaxAutoServiceUpdated(uint256 value)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingMaxAutoServiceUpdated(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingMaxAutoServiceUpdated) (event.Subscription, error) {

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingMaxAutoServiceUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingMaxAutoServiceUpdated)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingMaxAutoServiceUpdated", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingMaxAutoServiceUpdated(log types.Log) (*ProviderPayoutsDelegateStakingMaxAutoServiceUpdated, error) {
	event := new(ProviderPayoutsDelegateStakingMaxAutoServiceUpdated)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingMaxAutoServiceUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingMinPrincipalUpdatedIterator is returned from FilterDelegateStakingMinPrincipalUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMinPrincipalUpdated events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingMinPrincipalUpdatedIterator struct {
	Event *ProviderPayoutsDelegateStakingMinPrincipalUpdated // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingMinPrincipalUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingMinPrincipalUpdated)
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
		it.Event = new(ProviderPayoutsDelegateStakingMinPrincipalUpdated)
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
func (it *ProviderPayoutsDelegateStakingMinPrincipalUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingMinPrincipalUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingMinPrincipalUpdated represents a DelegateStakingMinPrincipalUpdated event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingMinPrincipalUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMinPrincipalUpdated is a free log retrieval operation binding the contract event 0x06ae87e705be2a5e7c6ff17a5d3d836834554f7d0aadaef3abb4343a202590a2.
//
// Solidity: event DelegateStakingMinPrincipalUpdated(uint256 value)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingMinPrincipalUpdated(opts *bind.FilterOpts) (*ProviderPayoutsDelegateStakingMinPrincipalUpdatedIterator, error) {

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingMinPrincipalUpdated")
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingMinPrincipalUpdatedIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingMinPrincipalUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMinPrincipalUpdated is a free log subscription operation binding the contract event 0x06ae87e705be2a5e7c6ff17a5d3d836834554f7d0aadaef3abb4343a202590a2.
//
// Solidity: event DelegateStakingMinPrincipalUpdated(uint256 value)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingMinPrincipalUpdated(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingMinPrincipalUpdated) (event.Subscription, error) {

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingMinPrincipalUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingMinPrincipalUpdated)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingMinPrincipalUpdated", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingMinPrincipalUpdated(log types.Log) (*ProviderPayoutsDelegateStakingMinPrincipalUpdated, error) {
	event := new(ProviderPayoutsDelegateStakingMinPrincipalUpdated)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingMinPrincipalUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingPoolDrawnIterator is returned from FilterDelegateStakingPoolDrawn and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolDrawn events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingPoolDrawnIterator struct {
	Event *ProviderPayoutsDelegateStakingPoolDrawn // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingPoolDrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingPoolDrawn)
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
		it.Event = new(ProviderPayoutsDelegateStakingPoolDrawn)
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
func (it *ProviderPayoutsDelegateStakingPoolDrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingPoolDrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingPoolDrawn represents a DelegateStakingPoolDrawn event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingPoolDrawn struct {
	SessionId [32]byte
	Hot       common.Address
	Funder    common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolDrawn is a free log retrieval operation binding the contract event 0xd178a2f3e5da2cc00732e1b88d2670a1eaa4f88a0b86e8790237a677f1fa0985.
//
// Solidity: event DelegateStakingPoolDrawn(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingPoolDrawn(opts *bind.FilterOpts, sessionId [][32]byte, hot []common.Address, funder []common.Address) (*ProviderPayoutsDelegateStakingPoolDrawnIterator, error) {

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

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingPoolDrawn", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingPoolDrawnIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingPoolDrawn", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolDrawn is a free log subscription operation binding the contract event 0xd178a2f3e5da2cc00732e1b88d2670a1eaa4f88a0b86e8790237a677f1fa0985.
//
// Solidity: event DelegateStakingPoolDrawn(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingPoolDrawn(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingPoolDrawn, sessionId [][32]byte, hot []common.Address, funder []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingPoolDrawn", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingPoolDrawn)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingPoolDrawn", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingPoolDrawn(log types.Log) (*ProviderPayoutsDelegateStakingPoolDrawn, error) {
	event := new(ProviderPayoutsDelegateStakingPoolDrawn)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingPoolDrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingPoolFunderApprovalRevokedIterator is returned from FilterDelegateStakingPoolFunderApprovalRevoked and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolFunderApprovalRevoked events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingPoolFunderApprovalRevokedIterator struct {
	Event *ProviderPayoutsDelegateStakingPoolFunderApprovalRevoked // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingPoolFunderApprovalRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingPoolFunderApprovalRevoked)
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
		it.Event = new(ProviderPayoutsDelegateStakingPoolFunderApprovalRevoked)
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
func (it *ProviderPayoutsDelegateStakingPoolFunderApprovalRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingPoolFunderApprovalRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingPoolFunderApprovalRevoked represents a DelegateStakingPoolFunderApprovalRevoked event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingPoolFunderApprovalRevoked struct {
	Hot    common.Address
	Funder common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolFunderApprovalRevoked is a free log retrieval operation binding the contract event 0x5ebb99e7e14d99bff22536b105e92a32d4fbb701f93b0d5b008cc0a0b876aeb2.
//
// Solidity: event DelegateStakingPoolFunderApprovalRevoked(address indexed hot, address indexed funder)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingPoolFunderApprovalRevoked(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*ProviderPayoutsDelegateStakingPoolFunderApprovalRevokedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingPoolFunderApprovalRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingPoolFunderApprovalRevokedIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingPoolFunderApprovalRevoked", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolFunderApprovalRevoked is a free log subscription operation binding the contract event 0x5ebb99e7e14d99bff22536b105e92a32d4fbb701f93b0d5b008cc0a0b876aeb2.
//
// Solidity: event DelegateStakingPoolFunderApprovalRevoked(address indexed hot, address indexed funder)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingPoolFunderApprovalRevoked(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingPoolFunderApprovalRevoked, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingPoolFunderApprovalRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingPoolFunderApprovalRevoked)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingPoolFunderApprovalRevoked", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingPoolFunderApprovalRevoked(log types.Log) (*ProviderPayoutsDelegateStakingPoolFunderApprovalRevoked, error) {
	event := new(ProviderPayoutsDelegateStakingPoolFunderApprovalRevoked)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingPoolFunderApprovalRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingPoolFunderApprovedIterator is returned from FilterDelegateStakingPoolFunderApproved and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolFunderApproved events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingPoolFunderApprovedIterator struct {
	Event *ProviderPayoutsDelegateStakingPoolFunderApproved // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingPoolFunderApprovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingPoolFunderApproved)
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
		it.Event = new(ProviderPayoutsDelegateStakingPoolFunderApproved)
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
func (it *ProviderPayoutsDelegateStakingPoolFunderApprovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingPoolFunderApprovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingPoolFunderApproved represents a DelegateStakingPoolFunderApproved event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingPoolFunderApproved struct {
	Hot    common.Address
	Funder common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolFunderApproved is a free log retrieval operation binding the contract event 0x2a6f85b141a6737d1530bd4b8c916eee97ba403dae709170266f72b6ec58f739.
//
// Solidity: event DelegateStakingPoolFunderApproved(address indexed hot, address indexed funder)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingPoolFunderApproved(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*ProviderPayoutsDelegateStakingPoolFunderApprovedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingPoolFunderApproved", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingPoolFunderApprovedIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingPoolFunderApproved", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolFunderApproved is a free log subscription operation binding the contract event 0x2a6f85b141a6737d1530bd4b8c916eee97ba403dae709170266f72b6ec58f739.
//
// Solidity: event DelegateStakingPoolFunderApproved(address indexed hot, address indexed funder)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingPoolFunderApproved(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingPoolFunderApproved, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingPoolFunderApproved", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingPoolFunderApproved)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingPoolFunderApproved", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingPoolFunderApproved(log types.Log) (*ProviderPayoutsDelegateStakingPoolFunderApproved, error) {
	event := new(ProviderPayoutsDelegateStakingPoolFunderApproved)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingPoolFunderApproved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingPoolLockCreatedIterator is returned from FilterDelegateStakingPoolLockCreated and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolLockCreated events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingPoolLockCreatedIterator struct {
	Event *ProviderPayoutsDelegateStakingPoolLockCreated // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingPoolLockCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingPoolLockCreated)
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
		it.Event = new(ProviderPayoutsDelegateStakingPoolLockCreated)
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
func (it *ProviderPayoutsDelegateStakingPoolLockCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingPoolLockCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingPoolLockCreated represents a DelegateStakingPoolLockCreated event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingPoolLockCreated struct {
	Hot       common.Address
	ReleaseAt *big.Int
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolLockCreated is a free log retrieval operation binding the contract event 0x87eb288bde7ca5d84aa5e85a8b81f8d7c31d8e19049823dab5552d8bb461dc0e.
//
// Solidity: event DelegateStakingPoolLockCreated(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingPoolLockCreated(opts *bind.FilterOpts, hot []common.Address, releaseAt []*big.Int) (*ProviderPayoutsDelegateStakingPoolLockCreatedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingPoolLockCreated", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingPoolLockCreatedIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingPoolLockCreated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolLockCreated is a free log subscription operation binding the contract event 0x87eb288bde7ca5d84aa5e85a8b81f8d7c31d8e19049823dab5552d8bb461dc0e.
//
// Solidity: event DelegateStakingPoolLockCreated(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingPoolLockCreated(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingPoolLockCreated, hot []common.Address, releaseAt []*big.Int) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingPoolLockCreated", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingPoolLockCreated)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingPoolLockCreated", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingPoolLockCreated(log types.Log) (*ProviderPayoutsDelegateStakingPoolLockCreated, error) {
	event := new(ProviderPayoutsDelegateStakingPoolLockCreated)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingPoolLockCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingPoolLockReleasedIterator is returned from FilterDelegateStakingPoolLockReleased and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolLockReleased events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingPoolLockReleasedIterator struct {
	Event *ProviderPayoutsDelegateStakingPoolLockReleased // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingPoolLockReleasedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingPoolLockReleased)
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
		it.Event = new(ProviderPayoutsDelegateStakingPoolLockReleased)
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
func (it *ProviderPayoutsDelegateStakingPoolLockReleasedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingPoolLockReleasedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingPoolLockReleased represents a DelegateStakingPoolLockReleased event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingPoolLockReleased struct {
	Hot       common.Address
	ReleaseAt *big.Int
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolLockReleased is a free log retrieval operation binding the contract event 0xd04817baed5e7c2b01f59b2218c05170e733b0e2a3864693721a2fc89e5f7a4f.
//
// Solidity: event DelegateStakingPoolLockReleased(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingPoolLockReleased(opts *bind.FilterOpts, hot []common.Address, releaseAt []*big.Int) (*ProviderPayoutsDelegateStakingPoolLockReleasedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingPoolLockReleased", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingPoolLockReleasedIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingPoolLockReleased", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolLockReleased is a free log subscription operation binding the contract event 0xd04817baed5e7c2b01f59b2218c05170e733b0e2a3864693721a2fc89e5f7a4f.
//
// Solidity: event DelegateStakingPoolLockReleased(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingPoolLockReleased(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingPoolLockReleased, hot []common.Address, releaseAt []*big.Int) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingPoolLockReleased", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingPoolLockReleased)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingPoolLockReleased", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingPoolLockReleased(log types.Log) (*ProviderPayoutsDelegateStakingPoolLockReleased, error) {
	event := new(ProviderPayoutsDelegateStakingPoolLockReleased)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingPoolLockReleased", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingPoolReturnedIterator is returned from FilterDelegateStakingPoolReturned and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolReturned events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingPoolReturnedIterator struct {
	Event *ProviderPayoutsDelegateStakingPoolReturned // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingPoolReturnedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingPoolReturned)
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
		it.Event = new(ProviderPayoutsDelegateStakingPoolReturned)
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
func (it *ProviderPayoutsDelegateStakingPoolReturnedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingPoolReturnedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingPoolReturned represents a DelegateStakingPoolReturned event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingPoolReturned struct {
	SessionId [32]byte
	Hot       common.Address
	Funder    common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolReturned is a free log retrieval operation binding the contract event 0x66524d7951752d8716b27064a7d50819e4db87609ea9eca1f6b1d42f359a061d.
//
// Solidity: event DelegateStakingPoolReturned(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingPoolReturned(opts *bind.FilterOpts, sessionId [][32]byte, hot []common.Address, funder []common.Address) (*ProviderPayoutsDelegateStakingPoolReturnedIterator, error) {

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

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingPoolReturned", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingPoolReturnedIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingPoolReturned", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolReturned is a free log subscription operation binding the contract event 0x66524d7951752d8716b27064a7d50819e4db87609ea9eca1f6b1d42f359a061d.
//
// Solidity: event DelegateStakingPoolReturned(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingPoolReturned(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingPoolReturned, sessionId [][32]byte, hot []common.Address, funder []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingPoolReturned", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingPoolReturned)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingPoolReturned", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingPoolReturned(log types.Log) (*ProviderPayoutsDelegateStakingPoolReturned, error) {
	event := new(ProviderPayoutsDelegateStakingPoolReturned)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingPoolReturned", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingPoolSessionSettledIterator is returned from FilterDelegateStakingPoolSessionSettled and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolSessionSettled events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingPoolSessionSettledIterator struct {
	Event *ProviderPayoutsDelegateStakingPoolSessionSettled // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingPoolSessionSettledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingPoolSessionSettled)
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
		it.Event = new(ProviderPayoutsDelegateStakingPoolSessionSettled)
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
func (it *ProviderPayoutsDelegateStakingPoolSessionSettledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingPoolSessionSettledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingPoolSessionSettled represents a DelegateStakingPoolSessionSettled event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingPoolSessionSettled struct {
	SessionId [32]byte
	Hot       common.Address
	Provider  common.Address
	Caller    common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolSessionSettled is a free log retrieval operation binding the contract event 0x518bf7ca4dc0c0c2144467883693da5fd9b0532de6539122309ecff49a330215.
//
// Solidity: event DelegateStakingPoolSessionSettled(bytes32 indexed sessionId, address indexed hot, address indexed provider, address caller)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingPoolSessionSettled(opts *bind.FilterOpts, sessionId [][32]byte, hot []common.Address, provider []common.Address) (*ProviderPayoutsDelegateStakingPoolSessionSettledIterator, error) {

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

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingPoolSessionSettled", sessionIdRule, hotRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingPoolSessionSettledIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingPoolSessionSettled", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolSessionSettled is a free log subscription operation binding the contract event 0x518bf7ca4dc0c0c2144467883693da5fd9b0532de6539122309ecff49a330215.
//
// Solidity: event DelegateStakingPoolSessionSettled(bytes32 indexed sessionId, address indexed hot, address indexed provider, address caller)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingPoolSessionSettled(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingPoolSessionSettled, sessionId [][32]byte, hot []common.Address, provider []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingPoolSessionSettled", sessionIdRule, hotRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingPoolSessionSettled)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingPoolSessionSettled", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingPoolSessionSettled(log types.Log) (*ProviderPayoutsDelegateStakingPoolSessionSettled, error) {
	event := new(ProviderPayoutsDelegateStakingPoolSessionSettled)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingPoolSessionSettled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingSettlementGraceUpdatedIterator is returned from FilterDelegateStakingSettlementGraceUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingSettlementGraceUpdated events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingSettlementGraceUpdatedIterator struct {
	Event *ProviderPayoutsDelegateStakingSettlementGraceUpdated // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingSettlementGraceUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingSettlementGraceUpdated)
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
		it.Event = new(ProviderPayoutsDelegateStakingSettlementGraceUpdated)
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
func (it *ProviderPayoutsDelegateStakingSettlementGraceUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingSettlementGraceUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingSettlementGraceUpdated represents a DelegateStakingSettlementGraceUpdated event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingSettlementGraceUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingSettlementGraceUpdated is a free log retrieval operation binding the contract event 0x23d4c1527dfe24f1a043524d5c7c2b730f3fb2841478487080916494bc8487f2.
//
// Solidity: event DelegateStakingSettlementGraceUpdated(uint256 value)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingSettlementGraceUpdated(opts *bind.FilterOpts) (*ProviderPayoutsDelegateStakingSettlementGraceUpdatedIterator, error) {

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingSettlementGraceUpdated")
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingSettlementGraceUpdatedIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingSettlementGraceUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingSettlementGraceUpdated is a free log subscription operation binding the contract event 0x23d4c1527dfe24f1a043524d5c7c2b730f3fb2841478487080916494bc8487f2.
//
// Solidity: event DelegateStakingSettlementGraceUpdated(uint256 value)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingSettlementGraceUpdated(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingSettlementGraceUpdated) (event.Subscription, error) {

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingSettlementGraceUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingSettlementGraceUpdated)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingSettlementGraceUpdated", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingSettlementGraceUpdated(log types.Log) (*ProviderPayoutsDelegateStakingSettlementGraceUpdated, error) {
	event := new(ProviderPayoutsDelegateStakingSettlementGraceUpdated)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingSettlementGraceUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingWithdrawalCancelledIterator is returned from FilterDelegateStakingWithdrawalCancelled and is used to iterate over the raw logs and unpacked data for DelegateStakingWithdrawalCancelled events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingWithdrawalCancelledIterator struct {
	Event *ProviderPayoutsDelegateStakingWithdrawalCancelled // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingWithdrawalCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingWithdrawalCancelled)
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
		it.Event = new(ProviderPayoutsDelegateStakingWithdrawalCancelled)
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
func (it *ProviderPayoutsDelegateStakingWithdrawalCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingWithdrawalCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingWithdrawalCancelled represents a DelegateStakingWithdrawalCancelled event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingWithdrawalCancelled struct {
	Hot    common.Address
	Funder common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingWithdrawalCancelled is a free log retrieval operation binding the contract event 0x00041bc92ab137f1f37671a48ce9dc19f99eb0d99b0bf41beb1ffca2dd9cb557.
//
// Solidity: event DelegateStakingWithdrawalCancelled(address indexed hot, address indexed funder, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingWithdrawalCancelled(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*ProviderPayoutsDelegateStakingWithdrawalCancelledIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingWithdrawalCancelled", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingWithdrawalCancelledIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingWithdrawalCancelled", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingWithdrawalCancelled is a free log subscription operation binding the contract event 0x00041bc92ab137f1f37671a48ce9dc19f99eb0d99b0bf41beb1ffca2dd9cb557.
//
// Solidity: event DelegateStakingWithdrawalCancelled(address indexed hot, address indexed funder, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingWithdrawalCancelled(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingWithdrawalCancelled, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingWithdrawalCancelled", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingWithdrawalCancelled)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingWithdrawalCancelled", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingWithdrawalCancelled(log types.Log) (*ProviderPayoutsDelegateStakingWithdrawalCancelled, error) {
	event := new(ProviderPayoutsDelegateStakingWithdrawalCancelled)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingWithdrawalCancelled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingWithdrawalPaidIterator is returned from FilterDelegateStakingWithdrawalPaid and is used to iterate over the raw logs and unpacked data for DelegateStakingWithdrawalPaid events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingWithdrawalPaidIterator struct {
	Event *ProviderPayoutsDelegateStakingWithdrawalPaid // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingWithdrawalPaidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingWithdrawalPaid)
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
		it.Event = new(ProviderPayoutsDelegateStakingWithdrawalPaid)
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
func (it *ProviderPayoutsDelegateStakingWithdrawalPaidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingWithdrawalPaidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingWithdrawalPaid represents a DelegateStakingWithdrawalPaid event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingWithdrawalPaid struct {
	Hot    common.Address
	Funder common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingWithdrawalPaid is a free log retrieval operation binding the contract event 0x7827e4dd700b7a2a556c261a5a293c5c936068522520173931017b0e9959d935.
//
// Solidity: event DelegateStakingWithdrawalPaid(address indexed hot, address indexed funder, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingWithdrawalPaid(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*ProviderPayoutsDelegateStakingWithdrawalPaidIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingWithdrawalPaid", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingWithdrawalPaidIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingWithdrawalPaid", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingWithdrawalPaid is a free log subscription operation binding the contract event 0x7827e4dd700b7a2a556c261a5a293c5c936068522520173931017b0e9959d935.
//
// Solidity: event DelegateStakingWithdrawalPaid(address indexed hot, address indexed funder, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingWithdrawalPaid(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingWithdrawalPaid, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingWithdrawalPaid", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingWithdrawalPaid)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingWithdrawalPaid", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingWithdrawalPaid(log types.Log) (*ProviderPayoutsDelegateStakingWithdrawalPaid, error) {
	event := new(ProviderPayoutsDelegateStakingWithdrawalPaid)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingWithdrawalPaid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsDelegateStakingWithdrawalRequestedIterator is returned from FilterDelegateStakingWithdrawalRequested and is used to iterate over the raw logs and unpacked data for DelegateStakingWithdrawalRequested events raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingWithdrawalRequestedIterator struct {
	Event *ProviderPayoutsDelegateStakingWithdrawalRequested // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsDelegateStakingWithdrawalRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsDelegateStakingWithdrawalRequested)
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
		it.Event = new(ProviderPayoutsDelegateStakingWithdrawalRequested)
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
func (it *ProviderPayoutsDelegateStakingWithdrawalRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsDelegateStakingWithdrawalRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsDelegateStakingWithdrawalRequested represents a DelegateStakingWithdrawalRequested event raised by the ProviderPayouts contract.
type ProviderPayoutsDelegateStakingWithdrawalRequested struct {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterDelegateStakingWithdrawalRequested(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*ProviderPayoutsDelegateStakingWithdrawalRequestedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "DelegateStakingWithdrawalRequested", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsDelegateStakingWithdrawalRequestedIterator{contract: _ProviderPayouts.contract, event: "DelegateStakingWithdrawalRequested", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingWithdrawalRequested is a free log subscription operation binding the contract event 0xab9f026078e090833bf0fd1211137db97a7a96f822c650c7997cdb7bfc859799.
//
// Solidity: event DelegateStakingWithdrawalRequested(address indexed hot, address indexed funder, uint256 amount, uint256 paid, uint256 pending)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchDelegateStakingWithdrawalRequested(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsDelegateStakingWithdrawalRequested, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "DelegateStakingWithdrawalRequested", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsDelegateStakingWithdrawalRequested)
				if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingWithdrawalRequested", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseDelegateStakingWithdrawalRequested(log types.Log) (*ProviderPayoutsDelegateStakingWithdrawalRequested, error) {
	event := new(ProviderPayoutsDelegateStakingWithdrawalRequested)
	if err := _ProviderPayouts.contract.UnpackLog(event, "DelegateStakingWithdrawalRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the ProviderPayouts contract.
type ProviderPayoutsInitializedIterator struct {
	Event *ProviderPayoutsInitialized // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsInitialized)
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
		it.Event = new(ProviderPayoutsInitialized)
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
func (it *ProviderPayoutsInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsInitialized represents a Initialized event raised by the ProviderPayouts contract.
type ProviderPayoutsInitialized struct {
	StorageSlot [32]byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterInitialized(opts *bind.FilterOpts) (*ProviderPayoutsInitializedIterator, error) {

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsInitializedIterator{contract: _ProviderPayouts.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsInitialized) (event.Subscription, error) {

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsInitialized)
				if err := _ProviderPayouts.contract.UnpackLog(event, "Initialized", log); err != nil {
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

// ParseInitialized is a log parse operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseInitialized(log types.Log) (*ProviderPayoutsInitialized, error) {
	event := new(ProviderPayoutsInitialized)
	if err := _ProviderPayouts.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsProviderDebtPaidIterator is returned from FilterProviderDebtPaid and is used to iterate over the raw logs and unpacked data for ProviderDebtPaid events raised by the ProviderPayouts contract.
type ProviderPayoutsProviderDebtPaidIterator struct {
	Event *ProviderPayoutsProviderDebtPaid // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsProviderDebtPaidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsProviderDebtPaid)
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
		it.Event = new(ProviderPayoutsProviderDebtPaid)
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
func (it *ProviderPayoutsProviderDebtPaidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsProviderDebtPaidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsProviderDebtPaid represents a ProviderDebtPaid event raised by the ProviderPayouts contract.
type ProviderPayoutsProviderDebtPaid struct {
	Provider     common.Address
	PayoutTarget common.Address
	Amount       *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterProviderDebtPaid is a free log retrieval operation binding the contract event 0x8d79baa0ce0a07e4a85c9f1c273815f861513a2e9fdb2082bf443519534aa2b0.
//
// Solidity: event ProviderDebtPaid(address indexed provider, address indexed payoutTarget, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterProviderDebtPaid(opts *bind.FilterOpts, provider []common.Address, payoutTarget []common.Address) (*ProviderPayoutsProviderDebtPaidIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var payoutTargetRule []interface{}
	for _, payoutTargetItem := range payoutTarget {
		payoutTargetRule = append(payoutTargetRule, payoutTargetItem)
	}

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "ProviderDebtPaid", providerRule, payoutTargetRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsProviderDebtPaidIterator{contract: _ProviderPayouts.contract, event: "ProviderDebtPaid", logs: logs, sub: sub}, nil
}

// WatchProviderDebtPaid is a free log subscription operation binding the contract event 0x8d79baa0ce0a07e4a85c9f1c273815f861513a2e9fdb2082bf443519534aa2b0.
//
// Solidity: event ProviderDebtPaid(address indexed provider, address indexed payoutTarget, uint256 amount)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchProviderDebtPaid(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsProviderDebtPaid, provider []common.Address, payoutTarget []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var payoutTargetRule []interface{}
	for _, payoutTargetItem := range payoutTarget {
		payoutTargetRule = append(payoutTargetRule, payoutTargetItem)
	}

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "ProviderDebtPaid", providerRule, payoutTargetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsProviderDebtPaid)
				if err := _ProviderPayouts.contract.UnpackLog(event, "ProviderDebtPaid", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseProviderDebtPaid(log types.Log) (*ProviderPayoutsProviderDebtPaid, error) {
	event := new(ProviderPayoutsProviderDebtPaid)
	if err := _ProviderPayouts.contract.UnpackLog(event, "ProviderDebtPaid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsProviderPayoutTargetUpdatedIterator is returned from FilterProviderPayoutTargetUpdated and is used to iterate over the raw logs and unpacked data for ProviderPayoutTargetUpdated events raised by the ProviderPayouts contract.
type ProviderPayoutsProviderPayoutTargetUpdatedIterator struct {
	Event *ProviderPayoutsProviderPayoutTargetUpdated // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsProviderPayoutTargetUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsProviderPayoutTargetUpdated)
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
		it.Event = new(ProviderPayoutsProviderPayoutTargetUpdated)
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
func (it *ProviderPayoutsProviderPayoutTargetUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsProviderPayoutTargetUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsProviderPayoutTargetUpdated represents a ProviderPayoutTargetUpdated event raised by the ProviderPayouts contract.
type ProviderPayoutsProviderPayoutTargetUpdated struct {
	Provider     common.Address
	PayoutTarget common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterProviderPayoutTargetUpdated is a free log retrieval operation binding the contract event 0x1cba50759ff5d1e813b2ed6b4f9ce6cd19137e624f4bbde798eb0572b74a6968.
//
// Solidity: event ProviderPayoutTargetUpdated(address indexed provider, address indexed payoutTarget)
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterProviderPayoutTargetUpdated(opts *bind.FilterOpts, provider []common.Address, payoutTarget []common.Address) (*ProviderPayoutsProviderPayoutTargetUpdatedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var payoutTargetRule []interface{}
	for _, payoutTargetItem := range payoutTarget {
		payoutTargetRule = append(payoutTargetRule, payoutTargetItem)
	}

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "ProviderPayoutTargetUpdated", providerRule, payoutTargetRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsProviderPayoutTargetUpdatedIterator{contract: _ProviderPayouts.contract, event: "ProviderPayoutTargetUpdated", logs: logs, sub: sub}, nil
}

// WatchProviderPayoutTargetUpdated is a free log subscription operation binding the contract event 0x1cba50759ff5d1e813b2ed6b4f9ce6cd19137e624f4bbde798eb0572b74a6968.
//
// Solidity: event ProviderPayoutTargetUpdated(address indexed provider, address indexed payoutTarget)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchProviderPayoutTargetUpdated(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsProviderPayoutTargetUpdated, provider []common.Address, payoutTarget []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var payoutTargetRule []interface{}
	for _, payoutTargetItem := range payoutTarget {
		payoutTargetRule = append(payoutTargetRule, payoutTargetItem)
	}

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "ProviderPayoutTargetUpdated", providerRule, payoutTargetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsProviderPayoutTargetUpdated)
				if err := _ProviderPayouts.contract.UnpackLog(event, "ProviderPayoutTargetUpdated", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseProviderPayoutTargetUpdated(log types.Log) (*ProviderPayoutsProviderPayoutTargetUpdated, error) {
	event := new(ProviderPayoutsProviderPayoutTargetUpdated)
	if err := _ProviderPayouts.contract.UnpackLog(event, "ProviderPayoutTargetUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsProviderRewardAccruedIterator is returned from FilterProviderRewardAccrued and is used to iterate over the raw logs and unpacked data for ProviderRewardAccrued events raised by the ProviderPayouts contract.
type ProviderPayoutsProviderRewardAccruedIterator struct {
	Event *ProviderPayoutsProviderRewardAccrued // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsProviderRewardAccruedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsProviderRewardAccrued)
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
		it.Event = new(ProviderPayoutsProviderRewardAccrued)
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
func (it *ProviderPayoutsProviderRewardAccruedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsProviderRewardAccruedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsProviderRewardAccrued represents a ProviderRewardAccrued event raised by the ProviderPayouts contract.
type ProviderPayoutsProviderRewardAccrued struct {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterProviderRewardAccrued(opts *bind.FilterOpts, sessionId [][32]byte, provider []common.Address) (*ProviderPayoutsProviderRewardAccruedIterator, error) {

	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "ProviderRewardAccrued", sessionIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsProviderRewardAccruedIterator{contract: _ProviderPayouts.contract, event: "ProviderRewardAccrued", logs: logs, sub: sub}, nil
}

// WatchProviderRewardAccrued is a free log subscription operation binding the contract event 0x2e9f0aeeab6a0852b0c976f7e0d74566d7b177b6612d645afd583e6ec4c0b2a6.
//
// Solidity: event ProviderRewardAccrued(bytes32 indexed sessionId, address indexed provider, uint256 accrued, uint256 paid, uint256 owed)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchProviderRewardAccrued(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsProviderRewardAccrued, sessionId [][32]byte, provider []common.Address) (event.Subscription, error) {

	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "ProviderRewardAccrued", sessionIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsProviderRewardAccrued)
				if err := _ProviderPayouts.contract.UnpackLog(event, "ProviderRewardAccrued", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseProviderRewardAccrued(log types.Log) (*ProviderPayoutsProviderRewardAccrued, error) {
	event := new(ProviderPayoutsProviderRewardAccrued)
	if err := _ProviderPayouts.contract.UnpackLog(event, "ProviderRewardAccrued", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ProviderPayoutsSessionRatingIterator is returned from FilterSessionRating and is used to iterate over the raw logs and unpacked data for SessionRating events raised by the ProviderPayouts contract.
type ProviderPayoutsSessionRatingIterator struct {
	Event *ProviderPayoutsSessionRating // Event containing the contract specifics and raw log

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
func (it *ProviderPayoutsSessionRatingIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProviderPayoutsSessionRating)
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
		it.Event = new(ProviderPayoutsSessionRating)
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
func (it *ProviderPayoutsSessionRatingIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProviderPayoutsSessionRatingIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProviderPayoutsSessionRating represents a SessionRating event raised by the ProviderPayouts contract.
type ProviderPayoutsSessionRating struct {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) FilterSessionRating(opts *bind.FilterOpts, sessionId [][32]byte, modelId [][32]byte, provider []common.Address) (*ProviderPayoutsSessionRatingIterator, error) {

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

	logs, sub, err := _ProviderPayouts.contract.FilterLogs(opts, "SessionRating", sessionIdRule, modelIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &ProviderPayoutsSessionRatingIterator{contract: _ProviderPayouts.contract, event: "SessionRating", logs: logs, sub: sub}, nil
}

// WatchSessionRating is a free log subscription operation binding the contract event 0x62fefdc7afa78415878080df4ed9b3c7dbbad510375c160ca7559f9f4d925ecb.
//
// Solidity: event SessionRating(bytes32 indexed sessionId, bytes32 indexed modelId, address indexed provider, uint256 closeoutType, uint32 tpsScaled1000, uint32 ttftMs)
func (_ProviderPayouts *ProviderPayoutsFilterer) WatchSessionRating(opts *bind.WatchOpts, sink chan<- *ProviderPayoutsSessionRating, sessionId [][32]byte, modelId [][32]byte, provider []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _ProviderPayouts.contract.WatchLogs(opts, "SessionRating", sessionIdRule, modelIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProviderPayoutsSessionRating)
				if err := _ProviderPayouts.contract.UnpackLog(event, "SessionRating", log); err != nil {
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
func (_ProviderPayouts *ProviderPayoutsFilterer) ParseSessionRating(log types.Log) (*ProviderPayoutsSessionRating, error) {
	event := new(ProviderPayoutsSessionRating)
	if err := _ProviderPayouts.contract.UnpackLog(event, "SessionRating", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
