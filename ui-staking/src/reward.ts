export interface Pool {
  lastRewardTime: bigint;
  rewardPerSecondScaled: bigint;
  accRewardPerShareScaled: bigint;
  totalShares: bigint;
}

export interface UserStake {
  shareAmount: bigint;
  rewardDebt: bigint;
  lockEndsAt: bigint;
  stakeAmount: bigint;
}

export const getRewardPerShareScaled = (pool: Pool, timestamp: bigint): bigint => {
  const rewardScaled = (timestamp - pool.lastRewardTime) * pool.rewardPerSecondScaled;
  return pool.accRewardPerShareScaled + rewardScaled / pool.totalShares;
};

export const getReward = (
  userStake: UserStake,
  pool: Pool,
  timestamp: bigint,
  precision: bigint
): bigint => {
  const rewardPerShareScaled = getRewardPerShareScaled(pool, timestamp);
  return (userStake.shareAmount * rewardPerShareScaled) / precision - userStake.rewardDebt;
};
