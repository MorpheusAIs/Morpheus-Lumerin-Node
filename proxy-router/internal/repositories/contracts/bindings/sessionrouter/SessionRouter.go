// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package sessionrouter

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

// SessionRouterMetaData contains all meta data concerning the SessionRouter contract.
var SessionRouterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"DelegateStakingActiveFundersLimitExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingAllowanceNotActive\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingCapBelowLifetimeFunding\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingFundingCapExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInsufficientCapacity\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInsufficientLiquidity\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInsufficientPrincipal\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidExpiry\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidPoolSettlement\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidWithdrawalCursor\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingLimitExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPoolFunderNotApproved\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPoolSessionAlreadyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPoolSessionNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPrincipalTooLow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingSessionSlotNotEmpty\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingSettlementNotReady\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingWithdrawalNotFound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"delegatee\",\"type\":\"address\"}],\"name\":\"InsufficientRightsForOperation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account_\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionAlreadyClosed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionApproveExpired\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionApprovedForAnotherChainId\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionApprovedForAnotherUser\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionBidNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionDuplicateApproval\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionMaxDurationTooShort\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionNotEndedOrNotExist\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionPoolIndexOutOfBounds\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionPoolSessionCannotBeClosed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionProviderNothingToClaimInThisPeriod\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionProviderSignatureMismatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionReceiptExpired\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionReceiptForAnotherChainId\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionStakeTooLow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionTooShort\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionUserAmountToWithdrawIsZero\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingAllowanceFunded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"cumulativeFundingCap\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"expiry\",\"type\":\"uint128\"}],\"name\":\"DelegateStakingAllowanceGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"}],\"name\":\"DelegateStakingAllowanceRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMaxActiveFundersUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMaxAutoReleaseDaysUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMaxAutoServiceUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMinPrincipalUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolDrawn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"}],\"name\":\"DelegateStakingPoolFunderApprovalRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"}],\"name\":\"DelegateStakingPoolFunderApproved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint128\",\"name\":\"releaseAt\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolLockCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint128\",\"name\":\"releaseAt\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolLockReleased\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolReturned\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"DelegateStakingPoolSessionSettled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingSettlementGraceUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingWithdrawalCancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingWithdrawalPaid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"paid\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"pending\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingWithdrawalRequested\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"storageSlot\",\"type\":\"bytes32\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"payoutTarget\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ProviderDebtPaid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"payoutTarget\",\"type\":\"address\"}],\"name\":\"ProviderPayoutTargetUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"accrued\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"paid\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"owed\",\"type\":\"uint256\"}],\"name\":\"ProviderRewardAccrued\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerId\",\"type\":\"address\"}],\"name\":\"SessionClosed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerId\",\"type\":\"address\"}],\"name\":\"SessionOpened\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"tpsScaled1000\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"ttftMs\",\"type\":\"uint32\"}],\"name\":\"SessionRating\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"}],\"name\":\"UserWithdrawn\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BIDS_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"COMPUTE_POOL_INDEX\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_MARKETPLACE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_MODEL\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_PROVIDER\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_SESSION\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DIAMOND_OWNABLE_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_SESSION_DURATION\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PROVIDERS_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SESSIONS_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SIGNATURE_TTL\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"STATS_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"fundingAccount_\",\"type\":\"address\"},{\"internalType\":\"uint128\",\"name\":\"maxSessionDuration_\",\"type\":\"uint128\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"initialReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardDecrease\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"payoutStart\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"decreaseInterval\",\"type\":\"uint128\"}],\"internalType\":\"structISessionStorage.Pool[]\",\"name\":\"pools_\",\"type\":\"tuple[]\"}],\"name\":\"__SessionRouter_init\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId_\",\"type\":\"bytes32\"}],\"name\":\"claimForProvider\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"receiptEncoded_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature_\",\"type\":\"bytes\"}],\"name\":\"closeSession\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getActiveProviders\",\"outputs\":[{\"internalType\":\"address[]\",\"name\":\"\",\"type\":\"address[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId_\",\"type\":\"bytes32\"}],\"name\":\"getBid\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"deletedAt\",\"type\":\"uint128\"}],\"internalType\":\"structIBidStorage.Bid\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getFundingAccount\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getIsProviderActive\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"approval_\",\"type\":\"bytes\"}],\"name\":\"getIsProviderApprovalUsed\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getMaxSessionDuration\",\"outputs\":[{\"internalType\":\"uint128\",\"name\":\"\",\"type\":\"uint128\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getModelActiveBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getModelBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getModelSessions\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"}],\"name\":\"getModelStats\",\"outputs\":[{\"components\":[{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"tpsScaled1000\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"ttftMs\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"totalDuration\",\"type\":\"tuple\"},{\"internalType\":\"uint32\",\"name\":\"count\",\"type\":\"uint32\"}],\"internalType\":\"structIStatsStorage.ModelStats\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index_\",\"type\":\"uint256\"}],\"name\":\"getPool\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"initialReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardDecrease\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"payoutStart\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"decreaseInterval\",\"type\":\"uint128\"}],\"internalType\":\"structISessionStorage.Pool\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getPools\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"initialReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardDecrease\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"payoutStart\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"decreaseInterval\",\"type\":\"uint128\"}],\"internalType\":\"structISessionStorage.Pool[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getProvider\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"endpoint\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"limitPeriodEnd\",\"type\":\"uint128\"},{\"internalType\":\"uint256\",\"name\":\"limitPeriodEarned\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"internalType\":\"structIProviderStorage.Provider\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getProviderActiveBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getProviderBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getProviderMinimumStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getProviderModelStats\",\"outputs\":[{\"components\":[{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"tpsScaled1000\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"ttftMs\",\"type\":\"tuple\"},{\"internalType\":\"uint32\",\"name\":\"totalDuration\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"successCount\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"totalCount\",\"type\":\"uint32\"}],\"internalType\":\"structIStatsStorage.ProviderModelStats\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getProviderSessions\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getProvidersTotalClaimed\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRegistry\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId_\",\"type\":\"bytes32\"}],\"name\":\"getSession\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"closeoutReceipt\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerWithdrawnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"openedAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"endsAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"closedAt\",\"type\":\"uint128\"},{\"internalType\":\"bool\",\"name\":\"isActive\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"isDirectPaymentFromUser\",\"type\":\"bool\"}],\"internalType\":\"structISessionStorage.Session\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"}],\"name\":\"getTotalSessions\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getUserSessions\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId_\",\"type\":\"bytes32\"}],\"name\":\"isBidActive\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegatee_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"delegator_\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"rights_\",\"type\":\"bytes32\"}],\"name\":\"isRightsDelegated\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isDirectPaymentFromUser_\",\"type\":\"bool\"},{\"internalType\":\"bytes\",\"name\":\"approvalEncoded_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature_\",\"type\":\"bytes\"}],\"name\":\"openSession\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint128\",\"name\":\"maxSessionDuration_\",\"type\":\"uint128\"}],\"name\":\"setMaxSessionDuration\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index_\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"initialReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardDecrease\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"payoutStart\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"decreaseInterval\",\"type\":\"uint128\"}],\"internalType\":\"structISessionStorage.Pool\",\"name\":\"pool_\",\"type\":\"tuple\"}],\"name\":\"setPoolConfig\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user_\",\"type\":\"address\"},{\"internalType\":\"uint8\",\"name\":\"iterations_\",\"type\":\"uint8\"}],\"name\":\"withdrawUserStakes\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// SessionRouterABI is the input ABI used to generate the binding from.
// Deprecated: Use SessionRouterMetaData.ABI instead.
var SessionRouterABI = SessionRouterMetaData.ABI

// SessionRouter is an auto generated Go binding around an Ethereum contract.
type SessionRouter struct {
	SessionRouterCaller     // Read-only binding to the contract
	SessionRouterTransactor // Write-only binding to the contract
	SessionRouterFilterer   // Log filterer for contract events
}

// SessionRouterCaller is an auto generated read-only Go binding around an Ethereum contract.
type SessionRouterCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SessionRouterTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SessionRouterTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SessionRouterFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SessionRouterFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SessionRouterSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SessionRouterSession struct {
	Contract     *SessionRouter    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SessionRouterCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SessionRouterCallerSession struct {
	Contract *SessionRouterCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// SessionRouterTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SessionRouterTransactorSession struct {
	Contract     *SessionRouterTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// SessionRouterRaw is an auto generated low-level Go binding around an Ethereum contract.
type SessionRouterRaw struct {
	Contract *SessionRouter // Generic contract binding to access the raw methods on
}

// SessionRouterCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SessionRouterCallerRaw struct {
	Contract *SessionRouterCaller // Generic read-only contract binding to access the raw methods on
}

// SessionRouterTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SessionRouterTransactorRaw struct {
	Contract *SessionRouterTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSessionRouter creates a new instance of SessionRouter, bound to a specific deployed contract.
func NewSessionRouter(address common.Address, backend bind.ContractBackend) (*SessionRouter, error) {
	contract, err := bindSessionRouter(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SessionRouter{SessionRouterCaller: SessionRouterCaller{contract: contract}, SessionRouterTransactor: SessionRouterTransactor{contract: contract}, SessionRouterFilterer: SessionRouterFilterer{contract: contract}}, nil
}

// NewSessionRouterCaller creates a new read-only instance of SessionRouter, bound to a specific deployed contract.
func NewSessionRouterCaller(address common.Address, caller bind.ContractCaller) (*SessionRouterCaller, error) {
	contract, err := bindSessionRouter(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SessionRouterCaller{contract: contract}, nil
}

// NewSessionRouterTransactor creates a new write-only instance of SessionRouter, bound to a specific deployed contract.
func NewSessionRouterTransactor(address common.Address, transactor bind.ContractTransactor) (*SessionRouterTransactor, error) {
	contract, err := bindSessionRouter(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SessionRouterTransactor{contract: contract}, nil
}

// NewSessionRouterFilterer creates a new log filterer instance of SessionRouter, bound to a specific deployed contract.
func NewSessionRouterFilterer(address common.Address, filterer bind.ContractFilterer) (*SessionRouterFilterer, error) {
	contract, err := bindSessionRouter(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SessionRouterFilterer{contract: contract}, nil
}

// bindSessionRouter binds a generic wrapper to an already deployed contract.
func bindSessionRouter(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SessionRouterMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SessionRouter *SessionRouterRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SessionRouter.Contract.SessionRouterCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SessionRouter *SessionRouterRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SessionRouter.Contract.SessionRouterTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SessionRouter *SessionRouterRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SessionRouter.Contract.SessionRouterTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SessionRouter *SessionRouterCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SessionRouter.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SessionRouter *SessionRouterTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SessionRouter.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SessionRouter *SessionRouterTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SessionRouter.Contract.contract.Transact(opts, method, params...)
}

// BIDSSTORAGESLOT is a free data retrieval call binding the contract method 0x266ccff0.
//
// Solidity: function BIDS_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCaller) BIDSSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "BIDS_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BIDSSTORAGESLOT is a free data retrieval call binding the contract method 0x266ccff0.
//
// Solidity: function BIDS_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterSession) BIDSSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.BIDSSTORAGESLOT(&_SessionRouter.CallOpts)
}

// BIDSSTORAGESLOT is a free data retrieval call binding the contract method 0x266ccff0.
//
// Solidity: function BIDS_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) BIDSSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.BIDSSTORAGESLOT(&_SessionRouter.CallOpts)
}

// COMPUTEPOOLINDEX is a free data retrieval call binding the contract method 0xc56d09a0.
//
// Solidity: function COMPUTE_POOL_INDEX() view returns(uint256)
func (_SessionRouter *SessionRouterCaller) COMPUTEPOOLINDEX(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "COMPUTE_POOL_INDEX")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// COMPUTEPOOLINDEX is a free data retrieval call binding the contract method 0xc56d09a0.
//
// Solidity: function COMPUTE_POOL_INDEX() view returns(uint256)
func (_SessionRouter *SessionRouterSession) COMPUTEPOOLINDEX() (*big.Int, error) {
	return _SessionRouter.Contract.COMPUTEPOOLINDEX(&_SessionRouter.CallOpts)
}

// COMPUTEPOOLINDEX is a free data retrieval call binding the contract method 0xc56d09a0.
//
// Solidity: function COMPUTE_POOL_INDEX() view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) COMPUTEPOOLINDEX() (*big.Int, error) {
	return _SessionRouter.Contract.COMPUTEPOOLINDEX(&_SessionRouter.CallOpts)
}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_SessionRouter *SessionRouterCaller) DELEGATIONRULESMARKETPLACE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "DELEGATION_RULES_MARKETPLACE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_SessionRouter *SessionRouterSession) DELEGATIONRULESMARKETPLACE() ([32]byte, error) {
	return _SessionRouter.Contract.DELEGATIONRULESMARKETPLACE(&_SessionRouter.CallOpts)
}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) DELEGATIONRULESMARKETPLACE() ([32]byte, error) {
	return _SessionRouter.Contract.DELEGATIONRULESMARKETPLACE(&_SessionRouter.CallOpts)
}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_SessionRouter *SessionRouterCaller) DELEGATIONRULESMODEL(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "DELEGATION_RULES_MODEL")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_SessionRouter *SessionRouterSession) DELEGATIONRULESMODEL() ([32]byte, error) {
	return _SessionRouter.Contract.DELEGATIONRULESMODEL(&_SessionRouter.CallOpts)
}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) DELEGATIONRULESMODEL() ([32]byte, error) {
	return _SessionRouter.Contract.DELEGATIONRULESMODEL(&_SessionRouter.CallOpts)
}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_SessionRouter *SessionRouterCaller) DELEGATIONRULESPROVIDER(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "DELEGATION_RULES_PROVIDER")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_SessionRouter *SessionRouterSession) DELEGATIONRULESPROVIDER() ([32]byte, error) {
	return _SessionRouter.Contract.DELEGATIONRULESPROVIDER(&_SessionRouter.CallOpts)
}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) DELEGATIONRULESPROVIDER() ([32]byte, error) {
	return _SessionRouter.Contract.DELEGATIONRULESPROVIDER(&_SessionRouter.CallOpts)
}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_SessionRouter *SessionRouterCaller) DELEGATIONRULESSESSION(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "DELEGATION_RULES_SESSION")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_SessionRouter *SessionRouterSession) DELEGATIONRULESSESSION() ([32]byte, error) {
	return _SessionRouter.Contract.DELEGATIONRULESSESSION(&_SessionRouter.CallOpts)
}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) DELEGATIONRULESSESSION() ([32]byte, error) {
	return _SessionRouter.Contract.DELEGATIONRULESSESSION(&_SessionRouter.CallOpts)
}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCaller) DELEGATIONSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "DELEGATION_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterSession) DELEGATIONSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.DELEGATIONSTORAGESLOT(&_SessionRouter.CallOpts)
}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) DELEGATIONSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.DELEGATIONSTORAGESLOT(&_SessionRouter.CallOpts)
}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCaller) DIAMONDOWNABLESTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "DIAMOND_OWNABLE_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterSession) DIAMONDOWNABLESTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.DIAMONDOWNABLESTORAGESLOT(&_SessionRouter.CallOpts)
}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) DIAMONDOWNABLESTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.DIAMONDOWNABLESTORAGESLOT(&_SessionRouter.CallOpts)
}

