// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IProviderStorage} from "../../interfaces/storage/IProviderStorage.sol";

contract ProviderStorage is IProviderStorage {
    struct PRVDRStorage {
        uint256 providerMinimumStake;
        mapping(address => Provider) providers;
    }

    // Reward for this period will be limited by the stake
    uint128 constant PROVIDER_REWARD_LIMITER_PERIOD = 365 days;

    bytes32 public constant PROVIDER_STORAGE_SLOT = keccak256("diamond.standard.provider.storage");

    /** PUBLIC, GETTERS */
    function getProvider(address provider_) external view returns (Provider memory) {
        return providers(provider_);
    }

    function getProviderMinimumStake() public view returns (uint256) {
        return _getProviderStorage().providerMinimumStake;
    }

    function getIsProviderActive(address provider_) public view returns (bool) {
        return !providers(provider_).isDeleted;
    }

    /** INTERNAL, GETTERS */
    function providers(address provider_) internal view returns (Provider storage) {
        return _getProviderStorage().providers[provider_];
    }

    /** INTERNAL, SETTERS */
    function setProviderMinimumStake(uint256 providerMinimumStake_) internal {
        _getProviderStorage().providerMinimumStake = providerMinimumStake_;
    }

    /** PRIVATE */
    function _getProviderStorage() private pure returns (PRVDRStorage storage ds) {
        bytes32 slot_ = PROVIDER_STORAGE_SLOT;

        assembly {
            ds.slot := slot_
        }
    }
}
