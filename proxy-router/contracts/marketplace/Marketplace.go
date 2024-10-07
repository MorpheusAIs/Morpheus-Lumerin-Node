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

// IBidStorageBid is an auto generated low-level Go binding around an user-defined struct.
type IBidStorageBid struct {
	Provider       common.Address
	ModelId        [32]byte
	PricePerSecond *big.Int
	Nonce          *big.Int
	CreatedAt      *big.Int
	DeletedAt      *big.Int
}

// IModelStorageModel is an auto generated low-level Go binding around an user-defined struct.
type IModelStorageModel struct {
	IpfsCID   [32]byte
	Fee       *big.Int
	Stake     *big.Int
	Owner     common.Address
	Name      string
	Tags      []string
	CreatedAt *big.Int
	IsDeleted bool
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

// MarketplaceMetaData contains all meta data concerning the Marketplace contract.
var MarketplaceMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"ActiveBidNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"BidTaken\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ModelNotFound\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotEnoughBalance\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"NotOwnerOrProvider\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account_\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"inputs\":[],\"name\":\"ProviderNotFound\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"BidDeleted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"name\":\"BidPosted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"bidFee\",\"type\":\"uint256\"}],\"name\":\"FeeUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"storageSlot\",\"type\":\"bytes32\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BID_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DIAMOND_OWNABLE_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MARKETPLACE_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MODEL_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"PROVIDER_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"token_\",\"type\":\"address\"}],\"name\":\"__Marketplace_init\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"}],\"name\":\"bids\",\"outputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"deletedAt\",\"type\":\"uint128\"}],\"internalType\":\"structIBidStorage.Bid\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId_\",\"type\":\"bytes32\"}],\"name\":\"deleteModelBid\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getBidFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"nonce_\",\"type\":\"uint256\"}],\"name\":\"getBidId\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId\",\"type\":\"bytes32\"}],\"name\":\"getModel\",\"outputs\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"ipfsCID\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"fee\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"name\",\"type\":\"string\"},{\"internalType\":\"string[]\",\"name\":\"tags\",\"type\":\"string[]\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"internalType\":\"structIModelStorage.Model\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider\",\"type\":\"address\"}],\"name\":\"getProvider\",\"outputs\":[{\"components\":[{\"internalType\":\"string\",\"name\":\"endpoint\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"stake\",\"type\":\"uint256\"},{\"internalType\":\"uint128\",\"name\":\"createdAt\",\"type\":\"uint128\"},{\"internalType\":\"uint128\",\"name\":\"limitPeriodEnd\",\"type\":\"uint128\"},{\"internalType\":\"uint256\",\"name\":\"limitPeriodEarned\",\"type\":\"uint256\"},{\"internalType\":\"bool\",\"name\":\"isDeleted\",\"type\":\"bool\"}],\"internalType\":\"structIProviderStorage.Provider\",\"name\":\"\",\"type\":\"tuple\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"}],\"name\":\"getProviderModelId\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"modelActiveBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"modelBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"modelMinimumStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"index\",\"type\":\"uint256\"}],\"name\":\"models\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"modelId_\",\"type\":\"bytes32\"},{\"internalType\":\"uint256\",\"name\":\"pricePerSecond_\",\"type\":\"uint256\"}],\"name\":\"postModelBid\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"bidId\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"providerActiveBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"provider_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"offset_\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"limit_\",\"type\":\"uint256\"}],\"name\":\"providerBids\",\"outputs\":[{\"internalType\":\"bytes32[]\",\"name\":\"\",\"type\":\"bytes32[]\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"providerMinimumStake\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"bidFee_\",\"type\":\"uint256\"}],\"name\":\"setBidFee\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"recipient_\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"amount_\",\"type\":\"uint256\"}],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
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

// BIDSTORAGESLOT is a free data retrieval call binding the contract method 0x4fa816f2.
//
// Solidity: function BID_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceCaller) BIDSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "BID_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BIDSTORAGESLOT is a free data retrieval call binding the contract method 0x4fa816f2.
//
// Solidity: function BID_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceSession) BIDSTORAGESLOT() ([32]byte, error) {
	return _Marketplace.Contract.BIDSTORAGESLOT(&_Marketplace.CallOpts)
}