// MINSESSIONDURATION is a free data retrieval call binding the contract method 0x7d980286.
//
// Solidity: function MIN_SESSION_DURATION() view returns(uint32)
func (_SessionRouter *SessionRouterCaller) MINSESSIONDURATION(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "MIN_SESSION_DURATION")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// MINSESSIONDURATION is a free data retrieval call binding the contract method 0x7d980286.
//
// Solidity: function MIN_SESSION_DURATION() view returns(uint32)
func (_SessionRouter *SessionRouterSession) MINSESSIONDURATION() (uint32, error) {
	return _SessionRouter.Contract.MINSESSIONDURATION(&_SessionRouter.CallOpts)
}

// MINSESSIONDURATION is a free data retrieval call binding the contract method 0x7d980286.
//
// Solidity: function MIN_SESSION_DURATION() view returns(uint32)
func (_SessionRouter *SessionRouterCallerSession) MINSESSIONDURATION() (uint32, error) {
	return _SessionRouter.Contract.MINSESSIONDURATION(&_SessionRouter.CallOpts)
}

// PROVIDERSSTORAGESLOT is a free data retrieval call binding the contract method 0xc51830f6.
//
// Solidity: function PROVIDERS_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCaller) PROVIDERSSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "PROVIDERS_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PROVIDERSSTORAGESLOT is a free data retrieval call binding the contract method 0xc51830f6.
//
// Solidity: function PROVIDERS_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterSession) PROVIDERSSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.PROVIDERSSTORAGESLOT(&_SessionRouter.CallOpts)
}

// PROVIDERSSTORAGESLOT is a free data retrieval call binding the contract method 0xc51830f6.
//
// Solidity: function PROVIDERS_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) PROVIDERSSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.PROVIDERSSTORAGESLOT(&_SessionRouter.CallOpts)
}

// SESSIONSSTORAGESLOT is a free data retrieval call binding the contract method 0xb392636e.
//
// Solidity: function SESSIONS_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCaller) SESSIONSSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "SESSIONS_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SESSIONSSTORAGESLOT is a free data retrieval call binding the contract method 0xb392636e.
//
// Solidity: function SESSIONS_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterSession) SESSIONSSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.SESSIONSSTORAGESLOT(&_SessionRouter.CallOpts)
}

// SESSIONSSTORAGESLOT is a free data retrieval call binding the contract method 0xb392636e.
//
// Solidity: function SESSIONS_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) SESSIONSSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.SESSIONSSTORAGESLOT(&_SessionRouter.CallOpts)
}

// SIGNATURETTL is a free data retrieval call binding the contract method 0xe7d791d0.
//
// Solidity: function SIGNATURE_TTL() view returns(uint32)
func (_SessionRouter *SessionRouterCaller) SIGNATURETTL(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "SIGNATURE_TTL")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// SIGNATURETTL is a free data retrieval call binding the contract method 0xe7d791d0.
//
// Solidity: function SIGNATURE_TTL() view returns(uint32)
func (_SessionRouter *SessionRouterSession) SIGNATURETTL() (uint32, error) {
	return _SessionRouter.Contract.SIGNATURETTL(&_SessionRouter.CallOpts)
}

// SIGNATURETTL is a free data retrieval call binding the contract method 0xe7d791d0.
//
// Solidity: function SIGNATURE_TTL() view returns(uint32)
func (_SessionRouter *SessionRouterCallerSession) SIGNATURETTL() (uint32, error) {
	return _SessionRouter.Contract.SIGNATURETTL(&_SessionRouter.CallOpts)
}

// STATSSTORAGESLOT is a free data retrieval call binding the contract method 0x87040789.
//
// Solidity: function STATS_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCaller) STATSSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "STATS_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// STATSSTORAGESLOT is a free data retrieval call binding the contract method 0x87040789.
//
// Solidity: function STATS_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterSession) STATSSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.STATSSTORAGESLOT(&_SessionRouter.CallOpts)
}

// STATSSTORAGESLOT is a free data retrieval call binding the contract method 0x87040789.
//
// Solidity: function STATS_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) STATSSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.STATSSTORAGESLOT(&_SessionRouter.CallOpts)
}

// GetActiveProviders is a free data retrieval call binding the contract method 0xd5472642.
//
// Solidity: function getActiveProviders(uint256 offset_, uint256 limit_) view returns(address[], uint256)
func (_SessionRouter *SessionRouterCaller) GetActiveProviders(opts *bind.CallOpts, offset_ *big.Int, limit_ *big.Int) ([]common.Address, *big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getActiveProviders", offset_, limit_)

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
func (_SessionRouter *SessionRouterSession) GetActiveProviders(offset_ *big.Int, limit_ *big.Int) ([]common.Address, *big.Int, error) {
	return _SessionRouter.Contract.GetActiveProviders(&_SessionRouter.CallOpts, offset_, limit_)
}

// GetActiveProviders is a free data retrieval call binding the contract method 0xd5472642.
//
// Solidity: function getActiveProviders(uint256 offset_, uint256 limit_) view returns(address[], uint256)
func (_SessionRouter *SessionRouterCallerSession) GetActiveProviders(offset_ *big.Int, limit_ *big.Int) ([]common.Address, *big.Int, error) {
	return _SessionRouter.Contract.GetActiveProviders(&_SessionRouter.CallOpts, offset_, limit_)
}

// GetBid is a free data retrieval call binding the contract method 0x91704e1e.
//
// Solidity: function getBid(bytes32 bidId_) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_SessionRouter *SessionRouterCaller) GetBid(opts *bind.CallOpts, bidId_ [32]byte) (IBidStorageBid, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getBid", bidId_)

	if err != nil {
		return *new(IBidStorageBid), err
	}

	out0 := *abi.ConvertType(out[0], new(IBidStorageBid)).(*IBidStorageBid)

	return out0, err

}

// GetBid is a free data retrieval call binding the contract method 0x91704e1e.
//
// Solidity: function getBid(bytes32 bidId_) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_SessionRouter *SessionRouterSession) GetBid(bidId_ [32]byte) (IBidStorageBid, error) {
	return _SessionRouter.Contract.GetBid(&_SessionRouter.CallOpts, bidId_)
}

// GetBid is a free data retrieval call binding the contract method 0x91704e1e.
//
// Solidity: function getBid(bytes32 bidId_) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_SessionRouter *SessionRouterCallerSession) GetBid(bidId_ [32]byte) (IBidStorageBid, error) {
	return _SessionRouter.Contract.GetBid(&_SessionRouter.CallOpts, bidId_)
}

// GetFundingAccount is a free data retrieval call binding the contract method 0x775c3727.
//
// Solidity: function getFundingAccount() view returns(address)
func (_SessionRouter *SessionRouterCaller) GetFundingAccount(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getFundingAccount")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetFundingAccount is a free data retrieval call binding the contract method 0x775c3727.
//
// Solidity: function getFundingAccount() view returns(address)
func (_SessionRouter *SessionRouterSession) GetFundingAccount() (common.Address, error) {
	return _SessionRouter.Contract.GetFundingAccount(&_SessionRouter.CallOpts)
}

// GetFundingAccount is a free data retrieval call binding the contract method 0x775c3727.
//
// Solidity: function getFundingAccount() view returns(address)
func (_SessionRouter *SessionRouterCallerSession) GetFundingAccount() (common.Address, error) {
	return _SessionRouter.Contract.GetFundingAccount(&_SessionRouter.CallOpts)
}

// GetIsProviderActive is a free data retrieval call binding the contract method 0x63ef175d.
//
// Solidity: function getIsProviderActive(address provider_) view returns(bool)
func (_SessionRouter *SessionRouterCaller) GetIsProviderActive(opts *bind.CallOpts, provider_ common.Address) (bool, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getIsProviderActive", provider_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// GetIsProviderActive is a free data retrieval call binding the contract method 0x63ef175d.
//
// Solidity: function getIsProviderActive(address provider_) view returns(bool)
func (_SessionRouter *SessionRouterSession) GetIsProviderActive(provider_ common.Address) (bool, error) {
	return _SessionRouter.Contract.GetIsProviderActive(&_SessionRouter.CallOpts, provider_)
}

// GetIsProviderActive is a free data retrieval call binding the contract method 0x63ef175d.
//
// Solidity: function getIsProviderActive(address provider_) view returns(bool)
func (_SessionRouter *SessionRouterCallerSession) GetIsProviderActive(provider_ common.Address) (bool, error) {
	return _SessionRouter.Contract.GetIsProviderActive(&_SessionRouter.CallOpts, provider_)
}

// GetIsProviderApprovalUsed is a free data retrieval call binding the contract method 0xdb1cf1e2.
//
// Solidity: function getIsProviderApprovalUsed(bytes approval_) view returns(bool)
func (_SessionRouter *SessionRouterCaller) GetIsProviderApprovalUsed(opts *bind.CallOpts, approval_ []byte) (bool, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getIsProviderApprovalUsed", approval_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// GetIsProviderApprovalUsed is a free data retrieval call binding the contract method 0xdb1cf1e2.
//
// Solidity: function getIsProviderApprovalUsed(bytes approval_) view returns(bool)
func (_SessionRouter *SessionRouterSession) GetIsProviderApprovalUsed(approval_ []byte) (bool, error) {
	return _SessionRouter.Contract.GetIsProviderApprovalUsed(&_SessionRouter.CallOpts, approval_)
}

// GetIsProviderApprovalUsed is a free data retrieval call binding the contract method 0xdb1cf1e2.
//
// Solidity: function getIsProviderApprovalUsed(bytes approval_) view returns(bool)
func (_SessionRouter *SessionRouterCallerSession) GetIsProviderApprovalUsed(approval_ []byte) (bool, error) {
	return _SessionRouter.Contract.GetIsProviderApprovalUsed(&_SessionRouter.CallOpts, approval_)
}

// GetMaxSessionDuration is a free data retrieval call binding the contract method 0xa9756858.
//
// Solidity: function getMaxSessionDuration() view returns(uint128)
func (_SessionRouter *SessionRouterCaller) GetMaxSessionDuration(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getMaxSessionDuration")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetMaxSessionDuration is a free data retrieval call binding the contract method 0xa9756858.
//
// Solidity: function getMaxSessionDuration() view returns(uint128)
func (_SessionRouter *SessionRouterSession) GetMaxSessionDuration() (*big.Int, error) {
	return _SessionRouter.Contract.GetMaxSessionDuration(&_SessionRouter.CallOpts)
}

// GetMaxSessionDuration is a free data retrieval call binding the contract method 0xa9756858.
//
// Solidity: function getMaxSessionDuration() view returns(uint128)
func (_SessionRouter *SessionRouterCallerSession) GetMaxSessionDuration() (*big.Int, error) {
	return _SessionRouter.Contract.GetMaxSessionDuration(&_SessionRouter.CallOpts)
}

// GetModelActiveBids is a free data retrieval call binding the contract method 0x8a683b6e.
//
// Solidity: function getModelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_SessionRouter *SessionRouterCaller) GetModelActiveBids(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getModelActiveBids", modelId_, offset_, limit_)

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
func (_SessionRouter *SessionRouterSession) GetModelActiveBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _SessionRouter.Contract.GetModelActiveBids(&_SessionRouter.CallOpts, modelId_, offset_, limit_)
}

// GetModelActiveBids is a free data retrieval call binding the contract method 0x8a683b6e.
//
// Solidity: function getModelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_SessionRouter *SessionRouterCallerSession) GetModelActiveBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _SessionRouter.Contract.GetModelActiveBids(&_SessionRouter.CallOpts, modelId_, offset_, limit_)
}

// GetModelBids is a free data retrieval call binding the contract method 0xfade17b1.
//
// Solidity: function getModelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_SessionRouter *SessionRouterCaller) GetModelBids(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getModelBids", modelId_, offset_, limit_)

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
func (_SessionRouter *SessionRouterSession) GetModelBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _SessionRouter.Contract.GetModelBids(&_SessionRouter.CallOpts, modelId_, offset_, limit_)
}

// GetModelBids is a free data retrieval call binding the contract method 0xfade17b1.
//
// Solidity: function getModelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_SessionRouter *SessionRouterCallerSession) GetModelBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _SessionRouter.Contract.GetModelBids(&_SessionRouter.CallOpts, modelId_, offset_, limit_)
}

