// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {StakingMasterChef} from "../../staking/StakingMasterChef.sol";

contract StakingMasterChefV2 is StakingMasterChef {
    function version() external pure returns (uint256) {
        return 2;
    }
}
