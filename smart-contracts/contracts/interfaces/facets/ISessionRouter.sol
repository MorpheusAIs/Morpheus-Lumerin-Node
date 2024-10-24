// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ISessionStorage} from "../storage/ISessionStorage.sol";
import {IBidStorage} from "../storage/IBidStorage.sol";
import {IStatsStorage} from "../storage/IStatsStorage.sol";

interface ISessionRouter is ISessionStorage {
    event SessionOpened(address indexed user, bytes32 indexed sessionId, address indexed providerId);
    event SessionClosed(address indexed user, bytes32 indexed sessionId, address indexed providerId);
    event UserWithdrawn(address indexed user, uint256 amount_);
    error SessionProviderSignatureMismatch();
    error SesssionApproveExpired();
    error SesssionApprovedForAnotherChainId();
    error SessionDuplicateApproval();
    error SessionApprovedForAnotherUser();
    error SesssionReceiptForAnotherChainId();
    error SesssionReceiptExpired();
    error SessionTooShort();
    error SessionAlreadyClosed();
    error SessionNotEndedOrNotExist();
    error SessionProviderNothingToClaimInThisPeriod();
    error SessionBidNotFound();
    error SessionPoolIndexOutOfBounds();
    error SessionUserAmountToWithdrawIsZero();
    error SessionMaxDurationTooShort();
    error SessionStakeTooLow();

    /**
     * The function to initialize the facet.
     * @param fundingAccount_ The funding address (treaasury)
     * @param maxSessionDuration_ The max session duration
     * @param pools_ The pools data
     */
    function __SessionRouter_init(address fundingAccount_, uint128 maxSessionDuration_, Pool[] calldata pools_) external;

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
     * @param amount_ The stake amount.
     * @param isDirectPaymentFromUser_ If active, provider rewarded from the user stake.
     * @param approvalEncoded_ Provider approval.
     * @param signature_ Provider signature.
     */
    function openSession(
        uint256 amount_,
        bool isDirectPaymentFromUser_,
        bytes calldata approvalEncoded_,
        bytes calldata signature_
    ) external returns (bytes32);

    /**
     * The function to get session ID/
     * @param user_ The user address.
     * @param provider_ The provider address.
     * @param bidId_ The bid ID.
     * @param sessionNonce_ The session nounce.
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
     */
    function getSessionEnd(uint256 amount_, uint256 pricePerSecond_, uint128 openedAt_) external view returns (uint128);

    /**
     * Returns stipend of user based on their stake
     * (User session stake amount / MOR Supply without Compute) * (MOR Compute Supply / 100)
     * @param amount_ The amount of tokens.
     * @param timestamp_ The timestamp when the TX executes.
     */
    function stakeToStipend(uint256 amount_, uint128 timestamp_) external view returns (uint256);

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
     * Returns stake of user based on their stipend.
     * @param stipend_ The stake amount.
     * @param timestamp_ The timestamp when the TX executed.
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
     * The function to withdraw user stakes.
     * @param iterations_ The loop interaction amount.
     */
    function withdrawUserStakes(uint8 iterations_) external;

    /**
     * Returns today's budget in MOR. 1%.
     * @param timestamp_ The timestamp when the TX executed.
     */
    function getTodaysBudget(uint128 timestamp_) external view returns (uint256);

    /**
     * Returns today's compute balance in MOR without claimed amount.
     * @param timestamp_ The timestamp when the TX executed.
     */
    function getComputeBalance(uint128 timestamp_) external view returns (uint256);

    /**
     * Total amount of MOR tokens that were distributed across all pools
     * without compute pool rewards and with compute claimed rewards.
     * @param timestamp_ The timestamp when the TX executed.
     */
    function totalMORSupply(uint128 timestamp_) external view returns (uint256);

    /**
     * The function to return the timestamp on the start of day
     * @param timestamp_ The timestamp when the TX executed.
     */
    function startOfTheDay(uint128 timestamp_) external pure returns (uint128);
}