// GetModelSessions is a free data retrieval call binding the contract method 0x1d78a872.
//
// Solidity: function getModelSessions(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_SessionRouter *SessionRouterCaller) GetModelSessions(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getModelSessions", modelId_, offset_, limit_)

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
func (_SessionRouter *SessionRouterSession) GetModelSessions(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _SessionRouter.Contract.GetModelSessions(&_SessionRouter.CallOpts, modelId_, offset_, limit_)
}

// GetModelSessions is a free data retrieval call binding the contract method 0x1d78a872.
//
// Solidity: function getModelSessions(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_SessionRouter *SessionRouterCallerSession) GetModelSessions(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _SessionRouter.Contract.GetModelSessions(&_SessionRouter.CallOpts, modelId_, offset_, limit_)
}

// GetModelStats is a free data retrieval call binding the contract method 0xce535723.
//
// Solidity: function getModelStats(bytes32 modelId_) view returns(((int64,int64),(int64,int64),(int64,int64),uint32))
func (_SessionRouter *SessionRouterCaller) GetModelStats(opts *bind.CallOpts, modelId_ [32]byte) (IStatsStorageModelStats, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getModelStats", modelId_)

	if err != nil {
		return *new(IStatsStorageModelStats), err
	}

	out0 := *abi.ConvertType(out[0], new(IStatsStorageModelStats)).(*IStatsStorageModelStats)

	return out0, err

}

// GetModelStats is a free data retrieval call binding the contract method 0xce535723.
//
// Solidity: function getModelStats(bytes32 modelId_) view returns(((int64,int64),(int64,int64),(int64,int64),uint32))
func (_SessionRouter *SessionRouterSession) GetModelStats(modelId_ [32]byte) (IStatsStorageModelStats, error) {
	return _SessionRouter.Contract.GetModelStats(&_SessionRouter.CallOpts, modelId_)
}

// GetModelStats is a free data retrieval call binding the contract method 0xce535723.
//
// Solidity: function getModelStats(bytes32 modelId_) view returns(((int64,int64),(int64,int64),(int64,int64),uint32))
func (_SessionRouter *SessionRouterCallerSession) GetModelStats(modelId_ [32]byte) (IStatsStorageModelStats, error) {
	return _SessionRouter.Contract.GetModelStats(&_SessionRouter.CallOpts, modelId_)
}

// GetPool is a free data retrieval call binding the contract method 0x068bcd8d.
//
// Solidity: function getPool(uint256 index_) view returns((uint256,uint256,uint128,uint128))
func (_SessionRouter *SessionRouterCaller) GetPool(opts *bind.CallOpts, index_ *big.Int) (ISessionStoragePool, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getPool", index_)

	if err != nil {
		return *new(ISessionStoragePool), err
	}

	out0 := *abi.ConvertType(out[0], new(ISessionStoragePool)).(*ISessionStoragePool)

	return out0, err

}

// GetPool is a free data retrieval call binding the contract method 0x068bcd8d.
//
// Solidity: function getPool(uint256 index_) view returns((uint256,uint256,uint128,uint128))
func (_SessionRouter *SessionRouterSession) GetPool(index_ *big.Int) (ISessionStoragePool, error) {
	return _SessionRouter.Contract.GetPool(&_SessionRouter.CallOpts, index_)
}

// GetPool is a free data retrieval call binding the contract method 0x068bcd8d.
//
// Solidity: function getPool(uint256 index_) view returns((uint256,uint256,uint128,uint128))
func (_SessionRouter *SessionRouterCallerSession) GetPool(index_ *big.Int) (ISessionStoragePool, error) {
	return _SessionRouter.Contract.GetPool(&_SessionRouter.CallOpts, index_)
}

// GetPools is a free data retrieval call binding the contract method 0x673a2a1f.
//
// Solidity: function getPools() view returns((uint256,uint256,uint128,uint128)[])
func (_SessionRouter *SessionRouterCaller) GetPools(opts *bind.CallOpts) ([]ISessionStoragePool, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getPools")

	if err != nil {
		return *new([]ISessionStoragePool), err
	}

	out0 := *abi.ConvertType(out[0], new([]ISessionStoragePool)).(*[]ISessionStoragePool)

	return out0, err

}

// GetPools is a free data retrieval call binding the contract method 0x673a2a1f.
//
// Solidity: function getPools() view returns((uint256,uint256,uint128,uint128)[])
func (_SessionRouter *SessionRouterSession) GetPools() ([]ISessionStoragePool, error) {
	return _SessionRouter.Contract.GetPools(&_SessionRouter.CallOpts)
}

// GetPools is a free data retrieval call binding the contract method 0x673a2a1f.
//
// Solidity: function getPools() view returns((uint256,uint256,uint128,uint128)[])
func (_SessionRouter *SessionRouterCallerSession) GetPools() ([]ISessionStoragePool, error) {
	return _SessionRouter.Contract.GetPools(&_SessionRouter.CallOpts)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider_) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_SessionRouter *SessionRouterCaller) GetProvider(opts *bind.CallOpts, provider_ common.Address) (IProviderStorageProvider, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getProvider", provider_)

	if err != nil {
		return *new(IProviderStorageProvider), err
	}

	out0 := *abi.ConvertType(out[0], new(IProviderStorageProvider)).(*IProviderStorageProvider)

	return out0, err

}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider_) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_SessionRouter *SessionRouterSession) GetProvider(provider_ common.Address) (IProviderStorageProvider, error) {
	return _SessionRouter.Contract.GetProvider(&_SessionRouter.CallOpts, provider_)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider_) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_SessionRouter *SessionRouterCallerSession) GetProvider(provider_ common.Address) (IProviderStorageProvider, error) {
	return _SessionRouter.Contract.GetProvider(&_SessionRouter.CallOpts, provider_)
}

// GetProviderActiveBids is a free data retrieval call binding the contract method 0xaf5b77ca.
//
// Solidity: function getProviderActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_SessionRouter *SessionRouterCaller) GetProviderActiveBids(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getProviderActiveBids", provider_, offset_, limit_)

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
func (_SessionRouter *SessionRouterSession) GetProviderActiveBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _SessionRouter.Contract.GetProviderActiveBids(&_SessionRouter.CallOpts, provider_, offset_, limit_)
}

// GetProviderActiveBids is a free data retrieval call binding the contract method 0xaf5b77ca.
//
// Solidity: function getProviderActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_SessionRouter *SessionRouterCallerSession) GetProviderActiveBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _SessionRouter.Contract.GetProviderActiveBids(&_SessionRouter.CallOpts, provider_, offset_, limit_)
}

// GetProviderBids is a free data retrieval call binding the contract method 0x59d435c4.
//
// Solidity: function getProviderBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_SessionRouter *SessionRouterCaller) GetProviderBids(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getProviderBids", provider_, offset_, limit_)

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
func (_SessionRouter *SessionRouterSession) GetProviderBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _SessionRouter.Contract.GetProviderBids(&_SessionRouter.CallOpts, provider_, offset_, limit_)
}

// GetProviderBids is a free data retrieval call binding the contract method 0x59d435c4.
//
// Solidity: function getProviderBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_SessionRouter *SessionRouterCallerSession) GetProviderBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _SessionRouter.Contract.GetProviderBids(&_SessionRouter.CallOpts, provider_, offset_, limit_)
}

// GetProviderMinimumStake is a free data retrieval call binding the contract method 0x53c029f6.
//
// Solidity: function getProviderMinimumStake() view returns(uint256)
func (_SessionRouter *SessionRouterCaller) GetProviderMinimumStake(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getProviderMinimumStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProviderMinimumStake is a free data retrieval call binding the contract method 0x53c029f6.
//
// Solidity: function getProviderMinimumStake() view returns(uint256)
func (_SessionRouter *SessionRouterSession) GetProviderMinimumStake() (*big.Int, error) {
	return _SessionRouter.Contract.GetProviderMinimumStake(&_SessionRouter.CallOpts)
}

// GetProviderMinimumStake is a free data retrieval call binding the contract method 0x53c029f6.
//
// Solidity: function getProviderMinimumStake() view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) GetProviderMinimumStake() (*big.Int, error) {
	return _SessionRouter.Contract.GetProviderMinimumStake(&_SessionRouter.CallOpts)
}

// GetProviderModelStats is a free data retrieval call binding the contract method 0x1b26c116.
//
// Solidity: function getProviderModelStats(bytes32 modelId_, address provider_) view returns(((int64,int64),(int64,int64),uint32,uint32,uint32))
func (_SessionRouter *SessionRouterCaller) GetProviderModelStats(opts *bind.CallOpts, modelId_ [32]byte, provider_ common.Address) (IStatsStorageProviderModelStats, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getProviderModelStats", modelId_, provider_)

	if err != nil {
		return *new(IStatsStorageProviderModelStats), err
	}

	out0 := *abi.ConvertType(out[0], new(IStatsStorageProviderModelStats)).(*IStatsStorageProviderModelStats)

	return out0, err

}

// GetProviderModelStats is a free data retrieval call binding the contract method 0x1b26c116.
//
// Solidity: function getProviderModelStats(bytes32 modelId_, address provider_) view returns(((int64,int64),(int64,int64),uint32,uint32,uint32))
func (_SessionRouter *SessionRouterSession) GetProviderModelStats(modelId_ [32]byte, provider_ common.Address) (IStatsStorageProviderModelStats, error) {
	return _SessionRouter.Contract.GetProviderModelStats(&_SessionRouter.CallOpts, modelId_, provider_)
}

// GetProviderModelStats is a free data retrieval call binding the contract method 0x1b26c116.
//
// Solidity: function getProviderModelStats(bytes32 modelId_, address provider_) view returns(((int64,int64),(int64,int64),uint32,uint32,uint32))
func (_SessionRouter *SessionRouterCallerSession) GetProviderModelStats(modelId_ [32]byte, provider_ common.Address) (IStatsStorageProviderModelStats, error) {
	return _SessionRouter.Contract.GetProviderModelStats(&_SessionRouter.CallOpts, modelId_, provider_)
}

// GetProviderSessions is a free data retrieval call binding the contract method 0x87bced7d.
//
// Solidity: function getProviderSessions(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_SessionRouter *SessionRouterCaller) GetProviderSessions(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getProviderSessions", provider_, offset_, limit_)

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
func (_SessionRouter *SessionRouterSession) GetProviderSessions(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _SessionRouter.Contract.GetProviderSessions(&_SessionRouter.CallOpts, provider_, offset_, limit_)
}

// GetProviderSessions is a free data retrieval call binding the contract method 0x87bced7d.
//
// Solidity: function getProviderSessions(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_SessionRouter *SessionRouterCallerSession) GetProviderSessions(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _SessionRouter.Contract.GetProviderSessions(&_SessionRouter.CallOpts, provider_, offset_, limit_)
}

// GetProvidersTotalClaimed is a free data retrieval call binding the contract method 0xbdfbbada.
//
// Solidity: function getProvidersTotalClaimed() view returns(uint256)
func (_SessionRouter *SessionRouterCaller) GetProvidersTotalClaimed(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getProvidersTotalClaimed")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProvidersTotalClaimed is a free data retrieval call binding the contract method 0xbdfbbada.
//
// Solidity: function getProvidersTotalClaimed() view returns(uint256)
func (_SessionRouter *SessionRouterSession) GetProvidersTotalClaimed() (*big.Int, error) {
	return _SessionRouter.Contract.GetProvidersTotalClaimed(&_SessionRouter.CallOpts)
}

// GetProvidersTotalClaimed is a free data retrieval call binding the contract method 0xbdfbbada.
//
// Solidity: function getProvidersTotalClaimed() view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) GetProvidersTotalClaimed() (*big.Int, error) {
	return _SessionRouter.Contract.GetProvidersTotalClaimed(&_SessionRouter.CallOpts)
}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_SessionRouter *SessionRouterCaller) GetRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_SessionRouter *SessionRouterSession) GetRegistry() (common.Address, error) {
	return _SessionRouter.Contract.GetRegistry(&_SessionRouter.CallOpts)
}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_SessionRouter *SessionRouterCallerSession) GetRegistry() (common.Address, error) {
	return _SessionRouter.Contract.GetRegistry(&_SessionRouter.CallOpts)
}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId_) view returns((address,bytes32,uint256,bytes,uint256,uint256,uint128,uint128,uint128,bool,bool))
func (_SessionRouter *SessionRouterCaller) GetSession(opts *bind.CallOpts, sessionId_ [32]byte) (ISessionStorageSession, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getSession", sessionId_)

	if err != nil {
		return *new(ISessionStorageSession), err
	}

	out0 := *abi.ConvertType(out[0], new(ISessionStorageSession)).(*ISessionStorageSession)

	return out0, err

}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId_) view returns((address,bytes32,uint256,bytes,uint256,uint256,uint128,uint128,uint128,bool,bool))
func (_SessionRouter *SessionRouterSession) GetSession(sessionId_ [32]byte) (ISessionStorageSession, error) {
	return _SessionRouter.Contract.GetSession(&_SessionRouter.CallOpts, sessionId_)
}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId_) view returns((address,bytes32,uint256,bytes,uint256,uint256,uint128,uint128,uint128,bool,bool))
func (_SessionRouter *SessionRouterCallerSession) GetSession(sessionId_ [32]byte) (ISessionStorageSession, error) {
	return _SessionRouter.Contract.GetSession(&_SessionRouter.CallOpts, sessionId_)
}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_SessionRouter *SessionRouterCaller) GetToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_SessionRouter *SessionRouterSession) GetToken() (common.Address, error) {
	return _SessionRouter.Contract.GetToken(&_SessionRouter.CallOpts)
}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_SessionRouter *SessionRouterCallerSession) GetToken() (common.Address, error) {
	return _SessionRouter.Contract.GetToken(&_SessionRouter.CallOpts)
}

