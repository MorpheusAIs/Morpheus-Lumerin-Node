import { loadFixture, time } from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import { stake } from "./fixtures";
import { expectError } from "./utils";
import { DAY, SECOND } from "../utils/time";

describe("Staking", function () {
  it("should stake", async function () {
    const { staking, expected } = await loadFixture(stake);

    expect(await staking.read.userStake([expected.account])).equal(expected.stakeAmount);
    expect(await staking.read.withdrawableStakeBalance([expected.account])).equal(
      expected.stakeAmount
    );
  });

  it("should unstake", async function () {
    const { staking, expected } = await loadFixture(stake);

    await staking.write.unstake([expected.account, expected.stakeAmount, expected.account], {
      account: expected.account,
    });

    expect(await staking.read.userStake([expected.account])).equal(0n);
  });

  it("should return daily stipend", async function () {
    const { staking, expected } = await loadFixture(stake);

    const balance = await staking.read.balanceOfDailyStipend([expected.account]);
    expect(balance).eq(expected.expectedStipend);
  });

  it("should spend daily stipend", async function () {
    const { staking, expected, tokenMOR } = await loadFixture(stake);

    await staking.write.transferDailyStipend(
      [expected.account, expected.transferTo.account.address, expected.spendAmount],
      { account: expected.account }
    );

    expect(await tokenMOR.read.balanceOf([expected.transferTo.account.address])).eq(
      expected.spendAmount
    );
    expect(await staking.read.balanceOfDailyStipend([expected.account])).eq(
      expected.expectedStipend - expected.spendAmount
    );
  });

  it("should error when spending more than daily stipend", async function () {
    const { staking, expected } = await loadFixture(stake);
    try {
      await staking.write.transferDailyStipend(
        [expected.account, expected.transferTo.account.address, expected.expectedStipend + 1n],
        { account: expected.account }
      );
      expect.fail("Should have thrown an error");
    } catch (e) {
      expectError(e, staking.abi, "NotEnoughDailyStipend");
    }
  });

  it("should reset daily stipend value on the next day", async function () {
    const { staking, expected, tokenMOR } = await loadFixture(stake);

    await staking.write.transferDailyStipend(
      [expected.account, expected.transferTo.account.address, expected.spendAmount],
      { account: expected.account }
    );

    expect(await tokenMOR.read.balanceOf([expected.transferTo.account.address])).eq(
      expected.spendAmount
    );
    expect(await staking.read.balanceOfDailyStipend([expected.account])).eq(
      expected.expectedStipend - expected.spendAmount
    );

    await time.increase(DAY / SECOND);
    expect(await staking.read.balanceOfDailyStipend([expected.account])).eq(
      expected.expectedStipend
    );
  });

  it("should lock unstaking till next day after using stipend", async function () {
    const { staking, expected, tokenMOR } = await loadFixture(stake);

    await staking.write.transferDailyStipend(
      [expected.account, expected.transferTo.account.address, expected.spendAmount],
      { account: expected.account }
    );

    expect(await staking.read.userStake([expected.account])).eq(expected.stakeAmount);

    const withdrawableStakeBefore = await staking.read.withdrawableStakeBalance([expected.account]);
    await time.increase(DAY / SECOND);
    const withdrawableStakeAfter = await staking.read.withdrawableStakeBalance([expected.account]);

    expect(withdrawableStakeBefore < withdrawableStakeAfter).eq(true);
    expect(withdrawableStakeAfter).eq(expected.stakeAmount);

    const balanceBefore = await tokenMOR.read.balanceOf([expected.account]);
    await staking.write.unstake([expected.account, expected.stakeAmount, expected.account], {
      account: expected.account,
    });
    const balanceAfter = await tokenMOR.read.balanceOf([expected.account]);

    expect(await staking.read.userStake([expected.account])).eq(0n);
    expect(balanceAfter - balanceBefore).eq(expected.stakeAmount);
  });

  it("should return stipend to the staking contract and increase daily stipend", async function () {
    const { staking, expected, tokenMOR } = await loadFixture(stake);

    const stipendBeforeTransfer = await staking.read.balanceOfDailyStipend([expected.account]);

    await staking.write.transferDailyStipend(
      [expected.account, expected.transferTo.account.address, expected.spendAmount],
      { account: expected.account }
    );

    const stipendAfterTransfer = await staking.read.balanceOfDailyStipend([expected.account]);

    await tokenMOR.write.approve([staking.address, expected.spendAmount], {
      account: expected.transferTo.account,
    });
    await staking.write.returnStipend([expected.account, expected.spendAmount], {
      account: expected.transferTo.account,
    });

    const stipendAfterRefund = await staking.read.balanceOfDailyStipend([expected.account]);

    expect(stipendAfterTransfer).eq(stipendBeforeTransfer - expected.spendAmount);
    expect(stipendBeforeTransfer).eq(stipendAfterRefund);
  });
});
