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

// Pool is an auto generated low-level Go binding around an user-defined struct.
type Pool struct {
	InitialReward    *big.Int
	RewardDecrease   *big.Int
	PayoutStart      *big.Int
	DecreaseInterval *big.Int
}

// Session is an auto generated low-level Go binding around an user-defined struct.
type Session struct {
	Id                      [32]byte
	User                    common.Address
	Provider                common.Address
	ModelAgentId            [32]byte
	BidID                   [32]byte
	Stake                   *big.Int
	PricePerSecond          *big.Int
	CloseoutReceipt         []byte
	CloseoutType            *big.Int
	ProviderWithdrawnAmount *big.Int
	OpenedAt                *big.Int
	EndsAt                  *big.Int
	ClosedAt                *big.Int
}

// SessionRouterMetaData contains all meta data concerning the SessionRouter contract.
var SessionRouterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"BidNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"CannotDecodeAbi\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"DuplicateApproval\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ECDSAInvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"ECDSAInvalidSignatureLength\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"ECDSAInvalidSignatureS\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"KeyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"KeyNotFound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_contractOwner\",\"type\":\"address\"}],\"name\":\"NotContractOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotEnoughWithdrawableBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSenderOrOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotUserOrProvider\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProviderSignatureMismatch\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionAlreadyClosed\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionTooShort\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SignatureExpired\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerId\",\"type\":\"address\"}],\"name\":\"SessionClosed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerId\",\"type\":\"address\"}],\"name\":\"SessionOpened\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MAX_SESSION_DURATION\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_SESSION_DURATION\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"SIGNATURE_TTL\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"activeSessionsCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amountToWithdraw\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"claimProviderBalance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"receiptEncoded\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"closeSession\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"}],\"name\":\"deleteHistory\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"getActiveSessionsByProvider\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"bidID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"closeoutReceipt\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerWithdrawnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"openedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endsAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"closedAt\",\"type\":\"uint256\"}],\"internalType\":\"structSession[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"}],\"name\":\"getActiveSessionsByUser\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"bidID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"closeoutReceipt\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerWithdrawnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"openedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endsAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"closedAt\",\"type\":\"uint256\"}],\"internalType\":\"structSession[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"getComputeBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"}],\"name\":\"getProviderClaimableBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"}],\"name\":\"getSession\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"bidID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"closeoutReceipt\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerWithdrawnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"openedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endsAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"closedAt\",\"type\":\"uint256\"}],\"internalType\":\"structSession\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"limit\",\"type\":\"uint8\"}],\"name\":\"getSessionsByModel\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"bidID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"closeoutReceipt\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerWithdrawnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"openedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endsAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"closedAt\",\"type\":\"uint256\"}],\"internalType\":\"structSession[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"limit\",\"type\":\"uint8\"}],\"name\":\"getSessionsByProvider\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"bidID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"closeoutReceipt\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerWithdrawnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"openedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endsAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"closedAt\",\"type\":\"uint256\"}],\"internalType\":\"structSession[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"limit\",\"type\":\"uint8\"}],\"name\":\"getSessionsByUser\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"bidID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"closeoutReceipt\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"providerWithdrawnAmount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"openedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"endsAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"closedAt\",\"type\":\"uint256\"}],\"internalType\":\"structSession[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"getTodaysBudget\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"signer\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"receipt\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"isValidReceipt\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_stake\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"providerApproval\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"openSession\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"sessionsCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"uint256\",\"name\":\"initialReward\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"rewardDecrease\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"payoutStart\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"decreaseInterval\",\"type\":\"uint128\"}],\"internalType\":\"structPool\",\"name\":\"pool\",\"type\":\"tuple\"}],\"name\":\"setPoolConfig\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"sessionStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"stakeToStipend\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"startOfTheDay\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"stipend\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"}],\"name\":\"stipendToStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"targetReward\",\"type\":\"uint256\"}],\"name\":\"whenComputeBalanceIsLessThan\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"sessionStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"}],\"name\":\"whenStipendLessThanDailyPrice\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amountToWithdraw\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"withdrawUserStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"userAddr\",\"type\":\"address\"}],\"name\":\"withdrawableUserStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"avail\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"hold\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
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

