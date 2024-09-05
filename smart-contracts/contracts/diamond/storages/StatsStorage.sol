// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { LibSD } from "../../libs/LibSD.sol";

contract StatsStorage {
  struct ProviderModelStats {
    LibSD.SD tpsScaled1000; // tokens per second running average
    LibSD.SD ttftMs; // time to first token running average in milliseconds
    uint32 totalDuration; // total duration of sessions
    uint32 successCount; // number of observations
    uint32 totalCount;
    // TODO: consider adding SD with weldford algorithm
  }

  struct ModelStats {
    LibSD.SD tpsScaled1000;
    LibSD.SD ttftMs;
    LibSD.SD totalDuration;
    uint32 count;
  }

  struct STTSStorage {
    mapping(bytes32 => mapping(address => ProviderModelStats)) stats; // modelId => provider => stats
    mapping(bytes32 => ModelStats) modelStats;
  }

  bytes32 public constant STATS_STORAGE_SLOT = keccak256("diamond.stats.storage");

  function _getModelStats(bytes32 modelAgentId) internal view returns (ModelStats storage) {
    return _getStatsStorage().modelStats[modelAgentId];
  }

  function _getProviderModelStats(
    bytes32 modelAgentId,
    address provider
  ) internal view returns (ProviderModelStats storage) {
    return _getStatsStorage().stats[modelAgentId][provider];
  }

  function _getStatsStorage() internal pure returns (STTSStorage storage _ds) {
    bytes32 slot_ = STATS_STORAGE_SLOT;

    assembly {
      _ds.slot := slot_
    }
  }
}
