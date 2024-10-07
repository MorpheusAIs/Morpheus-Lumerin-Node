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
	Id                      [32]byte
	User                    common.Address
	Provider                common.Address
	ModelId                 [32]byte
	BidId                   [32]byte
	Stake                   *big.Int
	PricePerSecond          *big.Int
	CloseoutReceipt         []byte
	CloseoutType            *big.Int
	ProviderWithdrawnAmount *big.Int
	OpenedAt                *big.Int
	EndsAt                  *big.Int
	ClosedAt                *big.Int
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
	ABI: "[{\"inputs\":[],\"name\":\"AmountToWithdrawIsZero\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ApprovedForAnotherUser\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"BidNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CannotDecodeAbi\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DuplicateApproval\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotEnoughWithdrawableBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotOwnerOrProvider\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotOwnerOrUser\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account_\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"PoolIndexOutOfBounds\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProviderSignatureMismatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionAlreadyClosed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionNotClosed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionTooShort\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SignatureExpired\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"WithdrawableBalanceLimitByStakeReached\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"WrongChainId\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"storageSlot\",\"type\":\"bytes32\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerId\",\"type\":\"address\"}],\"name\":\"SessionClosed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerId\",\"type\":\"address\"}],\"name\":\"SessionOpened\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BID_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"COMPUTE_POOL_INDEX\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DIAMOND_OWNABLE_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_SESSION_DURATION\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_SESSION_DURATION\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PROVIDER_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SESSION_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SIGNATURE_TTL\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"STATS_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"fundingAccount_\",\"type\":\"address\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"initialReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardDecrease\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"payoutStart\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"decreaseInterval\",\"type\":\"uint128\"}],\"internalType\":\"structISessionStorage.Pool[]\",\"name\":\"pools_\",\"type\":\"tuple[]\"}],\"name\":\"__SessionRouter_init\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"}],\"name\":\"bids\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"deletedAt\",\"type\":\"uint128\"}],\"internalType\":\"structIBidStorage.Bid\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amountToWithdraw_\",\"type\":\"uint256\"}],\"name\":\"claimProviderBalance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"receiptEncoded_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature_\",\"type\":\"bytes\"}],\"name\":\"closeSession\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId_\",\"type\":\"bytes32\"}],\"name\":\"deleteHistory\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"limit_\",\"type\":\"uint8\"}],\"name\":\"getActiveBidsRatingByModel\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"deletedAt\",\"type\":\"uint128\"}],\"internalType\":\"structIBidStorage.Bid[]\",\"name\":\"\",\"type\":\"tuple[]\"},{\"components\":[{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"tpsScaled1000\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"ttftMs\",\"type\":\"tuple\"},{\"internalType\":\"uint32\",\"name\":\"totalDuration\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"successCount\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"totalCount\",\"type\":\"uint32\"}],\"internalType\":\"structIStatsStorage.ProviderModelStats[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp_\",\"type\":\"uint256\"}],\"name\":\"getComputeBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getFundingAccount\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"getProvider\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"endpoint\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"limitPeriodEnd\",\"type\":\"uint128\"},{\"internalType\":\"uint256\",\"name\":\"limitPeriodEarned\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"internalType\":\"structIProviderStorage.Provider\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId_\",\"type\":\"bytes32\"}],\"name\":\"getProviderClaimableBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"stake_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"sessionNonce_\",\"type\":\"uint256\"}],\"name\":\"getSessionId\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"getSessionsByUser\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp_\",\"type\":\"uint256\"}],\"name\":\"getTodaysBudget\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"modelActiveBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"modelBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"providerApproval_\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature_\",\"type\":\"bytes\"}],\"name\":\"openSession\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pools\",\"outputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"initialReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardDecrease\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"payoutStart\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"decreaseInterval\",\"type\":\"uint128\"}],\"internalType\":\"structISessionStorage.Pool[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"providerActiveBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"providerBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"providerMinimumStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"}],\"name\":\"sessions\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"closeoutReceipt\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerWithdrawnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"openedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endsAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"closedAt\",\"type\":\"uint256\"}],\"internalType\":\"structISessionStorage.Session\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"initialReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardDecrease\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"payoutStart\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"decreaseInterval\",\"type\":\"uint128\"}],\"internalType\":\"structISessionStorage.Pool\",\"name\":\"pool\",\"type\":\"tuple\"}],\"name\":\"setPoolConfig\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"sessionStake_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp_\",\"type\":\"uint256\"}],\"name\":\"stakeToStipend\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp_\",\"type\":\"uint256\"}],\"name\":\"startOfTheDay\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"stipend_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp_\",\"type\":\"uint256\"}],\"name\":\"stipendToStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp_\",\"type\":\"uint256\"}],\"name\":\"totalMORSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"sessionStake_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"openedAt_\",\"type\":\"uint256\"}],\"name\":\"whenSessionEnds\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amountToWithdraw_\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"iterations_\",\"type\":\"uint8\"}],\"name\":\"withdrawUserStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user_\",\"type\":\"address\"},{\"internalType\":\"uint8\",\"name\":\"iterations_\",\"type\":\"uint8\"}],\"name\":\"withdrawableUserStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"avail_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"hold_\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
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

