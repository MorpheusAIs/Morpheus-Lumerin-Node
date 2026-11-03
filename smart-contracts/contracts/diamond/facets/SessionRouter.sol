// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IERC20, SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

import {ISessionRouter} from "../../interfaces/facets/ISessionRouter.sol";
import {OwnableDiamondStorage} from "../presets/OwnableDiamondStorage.sol";
import {SessionSettlement} from "../presets/SessionSettlement.sol";

contract SessionRouter is ISessionRouter, OwnableDiamondStorage, SessionSettlement {
    using SafeERC20 for IERC20;

    function __SessionRouter_init(
        address fundingAccount_,
        uint128 maxSessionDuration_,
        Pool[] calldata pools_
    ) external initializer(SESSIONS_STORAGE_SLOT) {
        SessionsStorage storage sessionsStorage = _getSessionsStorage();

        setMaxSessionDuration(maxSessionDuration_);
        sessionsStorage.fundingAccount = fundingAccount_;
        for (uint256 i = 0; i < pools_.length; ) {
            sessionsStorage.pools.push(pools_[i]);
            unchecked {
                i++;
            }
        }
    }

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

    function openSession(
        address user_,
        uint256 amount_,
        bool isDirectPaymentFromUser_,
        bytes calldata approvalEncoded_,
        bytes calldata signature_
    ) external returns (bytes32) {
        _validateDelegatee(_msgSender(), user_, DELEGATION_RULES_SESSION);

        SessionsStorage storage sessionsStorage = _getSessionsStorage();
        (bytes32 bidId_, ) = _extractProviderApproval(approvalEncoded_);
        address provider_ = _getBidsStorage().bids[bidId_].provider;
        bytes32 sessionId_ = _getSessionId(user_, provider_, bidId_, sessionsStorage.sessionNonce++);

        IERC20(_getBidsStorage().token).safeTransferFrom(user_, address(this), amount_);
        uint128 endsAt_ = _validateSession(
            bidId_,
            amount_,
            isDirectPaymentFromUser_,
            approvalEncoded_,
            signature_
        );
        _writeSession(sessionId_, user_, bidId_, amount_, isDirectPaymentFromUser_, approvalEncoded_, endsAt_);

        emit SessionOpened(user_, sessionId_, provider_);
        return sessionId_;
    }

    function closeSession(bytes calldata receiptEncoded_, bytes calldata signature_) external {
        (bytes32 sessionId_, uint32 tpsScaled1000_, uint32 ttftMs_) = _extractProviderReceipt(receiptEncoded_);
        if (_isPoolSession(sessionId_)) {
            revert SessionPoolSessionCannotBeClosed();
        }

        (
            Session storage session,
            Bid storage bid,
            bool noDispute_,
            UserSettlement memory userSettlement_
        ) = _closeSession(sessionId_, receiptEncoded_, signature_, _msgSender());

        if (userSettlement_.amountToLock != 0) {
            _getSessionsStorage().userStakesOnHold[session.user].push(
                OnHold(userSettlement_.amountToLock, userSettlement_.releaseAt)
            );
        }
        IERC20(_getBidsStorage().token).safeTransfer(session.user, userSettlement_.amountToReturn);

        _finalizeClosedSession(sessionId_, noDispute_, tpsScaled1000_, ttftMs_, session, bid);
    }

    function claimForProvider(bytes32 sessionId_) external {
        Session storage session = _getSessionsStorage().sessions[sessionId_];
        if (session.openedAt == 0) {
            revert SessionNotEndedOrNotExist();
        }
        Bid storage bid = _getBidsStorage().bids[session.bidId];

        uint256 rewards_ = _getProviderRewards(session, bid, true);
        uint256 onHoldAmount_ = _getProviderOnHoldAmount(session, bid);
        uint256 amount_ = rewards_ > onHoldAmount_ ? rewards_ - onHoldAmount_ : 0;

        _accrueProviderReward(sessionId_, session, bid, amount_, true);
        _payProviderDebt(bid.provider);
    }

    function withdrawUserStakes(address user_, uint8 iterations_) external {
        _validateDelegatee(_msgSender(), user_, DELEGATION_RULES_SESSION);

        OnHold[] storage onHoldEntries = _getSessionsStorage().userStakesOnHold[user_];
        uint8 count_ = iterations_ >= onHoldEntries.length ? uint8(onHoldEntries.length) : iterations_;
        uint256 length_ = onHoldEntries.length;
        uint256 amount_;

        if (length_ == 0 || count_ == 0) {
            revert SessionUserAmountToWithdrawIsZero();
        }

        uint8 removedCount_;
        for (uint256 i = length_; i > 0 && removedCount_ < count_; ) {
            unchecked {
                i--;
            }
            if (block.timestamp < onHoldEntries[i].releaseAt) {
                continue;
            }

            amount_ += onHoldEntries[i].amount;
            onHoldEntries[i] = onHoldEntries[length_ - 1];
            onHoldEntries.pop();
            unchecked {
                length_--;
                removedCount_++;
            }
        }

        if (amount_ == 0) {
            revert SessionUserAmountToWithdrawIsZero();
        }

        IERC20(_getBidsStorage().token).safeTransfer(user_, amount_);
        emit UserWithdrawn(user_, amount_);
    }
}
