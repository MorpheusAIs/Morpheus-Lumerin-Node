import { expect } from "chai";
import { formatUnits } from "./units.ts";

const decimals = 18n;
const d = 10n ** decimals; // decimal multiplier
const dt = 10n ** (decimals - 3n); // decimal multiplier /1000 for convenience
const dm = 10n ** (decimals - 6n); // decimal multiplier /1000000 for convenience
const table = [
  [1000n * dt, "1.000"],
  [1500n * dt, "1.500"],
  [1999n * dt, "1.999"],
  [1000999n * dm, "1.001"],
  [1200123n * dt, "1 200"],
  [1200560n * dt, "1 201"],
  [1000000n * d, "1 000 000"],
  [0n, "0"],
  [1n * dm, "0"],
] as const;

describe("unit tests", () => {
  for (const [input, expected] of table) {
    it(`should format ${input} as ${expected}`, () => {
      expect(formatUnits(input, decimals)).to.equal(expected);
    });
  }
});