// BIDSTORAGESLOT is a free data retrieval call binding the contract method 0x4fa816f2.
//
// Solidity: function BID_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceCallerSession) BIDSTORAGESLOT() ([32]byte, error) {
	return _Marketplace.Contract.BIDSTORAGESLOT(&_Marketplace.CallOpts)
}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceCaller) DIAMONDOWNABLESTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "DIAMOND_OWNABLE_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceSession) DIAMONDOWNABLESTORAGESLOT() ([32]byte, error) {
	return _Marketplace.Contract.DIAMONDOWNABLESTORAGESLOT(&_Marketplace.CallOpts)
}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceCallerSession) DIAMONDOWNABLESTORAGESLOT() ([32]byte, error) {
	return _Marketplace.Contract.DIAMONDOWNABLESTORAGESLOT(&_Marketplace.CallOpts)
}

// MARKETPLACESTORAGESLOT is a free data retrieval call binding the contract method 0xb855c3b5.
//
// Solidity: function MARKETPLACE_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceCaller) MARKETPLACESTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "MARKETPLACE_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MARKETPLACESTORAGESLOT is a free data retrieval call binding the contract method 0xb855c3b5.
//
// Solidity: function MARKETPLACE_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceSession) MARKETPLACESTORAGESLOT() ([32]byte, error) {
	return _Marketplace.Contract.MARKETPLACESTORAGESLOT(&_Marketplace.CallOpts)
}

// MARKETPLACESTORAGESLOT is a free data retrieval call binding the contract method 0xb855c3b5.
//
// Solidity: function MARKETPLACE_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceCallerSession) MARKETPLACESTORAGESLOT() ([32]byte, error) {
	return _Marketplace.Contract.MARKETPLACESTORAGESLOT(&_Marketplace.CallOpts)
}

// MODELSTORAGESLOT is a free data retrieval call binding the contract method 0xeda926f2.
//
// Solidity: function MODEL_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceCaller) MODELSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "MODEL_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// MODELSTORAGESLOT is a free data retrieval call binding the contract method 0xeda926f2.
//
// Solidity: function MODEL_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceSession) MODELSTORAGESLOT() ([32]byte, error) {
	return _Marketplace.Contract.MODELSTORAGESLOT(&_Marketplace.CallOpts)
}

// MODELSTORAGESLOT is a free data retrieval call binding the contract method 0xeda926f2.
//
// Solidity: function MODEL_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceCallerSession) MODELSTORAGESLOT() ([32]byte, error) {
	return _Marketplace.Contract.MODELSTORAGESLOT(&_Marketplace.CallOpts)
}

// PROVIDERSTORAGESLOT is a free data retrieval call binding the contract method 0x490713b1.
//
// Solidity: function PROVIDER_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceCaller) PROVIDERSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "PROVIDER_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// PROVIDERSTORAGESLOT is a free data retrieval call binding the contract method 0x490713b1.
//
// Solidity: function PROVIDER_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceSession) PROVIDERSTORAGESLOT() ([32]byte, error) {
	return _Marketplace.Contract.PROVIDERSTORAGESLOT(&_Marketplace.CallOpts)
}

// PROVIDERSTORAGESLOT is a free data retrieval call binding the contract method 0x490713b1.
//
// Solidity: function PROVIDER_STORAGE_SLOT() view returns(bytes32)
func (_Marketplace *MarketplaceCallerSession) PROVIDERSTORAGESLOT() ([32]byte, error) {
	return _Marketplace.Contract.PROVIDERSTORAGESLOT(&_Marketplace.CallOpts)
}

// Bids is a free data retrieval call binding the contract method 0x8f98eeda.
//
// Solidity: function bids(bytes32 bidId) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_Marketplace *MarketplaceCaller) Bids(opts *bind.CallOpts, bidId [32]byte) (IBidStorageBid, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "bids", bidId)

	if err != nil {
		return *new(IBidStorageBid), err
	}

	out0 := *abi.ConvertType(out[0], new(IBidStorageBid)).(*IBidStorageBid)

	return out0, err

}

