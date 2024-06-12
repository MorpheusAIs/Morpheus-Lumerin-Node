// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.0;

import { LinearDistributionIntervalDecrease } from "./libraries/LinearDistributionIntervalDecrease.sol";

contract Linear {
  uint256 public constant initialAmount = 1000;

  constructor() {}

  function getPeriodReward(
    uint256 initialAmount_,
    uint256 decreaseAmount_,
    uint128 payoutStart_,
    uint128 interval_,
    uint128 startTime_,
    uint128 endTime_
  ) public pure returns (uint256) {
    return
      LinearDistributionIntervalDecrease.getPeriodReward(
        initialAmount_,
        decreaseAmount_,
        payoutStart_,
        interval_,
        startTime_,
        endTime_
      );
  }
}
