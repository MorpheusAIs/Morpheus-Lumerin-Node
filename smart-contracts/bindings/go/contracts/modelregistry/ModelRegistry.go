// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package modelregistry

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

// ModelRegistryModel is an auto generated low-level Go binding around an user-defined struct.
type ModelRegistryModel struct {
	IpfsCID   [32]byte
	Fee       *big.Int
	Stake     *big.Int
	Owner     common.Address
	Name      string
	Tags      []string
	Timestamp *big.Int
	IsDeleted bool
}

// ModelRegistryMetaData contains all meta data concerning the ModelRegistry contract.
var ModelRegistryMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"KeyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"KeyNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ModelNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSenderOrOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"StakeTooLow\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroKey\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"}],\"name\":\"Deregistered\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newStake\",\"type\":\"uint256\"}],\"name\":\"MinStakeUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"}],\"name\":\"RegisteredUpdated\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"deregister\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"id\",\"type\":\"bytes32\"}],\"name\":\"exists\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getAll\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ipfsCID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string[]\",\"name\":\"tags\",\"type\":\"string[]\"},{\"internalType\":\"uint128\",\"name\":\"timestamp\",\"type\":\"uint128\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"internalType\":\"structModelRegistry.Model[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"getByIndex\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ipfsCID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string[]\",\"name\":\"tags\",\"type\":\"string[]\"},{\"internalType\":\"uint128\",\"name\":\"timestamp\",\"type\":\"uint128\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"internalType\":\"structModelRegistry.Model\",\"name\":\"model\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getCount\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"count\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getIds\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"map\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"ipfsCID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"uint128\",\"name\":\"timestamp\",\"type\":\"uint128\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"minStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"models\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"ipfsCID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"addStake\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string[]\",\"name\":\"tags\",\"type\":\"string[]\"}],\"name\":\"register\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_minStake\",\"type\":\"uint256\"}],\"name\":\"setMinStake\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"internalType\":\"contractERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ModelRegistryABI is the input ABI used to generate the binding from.
// Deprecated: Use ModelRegistryMetaData.ABI instead.
var ModelRegistryABI = ModelRegistryMetaData.ABI

// ModelRegistry is an auto generated Go binding around an Ethereum contract.
type ModelRegistry struct {
	ModelRegistryCaller     // Read-only binding to the contract
	ModelRegistryTransactor // Write-only binding to the contract
	ModelRegistryFilterer   // Log filterer for contract events
}

// ModelRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type ModelRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ModelRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ModelRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ModelRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ModelRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ModelRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ModelRegistrySession struct {
	Contract     *ModelRegistry    // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ModelRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ModelRegistryCallerSession struct {
	Contract *ModelRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts        // Call options to use throughout this session
}

// ModelRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ModelRegistryTransactorSession struct {
	Contract     *ModelRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts        // Transaction auth options to use throughout this session
}

// ModelRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type ModelRegistryRaw struct {
	Contract *ModelRegistry // Generic contract binding to access the raw methods on
}

// ModelRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ModelRegistryCallerRaw struct {
	Contract *ModelRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// ModelRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ModelRegistryTransactorRaw struct {
	Contract *ModelRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewModelRegistry creates a new instance of ModelRegistry, bound to a specific deployed contract.
func NewModelRegistry(address common.Address, backend bind.ContractBackend) (*ModelRegistry, error) {
	contract, err := bindModelRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ModelRegistry{ModelRegistryCaller: ModelRegistryCaller{contract: contract}, ModelRegistryTransactor: ModelRegistryTransactor{contract: contract}, ModelRegistryFilterer: ModelRegistryFilterer{contract: contract}}, nil
}

// NewModelRegistryCaller creates a new read-only instance of ModelRegistry, bound to a specific deployed contract.
func NewModelRegistryCaller(address common.Address, caller bind.ContractCaller) (*ModelRegistryCaller, error) {
	contract, err := bindModelRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ModelRegistryCaller{contract: contract}, nil
}

// NewModelRegistryTransactor creates a new write-only instance of ModelRegistry, bound to a specific deployed contract.
func NewModelRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*ModelRegistryTransactor, error) {
	contract, err := bindModelRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ModelRegistryTransactor{contract: contract}, nil
}

