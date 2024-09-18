import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { expect } from 'chai';
import { ethers } from 'hardhat';

import { Reverter } from '../helpers/reverter';

import { LumerinToken, MorpheusToken, StakingMasterChefV2 } from '@/generated-types/ethers';
import { StakingMasterChef } from '@/generated-types/ethers/contracts/staking/StakingMasterChef';
import { ZERO_ADDR } from '@/scripts/utils/constants';
import { wei } from '@/scripts/utils/utils';
import { getCurrentBlockTime } from '@/utils/block-helper';
import { getDefaultDurations } from '@/utils/staking-helper';
import { DAY, SECOND } from '@/utils/time';

describe('Staking', () => {
  const reverter = new Reverter();

  let startDate: bigint;

  let OWNER: SignerWithAddress;
  let ALICE: SignerWithAddress;
  let BOB: SignerWithAddress;
  let CAROL: SignerWithAddress;

  let staking: StakingMasterChef;
  let MOR: MorpheusToken;
  let LMR: LumerinToken;

  let pool: any;

  before(async () => {
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

  describe('UUPS proxy functionality', () => {
    describe('#constructor', () => {
      it('should disable initialize function', async () => {
        const reason = 'Initializable: contract is already initialized';

        const staking = await (await ethers.getContractFactory('StakingMasterChef')).deploy();

        await expect(staking.__StakingMasterChef_init(LMR, MOR)).to.be.revertedWith(reason);
      });
    });

    describe('#Distribution_init', () => {
      it('should set correct data after creation', async () => {
        expect(await staking.stakingToken()).to.eq(await LMR.getAddress());
        expect(await staking.rewardToken()).to.eq(await MOR.getAddress());
      });
      it('should revert if try to call init function twice', async () => {
        const reason = 'Initializable: contract is already initialized';

        await expect(staking.__StakingMasterChef_init(LMR, MOR)).to.be.rejectedWith(reason);
      });
    });

    describe('#_authorizeUpgrade', () => {
      it('should correctly upgrade', async () => {
        const stakingV2Factory = await ethers.getContractFactory('StakingMasterChefV2');
        const stakingV2Implementation = await stakingV2Factory.deploy();

        await staking.upgradeTo(await stakingV2Implementation.getAddress());

        const stakingV2 = stakingV2Factory.attach(await staking.getAddress()) as StakingMasterChefV2;

        expect(await stakingV2.version()).to.eq(2);
      });
      it('should revert if caller is not the owner', async () => {
        await expect(staking.connect(ALICE).upgradeTo(ZERO_ADDR)).to.be.revertedWith(
          'Ownable: caller is not the owner',
        );
      });
    });
  });
});
