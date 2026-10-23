// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Address} from "@openzeppelin/contracts/utils/Address.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";
import {ECDSA} from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import {IERC20, SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {EnumerableSet} from "@openzeppelin/contracts/utils/structs/EnumerableSet.sol";

import {LinearDistributionIntervalDecrease} from "morpheus-smart-contracts/contracts/libs/LinearDistributionIntervalDecrease.sol";

import {ISessionRouter} from "../../interfaces/facets/ISessionRouter.sol";
import {ISessionSettlement} from "../../interfaces/facets/ISessionSettlement.sol";
import {IBidStorage} from "../../interfaces/storage/IBidStorage.sol";
import {IProviderStorage} from "../../interfaces/storage/IProviderStorage.sol";
import {ISessionStorage} from "../../interfaces/storage/ISessionStorage.sol";
import {IStatsStorage} from "../../interfaces/storage/IStatsStorage.sol";
import {LibSD} from "../../libs/LibSD.sol";
import {BidStorage} from "../storages/BidStorage.sol";
import {DelegateStakingStorage} from "../storages/DelegateStakingStorage.sol";
import {DelegationStorage} from "../storages/DelegationStorage.sol";
import {ProviderPayoutStorage} from "../storages/ProviderPayoutStorage.sol";
import {ProviderStorage} from "../storages/ProviderStorage.sol";
import {SessionStorage} from "../storages/SessionStorage.sol";
import {StatsStorage} from "../storages/StatsStorage.sol";

abstract contract SessionSettlement is
    ISessionSettlement,
    SessionStorage,
    ProviderStorage,
    BidStorage,
    StatsStorage,
    DelegationStorage,
    ProviderPayoutStorage,
    DelegateStakingStorage
{
    using Address for address;
    using EnumerableSet for EnumerableSet.Bytes32Set;
    using LibSD for LibSD.SD;
    using Math for uint256;
    using SafeERC20 for IERC20;

    uint256 internal constant PROVIDER_OWNER_STATICCALL_GAS = 100_000;

    struct UserSettlement {
        uint256 amountToReturn;
        uint256 amountToLock;
        uint128 releaseAt;
    }

    function _extractProviderApproval(
        bytes calldata providerApproval_
    ) internal view returns (bytes32 bidId_, address user_) {
        uint256 chainId_;
        uint128 timestamp_;
        (bidId_, chainId_, user_, timestamp_) = abi.decode(providerApproval_, (bytes32, uint256, address, uint128));

        if (chainId_ != block.chainid) {
            revert ISessionRouter.SessionApprovedForAnotherChainId();
        }
        if (block.timestamp > timestamp_ + SIGNATURE_TTL) {
            revert ISessionRouter.SessionApproveExpired();
        }
    }

    function _validateSession(
        bytes32 bidId_,
        uint256 amount_,
        bool isDirectPaymentFromUser_,
        bytes calldata approvalEncoded_,
        bytes calldata signature_
    ) internal view returns (uint128) {
        BidStorage.BidsStorage storage bidsStorage = _getBidsStorage();
        IBidStorage.Bid storage bid = bidsStorage.bids[bidId_];

        if (bid.createdAt == 0 || bid.deletedAt != 0) {
            revert ISessionRouter.SessionBidNotFound();
        }
        if (!_isValidProviderReceipt(bid.provider, approvalEncoded_, signature_)) {
            revert ISessionRouter.SessionProviderSignatureMismatch();
        }
        if (_getSessionsStorage().isProviderApprovalUsed[approvalEncoded_]) {
            revert ISessionRouter.SessionDuplicateApproval();
        }

        uint128 endsAt_ = _getSessionEnd(amount_, bid.pricePerSecond, uint128(block.timestamp));
        uint128 duration_ = endsAt_ - uint128(block.timestamp);

        if (duration_ < MIN_SESSION_DURATION) {
            revert ISessionRouter.SessionTooShort();
        }
        if (isDirectPaymentFromUser_ && (duration_ * bid.pricePerSecond) > amount_) {
            revert ISessionRouter.SessionStakeTooLow();
        }

        return endsAt_;
    }

    function _writeSession(
        bytes32 sessionId_,
        address user_,
        bytes32 bidId_,
        uint256 amount_,
        bool isDirectPaymentFromUser_,
        bytes calldata approvalEncoded_,
        uint128 endsAt_
    ) internal {
        SessionStorage.SessionsStorage storage sessionsStorage = _getSessionsStorage();
        BidStorage.BidsStorage storage bidsStorage = _getBidsStorage();
        ISessionStorage.Session storage session = sessionsStorage.sessions[sessionId_];
        uint128 openedAt_ = uint128(block.timestamp);

        session.user = user_;
        session.stake = amount_;
        session.bidId = bidId_;
        session.openedAt = openedAt_;
        session.endsAt = endsAt_;
        session.isActive = true;
        session.isDirectPaymentFromUser = isDirectPaymentFromUser_;

        IBidStorage.Bid storage bid = bidsStorage.bids[bidId_];
        sessionsStorage.userSessions[user_].add(sessionId_);
        sessionsStorage.providerSessions[bid.provider].add(sessionId_);
        sessionsStorage.modelSessions[bid.modelId].add(sessionId_);
        sessionsStorage.isProviderApprovalUsed[approvalEncoded_] = true;
    }

    function _closeSession(
        bytes32 sessionId_,
        bytes calldata receiptEncoded_,
        bytes calldata signature_,
        address caller_
    )
        internal
        returns (
            ISessionStorage.Session storage session_,
            IBidStorage.Bid storage bid_,
            bool noDispute_,
            UserSettlement memory settlement_
        )
    {
        session_ = _getSessionsStorage().sessions[sessionId_];
        bid_ = _getBidsStorage().bids[session_.bidId];

        _validateDelegatee(caller_, session_.user, DELEGATION_RULES_SESSION);
        if (session_.closedAt != 0) {
            revert ISessionRouter.SessionAlreadyClosed();
        }

        session_.isActive = false;
        session_.closeoutReceipt = receiptEncoded_;
        session_.closedAt = uint128(block.timestamp);

        noDispute_ = _isValidProviderReceipt(bid_.provider, receiptEncoded_, signature_);
        settlement_ = _calculateUserSettlement(session_, bid_);
    }

    function _finalizeClosedSession(
        bytes32 sessionId_,
        bool noDispute_,
        uint32 tpsScaled1000_,
        uint32 ttftMs_,
        ISessionStorage.Session storage session_,
        IBidStorage.Bid storage bid_
    ) internal {
        _rewardProviderAfterClose(sessionId_, noDispute_, session_, bid_);
        _setStats(sessionId_, noDispute_, ttftMs_, tpsScaled1000_, session_, bid_);

        emit ISessionRouter.SessionClosed(session_.user, sessionId_, bid_.provider);
    }

    function _extractProviderReceipt(
        bytes calldata receiptEncoded_
    ) internal view returns (bytes32 sessionId_, uint32 tpsScaled1000_, uint32 ttftMs_) {
        uint256 chainId_;
        uint128 timestamp_;
        (sessionId_, chainId_, timestamp_, tpsScaled1000_, ttftMs_) = abi.decode(
            receiptEncoded_,
            (bytes32, uint256, uint128, uint32, uint32)
        );

        if (chainId_ != block.chainid) {
            revert ISessionRouter.SessionReceiptForAnotherChainId();
        }
        if (block.timestamp > timestamp_ + SIGNATURE_TTL) {
            revert ISessionRouter.SessionReceiptExpired();
        }
    }

    function _getSessionId(
        address user_,
        address provider_,
        bytes32 bidId_,
        uint256 sessionNonce_
    ) internal pure returns (bytes32) {
        return keccak256(abi.encodePacked(user_, provider_, bidId_, sessionNonce_));
    }

    function _getSessionEnd(
        uint256 amount_,
        uint256 pricePerSecond_,
        uint128 openedAt_
    ) internal view returns (uint128) {
        uint128 duration_ = uint128(_stakeToStipend(amount_, openedAt_) / pricePerSecond_);
        uint128 maxSessionDuration_ = _getSessionsStorage().maxSessionDuration;

        if (duration_ > maxSessionDuration_) {
            duration_ = maxSessionDuration_;
        }

        return openedAt_ + duration_;
    }

    function _stakeToStipend(uint256 amount_, uint128 timestamp_) internal view returns (uint256) {
        uint256 totalMorSupply_ = _totalMORSupply(timestamp_);
        if (totalMorSupply_ == 0) {
            return 0;
        }

        return (amount_ * _getComputeBalance(timestamp_)) / (totalMorSupply_ * 100);
    }

    function _stipendToStake(uint256 stipend_, uint128 timestamp_) internal view returns (uint256) {
        uint256 computeBalance_ = _getComputeBalance(timestamp_);
        if (computeBalance_ == 0) {
            return 0;
        }

        return (stipend_ * _totalMORSupply(timestamp_) * 100) / computeBalance_;
    }

    function _calculateUserSettlement(
        ISessionStorage.Session storage session,
        IBidStorage.Bid storage bid
    ) internal view returns (UserSettlement memory settlement_) {
        uint128 sessionEnd_ = uint128(uint256(session.closedAt).min(session.endsAt));
        uint128 startOfEndDay_ = _startOfTheDay(sessionEnd_);
        settlement_.releaseAt = startOfEndDay_ + 1 days;

        uint256 userStakeToProvider_ = session.isDirectPaymentFromUser ? _getProviderRewards(session, bid, false) : 0;
        uint256 userStake_ = session.stake - userStakeToProvider_;

        if (block.timestamp < settlement_.releaseAt) {
            uint256 userDuration_ = sessionEnd_ - uint256(session.openedAt).max(startOfEndDay_);
            uint256 userInitialLock_ = userDuration_ * bid.pricePerSecond;
            settlement_.amountToLock = userStake_.min(_stipendToStake(userInitialLock_, startOfEndDay_));
        }

        settlement_.amountToReturn = userStake_ - settlement_.amountToLock;
    }

    function _rewardProviderAfterClose(
        bytes32 sessionId_,
        bool noDispute_,
        ISessionStorage.Session storage session,
        IBidStorage.Bid storage bid
    ) internal {
        uint256 providerAmountToWithdraw_ = _getProviderRewards(session, bid, true);

        if (!noDispute_ && session.closedAt < session.endsAt) {
            uint256 onHoldAmount_ = _getProviderOnHoldAmount(session, bid);
            providerAmountToWithdraw_ -= onHoldAmount_;
        }

        _accrueProviderReward(sessionId_, session, bid, providerAmountToWithdraw_, true);
    }

    function _accrueProviderReward(
        bytes32 sessionId_,
        ISessionStorage.Session storage session,
        IBidStorage.Bid storage bid,
        uint256 amount_,
        bool payNow_
    ) internal {
        IProviderStorage.Provider storage provider = _getProvidersStorage().providers[bid.provider];

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

        SessionStorage.SessionsStorage storage sessionsStorage = _getSessionsStorage();
        ProviderPayoutsStorage storage payoutStorage = _getProviderPayoutsStorage();
        IERC20 token_;
        uint256 paid_;
        uint256 owed_;

        if (session.isDirectPaymentFromUser) {
            paid_ = amount_;
        } else if (payNow_) {
            token_ = IERC20(_getBidsStorage().token);
            if (_canFundingPayAfterDebts(token_, sessionsStorage.fundingAccount, payoutStorage, amount_)) {
                paid_ = amount_;
            } else {
                owed_ = amount_;
            }
        } else {
            owed_ = amount_;
        }

        if (paid_ != 0) {
            sessionsStorage.providersTotalClaimed += paid_;
            payoutStorage.lifetimeClaimed[bid.provider] += paid_;
        } else {
            payoutStorage.providerOwed[bid.provider] += owed_;
            payoutStorage.pendingProviderClaims += owed_;
        }

        if (paid_ != 0) {
            address payoutTarget_ = _getPayoutTarget(bid.provider);
            if (session.isDirectPaymentFromUser) {
                token_ = IERC20(_getBidsStorage().token);
                token_.safeTransfer(payoutTarget_, paid_);
            } else {
                token_.safeTransferFrom(sessionsStorage.fundingAccount, payoutTarget_, paid_);
            }
        }

        emit ProviderRewardAccrued(sessionId_, bid.provider, amount_, paid_, owed_);
    }

    /**
     * @dev Outstanding provider debts are reserved first: a new reward is paid at accrual only if the
     * funding account covers all pending debts plus this amount. Debts themselves are paid on claim.
     */
    function _canFundingPayAfterDebts(
        IERC20 token_,
        address fundingAccount_,
        ProviderPayoutsStorage storage payoutStorage,
        uint256 amount_
    ) private view returns (bool) {
        uint256 reserved_ = payoutStorage.pendingProviderClaims;
        uint256 balance_ = token_.balanceOf(fundingAccount_);
        if (balance_ < reserved_ || balance_ - reserved_ < amount_) {
            return false;
        }

        uint256 allowance_ = token_.allowance(fundingAccount_, address(this));
        return allowance_ >= reserved_ && allowance_ - reserved_ >= amount_;
    }

    function _payProviderDebt(address provider_) internal returns (uint256 amount_) {
        ProviderPayoutsStorage storage payoutStorage = _getProviderPayoutsStorage();
        uint256 owed_ = payoutStorage.providerOwed[provider_];
        if (owed_ == 0) {
            return 0;
        }

        SessionStorage.SessionsStorage storage sessionsStorage = _getSessionsStorage();
        IERC20 token_ = IERC20(_getBidsStorage().token);
        amount_ = owed_.min(token_.balanceOf(sessionsStorage.fundingAccount));
        if (amount_ != 0) {
            amount_ = amount_.min(token_.allowance(sessionsStorage.fundingAccount, address(this)));
        }

        if (amount_ == 0) {
            return 0;
        }

        payoutStorage.providerOwed[provider_] -= amount_;
        payoutStorage.pendingProviderClaims -= amount_;
        sessionsStorage.providersTotalClaimed += amount_;
        payoutStorage.lifetimeClaimed[provider_] += amount_;

        address payoutTarget_ = _getPayoutTarget(provider_);
        token_.safeTransferFrom(sessionsStorage.fundingAccount, payoutTarget_, amount_);

        emit ProviderDebtPaid(provider_, payoutTarget_, amount_);
    }

    function _getProviderRewards(
        ISessionStorage.Session storage session,
        IBidStorage.Bid storage bid,
        bool includeWithdrawnAmount_
    ) internal view returns (uint256) {
        uint256 sessionEnd_ = session.closedAt == 0 ? session.endsAt : uint256(session.closedAt).min(session.endsAt);
        if (block.timestamp < sessionEnd_) {
            return 0;
        }

        uint256 reward_ = (sessionEnd_ - session.openedAt) * bid.pricePerSecond;
        uint256 withdrawnAmount_ = includeWithdrawnAmount_ ? session.providerWithdrawnAmount : 0;
        return reward_ - withdrawnAmount_;
    }

    function _getProviderOnHoldAmount(
        ISessionStorage.Session storage session,
        IBidStorage.Bid storage bid
    ) internal view returns (uint256) {
        uint128 startOfClosedAt_ = _startOfTheDay(session.closedAt);
        if (block.timestamp >= startOfClosedAt_ + 1 days) {
            return 0;
        }

        return (session.closedAt - uint256(startOfClosedAt_).max(session.openedAt)) * bid.pricePerSecond;
    }

    function _setStats(
        bytes32 sessionId_,
        bool noDispute_,
        uint32 ttftMs_,
        uint32 tpsScaled1000_,
        ISessionStorage.Session storage session,
        IBidStorage.Bid storage bid
    ) internal {
        IStatsStorage.ProviderModelStats storage prStats = _providerModelStats(bid.modelId, bid.provider);
        IStatsStorage.ModelStats storage modelStats = _modelStats(bid.modelId);

        prStats.totalCount++;

        if (noDispute_) {
            if (prStats.successCount > 0) {
                modelStats.tpsScaled1000.remove(int32(prStats.tpsScaled1000.mean), int32(modelStats.count - 1));
                modelStats.ttftMs.remove(int32(prStats.ttftMs.mean), int32(modelStats.count - 1));
            } else {
                modelStats.count++;
            }

            prStats.successCount++;
            prStats.totalDuration += uint32(session.closedAt - session.openedAt);
            prStats.tpsScaled1000.add(int32(tpsScaled1000_), int32(prStats.successCount));
            prStats.ttftMs.add(int32(ttftMs_), int32(prStats.successCount));

            modelStats.totalDuration.add(int32(prStats.totalDuration), int32(modelStats.count));
            modelStats.tpsScaled1000.add(int32(prStats.tpsScaled1000.mean), int32(modelStats.count));
            modelStats.ttftMs.add(int32(prStats.ttftMs.mean), int32(modelStats.count));
        } else {
            session.closeoutType = 1;
        }

        emit SessionRating(sessionId_, bid.modelId, bid.provider, session.closeoutType, tpsScaled1000_, ttftMs_);
    }

    function _getComputeBalance(uint128 timestamp_) internal view returns (uint256) {
        SessionStorage.SessionsStorage storage sessionsStorage = _getSessionsStorage();
        ISessionStorage.Pool memory pool_ = sessionsStorage.pools[COMPUTE_POOL_INDEX];
        uint256 periodReward_ = LinearDistributionIntervalDecrease.getPeriodReward(
            pool_.initialReward,
            pool_.rewardDecrease,
            pool_.payoutStart,
            pool_.decreaseInterval,
            pool_.payoutStart,
            _startOfTheDay(timestamp_)
        );
        uint256 emissionDebit_ = sessionsStorage.providersTotalClaimed +
            _getProviderPayoutsStorage().pendingProviderClaims;

        return periodReward_ > emissionDebit_ ? periodReward_ - emissionDebit_ : 0;
    }

    function _totalMORSupply(uint128 timestamp_) internal view returns (uint256) {
        uint128 startOfTheDay_ = _startOfTheDay(timestamp_);
        uint256 totalSupply_;
        SessionStorage.SessionsStorage storage sessionsStorage = _getSessionsStorage();
        uint256 poolsLength_ = sessionsStorage.pools.length;

        for (uint256 i = 0; i < poolsLength_; ) {
            if (i == COMPUTE_POOL_INDEX) {
                unchecked {
                    i++;
                }
                continue;
            }

            ISessionStorage.Pool memory pool_ = sessionsStorage.pools[i];
            totalSupply_ += LinearDistributionIntervalDecrease.getPeriodReward(
                pool_.initialReward,
                pool_.rewardDecrease,
                pool_.payoutStart,
                pool_.decreaseInterval,
                pool_.payoutStart,
                startOfTheDay_
            );
            unchecked {
                i++;
            }
        }

        uint256 emissionDebit_ = sessionsStorage.providersTotalClaimed +
            _getProviderPayoutsStorage().pendingProviderClaims;

        return totalSupply_ + emissionDebit_;
    }

    function _startOfTheDay(uint128 timestamp_) internal pure returns (uint128) {
        return timestamp_ - (timestamp_ % 1 days);
    }

    function _isValidProviderReceipt(
        address provider_,
        bytes calldata receipt_,
        bytes calldata signature_
    ) internal view returns (bool) {
        bytes32 receiptHash_ = ECDSA.toEthSignedMessageHash(keccak256(receipt_));
        address signer_ = ECDSA.recover(receiptHash_, signature_);

        if (signer_ == provider_ || isRightsDelegated(signer_, provider_, DELEGATION_RULES_PROVIDER)) {
            return true;
        }

        if (provider_.isContract()) {
            bool success_;
            address owner_;

            assembly {
                let ptr_ := mload(0x40)
                mstore(ptr_, shl(224, 0x8da5cb5b))
                success_ := staticcall(PROVIDER_OWNER_STATICCALL_GAS, provider_, ptr_, 4, ptr_, 32)
                success_ := and(success_, eq(returndatasize(), 32))
                let ownerWord_ := mload(ptr_)
                success_ := and(success_, iszero(shr(160, ownerWord_)))
                owner_ := ownerWord_
            }

            if (success_ && signer_ == owner_) {
                return true;
            }
        }

        return false;
    }
}
