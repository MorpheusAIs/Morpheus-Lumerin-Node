import { getCurrentBlockTime, setTime } from "@/utils/block-helper";
import { getDefaultDurations } from "@/utils/staking-helper";
import { DAY } from "@/utils/time";
import { LumerinToken, MorpheusToken, StakingMasterChef } from "@ethers-v6";
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import { time } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { Reverter } from "../helpers/reverter";

describe("Staking contract - getReward", () => {
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

    it("Should get reward correctly for user that staked", async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const aliceStakeId = await staking
        .connect(ALICE)
        .stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);
      const aliceStakeTime = await getCurrentBlockTime();

      ////

      await time.increase(lockDuration);

      const reward = await staking.connect(ALICE).getReward(
        ALICE,
        poolId,
        aliceStakeId,
        { blockTag: "pending" }, // we need to check reward against the block that will be mined next
      );

      const tx = await staking
        .connect(ALICE)
        .withdrawReward(poolId, aliceStakeId, ALICE);
      const aliceWithdrawTime = await getCurrentBlockTime();

      const stakeDuration = aliceWithdrawTime - aliceStakeTime;

      await expect(tx).to.changeTokenBalance(MOR, ALICE.address, reward);
      expect(reward).to.equal(stakeDuration * pool.rewardPerSecond);
    });

    it("Should not error if stakeId is wrong", async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const aliceStakeId = await staking
        .connect(ALICE)
        .stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);

      ////

      expect(
        await staking.getReward(ALICE, poolId, aliceStakeId + 1n),
      ).to.equal(0n);
    });

    it("Should not error if user has not staked", async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const aliceStakeId = await staking
        .connect(ALICE)
        .stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);

      ////

      expect(await staking.getReward(BOB, poolId, aliceStakeId)).to.equal(0n);
    });

    it("Should not error if pool doesn't exist", async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const aliceStakeId = await staking
        .connect(ALICE)
        .stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);
      const aliceStakeTime = await getCurrentBlockTime();

      ////

      expect(
        await staking.getReward(ALICE, poolId + 1n, aliceStakeId),
      ).to.equal(0n);
    });

    it("Should return 0 if user withdrawn all rewards", async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const aliceStakeId = await staking
        .connect(ALICE)
        .stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);
      const aliceStakeTime = await getCurrentBlockTime();

      ////

      await setTime(Number((await getCurrentBlockTime()) + 10n * DAY));

      const rewardBefore = await staking.getReward(
        ALICE.address,
        poolId,
        aliceStakeId,
      );
      expect(rewardBefore > 0).to.be.true;

      await staking.connect(ALICE).withdrawReward(poolId, aliceStakeId, ALICE);

      const reward = await staking.getReward(
        ALICE.address,
        poolId,
        aliceStakeId,
      );
      expect(reward).to.equal(0n);
    });
  });
});