// BIDSTORAGESLOT is a free data retrieval call binding the contract method 0x4fa816f2.
//
// Solidity: function BID_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCaller) BIDSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "BID_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BIDSTORAGESLOT is a free data retrieval call binding the contract method 0x4fa816f2.
//
// Solidity: function BID_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterSession) BIDSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.BIDSTORAGESLOT(&_SessionRouter.CallOpts)
}

// BIDSTORAGESLOT is a free data retrieval call binding the contract method 0x4fa816f2.
//
// Solidity: function BID_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) BIDSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.BIDSTORAGESLOT(&_SessionRouter.CallOpts)
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

// MAXSESSIONDURATION is a free data retrieval call binding the contract method 0xcd8cd4ad.
//
// Solidity: function MAX_SESSION_DURATION() view returns(uint32)
func (_SessionRouter *SessionRouterCaller) MAXSESSIONDURATION(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "MAX_SESSION_DURATION")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// MAXSESSIONDURATION is a free data retrieval call binding the contract method 0xcd8cd4ad.
//
// Solidity: function MAX_SESSION_DURATION() view returns(uint32)
func (_SessionRouter *SessionRouterSession) MAXSESSIONDURATION() (uint32, error) {
	return _SessionRouter.Contract.MAXSESSIONDURATION(&_SessionRouter.CallOpts)
}

// MAXSESSIONDURATION is a free data retrieval call binding the contract method 0xcd8cd4ad.
//
// Solidity: function MAX_SESSION_DURATION() view returns(uint32)
func (_SessionRouter *SessionRouterCallerSession) MAXSESSIONDURATION() (uint32, error) {
	return _SessionRouter.Contract.MAXSESSIONDURATION(&_SessionRouter.CallOpts)
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

// PROVIDERSTORAGESLOT is a free data retrieval call binding the contract method 0x490713b1.
//
// Solidity: function PROVIDER_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCaller) PROVIDERSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "PROVIDER_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PROVIDERSTORAGESLOT is a free data retrieval call binding the contract method 0x490713b1.
//
// Solidity: function PROVIDER_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterSession) PROVIDERSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.PROVIDERSTORAGESLOT(&_SessionRouter.CallOpts)
}

// PROVIDERSTORAGESLOT is a free data retrieval call binding the contract method 0x490713b1.
//
// Solidity: function PROVIDER_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) PROVIDERSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.PROVIDERSTORAGESLOT(&_SessionRouter.CallOpts)
}

// SESSIONSTORAGESLOT is a free data retrieval call binding the contract method 0x0cbfb226.
//
// Solidity: function SESSION_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCaller) SESSIONSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "SESSION_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// SESSIONSTORAGESLOT is a free data retrieval call binding the contract method 0x0cbfb226.
//
// Solidity: function SESSION_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterSession) SESSIONSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.SESSIONSTORAGESLOT(&_SessionRouter.CallOpts)
}

// SESSIONSTORAGESLOT is a free data retrieval call binding the contract method 0x0cbfb226.
//
// Solidity: function SESSION_STORAGE_SLOT() view returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) SESSIONSTORAGESLOT() ([32]byte, error) {
	return _SessionRouter.Contract.SESSIONSTORAGESLOT(&_SessionRouter.CallOpts)
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

// Bids is a free data retrieval call binding the contract method 0x8f98eeda.
//
// Solidity: function bids(bytes32 bidId) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_SessionRouter *SessionRouterCaller) Bids(opts *bind.CallOpts, bidId [32]byte) (IBidStorageBid, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "bids", bidId)

	if err != nil {
		return *new(IBidStorageBid), err
	}

	out0 := *abi.ConvertType(out[0], new(IBidStorageBid)).(*IBidStorageBid)

	return out0, err

}

// Bids is a free data retrieval call binding the contract method 0x8f98eeda.
//
// Solidity: function bids(bytes32 bidId) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_SessionRouter *SessionRouterSession) Bids(bidId [32]byte) (IBidStorageBid, error) {
	return _SessionRouter.Contract.Bids(&_SessionRouter.CallOpts, bidId)
}

// Bids is a free data retrieval call binding the contract method 0x8f98eeda.
//
// Solidity: function bids(bytes32 bidId) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_SessionRouter *SessionRouterCallerSession) Bids(bidId [32]byte) (IBidStorageBid, error) {
	return _SessionRouter.Contract.Bids(&_SessionRouter.CallOpts, bidId)
}

