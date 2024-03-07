/**
 * Performs rounding of the time to the nearest granularityMs
 * @param {Date} input
 * @param {number} granularityMs
 * @returns {Date} result
 */
export function roundTime(input, granularityMs) {
  // used to avoid rounding errors due to possible extra seconds in the leap years
  const relativeBase = getStartOfTheDay(
    new Date(input.getTime() - granularityMs)
  );
  const relativeTime = input.getTime() - relativeBase.getTime();
  const remainder = relativeTime % granularityMs;
  if (remainder / granularityMs < 0.5) {
    return new Date(input.getTime() - remainder);
  }
  return new Date(input.getTime() - remainder + granularityMs);
}

/**
 * @param {Date} date
 * @returns {Date}
 */
export function getStartOfTheDay(date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}
