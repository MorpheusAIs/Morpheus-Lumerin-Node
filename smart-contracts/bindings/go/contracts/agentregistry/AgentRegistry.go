// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package agentregistry

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

// AgentRegistryAgent is an auto generated low-level Go binding around an user-defined struct.
type AgentRegistryAgent struct {
	AgentId   [32]byte
	Fee       *big.Int
	Stake     *big.Int
	Timestamp *big.Int
	Owner     common.Address
	Name      string
	Tags      []string
}

// AgentRegistryMetaData contains all meta data concerning the AgentRegistry contract.
var AgentRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"KeyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"KeyNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ModelNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSenderOrOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"StakeTooLow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroKey\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"agentId\",\"type\":\"bytes32\"}],\"name\":\"Deregistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newStake\",\"type\":\"uint256\"}],\"name\":\"MinStakeUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"agentId\",\"type\":\"bytes32\"}],\"name\":\"RegisteredUpdated\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"deregister\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"exists\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAll\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"agentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string[]\",\"name\":\"tags\",\"type\":\"string[]\"}],\"internalType\":\"structAgentRegistry.Agent[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getIds\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"map\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"agentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"addStake\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"agentId\",\"type\":\"bytes32\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string[]\",\"name\":\"tags\",\"type\":\"string[]\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_minStake\",\"type\":\"uint256\"}],\"name\":\"setMinStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"internalType\":\"contractERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// AgentRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use AgentRegistryMetaData.ABI instead.
var AgentRegistryABI = AgentRegistryMetaData.ABI

// AgentRegistry is an auto generated Go binding around an Ethereum contract.
type AgentRegistry struct {
	AgentRegistryCaller     // Read-only binding to the contract
	AgentRegistryTransactor // Write-only binding to the contract
	AgentRegistryFilterer   // Log filterer for contract events
}

// AgentRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type AgentRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AgentRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type AgentRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AgentRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type AgentRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// AgentRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type AgentRegistrySession struct {
	Contract     *AgentRegistry    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// AgentRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type AgentRegistryCallerSession struct {
	Contract *AgentRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// AgentRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type AgentRegistryTransactorSession struct {
	Contract     *AgentRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// AgentRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type AgentRegistryRaw struct {
	Contract *AgentRegistry // Generic contract binding to access the raw methods on
}

// AgentRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type AgentRegistryCallerRaw struct {
	Contract *AgentRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// AgentRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type AgentRegistryTransactorRaw struct {
	Contract *AgentRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewAgentRegistry creates a new instance of AgentRegistry, bound to a specific deployed contract.
func NewAgentRegistry(address common.Address, backend bind.ContractBackend) (*AgentRegistry, error) {
	contract, err := bindAgentRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &AgentRegistry{AgentRegistryCaller: AgentRegistryCaller{contract: contract}, AgentRegistryTransactor: AgentRegistryTransactor{contract: contract}, AgentRegistryFilterer: AgentRegistryFilterer{contract: contract}}, nil
}

// NewAgentRegistryCaller creates a new read-only instance of AgentRegistry, bound to a specific deployed contract.
func NewAgentRegistryCaller(address common.Address, caller bind.ContractCaller) (*AgentRegistryCaller, error) {
	contract, err := bindAgentRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &AgentRegistryCaller{contract: contract}, nil
}

// NewAgentRegistryTransactor creates a new write-only instance of AgentRegistry, bound to a specific deployed contract.
func NewAgentRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*AgentRegistryTransactor, error) {
	contract, err := bindAgentRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &AgentRegistryTransactor{contract: contract}, nil
}

// NewAgentRegistryFilterer creates a new log filterer instance of AgentRegistry, bound to a specific deployed contract.
func NewAgentRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*AgentRegistryFilterer, error) {
	contract, err := bindAgentRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &AgentRegistryFilterer{contract: contract}, nil
}

