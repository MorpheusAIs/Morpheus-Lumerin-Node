// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {Math} from "@openzeppelin/contracts/utils/math/Math.sol";
import {Ownable} from "@openzeppelin/contracts/access/Ownable.sol";
import {SafeERC20, IERC20} from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import {UUPSUpgradeable} from "@openzeppelin/contracts-upgradeable/proxy/utils/UUPSUpgradeable.sol";
import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";

import {PRECISION} from "@solarity/solidity-lib/utils/Globals.sol";

import {IStakingMasterChef} from "../interfaces/staking/IStakingMasterChef.sol";

// Due to impossibility of staking before the start time, it may be impossible to distribute the entire reward
// Refer to https://www.rareskills.io/post/staking-algorithm
contract StakingMasterChef is IStakingMasterChef, OwnableUpgradeable, UUPSUpgradeable {
    using Math for uint256;
    using SafeERC20 for IERC20;

    IERC20 public stakingToken;
    IERC20 public rewardToken;

    uint256 nextPoolId;

    mapping(uint256 poolId => Pool) public pools;
    mapping(uint256 poolId => PoolRateData) public poolRatesData;
    mapping(uint256 poolId => mapping(uint256 lockDuration => uint256 multiplierScaled)) public locks;
    mapping(uint256 poolId => mapping(address user => UserStake[])) public poolUserStakes;

    modifier poolExists(uint256 poolId_) {
        if (pools[poolId_].startTime == 0) {
            revert PoolNotExists();
        }
        _;
    }

    constructor() {
        _disableInitializers();
    }

    function __StakingMasterChef_init(address stakingToken_, address rewardToken_) public initializer {
        __Ownable_init();

        stakingToken = IERC20(stakingToken_);
        rewardToken = IERC20(rewardToken_);
    }

    function addPool(
        uint256 startTime_,
        uint256 duration_,
        uint256 totalReward_,
        uint256[] calldata lockDurations_,
        uint256[] calldata multipliersScaled_
    ) external onlyOwner returns (uint256) {
        if (startTime_ <= block.timestamp) {
            revert StartTimeIsPast();
        }
        if (duration_ == 0) {
            revert InvalidDuration();
        }
        if (totalReward_ == 0) {
            revert InvalidReward();
        }
        if (lockDurations_.length == 0) {
            revert InvalidLocksCount();
        }
        if (lockDurations_.length != multipliersScaled_.length) {
            revert InvalidLocksCount();
        }

        rewardToken.safeTransferFrom(_msgSender(), address(this), totalReward_);

        uint256 endTime_ = startTime_ + duration_;
        uint256 poolId_ = nextPoolId++;
        pools[poolId_] = Pool({
            startTime: startTime_,
            endTime: endTime_,
            rewardPerSecondScaled: (totalReward_ * PRECISION) / duration_,
            isTerminatedForced: false
        });

        poolRatesData[poolId_].lastRewardTime = startTime_;

        mapping(uint256 => uint256) storage poolLocks = locks[poolId_];
        for (uint256 i = 0; i < lockDurations_.length; i++) {
            poolLocks[lockDurations_[i]] = multipliersScaled_[i];
        }

        emit PoolAdded(poolId_, startTime_, endTime_);

        return poolId_;
    }

    function terminatePool(uint256 poolId_, address recipient_) external onlyOwner poolExists(poolId_) {
        Pool storage pool = pools[poolId_];

        if (block.timestamp >= pool.endTime) {
            revert PoolAlreadyTerminated();
        }

        uint256 timeTillEnd_ = pool.endTime - block.timestamp;
        // TODO: check if block.timestamp < startTime
        uint256 undistributedReward_ = (timeTillEnd_ * pool.rewardPerSecondScaled) / PRECISION;

        _recalculatePoolReward(poolId_);

        pool.endTime = block.timestamp;
        pool.isTerminatedForced = true;

        rewardToken.safeTransfer(recipient_, undistributedReward_);

        emit PoolTerminated(poolId_);
    }

    function stake(
        uint256 poolId_,
        uint256 amount_,
        uint256 lockDuration_
    ) external poolExists(poolId_) returns (uint256) {
        // TODO: add ability to stake and withdraw before start time
        if (amount_ == 0) {
            revert InvalidAmount();
        }

        Pool storage pool = pools[poolId_];
        if (block.timestamp < pool.startTime) {
            revert StakingNotStarted();
        }
        if (block.timestamp >= pool.endTime) {
            revert StakingFinished();
        }

        uint256 multiplierScaled_ = locks[poolId_][lockDuration_];
        if (multiplierScaled_ == 0) {
            revert InvalidLockDuration();
        }

        uint256 lockEnd_ = lockDuration_ + block.timestamp;
        if (lockEnd_ > pool.endTime) {
            revert LockReleaseTimePastPoolEndTime();
        }

        PoolRateData storage poolRateData = poolRatesData[poolId_];
        address user_ = _msgSender();

        stakingToken.safeTransferFrom(user_, address(this), amount_);

        _recalculatePoolReward(poolId_);

        uint256 userShares_ = (amount_ * multiplierScaled_) / PRECISION;

        UserStake[] storage userStakes = poolUserStakes[poolId_][user_];
        uint256 stakeId_ = userStakes.length;

        userStakes.push(
            UserStake({
                stakeAmount: amount_,
                shareAmount: userShares_,
                rewardDebt: (userShares_ * poolRateData.accRewardPerShareScaled) / PRECISION,
                lockEnd: lockEnd_
            })
        );

        poolRateData.totalShares += userShares_;

        emit Staked(user_, poolId_, stakeId_, amount_);

        return stakeId_;
    }

    function unstake(uint256 poolId_, uint256 stakeId_, address recipient_) external poolExists(poolId_) {
        address user_ = _msgSender();

        UserStake[] storage userStakes = poolUserStakes[poolId_][user_];
        if (stakeId_ >= userStakes.length) {
            revert StakeNotExists();
        }

        UserStake storage userStake = userStakes[stakeId_];
        if (userStake.shareAmount == 0) {
            revert StakeUnstaked();
        }

        Pool storage pool = pools[poolId_];

        if (userStake.lockEnd > block.timestamp && !pool.isTerminatedForced) {
            revert LockNotEnded();
        }

        _recalculatePoolReward(poolId_);

        PoolRateData storage poolRateData = poolRatesData[poolId_];

        uint256 unstakedAmount_ = userStake.stakeAmount;
        uint256 reward_ = (userStake.shareAmount * poolRateData.accRewardPerShareScaled) /
            PRECISION -
            userStake.rewardDebt;

        poolRateData.totalShares -= userStake.shareAmount;

        delete userStakes[stakeId_];

        stakingToken.safeTransfer(user_, unstakedAmount_);
        rewardToken.safeTransfer(recipient_, reward_);

        emit Unstaked(user_, poolId_, stakeId_, unstakedAmount_);
    }

    function withdrawReward(uint256 poolId_, uint256 stakeId_, address recipient_) external poolExists(poolId_) {
        address user_ = _msgSender();

        UserStake[] storage userStakes = poolUserStakes[poolId_][user_];
        if (stakeId_ >= userStakes.length) {
            revert StakeNotExists();
        }

        UserStake storage userStake = userStakes[stakeId_];
        if (userStake.shareAmount == 0) {
            revert StakeUnstaked();
        }

        _recalculatePoolReward(poolId_);

        PoolRateData storage poolRateData = poolRatesData[poolId_];

        uint256 rewardFromStart_ = (userStake.shareAmount * poolRateData.accRewardPerShareScaled) / PRECISION;
        uint256 reward_ = rewardFromStart_ - userStake.rewardDebt;
        if (reward_ == 0) {
            revert NoRewardAvailable();
        }

        userStake.rewardDebt = rewardFromStart_;

        rewardToken.safeTransfer(recipient_, reward_);

        emit RewardWithdrawed(user_, poolId_, stakeId_, reward_);
    }

    function recalculatePoolReward(uint256 poolId_) external poolExists(poolId_) {
        _recalculatePoolReward(poolId_);
    }

    /// @dev Update reward variables of the given pool to be up-to-date.
    function _recalculatePoolReward(uint256 poolId_) private {
        PoolRateData storage poolRateData = poolRatesData[poolId_];

        if (block.timestamp <= poolRateData.lastRewardTime) {
            return;
        }
        if (pools[poolId_].endTime <= poolRateData.lastRewardTime) {
            return;
        }

        poolRateData.accRewardPerShareScaled = _getRewardPerShareScaled(poolId_);
        poolRateData.lastRewardTime = block.timestamp.min(pools[poolId_].endTime);
    }

    function getReward(address user_, uint256 poolId_, uint256 stakeId_) external view returns (uint256) {
        UserStake[] storage userStakes = poolUserStakes[poolId_][user_];
        if (stakeId_ >= userStakes.length) {
            return 0;
        }

        UserStake storage userStake = userStakes[stakeId_];
        if (userStake.shareAmount == 0) {
            return 0;
        }

        uint256 totalUserReward_ = (userStake.shareAmount * _getRewardPerShareScaled(poolId_)) / PRECISION;

        return totalUserReward_ - userStake.rewardDebt;
    }

    /// @dev calculate reward per share scaled without updating the pool
    function _getRewardPerShareScaled(uint256 poolId_) private view returns (uint256) {
        PoolRateData storage poolRateData = poolRatesData[poolId_];

        if (poolRateData.totalShares == 0) {
            return poolRateData.accRewardPerShareScaled;
        }

        uint256 timestamp_ = block.timestamp.min(pools[poolId_].endTime);
        if (timestamp_ <= poolRateData.lastRewardTime) {
            return poolRateData.accRewardPerShareScaled;
        }

        uint256 timeSinceLastReward_ = timestamp_ - poolRateData.lastRewardTime;
        uint256 rewardScaled_ = timeSinceLastReward_ * pools[poolId_].rewardPerSecondScaled;
        uint256 rewardPerShareSinceLastReward_ = rewardScaled_ / poolRateData.totalShares;

        return poolRateData.accRewardPerShareScaled + rewardPerShareSinceLastReward_;
    }

    function _authorizeUpgrade(address) internal view override onlyOwner {}
}
