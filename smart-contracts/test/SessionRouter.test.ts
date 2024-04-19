import { loadFixture, time } from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import hre from "hardhat";
import { PublicClient, getAddress, keccak256, parseEventLogs } from "viem";
import { deploySessionRouter, encodedReport } from "./fixtures";
import { expectError, getHex, randomBytes32 } from "./utils";
import { DAY, HOUR, MINUTE, SECOND } from "../utils/time";
import { expectAlmostEqual } from "../utils/compare";

describe("Session router", function () {
  describe("Deployment", function () {
    it("Should set the right owner", async function () {
      const { sessionRouter, owner } = await loadFixture(deploySessionRouter);

      expect(await sessionRouter.read.owner()).to.equal(getAddress(owner.account.address));
    });

    it("Should set the right token", async function () {
      const { sessionRouter, tokenMOR } = await loadFixture(deploySessionRouter);

      expect(await sessionRouter.read.token()).to.equal(getAddress(tokenMOR.address));
    });

    it("Should set registries correctly", async function () {
      const { sessionRouter, marketplace, staking } = await loadFixture(deploySessionRouter);

      expect(await sessionRouter.read.marketplace()).eq(getAddress(marketplace.address));
      expect(await sessionRouter.read.stakingDailyStipend()).eq(getAddress(staking.address));
    });
  });

  describe("session actions", function () {
    it("should open session", async function () {
      const { sessionRouter, staking, expectedBid, user } = await loadFixture(deploySessionRouter);
      const budget = expectedBid.amount * BigInt(HOUR / SECOND);

      await sessionRouter.write.openSession([expectedBid.id, budget], {
        account: user.account,
      });
    });

    it("should fail to open session with invalid bid", async function () {
      const { sessionRouter, user } = await loadFixture(deploySessionRouter);

      try {
        await sessionRouter.write.openSession([randomBytes32(), 0n], {
          account: user.account,
        });
        expect.fail("Should have thrown an error");
      } catch (error) {
        expectError(error, sessionRouter.abi, "BidNotFound");
      }
    });

    it("should fail to open session with duration less than minimum", async function () {
      const { sessionRouter, staking, expectedBid, expectedStake, user } = await loadFixture(
        deploySessionRouter
      );

      await staking.write.unstake(
        [user.account.address, expectedStake.stakeAmount, user.account.address],
        {
          account: user.account,
        }
      );

      try {
        await sessionRouter.write.openSession([expectedBid.id, 0n], {
          account: user.account,
        });
        expect.fail("Should have thrown an error");
      } catch (error) {
        expectError(error, sessionRouter.abi, "SessionTooShort");
      }
    });

    it("should not open session with same bid simultaneously");

    it("should open session with same bid after previous session is closed");

    it("should emit session opened with session id", async function () {
      const { sessionRouter, provider, expectedBid, staking, user, publicClient, tokenMOR } =
        await loadFixture(deploySessionRouter);

      const budget = expectedBid.amount * BigInt(HOUR / SECOND);
      const openSession = await sessionRouter.write.openSession([expectedBid.id, budget], {
        account: user.account.address,
      });
      const receipt = await publicClient.waitForTransactionReceipt({ hash: openSession });
      const events = parseEventLogs({
        abi: sessionRouter.abi,
        logs: receipt.logs,
        eventName: "SessionOpened",
      });
      const [event] = events;

      expect(events.length).to.equal(1);
      expect(event.args.userAddress).to.equal(getAddress(user.account.address));
      expect(event.args.providerId).to.equal(getAddress(provider.account.address));

      const sessionId = await getSessionId(publicClient, openSession);
      expect(sessionId).to.equal(event.args.sessionId);
    });

    it("should open and close early", async function () {
      const { sessionRouter, provider, expectedBid, staking, user, publicClient, tokenMOR } =
        await loadFixture(deploySessionRouter);
      const budget = expectedBid.amount * BigInt(HOUR / SECOND);

      // save balance before opening session
      const balanceBeforeOpen = await staking.read.balanceOfDailyStipend([user.account.address]);
      const providerBalanceBefore = await tokenMOR.read.balanceOf([provider.account.address]);

      // open session
      const openTx = await sessionRouter.write.openSession([expectedBid.id, budget], {
        account: user.account.address,
      });
      const sessionId = await getSessionId(publicClient, openTx);

      await time.increase((30 * MINUTE) / SECOND);
      const balanceBeforeClose = await staking.read.balanceOfDailyStipend([user.account.address]);

      // close session
      const signature = await provider.signMessage({
        message: { raw: keccak256(encodedReport) },
      });
      await sessionRouter.write.closeSession([sessionId, encodedReport, signature], {
        account: user.account,
      });

      // verify session is closed without dispute
      const session = await sessionRouter.read.getSession([sessionId]);
      expect(session.closeoutType).to.equal(0n);

      // verify balances
      const balanceAfterClose = await staking.read.balanceOfDailyStipend([user.account.address]);
      const providerBalanceAfter = await tokenMOR.read.balanceOf([provider.account.address]);

      const stipendLocked = balanceBeforeOpen - balanceBeforeClose;
      const stipendSpent = balanceBeforeOpen - balanceAfterClose;
      const providerEarned = providerBalanceAfter - providerBalanceBefore;

      expect(stipendSpent).to.equal(providerEarned);
      expectAlmostEqual(stipendLocked / 2n, stipendSpent, 0.05);
    });

    it("should open and close with user report - dispute", async function () {
      const { sessionRouter, provider, expectedBid, staking, user, publicClient, tokenMOR } =
        await loadFixture(deploySessionRouter);
      const budget = expectedBid.amount * BigInt(HOUR / SECOND);

      // save balance before opening session
      const balanceBeforeOpen = await staking.read.balanceOfDailyStipend([user.account.address]);
      const providerBalanceBefore = await tokenMOR.read.balanceOf([provider.account.address]);

      // open session
      const openTx = await sessionRouter.write.openSession([expectedBid.id, budget], {
        account: user.account.address,
      });
      const sessionId = await getSessionId(publicClient, openTx);

      await time.increase((30 * MINUTE) / SECOND);
      const balanceBeforeClose = await staking.read.balanceOfDailyStipend([user.account.address]);

      // close session with invalid signature
      const signature = getHex(Buffer.from(""), 0);
      await sessionRouter.write.closeSession([sessionId, encodedReport, signature], {
        account: user.account,
      });

      // verify session is closed with dispute
      const session = await sessionRouter.read.getSession([sessionId]);
      expect(session.closeoutType).to.equal(1n);

      // verify balances
      const balanceAfterClose = await staking.read.balanceOfDailyStipend([user.account.address]);
      const providerBalanceAfter = await tokenMOR.read.balanceOf([provider.account.address]);
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
      await sessionRouter.write.claimProviderBalance([total2, provider.account.address], {
        account: provider.account.address,
      });

      const [total3, onHold3] = await sessionRouter.read.getProviderBalance([
        provider.account.address,
      ]);
      expect(total3).to.equal(0n);
      expect(onHold3).to.equal(0n);

      // verify provider balance after claim
      const providerBalanceAfterClaim = await tokenMOR.read.balanceOf([provider.account.address]);
      const providerClaimed = providerBalanceAfterClaim - providerBalanceAfter;
      expect(providerClaimed).to.equal(total2);
    });
  });
});

async function getSessionId(
  publicClient: PublicClient,
  txHash: `0x${string}`
): Promise<`0x${string}`> {
  const receipt = await publicClient.waitForTransactionReceipt({ hash: txHash });
  const artifact = await hre.artifacts.readArtifact("SessionRouter");
  const [event] = parseEventLogs({
    abi: artifact.abi,
    logs: receipt.logs,
    eventName: "SessionOpened",
  });
  return event.args.sessionId;
}