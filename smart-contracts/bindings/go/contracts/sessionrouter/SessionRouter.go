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

// SessionRouterSession is an auto generated low-level Go binding around an user-defined struct.
type SessionRouterSession struct {
	Id              [32]byte
	User            common.Address
	Provider        common.Address
	ModelAgentId    [32]byte
	Budget          *big.Int
	Price           *big.Int
	CloseoutReceipt []byte
	CloseoutType    *big.Int
	OpenedAt        *big.Int
	ClosedAt        *big.Int
}

// SessionRouterMetaData contains all meta data concerning the SessionRouter contract.
var SessionRouterMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"BidNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"InvalidSignature\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotEnoughBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotEnoughStipend\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSenderOrOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotUser\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotUserOrProvider\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"SessionTooShort\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"ProviderClaimed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerId\",\"type\":\"address\"}],\"name\":\"SessionClosed\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"providerId\",\"type\":\"address\"}],\"name\":\"SessionOpened\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Staked\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"userAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Unstaked\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"amountToWithdraw\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"}],\"name\":\"claimProviderBalance\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"receiptEncoded\",\"type\":\"bytes\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"name\":\"closeSession\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"}],\"name\":\"deleteHistory\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_encodedMessage\",\"type\":\"bytes\"}],\"name\":\"getEthSignedMessageHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"providerAddr\",\"type\":\"address\"}],\"name\":\"getProviderBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"total\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"hold\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"providerAddr\",\"type\":\"address\"}],\"name\":\"getProviderClaimBalance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"}],\"name\":\"getSession\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"budget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"closeoutReceipt\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"openedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"closedAt\",\"type\":\"uint256\"}],\"internalType\":\"structSessionRouter.Session\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_stakingDailyStipend\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_marketplace\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"map\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"marketplace\",\"outputs\":[{\"internalType\":\"contractMarketplace\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"budget\",\"type\":\"uint256\"}],\"name\":\"openSession\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"sessionId\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"providerOnHold\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"releaseAt\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"message\",\"type\":\"bytes32\"},{\"internalType\":\"bytes\",\"name\":\"sig\",\"type\":\"bytes\"}],\"name\":\"recoverSigner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"sessions\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"},{\"internalType\":\"address\",\"name\":\"user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"budget\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"price\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"closeoutReceipt\",\"type\":\"bytes\"},{\"internalType\":\"uint256\",\"name\":\"closeoutType\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"openedAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"closedAt\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"int256\",\"name\":\"delay\",\"type\":\"int256\"}],\"name\":\"setStakeDelay\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"sig\",\"type\":\"bytes\"}],\"name\":\"splitSignature\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"v\",\"type\":\"uint8\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stakeDelay\",\"outputs\":[{\"internalType\":\"int256\",\"name\":\"\",\"type\":\"int256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"stakingDailyStipend\",\"outputs\":[{\"internalType\":\"contractStakingDailyStipend\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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

