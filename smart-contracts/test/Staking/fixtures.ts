import hre from "hardhat";
import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { DAY, MINUTE, SECOND } from "../../utils/time";
import { getStakeId } from "./utils";

export async function setupStaking() {
  const [owner, alice, bob, carol] = await hre.viem.getWalletClients();

  const tokenMOR = await hre.viem.deployContract("MorpheusToken", []);
  const tokenLMR = await deployLMR();

  const startDate =
    BigInt(new Date("2024-07-16T01:00:00.000Z").getTime()) / 1000n;
  const duration = 400n * BigInt(DAY / SECOND);
  const endDate = startDate + duration;
  const rewardPerSecond = 100n;

  const expPool = {
    id: 0n,
    rewardPerSecond,
    stakingToken: tokenLMR,
    rewardToken: tokenMOR,
    totalReward: rewardPerSecond * duration,
    lockDurations: <{ durationSeconds: bigint; multiplierScaled: bigint }[]>[],
    precision: 0n,
    startDate,
    endDate,
    duration,
  };

  const staking = await hre.viem.deployContract("StakingMasterChef", [
    tokenLMR.address,
    tokenMOR.address,
  ]);

  const PRECISION = await staking.read.PRECISION();

  expPool.precision = PRECISION;
  expPool.lockDurations = getDefaultDurations(PRECISION);

  await tokenMOR.write.approve([staking.address, expPool.totalReward]);

  // create a pool
  await staking.write.addPool([
    expPool.startDate,
    expPool.duration,
    expPool.totalReward,
    expPool.lockDurations,
  ]);

  expPool.id = 0n;

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
    { account: alice.account },
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

export async function deployLMR() {
  console.log("Deploying LMR token...");
  const tx = await hre.viem.sendDeploymentTransaction("LumerinToken", []);
  console.log("LMR token deployed to address:", tx.contract.address);
  console.log("Transaction hash:", tx.deploymentTransaction.hash);
  return tx.contract;
}

export function getDefaultDurations(PRECISION: bigint) {
  return [
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
}

export function getDefaultDurationsShort(PRECISION: bigint) {
  return [
    {
      durationSeconds: BigInt((30 * SECOND) / SECOND),
      multiplierScaled: 1n * PRECISION,
    },
    {
      durationSeconds: BigInt((2 * MINUTE) / SECOND),
      multiplierScaled: (115n * PRECISION) / 100n,
    },
    {
      durationSeconds: BigInt((3 * MINUTE) / SECOND),
      multiplierScaled: (135n * PRECISION) / 100n,
    },
    {
      durationSeconds: BigInt((4 * MINUTE) / SECOND),
      multiplierScaled: (150n * PRECISION) / 100n,
    },
  ];
}

export async function deployStaking(
  lmrAddress: `0x${string}`,
  morAddress: `0x${string}`,
) {
  console.log("Deploying staking contract ...");
  const staking = await hre.viem.deployContract("StakingMasterChef", [
    lmrAddress,
    morAddress,
  ]);
  console.log("Staking deployed to:", staking.address);
  const precision = await staking.read.PRECISION();

  return { staking, precision };
}

export async function setupPools(
  stakingAddress: `0x${string}`,
  pools: {
    startDate: bigint;
    durationSeconds: bigint;
    totalReward: bigint;
    lockDurations: { durationSeconds: bigint; multiplierScaled: bigint }[];
  }[],
) {
  const staking = await hre.viem.getContractAt(
    "StakingMasterChef",
    stakingAddress,
  );
  const precision = await staking.read.PRECISION();

  for (const pool of pools) {
    console.log("Adding pool ...");
    await staking.write.addPool([
      pool.startDate,
      pool.durationSeconds,
      pool.totalReward,
      pool.lockDurations.map((ld) => ({
        durationSeconds: ld.durationSeconds,
        multiplierScaled: ld.multiplierScaled,
      })),
    ]);
    console.log(
      `Pool added: startTime=${pool.startDate}, duration=${pool.durationSeconds} seconds, totalReward=${pool.totalReward}`,
    );
    console.log(
      pool.lockDurations
        .map(
          (ld, i) =>
            `id=${i} duration=${ld.durationSeconds} seconds, multiplier=${Number(ld.multiplierScaled) / Number(precision)}`,
        )
        .join("\n"),
    );
  }
}
