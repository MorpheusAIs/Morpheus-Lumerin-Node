import { expect } from "chai";
import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { openEarlyCloseSession } from "../fixtures";
import { DAY, HOUR, SECOND } from "../../utils/time";
import { catchError, randomAddress } from "../utils";
import { expectAlmostEqual } from "../../utils/compare";

describe("User on hold tests", () => {
  it("user stake should be locked right after closeout", async () => {
    const { sessionRouter, user, expectedOnHold } = await loadFixture(
      openEarlyCloseSession,
    );

    // right after closeout
    const [available, onHold] = await sessionRouter.read.withdrawableUserStake([
      user.account.address,
    ]);
    expect(available).to.equal(0n);
    expectAlmostEqual(onHold, expectedOnHold, 0.01);
  });

  it("user stake should be locked before the next day", async () => {
    const { sessionRouter, user, expectedOnHold } = await loadFixture(
      openEarlyCloseSession,
    );

    // before next day
    await time.increaseTo(startOfTomorrow(await time.latest()) - HOUR / SECOND);
    const [available3, onHold3] =
      await sessionRouter.read.withdrawableUserStake([user.account.address]);
    expect(available3).to.equal(0n);
    expectAlmostEqual(onHold3, expectedOnHold, 0.01);
  });

  it("user stake should be available on the next day and withdrawable", async () => {
    const { sessionRouter, user, expectedOnHold, tokenMOR } = await loadFixture(
      openEarlyCloseSession,
    );

    await time.increaseTo(startOfTomorrow(await time.latest()));
    const [available2, onHold2] =
      await sessionRouter.read.withdrawableUserStake([user.account.address]);
    expectAlmostEqual(available2, expectedOnHold, 0.01);
    expect(onHold2).to.equal(0n);

    const randomAddr = randomAddress();
    await sessionRouter.write.withdrawUserStake([available2, randomAddr], {
      account: user.account.address,
    });
    const balance = await tokenMOR.read.balanceOf([randomAddr]);
    expectAlmostEqual(balance, expectedOnHold, 0.01);
  });

  it("user shouldn't be able to withdraw more than there is available stake", async () => {
    const { sessionRouter, user } = await loadFixture(openEarlyCloseSession);

    await time.increaseTo(startOfTomorrow(await time.latest()));
    const [available2] = await sessionRouter.read.withdrawableUserStake([
      user.account.address,
    ]);

    // check that user can't withdraw twice
    await catchError(
      sessionRouter.abi,
      "NotEnoughWithdrawableBalance",
      async () => {
        await sessionRouter.write.withdrawUserStake(
          [available2 * 2n, randomAddress()],
          {
            account: user.account.address,
          },
        );
      },
    );
  });
});

function startOfTomorrow(epochSeconds: number): number {
  const startOfToday = epochSeconds - (epochSeconds % (DAY / SECOND));
  return startOfToday + DAY / SECOND;
}
