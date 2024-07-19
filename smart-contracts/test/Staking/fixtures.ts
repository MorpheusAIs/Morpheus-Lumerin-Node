import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import hre from "hardhat";

export async function setupStaking() {
  const [owner, alice, bob, carol] = await hre.viem.getWalletClients();

  const tokenMOR = await hre.viem.deployContract("MorpheusToken", []);
  const tokenLMR = await hre.viem.deployContract("LumerinToken", []);

  const expPool = {
    rewardPerSecond: 100n,
    stakingToken: tokenLMR,
    rewardToken: tokenMOR,
    totalReward: 1_000_000_000_000n,
  };

  const staking = await hre.viem.deployContract("StakingMasterChef", [
    tokenMOR.address,
    tokenLMR.address,
    owner.account.address,
    expPool.rewardPerSecond,
  ]);

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
  };
}

export async function aliceStakes() {
  const data = await loadFixture(setupStaking);
  const {
    contracts: { staking, tokenLMR },
    accounts: { alice },
  } = data;

  const stakingAmount = 1000n;
  await tokenLMR.write.approve([staking.address, stakingAmount], {
    account: alice.account,
  });
  const depositTx = await staking.write.deposit([stakingAmount, 0], {
    account: alice.account,
  });

  return { ...data, stakes: { alice: { depositTx, stakingAmount } } };
}
