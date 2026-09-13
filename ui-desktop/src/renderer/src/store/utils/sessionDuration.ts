/**
 * The contract's own floor is five minutes, but providers set a minimum session
 * cost, so a five-minute session is priced below what any of them will accept
 * and the request is refused on-chain. Fifteen minutes is the shortest length
 * that actually opens a session, which is why this floor is stricter than the
 * contract's and is not read from it.
 */
export const MIN_SESSION_DURATION_SECONDS = 15 * 60;
/**
 * Fallback ceiling, used only until the contract's own getMaxSessionDuration
 * has been read. Prefer the value from the chain: it is owner-settable and a
 * hardcoded ceiling silently offers lengths that getSessionEnd would clamp.
 */
export const MAX_SESSION_DURATION_SECONDS = 24 * 60 * 60;
export const DEFAULT_SESSION_DURATION_SECONDS = 60 * 60;

/**
 * One extra second of compute is paid for on top of the length requested.
 *
 * The contract floors twice on the way from an amount to a session end:
 * stakeToStipend floors, then getSessionEnd floors again dividing the stipend
 * by the price. An amount computed to land exactly on the requested length can
 * come out a second short. A second of headroom removes both truncations and
 * costs well under a tenth of a percent at any length offered here.
 */
export const SESSION_DURATION_HEADROOM_SECONDS = 1;

export const SESSION_DURATION_OPTIONS = [
  { seconds: 15 * 60, label: '15 minutes' },
  { seconds: 30 * 60, label: '30 minutes' },
  { seconds: 60 * 60, label: '1 hour' },
  { seconds: 3 * 60 * 60, label: '3 hours' },
  { seconds: 6 * 60 * 60, label: '6 hours' },
  { seconds: 12 * 60 * 60, label: '12 hours' },
  { seconds: 24 * 60 * 60, label: '24 hours' },
] as const;

export type SessionDurationOption = (typeof SESSION_DURATION_OPTIONS)[number];

type StakingInfo = { budget: number; supply: number };

/**
 * The lengths worth offering given the contract's ceiling. Anything longer than
 * the ceiling would be silently shortened by getSessionEnd, so the user would
 * pay for a duration they do not get. The shortest option always survives, so
 * the picker is never empty even if the ceiling is misreported.
 */
export function sessionDurationOptions(
  maxSeconds: number = MAX_SESSION_DURATION_SECONDS,
): readonly SessionDurationOption[] {
  const ceiling =
    Number.isFinite(maxSeconds) && maxSeconds > 0
      ? maxSeconds
      : MAX_SESSION_DURATION_SECONDS;
  const allowed = SESSION_DURATION_OPTIONS.filter(
    (option) => option.seconds <= ceiling,
  );
  return allowed.length > 0 ? allowed : [SESSION_DURATION_OPTIONS[0]];
}

export function clampSessionDuration(
  seconds: number,
  maxSeconds: number = MAX_SESSION_DURATION_SECONDS,
): number {
  const ceiling =
    Number.isFinite(maxSeconds) && maxSeconds >= MIN_SESSION_DURATION_SECONDS
      ? maxSeconds
      : MAX_SESSION_DURATION_SECONDS;
  if (!Number.isFinite(seconds)) {
    return Math.min(ceiling, DEFAULT_SESSION_DURATION_SECONDS);
  }
  return Math.min(
    ceiling,
    Math.max(MIN_SESSION_DURATION_SECONDS, Math.round(seconds)),
  );
}

/**
 * The MOR the diamond pulls from the wallet to open a session of this length.
 *
 * This is not the price of the compute. SessionRouter.getSessionEnd prices
 * every session as stakeToStipend(amount) / pricePerSecond, so the amount that
 * buys a given duration is the session cost converted through the emissions
 * ratio, currently a few hundred times larger than the cost itself. That is
 * true whether or not the session is direct pay: the payment method decides
 * only who the provider is paid by at close, not how much the session costs to
 * open. The desktop app used to compensate for this by inflating the duration
 * it sent for direct pay while the router applied no conversion at all; both
 * sides now use this one number.
 *
 * Mirrors computeSessionTokenAmount in proxy-router. Keep the two in step.
 */
export function estimateSessionTokenAmount(
  pricePerSecondWei: number,
  desiredSeconds: number,
  stakingInfo: StakingInfo,
): number {
  const budget = Number(stakingInfo.budget);
  const supply = Number(stakingInfo.supply);
  if (!(budget > 0) || !(supply > 0)) {
    throw new Error('Pricing data is not ready');
  }
  const paidSeconds =
    clampSessionDuration(desiredSeconds) + SESSION_DURATION_HEADROOM_SECONDS;
  const cost = pricePerSecondWei * paidSeconds;
  return Math.ceil((cost * supply) / budget);
}

/**
 * What the session is actually worth in MOR: price per second times the length.
 * Shown next to the amount above so the user can see that the difference is the
 * emissions conversion and not a fee.
 */
export function sessionComputeCost(
  pricePerSecondWei: number,
  desiredSeconds: number,
): number {
  return pricePerSecondWei * clampSessionDuration(desiredSeconds);
}
