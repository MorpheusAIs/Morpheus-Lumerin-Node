// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {IStatsStorage} from "../../interfaces/storage/IStatsStorage.sol";

import {LibSD} from "../../libs/LibSD.sol";

contract StatsStorage is IStatsStorage {
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

    function _getModelStats(bytes32 modelId) internal view returns (ModelStats storage) {
        return _getStatsStorage().modelStats[modelId];
    }

    function _getProviderModelStats(
        bytes32 modelId,
        address provider
    ) internal view returns (ProviderModelStats storage) {
        return _getStatsStorage().stats[modelId][provider];
    }

    function _getStatsStorage() internal pure returns (STTSStorage storage _ds) {
        bytes32 slot_ = STATS_STORAGE_SLOT;

        assembly {
            _ds.slot := slot_
        }
    }
}