// GetEthSignedMessageHash is a free data retrieval call binding the contract method 0xf9aea466.
//
// Solidity: function getEthSignedMessageHash(bytes _encodedMessage) pure returns(bytes32)
func (_SessionRouter *SessionRouterCaller) GetEthSignedMessageHash(opts *bind.CallOpts, _encodedMessage []byte) ([32]byte, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getEthSignedMessageHash", _encodedMessage)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetEthSignedMessageHash is a free data retrieval call binding the contract method 0xf9aea466.
//
// Solidity: function getEthSignedMessageHash(bytes _encodedMessage) pure returns(bytes32)
func (_SessionRouter *SessionRouterSession) GetEthSignedMessageHash(_encodedMessage []byte) ([32]byte, error) {
	return _SessionRouter.Contract.GetEthSignedMessageHash(&_SessionRouter.CallOpts, _encodedMessage)
}

// GetEthSignedMessageHash is a free data retrieval call binding the contract method 0xf9aea466.
//
// Solidity: function getEthSignedMessageHash(bytes _encodedMessage) pure returns(bytes32)
func (_SessionRouter *SessionRouterCallerSession) GetEthSignedMessageHash(_encodedMessage []byte) ([32]byte, error) {
	return _SessionRouter.Contract.GetEthSignedMessageHash(&_SessionRouter.CallOpts, _encodedMessage)
}

// GetProviderBalance is a free data retrieval call binding the contract method 0x832eea0c.
//
// Solidity: function getProviderBalance(address providerAddr) view returns(uint256 total, uint256 hold)
func (_SessionRouter *SessionRouterCaller) GetProviderBalance(opts *bind.CallOpts, providerAddr common.Address) (struct {
	Total *big.Int
	Hold  *big.Int
}, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getProviderBalance", providerAddr)

	outstruct := new(struct {
		Total *big.Int
		Hold  *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Total = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.Hold = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetProviderBalance is a free data retrieval call binding the contract method 0x832eea0c.
//
// Solidity: function getProviderBalance(address providerAddr) view returns(uint256 total, uint256 hold)
func (_SessionRouter *SessionRouterSession) GetProviderBalance(providerAddr common.Address) (struct {
	Total *big.Int
	Hold  *big.Int
}, error) {
	return _SessionRouter.Contract.GetProviderBalance(&_SessionRouter.CallOpts, providerAddr)
}

// GetProviderBalance is a free data retrieval call binding the contract method 0x832eea0c.
//
// Solidity: function getProviderBalance(address providerAddr) view returns(uint256 total, uint256 hold)
func (_SessionRouter *SessionRouterCallerSession) GetProviderBalance(providerAddr common.Address) (struct {
	Total *big.Int
	Hold  *big.Int
}, error) {
	return _SessionRouter.Contract.GetProviderBalance(&_SessionRouter.CallOpts, providerAddr)
}

// GetProviderClaimBalance is a free data retrieval call binding the contract method 0x6c98599b.
//
// Solidity: function getProviderClaimBalance(address providerAddr) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) GetProviderClaimBalance(opts *bind.CallOpts, providerAddr common.Address) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getProviderClaimBalance", providerAddr)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetProviderClaimBalance is a free data retrieval call binding the contract method 0x6c98599b.
//
// Solidity: function getProviderClaimBalance(address providerAddr) view returns(uint256)
func (_SessionRouter *SessionRouterSession) GetProviderClaimBalance(providerAddr common.Address) (*big.Int, error) {
	return _SessionRouter.Contract.GetProviderClaimBalance(&_SessionRouter.CallOpts, providerAddr)
}

// GetProviderClaimBalance is a free data retrieval call binding the contract method 0x6c98599b.
//
// Solidity: function getProviderClaimBalance(address providerAddr) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) GetProviderClaimBalance(providerAddr common.Address) (*big.Int, error) {
	return _SessionRouter.Contract.GetProviderClaimBalance(&_SessionRouter.CallOpts, providerAddr)
}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId) view returns((bytes32,address,address,bytes32,uint256,uint256,bytes,uint256,uint256,uint256))
func (_SessionRouter *SessionRouterCaller) GetSession(opts *bind.CallOpts, sessionId [32]byte) (SessionRouterSession, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "getSession", sessionId)

	if err != nil {
		return *new(SessionRouterSession), err
	}

	out0 := *abi.ConvertType(out[0], new(SessionRouterSession)).(*SessionRouterSession)

	return out0, err

}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId) view returns((bytes32,address,address,bytes32,uint256,uint256,bytes,uint256,uint256,uint256))
func (_SessionRouter *SessionRouterSession) GetSession(sessionId [32]byte) (SessionRouterSession, error) {
	return _SessionRouter.Contract.GetSession(&_SessionRouter.CallOpts, sessionId)
}

// GetSession is a free data retrieval call binding the contract method 0x39b240bd.
//
// Solidity: function getSession(bytes32 sessionId) view returns((bytes32,address,address,bytes32,uint256,uint256,bytes,uint256,uint256,uint256))
func (_SessionRouter *SessionRouterCallerSession) GetSession(sessionId [32]byte) (SessionRouterSession, error) {
	return _SessionRouter.Contract.GetSession(&_SessionRouter.CallOpts, sessionId)
}