// ActiveSessionsCount is a free data retrieval call binding the contract method 0x782ea85c.
//
// Solidity: function activeSessionsCount() view returns(uint256)
func (_SessionRouter *SessionRouterCaller) ActiveSessionsCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "activeSessionsCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ActiveSessionsCount is a free data retrieval call binding the contract method 0x782ea85c.
//
// Solidity: function activeSessionsCount() view returns(uint256)
func (_SessionRouter *SessionRouterSession) ActiveSessionsCount() (*big.Int, error) {
	return _SessionRouter.Contract.ActiveSessionsCount(&_SessionRouter.CallOpts)
}

// ActiveSessionsCount is a free data retrieval call binding the contract method 0x782ea85c.
//
// Solidity: function activeSessionsCount() view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) ActiveSessionsCount() (*big.Int, error) {
	return _SessionRouter.Contract.ActiveSessionsCount(&_SessionRouter.CallOpts)
}

// GetActiveSessionsByProvider is a free data retrieval call binding the contract method 0xcba645ab.
//
// Solidity: function getActiveSessionsByProvider(address provider) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterCaller) GetActiveSessionsByProvider(opts *bind.CallOpts, provider common.Address) ([]Session, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getActiveSessionsByProvider", provider)

	if err != nil {
		return *new([]Session), err
	}

	out0 := *abi.ConvertType(out[0], new([]Session)).(*[]Session)

	return out0, err

}

// GetActiveSessionsByProvider is a free data retrieval call binding the contract method 0xcba645ab.
//
// Solidity: function getActiveSessionsByProvider(address provider) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterSession) GetActiveSessionsByProvider(provider common.Address) ([]Session, error) {
	return _SessionRouter.Contract.GetActiveSessionsByProvider(&_SessionRouter.CallOpts, provider)
}

// GetActiveSessionsByProvider is a free data retrieval call binding the contract method 0xcba645ab.
//
// Solidity: function getActiveSessionsByProvider(address provider) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterCallerSession) GetActiveSessionsByProvider(provider common.Address) ([]Session, error) {
	return _SessionRouter.Contract.GetActiveSessionsByProvider(&_SessionRouter.CallOpts, provider)
}

// GetActiveSessionsByUser is a free data retrieval call binding the contract method 0xb3da8c38.
//
// Solidity: function getActiveSessionsByUser(address user) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterCaller) GetActiveSessionsByUser(opts *bind.CallOpts, user common.Address) ([]Session, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getActiveSessionsByUser", user)

	if err != nil {
		return *new([]Session), err
	}

	out0 := *abi.ConvertType(out[0], new([]Session)).(*[]Session)

	return out0, err

}

// GetActiveSessionsByUser is a free data retrieval call binding the contract method 0xb3da8c38.
//
// Solidity: function getActiveSessionsByUser(address user) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterSession) GetActiveSessionsByUser(user common.Address) ([]Session, error) {
	return _SessionRouter.Contract.GetActiveSessionsByUser(&_SessionRouter.CallOpts, user)
}

// GetActiveSessionsByUser is a free data retrieval call binding the contract method 0xb3da8c38.
//
// Solidity: function getActiveSessionsByUser(address user) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterCallerSession) GetActiveSessionsByUser(user common.Address) ([]Session, error) {
	return _SessionRouter.Contract.GetActiveSessionsByUser(&_SessionRouter.CallOpts, user)
}

// GetComputeBalance is a free data retrieval call binding the contract method 0x76738e9e.
//
// Solidity: function getComputeBalance(uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) GetComputeBalance(opts *bind.CallOpts, timestamp *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getComputeBalance", timestamp)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetComputeBalance is a free data retrieval call binding the contract method 0x76738e9e.
//
// Solidity: function getComputeBalance(uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterSession) GetComputeBalance(timestamp *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.GetComputeBalance(&_SessionRouter.CallOpts, timestamp)
}

// GetComputeBalance is a free data retrieval call binding the contract method 0x76738e9e.
//
// Solidity: function getComputeBalance(uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) GetComputeBalance(timestamp *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.GetComputeBalance(&_SessionRouter.CallOpts, timestamp)
}

