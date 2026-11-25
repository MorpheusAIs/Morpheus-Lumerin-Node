// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package delegatestakingfunding

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

// IDelegateStakingStorageDelegateStakingLimits is an auto generated low-level Go binding around an user-defined struct.
type IDelegateStakingStorageDelegateStakingLimits struct {
	MaxActiveFunders   *big.Int
	MaxAutoService     *big.Int
	MaxAutoReleaseDays *big.Int
	MaxSettlementGrace *big.Int
	MaxMinPrincipal    *big.Int
}

// IDelegateStakingStorageDelegateStakingParameters is an auto generated low-level Go binding around an user-defined struct.
type IDelegateStakingStorageDelegateStakingParameters struct {
	MaxActiveFunders   *big.Int
	MinPrincipal       *big.Int
	SettlementGrace    *big.Int
	MaxAutoService     *big.Int
	MaxAutoReleaseDays *big.Int
}

// IDelegateStakingStoragePoolBalances is an auto generated low-level Go binding around an user-defined struct.
type IDelegateStakingStoragePoolBalances struct {
	Free         *big.Int
	InSession    *big.Int
	Locked       *big.Int
	PendingTotal *big.Int
}

// IDelegateStakingStoragePoolFunder is an auto generated low-level Go binding around an user-defined struct.
type IDelegateStakingStoragePoolFunder struct {
	Funder       common.Address
	Allowance    IDelegateStakingStorageStakingAllowance
	InActiveList bool
}

// IDelegateStakingStoragePoolLockBucket is an auto generated low-level Go binding around an user-defined struct.
type IDelegateStakingStoragePoolLockBucket struct {
	ReleaseAt *big.Int
	Amount    *big.Int
}

// IDelegateStakingStoragePoolSession is an auto generated low-level Go binding around an user-defined struct.
type IDelegateStakingStoragePoolSession struct {
	Hot     common.Address
	Draws   []IDelegateStakingStoragePoolSessionDraw
	Exists  bool
	Closed  bool
	Settled bool
}

// IDelegateStakingStoragePoolSessionDraw is an auto generated low-level Go binding around an user-defined struct.
type IDelegateStakingStoragePoolSessionDraw struct {
	Funder common.Address
	Amount *big.Int
}

// IDelegateStakingStoragePoolWithdrawal is an auto generated low-level Go binding around an user-defined struct.
type IDelegateStakingStoragePoolWithdrawal struct {
	Funder common.Address
	Amount *big.Int
}

// IDelegateStakingStorageStakingAllowance is an auto generated low-level Go binding around an user-defined struct.
type IDelegateStakingStorageStakingAllowance struct {
	CumulativeFundingCap *big.Int
	Expiry               *big.Int
	Revoked              bool
	LifetimeFunded       *big.Int
	CurrentPrincipal     *big.Int
	Exposure             *big.Int
	LockedExposure       *big.Int
	Pending              *big.Int
	Active               bool
	Approved             bool
}

// DelegateStakingFundingMetaData contains all meta data concerning the DelegateStakingFunding contract.
var DelegateStakingFundingMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"DelegateStakingActiveFundersLimitExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingAllowanceNotActive\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingCapBelowLifetimeFunding\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingFundingCapExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInsufficientCapacity\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInsufficientLiquidity\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInsufficientPrincipal\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidAddress\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidAmount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidExpiry\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidPoolSettlement\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingInvalidWithdrawalCursor\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingLimitExceeded\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPoolFunderNotApproved\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPoolSessionAlreadyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPoolSessionNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingPrincipalTooLow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingSessionSlotNotEmpty\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingSettlementNotReady\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DelegateStakingWithdrawalNotFound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"delegatee\",\"type\":\"address\"}],\"name\":\"InsufficientRightsForOperation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account_\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingAllowanceFunded\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"cumulativeFundingCap\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint128\",\"name\":\"expiry\",\"type\":\"uint128\"}],\"name\":\"DelegateStakingAllowanceGranted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"}],\"name\":\"DelegateStakingAllowanceRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMaxActiveFundersUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMaxAutoReleaseDaysUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMaxAutoServiceUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingMinPrincipalUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolDrawn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"}],\"name\":\"DelegateStakingPoolFunderApprovalRevoked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"}],\"name\":\"DelegateStakingPoolFunderApproved\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint128\",\"name\":\"releaseAt\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolLockCreated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"uint128\",\"name\":\"releaseAt\",\"type\":\"uint128\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolLockReleased\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingPoolReturned\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"caller\",\"type\":\"address\"}],\"name\":\"DelegateStakingPoolSessionSettled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingSettlementGraceUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingWithdrawalCancelled\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingWithdrawalPaid\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"paid\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"pending\",\"type\":\"uint256\"}],\"name\":\"DelegateStakingWithdrawalRequested\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"storageSlot\",\"type\":\"bytes32\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BIDS_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_MARKETPLACE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_MODEL\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_PROVIDER\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_SESSION\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DIAMOND_OWNABLE_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"funder_\",\"type\":\"address\"}],\"name\":\"approvePoolFunder\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"funder_\",\"type\":\"address\"}],\"name\":\"cancelStakingWithdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"}],\"name\":\"fundStakingAllowance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"}],\"name\":\"getAvailablePoolStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId_\",\"type\":\"bytes32\"}],\"name\":\"getBid\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"deletedAt\",\"type\":\"uint128\"}],\"internalType\":\"structIBidStorage.Bid\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDelegateStakingDelegationRight\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDelegateStakingLimits\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"maxActiveFunders\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxAutoService\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxAutoReleaseDays\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxSettlementGrace\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxMinPrincipal\",\"type\":\"uint256\"}],\"internalType\":\"structIDelegateStakingStorage.DelegateStakingLimits\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getDelegateStakingParameters\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"maxActiveFunders\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"minPrincipal\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"settlementGrace\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxAutoService\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"maxAutoReleaseDays\",\"type\":\"uint256\"}],\"internalType\":\"structIDelegateStakingStorage.DelegateStakingParameters\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getModelActiveBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getModelBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"}],\"name\":\"getPoolBalances\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"free\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"inSession\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"locked\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pendingTotal\",\"type\":\"uint256\"}],\"internalType\":\"structIDelegateStakingStorage.PoolBalances\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getPoolFunders\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"cumulativeFundingCap\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"expiry\",\"type\":\"uint128\"},{\"internalType\":\"bool\",\"name\":\"revoked\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"lifetimeFunded\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"currentPrincipal\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"exposure\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockedExposure\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pending\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"internalType\":\"structIDelegateStakingStorage.StakingAllowance\",\"name\":\"allowance\",\"type\":\"tuple\"},{\"internalType\":\"bool\",\"name\":\"inActiveList\",\"type\":\"bool\"}],\"internalType\":\"structIDelegateStakingStorage.PoolFunder[]\",\"name\":\"funders_\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"total_\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getPoolLockBuckets\",\"outputs\":[{\"components\":[{\"internalType\":\"uint128\",\"name\":\"releaseAt\",\"type\":\"uint128\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIDelegateStakingStorage.PoolLockBucket[]\",\"name\":\"buckets_\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256\",\"name\":\"total_\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId_\",\"type\":\"bytes32\"}],\"name\":\"getPoolSession\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"hot\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIDelegateStakingStorage.PoolSessionDraw[]\",\"name\":\"draws\",\"type\":\"tuple[]\"},{\"internalType\":\"bool\",\"name\":\"exists\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"closed\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"settled\",\"type\":\"bool\"}],\"internalType\":\"structIDelegateStakingStorage.PoolSession\",\"name\":\"session_\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"cursor_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getPoolWithdrawalQueue\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"funder\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"internalType\":\"structIDelegateStakingStorage.PoolWithdrawal[]\",\"name\":\"withdrawals_\",\"type\":\"tuple[]\"},{\"internalType\":\"address\",\"name\":\"nextCursor_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"total_\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getProviderActiveBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getProviderBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRegistry\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"funder_\",\"type\":\"address\"}],\"name\":\"getStakingAllowance\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"cumulativeFundingCap\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"expiry\",\"type\":\"uint128\"},{\"internalType\":\"bool\",\"name\":\"revoked\",\"type\":\"bool\"},{\"internalType\":\"uint256\",\"name\":\"lifetimeFunded\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"currentPrincipal\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"exposure\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"lockedExposure\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pending\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"active\",\"type\":\"bool\"},{\"internalType\":\"bool\",\"name\":\"approved\",\"type\":\"bool\"}],\"internalType\":\"structIDelegateStakingStorage.StakingAllowance\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"funder_\",\"type\":\"address\"}],\"name\":\"getStakingWithdrawal\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"cap_\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"expiry_\",\"type\":\"uint128\"}],\"name\":\"grantStakingAllowance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId_\",\"type\":\"bytes32\"}],\"name\":\"isBidActive\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"funder_\",\"type\":\"address\"}],\"name\":\"isPoolFunderApproved\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegatee_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"delegator_\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"rights_\",\"type\":\"bytes32\"}],\"name\":\"isRightsDelegated\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"maxEntries_\",\"type\":\"uint256\"}],\"name\":\"purgeInactiveFunders\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"maxBuckets_\",\"type\":\"uint256\"}],\"name\":\"releasePoolLocks\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"released_\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"funder_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"}],\"name\":\"requestStakingWithdrawal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"funder_\",\"type\":\"address\"}],\"name\":\"revokePoolFunderApproval\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"}],\"name\":\"revokeStakingAllowance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"hot_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"maxEntries_\",\"type\":\"uint256\"}],\"name\":\"servePoolWithdrawals\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"value_\",\"type\":\"uint256\"}],\"name\":\"setMaxActiveFunders\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"value_\",\"type\":\"uint256\"}],\"name\":\"setMaxAutoReleaseDays\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"value_\",\"type\":\"uint256\"}],\"name\":\"setMaxAutoService\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"value_\",\"type\":\"uint256\"}],\"name\":\"setMinPrincipal\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"value_\",\"type\":\"uint256\"}],\"name\":\"setSettlementGrace\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// DelegateStakingFundingABI is the input ABI used to generate the binding from.
// Deprecated: Use DelegateStakingFundingMetaData.ABI instead.
var DelegateStakingFundingABI = DelegateStakingFundingMetaData.ABI

// DelegateStakingFunding is an auto generated Go binding around an Ethereum contract.
type DelegateStakingFunding struct {
	DelegateStakingFundingCaller     // Read-only binding to the contract
	DelegateStakingFundingTransactor // Write-only binding to the contract
	DelegateStakingFundingFilterer   // Log filterer for contract events
}

// DelegateStakingFundingCaller is an auto generated read-only Go binding around an Ethereum contract.
type DelegateStakingFundingCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DelegateStakingFundingTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DelegateStakingFundingTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DelegateStakingFundingFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DelegateStakingFundingFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DelegateStakingFundingSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DelegateStakingFundingSession struct {
	Contract     *DelegateStakingFunding // Generic contract binding to set the session for
	CallOpts     bind.CallOpts           // Call options to use throughout this session
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// DelegateStakingFundingCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DelegateStakingFundingCallerSession struct {
	Contract *DelegateStakingFundingCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                 // Call options to use throughout this session
}

// DelegateStakingFundingTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DelegateStakingFundingTransactorSession struct {
	Contract     *DelegateStakingFundingTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                 // Transaction auth options to use throughout this session
}

// DelegateStakingFundingRaw is an auto generated low-level Go binding around an Ethereum contract.
type DelegateStakingFundingRaw struct {
	Contract *DelegateStakingFunding // Generic contract binding to access the raw methods on
}

// DelegateStakingFundingCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DelegateStakingFundingCallerRaw struct {
	Contract *DelegateStakingFundingCaller // Generic read-only contract binding to access the raw methods on
}

// DelegateStakingFundingTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DelegateStakingFundingTransactorRaw struct {
	Contract *DelegateStakingFundingTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDelegateStakingFunding creates a new instance of DelegateStakingFunding, bound to a specific deployed contract.
func NewDelegateStakingFunding(address common.Address, backend bind.ContractBackend) (*DelegateStakingFunding, error) {
	contract, err := bindDelegateStakingFunding(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFunding{DelegateStakingFundingCaller: DelegateStakingFundingCaller{contract: contract}, DelegateStakingFundingTransactor: DelegateStakingFundingTransactor{contract: contract}, DelegateStakingFundingFilterer: DelegateStakingFundingFilterer{contract: contract}}, nil
}

// NewDelegateStakingFundingCaller creates a new read-only instance of DelegateStakingFunding, bound to a specific deployed contract.
func NewDelegateStakingFundingCaller(address common.Address, caller bind.ContractCaller) (*DelegateStakingFundingCaller, error) {
	contract, err := bindDelegateStakingFunding(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingCaller{contract: contract}, nil
}

// NewDelegateStakingFundingTransactor creates a new write-only instance of DelegateStakingFunding, bound to a specific deployed contract.
func NewDelegateStakingFundingTransactor(address common.Address, transactor bind.ContractTransactor) (*DelegateStakingFundingTransactor, error) {
	contract, err := bindDelegateStakingFunding(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingTransactor{contract: contract}, nil
}

// NewDelegateStakingFundingFilterer creates a new log filterer instance of DelegateStakingFunding, bound to a specific deployed contract.
func NewDelegateStakingFundingFilterer(address common.Address, filterer bind.ContractFilterer) (*DelegateStakingFundingFilterer, error) {
	contract, err := bindDelegateStakingFunding(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingFilterer{contract: contract}, nil
}

// bindDelegateStakingFunding binds a generic wrapper to an already deployed contract.
func bindDelegateStakingFunding(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := DelegateStakingFundingMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DelegateStakingFunding *DelegateStakingFundingRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DelegateStakingFunding.Contract.DelegateStakingFundingCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DelegateStakingFunding *DelegateStakingFundingRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.DelegateStakingFundingTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DelegateStakingFunding *DelegateStakingFundingRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.DelegateStakingFundingTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DelegateStakingFunding *DelegateStakingFundingCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _DelegateStakingFunding.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DelegateStakingFunding *DelegateStakingFundingTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DelegateStakingFunding *DelegateStakingFundingTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.contract.Transact(opts, method, params...)
}

// BIDSSTORAGESLOT is a free data retrieval call binding the contract method 0x266ccff0.
//
// Solidity: function BIDS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) BIDSSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "BIDS_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BIDSSTORAGESLOT is a free data retrieval call binding the contract method 0x266ccff0.
//
// Solidity: function BIDS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingSession) BIDSSTORAGESLOT() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.BIDSSTORAGESLOT(&_DelegateStakingFunding.CallOpts)
}

// BIDSSTORAGESLOT is a free data retrieval call binding the contract method 0x266ccff0.
//
// Solidity: function BIDS_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) BIDSSTORAGESLOT() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.BIDSSTORAGESLOT(&_DelegateStakingFunding.CallOpts)
}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) DELEGATIONRULESMARKETPLACE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "DELEGATION_RULES_MARKETPLACE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingSession) DELEGATIONRULESMARKETPLACE() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.DELEGATIONRULESMARKETPLACE(&_DelegateStakingFunding.CallOpts)
}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) DELEGATIONRULESMARKETPLACE() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.DELEGATIONRULESMARKETPLACE(&_DelegateStakingFunding.CallOpts)
}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) DELEGATIONRULESMODEL(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "DELEGATION_RULES_MODEL")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingSession) DELEGATIONRULESMODEL() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.DELEGATIONRULESMODEL(&_DelegateStakingFunding.CallOpts)
}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) DELEGATIONRULESMODEL() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.DELEGATIONRULESMODEL(&_DelegateStakingFunding.CallOpts)
}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) DELEGATIONRULESPROVIDER(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "DELEGATION_RULES_PROVIDER")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingSession) DELEGATIONRULESPROVIDER() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.DELEGATIONRULESPROVIDER(&_DelegateStakingFunding.CallOpts)
}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) DELEGATIONRULESPROVIDER() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.DELEGATIONRULESPROVIDER(&_DelegateStakingFunding.CallOpts)
}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) DELEGATIONRULESSESSION(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "DELEGATION_RULES_SESSION")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingSession) DELEGATIONRULESSESSION() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.DELEGATIONRULESSESSION(&_DelegateStakingFunding.CallOpts)
}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) DELEGATIONRULESSESSION() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.DELEGATIONRULESSESSION(&_DelegateStakingFunding.CallOpts)
}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) DELEGATIONSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "DELEGATION_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingSession) DELEGATIONSTORAGESLOT() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.DELEGATIONSTORAGESLOT(&_DelegateStakingFunding.CallOpts)
}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) DELEGATIONSTORAGESLOT() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.DELEGATIONSTORAGESLOT(&_DelegateStakingFunding.CallOpts)
}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) DIAMONDOWNABLESTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "DIAMOND_OWNABLE_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingSession) DIAMONDOWNABLESTORAGESLOT() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.DIAMONDOWNABLESTORAGESLOT(&_DelegateStakingFunding.CallOpts)
}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) DIAMONDOWNABLESTORAGESLOT() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.DIAMONDOWNABLESTORAGESLOT(&_DelegateStakingFunding.CallOpts)
}

// GetAvailablePoolStake is a free data retrieval call binding the contract method 0x33707678.
//
// Solidity: function getAvailablePoolStake(address hot_) view returns(uint256)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetAvailablePoolStake(opts *bind.CallOpts, hot_ common.Address) (*big.Int, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getAvailablePoolStake", hot_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetAvailablePoolStake is a free data retrieval call binding the contract method 0x33707678.
//
// Solidity: function getAvailablePoolStake(address hot_) view returns(uint256)
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetAvailablePoolStake(hot_ common.Address) (*big.Int, error) {
	return _DelegateStakingFunding.Contract.GetAvailablePoolStake(&_DelegateStakingFunding.CallOpts, hot_)
}

// GetAvailablePoolStake is a free data retrieval call binding the contract method 0x33707678.
//
// Solidity: function getAvailablePoolStake(address hot_) view returns(uint256)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetAvailablePoolStake(hot_ common.Address) (*big.Int, error) {
	return _DelegateStakingFunding.Contract.GetAvailablePoolStake(&_DelegateStakingFunding.CallOpts, hot_)
}

// GetBid is a free data retrieval call binding the contract method 0x91704e1e.
//
// Solidity: function getBid(bytes32 bidId_) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetBid(opts *bind.CallOpts, bidId_ [32]byte) (IBidStorageBid, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getBid", bidId_)

	if err != nil {
		return *new(IBidStorageBid), err
	}

	out0 := *abi.ConvertType(out[0], new(IBidStorageBid)).(*IBidStorageBid)

	return out0, err

}

// GetBid is a free data retrieval call binding the contract method 0x91704e1e.
//
// Solidity: function getBid(bytes32 bidId_) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetBid(bidId_ [32]byte) (IBidStorageBid, error) {
	return _DelegateStakingFunding.Contract.GetBid(&_DelegateStakingFunding.CallOpts, bidId_)
}

// GetBid is a free data retrieval call binding the contract method 0x91704e1e.
//
// Solidity: function getBid(bytes32 bidId_) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetBid(bidId_ [32]byte) (IBidStorageBid, error) {
	return _DelegateStakingFunding.Contract.GetBid(&_DelegateStakingFunding.CallOpts, bidId_)
}

// GetDelegateStakingDelegationRight is a free data retrieval call binding the contract method 0x6d687e78.
//
// Solidity: function getDelegateStakingDelegationRight() pure returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetDelegateStakingDelegationRight(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getDelegateStakingDelegationRight")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetDelegateStakingDelegationRight is a free data retrieval call binding the contract method 0x6d687e78.
//
// Solidity: function getDelegateStakingDelegationRight() pure returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetDelegateStakingDelegationRight() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.GetDelegateStakingDelegationRight(&_DelegateStakingFunding.CallOpts)
}

// GetDelegateStakingDelegationRight is a free data retrieval call binding the contract method 0x6d687e78.
//
// Solidity: function getDelegateStakingDelegationRight() pure returns(bytes32)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetDelegateStakingDelegationRight() ([32]byte, error) {
	return _DelegateStakingFunding.Contract.GetDelegateStakingDelegationRight(&_DelegateStakingFunding.CallOpts)
}

// GetDelegateStakingLimits is a free data retrieval call binding the contract method 0xce9361b0.
//
// Solidity: function getDelegateStakingLimits() pure returns((uint256,uint256,uint256,uint256,uint256))
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetDelegateStakingLimits(opts *bind.CallOpts) (IDelegateStakingStorageDelegateStakingLimits, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getDelegateStakingLimits")

	if err != nil {
		return *new(IDelegateStakingStorageDelegateStakingLimits), err
	}

	out0 := *abi.ConvertType(out[0], new(IDelegateStakingStorageDelegateStakingLimits)).(*IDelegateStakingStorageDelegateStakingLimits)

	return out0, err

}

// GetDelegateStakingLimits is a free data retrieval call binding the contract method 0xce9361b0.
//
// Solidity: function getDelegateStakingLimits() pure returns((uint256,uint256,uint256,uint256,uint256))
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetDelegateStakingLimits() (IDelegateStakingStorageDelegateStakingLimits, error) {
	return _DelegateStakingFunding.Contract.GetDelegateStakingLimits(&_DelegateStakingFunding.CallOpts)
}

// GetDelegateStakingLimits is a free data retrieval call binding the contract method 0xce9361b0.
//
// Solidity: function getDelegateStakingLimits() pure returns((uint256,uint256,uint256,uint256,uint256))
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetDelegateStakingLimits() (IDelegateStakingStorageDelegateStakingLimits, error) {
	return _DelegateStakingFunding.Contract.GetDelegateStakingLimits(&_DelegateStakingFunding.CallOpts)
}

// GetDelegateStakingParameters is a free data retrieval call binding the contract method 0xa6800ab0.
//
// Solidity: function getDelegateStakingParameters() view returns((uint256,uint256,uint256,uint256,uint256))
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetDelegateStakingParameters(opts *bind.CallOpts) (IDelegateStakingStorageDelegateStakingParameters, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getDelegateStakingParameters")

	if err != nil {
		return *new(IDelegateStakingStorageDelegateStakingParameters), err
	}

	out0 := *abi.ConvertType(out[0], new(IDelegateStakingStorageDelegateStakingParameters)).(*IDelegateStakingStorageDelegateStakingParameters)

	return out0, err

}

// GetDelegateStakingParameters is a free data retrieval call binding the contract method 0xa6800ab0.
//
// Solidity: function getDelegateStakingParameters() view returns((uint256,uint256,uint256,uint256,uint256))
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetDelegateStakingParameters() (IDelegateStakingStorageDelegateStakingParameters, error) {
	return _DelegateStakingFunding.Contract.GetDelegateStakingParameters(&_DelegateStakingFunding.CallOpts)
}

// GetDelegateStakingParameters is a free data retrieval call binding the contract method 0xa6800ab0.
//
// Solidity: function getDelegateStakingParameters() view returns((uint256,uint256,uint256,uint256,uint256))
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetDelegateStakingParameters() (IDelegateStakingStorageDelegateStakingParameters, error) {
	return _DelegateStakingFunding.Contract.GetDelegateStakingParameters(&_DelegateStakingFunding.CallOpts)
}

// GetModelActiveBids is a free data retrieval call binding the contract method 0x8a683b6e.
//
// Solidity: function getModelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetModelActiveBids(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getModelActiveBids", modelId_, offset_, limit_)

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
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetModelActiveBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStakingFunding.Contract.GetModelActiveBids(&_DelegateStakingFunding.CallOpts, modelId_, offset_, limit_)
}

