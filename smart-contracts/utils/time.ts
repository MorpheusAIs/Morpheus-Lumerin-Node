export const MILLISECOND = 1;
export const SECOND = 1000 * MILLISECOND;
export const MINUTE = 60 * SECOND;
export const HOUR = 60 * MINUTE;
export const DAY = 24 * HOUR;

export function startOfNextDay(timestamp: bigint): bigint {
  return startOfDay(timestamp) + BigInt(DAY / SECOND);
}

export function startOfDay(timestamp: bigint): bigint {
  return timestamp - (timestamp % BigInt(DAY / SECOND));
}

export function now(): bigint {
  return BigInt(Math.floor(Date.now() / 1000));
}
