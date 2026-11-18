// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IProviderPayoutStorage} from "../../interfaces/storage/IProviderPayoutStorage.sol";

contract ProviderPayoutStorage is IProviderPayoutStorage {
    struct ProviderPayoutsStorage {
        mapping(address provider => address payoutTarget) payoutTargets;
        mapping(address provider => uint256 amount) providerOwed;
        uint256 pendingProviderClaims;
        mapping(address provider => uint256 amount) lifetimeClaimed;
        uint128 lifetimeClaimedStartedAt;
    }

    bytes32 internal constant PROVIDER_PAYOUT_STORAGE_SLOT = keccak256("diamond.standard.provider.payout.storage");
    bytes32 internal constant _DELEGATION_RULES_PAYOUT = keccak256("delegation.rules.payout");

    function _getProviderPayoutsStorage() internal pure returns (ProviderPayoutsStorage storage ds) {
        bytes32 slot_ = PROVIDER_PAYOUT_STORAGE_SLOT;

        assembly {
            ds.slot := slot_
        }
    }

    function _getPayoutTarget(address provider_) internal view returns (address) {
        address payoutTarget_ = _getProviderPayoutsStorage().payoutTargets[provider_];
        return payoutTarget_ == address(0) ? provider_ : payoutTarget_;
    }
}
