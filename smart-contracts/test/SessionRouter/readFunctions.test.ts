import {
  loadFixture,
  time,
} from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import { deploySingleBid, getProviderApproval, getReport } from "../fixtures";
import { getSessionId } from "../utils";
import hre from "hardhat";
import { HOUR, SECOND } from "../../utils/time";

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
        balance: 286534931460577320000000n,
      };

      await sessionRouter.write.setPoolConfig([
        3n,
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

  describe("getProviderClaimableBalance", function () {
    it("should be correct for contract that closed early due to dispute [H-6]", async function () {
      const {
        sessionRouter,
        provider,
        expectedSession: exp,
        user,
        publicClient,
      } = await loadFixture(deploySingleBid);

      // open session
      const { msg, signature } = await getProviderApproval(
        provider,
        user.account.address,
        exp.bidID,
      );
      const openTx = await sessionRouter.write.openSession(
        [exp.stake, msg, signature],
        { account: user.account.address },
      );
      const sessionId = await getSessionId(publicClient, hre, openTx);

      await time.increase(exp.durationSeconds / 2n - 1n);

      // close session with dispute / user report
      const report = await getReport(user, sessionId, 10, 10);
      await sessionRouter.write.closeSession([report.msg, report.sig], {
        account: user.account,
      });

      // verify session is closed with dispute
      const session = await sessionRouter.read.getSession([sessionId]);
      expect(session.closeoutType).to.equal(1n);

      const sessionCost =
        session.pricePerSecond * (session.closedAt - session.openedAt);

      // immediately after claimable balance should be 0
      const claimable = await sessionRouter.read.getProviderClaimableBalance([
        sessionId,
      ]);
      expect(claimable).to.equal(0n);

      // after 24 hours claimable balance should be correct
      await time.increase((24 * HOUR) / SECOND);
      const claimable2 = await sessionRouter.read.getProviderClaimableBalance([
        sessionId,
      ]);
      expect(claimable2).to.equal(sessionCost);
    });
  });
});