// GetModelActiveBids is a free data retrieval call binding the contract method 0x8a683b6e.
//
// Solidity: function getModelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetModelActiveBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStakingFunding.Contract.GetModelActiveBids(&_DelegateStakingFunding.CallOpts, modelId_, offset_, limit_)
}

// GetModelBids is a free data retrieval call binding the contract method 0xfade17b1.
//
// Solidity: function getModelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetModelBids(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getModelBids", modelId_, offset_, limit_)

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
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetModelBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStakingFunding.Contract.GetModelBids(&_DelegateStakingFunding.CallOpts, modelId_, offset_, limit_)
}

// GetModelBids is a free data retrieval call binding the contract method 0xfade17b1.
//
// Solidity: function getModelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetModelBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStakingFunding.Contract.GetModelBids(&_DelegateStakingFunding.CallOpts, modelId_, offset_, limit_)
}

// GetPoolBalances is a free data retrieval call binding the contract method 0x7b50f8c5.
//
// Solidity: function getPoolBalances(address hot_) view returns((uint256,uint256,uint256,uint256))
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetPoolBalances(opts *bind.CallOpts, hot_ common.Address) (IDelegateStakingStoragePoolBalances, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getPoolBalances", hot_)

	if err != nil {
		return *new(IDelegateStakingStoragePoolBalances), err
	}

	out0 := *abi.ConvertType(out[0], new(IDelegateStakingStoragePoolBalances)).(*IDelegateStakingStoragePoolBalances)

	return out0, err

}

// GetPoolBalances is a free data retrieval call binding the contract method 0x7b50f8c5.
//
// Solidity: function getPoolBalances(address hot_) view returns((uint256,uint256,uint256,uint256))
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetPoolBalances(hot_ common.Address) (IDelegateStakingStoragePoolBalances, error) {
	return _DelegateStakingFunding.Contract.GetPoolBalances(&_DelegateStakingFunding.CallOpts, hot_)
}

// GetPoolBalances is a free data retrieval call binding the contract method 0x7b50f8c5.
//
// Solidity: function getPoolBalances(address hot_) view returns((uint256,uint256,uint256,uint256))
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetPoolBalances(hot_ common.Address) (IDelegateStakingStoragePoolBalances, error) {
	return _DelegateStakingFunding.Contract.GetPoolBalances(&_DelegateStakingFunding.CallOpts, hot_)
}

// GetPoolFunders is a free data retrieval call binding the contract method 0xd00fff9b.
//
// Solidity: function getPoolFunders(address hot_, uint256 offset_, uint256 limit_) view returns((address,(uint256,uint128,bool,uint256,uint256,uint256,uint256,uint256,bool,bool),bool)[] funders_, uint256 total_)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetPoolFunders(opts *bind.CallOpts, hot_ common.Address, offset_ *big.Int, limit_ *big.Int) (struct {
	Funders []IDelegateStakingStoragePoolFunder
	Total   *big.Int
}, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getPoolFunders", hot_, offset_, limit_)

	outstruct := new(struct {
		Funders []IDelegateStakingStoragePoolFunder
		Total   *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Funders = *abi.ConvertType(out[0], new([]IDelegateStakingStoragePoolFunder)).(*[]IDelegateStakingStoragePoolFunder)
	outstruct.Total = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetPoolFunders is a free data retrieval call binding the contract method 0xd00fff9b.
//
// Solidity: function getPoolFunders(address hot_, uint256 offset_, uint256 limit_) view returns((address,(uint256,uint128,bool,uint256,uint256,uint256,uint256,uint256,bool,bool),bool)[] funders_, uint256 total_)
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetPoolFunders(hot_ common.Address, offset_ *big.Int, limit_ *big.Int) (struct {
	Funders []IDelegateStakingStoragePoolFunder
	Total   *big.Int
}, error) {
	return _DelegateStakingFunding.Contract.GetPoolFunders(&_DelegateStakingFunding.CallOpts, hot_, offset_, limit_)
}

// GetPoolFunders is a free data retrieval call binding the contract method 0xd00fff9b.
//
// Solidity: function getPoolFunders(address hot_, uint256 offset_, uint256 limit_) view returns((address,(uint256,uint128,bool,uint256,uint256,uint256,uint256,uint256,bool,bool),bool)[] funders_, uint256 total_)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetPoolFunders(hot_ common.Address, offset_ *big.Int, limit_ *big.Int) (struct {
	Funders []IDelegateStakingStoragePoolFunder
	Total   *big.Int
}, error) {
	return _DelegateStakingFunding.Contract.GetPoolFunders(&_DelegateStakingFunding.CallOpts, hot_, offset_, limit_)
}

// GetPoolLockBuckets is a free data retrieval call binding the contract method 0x6b6403a7.
//
// Solidity: function getPoolLockBuckets(address hot_, uint256 offset_, uint256 limit_) view returns((uint128,uint256)[] buckets_, uint256 total_)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetPoolLockBuckets(opts *bind.CallOpts, hot_ common.Address, offset_ *big.Int, limit_ *big.Int) (struct {
	Buckets []IDelegateStakingStoragePoolLockBucket
	Total   *big.Int
}, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getPoolLockBuckets", hot_, offset_, limit_)

	outstruct := new(struct {
		Buckets []IDelegateStakingStoragePoolLockBucket
		Total   *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Buckets = *abi.ConvertType(out[0], new([]IDelegateStakingStoragePoolLockBucket)).(*[]IDelegateStakingStoragePoolLockBucket)
	outstruct.Total = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetPoolLockBuckets is a free data retrieval call binding the contract method 0x6b6403a7.
//
// Solidity: function getPoolLockBuckets(address hot_, uint256 offset_, uint256 limit_) view returns((uint128,uint256)[] buckets_, uint256 total_)
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetPoolLockBuckets(hot_ common.Address, offset_ *big.Int, limit_ *big.Int) (struct {
	Buckets []IDelegateStakingStoragePoolLockBucket
	Total   *big.Int
}, error) {
	return _DelegateStakingFunding.Contract.GetPoolLockBuckets(&_DelegateStakingFunding.CallOpts, hot_, offset_, limit_)
}

// GetPoolLockBuckets is a free data retrieval call binding the contract method 0x6b6403a7.
//
// Solidity: function getPoolLockBuckets(address hot_, uint256 offset_, uint256 limit_) view returns((uint128,uint256)[] buckets_, uint256 total_)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetPoolLockBuckets(hot_ common.Address, offset_ *big.Int, limit_ *big.Int) (struct {
	Buckets []IDelegateStakingStoragePoolLockBucket
	Total   *big.Int
}, error) {
	return _DelegateStakingFunding.Contract.GetPoolLockBuckets(&_DelegateStakingFunding.CallOpts, hot_, offset_, limit_)
}

// GetPoolSession is a free data retrieval call binding the contract method 0x1752ba33.
//
// Solidity: function getPoolSession(bytes32 sessionId_) view returns((address,(address,uint256)[],bool,bool,bool) session_)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetPoolSession(opts *bind.CallOpts, sessionId_ [32]byte) (IDelegateStakingStoragePoolSession, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getPoolSession", sessionId_)

	if err != nil {
		return *new(IDelegateStakingStoragePoolSession), err
	}

	out0 := *abi.ConvertType(out[0], new(IDelegateStakingStoragePoolSession)).(*IDelegateStakingStoragePoolSession)

	return out0, err

}

// GetPoolSession is a free data retrieval call binding the contract method 0x1752ba33.
//
// Solidity: function getPoolSession(bytes32 sessionId_) view returns((address,(address,uint256)[],bool,bool,bool) session_)
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetPoolSession(sessionId_ [32]byte) (IDelegateStakingStoragePoolSession, error) {
	return _DelegateStakingFunding.Contract.GetPoolSession(&_DelegateStakingFunding.CallOpts, sessionId_)
}

// GetPoolSession is a free data retrieval call binding the contract method 0x1752ba33.
//
// Solidity: function getPoolSession(bytes32 sessionId_) view returns((address,(address,uint256)[],bool,bool,bool) session_)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetPoolSession(sessionId_ [32]byte) (IDelegateStakingStoragePoolSession, error) {
	return _DelegateStakingFunding.Contract.GetPoolSession(&_DelegateStakingFunding.CallOpts, sessionId_)
}

// GetPoolWithdrawalQueue is a free data retrieval call binding the contract method 0x23ddb0fc.
//
// Solidity: function getPoolWithdrawalQueue(address hot_, address cursor_, uint256 limit_) view returns((address,uint256)[] withdrawals_, address nextCursor_, uint256 total_)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetPoolWithdrawalQueue(opts *bind.CallOpts, hot_ common.Address, cursor_ common.Address, limit_ *big.Int) (struct {
	Withdrawals []IDelegateStakingStoragePoolWithdrawal
	NextCursor  common.Address
	Total       *big.Int
}, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getPoolWithdrawalQueue", hot_, cursor_, limit_)

	outstruct := new(struct {
		Withdrawals []IDelegateStakingStoragePoolWithdrawal
		NextCursor  common.Address
		Total       *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Withdrawals = *abi.ConvertType(out[0], new([]IDelegateStakingStoragePoolWithdrawal)).(*[]IDelegateStakingStoragePoolWithdrawal)
	outstruct.NextCursor = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.Total = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetPoolWithdrawalQueue is a free data retrieval call binding the contract method 0x23ddb0fc.
//
// Solidity: function getPoolWithdrawalQueue(address hot_, address cursor_, uint256 limit_) view returns((address,uint256)[] withdrawals_, address nextCursor_, uint256 total_)
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetPoolWithdrawalQueue(hot_ common.Address, cursor_ common.Address, limit_ *big.Int) (struct {
	Withdrawals []IDelegateStakingStoragePoolWithdrawal
	NextCursor  common.Address
	Total       *big.Int
}, error) {
	return _DelegateStakingFunding.Contract.GetPoolWithdrawalQueue(&_DelegateStakingFunding.CallOpts, hot_, cursor_, limit_)
}

// GetPoolWithdrawalQueue is a free data retrieval call binding the contract method 0x23ddb0fc.
//
// Solidity: function getPoolWithdrawalQueue(address hot_, address cursor_, uint256 limit_) view returns((address,uint256)[] withdrawals_, address nextCursor_, uint256 total_)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetPoolWithdrawalQueue(hot_ common.Address, cursor_ common.Address, limit_ *big.Int) (struct {
	Withdrawals []IDelegateStakingStoragePoolWithdrawal
	NextCursor  common.Address
	Total       *big.Int
}, error) {
	return _DelegateStakingFunding.Contract.GetPoolWithdrawalQueue(&_DelegateStakingFunding.CallOpts, hot_, cursor_, limit_)
}

// GetProviderActiveBids is a free data retrieval call binding the contract method 0xaf5b77ca.
//
// Solidity: function getProviderActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetProviderActiveBids(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getProviderActiveBids", provider_, offset_, limit_)

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
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetProviderActiveBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStakingFunding.Contract.GetProviderActiveBids(&_DelegateStakingFunding.CallOpts, provider_, offset_, limit_)
}

// GetProviderActiveBids is a free data retrieval call binding the contract method 0xaf5b77ca.
//
// Solidity: function getProviderActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetProviderActiveBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStakingFunding.Contract.GetProviderActiveBids(&_DelegateStakingFunding.CallOpts, provider_, offset_, limit_)
}

// GetProviderBids is a free data retrieval call binding the contract method 0x59d435c4.
//
// Solidity: function getProviderBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetProviderBids(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getProviderBids", provider_, offset_, limit_)

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
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetProviderBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStakingFunding.Contract.GetProviderBids(&_DelegateStakingFunding.CallOpts, provider_, offset_, limit_)
}

