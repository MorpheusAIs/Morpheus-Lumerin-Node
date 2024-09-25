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

// Bid is an auto generated low-level Go binding around an user-defined struct.
type Bid struct {
	Provider       common.Address
	ModelAgentId   [32]byte
	PricePerSecond *big.Int
	Nonce          *big.Int
	CreatedAt      *big.Int
	DeletedAt      *big.Int
}

// LibSDSD is an auto generated low-level Go binding around an user-defined struct.
type LibSDSD struct {
	Mean  int64
	SqSum int64
}

// ModelStats is an auto generated low-level Go binding around an user-defined struct.
type ModelStats struct {
	TpsScaled1000 LibSDSD
	TtftMs        LibSDSD
	TotalDuration LibSDSD
	Count         uint32
}

// ProviderModelStats is an auto generated low-level Go binding around an user-defined struct.
type ProviderModelStats struct {
	TpsScaled1000 LibSDSD
	TtftMs        LibSDSD
	TotalDuration uint32
	SuccessCount  uint32
	TotalCount    uint32
}

// MarketplaceMetaData contains all meta data concerning the Marketplace contract.
var MarketplaceMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"ActiveBidNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"BidTaken\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"KeyExists\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"KeyNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ModelOrAgentNotFound\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_user\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_contractOwner\",\"type\":\"address\"}],\"name\":\"NotContractOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotEnoughBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotSenderOrOwner\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"PricePerSecondIsZero\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProviderNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ZeroKey\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"BidDeleted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"BidPosted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"bidFee\",\"type\":\"uint256\"}],\"name\":\"FeeUpdated\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"bidFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"}],\"name\":\"bidMap\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"deletedAt\",\"type\":\"uint128\"}],\"internalType\":\"structBid\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"}],\"name\":\"deleteModelAgentBid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"}],\"name\":\"getActiveBidsByModelAgent\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"deletedAt\",\"type\":\"uint128\"}],\"internalType\":\"structBid[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"getActiveBidsByProvider\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"deletedAt\",\"type\":\"uint128\"}],\"internalType\":\"structBid[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"limit\",\"type\":\"uint8\"}],\"name\":\"getActiveBidsRatingByModelAgent\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"deletedAt\",\"type\":\"uint128\"}],\"internalType\":\"structBid[]\",\"name\":\"\",\"type\":\"tuple[]\"},{\"components\":[{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"tpsScaled1000\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"ttftMs\",\"type\":\"tuple\"},{\"internalType\":\"uint32\",\"name\":\"totalDuration\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"successCount\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"totalCount\",\"type\":\"uint32\"}],\"internalType\":\"structProviderModelStats[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"limit\",\"type\":\"uint8\"}],\"name\":\"getBidsByModelAgent\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"deletedAt\",\"type\":\"uint128\"}],\"internalType\":\"structBid[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset\",\"type\":\"uint256\"},{\"internalType\":\"uint8\",\"name\":\"limit\",\"type\":\"uint8\"}],\"name\":\"getBidsByProvider\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"},{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelAgentId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"deletedAt\",\"type\":\"uint128\"}],\"internalType\":\"structBid[]\",\"name\":\"\",\"type\":\"tuple[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelID\",\"type\":\"bytes32\"}],\"name\":\"getModelStats\",\"outputs\":[{\"components\":[{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"tpsScaled1000\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"ttftMs\",\"type\":\"tuple\"},{\"components\":[{\"internalType\":\"int64\",\"name\":\"mean\",\"type\":\"int64\"},{\"internalType\":\"int64\",\"name\":\"sqSum\",\"type\":\"int64\"}],\"internalType\":\"structLibSD.SD\",\"name\":\"totalDuration\",\"type\":\"tuple\"},{\"internalType\":\"uint32\",\"name\":\"count\",\"type\":\"uint32\"}],\"internalType\":\"structModelStats\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"providerAddr\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"}],\"name\":\"postModelBid\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_bidFee\",\"type\":\"uint256\"}],\"name\":\"setBidFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"addr\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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

// BidFee is a free data retrieval call binding the contract method 0xe14a2115.
//
// Solidity: function bidFee() view returns(uint256)
func (_Marketplace *MarketplaceCaller) BidFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "bidFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BidFee is a free data retrieval call binding the contract method 0xe14a2115.
//
// Solidity: function bidFee() view returns(uint256)
func (_Marketplace *MarketplaceSession) BidFee() (*big.Int, error) {
	return _Marketplace.Contract.BidFee(&_Marketplace.CallOpts)
}

// BidFee is a free data retrieval call binding the contract method 0xe14a2115.
//
// Solidity: function bidFee() view returns(uint256)
func (_Marketplace *MarketplaceCallerSession) BidFee() (*big.Int, error) {
	return _Marketplace.Contract.BidFee(&_Marketplace.CallOpts)
}

// BidMap is a free data retrieval call binding the contract method 0xf141c9a3.
//
// Solidity: function bidMap(bytes32 bidId) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_Marketplace *MarketplaceCaller) BidMap(opts *bind.CallOpts, bidId [32]byte) (Bid, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "bidMap", bidId)

	if err != nil {
		return *new(Bid), err
	}

	out0 := *abi.ConvertType(out[0], new(Bid)).(*Bid)

	return out0, err

}

