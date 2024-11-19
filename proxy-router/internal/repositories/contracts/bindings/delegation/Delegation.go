// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package delegation

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

// DelegationMetaData contains all meta data concerning the Delegation contract.
var DelegationMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegator\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"delegatee\",\"type\":\"address\"}],\"name\":\"InsufficientRightsForOperation\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account_\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"registry\",\"type\":\"address\"}],\"name\":\"DelegationRegistryUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"storageSlot\",\"type\":\"bytes32\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_MARKETPLACE\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_MODEL\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_PROVIDER\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_RULES_SESSION\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DELEGATION_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DIAMOND_OWNABLE_STORAGE_SLOT\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"registry_\",\"type\":\"address\"}],\"name\":\"__Delegation_init\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getRegistry\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"delegatee_\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"delegator_\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"rights_\",\"type\":\"bytes32\"}],\"name\":\"isRightsDelegated\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"registry_\",\"type\":\"address\"}],\"name\":\"setRegistry\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// DelegationABI is the input ABI used to generate the binding from.
// Deprecated: Use DelegationMetaData.ABI instead.
var DelegationABI = DelegationMetaData.ABI

// Delegation is an auto generated Go binding around an Ethereum contract.
type Delegation struct {
	DelegationCaller     // Read-only binding to the contract
	DelegationTransactor // Write-only binding to the contract
	DelegationFilterer   // Log filterer for contract events
}

