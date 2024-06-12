import { expect } from "chai";

type NumberLike = number | bigint;

export const expectAlmostEqual = (
  expected: NumberLike,
  actual: NumberLike,
  epsilon: NumberLike,
  message?: string,
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
  epsilon: NumberLike,
): boolean => {
  return RelativeError(expected, actual) <= epsilon;
};

export const RelativeError = (
  target: NumberLike,
  actual: NumberLike,
): number => {
  return Math.abs(Number(actual) - Number(target)) / Math.abs(Number(target));
};

export const expectAlmostEqualDelta = (
  expected: NumberLike,
  actual: NumberLike,
  delta: NumberLike,
  message?: string,
): void => {
  const min = Number(expected) - Number(delta);
  const max = Number(expected) + Number(delta);

  let msg = `expected ${actual} to be within ${expected} +/- ${delta} (${min} - ${max})`;
  if (message) {
    msg += `: ${message}`;
  }
  expect(AlmostEqualDelta(expected, actual, delta)).to.be.eq(true, msg);
};

export const AlmostEqualDelta = (
  expected: NumberLike,
  actual: NumberLike,
  delta: NumberLike,
): boolean => {
  return AbsoluteError(expected, actual) <= delta;
};

export const AbsoluteError = (
  target: NumberLike,
  actual: NumberLike,
): number => {
  return Math.abs(Number(actual) - Number(target));
};
