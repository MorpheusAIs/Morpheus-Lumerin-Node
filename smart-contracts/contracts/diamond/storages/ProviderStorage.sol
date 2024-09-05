// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { IProviderStorage } from "../../interfaces/storage/IProviderStorage.sol";

contract ProviderStorage is IProviderStorage {
  struct PRVDRStorage {
    mapping(address => Provider) providerMap; // provider address => Provider
    address[] providers; // all providers ids
    mapping(address => bool) activeProviders;
    uint256 providerMinimumStake;
  }

  uint128 constant PROVIDER_REWARD_LIMITER_PERIOD = 365 days; // reward for this period will be limited by the stake

  bytes32 public constant PROVIDER_STORAGE_SLOT = keccak256("diamond.standard.provider.storage");

  function getProvider(address provider) external view returns (Provider memory) {
    return _getProviderStorage().providerMap[provider];
  }

  function setActiveProvider(address provider, bool isActive) internal {
    _getProviderStorage().activeProviders[provider] = isActive;
  }

  function addProvider(address provider) internal {
    _getProviderStorage().providers.push(provider);
  }

  function setProvider(address provider, Provider memory provider_) internal {
    _getProviderStorage().providerMap[provider] = provider_;
  }

  function setProviderMinimumStake(uint256 _providerMinimumStake) internal {
    _getProviderStorage().providerMinimumStake = _providerMinimumStake;
  }

  function providerMap(address addr) internal view returns (Provider storage) {
    return _getProviderStorage().providerMap[addr];
  }

  function providers(uint256 index) internal view returns (address) {
    return _getProviderStorage().providers[index];
  }

  function isProviderActive(address provider) internal view returns (bool) {
    return _getProviderStorage().activeProviders[provider];
  }

  function providerMinimumStake() internal view returns (uint256) {
    return _getProviderStorage().providerMinimumStake;
  }

  function providerExists(address provider) internal view returns (bool) {
    return _getProviderStorage().activeProviders[provider];
  }

  function _getProviderStorage() internal pure returns (PRVDRStorage storage _ds) {
    bytes32 slot_ = PROVIDER_STORAGE_SLOT;

    assembly {
      _ds.slot := slot_
    }
  }
}