// GetActiveBidsRatingByModel is a free data retrieval call binding the contract method 0x3b04deec.
//
// Solidity: function getActiveBidsRatingByModel(bytes32 modelId_, uint256 offset_, uint8 limit_) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[], ((int64,int64),(int64,int64),uint32,uint32,uint32)[])
func (_SessionRouter *SessionRouterCaller) GetActiveBidsRatingByModel(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ uint8) ([][32]byte, []IBidStorageBid, []IStatsStorageProviderModelStats, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getActiveBidsRatingByModel", modelId_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), *new([]IBidStorageBid), *new([]IStatsStorageProviderModelStats), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)
	out1 := *abi.ConvertType(out[1], new([]IBidStorageBid)).(*[]IBidStorageBid)
	out2 := *abi.ConvertType(out[2], new([]IStatsStorageProviderModelStats)).(*[]IStatsStorageProviderModelStats)

	return out0, out1, out2, err

}

// GetActiveBidsRatingByModel is a free data retrieval call binding the contract method 0x3b04deec.
//
// Solidity: function getActiveBidsRatingByModel(bytes32 modelId_, uint256 offset_, uint8 limit_) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[], ((int64,int64),(int64,int64),uint32,uint32,uint32)[])
func (_SessionRouter *SessionRouterSession) GetActiveBidsRatingByModel(modelId_ [32]byte, offset_ *big.Int, limit_ uint8) ([][32]byte, []IBidStorageBid, []IStatsStorageProviderModelStats, error) {
	return _SessionRouter.Contract.GetActiveBidsRatingByModel(&_SessionRouter.CallOpts, modelId_, offset_, limit_)
}

// GetActiveBidsRatingByModel is a free data retrieval call binding the contract method 0x3b04deec.
//
// Solidity: function getActiveBidsRatingByModel(bytes32 modelId_, uint256 offset_, uint8 limit_) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[], ((int64,int64),(int64,int64),uint32,uint32,uint32)[])
func (_SessionRouter *SessionRouterCallerSession) GetActiveBidsRatingByModel(modelId_ [32]byte, offset_ *big.Int, limit_ uint8) ([][32]byte, []IBidStorageBid, []IStatsStorageProviderModelStats, error) {
	return _SessionRouter.Contract.GetActiveBidsRatingByModel(&_SessionRouter.CallOpts, modelId_, offset_, limit_)
}

// GetComputeBalance is a free data retrieval call binding the contract method 0x76738e9e.
//
// Solidity: function getComputeBalance(uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) GetComputeBalance(opts *bind.CallOpts, timestamp_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getComputeBalance", timestamp_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetComputeBalance is a free data retrieval call binding the contract method 0x76738e9e.
//
// Solidity: function getComputeBalance(uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterSession) GetComputeBalance(timestamp_ *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.GetComputeBalance(&_SessionRouter.CallOpts, timestamp_)
}

// GetComputeBalance is a free data retrieval call binding the contract method 0x76738e9e.
//
// Solidity: function getComputeBalance(uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) GetComputeBalance(timestamp_ *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.GetComputeBalance(&_SessionRouter.CallOpts, timestamp_)
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

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_SessionRouter *SessionRouterCaller) GetProvider(opts *bind.CallOpts, provider common.Address) (IProviderStorageProvider, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getProvider", provider)

	if err != nil {
		return *new(IProviderStorageProvider), err
	}

	out0 := *abi.ConvertType(out[0], new(IProviderStorageProvider)).(*IProviderStorageProvider)

	return out0, err

}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_SessionRouter *SessionRouterSession) GetProvider(provider common.Address) (IProviderStorageProvider, error) {
	return _SessionRouter.Contract.GetProvider(&_SessionRouter.CallOpts, provider)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_SessionRouter *SessionRouterCallerSession) GetProvider(provider common.Address) (IProviderStorageProvider, error) {
	return _SessionRouter.Contract.GetProvider(&_SessionRouter.CallOpts, provider)
}

// GetProviderClaimableBalance is a free data retrieval call binding the contract method 0xa8ca6323.
//
// Solidity: function getProviderClaimableBalance(bytes32 sessionId_) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) GetProviderClaimableBalance(opts *bind.CallOpts, sessionId_ [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getProviderClaimableBalance", sessionId_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProviderClaimableBalance is a free data retrieval call binding the contract method 0xa8ca6323.
//
// Solidity: function getProviderClaimableBalance(bytes32 sessionId_) view returns(uint256)
func (_SessionRouter *SessionRouterSession) GetProviderClaimableBalance(sessionId_ [32]byte) (*big.Int, error) {
	return _SessionRouter.Contract.GetProviderClaimableBalance(&_SessionRouter.CallOpts, sessionId_)
}

// GetProviderClaimableBalance is a free data retrieval call binding the contract method 0xa8ca6323.
//
// Solidity: function getProviderClaimableBalance(bytes32 sessionId_) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) GetProviderClaimableBalance(sessionId_ [32]byte) (*big.Int, error) {
	return _SessionRouter.Contract.GetProviderClaimableBalance(&_SessionRouter.CallOpts, sessionId_)
}