// Bids is a free data retrieval call binding the contract method 0x8f98eeda.
//
// Solidity: function bids(bytes32 bidId) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_Marketplace *MarketplaceSession) Bids(bidId [32]byte) (IBidStorageBid, error) {
	return _Marketplace.Contract.Bids(&_Marketplace.CallOpts, bidId)
}

// Bids is a free data retrieval call binding the contract method 0x8f98eeda.
//
// Solidity: function bids(bytes32 bidId) view returns((address,bytes32,uint256,uint256,uint128,uint128))
func (_Marketplace *MarketplaceCallerSession) Bids(bidId [32]byte) (IBidStorageBid, error) {
	return _Marketplace.Contract.Bids(&_Marketplace.CallOpts, bidId)
}

// GetBidFee is a free data retrieval call binding the contract method 0x8dbb4647.
//
// Solidity: function getBidFee() view returns(uint256)
func (_Marketplace *MarketplaceCaller) GetBidFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getBidFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetBidFee is a free data retrieval call binding the contract method 0x8dbb4647.
//
// Solidity: function getBidFee() view returns(uint256)
func (_Marketplace *MarketplaceSession) GetBidFee() (*big.Int, error) {
	return _Marketplace.Contract.GetBidFee(&_Marketplace.CallOpts)
}

// GetBidFee is a free data retrieval call binding the contract method 0x8dbb4647.
//
// Solidity: function getBidFee() view returns(uint256)
func (_Marketplace *MarketplaceCallerSession) GetBidFee() (*big.Int, error) {
	return _Marketplace.Contract.GetBidFee(&_Marketplace.CallOpts)
}

// GetBidId is a free data retrieval call binding the contract method 0x747ddd5b.
//
// Solidity: function getBidId(address provider_, bytes32 modelId_, uint256 nonce_) pure returns(bytes32)
func (_Marketplace *MarketplaceCaller) GetBidId(opts *bind.CallOpts, provider_ common.Address, modelId_ [32]byte, nonce_ *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getBidId", provider_, modelId_, nonce_)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetBidId is a free data retrieval call binding the contract method 0x747ddd5b.
//
// Solidity: function getBidId(address provider_, bytes32 modelId_, uint256 nonce_) pure returns(bytes32)
func (_Marketplace *MarketplaceSession) GetBidId(provider_ common.Address, modelId_ [32]byte, nonce_ *big.Int) ([32]byte, error) {
	return _Marketplace.Contract.GetBidId(&_Marketplace.CallOpts, provider_, modelId_, nonce_)
}

// GetBidId is a free data retrieval call binding the contract method 0x747ddd5b.
//
// Solidity: function getBidId(address provider_, bytes32 modelId_, uint256 nonce_) pure returns(bytes32)
func (_Marketplace *MarketplaceCallerSession) GetBidId(provider_ common.Address, modelId_ [32]byte, nonce_ *big.Int) ([32]byte, error) {
	return _Marketplace.Contract.GetBidId(&_Marketplace.CallOpts, provider_, modelId_, nonce_)
}

// GetModel is a free data retrieval call binding the contract method 0x21e7c498.
//
// Solidity: function getModel(bytes32 modelId) view returns((bytes32,uint256,uint256,address,string,string[],uint128,bool))
func (_Marketplace *MarketplaceCaller) GetModel(opts *bind.CallOpts, modelId [32]byte) (IModelStorageModel, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getModel", modelId)

	if err != nil {
		return *new(IModelStorageModel), err
	}

	out0 := *abi.ConvertType(out[0], new(IModelStorageModel)).(*IModelStorageModel)

	return out0, err

}

// GetModel is a free data retrieval call binding the contract method 0x21e7c498.
//
// Solidity: function getModel(bytes32 modelId) view returns((bytes32,uint256,uint256,address,string,string[],uint128,bool))
func (_Marketplace *MarketplaceSession) GetModel(modelId [32]byte) (IModelStorageModel, error) {
	return _Marketplace.Contract.GetModel(&_Marketplace.CallOpts, modelId)
}

// GetModel is a free data retrieval call binding the contract method 0x21e7c498.
//
// Solidity: function getModel(bytes32 modelId) view returns((bytes32,uint256,uint256,address,string,string[],uint128,bool))
func (_Marketplace *MarketplaceCallerSession) GetModel(modelId [32]byte) (IModelStorageModel, error) {
	return _Marketplace.Contract.GetModel(&_Marketplace.CallOpts, modelId)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_Marketplace *MarketplaceCaller) GetProvider(opts *bind.CallOpts, provider common.Address) (IProviderStorageProvider, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getProvider", provider)

	if err != nil {
		return *new(IProviderStorageProvider), err
	}

	out0 := *abi.ConvertType(out[0], new(IProviderStorageProvider)).(*IProviderStorageProvider)

	return out0, err

}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_Marketplace *MarketplaceSession) GetProvider(provider common.Address) (IProviderStorageProvider, error) {
	return _Marketplace.Contract.GetProvider(&_Marketplace.CallOpts, provider)
}