// BidMap is a free data retrieval call binding the contract method 0xf141c9a3.
//
// Solidity: function bidMap(bytes32 bidId) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_Marketplace *MarketplaceSession) BidMap(bidId [32]byte) (Bid, error) {
	return _Marketplace.Contract.BidMap(&_Marketplace.CallOpts, bidId)
}

// BidMap is a free data retrieval call binding the contract method 0xf141c9a3.
//
// Solidity: function bidMap(bytes32 bidId) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_Marketplace *MarketplaceCallerSession) BidMap(bidId [32]byte) (Bid, error) {
	return _Marketplace.Contract.BidMap(&_Marketplace.CallOpts, bidId)
}

// GetActiveBidsByModelAgent is a free data retrieval call binding the contract method 0x873d94d5.
//
// Solidity: function getActiveBidsByModelAgent(bytes32 modelAgentId) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[])
func (_Marketplace *MarketplaceCaller) GetActiveBidsByModelAgent(opts *bind.CallOpts, modelAgentId [32]byte) ([][32]byte, []Bid, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getActiveBidsByModelAgent", modelAgentId)

	if err != nil {
		return *new([][32]byte), *new([]Bid), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)
	out1 := *abi.ConvertType(out[1], new([]Bid)).(*[]Bid)

	return out0, out1, err

}

// GetActiveBidsByModelAgent is a free data retrieval call binding the contract method 0x873d94d5.
//
// Solidity: function getActiveBidsByModelAgent(bytes32 modelAgentId) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[])
func (_Marketplace *MarketplaceSession) GetActiveBidsByModelAgent(modelAgentId [32]byte) ([][32]byte, []Bid, error) {
	return _Marketplace.Contract.GetActiveBidsByModelAgent(&_Marketplace.CallOpts, modelAgentId)
}

// GetActiveBidsByModelAgent is a free data retrieval call binding the contract method 0x873d94d5.
//
// Solidity: function getActiveBidsByModelAgent(bytes32 modelAgentId) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[])
func (_Marketplace *MarketplaceCallerSession) GetActiveBidsByModelAgent(modelAgentId [32]byte) ([][32]byte, []Bid, error) {
	return _Marketplace.Contract.GetActiveBidsByModelAgent(&_Marketplace.CallOpts, modelAgentId)
}

// GetActiveBidsByProvider is a free data retrieval call binding the contract method 0x9fdaffd0.
//
// Solidity: function getActiveBidsByProvider(address provider) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[])
func (_Marketplace *MarketplaceCaller) GetActiveBidsByProvider(opts *bind.CallOpts, provider common.Address) ([][32]byte, []Bid, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getActiveBidsByProvider", provider)

	if err != nil {
		return *new([][32]byte), *new([]Bid), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)
	out1 := *abi.ConvertType(out[1], new([]Bid)).(*[]Bid)

	return out0, out1, err

}

