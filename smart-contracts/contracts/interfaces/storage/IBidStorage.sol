// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IERC20} from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

interface IBidStorage {
    struct Bid {
        address provider;
        bytes32 modelId;
        uint256 pricePerSecond; // hourly price
        uint256 nonce;
        uint128 createdAt;
        uint128 deletedAt;
    }

    function bids(bytes32 bidId) external view returns (Bid memory);

    function providerActiveBids(
        address provider_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory);

    function modelActiveBids(
        bytes32 modelId_,
        uint256 offset_,
        uint256 limit_
    ) external view returns (bytes32[] memory);

    function providerBids(address provider_, uint256 offset_, uint256 limit_) external view returns (bytes32[] memory);

    function modelBids(bytes32 modelId_, uint256 offset_, uint256 limit_) external view returns (bytes32[] memory);

    function getToken() external view returns (IERC20);
}
