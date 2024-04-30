// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package marketplace

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

// MarketplaceBid is an auto generated low-level Go binding around an user-defined struct.
type MarketplaceBid struct {
	Provider     common.Address
	ModelAgentId [32]byte
	Amount       *big.Int
	Nonce        *big.Int
	CreatedAt    *big.Int
	DeletedAt    *big.Int
}

// MarketplaceMetaData contains all meta data concerning the Marketplace contract.
var MarketplaceMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"ActiveBidNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"KeyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"KeyNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ModelOrAgentNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotEnoughWithdrawableBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSenderOrOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroKey\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"BidDeleted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"BidPosted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"modelFee\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"agentFee\",\"type\":\"uint256\"}],\"name\":\"FeeUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"agentBidFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"}],\"name\":\"deleteModelAgentBid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"}],\"name\":\"getActiveBidsByModelAgent\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deletedAt\",\"type\":\"uint256\"}],\"internalType\":\"structMarketplace.Bid[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"getActiveBidsByProvider\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deletedAt\",\"type\":\"uint256\"}],\"internalType\":\"structMarketplace.Bid[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"limit\",\"type\":\"uint8\"}],\"name\":\"getBidsByModelAgent\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deletedAt\",\"type\":\"uint256\"}],\"internalType\":\"structMarketplace.Bid[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"limit\",\"type\":\"uint8\"}],\"name\":\"getBidsByProvider\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deletedAt\",\"type\":\"uint256\"}],\"internalType\":\"structMarketplace.Bid[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_token\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"modelRegistryAddr\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"providerRegistryAddr\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"map\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"createdAt\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"deletedAt\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"modelBidFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"modelRegistry\",\"outputs\":[{\"internalType\":\"contractModelRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"providerAddr\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"postModelBid\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"providerModelAgentNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"providerRegistry\",\"outputs\":[{\"internalType\":\"contractProviderRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"modelFee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"agentFee\",\"type\":\"uint256\"}],\"name\":\"setBidFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"token\",\"outputs\":[{\"internalType\":\"contractERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// MarketplaceABI is the input ABI used to generate the binding from.
// Deprecated: Use MarketplaceMetaData.ABI instead.
var MarketplaceABI = MarketplaceMetaData.ABI

// Marketplace is an auto generated Go binding around an Ethereum contract.
type Marketplace struct {
	MarketplaceCaller     // Read-only binding to the contract
	MarketplaceTransactor // Write-only binding to the contract
	MarketplaceFilterer   // Log filterer for contract events
}