// GetProviderClaimableBalance is a free data retrieval call binding the contract method 0xa8ca6323.
//
// Solidity: function getProviderClaimableBalance(bytes32 sessionId) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) GetProviderClaimableBalance(opts *bind.CallOpts, sessionId [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getProviderClaimableBalance", sessionId)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProviderClaimableBalance is a free data retrieval call binding the contract method 0xa8ca6323.
//
// Solidity: function getProviderClaimableBalance(bytes32 sessionId) view returns(uint256)
func (_SessionRouter *SessionRouterSession) GetProviderClaimableBalance(sessionId [32]byte) (*big.Int, error) {
	return _SessionRouter.Contract.GetProviderClaimableBalance(&_SessionRouter.CallOpts, sessionId)
}

// GetProviderClaimableBalance is a free data retrieval call binding the contract method 0xa8ca6323.
//
// Solidity: function getProviderClaimableBalance(bytes32 sessionId) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) GetProviderClaimableBalance(sessionId [32]byte) (*big.Int, error) {
	return _SessionRouter.Contract.GetProviderClaimableBalance(&_SessionRouter.CallOpts, sessionId)
}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256))
func (_SessionRouter *SessionRouterCaller) GetSession(opts *bind.CallOpts, sessionId [32]byte) (Session, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getSession", sessionId)

	if err != nil {
		return *new(Session), err
	}

	out0 := *abi.ConvertType(out[0], new(Session)).(*Session)

	return out0, err

}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256))
func (_SessionRouter *SessionRouterSession) GetSession(sessionId [32]byte) (Session, error) {
	return _SessionRouter.Contract.GetSession(&_SessionRouter.CallOpts, sessionId)
}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256))
func (_SessionRouter *SessionRouterCallerSession) GetSession(sessionId [32]byte) (Session, error) {
	return _SessionRouter.Contract.GetSession(&_SessionRouter.CallOpts, sessionId)
}

// GetSessionsByModel is a free data retrieval call binding the contract method 0x67a057f6.
//
// Solidity: function getSessionsByModel(bytes32 modelId, uint256 offset, uint8 limit) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterCaller) GetSessionsByModel(opts *bind.CallOpts, modelId [32]byte, offset *big.Int, limit uint8) ([]Session, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getSessionsByModel", modelId, offset, limit)

	if err != nil {
		return *new([]Session), err
	}

	out0 := *abi.ConvertType(out[0], new([]Session)).(*[]Session)

	return out0, err

}

// GetSessionsByModel is a free data retrieval call binding the contract method 0x67a057f6.
//
// Solidity: function getSessionsByModel(bytes32 modelId, uint256 offset, uint8 limit) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterSession) GetSessionsByModel(modelId [32]byte, offset *big.Int, limit uint8) ([]Session, error) {
	return _SessionRouter.Contract.GetSessionsByModel(&_SessionRouter.CallOpts, modelId, offset, limit)
}

// GetSessionsByModel is a free data retrieval call binding the contract method 0x67a057f6.
//
// Solidity: function getSessionsByModel(bytes32 modelId, uint256 offset, uint8 limit) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterCallerSession) GetSessionsByModel(modelId [32]byte, offset *big.Int, limit uint8) ([]Session, error) {
	return _SessionRouter.Contract.GetSessionsByModel(&_SessionRouter.CallOpts, modelId, offset, limit)
}

// GetSessionsByProvider is a free data retrieval call binding the contract method 0x8ea1ac0e.
//
// Solidity: function getSessionsByProvider(address provider, uint256 offset, uint8 limit) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterCaller) GetSessionsByProvider(opts *bind.CallOpts, provider common.Address, offset *big.Int, limit uint8) ([]Session, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getSessionsByProvider", provider, offset, limit)

	if err != nil {
		return *new([]Session), err
	}

	out0 := *abi.ConvertType(out[0], new([]Session)).(*[]Session)

	return out0, err

}

// GetSessionsByProvider is a free data retrieval call binding the contract method 0x8ea1ac0e.
//
// Solidity: function getSessionsByProvider(address provider, uint256 offset, uint8 limit) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterSession) GetSessionsByProvider(provider common.Address, offset *big.Int, limit uint8) ([]Session, error) {
	return _SessionRouter.Contract.GetSessionsByProvider(&_SessionRouter.CallOpts, provider, offset, limit)
}

