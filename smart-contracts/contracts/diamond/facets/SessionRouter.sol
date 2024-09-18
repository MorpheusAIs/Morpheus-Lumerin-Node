// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {SafeERC20, IERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

import {OwnableDiamondStorage} from "../presets/OwnableDiamondStorage.sol";

import {BidStorage, EnumerableSet} from "../storages/BidStorage.sol";
import {StatsStorage} from "../storages/StatsStorage.sol";
import {SessionStorage} from "../storages/SessionStorage.sol";
import {ProviderStorage} from "../storages/ProviderStorage.sol";

import {LibSD} from "../../libs/LibSD.sol";

import {ISessionRouter} from "../../interfaces/facets/ISessionRouter.sol";

import {LinearDistributionIntervalDecrease} from "morpheus-smart-contracts/contracts/libs/LinearDistributionIntervalDecrease.sol";

contract SessionRouter is
    ISessionRouter,
    OwnableDiamondStorage,
    SessionStorage,
    ProviderStorage,
    BidStorage,
    StatsStorage
{
    using Math for uint256;
    using LibSD for LibSD.SD;
    using SafeERC20 for IERC20;
    using EnumerableSet for EnumerableSet.Bytes32Set;

    uint32 public constant MIN_SESSION_DURATION = 5 minutes;
    uint32 public constant MAX_SESSION_DURATION = 1 days;
    uint32 public constant SIGNATURE_TTL = 10 minutes;
    uint256 public constant COMPUTE_POOL_INDEX = 3;

    function __SessionRouter_init(
        address fundingAccount_,
        Pool[] calldata pools_
    ) external initializer(SESSION_STORAGE_SLOT) {
        SNStorage storage s = _getSessionStorage();

        s.fundingAccount = fundingAccount_;

        for (uint256 i = 0; i < pools_.length; i++) {
            s.pools.push(pools_[i]);
        }
    }

    function openSession(
        uint256 amount_,
        bytes calldata providerApproval_,
        bytes calldata signature_
    ) external returns (bytes32) {
        bytes32 bidId_ = _extractProviderApproval(providerApproval_);

        Bid memory bid_ = getBid(bidId_);
        if (bid_.deletedAt != 0 || bid_.createdAt == 0) {
            // wtf?
            revert BidNotFound();
        }
        if (!_isValidReceipt(bid_.provider, providerApproval_, signature_)) {
            revert ProviderSignatureMismatch();
        }
        if (isApproved(providerApproval_)) {
            revert DuplicateApproval();
        }
        approve(providerApproval_);

        uint256 endsAt_ = whenSessionEnds(amount_, bid_.pricePerSecond, block.timestamp);
        if (endsAt_ - block.timestamp < MIN_SESSION_DURATION) {
            revert SessionTooShort();
        }

        bytes32 sessionId_ = getSessionId(_msgSender(), bid_.provider, amount_, incrementSessionNonce());
        setSession(
            sessionId_,
            Session({
                id: sessionId_,
                user: _msgSender(),
                provider: bid_.provider,
                modelId: bid_.modelId,
                bidID: bidId_,
                stake: amount_,
                pricePerSecond: bid_.pricePerSecond,
                closeoutReceipt: "",
                closeoutType: 0,
                providerWithdrawnAmount: 0,
                openedAt: uint128(block.timestamp),
                endsAt: endsAt_,
                closedAt: 0
            })
        );

        addUserSessionId(_msgSender(), sessionId_);
        addProviderSessionId(bid_.provider, sessionId_);
        addModelSessionId(bid_.modelId, sessionId_);

        setUserSessionActive(_msgSender(), sessionId_, true);
        setProviderSessionActive(bid_.provider, sessionId_, true);

        // try to use locked stake first, but limit iterations to 20
        // if user has more than 20 onHold entries, they will have to use withdrawUserStake separately
        amount_ -= _removeUserStake(amount_, 10);

        getToken().safeTransferFrom(_msgSender(), address(this), amount_);

        emit SessionOpened(_msgSender(), sessionId_, bid_.provider);

        return sessionId_;
    }

    function closeSession(bytes calldata receiptEncoded_, bytes calldata signature_) external {
        (bytes32 sessionId_, uint32 tpsScaled1000_, uint32 ttftMs_) = _extractReceipt(receiptEncoded_);

        Session storage session = _getSession(sessionId_);
        if (session.openedAt == 0) {
            revert SessionNotFound();
        }
        if (!_ownerOrUser(session.user)) {
            revert NotOwnerOrUser();
        }

        if (session.closedAt != 0) {
            revert SessionAlreadyClosed();
        }

        // update indexes
        setUserSessionActive(session.user, sessionId_, false);
        setProviderSessionActive(session.provider, sessionId_, false);

        // update session record
        session.closeoutReceipt = receiptEncoded_; //TODO: remove that field in favor of tps and ttftMs
        session.closedAt = uint128(block.timestamp);

        // calculate provider withdraw
        uint256 providerWithdraw_;
        uint256 startOfToday_ = startOfTheDay(block.timestamp);
        bool isClosingLate_ = startOfToday_ > startOfTheDay(session.endsAt);
        bool noDispute_ = _isValidReceipt(session.provider, receiptEncoded_, signature_);

        if (noDispute_ || isClosingLate_) {
            // session was closed without dispute or next day after it expected to end
            uint256 duration_ = session.endsAt.min(block.timestamp) - session.openedAt;
            uint256 cost_ = duration_ * session.pricePerSecond;
            providerWithdraw_ = cost_ - session.providerWithdrawnAmount;
        } else {
            // session was closed on the same day or earlier with dispute
            // withdraw all funds except for today's session cost
            uint256 durationTillToday_ = startOfToday_ - session.openedAt.min(startOfToday_);
            uint256 costTillToday_ = durationTillToday_ * session.pricePerSecond;
            providerWithdraw_ = costTillToday_ - session.providerWithdrawnAmount;
        }

        // updating provider stats
        ProviderModelStats storage prStats = _getProviderModelStats(session.modelId, session.provider);
        ModelStats storage modelStats = _getModelStats(session.modelId);

        prStats.totalCount++;

        if (noDispute_) {
            if (prStats.successCount > 0) {
                // stats for this provider-model pair already contribute to average model stats
                modelStats.tpsScaled1000.remove(int32(prStats.tpsScaled1000.mean), int32(modelStats.count - 1));
                modelStats.ttftMs.remove(int32(prStats.ttftMs.mean), int32(modelStats.count - 1));
            } else {
                // stats for this provider-model pair do not contribute
                modelStats.count++;
            }

            // update provider-model stats
            prStats.successCount++;
            prStats.totalDuration += uint32(session.closedAt - session.openedAt);
            prStats.tpsScaled1000.add(int32(tpsScaled1000_), int32(prStats.successCount));
            prStats.ttftMs.add(int32(ttftMs_), int32(prStats.successCount));

            // update model stats
            modelStats.totalDuration.add(int32(prStats.totalDuration), int32(modelStats.count));
            modelStats.tpsScaled1000.add(int32(prStats.tpsScaled1000.mean), int32(modelStats.count));
            modelStats.ttftMs.add(int32(prStats.ttftMs.mean), int32(modelStats.count));
        } else {
            session.closeoutType = 1;
        }

        // we have to lock today's stake so the user won't get the reward twice
        uint256 userStakeToLock_ = 0;
        if (!isClosingLate_) {
            // session was closed on the same day
            // lock today's stake
            uint256 todaysDuration_ = session.endsAt.min(block.timestamp) - session.openedAt.max(startOfToday_);
            uint256 todaysCost_ = todaysDuration_ * session.pricePerSecond;
            userStakeToLock_ = session.stake.min(stipendToStake(todaysCost_, startOfToday_));
            addOnHold(session.user, OnHold(userStakeToLock_, uint128(startOfToday_ + 1 days)));
        }
        uint256 userWithdraw_ = session.stake - userStakeToLock_;

        emit SessionClosed(session.user, sessionId_, session.provider);

        // withdraw provider
        _rewardProvider(session, providerWithdraw_, false);

        // withdraw user
        getToken().safeTransfer(session.user, userWithdraw_);
    }

    /// @notice allows provider to claim their funds
    function claimProviderBalance(bytes32 sessionId_, uint256 amountToWithdraw_) external {
        Session storage session = _getSession(sessionId_);
        if (!_ownerOrProvider(session.provider)) {
            revert NotOwnerOrProvider();
        }
        if (session.openedAt == 0) {
            revert SessionNotFound();
        }

        uint256 withdrawableAmount = _getProviderClaimableBalance(session);
        if (amountToWithdraw_ > withdrawableAmount) {
            revert NotEnoughWithdrawableBalance();
        }

        _rewardProvider(session, amountToWithdraw_, true);
    }

    /// @notice deletes session from the history
    function deleteHistory(bytes32 sessionId_) external {
        Session storage session = _getSession(sessionId_);
        if (!_ownerOrUser(session.user)) {
            revert NotOwnerOrUser();
        }
        if (session.closedAt == 0) {
            revert SessionNotClosed();
        }

        session.user = address(0);
    }

    /// @notice withdraws user stake
    /// @param amountToWithdraw_ amount of funds to withdraw, maxUint256 means all available
    /// @param iterations_ number of entries to process
    function withdrawUserStake(uint256 amountToWithdraw_, uint8 iterations_) external {
        // withdraw all available funds if amountToWithdraw is 0
        if (amountToWithdraw_ == 0) {
            revert AmountToWithdrawIsZero();
        }

        uint256 removed_ = _removeUserStake(amountToWithdraw_, iterations_);
        if (removed_ < amountToWithdraw_) {
            revert NotEnoughWithdrawableBalance();
        }

        getToken().safeTransfer(_msgSender(), amountToWithdraw_);
    }

    /// @dev removes user stake amount from onHold entries
    function _removeUserStake(uint256 amountToRemove_, uint8 iterations_) private returns (uint256) {
        uint256 balance_ = 0;

        OnHold[] storage onHoldEntries = getOnHold(_msgSender());
        iterations_ = iterations_ > onHoldEntries.length ? uint8(onHoldEntries.length) : iterations_;

        // the only loop that is not avoidable
        for (uint256 i = 0; i < onHoldEntries.length && iterations_ > 0; i++) {
            if (block.timestamp < onHoldEntries[i].releaseAt) {
                continue;
            }

            balance_ += onHoldEntries[i].amount;

            if (balance_ >= amountToRemove_) {
                onHoldEntries[i].amount = balance_ - amountToRemove_;
                return amountToRemove_;
            }

            // Remove entry by swapping with last element and popping
            uint256 lastIndex_ = onHoldEntries.length - 1;
            if (i < lastIndex_) {
                onHoldEntries[i] = onHoldEntries[lastIndex_];
                i--; // TODO: is it correct?
            }
            onHoldEntries.pop();

            iterations_--;
        }

        return balance_;
    }

    /////////////////////////
    //   STATS FUNCTIONS   //
    /////////////////////////

    /// @notice sets distibution pool configuration
    /// @dev parameters should be the same as in Ethereum L1 Distribution contract
    /// @dev at address 0x47176B2Af9885dC6C4575d4eFd63895f7Aaa4790
    /// @dev call 'Distribution.pools(3)' where '3' is a poolId
    function setPoolConfig(uint256 index, Pool calldata pool) public onlyOwner {
        if (index >= getPools().length) {
            revert PoolIndexOutOfBounds();
        }
        _getSessionStorage().pools[index] = pool;
    }

    function _maybeResetProviderRewardLimiter(Provider storage provider) private {
        if (block.timestamp > provider.limitPeriodEnd) {
            provider.limitPeriodEnd += PROVIDER_REWARD_LIMITER_PERIOD;
            provider.limitPeriodEarned = 0;
        }
    }

    /// @notice sends provider reward considering stake as the limit for the reward
    /// @param session session storage object
    /// @param reward_ amount of reward to send
    /// @param revertOnReachingLimit_ if true function will revert if reward is more than stake, otherwise just limit the reward
    function _rewardProvider(Session storage session, uint256 reward_, bool revertOnReachingLimit_) private {
        Provider storage provider = providers(session.provider);
        _maybeResetProviderRewardLimiter(provider);
        uint256 limit_ = provider.stake - provider.limitPeriodEarned;

        if (reward_ > limit_) {
            if (revertOnReachingLimit_) {
                revert WithdrawableBalanceLimitByStakeReached();
            }
            reward_ = limit_;
        }

        getToken().safeTransferFrom(getFundingAccount(), session.provider, reward_);

        session.providerWithdrawnAmount += reward_;
        increaseTotalClaimed(reward_);
        provider.limitPeriodEarned += reward_;
    }

    /// @notice returns amount of withdrawable user stake and one on hold
    function withdrawableUserStake(
        address user_,
        uint8 iterations_
    ) external view returns (uint256 avail_, uint256 hold_) {
        OnHold[] memory onHold = getOnHold(user_);
        iterations_ = iterations_ > onHold.length ? uint8(onHold.length) : iterations_;

        for (uint256 i = 0; i < onHold.length; i++) {
            uint256 amount = onHold[i].amount;
            if (block.timestamp < onHold[i].releaseAt) {
                hold_ += amount;
            } else {
                avail_ += amount;
            }
        }
    }

    function getSessionId(
        address user_,
        address provider_,
        uint256 stake_,
        uint256 sessionNonce_
    ) public pure returns (bytes32) {
        return keccak256(abi.encodePacked(user_, provider_, stake_, sessionNonce_));
    }

    /// @notice returns stipend of user based on their stake
    function stakeToStipend(uint256 sessionStake_, uint256 timestamp_) public view returns (uint256) {
        // inlined getTodaysBudget call to get a better precision
        return (sessionStake_ * getComputeBalance(timestamp_)) / (totalMORSupply(timestamp_) * 100);
    }

    /// @notice returns stake of user based on their stipend
    function stipendToStake(uint256 stipend_, uint256 timestamp_) public view returns (uint256) {
        // inlined getTodaysBudget call to get a better precision
        // return (stipend * totalMORSupply(timestamp)) / getTodaysBudget(timestamp);
        return (stipend_ * totalMORSupply(timestamp_) * 100) / getComputeBalance(timestamp_);
    }

    /// @dev make it pure
    function whenSessionEnds(
        uint256 sessionStake_,
        uint256 pricePerSecond_,
        uint256 openedAt_
    ) public view returns (uint256) {
        // if session stake is more than daily price then session will last for its max duration
        uint256 duration = stakeToStipend(sessionStake_, openedAt_) / pricePerSecond_;
        if (duration >= MAX_SESSION_DURATION) {
            return openedAt_ + MAX_SESSION_DURATION;
        }

        return openedAt_ + duration;
    }

    /// @notice returns today's budget in MOR
    function getTodaysBudget(uint256 timestamp_) public view returns (uint256) {
        return getComputeBalance(timestamp_) / 100; // 1% of Compute Balance
    }

    /// @notice returns today's compute balance in MOR
    function getComputeBalance(uint256 timestamp_) public view returns (uint256) {
        Pool memory pool = getPool(COMPUTE_POOL_INDEX);
        uint256 periodReward = LinearDistributionIntervalDecrease.getPeriodReward(
            pool.initialReward,
            pool.rewardDecrease,
            pool.payoutStart,
            pool.decreaseInterval,
            pool.payoutStart,
            uint128(startOfTheDay(timestamp_))
        );

        return periodReward - totalClaimed();
    }

    // returns total amount of MOR tokens that were distributed across all pools
    function totalMORSupply(uint256 timestamp_) public view returns (uint256) {
        uint256 startOfTheDay_ = startOfTheDay(timestamp_);
        uint256 totalSupply_ = 0;

        Pool[] memory pools = getPools();
        for (uint256 i = 0; i < pools.length; i++) {
            if (i == COMPUTE_POOL_INDEX) continue; // skip compute pool (it's calculated separately)

            Pool memory pool = pools[i];

            totalSupply_ += LinearDistributionIntervalDecrease.getPeriodReward(
                pool.initialReward,
                pool.rewardDecrease,
                pool.payoutStart,
                pool.decreaseInterval,
                pool.payoutStart,
                uint128(startOfTheDay_)
            );
        }

        return totalSupply_ + totalClaimed();
    }

    function getActiveBidsRatingByModel(
        bytes32 modelId_,
        uint256 offset_,
        uint8 limit_
    ) external view returns (bytes32[] memory, Bid[] memory, ProviderModelStats[] memory) {
        bytes32[] memory modelBidsSet_ = modelActiveBids(modelId_, offset_, limit_);
        uint256 length_ = modelBidsSet_.length;

        Bid[] memory bids_ = new Bid[](length_);
        bytes32[] memory bidIds_ = new bytes32[](length_);
        ProviderModelStats[] memory stats_ = new ProviderModelStats[](length_);

        for (uint i = 0; i < length_; i++) {
            bytes32 id_ = modelBidsSet_[i];
            bidIds_[i] = id_;
            Bid memory bid_ = getBid(id_);
            bids_[i] = bid_;
            stats_[i] = _getProviderModelStats(modelId_, bid_.provider);
        }

        return (bidIds_, bids_, stats_);
    }

    function startOfTheDay(uint256 timestamp_) public pure returns (uint256) {
        return timestamp_ - (timestamp_ % 1 days);
    }

    function _extractProviderApproval(bytes calldata providerApproval_) private view returns (bytes32) {
        (bytes32 bidId_, uint256 chainId_, address user_, uint128 timestamp_) = abi.decode(
            providerApproval_,
            (bytes32, uint256, address, uint128)
        );

        if (user_ != _msgSender()) {
            revert ApprovedForAnotherUser();
        }
        if (chainId_ != block.chainid) {
            revert WrongChaidId();
        }
        if (timestamp_ < block.timestamp - SIGNATURE_TTL) {
            revert SignatureExpired();
        }

        return bidId_;
    }

    function _extractReceipt(bytes calldata receiptEncoded_) private view returns (bytes32, uint32, uint32) {
        (bytes32 sessionId_, uint256 chainId_, uint128 timestamp_, uint32 tpsScaled1000_, uint32 ttftMs_) = abi.decode(
            receiptEncoded_,
            (bytes32, uint256, uint128, uint32, uint32)
        );

        if (chainId_ != block.chainid) {
            revert WrongChaidId();
        }
        if (timestamp_ < block.timestamp - SIGNATURE_TTL) {
            revert SignatureExpired();
        }

        return (sessionId_, tpsScaled1000_, ttftMs_);
    }

    function _getProviderClaimableBalance(Session memory session_) private view returns (uint256) {
        // if session was closed with no dispute - provider already got all funds
        //
        // if session was closed with dispute   -
        // if session was ended but not closed  -
        // if session was not ended             - provider can claim all funds except for today's session cost

        uint256 claimIntervalEnd_ = session_.closedAt.min(session_.endsAt.min(startOfTheDay(block.timestamp)));
        uint256 claimableDuration_ = claimIntervalEnd_.max(session_.openedAt) - session_.openedAt;
        uint256 totalCost_ = claimableDuration_ * session_.pricePerSecond;
        uint256 withdrawableAmount_ = totalCost_ - session_.providerWithdrawnAmount;

        return withdrawableAmount_;
    }

    /// @notice returns total claimanble balance for the provider for particular session
    function getProviderClaimableBalance(bytes32 sessionId_) public view returns (uint256) {
        Session memory session_ = _getSession(sessionId_);
        if (session_.openedAt == 0) {
            revert SessionNotFound();
        }

        return _getProviderClaimableBalance(session_);
    }

    /// @notice checks if receipt is valid
    function _isValidReceipt(
        address signer_,
        bytes calldata receipt_,
        bytes calldata signature_
    ) private pure returns (bool) {
        if (signature_.length == 0) {
            return false;
        }

        bytes32 receiptHash_ = ECDSA.toEthSignedMessageHash(keccak256(receipt_));

        return ECDSA.recover(receiptHash_, signature_) == signer_;
    }

    function _ownerOrProvider(address provider_) private view returns (bool) {
        return _msgSender() == owner() || _msgSender() == provider_;
    }

    function _ownerOrUser(address user_) private view returns (bool) {
        return _msgSender() == owner() || _msgSender() == user_;
    }
}
