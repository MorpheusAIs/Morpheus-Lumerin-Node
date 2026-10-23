// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ISessionStorage} from "../storage/ISessionStorage.sol";

interface ISessionRouter is ISessionStorage {
    event SessionOpened(address indexed user, bytes32 indexed sessionId, address indexed providerId);
    event SessionClosed(address indexed user, bytes32 indexed sessionId, address indexed providerId);
    event UserWithdrawn(address indexed user, uint256 amount_);
    error SessionProviderSignatureMismatch();
    error SessionApproveExpired();
    error SessionApprovedForAnotherChainId();
    error SessionDuplicateApproval();
    error SessionApprovedForAnotherUser();
    error SessionReceiptForAnotherChainId();
    error SessionReceiptExpired();
    error SessionTooShort();
    error SessionAlreadyClosed();
    error SessionNotEndedOrNotExist();
    error SessionProviderNothingToClaimInThisPeriod();
    error SessionBidNotFound();
    error SessionPoolIndexOutOfBounds();
    error SessionUserAmountToWithdrawIsZero();
    error SessionMaxDurationTooShort();
    error SessionStakeTooLow();
    error SessionPoolSessionCannotBeClosed();

    /**
     * The function to initialize the facet.
     * @param fundingAccount_ The funding address (treaasury)
     * @param maxSessionDuration_ The max session duration
     * @param pools_ The pools data
     */
    function __SessionRouter_init(
        address fundingAccount_,
        uint128 maxSessionDuration_,
        Pool[] calldata pools_
    ) external;

    /**
     * @notice Sets distibution pool configuration
     * @dev parameters should be the same as in Ethereum L1 Distribution contract
     * @dev at address 0x47176B2Af9885dC6C4575d4eFd63895f7Aaa4790
     * @dev call 'Distribution.pools(3)' where '3' is a poolId.
     * @param index_ The pool index.
     * @param pool_ The pool data.
     */
    function setPoolConfig(uint256 index_, Pool calldata pool_) external;

    /**
     * The function to set the max session duration.
     * @param maxSessionDuration_ The max session duration.
     */
    function setMaxSessionDuration(uint128 maxSessionDuration_) external;

    /**
     * The function to open the session.
     * @param user_ The user address.
     * @param amount_ The stake amount.
     * @param isDirectPaymentFromUser_ If active, provider rewarded from the user stake.
     * @param approvalEncoded_ Provider approval.
     * @param signature_ Provider signature.
     */
    function openSession(
        address user_,
        uint256 amount_,
        bool isDirectPaymentFromUser_,
        bytes calldata approvalEncoded_,
        bytes calldata signature_
    ) external returns (bytes32);

    /**
     * The function to close session.
     * @param receiptEncoded_ Provider receipt
     * @param signature_ Provider signature
     */
    function closeSession(bytes calldata receiptEncoded_, bytes calldata signature_) external;

    /**
     * Allows providers to receive their funds after the end or closure of the session.
     * @param sessionId_ The session ID.
     */
    function claimForProvider(bytes32 sessionId_) external;

    /**
     * The function to withdraw user stakes.
     * @param user_ The user address.
     * @param iterations_ The loop interaction amount.
     */
    function withdrawUserStakes(address user_, uint8 iterations_) external;
}
