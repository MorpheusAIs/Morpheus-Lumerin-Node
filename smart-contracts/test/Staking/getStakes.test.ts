import { LumerinToken, MorpheusToken, StakingMasterChef } from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { time } from '@nomicfoundation/hardhat-network-helpers';
import { expect } from 'chai';
import { ethers } from 'hardhat';

import { Reverter } from '../helpers/reverter';

import { getCurrentBlockTime, setTime } from '@/utils/block-helper';
import { getDefaultDurations } from '@/utils/staking-helper';
import { DAY } from '@/utils/time';

describe('Staking contract - getStake', () => {
  const reverter = new Reverter();

  const startDate = BigInt(new Date('2024-07-16T01:00:00.000Z').getTime()) / 1000n;
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

    const startDate = BigInt(new Date('2024-07-16T01:00:00.000Z').getTime()) / 1000n;
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

    it('Should get user stake', async () => {
      //// aliceStakes
      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const aliceStakeId = await staking.connect(ALICE).stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);
      const aliceStakeTime = await getCurrentBlockTime();

      //// bobStakes
      await time.increase(5n * DAY);

      await LMR.connect(BOB).approve(staking, stakingAmount);
      const bobStakeId = await staking.connect(BOB).stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(BOB).stake(poolId, stakingAmount, lockDuration);
      const bobStakeTime = await getCurrentBlockTime();

      ////

      await LMR.connect(ALICE).approve(staking, stakingAmount);
      const stakeId = await staking.connect(ALICE).stake.staticCall(poolId, stakingAmount, lockDuration);
      await staking.connect(ALICE).stake(poolId, stakingAmount, lockDuration);

      const userStake = await staking.poolUserStakes(poolId, ALICE, pool.id);

      const lockEndsAt = aliceStakeTime + lockDuration;

      expect(userStake).to.deep.equal([stakingAmount, stakingAmount, 0n, lockEndsAt]);

      const aliceStake0 = await staking.poolUserStakes(poolId, ALICE, 0);
      const aliceStake1 = await staking.poolUserStakes(poolId, ALICE, 1);

      expect(aliceStake0.stakeAmount).equal(stakingAmount);
      expect(aliceStake1.stakeAmount).equal(stakingAmount);

      const bobStake0 = await staking.poolUserStakes(poolId, BOB, 0);

      expect(bobStake0.stakeAmount).equal(stakingAmount);
    });
  });
});
