import {
  loadFixture,
  time,
} from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import hre from "hardhat";
import { keccak256 } from "viem";
import { deploySingleBid, getProviderApproval, getReport } from "../fixtures";
import { getSessionId } from "../utils";
import { DAY, HOUR, SECOND } from "../../utils/time";
import { expectAlmostEqual } from "../../utils/compare";

describe("session closeout", function () {
  it("should open short (<1D) session and close later", async function () {
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
    const { msg, signature } = await getProviderApproval(provider, exp.bidID);
    const openTx = await sessionRouter.write.openSession(
      [exp.stake, msg, signature],
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

    expectAlmostEqual(userStakeReturned, exp.stake / 2n, 0.0001);
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

    return;
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