// GetSessionsByProvider is a free data retrieval call binding the contract method 0x8ea1ac0e.
//
// Solidity: function getSessionsByProvider(address provider, uint256 offset, uint8 limit) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterCallerSession) GetSessionsByProvider(provider common.Address, offset *big.Int, limit uint8) ([]Session, error) {
	return _SessionRouter.Contract.GetSessionsByProvider(&_SessionRouter.CallOpts, provider, offset, limit)
}

// GetSessionsByUser is a free data retrieval call binding the contract method 0xb954275b.
//
// Solidity: function getSessionsByUser(address user, uint256 offset, uint8 limit) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterCaller) GetSessionsByUser(opts *bind.CallOpts, user common.Address, offset *big.Int, limit uint8) ([]Session, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getSessionsByUser", user, offset, limit)

	if err != nil {
		return *new([]Session), err
	}

	out0 := *abi.ConvertType(out[0], new([]Session)).(*[]Session)

	return out0, err

}

// GetSessionsByUser is a free data retrieval call binding the contract method 0xb954275b.
//
// Solidity: function getSessionsByUser(address user, uint256 offset, uint8 limit) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterSession) GetSessionsByUser(user common.Address, offset *big.Int, limit uint8) ([]Session, error) {
	return _SessionRouter.Contract.GetSessionsByUser(&_SessionRouter.CallOpts, user, offset, limit)
}

// GetSessionsByUser is a free data retrieval call binding the contract method 0xb954275b.
//
// Solidity: function getSessionsByUser(address user, uint256 offset, uint8 limit) view returns((bytes32,address,address,bytes32,bytes32,uint256,uint256,bytes,uint256,uint256,uint256,uint256,uint256)[])
func (_SessionRouter *SessionRouterCallerSession) GetSessionsByUser(user common.Address, offset *big.Int, limit uint8) ([]Session, error) {
	return _SessionRouter.Contract.GetSessionsByUser(&_SessionRouter.CallOpts, user, offset, limit)
}

// GetTodaysBudget is a free data retrieval call binding the contract method 0x351ffeb0.
//
// Solidity: function getTodaysBudget(uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) GetTodaysBudget(opts *bind.CallOpts, timestamp *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getTodaysBudget", timestamp)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetTodaysBudget is a free data retrieval call binding the contract method 0x351ffeb0.
//
// Solidity: function getTodaysBudget(uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterSession) GetTodaysBudget(timestamp *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.GetTodaysBudget(&_SessionRouter.CallOpts, timestamp)
}

// GetTodaysBudget is a free data retrieval call binding the contract method 0x351ffeb0.
//
// Solidity: function getTodaysBudget(uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) GetTodaysBudget(timestamp *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.GetTodaysBudget(&_SessionRouter.CallOpts, timestamp)
}

// IsValidReceipt is a free data retrieval call binding the contract method 0x626dd729.
//
// Solidity: function isValidReceipt(address signer, bytes receipt, bytes signature) pure returns(bool)
func (_SessionRouter *SessionRouterCaller) IsValidReceipt(opts *bind.CallOpts, signer common.Address, receipt []byte, signature []byte) (bool, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "isValidReceipt", signer, receipt, signature)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsValidReceipt is a free data retrieval call binding the contract method 0x626dd729.
//
// Solidity: function isValidReceipt(address signer, bytes receipt, bytes signature) pure returns(bool)
func (_SessionRouter *SessionRouterSession) IsValidReceipt(signer common.Address, receipt []byte, signature []byte) (bool, error) {
	return _SessionRouter.Contract.IsValidReceipt(&_SessionRouter.CallOpts, signer, receipt, signature)
}

// IsValidReceipt is a free data retrieval call binding the contract method 0x626dd729.
//
// Solidity: function isValidReceipt(address signer, bytes receipt, bytes signature) pure returns(bool)
func (_SessionRouter *SessionRouterCallerSession) IsValidReceipt(signer common.Address, receipt []byte, signature []byte) (bool, error) {
	return _SessionRouter.Contract.IsValidReceipt(&_SessionRouter.CallOpts, signer, receipt, signature)
}

// SessionsCount is a free data retrieval call binding the contract method 0x312f6307.
//
// Solidity: function sessionsCount() view returns(uint256)
func (_SessionRouter *SessionRouterCaller) SessionsCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "sessionsCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// SessionsCount is a free data retrieval call binding the contract method 0x312f6307.
//
// Solidity: function sessionsCount() view returns(uint256)
func (_SessionRouter *SessionRouterSession) SessionsCount() (*big.Int, error) {
	return _SessionRouter.Contract.SessionsCount(&_SessionRouter.CallOpts)
}

