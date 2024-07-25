import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { aliceStakes } from "./fixtures";
import { expect } from "chai";
import { elapsedTxs } from "./utils";
import { catchError } from "../utils";

describe("Staking contract - unstake", () => {
  it("Should unstake correctly", async () => {
    const {
      accounts: { alice },
      contracts: { staking, tokenMOR },
      stakes,
      expPool,
    } = await loadFixture(aliceStakes);

    await time.increase(
      expPool.lockDurations[stakes.alice.lockDurationId].durationSeconds,
    );

    const morBalanceBefore = await tokenMOR.read.balanceOf([
      alice.account.address,
    ]);
    const withdrawTx = await staking.write.unstake(
      [stakes.alice.poolId, stakes.alice.stakeId],
      { account: alice.account },
    );
    const morBalanceAfter = await tokenMOR.read.balanceOf([
      alice.account.address,
    ]);
    const stakeDuration = await elapsedTxs(stakes.alice.depositTx, withdrawTx);

    expect(morBalanceAfter - morBalanceBefore).to.equal(
      stakeDuration * expPool.rewardPerSecond,
    );

    const events = await staking.getEvents.Unstake({
      userAddress: alice.account.address,
      poolId: stakes.alice.poolId,
    });

    expect(events.length).to.equal(1);
    const [event] = events;
    expect(event.args.stakeId).to.equal(stakes.alice.stakeId);
    expect(event.args.amount).to.equal(stakes.alice.stakingAmount);
  });

  it("Should error if pool does not exist", async () => {
    const {
      contracts: { staking },
      accounts,
      stakes: { alice },
      expPool,
    } = await loadFixture(aliceStakes);

    await catchError(staking.abi, "PoolNotFound", async () => {
      await staking.write.unstake([expPool.id + 1n, alice.stakeId], {
        account: accounts.alice.account,
      });
    });
  });

  it("Should error if stake does not exist", async () => {
    const {
      contracts: { staking },
      stakes: { alice },
      accounts,
    } = await loadFixture(aliceStakes);

    await catchError(staking.abi, "StakeNotFound", async () => {
      await staking.write.unstake([alice.poolId, alice.stakeId + 1n], {
        account: accounts.alice.account,
      });
    });
  });

  it("Should error if unstaking before lock duration", async () => {
    const {
      contracts: { staking },
      stakes: {
        alice: { poolId, stakeId },
      },
      accounts: { alice },
    } = await loadFixture(aliceStakes);

    await catchError(staking.abi, "LockDurationNotOver", async () => {
      await staking.write.unstake([poolId, stakeId], {
        account: alice.account,
      });
    });
  });
});
