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
  const start = new Date("2024-05-08T00:00:00Z").getTime() / 1000;

  const dailyReward: bigint[] = [];

  for (let i = 0; i < 100; i++) {
    const date = start + i * 6 * 60 * 60;
    const reward = await linear.read.getPeriodReward([
      data.initialReward,
      data.rewardDecrease,
      data.payoutStart,
      data.decreaseInterval,
      data.payoutStart,
      BigInt(date),
    ]);
    dailyReward.push(reward);
  }
  const content = dailyReward.join("\n");
  fs.writeFileSync("dailyReward.json", content);
}

main();
