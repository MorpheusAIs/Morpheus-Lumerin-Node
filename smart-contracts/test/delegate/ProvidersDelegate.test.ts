import {
  LumerinDiamond,
  Marketplace,
  ModelRegistry,
  MorpheusToken,
  ProviderRegistry,
  ProvidersDelegate,
  SessionRouter,
} from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { expect } from 'chai';
import { ethers } from 'hardhat';

import { payoutStart } from '../helpers/pool-helper';

import { DelegateRegistry } from '@/generated-types/ethers/contracts/mock/delegate-registry/src';
import { ZERO_ADDR } from '@/scripts/utils/constants';
import { getHex, wei } from '@/scripts/utils/utils';
import {
  deployDelegateRegistry,
  deployFacetDelegation,
  deployFacetMarketplace,
  deployFacetModelRegistry,
  deployFacetProviderRegistry,
  deployFacetSessionRouter,
  deployLumerinDiamond,
  deployMORToken,
  deployProvidersDelegate,
} from '@/test/helpers/deployers';
import { Reverter } from '@/test/helpers/reverter';
import { setTime } from '@/utils/block-helper';
import { getProviderApproval, getReceipt } from '@/utils/provider-helper';
import { DAY, YEAR } from '@/utils/time';

describe('ProvidersDelegate', () => {
  const reverter = new Reverter();

  let OWNER: SignerWithAddress;
  let DELEGATOR: SignerWithAddress;
  let TREASURY: SignerWithAddress;
  let KYLE: SignerWithAddress;
  let SHEV: SignerWithAddress;
  let ALAN: SignerWithAddress;

  let diamond: LumerinDiamond;
  let providerRegistry: ProviderRegistry;
  let modelRegistry: ModelRegistry;
  let providersDelegate: ProvidersDelegate;
  let marketplace: Marketplace;
  let sessionRouter: SessionRouter;

  let token: MorpheusToken;
  let delegateRegistry: DelegateRegistry;

  before(async () => {
    [OWNER, DELEGATOR, TREASURY, KYLE, SHEV, ALAN] = await ethers.getSigners();

    [diamond, token, delegateRegistry] = await Promise.all([
      deployLumerinDiamond(),
      deployMORToken(),
      deployDelegateRegistry(),
    ]);

    [providerRegistry, modelRegistry, sessionRouter, marketplace] = await Promise.all([
      deployFacetProviderRegistry(diamond),
      deployFacetModelRegistry(diamond),
      deployFacetSessionRouter(diamond, OWNER),
      deployFacetMarketplace(diamond, token, wei(0.0001), wei(900)),
      deployFacetDelegation(diamond, delegateRegistry),
    ]);

    providersDelegate = await deployProvidersDelegate(
      diamond,
      await TREASURY.getAddress(),
      wei(0.2, 25),
      'DLNAME',
      'ENDPOINT',
      payoutStart + 3 * DAY,
    );

    await token.transfer(KYLE, wei(1000));
    await token.transfer(SHEV, wei(1000));
    await token.transfer(DELEGATOR, wei(1000));
    await token.transfer(ALAN, wei(1000));

    await token.connect(OWNER).approve(sessionRouter, wei(1000));
    await token.connect(ALAN).approve(sessionRouter, wei(1000));
    await token.connect(KYLE).approve(providersDelegate, wei(1000));
    await token.connect(SHEV).approve(providersDelegate, wei(1000));
    await token.connect(ALAN).approve(providersDelegate, wei(1000));
    await token.connect(DELEGATOR).approve(modelRegistry, wei(1000));

    await reverter.snapshot();
  });

  afterEach(reverter.revert);

  describe('#providersDelegate_init', () => {
    it('should revert if try to call init function twice', async () => {
      await expect(
        providersDelegate.ProvidersDelegate_init(OWNER, await TREASURY.getAddress(), 1, '', '', 0),
      ).to.be.rejectedWith('Initializable: contract is already initialized');
    });
    it('should throw error when fee is invalid', async () => {
      await expect(
        deployProvidersDelegate(diamond, await TREASURY.getAddress(), wei(1.1, 25), 'DLNAME', 'ENDPOINT', 0n),
      ).to.be.revertedWithCustomError(providersDelegate, 'InvalidFee');
    });
  });

  describe('#setName', () => {
    it('should set the provider name', async () => {
      await providersDelegate.setName('TEST');

      expect(await providersDelegate.name()).eq('TEST');
    });
    it('should throw error when name is zero', async () => {
      await expect(providersDelegate.setName('')).to.be.revertedWithCustomError(providersDelegate, 'InvalidNameLength');
    });
    it('should throw error when caller is not an owner', async () => {
      await expect(providersDelegate.connect(KYLE).setName('')).to.be.revertedWith('Ownable: caller is not the owner');
    });
  });

  describe('#setEndpoint', () => {
    it('should set the provider endpoint', async () => {
      await providersDelegate.setEndpoint('TEST');

      expect(await providersDelegate.endpoint()).eq('TEST');
    });
    it('should throw error when endpoint is zero', async () => {
      await expect(providersDelegate.setEndpoint('')).to.be.revertedWithCustomError(
        providersDelegate,
        'InvalidEndpointLength',
      );
    });
    it('should throw error when caller is not an owner', async () => {
      await expect(providersDelegate.connect(KYLE).setEndpoint('')).to.be.revertedWith(
        'Ownable: caller is not the owner',
      );
    });
  });

  describe('#setFeeTreasuryTreasury', () => {
    it('should set the provider fee', async () => {
      await providersDelegate.setFeeTreasury(KYLE);

      expect(await providersDelegate.feeTreasury()).eq(KYLE);
    });
    it('should throw error when fee treasury is invalid', async () => {
      await expect(providersDelegate.setFeeTreasury(ZERO_ADDR)).to.be.revertedWithCustomError(
        providersDelegate,
        'InvalidFeeTreasuryAddress',
      );
    });
    it('should throw error when caller is not an owner', async () => {
      await expect(providersDelegate.connect(KYLE).setFeeTreasury(KYLE)).to.be.revertedWith(
        'Ownable: caller is not the owner',
      );
    });
  });

  describe('#setIsStakeClosed', () => {
    it('should set the isStakeClosed flag', async () => {
      await providersDelegate.setIsStakeClosed(true);

      expect(await providersDelegate.isStakeClosed()).eq(true);
    });
    it('should throw error when caller is not an owner', async () => {
      await expect(providersDelegate.connect(KYLE).setIsStakeClosed(true)).to.be.revertedWith(
        'Ownable: caller is not the owner',
      );
    });
  });

  describe('#stake', () => {
    it('should stake tokens, one staker', async () => {
      await providersDelegate.connect(KYLE).stake(wei(100));

      const staker = await providersDelegate.stakers(KYLE);
      expect(staker.staked).to.eq(wei(100));
      expect(staker.pendingRewards).to.eq(wei(0));
      expect(staker.isRestakeDisabled).to.eq(false);
      expect(await providersDelegate.totalStaked()).to.eq(wei(100));

      expect(await token.balanceOf(providersDelegate)).to.eq(wei(0));
      expect(await token.balanceOf(providerRegistry)).to.eq(wei(100));
      expect(await token.balanceOf(KYLE)).to.eq(wei(900));
    });
    it('should stake tokens, two staker', async () => {
      await providersDelegate.connect(KYLE).stake(wei(100));

      const staker1 = await providersDelegate.stakers(KYLE);
      expect(staker1.staked).to.eq(wei(100));
      expect(staker1.pendingRewards).to.eq(wei(0));
      expect(staker1.isRestakeDisabled).to.eq(false);
      expect(await providersDelegate.totalStaked()).to.eq(wei(100));

      await providersDelegate.connect(SHEV).stake(wei(200));

      const staker2 = await providersDelegate.stakers(SHEV);
      expect(staker2.staked).to.eq(wei(200));
      expect(staker2.pendingRewards).to.eq(wei(0));
      expect(staker2.isRestakeDisabled).to.eq(false);
      expect(await providersDelegate.totalStaked()).to.eq(wei(300));

      expect(await token.balanceOf(providersDelegate)).to.eq(wei(0));
      expect(await token.balanceOf(providerRegistry)).to.eq(wei(300));
      expect(await token.balanceOf(KYLE)).to.eq(wei(900));
      expect(await token.balanceOf(SHEV)).to.eq(wei(800));
    });
    it('should stake tokens when stake closed but deregistration is available', async () => {
      await providersDelegate.setIsStakeClosed(true);
      await setTime(payoutStart + 4 * DAY);
      await providersDelegate.connect(KYLE).stake(wei(100));
    });
    it('should throw error when the stake is too low', async () => {
      await expect(providersDelegate.connect(KYLE).stake(wei(0))).to.be.revertedWithCustomError(
        providersDelegate,
        'InsufficientAmount',
      );
    });
    it('should throw error when the stake closed', async () => {
      await providersDelegate.setIsStakeClosed(true);
      await expect(providersDelegate.connect(KYLE).stake(wei(1))).to.be.revertedWithCustomError(
        providersDelegate,
        'StakeClosed',
      );
    });
  });

  describe('#claim', () => {
    beforeEach(async () => {
      await setTime(5000);
    });
    it('should correctly claim, one staker, full claim', async () => {
      await providersDelegate.connect(KYLE).stake(wei(100));

      await token.transfer(providersDelegate, wei(10));

      expect(await providersDelegate.getCurrentStakerRewards(KYLE)).to.eq(wei(10));

      await providersDelegate.connect(KYLE).claim(KYLE, wei(9999));
      expect(await token.balanceOf(KYLE)).to.eq(wei(908));
      expect(await token.balanceOf(TREASURY)).to.eq(wei(2));
    });
    it('should correctly claim, one staker, partial claim', async () => {
      await providersDelegate.connect(KYLE).stake(wei(100));

      await token.transfer(providersDelegate, wei(20));

      await providersDelegate.connect(KYLE).claim(KYLE, wei(5));
      expect(await token.balanceOf(KYLE)).to.eq(wei(904));
      expect(await token.balanceOf(TREASURY)).to.eq(wei(1));

      await providersDelegate.connect(KYLE).claim(KYLE, wei(10));
      expect(await token.balanceOf(KYLE)).to.eq(wei(912));
      expect(await token.balanceOf(TREASURY)).to.eq(wei(3));

      await providersDelegate.connect(KYLE).claim(KYLE, wei(5));
      expect(await token.balanceOf(KYLE)).to.eq(wei(916));
      expect(await token.balanceOf(TREASURY)).to.eq(wei(4));
    });
    it('should correctly claim, two stakers, full claim, enter when no rewards distributed', async () => {
      await providersDelegate.connect(KYLE).stake(wei(100));
      await providersDelegate.connect(SHEV).stake(wei(300));

      await token.transfer(providersDelegate, wei(40));

      await providersDelegate.connect(KYLE).claim(KYLE, wei(9999));
      expect(await token.balanceOf(KYLE)).to.eq(wei(908));
      expect(await token.balanceOf(TREASURY)).to.eq(wei(2));

      await providersDelegate.connect(SHEV).claim(SHEV, wei(9999));
      expect(await token.balanceOf(SHEV)).to.eq(wei(724));
      expect(await token.balanceOf(TREASURY)).to.eq(wei(2 + 6));
    });
    it('should correctly claim, two stakers, partial claim, enter when rewards distributed', async () => {
      await providersDelegate.connect(KYLE).stake(wei(100));

      await token.transfer(providersDelegate, wei(10));

      await providersDelegate.connect(SHEV).stake(wei(300));

      await token.transfer(providersDelegate, wei(40));

      await providersDelegate.connect(KYLE).claim(KYLE, wei(9999));
      expect(await token.balanceOf(KYLE)).to.eq(wei(916)); // 10 + 25% from 40
      expect(await token.balanceOf(TREASURY)).to.eq(wei(4));

      await providersDelegate.connect(SHEV).claim(SHEV, wei(20));
      expect(await token.balanceOf(SHEV)).to.eq(wei(716));
      expect(await token.balanceOf(TREASURY)).to.eq(wei(4 + 4));

      await token.transfer(providersDelegate, wei(100));

      await providersDelegate.connect(SHEV).claim(SHEV, wei(20));
      expect(await token.balanceOf(SHEV)).to.eq(wei(732));
      expect(await token.balanceOf(TREASURY)).to.eq(wei(4 + 4 + 4));

      await providersDelegate.connect(KYLE).claim(KYLE, wei(9999));
      expect(await token.balanceOf(KYLE)).to.eq(wei(936));
      expect(await token.balanceOf(TREASURY)).to.eq(wei(4 + 4 + 4 + 5));

      await providersDelegate.connect(SHEV).claim(SHEV, wei(999));
      expect(await token.balanceOf(SHEV)).to.eq(wei(784));
      expect(await token.balanceOf(TREASURY)).to.eq(wei(4 + 4 + 4 + 5 + 13));
    });
    it('should throw error when nothing to claim', async () => {
      await expect(providersDelegate.connect(KYLE).claim(KYLE, wei(999))).to.be.revertedWithCustomError(
        providersDelegate,
        'ClaimAmountIsZero',
      );
    });
  });

  describe('#restake', () => {
    beforeEach(async () => {
      await setTime(5000);
    });
    it('should correctly restake, two stakers, full restake', async () => {
      await providersDelegate.connect(KYLE).stake(wei(100));
      await providersDelegate.connect(SHEV).stake(wei(300));

      await token.transfer(providersDelegate, wei(100));

      await providersDelegate.connect(OWNER).restake(KYLE, wei(9999));
      expect((await providersDelegate.stakers(KYLE)).staked).to.eq(wei(120));
      expect(await token.balanceOf(KYLE)).to.eq(wei(900));
      expect(await token.balanceOf(TREASURY)).to.eq(wei(5));

      await token.transfer(providersDelegate, wei(100));

      await providersDelegate.connect(KYLE).claim(KYLE, wei(9999));
      expect(await token.balanceOf(KYLE)).to.closeTo(wei(900 + 28.57 * 0.8), wei(0.01));
      expect(await token.balanceOf(TREASURY)).to.closeTo(wei(5 + 28.57 * 0.2), wei(0.01));

      await providersDelegate.connect(SHEV).claim(SHEV, wei(9999));
      expect(await token.balanceOf(SHEV)).to.closeTo(wei(700 + 75 * 0.8 + 71.42 * 0.8), wei(0.01));
      expect(await token.balanceOf(TREASURY)).to.closeTo(wei(5 + 28.57 * 0.2 + 75 * 0.2 + 71.42 * 0.2), wei(0.01));
    });
    it('should correctly restake, two stakers, partial restake', async () => {
      await providersDelegate.connect(KYLE).stake(wei(100));
      await providersDelegate.connect(SHEV).stake(wei(300));

      await token.transfer(providersDelegate, wei(100));

      await providersDelegate.connect(OWNER).restake(KYLE, wei(20));
      expect((await providersDelegate.stakers(KYLE)).staked).to.eq(wei(116));
      expect(await token.balanceOf(KYLE)).to.eq(wei(900));
      expect(await token.balanceOf(TREASURY)).to.eq(wei(4));

      await token.transfer(providersDelegate, wei(100));

      await providersDelegate.connect(KYLE).claim(KYLE, wei(9999));
      expect(await token.balanceOf(KYLE)).to.closeTo(wei(900 + 5 * 0.8 + 27.88 * 0.8), wei(0.01));
      expect(await token.balanceOf(TREASURY)).to.closeTo(wei(4 + 5 * 0.2 + 27.88 * 0.2), wei(0.01));

      await providersDelegate.connect(SHEV).claim(SHEV, wei(9999));
      expect(await token.balanceOf(SHEV)).to.closeTo(wei(700 + 75 * 0.8 + 72.11 * 0.8), wei(0.01));
      expect(await token.balanceOf(TREASURY)).to.closeTo(
        wei(4 + 5 * 0.2 + 27.88 * 0.2 + 75 * 0.2 + 72.11 * 0.2),
        wei(0.01),
      );
    });
    it('should correctly restake with zero fee', async () => {
      await providersDelegate.connect(KYLE).stake(wei(100));
      await token.transfer(providersDelegate, wei(10));
      await providersDelegate.connect(OWNER).restake(KYLE, 1);

      expect(await token.balanceOf(TREASURY)).to.eq(wei(0));
    });
    it('should throw error when restake caller is invalid', async () => {
      await expect(providersDelegate.connect(KYLE).restake(SHEV, wei(999))).to.be.revertedWithCustomError(
        providersDelegate,
        'RestakeInvalidCaller',
      );
    });
    it('should throw error when restake is disabled', async () => {
      await providersDelegate.connect(SHEV).setIsRestakeDisabled(true);
      await expect(providersDelegate.restake(SHEV, wei(999))).to.be.revertedWithCustomError(
        providersDelegate,
        'RestakeDisabled',
      );
    });
  });

  describe('#providerDeregister', () => {
    it('should deregister the provider', async () => {
      await providersDelegate.connect(KYLE).stake(wei(100));
      await providersDelegate.providerDeregister([]);

      await providersDelegate.connect(KYLE).claim(KYLE, wei(9999));
      expect(await token.balanceOf(KYLE)).to.eq(wei(1000));
      expect(await token.balanceOf(TREASURY)).to.eq(wei(0));
    });
    it('should throw error when caller is not an owner', async () => {
      await expect(providersDelegate.connect(KYLE).providerDeregister([])).to.be.revertedWith(
        'Ownable: caller is not the owner',
      );
    });
  });

  describe('#postModelBid, #deleteModelBids', () => {
    const baseModelId = getHex(Buffer.from('1'));

    it('should deregister the model bid and delete it', async () => {
      // Register provider
      await providersDelegate.connect(SHEV).stake(wei(300));

      // Register model
      await modelRegistry
        .connect(DELEGATOR)
        .modelRegister(DELEGATOR, baseModelId, getHex(Buffer.from('ipfs://ipfsaddress')), 0, wei(100), 'name', [
          'tag_1',
        ]);
      const modelId = await modelRegistry.getModelId(DELEGATOR, baseModelId);

      // Register bid
      await providersDelegate.postModelBid(modelId, wei(0.0001));
      let bidId = await marketplace.getBidId(await providersDelegate.getAddress(), modelId, 0);

      await providersDelegate.deleteModelBids([bidId]);

      // Register bid again and deregister not from OWNER
      await providersDelegate.postModelBid(modelId, wei(0.0001));
      bidId = await marketplace.getBidId(await providersDelegate.getAddress(), modelId, 1);

      await setTime(payoutStart + 10 * DAY);
      await providersDelegate.connect(ALAN).deleteModelBids([bidId]);
    });
    it('should throw error when caller is not an owner for `postModelBid`', async () => {
      await expect(providersDelegate.connect(KYLE).postModelBid(baseModelId, wei(0.0001))).to.be.revertedWith(
        'Ownable: caller is not the owner',
      );
    });
    it('should throw error when caller is not an owner for `deleteModelBids`', async () => {
      await expect(providersDelegate.connect(KYLE).deleteModelBids([baseModelId])).to.be.revertedWith(
        'Ownable: caller is not the owner',
      );
    });
    it('should throw error when try to add bid after the deregistration opened', async () => {
      await setTime(payoutStart + 10 * DAY);

      await expect(providersDelegate.postModelBid(baseModelId, wei(0.0001))).to.be.revertedWithCustomError(
        providersDelegate,
        'BidCannotBeCreatedDuringThisPeriod',
      );
    });
  });

  describe('#version', () => {
    it('should return the correct contract version', async () => {
      expect(await providersDelegate.version()).to.eq(1);
    });
  });

  describe('full flow', () => {
    const baseModelId = getHex(Buffer.from('1'));

    it('should claim correct reward amount', async () => {
      // Register provider
      await providersDelegate.connect(KYLE).stake(wei(100));
      await providersDelegate.connect(SHEV).stake(wei(300));

      // Register model
      await modelRegistry
        .connect(DELEGATOR)
        .modelRegister(DELEGATOR, baseModelId, getHex(Buffer.from('ipfs://ipfsaddress')), 0, wei(100), 'name', [
          'tag_1',
        ]);
      const modelId = await modelRegistry.getModelId(DELEGATOR, baseModelId);

      // Register bid
      await providersDelegate.postModelBid(modelId, wei(0.0001));
      const bidId = await marketplace.getBidId(await providersDelegate.getAddress(), modelId, 0);

      await setTime(payoutStart + 10 * DAY);
      const { msg, signature } = await getProviderApproval(OWNER, ALAN, bidId);
      await sessionRouter.connect(ALAN).openSession(ALAN, wei(50), false, msg, signature);
      const sessionId = await sessionRouter.getSessionId(ALAN, providersDelegate, bidId, 0);

      const sessionTreasuryBalanceBefore = await token.balanceOf(OWNER);

      await setTime(payoutStart + 15 * DAY);
      const { msg: receiptMsg } = await getReceipt(OWNER, sessionId, 0, 0);
      const { signature: receiptSig } = await getReceipt(OWNER, sessionId, 0, 0);
      await sessionRouter.connect(ALAN).closeSession(receiptMsg, receiptSig);

      const sessionTreasuryBalanceAfter = await token.balanceOf(OWNER);
      const reward = sessionTreasuryBalanceBefore - sessionTreasuryBalanceAfter;

      await providersDelegate.claim(KYLE, wei(9999));
      await providersDelegate.claim(SHEV, wei(9999));
      expect(await token.balanceOf(KYLE)).to.eq(wei(900) + BigInt(Number(reward.toString()) * 0.25 * 0.8));
      expect(await token.balanceOf(SHEV)).to.eq(wei(700) + BigInt(Number(reward.toString()) * 0.75 * 0.8));
      expect(await token.balanceOf(TREASURY)).to.eq(BigInt(Number(reward.toString()) * 0.2));
    });

    it('should correctly deregister provider without fees', async () => {
      await setTime(payoutStart + 1 * DAY);

      // Register provider
      await providersDelegate.connect(KYLE).stake(wei(100));
      await providersDelegate.connect(SHEV).stake(wei(300));

      // Register model
      await modelRegistry
        .connect(DELEGATOR)
        .modelRegister(DELEGATOR, baseModelId, getHex(Buffer.from('ipfs://ipfsaddress')), 0, wei(100), 'name', [
          'tag_1',
        ]);
      const modelId = await modelRegistry.getModelId(DELEGATOR, baseModelId);

      // Register bid
      await providersDelegate.postModelBid(modelId, wei(0.0001));
      const bidId = await marketplace.getBidId(await providersDelegate.getAddress(), modelId, 0);

      // Open session
      await setTime(payoutStart + 10 * DAY);
      const { msg, signature } = await getProviderApproval(OWNER, ALAN, bidId);
      await sessionRouter.connect(ALAN).openSession(ALAN, wei(50), false, msg, signature);
      const sessionId = await sessionRouter.getSessionId(ALAN, providersDelegate, bidId, 0);

      // Close session
      await setTime(payoutStart + 15 * DAY);
      const { msg: receiptMsg } = await getReceipt(OWNER, sessionId, 0, 0);
      const { signature: receiptSig } = await getReceipt(OWNER, sessionId, 0, 0);
      await sessionRouter.connect(ALAN).closeSession(receiptMsg, receiptSig);

      // Add the new Staker
      await providersDelegate.connect(ALAN).stake(wei(1000));

      // Deregister the providers
      await providersDelegate.connect(KYLE).providerDeregister([bidId]);

      // Claim rewards
      await providersDelegate.claim(KYLE, wei(9999));
      await providersDelegate.claim(SHEV, wei(9999));
      await providersDelegate.claim(ALAN, wei(9999));
      expect(await token.balanceOf(KYLE)).to.closeTo(wei(1000), wei(0.1));
      expect(await token.balanceOf(SHEV)).to.closeTo(wei(1000), wei(0.1));
      expect(await token.balanceOf(ALAN)).to.closeTo(wei(1000), wei(0.2));
      expect(await token.balanceOf(TREASURY)).to.eq(wei(0));

      // Withdraw all stake after the first limit period
      expect((await providerRegistry.getProvider(await providersDelegate.getAddress()))[1]).to.greaterThan(0);
      await setTime(payoutStart + 15 * DAY + 1.1 * YEAR);
      await providersDelegate.connect(KYLE).stake(wei(1));
      await providersDelegate.connect(KYLE).providerDeregister([]);
      expect((await providerRegistry.getProvider(await providersDelegate.getAddress()))[1]).to.eq(0);
    });
  });
});

// npm run generate-types && npx hardhat test "test/delegate/ProvidersDelegate.test.ts"
// npx hardhat coverage --solcoverjs ./.solcover.ts --testfiles "test/delegate/ProvidersDelegate.test.ts"
