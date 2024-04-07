import { expect } from "chai";

type NumberLike = number | bigint;

export const expectAlmostEqual = (expected: NumberLike, actual: NumberLike, epsilon: NumberLike, message?: string): void => {
  let msg = `expected ${actual} to be within ${expected} +/- ${epsilon} epsilon`;
  if (message){
    msg += `: ${message}`;
  }
  expect(AlmostEqual(expected, actual, epsilon)).to.be.eq(true, msg);
}

export const AlmostEqual = (expected: NumberLike, actual: NumberLike, epsilon: NumberLike): boolean => {
  return RelativeError(expected, actual) <= epsilon;
}

export const RelativeError = (target: NumberLike, actual: NumberLike): number => {
  return Math.abs(Number(actual) - Number(target)) / Math.abs(Number(target));
}