// GetProviderBids is a free data retrieval call binding the contract method 0x59d435c4.
//
// Solidity: function getProviderBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[], uint256)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetProviderBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, *big.Int, error) {
	return _DelegateStakingFunding.Contract.GetProviderBids(&_DelegateStakingFunding.CallOpts, provider_, offset_, limit_)
}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetRegistry() (common.Address, error) {
	return _DelegateStakingFunding.Contract.GetRegistry(&_DelegateStakingFunding.CallOpts)
}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetRegistry() (common.Address, error) {
	return _DelegateStakingFunding.Contract.GetRegistry(&_DelegateStakingFunding.CallOpts)
}

// GetStakingAllowance is a free data retrieval call binding the contract method 0x064f8571.
//
// Solidity: function getStakingAllowance(address hot_, address funder_) view returns((uint256,uint128,bool,uint256,uint256,uint256,uint256,uint256,bool,bool))
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetStakingAllowance(opts *bind.CallOpts, hot_ common.Address, funder_ common.Address) (IDelegateStakingStorageStakingAllowance, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getStakingAllowance", hot_, funder_)

	if err != nil {
		return *new(IDelegateStakingStorageStakingAllowance), err
	}

	out0 := *abi.ConvertType(out[0], new(IDelegateStakingStorageStakingAllowance)).(*IDelegateStakingStorageStakingAllowance)

	return out0, err

}

// GetStakingAllowance is a free data retrieval call binding the contract method 0x064f8571.
//
// Solidity: function getStakingAllowance(address hot_, address funder_) view returns((uint256,uint128,bool,uint256,uint256,uint256,uint256,uint256,bool,bool))
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetStakingAllowance(hot_ common.Address, funder_ common.Address) (IDelegateStakingStorageStakingAllowance, error) {
	return _DelegateStakingFunding.Contract.GetStakingAllowance(&_DelegateStakingFunding.CallOpts, hot_, funder_)
}

// GetStakingAllowance is a free data retrieval call binding the contract method 0x064f8571.
//
// Solidity: function getStakingAllowance(address hot_, address funder_) view returns((uint256,uint128,bool,uint256,uint256,uint256,uint256,uint256,bool,bool))
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetStakingAllowance(hot_ common.Address, funder_ common.Address) (IDelegateStakingStorageStakingAllowance, error) {
	return _DelegateStakingFunding.Contract.GetStakingAllowance(&_DelegateStakingFunding.CallOpts, hot_, funder_)
}

// GetStakingWithdrawal is a free data retrieval call binding the contract method 0x90bdca5b.
//
// Solidity: function getStakingWithdrawal(address hot_, address funder_) view returns(uint256)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetStakingWithdrawal(opts *bind.CallOpts, hot_ common.Address, funder_ common.Address) (*big.Int, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getStakingWithdrawal", hot_, funder_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetStakingWithdrawal is a free data retrieval call binding the contract method 0x90bdca5b.
//
// Solidity: function getStakingWithdrawal(address hot_, address funder_) view returns(uint256)
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetStakingWithdrawal(hot_ common.Address, funder_ common.Address) (*big.Int, error) {
	return _DelegateStakingFunding.Contract.GetStakingWithdrawal(&_DelegateStakingFunding.CallOpts, hot_, funder_)
}

// GetStakingWithdrawal is a free data retrieval call binding the contract method 0x90bdca5b.
//
// Solidity: function getStakingWithdrawal(address hot_, address funder_) view returns(uint256)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetStakingWithdrawal(hot_ common.Address, funder_ common.Address) (*big.Int, error) {
	return _DelegateStakingFunding.Contract.GetStakingWithdrawal(&_DelegateStakingFunding.CallOpts, hot_, funder_)
}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) GetToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "getToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_DelegateStakingFunding *DelegateStakingFundingSession) GetToken() (common.Address, error) {
	return _DelegateStakingFunding.Contract.GetToken(&_DelegateStakingFunding.CallOpts)
}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) GetToken() (common.Address, error) {
	return _DelegateStakingFunding.Contract.GetToken(&_DelegateStakingFunding.CallOpts)
}

// IsBidActive is a free data retrieval call binding the contract method 0x1345df58.
//
// Solidity: function isBidActive(bytes32 bidId_) view returns(bool)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) IsBidActive(opts *bind.CallOpts, bidId_ [32]byte) (bool, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "isBidActive", bidId_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsBidActive is a free data retrieval call binding the contract method 0x1345df58.
//
// Solidity: function isBidActive(bytes32 bidId_) view returns(bool)
func (_DelegateStakingFunding *DelegateStakingFundingSession) IsBidActive(bidId_ [32]byte) (bool, error) {
	return _DelegateStakingFunding.Contract.IsBidActive(&_DelegateStakingFunding.CallOpts, bidId_)
}

// IsBidActive is a free data retrieval call binding the contract method 0x1345df58.
//
// Solidity: function isBidActive(bytes32 bidId_) view returns(bool)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) IsBidActive(bidId_ [32]byte) (bool, error) {
	return _DelegateStakingFunding.Contract.IsBidActive(&_DelegateStakingFunding.CallOpts, bidId_)
}

// IsPoolFunderApproved is a free data retrieval call binding the contract method 0x787636bc.
//
// Solidity: function isPoolFunderApproved(address hot_, address funder_) view returns(bool)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) IsPoolFunderApproved(opts *bind.CallOpts, hot_ common.Address, funder_ common.Address) (bool, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "isPoolFunderApproved", hot_, funder_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsPoolFunderApproved is a free data retrieval call binding the contract method 0x787636bc.
//
// Solidity: function isPoolFunderApproved(address hot_, address funder_) view returns(bool)
func (_DelegateStakingFunding *DelegateStakingFundingSession) IsPoolFunderApproved(hot_ common.Address, funder_ common.Address) (bool, error) {
	return _DelegateStakingFunding.Contract.IsPoolFunderApproved(&_DelegateStakingFunding.CallOpts, hot_, funder_)
}

// IsPoolFunderApproved is a free data retrieval call binding the contract method 0x787636bc.
//
// Solidity: function isPoolFunderApproved(address hot_, address funder_) view returns(bool)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) IsPoolFunderApproved(hot_ common.Address, funder_ common.Address) (bool, error) {
	return _DelegateStakingFunding.Contract.IsPoolFunderApproved(&_DelegateStakingFunding.CallOpts, hot_, funder_)
}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) IsRightsDelegated(opts *bind.CallOpts, delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "isRightsDelegated", delegatee_, delegator_, rights_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_DelegateStakingFunding *DelegateStakingFundingSession) IsRightsDelegated(delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	return _DelegateStakingFunding.Contract.IsRightsDelegated(&_DelegateStakingFunding.CallOpts, delegatee_, delegator_, rights_)
}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) IsRightsDelegated(delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	return _DelegateStakingFunding.Contract.IsRightsDelegated(&_DelegateStakingFunding.CallOpts, delegatee_, delegator_, rights_)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_DelegateStakingFunding *DelegateStakingFundingCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _DelegateStakingFunding.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_DelegateStakingFunding *DelegateStakingFundingSession) Owner() (common.Address, error) {
	return _DelegateStakingFunding.Contract.Owner(&_DelegateStakingFunding.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_DelegateStakingFunding *DelegateStakingFundingCallerSession) Owner() (common.Address, error) {
	return _DelegateStakingFunding.Contract.Owner(&_DelegateStakingFunding.CallOpts)
}

// ApprovePoolFunder is a paid mutator transaction binding the contract method 0x4aaf3469.
//
// Solidity: function approvePoolFunder(address hot_, address funder_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) ApprovePoolFunder(opts *bind.TransactOpts, hot_ common.Address, funder_ common.Address) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "approvePoolFunder", hot_, funder_)
}

// ApprovePoolFunder is a paid mutator transaction binding the contract method 0x4aaf3469.
//
// Solidity: function approvePoolFunder(address hot_, address funder_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingSession) ApprovePoolFunder(hot_ common.Address, funder_ common.Address) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.ApprovePoolFunder(&_DelegateStakingFunding.TransactOpts, hot_, funder_)
}

// ApprovePoolFunder is a paid mutator transaction binding the contract method 0x4aaf3469.
//
// Solidity: function approvePoolFunder(address hot_, address funder_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) ApprovePoolFunder(hot_ common.Address, funder_ common.Address) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.ApprovePoolFunder(&_DelegateStakingFunding.TransactOpts, hot_, funder_)
}

// CancelStakingWithdrawal is a paid mutator transaction binding the contract method 0xcd2af04d.
//
// Solidity: function cancelStakingWithdrawal(address hot_, address funder_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) CancelStakingWithdrawal(opts *bind.TransactOpts, hot_ common.Address, funder_ common.Address) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "cancelStakingWithdrawal", hot_, funder_)
}

// CancelStakingWithdrawal is a paid mutator transaction binding the contract method 0xcd2af04d.
//
// Solidity: function cancelStakingWithdrawal(address hot_, address funder_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingSession) CancelStakingWithdrawal(hot_ common.Address, funder_ common.Address) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.CancelStakingWithdrawal(&_DelegateStakingFunding.TransactOpts, hot_, funder_)
}

// CancelStakingWithdrawal is a paid mutator transaction binding the contract method 0xcd2af04d.
//
// Solidity: function cancelStakingWithdrawal(address hot_, address funder_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) CancelStakingWithdrawal(hot_ common.Address, funder_ common.Address) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.CancelStakingWithdrawal(&_DelegateStakingFunding.TransactOpts, hot_, funder_)
}

// FundStakingAllowance is a paid mutator transaction binding the contract method 0x42ee6c53.
//
// Solidity: function fundStakingAllowance(address hot_, uint256 amount_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) FundStakingAllowance(opts *bind.TransactOpts, hot_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "fundStakingAllowance", hot_, amount_)
}

// FundStakingAllowance is a paid mutator transaction binding the contract method 0x42ee6c53.
//
// Solidity: function fundStakingAllowance(address hot_, uint256 amount_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingSession) FundStakingAllowance(hot_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.FundStakingAllowance(&_DelegateStakingFunding.TransactOpts, hot_, amount_)
}

// FundStakingAllowance is a paid mutator transaction binding the contract method 0x42ee6c53.
//
// Solidity: function fundStakingAllowance(address hot_, uint256 amount_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) FundStakingAllowance(hot_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.FundStakingAllowance(&_DelegateStakingFunding.TransactOpts, hot_, amount_)
}

// GrantStakingAllowance is a paid mutator transaction binding the contract method 0x2623a8a7.
//
// Solidity: function grantStakingAllowance(address hot_, uint256 cap_, uint128 expiry_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) GrantStakingAllowance(opts *bind.TransactOpts, hot_ common.Address, cap_ *big.Int, expiry_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "grantStakingAllowance", hot_, cap_, expiry_)
}

// GrantStakingAllowance is a paid mutator transaction binding the contract method 0x2623a8a7.
//
// Solidity: function grantStakingAllowance(address hot_, uint256 cap_, uint128 expiry_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingSession) GrantStakingAllowance(hot_ common.Address, cap_ *big.Int, expiry_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.GrantStakingAllowance(&_DelegateStakingFunding.TransactOpts, hot_, cap_, expiry_)
}

// GrantStakingAllowance is a paid mutator transaction binding the contract method 0x2623a8a7.
//
// Solidity: function grantStakingAllowance(address hot_, uint256 cap_, uint128 expiry_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) GrantStakingAllowance(hot_ common.Address, cap_ *big.Int, expiry_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.GrantStakingAllowance(&_DelegateStakingFunding.TransactOpts, hot_, cap_, expiry_)
}

