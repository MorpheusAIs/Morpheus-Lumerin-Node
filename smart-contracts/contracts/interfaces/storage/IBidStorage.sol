// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IBidStorage {
  struct Bid {
    address provider;
    bytes32 modelAgentId;
    uint256 pricePerSecond; // hourly price
    uint256 nonce;
    uint128 createdAt;
    uint128 deletedAt;
  }

  function bidMap(bytes32 bidId) external view returns (Bid memory);

  function providerActiveBids(
    address provider_,
    uint256 offset_,
    uint256 limit_
  ) external view returns (bytes32[] memory);

  function modelAgentActiveBids(
    bytes32 modelAgentId_,
    uint256 offset_,
    uint256 limit_
  ) external view returns (bytes32[] memory);

  function providerBids(address provider_, uint256 offset_, uint256 limit_) external view returns (bytes32[] memory);

  function modelAgentBids(
    bytes32 modelAgentId_,
    uint256 offset_,
    uint256 limit_
  ) external view returns (bytes32[] memory);
}
