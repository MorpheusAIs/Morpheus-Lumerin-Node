// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { SafeERC20, IERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

import { DiamondOwnableStorage } from "../presets/DiamondOwnableStorage.sol";

import { BidStorage } from "../storages/BidStorage.sol";
import { ProviderStorage } from "../storages/ProviderStorage.sol";

import { IProviderRegistry } from "../../interfaces/facets/IProviderRegistry.sol";

contract ProviderRegistry is IProviderRegistry, DiamondOwnableStorage, ProviderStorage, BidStorage {
  using SafeERC20 for IERC20;

  function __ProviderRegistry_init() external initializer(PROVIDER_STORAGE_SLOT) {}

  /// @notice Sets the minimum stake required for a provider
  function providerSetMinStake(uint256 providerMinimumStake_) external onlyOwner {
    setProviderMinimumStake(providerMinimumStake_);
    emit ProviderMinStakeUpdated(providerMinimumStake_);
  }

  /// @notice Registers a provider
  /// @param   providerAddress_      provider address
  /// @param   amount_  amount of stake to add
  /// @param   endpoint_  provider endpoint (host.com:1234)
  function providerRegister(address providerAddress_, uint256 amount_, string calldata endpoint_) external {
    if (!_ownerOrProvider(providerAddress_)) {
      revert NotOwnerOrProvider();
    }

    Provider memory provider_ = providerMap(providerAddress_);
    uint256 newStake_ = provider_.stake + amount_;
    if (newStake_ < providerMinimumStake()) {
      revert StakeTooLow();
    }

    getToken().safeTransferFrom(_msgSender(), address(this), amount_);

    // if we add stake to an existing provider the limiter period is not reset
    uint128 createdAt_ = provider_.createdAt;
    uint128 periodEnd_ = provider_.limitPeriodEnd;
    if (createdAt_ == 0) {
      setActiveProvider(providerAddress_, true);
      addProvider(providerAddress_);
      createdAt_ = uint128(block.timestamp);
      periodEnd_ = createdAt_ + PROVIDER_REWARD_LIMITER_PERIOD;
    } else if (provider_.isDeleted) {
      setActiveProvider(providerAddress_, true);
    }

    setProvider(
      providerAddress_,
      Provider(endpoint_, newStake_, createdAt_, periodEnd_, provider_.limitPeriodEarned, false)
    );

    emit ProviderRegisteredUpdated(providerAddress_);
  }

  /// @notice Deregisters a provider
  function providerDeregister(address provider_) external {
    if (!_ownerOrProvider(provider_)) {
      revert NotOwnerOrProvider();
    }
    if (!isProviderExists(provider_)) {
      revert ProviderNotFound();
    }
    if (!isProviderActiveBidsEmpty(provider_)) {
      revert ProviderHasActiveBids();
    }

    setActiveProvider(provider_, false);

    Provider storage provider = providerMap(provider_);
    uint256 withdrawable_ = _getWithdrawableStake(provider);
    provider.stake -= withdrawable_;
    provider.isDeleted = true;

    getToken().safeTransfer(_msgSender(), withdrawable_);

    emit ProviderDeregistered(provider_);
  }

  /// @notice Withdraws stake from a provider after it has been deregistered
  ///         Allows to withdraw the stake after provider reward period has ended
  function providerWithdrawStake(address provider_) external {
    Provider storage provider = providerMap(provider_);
    if (!provider.isDeleted) {
      revert ErrProviderNotDeleted();
    }
    if (provider.stake == 0) {
      revert ErrNoStake();
    }

    uint256 withdrawable_ = _getWithdrawableStake(provider);
    if (withdrawable_ == 0) {
      revert ErrNoWithdrawableStake();
    }

    provider.stake -= withdrawable_;

    getToken().safeTransfer(provider_, withdrawable_);

    emit ProviderWithdrawnStake(provider_, withdrawable_);
  }

  function isProviderExists(address provider_) public view returns (bool) {
    return providerMap(provider_).createdAt != 0;
  }

  /// @notice Returns the withdrawable stake for a provider
  /// @dev    If the provider already earned this period then withdrawable stake
  ///         is limited by the amount earning that remains in the current period.
  ///         It is done to prevent the provider from withdrawing and then staking
  ///         again from a different address, which bypasses the limitation.
  function _getWithdrawableStake(Provider memory provider_) private view returns (uint256) {
    if (uint128(block.timestamp) > provider_.limitPeriodEnd) {
      return provider_.stake;
    }

    return provider_.stake - provider_.limitPeriodEarned;
  }

  function _ownerOrProvider(address provider_) internal view returns (bool) {
    return _msgSender() == owner() || _msgSender() == provider_;
  }
}