// PurgeInactiveFunders is a paid mutator transaction binding the contract method 0x2765092e.
//
// Solidity: function purgeInactiveFunders(address hot_, uint256 maxEntries_) returns(uint256)
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) PurgeInactiveFunders(opts *bind.TransactOpts, hot_ common.Address, maxEntries_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "purgeInactiveFunders", hot_, maxEntries_)
}

// PurgeInactiveFunders is a paid mutator transaction binding the contract method 0x2765092e.
//
// Solidity: function purgeInactiveFunders(address hot_, uint256 maxEntries_) returns(uint256)
func (_DelegateStakingFunding *DelegateStakingFundingSession) PurgeInactiveFunders(hot_ common.Address, maxEntries_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.PurgeInactiveFunders(&_DelegateStakingFunding.TransactOpts, hot_, maxEntries_)
}

// PurgeInactiveFunders is a paid mutator transaction binding the contract method 0x2765092e.
//
// Solidity: function purgeInactiveFunders(address hot_, uint256 maxEntries_) returns(uint256)
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) PurgeInactiveFunders(hot_ common.Address, maxEntries_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.PurgeInactiveFunders(&_DelegateStakingFunding.TransactOpts, hot_, maxEntries_)
}

// ReleasePoolLocks is a paid mutator transaction binding the contract method 0x6e5461bc.
//
// Solidity: function releasePoolLocks(address hot_, uint256 maxBuckets_) returns(uint256 released_)
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) ReleasePoolLocks(opts *bind.TransactOpts, hot_ common.Address, maxBuckets_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "releasePoolLocks", hot_, maxBuckets_)
}

// ReleasePoolLocks is a paid mutator transaction binding the contract method 0x6e5461bc.
//
// Solidity: function releasePoolLocks(address hot_, uint256 maxBuckets_) returns(uint256 released_)
func (_DelegateStakingFunding *DelegateStakingFundingSession) ReleasePoolLocks(hot_ common.Address, maxBuckets_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.ReleasePoolLocks(&_DelegateStakingFunding.TransactOpts, hot_, maxBuckets_)
}

// ReleasePoolLocks is a paid mutator transaction binding the contract method 0x6e5461bc.
//
// Solidity: function releasePoolLocks(address hot_, uint256 maxBuckets_) returns(uint256 released_)
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) ReleasePoolLocks(hot_ common.Address, maxBuckets_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.ReleasePoolLocks(&_DelegateStakingFunding.TransactOpts, hot_, maxBuckets_)
}

// RequestStakingWithdrawal is a paid mutator transaction binding the contract method 0xd94ad88b.
//
// Solidity: function requestStakingWithdrawal(address hot_, address funder_, uint256 amount_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) RequestStakingWithdrawal(opts *bind.TransactOpts, hot_ common.Address, funder_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "requestStakingWithdrawal", hot_, funder_, amount_)
}

// RequestStakingWithdrawal is a paid mutator transaction binding the contract method 0xd94ad88b.
//
// Solidity: function requestStakingWithdrawal(address hot_, address funder_, uint256 amount_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingSession) RequestStakingWithdrawal(hot_ common.Address, funder_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.RequestStakingWithdrawal(&_DelegateStakingFunding.TransactOpts, hot_, funder_, amount_)
}

// RequestStakingWithdrawal is a paid mutator transaction binding the contract method 0xd94ad88b.
//
// Solidity: function requestStakingWithdrawal(address hot_, address funder_, uint256 amount_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) RequestStakingWithdrawal(hot_ common.Address, funder_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.RequestStakingWithdrawal(&_DelegateStakingFunding.TransactOpts, hot_, funder_, amount_)
}

// RevokePoolFunderApproval is a paid mutator transaction binding the contract method 0x28a48031.
//
// Solidity: function revokePoolFunderApproval(address hot_, address funder_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) RevokePoolFunderApproval(opts *bind.TransactOpts, hot_ common.Address, funder_ common.Address) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "revokePoolFunderApproval", hot_, funder_)
}

// RevokePoolFunderApproval is a paid mutator transaction binding the contract method 0x28a48031.
//
// Solidity: function revokePoolFunderApproval(address hot_, address funder_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingSession) RevokePoolFunderApproval(hot_ common.Address, funder_ common.Address) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.RevokePoolFunderApproval(&_DelegateStakingFunding.TransactOpts, hot_, funder_)
}

// RevokePoolFunderApproval is a paid mutator transaction binding the contract method 0x28a48031.
//
// Solidity: function revokePoolFunderApproval(address hot_, address funder_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) RevokePoolFunderApproval(hot_ common.Address, funder_ common.Address) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.RevokePoolFunderApproval(&_DelegateStakingFunding.TransactOpts, hot_, funder_)
}

// RevokeStakingAllowance is a paid mutator transaction binding the contract method 0x6a81e054.
//
// Solidity: function revokeStakingAllowance(address hot_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) RevokeStakingAllowance(opts *bind.TransactOpts, hot_ common.Address) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "revokeStakingAllowance", hot_)
}

// RevokeStakingAllowance is a paid mutator transaction binding the contract method 0x6a81e054.
//
// Solidity: function revokeStakingAllowance(address hot_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingSession) RevokeStakingAllowance(hot_ common.Address) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.RevokeStakingAllowance(&_DelegateStakingFunding.TransactOpts, hot_)
}

// RevokeStakingAllowance is a paid mutator transaction binding the contract method 0x6a81e054.
//
// Solidity: function revokeStakingAllowance(address hot_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) RevokeStakingAllowance(hot_ common.Address) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.RevokeStakingAllowance(&_DelegateStakingFunding.TransactOpts, hot_)
}

// ServePoolWithdrawals is a paid mutator transaction binding the contract method 0x19c2b604.
//
// Solidity: function servePoolWithdrawals(address hot_, uint256 maxEntries_) returns(uint256)
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) ServePoolWithdrawals(opts *bind.TransactOpts, hot_ common.Address, maxEntries_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "servePoolWithdrawals", hot_, maxEntries_)
}

// ServePoolWithdrawals is a paid mutator transaction binding the contract method 0x19c2b604.
//
// Solidity: function servePoolWithdrawals(address hot_, uint256 maxEntries_) returns(uint256)
func (_DelegateStakingFunding *DelegateStakingFundingSession) ServePoolWithdrawals(hot_ common.Address, maxEntries_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.ServePoolWithdrawals(&_DelegateStakingFunding.TransactOpts, hot_, maxEntries_)
}

// ServePoolWithdrawals is a paid mutator transaction binding the contract method 0x19c2b604.
//
// Solidity: function servePoolWithdrawals(address hot_, uint256 maxEntries_) returns(uint256)
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) ServePoolWithdrawals(hot_ common.Address, maxEntries_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.ServePoolWithdrawals(&_DelegateStakingFunding.TransactOpts, hot_, maxEntries_)
}

// SetMaxActiveFunders is a paid mutator transaction binding the contract method 0x422cd3e0.
//
// Solidity: function setMaxActiveFunders(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) SetMaxActiveFunders(opts *bind.TransactOpts, value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "setMaxActiveFunders", value_)
}

// SetMaxActiveFunders is a paid mutator transaction binding the contract method 0x422cd3e0.
//
// Solidity: function setMaxActiveFunders(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingSession) SetMaxActiveFunders(value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.SetMaxActiveFunders(&_DelegateStakingFunding.TransactOpts, value_)
}

// SetMaxActiveFunders is a paid mutator transaction binding the contract method 0x422cd3e0.
//
// Solidity: function setMaxActiveFunders(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) SetMaxActiveFunders(value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.SetMaxActiveFunders(&_DelegateStakingFunding.TransactOpts, value_)
}

// SetMaxAutoReleaseDays is a paid mutator transaction binding the contract method 0x661619fb.
//
// Solidity: function setMaxAutoReleaseDays(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) SetMaxAutoReleaseDays(opts *bind.TransactOpts, value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "setMaxAutoReleaseDays", value_)
}

// SetMaxAutoReleaseDays is a paid mutator transaction binding the contract method 0x661619fb.
//
// Solidity: function setMaxAutoReleaseDays(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingSession) SetMaxAutoReleaseDays(value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.SetMaxAutoReleaseDays(&_DelegateStakingFunding.TransactOpts, value_)
}

// SetMaxAutoReleaseDays is a paid mutator transaction binding the contract method 0x661619fb.
//
// Solidity: function setMaxAutoReleaseDays(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) SetMaxAutoReleaseDays(value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.SetMaxAutoReleaseDays(&_DelegateStakingFunding.TransactOpts, value_)
}

// SetMaxAutoService is a paid mutator transaction binding the contract method 0x5f4ca6c6.
//
// Solidity: function setMaxAutoService(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) SetMaxAutoService(opts *bind.TransactOpts, value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "setMaxAutoService", value_)
}

// SetMaxAutoService is a paid mutator transaction binding the contract method 0x5f4ca6c6.
//
// Solidity: function setMaxAutoService(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingSession) SetMaxAutoService(value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.SetMaxAutoService(&_DelegateStakingFunding.TransactOpts, value_)
}

// SetMaxAutoService is a paid mutator transaction binding the contract method 0x5f4ca6c6.
//
// Solidity: function setMaxAutoService(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) SetMaxAutoService(value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.SetMaxAutoService(&_DelegateStakingFunding.TransactOpts, value_)
}

// SetMinPrincipal is a paid mutator transaction binding the contract method 0x110cc7a7.
//
// Solidity: function setMinPrincipal(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) SetMinPrincipal(opts *bind.TransactOpts, value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "setMinPrincipal", value_)
}

// SetMinPrincipal is a paid mutator transaction binding the contract method 0x110cc7a7.
//
// Solidity: function setMinPrincipal(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingSession) SetMinPrincipal(value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.SetMinPrincipal(&_DelegateStakingFunding.TransactOpts, value_)
}

// SetMinPrincipal is a paid mutator transaction binding the contract method 0x110cc7a7.
//
// Solidity: function setMinPrincipal(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) SetMinPrincipal(value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.SetMinPrincipal(&_DelegateStakingFunding.TransactOpts, value_)
}

// SetSettlementGrace is a paid mutator transaction binding the contract method 0x87779611.
//
// Solidity: function setSettlementGrace(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactor) SetSettlementGrace(opts *bind.TransactOpts, value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.contract.Transact(opts, "setSettlementGrace", value_)
}

// SetSettlementGrace is a paid mutator transaction binding the contract method 0x87779611.
//
// Solidity: function setSettlementGrace(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingSession) SetSettlementGrace(value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.SetSettlementGrace(&_DelegateStakingFunding.TransactOpts, value_)
}

// SetSettlementGrace is a paid mutator transaction binding the contract method 0x87779611.
//
// Solidity: function setSettlementGrace(uint256 value_) returns()
func (_DelegateStakingFunding *DelegateStakingFundingTransactorSession) SetSettlementGrace(value_ *big.Int) (*types.Transaction, error) {
	return _DelegateStakingFunding.Contract.SetSettlementGrace(&_DelegateStakingFunding.TransactOpts, value_)
}

