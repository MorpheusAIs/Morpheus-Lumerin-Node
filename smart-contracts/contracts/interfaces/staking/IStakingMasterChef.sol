// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IStakingMasterChef {
    /**
     * The structure that represents a pool in the staking contract.
     * @param rewardPerSecondScaled reward tokens per second, times `PRECISION`
     * @param startTime start time of the staking for this pool
     * @param endTime end time of the staking for this pool - after this time, no more rewards will be distributed
     * @param isTerminatedForced whether the pool is stopped by owner
     */
    struct Pool {
        uint256 rewardPerSecondScaled;
        uint256 startTime;
        uint256 endTime;
        bool isTerminatedForced;
    }

    /**
     * The structure that represents a pool rate data in the staking contract.
     * @param lastRewardTime last time rewards were distributed
     * @param accRewardPerShareScaled accumulated reward per share, times `PRECISION`
     * @param totalShares total shares of reward token
     */
    struct PoolRateData {
        uint256 lastRewardTime;
        uint256 accRewardPerShareScaled;
        uint256 totalShares;
    }

    /**
     * The structure that represents a user stake in the staking contract.
     * @param stakeAmount amount of staked tokens
     * @param shareAmount shares received after staking
     * @param rewardDebt reward debt
     * @param lockEnd when staking lock duration ends
     */
    struct UserStake {
        uint256 stakeAmount;
        uint256 shareAmount;
        uint256 rewardDebt;
        uint256 lockEnd;
    }

    /**
     * @dev Emitted when a user stakes tokens.
     * @param user The address of the user who staked the tokens.
     * @param poolId The ID of the pool in which the user staked the tokens.
     * @param stakeId The ID of the stake.
     * @param amount The amount of tokens staked.
     */
    event Staked(address indexed user, uint256 indexed poolId, uint256 stakeId, uint256 amount);

    /**
     * @dev Emitted when a user unstakes tokens.
     * @param user The address of the user who unstaked the tokens.
     * @param poolId The ID of the pool in which the user unstaked the tokens.
     * @param stakeId The ID of the stake.
     * @param amount The amount of tokens unstaked.
     */
    event Unstaked(address indexed user, uint256 indexed poolId, uint256 stakeId, uint256 amount);

    /**
     * @dev Emitted when a user withdraws rewards.
     * @param userAddress The address of the user who withdrew the rewards.
     * @param poolId The ID of the pool in which the user withdrew the rewards.
     * @param stakeId The ID of the stake.
     * @param amount The amount of rewards withdrawn.
     */
    event RewardWithdrawed(address indexed userAddress, uint256 indexed poolId, uint256 stakeId, uint256 amount);

    /**
     * @dev Emitted when a new pool is added.
     * @param poolId The ID of the pool.
     * @param startTime The start time of the pool.
     * @param endTime The end time of the pool.
     */
    event PoolAdded(uint256 indexed poolId, uint256 startTime, uint256 endTime);

    /**
     * @dev Emitted when a pool is terminated.
     * @param poolId The ID of the pool.
     */
    event PoolTerminated(uint256 indexed poolId);

    error PoolNotExists();
    error StakeNotExists();
    error StakeUnstaked();
    error StakingNotStarted();
    error StakingFinished();
    error LockNotEnded();
    error LockReleaseTimePastPoolEndTime();
    error NoRewardAvailable();
    error PoolAlreadyTerminated();
    error InvalidDuration();
    error InvalidReward();
    error StartTimeIsPast();
    error InvalidLockId();
    error InvalidLocksCount();
    error InvalidLockDuration();
    error InvalidAmount();

    /**
     * @notice Adds a new pool, requires approval for the reward token
     * @dev if lock duration is longer than pool duration it is accepted, but it won't be possible to stake with that lock duration
     * @param startTime_ time when the staking starts, epoch seconds
     * @param duration_ pool duration in seconds
     * @param totalReward_ total reward for the pool
     * @param lockDurations_ predefined lock durations for this pool with corresponding multipliers
     * @param multipliersScaled_ multipliers for each lock duration, times `PRECISION`
     * @return The id of the new pool
     */
    function addPool(
        uint256 startTime_,
        uint256 duration_,
        uint256 totalReward_,
        uint256[] calldata lockDurations_,
        uint256[] calldata multipliersScaled_
    ) external returns (uint256);

    /**
     * @notice Terminate the pool, no more rewards will be distributed
     * @param poolId_ the id of the pool
     * @param recipient_ the address to receive the undistributed reward
     */
    function terminatePool(uint256 poolId_, address recipient_) external;

    /**
     * @notice Stakes tokens for a user
     * @param poolId_ the id of the pool
     * @param amount_ the amount of tokens to stake
     * @param lockDuration_ the duration of the lock
     * @return The id of the new stake
     */
    function stake(uint256 poolId_, uint256 amount_, uint256 lockDuration_) external returns (uint256);

    /**
     * @notice Unstakes tokens for a user
     * @param poolId_ the id of the pool
     * @param stakeId_ the id of the stake
     * @param recipient_ the address to receive the reward tokens
     */
    function unstake(uint256 poolId_, uint256 stakeId_, address recipient_) external;

    /**
     * @notice Withdraws rewards for a user
     * @param poolId_ the id of the pool
     * @param stakeId_ the id of the stake
     * @param recipient_ the address to receive the rewards
     */
    function withdrawReward(uint256 poolId_, uint256 stakeId_, address recipient_) external;

    /**
     * @notice Manually update pool reward variables
     * @param poolId_ the id of the pool
     */
    function recalculatePoolReward(uint256 poolId_) external;

    /**
     * @notice View function to see up-to-date reward of a user
     * @param user_ the user address
     * @param poolId_ the id of the pool
     * @param stakeId_ the id of the stake
     * @return The reward amount of the reward token
     */
    function getReward(address user_, uint256 poolId_, uint256 stakeId_) external view returns (uint256);
}
