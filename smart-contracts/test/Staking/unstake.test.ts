import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { aliceStakes, setupStaking } from "./fixtures";
import { expect } from "chai";
import { elapsedTxs, getPoolId } from "./utils";
import { catchError, getTxDeltaBalance } from "../utils";
import { DAY, HOUR, SECOND } from "../../utils/time";

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

    await catchError(staking.abi, "PoolOrStakeNotExists", async () => {
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

    await catchError(staking.abi, "PoolOrStakeNotExists", async () => {
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

    await catchError(staking.abi, "LockNotEnded", async () => {
      await staking.write.unstake([poolId, stakeId], {
        account: alice.account,
      });
    });
  });

  it("should allow unstaking before start date", async () => {
    const {
      contracts: { staking, tokenLMR, tokenMOR },
      expPool,
      accounts: { alice },
      pubClient,
    } = await loadFixture(setupStaking);

    const now = await time.latest();
    const startTime = BigInt(now + DAY / SECOND);
    const duration = 10n * BigInt(DAY / SECOND);
    const rewardPerSecond = 100n;
    const totalReward = rewardPerSecond * BigInt(duration);

    await tokenMOR.write.approve([staking.address, totalReward]);
    const tx = await staking.write.addPool([
      startTime,
      duration,
      totalReward,
      [
        {
          durationSeconds: BigInt(DAY / SECOND),
          multiplierScaled: 1n * expPool.precision,
        },
      ],
    ]);

    const poolId = await getPoolId(tx);

    const stakeAmount = 1000n;
    await tokenLMR.write.approve([staking.address, stakeAmount], {
      account: alice.account,
    });

    await staking.write.stake([poolId, stakeAmount, 0], {
      account: alice.account,
    });

    await time.increaseTo(startTime - BigInt(HOUR / SECOND));
    await staking.write.unstake([poolId, 0n], {
      account: alice.account,
    });
  });

  it("should not count prestaking period into rewards", async () => {
    const {
      contracts: { staking, tokenLMR, tokenMOR },
      expPool,
      accounts: { alice },
      pubClient,
    } = await loadFixture(setupStaking);

    const now = await time.latest();
    const startTime = BigInt(now + DAY / SECOND);
    const duration = 10n * BigInt(DAY / SECOND);
    const rewardPerSecond = 100n;
    const totalReward = rewardPerSecond * BigInt(duration);

    await tokenMOR.write.approve([staking.address, totalReward]);
    const tx = await staking.write.addPool([
      startTime,
      duration,
      totalReward,
      [
        {
          durationSeconds: BigInt(DAY / SECOND),
          multiplierScaled: 1n * expPool.precision,
        },
      ],
    ]);

    const poolId = await getPoolId(tx);
    // TODO: add above part to fixture (setupStaking in future)

    const stakeAmount = 1000n;
    await tokenLMR.write.approve([staking.address, stakeAmount], {
      account: alice.account,
    });

    await staking.write.stake([poolId, stakeAmount, 0], {
      account: alice.account,
    });

    await time.increaseTo(startTime - BigInt(HOUR / SECOND));

    const tx2 = await staking.write.unstake([poolId, 0n], {
      account: alice.account,
    });

    const deltaMOR = await getTxDeltaBalance(
      pubClient,
      tx2,
      alice.account.address,
      tokenMOR,
    );

    const deltaLMR = await getTxDeltaBalance(
      pubClient,
      tx2,
      alice.account.address,
      tokenLMR,
    );

    expect(deltaMOR).to.equal(0n);
    expect(deltaLMR).to.equal(stakeAmount);
  });
});