// GetSessionId is a free data retrieval call binding the contract method 0x0f9de78a.
//
// Solidity: function getSessionId(address user_, address provider_, uint256 stake_, uint256 sessionNonce_) pure returns(bytes32)
func (_SessionRouter *SessionRouterCaller) GetSessionId(opts *bind.CallOpts, user_ common.Address, provider_ common.Address, stake_ *big.Int, sessionNonce_ *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getSessionId", user_, provider_, stake_, sessionNonce_)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetSessionId is a free data retrieval call binding the contract method 0x0f9de78a.
//
// Solidity: function getSessionId(address user_, address provider_, uint256 stake_, uint256 sessionNonce_) pure returns(bytes32)
func (_SessionRouter *SessionRouterSession) GetSessionId(user_ common.Address, provider_ common.Address, stake_ *big.Int, sessionNonce_ *big.Int) ([32]byte, error) {
	return _SessionRouter.Contract.GetSessionId(&_SessionRouter.CallOpts, user_, provider_, stake_, sessionNonce_)
}

// GetSessionId is a free data retrieval call binding the contract method 0x0f9de78a.
//
// Solidity: function getSessionId(address user_, address provider_, uint256 stake_, uint256 sessionNonce_) pure returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) GetSessionId(user_ common.Address, provider_ common.Address, stake_ *big.Int, sessionNonce_ *big.Int) ([32]byte, error) {
	return _SessionRouter.Contract.GetSessionId(&_SessionRouter.CallOpts, user_, provider_, stake_, sessionNonce_)
}

// GetSessionsByUser is a free data retrieval call binding the contract method 0x4add952d.
//
// Solidity: function getSessionsByUser(address user, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterCaller) GetSessionsByUser(opts *bind.CallOpts, user common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getSessionsByUser", user, offset_, limit_)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetSessionsByUser is a free data retrieval call binding the contract method 0x4add952d.
//
// Solidity: function getSessionsByUser(address user, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterSession) GetSessionsByUser(user common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _SessionRouter.Contract.GetSessionsByUser(&_SessionRouter.CallOpts, user, offset_, limit_)
}

// GetSessionsByUser is a free data retrieval call binding the contract method 0x4add952d.
//
// Solidity: function getSessionsByUser(address user, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterCallerSession) GetSessionsByUser(user common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _SessionRouter.Contract.GetSessionsByUser(&_SessionRouter.CallOpts, user, offset_, limit_)
}

// GetTodaysBudget is a free data retrieval call binding the contract method 0x351ffeb0.
//
// Solidity: function getTodaysBudget(uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) GetTodaysBudget(opts *bind.CallOpts, timestamp_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getTodaysBudget", timestamp_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTodaysBudget is a free data retrieval call binding the contract method 0x351ffeb0.
//
// Solidity: function getTodaysBudget(uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterSession) GetTodaysBudget(timestamp_ *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.GetTodaysBudget(&_SessionRouter.CallOpts, timestamp_)
}

// GetTodaysBudget is a free data retrieval call binding the contract method 0x351ffeb0.
//
// Solidity: function getTodaysBudget(uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) GetTodaysBudget(timestamp_ *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.GetTodaysBudget(&_SessionRouter.CallOpts, timestamp_)
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

// ModelActiveBids is a free data retrieval call binding the contract method 0x3fd8e5e3.
//
// Solidity: function modelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterCaller) ModelActiveBids(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "modelActiveBids", modelId_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// ModelActiveBids is a free data retrieval call binding the contract method 0x3fd8e5e3.
//
// Solidity: function modelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterSession) ModelActiveBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _SessionRouter.Contract.ModelActiveBids(&_SessionRouter.CallOpts, modelId_, offset_, limit_)
}

// ModelActiveBids is a free data retrieval call binding the contract method 0x3fd8e5e3.
//
// Solidity: function modelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterCallerSession) ModelActiveBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _SessionRouter.Contract.ModelActiveBids(&_SessionRouter.CallOpts, modelId_, offset_, limit_)
}

// ModelBids is a free data retrieval call binding the contract method 0x5954d1b3.
//
// Solidity: function modelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterCaller) ModelBids(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "modelBids", modelId_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// ModelBids is a free data retrieval call binding the contract method 0x5954d1b3.
//
// Solidity: function modelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterSession) ModelBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _SessionRouter.Contract.ModelBids(&_SessionRouter.CallOpts, modelId_, offset_, limit_)
}

// ModelBids is a free data retrieval call binding the contract method 0x5954d1b3.
//
// Solidity: function modelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterCallerSession) ModelBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _SessionRouter.Contract.ModelBids(&_SessionRouter.CallOpts, modelId_, offset_, limit_)
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

// Pools is a free data retrieval call binding the contract method 0xc5c51dca.
//
// Solidity: function pools() view returns((uint256,uint256,uint128,uint128)[])
func (_SessionRouter *SessionRouterCaller) Pools(opts *bind.CallOpts) ([]ISessionStoragePool, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "pools")

	if err != nil {
		return *new([]ISessionStoragePool), err
	}

	out0 := *abi.ConvertType(out[0], new([]ISessionStoragePool)).(*[]ISessionStoragePool)

	return out0, err

}

