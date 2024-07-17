import { loadFixture } from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import { openEarlyCloseSession, openSession } from "../fixtures";
import { catchError } from "../utils";
import { getAddress, zeroAddress } from "viem";

describe("Session router", function () {
  describe("session write functions", function () {
    it("should block erase if session not closed", async function () {
      const { sessionRouter, user, sessionId } = await loadFixture(openSession);

      // check history
      const sessionIds = await sessionRouter.read.getSessionsByUser([
        user.account.address,
        0n,
        10,
      ]);
      expect(sessionIds.length).to.equal(1);
      expect(sessionIds[0].id).to.equal(sessionId);

      // erase history fails
      await catchError(sessionRouter.abi, "SessionNotClosed", async () => {
        await sessionRouter.write.deleteHistory([sessionId], {
          account: user.account,
        });
      });
    });

    it("erase history", async function () {
      const { sessionRouter, user, sessionId } = await loadFixture(
        openEarlyCloseSession,
      );

      // check history
      const sessionIds = await sessionRouter.read.getSessionsByUser([
        user.account.address,
        0n,
        10,
      ]);
      expect(sessionIds.length).to.equal(1);
      expect(sessionIds[0].id).to.equal(sessionId);
      expect(sessionIds[0].user).to.equal(getAddress(user.account.address));

      // erase history
      await sessionRouter.write.deleteHistory([sessionId], {
        account: user.account,
      });

      const session = await sessionRouter.read.getSession([sessionId]);
      expect(session.user).to.equal(zeroAddress);

      // TODO: fix history so user is not exposed using getSessionsByUser
      // const sessionIds2 = await sessionRouter.read.getSessionsByUser([
      //   user.account.address,
      //   0n,
      //   10,
      // ]);
      // expect(sessionIds2.length).to.equal(0);
    });
  });
});
