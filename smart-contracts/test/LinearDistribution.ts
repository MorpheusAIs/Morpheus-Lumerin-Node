import hre from "hardhat";
import fs from "node:fs";

async function main() {
  const data = {
    payoutStart: 1707393600n,
    decreaseInterval: 86400n,
    initialReward: 3_456_000_000_000_000_000_000n,
    rewardDecrease: 592558728240000000n,
  };

  const linear = await hre.viem.deployContract("Linear", []);
  const now = BigInt(new Date().getTime()) / 1000n;
  const start = (now / 86400n + 1n) * 86400n;
  const end = start + 86400n;

  const reward = await linear.read.getPeriodReward([
    data.initialReward,
    data.rewardDecrease,
    data.payoutStart,
    data.decreaseInterval,
    start,
    end,
  ]);

  console.log("Reward: ", reward.toString());
}

main();