// SessionsCount is a free data retrieval call binding the contract method 0x312f6307.
//
// Solidity: function sessionsCount() view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) SessionsCount() (*big.Int, error) {
	return _SessionRouter.Contract.SessionsCount(&_SessionRouter.CallOpts)
}

// StakeToStipend is a free data retrieval call binding the contract method 0x0a23b21f.
//
// Solidity: function stakeToStipend(uint256 sessionStake, uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) StakeToStipend(opts *bind.CallOpts, sessionStake *big.Int, timestamp *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "stakeToStipend", sessionStake, timestamp)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StakeToStipend is a free data retrieval call binding the contract method 0x0a23b21f.
//
// Solidity: function stakeToStipend(uint256 sessionStake, uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterSession) StakeToStipend(sessionStake *big.Int, timestamp *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.StakeToStipend(&_SessionRouter.CallOpts, sessionStake, timestamp)
}

// StakeToStipend is a free data retrieval call binding the contract method 0x0a23b21f.
//
// Solidity: function stakeToStipend(uint256 sessionStake, uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) StakeToStipend(sessionStake *big.Int, timestamp *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.StakeToStipend(&_SessionRouter.CallOpts, sessionStake, timestamp)
}

// StartOfTheDay is a free data retrieval call binding the contract method 0xeedd0a72.
//
// Solidity: function startOfTheDay(uint256 timestamp) pure returns(uint256)
func (_SessionRouter *SessionRouterCaller) StartOfTheDay(opts *bind.CallOpts, timestamp *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "startOfTheDay", timestamp)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StartOfTheDay is a free data retrieval call binding the contract method 0xeedd0a72.
//
// Solidity: function startOfTheDay(uint256 timestamp) pure returns(uint256)
func (_SessionRouter *SessionRouterSession) StartOfTheDay(timestamp *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.StartOfTheDay(&_SessionRouter.CallOpts, timestamp)
}

// StartOfTheDay is a free data retrieval call binding the contract method 0xeedd0a72.
//
// Solidity: function startOfTheDay(uint256 timestamp) pure returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) StartOfTheDay(timestamp *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.StartOfTheDay(&_SessionRouter.CallOpts, timestamp)
}

// StipendToStake is a free data retrieval call binding the contract method 0xac3c19ce.
//
// Solidity: function stipendToStake(uint256 stipend, uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) StipendToStake(opts *bind.CallOpts, stipend *big.Int, timestamp *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "stipendToStake", stipend, timestamp)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StipendToStake is a free data retrieval call binding the contract method 0xac3c19ce.
//
// Solidity: function stipendToStake(uint256 stipend, uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterSession) StipendToStake(stipend *big.Int, timestamp *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.StipendToStake(&_SessionRouter.CallOpts, stipend, timestamp)
}

// StipendToStake is a free data retrieval call binding the contract method 0xac3c19ce.
//
// Solidity: function stipendToStake(uint256 stipend, uint256 timestamp) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) StipendToStake(stipend *big.Int, timestamp *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.StipendToStake(&_SessionRouter.CallOpts, stipend, timestamp)
}

// WhenComputeBalanceIsLessThan is a free data retrieval call binding the contract method 0xfa64e1db.
//
// Solidity: function whenComputeBalanceIsLessThan(uint256 targetReward) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) WhenComputeBalanceIsLessThan(opts *bind.CallOpts, targetReward *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "whenComputeBalanceIsLessThan", targetReward)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WhenComputeBalanceIsLessThan is a free data retrieval call binding the contract method 0xfa64e1db.
//
// Solidity: function whenComputeBalanceIsLessThan(uint256 targetReward) view returns(uint256)
func (_SessionRouter *SessionRouterSession) WhenComputeBalanceIsLessThan(targetReward *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.WhenComputeBalanceIsLessThan(&_SessionRouter.CallOpts, targetReward)
}

