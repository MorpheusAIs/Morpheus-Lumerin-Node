import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { aliceStakes, setupStaking } from "./fixtures";
import { expect } from "chai";
import { getPoolId } from "./utils";
import { DAY, SECOND } from "../../utils/time";
import { catchError } from "../utils";

describe("Staking contract - stake", () => {
  it("Should stake correctly and emit event", async () => {
    const data = await loadFixture(aliceStakes);
    const {
      contracts: { staking },
      stakes: {
        alice: { poolId, stakingAmount, stakeId },
      },
      accounts: { alice },
    } = data;

    const events = await staking.getEvents.Stake({
      userAddress: alice.account.address,
      poolId,
    });

    expect(events.length).to.equal(1);
    const [event] = events;
    expect(event.args.stakeId).to.equal(stakeId);
    expect(event.args.amount).to.equal(stakingAmount);
  });

  it("should error if pool does not exist", async () => {
    const {
      contracts: { staking },
      accounts: { alice },
    } = await loadFixture(aliceStakes);

    await catchError(staking.abi, "PoolOrStakeNotFound", async () => {
      await staking.write.stake([100n, 1000n, 0], {
        account: alice.account,
      });
    });
  });

  it("should error if staking before start date", async () => {
    const {
      contracts: { staking, tokenLMR, tokenMOR },
      expPool,
      accounts: { alice },
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

    await catchError(staking.abi, "StakingNotStarted", async () => {
      await staking.write.stake([poolId, stakeAmount, 0], {
        account: alice.account,
      });
    });

    await time.increaseTo(startTime);

    await staking.write.stake([expPool.id, stakeAmount, 0], {
      account: alice.account,
    });

    // no error
  });

  it("should error if staking after end date", async () => {
    const {
      contracts: { staking, tokenLMR },
      expPool,
      accounts: { alice },
    } = await loadFixture(setupStaking);
    await time.increaseTo(expPool.endDate);

    const stakeAmount = 1000n;
    await tokenLMR.write.approve([staking.address, stakeAmount], {
      account: alice.account,
    });
    await catchError(staking.abi, "StakingFinished", async () => {
      await staking.write.stake([expPool.id, stakeAmount, 0], {
        account: alice.account,
      });
    });
  });

  it("Should error if staking durartion is too long", async () => {
    const {
      contracts: { staking, tokenLMR },
      expPool,
      accounts: { alice },
    } = await loadFixture(setupStaking);

    await time.increaseTo(expPool.endDate - BigInt(DAY / SECOND));
    const stakeAmount = 1000n;
    await tokenLMR.write.approve([staking.address, stakeAmount], {
      account: alice.account,
    });

    await catchError(
      staking.abi,
      "LockDurationExceedsStakingRange",
      async () => {
        await staking.write.stake([expPool.id, stakeAmount, 0], {
          account: alice.account,
        });
      },
    );
  });

  it("should error if not enough allowance", async () => {
    const {
      contracts: { staking, tokenLMR },
      accounts: { alice },
    } = await loadFixture(setupStaking);

    await catchError(tokenLMR.abi, "ERC20InsufficientAllowance", async () => {
      await staking.write.stake([0n, 2_000_000n, 0], {
        account: alice.account,
      });
    });
  });

  it("should error if not enough allowance", async () => {
    const {
      contracts: { staking, tokenLMR },
      accounts: { alice },
    } = await loadFixture(setupStaking);
    const amount = 2_000_000n;
    await tokenLMR.write.approve([staking.address, amount], {
      account: alice.account,
    });
    await catchError(tokenLMR.abi, "ERC20InsufficientBalance", async () => {
      await staking.write.stake([0n, amount, 0], {
        account: alice.account,
      });
    });
  });
});
