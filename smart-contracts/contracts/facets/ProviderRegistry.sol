// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;
import { OwnableUpgradeable } from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import { ERC20 } from "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import { AddressSet } from "../libraries/KeySet.sol";
import { AppStorage, Provider } from "../AppStorage.sol";
import { LibOwner } from '../libraries/LibOwner.sol';

contract ProviderRegistry {
  using AddressSet for AddressSet.Set;
  AppStorage internal s;

  event ProviderRegisteredUpdated(address indexed provider);
  event ProviderDeregistered(address indexed provider);
  event ProviderMinStakeUpdated(uint256 newStake);
  
  error StakeTooLow();

  function providerMap(address addr) public view returns (string memory, uint256, uint128, bool) {
    Provider memory provider = s.providerMap[addr];
    return (provider.endpoint, provider.stake, provider.timestamp, provider.isDeleted);
  }

  function providers(uint256 index) public view returns (address) {
      return s.providers[index];
  }

  function providerGetIds() public view returns (address[] memory) {
      return s.activeProviders.keys();
  }

  function providerGetCount() public view returns (uint count) {
      return s.activeProviders.count();
  }

  function providerGetByIndex(uint index) public view returns (address addr, Provider memory provider) {
      addr = s.activeProviders.keyAtIndex(index);
      return (addr, s.providerMap[addr]);
  }

  function providerGetAll() public view returns (address[] memory, Provider[] memory) {
      uint256 count = s.activeProviders.count();
      address[] memory _addrs = new address[](count);
      Provider[] memory _providers = new Provider[](count);
      
      for (uint i = 0; i < count; i++) {
          address addr = s.activeProviders.keyAtIndex(i);
          _addrs[i] = addr;
          _providers[i] = s.providerMap[addr];
      }

      return (_addrs, _providers);
  }

  // registers new provider or updates existing
  function providerRegister(address addr, uint256 addStake, string memory endpoint) public {
    LibOwner._senderOrOwner(addr);
    Provider memory provider = s.providerMap[addr];
    uint256 newStake = provider.stake + addStake;
    if (newStake < s.providerMinStake) {
        revert StakeTooLow();
    }
    if (provider.timestamp == 0) {
        s.activeProviders.insert(addr);
        s.providers.push(addr);
    } else {
        LibOwner._senderOrOwner(addr);
    }

    s.providerMap[addr] = Provider(endpoint, newStake, uint128(block.timestamp), false);

    emit ProviderRegisteredUpdated(addr);

    s.token.transferFrom(msg.sender, address(this), addStake); // reverts with ERC20InsufficientAllowance
  }

  function providerDeregister(address addr) public {
    LibOwner._senderOrOwner(addr);
    s.activeProviders.remove(addr);
    
    s.providerMap[addr].isDeleted = true;
    uint256 stake = s.providerMap[addr].stake;
    
    emit ProviderDeregistered(addr);
    s.token.transfer(addr, stake);
  }

  function providerSetMinStake(uint256 _minStake) public {
    LibOwner._onlyOwner();
    s.providerMinStake = _minStake;
    emit ProviderMinStakeUpdated(_minStake);
  }

  function providrerExists(address addr) public view returns (bool) {
    return s.activeProviders.exists(addr);
  }

  function providerMinStake() public view returns (uint256) {
    return s.providerMinStake;
  }
}