// Pools is a free data retrieval call binding the contract method 0xc5c51dca.
//
// Solidity: function pools() view returns((uint256,uint256,uint128,uint128)[])
func (_SessionRouter *SessionRouterSession) Pools() ([]ISessionStoragePool, error) {
	return _SessionRouter.Contract.Pools(&_SessionRouter.CallOpts)
}

// Pools is a free data retrieval call binding the contract method 0xc5c51dca.
//
// Solidity: function pools() view returns((uint256,uint256,uint128,uint128)[])
func (_SessionRouter *SessionRouterCallerSession) Pools() ([]ISessionStoragePool, error) {
	return _SessionRouter.Contract.Pools(&_SessionRouter.CallOpts)
}

// ProviderActiveBids is a free data retrieval call binding the contract method 0x6dd7d31c.
//
// Solidity: function providerActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterCaller) ProviderActiveBids(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "providerActiveBids", provider_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// ProviderActiveBids is a free data retrieval call binding the contract method 0x6dd7d31c.
//
// Solidity: function providerActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterSession) ProviderActiveBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _SessionRouter.Contract.ProviderActiveBids(&_SessionRouter.CallOpts, provider_, offset_, limit_)
}

// ProviderActiveBids is a free data retrieval call binding the contract method 0x6dd7d31c.
//
// Solidity: function providerActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterCallerSession) ProviderActiveBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _SessionRouter.Contract.ProviderActiveBids(&_SessionRouter.CallOpts, provider_, offset_, limit_)
}

// ProviderBids is a free data retrieval call binding the contract method 0x22fbda9a.
//
// Solidity: function providerBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterCaller) ProviderBids(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "providerBids", provider_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// ProviderBids is a free data retrieval call binding the contract method 0x22fbda9a.
//
// Solidity: function providerBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterSession) ProviderBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _SessionRouter.Contract.ProviderBids(&_SessionRouter.CallOpts, provider_, offset_, limit_)
}

// ProviderBids is a free data retrieval call binding the contract method 0x22fbda9a.
//
// Solidity: function providerBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_SessionRouter *SessionRouterCallerSession) ProviderBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _SessionRouter.Contract.ProviderBids(&_SessionRouter.CallOpts, provider_, offset_, limit_)
}

// ProviderMinimumStake is a free data retrieval call binding the contract method 0x9476c58e.
//
// Solidity: function providerMinimumStake() view returns(uint256)
func (_SessionRouter *SessionRouterCaller) ProviderMinimumStake(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "providerMinimumStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProviderMinimumStake is a free data retrieval call binding the contract method 0x9476c58e.
//
// Solidity: function providerMinimumStake() view returns(uint256)
func (_SessionRouter *SessionRouterSession) ProviderMinimumStake() (*big.Int, error) {
	return _SessionRouter.Contract.ProviderMinimumStake(&_SessionRouter.CallOpts)
}

// ProviderMinimumStake is a free data retrieval call binding the contract method 0x9476c58e.
//
// Solidity: function providerMinimumStake() view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) ProviderMinimumStake() (*big.Int, error) {
	return _SessionRouter.Contract.ProviderMinimumStake(&_SessionRouter.CallOpts)
}

// Sessions is a free data retrieval call binding the contract method 0x7dbd2832.
//
// Solidity: function sessions(bytes32 sessionId) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256))
func (_SessionRouter *SessionRouterCaller) Sessions(opts *bind.CallOpts, sessionId [32]byte) (ISessionStorageSession, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "sessions", sessionId)

	if err != nil {
		return *new(ISessionStorageSession), err
	}

	out0 := *abi.ConvertType(out[0], new(ISessionStorageSession)).(*ISessionStorageSession)

	return out0, err

}

// Sessions is a free data retrieval call binding the contract method 0x7dbd2832.
//
// Solidity: function sessions(bytes32 sessionId) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256))
func (_SessionRouter *SessionRouterSession) Sessions(sessionId [32]byte) (ISessionStorageSession, error) {
	return _SessionRouter.Contract.Sessions(&_SessionRouter.CallOpts, sessionId)
}

// Sessions is a free data retrieval call binding the contract method 0x7dbd2832.
//
// Solidity: function sessions(bytes32 sessionId) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256))
func (_SessionRouter *SessionRouterCallerSession) Sessions(sessionId [32]byte) (ISessionStorageSession, error) {
	return _SessionRouter.Contract.Sessions(&_SessionRouter.CallOpts, sessionId)
}