// WhenComputeBalanceIsLessThan is a free data retrieval call binding the contract method 0xfa64e1db.
//
// Solidity: function whenComputeBalanceIsLessThan(uint256 targetReward) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) WhenComputeBalanceIsLessThan(targetReward *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.WhenComputeBalanceIsLessThan(&_SessionRouter.CallOpts, targetReward)
}

// WhenStipendLessThanDailyPrice is a free data retrieval call binding the contract method 0xf13619da.
//
// Solidity: function whenStipendLessThanDailyPrice(uint256 sessionStake, uint256 pricePerSecond) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) WhenStipendLessThanDailyPrice(opts *bind.CallOpts, sessionStake *big.Int, pricePerSecond *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "whenStipendLessThanDailyPrice", sessionStake, pricePerSecond)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// WhenStipendLessThanDailyPrice is a free data retrieval call binding the contract method 0xf13619da.
//
// Solidity: function whenStipendLessThanDailyPrice(uint256 sessionStake, uint256 pricePerSecond) view returns(uint256)
func (_SessionRouter *SessionRouterSession) WhenStipendLessThanDailyPrice(sessionStake *big.Int, pricePerSecond *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.WhenStipendLessThanDailyPrice(&_SessionRouter.CallOpts, sessionStake, pricePerSecond)
}

// WhenStipendLessThanDailyPrice is a free data retrieval call binding the contract method 0xf13619da.
//
// Solidity: function whenStipendLessThanDailyPrice(uint256 sessionStake, uint256 pricePerSecond) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) WhenStipendLessThanDailyPrice(sessionStake *big.Int, pricePerSecond *big.Int) (*big.Int, error) {
	return _SessionRouter.Contract.WhenStipendLessThanDailyPrice(&_SessionRouter.CallOpts, sessionStake, pricePerSecond)
}

// WithdrawableUserStake is a free data retrieval call binding the contract method 0x536f1f82.
//
// Solidity: function withdrawableUserStake(address userAddr) view returns(uint256 avail, uint256 hold)
func (_SessionRouter *SessionRouterCaller) WithdrawableUserStake(opts *bind.CallOpts, userAddr common.Address) (struct {
	Avail *big.Int
	Hold  *big.Int
}, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "withdrawableUserStake", userAddr)

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

// WithdrawableUserStake is a free data retrieval call binding the contract method 0x536f1f82.
//
// Solidity: function withdrawableUserStake(address userAddr) view returns(uint256 avail, uint256 hold)
func (_SessionRouter *SessionRouterSession) WithdrawableUserStake(userAddr common.Address) (struct {
	Avail *big.Int
	Hold  *big.Int
}, error) {
	return _SessionRouter.Contract.WithdrawableUserStake(&_SessionRouter.CallOpts, userAddr)
}

// WithdrawableUserStake is a free data retrieval call binding the contract method 0x536f1f82.
//
// Solidity: function withdrawableUserStake(address userAddr) view returns(uint256 avail, uint256 hold)
func (_SessionRouter *SessionRouterCallerSession) WithdrawableUserStake(userAddr common.Address) (struct {
	Avail *big.Int
	Hold  *big.Int
}, error) {
	return _SessionRouter.Contract.WithdrawableUserStake(&_SessionRouter.CallOpts, userAddr)
}

// ClaimProviderBalance is a paid mutator transaction binding the contract method 0xbab3de02.
//
// Solidity: function claimProviderBalance(bytes32 sessionId, uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterTransactor) ClaimProviderBalance(opts *bind.TransactOpts, sessionId [32]byte, amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "claimProviderBalance", sessionId, amountToWithdraw, to)
}

// ClaimProviderBalance is a paid mutator transaction binding the contract method 0xbab3de02.
//
// Solidity: function claimProviderBalance(bytes32 sessionId, uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterSession) ClaimProviderBalance(sessionId [32]byte, amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.Contract.ClaimProviderBalance(&_SessionRouter.TransactOpts, sessionId, amountToWithdraw, to)
}

// ClaimProviderBalance is a paid mutator transaction binding the contract method 0xbab3de02.
//
// Solidity: function claimProviderBalance(bytes32 sessionId, uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterTransactorSession) ClaimProviderBalance(sessionId [32]byte, amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.Contract.ClaimProviderBalance(&_SessionRouter.TransactOpts, sessionId, amountToWithdraw, to)
}

