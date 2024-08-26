import { IERC20 } from "@openzeppelin/contracts/interfaces/IERC20.sol";
import { Ownable } from "@openzeppelin/contracts/access/Ownable.sol";

// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

// Refer to https://www.rareskills.io/post/staking-algorithm
contract StakingMasterChef is Ownable {
  struct Pool {
    uint256 rewardPerSecondScaled; // reward tokens per second, times `PRECISION`
    uint256 lastRewardTime; // last time rewards were distributed
    uint256 accRewardPerShareScaled; // accumulated reward per share, times `PRECISION`
    uint256 totalShares; // total shares of reward token
    uint256 startTime; // start time of the staking for this pool
    uint256 endTime; // end time of the staking for this pool - after this time, no more rewards will be distributed
    Lock[] locks; // locks available for this pool: durations with corresponding multipliers
  }

  struct Lock {
    uint256 durationSeconds; // lock duration in seconds
    uint256 multiplierScaled; // multiplier for the lock duration, times `PRECISION`
  }

  struct UserStake {
    uint256 stakeAmount; // amount of staked tokens
    uint256 shareAmount; // shares received after staking
    uint256 rewardDebt; // reward debt
    uint256 lockEndsAt; // when staking lock duration ends
  }

  uint256 public constant PRECISION = 1e12; // precision multiplier for decimal calculations

  IERC20 public immutable stakingToken;
  IERC20 public immutable rewardToken;

  Pool[] public pools; // poolId => Pool
  mapping(uint256 => mapping(address => UserStake[])) poolUserStakes; // poolId => userAddress => stakeId => UserStake

  event Stake(address indexed userAddress, uint256 indexed poolId, uint256 stakeId, uint256 amount);
  event Unstake(address indexed userAddress, uint256 indexed poolId, uint256 stakeId, uint256 amount);
  event RewardWithdrawal(address indexed userAddress, uint256 indexed poolId, uint256 stakeId, uint256 amount);
  event PoolAdded(uint256 indexed poolId, uint256 startTime, uint256 endTime);
  event PoolStopped(uint256 indexed poolId);

  error PoolOrStakeNotExists();
  error StakingFinished();
  error LockNotEnded();
  error LockReleaseTimePastPoolEndTime(); // lock duration exceeds staking range, choose a shorter lock duration
  error NoRewardAvailable();

  constructor(IERC20 _stakingToken, IERC20 _rewardToken) Ownable(_msgSender()) {
    stakingToken = _stakingToken;
    rewardToken = _rewardToken;
  }

  /// @notice Adds a new pool, requires approval for the reward token
  /// @dev if lock duration is longer than pool duration it is accepted, but it won't be possible to stake with that lock duration
  /// @param _startTime time when the staking starts, epoch seconds
  /// @param _duration pool duration in seconds
  /// @param _lockDurations predefined lock durations for this pool with corresponding multipliers
  /// @return poolId the id of the new pool
  function addPool(
    uint256 _startTime,
    uint256 _duration,
    uint256 _totalReward,
    Lock[] memory _lockDurations
  ) external onlyOwner returns (uint256) {
    uint256 endTime = _startTime + _duration;
    uint256 poolId = pools.length;
    pools.push(
      Pool({
        startTime: _startTime,
        lastRewardTime: _startTime,
        endTime: endTime,
        rewardPerSecondScaled: (_totalReward * PRECISION) / _duration,
        locks: _lockDurations,
        accRewardPerShareScaled: 0,
        totalShares: 0
      })
    );
    emit PoolAdded(poolId, _startTime, endTime);

    rewardToken.transferFrom(_msgSender(), address(this), _totalReward);

    return poolId;
  }

  /// @notice Get the number of pools
  /// @return count the number of pools
  function getPoolsCount() external view returns (uint256) {
    return pools.length;
  }

  /// @notice Get the available lock durations of a pool with the corresponding multipliers
  /// @param _poolId the id of the pool
  /// @return locks locks for this pool
  function getLockDurations(uint256 _poolId) external view poolExists(_poolId) returns (Lock[] memory) {
    return pools[_poolId].locks;
  }

  /// @notice Stops the pool, no more rewards will be distributed
  /// @param _poolId the id of the pool
  function stopPool(uint256 _poolId) external onlyOwner poolExists(_poolId) {
    Pool storage pool = pools[_poolId]; // errors if poolId is invalid
    _recalculatePoolReward(pool);
    uint256 oldEndTime = pool.endTime;
    pool.endTime = block.timestamp;
    emit PoolStopped(_poolId);

    uint256 undistributedReward = ((oldEndTime - block.timestamp) * pool.rewardPerSecondScaled) / PRECISION;
    safeTransfer(_msgSender(), undistributedReward);
  }

  /// @notice Manually update pool reward variables
  /// @param _poolId the id of the pool
  function recalculatePoolReward(uint256 _poolId) external poolExists(_poolId) {
    Pool storage pool = pools[_poolId]; // errors if poolId is invalid
    _recalculatePoolReward(pool);
  }

  /// @dev Update reward variables of the given pool to be up-to-date.
  function _recalculatePoolReward(Pool storage _pool) private {
    uint256 timestamp = min(block.timestamp, _pool.endTime);
    if (timestamp <= _pool.lastRewardTime) {
      return;
    }

    if (_pool.totalShares != 0) {
      _pool.accRewardPerShareScaled = getRewardPerShareScaled(_pool, timestamp);
    }

    _pool.lastRewardTime = timestamp;
  }

  /// @dev calculate reward per share scaled without updating the pool
  function getRewardPerShareScaled(Pool storage _pool, uint256 _timestamp) private view returns (uint256) {
    uint256 rewardScaled = (_timestamp - _pool.lastRewardTime) * _pool.rewardPerSecondScaled;
    return _pool.accRewardPerShareScaled + (rewardScaled / _pool.totalShares);
  }

  /// @notice Deposit staking token
  /// @param _poolId the id of the pool
  /// @param _amount the amount of staking token
  /// @param _lockId the id for the predefined lock duration of the pool, earlier withdrawal is not possible
  /// @return stakeId the id of the new stake
  function stake(uint256 _poolId, uint256 _amount, uint8 _lockId) external poolExists(_poolId) returns (uint256) {
    Pool storage pool = pools[_poolId];
    if (block.timestamp >= pool.endTime) {
      revert StakingFinished();
    }
    Lock storage lock = pool.locks[_lockId];
    uint256 lockEndsAt = max(block.timestamp, pool.startTime) + lock.durationSeconds;
    if (lockEndsAt > pool.endTime) {
      revert LockReleaseTimePastPoolEndTime();
    }

    _recalculatePoolReward(pool);

    uint256 userShares = (_amount * lock.multiplierScaled) / PRECISION;
    pool.totalShares += userShares;

    UserStake[] storage userStakes = poolUserStakes[_poolId][_msgSender()];
    uint256 stakeId = userStakes.length;
    userStakes.push(
      UserStake({
        stakeAmount: _amount,
        shareAmount: userShares,
        rewardDebt: (userShares * pool.accRewardPerShareScaled) / PRECISION,
        lockEndsAt: lockEndsAt
      })
    );

    emit Stake(_msgSender(), _poolId, stakeId, _amount);
    stakingToken.transferFrom(address(_msgSender()), address(this), _amount);

    return stakeId;
  }

  /// @notice Withdraw staking token and reward
  /// @param _poolId the id of the pool
  /// @param _stakeId the id of the stake
  function unstake(uint256 _poolId, uint256 _stakeId) external {
    UserStake[] storage userStakes = poolUserStakes[_poolId][_msgSender()];
    if (_stakeId >= userStakes.length) {
      revert PoolOrStakeNotExists();
    }
    UserStake storage userStake = userStakes[_stakeId];
    Pool storage pool = pools[_poolId]; // errors if poolId is invalid

    // lockEndsAt cannot be larger than pool.endTime if stopPool is not called
    // if stopPool is called, lockEndsAt is not checked
    if ((block.timestamp > pool.startTime) && (block.timestamp < min(pool.endTime, userStake.lockEndsAt))) {
      revert LockNotEnded();
    }

    _recalculatePoolReward(pool);

    uint256 unstakeAmount = userStake.stakeAmount;
    uint256 reward = (userStake.shareAmount * pool.accRewardPerShareScaled) / PRECISION - userStake.rewardDebt;

    pool.totalShares -= userStake.shareAmount;

    userStake.rewardDebt = 0;
    userStake.stakeAmount = 0;
    userStake.shareAmount = 0;
    userStake.lockEndsAt = 0;

    emit Unstake(_msgSender(), _poolId, _stakeId, unstakeAmount);

    safeTransfer(_msgSender(), reward);
    stakingToken.transfer(address(_msgSender()), unstakeAmount);
  }

  /// @notice Get stake of a user in a pool
  /// @param _addr user address
  /// @param _poolId pool id
  /// @param _stakeId stake id
  /// @return userStake the stake information
  function getStake(address _addr, uint256 _poolId, uint256 _stakeId) external view returns (UserStake memory) {
    UserStake[] storage userStakes = poolUserStakes[_poolId][_addr];
    if (_stakeId >= userStakes.length) {
      revert PoolOrStakeNotExists();
    }
    return userStakes[_stakeId];
  }

  /// @notice Get all stakes of a user in a pool
  /// @param _addr user address
  /// @param _poolId pool id
  /// @return userStakes the stakes of the user
  function getStakes(address _addr, uint256 _poolId) external view poolExists(_poolId) returns (UserStake[] memory) {
    return poolUserStakes[_poolId][_addr];
  }

  /// @notice View function to see up-to-date reward of a user
  /// @param _user the user address
  /// @param _poolId the id of the pool
  /// @param _stakeId the id of the stake
  /// @return reward the reward amount of the reward token
  function getReward(address _user, uint256 _poolId, uint256 _stakeId) external view returns (uint256) {
    // we don't need to check pool.startTime because
    // staking is not allowed before startTime
    UserStake[] storage userStakes = poolUserStakes[_poolId][_user];
    if (_stakeId >= userStakes.length) {
      revert PoolOrStakeNotExists();
    }

    UserStake storage userStake = userStakes[_stakeId];
    if (userStake.shareAmount == 0) {
      // early exit if user has no stake
      // also avoids division by zero if pool has no shares
      // cause the only way it can happen is when everybody
      // unstaked and user calls getReward again
      return 0;
    }

    Pool storage pool = pools[_poolId];
    if (block.timestamp < pool.startTime) {
      return 0;
    }

    uint256 timestamp = min(block.timestamp, pool.endTime);
    return (userStake.shareAmount * getRewardPerShareScaled(pool, timestamp)) / PRECISION - userStake.rewardDebt;
  }

  /// @notice Withdraw reward token
  /// @param _poolId the id of the pool
  /// @param _stakeId the id of the stake
  function withdrawReward(uint256 _poolId, uint256 _stakeId) external {
    UserStake[] storage userStakes = poolUserStakes[_poolId][_msgSender()];
    if (_stakeId >= userStakes.length) {
      revert PoolOrStakeNotExists();
    }
    UserStake storage userStake = userStakes[_stakeId];
    Pool storage pool = pools[_poolId];
    _recalculatePoolReward(pool);

    uint256 rewardFromStart = (userStake.shareAmount * pool.accRewardPerShareScaled) / PRECISION;
    uint256 reward = rewardFromStart - userStake.rewardDebt;
    if (reward == 0) {
      revert NoRewardAvailable();
    }
    userStake.rewardDebt = rewardFromStart;

    emit RewardWithdrawal(_msgSender(), _poolId, _stakeId, reward);
    safeTransfer(_msgSender(), reward);
  }

  /// @dev Safe reward transfer function, just in case if rounding error causes pool to not have enough reward token.
  function safeTransfer(address _to, uint256 _amount) private {
    uint256 rewardBalance = rewardToken.balanceOf(address(this));
    rewardToken.transfer(_to, min(rewardBalance, _amount));
  }

  function max(uint256 _a, uint256 _b) private pure returns (uint256) {
    return _a > _b ? _a : _b;
  }

  function min(uint256 _a, uint256 _b) private pure returns (uint256) {
    return _a < _b ? _a : _b;
  }

  modifier poolExists(uint256 _poolId) {
    if (_poolId >= pools.length) {
      revert PoolOrStakeNotExists();
    }
    _;
  }
}
