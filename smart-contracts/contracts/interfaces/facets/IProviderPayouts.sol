// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IProviderPayoutStorage} from "../storage/IProviderPayoutStorage.sol";

interface IProviderPayouts is IProviderPayoutStorage {
    error ProviderPayoutTargetInvalid();

    /**
     * The function to return the payout delegation right.
     * @return `keccak256("delegation.rules.payout")`.
     */
    function DELEGATION_RULES_PAYOUT() external pure returns (bytes32);

    /**
     * The function to initialize the facet.
     */
    function __ProviderPayouts_init() external;

    /**
     * The function to set the provider payout target.
     * A full delegation does not grant this right.
     * @param provider_ The provider address.
     * @param target_ The payout target. Zero or the provider address resets it to the default.
     */
    function setPayoutTarget(address provider_, address target_) external;

    /**
     * The function to return the provider payout target.
     * @param provider_ The provider address.
     * @return The effective payout target.
     */
    function getPayoutTarget(address provider_) external view returns (address);

    /**
     * The function to return the unpaid provider reward.
     * @param provider_ The provider address.
     * @return The unpaid reward amount.
     */
    function getProviderOwed(address provider_) external view returns (uint256);

    /**
     * The function to return the total unpaid provider rewards.
     * @return The total unpaid reward amount.
     */
    function getPendingProviderClaims() external view returns (uint256);

    /**
     * The function to return provider payouts and the accounting start timestamp.
     * @param provider_ The provider address.
     * @return The amount paid since the accounting start timestamp.
     * @return The accounting start timestamp.
     */
    function getProviderLifetimeClaimed(address provider_) external view returns (uint256, uint128);

    /**
     * The function to get session ID/
     * @param user_ The user address.
     * @param provider_ The provider address.
     * @param bidId_ The bid ID.
     * @param sessionNonce_ The session nounce.
     * @return The session ID.
     */
    function getSessionId(
        address user_,
        address provider_,
        bytes32 bidId_,
        uint256 sessionNonce_
    ) external pure returns (bytes32);

    /**
     * The function to returns the session end timestamp
     * @param amount_ The stake amount.
     * @param pricePerSecond_ The price per second.
     * @param openedAt_ The opened at timestamp.
     * @return The session end timestamp.
     */
    function getSessionEnd(uint256 amount_, uint256 pricePerSecond_, uint128 openedAt_) external view returns (uint128);

    /**
     * Returns stipend of user based on their stake.
     * (User session stake amount / MOR Supply without Compute) * (MOR Compute Supply / 100)
     * @param amount_ The amount of tokens.
     * @param timestamp_ The timestamp when the TX executes.
     * @return The stipend amount.
     */
    function stakeToStipend(uint256 amount_, uint128 timestamp_) external view returns (uint256);

    /**
     * Returns stake of user based on their stipend.
     * @param stipend_ The stake amount.
     * @param timestamp_ The timestamp when the TX executed.
     * @return The stake amount.
     */
    function stipendToStake(uint256 stipend_, uint128 timestamp_) external view returns (uint256);

    /**
     * The function to return available and locked amount of the user tokens.
     * @param user_ The user address.
     * @param iterations_ The loop interaction amount.
     * @return available_ The available to withdraw.
     * @return hold_ The locked amount.
     */
    function getUserStakesOnHold(
        address user_,
        uint8 iterations_
    ) external view returns (uint256 available_, uint256 hold_);

    /**
     * Returns today's budget in MOR. 1%.
     * @param timestamp_ The timestamp when the TX executed.
     * @return Today's budget.
     */
    function getTodaysBudget(uint128 timestamp_) external view returns (uint256);

    /**
     * Returns today's compute balance in MOR without claimed amount.
     * @param timestamp_ The timestamp when the TX executed.
     * @return Today's unclaimed compute balance.
     */
    function getComputeBalance(uint128 timestamp_) external view returns (uint256);

    /**
     * Total amount of MOR tokens that were distributed across all pools
     * without compute pool rewards and with compute claimed rewards.
     * @param timestamp_ The timestamp when the TX executed.
     * @return The total distributed MOR supply.
     */
    function totalMORSupply(uint128 timestamp_) external view returns (uint256);

    /**
     * The function to return the timestamp on the start of day
     * @param timestamp_ The timestamp when the TX executed.
     * @return The timestamp at the start of the day.
     */
    function startOfTheDay(uint128 timestamp_) external pure returns (uint128);
}