// GetActiveBidsByProvider is a free data retrieval call binding the contract method 0x9fdaffd0.
//
// Solidity: function getActiveBidsByProvider(address provider) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[])
func (_Marketplace *MarketplaceSession) GetActiveBidsByProvider(provider common.Address) ([][32]byte, []Bid, error) {
	return _Marketplace.Contract.GetActiveBidsByProvider(&_Marketplace.CallOpts, provider)
}

// GetActiveBidsByProvider is a free data retrieval call binding the contract method 0x9fdaffd0.
//
// Solidity: function getActiveBidsByProvider(address provider) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[])
func (_Marketplace *MarketplaceCallerSession) GetActiveBidsByProvider(provider common.Address) ([][32]byte, []Bid, error) {
	return _Marketplace.Contract.GetActiveBidsByProvider(&_Marketplace.CallOpts, provider)
}

// GetActiveBidsRatingByModelAgent is a free data retrieval call binding the contract method 0xa69a4dd4.
//
// Solidity: function getActiveBidsRatingByModelAgent(bytes32 modelAgentId, uint256 offset, uint8 limit) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[], ((int64,int64),(int64,int64),uint32,uint32,uint32)[])
func (_Marketplace *MarketplaceCaller) GetActiveBidsRatingByModelAgent(opts *bind.CallOpts, modelAgentId [32]byte, offset *big.Int, limit uint8) ([][32]byte, []Bid, []ProviderModelStats, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getActiveBidsRatingByModelAgent", modelAgentId, offset, limit)

	if err != nil {
		return *new([][32]byte), *new([]Bid), *new([]ProviderModelStats), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)
	out1 := *abi.ConvertType(out[1], new([]Bid)).(*[]Bid)
	out2 := *abi.ConvertType(out[2], new([]ProviderModelStats)).(*[]ProviderModelStats)

	return out0, out1, out2, err

}

// GetActiveBidsRatingByModelAgent is a free data retrieval call binding the contract method 0xa69a4dd4.
//
// Solidity: function getActiveBidsRatingByModelAgent(bytes32 modelAgentId, uint256 offset, uint8 limit) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[], ((int64,int64),(int64,int64),uint32,uint32,uint32)[])
func (_Marketplace *MarketplaceSession) GetActiveBidsRatingByModelAgent(modelAgentId [32]byte, offset *big.Int, limit uint8) ([][32]byte, []Bid, []ProviderModelStats, error) {
	return _Marketplace.Contract.GetActiveBidsRatingByModelAgent(&_Marketplace.CallOpts, modelAgentId, offset, limit)
}

// GetActiveBidsRatingByModelAgent is a free data retrieval call binding the contract method 0xa69a4dd4.
//
// Solidity: function getActiveBidsRatingByModelAgent(bytes32 modelAgentId, uint256 offset, uint8 limit) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[], ((int64,int64),(int64,int64),uint32,uint32,uint32)[])
func (_Marketplace *MarketplaceCallerSession) GetActiveBidsRatingByModelAgent(modelAgentId [32]byte, offset *big.Int, limit uint8) ([][32]byte, []Bid, []ProviderModelStats, error) {
	return _Marketplace.Contract.GetActiveBidsRatingByModelAgent(&_Marketplace.CallOpts, modelAgentId, offset, limit)
}

// GetBidsByModelAgent is a free data retrieval call binding the contract method 0xa87665ec.
//
// Solidity: function getBidsByModelAgent(bytes32 modelAgentId, uint256 offset, uint8 limit) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[])
func (_Marketplace *MarketplaceCaller) GetBidsByModelAgent(opts *bind.CallOpts, modelAgentId [32]byte, offset *big.Int, limit uint8) ([][32]byte, []Bid, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getBidsByModelAgent", modelAgentId, offset, limit)

	if err != nil {
		return *new([][32]byte), *new([]Bid), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)
	out1 := *abi.ConvertType(out[1], new([]Bid)).(*[]Bid)

	return out0, out1, err

}