// DelegationCaller is an auto generated read-only Go binding around an Ethereum contract.
type DelegationCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DelegationTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DelegationTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DelegationFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DelegationFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DelegationSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DelegationSession struct {
	Contract     *Delegation       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DelegationCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DelegationCallerSession struct {
	Contract *DelegationCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// DelegationTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DelegationTransactorSession struct {
	Contract     *DelegationTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// DelegationRaw is an auto generated low-level Go binding around an Ethereum contract.
type DelegationRaw struct {
	Contract *Delegation // Generic contract binding to access the raw methods on
}

// DelegationCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DelegationCallerRaw struct {
	Contract *DelegationCaller // Generic read-only contract binding to access the raw methods on
}

// DelegationTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DelegationTransactorRaw struct {
	Contract *DelegationTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDelegation creates a new instance of Delegation, bound to a specific deployed contract.
func NewDelegation(address common.Address, backend bind.ContractBackend) (*Delegation, error) {
	contract, err := bindDelegation(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Delegation{DelegationCaller: DelegationCaller{contract: contract}, DelegationTransactor: DelegationTransactor{contract: contract}, DelegationFilterer: DelegationFilterer{contract: contract}}, nil
}

// NewDelegationCaller creates a new read-only instance of Delegation, bound to a specific deployed contract.
func NewDelegationCaller(address common.Address, caller bind.ContractCaller) (*DelegationCaller, error) {
	contract, err := bindDelegation(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DelegationCaller{contract: contract}, nil
}

// NewDelegationTransactor creates a new write-only instance of Delegation, bound to a specific deployed contract.
func NewDelegationTransactor(address common.Address, transactor bind.ContractTransactor) (*DelegationTransactor, error) {
	contract, err := bindDelegation(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DelegationTransactor{contract: contract}, nil
}

// NewDelegationFilterer creates a new log filterer instance of Delegation, bound to a specific deployed contract.
func NewDelegationFilterer(address common.Address, filterer bind.ContractFilterer) (*DelegationFilterer, error) {
	contract, err := bindDelegation(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DelegationFilterer{contract: contract}, nil
}

// bindDelegation binds a generic wrapper to an already deployed contract.
func bindDelegation(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := DelegationMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Delegation *DelegationRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Delegation.Contract.DelegationCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Delegation *DelegationRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Delegation.Contract.DelegationTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Delegation *DelegationRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Delegation.Contract.DelegationTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Delegation *DelegationCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Delegation.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Delegation *DelegationTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Delegation.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Delegation *DelegationTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Delegation.Contract.contract.Transact(opts, method, params...)
}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_Delegation *DelegationCaller) DELEGATIONRULESMARKETPLACE(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Delegation.contract.Call(opts, &out, "DELEGATION_RULES_MARKETPLACE")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_Delegation *DelegationSession) DELEGATIONRULESMARKETPLACE() ([32]byte, error) {
	return _Delegation.Contract.DELEGATIONRULESMARKETPLACE(&_Delegation.CallOpts)
}

// DELEGATIONRULESMARKETPLACE is a free data retrieval call binding the contract method 0xad34a150.
//
// Solidity: function DELEGATION_RULES_MARKETPLACE() view returns(bytes32)
func (_Delegation *DelegationCallerSession) DELEGATIONRULESMARKETPLACE() ([32]byte, error) {
	return _Delegation.Contract.DELEGATIONRULESMARKETPLACE(&_Delegation.CallOpts)
}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_Delegation *DelegationCaller) DELEGATIONRULESMODEL(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Delegation.contract.Call(opts, &out, "DELEGATION_RULES_MODEL")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_Delegation *DelegationSession) DELEGATIONRULESMODEL() ([32]byte, error) {
	return _Delegation.Contract.DELEGATIONRULESMODEL(&_Delegation.CallOpts)
}

// DELEGATIONRULESMODEL is a free data retrieval call binding the contract method 0x86878047.
//
// Solidity: function DELEGATION_RULES_MODEL() view returns(bytes32)
func (_Delegation *DelegationCallerSession) DELEGATIONRULESMODEL() ([32]byte, error) {
	return _Delegation.Contract.DELEGATIONRULESMODEL(&_Delegation.CallOpts)
}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_Delegation *DelegationCaller) DELEGATIONRULESPROVIDER(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Delegation.contract.Call(opts, &out, "DELEGATION_RULES_PROVIDER")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_Delegation *DelegationSession) DELEGATIONRULESPROVIDER() ([32]byte, error) {
	return _Delegation.Contract.DELEGATIONRULESPROVIDER(&_Delegation.CallOpts)
}

// DELEGATIONRULESPROVIDER is a free data retrieval call binding the contract method 0x58aeef93.
//
// Solidity: function DELEGATION_RULES_PROVIDER() view returns(bytes32)
func (_Delegation *DelegationCallerSession) DELEGATIONRULESPROVIDER() ([32]byte, error) {
	return _Delegation.Contract.DELEGATIONRULESPROVIDER(&_Delegation.CallOpts)
}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_Delegation *DelegationCaller) DELEGATIONRULESSESSION(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Delegation.contract.Call(opts, &out, "DELEGATION_RULES_SESSION")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_Delegation *DelegationSession) DELEGATIONRULESSESSION() ([32]byte, error) {
	return _Delegation.Contract.DELEGATIONRULESSESSION(&_Delegation.CallOpts)
}

// DELEGATIONRULESSESSION is a free data retrieval call binding the contract method 0xd1b43638.
//
// Solidity: function DELEGATION_RULES_SESSION() view returns(bytes32)
func (_Delegation *DelegationCallerSession) DELEGATIONRULESSESSION() ([32]byte, error) {
	return _Delegation.Contract.DELEGATIONRULESSESSION(&_Delegation.CallOpts)
}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_Delegation *DelegationCaller) DELEGATIONSTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Delegation.contract.Call(opts, &out, "DELEGATION_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_Delegation *DelegationSession) DELEGATIONSTORAGESLOT() ([32]byte, error) {
	return _Delegation.Contract.DELEGATIONSTORAGESLOT(&_Delegation.CallOpts)
}

// DELEGATIONSTORAGESLOT is a free data retrieval call binding the contract method 0xdd9b48cb.
//
// Solidity: function DELEGATION_STORAGE_SLOT() view returns(bytes32)
func (_Delegation *DelegationCallerSession) DELEGATIONSTORAGESLOT() ([32]byte, error) {
	return _Delegation.Contract.DELEGATIONSTORAGESLOT(&_Delegation.CallOpts)
}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_Delegation *DelegationCaller) DIAMONDOWNABLESTORAGESLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Delegation.contract.Call(opts, &out, "DIAMOND_OWNABLE_STORAGE_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_Delegation *DelegationSession) DIAMONDOWNABLESTORAGESLOT() ([32]byte, error) {
	return _Delegation.Contract.DIAMONDOWNABLESTORAGESLOT(&_Delegation.CallOpts)
}

