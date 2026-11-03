// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface ISessionSettlement {
    event SessionRating(
        bytes32 indexed sessionId,
        bytes32 indexed modelId,
        address indexed provider,
        uint256 closeoutType,
        uint32 tpsScaled1000,
        uint32 ttftMs
    );
}