// CloseSession is a paid mutator transaction binding the contract method 0x42f77a31.
//
// Solidity: function closeSession(bytes receiptEncoded, bytes signature) returns()
func (_SessionRouter *SessionRouterTransactor) CloseSession(opts *bind.TransactOpts, receiptEncoded []byte, signature []byte) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "closeSession", receiptEncoded, signature)
}

// CloseSession is a paid mutator transaction binding the contract method 0x42f77a31.
//
// Solidity: function closeSession(bytes receiptEncoded, bytes signature) returns()
func (_SessionRouter *SessionRouterSession) CloseSession(receiptEncoded []byte, signature []byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.CloseSession(&_SessionRouter.TransactOpts, receiptEncoded, signature)
}

// CloseSession is a paid mutator transaction binding the contract method 0x42f77a31.
//
// Solidity: function closeSession(bytes receiptEncoded, bytes signature) returns()
func (_SessionRouter *SessionRouterTransactorSession) CloseSession(receiptEncoded []byte, signature []byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.CloseSession(&_SessionRouter.TransactOpts, receiptEncoded, signature)
}

// DeleteHistory is a paid mutator transaction binding the contract method 0xf074ca6b.
//
// Solidity: function deleteHistory(bytes32 sessionId) returns()
func (_SessionRouter *SessionRouterTransactor) DeleteHistory(opts *bind.TransactOpts, sessionId [32]byte) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "deleteHistory", sessionId)
}

// DeleteHistory is a paid mutator transaction binding the contract method 0xf074ca6b.
//
// Solidity: function deleteHistory(bytes32 sessionId) returns()
func (_SessionRouter *SessionRouterSession) DeleteHistory(sessionId [32]byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.DeleteHistory(&_SessionRouter.TransactOpts, sessionId)
}

// DeleteHistory is a paid mutator transaction binding the contract method 0xf074ca6b.
//
// Solidity: function deleteHistory(bytes32 sessionId) returns()
func (_SessionRouter *SessionRouterTransactorSession) DeleteHistory(sessionId [32]byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.DeleteHistory(&_SessionRouter.TransactOpts, sessionId)
}

// OpenSession is a paid mutator transaction binding the contract method 0x1f71815e.
//
// Solidity: function openSession(uint256 _stake, bytes providerApproval, bytes signature) returns(bytes32 sessionId)
func (_SessionRouter *SessionRouterTransactor) OpenSession(opts *bind.TransactOpts, _stake *big.Int, providerApproval []byte, signature []byte) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "openSession", _stake, providerApproval, signature)
}

// OpenSession is a paid mutator transaction binding the contract method 0x1f71815e.
//
// Solidity: function openSession(uint256 _stake, bytes providerApproval, bytes signature) returns(bytes32 sessionId)
func (_SessionRouter *SessionRouterSession) OpenSession(_stake *big.Int, providerApproval []byte, signature []byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.OpenSession(&_SessionRouter.TransactOpts, _stake, providerApproval, signature)
}

// OpenSession is a paid mutator transaction binding the contract method 0x1f71815e.
//
// Solidity: function openSession(uint256 _stake, bytes providerApproval, bytes signature) returns(bytes32 sessionId)
func (_SessionRouter *SessionRouterTransactorSession) OpenSession(_stake *big.Int, providerApproval []byte, signature []byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.OpenSession(&_SessionRouter.TransactOpts, _stake, providerApproval, signature)
}

// SetPoolConfig is a paid mutator transaction binding the contract method 0x8b1af52a.
//
// Solidity: function setPoolConfig((uint256,uint256,uint128,uint128) pool) returns()
func (_SessionRouter *SessionRouterTransactor) SetPoolConfig(opts *bind.TransactOpts, pool Pool) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "setPoolConfig", pool)
}

// SetPoolConfig is a paid mutator transaction binding the contract method 0x8b1af52a.
//
// Solidity: function setPoolConfig((uint256,uint256,uint128,uint128) pool) returns()
func (_SessionRouter *SessionRouterSession) SetPoolConfig(pool Pool) (*types.Transaction, error) {
	return _SessionRouter.Contract.SetPoolConfig(&_SessionRouter.TransactOpts, pool)
}