// StakeToStipend is a free data retrieval call binding the contract method 0x0a23b21f.
//
// Solidity: function stakeToStipend(uint256 sessionStake_, uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) StakeToStipend(opts *bind.CallOpts, sessionStake_ *big.Int, timestamp_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "stakeToStipend", sessionStake_, timestamp_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StakeToStipend is a free data retrieval call binding the contract method 0x0a23b21f.
//
// Solidity: function stakeToStipend(uint256 sessionStake_, uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterSession) StakeToStipend(sessionStake_ *big.Int, timestamp_ *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.StakeToStipend(&_SessionRouter.CallOpts, sessionStake_, timestamp_)
}

// StakeToStipend is a free data retrieval call binding the contract method 0x0a23b21f.
//
// Solidity: function stakeToStipend(uint256 sessionStake_, uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) StakeToStipend(sessionStake_ *big.Int, timestamp_ *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.StakeToStipend(&_SessionRouter.CallOpts, sessionStake_, timestamp_)
}

// StartOfTheDay is a free data retrieval call binding the contract method 0xeedd0a72.
//
// Solidity: function startOfTheDay(uint256 timestamp_) pure returns(uint256)
func (_SessionRouter *SessionRouterCaller) StartOfTheDay(opts *bind.CallOpts, timestamp_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "startOfTheDay", timestamp_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StartOfTheDay is a free data retrieval call binding the contract method 0xeedd0a72.
//
// Solidity: function startOfTheDay(uint256 timestamp_) pure returns(uint256)
func (_SessionRouter *SessionRouterSession) StartOfTheDay(timestamp_ *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.StartOfTheDay(&_SessionRouter.CallOpts, timestamp_)
}

// StartOfTheDay is a free data retrieval call binding the contract method 0xeedd0a72.
//
// Solidity: function startOfTheDay(uint256 timestamp_) pure returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) StartOfTheDay(timestamp_ *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.StartOfTheDay(&_SessionRouter.CallOpts, timestamp_)
}

// StipendToStake is a free data retrieval call binding the contract method 0xac3c19ce.
//
// Solidity: function stipendToStake(uint256 stipend_, uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) StipendToStake(opts *bind.CallOpts, stipend_ *big.Int, timestamp_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "stipendToStake", stipend_, timestamp_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StipendToStake is a free data retrieval call binding the contract method 0xac3c19ce.
//
// Solidity: function stipendToStake(uint256 stipend_, uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterSession) StipendToStake(stipend_ *big.Int, timestamp_ *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.StipendToStake(&_SessionRouter.CallOpts, stipend_, timestamp_)
}

// StipendToStake is a free data retrieval call binding the contract method 0xac3c19ce.
//
// Solidity: function stipendToStake(uint256 stipend_, uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) StipendToStake(stipend_ *big.Int, timestamp_ *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.StipendToStake(&_SessionRouter.CallOpts, stipend_, timestamp_)
}

// TotalMORSupply is a free data retrieval call binding the contract method 0xf1d5440c.
//
// Solidity: function totalMORSupply(uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) TotalMORSupply(opts *bind.CallOpts, timestamp_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "totalMORSupply", timestamp_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalMORSupply is a free data retrieval call binding the contract method 0xf1d5440c.
//
// Solidity: function totalMORSupply(uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterSession) TotalMORSupply(timestamp_ *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.TotalMORSupply(&_SessionRouter.CallOpts, timestamp_)
}

// TotalMORSupply is a free data retrieval call binding the contract method 0xf1d5440c.
//
// Solidity: function totalMORSupply(uint256 timestamp_) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) TotalMORSupply(timestamp_ *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.TotalMORSupply(&_SessionRouter.CallOpts, timestamp_)
}

// WhenSessionEnds is a free data retrieval call binding the contract method 0x9bc08456.
//
// Solidity: function whenSessionEnds(uint256 sessionStake_, uint256 pricePerSecond_, uint256 openedAt_) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) WhenSessionEnds(opts *bind.CallOpts, sessionStake_ *big.Int, pricePerSecond_ *big.Int, openedAt_ *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "whenSessionEnds", sessionStake_, pricePerSecond_, openedAt_)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WhenSessionEnds is a free data retrieval call binding the contract method 0x9bc08456.
//
// Solidity: function whenSessionEnds(uint256 sessionStake_, uint256 pricePerSecond_, uint256 openedAt_) view returns(uint256)
func (_SessionRouter *SessionRouterSession) WhenSessionEnds(sessionStake_ *big.Int, pricePerSecond_ *big.Int, openedAt_ *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.WhenSessionEnds(&_SessionRouter.CallOpts, sessionStake_, pricePerSecond_, openedAt_)
}

// WhenSessionEnds is a free data retrieval call binding the contract method 0x9bc08456.
//
// Solidity: function whenSessionEnds(uint256 sessionStake_, uint256 pricePerSecond_, uint256 openedAt_) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) WhenSessionEnds(sessionStake_ *big.Int, pricePerSecond_ *big.Int, openedAt_ *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.WhenSessionEnds(&_SessionRouter.CallOpts, sessionStake_, pricePerSecond_, openedAt_)
}

