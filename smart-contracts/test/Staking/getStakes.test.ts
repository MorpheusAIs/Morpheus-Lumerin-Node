import { loadFixture, time } from "@nomicfoundation/hardhat-network-helpers";
import { aliceStakes } from "./fixtures";
import { expect } from "chai";
import { catchError, getTxDeltaBalance } from "../utils";
import { DAY, SECOND } from "../../utils/time";
import { elapsedTxs, getStakeId } from "./utils";

describe("Staking contract - getStake", () => {
  it("Should get user stake", async () => {
    const {
      accounts: { alice },
      contracts: { staking, tokenMOR, tokenLMR },
      stakes,
      expPool,
      pubClient,
    } = await loadFixture(aliceStakes);

    const stakingAmount = 1000n;
    const lockDurationId = 0;
    const poolId = 0n;

    // await tokenLMR.write.approve([staking.address, stakingAmount], {
    //   account: alice.account,
    // });
    // const depositTx = await staking.write.stake(
    //   [poolId, stakingAmount, lockDurationId],
    //   { account: alice.account },
    // );

    // const stakeId = await getStakeId(depositTx);

    // const userStake = await staking.read.poolUserStakes([
    //   poolId,
    //   alice.account.address,
    //   stakes.alice.stakeId,
    // ]);

    const userStake = await staking.read.getStake([
      alice.account.address,
      poolId,
      0n,
    ]);

    const userStakes = await staking.read.getStakes([
      alice.account.address,
      poolId,
    ]);

    console.log(userStake);
    console.log(userStakes);
  });
});
