// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface ISessionStorage {
    struct Session {
        bytes32 id;
        address user;
        address provider;
        bytes32 modelId;
        bytes32 bidID;
        uint256 stake;
        uint256 pricePerSecond;
        bytes closeoutReceipt;
        uint256 closeoutType;
        // amount of funds that was already withdrawn by provider (we allow to withdraw for the previous day)
        uint256 providerWithdrawnAmount;
        uint256 openedAt;
        uint256 endsAt; // expected end time considering the stake provided
        uint256 closedAt;
    }

    struct OnHold {
        uint256 amount;
        uint128 releaseAt; // in epoch seconds TODO: consider using hours to reduce storage cost
    }

    struct Pool {
        uint256 initialReward;
        uint256 rewardDecrease;
        uint128 payoutStart;
        uint128 decreaseInterval;
    }

    function sessions(bytes32 sessionId) external view returns (Session memory);

    function getSessionsByUser(address user, uint256 offset_, uint256 limit_) external view returns (bytes32[] memory);

    function getFundingAccount() external view returns (address);

    function pools() external view returns (Pool[] memory);
}
