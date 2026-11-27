// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IDelegateStakingStorage} from "../storage/IDelegateStakingStorage.sol";

interface IDelegateStaking is IDelegateStakingStorage {
    event SessionOpened(address indexed user, bytes32 indexed sessionId, address indexed providerId);
    event SessionClosed(address indexed user, bytes32 indexed sessionId, address indexed providerId);

    /**
     * @notice Opens a stake-pool session backed by the hot wallet's pool.
     * @param hot_ The hot wallet whose pool backs the session.
     * @param amount_ The stake amount to draw from the pool.
     * @param approvalEncoded_ The provider approval.
     * @param signature_ The provider signature.
     * @return The opened session ID.
     */
    function openSessionFromPool(
        address hot_,
        uint256 amount_,
        bytes calldata approvalEncoded_,
        bytes calldata signature_
    ) external returns (bytes32);

    /**
     * @notice Closes a pool session using a provider receipt.
     * @param receiptEncoded_ The provider receipt.
     * @param signature_ The provider signature.
     */
    function closePoolSession(bytes calldata receiptEncoded_, bytes calldata signature_) external;

    /**
     * @notice Permissionlessly settles an expired pool session after its grace period.
     * @param sessionId_ The session ID.
     */
    function settleExpiredSession(bytes32 sessionId_) external;
}