// GetBidsByModelAgent is a free data retrieval call binding the contract method 0xa87665ec.
//
// Solidity: function getBidsByModelAgent(bytes32 modelAgentId, uint256 offset, uint8 limit) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[])
func (_Marketplace *MarketplaceSession) GetBidsByModelAgent(modelAgentId [32]byte, offset *big.Int, limit uint8) ([][32]byte, []Bid, error) {
	return _Marketplace.Contract.GetBidsByModelAgent(&_Marketplace.CallOpts, modelAgentId, offset, limit)
}

// GetBidsByModelAgent is a free data retrieval call binding the contract method 0xa87665ec.
//
// Solidity: function getBidsByModelAgent(bytes32 modelAgentId, uint256 offset, uint8 limit) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[])
func (_Marketplace *MarketplaceCallerSession) GetBidsByModelAgent(modelAgentId [32]byte, offset *big.Int, limit uint8) ([][32]byte, []Bid, error) {
	return _Marketplace.Contract.GetBidsByModelAgent(&_Marketplace.CallOpts, modelAgentId, offset, limit)
}

// GetBidsByProvider is a free data retrieval call binding the contract method 0x2f817685.
//
// Solidity: function getBidsByProvider(address provider, uint256 offset, uint8 limit) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[])
func (_Marketplace *MarketplaceCaller) GetBidsByProvider(opts *bind.CallOpts, provider common.Address, offset *big.Int, limit uint8) ([][32]byte, []Bid, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getBidsByProvider", provider, offset, limit)

	if err != nil {
		return *new([][32]byte), *new([]Bid), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)
	out1 := *abi.ConvertType(out[1], new([]Bid)).(*[]Bid)

	return out0, out1, err

}

// GetBidsByProvider is a free data retrieval call binding the contract method 0x2f817685.
//
// Solidity: function getBidsByProvider(address provider, uint256 offset, uint8 limit) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[])
func (_Marketplace *MarketplaceSession) GetBidsByProvider(provider common.Address, offset *big.Int, limit uint8) ([][32]byte, []Bid, error) {
	return _Marketplace.Contract.GetBidsByProvider(&_Marketplace.CallOpts, provider, offset, limit)
}

// GetBidsByProvider is a free data retrieval call binding the contract method 0x2f817685.
//
// Solidity: function getBidsByProvider(address provider, uint256 offset, uint8 limit) view returns(bytes32[], (address,bytes32,uint256,uint256,uint128,uint128)[])
func (_Marketplace *MarketplaceCallerSession) GetBidsByProvider(provider common.Address, offset *big.Int, limit uint8) ([][32]byte, []Bid, error) {
	return _Marketplace.Contract.GetBidsByProvider(&_Marketplace.CallOpts, provider, offset, limit)
}

// GetModelStats is a free data retrieval call binding the contract method 0xce535723.
//
// Solidity: function getModelStats(bytes32 modelID) view returns(((int64,int64),(int64,int64),(int64,int64),uint32))
func (_Marketplace *MarketplaceCaller) GetModelStats(opts *bind.CallOpts, modelID [32]byte) (ModelStats, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getModelStats", modelID)

	if err != nil {
		return *new(ModelStats), err
	}

	out0 := *abi.ConvertType(out[0], new(ModelStats)).(*ModelStats)

	return out0, err

}

// GetModelStats is a free data retrieval call binding the contract method 0xce535723.
//
// Solidity: function getModelStats(bytes32 modelID) view returns(((int64,int64),(int64,int64),(int64,int64),uint32))
func (_Marketplace *MarketplaceSession) GetModelStats(modelID [32]byte) (ModelStats, error) {
	return _Marketplace.Contract.GetModelStats(&_Marketplace.CallOpts, modelID)
}

