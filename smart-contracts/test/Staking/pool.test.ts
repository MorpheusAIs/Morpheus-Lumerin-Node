import { LumerinToken, MorpheusToken, StakingMasterChef } from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { expect } from 'chai';
import { ethers } from 'hardhat';

import { Reverter } from '../helpers/reverter';

import { PRECISION } from '@/scripts/utils/constants';
import { getCurrentBlockTime, setTime } from '@/utils/block-helper';
import { getDefaultDurations } from '@/utils/staking-helper';
import { DAY } from '@/utils/time';

describe('Staking contract', () => {
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

  let pool: {
    id: bigint;
    rewardPerSecond: bigint;
    stakingToken: LumerinToken;
    rewardToken: MorpheusToken;
    totalReward: bigint;
    lockDurations: bigint[];
    multipliersScaled_: bigint[];
    precision: bigint;
    startDate: bigint;
    endDate: bigint;
    duration: bigint;
  };

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
      precision: PRECISION,
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

    describe('Add pool', () => {
      it('Should verify adding pool', async () => {
        const poolInfo = await staking.pools(pool.id);
        expect(poolInfo).deep.equal([pool.rewardPerSecond * pool.precision, pool.startDate, pool.endDate, false]);
      });

      it('Should error adding pool if not owner', async () => {
        await MOR.approve(staking, pool.totalReward);

        await expect(
          staking
            .connect(ALICE)
            .addPool(pool.startDate, pool.duration, pool.totalReward, pool.lockDurations, pool.multipliersScaled_),
        ).to.be.revertedWith('Ownable: caller is not the owner');
      });

      it('Should error adding pool if not approved', async () => {
        await expect(
          staking.addPool(pool.startDate, pool.duration, pool.totalReward, pool.lockDurations, pool.multipliersScaled_),
        ).to.be.revertedWithCustomError(staking, 'StartTimeIsPast');
      });

      it('Should error adding pool if not enough balance', async () => {
        const balance = await MOR.balanceOf(OWNER);
        await MOR.transfer(ALICE, balance);

        await MOR.approve(staking, pool.totalReward);

        await expect(
          staking.addPool(pool.startDate, pool.duration, pool.totalReward, pool.lockDurations, pool.multipliersScaled_),
        ).to.be.revertedWithCustomError(staking, 'StartTimeIsPast');
      });

      it('Should error adding pool if `duration_ == 0`', async () => {
        await expect(
          staking.addPool(pool.startDate + DAY, 0, pool.totalReward, pool.lockDurations, pool.multipliersScaled_),
        ).to.be.revertedWithCustomError(staking, 'InvalidDuration');
      });

      it('Should error adding pool if `totalReward_ == 0`', async () => {
        await expect(
          staking.addPool(pool.startDate + DAY, pool.duration, 0, pool.lockDurations, pool.multipliersScaled_),
        ).to.be.revertedWithCustomError(staking, 'InvalidReward');
      });

      it('Should error adding pool if `lockDurations_.length == 0`', async () => {
        await expect(
          staking.addPool(pool.startDate + DAY, pool.duration, pool.totalReward, [], pool.multipliersScaled_),
        ).to.be.revertedWithCustomError(staking, 'InvalidLocksCount');
      });

      it('Should error adding pool if `lockDurations_.length != multipliersScaled_.length`', async () => {
        await expect(
          staking.addPool(pool.startDate + DAY, pool.duration, pool.totalReward, pool.lockDurations, []),
        ).to.be.revertedWithCustomError(staking, 'InvalidLocksCount');
      });
    });

    describe('Stop pool', () => {
      it('Should stop pool', async () => {
        await staking.terminatePool(pool.id, OWNER);
        const timestamp = await getCurrentBlockTime();

        const [, startTime, endTime] = await staking.pools(pool.id);

        expect(startTime).equal(pool.startDate);
        expect(endTime).equal(timestamp);
      });

      it('Should pay back unused reward', async () => {
        //// aliceStakes
        await LMR.connect(ALICE).approve(staking, stakingAmount);
        const aliceStakeId = await staking.connect(ALICE).stake.staticCall(poolId, stakingAmount, lockDuration);
        await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);
        const aliceStakeTime = await getCurrentBlockTime();

        ////

        await setTime(Number((await getCurrentBlockTime()) + DAY));
        const terminateTx = await staking.terminatePool(pool.id, OWNER);
        const stoppedAt = await getCurrentBlockTime();

        const unstakeTx = await staking.connect(ALICE).unstake(pool.id, aliceStakeId, ALICE);

        const expPayback = (pool.endDate - stoppedAt) * pool.rewardPerSecond;
        const expAliceReward = (stoppedAt - aliceStakeTime) * pool.rewardPerSecond;

        await expect(terminateTx).to.changeTokenBalance(MOR, OWNER, expPayback);
        await expect(unstakeTx).to.changeTokenBalance(MOR, ALICE, expAliceReward);
      });

      it('Should error stopping pool if not owner', async () => {
        await expect(staking.connect(ALICE).terminatePool(pool.id, ALICE.address)).to.be.revertedWith(
          'Ownable: caller is not the owner',
        );
      });

      it('Should not be able to stake after pool is stopped', async () => {
        //// aliceStakes
        await LMR.connect(ALICE).approve(staking, stakingAmount);
        await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);

        await staking.terminatePool(pool.id, OWNER);

        await LMR.connect(BOB).approve(staking, stakingAmount);

        await expect(staking.connect(BOB).stake(pool.id, stakingAmount, lockDuration)).to.be.revertedWithCustomError(
          staking,
          'StakingFinished',
        );
      });

      it('Should be able to unstake after pool is stopped', async () => {
        //// aliceStakes
        await LMR.connect(ALICE).approve(staking, stakingAmount);
        const aliceStakeId = await staking.connect(ALICE).stake.staticCall(poolId, stakingAmount, lockDuration);
        await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);
        const aliceStakeTime = await getCurrentBlockTime();

        ////

        await staking.terminatePool(pool.id, OWNER);
        const stopTime = await getCurrentBlockTime();

        const lmrBalanceBefore = await LMR.balanceOf(ALICE);
        const morBalanceBefore = await MOR.balanceOf(ALICE);

        await staking.connect(ALICE).unstake(poolId, aliceStakeId, ALICE);

        const lmrBalanceAfter = await LMR.balanceOf(ALICE);
        const morBalanceAfter = await MOR.balanceOf(ALICE);

        expect(lmrBalanceAfter - lmrBalanceBefore).to.equal(stakingAmount, 'should return staked balance');
        expect(morBalanceAfter - morBalanceBefore).to.equal(
          (stopTime - aliceStakeTime) * pool.rewardPerSecond,
          'should return earned balance',
        );
      });

      it('Should error stopping pool if already stopped (1)', async () => {
        await setTime(Number(pool.endDate + 1n));

        await expect(staking.terminatePool(pool.id, OWNER)).to.be.revertedWithCustomError(
          staking,
          'PoolAlreadyTerminated',
        );
      });

      it('Should error stopping pool if already stopped (2)', async () => {
        await staking.terminatePool(pool.id, OWNER);

        await expect(staking.terminatePool(pool.id, OWNER)).to.be.revertedWithCustomError(
          staking,
          'PoolAlreadyTerminated',
        );
      });

      it('should error if poolId is wrong', async () => {
        await expect(staking.terminatePool(poolId + 1n, OWNER)).to.be.revertedWithCustomError(staking, 'PoolNotExists');
      });
    });

    describe('Staking contract - updatePoolReward', () => {
      it('should update reward manually', async () => {
        //// aliceStakes
        await LMR.connect(ALICE).approve(staking, stakingAmount);
        await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);
        const aliceStakeTime = await getCurrentBlockTime();

        ////

        await setTime(Number((await getCurrentBlockTime()) + DAY));
        const [lastRewardTimeBf, rewardPerShareBf] = await staking.poolRatesData(pool.id);
        await staking.recalculatePoolReward(pool.id);
        const [lastRewardTimeAf, rewardPerShareAf] = await staking.poolRatesData(pool.id);

        expect(aliceStakeTime).to.be.eq(lastRewardTimeBf);
        expect(lastRewardTimeAf > lastRewardTimeBf).to.be.true;
        expect(rewardPerShareAf > rewardPerShareBf).to.be.true;
      });

      it('should error if `poolId` is not exists', async () => {
        await expect(staking.recalculatePoolReward(poolId + 1n)).to.be.revertedWithCustomError(
          staking,
          'PoolNotExists',
        );
      });
    });
  });
});
