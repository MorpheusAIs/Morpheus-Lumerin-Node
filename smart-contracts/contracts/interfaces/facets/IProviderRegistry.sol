// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IProviderStorage} from "../storage/IProviderStorage.sol";

interface IProviderRegistry is IProviderStorage {
    event ProviderRegistered(address indexed provider);
    event ProviderDeregistered(address indexed provider);
    event ProviderMinimumStakeUpdated(uint256 providerMinimumStake);
    event ProviderWithdrawn(address indexed provider, uint256 amount);
    error ProviderStakeTooLow(uint256 amount, uint256 minAmount);
    error ProviderNotDeregistered();
    error ProviderNoStake();
    error ProviderNothingToWithdraw();
    error ProviderHasActiveBids();
    error ProviderNotFound();
    error ProviderHasAlreadyDeregistered();

    /**
     * The function to initialize the facet.
     */
    function __ProviderRegistry_init() external;

    /**
     * @notice The function to the minimum stake required for a provider
     * @param providerMinimumStake_ The minimal stake
     */
    function providerSetMinStake(uint256 providerMinimumStake_) external;

    /**
     * @notice The function to register the provider.
     * @param provider_ The provider address.
     * @param amount_ The amount of stake to add.
     * @param endpoint_ The provider endpoint (host.com:1234).
     */
    function providerRegister(address provider_, uint256 amount_, string calldata endpoint_) external;

    /**
     * @notice The function to deregister the provider.
     * @param provider_ The provider address.
     */
    function providerDeregister(address provider_) external;

    /**
     * @notice Returns provider earnings status and reward capacity for current limiter period (RFP M6).
     * @param provider_ The provider address.
     * @return stake Current total stake.
     * @return earnedThisPeriod Amount of MOR earned in the current limiter period.
     * @return remainingCapacity Remaining reward earning capacity before stake ceiling.
     * @return periodEnd Timestamp when the current 365-day limiter period ends.
     */
    function getProviderEarningsStatus(address provider_) external view returns (
        uint256 stake,
        uint256 earnedThisPeriod,
        uint256 remainingCapacity,
        uint128 periodEnd
    );
}