// GetModelStats is a free data retrieval call binding the contract method 0xce535723.
//
// Solidity: function getModelStats(bytes32 modelID) view returns(((int64,int64),(int64,int64),(int64,int64),uint32))
func (_Marketplace *MarketplaceCallerSession) GetModelStats(modelID [32]byte) (ModelStats, error) {
	return _Marketplace.Contract.GetModelStats(&_Marketplace.CallOpts, modelID)
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

// PostModelBid is a paid mutator transaction binding the contract method 0xede96bb1.
//
// Solidity: function postModelBid(address providerAddr, bytes32 modelId, uint256 pricePerSecond) returns(bytes32 bidId)
func (_Marketplace *MarketplaceTransactor) PostModelBid(opts *bind.TransactOpts, providerAddr common.Address, modelId [32]byte, pricePerSecond *big.Int) (*types.Transaction, error) {
	return _Marketplace.contract.Transact(opts, "postModelBid", providerAddr, modelId, pricePerSecond)
}

// PostModelBid is a paid mutator transaction binding the contract method 0xede96bb1.
//
// Solidity: function postModelBid(address providerAddr, bytes32 modelId, uint256 pricePerSecond) returns(bytes32 bidId)
func (_Marketplace *MarketplaceSession) PostModelBid(providerAddr common.Address, modelId [32]byte, pricePerSecond *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.PostModelBid(&_Marketplace.TransactOpts, providerAddr, modelId, pricePerSecond)
}

// PostModelBid is a paid mutator transaction binding the contract method 0xede96bb1.
//
// Solidity: function postModelBid(address providerAddr, bytes32 modelId, uint256 pricePerSecond) returns(bytes32 bidId)
func (_Marketplace *MarketplaceTransactorSession) PostModelBid(providerAddr common.Address, modelId [32]byte, pricePerSecond *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.PostModelBid(&_Marketplace.TransactOpts, providerAddr, modelId, pricePerSecond)
}

// SetBidFee is a paid mutator transaction binding the contract method 0x013869bf.
//
// Solidity: function setBidFee(uint256 _bidFee) returns()
func (_Marketplace *MarketplaceTransactor) SetBidFee(opts *bind.TransactOpts, _bidFee *big.Int) (*types.Transaction, error) {
	return _Marketplace.contract.Transact(opts, "setBidFee", _bidFee)
}

// SetBidFee is a paid mutator transaction binding the contract method 0x013869bf.
//
// Solidity: function setBidFee(uint256 _bidFee) returns()
func (_Marketplace *MarketplaceSession) SetBidFee(_bidFee *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.SetBidFee(&_Marketplace.TransactOpts, _bidFee)
}

// SetBidFee is a paid mutator transaction binding the contract method 0x013869bf.
//
// Solidity: function setBidFee(uint256 _bidFee) returns()
func (_Marketplace *MarketplaceTransactorSession) SetBidFee(_bidFee *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.SetBidFee(&_Marketplace.TransactOpts, _bidFee)
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
	BidFee *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterFeeUpdated is a free log retrieval operation binding the contract event 0x8c4d35e54a3f2ef1134138fd8ea3daee6a3c89e10d2665996babdf70261e2c76.
//
// Solidity: event FeeUpdated(uint256 bidFee)
func (_Marketplace *MarketplaceFilterer) FilterFeeUpdated(opts *bind.FilterOpts) (*MarketplaceFeeUpdatedIterator, error) {

	logs, sub, err := _Marketplace.contract.FilterLogs(opts, "FeeUpdated")
	if err != nil {
		return nil, err
	}
	return &MarketplaceFeeUpdatedIterator{contract: _Marketplace.contract, event: "FeeUpdated", logs: logs, sub: sub}, nil
}

// WatchFeeUpdated is a free log subscription operation binding the contract event 0x8c4d35e54a3f2ef1134138fd8ea3daee6a3c89e10d2665996babdf70261e2c76.
//
// Solidity: event FeeUpdated(uint256 bidFee)
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

// ParseFeeUpdated is a log parse operation binding the contract event 0x8c4d35e54a3f2ef1134138fd8ea3daee6a3c89e10d2665996babdf70261e2c76.
//
// Solidity: event FeeUpdated(uint256 bidFee)
func (_Marketplace *MarketplaceFilterer) ParseFeeUpdated(log types.Log) (*MarketplaceFeeUpdated, error) {
	event := new(MarketplaceFeeUpdated)
	if err := _Marketplace.contract.UnpackLog(event, "FeeUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