// WithdrawableUserStake is a free data retrieval call binding the contract method 0x65bf0e3c.
//
// Solidity: function withdrawableUserStake(address user_, uint8 iterations_) view returns(uint256 avail_, uint256 hold_)
func (_SessionRouter *SessionRouterCaller) WithdrawableUserStake(opts *bind.CallOpts, user_ common.Address, iterations_ uint8) (struct {
	Avail *big.Int
	Hold  *big.Int
}, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "withdrawableUserStake", user_, iterations_)

	outstruct := new(struct {
		Avail *big.Int
		Hold  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Avail = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Hold = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// WithdrawableUserStake is a free data retrieval call binding the contract method 0x65bf0e3c.
//
// Solidity: function withdrawableUserStake(address user_, uint8 iterations_) view returns(uint256 avail_, uint256 hold_)
func (_SessionRouter *SessionRouterSession) WithdrawableUserStake(user_ common.Address, iterations_ uint8) (struct {
	Avail *big.Int
	Hold  *big.Int
}, error) {
	return _SessionRouter.Contract.WithdrawableUserStake(&_SessionRouter.CallOpts, user_, iterations_)
}

// WithdrawableUserStake is a free data retrieval call binding the contract method 0x65bf0e3c.
//
// Solidity: function withdrawableUserStake(address user_, uint8 iterations_) view returns(uint256 avail_, uint256 hold_)
func (_SessionRouter *SessionRouterCallerSession) WithdrawableUserStake(user_ common.Address, iterations_ uint8) (struct {
	Avail *big.Int
	Hold  *big.Int
}, error) {
	return _SessionRouter.Contract.WithdrawableUserStake(&_SessionRouter.CallOpts, user_, iterations_)
}

// SessionRouterInit is a paid mutator transaction binding the contract method 0x44ceb8e0.
//
// Solidity: function __SessionRouter_init(address fundingAccount_, (uint256,uint256,uint128,uint128)[] pools_) returns()
func (_SessionRouter *SessionRouterTransactor) SessionRouterInit(opts *bind.TransactOpts, fundingAccount_ common.Address, pools_ []ISessionStoragePool) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "__SessionRouter_init", fundingAccount_, pools_)
}

// SessionRouterInit is a paid mutator transaction binding the contract method 0x44ceb8e0.
//
// Solidity: function __SessionRouter_init(address fundingAccount_, (uint256,uint256,uint128,uint128)[] pools_) returns()
func (_SessionRouter *SessionRouterSession) SessionRouterInit(fundingAccount_ common.Address, pools_ []ISessionStoragePool) (*types.Transaction, error) {
	return _SessionRouter.Contract.SessionRouterInit(&_SessionRouter.TransactOpts, fundingAccount_, pools_)
}

// SessionRouterInit is a paid mutator transaction binding the contract method 0x44ceb8e0.
//
// Solidity: function __SessionRouter_init(address fundingAccount_, (uint256,uint256,uint128,uint128)[] pools_) returns()
func (_SessionRouter *SessionRouterTransactorSession) SessionRouterInit(fundingAccount_ common.Address, pools_ []ISessionStoragePool) (*types.Transaction, error) {
	return _SessionRouter.Contract.SessionRouterInit(&_SessionRouter.TransactOpts, fundingAccount_, pools_)
}

// ClaimProviderBalance is a paid mutator transaction binding the contract method 0xf42d165a.
//
// Solidity: function claimProviderBalance(bytes32 sessionId_, uint256 amountToWithdraw_) returns()
func (_SessionRouter *SessionRouterTransactor) ClaimProviderBalance(opts *bind.TransactOpts, sessionId_ [32]byte, amountToWithdraw_ *big.Int) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "claimProviderBalance", sessionId_, amountToWithdraw_)
}

// ClaimProviderBalance is a paid mutator transaction binding the contract method 0xf42d165a.
//
// Solidity: function claimProviderBalance(bytes32 sessionId_, uint256 amountToWithdraw_) returns()
func (_SessionRouter *SessionRouterSession) ClaimProviderBalance(sessionId_ [32]byte, amountToWithdraw_ *big.Int) (*types.Transaction, error) {
	return _SessionRouter.Contract.ClaimProviderBalance(&_SessionRouter.TransactOpts, sessionId_, amountToWithdraw_)
}

// ClaimProviderBalance is a paid mutator transaction binding the contract method 0xf42d165a.
//
// Solidity: function claimProviderBalance(bytes32 sessionId_, uint256 amountToWithdraw_) returns()
func (_SessionRouter *SessionRouterTransactorSession) ClaimProviderBalance(sessionId_ [32]byte, amountToWithdraw_ *big.Int) (*types.Transaction, error) {
	return _SessionRouter.Contract.ClaimProviderBalance(&_SessionRouter.TransactOpts, sessionId_, amountToWithdraw_)
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

// DeleteHistory is a paid mutator transaction binding the contract method 0xf074ca6b.
//
// Solidity: function deleteHistory(bytes32 sessionId_) returns()
func (_SessionRouter *SessionRouterTransactor) DeleteHistory(opts *bind.TransactOpts, sessionId_ [32]byte) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "deleteHistory", sessionId_)
}

