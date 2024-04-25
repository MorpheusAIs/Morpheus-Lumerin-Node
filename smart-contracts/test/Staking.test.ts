import { loadFixture, time } from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import { deployMarketplace } from "./fixtures";
import { expectError } from "./utils";
import { DAY, SECOND } from "../utils/time";

describe("Staking", function () {
  it("should stake", async function () {
    const { sessionRouter, expectedStake } = await loadFixture(deployMarketplace);

    expect(await sessionRouter.read.userStake([expectedStake.account])).equal(
      expectedStake.stakeAmount
    );
    expect(await sessionRouter.read.withdrawableStakeBalance([expectedStake.account])).equal(
      expectedStake.stakeAmount
    );
  });

  it("should unstake", async function () {
    const { sessionRouter, expectedStake } = await loadFixture(deployMarketplace);

    await sessionRouter.write.unstake(
      [expectedStake.account, expectedStake.stakeAmount, expectedStake.account],
      {
        account: expectedStake.account,
      }
    );

    expect(await sessionRouter.read.userStake([expectedStake.account])).equal(0n);
  });

  // rewrite with real purchases
  describe.skip("Stipend", function () {
    it("should return daily stipend", async function () {
      const { sessionRouter, expectedStake } = await loadFixture(deployMarketplace);

      const balance = await sessionRouter.read.balanceOfDailyStipend([expectedStake.account]);
      expect(balance).eq(expectedStake.expectedStipend);
    });

    it("should spend daily stipend", async function () {
      const { sessionRouter, expectedStake, tokenMOR } = await loadFixture(deployMarketplace);

      // await sessionRouter.write.transferDailyStipend(
      //   [expectedStake.account, expectedStake.transferTo, expectedStake.spendAmount],
      //   { account: expectedStake.account }
      // );

      expect(await tokenMOR.read.balanceOf([expectedStake.transferTo])).eq(
        expectedStake.spendAmount
      );
      expect(await sessionRouter.read.balanceOfDailyStipend([expectedStake.account])).eq(
        expectedStake.expectedStipend - expectedStake.spendAmount
      );
    });

    it("should count spend amount correctly when spent second time same day", async function () {
      const { sessionRouter, expectedStake, tokenMOR } = await loadFixture(deployMarketplace);

      // await sessionRouter.write.transferDailyStipend(
      //   [expectedStake.account, expectedStake.transferTo, expectedStake.spendAmount],
      //   { account: expectedStake.account }
      // );
      expect(await tokenMOR.read.balanceOf([expectedStake.transferTo])).eq(
        expectedStake.spendAmount
      );
      expect(await sessionRouter.read.balanceOfDailyStipend([expectedStake.account])).eq(
        expectedStake.expectedStipend - expectedStake.spendAmount
      );

      // await sessionRouter.write.transferDailyStipend(
      //   [expectedStake.account, expectedStake.transferTo, expectedStake.spendAmount],
      //   { account: expectedStake.account }
      // );
      expect(await tokenMOR.read.balanceOf([expectedStake.transferTo])).eq(
        expectedStake.spendAmount * 2n
      );
      expect(await sessionRouter.read.balanceOfDailyStipend([expectedStake.account])).eq(
        expectedStake.expectedStipend - expectedStake.spendAmount * 2n
      );
    });

    it("should error when spending more than daily stipend", async function () {
      const { sessionRouter, expectedStake } = await loadFixture(deployMarketplace);
      try {
        // await sessionRouter.write.transferDailyStipend(
        //   [expectedStake.account, expectedStake.transferTo, expectedStake.expectedStipend + 1n],
        //   { account: expectedStake.account }
        // );
        expect.fail("Should have thrown an error");
      } catch (e) {
        expectError(e, sessionRouter.abi, "NotEnoughStipend");
      }
    });

    it("should reset daily stipend value on the next day", async function () {
      const { sessionRouter, expectedStake, tokenMOR } = await loadFixture(deployMarketplace);

      // await sessionRouter.write.transferDailyStipend(
      //   [expectedStake.account, expectedStake.transferTo, expectedStake.spendAmount],
      //   { account: expectedStake.account }
      // );

      expect(await tokenMOR.read.balanceOf([expectedStake.transferTo])).eq(
        expectedStake.spendAmount
      );
      expect(await sessionRouter.read.balanceOfDailyStipend([expectedStake.account])).eq(
        expectedStake.expectedStipend - expectedStake.spendAmount
      );

      await time.increase(DAY / SECOND);
      expect(await sessionRouter.read.balanceOfDailyStipend([expectedStake.account])).eq(
        expectedStake.expectedStipend
      );
    });

    it("should lock unstaking till next day after using stipend", async function () {
      const { sessionRouter, expectedStake, tokenMOR } = await loadFixture(deployMarketplace);

      // await sessionRouter.write.transferDailyStipend(
      //   [expectedStake.account, expectedStake.transferTo, expectedStake.spendAmount],
      //   { account: expectedStake.account }
      // );

      expect(await sessionRouter.read.userStake([expectedStake.account])).eq(
        expectedStake.stakeAmount
      );

      const withdrawableStakeBefore = await sessionRouter.read.withdrawableStakeBalance([
        expectedStake.account,
      ]);
      await time.increase(DAY / SECOND);
      const withdrawableStakeAfter = await sessionRouter.read.withdrawableStakeBalance([
        expectedStake.account,
      ]);

      expect(withdrawableStakeBefore < withdrawableStakeAfter).eq(true);
      expect(withdrawableStakeAfter).eq(expectedStake.stakeAmount);

      const balanceBefore = await tokenMOR.read.balanceOf([expectedStake.account]);
      await sessionRouter.write.unstake(
        [expectedStake.account, expectedStake.stakeAmount, expectedStake.account],
        {
          account: expectedStake.account,
        }
      );
      const balanceAfter = await tokenMOR.read.balanceOf([expectedStake.account]);

      expect(await sessionRouter.read.userStake([expectedStake.account])).eq(0n);
      expect(balanceAfter - balanceBefore).eq(expectedStake.stakeAmount);
    });

    it("should return stipend to the staking contract and increase daily stipend", async function () {
      const { sessionRouter, expectedStake, tokenMOR } = await loadFixture(deployMarketplace);

      const stipendBeforeTransfer = await sessionRouter.read.balanceOfDailyStipend([
        expectedStake.account,
      ]);

      // await sessionRouter.write.transferDailyStipend(
      //   [expectedStake.account, expectedStake.transferTo, expectedStake.spendAmount],
      //   { account: expectedStake.account }
      // );

      const stipendAfterTransfer = await sessionRouter.read.balanceOfDailyStipend([
        expectedStake.account,
      ]);

      await tokenMOR.write.approve([sessionRouter.address, expectedStake.spendAmount], {
        account: expectedStake.transferTo,
      });
      await sessionRouter.write.returnStipend([expectedStake.account, expectedStake.spendAmount], {
        account: expectedStake.transferTo,
      });

      const stipendAfterRefund = await sessionRouter.read.balanceOfDailyStipend([
        expectedStake.account,
      ]);

      expect(stipendAfterTransfer).eq(stipendBeforeTransfer - expectedStake.spendAmount);
      expect(stipendBeforeTransfer).eq(stipendAfterRefund);
    });
  });
});
