// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {SafeERC20, IERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {EnumerableSet} from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";

import {OwnableDiamondStorage} from "../presets/OwnableDiamondStorage.sol";

import {BidStorage} from "../storages/BidStorage.sol";
import {ProviderStorage} from "../storages/ProviderStorage.sol";
import {DelegationStorage} from "../storages/DelegationStorage.sol";

import {IProviderRegistry} from "../../interfaces/facets/IProviderRegistry.sol";
import {IDelegateRegistry} from "../../interfaces/deps/IDelegateRegistry.sol";

contract ProviderRegistry is IProviderRegistry, OwnableDiamondStorage, ProviderStorage, BidStorage, DelegationStorage {
    using EnumerableSet for EnumerableSet.AddressSet;
    using SafeERC20 for IERC20;

    function __ProviderRegistry_init() external initializer(PROVIDERS_STORAGE_SLOT) {}

    function providerSetMinStake(uint256 providerMinimumStake_) external onlyOwner {
        PovidersStorage storage providersStorage = _getProvidersStorage();
        providersStorage.providerMinimumStake = providerMinimumStake_;

        emit ProviderMinimumStakeUpdated(providerMinimumStake_);
    }

    function providerRegister(address provider_, uint256 amount_, string calldata endpoint_) external {
        _validateDelegatee(_msgSender(), provider_, DELEGATION_RULES_PROVIDER);

        BidsStorage storage bidsStorage = _getBidsStorage();
        if (amount_ > 0) {
            IERC20(bidsStorage.token).safeTransferFrom(provider_, address(this), amount_);
        }

        PovidersStorage storage providersStorage = _getProvidersStorage();
        Provider storage provider = providersStorage.providers[provider_];

        uint256 newStake_ = provider.stake + amount_;
        uint256 minStake_ = providersStorage.providerMinimumStake;
        if (newStake_ < minStake_) {
            revert ProviderStakeTooLow(newStake_, minStake_);
        }

        if (provider.createdAt == 0) {
            provider.createdAt = uint128(block.timestamp);
            provider.limitPeriodEnd = uint128(block.timestamp) + PROVIDER_REWARD_LIMITER_PERIOD;
        } else if (provider.isDeleted) {
            provider.isDeleted = false;
        }

        provider.endpoint = endpoint_;
        provider.stake = newStake_;

        providersStorage.activeProviders.add(provider_);

        emit ProviderRegistered(provider_);
    }

    function providerDeregister(address provider_) external {
        _validateDelegatee(_msgSender(), provider_, DELEGATION_RULES_PROVIDER);

        PovidersStorage storage providersStorage = _getProvidersStorage();
        Provider storage provider = providersStorage.providers[provider_];

        if (provider.createdAt == 0) {
            revert ProviderNotFound();
        }
        if (!_isProviderActiveBidsEmpty(provider_)) {
            revert ProviderHasActiveBids();
        }
        if (provider.isDeleted) {
            revert ProviderHasAlreadyDeregistered();
        }

        uint256 withdrawAmount_ = _getWithdrawAmount(provider);

        provider.stake -= withdrawAmount_;
        provider.isDeleted = true;

        providersStorage.activeProviders.remove(provider_);

        if (withdrawAmount_ > 0) {
            BidsStorage storage bidsStorage = _getBidsStorage();
            IERC20(bidsStorage.token).safeTransfer(provider_, withdrawAmount_);
        }

        emit ProviderDeregistered(provider_);
    }

    /**
     * @notice Returns the withdrawable stake for a provider
     * @dev If the provider already earned this period then withdrawable stake
     * is limited by the amount earning that remains in the current period.
     * It is done to prevent the provider from withdrawing and then staking
     * again from a different address, which bypasses the limitation.
     */
    function _getWithdrawAmount(Provider storage provider) private view returns (uint256) {
        if (block.timestamp > provider.limitPeriodEnd) {
            return provider.stake;
        }

        return provider.stake - provider.limitPeriodEarned;
    }
}
