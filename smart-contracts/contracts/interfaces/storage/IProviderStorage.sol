// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IProviderStorage {
    /**
     * The structure that stores the provider data.
     * @param endpoint Example 'domain.com:1234'. Readonly for now.
     * @param stake The stake amount.
     * @param createdAt The timestamp when the provider is created.
     * @param limitPeriodEnd Timestamp that indicate limit period end for provider rewards.
     * @param limitPeriodEarned The amount of tokens that provider can receive before `limitPeriodEnd`.
     * @param isDeleted The provider status.
     */
    struct Provider {
        string endpoint;
        uint256 stake;
        uint128 createdAt;
        uint128 limitPeriodEnd;
        uint256 limitPeriodEarned;
        bool isDeleted;
    }

    /**
     * The function returns provider structure.
     * @param provider_ Provider address
     */
    function getProvider(address provider_) external view returns (Provider memory);

    /**
     * The function returns provider minimal stake.
     */
    function getProviderMinimumStake() external view returns (uint256);

    /**
     * The function returns list of active providers.
     * @param offset_ Offset for the pagination.
     * @param limit_ Number of entities to return.
     */
    function getActiveProviders(uint256 offset_, uint256 limit_) external view returns (address[] memory);

    /**
     * The function returns provider status, active or not.
     * @param provider_ Provider address
     */
    function getIsProviderActive(address provider_) external view returns (bool);
}
