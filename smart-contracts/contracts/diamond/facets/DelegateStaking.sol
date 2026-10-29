// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Context} from "@openzeppelin/contracts/utils/Context.sol";

import {IDelegateStaking} from "../../interfaces/facets/IDelegateStaking.sol";
import {ISessionRouter} from "../../interfaces/facets/ISessionRouter.sol";
import {DelegateStakingPool} from "../presets/DelegateStakingPool.sol";
import {SessionSettlement} from "../presets/SessionSettlement.sol";

contract DelegateStaking is IDelegateStaking, Context, SessionSettlement, DelegateStakingPool {
    function openSessionFromPool(
        address hot_,
        uint256 amount_,
        bytes calldata approvalEncoded_,
        bytes calldata signature_
    ) external returns (bytes32) {
        _validateDelegatee(_msgSender(), hot_, DELEGATION_RULES_SESSION);

        bytes32 bidId_;
        {
            address approvedUser_;
            (bidId_, approvedUser_) = _extractProviderApproval(approvalEncoded_);
            if (approvedUser_ != hot_) {
                revert ISessionRouter.SessionApprovedForAnotherUser();
            }
        }

        bytes32 sessionId_;
        address provider_;
        {
            SessionsStorage storage sessionsStorage_ = _getSessionsStorage();
            provider_ = _getBidsStorage().bids[bidId_].provider;
            sessionId_ = _getSessionId(hot_, provider_, bidId_, sessionsStorage_.sessionNonce++);
            if (sessionsStorage_.sessions[sessionId_].openedAt != 0) {
                revert DelegateStakingSessionSlotNotEmpty();
            }
        }

        uint128 endsAt_ = _validateSession(bidId_, amount_, false, approvalEncoded_, signature_);
        _drawPoolStake(hot_, sessionId_, amount_);
        _writeSession(sessionId_, hot_, bidId_, amount_, false, approvalEncoded_, endsAt_);

        emit SessionOpened(hot_, sessionId_, provider_);
        return sessionId_;
    }

    function closePoolSession(bytes calldata receiptEncoded_, bytes calldata signature_) external {
        (bytes32 sessionId_, uint32 tpsScaled1000_, uint32 ttftMs_) = _extractProviderReceipt(receiptEncoded_);
        PoolSessionData storage poolSession_ = _getDelegateStakingStorage().poolSessions[sessionId_];
        if (!poolSession_.exists) {
            revert DelegateStakingPoolSessionNotFound();
        }

        (
            Session storage session_,
            Bid storage bid_,
            bool noDispute_,
            UserSettlement memory settlement_
        ) = _closeSession(sessionId_, receiptEncoded_, signature_, _msgSender());
        _returnPoolStake(sessionId_, session_.stake, settlement_.amountToLock, settlement_.releaseAt, false);
        _finalizeClosedSession(sessionId_, noDispute_, tpsScaled1000_, ttftMs_, session_, bid_);
    }

    function settleExpiredSession(bytes32 sessionId_) external {
        PoolSessionData storage poolSession_ = _getDelegateStakingStorage().poolSessions[sessionId_];
        if (!poolSession_.exists) {
            revert DelegateStakingPoolSessionNotFound();
        }

        Session storage session_ = _getSessionsStorage().sessions[sessionId_];
        if (session_.closedAt != 0) {
            revert ISessionRouter.SessionAlreadyClosed();
        }
        if (block.timestamp < uint256(session_.endsAt) + _settlementGrace()) {
            revert DelegateStakingSettlementNotReady();
        }

        Bid storage bid_ = _getBidsStorage().bids[session_.bidId];
        address hot_ = poolSession_.hot;
        session_.isActive = false;
        session_.closedAt = uint128(block.timestamp);

        UserSettlement memory settlement_ = _calculateUserSettlement(session_, bid_);
        _returnPoolStake(sessionId_, session_.stake, settlement_.amountToLock, settlement_.releaseAt, true);
        _accrueProviderReward(sessionId_, session_, bid_, _getProviderRewards(session_, bid_, true), false);

        emit SessionClosed(session_.user, sessionId_, bid_.provider);
        emit DelegateStakingPoolSessionSettled(sessionId_, hot_, bid_.provider, _msgSender());
    }
}
