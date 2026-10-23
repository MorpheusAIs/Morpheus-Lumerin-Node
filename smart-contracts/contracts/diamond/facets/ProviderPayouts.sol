// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IDelegateRegistry} from "../../interfaces/deps/IDelegateRegistry.sol";
import {IProviderPayouts} from "../../interfaces/facets/IProviderPayouts.sol";
import {OwnableDiamondStorage} from "../presets/OwnableDiamondStorage.sol";
import {SessionSettlement} from "../presets/SessionSettlement.sol";
import {ProviderPayoutStorage} from "../storages/ProviderPayoutStorage.sol";

contract ProviderPayouts is IProviderPayouts, OwnableDiamondStorage, SessionSettlement {
    function DELEGATION_RULES_PAYOUT() external pure returns (bytes32) {
        return ProviderPayoutStorage._DELEGATION_RULES_PAYOUT;
    }

    function __ProviderPayouts_init() external initializer(PROVIDER_PAYOUT_STORAGE_SLOT) {
        _getProviderPayoutsStorage().lifetimeClaimedStartedAt = uint128(block.timestamp);
    }

    function setPayoutTarget(address provider_, address target_) external {
        address sender_ = _msgSender();

        if (sender_ != provider_ && !_hasExplicitPayoutRight(provider_, sender_)) {
            revert InsufficientRightsForOperation(provider_, sender_);
        }
        if (target_ == address(this)) {
            revert ProviderPayoutTargetInvalid();
        }

        ProviderPayoutsStorage storage payoutStorage = _getProviderPayoutsStorage();
        payoutStorage.payoutTargets[provider_] = target_ == provider_ ? address(0) : target_;

        emit ProviderPayoutTargetUpdated(provider_, _getPayoutTarget(provider_));
    }

    function getPayoutTarget(address provider_) external view returns (address) {
        return _getPayoutTarget(provider_);
    }

    function getProviderOwed(address provider_) external view returns (uint256) {
        return _getProviderPayoutsStorage().providerOwed[provider_];
    }

    function getPendingProviderClaims() external view returns (uint256) {
        return _getProviderPayoutsStorage().pendingProviderClaims;
    }

    function getProviderLifetimeClaimed(address provider_) external view returns (uint256, uint128) {
        ProviderPayoutsStorage storage payoutStorage = _getProviderPayoutsStorage();
        return (payoutStorage.lifetimeClaimed[provider_], payoutStorage.lifetimeClaimedStartedAt);
    }

    function getSessionId(
        address user_,
        address provider_,
        bytes32 bidId_,
        uint256 sessionNonce_
    ) external pure returns (bytes32) {
        return _getSessionId(user_, provider_, bidId_, sessionNonce_);
    }

    function getSessionEnd(uint256 amount_, uint256 pricePerSecond_, uint128 openedAt_) external view returns (uint128) {
        return _getSessionEnd(amount_, pricePerSecond_, openedAt_);
    }

    function stakeToStipend(uint256 amount_, uint128 timestamp_) external view returns (uint256) {
        return _stakeToStipend(amount_, timestamp_);
    }

    function stipendToStake(uint256 stipend_, uint128 timestamp_) external view returns (uint256) {
        return _stipendToStake(stipend_, timestamp_);
    }

    function getUserStakesOnHold(
        address user_,
        uint8 iterations_
    ) external view returns (uint256 available_, uint256 hold_) {
        OnHold[] memory onHold = _getSessionsStorage().userStakesOnHold[user_];
        iterations_ = iterations_ > onHold.length ? uint8(onHold.length) : iterations_;

        for (uint256 i = 0; i < onHold.length; i++) {
            uint256 amount = onHold[i].amount;

            if (block.timestamp < onHold[i].releaseAt) {
                hold_ += amount;
            } else {
                available_ += amount;
            }
        }
    }

    function getTodaysBudget(uint128 timestamp_) external view returns (uint256) {
        return _getComputeBalance(timestamp_) / 100;
    }

    function getComputeBalance(uint128 timestamp_) external view returns (uint256) {
        return _getComputeBalance(timestamp_);
    }

    function totalMORSupply(uint128 timestamp_) external view returns (uint256) {
        return _totalMORSupply(timestamp_);
    }

    function startOfTheDay(uint128 timestamp_) external pure returns (uint128) {
        return _startOfTheDay(timestamp_);
    }

    /**
     * @dev Only an explicit payout delegation counts: an empty-rights (full) delegation does not grant
     * this right, and it does not take away an explicit one. The provider controls its outgoing list.
     */
    function _hasExplicitPayoutRight(address provider_, address delegate_) private view returns (bool) {
        IDelegateRegistry.Delegation[] memory delegations_ = IDelegateRegistry(_getDelegationStorage().registry)
            .getOutgoingDelegations(provider_);

        for (uint256 i = 0; i < delegations_.length; ) {
            IDelegateRegistry.Delegation memory delegation_ = delegations_[i];
            if (
                delegation_.to == delegate_ &&
                delegation_.rights == ProviderPayoutStorage._DELEGATION_RULES_PAYOUT &&
                (delegation_.type_ == IDelegateRegistry.DelegationType.ALL ||
                    (delegation_.type_ == IDelegateRegistry.DelegationType.CONTRACT &&
                        delegation_.contract_ == address(this)))
            ) {
                return true;
            }
            unchecked {
                i++;
            }
        }

        return false;
    }
}