// GetProvider is a free data retrieval call binding the contract method 0x55f21eb7.
//
// Solidity: function getProvider(address provider) view returns((string,uint256,uint128,uint128,uint256,bool))
func (_Marketplace *MarketplaceCallerSession) GetProvider(provider common.Address) (IProviderStorageProvider, error) {
	return _Marketplace.Contract.GetProvider(&_Marketplace.CallOpts, provider)
}

// GetProviderModelId is a free data retrieval call binding the contract method 0x1cc9de8c.
//
// Solidity: function getProviderModelId(address provider_, bytes32 modelId_) pure returns(bytes32)
func (_Marketplace *MarketplaceCaller) GetProviderModelId(opts *bind.CallOpts, provider_ common.Address, modelId_ [32]byte) ([32]byte, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getProviderModelId", provider_, modelId_)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetProviderModelId is a free data retrieval call binding the contract method 0x1cc9de8c.
//
// Solidity: function getProviderModelId(address provider_, bytes32 modelId_) pure returns(bytes32)
func (_Marketplace *MarketplaceSession) GetProviderModelId(provider_ common.Address, modelId_ [32]byte) ([32]byte, error) {
	return _Marketplace.Contract.GetProviderModelId(&_Marketplace.CallOpts, provider_, modelId_)
}

// GetProviderModelId is a free data retrieval call binding the contract method 0x1cc9de8c.
//
// Solidity: function getProviderModelId(address provider_, bytes32 modelId_) pure returns(bytes32)
func (_Marketplace *MarketplaceCallerSession) GetProviderModelId(provider_ common.Address, modelId_ [32]byte) ([32]byte, error) {
	return _Marketplace.Contract.GetProviderModelId(&_Marketplace.CallOpts, provider_, modelId_)
}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_Marketplace *MarketplaceCaller) GetToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "getToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_Marketplace *MarketplaceSession) GetToken() (common.Address, error) {
	return _Marketplace.Contract.GetToken(&_Marketplace.CallOpts)
}

// GetToken is a free data retrieval call binding the contract method 0x21df0da7.
//
// Solidity: function getToken() view returns(address)
func (_Marketplace *MarketplaceCallerSession) GetToken() (common.Address, error) {
	return _Marketplace.Contract.GetToken(&_Marketplace.CallOpts)
}

// ModelActiveBids is a free data retrieval call binding the contract method 0x3fd8e5e3.
//
// Solidity: function modelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_Marketplace *MarketplaceCaller) ModelActiveBids(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "modelActiveBids", modelId_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// ModelActiveBids is a free data retrieval call binding the contract method 0x3fd8e5e3.
//
// Solidity: function modelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_Marketplace *MarketplaceSession) ModelActiveBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _Marketplace.Contract.ModelActiveBids(&_Marketplace.CallOpts, modelId_, offset_, limit_)
}

// ModelActiveBids is a free data retrieval call binding the contract method 0x3fd8e5e3.
//
// Solidity: function modelActiveBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_Marketplace *MarketplaceCallerSession) ModelActiveBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _Marketplace.Contract.ModelActiveBids(&_Marketplace.CallOpts, modelId_, offset_, limit_)
}

