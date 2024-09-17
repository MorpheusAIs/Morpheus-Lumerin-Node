// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {LinearDistributionIntervalDecrease} from "morpheus-smart-contracts/contracts/libs/LinearDistributionIntervalDecrease.sol";

import {ERC1967Proxy} from "@openzeppelin/contracts/proxy/ERC1967/ERC1967Proxy.sol";

contract LinearDistributionIntervalDecreaseMock {
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
