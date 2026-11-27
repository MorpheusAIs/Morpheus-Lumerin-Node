// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IDelegateStakingStorage} from "../storage/IDelegateStakingStorage.sol";

interface IDelegateStakingFunding is IDelegateStakingStorage {
    /**
     * @notice Approves a funder to fund and participate in a hot wallet's pool.
     * @param hot_ The hot wallet granting the approval.
     * @param funder_ The funder being approved.
     */
    function approvePoolFunder(address hot_, address funder_) external;

    /**
     * @notice Revokes a funder's approval to fund and participate in a hot wallet's pool.
     * @param hot_ The hot wallet revoking the approval.
     * @param funder_ The funder whose approval is revoked.
     */
    function revokePoolFunderApproval(address hot_, address funder_) external;

    /**
     * @notice Creates or updates the caller's funding grant for a hot wallet.
     * @param hot_ The hot wallet receiving the grant.
     * @param cap_ The cumulative funding cap.
     * @param expiry_ The grant expiry timestamp.
     */
    function grantStakingAllowance(address hot_, uint256 cap_, uint128 expiry_) external;

    /**
     * @notice Revokes the caller's funding grant for a hot wallet.
     * @param hot_ The hot wallet whose grant is revoked.
     */
    function revokeStakingAllowance(address hot_) external;

    /**
     * @notice Funds the caller's approved live grant.
     * @param hot_ The hot wallet receiving the funding.
     * @param amount_ The amount to fund.
     */
    function fundStakingAllowance(address hot_, uint256 amount_) external;

    /**
     * @notice Withdraws available principal and queues any unpaid remainder.
     * A repeated request merges into the funder's queued entry and moves it to the queue tail.
     * @param hot_ The hot wallet whose pool holds the principal.
     * @param funder_ The funder whose principal is withdrawn.
     * @param amount_ The amount to withdraw.
     */
    function requestStakingWithdrawal(address hot_, address funder_, uint256 amount_) external;

    /**
     * @notice Cancels a funder's pending withdrawal.
     * @param hot_ The hot wallet whose pool contains the withdrawal.
     * @param funder_ The funder whose withdrawal is cancelled.
     */
    function cancelStakingWithdrawal(address hot_, address funder_) external;

    /**
     * @notice Services up to `maxEntries_` pending withdrawals.
     * @param hot_ The hot wallet whose withdrawal queue is serviced.
     * @param maxEntries_ The maximum number of queue entries to service.
     * @return The number of queue entries serviced.
     */
    function servePoolWithdrawals(address hot_, uint256 maxEntries_) external returns (uint256);

    /**
     * @notice Releases up to `maxBuckets_` matured lock buckets and services withdrawals.
     * @param hot_ The hot wallet whose lock buckets are released.
     * @param maxBuckets_ The maximum number of lock buckets to release.
     * @return The number of lock buckets released.
     */
    function releasePoolLocks(address hot_, uint256 maxBuckets_) external returns (uint256);

    /**
     * @notice Removes up to `maxEntries_` ineligible funders from the active list.
     * @param hot_ The hot wallet whose active funder list is purged.
     * @param maxEntries_ The maximum number of active entries to inspect.
     * @return The number of inactive funders removed.
     */
    function purgeInactiveFunders(address hot_, uint256 maxEntries_) external returns (uint256);

    /**
     * @notice Sets the active-funder admission limit; zero restores the default.
     * @param value_ The new limit.
     */
    function setMaxActiveFunders(uint256 value_) external;

    /**
     * @notice Sets the minimum principal after funding; zero restores the default.
     * @param value_ The new minimum principal.
     */
    function setMinPrincipal(uint256 value_) external;

    /**
     * @notice Sets the expired-session settlement grace; zero restores the default.
     * @param value_ The new settlement grace period.
     */
    function setSettlementGrace(uint256 value_) external;

    /**
     * @notice Sets the automatic withdrawal-service limit; zero restores the default.
     * @param value_ The new automatic service limit.
     */
    function setMaxAutoService(uint256 value_) external;

