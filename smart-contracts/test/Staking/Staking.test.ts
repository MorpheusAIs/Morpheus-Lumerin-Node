import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import hre from "hardhat";
import { aliceStakes, setupStaking } from "./fixtures";
import { DAY, HOUR, SECOND } from "../../utils/time";
import { expect } from "chai";
import { getTxTimestamp } from "../utils";

describe.only("Staking contract", () => {
  it("Should return correct amount for single staker", async () => {
    const {
      accounts: { alice },
      contracts: { staking, tokenMOR },
      stakes,
      expPool,
    } = await loadFixture(aliceStakes);

    await time.increase(7 * (DAY / SECOND));

    const morBalanceBefore = await tokenMOR.read.balanceOf([
      alice.account.address,
    ]);
    const withdrawTx = await staking.write.withdraw(
      [0n, stakes.alice.stakingAmount],
      {
        account: alice.account,
      },
    );
    const morBalanceAfter = await tokenMOR.read.balanceOf([
      alice.account.address,
    ]);
    const stakeDuration = await elapsedTxs(stakes.alice.depositTx, withdrawTx);

    expect(morBalanceAfter - morBalanceBefore).to.equal(
      stakeDuration * expPool.rewardPerSecond,
    );
  });

  it("Should return correct amount for two stakers", async () => {
    const {
      accounts,
      contracts: { staking, tokenMOR, tokenLMR },
      stakes,
      expPool,
    } = await loadFixture(aliceStakes);

    await time.increase(5 * (DAY / SECOND));

    const stakingAmount = 1000n;
    await tokenLMR.write.approve([staking.address, stakingAmount], {
      account: accounts.bob.account,
    });
    const bobDepositTx = await staking.write.deposit([0n, stakingAmount], {
      account: accounts.bob.account,
    });
    await time.increase(5 * (DAY / SECOND));

    const aliceMorBalanceBefore = await tokenMOR.read.balanceOf([
      accounts.alice.account.address,
    ]);
    const aliceWithdrawTx = await staking.write.withdraw(
      [0n, stakes.alice.stakingAmount],
      {
        account: accounts.alice.account,
      },
    );
    const aliceMorBalanceAfter = await tokenMOR.read.balanceOf([
      accounts.alice.account.address,
    ]);

    const durationAliceSingle = await elapsedTxs(
      stakes.alice.depositTx,
      bobDepositTx,
    );
    const earnAliceSingle = durationAliceSingle * expPool.rewardPerSecond;
    const durationAliceDouble = await elapsedTxs(bobDepositTx, aliceWithdrawTx);
    const earnAliceDouble =
      (durationAliceDouble * expPool.rewardPerSecond) / 2n;
    const totalEarnAlice = earnAliceSingle + earnAliceDouble;
    expect(aliceMorBalanceAfter - aliceMorBalanceBefore).to.equal(
      totalEarnAlice,
    );

    await time.increase(1 * (HOUR / SECOND));

    const bobMorBalanceBefore = await tokenMOR.read.balanceOf([
      accounts.bob.account.address,
    ]);
    const bobWithdrawTx = await staking.write.withdraw([0n, stakingAmount], {
      account: accounts.bob.account,
    });
    const bobMorBalanceAfter = await tokenMOR.read.balanceOf([
      accounts.bob.account.address,
    ]);

    const durationBobDouble = await elapsedTxs(bobDepositTx, aliceWithdrawTx);
    const earnBobDouble = (durationBobDouble * expPool.rewardPerSecond) / 2n;
    const durationBobSingle = await elapsedTxs(aliceWithdrawTx, bobWithdrawTx);
    const earnBobSingle = durationBobSingle * expPool.rewardPerSecond;
    const totalEarnBob = earnBobSingle + earnBobDouble;
    expect(bobMorBalanceAfter - bobMorBalanceBefore).to.equal(totalEarnBob);
  });
});

/** Elapsed time between two transactions */
async function elapsedTxs(
  tx1: `0x${string}`,
  tx2: `0x${string}`,
): Promise<bigint> {
  const pc = await hre.viem.getPublicClient();
  return (await getTxTimestamp(pc, tx2)) - (await getTxTimestamp(pc, tx1));
}
