// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {ISessionStorage} from "../storage/ISessionStorage.sol";
import {IBidStorage} from "../storage/IBidStorage.sol";
import {IStatsStorage} from "../storage/IStatsStorage.sol";

interface ISessionRouter is ISessionStorage {
    event SessionOpened(address indexed user, bytes32 indexed sessionId, address indexed providerId);
    event SessionClosed(address indexed user, bytes32 indexed sessionId, address indexed providerId);

    error NotEnoughWithdrawableBalance(); // means that there is not enough funds at all or some funds are still locked
    error WithdrawableBalanceLimitByStakeReached(); // means that user can't withdraw more funds because of the limit which equals to the stake
    error ProviderSignatureMismatch();
    error SignatureExpired();
    error WrongChaidId();
    error DuplicateApproval();
    error ApprovedForAnotherUser(); // means that approval generated for another user address, protection from front-running

    error SessionTooShort();
    error SessionNotFound();
    error SessionAlreadyClosed();
    error SessionNotClosed();

    error BidNotFound();
    error CannotDecodeAbi();

    error AmountToWithdrawIsZero();
    error NotOwnerOrProvider();
    error NotOwnerOrUser();
    error PoolIndexOutOfBounds();

    function __SessionRouter_init(address fundingAccount_, Pool[] memory pools_) external;

    function openSession(
        uint256 amount_,
        bytes memory providerApproval_,
        bytes memory signature_
    ) external returns (bytes32);

    function closeSession(bytes memory receiptEncoded_, bytes memory signature_) external;

    function claimProviderBalance(bytes32 sessionId_, uint256 amountToWithdraw_) external;

    function deleteHistory(bytes32 sessionId_) external;

    function withdrawUserStake(uint256 amountToWithdraw_, uint8 iterations_) external;

    function withdrawableUserStake(
        address user_,
        uint8 iterations_
    ) external view returns (uint256 avail_, uint256 hold_);

    function getSessionId(
        address user_,
        address provider_,
        uint256 stake_,
        uint256 sessionNonce_
    ) external pure returns (bytes32);

    function stakeToStipend(uint256 sessionStake_, uint256 timestamp_) external view returns (uint256);

    function stipendToStake(uint256 stipend_, uint256 timestamp_) external view returns (uint256);

    function whenSessionEnds(
        uint256 sessionStake_,
        uint256 pricePerSecond_,
        uint256 openedAt_
    ) external view returns (uint256);

    function getTodaysBudget(uint256 timestamp_) external view returns (uint256);

    function getComputeBalance(uint256 timestamp_) external view returns (uint256);

    function totalMORSupply(uint256 timestamp_) external view returns (uint256);

    function startOfTheDay(uint256 timestamp_) external pure returns (uint256);

    function getProviderClaimableBalance(bytes32 sessionId_) external view returns (uint256);

    function setPoolConfig(uint256 index, Pool calldata pool) external;

    function SIGNATURE_TTL() external view returns (uint32);

    function getActiveBidsRatingByModelAgent(
        bytes32 modelAgentId_,
        uint256 offset_,
        uint8 limit_
    ) external view returns (bytes32[] memory, IBidStorage.Bid[] memory, IStatsStorage.ProviderModelStats[] memory);
}