// SetPoolConfig is a paid mutator transaction binding the contract method 0x8b1af52a.
//
// Solidity: function setPoolConfig((uint256,uint256,uint128,uint128) pool) returns()
func (_SessionRouter *SessionRouterTransactorSession) SetPoolConfig(pool Pool) (*types.Transaction, error) {
	return _SessionRouter.Contract.SetPoolConfig(&_SessionRouter.TransactOpts, pool)
}

// WithdrawUserStake is a paid mutator transaction binding the contract method 0xcd308cb1.
//
// Solidity: function withdrawUserStake(uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterTransactor) WithdrawUserStake(opts *bind.TransactOpts, amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "withdrawUserStake", amountToWithdraw, to)
}

// WithdrawUserStake is a paid mutator transaction binding the contract method 0xcd308cb1.
//
// Solidity: function withdrawUserStake(uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterSession) WithdrawUserStake(amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.Contract.WithdrawUserStake(&_SessionRouter.TransactOpts, amountToWithdraw, to)
}

// WithdrawUserStake is a paid mutator transaction binding the contract method 0xcd308cb1.
//
// Solidity: function withdrawUserStake(uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterTransactorSession) WithdrawUserStake(amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.Contract.WithdrawUserStake(&_SessionRouter.TransactOpts, amountToWithdraw, to)
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
	UserAddress common.Address
	SessionId   [32]byte
	ProviderId  common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterSessionClosed is a free log retrieval operation binding the contract event 0x337fbb0a41a596db800dc836595a57815f967185e3596615c646f2455ac3914a.
//
// Solidity: event SessionClosed(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) FilterSessionClosed(opts *bind.FilterOpts, userAddress []common.Address, sessionId [][32]byte, providerId []common.Address) (*SessionRouterSessionClosedIterator, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}
	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "SessionClosed", userAddressRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterSessionClosedIterator{contract: _SessionRouter.contract, event: "SessionClosed", logs: logs, sub: sub}, nil
}

// WatchSessionClosed is a free log subscription operation binding the contract event 0x337fbb0a41a596db800dc836595a57815f967185e3596615c646f2455ac3914a.
//
// Solidity: event SessionClosed(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) WatchSessionClosed(opts *bind.WatchOpts, sink chan<- *SessionRouterSessionClosed, userAddress []common.Address, sessionId [][32]byte, providerId []common.Address) (event.Subscription, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}
	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "SessionClosed", userAddressRule, sessionIdRule, providerIdRule)
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
// Solidity: event SessionClosed(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId)
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
	UserAddress common.Address
	SessionId   [32]byte
	ProviderId  common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterSessionOpened is a free log retrieval operation binding the contract event 0x2bd7c890baf595977d256a6e784512c873ac58ba612b4895dbb7f784bfbf4839.
//
// Solidity: event SessionOpened(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) FilterSessionOpened(opts *bind.FilterOpts, userAddress []common.Address, sessionId [][32]byte, providerId []common.Address) (*SessionRouterSessionOpenedIterator, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}
	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "SessionOpened", userAddressRule, sessionIdRule, providerIdRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterSessionOpenedIterator{contract: _SessionRouter.contract, event: "SessionOpened", logs: logs, sub: sub}, nil
}

// WatchSessionOpened is a free log subscription operation binding the contract event 0x2bd7c890baf595977d256a6e784512c873ac58ba612b4895dbb7f784bfbf4839.
//
// Solidity: event SessionOpened(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) WatchSessionOpened(opts *bind.WatchOpts, sink chan<- *SessionRouterSessionOpened, userAddress []common.Address, sessionId [][32]byte, providerId []common.Address) (event.Subscription, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}
	var sessionIdRule []interface{}
	for _, sessionIdItem := range sessionId {
		sessionIdRule = append(sessionIdRule, sessionIdItem)
	}
	var providerIdRule []interface{}
	for _, providerIdItem := range providerId {
		providerIdRule = append(providerIdRule, providerIdItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "SessionOpened", userAddressRule, sessionIdRule, providerIdRule)
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
// Solidity: event SessionOpened(address indexed userAddress, bytes32 indexed sessionId, address indexed providerId)
func (_SessionRouter *SessionRouterFilterer) ParseSessionOpened(log types.Log) (*SessionRouterSessionOpened, error) {
	event := new(SessionRouterSessionOpened)
	if err := _SessionRouter.contract.UnpackLog(event, "SessionOpened", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
