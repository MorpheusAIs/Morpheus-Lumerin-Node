import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { aliceAndBobStake, aliceStakes } from "./fixtures";
import { expect } from "chai";
import { elapsedTxs, getStakeId } from "./utils";
import { DAY, SECOND } from "../../utils/time";
import { getTxDeltaBalance } from "../utils";

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

  it("should withdraw undistributed rewards if there is a period with no stakers", async () => {
    const {
      contracts: { staking, tokenMOR, tokenLMR },
      stakes,
      expPool,
      accounts: { owner, alice },
      pubClient,
    } = await loadFixture(aliceStakes);

    await time.increase(expPool.duration / 3n);
    const tx = await staking.write.unstake(
      [stakes.alice.poolId, stakes.alice.stakeId],
      {
        account: alice.account,
      },
    );
    const earned1 = await getTxDeltaBalance(
      pubClient,
      tx,
      alice.account.address,
      tokenMOR,
    );
    console.log("Alice earned 1", earned1.toString());

    await time.increase(expPool.duration / 3n);
    await tokenLMR.write.approve(
      [staking.address, stakes.alice.stakingAmount],
      { account: alice.account },
    );
    const stakeTx2 = await staking.write.stake(
      [stakes.alice.poolId, stakes.alice.stakingAmount, 0],
      { account: alice.account },
    );
    const stakeId2 = await getStakeId(stakeTx2);

    await time.increase(expPool.duration / 3n);
    const tx2 = await staking.write.unstake([stakes.alice.poolId, stakeId2], {
      account: alice.account,
    });
    const earned2 = await getTxDeltaBalance(
      pubClient,
      tx2,
      alice.account.address,
      tokenMOR,
    );
    console.log("Alice earned 2", earned2.toString());

    const balance = await tokenMOR.read.balanceOf([staking.address]);
    console.log("total reward", expPool.totalReward.toString());

    const [, , , , , , , undistributedReward] = await staking.read.pools([
      stakes.alice.poolId,
    ]);
    console.log("unused", undistributedReward);
    expect(balance).to.equal(undistributedReward);

    const tx3 = await staking.write.withdrawUndistributedReward([
      stakes.alice.poolId,
    ]);
    const earned3 = await getTxDeltaBalance(
      pubClient,
      tx3,
      owner.account.address,
      tokenMOR,
    );
    console.log("Owner earned", earned3.toString());
    expect(earned3).to.equal(undistributedReward);
  });
});