// DelegateStakingFundingDelegateStakingAllowanceFundedIterator is returned from FilterDelegateStakingAllowanceFunded and is used to iterate over the raw logs and unpacked data for DelegateStakingAllowanceFunded events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingAllowanceFundedIterator struct {
	Event *DelegateStakingFundingDelegateStakingAllowanceFunded // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingAllowanceFundedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingAllowanceFunded)
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
		it.Event = new(DelegateStakingFundingDelegateStakingAllowanceFunded)
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
func (it *DelegateStakingFundingDelegateStakingAllowanceFundedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingAllowanceFundedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingAllowanceFunded represents a DelegateStakingAllowanceFunded event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingAllowanceFunded struct {
	Hot    common.Address
	Funder common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingAllowanceFunded is a free log retrieval operation binding the contract event 0xab72e208ad5755afbd9bcdfe0b4304d209b428685d5ee350433f417b117f38f0.
//
// Solidity: event DelegateStakingAllowanceFunded(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingAllowanceFunded(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingFundingDelegateStakingAllowanceFundedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingAllowanceFunded", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingAllowanceFundedIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingAllowanceFunded", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingAllowanceFunded is a free log subscription operation binding the contract event 0xab72e208ad5755afbd9bcdfe0b4304d209b428685d5ee350433f417b117f38f0.
//
// Solidity: event DelegateStakingAllowanceFunded(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingAllowanceFunded(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingAllowanceFunded, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingAllowanceFunded", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingAllowanceFunded)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingAllowanceFunded", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingAllowanceFunded(log types.Log) (*DelegateStakingFundingDelegateStakingAllowanceFunded, error) {
	event := new(DelegateStakingFundingDelegateStakingAllowanceFunded)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingAllowanceFunded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingAllowanceGrantedIterator is returned from FilterDelegateStakingAllowanceGranted and is used to iterate over the raw logs and unpacked data for DelegateStakingAllowanceGranted events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingAllowanceGrantedIterator struct {
	Event *DelegateStakingFundingDelegateStakingAllowanceGranted // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingAllowanceGrantedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingAllowanceGranted)
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
		it.Event = new(DelegateStakingFundingDelegateStakingAllowanceGranted)
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
func (it *DelegateStakingFundingDelegateStakingAllowanceGrantedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingAllowanceGrantedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingAllowanceGranted represents a DelegateStakingAllowanceGranted event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingAllowanceGranted struct {
	Hot                  common.Address
	Funder               common.Address
	CumulativeFundingCap *big.Int
	Expiry               *big.Int
	Raw                  types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingAllowanceGranted is a free log retrieval operation binding the contract event 0x743ecb36e3b0fa6c16a5e8cd3ae70e657b294fd68ffbdfdc73d3a89d04addf46.
//
// Solidity: event DelegateStakingAllowanceGranted(address indexed hot, address indexed funder, uint256 cumulativeFundingCap, uint128 expiry)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingAllowanceGranted(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingFundingDelegateStakingAllowanceGrantedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingAllowanceGranted", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingAllowanceGrantedIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingAllowanceGranted", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingAllowanceGranted is a free log subscription operation binding the contract event 0x743ecb36e3b0fa6c16a5e8cd3ae70e657b294fd68ffbdfdc73d3a89d04addf46.
//
// Solidity: event DelegateStakingAllowanceGranted(address indexed hot, address indexed funder, uint256 cumulativeFundingCap, uint128 expiry)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingAllowanceGranted(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingAllowanceGranted, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingAllowanceGranted", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingAllowanceGranted)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingAllowanceGranted", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingAllowanceGranted(log types.Log) (*DelegateStakingFundingDelegateStakingAllowanceGranted, error) {
	event := new(DelegateStakingFundingDelegateStakingAllowanceGranted)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingAllowanceGranted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingAllowanceRevokedIterator is returned from FilterDelegateStakingAllowanceRevoked and is used to iterate over the raw logs and unpacked data for DelegateStakingAllowanceRevoked events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingAllowanceRevokedIterator struct {
	Event *DelegateStakingFundingDelegateStakingAllowanceRevoked // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingAllowanceRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingAllowanceRevoked)
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
		it.Event = new(DelegateStakingFundingDelegateStakingAllowanceRevoked)
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
func (it *DelegateStakingFundingDelegateStakingAllowanceRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingAllowanceRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingAllowanceRevoked represents a DelegateStakingAllowanceRevoked event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingAllowanceRevoked struct {
	Hot    common.Address
	Funder common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingAllowanceRevoked is a free log retrieval operation binding the contract event 0x641afbae43e9aa28104693da9d716e2295e9efaf231c0bfec7dd6f949d3bd726.
//
// Solidity: event DelegateStakingAllowanceRevoked(address indexed hot, address indexed funder)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingAllowanceRevoked(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingFundingDelegateStakingAllowanceRevokedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingAllowanceRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingAllowanceRevokedIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingAllowanceRevoked", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingAllowanceRevoked is a free log subscription operation binding the contract event 0x641afbae43e9aa28104693da9d716e2295e9efaf231c0bfec7dd6f949d3bd726.
//
// Solidity: event DelegateStakingAllowanceRevoked(address indexed hot, address indexed funder)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingAllowanceRevoked(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingAllowanceRevoked, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingAllowanceRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingAllowanceRevoked)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingAllowanceRevoked", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingAllowanceRevoked(log types.Log) (*DelegateStakingFundingDelegateStakingAllowanceRevoked, error) {
	event := new(DelegateStakingFundingDelegateStakingAllowanceRevoked)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingAllowanceRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingMaxActiveFundersUpdatedIterator is returned from FilterDelegateStakingMaxActiveFundersUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMaxActiveFundersUpdated events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingMaxActiveFundersUpdatedIterator struct {
	Event *DelegateStakingFundingDelegateStakingMaxActiveFundersUpdated // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingMaxActiveFundersUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingMaxActiveFundersUpdated)
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
		it.Event = new(DelegateStakingFundingDelegateStakingMaxActiveFundersUpdated)
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
func (it *DelegateStakingFundingDelegateStakingMaxActiveFundersUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingMaxActiveFundersUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingMaxActiveFundersUpdated represents a DelegateStakingMaxActiveFundersUpdated event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingMaxActiveFundersUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMaxActiveFundersUpdated is a free log retrieval operation binding the contract event 0x6018427c349577edf6a6c0b891cf09b86b47cc1834917516e6da404941ef43f5.
//
// Solidity: event DelegateStakingMaxActiveFundersUpdated(uint256 value)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingMaxActiveFundersUpdated(opts *bind.FilterOpts) (*DelegateStakingFundingDelegateStakingMaxActiveFundersUpdatedIterator, error) {

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingMaxActiveFundersUpdated")
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingMaxActiveFundersUpdatedIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingMaxActiveFundersUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMaxActiveFundersUpdated is a free log subscription operation binding the contract event 0x6018427c349577edf6a6c0b891cf09b86b47cc1834917516e6da404941ef43f5.
//
// Solidity: event DelegateStakingMaxActiveFundersUpdated(uint256 value)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingMaxActiveFundersUpdated(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingMaxActiveFundersUpdated) (event.Subscription, error) {

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingMaxActiveFundersUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingMaxActiveFundersUpdated)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingMaxActiveFundersUpdated", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingMaxActiveFundersUpdated(log types.Log) (*DelegateStakingFundingDelegateStakingMaxActiveFundersUpdated, error) {
	event := new(DelegateStakingFundingDelegateStakingMaxActiveFundersUpdated)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingMaxActiveFundersUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdatedIterator is returned from FilterDelegateStakingMaxAutoReleaseDaysUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMaxAutoReleaseDaysUpdated events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdatedIterator struct {
	Event *DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdated // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdated)
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
		it.Event = new(DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdated)
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
func (it *DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdated represents a DelegateStakingMaxAutoReleaseDaysUpdated event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMaxAutoReleaseDaysUpdated is a free log retrieval operation binding the contract event 0xc9ca6843c5a726f8df173b27882fb647575843c62e31760e9b7f90c94518aa2a.
//
// Solidity: event DelegateStakingMaxAutoReleaseDaysUpdated(uint256 value)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingMaxAutoReleaseDaysUpdated(opts *bind.FilterOpts) (*DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdatedIterator, error) {

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingMaxAutoReleaseDaysUpdated")
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdatedIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingMaxAutoReleaseDaysUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMaxAutoReleaseDaysUpdated is a free log subscription operation binding the contract event 0xc9ca6843c5a726f8df173b27882fb647575843c62e31760e9b7f90c94518aa2a.
//
// Solidity: event DelegateStakingMaxAutoReleaseDaysUpdated(uint256 value)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingMaxAutoReleaseDaysUpdated(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdated) (event.Subscription, error) {

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingMaxAutoReleaseDaysUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdated)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingMaxAutoReleaseDaysUpdated", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingMaxAutoReleaseDaysUpdated(log types.Log) (*DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdated, error) {
	event := new(DelegateStakingFundingDelegateStakingMaxAutoReleaseDaysUpdated)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingMaxAutoReleaseDaysUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingMaxAutoServiceUpdatedIterator is returned from FilterDelegateStakingMaxAutoServiceUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMaxAutoServiceUpdated events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingMaxAutoServiceUpdatedIterator struct {
	Event *DelegateStakingFundingDelegateStakingMaxAutoServiceUpdated // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingMaxAutoServiceUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingMaxAutoServiceUpdated)
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
		it.Event = new(DelegateStakingFundingDelegateStakingMaxAutoServiceUpdated)
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
func (it *DelegateStakingFundingDelegateStakingMaxAutoServiceUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingMaxAutoServiceUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingMaxAutoServiceUpdated represents a DelegateStakingMaxAutoServiceUpdated event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingMaxAutoServiceUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMaxAutoServiceUpdated is a free log retrieval operation binding the contract event 0x4f927d3e460674eab7e38dac3592136fe961de0c14533b919652a814a7510adc.
//
// Solidity: event DelegateStakingMaxAutoServiceUpdated(uint256 value)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingMaxAutoServiceUpdated(opts *bind.FilterOpts) (*DelegateStakingFundingDelegateStakingMaxAutoServiceUpdatedIterator, error) {

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingMaxAutoServiceUpdated")
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingMaxAutoServiceUpdatedIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingMaxAutoServiceUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMaxAutoServiceUpdated is a free log subscription operation binding the contract event 0x4f927d3e460674eab7e38dac3592136fe961de0c14533b919652a814a7510adc.
//
// Solidity: event DelegateStakingMaxAutoServiceUpdated(uint256 value)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingMaxAutoServiceUpdated(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingMaxAutoServiceUpdated) (event.Subscription, error) {

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingMaxAutoServiceUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingMaxAutoServiceUpdated)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingMaxAutoServiceUpdated", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingMaxAutoServiceUpdated(log types.Log) (*DelegateStakingFundingDelegateStakingMaxAutoServiceUpdated, error) {
	event := new(DelegateStakingFundingDelegateStakingMaxAutoServiceUpdated)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingMaxAutoServiceUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingMinPrincipalUpdatedIterator is returned from FilterDelegateStakingMinPrincipalUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingMinPrincipalUpdated events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingMinPrincipalUpdatedIterator struct {
	Event *DelegateStakingFundingDelegateStakingMinPrincipalUpdated // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingMinPrincipalUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingMinPrincipalUpdated)
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
		it.Event = new(DelegateStakingFundingDelegateStakingMinPrincipalUpdated)
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
func (it *DelegateStakingFundingDelegateStakingMinPrincipalUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingMinPrincipalUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingMinPrincipalUpdated represents a DelegateStakingMinPrincipalUpdated event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingMinPrincipalUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingMinPrincipalUpdated is a free log retrieval operation binding the contract event 0x06ae87e705be2a5e7c6ff17a5d3d836834554f7d0aadaef3abb4343a202590a2.
//
// Solidity: event DelegateStakingMinPrincipalUpdated(uint256 value)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingMinPrincipalUpdated(opts *bind.FilterOpts) (*DelegateStakingFundingDelegateStakingMinPrincipalUpdatedIterator, error) {

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingMinPrincipalUpdated")
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingMinPrincipalUpdatedIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingMinPrincipalUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingMinPrincipalUpdated is a free log subscription operation binding the contract event 0x06ae87e705be2a5e7c6ff17a5d3d836834554f7d0aadaef3abb4343a202590a2.
//
// Solidity: event DelegateStakingMinPrincipalUpdated(uint256 value)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingMinPrincipalUpdated(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingMinPrincipalUpdated) (event.Subscription, error) {

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingMinPrincipalUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingMinPrincipalUpdated)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingMinPrincipalUpdated", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingMinPrincipalUpdated(log types.Log) (*DelegateStakingFundingDelegateStakingMinPrincipalUpdated, error) {
	event := new(DelegateStakingFundingDelegateStakingMinPrincipalUpdated)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingMinPrincipalUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingPoolDrawnIterator is returned from FilterDelegateStakingPoolDrawn and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolDrawn events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingPoolDrawnIterator struct {
	Event *DelegateStakingFundingDelegateStakingPoolDrawn // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingPoolDrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingPoolDrawn)
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
		it.Event = new(DelegateStakingFundingDelegateStakingPoolDrawn)
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
func (it *DelegateStakingFundingDelegateStakingPoolDrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingPoolDrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingPoolDrawn represents a DelegateStakingPoolDrawn event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingPoolDrawn struct {
	SessionId [32]byte
	Hot       common.Address
	Funder    common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolDrawn is a free log retrieval operation binding the contract event 0xd178a2f3e5da2cc00732e1b88d2670a1eaa4f88a0b86e8790237a677f1fa0985.
//
// Solidity: event DelegateStakingPoolDrawn(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingPoolDrawn(opts *bind.FilterOpts, sessionId [][32]byte, hot []common.Address, funder []common.Address) (*DelegateStakingFundingDelegateStakingPoolDrawnIterator, error) {

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

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingPoolDrawn", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingPoolDrawnIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingPoolDrawn", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolDrawn is a free log subscription operation binding the contract event 0xd178a2f3e5da2cc00732e1b88d2670a1eaa4f88a0b86e8790237a677f1fa0985.
//
// Solidity: event DelegateStakingPoolDrawn(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingPoolDrawn(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingPoolDrawn, sessionId [][32]byte, hot []common.Address, funder []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingPoolDrawn", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingPoolDrawn)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingPoolDrawn", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingPoolDrawn(log types.Log) (*DelegateStakingFundingDelegateStakingPoolDrawn, error) {
	event := new(DelegateStakingFundingDelegateStakingPoolDrawn)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingPoolDrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingPoolFunderApprovalRevokedIterator is returned from FilterDelegateStakingPoolFunderApprovalRevoked and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolFunderApprovalRevoked events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingPoolFunderApprovalRevokedIterator struct {
	Event *DelegateStakingFundingDelegateStakingPoolFunderApprovalRevoked // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingPoolFunderApprovalRevokedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingPoolFunderApprovalRevoked)
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
		it.Event = new(DelegateStakingFundingDelegateStakingPoolFunderApprovalRevoked)
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
func (it *DelegateStakingFundingDelegateStakingPoolFunderApprovalRevokedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingPoolFunderApprovalRevokedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingPoolFunderApprovalRevoked represents a DelegateStakingPoolFunderApprovalRevoked event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingPoolFunderApprovalRevoked struct {
	Hot    common.Address
	Funder common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolFunderApprovalRevoked is a free log retrieval operation binding the contract event 0x5ebb99e7e14d99bff22536b105e92a32d4fbb701f93b0d5b008cc0a0b876aeb2.
//
// Solidity: event DelegateStakingPoolFunderApprovalRevoked(address indexed hot, address indexed funder)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingPoolFunderApprovalRevoked(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingFundingDelegateStakingPoolFunderApprovalRevokedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingPoolFunderApprovalRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingPoolFunderApprovalRevokedIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingPoolFunderApprovalRevoked", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolFunderApprovalRevoked is a free log subscription operation binding the contract event 0x5ebb99e7e14d99bff22536b105e92a32d4fbb701f93b0d5b008cc0a0b876aeb2.
//
// Solidity: event DelegateStakingPoolFunderApprovalRevoked(address indexed hot, address indexed funder)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingPoolFunderApprovalRevoked(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingPoolFunderApprovalRevoked, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingPoolFunderApprovalRevoked", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingPoolFunderApprovalRevoked)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingPoolFunderApprovalRevoked", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingPoolFunderApprovalRevoked(log types.Log) (*DelegateStakingFundingDelegateStakingPoolFunderApprovalRevoked, error) {
	event := new(DelegateStakingFundingDelegateStakingPoolFunderApprovalRevoked)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingPoolFunderApprovalRevoked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingPoolFunderApprovedIterator is returned from FilterDelegateStakingPoolFunderApproved and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolFunderApproved events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingPoolFunderApprovedIterator struct {
	Event *DelegateStakingFundingDelegateStakingPoolFunderApproved // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingPoolFunderApprovedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingPoolFunderApproved)
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
		it.Event = new(DelegateStakingFundingDelegateStakingPoolFunderApproved)
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
func (it *DelegateStakingFundingDelegateStakingPoolFunderApprovedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingPoolFunderApprovedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingPoolFunderApproved represents a DelegateStakingPoolFunderApproved event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingPoolFunderApproved struct {
	Hot    common.Address
	Funder common.Address
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolFunderApproved is a free log retrieval operation binding the contract event 0x2a6f85b141a6737d1530bd4b8c916eee97ba403dae709170266f72b6ec58f739.
//
// Solidity: event DelegateStakingPoolFunderApproved(address indexed hot, address indexed funder)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingPoolFunderApproved(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingFundingDelegateStakingPoolFunderApprovedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingPoolFunderApproved", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingPoolFunderApprovedIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingPoolFunderApproved", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolFunderApproved is a free log subscription operation binding the contract event 0x2a6f85b141a6737d1530bd4b8c916eee97ba403dae709170266f72b6ec58f739.
//
// Solidity: event DelegateStakingPoolFunderApproved(address indexed hot, address indexed funder)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingPoolFunderApproved(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingPoolFunderApproved, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingPoolFunderApproved", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingPoolFunderApproved)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingPoolFunderApproved", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingPoolFunderApproved(log types.Log) (*DelegateStakingFundingDelegateStakingPoolFunderApproved, error) {
	event := new(DelegateStakingFundingDelegateStakingPoolFunderApproved)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingPoolFunderApproved", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingPoolLockCreatedIterator is returned from FilterDelegateStakingPoolLockCreated and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolLockCreated events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingPoolLockCreatedIterator struct {
	Event *DelegateStakingFundingDelegateStakingPoolLockCreated // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingPoolLockCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingPoolLockCreated)
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
		it.Event = new(DelegateStakingFundingDelegateStakingPoolLockCreated)
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
func (it *DelegateStakingFundingDelegateStakingPoolLockCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingPoolLockCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingPoolLockCreated represents a DelegateStakingPoolLockCreated event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingPoolLockCreated struct {
	Hot       common.Address
	ReleaseAt *big.Int
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolLockCreated is a free log retrieval operation binding the contract event 0x87eb288bde7ca5d84aa5e85a8b81f8d7c31d8e19049823dab5552d8bb461dc0e.
//
// Solidity: event DelegateStakingPoolLockCreated(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingPoolLockCreated(opts *bind.FilterOpts, hot []common.Address, releaseAt []*big.Int) (*DelegateStakingFundingDelegateStakingPoolLockCreatedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingPoolLockCreated", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingPoolLockCreatedIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingPoolLockCreated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolLockCreated is a free log subscription operation binding the contract event 0x87eb288bde7ca5d84aa5e85a8b81f8d7c31d8e19049823dab5552d8bb461dc0e.
//
// Solidity: event DelegateStakingPoolLockCreated(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingPoolLockCreated(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingPoolLockCreated, hot []common.Address, releaseAt []*big.Int) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingPoolLockCreated", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingPoolLockCreated)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingPoolLockCreated", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingPoolLockCreated(log types.Log) (*DelegateStakingFundingDelegateStakingPoolLockCreated, error) {
	event := new(DelegateStakingFundingDelegateStakingPoolLockCreated)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingPoolLockCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingPoolLockReleasedIterator is returned from FilterDelegateStakingPoolLockReleased and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolLockReleased events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingPoolLockReleasedIterator struct {
	Event *DelegateStakingFundingDelegateStakingPoolLockReleased // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingPoolLockReleasedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingPoolLockReleased)
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
		it.Event = new(DelegateStakingFundingDelegateStakingPoolLockReleased)
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
func (it *DelegateStakingFundingDelegateStakingPoolLockReleasedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingPoolLockReleasedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingPoolLockReleased represents a DelegateStakingPoolLockReleased event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingPoolLockReleased struct {
	Hot       common.Address
	ReleaseAt *big.Int
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolLockReleased is a free log retrieval operation binding the contract event 0xd04817baed5e7c2b01f59b2218c05170e733b0e2a3864693721a2fc89e5f7a4f.
//
// Solidity: event DelegateStakingPoolLockReleased(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingPoolLockReleased(opts *bind.FilterOpts, hot []common.Address, releaseAt []*big.Int) (*DelegateStakingFundingDelegateStakingPoolLockReleasedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingPoolLockReleased", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingPoolLockReleasedIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingPoolLockReleased", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolLockReleased is a free log subscription operation binding the contract event 0xd04817baed5e7c2b01f59b2218c05170e733b0e2a3864693721a2fc89e5f7a4f.
//
// Solidity: event DelegateStakingPoolLockReleased(address indexed hot, uint128 indexed releaseAt, uint256 amount)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingPoolLockReleased(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingPoolLockReleased, hot []common.Address, releaseAt []*big.Int) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var releaseAtRule []interface{}
	for _, releaseAtItem := range releaseAt {
		releaseAtRule = append(releaseAtRule, releaseAtItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingPoolLockReleased", hotRule, releaseAtRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingPoolLockReleased)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingPoolLockReleased", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingPoolLockReleased(log types.Log) (*DelegateStakingFundingDelegateStakingPoolLockReleased, error) {
	event := new(DelegateStakingFundingDelegateStakingPoolLockReleased)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingPoolLockReleased", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingPoolReturnedIterator is returned from FilterDelegateStakingPoolReturned and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolReturned events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingPoolReturnedIterator struct {
	Event *DelegateStakingFundingDelegateStakingPoolReturned // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingPoolReturnedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingPoolReturned)
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
		it.Event = new(DelegateStakingFundingDelegateStakingPoolReturned)
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
func (it *DelegateStakingFundingDelegateStakingPoolReturnedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingPoolReturnedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingPoolReturned represents a DelegateStakingPoolReturned event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingPoolReturned struct {
	SessionId [32]byte
	Hot       common.Address
	Funder    common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolReturned is a free log retrieval operation binding the contract event 0x66524d7951752d8716b27064a7d50819e4db87609ea9eca1f6b1d42f359a061d.
//
// Solidity: event DelegateStakingPoolReturned(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingPoolReturned(opts *bind.FilterOpts, sessionId [][32]byte, hot []common.Address, funder []common.Address) (*DelegateStakingFundingDelegateStakingPoolReturnedIterator, error) {

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

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingPoolReturned", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingPoolReturnedIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingPoolReturned", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolReturned is a free log subscription operation binding the contract event 0x66524d7951752d8716b27064a7d50819e4db87609ea9eca1f6b1d42f359a061d.
//
// Solidity: event DelegateStakingPoolReturned(bytes32 indexed sessionId, address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingPoolReturned(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingPoolReturned, sessionId [][32]byte, hot []common.Address, funder []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingPoolReturned", sessionIdRule, hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingPoolReturned)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingPoolReturned", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingPoolReturned(log types.Log) (*DelegateStakingFundingDelegateStakingPoolReturned, error) {
	event := new(DelegateStakingFundingDelegateStakingPoolReturned)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingPoolReturned", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingPoolSessionSettledIterator is returned from FilterDelegateStakingPoolSessionSettled and is used to iterate over the raw logs and unpacked data for DelegateStakingPoolSessionSettled events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingPoolSessionSettledIterator struct {
	Event *DelegateStakingFundingDelegateStakingPoolSessionSettled // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingPoolSessionSettledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingPoolSessionSettled)
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
		it.Event = new(DelegateStakingFundingDelegateStakingPoolSessionSettled)
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
func (it *DelegateStakingFundingDelegateStakingPoolSessionSettledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingPoolSessionSettledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingPoolSessionSettled represents a DelegateStakingPoolSessionSettled event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingPoolSessionSettled struct {
	SessionId [32]byte
	Hot       common.Address
	Provider  common.Address
	Caller    common.Address
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingPoolSessionSettled is a free log retrieval operation binding the contract event 0x518bf7ca4dc0c0c2144467883693da5fd9b0532de6539122309ecff49a330215.
//
// Solidity: event DelegateStakingPoolSessionSettled(bytes32 indexed sessionId, address indexed hot, address indexed provider, address caller)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingPoolSessionSettled(opts *bind.FilterOpts, sessionId [][32]byte, hot []common.Address, provider []common.Address) (*DelegateStakingFundingDelegateStakingPoolSessionSettledIterator, error) {

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

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingPoolSessionSettled", sessionIdRule, hotRule, providerRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingPoolSessionSettledIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingPoolSessionSettled", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingPoolSessionSettled is a free log subscription operation binding the contract event 0x518bf7ca4dc0c0c2144467883693da5fd9b0532de6539122309ecff49a330215.
//
// Solidity: event DelegateStakingPoolSessionSettled(bytes32 indexed sessionId, address indexed hot, address indexed provider, address caller)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingPoolSessionSettled(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingPoolSessionSettled, sessionId [][32]byte, hot []common.Address, provider []common.Address) (event.Subscription, error) {

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

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingPoolSessionSettled", sessionIdRule, hotRule, providerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingPoolSessionSettled)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingPoolSessionSettled", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingPoolSessionSettled(log types.Log) (*DelegateStakingFundingDelegateStakingPoolSessionSettled, error) {
	event := new(DelegateStakingFundingDelegateStakingPoolSessionSettled)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingPoolSessionSettled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingSettlementGraceUpdatedIterator is returned from FilterDelegateStakingSettlementGraceUpdated and is used to iterate over the raw logs and unpacked data for DelegateStakingSettlementGraceUpdated events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingSettlementGraceUpdatedIterator struct {
	Event *DelegateStakingFundingDelegateStakingSettlementGraceUpdated // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingSettlementGraceUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingSettlementGraceUpdated)
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
		it.Event = new(DelegateStakingFundingDelegateStakingSettlementGraceUpdated)
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
func (it *DelegateStakingFundingDelegateStakingSettlementGraceUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingSettlementGraceUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingSettlementGraceUpdated represents a DelegateStakingSettlementGraceUpdated event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingSettlementGraceUpdated struct {
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingSettlementGraceUpdated is a free log retrieval operation binding the contract event 0x23d4c1527dfe24f1a043524d5c7c2b730f3fb2841478487080916494bc8487f2.
//
// Solidity: event DelegateStakingSettlementGraceUpdated(uint256 value)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingSettlementGraceUpdated(opts *bind.FilterOpts) (*DelegateStakingFundingDelegateStakingSettlementGraceUpdatedIterator, error) {

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingSettlementGraceUpdated")
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingSettlementGraceUpdatedIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingSettlementGraceUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingSettlementGraceUpdated is a free log subscription operation binding the contract event 0x23d4c1527dfe24f1a043524d5c7c2b730f3fb2841478487080916494bc8487f2.
//
// Solidity: event DelegateStakingSettlementGraceUpdated(uint256 value)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingSettlementGraceUpdated(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingSettlementGraceUpdated) (event.Subscription, error) {

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingSettlementGraceUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingSettlementGraceUpdated)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingSettlementGraceUpdated", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingSettlementGraceUpdated(log types.Log) (*DelegateStakingFundingDelegateStakingSettlementGraceUpdated, error) {
	event := new(DelegateStakingFundingDelegateStakingSettlementGraceUpdated)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingSettlementGraceUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingWithdrawalCancelledIterator is returned from FilterDelegateStakingWithdrawalCancelled and is used to iterate over the raw logs and unpacked data for DelegateStakingWithdrawalCancelled events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingWithdrawalCancelledIterator struct {
	Event *DelegateStakingFundingDelegateStakingWithdrawalCancelled // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingWithdrawalCancelledIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingWithdrawalCancelled)
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
		it.Event = new(DelegateStakingFundingDelegateStakingWithdrawalCancelled)
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
func (it *DelegateStakingFundingDelegateStakingWithdrawalCancelledIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingWithdrawalCancelledIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingWithdrawalCancelled represents a DelegateStakingWithdrawalCancelled event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingWithdrawalCancelled struct {
	Hot    common.Address
	Funder common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingWithdrawalCancelled is a free log retrieval operation binding the contract event 0x00041bc92ab137f1f37671a48ce9dc19f99eb0d99b0bf41beb1ffca2dd9cb557.
//
// Solidity: event DelegateStakingWithdrawalCancelled(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingWithdrawalCancelled(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingFundingDelegateStakingWithdrawalCancelledIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingWithdrawalCancelled", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingWithdrawalCancelledIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingWithdrawalCancelled", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingWithdrawalCancelled is a free log subscription operation binding the contract event 0x00041bc92ab137f1f37671a48ce9dc19f99eb0d99b0bf41beb1ffca2dd9cb557.
//
// Solidity: event DelegateStakingWithdrawalCancelled(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingWithdrawalCancelled(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingWithdrawalCancelled, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingWithdrawalCancelled", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingWithdrawalCancelled)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingWithdrawalCancelled", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingWithdrawalCancelled(log types.Log) (*DelegateStakingFundingDelegateStakingWithdrawalCancelled, error) {
	event := new(DelegateStakingFundingDelegateStakingWithdrawalCancelled)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingWithdrawalCancelled", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingWithdrawalPaidIterator is returned from FilterDelegateStakingWithdrawalPaid and is used to iterate over the raw logs and unpacked data for DelegateStakingWithdrawalPaid events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingWithdrawalPaidIterator struct {
	Event *DelegateStakingFundingDelegateStakingWithdrawalPaid // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingWithdrawalPaidIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingWithdrawalPaid)
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
		it.Event = new(DelegateStakingFundingDelegateStakingWithdrawalPaid)
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
func (it *DelegateStakingFundingDelegateStakingWithdrawalPaidIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingWithdrawalPaidIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingWithdrawalPaid represents a DelegateStakingWithdrawalPaid event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingWithdrawalPaid struct {
	Hot    common.Address
	Funder common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDelegateStakingWithdrawalPaid is a free log retrieval operation binding the contract event 0x7827e4dd700b7a2a556c261a5a293c5c936068522520173931017b0e9959d935.
//
// Solidity: event DelegateStakingWithdrawalPaid(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingWithdrawalPaid(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingFundingDelegateStakingWithdrawalPaidIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingWithdrawalPaid", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingWithdrawalPaidIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingWithdrawalPaid", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingWithdrawalPaid is a free log subscription operation binding the contract event 0x7827e4dd700b7a2a556c261a5a293c5c936068522520173931017b0e9959d935.
//
// Solidity: event DelegateStakingWithdrawalPaid(address indexed hot, address indexed funder, uint256 amount)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingWithdrawalPaid(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingWithdrawalPaid, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingWithdrawalPaid", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingWithdrawalPaid)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingWithdrawalPaid", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingWithdrawalPaid(log types.Log) (*DelegateStakingFundingDelegateStakingWithdrawalPaid, error) {
	event := new(DelegateStakingFundingDelegateStakingWithdrawalPaid)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingWithdrawalPaid", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingDelegateStakingWithdrawalRequestedIterator is returned from FilterDelegateStakingWithdrawalRequested and is used to iterate over the raw logs and unpacked data for DelegateStakingWithdrawalRequested events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingWithdrawalRequestedIterator struct {
	Event *DelegateStakingFundingDelegateStakingWithdrawalRequested // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingDelegateStakingWithdrawalRequestedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingDelegateStakingWithdrawalRequested)
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
		it.Event = new(DelegateStakingFundingDelegateStakingWithdrawalRequested)
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
func (it *DelegateStakingFundingDelegateStakingWithdrawalRequestedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingDelegateStakingWithdrawalRequestedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingDelegateStakingWithdrawalRequested represents a DelegateStakingWithdrawalRequested event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingDelegateStakingWithdrawalRequested struct {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterDelegateStakingWithdrawalRequested(opts *bind.FilterOpts, hot []common.Address, funder []common.Address) (*DelegateStakingFundingDelegateStakingWithdrawalRequestedIterator, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "DelegateStakingWithdrawalRequested", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingDelegateStakingWithdrawalRequestedIterator{contract: _DelegateStakingFunding.contract, event: "DelegateStakingWithdrawalRequested", logs: logs, sub: sub}, nil
}

// WatchDelegateStakingWithdrawalRequested is a free log subscription operation binding the contract event 0xab9f026078e090833bf0fd1211137db97a7a96f822c650c7997cdb7bfc859799.
//
// Solidity: event DelegateStakingWithdrawalRequested(address indexed hot, address indexed funder, uint256 amount, uint256 paid, uint256 pending)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchDelegateStakingWithdrawalRequested(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingDelegateStakingWithdrawalRequested, hot []common.Address, funder []common.Address) (event.Subscription, error) {

	var hotRule []interface{}
	for _, hotItem := range hot {
		hotRule = append(hotRule, hotItem)
	}
	var funderRule []interface{}
	for _, funderItem := range funder {
		funderRule = append(funderRule, funderItem)
	}

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "DelegateStakingWithdrawalRequested", hotRule, funderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingDelegateStakingWithdrawalRequested)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingWithdrawalRequested", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseDelegateStakingWithdrawalRequested(log types.Log) (*DelegateStakingFundingDelegateStakingWithdrawalRequested, error) {
	event := new(DelegateStakingFundingDelegateStakingWithdrawalRequested)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "DelegateStakingWithdrawalRequested", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegateStakingFundingInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the DelegateStakingFunding contract.
type DelegateStakingFundingInitializedIterator struct {
	Event *DelegateStakingFundingInitialized // Event containing the contract specifics and raw log

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
func (it *DelegateStakingFundingInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegateStakingFundingInitialized)
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
		it.Event = new(DelegateStakingFundingInitialized)
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
func (it *DelegateStakingFundingInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegateStakingFundingInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegateStakingFundingInitialized represents a Initialized event raised by the DelegateStakingFunding contract.
type DelegateStakingFundingInitialized struct {
	StorageSlot [32]byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) FilterInitialized(opts *bind.FilterOpts) (*DelegateStakingFundingInitializedIterator, error) {

	logs, sub, err := _DelegateStakingFunding.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &DelegateStakingFundingInitializedIterator{contract: _DelegateStakingFunding.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *DelegateStakingFundingInitialized) (event.Subscription, error) {

	logs, sub, err := _DelegateStakingFunding.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegateStakingFundingInitialized)
				if err := _DelegateStakingFunding.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_DelegateStakingFunding *DelegateStakingFundingFilterer) ParseInitialized(log types.Log) (*DelegateStakingFundingInitialized, error) {
	event := new(DelegateStakingFundingInitialized)
	if err := _DelegateStakingFunding.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
