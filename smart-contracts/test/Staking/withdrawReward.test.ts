import hre from "hardhat";
import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { aliceStakes } from "./fixtures";
import { expect } from "chai";
import { elapsedTxs } from "./utils";
import { catchError, getTxDeltaBalance, mine, setAutomine } from "../utils";

describe("Staking contract - withdrawReward", () => {
  it("should withdraw reward correctly", async () => {
    const {
      accounts: { alice },
      contracts: { staking, tokenMOR },
      stakes,
      expPool,
      pubClient,
    } = await loadFixture(aliceStakes);

    const duration =
      expPool.lockDurations[stakes.alice.lockDurationId].durationSeconds;

    await time.increase(duration);

    const rewardTx = await staking.write.withdrawReward(
      [stakes.alice.poolId, stakes.alice.stakeId],
      { account: alice.account },
    );

    const reward = await getTxDeltaBalance(
      pubClient,
      rewardTx,
      alice.account.address,
      tokenMOR,
    );
    const stakeDuration = await elapsedTxs(stakes.alice.depositTx, rewardTx);

    expect(reward).to.equal(stakeDuration * expPool.rewardPerSecond);

    const events = await staking.getEvents.RewardWithdrawal({
      userAddress: alice.account.address,
      poolId: stakes.alice.poolId,
    });

    expect(events.length).to.equal(1);
    const [event] = events;
    expect(event.args.stakeId).to.equal(stakes.alice.stakeId);
    expect(event.args.amount).to.equal(reward);
  });

  it("should error if poolId is wrong", async () => {
    const {
      accounts: { alice },
      contracts: { staking },
      stakes,
    } = await loadFixture(aliceStakes);

    await catchError(staking.abi, "PoolNotFound", async () => {
      await staking.write.withdrawReward(
        [stakes.alice.poolId + 1n, stakes.alice.stakeId],
        { account: alice.account },
      );
    });
  });

  it("should error if stakeId is wrong", async () => {
    const {
      accounts: { alice },
      contracts: { staking },
      stakes,
    } = await loadFixture(aliceStakes);

    await catchError(staking.abi, "StakeNotFound", async () => {
      await staking.write.withdrawReward(
        [stakes.alice.poolId, stakes.alice.stakeId + 1n],
        { account: alice.account },
      );
    });
  });

  it("should error if no reward yet", async () => {
    const {
      accounts: { alice },
      contracts: { staking },
      stakes,
      pubClient,
    } = await loadFixture(aliceStakes);

    await setAutomine(hre, false);
    const rewardTx = await staking.write.withdrawReward(
      [stakes.alice.poolId, stakes.alice.stakeId],
      { account: alice.account },
    );

    await catchError(staking.abi, "NoRewardAvailable", async () => {
      await staking.write.withdrawReward(
        [stakes.alice.poolId, stakes.alice.stakeId],
        { account: alice.account },
      );
    });
    await mine(hre);
    await setAutomine(hre, true);

    await pubClient.waitForTransactionReceipt({
      hash: rewardTx,
    });
  });
});
