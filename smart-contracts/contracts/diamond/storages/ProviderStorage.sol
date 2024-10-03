// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IProviderStorage} from "../../interfaces/storage/IProviderStorage.sol";

contract ProviderStorage is IProviderStorage {
    struct PRVDRStorage {
        uint256 providerMinimumStake;
        mapping(address => Provider) providers;
        mapping(address => bool) isProviderActive;
    }

    uint128 constant PROVIDER_REWARD_LIMITER_PERIOD = 365 days; // reward for this period will be limited by the stake

    bytes32 public constant PROVIDER_STORAGE_SLOT = keccak256("diamond.standard.provider.storage");

    function getProvider(address provider) external view returns (Provider memory) {
        return _getProviderStorage().providers[provider];
    }

    function providerMinimumStake() public view returns (uint256) {
        return _getProviderStorage().providerMinimumStake;
    }

    function setProviderActive(address provider, bool isActive) internal {
        _getProviderStorage().isProviderActive[provider] = isActive;
    }

    function setProvider(address provider, Provider memory provider_) internal {
        _getProviderStorage().providers[provider] = provider_;
    }

    function setProviderMinimumStake(uint256 _providerMinimumStake) internal {
        _getProviderStorage().providerMinimumStake = _providerMinimumStake;
    }

    function providers(address addr) internal view returns (Provider storage) {
        return _getProviderStorage().providers[addr];
    }

    function isProviderActive(address provider) internal view returns (bool) {
        return _getProviderStorage().isProviderActive[provider];
    }

    function _getProviderStorage() internal pure returns (PRVDRStorage storage _ds) {
        bytes32 slot_ = PROVIDER_STORAGE_SLOT;

        assembly {
            _ds.slot := slot_
        }
    }
}