// ModelBids is a free data retrieval call binding the contract method 0x5954d1b3.
//
// Solidity: function modelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_Marketplace *MarketplaceCaller) ModelBids(opts *bind.CallOpts, modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "modelBids", modelId_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// ModelBids is a free data retrieval call binding the contract method 0x5954d1b3.
//
// Solidity: function modelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_Marketplace *MarketplaceSession) ModelBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _Marketplace.Contract.ModelBids(&_Marketplace.CallOpts, modelId_, offset_, limit_)
}

// ModelBids is a free data retrieval call binding the contract method 0x5954d1b3.
//
// Solidity: function modelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_Marketplace *MarketplaceCallerSession) ModelBids(modelId_ [32]byte, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _Marketplace.Contract.ModelBids(&_Marketplace.CallOpts, modelId_, offset_, limit_)
}

// ModelMinimumStake is a free data retrieval call binding the contract method 0xc4288ed4.
//
// Solidity: function modelMinimumStake() view returns(uint256)
func (_Marketplace *MarketplaceCaller) ModelMinimumStake(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "modelMinimumStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ModelMinimumStake is a free data retrieval call binding the contract method 0xc4288ed4.
//
// Solidity: function modelMinimumStake() view returns(uint256)
func (_Marketplace *MarketplaceSession) ModelMinimumStake() (*big.Int, error) {
	return _Marketplace.Contract.ModelMinimumStake(&_Marketplace.CallOpts)
}

// ModelMinimumStake is a free data retrieval call binding the contract method 0xc4288ed4.
//
// Solidity: function modelMinimumStake() view returns(uint256)
func (_Marketplace *MarketplaceCallerSession) ModelMinimumStake() (*big.Int, error) {
	return _Marketplace.Contract.ModelMinimumStake(&_Marketplace.CallOpts)
}

// Models is a free data retrieval call binding the contract method 0x6a030ca9.
//
// Solidity: function models(uint256 index) view returns(bytes32)
func (_Marketplace *MarketplaceCaller) Models(opts *bind.CallOpts, index *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "models", index)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// Models is a free data retrieval call binding the contract method 0x6a030ca9.
//
// Solidity: function models(uint256 index) view returns(bytes32)
func (_Marketplace *MarketplaceSession) Models(index *big.Int) ([32]byte, error) {
	return _Marketplace.Contract.Models(&_Marketplace.CallOpts, index)
}

// Models is a free data retrieval call binding the contract method 0x6a030ca9.
//
// Solidity: function models(uint256 index) view returns(bytes32)
func (_Marketplace *MarketplaceCallerSession) Models(index *big.Int) ([32]byte, error) {
	return _Marketplace.Contract.Models(&_Marketplace.CallOpts, index)
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

// ProviderActiveBids is a free data retrieval call binding the contract method 0x6dd7d31c.
//
// Solidity: function providerActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_Marketplace *MarketplaceCaller) ProviderActiveBids(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "providerActiveBids", provider_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// ProviderActiveBids is a free data retrieval call binding the contract method 0x6dd7d31c.
//
// Solidity: function providerActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_Marketplace *MarketplaceSession) ProviderActiveBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _Marketplace.Contract.ProviderActiveBids(&_Marketplace.CallOpts, provider_, offset_, limit_)
}

// ProviderActiveBids is a free data retrieval call binding the contract method 0x6dd7d31c.
//
// Solidity: function providerActiveBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_Marketplace *MarketplaceCallerSession) ProviderActiveBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _Marketplace.Contract.ProviderActiveBids(&_Marketplace.CallOpts, provider_, offset_, limit_)
}

// ProviderBids is a free data retrieval call binding the contract method 0x22fbda9a.
//
// Solidity: function providerBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_Marketplace *MarketplaceCaller) ProviderBids(opts *bind.CallOpts, provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "providerBids", provider_, offset_, limit_)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// ProviderBids is a free data retrieval call binding the contract method 0x22fbda9a.
//
// Solidity: function providerBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_Marketplace *MarketplaceSession) ProviderBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _Marketplace.Contract.ProviderBids(&_Marketplace.CallOpts, provider_, offset_, limit_)
}

