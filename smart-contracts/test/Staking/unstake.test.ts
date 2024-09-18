import { LumerinToken, MorpheusToken, StakingMasterChef } from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { expect } from 'chai';
import { ethers } from 'hardhat';

import { Reverter } from '../helpers/reverter';

import { getCurrentBlockTime, setTime } from '@/utils/block-helper';
import { getDefaultDurations } from '@/utils/staking-helper';
import { DAY } from '@/utils/time';

describe('Staking contract - unstake', () => {
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

    it('Should unstake correctly', async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const aliceStakeId = await staking.connect(ALICE).stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);
      const aliceStakeTime = await getCurrentBlockTime();

      ////

      await setTime(Number((await getCurrentBlockTime()) + lockDuration));

      const morBalanceBefore = await MOR.balanceOf(ALICE);
      const tx = await staking.connect(ALICE).unstake(poolId, aliceStakeId, ALICE);
      const withdrawTime = await getCurrentBlockTime();
      const morBalanceAfter = await MOR.balanceOf(ALICE);
      const stakeDuration = withdrawTime - aliceStakeTime;

      expect(morBalanceAfter - morBalanceBefore).to.equal(stakeDuration * pool.rewardPerSecond);

      await expect(tx).to.emit(staking, 'Unstaked').withArgs(ALICE, poolId, aliceStakeId, stakingAmount);
    });

    it('Should error if pool does not exist', async () => {
      await expect(staking.connect(ALICE).unstake(poolId + 1n, 0, ALICE)).to.be.revertedWithCustomError(
        staking,
        'PoolNotExists',
      );
    });

    it('Should error if stake does not exist', async () => {
      await expect(staking.connect(ALICE).unstake(poolId, 1, ALICE)).to.be.revertedWithCustomError(
        staking,
        'StakeNotExists',
      );
    });

    it('Should error if unstaking before lock duration', async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const aliceStakeId = await staking.connect(ALICE).stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);

      await expect(staking.connect(ALICE).unstake(poolId, aliceStakeId, ALICE)).to.be.revertedWithCustomError(
        staking,
        'LockNotEnded',
      );
    });

    it('Should error if nothing to unstake', async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, 2n * stakingAmount);
      const aliceStakeId = await staking.connect(ALICE).stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);

      await setTime(Number((await getCurrentBlockTime()) + lockDuration));

      await staking.connect(ALICE).unstake(poolId, aliceStakeId, ALICE);

      await expect(staking.connect(ALICE).unstake(poolId, aliceStakeId, ALICE)).to.be.revertedWithCustomError(
        staking,
        'StakeUnstaked',
      );
    });
  });
});
