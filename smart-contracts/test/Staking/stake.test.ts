import { getCurrentBlockTime, setTime } from "@/utils/block-helper";
import { getDefaultDurations } from "@/utils/staking-helper";
import { DAY } from "@/utils/time";
import { LumerinToken, MorpheusToken, StakingMasterChef } from "@ethers-v6";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { Reverter } from "../helpers/reverter";

describe("Staking contract - stake", () => {
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

    const startDate =
      BigInt(new Date("2024-07-16T01:00:00.000Z").getTime()) / 1000n;
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

    it("Should stake correctly and emit event", async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const aliceStakeId = await staking
        .connect(ALICE)
        .stake.staticCall(poolId, stakingAmount, lockDuration);
      const tx = await staking
        .connect(ALICE)
        .stake(poolId, stakingAmount, lockDuration);
      const aliceStakeTime = await getCurrentBlockTime();

      ////

      await expect(tx)
        .to.emit(staking, "Staked")
        .withArgs(ALICE, poolId, aliceStakeId, stakingAmount);
    });

    it("should error if pool does not exist", async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const aliceStakeId = await staking
        .connect(ALICE)
        .stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);
      const aliceStakeTime = await getCurrentBlockTime();

      ////

      await expect(staking.stake(100, 1000, 0)).to.be.revertedWithCustomError(
        staking,
        "PoolNotExists",
      );
    });

    it("should error if staking before start date", async () => {
      const now = await getCurrentBlockTime();
      const startTime = now + DAY;
      const duration = 10n * DAY;
      const rewardPerSecond = 100n;
      const totalReward = rewardPerSecond * duration;

      await MOR.approve(staking, totalReward);
      const newPoolId = await staking.addPool.staticCall(
        startTime,
        duration,
        totalReward,
        [DAY],
        [pool.precision],
      );
      await staking.addPool(
        startTime,
        duration,
        totalReward,
        [DAY],
        [pool.precision],
      );

      const stakeAmount = 1000n;
      await LMR.connect(ALICE).approve(staking, stakeAmount);

      await expect(
        staking.connect(ALICE).stake(newPoolId, stakeAmount, 0),
      ).to.be.revertedWithCustomError(staking, "StakingNotStarted");

      await setTime(Number(startTime));

      await staking.connect(ALICE).stake(poolId, stakeAmount, lockDuration);
    });

    it("should error if staking after end date", async () => {
      await setTime(Number(pool.endDate));

      await LMR.connect(ALICE).approve(staking, stakingAmount);

      await expect(
        staking.connect(ALICE).stake(poolId, stakingAmount, 0),
      ).to.be.revertedWithCustomError(staking, "StakingFinished");
    });

    it("Should error if staking duration is too long", async () => {
      await setTime(Number(pool.endDate - DAY));
      await LMR.connect(ALICE).approve(staking, stakingAmount);

      await expect(
        staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration),
      ).to.be.revertedWithCustomError(
        staking,
        "LockReleaseTimePastPoolEndTime",
      );
    });

    it("should error if not enough allowance", async () => {
      await expect(
        staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration),
      ).to.be.revertedWith("ERC20: insufficient allowance");
    });

    it("should error if not enough tokens", async () => {
      const amount = 2_000_000n;
      await LMR.connect(ALICE).approve(staking, amount);
      await expect(
        staking.connect(ALICE).stake(poolId, amount, lockDuration),
      ).to.be.revertedWith("ERC20: transfer amount exceeds balance");
    });
  });
});
