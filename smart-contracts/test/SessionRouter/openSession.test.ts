import {
  loadFixture,
  time,
} from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import hre from "hardhat";
import { deploySingleBid, getProviderApproval } from "../fixtures";
import {
  NewDate,
  catchError,
  getHex,
  getSessionId,
  getTxTimestamp,
  nowChain,
  randomBytes,
  randomBytes32,
  startOfTheDay,
} from "../utils";
import { DAY, HOUR, MINUTE, SECOND } from "../../utils/time";
import { UnknownRpcError, formatUnits } from "viem";

describe("session actions", function () {
  describe("positive cases", function () {
    it("should open session without error", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        publicClient,
        provider,
      } = await loadFixture(deploySingleBid);

      const { msg, signature } = await getProviderApproval(provider, exp.bidID);
      const openTx = await sessionRouter.write.openSession(
        [exp.stake, msg, signature],
        { account: user.account },
      );

      const sessionId = await getSessionId(publicClient, hre, openTx);
      expect(sessionId).to.be.a("string");
    });

    it("should emit SessionOpened event", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        provider,
        publicClient,
      } = await loadFixture(deploySingleBid);

      const { msg, signature } = await getProviderApproval(provider, exp.bidID);
      const hash = await sessionRouter.write.openSession(
        [exp.stake, msg, signature],
        { account: user.account },
      );

      await publicClient.waitForTransactionReceipt({ hash });
      const ev = await sessionRouter.getEvents.SessionOpened({
        providerId: exp.provider,
        userAddress: exp.user,
      });
      expect(ev.length).to.eq(1);
    });

    it("should verify session fields after opening", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        publicClient,
        provider,
      } = await loadFixture(deploySingleBid);

      const { msg, signature } = await getProviderApproval(provider, exp.bidID);
      const openTx = await sessionRouter.write.openSession(
        [exp.stake, msg, signature],
        { account: user.account },
      );

      const sessionId = await getSessionId(publicClient, hre, openTx);
      const session = await sessionRouter.read.getSession([sessionId]);
      const createdAt = await getTxTimestamp(publicClient, openTx);

      expect(session).to.deep.include({
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
        // endsAt: verified in session end time tests
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
        provider,
      } = await loadFixture(deploySingleBid);

      const srBefore = await tokenMOR.read.balanceOf([sessionRouter.address]);
      const userBefore = await tokenMOR.read.balanceOf([user.account.address]);

      const { msg, signature } = await getProviderApproval(provider, exp.bidID);
      const txHash = await sessionRouter.write.openSession(
        [exp.stake, msg, signature],
        { account: user.account },
      );
      await publicClient.waitForTransactionReceipt({ hash: txHash });

      const srAfter = await tokenMOR.read.balanceOf([sessionRouter.address]);
      const userAfter = await tokenMOR.read.balanceOf([user.account.address]);

      expect(srAfter - srBefore).to.equal(exp.stake);
      expect(userBefore - userAfter).to.equal(exp.stake);
    });
  });

  describe("negative cases", function () {
    it("should error when approval bytes is invalid abi data", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        provider,
      } = await loadFixture(deploySingleBid);
      try {
        await sessionRouter.write.openSession([exp.stake, "0x0", "0x0"], {
          account: user.account,
        });
      } catch (e) {
        expect((e as Error).cause).to.be.an.instanceOf(UnknownRpcError);
      }
    });

    it("should error when approval expired", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        provider,
      } = await loadFixture(deploySingleBid);

      const { msg, signature } = await getProviderApproval(provider, exp.bidID);
      const ttl = await sessionRouter.read.SIGNATURE_TTL();
      await time.increase(ttl + 1);

      await catchError(sessionRouter.abi, "SignatureExpired", async () => {
        await sessionRouter.write.openSession([exp.stake, msg, signature], {
          account: user.account,
        });
      });
    });

    it("should error when bid not exist", async function () {
      const {
        sessionRouter,
        user,
        expectedSession: exp,
        provider,
      } = await loadFixture(deploySingleBid);

      const { msg, signature } = await getProviderApproval(
        provider,
        randomBytes32(),
      );
      await catchError(sessionRouter.abi, "BidNotFound", async () => {
        await sessionRouter.write.openSession([exp.stake, msg, signature], {
          account: user.account,
        });
      });
    });

    it("should error when bid is deleted", async function () {
      const {
        sessionRouter,
        user,
        expectedSession: exp,
        marketplace,
        provider,
      } = await loadFixture(deploySingleBid);

      await marketplace.write.deleteModelAgentBid([exp.bidID], {
        account: provider.account,
      });

      const { msg, signature } = await getProviderApproval(provider, exp.bidID);
      await catchError(sessionRouter.abi, "BidNotFound", async () => {
        await sessionRouter.write.openSession([exp.stake, msg, signature], {
          account: user.account,
        });
      });
    });

    it("should error when signature has invalid length", async function () {
      const {
        sessionRouter,
        user,
        expectedSession: exp,
        provider,
      } = await loadFixture(deploySingleBid);

      const { msg } = await getProviderApproval(provider, exp.bidID);
      await catchError(
        sessionRouter.abi,
        "ECDSAInvalidSignatureLength",
        async () => {
          await sessionRouter.write.openSession([exp.stake, msg, "0x0"], {
            account: user.account,
          });
        },
      );
    });

    it("should error when signature is invalid", async function () {
      const {
        sessionRouter,
        user,
        expectedSession: exp,
        provider,
      } = await loadFixture(deploySingleBid);

      const { msg } = await getProviderApproval(provider, exp.bidID);
      const sig = randomBytes(65);

      await catchError(
        sessionRouter.abi,
        ["ECDSAInvalidSignatureS", "ECDSAInvalidSignature"],
        async () => {
          await sessionRouter.write.openSession([exp.stake, msg, sig], {
            account: user.account,
          });
        },
      );
    });

    it("should error when opening two bids with same signature", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        provider,
        approveUserFunds,
      } = await loadFixture(deploySingleBid);

      const { msg, signature } = await getProviderApproval(provider, exp.bidID);
      await sessionRouter.write.openSession([exp.stake, msg, signature], {
        account: user.account.address,
      });

      await approveUserFunds(exp.stake);

      await catchError(sessionRouter.abi, "DuplicateApproval", async () => {
        await sessionRouter.write.openSession([exp.stake, msg, signature], {
          account: user.account.address,
        });
      });
    });

    it("should not error when opening two bids same time", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        provider,
        approveUserFunds,
      } = await loadFixture(deploySingleBid);

      const appr1 = await getProviderApproval(provider, exp.bidID);
      await sessionRouter.write.openSession(
        [exp.stake, appr1.msg, appr1.signature],
        {
          account: user.account.address,
        },
      );

      await approveUserFunds(exp.stake);
      const appr2 = await getProviderApproval(provider, exp.bidID);
      await sessionRouter.write.openSession(
        [exp.stake, appr2.msg, appr2.signature],
        {
          account: user.account.address,
        },
      );
    });

    it("should error with insufficient allowance", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        provider,
        tokenMOR,
      } = await loadFixture(deploySingleBid);

      const { msg, signature } = await getProviderApproval(provider, exp.bidID);
      await catchError(tokenMOR.abi, "ERC20InsufficientAllowance", async () => {
        await sessionRouter.write.openSession(
          [exp.stake * 2n, msg, signature],
          {
            account: user.account,
          },
        );
      });
    });

    it("should error with insufficient allowance", async function () {
      const {
        sessionRouter,
        expectedSession: exp,
        user,
        provider,
        tokenMOR,
      } = await loadFixture(deploySingleBid);

      const stake =
        (await tokenMOR.read.balanceOf([user.account.address])) + 1n;
      await tokenMOR.write.approve([sessionRouter.address, stake], {
        account: user.account,
      });

      const { msg, signature } = await getProviderApproval(provider, exp.bidID);
      await catchError(tokenMOR.abi, "ERC20InsufficientBalance", async () => {
        await sessionRouter.write.openSession([stake, msg, signature], {
          account: user.account,
        });
      });
    });
  });
});

