import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { aliceStakes } from "./fixtures";
import { expect } from "chai";
import { catchError, getTxDeltaBalance } from "../utils";
import { DAY, SECOND } from "../../utils/time";
import { elapsedTxs } from "./utils";

describe("Staking contract - getReward", () => {
  it("Should get reward correctly for user that staked", async () => {
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

    const reward = await staking.read.getReward(
      [alice.account.address, stakes.alice.poolId, stakes.alice.stakeId],
      { account: alice.account, blockTag: "pending" }, // we need to check reward against the block that will be mined next
    );

    const rewardTx = await staking.write.withdrawReward(
      [stakes.alice.poolId, stakes.alice.stakeId],
      { account: alice.account },
    );

    const reward2 = await getTxDeltaBalance(
      pubClient,
      rewardTx,
      alice.account.address,
      tokenMOR,
    );
    const stakeDuration = await elapsedTxs(stakes.alice.depositTx, rewardTx);

    expect(reward).to.equal(reward2);
    expect(reward).to.equal(stakeDuration * expPool.rewardPerSecond);
  });

  it("Should error if stakeId is wrong", async () => {
    const {
      accounts: { alice },
      contracts: { staking },
      stakes,
    } = await loadFixture(aliceStakes);

    await catchError(staking.abi, "PoolOrStakeNotFound", async () => {
      await staking.read.getReward([
        alice.account.address,
        stakes.alice.poolId,
        stakes.alice.stakeId + 1n,
      ]);
    });
  });

  it("Should error if user has not staked", async () => {
    const {
      accounts: { bob },
      contracts: { staking },
      stakes,
    } = await loadFixture(aliceStakes);

    await catchError(staking.abi, "PoolOrStakeNotFound", async () => {
      await staking.read.getReward([
        bob.account.address,
        stakes.alice.poolId,
        stakes.alice.stakeId,
      ]);
    });
  });

  it("Should error if pool doesn't exist", async () => {
    const {
      accounts: { alice },
      contracts: { staking },
      stakes,
    } = await loadFixture(aliceStakes);

    await catchError(staking.abi, "PoolOrStakeNotFound", async () => {
      await staking.read.getReward([
        alice.account.address,
        stakes.alice.poolId + 1n,
        stakes.alice.stakeId,
      ]);
    });
  });

  it("Should return 0 if user withdrawn all rewards", async () => {
    const {
      accounts: { alice },
      contracts: { staking },
      stakes,
    } = await loadFixture(aliceStakes);

    await time.increase(10 * (DAY / SECOND));

    const rewardBefore = await staking.read.getReward([
      alice.account.address,
      stakes.alice.poolId,
      stakes.alice.stakeId,
    ]);
    expect(rewardBefore > 0).to.be.true;

    await staking.write.withdrawReward(
      [stakes.alice.poolId, stakes.alice.stakeId],
      { account: alice.account },
    );

    const reward = await staking.read.getReward([
      alice.account.address,
      stakes.alice.poolId,
      stakes.alice.stakeId,
    ]);
    expect(reward).to.equal(0n);
  });
});
