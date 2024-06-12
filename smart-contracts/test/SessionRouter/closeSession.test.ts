import {
  loadFixture,
  time,
} from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import hre from "hardhat";
import { deploySingleBid, getProviderApproval, getReport } from "../fixtures";
import { getSessionId } from "../utils";
import { DAY, SECOND } from "../../utils/time";
import { expectAlmostEqual, expectAlmostEqualDelta } from "../../utils/compare";

describe("session closeout", function () {
  it("should open short (<1D) session and close after expiration", async function () {
    const {
      sessionRouter,
      provider,
      expectedSession: exp,
      user,
      publicClient,
      tokenMOR,
    } = await loadFixture(deploySingleBid);

    // open session
    const { msg, signature } = await getProviderApproval(provider, exp.bidID);
    const openTx = await sessionRouter.write.openSession(
      [exp.stake, msg, signature],
      { account: user.account.address },
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
    const report = await getReport(provider, sessionId, 10);
    await sessionRouter.write.closeSession([report.msg, report.sig], {
      account: user.account,
    });

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

    const totalPrice =
      (session.endsAt - session.openedAt) * session.pricePerSecond;

    expectAlmostEqualDelta(
      0,
      userStakeReturned,
      Number(session.pricePerSecond) * 5,
    );
    expectAlmostEqual(providerEarned, totalPrice, 0.0001);
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
    const { msg, signature } = await getProviderApproval(provider, exp.bidID);
    const openTx = await sessionRouter.write.openSession(
      [exp.stake, msg, signature],
      { account: user.account.address },
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
    const report = await getReport(provider, sessionId, 10);
    await sessionRouter.write.closeSession([report.msg, report.sig], {
      account: user.account,
    });

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

    expectAlmostEqual(userStakeReturned, exp.stake / 2n, 0.001);
    expectAlmostEqual(providerEarned, exp.totalCost / 2n, 0.0001);
  });

  it("should open and close early with user report - dispute", async function () {
    const {
      sessionRouter,
      provider,
      expectedSession: exp,
      user,
      publicClient,
      tokenMOR,
    } = await loadFixture(deploySingleBid);

    // open session
    const { msg, signature } = await getProviderApproval(provider, exp.bidID);
    const openTx = await sessionRouter.write.openSession(
      [exp.stake, msg, signature],
      { account: user.account.address },
    );
    const sessionId = await getSessionId(publicClient, hre, openTx);

    // wait half of the session
    await time.increase(exp.durationSeconds / 2n - 1n);

    const userBalanceBefore = await tokenMOR.read.balanceOf([
      user.account.address,
    ]);
    const providerBalanceBefore = await tokenMOR.read.balanceOf([
      provider.account.address,
    ]);

    // close session with user signature
    const report = await getReport(user, sessionId, 10);
    await sessionRouter.write.closeSession([report.msg, report.sig], {
      account: user.account,
    });

    // verify session is closed with dispute
    const session = await sessionRouter.read.getSession([sessionId]);

    // verify balances
    const userBalanceAfter = await tokenMOR.read.balanceOf([
      user.account.address,
    ]);
    const providerBalanceAfter = await tokenMOR.read.balanceOf([
      provider.account.address,
    ]);

    const claimableProvider =
      await sessionRouter.read.getProviderClaimableBalance([session.id]);

    const [userAvail, userHold] =
      await sessionRouter.read.withdrawableUserStake([session.user]);

    expect(session.closeoutType).to.equal(1n);
    expect(providerBalanceAfter - providerBalanceBefore).to.equal(0n);
    expect(claimableProvider).to.equal(0n);
    expectAlmostEqual(
      exp.stake / 2n,
      userBalanceAfter - userBalanceBefore,
      0.05,
    );
    expect(userAvail).to.equal(0n);
    expectAlmostEqual(userHold, userBalanceAfter - userBalanceBefore, 0.05);

    // verify provider balance after dispute is released
    await time.increase((1 * DAY) / SECOND);
    const claimableProvider2 =
      await sessionRouter.read.getProviderClaimableBalance([sessionId]);
    expect(claimableProvider2).to.equal(exp.totalCost);

    // claim provider balance
    await sessionRouter.write.claimProviderBalance([
      sessionId,
      claimableProvider2,
      exp.provider,
    ]);

    // verify provider balance after claim
    const providerBalanceAfterClaim = await tokenMOR.read.balanceOf([
      provider.account.address,
    ]);
    const providerClaimed = providerBalanceAfterClaim - providerBalanceAfter;
    expect(providerClaimed).to.equal(exp.totalCost);
  });
});
