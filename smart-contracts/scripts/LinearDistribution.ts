import hre from "hardhat";

async function main() {
  const data = {
    payoutStart: 19870n * 24n * 3600n,
    decreaseInterval: 86400n,
    initialReward: 100_000_000n,
    rewardDecrease: 1_000_000n,
  };
  const linear = await hre.viem.deployContract(
    "LinearDistributionIntervalDecreaseMock",
    [],
  );

  for (let i = 0; i < 1000; i++) {
    const now = data.payoutStart + BigInt(i) * 24n * 3600n;
    const start = 0n;
    const end = now;

    const reward = await linear.read.getPeriodReward([
      data.initialReward,
      data.rewardDecrease,
      data.payoutStart,
      data.decreaseInterval,
      start,
      end,
    ]);

    const rewardSm = await linear.read.getPeriodReward([
      data.initialReward / 6n,
      data.rewardDecrease / 6n,
      data.payoutStart,
      data.decreaseInterval,
      start,
      end,
    ]);

    const total = 4n * reward + rewardSm;

    const ratio = Number(reward) / Number(total);

    console.log("Reward: ", reward.toString());
    console.log("RewardSm: ", rewardSm.toString());
    console.log("Ratio: ", ratio);
  }
}

main();