// MarketplaceCaller is an auto generated read-only Go binding around an Ethereum contract.
type MarketplaceCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MarketplaceTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MarketplaceTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MarketplaceFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MarketplaceFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MarketplaceSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MarketplaceSession struct {
	Contract     *Marketplace      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MarketplaceCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MarketplaceCallerSession struct {
	Contract *MarketplaceCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// MarketplaceTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MarketplaceTransactorSession struct {
	Contract     *MarketplaceTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// MarketplaceRaw is an auto generated low-level Go binding around an Ethereum contract.
type MarketplaceRaw struct {
	Contract *Marketplace // Generic contract binding to access the raw methods on
}

// MarketplaceCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MarketplaceCallerRaw struct {
	Contract *MarketplaceCaller // Generic read-only contract binding to access the raw methods on
}

// MarketplaceTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MarketplaceTransactorRaw struct {
	Contract *MarketplaceTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMarketplace creates a new instance of Marketplace, bound to a specific deployed contract.
func NewMarketplace(address common.Address, backend bind.ContractBackend) (*Marketplace, error) {
	contract, err := bindMarketplace(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Marketplace{MarketplaceCaller: MarketplaceCaller{contract: contract}, MarketplaceTransactor: MarketplaceTransactor{contract: contract}, MarketplaceFilterer: MarketplaceFilterer{contract: contract}}, nil
}

// NewMarketplaceCaller creates a new read-only instance of Marketplace, bound to a specific deployed contract.
func NewMarketplaceCaller(address common.Address, caller bind.ContractCaller) (*MarketplaceCaller, error) {
	contract, err := bindMarketplace(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MarketplaceCaller{contract: contract}, nil
}

// NewMarketplaceTransactor creates a new write-only instance of Marketplace, bound to a specific deployed contract.
func NewMarketplaceTransactor(address common.Address, transactor bind.ContractTransactor) (*MarketplaceTransactor, error) {
	contract, err := bindMarketplace(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MarketplaceTransactor{contract: contract}, nil
}

// NewMarketplaceFilterer creates a new log filterer instance of Marketplace, bound to a specific deployed contract.
func NewMarketplaceFilterer(address common.Address, filterer bind.ContractFilterer) (*MarketplaceFilterer, error) {
	contract, err := bindMarketplace(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MarketplaceFilterer{contract: contract}, nil
}

// bindMarketplace binds a generic wrapper to an already deployed contract.
func bindMarketplace(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := MarketplaceMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Marketplace *MarketplaceRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Marketplace.Contract.MarketplaceCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Marketplace *MarketplaceRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Marketplace.Contract.MarketplaceTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Marketplace *MarketplaceRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Marketplace.Contract.MarketplaceTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Marketplace *MarketplaceCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Marketplace.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Marketplace *MarketplaceTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Marketplace.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Marketplace *MarketplaceTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Marketplace.Contract.contract.Transact(opts, method, params...)
}

// AgentBidFee is a free data retrieval call binding the contract method 0x113d91c9.
//
// Solidity: function agentBidFee() view returns(uint256)
func (_Marketplace *MarketplaceCaller) AgentBidFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "agentBidFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// AgentBidFee is a free data retrieval call binding the contract method 0x113d91c9.
//
// Solidity: function agentBidFee() view returns(uint256)
func (_Marketplace *MarketplaceSession) AgentBidFee() (*big.Int, error) {
	return _Marketplace.Contract.AgentBidFee(&_Marketplace.CallOpts)
}

// AgentBidFee is a free data retrieval call binding the contract method 0x113d91c9.
//
// Solidity: function agentBidFee() view returns(uint256)
func (_Marketplace *MarketplaceCallerSession) AgentBidFee() (*big.Int, error) {
	return _Marketplace.Contract.AgentBidFee(&_Marketplace.CallOpts)
}

// GetActiveBidsByModelAgent is a free data retrieval call binding the contract method 0x873d94d5.
//
// Solidity: function getActiveBidsByModelAgent(bytes32 modelAgentId) view returns((address,bytes32,uint256,uint256,uint256,uint256)[])
func (_Marketplace *MarketplaceCaller) GetActiveBidsByModelAgent(opts *bind.CallOpts, modelAgentId [32]byte) ([]MarketplaceBid, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getActiveBidsByModelAgent", modelAgentId)

	if err != nil {
		return *new([]MarketplaceBid), err
	}

	out0 := *abi.ConvertType(out[0], new([]MarketplaceBid)).(*[]MarketplaceBid)

	return out0, err

}

// GetActiveBidsByModelAgent is a free data retrieval call binding the contract method 0x873d94d5.
//
// Solidity: function getActiveBidsByModelAgent(bytes32 modelAgentId) view returns((address,bytes32,uint256,uint256,uint256,uint256)[])
func (_Marketplace *MarketplaceSession) GetActiveBidsByModelAgent(modelAgentId [32]byte) ([]MarketplaceBid, error) {
	return _Marketplace.Contract.GetActiveBidsByModelAgent(&_Marketplace.CallOpts, modelAgentId)
}

// GetActiveBidsByModelAgent is a free data retrieval call binding the contract method 0x873d94d5.
//
// Solidity: function getActiveBidsByModelAgent(bytes32 modelAgentId) view returns((address,bytes32,uint256,uint256,uint256,uint256)[])
func (_Marketplace *MarketplaceCallerSession) GetActiveBidsByModelAgent(modelAgentId [32]byte) ([]MarketplaceBid, error) {
	return _Marketplace.Contract.GetActiveBidsByModelAgent(&_Marketplace.CallOpts, modelAgentId)
}

// GetActiveBidsByProvider is a free data retrieval call binding the contract method 0x9fdaffd0.
//
// Solidity: function getActiveBidsByProvider(address provider) view returns((address,bytes32,uint256,uint256,uint256,uint256)[])
func (_Marketplace *MarketplaceCaller) GetActiveBidsByProvider(opts *bind.CallOpts, provider common.Address) ([]MarketplaceBid, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getActiveBidsByProvider", provider)

	if err != nil {
		return *new([]MarketplaceBid), err
	}

	out0 := *abi.ConvertType(out[0], new([]MarketplaceBid)).(*[]MarketplaceBid)

	return out0, err

}

// GetActiveBidsByProvider is a free data retrieval call binding the contract method 0x9fdaffd0.
//
// Solidity: function getActiveBidsByProvider(address provider) view returns((address,bytes32,uint256,uint256,uint256,uint256)[])
func (_Marketplace *MarketplaceSession) GetActiveBidsByProvider(provider common.Address) ([]MarketplaceBid, error) {
	return _Marketplace.Contract.GetActiveBidsByProvider(&_Marketplace.CallOpts, provider)
}

// GetActiveBidsByProvider is a free data retrieval call binding the contract method 0x9fdaffd0.
//
// Solidity: function getActiveBidsByProvider(address provider) view returns((address,bytes32,uint256,uint256,uint256,uint256)[])
func (_Marketplace *MarketplaceCallerSession) GetActiveBidsByProvider(provider common.Address) ([]MarketplaceBid, error) {
	return _Marketplace.Contract.GetActiveBidsByProvider(&_Marketplace.CallOpts, provider)
}

// GetBidsByModelAgent is a free data retrieval call binding the contract method 0xa87665ec.
//
// Solidity: function getBidsByModelAgent(bytes32 modelAgentId, uint256 offset, uint8 limit) view returns((address,bytes32,uint256,uint256,uint256,uint256)[])
func (_Marketplace *MarketplaceCaller) GetBidsByModelAgent(opts *bind.CallOpts, modelAgentId [32]byte, offset *big.Int, limit uint8) ([]MarketplaceBid, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getBidsByModelAgent", modelAgentId, offset, limit)

	if err != nil {
		return *new([]MarketplaceBid), err
	}

	out0 := *abi.ConvertType(out[0], new([]MarketplaceBid)).(*[]MarketplaceBid)

	return out0, err

}

// GetBidsByModelAgent is a free data retrieval call binding the contract method 0xa87665ec.
//
// Solidity: function getBidsByModelAgent(bytes32 modelAgentId, uint256 offset, uint8 limit) view returns((address,bytes32,uint256,uint256,uint256,uint256)[])
func (_Marketplace *MarketplaceSession) GetBidsByModelAgent(modelAgentId [32]byte, offset *big.Int, limit uint8) ([]MarketplaceBid, error) {
	return _Marketplace.Contract.GetBidsByModelAgent(&_Marketplace.CallOpts, modelAgentId, offset, limit)
}

// GetBidsByModelAgent is a free data retrieval call binding the contract method 0xa87665ec.
//
// Solidity: function getBidsByModelAgent(bytes32 modelAgentId, uint256 offset, uint8 limit) view returns((address,bytes32,uint256,uint256,uint256,uint256)[])
func (_Marketplace *MarketplaceCallerSession) GetBidsByModelAgent(modelAgentId [32]byte, offset *big.Int, limit uint8) ([]MarketplaceBid, error) {
	return _Marketplace.Contract.GetBidsByModelAgent(&_Marketplace.CallOpts, modelAgentId, offset, limit)
}

// GetBidsByProvider is a free data retrieval call binding the contract method 0x2f817685.
//
// Solidity: function getBidsByProvider(address provider, uint256 offset, uint8 limit) view returns((address,bytes32,uint256,uint256,uint256,uint256)[])
func (_Marketplace *MarketplaceCaller) GetBidsByProvider(opts *bind.CallOpts, provider common.Address, offset *big.Int, limit uint8) ([]MarketplaceBid, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getBidsByProvider", provider, offset, limit)

	if err != nil {
		return *new([]MarketplaceBid), err
	}

	out0 := *abi.ConvertType(out[0], new([]MarketplaceBid)).(*[]MarketplaceBid)

	return out0, err

}

// GetBidsByProvider is a free data retrieval call binding the contract method 0x2f817685.
//
// Solidity: function getBidsByProvider(address provider, uint256 offset, uint8 limit) view returns((address,bytes32,uint256,uint256,uint256,uint256)[])
func (_Marketplace *MarketplaceSession) GetBidsByProvider(provider common.Address, offset *big.Int, limit uint8) ([]MarketplaceBid, error) {
	return _Marketplace.Contract.GetBidsByProvider(&_Marketplace.CallOpts, provider, offset, limit)
}

// GetBidsByProvider is a free data retrieval call binding the contract method 0x2f817685.
//
// Solidity: function getBidsByProvider(address provider, uint256 offset, uint8 limit) view returns((address,bytes32,uint256,uint256,uint256,uint256)[])
func (_Marketplace *MarketplaceCallerSession) GetBidsByProvider(provider common.Address, offset *big.Int, limit uint8) ([]MarketplaceBid, error) {
	return _Marketplace.Contract.GetBidsByProvider(&_Marketplace.CallOpts, provider, offset, limit)
}

// Map is a free data retrieval call binding the contract method 0x0ae186a8.
//
// Solidity: function map(bytes32 ) view returns(address provider, bytes32 modelAgentId, uint256 amount, uint256 nonce, uint256 createdAt, uint256 deletedAt)
func (_Marketplace *MarketplaceCaller) Map(opts *bind.CallOpts, arg0 [32]byte) (struct {
	Provider     common.Address
	ModelAgentId [32]byte
	Amount       *big.Int
	Nonce        *big.Int
	CreatedAt    *big.Int
	DeletedAt    *big.Int
}, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "map", arg0)

	outstruct := new(struct {
		Provider     common.Address
		ModelAgentId [32]byte
		Amount       *big.Int
		Nonce        *big.Int
		CreatedAt    *big.Int
		DeletedAt    *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Provider = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.ModelAgentId = *abi.ConvertType(out[1], new([32]byte)).(*[32]byte)
	outstruct.Amount = *abi.ConvertType(out[2], new(*big.Int)).(**big.Int)
	outstruct.Nonce = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.CreatedAt = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.DeletedAt = *abi.ConvertType(out[5], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// Map is a free data retrieval call binding the contract method 0x0ae186a8.
//
// Solidity: function map(bytes32 ) view returns(address provider, bytes32 modelAgentId, uint256 amount, uint256 nonce, uint256 createdAt, uint256 deletedAt)
func (_Marketplace *MarketplaceSession) Map(arg0 [32]byte) (struct {
	Provider     common.Address
	ModelAgentId [32]byte
	Amount       *big.Int
	Nonce        *big.Int
	CreatedAt    *big.Int
	DeletedAt    *big.Int
}, error) {
	return _Marketplace.Contract.Map(&_Marketplace.CallOpts, arg0)
}

// Map is a free data retrieval call binding the contract method 0x0ae186a8.
//
// Solidity: function map(bytes32 ) view returns(address provider, bytes32 modelAgentId, uint256 amount, uint256 nonce, uint256 createdAt, uint256 deletedAt)
func (_Marketplace *MarketplaceCallerSession) Map(arg0 [32]byte) (struct {
	Provider     common.Address
	ModelAgentId [32]byte
	Amount       *big.Int
	Nonce        *big.Int
	CreatedAt    *big.Int
	DeletedAt    *big.Int
}, error) {
	return _Marketplace.Contract.Map(&_Marketplace.CallOpts, arg0)
}

// ModelBidFee is a free data retrieval call binding the contract method 0x7c785e73.
//
// Solidity: function modelBidFee() view returns(uint256)
func (_Marketplace *MarketplaceCaller) ModelBidFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "modelBidFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ModelBidFee is a free data retrieval call binding the contract method 0x7c785e73.
//
// Solidity: function modelBidFee() view returns(uint256)
func (_Marketplace *MarketplaceSession) ModelBidFee() (*big.Int, error) {
	return _Marketplace.Contract.ModelBidFee(&_Marketplace.CallOpts)
}

// ModelBidFee is a free data retrieval call binding the contract method 0x7c785e73.
//
// Solidity: function modelBidFee() view returns(uint256)
func (_Marketplace *MarketplaceCallerSession) ModelBidFee() (*big.Int, error) {
	return _Marketplace.Contract.ModelBidFee(&_Marketplace.CallOpts)
}

// ModelRegistry is a free data retrieval call binding the contract method 0x2ba36f78.
//
// Solidity: function modelRegistry() view returns(address)
func (_Marketplace *MarketplaceCaller) ModelRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "modelRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ModelRegistry is a free data retrieval call binding the contract method 0x2ba36f78.
//
// Solidity: function modelRegistry() view returns(address)
func (_Marketplace *MarketplaceSession) ModelRegistry() (common.Address, error) {
	return _Marketplace.Contract.ModelRegistry(&_Marketplace.CallOpts)
}

// ModelRegistry is a free data retrieval call binding the contract method 0x2ba36f78.
//
// Solidity: function modelRegistry() view returns(address)
func (_Marketplace *MarketplaceCallerSession) ModelRegistry() (common.Address, error) {
	return _Marketplace.Contract.ModelRegistry(&_Marketplace.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Marketplace *MarketplaceCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Marketplace *MarketplaceSession) Owner() (common.Address, error) {
	return _Marketplace.Contract.Owner(&_Marketplace.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Marketplace *MarketplaceCallerSession) Owner() (common.Address, error) {
	return _Marketplace.Contract.Owner(&_Marketplace.CallOpts)
}

// ProviderModelAgentNonce is a free data retrieval call binding the contract method 0x93ec6bc5.
//
// Solidity: function providerModelAgentNonce(bytes32 ) view returns(uint256)
func (_Marketplace *MarketplaceCaller) ProviderModelAgentNonce(opts *bind.CallOpts, arg0 [32]byte) (*big.Int, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "providerModelAgentNonce", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProviderModelAgentNonce is a free data retrieval call binding the contract method 0x93ec6bc5.
//
// Solidity: function providerModelAgentNonce(bytes32 ) view returns(uint256)
func (_Marketplace *MarketplaceSession) ProviderModelAgentNonce(arg0 [32]byte) (*big.Int, error) {
	return _Marketplace.Contract.ProviderModelAgentNonce(&_Marketplace.CallOpts, arg0)
}

// ProviderModelAgentNonce is a free data retrieval call binding the contract method 0x93ec6bc5.
//
// Solidity: function providerModelAgentNonce(bytes32 ) view returns(uint256)
func (_Marketplace *MarketplaceCallerSession) ProviderModelAgentNonce(arg0 [32]byte) (*big.Int, error) {
	return _Marketplace.Contract.ProviderModelAgentNonce(&_Marketplace.CallOpts, arg0)
}

// ProviderRegistry is a free data retrieval call binding the contract method 0x545921d9.
//
// Solidity: function providerRegistry() view returns(address)
func (_Marketplace *MarketplaceCaller) ProviderRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "providerRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// ProviderRegistry is a free data retrieval call binding the contract method 0x545921d9.
//
// Solidity: function providerRegistry() view returns(address)
func (_Marketplace *MarketplaceSession) ProviderRegistry() (common.Address, error) {
	return _Marketplace.Contract.ProviderRegistry(&_Marketplace.CallOpts)
}

// ProviderRegistry is a free data retrieval call binding the contract method 0x545921d9.
//
// Solidity: function providerRegistry() view returns(address)
func (_Marketplace *MarketplaceCallerSession) ProviderRegistry() (common.Address, error) {
	return _Marketplace.Contract.ProviderRegistry(&_Marketplace.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Marketplace *MarketplaceCaller) Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Marketplace *MarketplaceSession) Token() (common.Address, error) {
	return _Marketplace.Contract.Token(&_Marketplace.CallOpts)
}

// Token is a free data retrieval call binding the contract method 0xfc0c546a.
//
// Solidity: function token() view returns(address)
func (_Marketplace *MarketplaceCallerSession) Token() (common.Address, error) {
	return _Marketplace.Contract.Token(&_Marketplace.CallOpts)
}

// DeleteModelAgentBid is a paid mutator transaction binding the contract method 0x42856b75.
//
// Solidity: function deleteModelAgentBid(bytes32 bidId) returns()
func (_Marketplace *MarketplaceTransactor) DeleteModelAgentBid(opts *bind.TransactOpts, bidId [32]byte) (*types.Transaction, error) {
	return _Marketplace.contract.Transact(opts, "deleteModelAgentBid", bidId)
}

// DeleteModelAgentBid is a paid mutator transaction binding the contract method 0x42856b75.
//
// Solidity: function deleteModelAgentBid(bytes32 bidId) returns()
func (_Marketplace *MarketplaceSession) DeleteModelAgentBid(bidId [32]byte) (*types.Transaction, error) {
	return _Marketplace.Contract.DeleteModelAgentBid(&_Marketplace.TransactOpts, bidId)
}

// DeleteModelAgentBid is a paid mutator transaction binding the contract method 0x42856b75.
//
// Solidity: function deleteModelAgentBid(bytes32 bidId) returns()
func (_Marketplace *MarketplaceTransactorSession) DeleteModelAgentBid(bidId [32]byte) (*types.Transaction, error) {
	return _Marketplace.Contract.DeleteModelAgentBid(&_Marketplace.TransactOpts, bidId)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _token, address modelRegistryAddr, address providerRegistryAddr) returns()
func (_Marketplace *MarketplaceTransactor) Initialize(opts *bind.TransactOpts, _token common.Address, modelRegistryAddr common.Address, providerRegistryAddr common.Address) (*types.Transaction, error) {
	return _Marketplace.contract.Transact(opts, "initialize", _token, modelRegistryAddr, providerRegistryAddr)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _token, address modelRegistryAddr, address providerRegistryAddr) returns()
func (_Marketplace *MarketplaceSession) Initialize(_token common.Address, modelRegistryAddr common.Address, providerRegistryAddr common.Address) (*types.Transaction, error) {
	return _Marketplace.Contract.Initialize(&_Marketplace.TransactOpts, _token, modelRegistryAddr, providerRegistryAddr)
}

// Initialize is a paid mutator transaction binding the contract method 0xc0c53b8b.
//
// Solidity: function initialize(address _token, address modelRegistryAddr, address providerRegistryAddr) returns()
func (_Marketplace *MarketplaceTransactorSession) Initialize(_token common.Address, modelRegistryAddr common.Address, providerRegistryAddr common.Address) (*types.Transaction, error) {
	return _Marketplace.Contract.Initialize(&_Marketplace.TransactOpts, _token, modelRegistryAddr, providerRegistryAddr)
}

// PostModelBid is a paid mutator transaction binding the contract method 0xede96bb1.
//
// Solidity: function postModelBid(address providerAddr, bytes32 modelId, uint256 amount) returns(bytes32 bidId)
func (_Marketplace *MarketplaceTransactor) PostModelBid(opts *bind.TransactOpts, providerAddr common.Address, modelId [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Marketplace.contract.Transact(opts, "postModelBid", providerAddr, modelId, amount)
}

// PostModelBid is a paid mutator transaction binding the contract method 0xede96bb1.
//
// Solidity: function postModelBid(address providerAddr, bytes32 modelId, uint256 amount) returns(bytes32 bidId)
func (_Marketplace *MarketplaceSession) PostModelBid(providerAddr common.Address, modelId [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.PostModelBid(&_Marketplace.TransactOpts, providerAddr, modelId, amount)
}

// PostModelBid is a paid mutator transaction binding the contract method 0xede96bb1.
//
// Solidity: function postModelBid(address providerAddr, bytes32 modelId, uint256 amount) returns(bytes32 bidId)
func (_Marketplace *MarketplaceTransactorSession) PostModelBid(providerAddr common.Address, modelId [32]byte, amount *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.PostModelBid(&_Marketplace.TransactOpts, providerAddr, modelId, amount)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Marketplace *MarketplaceTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Marketplace.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Marketplace *MarketplaceSession) RenounceOwnership() (*types.Transaction, error) {
	return _Marketplace.Contract.RenounceOwnership(&_Marketplace.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Marketplace *MarketplaceTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Marketplace.Contract.RenounceOwnership(&_Marketplace.TransactOpts)
}

// SetBidFee is a paid mutator transaction binding the contract method 0xcef0b7f4.
//
// Solidity: function setBidFee(uint256 modelFee, uint256 agentFee) returns()
func (_Marketplace *MarketplaceTransactor) SetBidFee(opts *bind.TransactOpts, modelFee *big.Int, agentFee *big.Int) (*types.Transaction, error) {
	return _Marketplace.contract.Transact(opts, "setBidFee", modelFee, agentFee)
}

// SetBidFee is a paid mutator transaction binding the contract method 0xcef0b7f4.
//
// Solidity: function setBidFee(uint256 modelFee, uint256 agentFee) returns()
func (_Marketplace *MarketplaceSession) SetBidFee(modelFee *big.Int, agentFee *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.SetBidFee(&_Marketplace.TransactOpts, modelFee, agentFee)
}

// SetBidFee is a paid mutator transaction binding the contract method 0xcef0b7f4.
//
// Solidity: function setBidFee(uint256 modelFee, uint256 agentFee) returns()
func (_Marketplace *MarketplaceTransactorSession) SetBidFee(modelFee *big.Int, agentFee *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.SetBidFee(&_Marketplace.TransactOpts, modelFee, agentFee)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Marketplace *MarketplaceTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Marketplace.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Marketplace *MarketplaceSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Marketplace.Contract.TransferOwnership(&_Marketplace.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Marketplace *MarketplaceTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Marketplace.Contract.TransferOwnership(&_Marketplace.TransactOpts, newOwner)
}

// Withdraw is a paid mutator transaction binding the contract method 0xf3fef3a3.
//
// Solidity: function withdraw(address addr, uint256 amount) returns()
func (_Marketplace *MarketplaceTransactor) Withdraw(opts *bind.TransactOpts, addr common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Marketplace.contract.Transact(opts, "withdraw", addr, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0xf3fef3a3.
//
// Solidity: function withdraw(address addr, uint256 amount) returns()
func (_Marketplace *MarketplaceSession) Withdraw(addr common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.Withdraw(&_Marketplace.TransactOpts, addr, amount)
}

// Withdraw is a paid mutator transaction binding the contract method 0xf3fef3a3.
//
// Solidity: function withdraw(address addr, uint256 amount) returns()
func (_Marketplace *MarketplaceTransactorSession) Withdraw(addr common.Address, amount *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.Withdraw(&_Marketplace.TransactOpts, addr, amount)
}

// MarketplaceBidDeletedIterator is returned from FilterBidDeleted and is used to iterate over the raw logs and unpacked data for BidDeleted events raised by the Marketplace contract.
type MarketplaceBidDeletedIterator struct {
	Event *MarketplaceBidDeleted // Event containing the contract specifics and raw log

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
func (it *MarketplaceBidDeletedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MarketplaceBidDeleted)
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
		it.Event = new(MarketplaceBidDeleted)
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
func (it *MarketplaceBidDeletedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MarketplaceBidDeletedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MarketplaceBidDeleted represents a BidDeleted event raised by the Marketplace contract.
type MarketplaceBidDeleted struct {
	Provider     common.Address
	ModelAgentId [32]byte
	Nonce        *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterBidDeleted is a free log retrieval operation binding the contract event 0x096f970f504563bca8ac4419b4299946965221e396c34aea149ac84947b9242f.
//
// Solidity: event BidDeleted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce)
func (_Marketplace *MarketplaceFilterer) FilterBidDeleted(opts *bind.FilterOpts, provider []common.Address, modelAgentId [][32]byte) (*MarketplaceBidDeletedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var modelAgentIdRule []interface{}
	for _, modelAgentIdItem := range modelAgentId {
		modelAgentIdRule = append(modelAgentIdRule, modelAgentIdItem)
	}

	logs, sub, err := _Marketplace.contract.FilterLogs(opts, "BidDeleted", providerRule, modelAgentIdRule)
	if err != nil {
		return nil, err
	}
	return &MarketplaceBidDeletedIterator{contract: _Marketplace.contract, event: "BidDeleted", logs: logs, sub: sub}, nil
}

// WatchBidDeleted is a free log subscription operation binding the contract event 0x096f970f504563bca8ac4419b4299946965221e396c34aea149ac84947b9242f.
//
// Solidity: event BidDeleted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce)
func (_Marketplace *MarketplaceFilterer) WatchBidDeleted(opts *bind.WatchOpts, sink chan<- *MarketplaceBidDeleted, provider []common.Address, modelAgentId [][32]byte) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var modelAgentIdRule []interface{}
	for _, modelAgentIdItem := range modelAgentId {
		modelAgentIdRule = append(modelAgentIdRule, modelAgentIdItem)
	}

	logs, sub, err := _Marketplace.contract.WatchLogs(opts, "BidDeleted", providerRule, modelAgentIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MarketplaceBidDeleted)
				if err := _Marketplace.contract.UnpackLog(event, "BidDeleted", log); err != nil {
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

// ParseBidDeleted is a log parse operation binding the contract event 0x096f970f504563bca8ac4419b4299946965221e396c34aea149ac84947b9242f.
//
// Solidity: event BidDeleted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce)
func (_Marketplace *MarketplaceFilterer) ParseBidDeleted(log types.Log) (*MarketplaceBidDeleted, error) {
	event := new(MarketplaceBidDeleted)
	if err := _Marketplace.contract.UnpackLog(event, "BidDeleted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MarketplaceBidPostedIterator is returned from FilterBidPosted and is used to iterate over the raw logs and unpacked data for BidPosted events raised by the Marketplace contract.
type MarketplaceBidPostedIterator struct {
	Event *MarketplaceBidPosted // Event containing the contract specifics and raw log

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
func (it *MarketplaceBidPostedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MarketplaceBidPosted)
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
		it.Event = new(MarketplaceBidPosted)
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
func (it *MarketplaceBidPostedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MarketplaceBidPostedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MarketplaceBidPosted represents a BidPosted event raised by the Marketplace contract.
type MarketplaceBidPosted struct {
	Provider     common.Address
	ModelAgentId [32]byte
	Nonce        *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterBidPosted is a free log retrieval operation binding the contract event 0xd138adff73af2621d26114cd9ee4f20dcd39ed78f9e0004215ed49aa22753ebe.
//
// Solidity: event BidPosted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce)
func (_Marketplace *MarketplaceFilterer) FilterBidPosted(opts *bind.FilterOpts, provider []common.Address, modelAgentId [][32]byte) (*MarketplaceBidPostedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var modelAgentIdRule []interface{}
	for _, modelAgentIdItem := range modelAgentId {
		modelAgentIdRule = append(modelAgentIdRule, modelAgentIdItem)
	}

	logs, sub, err := _Marketplace.contract.FilterLogs(opts, "BidPosted", providerRule, modelAgentIdRule)
	if err != nil {
		return nil, err
	}
	return &MarketplaceBidPostedIterator{contract: _Marketplace.contract, event: "BidPosted", logs: logs, sub: sub}, nil
}

// WatchBidPosted is a free log subscription operation binding the contract event 0xd138adff73af2621d26114cd9ee4f20dcd39ed78f9e0004215ed49aa22753ebe.
//
// Solidity: event BidPosted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce)
func (_Marketplace *MarketplaceFilterer) WatchBidPosted(opts *bind.WatchOpts, sink chan<- *MarketplaceBidPosted, provider []common.Address, modelAgentId [][32]byte) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var modelAgentIdRule []interface{}
	for _, modelAgentIdItem := range modelAgentId {
		modelAgentIdRule = append(modelAgentIdRule, modelAgentIdItem)
	}

	logs, sub, err := _Marketplace.contract.WatchLogs(opts, "BidPosted", providerRule, modelAgentIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MarketplaceBidPosted)
				if err := _Marketplace.contract.UnpackLog(event, "BidPosted", log); err != nil {
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

// ParseBidPosted is a log parse operation binding the contract event 0xd138adff73af2621d26114cd9ee4f20dcd39ed78f9e0004215ed49aa22753ebe.
//
// Solidity: event BidPosted(address indexed provider, bytes32 indexed modelAgentId, uint256 nonce)
func (_Marketplace *MarketplaceFilterer) ParseBidPosted(log types.Log) (*MarketplaceBidPosted, error) {
	event := new(MarketplaceBidPosted)
	if err := _Marketplace.contract.UnpackLog(event, "BidPosted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MarketplaceFeeUpdatedIterator is returned from FilterFeeUpdated and is used to iterate over the raw logs and unpacked data for FeeUpdated events raised by the Marketplace contract.
type MarketplaceFeeUpdatedIterator struct {
	Event *MarketplaceFeeUpdated // Event containing the contract specifics and raw log

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
func (it *MarketplaceFeeUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MarketplaceFeeUpdated)
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
		it.Event = new(MarketplaceFeeUpdated)
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
func (it *MarketplaceFeeUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MarketplaceFeeUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MarketplaceFeeUpdated represents a FeeUpdated event raised by the Marketplace contract.
type MarketplaceFeeUpdated struct {
	ModelFee *big.Int
	AgentFee *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterFeeUpdated is a free log retrieval operation binding the contract event 0x528d9479e9f9889a87a3c30c7f7ba537e5e59c4c85a37733b16e57c62df61302.
//
// Solidity: event FeeUpdated(uint256 modelFee, uint256 agentFee)
func (_Marketplace *MarketplaceFilterer) FilterFeeUpdated(opts *bind.FilterOpts) (*MarketplaceFeeUpdatedIterator, error) {

	logs, sub, err := _Marketplace.contract.FilterLogs(opts, "FeeUpdated")
	if err != nil {
		return nil, err
	}
	return &MarketplaceFeeUpdatedIterator{contract: _Marketplace.contract, event: "FeeUpdated", logs: logs, sub: sub}, nil
}

// WatchFeeUpdated is a free log subscription operation binding the contract event 0x528d9479e9f9889a87a3c30c7f7ba537e5e59c4c85a37733b16e57c62df61302.
//
// Solidity: event FeeUpdated(uint256 modelFee, uint256 agentFee)
func (_Marketplace *MarketplaceFilterer) WatchFeeUpdated(opts *bind.WatchOpts, sink chan<- *MarketplaceFeeUpdated) (event.Subscription, error) {

	logs, sub, err := _Marketplace.contract.WatchLogs(opts, "FeeUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MarketplaceFeeUpdated)
				if err := _Marketplace.contract.UnpackLog(event, "FeeUpdated", log); err != nil {
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

// ParseFeeUpdated is a log parse operation binding the contract event 0x528d9479e9f9889a87a3c30c7f7ba537e5e59c4c85a37733b16e57c62df61302.
//
// Solidity: event FeeUpdated(uint256 modelFee, uint256 agentFee)
func (_Marketplace *MarketplaceFilterer) ParseFeeUpdated(log types.Log) (*MarketplaceFeeUpdated, error) {
	event := new(MarketplaceFeeUpdated)
	if err := _Marketplace.contract.UnpackLog(event, "FeeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MarketplaceInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Marketplace contract.
type MarketplaceInitializedIterator struct {
	Event *MarketplaceInitialized // Event containing the contract specifics and raw log

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
func (it *MarketplaceInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MarketplaceInitialized)
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
		it.Event = new(MarketplaceInitialized)
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
func (it *MarketplaceInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MarketplaceInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MarketplaceInitialized represents a Initialized event raised by the Marketplace contract.
type MarketplaceInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Marketplace *MarketplaceFilterer) FilterInitialized(opts *bind.FilterOpts) (*MarketplaceInitializedIterator, error) {

	logs, sub, err := _Marketplace.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &MarketplaceInitializedIterator{contract: _Marketplace.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_Marketplace *MarketplaceFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *MarketplaceInitialized) (event.Subscription, error) {

	logs, sub, err := _Marketplace.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MarketplaceInitialized)
				if err := _Marketplace.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_Marketplace *MarketplaceFilterer) ParseInitialized(log types.Log) (*MarketplaceInitialized, error) {
	event := new(MarketplaceInitialized)
	if err := _Marketplace.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MarketplaceOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Marketplace contract.
type MarketplaceOwnershipTransferredIterator struct {
	Event *MarketplaceOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *MarketplaceOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MarketplaceOwnershipTransferred)
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
		it.Event = new(MarketplaceOwnershipTransferred)
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
func (it *MarketplaceOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MarketplaceOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MarketplaceOwnershipTransferred represents a OwnershipTransferred event raised by the Marketplace contract.
type MarketplaceOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Marketplace *MarketplaceFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*MarketplaceOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Marketplace.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &MarketplaceOwnershipTransferredIterator{contract: _Marketplace.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Marketplace *MarketplaceFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *MarketplaceOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Marketplace.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MarketplaceOwnershipTransferred)
				if err := _Marketplace.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_Marketplace *MarketplaceFilterer) ParseOwnershipTransferred(log types.Log) (*MarketplaceOwnershipTransferred, error) {
	event := new(MarketplaceOwnershipTransferred)
	if err := _Marketplace.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
