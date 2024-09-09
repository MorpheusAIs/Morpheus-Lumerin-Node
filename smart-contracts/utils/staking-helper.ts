import { PRECISION } from "../scripts/utils/constants";
import { DAY } from "./time";

export function getDefaultDurations() {
  return {
    durationSeconds: [7n * DAY, 30n * DAY, 180n * DAY, 365n * DAY],
    multiplierScaled: [
      1n * PRECISION,
      (115n * PRECISION) / 100n,
      (135n * PRECISION) / 100n,
      (150n * PRECISION) / 100n,
    ],
  };
}
