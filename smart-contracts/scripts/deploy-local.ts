import * as fixtures from "../test/fixtures";
import * as fixturesStaking from "../test/Staking/fixtures";
import { DAY, SECOND } from "../utils/time";

async function main() {
  const data = await fixtures.deploySingleBid();
  const lmr = await fixturesStaking.deployLMR();
  const { staking, precision } = await fixturesStaking.deployStaking(
    data.tokenMOR.address,
    lmr.address,
  );

  const startDate =
    BigInt(new Date("2024-07-16T01:00:00.000Z").getTime()) / 1000n;
  const duration = 400n * BigInt(DAY / SECOND);
  const rewardPerSecond = 100n;
  const totalReward = rewardPerSecond * duration;

  await lmr.write.approve([staking.address, totalReward]);
  await fixturesStaking.setupPools(staking.address, [
    {
      durationSeconds: duration,
      startDate,
      totalReward: totalReward,
      lockDurations: fixturesStaking.getDefaultDurations(precision),
    },
  ]);

  console.log(`
    MOR token       ${data.tokenMOR.address}
    LMR token       ${lmr.address}
    Diamond         ${data.marketplace.address}
    Staking         ${staking.address}


    Owner:          ${data.owner.account.address}
    Provider:       ${data.provider.account.address}
    User:           ${data.user.account.address}
  `);
}

main();
