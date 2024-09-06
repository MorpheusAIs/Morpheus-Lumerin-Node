// SPDX-License-Identifier: MIT
pragma solidity ^0.8.24;

import { LibSD } from "../../libs/LibSD.sol";

interface IStatsStorage {
  struct ProviderModelStats {
    LibSD.SD tpsScaled1000; // tokens per second running average
    LibSD.SD ttftMs; // time to first token running average in milliseconds
    uint32 totalDuration; // total duration of sessions
    uint32 successCount; // number of observations
    uint32 totalCount;
    // TODO: consider adding SD with weldford algorithm
  }
}
