import { loadFixture } from "@nomicfoundation/hardhat-toolbox-viem/network-helpers";
import { expect } from "chai";
import { deploySingleBid } from "../fixtures";

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
        balance: 3406521346191960000000n,
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
});
