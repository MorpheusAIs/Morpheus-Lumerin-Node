// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IERC20, SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";
import {SafeCast} from "@openzeppelin/contracts/utils/math/SafeCast.sol";

import {ISessionRouter} from "../../interfaces/facets/ISessionRouter.sol";
import {BidStorage} from "../storages/BidStorage.sol";
import {DelegateStakingStorage} from "../storages/DelegateStakingStorage.sol";
import {DelegationStorage} from "../storages/DelegationStorage.sol";

abstract contract DelegateStakingPool is BidStorage, DelegationStorage, DelegateStakingStorage {
    using Math for uint256;
    using SafeCast for uint256;
    using SafeERC20 for IERC20;

    function _isGrantLive(Grant storage grant_) internal view returns (bool) {
        return !grant_.revoked && grant_.expiry > block.timestamp;
    }

    function _isPoolFunderApproved(PoolData storage pool_, address hot_, address funder_) internal view returns (bool) {
        return funder_ != address(0) && (funder_ == hot_ || pool_.grants[funder_].approved);
    }

    function _isActiveFunder(PoolData storage pool_, address hot_, address funder_) internal view returns (bool) {
        Grant storage grant_ = pool_.grants[funder_];
        return _isPoolFunderApproved(pool_, hot_, funder_) && _isGrantLive(grant_) && grant_.currentPrincipal != 0;
    }

    function _getLockedExposure(Grant storage grant_) internal view returns (uint256) {
        return block.timestamp < grant_.lockedUntil ? grant_.lockedExposure : 0;
    }

    function _getAvailablePrincipal(Grant storage grant_) internal view returns (uint256) {
        uint256 used_ = uint256(grant_.exposure) + _getLockedExposure(grant_);
        return grant_.currentPrincipal > used_ ? grant_.currentPrincipal - used_ : 0;
    }

    function _addLockedExposure(Grant storage grant_, uint256 amount_, uint128 releaseAt_) internal {
        if (amount_ == 0) {
            return;
        }
        if (block.timestamp >= grant_.lockedUntil) {
            grant_.lockedExposure = amount_.toUint128();
        } else {
            grant_.lockedExposure = (uint256(grant_.lockedExposure) + amount_).toUint128();
        }
        if (releaseAt_ > grant_.lockedUntil) {
            grant_.lockedUntil = uint256(releaseAt_).toUint64();
        }
    }

    function _admitActiveFunder(PoolData storage pool_, address hot_, address funder_) internal {
        uint256 maxActiveFunders_ = _maxActiveFunders();
        if (pool_.activeCount >= maxActiveFunders_) {
            _purgeInactiveFunders(pool_, hot_, MAX_ACTIVE_FUNDERS_LIMIT, false);
        }
        if (pool_.activeCount >= maxActiveFunders_) {
            revert DelegateStakingActiveFundersLimitExceeded();
        }
        _appendActiveFunder(pool_, funder_);
    }

    function _appendListNode(
        mapping(address => LinkedNode) storage nodes_,
        address head_,
        address tail_,
        address funder_
    ) internal returns (address, address) {
        LinkedNode storage node_ = nodes_[funder_];
        node_.linked = true;
        node_.previous = tail_;

        if (tail_ == address(0)) {
            head_ = funder_;
        } else {
            nodes_[tail_].next = funder_;
        }
        return (head_, funder_);
    }

    function _removeListNode(
        mapping(address => LinkedNode) storage nodes_,
        address head_,
        address tail_,
        address funder_
    ) internal returns (address, address) {
        LinkedNode storage node_ = nodes_[funder_];
        address previous_ = node_.previous;
        address next_ = node_.next;
        if (previous_ == address(0)) {
            head_ = next_;
        } else {
            nodes_[previous_].next = next_;
        }
        if (next_ == address(0)) {
            tail_ = previous_;
        } else {
            nodes_[next_].previous = previous_;
        }
        delete nodes_[funder_];
        return (head_, tail_);
    }

    function _appendActiveFunder(PoolData storage pool_, address funder_) internal {
        (pool_.activeHead, pool_.activeTail) = _appendListNode(
            pool_.activeNodes,
            pool_.activeHead,
            pool_.activeTail,
            funder_
        );
        unchecked {
            pool_.activeCount++;
        }
    }

    function _removeActiveFunder(PoolData storage pool_, address funder_) internal {
        LinkedNode storage node_ = pool_.activeNodes[funder_];
        if (!node_.linked) {
            return;
        }

        address next_ = node_.next;
        (pool_.activeHead, pool_.activeTail) = _removeListNode(
            pool_.activeNodes,
            pool_.activeHead,
            pool_.activeTail,
            funder_
        );
        if (pool_.purgeCursor == funder_) {
            pool_.purgeCursor = next_ == address(0) ? pool_.activeHead : next_;
        }
        unchecked {
            pool_.activeCount--;
        }
    }

    function _purgeInactiveFunders(
        PoolData storage pool_,
        address hot_,
        uint256 maxEntries_,
        bool useCursor_
    ) internal returns (uint256 removed_) {
        address current_ = useCursor_ && pool_.purgeCursor != address(0) ? pool_.purgeCursor : pool_.activeHead;
        uint256 inspected_;

        while (current_ != address(0) && inspected_ < maxEntries_) {
            address next_ = pool_.activeNodes[current_].next;
            if (!_isActiveFunder(pool_, hot_, current_)) {
                _removeActiveFunder(pool_, current_);
                unchecked {
                    removed_++;
                }
            }
            current_ = next_;
            unchecked {
                inspected_++;
            }
        }

        if (useCursor_) {
            pool_.purgeCursor = current_ == address(0) ? pool_.activeHead : current_;
        }
    }

    function _appendWithdrawal(PoolData storage pool_, address funder_) internal {
        (pool_.withdrawalHead, pool_.withdrawalTail) = _appendListNode(
            pool_.withdrawalNodes,
            pool_.withdrawalHead,
            pool_.withdrawalTail,
            funder_
        );
        pool_.withdrawalCount++;
    }

    function _removeWithdrawal(PoolData storage pool_, address funder_) internal {
        (pool_.withdrawalHead, pool_.withdrawalTail) = _removeListNode(
            pool_.withdrawalNodes,
            pool_.withdrawalHead,
            pool_.withdrawalTail,
            funder_
        );
        pool_.withdrawalCount--;
    }

    function _servicePoolWithdrawals(address hot_, uint256 maxEntries_) internal returns (uint256 serviced_) {
        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        IERC20 token_ = IERC20(_getBidsStorage().token);

        while (pool_.withdrawalHead != address(0) && serviced_ < maxEntries_ && pool_.free != 0) {
            address funder_ = pool_.withdrawalHead;
            Grant storage grant_ = pool_.grants[funder_];
            uint256 amount_ = uint256(grant_.pending).min(pool_.free);

            grant_.pending = (uint256(grant_.pending) - amount_).toUint128();
            pool_.pendingTotal = (uint256(pool_.pendingTotal) - amount_).toUint128();
            pool_.free = (uint256(pool_.free) - amount_).toUint128();
            if (grant_.pending == 0) {
                _removeWithdrawal(pool_, funder_);
            }
            unchecked {
                serviced_++;
            }

            token_.safeTransfer(funder_, amount_);
            emit DelegateStakingWithdrawalPaid(hot_, funder_, amount_);
        }
    }

    function _releaseMaturedPoolLocks(address hot_, uint256 maxBuckets_) internal returns (uint256 released_) {
        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        uint256 length_ = pool_.lockReleaseDates.length;

        while (pool_.lockHead < length_ && released_ < maxBuckets_) {
            uint256 head_ = pool_.lockHead;
            uint256 releaseAt_ = pool_.lockReleaseDates[head_];
            if (block.timestamp < releaseAt_) {
                break;
            }

            uint256 amount_ = pool_.lockAmounts[releaseAt_];
            delete pool_.lockAmounts[releaseAt_];
            delete pool_.lockReleaseDates[head_];
            unchecked {
                pool_.lockHead = head_ + 1;
                released_++;
            }
            pool_.locked = (uint256(pool_.locked) - amount_).toUint128();
            pool_.free = (uint256(pool_.free) + amount_).toUint128();

            emit DelegateStakingPoolLockReleased(hot_, uint128(releaseAt_), amount_);
        }
    }

    function _addPoolLock(PoolData storage pool_, address hot_, uint128 releaseAt_, uint256 amount_) internal {
        uint256 length_ = pool_.lockReleaseDates.length;
        if (length_ != pool_.lockHead && pool_.lockReleaseDates[length_ - 1] > releaseAt_) {
            revert DelegateStakingInvalidPoolSettlement();
        }
        if (length_ == pool_.lockHead || pool_.lockReleaseDates[length_ - 1] < releaseAt_) {
            pool_.lockReleaseDates.push(releaseAt_);
        }
        pool_.lockAmounts[releaseAt_] += amount_;
        pool_.locked = (uint256(pool_.locked) + amount_).toUint128();

        emit DelegateStakingPoolLockCreated(hot_, releaseAt_, amount_);
    }

    function _getPoolCapacity(PoolData storage pool_, address hot_) internal view returns (uint256 capacity_) {
        address current_ = pool_.activeHead;
        uint256 inspected_;

        while (current_ != address(0) && inspected_ < MAX_ACTIVE_FUNDERS_LIMIT) {
            if (_isActiveFunder(pool_, hot_, current_)) {
                capacity_ += _getAvailablePrincipal(pool_.grants[current_]);
            }
            current_ = pool_.activeNodes[current_].next;
            unchecked {
                inspected_++;
            }
        }

        Grant storage selfGrant_ = pool_.grants[hot_];
        if (_isGrantLive(selfGrant_)) {
            capacity_ += _getAvailablePrincipal(selfGrant_);
        }
    }

    function _drawPoolStake(address hot_, bytes32 sessionId_, uint256 amount_) internal {
        DelegateStakingData storage stakingStorage_ = _getDelegateStakingStorage();
        PoolData storage pool_ = stakingStorage_.pools[hot_];
        PoolSessionData storage poolSession_ = stakingStorage_.poolSessions[sessionId_];

        if (poolSession_.exists) {
            revert DelegateStakingPoolSessionAlreadyExists();
        }

        _releaseMaturedPoolLocks(hot_, _maxAutoReleaseDays());
        if (pool_.pendingTotal > pool_.free || amount_ > pool_.free - pool_.pendingTotal) {
            revert DelegateStakingInsufficientLiquidity();
        }
        poolSession_.hot = hot_;
        poolSession_.exists = true;
        uint256 remaining_ = amount_;
        address current_ = pool_.activeHead;
        uint256 inspected_;

        while (current_ != address(0) && inspected_ < MAX_ACTIVE_FUNDERS_LIMIT) {
            address next_ = pool_.activeNodes[current_].next;
            Grant storage grant_ = pool_.grants[current_];
            if (!_isActiveFunder(pool_, hot_, current_)) {
                _removeActiveFunder(pool_, current_);
            } else if (remaining_ != 0) {
                uint256 drawn_ = _getAvailablePrincipal(grant_).min(remaining_);
                if (drawn_ != 0) {
                    grant_.exposure = (uint256(grant_.exposure) + drawn_).toUint128();
                    remaining_ -= drawn_;
                    poolSession_.draws.push(PoolSessionDrawData(current_, drawn_.toUint96()));
                    emit DelegateStakingPoolDrawn(sessionId_, hot_, current_, drawn_);
                }
            }
            current_ = next_;
            unchecked {
                inspected_++;
            }
        }

        if (remaining_ != 0) {
            Grant storage selfGrant_ = pool_.grants[hot_];
            uint256 drawn_ = _isGrantLive(selfGrant_) ? _getAvailablePrincipal(selfGrant_).min(remaining_) : 0;
            if (drawn_ != 0) {
                selfGrant_.exposure = (uint256(selfGrant_.exposure) + drawn_).toUint128();
                remaining_ -= drawn_;
                poolSession_.draws.push(PoolSessionDrawData(hot_, drawn_.toUint96()));
                emit DelegateStakingPoolDrawn(sessionId_, hot_, hot_, drawn_);
            }
        }

        if (remaining_ != 0) {
            revert DelegateStakingInsufficientCapacity();
        }
        pool_.free = (uint256(pool_.free) - amount_).toUint128();
        pool_.inSession = (uint256(pool_.inSession) + amount_).toUint128();

        _servicePoolWithdrawals(hot_, _maxAutoService());
    }

    function _returnPoolStake(
        bytes32 sessionId_,
        uint256 stake_,
        uint256 lock_,
        uint128 releaseAt_,
        bool settled_
    ) internal {
        PoolSessionData storage poolSession_ = _getDelegateStakingStorage().poolSessions[sessionId_];
        if (!poolSession_.exists) {
            revert DelegateStakingPoolSessionNotFound();
        }
        if (poolSession_.closed) {
            revert ISessionRouter.SessionAlreadyClosed();
        }

        address hot_ = poolSession_.hot;
        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        if (
            lock_ > stake_ ||
            pool_.inSession < stake_ ||
            poolSession_.draws.length > MAX_POOL_SESSION_DRAWS ||
            (lock_ != 0 && releaseAt_ <= block.timestamp)
        ) {
            revert DelegateStakingInvalidPoolSettlement();
        }

        _releaseMaturedPoolLocks(hot_, _maxAutoReleaseDays());
        poolSession_.closed = true;
        poolSession_.settled = settled_;
        pool_.inSession = (uint256(pool_.inSession) - stake_).toUint128();
        pool_.free = (uint256(pool_.free) + stake_ - lock_).toUint128();
        if (lock_ != 0) {
            _addPoolLock(pool_, hot_, releaseAt_, lock_);
        }

        if (_releasePoolDraws(poolSession_, pool_, sessionId_, stake_, lock_, releaseAt_) != stake_) {
            revert DelegateStakingInvalidPoolSettlement();
        }
        delete poolSession_.draws;

        _servicePoolWithdrawals(hot_, _maxAutoService());
    }

    /**
     * @dev Returns every draw to its funder and splits the session day lock across the draws
     * pro rata to the drawn amounts. Cumulative rounding makes the shares add up to `lock_`.
     */
    function _releasePoolDraws(
        PoolSessionData storage poolSession_,
        PoolData storage pool_,
        bytes32 sessionId_,
        uint256 stake_,
        uint256 lock_,
        uint128 releaseAt_
    ) internal returns (uint256 returned_) {
        address hot_ = poolSession_.hot;
        uint256 lockedSoFar_;
        uint256 drawsLength_ = poolSession_.draws.length;
        for (uint256 i = 0; i < drawsLength_; ) {
            PoolSessionDrawData storage draw_ = poolSession_.draws[i];
            address funder_ = draw_.funder;
            uint256 amount_ = draw_.amount;
            Grant storage grant_ = pool_.grants[funder_];
            if (grant_.exposure < amount_) {
                revert DelegateStakingInvalidPoolSettlement();
            }
            grant_.exposure = (uint256(grant_.exposure) - amount_).toUint128();
            returned_ += amount_;
            if (lock_ != 0 && returned_ <= stake_) {
                uint256 lockedTotal_ = (returned_ * lock_) / stake_;
                _addLockedExposure(grant_, lockedTotal_ - lockedSoFar_, releaseAt_);
                lockedSoFar_ = lockedTotal_;
            }
            emit DelegateStakingPoolReturned(sessionId_, hot_, funder_, amount_);
            unchecked {
                i++;
            }
        }
    }

    function _getStakingAllowance(
        PoolData storage pool_,
        address hot_,
        address funder_
    ) internal view returns (StakingAllowance memory) {
        Grant storage grant_ = pool_.grants[funder_];
        return
            StakingAllowance(
                grant_.cumulativeFundingCap,
                grant_.expiry,
                grant_.revoked,
                grant_.lifetimeFunded,
                grant_.currentPrincipal,
                grant_.exposure,
                _getLockedExposure(grant_),
                grant_.pending,
                _isGrantLive(grant_),
                _isPoolFunderApproved(pool_, hot_, funder_)
            );
    }
}
