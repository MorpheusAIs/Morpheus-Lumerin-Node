import hre from "hardhat";
import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { DAY, SECOND } from "../../utils/time";
import { getStakeId } from "./utils";

export async function setupStaking() {
  const [owner, alice, bob, carol] = await hre.viem.getWalletClients();

  const tokenMOR = await hre.viem.deployContract("MorpheusToken", []);
  const tokenLMR = await hre.viem.deployContract("LumerinToken", []);

  const startDate =
    BigInt(new Date("2024-07-16T01:00:00.000Z").getTime()) / 1000n;
  const endDate = startDate + BigInt((400 * DAY) / SECOND);

  const expPool = {
    id: 0n,
    rewardPerSecond: 100n,
    stakingToken: tokenLMR,
    rewardToken: tokenMOR,
    totalReward: 1_000_000_000_000n,
    lockDurations: <{ durationSeconds: bigint; multiplierScaled: bigint }[]>[],
    precision: 0n,
    startDate,
    endDate,
  };

  const staking = await hre.viem.deployContract("StakingMasterChef", [
    tokenLMR.address,
    tokenMOR.address,
    owner.account.address,
  ]);

  const PRECISION = await staking.read.PRECISION();

  expPool.precision = PRECISION;
  expPool.lockDurations = [
    {
      durationSeconds: BigInt((7 * DAY) / SECOND),
      multiplierScaled: 1n * PRECISION,
    },
    {
      durationSeconds: BigInt((30 * DAY) / SECOND),
      multiplierScaled: (115n * PRECISION) / 100n,
    },
    {
      durationSeconds: BigInt((180 * DAY) / SECOND),
      multiplierScaled: (135n * PRECISION) / 100n,
    },
    {
      durationSeconds: BigInt((365 * DAY) / SECOND),
      multiplierScaled: (150n * PRECISION) / 100n,
    },
  ];

  await staking.write.addPool([
    expPool.rewardPerSecond,
    expPool.startDate,
    expPool.endDate,
    expPool.lockDurations,
  ]);

  expPool.id = 0n;

  // approve funds for staking
  await tokenMOR.write.approve([staking.address, expPool.totalReward]);

  // top up accounts
  await tokenLMR.write.transfer([alice.account.address, 1_000_000n]);
  await tokenLMR.write.transfer([bob.account.address, 1_000_000n]);
  await tokenLMR.write.transfer([carol.account.address, 1_000_000n]);

  return {
    accounts: { owner, alice, bob, carol },
    contracts: { staking, tokenMOR, tokenLMR },
    expPool,
    pubClient: await hre.viem.getPublicClient(),
  };
}

export async function aliceStakes() {
  const data = await loadFixture(setupStaking);
  const {
    contracts: { staking, tokenLMR },
    accounts: { alice },
  } = data;

  const stakingAmount = 1000n;
  const lockDurationId = 0;
  const poolId = 0n;

  await tokenLMR.write.approve([staking.address, stakingAmount], {
    account: alice.account,
  });
  const depositTx = await staking.write.stake(
    [poolId, stakingAmount, lockDurationId],
    {
      account: alice.account,
    },
  );

  const stakeId = await getStakeId(depositTx);

  return {
    ...data,
    stakes: {
      alice: { depositTx, stakingAmount, stakeId, lockDurationId, poolId },
    },
  };
}

export async function aliceAndBobStake() {
  const data = await loadFixture(aliceStakes);
  const {
    accounts,
    contracts: { staking, tokenLMR },
    stakes: {
      alice: { poolId, lockDurationId },
    },
  } = data;

  await time.increase(5 * (DAY / SECOND));

  const stakingAmount = 1000n;
  await tokenLMR.write.approve([staking.address, stakingAmount], {
    account: accounts.bob.account,
  });
  const bobDepositTx = await staking.write.stake(
    [poolId, stakingAmount, lockDurationId],
    { account: accounts.bob.account },
  );
  const bobStakeId = await getStakeId(bobDepositTx);

  return {
    ...data,
    stakes: {
      ...data.stakes,
      bob: {
        depositTx: bobDepositTx,
        stakingAmount,
        stakeId: bobStakeId,
        poolId,
        lockDurationId,
      },
    },
  };
}