// bindAgentRegistry binds a generic wrapper to an already deployed contract.
func bindAgentRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := AgentRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AgentRegistry *AgentRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AgentRegistry.Contract.AgentRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AgentRegistry *AgentRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AgentRegistry.Contract.AgentRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AgentRegistry *AgentRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AgentRegistry.Contract.AgentRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_AgentRegistry *AgentRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _AgentRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_AgentRegistry *AgentRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AgentRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_AgentRegistry *AgentRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _AgentRegistry.Contract.contract.Transact(opts, method, params...)
}

// Exists is a free data retrieval call binding the contract method 0x38a699a4.
//
// Solidity: function exists(bytes32 id) view returns(bool)
func (_AgentRegistry *AgentRegistryCaller) Exists(opts *bind.CallOpts, id [32]byte) (bool, error) {
	var out []interface{}
	err := _AgentRegistry.contract.Call(opts, &out, "exists", id)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Exists is a free data retrieval call binding the contract method 0x38a699a4.
//
// Solidity: function exists(bytes32 id) view returns(bool)
func (_AgentRegistry *AgentRegistrySession) Exists(id [32]byte) (bool, error) {
	return _AgentRegistry.Contract.Exists(&_AgentRegistry.CallOpts, id)
}

// Exists is a free data retrieval call binding the contract method 0x38a699a4.
//
// Solidity: function exists(bytes32 id) view returns(bool)
func (_AgentRegistry *AgentRegistryCallerSession) Exists(id [32]byte) (bool, error) {
	return _AgentRegistry.Contract.Exists(&_AgentRegistry.CallOpts, id)
}

// GetAll is a free data retrieval call binding the contract method 0x53ed5143.
//
// Solidity: function getAll() view returns((bytes32,uint256,uint256,uint256,address,string,string[])[])
func (_AgentRegistry *AgentRegistryCaller) GetAll(opts *bind.CallOpts) ([]AgentRegistryAgent, error) {
	var out []interface{}
	err := _AgentRegistry.contract.Call(opts, &out, "getAll")

	if err != nil {
		return *new([]AgentRegistryAgent), err
	}

	out0 := *abi.ConvertType(out[0], new([]AgentRegistryAgent)).(*[]AgentRegistryAgent)

	return out0, err

}

// GetAll is a free data retrieval call binding the contract method 0x53ed5143.
//
// Solidity: function getAll() view returns((bytes32,uint256,uint256,uint256,address,string,string[])[])
func (_AgentRegistry *AgentRegistrySession) GetAll() ([]AgentRegistryAgent, error) {
	return _AgentRegistry.Contract.GetAll(&_AgentRegistry.CallOpts)
}

// GetAll is a free data retrieval call binding the contract method 0x53ed5143.
//
// Solidity: function getAll() view returns((bytes32,uint256,uint256,uint256,address,string,string[])[])
func (_AgentRegistry *AgentRegistryCallerSession) GetAll() ([]AgentRegistryAgent, error) {
	return _AgentRegistry.Contract.GetAll(&_AgentRegistry.CallOpts)
}

// GetIds is a free data retrieval call binding the contract method 0x2b105663.
//
// Solidity: function getIds() view returns(bytes32[])
func (_AgentRegistry *AgentRegistryCaller) GetIds(opts *bind.CallOpts) ([][32]byte, error) {
	var out []interface{}
	err := _AgentRegistry.contract.Call(opts, &out, "getIds")

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetIds is a free data retrieval call binding the contract method 0x2b105663.
//
// Solidity: function getIds() view returns(bytes32[])
func (_AgentRegistry *AgentRegistrySession) GetIds() ([][32]byte, error) {
	return _AgentRegistry.Contract.GetIds(&_AgentRegistry.CallOpts)
}

// GetIds is a free data retrieval call binding the contract method 0x2b105663.
//
// Solidity: function getIds() view returns(bytes32[])
func (_AgentRegistry *AgentRegistryCallerSession) GetIds() ([][32]byte, error) {
	return _AgentRegistry.Contract.GetIds(&_AgentRegistry.CallOpts)
}

// Map is a free data retrieval call binding the contract method 0x0ae186a8.
//
// Solidity: function map(bytes32 ) view returns(bytes32 agentId, uint256 fee, uint256 stake, uint256 timestamp, address owner, string name)
func (_AgentRegistry *AgentRegistryCaller) Map(opts *bind.CallOpts, arg0 [32]byte) (struct {
	AgentId   [32]byte
	Fee       *big.Int
	Stake     *big.Int
	Timestamp *big.Int
	Owner     common.Address
	Name      string
}, error) {
	var out []interface{}
	err := _AgentRegistry.contract.Call(opts, &out, "map", arg0)

	outstruct := new(struct {
		AgentId   [32]byte
		Fee       *big.Int
		Stake     *big.Int
		Timestamp *big.Int
		Owner     common.Address
		Name      string
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.AgentId = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.Fee = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Stake = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Timestamp = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.Owner = *abi.ConvertType(out[4], new(common.Address)).(*common.Address)
	outstruct.Name = *abi.ConvertType(out[5], new(string)).(*string)

	return *outstruct, err

}

// Map is a free data retrieval call binding the contract method 0x0ae186a8.
//
// Solidity: function map(bytes32 ) view returns(bytes32 agentId, uint256 fee, uint256 stake, uint256 timestamp, address owner, string name)
func (_AgentRegistry *AgentRegistrySession) Map(arg0 [32]byte) (struct {
	AgentId   [32]byte
	Fee       *big.Int
	Stake     *big.Int
	Timestamp *big.Int
	Owner     common.Address
	Name      string
}, error) {
	return _AgentRegistry.Contract.Map(&_AgentRegistry.CallOpts, arg0)
}

// Map is a free data retrieval call binding the contract method 0x0ae186a8.
//
// Solidity: function map(bytes32 ) view returns(bytes32 agentId, uint256 fee, uint256 stake, uint256 timestamp, address owner, string name)
func (_AgentRegistry *AgentRegistryCallerSession) Map(arg0 [32]byte) (struct {
	AgentId   [32]byte
	Fee       *big.Int
	Stake     *big.Int
	Timestamp *big.Int
	Owner     common.Address
	Name      string
}, error) {
	return _AgentRegistry.Contract.Map(&_AgentRegistry.CallOpts, arg0)
}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() view returns(uint256)
func (_AgentRegistry *AgentRegistryCaller) MinStake(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _AgentRegistry.contract.Call(opts, &out, "minStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() view returns(uint256)
func (_AgentRegistry *AgentRegistrySession) MinStake() (*big.Int, error) {
	return _AgentRegistry.Contract.MinStake(&_AgentRegistry.CallOpts)
}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() view returns(uint256)
func (_AgentRegistry *AgentRegistryCallerSession) MinStake() (*big.Int, error) {
	return _AgentRegistry.Contract.MinStake(&_AgentRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AgentRegistry *AgentRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AgentRegistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AgentRegistry *AgentRegistrySession) Owner() (common.Address, error) {
	return _AgentRegistry.Contract.Owner(&_AgentRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_AgentRegistry *AgentRegistryCallerSession) Owner() (common.Address, error) {
	return _AgentRegistry.Contract.Owner(&_AgentRegistry.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_AgentRegistry *AgentRegistryCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _AgentRegistry.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_AgentRegistry *AgentRegistrySession) Token() (common.Address, error) {
	return _AgentRegistry.Contract.Token(&_AgentRegistry.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_AgentRegistry *AgentRegistryCallerSession) Token() (common.Address, error) {
	return _AgentRegistry.Contract.Token(&_AgentRegistry.CallOpts)
}

// Deregister is a paid mutator transaction binding the contract method 0x20813154.
//
// Solidity: function deregister(bytes32 id) returns()
func (_AgentRegistry *AgentRegistryTransactor) Deregister(opts *bind.TransactOpts, id [32]byte) (*types.Transaction, error) {
	return _AgentRegistry.contract.Transact(opts, "deregister", id)
}

// Deregister is a paid mutator transaction binding the contract method 0x20813154.
//
// Solidity: function deregister(bytes32 id) returns()
func (_AgentRegistry *AgentRegistrySession) Deregister(id [32]byte) (*types.Transaction, error) {
	return _AgentRegistry.Contract.Deregister(&_AgentRegistry.TransactOpts, id)
}

// Deregister is a paid mutator transaction binding the contract method 0x20813154.
//
// Solidity: function deregister(bytes32 id) returns()
func (_AgentRegistry *AgentRegistryTransactorSession) Deregister(id [32]byte) (*types.Transaction, error) {
	return _AgentRegistry.Contract.Deregister(&_AgentRegistry.TransactOpts, id)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _token) returns()
func (_AgentRegistry *AgentRegistryTransactor) Initialize(opts *bind.TransactOpts, _token common.Address) (*types.Transaction, error) {
	return _AgentRegistry.contract.Transact(opts, "initialize", _token)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _token) returns()
func (_AgentRegistry *AgentRegistrySession) Initialize(_token common.Address) (*types.Transaction, error) {
	return _AgentRegistry.Contract.Initialize(&_AgentRegistry.TransactOpts, _token)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _token) returns()
func (_AgentRegistry *AgentRegistryTransactorSession) Initialize(_token common.Address) (*types.Transaction, error) {
	return _AgentRegistry.Contract.Initialize(&_AgentRegistry.TransactOpts, _token)
}

// Register is a paid mutator transaction binding the contract method 0xcca4cb64.
//
// Solidity: function register(uint256 addStake, uint256 fee, address owner, bytes32 agentId, string name, string[] tags) returns()
func (_AgentRegistry *AgentRegistryTransactor) Register(opts *bind.TransactOpts, addStake *big.Int, fee *big.Int, owner common.Address, agentId [32]byte, name string, tags []string) (*types.Transaction, error) {
	return _AgentRegistry.contract.Transact(opts, "register", addStake, fee, owner, agentId, name, tags)
}

// Register is a paid mutator transaction binding the contract method 0xcca4cb64.
//
// Solidity: function register(uint256 addStake, uint256 fee, address owner, bytes32 agentId, string name, string[] tags) returns()
func (_AgentRegistry *AgentRegistrySession) Register(addStake *big.Int, fee *big.Int, owner common.Address, agentId [32]byte, name string, tags []string) (*types.Transaction, error) {
	return _AgentRegistry.Contract.Register(&_AgentRegistry.TransactOpts, addStake, fee, owner, agentId, name, tags)
}

// Register is a paid mutator transaction binding the contract method 0xcca4cb64.
//
// Solidity: function register(uint256 addStake, uint256 fee, address owner, bytes32 agentId, string name, string[] tags) returns()
func (_AgentRegistry *AgentRegistryTransactorSession) Register(addStake *big.Int, fee *big.Int, owner common.Address, agentId [32]byte, name string, tags []string) (*types.Transaction, error) {
	return _AgentRegistry.Contract.Register(&_AgentRegistry.TransactOpts, addStake, fee, owner, agentId, name, tags)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AgentRegistry *AgentRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _AgentRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AgentRegistry *AgentRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _AgentRegistry.Contract.RenounceOwnership(&_AgentRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_AgentRegistry *AgentRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _AgentRegistry.Contract.RenounceOwnership(&_AgentRegistry.TransactOpts)
}

// SetMinStake is a paid mutator transaction binding the contract method 0x8c80fd90.
//
// Solidity: function setMinStake(uint256 _minStake) returns()
func (_AgentRegistry *AgentRegistryTransactor) SetMinStake(opts *bind.TransactOpts, _minStake *big.Int) (*types.Transaction, error) {
	return _AgentRegistry.contract.Transact(opts, "setMinStake", _minStake)
}

// SetMinStake is a paid mutator transaction binding the contract method 0x8c80fd90.
//
// Solidity: function setMinStake(uint256 _minStake) returns()
func (_AgentRegistry *AgentRegistrySession) SetMinStake(_minStake *big.Int) (*types.Transaction, error) {
	return _AgentRegistry.Contract.SetMinStake(&_AgentRegistry.TransactOpts, _minStake)
}

// SetMinStake is a paid mutator transaction binding the contract method 0x8c80fd90.
//
// Solidity: function setMinStake(uint256 _minStake) returns()
func (_AgentRegistry *AgentRegistryTransactorSession) SetMinStake(_minStake *big.Int) (*types.Transaction, error) {
	return _AgentRegistry.Contract.SetMinStake(&_AgentRegistry.TransactOpts, _minStake)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AgentRegistry *AgentRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _AgentRegistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AgentRegistry *AgentRegistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AgentRegistry.Contract.TransferOwnership(&_AgentRegistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_AgentRegistry *AgentRegistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _AgentRegistry.Contract.TransferOwnership(&_AgentRegistry.TransactOpts, newOwner)
}

// AgentRegistryDeregisteredIterator is returned from FilterDeregistered and is used to iterate over the raw logs and unpacked data for Deregistered events raised by the AgentRegistry contract.
type AgentRegistryDeregisteredIterator struct {
	Event *AgentRegistryDeregistered // Event containing the contract specifics and raw log

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
func (it *AgentRegistryDeregisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentRegistryDeregistered)
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
		it.Event = new(AgentRegistryDeregistered)
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
func (it *AgentRegistryDeregisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentRegistryDeregisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentRegistryDeregistered represents a Deregistered event raised by the AgentRegistry contract.
type AgentRegistryDeregistered struct {
	Owner   common.Address
	AgentId [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterDeregistered is a free log retrieval operation binding the contract event 0xcc0d8d49cb9a5a3eebc1e1f0b607b8aad83fbe9618e448df79b8a0cc33319472.
//
// Solidity: event Deregistered(address indexed owner, bytes32 indexed agentId)
func (_AgentRegistry *AgentRegistryFilterer) FilterDeregistered(opts *bind.FilterOpts, owner []common.Address, agentId [][32]byte) (*AgentRegistryDeregisteredIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var agentIdRule []interface{}
	for _, agentIdItem := range agentId {
		agentIdRule = append(agentIdRule, agentIdItem)
	}

	logs, sub, err := _AgentRegistry.contract.FilterLogs(opts, "Deregistered", ownerRule, agentIdRule)
	if err != nil {
		return nil, err
	}
	return &AgentRegistryDeregisteredIterator{contract: _AgentRegistry.contract, event: "Deregistered", logs: logs, sub: sub}, nil
}

// WatchDeregistered is a free log subscription operation binding the contract event 0xcc0d8d49cb9a5a3eebc1e1f0b607b8aad83fbe9618e448df79b8a0cc33319472.
//
// Solidity: event Deregistered(address indexed owner, bytes32 indexed agentId)
func (_AgentRegistry *AgentRegistryFilterer) WatchDeregistered(opts *bind.WatchOpts, sink chan<- *AgentRegistryDeregistered, owner []common.Address, agentId [][32]byte) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var agentIdRule []interface{}
	for _, agentIdItem := range agentId {
		agentIdRule = append(agentIdRule, agentIdItem)
	}

	logs, sub, err := _AgentRegistry.contract.WatchLogs(opts, "Deregistered", ownerRule, agentIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentRegistryDeregistered)
				if err := _AgentRegistry.contract.UnpackLog(event, "Deregistered", log); err != nil {
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

// ParseDeregistered is a log parse operation binding the contract event 0xcc0d8d49cb9a5a3eebc1e1f0b607b8aad83fbe9618e448df79b8a0cc33319472.
//
// Solidity: event Deregistered(address indexed owner, bytes32 indexed agentId)
func (_AgentRegistry *AgentRegistryFilterer) ParseDeregistered(log types.Log) (*AgentRegistryDeregistered, error) {
	event := new(AgentRegistryDeregistered)
	if err := _AgentRegistry.contract.UnpackLog(event, "Deregistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AgentRegistryInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the AgentRegistry contract.
type AgentRegistryInitializedIterator struct {
	Event *AgentRegistryInitialized // Event containing the contract specifics and raw log

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
func (it *AgentRegistryInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentRegistryInitialized)
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
		it.Event = new(AgentRegistryInitialized)
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
func (it *AgentRegistryInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentRegistryInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentRegistryInitialized represents a Initialized event raised by the AgentRegistry contract.
type AgentRegistryInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_AgentRegistry *AgentRegistryFilterer) FilterInitialized(opts *bind.FilterOpts) (*AgentRegistryInitializedIterator, error) {

	logs, sub, err := _AgentRegistry.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &AgentRegistryInitializedIterator{contract: _AgentRegistry.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_AgentRegistry *AgentRegistryFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *AgentRegistryInitialized) (event.Subscription, error) {

	logs, sub, err := _AgentRegistry.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentRegistryInitialized)
				if err := _AgentRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_AgentRegistry *AgentRegistryFilterer) ParseInitialized(log types.Log) (*AgentRegistryInitialized, error) {
	event := new(AgentRegistryInitialized)
	if err := _AgentRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AgentRegistryMinStakeUpdatedIterator is returned from FilterMinStakeUpdated and is used to iterate over the raw logs and unpacked data for MinStakeUpdated events raised by the AgentRegistry contract.
type AgentRegistryMinStakeUpdatedIterator struct {
	Event *AgentRegistryMinStakeUpdated // Event containing the contract specifics and raw log

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
func (it *AgentRegistryMinStakeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentRegistryMinStakeUpdated)
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
		it.Event = new(AgentRegistryMinStakeUpdated)
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
func (it *AgentRegistryMinStakeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentRegistryMinStakeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentRegistryMinStakeUpdated represents a MinStakeUpdated event raised by the AgentRegistry contract.
type AgentRegistryMinStakeUpdated struct {
	NewStake *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterMinStakeUpdated is a free log retrieval operation binding the contract event 0x47ab46f2c8d4258304a2f5551c1cbdb6981be49631365d1ba7191288a73f39ef.
//
// Solidity: event MinStakeUpdated(uint256 newStake)
func (_AgentRegistry *AgentRegistryFilterer) FilterMinStakeUpdated(opts *bind.FilterOpts) (*AgentRegistryMinStakeUpdatedIterator, error) {

	logs, sub, err := _AgentRegistry.contract.FilterLogs(opts, "MinStakeUpdated")
	if err != nil {
		return nil, err
	}
	return &AgentRegistryMinStakeUpdatedIterator{contract: _AgentRegistry.contract, event: "MinStakeUpdated", logs: logs, sub: sub}, nil
}

// WatchMinStakeUpdated is a free log subscription operation binding the contract event 0x47ab46f2c8d4258304a2f5551c1cbdb6981be49631365d1ba7191288a73f39ef.
//
// Solidity: event MinStakeUpdated(uint256 newStake)
func (_AgentRegistry *AgentRegistryFilterer) WatchMinStakeUpdated(opts *bind.WatchOpts, sink chan<- *AgentRegistryMinStakeUpdated) (event.Subscription, error) {

	logs, sub, err := _AgentRegistry.contract.WatchLogs(opts, "MinStakeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentRegistryMinStakeUpdated)
				if err := _AgentRegistry.contract.UnpackLog(event, "MinStakeUpdated", log); err != nil {
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

// ParseMinStakeUpdated is a log parse operation binding the contract event 0x47ab46f2c8d4258304a2f5551c1cbdb6981be49631365d1ba7191288a73f39ef.
//
// Solidity: event MinStakeUpdated(uint256 newStake)
func (_AgentRegistry *AgentRegistryFilterer) ParseMinStakeUpdated(log types.Log) (*AgentRegistryMinStakeUpdated, error) {
	event := new(AgentRegistryMinStakeUpdated)
	if err := _AgentRegistry.contract.UnpackLog(event, "MinStakeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AgentRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the AgentRegistry contract.
type AgentRegistryOwnershipTransferredIterator struct {
	Event *AgentRegistryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *AgentRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentRegistryOwnershipTransferred)
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
		it.Event = new(AgentRegistryOwnershipTransferred)
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
func (it *AgentRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the AgentRegistry contract.
type AgentRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AgentRegistry *AgentRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*AgentRegistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AgentRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &AgentRegistryOwnershipTransferredIterator{contract: _AgentRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_AgentRegistry *AgentRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *AgentRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _AgentRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentRegistryOwnershipTransferred)
				if err := _AgentRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_AgentRegistry *AgentRegistryFilterer) ParseOwnershipTransferred(log types.Log) (*AgentRegistryOwnershipTransferred, error) {
	event := new(AgentRegistryOwnershipTransferred)
	if err := _AgentRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// AgentRegistryRegisteredUpdatedIterator is returned from FilterRegisteredUpdated and is used to iterate over the raw logs and unpacked data for RegisteredUpdated events raised by the AgentRegistry contract.
type AgentRegistryRegisteredUpdatedIterator struct {
	Event *AgentRegistryRegisteredUpdated // Event containing the contract specifics and raw log

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
func (it *AgentRegistryRegisteredUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(AgentRegistryRegisteredUpdated)
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
		it.Event = new(AgentRegistryRegisteredUpdated)
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
func (it *AgentRegistryRegisteredUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *AgentRegistryRegisteredUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// AgentRegistryRegisteredUpdated represents a RegisteredUpdated event raised by the AgentRegistry contract.
type AgentRegistryRegisteredUpdated struct {
	Owner   common.Address
	AgentId [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRegisteredUpdated is a free log retrieval operation binding the contract event 0x4dffc4e92367641816479c647d1ff3f03202f6c73b97340cba51a7b577c55804.
//
// Solidity: event RegisteredUpdated(address indexed owner, bytes32 indexed agentId)
func (_AgentRegistry *AgentRegistryFilterer) FilterRegisteredUpdated(opts *bind.FilterOpts, owner []common.Address, agentId [][32]byte) (*AgentRegistryRegisteredUpdatedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var agentIdRule []interface{}
	for _, agentIdItem := range agentId {
		agentIdRule = append(agentIdRule, agentIdItem)
	}

	logs, sub, err := _AgentRegistry.contract.FilterLogs(opts, "RegisteredUpdated", ownerRule, agentIdRule)
	if err != nil {
		return nil, err
	}
	return &AgentRegistryRegisteredUpdatedIterator{contract: _AgentRegistry.contract, event: "RegisteredUpdated", logs: logs, sub: sub}, nil
}

// WatchRegisteredUpdated is a free log subscription operation binding the contract event 0x4dffc4e92367641816479c647d1ff3f03202f6c73b97340cba51a7b577c55804.
//
// Solidity: event RegisteredUpdated(address indexed owner, bytes32 indexed agentId)
func (_AgentRegistry *AgentRegistryFilterer) WatchRegisteredUpdated(opts *bind.WatchOpts, sink chan<- *AgentRegistryRegisteredUpdated, owner []common.Address, agentId [][32]byte) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var agentIdRule []interface{}
	for _, agentIdItem := range agentId {
		agentIdRule = append(agentIdRule, agentIdItem)
	}

	logs, sub, err := _AgentRegistry.contract.WatchLogs(opts, "RegisteredUpdated", ownerRule, agentIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(AgentRegistryRegisteredUpdated)
				if err := _AgentRegistry.contract.UnpackLog(event, "RegisteredUpdated", log); err != nil {
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

// ParseRegisteredUpdated is a log parse operation binding the contract event 0x4dffc4e92367641816479c647d1ff3f03202f6c73b97340cba51a7b577c55804.
//
// Solidity: event RegisteredUpdated(address indexed owner, bytes32 indexed agentId)
func (_AgentRegistry *AgentRegistryFilterer) ParseRegisteredUpdated(log types.Log) (*AgentRegistryRegisteredUpdated, error) {
	event := new(AgentRegistryRegisteredUpdated)
	if err := _AgentRegistry.contract.UnpackLog(event, "RegisteredUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
