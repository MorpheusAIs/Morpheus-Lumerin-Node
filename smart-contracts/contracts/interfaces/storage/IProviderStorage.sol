// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

interface IProviderStorage {
  struct Provider {
    string endpoint; // example 'domain.com:1234'
    uint256 stake; // stake amount, which also server as a reward limiter
    uint128 createdAt; // timestamp of the registration
    uint128 limitPeriodEnd; // timestamp of the limiter period end
    uint256 limitPeriodEarned; // total earned during the last limiter period
    bool isDeleted;
  }

  function getProvider(address provider) external view returns (Provider memory);

  function providers(uint256 index) external view returns (address);

  function providerMinimumStake() external view returns (uint256);
}
