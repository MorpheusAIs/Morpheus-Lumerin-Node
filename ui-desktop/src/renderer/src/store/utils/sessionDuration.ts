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

/**
 * A further 0.1% on top of the computed amount, to absorb drift between
 * quoting and mining.
 *
 * The amount needed scales with totalMORSupply, which the contract evaluates at
 * block.timestamp while the router quotes from a cached read taken up to a
 * minute earlier. Supply grows about 0.29% a day, so a quote that sits in the
 * mempool for a few minutes buys very slightly less time than it was priced
 * for — and on a 24 hour session that shortfall already exceeds the one second
 * of headroom above. Falling short means the open reverts with SessionTooShort
 * and the gas is spent for nothing.
 *
 * 0.1% covers roughly eight hours of drift. It is not a fee: staked MOR is
 * refunded in full at close, and on the direct-pay path the unused remainder is
 * returned, so the padding comes back either way.
 *
 * Mirrors sessionAmountSafetyBps in proxy-router. Keep the two in step.
 */
export const SESSION_AMOUNT_SAFETY_BPS = 10n;

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

type StakingInfo = { budget: string | number; supply: string | number };

/**
 * Wei-denominated values arrive from the proxy-router as JSON, which means a
 * decimal string for anything the Go side holds in a big.Int and a JS number
 * for the few that were already narrowed. Both are accepted; the string form is
 * the one that survives a round trip intact, so prefer passing it through.
 *
 * Throws rather than returning a sentinel: every caller is about to quote a
 * price, and a silently-zero price quotes a free session.
 */
function toWei(value: string | number, label: string): bigint {
  if (typeof value === 'string') {
    const trimmed = value.trim();
    if (!/^\d+$/.test(trimmed)) {
      throw new Error(`${label} is not a whole number of wei`);
    }
    return BigInt(trimmed);
  }
  if (!Number.isFinite(value) || value < 0) {
    throw new Error(`${label} is not a usable amount`);
  }
  // A number this large has already lost its low digits before reaching us.
  // Rounding here only makes that explicit; it does not add error.
  return BigInt(Math.round(value));
}

/** Integer ceiling division. `/` on bigint truncates toward zero. */
const divCeil = (numerator: bigint, denominator: bigint): bigint =>
  (numerator + denominator - 1n) / denominator;

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
 * `maxSeconds` must be the same contract ceiling the picker and the open path
 * clamp against. Leaving it off quotes the unclamped length while the session
 * opened is the clamped one, so the user is shown a price for time they do not
 * get.
 *
 * The arithmetic runs in BigInt. These are 18-decimal amounts multiplied by a
 * supply on the order of 1e25, so every intermediate is far past 2^53 and the
 * rounding-up this function promises would otherwise be decorative.
 *
 * Mirrors computeSessionTokenAmount in proxy-router. Keep the two in step,
 * including SESSION_AMOUNT_SAFETY_BPS.
 */
export function estimateSessionTokenAmount(
  pricePerSecondWei: string | number,
  desiredSeconds: number,
  stakingInfo: StakingInfo,
  maxSeconds?: number,
): number {
  const budget = toWei(stakingInfo.budget, "Today's budget");
  const supply = toWei(stakingInfo.supply, 'Token supply');
  if (budget <= 0n || supply <= 0n) {
    throw new Error('Pricing data is not ready');
  }
  const price = toWei(pricePerSecondWei, 'Price per second');
  if (price <= 0n) {
    throw new Error('Price per second is not ready');
  }
  const paidSeconds = BigInt(
    clampSessionDuration(desiredSeconds, maxSeconds) +
      SESSION_DURATION_HEADROOM_SECONDS,
  );
  const amount = divCeil(price * paidSeconds * supply, budget);
  return Number(
    divCeil(amount * (10_000n + SESSION_AMOUNT_SAFETY_BPS), 10_000n),
  );
}

/**
 * What the session is actually worth in MOR: price per second times the length.
 * Shown next to the amount above so the user can see that the difference is the
 * emissions conversion and not a fee.
 *
 * Takes the same `maxSeconds` as the estimate above, for the same reason: these
 * two numbers are displayed side by side and must describe one session.
 */
export function sessionComputeCost(
  pricePerSecondWei: string | number,
  desiredSeconds: number,
  maxSeconds?: number,
): number {
  const price = toWei(pricePerSecondWei, 'Price per second');
  return Number(
    price * BigInt(clampSessionDuration(desiredSeconds, maxSeconds)),
  );
}
