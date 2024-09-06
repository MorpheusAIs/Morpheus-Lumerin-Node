import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { aliceAndBobStake } from "./fixtures";
import { expect } from "chai";
import { getTxTimestamp } from "../utils";
import { getStakeId } from "./utils";

describe("Staking contract - getStake", () => {
  it("Should get user stake", async () => {
    const {
      accounts: { alice, bob },
      contracts: { staking, tokenLMR },
      stakes,
      expPool,
      pubClient,
    } = await loadFixture(aliceAndBobStake);

    const stakingAmount = 1000n;
    const lockDurationId = 0;
    const poolId = 0n;

    await tokenLMR.write.approve([staking.address, stakingAmount], {
      account: alice.account,
    });
    const depositTx = await staking.write.stake(
      [poolId, stakingAmount, lockDurationId],
      { account: alice.account },
    );

    const stakeId = await getStakeId(depositTx);

    const userStake = await staking.read.getStake([
      alice.account.address,
      poolId,
      expPool.id,
    ]);

    const lockEndsAt =
      (await getTxTimestamp(pubClient, stakes.alice.depositTx)) +
      expPool.lockDurations[stakes.alice.lockDurationId].durationSeconds;

    expect(userStake).to.deep.equal({
      stakeAmount: stakes.alice.stakingAmount,
      shareAmount: stakes.alice.stakingAmount,
      rewardDebt: 0n,
      lockEndsAt: lockEndsAt,
      stakedAt: await getTxTimestamp(pubClient, stakes.alice.depositTx),
    });

    const aliceStakes = await staking.read.getStakes([
      alice.account.address,
      poolId,
    ]);

    expect(aliceStakes.length).to.equal(2);
    expect(aliceStakes[0].stakeAmount).equal(stakes.alice.stakingAmount);
    expect(aliceStakes[1].stakeAmount).equal(stakingAmount);

    const bobStakes = await staking.read.getStakes([
      bob.account.address,
      stakes.bob.poolId,
    ]);

    expect(bobStakes.length).to.equal(1);
    expect(bobStakes[Number(stakes.bob.stakeId)].stakeAmount).equal(
      stakes.bob.stakingAmount,
    );
  });
});
