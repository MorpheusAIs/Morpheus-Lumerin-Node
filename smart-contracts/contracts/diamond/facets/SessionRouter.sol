// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {SafeERC20, IERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

import {LinearDistributionIntervalDecrease} from "morpheus-smart-contracts/contracts/libs/LinearDistributionIntervalDecrease.sol";

import {OwnableDiamondStorage} from "../presets/OwnableDiamondStorage.sol";

import {BidStorage, EnumerableSet} from "../storages/BidStorage.sol";
import {StatsStorage} from "../storages/StatsStorage.sol";
import {SessionStorage} from "../storages/SessionStorage.sol";
import {ProviderStorage} from "../storages/ProviderStorage.sol";
import {DelegationStorage} from "../storages/DelegationStorage.sol";

import {LibSD} from "../../libs/LibSD.sol";

import {ISessionRouter} from "../../interfaces/facets/ISessionRouter.sol";

import "hardhat/console.sol";

contract SessionRouter is
    ISessionRouter,
    OwnableDiamondStorage,
    SessionStorage,
    ProviderStorage,
    BidStorage,
    StatsStorage,
    DelegationStorage
{
    using Math for *;
    using LibSD for LibSD.SD;
    using SafeERC20 for IERC20;
    using EnumerableSet for EnumerableSet.Bytes32Set;

    function __SessionRouter_init(
        address fundingAccount_,
        uint128 maxSessionDuration_,
        Pool[] calldata pools_
    ) external initializer(SESSIONS_STORAGE_SLOT) {
        SessionsStorage storage sessionsStorage = _getSessionsStorage();

        setMaxSessionDuration(maxSessionDuration_);
        sessionsStorage.fundingAccount = fundingAccount_;
        for (uint256 i = 0; i < pools_.length; i++) {
            sessionsStorage.pools.push(pools_[i]);
        }
    }

    ////////////////////////////
    ///   CONTRACT CONFIGS   ///
    ////////////////////////////
    /**
     * @notice Sets distibution pool configuration
     * @dev parameters should be the same as in Ethereum L1 Distribution contract
     * @dev at address 0x47176B2Af9885dC6C4575d4eFd63895f7Aaa4790
     * @dev call 'Distribution.pools(3)' where '3' is a poolId
     */
    function setPoolConfig(uint256 index_, Pool calldata pool_) external onlyOwner {
        SessionsStorage storage sessionsStorage = _getSessionsStorage();

        if (index_ >= sessionsStorage.pools.length) {
            revert SessionPoolIndexOutOfBounds();
        }

        sessionsStorage.pools[index_] = pool_;
    }

    function setMaxSessionDuration(uint128 maxSessionDuration_) public onlyOwner {
        if (maxSessionDuration_ < MIN_SESSION_DURATION) {
            revert SessionMaxDurationTooShort();
        }

        _getSessionsStorage().maxSessionDuration = maxSessionDuration_;
    }

    ////////////////////////
    ///   OPEN SESSION   ///
    ////////////////////////
    function openSession(
        address user_,
        uint256 amount_,
        bool isDirectPaymentFromUser_,
        bytes calldata approvalEncoded_,
        bytes calldata signature_
    ) external returns (bytes32) {
        _validateDelegatee(_msgSender(), user_, DELEGATION_RULES_SESSION);

        SessionsStorage storage sessionsStorage = _getSessionsStorage();

        bytes32 bidId_ = _extractProviderApproval(approvalEncoded_);

        bytes32 sessionId_ = getSessionId(
            user_,
            _getBidsStorage().bids[bidId_].provider,
            bidId_,
            sessionsStorage.sessionNonce++
        );
        Session storage session = sessionsStorage.sessions[sessionId_];

        IERC20(_getBidsStorage().token).safeTransferFrom(user_, address(this), amount_);

        session.user = user_;
        session.stake = amount_;
        session.bidId = bidId_;
        session.openedAt = uint128(block.timestamp);
        session.endsAt = _validateSession(bidId_, amount_, isDirectPaymentFromUser_, approvalEncoded_, signature_);
        session.isActive = true;
        session.isDirectPaymentFromUser = isDirectPaymentFromUser_;

        sessionsStorage.userSessions[user_].add(sessionId_);
        sessionsStorage.providerSessions[_getBidsStorage().bids[bidId_].provider].add(sessionId_);
        sessionsStorage.modelSessions[_getBidsStorage().bids[bidId_].modelId].add(sessionId_);

        sessionsStorage.isProviderApprovalUsed[approvalEncoded_] = true;

        emit SessionOpened(user_, sessionId_, _getBidsStorage().bids[bidId_].provider);

        return sessionId_;
    }

    function _validateSession(
        bytes32 bidId_,
        uint256 amount_,
        bool isDirectPaymentFromUser_,
        bytes calldata approvalEncoded_,
        bytes calldata signature_
    ) private view returns (uint128) {
        if (!isBidActive(bidId_)) {
            revert SessionBidNotFound();
        }

        Bid storage bid = _getBidsStorage().bids[bidId_];
        if (!_isValidProviderReceipt(bid.provider, approvalEncoded_, signature_)) {
            revert SessionProviderSignatureMismatch();
        }
        if (_getSessionsStorage().isProviderApprovalUsed[approvalEncoded_]) {
            revert SessionDuplicateApproval();
        }

        uint128 endsAt_ = getSessionEnd(amount_, bid.pricePerSecond, uint128(block.timestamp));
        uint128 duration_ = endsAt_ - uint128(block.timestamp);

        if (duration_ < MIN_SESSION_DURATION) {
            revert SessionTooShort();
        }

        // This situation cannot be achieved in theory, but just in case, I'll leave it at that
        if (isDirectPaymentFromUser_ && (duration_ * bid.pricePerSecond) > amount_) {
            revert SessionStakeTooLow();
        }

        return endsAt_;
    }

    function getSessionId(
        address user_,
        address provider_,
        bytes32 bidId_,
        uint256 sessionNonce_
    ) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(user_, provider_, bidId_, sessionNonce_));
    }

    function getSessionEnd(uint256 amount_, uint256 pricePerSecond_, uint128 openedAt_) public view returns (uint128) {
        uint128 duration_ = uint128(stakeToStipend(amount_, openedAt_) / pricePerSecond_);

        if (duration_ > _getSessionsStorage().maxSessionDuration) {
            duration_ = _getSessionsStorage().maxSessionDuration;
        }

        return openedAt_ + duration_;
    }

    /**
     * @dev Returns stipend of user based on their stake
     * (User session stake amount / MOR Supply without Compute) * (MOR Compute Supply / 100)
     * (User share) * (Rewards for all computes)
     */
    function stakeToStipend(uint256 amount_, uint128 timestamp_) public view returns (uint256) {
        uint256 totalMorSupply_ = totalMORSupply(timestamp_);
        if (totalMorSupply_ == 0) {
            return 0;
        }

        return (amount_ * getComputeBalance(timestamp_)) / (totalMorSupply_ * 100);
    }

    function _extractProviderApproval(bytes calldata providerApproval_) private view returns (bytes32) {
        (bytes32 bidId_, uint256 chainId_, , uint128 timestamp_) = abi.decode(
            providerApproval_,
            (bytes32, uint256, address, uint128)
        );

        if (chainId_ != block.chainid) {
            revert SessionApprovedForAnotherChainId();
        }
        if (block.timestamp > timestamp_ + SIGNATURE_TTL) {
            revert SessionApproveExpired();
        }

        return bidId_;
    }

    ///////////////////////////////////
    ///   CLOSE SESSION, WITHDRAW   ///
    ///////////////////////////////////
    function closeSession(bytes calldata receiptEncoded_, bytes calldata signature_) external {
        (bytes32 sessionId_, uint32 tpsScaled1000_, uint32 ttftMs_) = _extractProviderReceipt(receiptEncoded_);

        Session storage session = _getSessionsStorage().sessions[sessionId_];
        Bid storage bid = _getBidsStorage().bids[session.bidId];

        _validateDelegatee(_msgSender(), session.user, DELEGATION_RULES_SESSION);

        if (session.closedAt != 0) {
            revert SessionAlreadyClosed();
        }

        session.isActive = false;
        session.closeoutReceipt = receiptEncoded_; // TODO: Remove that field in favor of tps and ttftMs
        session.closedAt = uint128(block.timestamp);

        bool noDispute_ = _isValidProviderReceipt(bid.provider, receiptEncoded_, signature_);

        _rewardUserAfterClose(session, bid);
        _rewardProviderAfterClose(noDispute_, session, bid);
        _setStats(noDispute_, ttftMs_, tpsScaled1000_, session, bid);

        emit SessionClosed(session.user, sessionId_, bid.provider);
    }

    function _extractProviderReceipt(bytes calldata receiptEncoded_) private view returns (bytes32, uint32, uint32) {
        (bytes32 sessionId_, uint256 chainId_, uint128 timestamp_, uint32 tpsScaled1000_, uint32 ttftMs_) = abi.decode(
            receiptEncoded_,
            (bytes32, uint256, uint128, uint32, uint32)
        );

        if (chainId_ != block.chainid) {
            revert SessionReceiptForAnotherChainId();
        }
        if (block.timestamp > timestamp_ + SIGNATURE_TTL) {
            revert SessionReceiptExpired();
        }

        return (sessionId_, tpsScaled1000_, ttftMs_);
    }

    function _getProviderRewards(
        Session storage session,
        Bid storage bid,
        bool isIncludeWithdrawnAmount_
    ) private view returns (uint256) {
        uint256 sessionEnd_ = session.closedAt == 0 ? session.endsAt : session.closedAt.min(session.endsAt);
        if (block.timestamp < sessionEnd_) {
            return 0;
        }

        uint256 withdrawnAmount = isIncludeWithdrawnAmount_ ? session.providerWithdrawnAmount : 0;

        return (sessionEnd_ - session.openedAt) * bid.pricePerSecond - withdrawnAmount;
    }

    function _getProviderOnHoldAmount(Session storage session, Bid storage bid) private view returns (uint256) {
        uint128 startOfClosedAt = startOfTheDay(session.closedAt);
        if (block.timestamp >= startOfClosedAt + 1 days) {
            return 0;
        }

        // `closedAt` - latest timestamp, cause `endsAt` bigger then `closedAt`
        // Lock the provider's tokens for the current day.
        // Withdrawal is allowed after a day after `startOfTheDay(session.closedAt)`.
        return (session.closedAt - startOfClosedAt.max(session.openedAt)) * bid.pricePerSecond;
    }

    function _rewardProviderAfterClose(bool noDispute_, Session storage session, Bid storage bid) internal {
        bool isClosingLate_ = session.closedAt >= session.endsAt;

        uint256 providerAmountToWithdraw_ = _getProviderRewards(session, bid, true);
        uint256 providerOnHoldAmount = 0;
        // Enter when the user has a dispute AND closing early
        if (!noDispute_ && !isClosingLate_) {
            providerOnHoldAmount = _getProviderOnHoldAmount(session, bid);
        }
        providerAmountToWithdraw_ -= providerOnHoldAmount;

        _claimForProvider(session, providerAmountToWithdraw_);
    }

    function _rewardUserAfterClose(Session storage session, Bid storage bid) private {
        uint128 startOfClosedAt_ = startOfTheDay(session.closedAt);
        bool isClosingLate_ = session.closedAt >= session.endsAt;

        uint256 userStakeToProvider = session.isDirectPaymentFromUser ? _getProviderRewards(session, bid, false) : 0;
        uint256 userStake = session.stake - userStakeToProvider;
        uint256 userStakeToLock_ = 0;
        if (!isClosingLate_) {
            uint256 userDuration_ = session.endsAt.min(session.closedAt) - session.openedAt.max(startOfClosedAt_);
            uint256 userInitialLock_ = userDuration_ * bid.pricePerSecond;
            userStakeToLock_ = userStake.min(stipendToStake(userInitialLock_, startOfClosedAt_));

            _getSessionsStorage().userStakesOnHold[session.user].push(
                OnHold(userStakeToLock_, uint128(startOfClosedAt_ + 1 days))
            );
        }
        uint256 userAmountToWithdraw_ = userStake - userStakeToLock_;
        IERC20(_getBidsStorage().token).safeTransfer(session.user, userAmountToWithdraw_);
    }

    function _setStats(
        bool noDispute_,
        uint32 ttftMs_,
        uint32 tpsScaled1000_,
        Session storage session,
        Bid storage bid
    ) internal {
        ProviderModelStats storage prStats = _providerModelStats(bid.modelId, bid.provider);
        ModelStats storage modelStats = _modelStats(bid.modelId);

        prStats.totalCount++;

        if (noDispute_) {
            if (prStats.successCount > 0) {
                // Stats for this provider-model pair already contribute to average model stats
                modelStats.tpsScaled1000.remove(int32(prStats.tpsScaled1000.mean), int32(modelStats.count - 1));
                modelStats.ttftMs.remove(int32(prStats.ttftMs.mean), int32(modelStats.count - 1));
            } else {
                // Stats for this provider-model pair do not contribute
                modelStats.count++;
            }

            // Update provider model stats
            prStats.successCount++;
            prStats.totalDuration += uint32(session.closedAt - session.openedAt);
            prStats.tpsScaled1000.add(int32(tpsScaled1000_), int32(prStats.successCount));
            prStats.ttftMs.add(int32(ttftMs_), int32(prStats.successCount));

            // Update model stats
            modelStats.totalDuration.add(int32(prStats.totalDuration), int32(modelStats.count));
            modelStats.tpsScaled1000.add(int32(prStats.tpsScaled1000.mean), int32(modelStats.count));
            modelStats.ttftMs.add(int32(prStats.ttftMs.mean), int32(modelStats.count));
        } else {
            session.closeoutType = 1;
        }
    }

    /**
     * @dev Allows providers to receive their funds after the end or closure of the session
     */
    function claimForProvider(bytes32 sessionId_) external {
        Session storage session = _getSessionsStorage().sessions[sessionId_];
        Bid storage bid = _getBidsStorage().bids[session.bidId];

        _validateDelegatee(_msgSender(), bid.provider, DELEGATION_RULES_SESSION);

        uint256 amount_ = _getProviderRewards(session, bid, true) - _getProviderOnHoldAmount(session, bid);

        _claimForProvider(session, amount_);
    }

    /**
     * @dev Sends provider reward considering stake as the limit for the reward
     * @param session Storage session object
     * @param amount_ Amount of reward to send
     */
    function _claimForProvider(Session storage session, uint256 amount_) private {
        Bid storage bid = _getBidsStorage().bids[session.bidId];
        Provider storage provider = _getProvidersStorage().providers[bid.provider];

        if (block.timestamp > provider.limitPeriodEnd) {
            provider.limitPeriodEnd = uint128(block.timestamp) + PROVIDER_REWARD_LIMITER_PERIOD;
            provider.limitPeriodEarned = 0;
        }

        uint256 providerClaimLimit_ = provider.stake - provider.limitPeriodEarned;

        amount_ = amount_.min(providerClaimLimit_);
        if (amount_ == 0) {
            return;
        }

        session.providerWithdrawnAmount += amount_;
        provider.limitPeriodEarned += amount_;
        _getSessionsStorage().providersTotalClaimed += amount_;

        if (session.isDirectPaymentFromUser) {
            IERC20(_getBidsStorage().token).safeTransfer(bid.provider, amount_);
        } else {
            IERC20(_getBidsStorage().token).safeTransferFrom(
                _getSessionsStorage().fundingAccount,
                bid.provider,
                amount_
            );
        }
    }

    /**
     * @notice Returns stake of user based on their stipend
     */
    function stipendToStake(uint256 stipend_, uint128 timestamp_) public view returns (uint256) {
        uint256 computeBalance_ = getComputeBalance(timestamp_);
        if (computeBalance_ == 0) {
            return 0;
        }

        return (stipend_ * totalMORSupply(timestamp_) * 100) / computeBalance_;
    }

    function getUserStakesOnHold(
        address user_,
        uint8 iterations_
    ) external view returns (uint256 available_, uint256 hold_) {
        OnHold[] memory onHold = _getSessionsStorage().userStakesOnHold[user_];
        iterations_ = iterations_ > onHold.length ? uint8(onHold.length) : iterations_;

        for (uint256 i = 0; i < onHold.length; i++) {
            uint256 amount = onHold[i].amount;

            if (block.timestamp < onHold[i].releaseAt) {
                hold_ += amount;
            } else {
                available_ += amount;
            }
        }
    }

    function withdrawUserStakes(address user_, uint8 iterations_) external {
        _validateDelegatee(_msgSender(), user_, DELEGATION_RULES_SESSION);

        OnHold[] storage onHoldEntries = _getSessionsStorage().userStakesOnHold[user_];
        uint8 count_ = iterations_ >= onHoldEntries.length ? uint8(onHoldEntries.length) : iterations_;
        uint256 length_ = onHoldEntries.length;
        uint256 amount_ = 0;

        if (length_ == 0 || count_ == 0) {
            revert SessionUserAmountToWithdrawIsZero();
        }

        uint8 removedCount_;
        for (uint256 i = length_; i > 0 && removedCount_ < count_; i--) {
            if (block.timestamp < onHoldEntries[i - 1].releaseAt) {
                continue;
            }

            amount_ += onHoldEntries[i - 1].amount;

            onHoldEntries[i - 1] = onHoldEntries[length_ - 1];
            onHoldEntries.pop();
            length_--;
            removedCount_++;
        }

        if (amount_ == 0) {
            revert SessionUserAmountToWithdrawIsZero();
        }

        IERC20(_getBidsStorage().token).safeTransfer(user_, amount_);

        emit UserWithdrawn(user_, amount_);
    }

    ////////////////////////
    ///   GLOBAL PUBLIC  ///
    ////////////////////////

    /**
     * @dev Returns today's budget in MOR. 1%
     */
    function getTodaysBudget(uint128 timestamp_) external view returns (uint256) {
        return getComputeBalance(timestamp_) / 100;
    }

    /**
     * @dev Returns today's compute balance in MOR without claimed amount
     */
    function getComputeBalance(uint128 timestamp_) public view returns (uint256) {
        SessionsStorage storage sessionsStorage = _getSessionsStorage();
        Pool memory pool_ = sessionsStorage.pools[COMPUTE_POOL_INDEX];

        uint256 periodReward = LinearDistributionIntervalDecrease.getPeriodReward(
            pool_.initialReward,
            pool_.rewardDecrease,
            pool_.payoutStart,
            pool_.decreaseInterval,
            pool_.payoutStart,
            uint128(startOfTheDay(timestamp_))
        );

        return periodReward - sessionsStorage.providersTotalClaimed;
    }

    /**
     * @dev Total amount of MOR tokens that were distributed across all pools
     * without compute pool rewards and with compute claimed rewards
     */
    function totalMORSupply(uint128 timestamp_) public view returns (uint256) {
        uint256 startOfTheDay_ = startOfTheDay(timestamp_);
        uint256 totalSupply_ = 0;

        SessionsStorage storage sessionsStorage = _getSessionsStorage();
        uint256 poolsLength_ = sessionsStorage.pools.length;

        for (uint256 i = 0; i < poolsLength_; i++) {
            if (i == COMPUTE_POOL_INDEX) continue;

            Pool memory pool_ = sessionsStorage.pools[i];

            totalSupply_ += LinearDistributionIntervalDecrease.getPeriodReward(
                pool_.initialReward,
                pool_.rewardDecrease,
                pool_.payoutStart,
                pool_.decreaseInterval,
                pool_.payoutStart,
                uint128(startOfTheDay_)
            );
        }

        return totalSupply_ + sessionsStorage.providersTotalClaimed;
    }

    function startOfTheDay(uint128 timestamp_) public pure returns (uint128) {
        return timestamp_ - (timestamp_ % 1 days);
    }

    //////////////////////////
    ///   GLOBAL PRIVATE   ///
    //////////////////////////

    function _isValidProviderReceipt(
        address provider_,
        bytes calldata receipt_,
        bytes calldata signature_
    ) private view returns (bool) {
        bytes32 receiptHash_ = ECDSA.toEthSignedMessageHash(keccak256(receipt_));
        address user_ = ECDSA.recover(receiptHash_, signature_);

        if (user_ == provider_ || isRightsDelegated(user_, provider_, DELEGATION_RULES_PROVIDER)) {
            return true;
        }

        if (provider_.code.length > 0) {
            (bool success, bytes memory result) = provider_.staticcall(abi.encodeWithSignature("owner()"));
            if (success && result.length == 32) {
                address owner_ = abi.decode(result, (address));
                if (user_ == owner_) {
                    return true;
                }
            }
        }

        return false;
    }
}
