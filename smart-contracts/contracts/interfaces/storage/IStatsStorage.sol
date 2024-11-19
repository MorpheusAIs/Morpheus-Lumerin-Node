// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import {LibSD} from "../../libs/LibSD.sol";

interface IStatsStorage {
    struct ModelStats {
        LibSD.SD tpsScaled1000;
        LibSD.SD ttftMs;
        LibSD.SD totalDuration;
        uint32 count;
    }

    /**
     * The structure that stores the provider model stats.
     * @param tpsScaled1000 Tokens per second running average
     * @param ttftMs Time to first token running average in milliseconds
     * @param totalDuration Total duration of sessions
     * @param successCount Number of observations
     * @param totalCount
     */
    struct ProviderModelStats {
        LibSD.SD tpsScaled1000;
        LibSD.SD ttftMs;
        uint32 totalDuration;
        uint32 successCount;
        uint32 totalCount;
        // TODO: consider adding SD with weldford algorithm
    }

    /**
     * @param modelId_ The model ID.
     * @param provider_ The provider address.
     */
    function getProviderModelStats(
        bytes32 modelId_,
        address provider_
    ) external view returns (ProviderModelStats memory);

    /**
     * @param modelId_ The model ID.
     */
    function getModelStats(bytes32 modelId_) external view returns (ModelStats memory);
}