// Map is a free data retrieval call binding the contract method 0x0ae186a8.
//
// Solidity: function map(bytes32 ) view returns(uint256)
func (_SessionRouter *SessionRouterCaller) Map(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "map", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Map is a free data retrieval call binding the contract method 0x0ae186a8.
//
// Solidity: function map(bytes32 ) view returns(uint256)
func (_SessionRouter *SessionRouterSession) Map(arg0 [32]byte) (*big.Int, error) {
	return _SessionRouter.Contract.Map(&_SessionRouter.CallOpts, arg0)
}

// Map is a free data retrieval call binding the contract method 0x0ae186a8.
//
// Solidity: function map(bytes32 ) view returns(uint256)
func (_SessionRouter *SessionRouterCallerSession) Map(arg0 [32]byte) (*big.Int, error) {
	return _SessionRouter.Contract.Map(&_SessionRouter.CallOpts, arg0)
}

// Marketplace is a free data retrieval call binding the contract method 0xabc8c7af.
//
// Solidity: function marketplace() view returns(address)
func (_SessionRouter *SessionRouterCaller) Marketplace(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "marketplace")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Marketplace is a free data retrieval call binding the contract method 0xabc8c7af.
//
// Solidity: function marketplace() view returns(address)
func (_SessionRouter *SessionRouterSession) Marketplace() (common.Address, error) {
	return _SessionRouter.Contract.Marketplace(&_SessionRouter.CallOpts)
}

// Marketplace is a free data retrieval call binding the contract method 0xabc8c7af.
//
// Solidity: function marketplace() view returns(address)
func (_SessionRouter *SessionRouterCallerSession) Marketplace() (common.Address, error) {
	return _SessionRouter.Contract.Marketplace(&_SessionRouter.CallOpts)
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

// ProviderOnHold is a free data retrieval call binding the contract method 0xdbf7d54f.
//
// Solidity: function providerOnHold(address , uint256 ) view returns(uint256 amount, uint256 releaseAt)
func (_SessionRouter *SessionRouterCaller) ProviderOnHold(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (struct {
	Amount    *big.Int
	ReleaseAt *big.Int
}, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "providerOnHold", arg0, arg1)

	outstruct := new(struct {
		Amount    *big.Int
		ReleaseAt *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Amount = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.ReleaseAt = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// ProviderOnHold is a free data retrieval call binding the contract method 0xdbf7d54f.
//
// Solidity: function providerOnHold(address , uint256 ) view returns(uint256 amount, uint256 releaseAt)
func (_SessionRouter *SessionRouterSession) ProviderOnHold(arg0 common.Address, arg1 *big.Int) (struct {
	Amount    *big.Int
	ReleaseAt *big.Int
}, error) {
	return _SessionRouter.Contract.ProviderOnHold(&_SessionRouter.CallOpts, arg0, arg1)
}

// ProviderOnHold is a free data retrieval call binding the contract method 0xdbf7d54f.
//
// Solidity: function providerOnHold(address , uint256 ) view returns(uint256 amount, uint256 releaseAt)
func (_SessionRouter *SessionRouterCallerSession) ProviderOnHold(arg0 common.Address, arg1 *big.Int) (struct {
	Amount    *big.Int
	ReleaseAt *big.Int
}, error) {
	return _SessionRouter.Contract.ProviderOnHold(&_SessionRouter.CallOpts, arg0, arg1)
}

// RecoverSigner is a free data retrieval call binding the contract method 0x97aba7f9.
//
// Solidity: function recoverSigner(bytes32 message, bytes sig) pure returns(address)
func (_SessionRouter *SessionRouterCaller) RecoverSigner(opts *bind.CallOpts, message [32]byte, sig []byte) (common.Address, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "recoverSigner", message, sig)

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RecoverSigner is a free data retrieval call binding the contract method 0x97aba7f9.
//
// Solidity: function recoverSigner(bytes32 message, bytes sig) pure returns(address)
func (_SessionRouter *SessionRouterSession) RecoverSigner(message [32]byte, sig []byte) (common.Address, error) {
	return _SessionRouter.Contract.RecoverSigner(&_SessionRouter.CallOpts, message, sig)
}

// RecoverSigner is a free data retrieval call binding the contract method 0x97aba7f9.
//
// Solidity: function recoverSigner(bytes32 message, bytes sig) pure returns(address)
func (_SessionRouter *SessionRouterCallerSession) RecoverSigner(message [32]byte, sig []byte) (common.Address, error) {
	return _SessionRouter.Contract.RecoverSigner(&_SessionRouter.CallOpts, message, sig)
}

// Sessions is a free data retrieval call binding the contract method 0x83c4b7a3.
//
// Solidity: function sessions(uint256 ) view returns(bytes32 id, address user, address provider, bytes32 modelAgentId, uint256 budget, uint256 price, bytes closeoutReceipt, uint256 closeoutType, uint256 openedAt, uint256 closedAt)
func (_SessionRouter *SessionRouterCaller) Sessions(opts *bind.CallOpts, arg0 *big.Int) (struct {
	Id              [32]byte
	User            common.Address
	Provider        common.Address
	ModelAgentId    [32]byte
	Budget          *big.Int
	Price           *big.Int
	CloseoutReceipt []byte
	CloseoutType    *big.Int
	OpenedAt        *big.Int
	ClosedAt        *big.Int
}, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "sessions", arg0)

	outstruct := new(struct {
		Id              [32]byte
		User            common.Address
		Provider        common.Address
		ModelAgentId    [32]byte
		Budget          *big.Int
		Price           *big.Int
		CloseoutReceipt []byte
		CloseoutType    *big.Int
		OpenedAt        *big.Int
		ClosedAt        *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Id = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.User = *abi.ConvertType(out[1], new(common.Address)).(*common.Address)
	outstruct.Provider = *abi.ConvertType(out[2], new(common.Address)).(*common.Address)
	outstruct.ModelAgentId = *abi.ConvertType(out[3], new([32]byte)).(*[32]byte)
	outstruct.Budget = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.Price = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.CloseoutReceipt = *abi.ConvertType(out[6], new([]byte)).(*[]byte)
	outstruct.CloseoutType = *abi.ConvertType(out[7], new(*big.Int)).(**big.Int)
	outstruct.OpenedAt = *abi.ConvertType(out[8], new(*big.Int)).(**big.Int)
	outstruct.ClosedAt = *abi.ConvertType(out[9], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Sessions is a free data retrieval call binding the contract method 0x83c4b7a3.
//
// Solidity: function sessions(uint256 ) view returns(bytes32 id, address user, address provider, bytes32 modelAgentId, uint256 budget, uint256 price, bytes closeoutReceipt, uint256 closeoutType, uint256 openedAt, uint256 closedAt)
func (_SessionRouter *SessionRouterSession) Sessions(arg0 *big.Int) (struct {
	Id              [32]byte
	User            common.Address
	Provider        common.Address
	ModelAgentId    [32]byte
	Budget          *big.Int
	Price           *big.Int
	CloseoutReceipt []byte
	CloseoutType    *big.Int
	OpenedAt        *big.Int
	ClosedAt        *big.Int
}, error) {
	return _SessionRouter.Contract.Sessions(&_SessionRouter.CallOpts, arg0)
}

// Sessions is a free data retrieval call binding the contract method 0x83c4b7a3.
//
// Solidity: function sessions(uint256 ) view returns(bytes32 id, address user, address provider, bytes32 modelAgentId, uint256 budget, uint256 price, bytes closeoutReceipt, uint256 closeoutType, uint256 openedAt, uint256 closedAt)
func (_SessionRouter *SessionRouterCallerSession) Sessions(arg0 *big.Int) (struct {
	Id              [32]byte
	User            common.Address
	Provider        common.Address
	ModelAgentId    [32]byte
	Budget          *big.Int
	Price           *big.Int
	CloseoutReceipt []byte
	CloseoutType    *big.Int
	OpenedAt        *big.Int
	ClosedAt        *big.Int
}, error) {
	return _SessionRouter.Contract.Sessions(&_SessionRouter.CallOpts, arg0)
}

// SplitSignature is a free data retrieval call binding the contract method 0xa7bb5803.
//
// Solidity: function splitSignature(bytes sig) pure returns(bytes32 r, bytes32 s, uint8 v)
func (_SessionRouter *SessionRouterCaller) SplitSignature(opts *bind.CallOpts, sig []byte) (struct {
	R [32]byte
	S [32]byte
	V uint8
}, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "splitSignature", sig)

	outstruct := new(struct {
		R [32]byte
		S [32]byte
		V uint8
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.R = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.S = *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)
	outstruct.V = *abi.ConvertType(out[2], new(uint8)).(*uint8)

	return *outstruct, err

}

// SplitSignature is a free data retrieval call binding the contract method 0xa7bb5803.
//
// Solidity: function splitSignature(bytes sig) pure returns(bytes32 r, bytes32 s, uint8 v)
func (_SessionRouter *SessionRouterSession) SplitSignature(sig []byte) (struct {
	R [32]byte
	S [32]byte
	V uint8
}, error) {
	return _SessionRouter.Contract.SplitSignature(&_SessionRouter.CallOpts, sig)
}

// SplitSignature is a free data retrieval call binding the contract method 0xa7bb5803.
//
// Solidity: function splitSignature(bytes sig) pure returns(bytes32 r, bytes32 s, uint8 v)
func (_SessionRouter *SessionRouterCallerSession) SplitSignature(sig []byte) (struct {
	R [32]byte
	S [32]byte
	V uint8
}, error) {
	return _SessionRouter.Contract.SplitSignature(&_SessionRouter.CallOpts, sig)
}

// StakeDelay is a free data retrieval call binding the contract method 0x946ada60.
//
// Solidity: function stakeDelay() view returns(int256)
func (_SessionRouter *SessionRouterCaller) StakeDelay(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "stakeDelay")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// StakeDelay is a free data retrieval call binding the contract method 0x946ada60.
//
// Solidity: function stakeDelay() view returns(int256)
func (_SessionRouter *SessionRouterSession) StakeDelay() (*big.Int, error) {
	return _SessionRouter.Contract.StakeDelay(&_SessionRouter.CallOpts)
}

// StakeDelay is a free data retrieval call binding the contract method 0x946ada60.
//
// Solidity: function stakeDelay() view returns(int256)
func (_SessionRouter *SessionRouterCallerSession) StakeDelay() (*big.Int, error) {
	return _SessionRouter.Contract.StakeDelay(&_SessionRouter.CallOpts)
}

// StakingDailyStipend is a free data retrieval call binding the contract method 0xfcfc2201.
//
// Solidity: function stakingDailyStipend() view returns(address)
func (_SessionRouter *SessionRouterCaller) StakingDailyStipend(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "stakingDailyStipend")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// StakingDailyStipend is a free data retrieval call binding the contract method 0xfcfc2201.
//
// Solidity: function stakingDailyStipend() view returns(address)
func (_SessionRouter *SessionRouterSession) StakingDailyStipend() (common.Address, error) {
	return _SessionRouter.Contract.StakingDailyStipend(&_SessionRouter.CallOpts)
}

// StakingDailyStipend is a free data retrieval call binding the contract method 0xfcfc2201.
//
// Solidity: function stakingDailyStipend() view returns(address)
func (_SessionRouter *SessionRouterCallerSession) StakingDailyStipend() (common.Address, error) {
	return _SessionRouter.Contract.StakingDailyStipend(&_SessionRouter.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_SessionRouter *SessionRouterCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SessionRouter.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_SessionRouter *SessionRouterSession) Token() (common.Address, error) {
	return _SessionRouter.Contract.Token(&_SessionRouter.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_SessionRouter *SessionRouterCallerSession) Token() (common.Address, error) {
	return _SessionRouter.Contract.Token(&_SessionRouter.CallOpts)
}

// ClaimProviderBalance is a paid mutator transaction binding the contract method 0xc9a93c1a.
//
// Solidity: function claimProviderBalance(uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterTransactor) ClaimProviderBalance(opts *bind.TransactOpts, amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "claimProviderBalance", amountToWithdraw, to)
}

// ClaimProviderBalance is a paid mutator transaction binding the contract method 0xc9a93c1a.
//
// Solidity: function claimProviderBalance(uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterSession) ClaimProviderBalance(amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.Contract.ClaimProviderBalance(&_SessionRouter.TransactOpts, amountToWithdraw, to)
}

// ClaimProviderBalance is a paid mutator transaction binding the contract method 0xc9a93c1a.
//
// Solidity: function claimProviderBalance(uint256 amountToWithdraw, address to) returns()
func (_SessionRouter *SessionRouterTransactorSession) ClaimProviderBalance(amountToWithdraw *big.Int, to common.Address) (*types.Transaction, error) {
	return _SessionRouter.Contract.ClaimProviderBalance(&_SessionRouter.TransactOpts, amountToWithdraw, to)
}

// CloseSession is a paid mutator transaction binding the contract method 0x9775d1ff.
//
// Solidity: function closeSession(bytes32 sessionId, bytes receiptEncoded, bytes signature) returns()
func (_SessionRouter *SessionRouterTransactor) CloseSession(opts *bind.TransactOpts, sessionId [32]byte, receiptEncoded []byte, signature []byte) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "closeSession", sessionId, receiptEncoded, signature)
}

// CloseSession is a paid mutator transaction binding the contract method 0x9775d1ff.
//
// Solidity: function closeSession(bytes32 sessionId, bytes receiptEncoded, bytes signature) returns()
func (_SessionRouter *SessionRouterSession) CloseSession(sessionId [32]byte, receiptEncoded []byte, signature []byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.CloseSession(&_SessionRouter.TransactOpts, sessionId, receiptEncoded, signature)
}

// CloseSession is a paid mutator transaction binding the contract method 0x9775d1ff.
//
// Solidity: function closeSession(bytes32 sessionId, bytes receiptEncoded, bytes signature) returns()
func (_SessionRouter *SessionRouterTransactorSession) CloseSession(sessionId [32]byte, receiptEncoded []byte, signature []byte) (*types.Transaction, error) {
	return _SessionRouter.Contract.CloseSession(&_SessionRouter.TransactOpts, sessionId, receiptEncoded, signature)
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

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _token, address _stakingDailyStipend, address _marketplace) returns()
func (_SessionRouter *SessionRouterTransactor) Initialize(opts *bind.TransactOpts, _token common.Address, _stakingDailyStipend common.Address, _marketplace common.Address) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "initialize", _token, _stakingDailyStipend, _marketplace)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _token, address _stakingDailyStipend, address _marketplace) returns()
func (_SessionRouter *SessionRouterSession) Initialize(_token common.Address, _stakingDailyStipend common.Address, _marketplace common.Address) (*types.Transaction, error) {
	return _SessionRouter.Contract.Initialize(&_SessionRouter.TransactOpts, _token, _stakingDailyStipend, _marketplace)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _token, address _stakingDailyStipend, address _marketplace) returns()
func (_SessionRouter *SessionRouterTransactorSession) Initialize(_token common.Address, _stakingDailyStipend common.Address, _marketplace common.Address) (*types.Transaction, error) {
	return _SessionRouter.Contract.Initialize(&_SessionRouter.TransactOpts, _token, _stakingDailyStipend, _marketplace)
}

// OpenSession is a paid mutator transaction binding the contract method 0x48c00c90.
//
// Solidity: function openSession(bytes32 bidId, uint256 budget) returns(bytes32 sessionId)
func (_SessionRouter *SessionRouterTransactor) OpenSession(opts *bind.TransactOpts, bidId [32]byte, budget *big.Int) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "openSession", bidId, budget)
}

// OpenSession is a paid mutator transaction binding the contract method 0x48c00c90.
//
// Solidity: function openSession(bytes32 bidId, uint256 budget) returns(bytes32 sessionId)
func (_SessionRouter *SessionRouterSession) OpenSession(bidId [32]byte, budget *big.Int) (*types.Transaction, error) {
	return _SessionRouter.Contract.OpenSession(&_SessionRouter.TransactOpts, bidId, budget)
}

// OpenSession is a paid mutator transaction binding the contract method 0x48c00c90.
//
// Solidity: function openSession(bytes32 bidId, uint256 budget) returns(bytes32 sessionId)
func (_SessionRouter *SessionRouterTransactorSession) OpenSession(bidId [32]byte, budget *big.Int) (*types.Transaction, error) {
	return _SessionRouter.Contract.OpenSession(&_SessionRouter.TransactOpts, bidId, budget)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SessionRouter *SessionRouterTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SessionRouter *SessionRouterSession) RenounceOwnership() (*types.Transaction, error) {
	return _SessionRouter.Contract.RenounceOwnership(&_SessionRouter.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SessionRouter *SessionRouterTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _SessionRouter.Contract.RenounceOwnership(&_SessionRouter.TransactOpts)
}

// SetStakeDelay is a paid mutator transaction binding the contract method 0x3cadd8bb.
//
// Solidity: function setStakeDelay(int256 delay) returns()
func (_SessionRouter *SessionRouterTransactor) SetStakeDelay(opts *bind.TransactOpts, delay *big.Int) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "setStakeDelay", delay)
}

// SetStakeDelay is a paid mutator transaction binding the contract method 0x3cadd8bb.
//
// Solidity: function setStakeDelay(int256 delay) returns()
func (_SessionRouter *SessionRouterSession) SetStakeDelay(delay *big.Int) (*types.Transaction, error) {
	return _SessionRouter.Contract.SetStakeDelay(&_SessionRouter.TransactOpts, delay)
}

// SetStakeDelay is a paid mutator transaction binding the contract method 0x3cadd8bb.
//
// Solidity: function setStakeDelay(int256 delay) returns()
func (_SessionRouter *SessionRouterTransactorSession) SetStakeDelay(delay *big.Int) (*types.Transaction, error) {
	return _SessionRouter.Contract.SetStakeDelay(&_SessionRouter.TransactOpts, delay)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SessionRouter *SessionRouterTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SessionRouter.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SessionRouter *SessionRouterSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SessionRouter.Contract.TransferOwnership(&_SessionRouter.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SessionRouter *SessionRouterTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SessionRouter.Contract.TransferOwnership(&_SessionRouter.TransactOpts, newOwner)
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
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_SessionRouter *SessionRouterFilterer) FilterInitialized(opts *bind.FilterOpts) (*SessionRouterInitializedIterator, error) {

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &SessionRouterInitializedIterator{contract: _SessionRouter.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
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

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_SessionRouter *SessionRouterFilterer) ParseInitialized(log types.Log) (*SessionRouterInitialized, error) {
	event := new(SessionRouterInitialized)
	if err := _SessionRouter.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SessionRouter contract.
type SessionRouterOwnershipTransferredIterator struct {
	Event *SessionRouterOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *SessionRouterOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterOwnershipTransferred)
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
		it.Event = new(SessionRouterOwnershipTransferred)
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
func (it *SessionRouterOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterOwnershipTransferred represents a OwnershipTransferred event raised by the SessionRouter contract.
type SessionRouterOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SessionRouter *SessionRouterFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SessionRouterOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterOwnershipTransferredIterator{contract: _SessionRouter.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SessionRouter *SessionRouterFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SessionRouterOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterOwnershipTransferred)
				if err := _SessionRouter.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SessionRouter *SessionRouterFilterer) ParseOwnershipTransferred(log types.Log) (*SessionRouterOwnershipTransferred, error) {
	event := new(SessionRouterOwnershipTransferred)
	if err := _SessionRouter.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterProviderClaimedIterator is returned from FilterProviderClaimed and is used to iterate over the raw logs and unpacked data for ProviderClaimed events raised by the SessionRouter contract.
type SessionRouterProviderClaimedIterator struct {
	Event *SessionRouterProviderClaimed // Event containing the contract specifics and raw log

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
func (it *SessionRouterProviderClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterProviderClaimed)
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
		it.Event = new(SessionRouterProviderClaimed)
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
func (it *SessionRouterProviderClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterProviderClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterProviderClaimed represents a ProviderClaimed event raised by the SessionRouter contract.
type SessionRouterProviderClaimed struct {
	ProviderAddress common.Address
	Amount          *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterProviderClaimed is a free log retrieval operation binding the contract event 0x1cd322e3d02eade120b8dceb43a6c1dee437af36e7acd81726c4b54adf5584c2.
//
// Solidity: event ProviderClaimed(address indexed providerAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) FilterProviderClaimed(opts *bind.FilterOpts, providerAddress []common.Address) (*SessionRouterProviderClaimedIterator, error) {

	var providerAddressRule []interface{}
	for _, providerAddressItem := range providerAddress {
		providerAddressRule = append(providerAddressRule, providerAddressItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "ProviderClaimed", providerAddressRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterProviderClaimedIterator{contract: _SessionRouter.contract, event: "ProviderClaimed", logs: logs, sub: sub}, nil
}

// WatchProviderClaimed is a free log subscription operation binding the contract event 0x1cd322e3d02eade120b8dceb43a6c1dee437af36e7acd81726c4b54adf5584c2.
//
// Solidity: event ProviderClaimed(address indexed providerAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) WatchProviderClaimed(opts *bind.WatchOpts, sink chan<- *SessionRouterProviderClaimed, providerAddress []common.Address) (event.Subscription, error) {

	var providerAddressRule []interface{}
	for _, providerAddressItem := range providerAddress {
		providerAddressRule = append(providerAddressRule, providerAddressItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "ProviderClaimed", providerAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterProviderClaimed)
				if err := _SessionRouter.contract.UnpackLog(event, "ProviderClaimed", log); err != nil {
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

// ParseProviderClaimed is a log parse operation binding the contract event 0x1cd322e3d02eade120b8dceb43a6c1dee437af36e7acd81726c4b54adf5584c2.
//
// Solidity: event ProviderClaimed(address indexed providerAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) ParseProviderClaimed(log types.Log) (*SessionRouterProviderClaimed, error) {
	event := new(SessionRouterProviderClaimed)
	if err := _SessionRouter.contract.UnpackLog(event, "ProviderClaimed", log); err != nil {
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

// SessionRouterStakedIterator is returned from FilterStaked and is used to iterate over the raw logs and unpacked data for Staked events raised by the SessionRouter contract.
type SessionRouterStakedIterator struct {
	Event *SessionRouterStaked // Event containing the contract specifics and raw log

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
func (it *SessionRouterStakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterStaked)
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
		it.Event = new(SessionRouterStaked)
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
func (it *SessionRouterStakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterStakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterStaked represents a Staked event raised by the SessionRouter contract.
type SessionRouterStaked struct {
	UserAddress common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterStaked is a free log retrieval operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed userAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) FilterStaked(opts *bind.FilterOpts, userAddress []common.Address) (*SessionRouterStakedIterator, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "Staked", userAddressRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterStakedIterator{contract: _SessionRouter.contract, event: "Staked", logs: logs, sub: sub}, nil
}

// WatchStaked is a free log subscription operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed userAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) WatchStaked(opts *bind.WatchOpts, sink chan<- *SessionRouterStaked, userAddress []common.Address) (event.Subscription, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "Staked", userAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterStaked)
				if err := _SessionRouter.contract.UnpackLog(event, "Staked", log); err != nil {
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

// ParseStaked is a log parse operation binding the contract event 0x9e71bc8eea02a63969f509818f2dafb9254532904319f9dbda79b67bd34a5f3d.
//
// Solidity: event Staked(address indexed userAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) ParseStaked(log types.Log) (*SessionRouterStaked, error) {
	event := new(SessionRouterStaked)
	if err := _SessionRouter.contract.UnpackLog(event, "Staked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SessionRouterUnstakedIterator is returned from FilterUnstaked and is used to iterate over the raw logs and unpacked data for Unstaked events raised by the SessionRouter contract.
type SessionRouterUnstakedIterator struct {
	Event *SessionRouterUnstaked // Event containing the contract specifics and raw log

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
func (it *SessionRouterUnstakedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SessionRouterUnstaked)
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
		it.Event = new(SessionRouterUnstaked)
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
func (it *SessionRouterUnstakedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SessionRouterUnstakedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SessionRouterUnstaked represents a Unstaked event raised by the SessionRouter contract.
type SessionRouterUnstaked struct {
	UserAddress common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterUnstaked is a free log retrieval operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed userAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) FilterUnstaked(opts *bind.FilterOpts, userAddress []common.Address) (*SessionRouterUnstakedIterator, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}

	logs, sub, err := _SessionRouter.contract.FilterLogs(opts, "Unstaked", userAddressRule)
	if err != nil {
		return nil, err
	}
	return &SessionRouterUnstakedIterator{contract: _SessionRouter.contract, event: "Unstaked", logs: logs, sub: sub}, nil
}

// WatchUnstaked is a free log subscription operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed userAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) WatchUnstaked(opts *bind.WatchOpts, sink chan<- *SessionRouterUnstaked, userAddress []common.Address) (event.Subscription, error) {

	var userAddressRule []interface{}
	for _, userAddressItem := range userAddress {
		userAddressRule = append(userAddressRule, userAddressItem)
	}

	logs, sub, err := _SessionRouter.contract.WatchLogs(opts, "Unstaked", userAddressRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SessionRouterUnstaked)
				if err := _SessionRouter.contract.UnpackLog(event, "Unstaked", log); err != nil {
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

// ParseUnstaked is a log parse operation binding the contract event 0x0f5bb82176feb1b5e747e28471aa92156a04d9f3ab9f45f28e2d704232b93f75.
//
// Solidity: event Unstaked(address indexed userAddress, uint256 amount)
func (_SessionRouter *SessionRouterFilterer) ParseUnstaked(log types.Log) (*SessionRouterUnstaked, error) {
	event := new(SessionRouterUnstaked)
	if err := _SessionRouter.contract.UnpackLog(event, "Unstaked", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
