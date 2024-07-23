import { IERC20 } from "@openzeppelin/contracts/interfaces/IERC20.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";

// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

// Refer to https://www.rareskills.io/post/staking-algorithm
contract StakingMasterChef {
  using SafeERC20 for IERC20;

  struct Pool {
    uint256 rewardPerSecond; // reward tokens per second
    uint256 lastRewardTime; // last time rewards were distributed
    uint256 accRewardPerShareScaled; // accumulated reward per share, times `PRECISION`
    uint256 totalShares; // total shares of reward token
    uint256 startTime; // start time of the staking for this pool
    uint256 endTime; // end time of the staking for this pool - after this time, no more rewards will be distributed
    LockDuration[] lockDuration; // lock durations for this pool with corresponding multipliers
  }

  struct LockDuration {
    uint256 durationSeconds; // lock duration in seconds
    uint256 multiplierScaled; // multiplier for the lock duration, times `PRECISION`
  }

  struct UserStake {
    uint256 amount; // amount of staked tokens
    uint256 shareAmount; // shares received after staking
    uint256 rewardDebt; // reward debt
    uint256 lockEndsAt; // when staking lock duration ends
  }

  uint256 public constant PRECISION = 1e12; // precision multiplier for decimal calculations

  IERC20 public immutable stakingToken;
  IERC20 public immutable rewardToken;

  address public owner; // the owner of the contract
  address public fundingAccount; // account which stores the reward tokens with allowance for this contract
  Pool[] public pools; // poolId => Pool
  mapping(uint256 => mapping(address => UserStake[])) poolUserStakes; // poolId => userAddress => UserStake

  event Stake(address indexed userAddress, uint256 indexed poolId, uint256 stakeId, uint256 amount);
  event Unstake(address indexed userAddress, uint256 indexed poolId, uint256 stakeId, uint256 amount);
  event RewardWithdrawal(address indexed userAddress, uint256 indexed poolId, uint256 stakeId, uint256 amount);
  event PoolAdded(uint256 indexed poolId, uint256 startTime, uint256 endTime);
  event PoolStopped(uint256 indexed poolId);

  error Unauthorized();
  error PoolNotFound();
  error StakeNotFound();
  error StakingNotStarted();
  error StakingFinished();
  error LockDurationNotOver();
  error LockDurationExceedsStakingRange(); // lock duration exceeds staking range, choose a shorter lock duration
  error NoRewardAvailable();

  constructor(IERC20 _stakingToken, IERC20 _rewardToken, address _fundingAccount) {
    owner = msg.sender;
    stakingToken = _stakingToken;
    rewardToken = _rewardToken;
    fundingAccount = _fundingAccount;
  }

  /// @notice Add a new pool
  /// @param _rewardPerSecond how many reward tokens are distributed per second of the staking between all of the participants
  /// @param _startTime time when the staking starts, epoch seconds
  /// @param _endTime time when the staking ends, epoch seconds
  /// @param _lockDurations predefined lock durations for this pool with corresponding multipliers
  /// @return poolId the id of the new pool
  function addPool(
    uint256 _rewardPerSecond,
    uint256 _startTime,
    uint256 _endTime,
    LockDuration[] memory _lockDurations
  ) external onlyOwner returns (uint256) {
    // TODO: enforce there is enough funds for the rewards

    // if lock duration is longer than pool duration it is accepted,
    // but nobody will be able to stake with that lock duration
    uint256 poolId = pools.length;
    pools.push(
      Pool({
        startTime: _startTime,
        lastRewardTime: _startTime,
        endTime: _endTime,
        rewardPerSecond: _rewardPerSecond,
        lockDuration: _lockDurations,
        accRewardPerShareScaled: 0,
        totalShares: 0
      })
    );
    emit PoolAdded(poolId, _startTime, _endTime);
    return poolId;
  }

  /// @notice Get the available lock durations of a pool with the corresponding multipliers
  /// @param poolId the id of the pool
  function getLockDurations(uint256 poolId) external view poolExists(poolId) returns (LockDuration[] memory) {
    return pools[poolId].lockDuration;
  }

  /// @notice Stops the pool, no more rewards will be distributed
  /// @param poolId the id of the pool
  function stopPool(uint256 poolId) external onlyOwner poolExists(poolId) {
    Pool storage pool = pools[poolId]; // errors if poolId is invalid
    _updatePoolReward(pool);
    pool.endTime = block.timestamp;
    emit PoolStopped(poolId);
    // TODO: withdraw all remaining rewards
  }

  /// @notice Manually update pool reward variables
  /// @param poolId the id of the pool
  function updatePoolReward(uint256 poolId) external poolExists(poolId) {
    //TODO: consider relalculatePoolReward, cause we're not updating the pool reward to some value
    Pool storage pool = pools[poolId]; // errors if poolId is invalid
    _updatePoolReward(pool);
  }

  /// @dev Update reward variables of the given pool to be up-to-date.
  function _updatePoolReward(Pool storage pool) private {
    uint256 timestamp = min(block.timestamp, pool.endTime);
    if (timestamp <= pool.lastRewardTime) {
      return;
    }

    if (pool.totalShares == 0) {
      pool.lastRewardTime = timestamp;
      return;
    }

    pool.accRewardPerShareScaled = getRewardPerShareScaled(pool, timestamp);
    pool.lastRewardTime = timestamp;
  }

  /// @dev calculate reward per share scaled without updating the pool
  function getRewardPerShareScaled(Pool memory pool, uint256 timestamp) private pure returns (uint256) {
    uint256 rewardScaled = (timestamp - pool.lastRewardTime) * pool.rewardPerSecond * PRECISION;
    return pool.accRewardPerShareScaled + (rewardScaled / pool.totalShares);
  }

  /// @notice Deposit staking token
  /// @param _poolId the id of the pool
  /// @param _amount the amount of staking token
  /// @param _lockDurationId the id for the predefined lock duration of the pool, earlier withdrawal is not possible
  /// @return stakeId the id of the new stake
  function stake(
    uint256 _poolId,
    uint256 _amount,
    uint8 _lockDurationId
  ) external poolExists(_poolId) returns (uint256) {
    Pool storage pool = pools[_poolId];
    if (block.timestamp < pool.startTime) {
      revert StakingNotStarted();
    }
    if (block.timestamp >= pool.endTime) {
      revert StakingFinished();
    }
    LockDuration memory lockDuration = pool.lockDuration[_lockDurationId];
    if (block.timestamp + lockDuration.durationSeconds > pool.endTime) {
      revert LockDurationExceedsStakingRange();
    }

    _updatePoolReward(pool);

    stakingToken.safeTransferFrom(address(msg.sender), address(this), _amount);

    uint256 userShares = (_amount * lockDuration.multiplierScaled) / PRECISION;
    pool.totalShares += userShares;

    uint256 stakeId = poolUserStakes[_poolId][msg.sender].length;
    poolUserStakes[_poolId][msg.sender].push(
      UserStake({
        amount: _amount,
        shareAmount: userShares,
        rewardDebt: (userShares * pool.accRewardPerShareScaled) / PRECISION,
        lockEndsAt: block.timestamp + lockDuration.durationSeconds
      })
    );

    emit Stake(msg.sender, _poolId, stakeId, _amount);
    return stakeId;
  }

  /// @notice Withdraw staking token and reward
  /// @param poolId the id of the pool
  /// @param stakeId the id of the stake
  function unstake(
    uint256 poolId,
    uint256 stakeId
  ) external poolExists(poolId) stakeExists(msg.sender, poolId, stakeId) {
    Pool storage pool = pools[poolId]; // errors if poolId is invalid
    UserStake storage userStake = poolUserStakes[poolId][msg.sender][stakeId];

    // lockEndsAt cannot be larger than pool.endTime if stopPool is not called
    // if stopPool is called, lockEndsAt is not checked
    if (block.timestamp < min(pool.endTime, userStake.lockEndsAt)) {
      revert LockDurationNotOver();
    }

    _updatePoolReward(pool);

    uint256 reward = (userStake.shareAmount * pool.accRewardPerShareScaled) / PRECISION - userStake.rewardDebt;

    uint256 stakeAmount = userStake.amount;
    pool.totalShares -= userStake.shareAmount;
    userStake.rewardDebt = (userStake.shareAmount * pool.accRewardPerShareScaled) / PRECISION;
    userStake.amount = 0;
    userStake.shareAmount = 0;

    safeTransfer(msg.sender, reward);
    stakingToken.safeTransfer(address(msg.sender), stakeAmount);
    emit Unstake(msg.sender, poolId, stakeId, stakeAmount);
  }

  /// @notice View function to see up-to-date reward of a user
  /// @param _user the user address
  /// @param poolId the id of the pool
  /// @param stakeId the id of the stake
  /// @return reward the reward amount of the reward token
  function getReward(
    address _user,
    uint256 poolId,
    uint256 stakeId
  ) external view poolExists(poolId) stakeExists(_user, poolId, stakeId) returns (uint256) {
    // we don't need to check pool.startTime because
    // staking is not allowed before startTime

    UserStake memory userStake = poolUserStakes[poolId][_user][stakeId];
    if (userStake.shareAmount == 0) {
      // early exit if user has no stake
      // also avoids division by zero if pool has no shares
      // cause the only way it can happen is when everybody
      // unstaked and user calls getReward again
      return 0;
    }

    Pool memory pool = pools[poolId];
    uint256 timestamp = min(block.timestamp, pool.endTime);
    return (userStake.shareAmount * getRewardPerShareScaled(pool, timestamp)) / PRECISION - userStake.rewardDebt;
  }

  /// @notice Withdraw reward token
  /// @param poolId the id of the pool
  /// @param stakeId the id of the stake
  function withdrawReward(
    uint256 poolId,
    uint256 stakeId
  ) external poolExists(poolId) stakeExists(msg.sender, poolId, stakeId) {
    Pool storage pool = pools[poolId];
    _updatePoolReward(pool);

    UserStake storage userStake = poolUserStakes[poolId][msg.sender][stakeId];
    uint256 rewardFromStart = (userStake.shareAmount * pool.accRewardPerShareScaled) / PRECISION;
    uint256 reward = rewardFromStart - userStake.rewardDebt;
    if (reward == 0) {
      revert NoRewardAvailable();
    }
    userStake.rewardDebt = rewardFromStart;

    emit RewardWithdrawal(msg.sender, poolId, stakeId, reward);
    safeTransfer(msg.sender, reward);
  }

  /// @dev Safe reward transfer function, just in case if rounding error causes pool to not have enough reward token.
  function safeTransfer(address _to, uint256 _amount) private {
    uint256 rewardBalance = rewardToken.allowance(fundingAccount, address(this));
    rewardToken.safeTransferFrom(fundingAccount, _to, min(rewardBalance, _amount));
  }

  function min(uint256 a, uint256 b) private pure returns (uint256) {
    return a < b ? a : b;
  }

  modifier onlyOwner() {
    if (msg.sender != owner) {
      revert Unauthorized();
    }
    _;
  }

  modifier poolExists(uint256 poolId) {
    if (poolId >= pools.length) {
      revert PoolNotFound();
    }
    _;
  }

  modifier stakeExists(
    address user,
    uint256 poolId,
    uint256 stakeId
  ) {
    // TODO: use only stakeExists modifier cause it imply existence of the pool
    // TODO: use hashing of mapping keys to check existence
    if (stakeId >= poolUserStakes[poolId][user].length) {
      revert StakeNotFound();
    }
    _;
  }
}
