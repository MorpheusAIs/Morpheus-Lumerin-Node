import { expect } from "chai";

export type NumberLike = number | bigint;

export const expectAlmostEqual = (
  expected: NumberLike,
  actual: NumberLike,
  epsilon: NumberLike,
  message?: string
): void => {
  const delta = Number(epsilon) * Number(expected);
  const min = Number(expected) - delta;
  const max = Number(expected) + delta;

  let msg = `expected ${actual} to be within ${expected} +/- ${epsilon} epsilon (${min} - ${max})`;
  if (message) {
    msg += `: ${message}`;
  }
  expect(AlmostEqual(expected, actual, epsilon)).to.be.eq(true, msg);
};

export const AlmostEqual = (
  expected: NumberLike,
  actual: NumberLike,
  epsilon: NumberLike
): boolean => {
  return RelativeError(expected, actual) <= epsilon;
};

export const RelativeError = (target: NumberLike, actual: NumberLike): number => {
  return Math.abs(Number(actual) - Number(target)) / Math.abs(Number(target));
}
