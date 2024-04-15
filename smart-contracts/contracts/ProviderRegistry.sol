// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;
import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import {AddressSet} from "./KeySet.sol";
import "hardhat/console.sol";

contract ProviderRegistry is OwnableUpgradeable {
  using AddressSet for AddressSet.Set;

  struct Provider {
    string endpoint; // example 'domain.com:1234'
    uint256 stake; // stake amount
    uint128 timestamp; // timestamp of the registration
    bool isDeleted;
  }

  error StakeTooLow();
  error NotSenderOrOwner();

  event RegisteredUpdated(address indexed provider);
  event Deregistered(address indexed provider);
  event MinStakeUpdated(uint256 newStake);

  // state
  uint256 public minStake;
  ERC20 public token;

  // providers storage
  mapping(address => Provider) public map; // provider address => Provider 
  address[] public providers; // all providers ids
  AddressSet.Set activeProviders; // active providers ids

  function initialize(address _token) public initializer {
      OwnableUpgradeable.__Ownable_init();
      token = ERC20(_token);
  }

  function getIds() public view returns (address[] memory) {
      return activeProviders.keys();
  }

  function getCount() public view returns (uint count) {
      return activeProviders.count();
  }

  function getByIndex(uint index) public view returns (address addr, Provider memory provider) {
      addr = activeProviders.keyAtIndex(index);
      return (addr, map[addr]);
  }

  function getAll() public view returns (address[] memory, Provider[] memory) {
      uint256 count = activeProviders.count();
      address[] memory _addrs = new address[](count);
      Provider[] memory _providers = new Provider[](count);
      
      for (uint i = 0; i < count; i++) {
          address addr = activeProviders.keyAtIndex(i);
          _addrs[i] = addr;
          _providers[i] = map[addr];
      }

      return (_addrs, _providers);
  }

  // registers new provider or updates existing
  function register(address addr, uint256 addStake, string memory endpoint) public senderOrOwner(addr) {
    Provider memory provider = map[addr];
    uint256 newStake = provider.stake + addStake;
    if (newStake < minStake) {
        revert StakeTooLow();
    }
    if (provider.timestamp == 0) {
        activeProviders.insert(addr);
        providers.push(addr);
    } else {
        _senderOrOwner(addr);
    }

    map[addr] = Provider(endpoint, newStake, uint128(block.timestamp), false);

    emit RegisteredUpdated(addr);
    token.transferFrom(_msgSender(), address(this), addStake); // reverts with ERC20InsufficientAllowance
  }

  function deregister(address addr) public senderOrOwner(addr) {
    activeProviders.remove(addr);
    
    map[addr].isDeleted = true;
    uint256 stake = map[addr].stake;
    
    emit Deregistered(addr);
    token.transfer(addr, stake);
  }

  function setMinStake(uint256 _minStake) public onlyOwner {
    minStake = _minStake;
    emit MinStakeUpdated(_minStake);
  }

  function exists(address addr) public view returns (bool) {
    return activeProviders.exists(addr);
  }

  modifier senderOrOwner(address addr) {
    _senderOrOwner(addr);
    _;
  }

  function _senderOrOwner(address resourceOwner) internal view {
    if (_msgSender() != resourceOwner && _msgSender() != owner()) {
      revert NotSenderOrOwner();
    }
  }
}