// DIAMONDOWNABLESTORAGESLOT is a free data retrieval call binding the contract method 0x4ac3371e.
//
// Solidity: function DIAMOND_OWNABLE_STORAGE_SLOT() view returns(bytes32)
func (_Delegation *DelegationCallerSession) DIAMONDOWNABLESTORAGESLOT() ([32]byte, error) {
	return _Delegation.Contract.DIAMONDOWNABLESTORAGESLOT(&_Delegation.CallOpts)
}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_Delegation *DelegationCaller) GetRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Delegation.contract.Call(opts, &out, "getRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_Delegation *DelegationSession) GetRegistry() (common.Address, error) {
	return _Delegation.Contract.GetRegistry(&_Delegation.CallOpts)
}

// GetRegistry is a free data retrieval call binding the contract method 0x5ab1bd53.
//
// Solidity: function getRegistry() view returns(address)
func (_Delegation *DelegationCallerSession) GetRegistry() (common.Address, error) {
	return _Delegation.Contract.GetRegistry(&_Delegation.CallOpts)
}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_Delegation *DelegationCaller) IsRightsDelegated(opts *bind.CallOpts, delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	var out []interface{}
	err := _Delegation.contract.Call(opts, &out, "isRightsDelegated", delegatee_, delegator_, rights_)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_Delegation *DelegationSession) IsRightsDelegated(delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	return _Delegation.Contract.IsRightsDelegated(&_Delegation.CallOpts, delegatee_, delegator_, rights_)
}

