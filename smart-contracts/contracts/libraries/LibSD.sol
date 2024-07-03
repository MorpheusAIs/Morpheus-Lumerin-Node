// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.24;

// inspired by https://github.com/rob-Hitchens/UnorderedKeySet/blob/master/contracts/HitchensUnorderedKeySet.sol
library LibSD {
  struct SD {
    int64 mean;
    int64 sqSum;
  }

  function add(SD storage self, int64 x, int64 updatedCount) internal {
    int64 oldMean = self.mean;
    self.mean = self.mean + (x - self.mean) / updatedCount;
    self.sqSum = self.sqSum + (x - self.mean) * (x - oldMean);
  }

  function remove(SD storage self, int64 x, int64 updatedCount) internal {
    if (updatedCount == 0) {
      self.mean = 0;
      self.sqSum = 0;
    } else {
      int64 oldMean = self.mean;
      self.mean = self.mean - (x - self.mean) / updatedCount;
      self.sqSum = self.sqSum - (x - self.mean) * (x - oldMean);
    }
  }

  function variance(SD storage self, int64 count) internal view returns (int64) {
    return self.sqSum / (count - 1);
  }
}