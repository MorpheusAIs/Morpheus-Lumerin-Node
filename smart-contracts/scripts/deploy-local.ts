import * as fixtures from "../test/fixtures";
import * as fixturesStaking from "../test/Staking/fixtures";
import { getStakeId } from "../test/Staking/utils";
import { DAY, HOUR, MINUTE, SECOND } from "../utils/time";
import hre from "hardhat";

async function main() {
  const data = await fixtures.deploySingleBid();
  const lmr = await fixturesStaking.deployLMR();
  const { staking, precision } = await fixturesStaking.deployStaking(
    lmr.address,
    data.tokenMOR.address,
  );

  const block = await data.publicClient.getBlock();
  const startDate = block.timestamp;
  const duration = 48n * BigInt(HOUR / SECOND);
  const rewardPerSecond = (115n * 10n ** 18n) / 1_000_000n;
  const totalReward = rewardPerSecond * duration;

  await data.tokenMOR.write.approve([staking.address, totalReward]);
  await fixturesStaking.setupPools(staking.address, [
    {
      durationSeconds: duration,
      startDate,
      totalReward: totalReward,
      lockDurations: fixturesStaking.getDefaultDurationsMedium(precision),
    },
  ]);

  const stakingAmount = 10n * 10n ** 8n;
  const lockDurationId = 0;
  const poolId = 0n;
  const [_, alice, bob] = await hre.viem.getWalletClients();

  await lmr.write.transfer([alice.account.address, stakingAmount * 100n]);
  await lmr.write.transfer([bob.account.address, stakingAmount * 100n]);

  /**
  for (let i = 0; i < 3; i++) {
    await lmr.write.approve([staking.address, stakingAmount], {
      account: alice.account,
    });
    await staking.write.stake([poolId, stakingAmount, lockDurationId], {
      account: alice.account,
    });
  }

  for (let i = 0; i < 3; i++) {
    await lmr.write.approve([staking.address, stakingAmount], {
      account: bob.account,
    });
    await staking.write.stake([poolId, stakingAmount, lockDurationId], {
      account: bob.account,
    });
  }
  */

  // const stakeId = await getStakeId(depositTx);

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