    /**
     * @notice Sets the automatic matured-bucket release limit; zero restores the default.
     * @param value_ The new automatic release limit.
     */
    function setMaxAutoReleaseDays(uint256 value_) external;

    /**
     * @notice Returns a funder's grant and accounting state.
     * @param hot_ The hot wallet whose pool contains the grant.
     * @param funder_ The funder whose allowance is returned.
     * @return The funder's allowance and accounting state.
     */
    function getStakingAllowance(address hot_, address funder_) external view returns (StakingAllowance memory);

    /**
     * @notice Returns whether a funder is approved to participate in a hot wallet's pool.
     * @param hot_ The hot wallet whose approval is queried.
     * @param funder_ The funder whose approval is queried.
     * @return Whether the funder is approved, including implicit self-escrow approval.
     */
    function isPoolFunderApproved(address hot_, address funder_) external view returns (bool);

    /**
     * @notice Returns currently drawable stake after pending withdrawals and grant capacity.
     * @param hot_ The hot wallet whose available pool stake is returned.
     * @return The stake drawable by the next pool session.
     */
    function getAvailablePoolStake(address hot_) external view returns (uint256);

    /**
     * @notice Returns a page of all funders that have successfully funded a hot wallet's pool.
     * @param hot_ The hot wallet whose funders are returned.
     * @param offset_ The zero-based page offset.
     * @param limit_ The maximum number of funders to return.
     * @return The page of funders.
     * @return The total number of registered funders.
     */
    function getPoolFunders(
        address hot_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (PoolFunder[] memory, uint256);

    /**
     * @notice Returns a funder's pending withdrawal amount.
     * @param hot_ The hot wallet whose pool contains the withdrawal.
     * @param funder_ The funder whose pending amount is returned.
     * @return The pending withdrawal amount.
     */
    function getStakingWithdrawal(address hot_, address funder_) external view returns (uint256);

    /**
     * @notice Returns a cursor-based page of the withdrawal queue in FIFO order.
     * @param hot_ The hot wallet whose withdrawal queue is returned.
     * @param cursor_ The first funder to include, or zero for the queue head.
     * @param limit_ The maximum number of withdrawals to return.
     * @return The page of queued withdrawals.
     * @return The next cursor, or zero when the page reaches the queue tail.
     * @return The total number of queued withdrawals.
     */
    function getPoolWithdrawalQueue(
        address hot_,
        address cursor_,
        uint256 limit_
    ) external view returns (PoolWithdrawal[] memory, address, uint256);

    /**
     * @notice Returns aggregate pool balances.
     * @param hot_ The hot wallet whose pool balances are returned.
     * @return The pool's actual aggregate balances.
     */
    function getPoolBalances(address hot_) external view returns (PoolBalances memory);

    /**
     * @notice Returns a page of unreleased lock buckets.
     * @param hot_ The hot wallet whose lock buckets are returned.
     * @param offset_ The zero-based offset among unreleased buckets.
     * @param limit_ The maximum number of buckets to return.
     * @return The page of unreleased lock buckets.
     * @return The total number of unreleased lock buckets.
     */
    function getPoolLockBuckets(
        address hot_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (PoolLockBucket[] memory, uint256);

    /**
     * @notice Returns pool funding attribution for a session.
     * @param sessionId_ The session ID.
     * @return The pool session and its funding draws.
     */
    function getPoolSession(bytes32 sessionId_) external view returns (PoolSession memory);

    /**
     * @notice Returns effective delegate-staking parameters.
     * @return The effective delegate-staking parameters.
     */
    function getDelegateStakingParameters() external view returns (DelegateStakingParameters memory);

    /**
     * @notice Returns hard limits for delegate-staking parameters.
     * @return The hard delegate-staking parameter limits.
     */
    function getDelegateStakingLimits() external pure returns (DelegateStakingLimits memory);

    /**
     * @notice Returns the delegation right for managing staking withdrawals.
     * @return The delegate-staking delegation right.
     */
    function getDelegateStakingDelegationRight() external pure returns (bytes32);
}