// ProviderBids is a free data retrieval call binding the contract method 0x22fbda9a.
//
// Solidity: function providerBids(address provider_, uint256 offset_, uint256 limit_) view returns(bytes32[])
func (_Marketplace *MarketplaceCallerSession) ProviderBids(provider_ common.Address, offset_ *big.Int, limit_ *big.Int) ([][32]byte, error) {
	return _Marketplace.Contract.ProviderBids(&_Marketplace.CallOpts, provider_, offset_, limit_)
}

// ProviderMinimumStake is a free data retrieval call binding the contract method 0x9476c58e.
//
// Solidity: function providerMinimumStake() view returns(uint256)
func (_Marketplace *MarketplaceCaller) ProviderMinimumStake(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _Marketplace.contract.Call(opts, &out, "providerMinimumStake")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// ProviderMinimumStake is a free data retrieval call binding the contract method 0x9476c58e.
//
// Solidity: function providerMinimumStake() view returns(uint256)
func (_Marketplace *MarketplaceSession) ProviderMinimumStake() (*big.Int, error) {
	return _Marketplace.Contract.ProviderMinimumStake(&_Marketplace.CallOpts)
}

// ProviderMinimumStake is a free data retrieval call binding the contract method 0x9476c58e.
//
// Solidity: function providerMinimumStake() view returns(uint256)
func (_Marketplace *MarketplaceCallerSession) ProviderMinimumStake() (*big.Int, error) {
	return _Marketplace.Contract.ProviderMinimumStake(&_Marketplace.CallOpts)
}

// MarketplaceInit is a paid mutator transaction binding the contract method 0xf6ec33fe.
//
// Solidity: function __Marketplace_init(address token_) returns()
func (_Marketplace *MarketplaceTransactor) MarketplaceInit(opts *bind.TransactOpts, token_ common.Address) (*types.Transaction, error) {
	return _Marketplace.contract.Transact(opts, "__Marketplace_init", token_)
}

// MarketplaceInit is a paid mutator transaction binding the contract method 0xf6ec33fe.
//
// Solidity: function __Marketplace_init(address token_) returns()
func (_Marketplace *MarketplaceSession) MarketplaceInit(token_ common.Address) (*types.Transaction, error) {
	return _Marketplace.Contract.MarketplaceInit(&_Marketplace.TransactOpts, token_)
}

// MarketplaceInit is a paid mutator transaction binding the contract method 0xf6ec33fe.
//
// Solidity: function __Marketplace_init(address token_) returns()
func (_Marketplace *MarketplaceTransactorSession) MarketplaceInit(token_ common.Address) (*types.Transaction, error) {
	return _Marketplace.Contract.MarketplaceInit(&_Marketplace.TransactOpts, token_)
}

// DeleteModelBid is a paid mutator transaction binding the contract method 0x8913dcaa.
//
// Solidity: function deleteModelBid(bytes32 bidId_) returns()
func (_Marketplace *MarketplaceTransactor) DeleteModelBid(opts *bind.TransactOpts, bidId_ [32]byte) (*types.Transaction, error) {
	return _Marketplace.contract.Transact(opts, "deleteModelBid", bidId_)
}

// DeleteModelBid is a paid mutator transaction binding the contract method 0x8913dcaa.
//
// Solidity: function deleteModelBid(bytes32 bidId_) returns()
func (_Marketplace *MarketplaceSession) DeleteModelBid(bidId_ [32]byte) (*types.Transaction, error) {
	return _Marketplace.Contract.DeleteModelBid(&_Marketplace.TransactOpts, bidId_)
}

// DeleteModelBid is a paid mutator transaction binding the contract method 0x8913dcaa.
//
// Solidity: function deleteModelBid(bytes32 bidId_) returns()
func (_Marketplace *MarketplaceTransactorSession) DeleteModelBid(bidId_ [32]byte) (*types.Transaction, error) {
	return _Marketplace.Contract.DeleteModelBid(&_Marketplace.TransactOpts, bidId_)
}

// PostModelBid is a paid mutator transaction binding the contract method 0xede96bb1.
//
// Solidity: function postModelBid(address provider_, bytes32 modelId_, uint256 pricePerSecond_) returns(bytes32 bidId)
func (_Marketplace *MarketplaceTransactor) PostModelBid(opts *bind.TransactOpts, provider_ common.Address, modelId_ [32]byte, pricePerSecond_ *big.Int) (*types.Transaction, error) {
	return _Marketplace.contract.Transact(opts, "postModelBid", provider_, modelId_, pricePerSecond_)
}

// PostModelBid is a paid mutator transaction binding the contract method 0xede96bb1.
//
// Solidity: function postModelBid(address provider_, bytes32 modelId_, uint256 pricePerSecond_) returns(bytes32 bidId)
func (_Marketplace *MarketplaceSession) PostModelBid(provider_ common.Address, modelId_ [32]byte, pricePerSecond_ *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.PostModelBid(&_Marketplace.TransactOpts, provider_, modelId_, pricePerSecond_)
}

// PostModelBid is a paid mutator transaction binding the contract method 0xede96bb1.
//
// Solidity: function postModelBid(address provider_, bytes32 modelId_, uint256 pricePerSecond_) returns(bytes32 bidId)
func (_Marketplace *MarketplaceTransactorSession) PostModelBid(provider_ common.Address, modelId_ [32]byte, pricePerSecond_ *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.PostModelBid(&_Marketplace.TransactOpts, provider_, modelId_, pricePerSecond_)
}

// SetBidFee is a paid mutator transaction binding the contract method 0x013869bf.
//
// Solidity: function setBidFee(uint256 bidFee_) returns()
func (_Marketplace *MarketplaceTransactor) SetBidFee(opts *bind.TransactOpts, bidFee_ *big.Int) (*types.Transaction, error) {
	return _Marketplace.contract.Transact(opts, "setBidFee", bidFee_)
}

// SetBidFee is a paid mutator transaction binding the contract method 0x013869bf.
//
// Solidity: function setBidFee(uint256 bidFee_) returns()
func (_Marketplace *MarketplaceSession) SetBidFee(bidFee_ *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.SetBidFee(&_Marketplace.TransactOpts, bidFee_)
}

// SetBidFee is a paid mutator transaction binding the contract method 0x013869bf.
//
// Solidity: function setBidFee(uint256 bidFee_) returns()
func (_Marketplace *MarketplaceTransactorSession) SetBidFee(bidFee_ *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.SetBidFee(&_Marketplace.TransactOpts, bidFee_)
}

// Withdraw is a paid mutator transaction binding the contract method 0xf3fef3a3.
//
// Solidity: function withdraw(address recipient_, uint256 amount_) returns()
func (_Marketplace *MarketplaceTransactor) Withdraw(opts *bind.TransactOpts, recipient_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _Marketplace.contract.Transact(opts, "withdraw", recipient_, amount_)
}

// Withdraw is a paid mutator transaction binding the contract method 0xf3fef3a3.
//
// Solidity: function withdraw(address recipient_, uint256 amount_) returns()
func (_Marketplace *MarketplaceSession) Withdraw(recipient_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.Withdraw(&_Marketplace.TransactOpts, recipient_, amount_)
}

// Withdraw is a paid mutator transaction binding the contract method 0xf3fef3a3.
//
// Solidity: function withdraw(address recipient_, uint256 amount_) returns()
func (_Marketplace *MarketplaceTransactorSession) Withdraw(recipient_ common.Address, amount_ *big.Int) (*types.Transaction, error) {
	return _Marketplace.Contract.Withdraw(&_Marketplace.TransactOpts, recipient_, amount_)
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
	Provider common.Address
	ModelId  [32]byte
	Nonce    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterBidDeleted is a free log retrieval operation binding the contract event 0x096f970f504563bca8ac4419b4299946965221e396c34aea149ac84947b9242f.
//
// Solidity: event BidDeleted(address indexed provider, bytes32 indexed modelId, uint256 nonce)
func (_Marketplace *MarketplaceFilterer) FilterBidDeleted(opts *bind.FilterOpts, provider []common.Address, modelId [][32]byte) (*MarketplaceBidDeletedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var modelIdRule []interface{}
	for _, modelIdItem := range modelId {
		modelIdRule = append(modelIdRule, modelIdItem)
	}

	logs, sub, err := _Marketplace.contract.FilterLogs(opts, "BidDeleted", providerRule, modelIdRule)
	if err != nil {
		return nil, err
	}
	return &MarketplaceBidDeletedIterator{contract: _Marketplace.contract, event: "BidDeleted", logs: logs, sub: sub}, nil
}

// WatchBidDeleted is a free log subscription operation binding the contract event 0x096f970f504563bca8ac4419b4299946965221e396c34aea149ac84947b9242f.
//
// Solidity: event BidDeleted(address indexed provider, bytes32 indexed modelId, uint256 nonce)
func (_Marketplace *MarketplaceFilterer) WatchBidDeleted(opts *bind.WatchOpts, sink chan<- *MarketplaceBidDeleted, provider []common.Address, modelId [][32]byte) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var modelIdRule []interface{}
	for _, modelIdItem := range modelId {
		modelIdRule = append(modelIdRule, modelIdItem)
	}

	logs, sub, err := _Marketplace.contract.WatchLogs(opts, "BidDeleted", providerRule, modelIdRule)
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
// Solidity: event BidDeleted(address indexed provider, bytes32 indexed modelId, uint256 nonce)
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
	Provider common.Address
	ModelId  [32]byte
	Nonce    *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterBidPosted is a free log retrieval operation binding the contract event 0xd138adff73af2621d26114cd9ee4f20dcd39ed78f9e0004215ed49aa22753ebe.
//
// Solidity: event BidPosted(address indexed provider, bytes32 indexed modelId, uint256 nonce)
func (_Marketplace *MarketplaceFilterer) FilterBidPosted(opts *bind.FilterOpts, provider []common.Address, modelId [][32]byte) (*MarketplaceBidPostedIterator, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var modelIdRule []interface{}
	for _, modelIdItem := range modelId {
		modelIdRule = append(modelIdRule, modelIdItem)
	}

	logs, sub, err := _Marketplace.contract.FilterLogs(opts, "BidPosted", providerRule, modelIdRule)
	if err != nil {
		return nil, err
	}
	return &MarketplaceBidPostedIterator{contract: _Marketplace.contract, event: "BidPosted", logs: logs, sub: sub}, nil
}

// WatchBidPosted is a free log subscription operation binding the contract event 0xd138adff73af2621d26114cd9ee4f20dcd39ed78f9e0004215ed49aa22753ebe.
//
// Solidity: event BidPosted(address indexed provider, bytes32 indexed modelId, uint256 nonce)
func (_Marketplace *MarketplaceFilterer) WatchBidPosted(opts *bind.WatchOpts, sink chan<- *MarketplaceBidPosted, provider []common.Address, modelId [][32]byte) (event.Subscription, error) {

	var providerRule []interface{}
	for _, providerItem := range provider {
		providerRule = append(providerRule, providerItem)
	}
	var modelIdRule []interface{}
	for _, modelIdItem := range modelId {
		modelIdRule = append(modelIdRule, modelIdItem)
	}

	logs, sub, err := _Marketplace.contract.WatchLogs(opts, "BidPosted", providerRule, modelIdRule)
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
// Solidity: event BidPosted(address indexed provider, bytes32 indexed modelId, uint256 nonce)
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
	StorageSlot [32]byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
func (_Marketplace *MarketplaceFilterer) FilterInitialized(opts *bind.FilterOpts) (*MarketplaceInitializedIterator, error) {

	logs, sub, err := _Marketplace.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &MarketplaceInitializedIterator{contract: _Marketplace.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
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

// ParseInitialized is a log parse operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
func (_Marketplace *MarketplaceFilterer) ParseInitialized(log types.Log) (*MarketplaceInitialized, error) {
	event := new(MarketplaceInitialized)
	if err := _Marketplace.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