describe("verify session end time", function () {
  it("session that doesn't span across midnight (1h)", async function () {
    const {
      sessionRouter,
      expectedSession: exp,
      user,
      publicClient,
      getStake,
      provider,
    } = await loadFixture(deploySingleBid);

    const durationSeconds = BigInt(HOUR / SECOND);
    const stake = await getStake(durationSeconds, exp.pricePerSecond);

    const { msg, signature } = await getProviderApproval(provider, exp.bidID);
    const txHash = await sessionRouter.write.openSession(
      [stake, msg, signature],
      { account: user.account },
    );

    const sessionId = await getSessionId(publicClient, hre, txHash);
    const session = await sessionRouter.read.getSession([sessionId]);
    const createdAt = await getTxTimestamp(publicClient, txHash);

    expect(session.endsAt).to.equal(createdAt + exp.durationSeconds);
  });

  it("session that spans across midnight (6h) should last 6h", async function () {
    const {
      sessionRouter,
      expectedSession: exp,
      user,
      publicClient,
      getStake,
      approveUserFunds,
      provider,
    } = await loadFixture(deploySingleBid);

    const tomorrow9pm =
      startOfTheDay(await nowChain()) +
      BigInt(DAY / SECOND) +
      21n * BigInt(HOUR / SECOND);
    await time.increaseTo(tomorrow9pm); // 9pm

    // the stake is enough to cover the first day (3h till midnight) and the next day (< 6h)
    const durationSeconds = 6n * BigInt(HOUR / SECOND);
    const stake = await getStake(durationSeconds, exp.pricePerSecond);
    await approveUserFunds(stake);

    const { msg, signature } = await getProviderApproval(provider, exp.bidID);
    const txHash = await sessionRouter.write.openSession(
      [stake, msg, signature],
      { account: user.account },
    );

    await time.increase((3 * HOUR) / SECOND + 1);

    const sessionId = await getSessionId(publicClient, hre, txHash);
    const session = await sessionRouter.read.getSession([sessionId]);
    const createdAt = await getTxTimestamp(publicClient, txHash);

    const endsAt = NewDate(session.endsAt);
    const expEndsAt = NewDate(createdAt + durationSeconds);

    expect(endsAt.getTime()).approximately(expEndsAt.getTime(), 10 * SECOND);
  });

  it("session that lasts multiple days", async function () {
    const {
      sessionRouter,
      expectedSession: exp,
      user,
      publicClient,
      approveUserFunds,
      provider,
    } = await loadFixture(deploySingleBid);

    const midnight = startOfTheDay(await nowChain()) + BigInt(DAY / SECOND);
    await time.increaseTo(midnight);

    // the stake is enough to cover the whole day + extra 1h
    const durationSeconds = 25n * BigInt(HOUR / SECOND);
    const stake = await sessionRouter.read.stipendToStake([
      durationSeconds * exp.pricePerSecond,
      await nowChain(),
    ]);

    await approveUserFunds(stake);

    const { msg, signature } = await getProviderApproval(provider, exp.bidID);
    const txHash = await sessionRouter.write.openSession(
      [stake, msg, signature],
      { account: user.account },
    );

    const sessionId = await getSessionId(publicClient, hre, txHash);
    const session = await sessionRouter.read.getSession([sessionId]);
    const durSeconds = Number(session.endsAt - session.openedAt);

    expect(durSeconds).to.equal(1 * (DAY / SECOND));
  });
});
