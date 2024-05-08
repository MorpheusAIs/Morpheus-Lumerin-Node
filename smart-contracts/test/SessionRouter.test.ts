import {
  loadFixture,
  time,
} from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import hre from "hardhat";
import { keccak256, maxUint256 } from "viem";
import { deploySingleBid, encodedReport } from "./fixtures";
import {
  catchError,
  getHex,
  getTxTimestamp,
  randomBytes32,
  getSessionId,
} from "./utils";
import { DAY, HOUR, MINUTE, SECOND, now, startOfNextDay } from "../utils/time";
import { expectAlmostEqual } from "../utils/compare";

describe("Session router", function () {
  describe("session read functions", function () {
    it("should get compute balance equal to one on L1", async function () {
      const { sessionRouter } = await loadFixture(deploySingleBid);
      const exp = {
        initialReward: 3456000000000000000000n,
        rewardDecrease: 592558728240000000n,
        payoutStart: 1707393600n,
        decreaseInterval: 86400n,
        blockTimeEpochSeconds:
          BigInt(new Date("2024-05-02T09:19:57Z").getTime()) / 1000n,
        balance: 287859686689348241525000n,
      };

      await sessionRouter.write.setPoolConfig([
        {
          initialReward: exp.initialReward,
          rewardDecrease: exp.rewardDecrease,
          payoutStart: exp.payoutStart,
          decreaseInterval: exp.decreaseInterval,
        },
      ]);

      const balance = await sessionRouter.read.getComputeBalance([
        exp.blockTimeEpochSeconds,
      ]);

      expect(balance).to.equal(exp.balance);
    });
  });

  describe("session actions", function () {
    it("should open session without error", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        publicClient,
      } = await loadFixture(deploySingleBid);

      const openTx = await sessionRouter.write.openSession(
        [exp.bidID, exp.stake],
        { account: user.account },
      );

      const sessionId = await getSessionId(publicClient, hre, openTx);
      expect(sessionId).to.be.a("string");
    });

    it("should verify session fields after opening", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        publicClient,
      } = await loadFixture(deploySingleBid);

      const txHash = await sessionRouter.write.openSession(
        [exp.bidID, exp.stake],
        { account: user.account },
      );

      const sessionId = await getSessionId(publicClient, hre, txHash);
      const session = await sessionRouter.read.getSession([sessionId]);
      const createdAt = await getTxTimestamp(publicClient, txHash);

      expect(session).to.deep.equal({
        id: sessionId,
        user: exp.user,
        provider: exp.provider,
        modelAgentId: exp.modelAgentId,
        bidID: exp.bidID,
        stake: exp.stake,
        pricePerSecond: exp.pricePerSecond,
        closeoutReceipt: getHex(Buffer.from(""), 0),
        closeoutType: 0n,
        providerWithdrawnAmount: 0n,
        openedAt: createdAt,
        closedAt: 0n,
      });
    });

    it("should verify balances after opening", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        publicClient,
        user,
        tokenMOR,
      } = await loadFixture(deploySingleBid);

      const srBefore = await tokenMOR.read.balanceOf([sessionRouter.address]);
      const userBefore = await tokenMOR.read.balanceOf([user.account.address]);

      const txHash = await sessionRouter.write.openSession(
        [exp.bidID, exp.stake],
        { account: user.account },
      );
      await publicClient.waitForTransactionReceipt({ hash: txHash });

      const srAfter = await tokenMOR.read.balanceOf([sessionRouter.address]);
      const userAfter = await tokenMOR.read.balanceOf([user.account.address]);

      expect(srAfter - srBefore).to.equal(exp.stake);
      expect(userBefore - userAfter).to.equal(exp.stake);
    });

    it("should error when opening session with missing bid", async function () {
      const {
        sessionRouter,
        user,
        expectedSession: exp,
      } = await loadFixture(deploySingleBid);

      await catchError(sessionRouter.abi, "BidNotFound", async () => {
        await sessionRouter.write.openSession([randomBytes32(), exp.stake], {
          account: user.account,
        });
      });
    });

    it("should error when opening with same bid simultaneously", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
      } = await loadFixture(deploySingleBid);

      await sessionRouter.write.openSession([exp.bidID, exp.stake], {
        account: user.account.address,
      });

      await catchError(sessionRouter.abi, "BidTaken", async () => {
        await sessionRouter.write.openSession([exp.bidID, exp.stake], {
          account: user.account.address,
        });
      });
    });
  });

  describe("session end time", function () {
    it("should open session (spans 1 day) and verify end time", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        publicClient,
      } = await loadFixture(deploySingleBid);
      // open session
      const openTx = await sessionRouter.write.openSession(
        [exp.bidID, exp.stake],
        {
          account: user.account.address,
        },
      );

      const startedAt = await getTxTimestamp(publicClient, openTx);
      const sessionId = await getSessionId(publicClient, hre, openTx);
      const endTime = await sessionRouter.read.getSessionEndTime([sessionId]);

      expect(endTime).to.equal(startedAt + exp.durationSeconds);
    });

    it("should open session (spans 2 days) and verify end time", async function () {
      //TODO: improve code to reduce time difference
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        publicClient,
      } = await loadFixture(deploySingleBid);

      // increase time to 23:00 next day
      await time.increaseTo(
        startOfNextDay(now()) + BigInt((23 * HOUR) / SECOND),
      );
      const openTx = await sessionRouter.write.openSession(
        [exp.bidID, exp.stake * 2n],
        {
          account: user.account.address,
        },
      );

      const startedAt = await getTxTimestamp(publicClient, openTx);
      const sessionId = await getSessionId(publicClient, hre, openTx);
      const endTime = await sessionRouter.read.getSessionEndTime([sessionId]);
      const expEndTime = startOfNextDay(startedAt) + exp.durationSeconds * 2n;

      expect(Number(endTime)).to.approximately(
        Number(expEndTime),
        (3 * MINUTE) / SECOND,
      );
    });

    it("should open session (longer than 1 day) and verify end time", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        publicClient,
        tokenMOR,
      } = await loadFixture(deploySingleBid);
      const stake = exp.stake * 25n;

      await tokenMOR.write.transfer([user.account.address, stake]);

      const openTx = await sessionRouter.write.openSession([exp.bidID, stake], {
        account: user.account.address,
      });

      const sessionId = await getSessionId(publicClient, hre, openTx);
      const endTime = await sessionRouter.read.getSessionEndTime([sessionId]);

      expect(endTime).to.be.equal(maxUint256);
    });

    it("should open session (longer than 1 day) that ends today and verify end time", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        owner,
        publicClient,
        tokenMOR,
      } = await loadFixture(deploySingleBid);
      const stake = exp.stake * 24n;

      await tokenMOR.write.transfer([user.account.address, stake]);

      const openTx = await sessionRouter.write.openSession([exp.bidID, stake], {
        account: user.account.address,
      });

      const sessionId = await getSessionId(publicClient, hre, openTx);
      const endTime = await sessionRouter.read.getSessionEndTime([sessionId]);
      console.log("endTime", endTime);

      await time.increase((1000 * DAY) / SECOND);
      const endTime2 = await sessionRouter.read.getSessionEndTime([sessionId]);
      console.log("endTime2", endTime2);

      // expect(endTime).to.be.equal(maxUint256);
    });
  });

  describe("session closeout", function () {
    it("should open short (< 1D) session and close later", async function () {
      const {
        sessionRouter,
        provider,
        expectedSession: exp,
        user,
        publicClient,
        tokenMOR,
      } = await loadFixture(deploySingleBid);
      // open session
      const openTx = await sessionRouter.write.openSession(
        [exp.bidID, exp.stake],
        {
          account: user.account.address,
        },
      );
      const sessionId = await getSessionId(publicClient, hre, openTx);

      await time.increase(exp.durationSeconds * 2n);

      const userBalanceBefore = await tokenMOR.read.balanceOf([
        user.account.address,
      ]);
      const providerBalanceBefore = await tokenMOR.read.balanceOf([
        provider.account.address,
      ]);

      // close session
      const signature = await provider.signMessage({
        message: { raw: keccak256(encodedReport) },
      });
      await sessionRouter.write.closeSession(
        [sessionId, encodedReport, signature],
        {
          account: user.account,
        },
      );

      // verify session is closed without dispute
      const session = await sessionRouter.read.getSession([sessionId]);
      expect(session.closeoutType).to.equal(0n);

      // verify balances
      const userBalanceAfter = await tokenMOR.read.balanceOf([
        user.account.address,
      ]);
      const providerBalanceAfter = await tokenMOR.read.balanceOf([
        provider.account.address,
      ]);

      const userStakeReturned = userBalanceAfter - userBalanceBefore;
      const providerEarned = providerBalanceAfter - providerBalanceBefore;

      expect(userStakeReturned).to.equal(0n);
      expectAlmostEqual(providerEarned, exp.totalCost, 0.0001);
    });

    it("should open short (<1D) session and close early", async function () {
      const {
        sessionRouter,
        provider,
        expectedSession: exp,
        user,
        publicClient,
        tokenMOR,
      } = await loadFixture(deploySingleBid);
      // open session
      const openTx = await sessionRouter.write.openSession(
        [exp.bidID, exp.stake],
        {
          account: user.account.address,
        },
      );
      const sessionId = await getSessionId(publicClient, hre, openTx);

      await time.increase(exp.durationSeconds / 2n - 1n);

      const userBalanceBefore = await tokenMOR.read.balanceOf([
        user.account.address,
      ]);
      const providerBalanceBefore = await tokenMOR.read.balanceOf([
        provider.account.address,
      ]);

      // close session
      const signature = await provider.signMessage({
        message: { raw: keccak256(encodedReport) },
      });
      await sessionRouter.write.closeSession(
        [sessionId, encodedReport, signature],
        {
          account: user.account,
        },
      );

      // verify session is closed without dispute
      const session = await sessionRouter.read.getSession([sessionId]);
      expect(session.closeoutType).to.equal(0n);

      // verify balances
      const userBalanceAfter = await tokenMOR.read.balanceOf([
        user.account.address,
      ]);
      const providerBalanceAfter = await tokenMOR.read.balanceOf([
        provider.account.address,
      ]);

      const userStakeReturned = userBalanceAfter - userBalanceBefore;
      const providerEarned = providerBalanceAfter - providerBalanceBefore;

      expectAlmostEqual(userStakeReturned, exp.stake / 2n, 0.0001);
      expectAlmostEqual(providerEarned, exp.totalCost / 2n, 0.0001);
    });

    it.skip("should open and close with user report - dispute", async function () {
      const {
        sessionRouter,
        provider,
        expectedBid,
        user,
        publicClient,
        tokenMOR,
      } = await loadFixture(deploySingleBid);
      const budget = expectedBid.pricePerSecond * BigInt(HOUR / SECOND);

      // save balance before opening session
      const balanceBeforeOpen = await sessionRouter.read.balanceOfDailyStipend([
        user.account.address,
      ]);
      const providerBalanceBefore = await tokenMOR.read.balanceOf([
        provider.account.address,
      ]);

      // open session
      const openTx = await sessionRouter.write.openSession(
        [expectedBid.id, budget],
        {
          account: user.account.address,
        },
      );
      const sessionId = await getSessionId(publicClient, hre, openTx);

      await time.increase((30 * MINUTE) / SECOND);
      const balanceBeforeClose = await sessionRouter.read.balanceOfDailyStipend(
        [user.account.address],
      );

      // close session with invalid signature
      const signature = getHex(Buffer.from(""), 0);
      await sessionRouter.write.closeSession(
        [sessionId, encodedReport, signature],
        {
          account: user.account,
        },
      );

      // verify session is closed with dispute
      const session = await sessionRouter.read.getSession([sessionId]);
      expect(session.closeoutType).to.equal(1n);

      // verify balances
      const balanceAfterClose = await sessionRouter.read.balanceOfDailyStipend([
        user.account.address,
      ]);
      const providerBalanceAfter = await tokenMOR.read.balanceOf([
        provider.account.address,
      ]);
      const [total, onHold] = await sessionRouter.read.getProviderBalance([
        provider.account.address,
      ]);

      const stipendLocked = balanceBeforeOpen - balanceBeforeClose;
      const stipendSpent = balanceBeforeOpen - balanceAfterClose;
      const providerEarned = providerBalanceAfter - providerBalanceBefore;

      expect(providerEarned).to.equal(0n);
      expect(onHold).to.equal(stipendSpent);
      expectAlmostEqual(stipendLocked / 2n, stipendSpent, 0.05);

      // verify provider balance after dispute is released
      await time.increase((1 * DAY) / SECOND);
      const [total2, onHold2] = await sessionRouter.read.getProviderBalance([
        provider.account.address,
      ]);
      expect(total2).to.equal(total);
      expect(onHold2).to.equal(0n);

      // verify user balance after dispute is claimable
      await sessionRouter.write.claimProviderBalance(
        [total2, provider.account.address],
        {
          account: provider.account.address,
        },
      );

      const [total3, onHold3] = await sessionRouter.read.getProviderBalance([
        provider.account.address,
      ]);
      expect(total3).to.equal(0n);
      expect(onHold3).to.equal(0n);

      // verify provider balance after claim
      const providerBalanceAfterClaim = await tokenMOR.read.balanceOf([
        provider.account.address,
      ]);
      const providerClaimed = providerBalanceAfterClaim - providerBalanceAfter;
      expect(providerClaimed).to.equal(total2);
    });

    it.skip("should open session with same bid after previous session is closed", async function () {
      const { sessionRouter, expectedBid, user, publicClient, provider } =
        await loadFixture(deploySingleBid);
      const budget = expectedBid.pricePerSecond * BigInt(HOUR / SECOND);

      // first purchase
      const openTx = await sessionRouter.write.openSession(
        [expectedBid.id, budget],
        {
          account: user.account.address,
        },
      );
      const sessionId = await getSessionId(publicClient, hre, openTx);

      // first closeout
      const signature = await provider.signMessage({
        message: { raw: keccak256(encodedReport) },
      });
      await sessionRouter.write.closeSession(
        [sessionId, encodedReport, signature],
        {
          account: user.account,
        },
      );

      // second purchase same bidId
      const openTx2 = await sessionRouter.write.openSession(
        [expectedBid.id, budget],
        {
          account: user.account.address,
        },
      );

      // expect no purchase error
    });
  });
});

