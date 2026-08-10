export const MIN_SESSION_DURATION_SECONDS = 5 * 60;
export const MAX_SESSION_DURATION_SECONDS = 24 * 60 * 60;
export const DEFAULT_SESSION_DURATION_SECONDS = 60 * 60;

export const SESSION_DURATION_OPTIONS = [
  { seconds: 5 * 60, label: '5 minutes' },
  { seconds: 15 * 60, label: '15 minutes' },
  { seconds: 30 * 60, label: '30 minutes' },
  { seconds: 60 * 60, label: '1 hour' },
  { seconds: 3 * 60 * 60, label: '3 hours' },
  { seconds: 6 * 60 * 60, label: '6 hours' },
  { seconds: 12 * 60 * 60, label: '12 hours' },
  { seconds: 24 * 60 * 60, label: '24 hours' },
] as const;

type StakingInfo = { budget: number; supply: number };

export function clampSessionDuration(seconds: number): number {
  if (!Number.isFinite(seconds)) {
    return DEFAULT_SESSION_DURATION_SECONDS;
  }
  return Math.min(
    MAX_SESSION_DURATION_SECONDS,
    Math.max(MIN_SESSION_DURATION_SECONDS, Math.round(seconds)),
  );
}

/**
 * The current contract calculates direct-pay session length through the staking
 * conversion as well. Compensate for that contract behaviour so the duration
 * selected in the UI is the duration the user receives on-chain.
 */
export function toSessionRequestDuration(
  desiredSeconds: number,
  directPayment: boolean,
  stakingInfo: StakingInfo,
): number {
  const desired = clampSessionDuration(desiredSeconds);
  if (!directPayment) {
    return desired;
  }
  const budget = Number(stakingInfo.budget);
  const supply = Number(stakingInfo.supply);
  if (!(budget > 0) || !(supply > 0)) {
    throw new Error('Pricing data is not ready');
  }
  return Math.ceil((desired * supply) / budget) + 1;
}

export function estimateSessionTokenAmount(
  pricePerSecondWei: number,
  desiredSeconds: number,
  directPayment: boolean,
  stakingInfo: StakingInfo,
): number {
  const requestDuration = toSessionRequestDuration(
    desiredSeconds,
    directPayment,
    stakingInfo,
  );
  const cost = pricePerSecondWei * requestDuration;
  if (directPayment) {
    return cost;
  }
  return (cost * Number(stakingInfo.supply)) / Number(stakingInfo.budget);
}