// GetTotalSessions is a free data retrieval call binding the contract method 0x76400935.
//
// Solidity: function getTotalSessions(address provider_) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) GetTotalSessions(opts *bind.CallOpts, provider_ common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getTotalSessions", provider_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTotalSessions is a free data retrieval call binding the contract method 0x76400935.
//
// Solidity: function getTotalSessions(address provider_) view returns(uint256)
func (_SessionRouter *SessionRouterSession) GetTotalSessions(provider_ common.Address) (*big.Int, error) {
	return _SessionRouter.Contract.GetTotalSessions(&_SessionRouter.CallOpts, provider_)
}

// GetTotalSessions is a free data retrieval call binding the contract method 0x76400935.
//
// Solidity: function getTotalSessions(address provider_) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) GetTotalSessions(provider_ common.Address) (*big.Int, error) {
	return _SessionRouter.Contract.GetTotalSessions(&_SessionRouter.CallOpts, provider_)
}

// GetUserSessions is a free data retrieval call binding the contract method 0xeb7764bb.
//
// Solidity: function getUserSessions(address user_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_SessionRouter *SessionRouterCaller) GetUserSessions(opts *bind.CallOpts, user_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getUserSessions", user_, offset_, limit_)

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
func (_SessionRouter *SessionRouterSession) GetUserSessions(user_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _SessionRouter.Contract.GetUserSessions(&_SessionRouter.CallOpts, user_, offset_, limit_)
}

// GetUserSessions is a free data retrieval call binding the contract method 0xeb7764bb.
//
// Solidity: function getUserSessions(address user_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_SessionRouter *SessionRouterCallerSession) GetUserSessions(user_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _SessionRouter.Contract.GetUserSessions(&_SessionRouter.CallOpts, user_, offset_, limit_)
}

// IsBidActive is a free data retrieval call binding the contract method 0x1345df58.
//
// Solidity: function isBidActive(bytes32 bidId_) view returns(bool)
func (_SessionRouter *SessionRouterCaller) IsBidActive(opts *bind.CallOpts, bidId_ [32]byte) (bool, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "isBidActive", bidId_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsBidActive is a free data retrieval call binding the contract method 0x1345df58.
//
// Solidity: function isBidActive(bytes32 bidId_) view returns(bool)
func (_SessionRouter *SessionRouterSession) IsBidActive(bidId_ [32]byte) (bool, error) {
	return _SessionRouter.Contract.IsBidActive(&_SessionRouter.CallOpts, bidId_)
}

// IsBidActive is a free data retrieval call binding the contract method 0x1345df58.
//
// Solidity: function isBidActive(bytes32 bidId_) view returns(bool)
func (_SessionRouter *SessionRouterCallerSession) IsBidActive(bidId_ [32]byte) (bool, error) {
	return _SessionRouter.Contract.IsBidActive(&_SessionRouter.CallOpts, bidId_)
}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_SessionRouter *SessionRouterCaller) IsRightsDelegated(opts *bind.CallOpts, delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "isRightsDelegated", delegatee_, delegator_, rights_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_SessionRouter *SessionRouterSession) IsRightsDelegated(delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	return _SessionRouter.Contract.IsRightsDelegated(&_SessionRouter.CallOpts, delegatee_, delegator_, rights_)
}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_SessionRouter *SessionRouterCallerSession) IsRightsDelegated(delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	return _SessionRouter.Contract.IsRightsDelegated(&_SessionRouter.CallOpts, delegatee_, delegator_, rights_)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SessionRouter *SessionRouterCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SessionRouter *SessionRouterSession) Owner() (common.Address, error) {
	return _SessionRouter.Contract.Owner(&_SessionRouter.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SessionRouter *SessionRouterCallerSession) Owner() (common.Address, error) {
	return _SessionRouter.Contract.Owner(&_SessionRouter.CallOpts)
}

// SessionRouterInit is a paid mutator transaction binding the contract method 0xf8ba944b.
//
// Solidity: function __SessionRouter_init(address fundingAccount_, uint128 maxSessionDuration_, (uint256,uint256,uint128,uint128)[] pools_) returns()
func (_SessionRouter *SessionRouterTransactor) SessionRouterInit(opts *bind.TransactOpts, fundingAccount_ common.Address, maxSessionDuration_ *big.Int, pools_ []ISessionStoragePool) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "__SessionRouter_init", fundingAccount_, maxSessionDuration_, pools_)
}

// SessionRouterInit is a paid mutator transaction binding the contract method 0xf8ba944b.
//
// Solidity: function __SessionRouter_init(address fundingAccount_, uint128 maxSessionDuration_, (uint256,uint256,uint128,uint128)[] pools_) returns()
func (_SessionRouter *SessionRouterSession) SessionRouterInit(fundingAccount_ common.Address, maxSessionDuration_ *big.Int, pools_ []ISessionStoragePool) (*types.Transaction, error) {
	return _SessionRouter.Contract.SessionRouterInit(&_SessionRouter.TransactOpts, fundingAccount_, maxSessionDuration_, pools_)
}

// SessionRouterInit is a paid mutator transaction binding the contract method 0xf8ba944b.
//
// Solidity: function __SessionRouter_init(address fundingAccount_, uint128 maxSessionDuration_, (uint256,uint256,uint128,uint128)[] pools_) returns()
func (_SessionRouter *SessionRouterTransactorSession) SessionRouterInit(fundingAccount_ common.Address, maxSessionDuration_ *big.Int, pools_ []ISessionStoragePool) (*types.Transaction, error) {
	return _SessionRouter.Contract.SessionRouterInit(&_SessionRouter.TransactOpts, fundingAccount_, maxSessionDuration_, pools_)
}

// ClaimForProvider is a paid mutator transaction binding the contract method 0xf2e96e8e.
//
// Solidity: function claimForProvider(bytes32 sessionId_) returns()
func (_SessionRouter *SessionRouterTransactor) ClaimForProvider(opts *bind.TransactOpts, sessionId_ [32]byte) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "claimForProvider", sessionId_)
}

// ClaimForProvider is a paid mutator transaction binding the contract method 0xf2e96e8e.
//
// Solidity: function claimForProvider(bytes32 sessionId_) returns()
func (_SessionRouter *SessionRouterSession) ClaimForProvider(sessionId_ [32]byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.ClaimForProvider(&_SessionRouter.TransactOpts, sessionId_)
}

// ClaimForProvider is a paid mutator transaction binding the contract method 0xf2e96e8e.
//
// Solidity: function claimForProvider(bytes32 sessionId_) returns()
func (_SessionRouter *SessionRouterTransactorSession) ClaimForProvider(sessionId_ [32]byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.ClaimForProvider(&_SessionRouter.TransactOpts, sessionId_)
}

// CloseSession is a paid mutator transaction binding the contract method 0x42f77a31.
//
// Solidity: function closeSession(bytes receiptEncoded_, bytes signature_) returns()
func (_SessionRouter *SessionRouterTransactor) CloseSession(opts *bind.TransactOpts, receiptEncoded_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "closeSession", receiptEncoded_, signature_)
}

// CloseSession is a paid mutator transaction binding the contract method 0x42f77a31.
//
// Solidity: function closeSession(bytes receiptEncoded_, bytes signature_) returns()
func (_SessionRouter *SessionRouterSession) CloseSession(receiptEncoded_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.CloseSession(&_SessionRouter.TransactOpts, receiptEncoded_, signature_)
}

// CloseSession is a paid mutator transaction binding the contract method 0x42f77a31.
//
// Solidity: function closeSession(bytes receiptEncoded_, bytes signature_) returns()
func (_SessionRouter *SessionRouterTransactorSession) CloseSession(receiptEncoded_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.CloseSession(&_SessionRouter.TransactOpts, receiptEncoded_, signature_)
}

// OpenSession is a paid mutator transaction binding the contract method 0xa85a1782.
//
// Solidity: function openSession(address user_, uint256 amount_, bool isDirectPaymentFromUser_, bytes approvalEncoded_, bytes signature_) returns(bytes32)
func (_SessionRouter *SessionRouterTransactor) OpenSession(opts *bind.TransactOpts, user_ common.Address, amount_ *big.Int, isDirectPaymentFromUser_ bool, approvalEncoded_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "openSession", user_, amount_, isDirectPaymentFromUser_, approvalEncoded_, signature_)
}

// OpenSession is a paid mutator transaction binding the contract method 0xa85a1782.
//
// Solidity: function openSession(address user_, uint256 amount_, bool isDirectPaymentFromUser_, bytes approvalEncoded_, bytes signature_) returns(bytes32)
func (_SessionRouter *SessionRouterSession) OpenSession(user_ common.Address, amount_ *big.Int, isDirectPaymentFromUser_ bool, approvalEncoded_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.OpenSession(&_SessionRouter.TransactOpts, user_, amount_, isDirectPaymentFromUser_, approvalEncoded_, signature_)
}

// OpenSession is a paid mutator transaction binding the contract method 0xa85a1782.
//
// Solidity: function openSession(address user_, uint256 amount_, bool isDirectPaymentFromUser_, bytes approvalEncoded_, bytes signature_) returns(bytes32)
func (_SessionRouter *SessionRouterTransactorSession) OpenSession(user_ common.Address, amount_ *big.Int, isDirectPaymentFromUser_ bool, approvalEncoded_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.OpenSession(&_SessionRouter.TransactOpts, user_, amount_, isDirectPaymentFromUser_, approvalEncoded_, signature_)
}

// SetMaxSessionDuration is a paid mutator transaction binding the contract method 0xe8664577.
//
// Solidity: function setMaxSessionDuration(uint128 maxSessionDuration_) returns()
func (_SessionRouter *SessionRouterTransactor) SetMaxSessionDuration(opts *bind.TransactOpts, maxSessionDuration_ *big.Int) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "setMaxSessionDuration", maxSessionDuration_)
}

// SetMaxSessionDuration is a paid mutator transaction binding the contract method 0xe8664577.
//
// Solidity: function setMaxSessionDuration(uint128 maxSessionDuration_) returns()
func (_SessionRouter *SessionRouterSession) SetMaxSessionDuration(maxSessionDuration_ *big.Int) (*types.Transaction, error) {
	return _SessionRouter.Contract.SetMaxSessionDuration(&_SessionRouter.TransactOpts, maxSessionDuration_)
}

// SetMaxSessionDuration is a paid mutator transaction binding the contract method 0xe8664577.
//
// Solidity: function setMaxSessionDuration(uint128 maxSessionDuration_) returns()
func (_SessionRouter *SessionRouterTransactorSession) SetMaxSessionDuration(maxSessionDuration_ *big.Int) (*types.Transaction, error) {
	return _SessionRouter.Contract.SetMaxSessionDuration(&_SessionRouter.TransactOpts, maxSessionDuration_)
}

// SetPoolConfig is a paid mutator transaction binding the contract method 0xd7178753.
//
// Solidity: function setPoolConfig(uint256 index_, (uint256,uint256,uint128,uint128) pool_) returns()
func (_SessionRouter *SessionRouterTransactor) SetPoolConfig(opts *bind.TransactOpts, index_ *big.Int, pool_ ISessionStoragePool) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "setPoolConfig", index_, pool_)
}

// SetPoolConfig is a paid mutator transaction binding the contract method 0xd7178753.
//
// Solidity: function setPoolConfig(uint256 index_, (uint256,uint256,uint128,uint128) pool_) returns()
func (_SessionRouter *SessionRouterSession) SetPoolConfig(index_ *big.Int, pool_ ISessionStoragePool) (*types.Transaction, error) {
	return _SessionRouter.Contract.SetPoolConfig(&_SessionRouter.TransactOpts, index_, pool_)
}

// SetPoolConfig is a paid mutator transaction binding the contract method 0xd7178753.
//
// Solidity: function setPoolConfig(uint256 index_, (uint256,uint256,uint128,uint128) pool_) returns()
func (_SessionRouter *SessionRouterTransactorSession) SetPoolConfig(index_ *big.Int, pool_ ISessionStoragePool) (*types.Transaction, error) {
	return _SessionRouter.Contract.SetPoolConfig(&_SessionRouter.TransactOpts, index_, pool_)
}

// WithdrawUserStakes is a paid mutator transaction binding the contract method 0xa98a7c6b.
//
// Solidity: function withdrawUserStakes(address user_, uint8 iterations_) returns()
func (_SessionRouter *SessionRouterTransactor) WithdrawUserStakes(opts *bind.TransactOpts, user_ common.Address, iterations_ uint8) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "withdrawUserStakes", user_, iterations_)
}

// WithdrawUserStakes is a paid mutator transaction binding the contract method 0xa98a7c6b.
//
// Solidity: function withdrawUserStakes(address user_, uint8 iterations_) returns()
func (_SessionRouter *SessionRouterSession) WithdrawUserStakes(user_ common.Address, iterations_ uint8) (*types.Transaction, error) {
	return _SessionRouter.Contract.WithdrawUserStakes(&_SessionRouter.TransactOpts, user_, iterations_)
}

