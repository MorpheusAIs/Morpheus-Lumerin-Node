import { getCurrentBlockTime, setTime } from "@/utils/block-helper";
import { getDefaultDurations } from "@/utils/staking-helper";
import { DAY } from "@/utils/time";
import { LumerinToken, MorpheusToken, StakingMasterChef } from "@ethers-v6";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { time } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { Reverter } from "../helpers/reverter";

describe("Staking contract - Complex reward scenarios", () => {
  const reverter = new Reverter();

  const startDate =
    BigInt(new Date("2024-07-16T01:00:00.000Z").getTime()) / 1000n;
  const stakingAmount = 1000n;
  const lockDuration = 7n * DAY;
  const poolId = 0n;

  let OWNER: SignerWithAddress;
  let ALICE: SignerWithAddress;
  let BOB: SignerWithAddress;
  let CAROL: SignerWithAddress;

  let staking: StakingMasterChef;
  let MOR: MorpheusToken;
  let LMR: LumerinToken;

  let pool: any;

  before("setup", async () => {
    [OWNER, ALICE, BOB, CAROL] = await ethers.getSigners();

    const [StakingMasterChef, ERC1967Proxy, MORFactory, LMRFactory] =
      await Promise.all([
        await ethers.getContractFactory("StakingMasterChef"),
        await ethers.getContractFactory("ERC1967Proxy"),
        await ethers.getContractFactory("MorpheusToken"),
        await ethers.getContractFactory("LumerinToken"),
      ]);

    let stakingImpl;
    [stakingImpl, MOR, LMR] = await Promise.all([
      StakingMasterChef.deploy(),
      MORFactory.deploy(),
      LMRFactory.deploy("Lumerin dev", "LMR"),
    ]);
    const stakingProxy = await ERC1967Proxy.deploy(stakingImpl, "0x");

    staking = StakingMasterChef.attach(
      stakingProxy.target,
    ) as StakingMasterChef;

    await staking.__StakingMasterChef_init(LMR, MOR);

    const duration = 400n * DAY;
    const endDate = startDate + duration;
    const rewardPerSecond = 100n;

    pool = {
      id: 0n,
      rewardPerSecond,
      stakingToken: LMR,
      rewardToken: MOR,
      totalReward: rewardPerSecond * duration,
      lockDurations: getDefaultDurations().durationSeconds,
      multipliersScaled_: getDefaultDurations().multiplierScaled,
      precision: 0n,
      startDate,
      endDate,
      duration,
    };

    await MOR.approve(staking, pool.totalReward);

    await staking.addPool(
      pool.startDate,
      pool.duration,
      pool.totalReward,
      pool.lockDurations,
      pool.multipliersScaled_,
    );

    await LMR.transfer(ALICE, 1_000_000n);
    await LMR.transfer(BOB, 1_000_000n);
    await LMR.transfer(CAROL, 1_000_000n);

    await reverter.snapshot();
  });

  afterEach(reverter.revert);

  describe("Actions", () => {
    beforeEach(async () => {
      await setTime(Number(startDate));
    });

    it("Should reward correctly two stakers", async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const aliceStakeId = await staking
        .connect(ALICE)
        .stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);
      const aliceStakeTime = await getCurrentBlockTime();

      //// bobStakes
      await time.increase(5n * DAY);

      await LMR.connect(BOB).approve(staking, stakingAmount);
      const bobStakeId = await staking
        .connect(BOB)
        .stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(BOB).stake(poolId, stakingAmount, lockDuration);
      const bobStakeTime = await getCurrentBlockTime();

      ////

      await setTime(Number((await getCurrentBlockTime()) + 5n * DAY));

      const aliceMorBalanceBefore = await MOR.balanceOf(ALICE);
      await staking.connect(ALICE).unstake(poolId, aliceStakeId, ALICE);
      const aliceWithdrawTime = await getCurrentBlockTime();
      const aliceMorBalanceAfter = await MOR.balanceOf(ALICE);

      const durationAliceSingle = bobStakeTime - aliceStakeTime;
      const earnAliceSingle = durationAliceSingle * pool.rewardPerSecond;
      const durationAliceDouble = aliceWithdrawTime - bobStakeTime;
      const earnAliceDouble = (durationAliceDouble * pool.rewardPerSecond) / 2n;
      const totalEarnAlice = earnAliceSingle + earnAliceDouble;
      expect(aliceMorBalanceAfter - aliceMorBalanceBefore).to.equal(
        totalEarnAlice,
      );

      await setTime(Number((await getCurrentBlockTime()) + 2n * DAY));

      const bobMorBalanceBefore = await MOR.balanceOf(BOB);
      await staking.connect(BOB).unstake(poolId, aliceStakeId, BOB);
      const bobWithdrawTime = await getCurrentBlockTime();
      const bobMorBalanceAfter = await MOR.balanceOf(BOB);

      const durationBobDouble = aliceWithdrawTime - bobStakeTime;
      const earnBobDouble = (durationBobDouble * pool.rewardPerSecond) / 2n;
      const durationBobSingle = bobWithdrawTime - aliceWithdrawTime;
      const earnBobSingle = durationBobSingle * pool.rewardPerSecond;
      const totalEarnBob = earnBobSingle + earnBobDouble;
      expect(bobMorBalanceAfter - bobMorBalanceBefore).to.equal(totalEarnBob);
    });

    it("should stop increasing reward after pool end date", async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const aliceStakeId = await staking
        .connect(ALICE)
        .stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);

      ////

      const reward1 = await staking.getReward(ALICE, poolId, aliceStakeId);

      await setTime(Number(pool.endDate));
      const reward2 = await staking.getReward(ALICE, poolId, aliceStakeId);

      await setTime(
        Number(
          (await getCurrentBlockTime()) +
            5n * BigInt(pool.endDate - pool.startDate),
        ),
      );
      const reward3 = await staking.getReward(ALICE, poolId, aliceStakeId);

      expect(reward1 < reward2).to.be.true;
      expect(reward2).to.equal(reward3);
    });
  });
});
