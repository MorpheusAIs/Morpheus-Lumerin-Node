import { LumerinToken, MorpheusToken, StakingMasterChef } from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { expect } from 'chai';
import { ethers } from 'hardhat';

import { Reverter } from '../helpers/reverter';

import { getCurrentBlockTime, setTime } from '@/utils/block-helper';
import { getDefaultDurations } from '@/utils/staking-helper';
import { DAY } from '@/utils/time';

describe('Staking contract - withdrawReward', () => {
  const reverter = new Reverter();

  let startDate: bigint;
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

  before('setup', async () => {
    [OWNER, ALICE, BOB, CAROL] = await ethers.getSigners();

    const [StakingMasterChef, ERC1967Proxy, MORFactory, LMRFactory] = await Promise.all([
      await ethers.getContractFactory('StakingMasterChef'),
      await ethers.getContractFactory('ERC1967Proxy'),
      await ethers.getContractFactory('MorpheusToken'),
      await ethers.getContractFactory('LumerinToken'),
    ]);

    let stakingImpl;
    [stakingImpl, MOR, LMR] = await Promise.all([
      StakingMasterChef.deploy(),
      MORFactory.deploy(),
      LMRFactory.deploy('Lumerin dev', 'LMR'),
    ]);
    const stakingProxy = await ERC1967Proxy.deploy(stakingImpl, '0x');

    staking = StakingMasterChef.attach(stakingProxy.target) as StakingMasterChef;

    await staking.__StakingMasterChef_init(LMR, MOR);

    startDate = (await getCurrentBlockTime()) + DAY;
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

    await staking.addPool(pool.startDate, pool.duration, pool.totalReward, pool.lockDurations, pool.multipliersScaled_);

    await LMR.transfer(ALICE, 1_000_000n);
    await LMR.transfer(BOB, 1_000_000n);
    await LMR.transfer(CAROL, 1_000_000n);

    await reverter.snapshot();
  });

  afterEach(reverter.revert);

  describe('Actions', () => {
    beforeEach(async () => {
      await setTime(Number(startDate));
    });

    it('should withdraw reward correctly', async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const aliceStakeId = await staking.connect(ALICE).stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);
      const aliceStakeTime = await getCurrentBlockTime();

      ////

      await setTime(Number((await getCurrentBlockTime()) + lockDuration));

      const tx = await staking.connect(ALICE).withdrawReward(poolId, aliceStakeId, ALICE);
      const rewardTime = await getCurrentBlockTime();

      const stakeDuration = rewardTime - aliceStakeTime;
      const reward = stakeDuration * pool.rewardPerSecond;

      await expect(tx).to.changeTokenBalance(MOR, ALICE, reward);

      await expect(tx).to.emit(staking, 'RewardWithdrawed').withArgs(ALICE, poolId, aliceStakeId, reward);
    });

    it('should error if poolId is wrong', async () => {
      await expect(staking.connect(ALICE).withdrawReward(poolId + 1n, 0, ALICE)).to.be.revertedWithCustomError(
        staking,
        'PoolNotExists',
      );
    });

    it('should error if stakeId is wrong', async () => {
      await expect(staking.connect(ALICE).withdrawReward(poolId, 1, ALICE)).to.be.revertedWithCustomError(
        staking,
        'StakeNotExists',
      );
    });

    it('should error if no reward yet', async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const aliceStakeId = await staking.connect(ALICE).stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);

      await setTime(pool.endDate + 1n);

      await staking.connect(ALICE).withdrawReward(poolId, aliceStakeId, ALICE);
      await expect(staking.connect(ALICE).withdrawReward(poolId, aliceStakeId, ALICE)).to.be.revertedWithCustomError(
        staking,
        'NoRewardAvailable',
      );
    });

    it('should error if rewards already have been withdrawn', async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, 2n * stakingAmount);
      const aliceStakeId = await staking.connect(ALICE).stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);

      await setTime(pool.endDate);

      await staking.connect(ALICE).withdrawReward(poolId, aliceStakeId, ALICE);
      await expect(staking.connect(ALICE).withdrawReward(poolId, aliceStakeId, ALICE)).to.be.revertedWithCustomError(
        staking,
        'StakeUnstaked',
      );
    });
  });
});