// IsRightsDelegated is a free data retrieval call binding the contract method 0x54126b8f.
//
// Solidity: function isRightsDelegated(address delegatee_, address delegator_, bytes32 rights_) view returns(bool)
func (_Delegation *DelegationCallerSession) IsRightsDelegated(delegatee_ common.Address, delegator_ common.Address, rights_ [32]byte) (bool, error) {
	return _Delegation.Contract.IsRightsDelegated(&_Delegation.CallOpts, delegatee_, delegator_, rights_)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Delegation *DelegationCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Delegation.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Delegation *DelegationSession) Owner() (common.Address, error) {
	return _Delegation.Contract.Owner(&_Delegation.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Delegation *DelegationCallerSession) Owner() (common.Address, error) {
	return _Delegation.Contract.Owner(&_Delegation.CallOpts)
}

// DelegationInit is a paid mutator transaction binding the contract method 0xe50e4c43.
//
// Solidity: function __Delegation_init(address registry_) returns()
func (_Delegation *DelegationTransactor) DelegationInit(opts *bind.TransactOpts, registry_ common.Address) (*types.Transaction, error) {
	return _Delegation.contract.Transact(opts, "__Delegation_init", registry_)
}

// DelegationInit is a paid mutator transaction binding the contract method 0xe50e4c43.
//
// Solidity: function __Delegation_init(address registry_) returns()
func (_Delegation *DelegationSession) DelegationInit(registry_ common.Address) (*types.Transaction, error) {
	return _Delegation.Contract.DelegationInit(&_Delegation.TransactOpts, registry_)
}

// DelegationInit is a paid mutator transaction binding the contract method 0xe50e4c43.
//
// Solidity: function __Delegation_init(address registry_) returns()
func (_Delegation *DelegationTransactorSession) DelegationInit(registry_ common.Address) (*types.Transaction, error) {
	return _Delegation.Contract.DelegationInit(&_Delegation.TransactOpts, registry_)
}

// SetRegistry is a paid mutator transaction binding the contract method 0xa91ee0dc.
//
// Solidity: function setRegistry(address registry_) returns()
func (_Delegation *DelegationTransactor) SetRegistry(opts *bind.TransactOpts, registry_ common.Address) (*types.Transaction, error) {
	return _Delegation.contract.Transact(opts, "setRegistry", registry_)
}

// SetRegistry is a paid mutator transaction binding the contract method 0xa91ee0dc.
//
// Solidity: function setRegistry(address registry_) returns()
func (_Delegation *DelegationSession) SetRegistry(registry_ common.Address) (*types.Transaction, error) {
	return _Delegation.Contract.SetRegistry(&_Delegation.TransactOpts, registry_)
}

// SetRegistry is a paid mutator transaction binding the contract method 0xa91ee0dc.
//
// Solidity: function setRegistry(address registry_) returns()
func (_Delegation *DelegationTransactorSession) SetRegistry(registry_ common.Address) (*types.Transaction, error) {
	return _Delegation.Contract.SetRegistry(&_Delegation.TransactOpts, registry_)
}

// DelegationDelegationRegistryUpdatedIterator is returned from FilterDelegationRegistryUpdated and is used to iterate over the raw logs and unpacked data for DelegationRegistryUpdated events raised by the Delegation contract.
type DelegationDelegationRegistryUpdatedIterator struct {
	Event *DelegationDelegationRegistryUpdated // Event containing the contract specifics and raw log

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
func (it *DelegationDelegationRegistryUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegationDelegationRegistryUpdated)
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
		it.Event = new(DelegationDelegationRegistryUpdated)
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
func (it *DelegationDelegationRegistryUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegationDelegationRegistryUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegationDelegationRegistryUpdated represents a DelegationRegistryUpdated event raised by the Delegation contract.
type DelegationDelegationRegistryUpdated struct {
	Registry common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterDelegationRegistryUpdated is a free log retrieval operation binding the contract event 0x836360d1b094a7de3c3eab3d1185f3a5939467c23d4a12709dbdbf8c8d7e2f3b.
//
// Solidity: event DelegationRegistryUpdated(address registry)
func (_Delegation *DelegationFilterer) FilterDelegationRegistryUpdated(opts *bind.FilterOpts) (*DelegationDelegationRegistryUpdatedIterator, error) {

	logs, sub, err := _Delegation.contract.FilterLogs(opts, "DelegationRegistryUpdated")
	if err != nil {
		return nil, err
	}
	return &DelegationDelegationRegistryUpdatedIterator{contract: _Delegation.contract, event: "DelegationRegistryUpdated", logs: logs, sub: sub}, nil
}

// WatchDelegationRegistryUpdated is a free log subscription operation binding the contract event 0x836360d1b094a7de3c3eab3d1185f3a5939467c23d4a12709dbdbf8c8d7e2f3b.
//
// Solidity: event DelegationRegistryUpdated(address registry)
func (_Delegation *DelegationFilterer) WatchDelegationRegistryUpdated(opts *bind.WatchOpts, sink chan<- *DelegationDelegationRegistryUpdated) (event.Subscription, error) {

	logs, sub, err := _Delegation.contract.WatchLogs(opts, "DelegationRegistryUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegationDelegationRegistryUpdated)
				if err := _Delegation.contract.UnpackLog(event, "DelegationRegistryUpdated", log); err != nil {
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

// ParseDelegationRegistryUpdated is a log parse operation binding the contract event 0x836360d1b094a7de3c3eab3d1185f3a5939467c23d4a12709dbdbf8c8d7e2f3b.
//
// Solidity: event DelegationRegistryUpdated(address registry)
func (_Delegation *DelegationFilterer) ParseDelegationRegistryUpdated(log types.Log) (*DelegationDelegationRegistryUpdated, error) {
	event := new(DelegationDelegationRegistryUpdated)
	if err := _Delegation.contract.UnpackLog(event, "DelegationRegistryUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// DelegationInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the Delegation contract.
type DelegationInitializedIterator struct {
	Event *DelegationInitialized // Event containing the contract specifics and raw log

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
func (it *DelegationInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DelegationInitialized)
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
		it.Event = new(DelegationInitialized)
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
func (it *DelegationInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DelegationInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DelegationInitialized represents a Initialized event raised by the Delegation contract.
type DelegationInitialized struct {
	StorageSlot [32]byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
func (_Delegation *DelegationFilterer) FilterInitialized(opts *bind.FilterOpts) (*DelegationInitializedIterator, error) {

	logs, sub, err := _Delegation.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &DelegationInitializedIterator{contract: _Delegation.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0xdc73717d728bcfa015e8117438a65319aa06e979ca324afa6e1ea645c28ea15d.
//
// Solidity: event Initialized(bytes32 storageSlot)
func (_Delegation *DelegationFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *DelegationInitialized) (event.Subscription, error) {

	logs, sub, err := _Delegation.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DelegationInitialized)
				if err := _Delegation.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_Delegation *DelegationFilterer) ParseInitialized(log types.Log) (*DelegationInitialized, error) {
	event := new(DelegationInitialized)
	if err := _Delegation.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