// WithdrawUserStakes is a paid mutator transaction binding the contract method 0xa98a7c6b.
//
// Solidity: function withdrawUserStakes(address user_, uint8 iterations_) returns()
func (_SessionRouter *SessionRouterTransactorSession) WithdrawUserStakes(user_ common.Address, iterations_ uint8) (*types.Transaction, error) {
	return _SessionRouter.Contract.WithdrawUserStakes(&_SessionRouter.TransactOpts, user_, iterations_)
}

// SessionRouterDelegateStakingAllowanceFundedIterator is returned from FilterDelegateStakingAllowanceFunded and is used to iterate over the raw logs and unpacked data for DelegateStakingAllowanceFunded events raised by the SessionRouter contract.
type SessionRouterDelegateStakingAllowanceFundedIterator struct {
	Event *SessionRouterDelegateStakingAllowanceFunded // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingAllowanceFundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingAllowanceFunded)
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
		it.Event = new(SessionRouterDelegateStakingAllowanceFunded)
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
func (it *SessionRouterDelegateStakingAllowanceFundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingAllowanceFundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingAllowanceFunded represents a DelegateStakingAllowanceFunded event raised by the SessionRouter contract.
type SessionRouterDelegateStakingAllowanceFunded struct {
	Hot    common.Address
	Funder common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingAllowanceFunded is a free log retrieval operation binding the contract event 0xab72e208ad5755afbd9bcdfe0b4304d209b428685d5ee350433f417b117f38f0.
//
// Solidity: event DelegateStakingAllowanceFunded(address indexed hot, address indexed funder, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingAllowanceFunded(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*SessionRouterDelegateStakingAllowanceFundedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingAllowanceFunded", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingAllowanceFundedIterator{contract: _SessionRouter.contract, event: "DelegateStakingAllowanceFunded", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingAllowanceFunded is a free log subscription operation binding the contract event 0xab72e208ad5755afbd9bcdfe0b4304d209b428685d5ee350433f417b117f38f0.
//
// Solidity: event DelegateStakingAllowanceFunded(address indexed hot, address indexed funder, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingAllowanceFunded(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingAllowanceFunded, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingAllowanceFunded", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingAllowanceFunded)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingAllowanceFunded", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingAllowanceFunded(log types.Log) (*SessionRouterDelegateStakingAllowanceFunded, error) {
	event := new(SessionRouterDelegateStakingAllowanceFunded)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingAllowanceFunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingAllowanceGrantedIterator is returned from FilterDelegateStakingAllowanceGranted and is used to iterate over the raw logs and unpacked data for DelegateStakingAllowanceGranted events raised by the SessionRouter contract.
type SessionRouterDelegateStakingAllowanceGrantedIterator struct {
	Event *SessionRouterDelegateStakingAllowanceGranted // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingAllowanceGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingAllowanceGranted)
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
		it.Event = new(SessionRouterDelegateStakingAllowanceGranted)
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
func (it *SessionRouterDelegateStakingAllowanceGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingAllowanceGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingAllowanceGranted represents a DelegateStakingAllowanceGranted event raised by the SessionRouter contract.
type SessionRouterDelegateStakingAllowanceGranted struct {
	Hot                  common.Address
	Funder               common.Address
	CumulativeFundingCap *big.Int
	Expiry               *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingAllowanceGranted is a free log retrieval operation binding the contract event 0x743ecb36e3b0fa6c16a5e8cd3ae70e657b294fd68ffbdfdc73d3a89d04addf46.
//
// Solidity: event DelegateStakingAllowanceGranted(address indexed hot, address indexed funder, uint256 cumulativeFundingCap, uint128 expiry)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingAllowanceGranted(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*SessionRouterDelegateStakingAllowanceGrantedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingAllowanceGranted", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingAllowanceGrantedIterator{contract: _SessionRouter.contract, event: "DelegateStakingAllowanceGranted", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingAllowanceGranted is a free log subscription operation binding the contract event 0x743ecb36e3b0fa6c16a5e8cd3ae70e657b294fd68ffbdfdc73d3a89d04addf46.
//
// Solidity: event DelegateStakingAllowanceGranted(address indexed hot, address indexed funder, uint256 cumulativeFundingCap, uint128 expiry)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingAllowanceGranted(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingAllowanceGranted, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingAllowanceGranted", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingAllowanceGranted)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingAllowanceGranted", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingAllowanceGranted(log types.Log) (*SessionRouterDelegateStakingAllowanceGranted, error) {
	event := new(SessionRouterDelegateStakingAllowanceGranted)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingAllowanceGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingAllowanceRevokedIterator is returned from FilterDelegateStakingAllowanceRevoked and is used to iterate over the raw logs and unpacked data for DelegateStakingAllowanceRevoked events raised by the SessionRouter contract.
type SessionRouterDelegateStakingAllowanceRevokedIterator struct {
	Event *SessionRouterDelegateStakingAllowanceRevoked // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingAllowanceRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingAllowanceRevoked)
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
		it.Event = new(SessionRouterDelegateStakingAllowanceRevoked)
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
func (it *SessionRouterDelegateStakingAllowanceRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingAllowanceRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingAllowanceRevoked represents a DelegateStakingAllowanceRevoked event raised by the SessionRouter contract.
type SessionRouterDelegateStakingAllowanceRevoked struct {
	Hot    common.Address
	Funder common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingAllowanceRevoked is a free log retrieval operation binding the contract event 0x641afbae43e9aa28104693da9d716e2295e9efaf231c0bfec7dd6f949d3bd726.
//
// Solidity: event DelegateStakingAllowanceRevoked(address indexed hot, address indexed funder)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingAllowanceRevoked(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*SessionRouterDelegateStakingAllowanceRevokedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingAllowanceRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingAllowanceRevokedIterator{contract: _SessionRouter.contract, event: "DelegateStakingAllowanceRevoked", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingAllowanceRevoked is a free log subscription operation binding the contract event 0x641afbae43e9aa28104693da9d716e2295e9efaf231c0bfec7dd6f949d3bd726.
//
// Solidity: event DelegateStakingAllowanceRevoked(address indexed hot, address indexed funder)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingAllowanceRevoked(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingAllowanceRevoked, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingAllowanceRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingAllowanceRevoked)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingAllowanceRevoked", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingAllowanceRevoked(log types.Log) (*SessionRouterDelegateStakingAllowanceRevoked, error) {
	event := new(SessionRouterDelegateStakingAllowanceRevoked)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingAllowanceRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingMaxActiveFundersUpdatedIterator is returned from FilterDelegateStakingMaxActiveFundersUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMaxActiveFundersUpdated events raised by the SessionRouter contract.
type SessionRouterDelegateStakingMaxActiveFundersUpdatedIterator struct {
	Event *SessionRouterDelegateStakingMaxActiveFundersUpdated // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingMaxActiveFundersUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingMaxActiveFundersUpdated)
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
		it.Event = new(SessionRouterDelegateStakingMaxActiveFundersUpdated)
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
func (it *SessionRouterDelegateStakingMaxActiveFundersUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingMaxActiveFundersUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingMaxActiveFundersUpdated represents a DelegateStakingMaxActiveFundersUpdated event raised by the SessionRouter contract.
type SessionRouterDelegateStakingMaxActiveFundersUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMaxActiveFundersUpdated is a free log retrieval operation binding the contract event 0x6018427c349577edf6a6c0b891cf09b86b47cc1834917516e6da404941ef43f5.
//
// Solidity: event DelegateStakingMaxActiveFundersUpdated(uint256 value)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingMaxActiveFundersUpdated(opts *bind.FilterOpts) (*SessionRouterDelegateStakingMaxActiveFundersUpdatedIterator, error) {

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingMaxActiveFundersUpdated")
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingMaxActiveFundersUpdatedIterator{contract: _SessionRouter.contract, event: "DelegateStakingMaxActiveFundersUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMaxActiveFundersUpdated is a free log subscription operation binding the contract event 0x6018427c349577edf6a6c0b891cf09b86b47cc1834917516e6da404941ef43f5.
//
// Solidity: event DelegateStakingMaxActiveFundersUpdated(uint256 value)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingMaxActiveFundersUpdated(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingMaxActiveFundersUpdated) (event.Subscription, error) {

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingMaxActiveFundersUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingMaxActiveFundersUpdated)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingMaxActiveFundersUpdated", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingMaxActiveFundersUpdated(log types.Log) (*SessionRouterDelegateStakingMaxActiveFundersUpdated, error) {
	event := new(SessionRouterDelegateStakingMaxActiveFundersUpdated)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingMaxActiveFundersUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingMaxAutoReleaseDaysUpdatedIterator is returned from FilterDelegateStakingMaxAutoReleaseDaysUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMaxAutoReleaseDaysUpdated events raised by the SessionRouter contract.
type SessionRouterDelegateStakingMaxAutoReleaseDaysUpdatedIterator struct {
	Event *SessionRouterDelegateStakingMaxAutoReleaseDaysUpdated // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingMaxAutoReleaseDaysUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingMaxAutoReleaseDaysUpdated)
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
		it.Event = new(SessionRouterDelegateStakingMaxAutoReleaseDaysUpdated)
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
func (it *SessionRouterDelegateStakingMaxAutoReleaseDaysUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingMaxAutoReleaseDaysUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingMaxAutoReleaseDaysUpdated represents a DelegateStakingMaxAutoReleaseDaysUpdated event raised by the SessionRouter contract.
type SessionRouterDelegateStakingMaxAutoReleaseDaysUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMaxAutoReleaseDaysUpdated is a free log retrieval operation binding the contract event 0xc9ca6843c5a726f8df173b27882fb647575843c62e31760e9b7f90c94518aa2a.
//
// Solidity: event DelegateStakingMaxAutoReleaseDaysUpdated(uint256 value)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingMaxAutoReleaseDaysUpdated(opts *bind.FilterOpts) (*SessionRouterDelegateStakingMaxAutoReleaseDaysUpdatedIterator, error) {

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingMaxAutoReleaseDaysUpdated")
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingMaxAutoReleaseDaysUpdatedIterator{contract: _SessionRouter.contract, event: "DelegateStakingMaxAutoReleaseDaysUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMaxAutoReleaseDaysUpdated is a free log subscription operation binding the contract event 0xc9ca6843c5a726f8df173b27882fb647575843c62e31760e9b7f90c94518aa2a.
//
// Solidity: event DelegateStakingMaxAutoReleaseDaysUpdated(uint256 value)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingMaxAutoReleaseDaysUpdated(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingMaxAutoReleaseDaysUpdated) (event.Subscription, error) {

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingMaxAutoReleaseDaysUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingMaxAutoReleaseDaysUpdated)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingMaxAutoReleaseDaysUpdated", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingMaxAutoReleaseDaysUpdated(log types.Log) (*SessionRouterDelegateStakingMaxAutoReleaseDaysUpdated, error) {
	event := new(SessionRouterDelegateStakingMaxAutoReleaseDaysUpdated)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingMaxAutoReleaseDaysUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingMaxAutoServiceUpdatedIterator is returned from FilterDelegateStakingMaxAutoServiceUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMaxAutoServiceUpdated events raised by the SessionRouter contract.
type SessionRouterDelegateStakingMaxAutoServiceUpdatedIterator struct {
	Event *SessionRouterDelegateStakingMaxAutoServiceUpdated // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingMaxAutoServiceUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingMaxAutoServiceUpdated)
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
		it.Event = new(SessionRouterDelegateStakingMaxAutoServiceUpdated)
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
func (it *SessionRouterDelegateStakingMaxAutoServiceUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingMaxAutoServiceUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingMaxAutoServiceUpdated represents a DelegateStakingMaxAutoServiceUpdated event raised by the SessionRouter contract.
type SessionRouterDelegateStakingMaxAutoServiceUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMaxAutoServiceUpdated is a free log retrieval operation binding the contract event 0x4f927d3e460674eab7e38dac3592136fe961de0c14533b919652a814a7510adc.
//
// Solidity: event DelegateStakingMaxAutoServiceUpdated(uint256 value)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingMaxAutoServiceUpdated(opts *bind.FilterOpts) (*SessionRouterDelegateStakingMaxAutoServiceUpdatedIterator, error) {

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingMaxAutoServiceUpdated")
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingMaxAutoServiceUpdatedIterator{contract: _SessionRouter.contract, event: "DelegateStakingMaxAutoServiceUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMaxAutoServiceUpdated is a free log subscription operation binding the contract event 0x4f927d3e460674eab7e38dac3592136fe961de0c14533b919652a814a7510adc.
//
// Solidity: event DelegateStakingMaxAutoServiceUpdated(uint256 value)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingMaxAutoServiceUpdated(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingMaxAutoServiceUpdated) (event.Subscription, error) {

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingMaxAutoServiceUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingMaxAutoServiceUpdated)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingMaxAutoServiceUpdated", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingMaxAutoServiceUpdated(log types.Log) (*SessionRouterDelegateStakingMaxAutoServiceUpdated, error) {
	event := new(SessionRouterDelegateStakingMaxAutoServiceUpdated)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingMaxAutoServiceUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingMinPrincipalUpdatedIterator is returned from FilterDelegateStakingMinPrincipalUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMinPrincipalUpdated events raised by the SessionRouter contract.
type SessionRouterDelegateStakingMinPrincipalUpdatedIterator struct {
	Event *SessionRouterDelegateStakingMinPrincipalUpdated // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingMinPrincipalUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingMinPrincipalUpdated)
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
		it.Event = new(SessionRouterDelegateStakingMinPrincipalUpdated)
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
func (it *SessionRouterDelegateStakingMinPrincipalUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingMinPrincipalUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingMinPrincipalUpdated represents a DelegateStakingMinPrincipalUpdated event raised by the SessionRouter contract.
type SessionRouterDelegateStakingMinPrincipalUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMinPrincipalUpdated is a free log retrieval operation binding the contract event 0x06ae87e705be2a5e7c6ff17a5d3d836834554f7d0aadaef3abb4343a202590a2.
//
// Solidity: event DelegateStakingMinPrincipalUpdated(uint256 value)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingMinPrincipalUpdated(opts *bind.FilterOpts) (*SessionRouterDelegateStakingMinPrincipalUpdatedIterator, error) {

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingMinPrincipalUpdated")
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingMinPrincipalUpdatedIterator{contract: _SessionRouter.contract, event: "DelegateStakingMinPrincipalUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMinPrincipalUpdated is a free log subscription operation binding the contract event 0x06ae87e705be2a5e7c6ff17a5d3d836834554f7d0aadaef3abb4343a202590a2.
//
// Solidity: event DelegateStakingMinPrincipalUpdated(uint256 value)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingMinPrincipalUpdated(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingMinPrincipalUpdated) (event.Subscription, error) {

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingMinPrincipalUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingMinPrincipalUpdated)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingMinPrincipalUpdated", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingMinPrincipalUpdated(log types.Log) (*SessionRouterDelegateStakingMinPrincipalUpdated, error) {
	event := new(SessionRouterDelegateStakingMinPrincipalUpdated)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingMinPrincipalUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingPoolDrawnIterator is returned from FilterDelegateStakingPoolDrawn and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolDrawn events raised by the SessionRouter contract.
type SessionRouterDelegateStakingPoolDrawnIterator struct {
	Event *SessionRouterDelegateStakingPoolDrawn // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingPoolDrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingPoolDrawn)
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
		it.Event = new(SessionRouterDelegateStakingPoolDrawn)
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
func (it *SessionRouterDelegateStakingPoolDrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingPoolDrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingPoolDrawn represents a DelegateStakingPoolDrawn event raised by the SessionRouter contract.
type SessionRouterDelegateStakingPoolDrawn struct {
	SessionId [32]byte
	Hot       common.Address
	Funder    common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolDrawn is a free log retrieval operation binding the contract event 0xd178a2f3e5da2cc00732e1b88d2670a1eaa4f88a0b86e8790237a677f1fa0985.
//
// Solidity: event DelegateStakingPoolDrawn(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingPoolDrawn(opts *bind.FilterOpts, sessionId [][32]byte, hot []common.Address, funder []common.Address) (*SessionRouterDelegateStakingPoolDrawnIterator, error) {

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

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingPoolDrawn", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingPoolDrawnIterator{contract: _SessionRouter.contract, event: "DelegateStakingPoolDrawn", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolDrawn is a free log subscription operation binding the contract event 0xd178a2f3e5da2cc00732e1b88d2670a1eaa4f88a0b86e8790237a677f1fa0985.
//
// Solidity: event DelegateStakingPoolDrawn(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingPoolDrawn(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingPoolDrawn, sessionId [][32]byte, hot []common.Address, funder []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingPoolDrawn", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingPoolDrawn)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingPoolDrawn", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingPoolDrawn(log types.Log) (*SessionRouterDelegateStakingPoolDrawn, error) {
	event := new(SessionRouterDelegateStakingPoolDrawn)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingPoolDrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingPoolFunderApprovalRevokedIterator is returned from FilterDelegateStakingPoolFunderApprovalRevoked and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolFunderApprovalRevoked events raised by the SessionRouter contract.
type SessionRouterDelegateStakingPoolFunderApprovalRevokedIterator struct {
	Event *SessionRouterDelegateStakingPoolFunderApprovalRevoked // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingPoolFunderApprovalRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingPoolFunderApprovalRevoked)
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
		it.Event = new(SessionRouterDelegateStakingPoolFunderApprovalRevoked)
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
func (it *SessionRouterDelegateStakingPoolFunderApprovalRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingPoolFunderApprovalRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingPoolFunderApprovalRevoked represents a DelegateStakingPoolFunderApprovalRevoked event raised by the SessionRouter contract.
type SessionRouterDelegateStakingPoolFunderApprovalRevoked struct {
	Hot    common.Address
	Funder common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolFunderApprovalRevoked is a free log retrieval operation binding the contract event 0x5ebb99e7e14d99bff22536b105e92a32d4fbb701f93b0d5b008cc0a0b876aeb2.
//
// Solidity: event DelegateStakingPoolFunderApprovalRevoked(address indexed hot, address indexed funder)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingPoolFunderApprovalRevoked(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*SessionRouterDelegateStakingPoolFunderApprovalRevokedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingPoolFunderApprovalRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingPoolFunderApprovalRevokedIterator{contract: _SessionRouter.contract, event: "DelegateStakingPoolFunderApprovalRevoked", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolFunderApprovalRevoked is a free log subscription operation binding the contract event 0x5ebb99e7e14d99bff22536b105e92a32d4fbb701f93b0d5b008cc0a0b876aeb2.
//
// Solidity: event DelegateStakingPoolFunderApprovalRevoked(address indexed hot, address indexed funder)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingPoolFunderApprovalRevoked(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingPoolFunderApprovalRevoked, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingPoolFunderApprovalRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingPoolFunderApprovalRevoked)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingPoolFunderApprovalRevoked", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingPoolFunderApprovalRevoked(log types.Log) (*SessionRouterDelegateStakingPoolFunderApprovalRevoked, error) {
	event := new(SessionRouterDelegateStakingPoolFunderApprovalRevoked)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingPoolFunderApprovalRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingPoolFunderApprovedIterator is returned from FilterDelegateStakingPoolFunderApproved and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolFunderApproved events raised by the SessionRouter contract.
type SessionRouterDelegateStakingPoolFunderApprovedIterator struct {
	Event *SessionRouterDelegateStakingPoolFunderApproved // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingPoolFunderApprovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingPoolFunderApproved)
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
		it.Event = new(SessionRouterDelegateStakingPoolFunderApproved)
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
func (it *SessionRouterDelegateStakingPoolFunderApprovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingPoolFunderApprovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingPoolFunderApproved represents a DelegateStakingPoolFunderApproved event raised by the SessionRouter contract.
type SessionRouterDelegateStakingPoolFunderApproved struct {
	Hot    common.Address
	Funder common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolFunderApproved is a free log retrieval operation binding the contract event 0x2a6f85b141a6737d1530bd4b8c916eee97ba403dae709170266f72b6ec58f739.
//
// Solidity: event DelegateStakingPoolFunderApproved(address indexed hot, address indexed funder)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingPoolFunderApproved(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*SessionRouterDelegateStakingPoolFunderApprovedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingPoolFunderApproved", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingPoolFunderApprovedIterator{contract: _SessionRouter.contract, event: "DelegateStakingPoolFunderApproved", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolFunderApproved is a free log subscription operation binding the contract event 0x2a6f85b141a6737d1530bd4b8c916eee97ba403dae709170266f72b6ec58f739.
//
// Solidity: event DelegateStakingPoolFunderApproved(address indexed hot, address indexed funder)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingPoolFunderApproved(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingPoolFunderApproved, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingPoolFunderApproved", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingPoolFunderApproved)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingPoolFunderApproved", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingPoolFunderApproved(log types.Log) (*SessionRouterDelegateStakingPoolFunderApproved, error) {
	event := new(SessionRouterDelegateStakingPoolFunderApproved)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingPoolFunderApproved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingPoolLockCreatedIterator is returned from FilterDelegateStakingPoolLockCreated and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolLockCreated events raised by the SessionRouter contract.
type SessionRouterDelegateStakingPoolLockCreatedIterator struct {
	Event *SessionRouterDelegateStakingPoolLockCreated // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingPoolLockCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingPoolLockCreated)
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
		it.Event = new(SessionRouterDelegateStakingPoolLockCreated)
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
func (it *SessionRouterDelegateStakingPoolLockCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingPoolLockCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingPoolLockCreated represents a DelegateStakingPoolLockCreated event raised by the SessionRouter contract.
type SessionRouterDelegateStakingPoolLockCreated struct {
	Hot       common.Address
	ReleaseAt *big.Int
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolLockCreated is a free log retrieval operation binding the contract event 0x87eb288bde7ca5d84aa5e85a8b81f8d7c31d8e19049823dab5552d8bb461dc0e.
//
// Solidity: event DelegateStakingPoolLockCreated(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingPoolLockCreated(opts *bind.FilterOpts, hot []common.Address, releaseAt []*big.Int) (*SessionRouterDelegateStakingPoolLockCreatedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingPoolLockCreated", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingPoolLockCreatedIterator{contract: _SessionRouter.contract, event: "DelegateStakingPoolLockCreated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolLockCreated is a free log subscription operation binding the contract event 0x87eb288bde7ca5d84aa5e85a8b81f8d7c31d8e19049823dab5552d8bb461dc0e.
//
// Solidity: event DelegateStakingPoolLockCreated(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingPoolLockCreated(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingPoolLockCreated, hot []common.Address, releaseAt []*big.Int) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingPoolLockCreated", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingPoolLockCreated)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingPoolLockCreated", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingPoolLockCreated(log types.Log) (*SessionRouterDelegateStakingPoolLockCreated, error) {
	event := new(SessionRouterDelegateStakingPoolLockCreated)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingPoolLockCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingPoolLockReleasedIterator is returned from FilterDelegateStakingPoolLockReleased and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolLockReleased events raised by the SessionRouter contract.
type SessionRouterDelegateStakingPoolLockReleasedIterator struct {
	Event *SessionRouterDelegateStakingPoolLockReleased // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingPoolLockReleasedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingPoolLockReleased)
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
		it.Event = new(SessionRouterDelegateStakingPoolLockReleased)
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
func (it *SessionRouterDelegateStakingPoolLockReleasedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingPoolLockReleasedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingPoolLockReleased represents a DelegateStakingPoolLockReleased event raised by the SessionRouter contract.
type SessionRouterDelegateStakingPoolLockReleased struct {
	Hot       common.Address
	ReleaseAt *big.Int
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolLockReleased is a free log retrieval operation binding the contract event 0xd04817baed5e7c2b01f59b2218c05170e733b0e2a3864693721a2fc89e5f7a4f.
//
// Solidity: event DelegateStakingPoolLockReleased(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingPoolLockReleased(opts *bind.FilterOpts, hot []common.Address, releaseAt []*big.Int) (*SessionRouterDelegateStakingPoolLockReleasedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingPoolLockReleased", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingPoolLockReleasedIterator{contract: _SessionRouter.contract, event: "DelegateStakingPoolLockReleased", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolLockReleased is a free log subscription operation binding the contract event 0xd04817baed5e7c2b01f59b2218c05170e733b0e2a3864693721a2fc89e5f7a4f.
//
// Solidity: event DelegateStakingPoolLockReleased(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingPoolLockReleased(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingPoolLockReleased, hot []common.Address, releaseAt []*big.Int) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingPoolLockReleased", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingPoolLockReleased)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingPoolLockReleased", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingPoolLockReleased(log types.Log) (*SessionRouterDelegateStakingPoolLockReleased, error) {
	event := new(SessionRouterDelegateStakingPoolLockReleased)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingPoolLockReleased", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingPoolReturnedIterator is returned from FilterDelegateStakingPoolReturned and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolReturned events raised by the SessionRouter contract.
type SessionRouterDelegateStakingPoolReturnedIterator struct {
	Event *SessionRouterDelegateStakingPoolReturned // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingPoolReturnedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingPoolReturned)
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
		it.Event = new(SessionRouterDelegateStakingPoolReturned)
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
func (it *SessionRouterDelegateStakingPoolReturnedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingPoolReturnedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingPoolReturned represents a DelegateStakingPoolReturned event raised by the SessionRouter contract.
type SessionRouterDelegateStakingPoolReturned struct {
	SessionId [32]byte
	Hot       common.Address
	Funder    common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolReturned is a free log retrieval operation binding the contract event 0x66524d7951752d8716b27064a7d50819e4db87609ea9eca1f6b1d42f359a061d.
//
// Solidity: event DelegateStakingPoolReturned(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingPoolReturned(opts *bind.FilterOpts, sessionId [][32]byte, hot []common.Address, funder []common.Address) (*SessionRouterDelegateStakingPoolReturnedIterator, error) {

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

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingPoolReturned", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingPoolReturnedIterator{contract: _SessionRouter.contract, event: "DelegateStakingPoolReturned", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolReturned is a free log subscription operation binding the contract event 0x66524d7951752d8716b27064a7d50819e4db87609ea9eca1f6b1d42f359a061d.
//
// Solidity: event DelegateStakingPoolReturned(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingPoolReturned(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingPoolReturned, sessionId [][32]byte, hot []common.Address, funder []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingPoolReturned", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingPoolReturned)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingPoolReturned", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingPoolReturned(log types.Log) (*SessionRouterDelegateStakingPoolReturned, error) {
	event := new(SessionRouterDelegateStakingPoolReturned)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingPoolReturned", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingPoolSessionSettledIterator is returned from FilterDelegateStakingPoolSessionSettled and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolSessionSettled events raised by the SessionRouter contract.
type SessionRouterDelegateStakingPoolSessionSettledIterator struct {
	Event *SessionRouterDelegateStakingPoolSessionSettled // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingPoolSessionSettledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingPoolSessionSettled)
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
		it.Event = new(SessionRouterDelegateStakingPoolSessionSettled)
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
func (it *SessionRouterDelegateStakingPoolSessionSettledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingPoolSessionSettledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingPoolSessionSettled represents a DelegateStakingPoolSessionSettled event raised by the SessionRouter contract.
type SessionRouterDelegateStakingPoolSessionSettled struct {
	SessionId [32]byte
	Hot       common.Address
	Provider  common.Address
	Caller    common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolSessionSettled is a free log retrieval operation binding the contract event 0x518bf7ca4dc0c0c2144467883693da5fd9b0532de6539122309ecff49a330215.
//
// Solidity: event DelegateStakingPoolSessionSettled(bytes32 indexed sessionId, address indexed hot, address indexed provider, address caller)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingPoolSessionSettled(opts *bind.FilterOpts, sessionId [][32]byte, hot []common.Address, provider []common.Address) (*SessionRouterDelegateStakingPoolSessionSettledIterator, error) {

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

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingPoolSessionSettled", sessionIdRule, hotRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingPoolSessionSettledIterator{contract: _SessionRouter.contract, event: "DelegateStakingPoolSessionSettled", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolSessionSettled is a free log subscription operation binding the contract event 0x518bf7ca4dc0c0c2144467883693da5fd9b0532de6539122309ecff49a330215.
//
// Solidity: event DelegateStakingPoolSessionSettled(bytes32 indexed sessionId, address indexed hot, address indexed provider, address caller)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingPoolSessionSettled(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingPoolSessionSettled, sessionId [][32]byte, hot []common.Address, provider []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingPoolSessionSettled", sessionIdRule, hotRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingPoolSessionSettled)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingPoolSessionSettled", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingPoolSessionSettled(log types.Log) (*SessionRouterDelegateStakingPoolSessionSettled, error) {
	event := new(SessionRouterDelegateStakingPoolSessionSettled)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingPoolSessionSettled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingSettlementGraceUpdatedIterator is returned from FilterDelegateStakingSettlementGraceUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingSettlementGraceUpdated events raised by the SessionRouter contract.
type SessionRouterDelegateStakingSettlementGraceUpdatedIterator struct {
	Event *SessionRouterDelegateStakingSettlementGraceUpdated // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingSettlementGraceUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingSettlementGraceUpdated)
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
		it.Event = new(SessionRouterDelegateStakingSettlementGraceUpdated)
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
func (it *SessionRouterDelegateStakingSettlementGraceUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingSettlementGraceUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingSettlementGraceUpdated represents a DelegateStakingSettlementGraceUpdated event raised by the SessionRouter contract.
type SessionRouterDelegateStakingSettlementGraceUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingSettlementGraceUpdated is a free log retrieval operation binding the contract event 0x23d4c1527dfe24f1a043524d5c7c2b730f3fb2841478487080916494bc8487f2.
//
// Solidity: event DelegateStakingSettlementGraceUpdated(uint256 value)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingSettlementGraceUpdated(opts *bind.FilterOpts) (*SessionRouterDelegateStakingSettlementGraceUpdatedIterator, error) {

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingSettlementGraceUpdated")
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingSettlementGraceUpdatedIterator{contract: _SessionRouter.contract, event: "DelegateStakingSettlementGraceUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingSettlementGraceUpdated is a free log subscription operation binding the contract event 0x23d4c1527dfe24f1a043524d5c7c2b730f3fb2841478487080916494bc8487f2.
//
// Solidity: event DelegateStakingSettlementGraceUpdated(uint256 value)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingSettlementGraceUpdated(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingSettlementGraceUpdated) (event.Subscription, error) {

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingSettlementGraceUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingSettlementGraceUpdated)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingSettlementGraceUpdated", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingSettlementGraceUpdated(log types.Log) (*SessionRouterDelegateStakingSettlementGraceUpdated, error) {
	event := new(SessionRouterDelegateStakingSettlementGraceUpdated)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingSettlementGraceUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingWithdrawalCancelledIterator is returned from FilterDelegateStakingWithdrawalCancelled and is used to iterate over the raw logs and unpacked data for DelegateStakingWithdrawalCancelled events raised by the SessionRouter contract.
type SessionRouterDelegateStakingWithdrawalCancelledIterator struct {
	Event *SessionRouterDelegateStakingWithdrawalCancelled // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingWithdrawalCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingWithdrawalCancelled)
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
		it.Event = new(SessionRouterDelegateStakingWithdrawalCancelled)
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
func (it *SessionRouterDelegateStakingWithdrawalCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingWithdrawalCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingWithdrawalCancelled represents a DelegateStakingWithdrawalCancelled event raised by the SessionRouter contract.
type SessionRouterDelegateStakingWithdrawalCancelled struct {
	Hot    common.Address
	Funder common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingWithdrawalCancelled is a free log retrieval operation binding the contract event 0x00041bc92ab137f1f37671a48ce9dc19f99eb0d99b0bf41beb1ffca2dd9cb557.
//
// Solidity: event DelegateStakingWithdrawalCancelled(address indexed hot, address indexed funder, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingWithdrawalCancelled(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*SessionRouterDelegateStakingWithdrawalCancelledIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingWithdrawalCancelled", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingWithdrawalCancelledIterator{contract: _SessionRouter.contract, event: "DelegateStakingWithdrawalCancelled", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingWithdrawalCancelled is a free log subscription operation binding the contract event 0x00041bc92ab137f1f37671a48ce9dc19f99eb0d99b0bf41beb1ffca2dd9cb557.
//
// Solidity: event DelegateStakingWithdrawalCancelled(address indexed hot, address indexed funder, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingWithdrawalCancelled(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingWithdrawalCancelled, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingWithdrawalCancelled", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingWithdrawalCancelled)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingWithdrawalCancelled", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingWithdrawalCancelled(log types.Log) (*SessionRouterDelegateStakingWithdrawalCancelled, error) {
	event := new(SessionRouterDelegateStakingWithdrawalCancelled)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingWithdrawalCancelled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingWithdrawalPaidIterator is returned from FilterDelegateStakingWithdrawalPaid and is used to iterate over the raw logs and unpacked data for DelegateStakingWithdrawalPaid events raised by the SessionRouter contract.
type SessionRouterDelegateStakingWithdrawalPaidIterator struct {
	Event *SessionRouterDelegateStakingWithdrawalPaid // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingWithdrawalPaidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingWithdrawalPaid)
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
		it.Event = new(SessionRouterDelegateStakingWithdrawalPaid)
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
func (it *SessionRouterDelegateStakingWithdrawalPaidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingWithdrawalPaidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingWithdrawalPaid represents a DelegateStakingWithdrawalPaid event raised by the SessionRouter contract.
type SessionRouterDelegateStakingWithdrawalPaid struct {
	Hot    common.Address
	Funder common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingWithdrawalPaid is a free log retrieval operation binding the contract event 0x7827e4dd700b7a2a556c261a5a293c5c936068522520173931017b0e9959d935.
//
// Solidity: event DelegateStakingWithdrawalPaid(address indexed hot, address indexed funder, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingWithdrawalPaid(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*SessionRouterDelegateStakingWithdrawalPaidIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingWithdrawalPaid", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingWithdrawalPaidIterator{contract: _SessionRouter.contract, event: "DelegateStakingWithdrawalPaid", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingWithdrawalPaid is a free log subscription operation binding the contract event 0x7827e4dd700b7a2a556c261a5a293c5c936068522520173931017b0e9959d935.
//
// Solidity: event DelegateStakingWithdrawalPaid(address indexed hot, address indexed funder, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingWithdrawalPaid(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingWithdrawalPaid, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingWithdrawalPaid", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingWithdrawalPaid)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingWithdrawalPaid", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingWithdrawalPaid(log types.Log) (*SessionRouterDelegateStakingWithdrawalPaid, error) {
	event := new(SessionRouterDelegateStakingWithdrawalPaid)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingWithdrawalPaid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterDelegateStakingWithdrawalRequestedIterator is returned from FilterDelegateStakingWithdrawalRequested and is used to iterate over the raw logs and unpacked data for DelegateStakingWithdrawalRequested events raised by the SessionRouter contract.
type SessionRouterDelegateStakingWithdrawalRequestedIterator struct {
	Event *SessionRouterDelegateStakingWithdrawalRequested // Event containing the contract specifics and raw log

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
func (it *SessionRouterDelegateStakingWithdrawalRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterDelegateStakingWithdrawalRequested)
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
		it.Event = new(SessionRouterDelegateStakingWithdrawalRequested)
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
func (it *SessionRouterDelegateStakingWithdrawalRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterDelegateStakingWithdrawalRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterDelegateStakingWithdrawalRequested represents a DelegateStakingWithdrawalRequested event raised by the SessionRouter contract.
type SessionRouterDelegateStakingWithdrawalRequested struct {
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
func (_SessionRouter *SessionRouterFilterer) FilterDelegateStakingWithdrawalRequested(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*SessionRouterDelegateStakingWithdrawalRequestedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "DelegateStakingWithdrawalRequested", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterDelegateStakingWithdrawalRequestedIterator{contract: _SessionRouter.contract, event: "DelegateStakingWithdrawalRequested", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingWithdrawalRequested is a free log subscription operation binding the contract event 0xab9f026078e090833bf0fd1211137db97a7a96f822c650c7997cdb7bfc859799.
//
// Solidity: event DelegateStakingWithdrawalRequested(address indexed hot, address indexed funder, uint256 amount, uint256 paid, uint256 pending)
func (_SessionRouter *SessionRouterFilterer) WatchDelegateStakingWithdrawalRequested(opts *bind.WatchOpts, sink chan<- *SessionRouterDelegateStakingWithdrawalRequested, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "DelegateStakingWithdrawalRequested", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterDelegateStakingWithdrawalRequested)
				if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingWithdrawalRequested", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseDelegateStakingWithdrawalRequested(log types.Log) (*SessionRouterDelegateStakingWithdrawalRequested, error) {
	event := new(SessionRouterDelegateStakingWithdrawalRequested)
	if err := _SessionRouter.contract.UnpackLog(event, "DelegateStakingWithdrawalRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the SessionRouter contract.
type SessionRouterInitializedIterator struct {
	Event *SessionRouterInitialized // Event containing the contract specifics and raw log

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
func (it *SessionRouterInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterInitialized)
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
		it.Event = new(SessionRouterInitialized)
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
func (it *SessionRouterInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterInitialized represents a Initialized event raised by the SessionRouter contract.
type SessionRouterInitialized struct {
	StorageSlot [32]byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
func (_SessionRouter *SessionRouterFilterer) FilterInitialized(opts *bind.FilterOpts) (*SessionRouterInitializedIterator, error) {

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &SessionRouterInitializedIterator{contract: _SessionRouter.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
func (_SessionRouter *SessionRouterFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *SessionRouterInitialized) (event.Subscription, error) {

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterInitialized)
				if err := _SessionRouter.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseInitialized(log types.Log) (*SessionRouterInitialized, error) {
	event := new(SessionRouterInitialized)
	if err := _SessionRouter.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterProviderDebtPaidIterator is returned from FilterProviderDebtPaid and is used to iterate over the raw logs and unpacked data for ProviderDebtPaid events raised by the SessionRouter contract.
type SessionRouterProviderDebtPaidIterator struct {
	Event *SessionRouterProviderDebtPaid // Event containing the contract specifics and raw log

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
func (it *SessionRouterProviderDebtPaidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterProviderDebtPaid)
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
		it.Event = new(SessionRouterProviderDebtPaid)
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
func (it *SessionRouterProviderDebtPaidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterProviderDebtPaidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterProviderDebtPaid represents a ProviderDebtPaid event raised by the SessionRouter contract.
type SessionRouterProviderDebtPaid struct {
	Provider     common.Address
	PayoutTarget common.Address
	Amount       *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterProviderDebtPaid is a free log retrieval operation binding the contract event 0x8d79baa0ce0a07e4a85c9f1c273815f861513a2e9fdb2082bf443519534aa2b0.
//
// Solidity: event ProviderDebtPaid(address indexed provider, address indexed payoutTarget, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) FilterProviderDebtPaid(opts *bind.FilterOpts, provider []common.Address, payoutTarget []common.Address) (*SessionRouterProviderDebtPaidIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var payoutTargetRule []interface{}
	for _, payoutTargetItem := range payoutTarget {
		payoutTargetRule = append(payoutTargetRule, payoutTargetItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "ProviderDebtPaid", providerRule, payoutTargetRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterProviderDebtPaidIterator{contract: _SessionRouter.contract, event: "ProviderDebtPaid", logs: logs, sub: sub}, nil
}

// WatchProviderDebtPaid is a free log subscription operation binding the contract event 0x8d79baa0ce0a07e4a85c9f1c273815f861513a2e9fdb2082bf443519534aa2b0.
//
// Solidity: event ProviderDebtPaid(address indexed provider, address indexed payoutTarget, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) WatchProviderDebtPaid(opts *bind.WatchOpts, sink chan<- *SessionRouterProviderDebtPaid, provider []common.Address, payoutTarget []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var payoutTargetRule []interface{}
	for _, payoutTargetItem := range payoutTarget {
		payoutTargetRule = append(payoutTargetRule, payoutTargetItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "ProviderDebtPaid", providerRule, payoutTargetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterProviderDebtPaid)
				if err := _SessionRouter.contract.UnpackLog(event, "ProviderDebtPaid", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseProviderDebtPaid(log types.Log) (*SessionRouterProviderDebtPaid, error) {
	event := new(SessionRouterProviderDebtPaid)
	if err := _SessionRouter.contract.UnpackLog(event, "ProviderDebtPaid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterProviderPayoutTargetUpdatedIterator is returned from FilterProviderPayoutTargetUpdated and is used to iterate over the raw logs and unpacked data for ProviderPayoutTargetUpdated events raised by the SessionRouter contract.
type SessionRouterProviderPayoutTargetUpdatedIterator struct {
	Event *SessionRouterProviderPayoutTargetUpdated // Event containing the contract specifics and raw log

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
func (it *SessionRouterProviderPayoutTargetUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterProviderPayoutTargetUpdated)
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
		it.Event = new(SessionRouterProviderPayoutTargetUpdated)
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
func (it *SessionRouterProviderPayoutTargetUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterProviderPayoutTargetUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterProviderPayoutTargetUpdated represents a ProviderPayoutTargetUpdated event raised by the SessionRouter contract.
type SessionRouterProviderPayoutTargetUpdated struct {
	Provider     common.Address
	PayoutTarget common.Address
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterProviderPayoutTargetUpdated is a free log retrieval operation binding the contract event 0x1cba50759ff5d1e813b2ed6b4f9ce6cd19137e624f4bbde798eb0572b74a6968.
//
// Solidity: event ProviderPayoutTargetUpdated(address indexed provider, address indexed payoutTarget)
func (_SessionRouter *SessionRouterFilterer) FilterProviderPayoutTargetUpdated(opts *bind.FilterOpts, provider []common.Address, payoutTarget []common.Address) (*SessionRouterProviderPayoutTargetUpdatedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var payoutTargetRule []interface{}
	for _, payoutTargetItem := range payoutTarget {
		payoutTargetRule = append(payoutTargetRule, payoutTargetItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "ProviderPayoutTargetUpdated", providerRule, payoutTargetRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterProviderPayoutTargetUpdatedIterator{contract: _SessionRouter.contract, event: "ProviderPayoutTargetUpdated", logs: logs, sub: sub}, nil
}

// WatchProviderPayoutTargetUpdated is a free log subscription operation binding the contract event 0x1cba50759ff5d1e813b2ed6b4f9ce6cd19137e624f4bbde798eb0572b74a6968.
//
// Solidity: event ProviderPayoutTargetUpdated(address indexed provider, address indexed payoutTarget)
func (_SessionRouter *SessionRouterFilterer) WatchProviderPayoutTargetUpdated(opts *bind.WatchOpts, sink chan<- *SessionRouterProviderPayoutTargetUpdated, provider []common.Address, payoutTarget []common.Address) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var payoutTargetRule []interface{}
	for _, payoutTargetItem := range payoutTarget {
		payoutTargetRule = append(payoutTargetRule, payoutTargetItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "ProviderPayoutTargetUpdated", providerRule, payoutTargetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterProviderPayoutTargetUpdated)
				if err := _SessionRouter.contract.UnpackLog(event, "ProviderPayoutTargetUpdated", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseProviderPayoutTargetUpdated(log types.Log) (*SessionRouterProviderPayoutTargetUpdated, error) {
	event := new(SessionRouterProviderPayoutTargetUpdated)
	if err := _SessionRouter.contract.UnpackLog(event, "ProviderPayoutTargetUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterProviderRewardAccruedIterator is returned from FilterProviderRewardAccrued and is used to iterate over the raw logs and unpacked data for ProviderRewardAccrued events raised by the SessionRouter contract.
type SessionRouterProviderRewardAccruedIterator struct {
	Event *SessionRouterProviderRewardAccrued // Event containing the contract specifics and raw log

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
func (it *SessionRouterProviderRewardAccruedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterProviderRewardAccrued)
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
		it.Event = new(SessionRouterProviderRewardAccrued)
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
func (it *SessionRouterProviderRewardAccruedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterProviderRewardAccruedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterProviderRewardAccrued represents a ProviderRewardAccrued event raised by the SessionRouter contract.
type SessionRouterProviderRewardAccrued struct {
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
func (_SessionRouter *SessionRouterFilterer) FilterProviderRewardAccrued(opts *bind.FilterOpts, sessionId [][32]byte, provider []common.Address) (*SessionRouterProviderRewardAccruedIterator, error) {

	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "ProviderRewardAccrued", sessionIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterProviderRewardAccruedIterator{contract: _SessionRouter.contract, event: "ProviderRewardAccrued", logs: logs, sub: sub}, nil
}

// WatchProviderRewardAccrued is a free log subscription operation binding the contract event 0x2e9f0aeeab6a0852b0c976f7e0d74566d7b177b6612d645afd583e6ec4c0b2a6.
//
// Solidity: event ProviderRewardAccrued(bytes32 indexed sessionId, address indexed provider, uint256 accrued, uint256 paid, uint256 owed)
func (_SessionRouter *SessionRouterFilterer) WatchProviderRewardAccrued(opts *bind.WatchOpts, sink chan<- *SessionRouterProviderRewardAccrued, sessionId [][32]byte, provider []common.Address) (event.Subscription, error) {

	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "ProviderRewardAccrued", sessionIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterProviderRewardAccrued)
				if err := _SessionRouter.contract.UnpackLog(event, "ProviderRewardAccrued", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseProviderRewardAccrued(log types.Log) (*SessionRouterProviderRewardAccrued, error) {
	event := new(SessionRouterProviderRewardAccrued)
	if err := _SessionRouter.contract.UnpackLog(event, "ProviderRewardAccrued", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterSessionClosedIterator is returned from FilterSessionClosed and is used to iterate over the raw logs and unpacked data for SessionClosed events raised by the SessionRouter contract.
type SessionRouterSessionClosedIterator struct {
	Event *SessionRouterSessionClosed // Event containing the contract specifics and raw log

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
func (it *SessionRouterSessionClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterSessionClosed)
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
		it.Event = new(SessionRouterSessionClosed)
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
func (it *SessionRouterSessionClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterSessionClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterSessionClosed represents a SessionClosed event raised by the SessionRouter contract.
type SessionRouterSessionClosed struct {
	User       common.Address
	SessionId  [32]byte
	ProviderId common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSessionClosed is a free log retrieval operation binding the contract event 0x337fbb0a41a596db800dc836595a57815f967185e3596615c646f2455ac3914a.
//
// Solidity: event SessionClosed(address indexed user, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) FilterSessionClosed(opts *bind.FilterOpts, user []common.Address, sessionId [][32]byte, providerId []common.Address) (*SessionRouterSessionClosedIterator, error) {

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

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "SessionClosed", userRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterSessionClosedIterator{contract: _SessionRouter.contract, event: "SessionClosed", logs: logs, sub: sub}, nil
}

// WatchSessionClosed is a free log subscription operation binding the contract event 0x337fbb0a41a596db800dc836595a57815f967185e3596615c646f2455ac3914a.
//
// Solidity: event SessionClosed(address indexed user, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) WatchSessionClosed(opts *bind.WatchOpts, sink chan<- *SessionRouterSessionClosed, user []common.Address, sessionId [][32]byte, providerId []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "SessionClosed", userRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterSessionClosed)
				if err := _SessionRouter.contract.UnpackLog(event, "SessionClosed", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseSessionClosed(log types.Log) (*SessionRouterSessionClosed, error) {
	event := new(SessionRouterSessionClosed)
	if err := _SessionRouter.contract.UnpackLog(event, "SessionClosed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterSessionOpenedIterator is returned from FilterSessionOpened and is used to iterate over the raw logs and unpacked data for SessionOpened events raised by the SessionRouter contract.
type SessionRouterSessionOpenedIterator struct {
	Event *SessionRouterSessionOpened // Event containing the contract specifics and raw log

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
func (it *SessionRouterSessionOpenedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterSessionOpened)
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
		it.Event = new(SessionRouterSessionOpened)
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
func (it *SessionRouterSessionOpenedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterSessionOpenedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterSessionOpened represents a SessionOpened event raised by the SessionRouter contract.
type SessionRouterSessionOpened struct {
	User       common.Address
	SessionId  [32]byte
	ProviderId common.Address
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterSessionOpened is a free log retrieval operation binding the contract event 0x2bd7c890baf595977d256a6e784512c873ac58ba612b4895dbb7f784bfbf4839.
//
// Solidity: event SessionOpened(address indexed user, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) FilterSessionOpened(opts *bind.FilterOpts, user []common.Address, sessionId [][32]byte, providerId []common.Address) (*SessionRouterSessionOpenedIterator, error) {

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

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "SessionOpened", userRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterSessionOpenedIterator{contract: _SessionRouter.contract, event: "SessionOpened", logs: logs, sub: sub}, nil
}

// WatchSessionOpened is a free log subscription operation binding the contract event 0x2bd7c890baf595977d256a6e784512c873ac58ba612b4895dbb7f784bfbf4839.
//
// Solidity: event SessionOpened(address indexed user, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) WatchSessionOpened(opts *bind.WatchOpts, sink chan<- *SessionRouterSessionOpened, user []common.Address, sessionId [][32]byte, providerId []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "SessionOpened", userRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterSessionOpened)
				if err := _SessionRouter.contract.UnpackLog(event, "SessionOpened", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseSessionOpened(log types.Log) (*SessionRouterSessionOpened, error) {
	event := new(SessionRouterSessionOpened)
	if err := _SessionRouter.contract.UnpackLog(event, "SessionOpened", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterSessionRatingIterator is returned from FilterSessionRating and is used to iterate over the raw logs and unpacked data for SessionRating events raised by the SessionRouter contract.
type SessionRouterSessionRatingIterator struct {
	Event *SessionRouterSessionRating // Event containing the contract specifics and raw log

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
func (it *SessionRouterSessionRatingIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterSessionRating)
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
		it.Event = new(SessionRouterSessionRating)
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
func (it *SessionRouterSessionRatingIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterSessionRatingIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterSessionRating represents a SessionRating event raised by the SessionRouter contract.
type SessionRouterSessionRating struct {
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
func (_SessionRouter *SessionRouterFilterer) FilterSessionRating(opts *bind.FilterOpts, sessionId [][32]byte, modelId [][32]byte, provider []common.Address) (*SessionRouterSessionRatingIterator, error) {

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

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "SessionRating", sessionIdRule, modelIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterSessionRatingIterator{contract: _SessionRouter.contract, event: "SessionRating", logs: logs, sub: sub}, nil
}

// WatchSessionRating is a free log subscription operation binding the contract event 0x62fefdc7afa78415878080df4ed9b3c7dbbad510375c160ca7559f9f4d925ecb.
//
// Solidity: event SessionRating(bytes32 indexed sessionId, bytes32 indexed modelId, address indexed provider, uint256 closeoutType, uint32 tpsScaled1000, uint32 ttftMs)
func (_SessionRouter *SessionRouterFilterer) WatchSessionRating(opts *bind.WatchOpts, sink chan<- *SessionRouterSessionRating, sessionId [][32]byte, modelId [][32]byte, provider []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "SessionRating", sessionIdRule, modelIdRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterSessionRating)
				if err := _SessionRouter.contract.UnpackLog(event, "SessionRating", log); err != nil {
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
func (_SessionRouter *SessionRouterFilterer) ParseSessionRating(log types.Log) (*SessionRouterSessionRating, error) {
	event := new(SessionRouterSessionRating)
	if err := _SessionRouter.contract.UnpackLog(event, "SessionRating", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterUserWithdrawnIterator is returned from FilterUserWithdrawn and is used to iterate over the raw logs and unpacked data for UserWithdrawn events raised by the SessionRouter contract.
type SessionRouterUserWithdrawnIterator struct {
	Event *SessionRouterUserWithdrawn // Event containing the contract specifics and raw log

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
func (it *SessionRouterUserWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterUserWithdrawn)
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
		it.Event = new(SessionRouterUserWithdrawn)
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
func (it *SessionRouterUserWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterUserWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterUserWithdrawn represents a UserWithdrawn event raised by the SessionRouter contract.
type SessionRouterUserWithdrawn struct {
	User   common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterUserWithdrawn is a free log retrieval operation binding the contract event 0xe6b386172074b393dc04ed6cb1a352475ffad5dd8cebc76231a3b683141ea6fb.
//
// Solidity: event UserWithdrawn(address indexed user, uint256 amount_)
func (_SessionRouter *SessionRouterFilterer) FilterUserWithdrawn(opts *bind.FilterOpts, user []common.Address) (*SessionRouterUserWithdrawnIterator, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "UserWithdrawn", userRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterUserWithdrawnIterator{contract: _SessionRouter.contract, event: "UserWithdrawn", logs: logs, sub: sub}, nil
}

// WatchUserWithdrawn is a free log subscription operation binding the contract event 0xe6b386172074b393dc04ed6cb1a352475ffad5dd8cebc76231a3b683141ea6fb.
//
// Solidity: event UserWithdrawn(address indexed user, uint256 amount_)
func (_SessionRouter *SessionRouterFilterer) WatchUserWithdrawn(opts *bind.WatchOpts, sink chan<- *SessionRouterUserWithdrawn, user []common.Address) (event.Subscription, error) {

	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "UserWithdrawn", userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterUserWithdrawn)
				if err := _SessionRouter.contract.UnpackLog(event, "UserWithdrawn", log); err != nil {
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

// ParseUserWithdrawn is a log parse operation binding the contract event 0xe6b386172074b393dc04ed6cb1a352475ffad5dd8cebc76231a3b683141ea6fb.
//
// Solidity: event UserWithdrawn(address indexed user, uint256 amount_)
func (_SessionRouter *SessionRouterFilterer) ParseUserWithdrawn(log types.Log) (*SessionRouterUserWithdrawn, error) {
	event := new(SessionRouterUserWithdrawn)
	if err := _SessionRouter.contract.UnpackLog(event, "UserWithdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