// DeleteHistory is a paid mutator transaction binding the contract method 0xf074ca6b.
//
// Solidity: function deleteHistory(bytes32 sessionId_) returns()
func (_SessionRouter *SessionRouterSession) DeleteHistory(sessionId_ [32]byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.DeleteHistory(&_SessionRouter.TransactOpts, sessionId_)
}

// DeleteHistory is a paid mutator transaction binding the contract method 0xf074ca6b.
//
// Solidity: function deleteHistory(bytes32 sessionId_) returns()
func (_SessionRouter *SessionRouterTransactorSession) DeleteHistory(sessionId_ [32]byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.DeleteHistory(&_SessionRouter.TransactOpts, sessionId_)
}

// OpenSession is a paid mutator transaction binding the contract method 0x1f71815e.
//
// Solidity: function openSession(uint256 amount_, bytes providerApproval_, bytes signature_) returns(bytes32)
func (_SessionRouter *SessionRouterTransactor) OpenSession(opts *bind.TransactOpts, amount_ *big.Int, providerApproval_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "openSession", amount_, providerApproval_, signature_)
}

// OpenSession is a paid mutator transaction binding the contract method 0x1f71815e.
//
// Solidity: function openSession(uint256 amount_, bytes providerApproval_, bytes signature_) returns(bytes32)
func (_SessionRouter *SessionRouterSession) OpenSession(amount_ *big.Int, providerApproval_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.OpenSession(&_SessionRouter.TransactOpts, amount_, providerApproval_, signature_)
}

// OpenSession is a paid mutator transaction binding the contract method 0x1f71815e.
//
// Solidity: function openSession(uint256 amount_, bytes providerApproval_, bytes signature_) returns(bytes32)
func (_SessionRouter *SessionRouterTransactorSession) OpenSession(amount_ *big.Int, providerApproval_ []byte, signature_ []byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.OpenSession(&_SessionRouter.TransactOpts, amount_, providerApproval_, signature_)
}

// SetPoolConfig is a paid mutator transaction binding the contract method 0xd7178753.
//
// Solidity: function setPoolConfig(uint256 index, (uint256,uint256,uint128,uint128) pool) returns()
func (_SessionRouter *SessionRouterTransactor) SetPoolConfig(opts *bind.TransactOpts, index *big.Int, pool ISessionStoragePool) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "setPoolConfig", index, pool)
}

// SetPoolConfig is a paid mutator transaction binding the contract method 0xd7178753.
//
// Solidity: function setPoolConfig(uint256 index, (uint256,uint256,uint128,uint128) pool) returns()
func (_SessionRouter *SessionRouterSession) SetPoolConfig(index *big.Int, pool ISessionStoragePool) (*types.Transaction, error) {
	return _SessionRouter.Contract.SetPoolConfig(&_SessionRouter.TransactOpts, index, pool)
}

// SetPoolConfig is a paid mutator transaction binding the contract method 0xd7178753.
//
// Solidity: function setPoolConfig(uint256 index, (uint256,uint256,uint128,uint128) pool) returns()
func (_SessionRouter *SessionRouterTransactorSession) SetPoolConfig(index *big.Int, pool ISessionStoragePool) (*types.Transaction, error) {
	return _SessionRouter.Contract.SetPoolConfig(&_SessionRouter.TransactOpts, index, pool)
}

// WithdrawUserStake is a paid mutator transaction binding the contract method 0x0fd2c44e.
//
// Solidity: function withdrawUserStake(uint256 amountToWithdraw_, uint8 iterations_) returns()
func (_SessionRouter *SessionRouterTransactor) WithdrawUserStake(opts *bind.TransactOpts, amountToWithdraw_ *big.Int, iterations_ uint8) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "withdrawUserStake", amountToWithdraw_, iterations_)
}

// WithdrawUserStake is a paid mutator transaction binding the contract method 0x0fd2c44e.
//
// Solidity: function withdrawUserStake(uint256 amountToWithdraw_, uint8 iterations_) returns()
func (_SessionRouter *SessionRouterSession) WithdrawUserStake(amountToWithdraw_ *big.Int, iterations_ uint8) (*types.Transaction, error) {
	return _SessionRouter.Contract.WithdrawUserStake(&_SessionRouter.TransactOpts, amountToWithdraw_, iterations_)
}

// WithdrawUserStake is a paid mutator transaction binding the contract method 0x0fd2c44e.
//
// Solidity: function withdrawUserStake(uint256 amountToWithdraw_, uint8 iterations_) returns()
func (_SessionRouter *SessionRouterTransactorSession) WithdrawUserStake(amountToWithdraw_ *big.Int, iterations_ uint8) (*types.Transaction, error) {
	return _SessionRouter.Contract.WithdrawUserStake(&_SessionRouter.TransactOpts, amountToWithdraw_, iterations_)
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
