import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { deployStaking } from "./fixtures";
import { expect } from "chai";
import { formatUnits, getAddress, parseUnits } from "viem";
import { DAY, SECOND } from "../utils/time";

describe("Staking rewards", function () {
  it("should verify deployment of staking rewards contract", async () => {
    const { staking, tokenLMR, tokenMOR } = await loadFixture(deployStaking);

    expect(await staking.read.stakingToken()).to.eq(
      getAddress(tokenLMR.address),
    );
    expect(await staking.read.rewardsToken()).to.eq(
      getAddress(tokenMOR.address),
    );
  });

  it.only("should add rewards", async () => {
    const { staking, tokenMOR, tokenLMR, user, decimalsMOR, decimalsLMR } =
      await loadFixture(deployStaking);

    const stake = 100n * 10n ** BigInt(decimalsLMR);

    await tokenLMR.write.approve([staking.address, stake], {
      account: user.account.address,
    });
    await staking.write.stake([stake], { account: user.account.address });
    await time.increase(DAY / SECOND);

    await staking.write.getReward({ account: user.account.address });
    const dayReward = await tokenMOR.read.balanceOf([user.account.address]);
    console.log("dayReward:", formatUnits(dayReward, decimalsMOR));
  });
});
