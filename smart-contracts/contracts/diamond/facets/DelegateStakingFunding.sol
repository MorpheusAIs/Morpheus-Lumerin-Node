// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IERC20, SafeERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";
import {SafeCast} from "@openzeppelin/contracts/utils/math/SafeCast.sol";
import {Paginator} from "@solarity/solidity-lib/libs/arrays/Paginator.sol";

import {IDelegateStakingFunding} from "../../interfaces/facets/IDelegateStakingFunding.sol";
import {DelegateStakingPool} from "../presets/DelegateStakingPool.sol";
import {OwnableDiamondStorage} from "../presets/OwnableDiamondStorage.sol";

contract DelegateStakingFunding is IDelegateStakingFunding, OwnableDiamondStorage, DelegateStakingPool {
    using Math for uint256;
    using Paginator for address[];
    using Paginator for uint256[];
    using SafeCast for uint256;
    using SafeERC20 for IERC20;

    function approvePoolFunder(address hot_, address funder_) external {
        _validateDelegatee(_msgSender(), hot_, DELEGATION_RULES_DELEGATE_STAKING);
        if (funder_ == address(0)) {
            revert DelegateStakingInvalidAddress();
        }

        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        pool_.grants[funder_].approved = true;
        if (funder_ != hot_ && _isActiveFunder(pool_, hot_, funder_) && !pool_.activeNodes[funder_].linked) {
            _admitActiveFunder(pool_, hot_, funder_);
        }

        emit DelegateStakingPoolFunderApproved(hot_, funder_);
    }

    function revokePoolFunderApproval(address hot_, address funder_) external {
        _validateDelegatee(_msgSender(), hot_, DELEGATION_RULES_DELEGATE_STAKING);
        if (funder_ == address(0)) {
            revert DelegateStakingInvalidAddress();
        }

        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        pool_.grants[funder_].approved = false;
        if (funder_ != hot_) {
            _removeActiveFunder(pool_, funder_);
        }

        emit DelegateStakingPoolFunderApprovalRevoked(hot_, funder_);
    }

    function grantStakingAllowance(address hot_, uint256 cap_, uint128 expiry_) external {
        if (hot_ == address(0)) {
            revert DelegateStakingInvalidAddress();
        }
        if (expiry_ <= block.timestamp) {
            revert DelegateStakingInvalidExpiry();
        }

        address funder_ = _msgSender();
        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        Grant storage grant_ = pool_.grants[funder_];
        if (cap_ < grant_.lifetimeFunded) {
            revert DelegateStakingCapBelowLifetimeFunding();
        }

        bool wasLive_ = _isGrantLive(grant_);
        grant_.cumulativeFundingCap = cap_;
        grant_.expiry = uint256(expiry_).toUint64();
        grant_.revoked = false;
        if (funder_ != hot_) {
            bool isActive_ = _isActiveFunder(pool_, hot_, funder_);
            if (isActive_ && !wasLive_) {
                // An expired node that no draw or purge has unlinked yet goes back to the FIFO tail.
                _removeActiveFunder(pool_, funder_);
            }
            if (isActive_ && !pool_.activeNodes[funder_].linked) {
                _admitActiveFunder(pool_, hot_, funder_);
            } else if (!isActive_) {
                _removeActiveFunder(pool_, funder_);
            }
        }

        emit DelegateStakingAllowanceGranted(hot_, funder_, cap_, expiry_);
    }

    function revokeStakingAllowance(address hot_) external {
        address funder_ = _msgSender();
        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        Grant storage grant_ = pool_.grants[funder_];
        if (!_isGrantLive(grant_)) {
            revert DelegateStakingAllowanceNotActive();
        }

        grant_.revoked = true;
        if (funder_ != hot_) {
            _removeActiveFunder(pool_, funder_);
        }

        emit DelegateStakingAllowanceRevoked(hot_, funder_);
    }

    function fundStakingAllowance(address hot_, uint256 amount_) external {
        if (amount_ == 0) {
            revert DelegateStakingInvalidAmount();
        }

        address funder_ = _msgSender();
        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        Grant storage grant_ = pool_.grants[funder_];
        if (!_isPoolFunderApproved(pool_, hot_, funder_)) {
            revert DelegateStakingPoolFunderNotApproved();
        }
        if (!_isGrantLive(grant_)) {
            revert DelegateStakingAllowanceNotActive();
        }
        if (
            grant_.lifetimeFunded > grant_.cumulativeFundingCap ||
            amount_ > grant_.cumulativeFundingCap - grant_.lifetimeFunded
        ) {
            revert DelegateStakingFundingCapExceeded();
        }
        if (grant_.currentPrincipal + amount_ < _minPrincipal()) {
            revert DelegateStakingPrincipalTooLow();
        }

        if (!grant_.known) {
            grant_.known = true;
            pool_.allFunders.push(funder_);
        }
        _releaseMaturedPoolLocks(hot_, _maxAutoReleaseDays());
        if (funder_ != hot_ && grant_.currentPrincipal == 0) {
            _removeActiveFunder(pool_, funder_);
        }
        grant_.lifetimeFunded += amount_;
        grant_.currentPrincipal = (uint256(grant_.currentPrincipal) + amount_).toUint128();
        pool_.free = (uint256(pool_.free) + amount_).toUint128();
        if (funder_ != hot_ && !pool_.activeNodes[funder_].linked) {
            _admitActiveFunder(pool_, hot_, funder_);
        }

        IERC20 token_ = IERC20(_getBidsStorage().token);
        token_.safeTransferFrom(funder_, address(this), amount_);
        emit DelegateStakingAllowanceFunded(hot_, funder_, amount_);

        _servicePoolWithdrawals(hot_, _maxAutoService());
    }

    function requestStakingWithdrawal(address hot_, address funder_, uint256 amount_) external {
        _validateDelegatee(_msgSender(), funder_, DELEGATION_RULES_DELEGATE_STAKING);
        if (amount_ == 0) {
            revert DelegateStakingInvalidAmount();
        }

        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        Grant storage grant_ = pool_.grants[funder_];
        if (amount_ > grant_.currentPrincipal) {
            revert DelegateStakingInsufficientPrincipal();
        }
        uint256 remainingPrincipal_ = grant_.currentPrincipal - amount_;
        if (remainingPrincipal_ != 0 && remainingPrincipal_ < _minPrincipal()) {
            revert DelegateStakingPrincipalTooLow();
        }

        _releaseMaturedPoolLocks(hot_, _maxAutoReleaseDays());
        grant_.currentPrincipal = remainingPrincipal_.toUint128();
        if (remainingPrincipal_ == 0 && funder_ != hot_) {
            _removeActiveFunder(pool_, funder_);
        }

        uint256 paid_;
        uint256 pending_ = amount_;
        if (pool_.withdrawalHead == address(0)) {
            paid_ = amount_.min(pool_.free);
            pool_.free = (uint256(pool_.free) - paid_).toUint128();
            pending_ -= paid_;
        }
        if (pending_ != 0) {
            grant_.pending = (uint256(grant_.pending) + pending_).toUint128();
            pool_.pendingTotal = (uint256(pool_.pendingTotal) + pending_).toUint128();
            if (!pool_.withdrawalNodes[funder_].linked) {
                _appendWithdrawal(pool_, funder_);
            }
        }

        emit DelegateStakingWithdrawalRequested(hot_, funder_, amount_, paid_, grant_.pending);

        if (paid_ != 0) {
            IERC20(_getBidsStorage().token).safeTransfer(funder_, paid_);
            emit DelegateStakingWithdrawalPaid(hot_, funder_, paid_);
        }
        _servicePoolWithdrawals(hot_, _maxAutoService());
    }

    function cancelStakingWithdrawal(address hot_, address funder_) external {
        _validateDelegatee(_msgSender(), funder_, DELEGATION_RULES_DELEGATE_STAKING);

        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        Grant storage grant_ = pool_.grants[funder_];
        uint256 amount_ = grant_.pending;
        if (amount_ == 0 || !pool_.withdrawalNodes[funder_].linked) {
            revert DelegateStakingWithdrawalNotFound();
        }

        uint256 restoredPrincipal_ = grant_.currentPrincipal + amount_;
        if (
            _isPoolFunderApproved(pool_, hot_, funder_) && _isGrantLive(grant_) && restoredPrincipal_ < _minPrincipal()
        ) {
            revert DelegateStakingPrincipalTooLow();
        }

        if (_isGrantLive(grant_) && funder_ != hot_ && grant_.currentPrincipal == 0) {
            _removeActiveFunder(pool_, funder_);
        }
        grant_.pending = 0;
        grant_.currentPrincipal = restoredPrincipal_.toUint128();
        pool_.pendingTotal = (uint256(pool_.pendingTotal) - amount_).toUint128();
        _removeWithdrawal(pool_, funder_);
        if (_isActiveFunder(pool_, hot_, funder_) && funder_ != hot_ && !pool_.activeNodes[funder_].linked) {
            _admitActiveFunder(pool_, hot_, funder_);
        }

        emit DelegateStakingWithdrawalCancelled(hot_, funder_, amount_);
    }

    function servePoolWithdrawals(address hot_, uint256 maxEntries_) external returns (uint256) {
        if (maxEntries_ > MAX_AUTO_SERVICE_LIMIT) {
            revert DelegateStakingLimitExceeded();
        }
        return _servicePoolWithdrawals(hot_, maxEntries_);
    }

    function releasePoolLocks(address hot_, uint256 maxBuckets_) external returns (uint256 released_) {
        if (maxBuckets_ > MAX_AUTO_RELEASE_DAYS_LIMIT) {
            revert DelegateStakingLimitExceeded();
        }

        released_ = _releaseMaturedPoolLocks(hot_, maxBuckets_);
        _servicePoolWithdrawals(hot_, _maxAutoService());
    }

    function purgeInactiveFunders(address hot_, uint256 maxEntries_) external returns (uint256) {
        if (maxEntries_ > MAX_ACTIVE_FUNDERS_LIMIT) {
            revert DelegateStakingLimitExceeded();
        }
        return _purgeInactiveFunders(_getDelegateStakingStorage().pools[hot_], hot_, maxEntries_, true);
    }

    function setMaxActiveFunders(uint256 value_) external onlyOwner {
        if (value_ > MAX_ACTIVE_FUNDERS_LIMIT) {
            revert DelegateStakingLimitExceeded();
        }
        _getDelegateStakingStorage().maxActiveFunders = value_.toUint8();
        emit DelegateStakingMaxActiveFundersUpdated(value_);
    }

    function setMinPrincipal(uint256 value_) external onlyOwner {
        if (value_ > MAX_MIN_PRINCIPAL_LIMIT) {
            revert DelegateStakingLimitExceeded();
        }
        _getDelegateStakingStorage().minPrincipal = value_.toUint96();
        emit DelegateStakingMinPrincipalUpdated(value_);
    }

    function setSettlementGrace(uint256 value_) external onlyOwner {
        if (value_ > MAX_SETTLEMENT_GRACE_LIMIT) {
            revert DelegateStakingLimitExceeded();
        }
        _getDelegateStakingStorage().settlementGrace = value_.toUint32();
        emit DelegateStakingSettlementGraceUpdated(value_);
    }

    function setMaxAutoService(uint256 value_) external onlyOwner {
        if (value_ > MAX_AUTO_SERVICE_LIMIT) {
            revert DelegateStakingLimitExceeded();
        }
        _getDelegateStakingStorage().maxAutoService = value_.toUint8();
        emit DelegateStakingMaxAutoServiceUpdated(value_);
    }

    function setMaxAutoReleaseDays(uint256 value_) external onlyOwner {
        if (value_ > MAX_AUTO_RELEASE_DAYS_LIMIT) {
            revert DelegateStakingLimitExceeded();
        }
        _getDelegateStakingStorage().maxAutoReleaseDays = value_.toUint8();
        emit DelegateStakingMaxAutoReleaseDaysUpdated(value_);
    }

    function getStakingAllowance(address hot_, address funder_) external view returns (StakingAllowance memory) {
        return _getStakingAllowance(_getDelegateStakingStorage().pools[hot_], hot_, funder_);
    }

    function isPoolFunderApproved(address hot_, address funder_) external view returns (bool) {
        return _isPoolFunderApproved(_getDelegateStakingStorage().pools[hot_], hot_, funder_);
    }

    function getAvailablePoolStake(address hot_) external view returns (uint256) {
        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        uint256 free_ = pool_.free;
        uint256 releaseLimit_ = _maxAutoReleaseDays();
        uint256 length_ = pool_.lockReleaseDates.length;
        uint256 released_;

        for (uint256 i = pool_.lockHead; i < length_ && released_ < releaseLimit_; ) {
            uint256 releaseAt_ = pool_.lockReleaseDates[i];
            if (block.timestamp < releaseAt_) {
                break;
            }
            free_ += pool_.lockAmounts[releaseAt_];
            unchecked {
                i++;
                released_++;
            }
        }

        uint256 liquidity_ = pool_.pendingTotal > free_ ? 0 : free_ - pool_.pendingTotal;
        return liquidity_.min(_getPoolCapacity(pool_, hot_));
    }

    function getPoolFunders(
        address hot_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (PoolFunder[] memory funders_, uint256 total_) {
        if (limit_ > MAX_ACTIVE_FUNDERS_LIMIT) {
            revert DelegateStakingLimitExceeded();
        }

        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        if (offset_ >= pool_.allFunders.length) {
            return (new PoolFunder[](0), pool_.allFunders.length);
        }

        address[] memory addresses_ = pool_.allFunders.part(offset_, limit_);
        funders_ = new PoolFunder[](addresses_.length);
        for (uint256 i = 0; i < addresses_.length; ) {
            funders_[i] = PoolFunder(
                addresses_[i],
                _getStakingAllowance(pool_, hot_, addresses_[i]),
                pool_.activeNodes[addresses_[i]].linked
            );
            unchecked {
                i++;
            }
        }
        return (funders_, pool_.allFunders.length);
    }

    function getStakingWithdrawal(address hot_, address funder_) external view returns (uint256) {
        return _getDelegateStakingStorage().pools[hot_].grants[funder_].pending;
    }

    function getPoolWithdrawalQueue(
        address hot_,
        address cursor_,
        uint256 limit_
    ) external view returns (PoolWithdrawal[] memory withdrawals_, address nextCursor_, uint256 total_) {
        if (limit_ > MAX_ACTIVE_FUNDERS_LIMIT) {
            revert DelegateStakingLimitExceeded();
        }

        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        total_ = pool_.withdrawalCount;
        if (cursor_ != address(0) && !pool_.withdrawalNodes[cursor_].linked) {
            revert DelegateStakingInvalidWithdrawalCursor();
        }

        address current_ = cursor_ == address(0) ? pool_.withdrawalHead : cursor_;
        address scan_ = current_;
        uint256 length_;
        while (scan_ != address(0) && length_ < limit_) {
            scan_ = pool_.withdrawalNodes[scan_].next;
            unchecked {
                length_++;
            }
        }

        withdrawals_ = new PoolWithdrawal[](length_);
        for (uint256 i = 0; i < length_; ) {
            withdrawals_[i] = PoolWithdrawal(current_, pool_.grants[current_].pending);
            current_ = pool_.withdrawalNodes[current_].next;
            unchecked {
                i++;
            }
        }
        nextCursor_ = scan_;
    }

    function getPoolBalances(address hot_) external view returns (PoolBalances memory) {
        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        return PoolBalances(pool_.free, pool_.inSession, pool_.locked, pool_.pendingTotal);
    }

    function getPoolLockBuckets(
        address hot_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (PoolLockBucket[] memory buckets_, uint256 total_) {
        if (limit_ > MAX_AUTO_RELEASE_DAYS_LIMIT) {
            revert DelegateStakingLimitExceeded();
        }

        PoolData storage pool_ = _getDelegateStakingStorage().pools[hot_];
        total_ = pool_.lockReleaseDates.length - pool_.lockHead;
        if (offset_ >= total_) {
            return (new PoolLockBucket[](0), total_);
        }

        uint256[] memory dates_ = pool_.lockReleaseDates.part(pool_.lockHead + offset_, limit_);
        buckets_ = new PoolLockBucket[](dates_.length);
        for (uint256 i = 0; i < dates_.length; ) {
            buckets_[i] = PoolLockBucket(uint128(dates_[i]), pool_.lockAmounts[dates_[i]]);
            unchecked {
                i++;
            }
        }
    }

    function getPoolSession(bytes32 sessionId_) external view returns (PoolSession memory session_) {
        PoolSessionData storage poolSession_ = _getDelegateStakingStorage().poolSessions[sessionId_];
        session_.hot = poolSession_.hot;
        session_.exists = poolSession_.exists;
        session_.closed = poolSession_.closed;
        session_.settled = poolSession_.settled;
        uint256 drawsLength_ = poolSession_.draws.length;
        session_.draws = new PoolSessionDraw[](drawsLength_);
        for (uint256 i = 0; i < drawsLength_; ) {
            PoolSessionDrawData storage draw_ = poolSession_.draws[i];
            session_.draws[i] = PoolSessionDraw(draw_.funder, draw_.amount);
            unchecked {
                i++;
            }
        }
    }

    function getDelegateStakingParameters() external view returns (DelegateStakingParameters memory) {
        return
            DelegateStakingParameters(
                _maxActiveFunders(),
                _minPrincipal(),
                _settlementGrace(),
                _maxAutoService(),
                _maxAutoReleaseDays()
            );
    }

    function getDelegateStakingLimits() external pure returns (DelegateStakingLimits memory) {
        return
            DelegateStakingLimits(
                MAX_ACTIVE_FUNDERS_LIMIT,
                MAX_AUTO_SERVICE_LIMIT,
                MAX_AUTO_RELEASE_DAYS_LIMIT,
                MAX_SETTLEMENT_GRACE_LIMIT,
                MAX_MIN_PRINCIPAL_LIMIT
            );
    }

    function getDelegateStakingDelegationRight() external pure returns (bytes32) {
        return DELEGATION_RULES_DELEGATE_STAKING;
    }
}