// NewModelRegistryFilterer creates a new log filterer instance of ModelRegistry, bound to a specific deployed contract.
func NewModelRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*ModelRegistryFilterer, error) {
	contract, err := bindModelRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ModelRegistryFilterer{contract: contract}, nil
}

// bindModelRegistry binds a generic wrapper to an already deployed contract.
func bindModelRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ModelRegistryMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ModelRegistry *ModelRegistryRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ModelRegistry.Contract.ModelRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ModelRegistry *ModelRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ModelRegistry.Contract.ModelRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ModelRegistry *ModelRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ModelRegistry.Contract.ModelRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ModelRegistry *ModelRegistryCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ModelRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ModelRegistry *ModelRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ModelRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ModelRegistry *ModelRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ModelRegistry.Contract.contract.Transact(opts, method, params...)
}

// Exists is a free data retrieval call binding the contract method 0x38a699a4.
//
// Solidity: function exists(bytes32 id) view returns(bool)
func (_ModelRegistry *ModelRegistryCaller) Exists(opts *bind.CallOpts, id [32]byte) (bool, error) {
	var out []interface{}
	err := _ModelRegistry.contract.Call(opts, &out, "exists", id)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Exists is a free data retrieval call binding the contract method 0x38a699a4.
//
// Solidity: function exists(bytes32 id) view returns(bool)
func (_ModelRegistry *ModelRegistrySession) Exists(id [32]byte) (bool, error) {
	return _ModelRegistry.Contract.Exists(&_ModelRegistry.CallOpts, id)
}

// Exists is a free data retrieval call binding the contract method 0x38a699a4.
//
// Solidity: function exists(bytes32 id) view returns(bool)
func (_ModelRegistry *ModelRegistryCallerSession) Exists(id [32]byte) (bool, error) {
	return _ModelRegistry.Contract.Exists(&_ModelRegistry.CallOpts, id)
}

// GetAll is a free data retrieval call binding the contract method 0x53ed5143.
//
// Solidity: function getAll() view returns((bytes32,uint256,uint256,address,string,string[],uint128,bool)[])
func (_ModelRegistry *ModelRegistryCaller) GetAll(opts *bind.CallOpts) ([]ModelRegistryModel, error) {
	var out []interface{}
	err := _ModelRegistry.contract.Call(opts, &out, "getAll")

	if err != nil {
		return *new([]ModelRegistryModel), err
	}

	out0 := *abi.ConvertType(out[0], new([]ModelRegistryModel)).(*[]ModelRegistryModel)

	return out0, err

}

// GetAll is a free data retrieval call binding the contract method 0x53ed5143.
//
// Solidity: function getAll() view returns((bytes32,uint256,uint256,address,string,string[],uint128,bool)[])
func (_ModelRegistry *ModelRegistrySession) GetAll() ([]ModelRegistryModel, error) {
	return _ModelRegistry.Contract.GetAll(&_ModelRegistry.CallOpts)
}

// GetAll is a free data retrieval call binding the contract method 0x53ed5143.
//
// Solidity: function getAll() view returns((bytes32,uint256,uint256,address,string,string[],uint128,bool)[])
func (_ModelRegistry *ModelRegistryCallerSession) GetAll() ([]ModelRegistryModel, error) {
	return _ModelRegistry.Contract.GetAll(&_ModelRegistry.CallOpts)
}

// GetByIndex is a free data retrieval call binding the contract method 0x2d883a73.
//
// Solidity: function getByIndex(uint256 index) view returns(bytes32 modelId, (bytes32,uint256,uint256,address,string,string[],uint128,bool) model)
func (_ModelRegistry *ModelRegistryCaller) GetByIndex(opts *bind.CallOpts, index *big.Int) (struct {
	ModelId [32]byte
	Model   ModelRegistryModel
}, error) {
	var out []interface{}
	err := _ModelRegistry.contract.Call(opts, &out, "getByIndex", index)

	outstruct := new(struct {
		ModelId [32]byte
		Model   ModelRegistryModel
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ModelId = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.Model = *abi.ConvertType(out[1], new(ModelRegistryModel)).(*ModelRegistryModel)

	return *outstruct, err

}

// GetByIndex is a free data retrieval call binding the contract method 0x2d883a73.
//
// Solidity: function getByIndex(uint256 index) view returns(bytes32 modelId, (bytes32,uint256,uint256,address,string,string[],uint128,bool) model)
func (_ModelRegistry *ModelRegistrySession) GetByIndex(index *big.Int) (struct {
	ModelId [32]byte
	Model   ModelRegistryModel
}, error) {
	return _ModelRegistry.Contract.GetByIndex(&_ModelRegistry.CallOpts, index)
}

// GetByIndex is a free data retrieval call binding the contract method 0x2d883a73.
//
// Solidity: function getByIndex(uint256 index) view returns(bytes32 modelId, (bytes32,uint256,uint256,address,string,string[],uint128,bool) model)
func (_ModelRegistry *ModelRegistryCallerSession) GetByIndex(index *big.Int) (struct {
	ModelId [32]byte
	Model   ModelRegistryModel
}, error) {
	return _ModelRegistry.Contract.GetByIndex(&_ModelRegistry.CallOpts, index)
}

// GetCount is a free data retrieval call binding the contract method 0xa87d942c.
//
// Solidity: function getCount() view returns(uint256 count)
func (_ModelRegistry *ModelRegistryCaller) GetCount(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ModelRegistry.contract.Call(opts, &out, "getCount")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetCount is a free data retrieval call binding the contract method 0xa87d942c.
//
// Solidity: function getCount() view returns(uint256 count)
func (_ModelRegistry *ModelRegistrySession) GetCount() (*big.Int, error) {
	return _ModelRegistry.Contract.GetCount(&_ModelRegistry.CallOpts)
}

// GetCount is a free data retrieval call binding the contract method 0xa87d942c.
//
// Solidity: function getCount() view returns(uint256 count)
func (_ModelRegistry *ModelRegistryCallerSession) GetCount() (*big.Int, error) {
	return _ModelRegistry.Contract.GetCount(&_ModelRegistry.CallOpts)
}

// GetIds is a free data retrieval call binding the contract method 0x2b105663.
//
// Solidity: function getIds() view returns(bytes32[])
func (_ModelRegistry *ModelRegistryCaller) GetIds(opts *bind.CallOpts) ([][32]byte, error) {
	var out []interface{}
	err := _ModelRegistry.contract.Call(opts, &out, "getIds")

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetIds is a free data retrieval call binding the contract method 0x2b105663.
//
// Solidity: function getIds() view returns(bytes32[])
func (_ModelRegistry *ModelRegistrySession) GetIds() ([][32]byte, error) {
	return _ModelRegistry.Contract.GetIds(&_ModelRegistry.CallOpts)
}

// GetIds is a free data retrieval call binding the contract method 0x2b105663.
//
// Solidity: function getIds() view returns(bytes32[])
func (_ModelRegistry *ModelRegistryCallerSession) GetIds() ([][32]byte, error) {
	return _ModelRegistry.Contract.GetIds(&_ModelRegistry.CallOpts)
}

// Map is a free data retrieval call binding the contract method 0x0ae186a8.
//
// Solidity: function map(bytes32 ) view returns(bytes32 ipfsCID, uint256 fee, uint256 stake, address owner, string name, uint128 timestamp, bool isDeleted)
func (_ModelRegistry *ModelRegistryCaller) Map(opts *bind.CallOpts, arg0 [32]byte) (struct {
	IpfsCID   [32]byte
	Fee       *big.Int
	Stake     *big.Int
	Owner     common.Address
	Name      string
	Timestamp *big.Int
	IsDeleted bool
}, error) {
	var out []interface{}
	err := _ModelRegistry.contract.Call(opts, &out, "map", arg0)

	outstruct := new(struct {
		IpfsCID   [32]byte
		Fee       *big.Int
		Stake     *big.Int
		Owner     common.Address
		Name      string
		Timestamp *big.Int
		IsDeleted bool
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.IpfsCID = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.Fee = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)
	outstruct.Stake = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Owner = *abi.ConvertType(out[3], new(common.Address)).(*common.Address)
	outstruct.Name = *abi.ConvertType(out[4], new(string)).(*string)
	outstruct.Timestamp = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)
	outstruct.IsDeleted = *abi.ConvertType(out[6], new(bool)).(*bool)

	return *outstruct, err

}

// Map is a free data retrieval call binding the contract method 0x0ae186a8.
//
// Solidity: function map(bytes32 ) view returns(bytes32 ipfsCID, uint256 fee, uint256 stake, address owner, string name, uint128 timestamp, bool isDeleted)
func (_ModelRegistry *ModelRegistrySession) Map(arg0 [32]byte) (struct {
	IpfsCID   [32]byte
	Fee       *big.Int
	Stake     *big.Int
	Owner     common.Address
	Name      string
	Timestamp *big.Int
	IsDeleted bool
}, error) {
	return _ModelRegistry.Contract.Map(&_ModelRegistry.CallOpts, arg0)
}

// Map is a free data retrieval call binding the contract method 0x0ae186a8.
//
// Solidity: function map(bytes32 ) view returns(bytes32 ipfsCID, uint256 fee, uint256 stake, address owner, string name, uint128 timestamp, bool isDeleted)
func (_ModelRegistry *ModelRegistryCallerSession) Map(arg0 [32]byte) (struct {
	IpfsCID   [32]byte
	Fee       *big.Int
	Stake     *big.Int
	Owner     common.Address
	Name      string
	Timestamp *big.Int
	IsDeleted bool
}, error) {
	return _ModelRegistry.Contract.Map(&_ModelRegistry.CallOpts, arg0)
}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() view returns(uint256)
func (_ModelRegistry *ModelRegistryCaller) MinStake(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ModelRegistry.contract.Call(opts, &out, "minStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() view returns(uint256)
func (_ModelRegistry *ModelRegistrySession) MinStake() (*big.Int, error) {
	return _ModelRegistry.Contract.MinStake(&_ModelRegistry.CallOpts)
}

// MinStake is a free data retrieval call binding the contract method 0x375b3c0a.
//
// Solidity: function minStake() view returns(uint256)
func (_ModelRegistry *ModelRegistryCallerSession) MinStake() (*big.Int, error) {
	return _ModelRegistry.Contract.MinStake(&_ModelRegistry.CallOpts)
}

// Models is a free data retrieval call binding the contract method 0x6a030ca9.
//
// Solidity: function models(uint256 ) view returns(bytes32)
func (_ModelRegistry *ModelRegistryCaller) Models(opts *bind.CallOpts, arg0 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _ModelRegistry.contract.Call(opts, &out, "models", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// Models is a free data retrieval call binding the contract method 0x6a030ca9.
//
// Solidity: function models(uint256 ) view returns(bytes32)
func (_ModelRegistry *ModelRegistrySession) Models(arg0 *big.Int) ([32]byte, error) {
	return _ModelRegistry.Contract.Models(&_ModelRegistry.CallOpts, arg0)
}

// Models is a free data retrieval call binding the contract method 0x6a030ca9.
//
// Solidity: function models(uint256 ) view returns(bytes32)
func (_ModelRegistry *ModelRegistryCallerSession) Models(arg0 *big.Int) ([32]byte, error) {
	return _ModelRegistry.Contract.Models(&_ModelRegistry.CallOpts, arg0)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ModelRegistry *ModelRegistryCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ModelRegistry.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ModelRegistry *ModelRegistrySession) Owner() (common.Address, error) {
	return _ModelRegistry.Contract.Owner(&_ModelRegistry.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ModelRegistry *ModelRegistryCallerSession) Owner() (common.Address, error) {
	return _ModelRegistry.Contract.Owner(&_ModelRegistry.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_ModelRegistry *ModelRegistryCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ModelRegistry.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_ModelRegistry *ModelRegistrySession) Token() (common.Address, error) {
	return _ModelRegistry.Contract.Token(&_ModelRegistry.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_ModelRegistry *ModelRegistryCallerSession) Token() (common.Address, error) {
	return _ModelRegistry.Contract.Token(&_ModelRegistry.CallOpts)
}

// Deregister is a paid mutator transaction binding the contract method 0x20813154.
//
// Solidity: function deregister(bytes32 id) returns()
func (_ModelRegistry *ModelRegistryTransactor) Deregister(opts *bind.TransactOpts, id [32]byte) (*types.Transaction, error) {
	return _ModelRegistry.contract.Transact(opts, "deregister", id)
}

// Deregister is a paid mutator transaction binding the contract method 0x20813154.
//
// Solidity: function deregister(bytes32 id) returns()
func (_ModelRegistry *ModelRegistrySession) Deregister(id [32]byte) (*types.Transaction, error) {
	return _ModelRegistry.Contract.Deregister(&_ModelRegistry.TransactOpts, id)
}

// Deregister is a paid mutator transaction binding the contract method 0x20813154.
//
// Solidity: function deregister(bytes32 id) returns()
func (_ModelRegistry *ModelRegistryTransactorSession) Deregister(id [32]byte) (*types.Transaction, error) {
	return _ModelRegistry.Contract.Deregister(&_ModelRegistry.TransactOpts, id)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _token) returns()
func (_ModelRegistry *ModelRegistryTransactor) Initialize(opts *bind.TransactOpts, _token common.Address) (*types.Transaction, error) {
	return _ModelRegistry.contract.Transact(opts, "initialize", _token)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _token) returns()
func (_ModelRegistry *ModelRegistrySession) Initialize(_token common.Address) (*types.Transaction, error) {
	return _ModelRegistry.Contract.Initialize(&_ModelRegistry.TransactOpts, _token)
}

// Initialize is a paid mutator transaction binding the contract method 0xc4d66de8.
//
// Solidity: function initialize(address _token) returns()
func (_ModelRegistry *ModelRegistryTransactorSession) Initialize(_token common.Address) (*types.Transaction, error) {
	return _ModelRegistry.Contract.Initialize(&_ModelRegistry.TransactOpts, _token)
}

// Register is a paid mutator transaction binding the contract method 0x9e4aaa05.
//
// Solidity: function register(bytes32 modelId, bytes32 ipfsCID, uint256 fee, uint256 addStake, address owner, string name, string[] tags) returns()
func (_ModelRegistry *ModelRegistryTransactor) Register(opts *bind.TransactOpts, modelId [32]byte, ipfsCID [32]byte, fee *big.Int, addStake *big.Int, owner common.Address, name string, tags []string) (*types.Transaction, error) {
	return _ModelRegistry.contract.Transact(opts, "register", modelId, ipfsCID, fee, addStake, owner, name, tags)
}

// Register is a paid mutator transaction binding the contract method 0x9e4aaa05.
//
// Solidity: function register(bytes32 modelId, bytes32 ipfsCID, uint256 fee, uint256 addStake, address owner, string name, string[] tags) returns()
func (_ModelRegistry *ModelRegistrySession) Register(modelId [32]byte, ipfsCID [32]byte, fee *big.Int, addStake *big.Int, owner common.Address, name string, tags []string) (*types.Transaction, error) {
	return _ModelRegistry.Contract.Register(&_ModelRegistry.TransactOpts, modelId, ipfsCID, fee, addStake, owner, name, tags)
}

// Register is a paid mutator transaction binding the contract method 0x9e4aaa05.
//
// Solidity: function register(bytes32 modelId, bytes32 ipfsCID, uint256 fee, uint256 addStake, address owner, string name, string[] tags) returns()
func (_ModelRegistry *ModelRegistryTransactorSession) Register(modelId [32]byte, ipfsCID [32]byte, fee *big.Int, addStake *big.Int, owner common.Address, name string, tags []string) (*types.Transaction, error) {
	return _ModelRegistry.Contract.Register(&_ModelRegistry.TransactOpts, modelId, ipfsCID, fee, addStake, owner, name, tags)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ModelRegistry *ModelRegistryTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ModelRegistry.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ModelRegistry *ModelRegistrySession) RenounceOwnership() (*types.Transaction, error) {
	return _ModelRegistry.Contract.RenounceOwnership(&_ModelRegistry.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ModelRegistry *ModelRegistryTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ModelRegistry.Contract.RenounceOwnership(&_ModelRegistry.TransactOpts)
}

// SetMinStake is a paid mutator transaction binding the contract method 0x8c80fd90.
//
// Solidity: function setMinStake(uint256 _minStake) returns()
func (_ModelRegistry *ModelRegistryTransactor) SetMinStake(opts *bind.TransactOpts, _minStake *big.Int) (*types.Transaction, error) {
	return _ModelRegistry.contract.Transact(opts, "setMinStake", _minStake)
}

// SetMinStake is a paid mutator transaction binding the contract method 0x8c80fd90.
//
// Solidity: function setMinStake(uint256 _minStake) returns()
func (_ModelRegistry *ModelRegistrySession) SetMinStake(_minStake *big.Int) (*types.Transaction, error) {
	return _ModelRegistry.Contract.SetMinStake(&_ModelRegistry.TransactOpts, _minStake)
}

// SetMinStake is a paid mutator transaction binding the contract method 0x8c80fd90.
//
// Solidity: function setMinStake(uint256 _minStake) returns()
func (_ModelRegistry *ModelRegistryTransactorSession) SetMinStake(_minStake *big.Int) (*types.Transaction, error) {
	return _ModelRegistry.Contract.SetMinStake(&_ModelRegistry.TransactOpts, _minStake)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ModelRegistry *ModelRegistryTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ModelRegistry.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ModelRegistry *ModelRegistrySession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ModelRegistry.Contract.TransferOwnership(&_ModelRegistry.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ModelRegistry *ModelRegistryTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ModelRegistry.Contract.TransferOwnership(&_ModelRegistry.TransactOpts, newOwner)
}

// ModelRegistryDeregisteredIterator is returned from FilterDeregistered and is used to iterate over the raw logs and unpacked data for Deregistered events raised by the ModelRegistry contract.
type ModelRegistryDeregisteredIterator struct {
	Event *ModelRegistryDeregistered // Event containing the contract specifics and raw log

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
func (it *ModelRegistryDeregisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ModelRegistryDeregistered)
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
		it.Event = new(ModelRegistryDeregistered)
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
func (it *ModelRegistryDeregisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ModelRegistryDeregisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ModelRegistryDeregistered represents a Deregistered event raised by the ModelRegistry contract.
type ModelRegistryDeregistered struct {
	Owner   common.Address
	ModelId [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterDeregistered is a free log retrieval operation binding the contract event 0xcc0d8d49cb9a5a3eebc1e1f0b607b8aad83fbe9618e448df79b8a0cc33319472.
//
// Solidity: event Deregistered(address indexed owner, bytes32 indexed modelId)
func (_ModelRegistry *ModelRegistryFilterer) FilterDeregistered(opts *bind.FilterOpts, owner []common.Address, modelId [][32]byte) (*ModelRegistryDeregisteredIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var modelIdRule []interface{}
	for _, modelIdItem := range modelId {
		modelIdRule = append(modelIdRule, modelIdItem)
	}

	logs, sub, err := _ModelRegistry.contract.FilterLogs(opts, "Deregistered", ownerRule, modelIdRule)
	if err != nil {
		return nil, err
	}
	return &ModelRegistryDeregisteredIterator{contract: _ModelRegistry.contract, event: "Deregistered", logs: logs, sub: sub}, nil
}

// WatchDeregistered is a free log subscription operation binding the contract event 0xcc0d8d49cb9a5a3eebc1e1f0b607b8aad83fbe9618e448df79b8a0cc33319472.
//
// Solidity: event Deregistered(address indexed owner, bytes32 indexed modelId)
func (_ModelRegistry *ModelRegistryFilterer) WatchDeregistered(opts *bind.WatchOpts, sink chan<- *ModelRegistryDeregistered, owner []common.Address, modelId [][32]byte) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var modelIdRule []interface{}
	for _, modelIdItem := range modelId {
		modelIdRule = append(modelIdRule, modelIdItem)
	}

	logs, sub, err := _ModelRegistry.contract.WatchLogs(opts, "Deregistered", ownerRule, modelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ModelRegistryDeregistered)
				if err := _ModelRegistry.contract.UnpackLog(event, "Deregistered", log); err != nil {
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
// Solidity: event Deregistered(address indexed owner, bytes32 indexed modelId)
func (_ModelRegistry *ModelRegistryFilterer) ParseDeregistered(log types.Log) (*ModelRegistryDeregistered, error) {
	event := new(ModelRegistryDeregistered)
	if err := _ModelRegistry.contract.UnpackLog(event, "Deregistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ModelRegistryInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the ModelRegistry contract.
type ModelRegistryInitializedIterator struct {
	Event *ModelRegistryInitialized // Event containing the contract specifics and raw log

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
func (it *ModelRegistryInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ModelRegistryInitialized)
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
		it.Event = new(ModelRegistryInitialized)
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
func (it *ModelRegistryInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ModelRegistryInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ModelRegistryInitialized represents a Initialized event raised by the ModelRegistry contract.
type ModelRegistryInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ModelRegistry *ModelRegistryFilterer) FilterInitialized(opts *bind.FilterOpts) (*ModelRegistryInitializedIterator, error) {

	logs, sub, err := _ModelRegistry.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ModelRegistryInitializedIterator{contract: _ModelRegistry.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ModelRegistry *ModelRegistryFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ModelRegistryInitialized) (event.Subscription, error) {

	logs, sub, err := _ModelRegistry.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ModelRegistryInitialized)
				if err := _ModelRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_ModelRegistry *ModelRegistryFilterer) ParseInitialized(log types.Log) (*ModelRegistryInitialized, error) {
	event := new(ModelRegistryInitialized)
	if err := _ModelRegistry.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ModelRegistryMinStakeUpdatedIterator is returned from FilterMinStakeUpdated and is used to iterate over the raw logs and unpacked data for MinStakeUpdated events raised by the ModelRegistry contract.
type ModelRegistryMinStakeUpdatedIterator struct {
	Event *ModelRegistryMinStakeUpdated // Event containing the contract specifics and raw log

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
func (it *ModelRegistryMinStakeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ModelRegistryMinStakeUpdated)
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
		it.Event = new(ModelRegistryMinStakeUpdated)
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
func (it *ModelRegistryMinStakeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ModelRegistryMinStakeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ModelRegistryMinStakeUpdated represents a MinStakeUpdated event raised by the ModelRegistry contract.
type ModelRegistryMinStakeUpdated struct {
	NewStake *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterMinStakeUpdated is a free log retrieval operation binding the contract event 0x47ab46f2c8d4258304a2f5551c1cbdb6981be49631365d1ba7191288a73f39ef.
//
// Solidity: event MinStakeUpdated(uint256 newStake)
func (_ModelRegistry *ModelRegistryFilterer) FilterMinStakeUpdated(opts *bind.FilterOpts) (*ModelRegistryMinStakeUpdatedIterator, error) {

	logs, sub, err := _ModelRegistry.contract.FilterLogs(opts, "MinStakeUpdated")
	if err != nil {
		return nil, err
	}
	return &ModelRegistryMinStakeUpdatedIterator{contract: _ModelRegistry.contract, event: "MinStakeUpdated", logs: logs, sub: sub}, nil
}

// WatchMinStakeUpdated is a free log subscription operation binding the contract event 0x47ab46f2c8d4258304a2f5551c1cbdb6981be49631365d1ba7191288a73f39ef.
//
// Solidity: event MinStakeUpdated(uint256 newStake)
func (_ModelRegistry *ModelRegistryFilterer) WatchMinStakeUpdated(opts *bind.WatchOpts, sink chan<- *ModelRegistryMinStakeUpdated) (event.Subscription, error) {

	logs, sub, err := _ModelRegistry.contract.WatchLogs(opts, "MinStakeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ModelRegistryMinStakeUpdated)
				if err := _ModelRegistry.contract.UnpackLog(event, "MinStakeUpdated", log); err != nil {
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
func (_ModelRegistry *ModelRegistryFilterer) ParseMinStakeUpdated(log types.Log) (*ModelRegistryMinStakeUpdated, error) {
	event := new(ModelRegistryMinStakeUpdated)
	if err := _ModelRegistry.contract.UnpackLog(event, "MinStakeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ModelRegistryOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ModelRegistry contract.
type ModelRegistryOwnershipTransferredIterator struct {
	Event *ModelRegistryOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *ModelRegistryOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ModelRegistryOwnershipTransferred)
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
		it.Event = new(ModelRegistryOwnershipTransferred)
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
func (it *ModelRegistryOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ModelRegistryOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ModelRegistryOwnershipTransferred represents a OwnershipTransferred event raised by the ModelRegistry contract.
type ModelRegistryOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ModelRegistry *ModelRegistryFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ModelRegistryOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ModelRegistry.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ModelRegistryOwnershipTransferredIterator{contract: _ModelRegistry.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ModelRegistry *ModelRegistryFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ModelRegistryOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ModelRegistry.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ModelRegistryOwnershipTransferred)
				if err := _ModelRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_ModelRegistry *ModelRegistryFilterer) ParseOwnershipTransferred(log types.Log) (*ModelRegistryOwnershipTransferred, error) {
	event := new(ModelRegistryOwnershipTransferred)
	if err := _ModelRegistry.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ModelRegistryRegisteredUpdatedIterator is returned from FilterRegisteredUpdated and is used to iterate over the raw logs and unpacked data for RegisteredUpdated events raised by the ModelRegistry contract.
type ModelRegistryRegisteredUpdatedIterator struct {
	Event *ModelRegistryRegisteredUpdated // Event containing the contract specifics and raw log

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
func (it *ModelRegistryRegisteredUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ModelRegistryRegisteredUpdated)
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
		it.Event = new(ModelRegistryRegisteredUpdated)
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
func (it *ModelRegistryRegisteredUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ModelRegistryRegisteredUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ModelRegistryRegisteredUpdated represents a RegisteredUpdated event raised by the ModelRegistry contract.
type ModelRegistryRegisteredUpdated struct {
	Owner   common.Address
	ModelId [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRegisteredUpdated is a free log retrieval operation binding the contract event 0x4dffc4e92367641816479c647d1ff3f03202f6c73b97340cba51a7b577c55804.
//
// Solidity: event RegisteredUpdated(address indexed owner, bytes32 indexed modelId)
func (_ModelRegistry *ModelRegistryFilterer) FilterRegisteredUpdated(opts *bind.FilterOpts, owner []common.Address, modelId [][32]byte) (*ModelRegistryRegisteredUpdatedIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var modelIdRule []interface{}
	for _, modelIdItem := range modelId {
		modelIdRule = append(modelIdRule, modelIdItem)
	}

	logs, sub, err := _ModelRegistry.contract.FilterLogs(opts, "RegisteredUpdated", ownerRule, modelIdRule)
	if err != nil {
		return nil, err
	}
	return &ModelRegistryRegisteredUpdatedIterator{contract: _ModelRegistry.contract, event: "RegisteredUpdated", logs: logs, sub: sub}, nil
}

// WatchRegisteredUpdated is a free log subscription operation binding the contract event 0x4dffc4e92367641816479c647d1ff3f03202f6c73b97340cba51a7b577c55804.
//
// Solidity: event RegisteredUpdated(address indexed owner, bytes32 indexed modelId)
func (_ModelRegistry *ModelRegistryFilterer) WatchRegisteredUpdated(opts *bind.WatchOpts, sink chan<- *ModelRegistryRegisteredUpdated, owner []common.Address, modelId [][32]byte) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var modelIdRule []interface{}
	for _, modelIdItem := range modelId {
		modelIdRule = append(modelIdRule, modelIdItem)
	}

	logs, sub, err := _ModelRegistry.contract.WatchLogs(opts, "RegisteredUpdated", ownerRule, modelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ModelRegistryRegisteredUpdated)
				if err := _ModelRegistry.contract.UnpackLog(event, "RegisteredUpdated", log); err != nil {
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
// Solidity: event RegisteredUpdated(address indexed owner, bytes32 indexed modelId)
func (_ModelRegistry *ModelRegistryFilterer) ParseRegisteredUpdated(log types.Log) (*ModelRegistryRegisteredUpdated, error) {
	event := new(ModelRegistryRegisteredUpdated)
	if err := _ModelRegistry.contract.UnpackLog(event, "RegisteredUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
