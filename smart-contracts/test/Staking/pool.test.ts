import hre from "hardhat";
import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { aliceStakes, setupStaking } from "./fixtures";
import { expect } from "chai";
import { catchError, getTxDeltaBalance, getTxTimestamp } from "../utils";
import { DAY, SECOND } from "../../utils/time";

describe("Staking contract - Add pool", () => {
  it("Should verify adding pool", async () => {
    const {
      contracts: { staking },
      expPool,
    } = await loadFixture(setupStaking);

    const poolInfo = await staking.read.pools([expPool.id]);
    expect(poolInfo).deep.equal([
      expPool.rewardPerSecond * expPool.precision,
      expPool.startDate,
      0n,
      0n,
      expPool.startDate,
      expPool.endDate,
    ]);

    const lockDurations = await staking.read.getLockDurations([expPool.id]);
    expect(lockDurations).deep.equal(expPool.lockDurations);
  });

  it("Should error adding pool if not owner", async () => {
    const {
      contracts: { staking, tokenMOR },
      accounts: { alice },
      expPool,
    } = await loadFixture(setupStaking);

    await tokenMOR.write.approve([staking.address, expPool.totalReward]);

    await catchError(staking.abi, "OwnableUnauthorizedAccount", async () => {
      await staking.write.addPool(
        [
          expPool.startDate,
          expPool.duration,
          expPool.totalReward,
          expPool.lockDurations,
        ],
        { account: alice.account },
      );
    });
  });

  it("Should error adding pool if not approved", async () => {
    const {
      contracts: { staking, tokenMOR },
      expPool,
    } = await loadFixture(setupStaking);

    await catchError(tokenMOR.abi, "ERC20InsufficientAllowance", async () => {
      await staking.write.addPool([
        expPool.startDate,
        expPool.duration,
        expPool.totalReward,
        expPool.lockDurations,
      ]);
    });
  });

  it("Should error adding pool if not enough balance", async () => {
    const {
      contracts: { staking, tokenMOR },
      accounts: { owner, alice },
      expPool,
    } = await loadFixture(setupStaking);

    const balance = await tokenMOR.read.balanceOf([owner.account.address]);
    await tokenMOR.write.transfer([alice.account.address, balance]);

    await tokenMOR.write.approve([staking.address, expPool.totalReward]);

    await catchError(tokenMOR.abi, "ERC20InsufficientBalance", async () => {
      await staking.write.addPool([
        expPool.startDate,
        expPool.duration,
        expPool.totalReward,
        expPool.lockDurations,
      ]);
    });
  });
});

describe("Staking contract - Stop pool", () => {
  it("Should stop pool", async () => {
    const {
      contracts: { staking },
      expPool,
    } = await loadFixture(aliceStakes);

    const pubClient = await hre.viem.getPublicClient();
    const stopTx = await staking.write.stopPool([expPool.id]);
    const timestamp = await getTxTimestamp(pubClient, stopTx);

    const [, , , , startTime, endTime] = await staking.read.pools([expPool.id]);

    expect(startTime).equal(expPool.startDate);
    expect(endTime).equal(timestamp);
  });

  it("Should pay back unused reward", async () => {
    const {
      contracts: { staking, tokenMOR },
      accounts: { owner, alice: aliceAccount },
      stakes: { alice },
      expPool,
    } = await loadFixture(aliceStakes);

    const pubClient = await hre.viem.getPublicClient();
    await time.increase((1 * DAY) / SECOND);
    const stopTx = await staking.write.stopPool([expPool.id]);
    const stoppedAt = await getTxTimestamp(pubClient, stopTx);

    const ownerPayback = await getTxDeltaBalance(
      pubClient,
      stopTx,
      owner.account.address,
      tokenMOR,
    );

    const unstakeTx = await staking.write.unstake(
      [alice.poolId, alice.stakeId],
      { account: aliceAccount.account },
    );
    const aliceReward = await getTxDeltaBalance(
      pubClient,
      unstakeTx,
      aliceAccount.account.address,
      tokenMOR,
    );
    const stakedAt = await getTxTimestamp(pubClient, alice.depositTx);

    const expPayback = (expPool.endDate - stoppedAt) * expPool.rewardPerSecond;
    const expAliceReward = (stoppedAt - stakedAt) * expPool.rewardPerSecond;

    expect(ownerPayback).equal(expPayback);
    expect(aliceReward).equal(expAliceReward);
  });

  it("Should error stopping pool if not owner", async () => {
    const {
      contracts: { staking },
      expPool,
      accounts: { alice },
    } = await loadFixture(aliceStakes);

    await catchError(staking.abi, "OwnableUnauthorizedAccount", async () => {
      await staking.write.stopPool([expPool.id], {
        account: alice.account,
      });
    });
  });

  it("Should not be able to stake after pool is stopped", async () => {
    const {
      contracts: { staking, tokenLMR },
      expPool,
      stakes: {
        alice: { poolId, lockDurationId },
      },
      accounts: { bob },
    } = await loadFixture(aliceStakes);

    await staking.write.stopPool([expPool.id]);

    const stakeAmount = 1000n;
    await tokenLMR.write.approve([staking.address, stakeAmount], {
      account: bob.account,
    });

    await catchError(staking.abi, "StakingFinished", async () => {
      await staking.write.stake([poolId, stakeAmount, lockDurationId], {
        account: bob.account,
      });
    });
  });

  it("Should be able to unstake after pool is stopped", async () => {
    const {
      contracts: { staking, tokenLMR, tokenMOR },
      accounts: { alice },
      expPool,
      stakes: {
        alice: { poolId, stakeId, stakingAmount, depositTx },
      },
      pubClient,
    } = await loadFixture(aliceStakes);

    const startTime = await getTxTimestamp(pubClient, depositTx);

    const stopTx = await staking.write.stopPool([expPool.id]);
    const stopTime = await getTxTimestamp(pubClient, stopTx);

    const lmrBalanceBefore = await tokenLMR.read.balanceOf([
      alice.account.address,
    ]);
    const morBalanceBefore = await tokenMOR.read.balanceOf([
      alice.account.address,
    ]);

    await staking.write.unstake([poolId, stakeId], {
      account: alice.account,
    });

    const lmrBalanceAfter = await tokenLMR.read.balanceOf([
      alice.account.address,
    ]);
    const morBalanceAfter = await tokenMOR.read.balanceOf([
      alice.account.address,
    ]);

    expect(lmrBalanceAfter - lmrBalanceBefore).to.equal(
      stakingAmount,
      "should return staked balance",
    );
    expect(morBalanceAfter - morBalanceBefore).to.equal(
      (stopTime - startTime) * expPool.rewardPerSecond,
      "should return earned balance",
    );
  });
});

describe("Staking contract - updatePoolReward", () => {
  it("should update reward manually", async () => {
    const {
      contracts: { staking },
      expPool,
      stakes,
      pubClient,
    } = await loadFixture(aliceStakes);

    const aliceStakeTime = await getTxTimestamp(
      pubClient,
      stakes.alice.depositTx,
    );

    await time.increase((1 * DAY) / SECOND);
    const [, lastRewardTimeBf, rewardPerShareBf] = await staking.read.pools([
      expPool.id,
    ]);
    await staking.write.recalculatePoolReward([expPool.id]);
    const [, lastRewardTimeAf, rewardPerShareAf] = await staking.read.pools([
      expPool.id,
    ]);

    expect(aliceStakeTime).to.be.eq(lastRewardTimeBf);
    expect(lastRewardTimeAf > lastRewardTimeBf).to.be.true;
    expect(rewardPerShareAf > rewardPerShareBf).to.be.true;
  });
});
