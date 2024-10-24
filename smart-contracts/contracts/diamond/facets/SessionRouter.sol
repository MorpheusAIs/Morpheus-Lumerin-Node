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

import {LibSD} from "../../libs/LibSD.sol";

import {ISessionRouter} from "../../interfaces/facets/ISessionRouter.sol";

contract SessionRouter is
    ISessionRouter,
    OwnableDiamondStorage,
    SessionStorage,
    ProviderStorage,
    BidStorage,
    StatsStorage
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
        SessionsStorage storage sessionsStorage = getSessionsStorage();

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
        if (index_ >= getSessionsStorage().pools.length) {
            revert SessionPoolIndexOutOfBounds();
        }

        getSessionsStorage().pools[index_] = pool_;
    }

    function setMaxSessionDuration(uint128 maxSessionDuration_) public onlyOwner {
        if (maxSessionDuration_ <= MIN_SESSION_DURATION) {
            revert SessionMaxDurationTooShort();
        }

        getSessionsStorage().maxSessionDuration = maxSessionDuration_;
    }

    ////////////////////////
    ///   OPEN SESSION   ///
    ////////////////////////
    function openSession(
        uint256 amount_,
        bool isDirectPaymentFromUser_,
        bytes calldata approvalEncoded_,
        bytes calldata signature_
    ) external returns (bytes32) {
        SessionsStorage storage sessionsStorage = getSessionsStorage();

        bytes32 bidId_ = _extractProviderApproval(approvalEncoded_);
        Bid storage bid = getBidsStorage().bids[bidId_];

        bytes32 sessionId_ = getSessionId(_msgSender(), bid.provider, bidId_, sessionsStorage.sessionNonce++);
        Session storage session = sessionsStorage.sessions[sessionId_];

        uint128 endsAt_ = _validateSession(bidId_, amount_, isDirectPaymentFromUser_, approvalEncoded_, signature_);

        session.user = _msgSender();
        session.stake = amount_;
        session.bidId = bidId_;
        session.openedAt = uint128(block.timestamp);
        session.endsAt = endsAt_;
        session.isActive = true;
        session.isDirectPaymentFromUser = isDirectPaymentFromUser_;

        sessionsStorage.userSessions[_msgSender()].add(sessionId_);
        sessionsStorage.providerSessions[bid.provider].add(sessionId_);
        sessionsStorage.modelSessions[bid.modelId].add(sessionId_);

        sessionsStorage.isProviderApprovalUsed[approvalEncoded_] = true;

        IERC20(getBidsStorage().token).safeTransferFrom(_msgSender(), address(this), amount_);

        emit SessionOpened(_msgSender(), sessionId_, bid.provider);

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

        Bid storage bid = getBidsStorage().bids[bidId_];
        if (!_isValidProviderReceipt(bid.provider, approvalEncoded_, signature_)) {
            revert SessionProviderSignatureMismatch();
        }
        if (getSessionsStorage().isProviderApprovalUsed[approvalEncoded_]) {
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

        if (duration_ > getSessionsStorage().maxSessionDuration) {
            duration_ = getSessionsStorage().maxSessionDuration;
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

        return (amount_ * getComputeBalance(timestamp_)) / (totalMORSupply(timestamp_) * 100);
    }

    function _extractProviderApproval(bytes calldata providerApproval_) private view returns (bytes32) {
        (bytes32 bidId_, uint256 chainId_, address user_, uint128 timestamp_) = abi.decode(
            providerApproval_,
            (bytes32, uint256, address, uint128)
        );

        if (user_ != _msgSender()) {
            revert SessionApprovedForAnotherUser();
        }
        if (chainId_ != block.chainid) {
            revert SesssionApprovedForAnotherChainId();
        }
        if (block.timestamp > timestamp_ + SIGNATURE_TTL) {
            revert SesssionApproveExpired();
        }

        return bidId_;
    }

    ///////////////////////////////////
    ///   CLOSE SESSION, WITHDRAW   ///
    ///////////////////////////////////
    function closeSession(bytes calldata receiptEncoded_, bytes calldata signature_) external {
        (bytes32 sessionId_, uint32 tpsScaled1000_, uint32 ttftMs_) = _extractProviderReceipt(receiptEncoded_);

        Session storage session = getSessionsStorage().sessions[sessionId_];
        Bid storage bid = getBidsStorage().bids[session.bidId];

        _onlyAccount(session.user);
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
            revert SesssionReceiptForAnotherChainId();
        }
        if (block.timestamp > timestamp_ + SIGNATURE_TTL) {
            revert SesssionReceiptExpired();
        }

        return (sessionId_, tpsScaled1000_, ttftMs_);
    }

    function _getProviderRewards(Session storage session, Bid storage bid, bool isIncludeWithdrawnAmount_) private view returns (uint256) {
        uint256 sessionEnd_ = session.closedAt == 0 ? session.endsAt : session.closedAt.min(session.endsAt);
        if (block.timestamp < sessionEnd_) {
            return 0;
        }

        uint256 withdrawnAmount = isIncludeWithdrawnAmount_ ?  session.providerWithdrawnAmount : 0;

        return (sessionEnd_ - session.openedAt) * bid.pricePerSecond - withdrawnAmount;
    }

    function _rewardProviderAfterClose(
        bool noDispute_,
        Session storage session, 
        Bid storage bid
    ) internal {
        uint128 startOfToday_ = startOfTheDay(uint128(block.timestamp));
        bool isClosingLate_ = uint128(block.timestamp) > session.endsAt;

        uint256 providerAmountToWithdraw_ = _getProviderRewards(session, bid, true);
        uint256 providerOnHoldAmount = 0;
        if (!noDispute_ && !isClosingLate_) {
            providerOnHoldAmount = (session.endsAt.min(session.closedAt) - startOfToday_.max(session.openedAt)) * bid.pricePerSecond;
        }
        providerAmountToWithdraw_ -= providerOnHoldAmount;

        _claimForProvider(session, providerAmountToWithdraw_);
    }

    function _rewardUserAfterClose(Session storage session, Bid storage bid) private {
        uint128 startOfToday_ = startOfTheDay(uint128(block.timestamp));
        bool isClosingLate_ = uint128(block.timestamp) > session.endsAt;

        uint256 userStakeToProvider = session.isDirectPaymentFromUser ? _getProviderRewards(session, bid, false) : 0;
        uint256 userStake = session.stake - userStakeToProvider;
        uint256 userStakeToLock_ = 0;
        if (!isClosingLate_) {
            // Session was closed on the same day, lock today's stake
            uint256 userDuration_ = session.endsAt.min(session.closedAt) - session.openedAt.max(startOfToday_);
            uint256 userInitialLock_ = userDuration_ * bid.pricePerSecond;
            userStakeToLock_ = userStake.min(stipendToStake(userInitialLock_, startOfToday_));

            getSessionsStorage().userStakesOnHold[session.user].push(OnHold(userStakeToLock_, uint128(startOfToday_ + 1 days)));
        }
        uint256 userAmountToWithdraw_ = userStake - userStakeToLock_;
        IERC20(getBidsStorage().token).safeTransfer(session.user, userAmountToWithdraw_);
    }

    function _setStats(
        bool noDispute_, 
        uint32 ttftMs_, 
        uint32 tpsScaled1000_, 
        Session storage session, 
        Bid storage bid
    ) internal {
        ProviderModelStats storage prStats = providerModelStats(bid.modelId, bid.provider);
        ModelStats storage modelStats = modelStats(bid.modelId);

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
        Session storage session = getSessionsStorage().sessions[sessionId_];
        Bid storage bid = getBidsStorage().bids[session.bidId];

        _onlyAccount(bid.provider);
        _claimForProvider(session, _getProviderRewards(session, bid, true));
    }

    /**
     * @dev Sends provider reward considering stake as the limit for the reward
     * @param session Storage session object
     * @param amount_ Amount of reward to send
     */
    function _claimForProvider(Session storage session, uint256 amount_) private {
        Bid storage bid = getBidsStorage().bids[session.bidId];
        Provider storage provider = getProvidersStorage().providers[bid.provider];

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
        getSessionsStorage().providersTotalClaimed += amount_;

        if (session.isDirectPaymentFromUser) {
            IERC20(getBidsStorage().token).safeTransfer(bid.provider, amount_);
        } else {
            IERC20(getBidsStorage().token).safeTransferFrom(getSessionsStorage().fundingAccount, bid.provider, amount_);
        }
    }

    /**
     * @notice Returns stake of user based on their stipend
     */
    function stipendToStake(uint256 stipend_, uint128 timestamp_) public view returns (uint256) {
        return (stipend_ * totalMORSupply(timestamp_) * 100) / getComputeBalance(timestamp_);
    }

    function getUserStakesOnHold(
        address user_,
        uint8 iterations_
    ) external view returns (uint256 available_, uint256 hold_) {
        OnHold[] memory onHold = getSessionsStorage().userStakesOnHold[user_];
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

    function withdrawUserStakes(uint8 iterations_) external {
        uint256 amount_ = 0;

        OnHold[] storage onHoldEntries = getSessionsStorage().userStakesOnHold[_msgSender()];
        uint8 i_ = iterations_ >= onHoldEntries.length ? uint8(onHoldEntries.length) : iterations_;
        if (i_ == 0) {
         revert SessionUserAmountToWithdrawIsZero();
        }
        i_--;

        while (i_ >= 0) {
            if (block.timestamp < onHoldEntries[i_].releaseAt) {
                if (i_ == 0) break;
                i_--;

                continue;
            }

            amount_ += onHoldEntries[i_].amount;
            onHoldEntries.pop();

            if (i_ == 0) break;
            i_--;
        }

        if (amount_ == 0) {
            revert SessionUserAmountToWithdrawIsZero();
        }

        IERC20(getBidsStorage().token).safeTransfer(_msgSender(), amount_);

        emit UserWithdrawn(_msgSender(), amount_);
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
        SessionsStorage storage sessionsStorage = getSessionsStorage();
        Pool storage pool = sessionsStorage.pools[COMPUTE_POOL_INDEX];

        uint256 periodReward = LinearDistributionIntervalDecrease.getPeriodReward(
            pool.initialReward,
            pool.rewardDecrease,
            pool.payoutStart,
            pool.decreaseInterval,
            pool.payoutStart,
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

        SessionsStorage storage sessionsStorage = getSessionsStorage();
        Pool[] storage pools = sessionsStorage.pools;

        for (uint256 i = 0; i < pools.length; i++) {
            if (i == COMPUTE_POOL_INDEX) continue;

            totalSupply_ += LinearDistributionIntervalDecrease.getPeriodReward(
                pools[i].initialReward,
                pools[i].rewardDecrease,
                pools[i].payoutStart,
                pools[i].decreaseInterval,
                pools[i].payoutStart,
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
    ) private pure returns (bool) {
        bytes32 receiptHash_ = ECDSA.toEthSignedMessageHash(keccak256(receipt_));

        return ECDSA.recover(receiptHash_, signature_) == provider_;
    }
}
