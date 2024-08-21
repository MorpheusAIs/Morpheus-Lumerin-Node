import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { aliceAndBobStake, aliceStakes } from "./fixtures";
import { expect } from "chai";
import { elapsedTxs } from "./utils";
import { DAY, SECOND } from "../../utils/time";

describe("Staking contract - Complex reward scenarios", () => {
  it("Should reward correctly two stakers", async () => {
    const data = loadFixture(aliceAndBobStake);
    const {
      contracts: { staking, tokenMOR },
      stakes: { alice, bob },
      expPool,
      accounts,
    } = await data;
    await time.increase(5 * (DAY / SECOND));

    const aliceMorBalanceBefore = await tokenMOR.read.balanceOf([
      accounts.alice.account.address,
    ]);
    const aliceWithdrawTx = await staking.write.unstake(
      [alice.poolId, alice.stakeId],
      { account: accounts.alice.account },
    );
    const aliceMorBalanceAfter = await tokenMOR.read.balanceOf([
      accounts.alice.account.address,
    ]);

    const durationAliceSingle = await elapsedTxs(
      alice.depositTx,
      bob.depositTx,
    );
    const earnAliceSingle = durationAliceSingle * expPool.rewardPerSecond;
    const durationAliceDouble = await elapsedTxs(
      bob.depositTx,
      aliceWithdrawTx,
    );
    const earnAliceDouble =
      (durationAliceDouble * expPool.rewardPerSecond) / 2n;
    const totalEarnAlice = earnAliceSingle + earnAliceDouble;
    expect(aliceMorBalanceAfter - aliceMorBalanceBefore).to.equal(
      totalEarnAlice,
    );

    await time.increase(2 * (DAY / SECOND));

    const bobMorBalanceBefore = await tokenMOR.read.balanceOf([
      accounts.bob.account.address,
    ]);
    const bobWithdrawTx = await staking.write.unstake(
      [bob.poolId, bob.stakeId],
      {
        account: accounts.bob.account,
      },
    );
    const bobMorBalanceAfter = await tokenMOR.read.balanceOf([
      accounts.bob.account.address,
    ]);

    const durationBobDouble = await elapsedTxs(bob.depositTx, aliceWithdrawTx);
    const earnBobDouble = (durationBobDouble * expPool.rewardPerSecond) / 2n;
    const durationBobSingle = await elapsedTxs(aliceWithdrawTx, bobWithdrawTx);
    const earnBobSingle = durationBobSingle * expPool.rewardPerSecond;
    const totalEarnBob = earnBobSingle + earnBobDouble;
    expect(bobMorBalanceAfter - bobMorBalanceBefore).to.equal(totalEarnBob);
  });

  it("should stop increasing reward after pool end date", async () => {
    const {
      contracts: { staking },

      stakes,
      expPool,
      accounts: { alice },
    } = await loadFixture(aliceStakes);
    const reward1 = await staking.read.getReward([
      alice.account.address,
      stakes.alice.poolId,
      stakes.alice.stakeId,
    ]);

    await time.increaseTo(expPool.endDate);
    const reward2 = await staking.read.getReward([
      alice.account.address,
      stakes.alice.poolId,
      stakes.alice.stakeId,
    ]);

    await time.increase(5n * (expPool.endDate - expPool.startDate));
    const reward3 = await staking.read.getReward([
      alice.account.address,
      stakes.alice.poolId,
      stakes.alice.stakeId,
    ]);

    expect(reward1 < reward2).to.be.true;
    expect(reward2).to.equal(reward3);
  });
});
