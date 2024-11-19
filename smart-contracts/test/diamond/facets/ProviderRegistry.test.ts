import { LumerinDiamond, Marketplace, ModelRegistry, MorpheusToken, ProviderRegistry } from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { expect } from 'chai';
import { ethers } from 'hardhat';

import { DelegateRegistry } from '@/generated-types/ethers/contracts/mock/delegate-registry/src';
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
} from '@/test/helpers/deployers';
import { Reverter } from '@/test/helpers/reverter';
import { setNextTime } from '@/utils/block-helper';
import { YEAR } from '@/utils/time';

describe('ProviderRegistry', () => {
  const reverter = new Reverter();

  let OWNER: SignerWithAddress;
  let PROVIDER: SignerWithAddress;

  let diamond: LumerinDiamond;
  let providerRegistry: ProviderRegistry;
  let modelRegistry: ModelRegistry;
  let marketplace: Marketplace;

  let token: MorpheusToken;
  let delegateRegistry: DelegateRegistry;

  const baseModelId = getHex(Buffer.from('1'));
  let modelId = getHex(Buffer.from(''));
  const ipfsCID = getHex(Buffer.from('ipfs://ipfsaddress'));

  before(async () => {
    [OWNER, PROVIDER] = await ethers.getSigners();

    [diamond, token, delegateRegistry] = await Promise.all([
      deployLumerinDiamond(),
      deployMORToken(),
      deployDelegateRegistry(),
    ]);

    [providerRegistry, modelRegistry, , marketplace] = await Promise.all([
      deployFacetProviderRegistry(diamond),
      deployFacetModelRegistry(diamond),
      deployFacetSessionRouter(diamond, OWNER),
      deployFacetMarketplace(diamond, token, wei(0.0001), wei(900)),
      deployFacetDelegation(diamond, delegateRegistry),
    ]);

    await token.transfer(PROVIDER, wei(1000));
    await token.connect(PROVIDER).approve(providerRegistry, wei(1000));
    await token.approve(providerRegistry, wei(1000));

    modelId = await modelRegistry.getModelId(PROVIDER, baseModelId);

    await reverter.snapshot();
  });

  afterEach(reverter.revert);

  describe('#__ProviderRegistry_init', () => {
    it('should revert if try to call init function twice', async () => {
      await expect(providerRegistry.__ProviderRegistry_init()).to.be.rejectedWith(
        'Initializable: contract is already initialized',
      );
    });
  });

  describe('#providerSetMinStake', async () => {
    it('should set min stake', async () => {
      const minStake = wei(100);

      await expect(providerRegistry.providerSetMinStake(minStake))
        .to.emit(providerRegistry, 'ProviderMinimumStakeUpdated')
        .withArgs(minStake);

      expect(await providerRegistry.getProviderMinimumStake()).eq(minStake);
    });

    it('should throw error when caller is not an owner', async () => {
      await expect(providerRegistry.connect(PROVIDER).providerSetMinStake(100)).to.be.revertedWithCustomError(
        diamond,
        'OwnableUnauthorizedAccount',
      );
    });
  });

  describe('#providerRegister', async () => {
    it('should register a new provider', async () => {
      await setNextTime(300);
      await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(100), 'test');

      const data = await providerRegistry.getProvider(PROVIDER);

      expect(data.endpoint).to.eq('test');
      expect(data.stake).to.eq(wei(100));
      expect(data.createdAt).to.eq(300);
      expect(data.limitPeriodEnd).to.eq(YEAR + 300);
      expect(data.limitPeriodEarned).to.eq(0);
      expect(data.isDeleted).to.eq(false);
      expect(await providerRegistry.getIsProviderActive(PROVIDER)).to.eq(true);

      expect(await token.balanceOf(providerRegistry)).to.eq(wei(100));
      expect(await token.balanceOf(PROVIDER)).to.eq(wei(900));

      expect(await providerRegistry.getActiveProviders(0, 10)).to.deep.eq([[PROVIDER.address], 1n]);

      await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(0), 'test');
    });
    it('should add stake to existed provider', async () => {
      await setNextTime(300);
      await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(100), 'test');
      await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(300), 'test2');

      const data = await providerRegistry.getProvider(PROVIDER);

      expect(data.endpoint).to.eq('test2');
      expect(data.stake).to.eq(wei(400));
      expect(data.createdAt).to.eq(300);
      expect(data.limitPeriodEnd).to.eq(YEAR + 300);
      expect(data.limitPeriodEarned).to.eq(0);
      expect(data.isDeleted).to.eq(false);
      expect(await providerRegistry.getIsProviderActive(PROVIDER)).to.eq(true);

      expect(await token.balanceOf(providerRegistry)).to.eq(wei(400));
      expect(await token.balanceOf(PROVIDER)).to.eq(wei(600));
    });
    it('should activate deregistered provider', async () => {
      await setNextTime(300);
      await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(100), 'test');
      await setNextTime(301 + YEAR);
      await providerRegistry.connect(PROVIDER).providerDeregister(PROVIDER);

      let data = await providerRegistry.getProvider(PROVIDER);
      expect(data.isDeleted).to.eq(true);
      expect(await providerRegistry.getIsProviderActive(PROVIDER)).to.eq(false);

      await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(1), 'test2');
      data = await providerRegistry.getProvider(PROVIDER);

      expect(data.endpoint).to.eq('test2');
      expect(data.stake).to.eq(wei(1));
      expect(data.createdAt).to.eq(300);
      expect(data.limitPeriodEnd).to.eq(YEAR + 300);
      expect(data.limitPeriodEarned).to.eq(0);
      expect(data.isDeleted).to.eq(false);
      expect(await providerRegistry.getIsProviderActive(PROVIDER)).to.eq(true);
    });
    it('should register a new provider from the delegatee address', async () => {
      await delegateRegistry
        .connect(PROVIDER)
        .delegateContract(OWNER, providerRegistry, await providerRegistry.DELEGATION_RULES_PROVIDER(), true);

      await setNextTime(300);
      await providerRegistry.connect(OWNER).providerRegister(PROVIDER, wei(100), 'test');

      const data = await providerRegistry.getProvider(PROVIDER);

      expect(data.endpoint).to.eq('test');
      expect(data.stake).to.eq(wei(100));
      expect(data.createdAt).to.eq(300);
      expect(data.limitPeriodEnd).to.eq(YEAR + 300);
      expect(data.limitPeriodEarned).to.eq(0);
      expect(data.isDeleted).to.eq(false);
      expect(await providerRegistry.getIsProviderActive(PROVIDER)).to.eq(true);

      expect(await token.balanceOf(providerRegistry)).to.eq(wei(100));
      expect(await token.balanceOf(PROVIDER)).to.eq(wei(900));

      expect(await providerRegistry.getActiveProviders(0, 10)).to.deep.eq([[PROVIDER.address], 1n]);

      await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(0), 'test');
    });
    it('should throw error when the stake is too low', async () => {
      await providerRegistry.providerSetMinStake(wei(2));
      await expect(
        providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(0), ''),
      ).to.be.revertedWithCustomError(providerRegistry, 'ProviderStakeTooLow');
    });
    it('should throw error when create provider without delegation or with incorrect rules', async () => {
      await expect(
        providerRegistry.connect(OWNER).providerRegister(PROVIDER, wei(0), ''),
      ).to.be.revertedWithCustomError(providerRegistry, 'InsufficientRightsForOperation');

      await delegateRegistry
        .connect(PROVIDER)
        .delegateContract(OWNER, providerRegistry, getHex(Buffer.from('123')), true);
      await expect(
        providerRegistry.connect(OWNER).providerRegister(PROVIDER, wei(0), ''),
      ).to.be.revertedWithCustomError(providerRegistry, 'InsufficientRightsForOperation');
    });
  });

  describe('#providerDeregister', async () => {
    it('should deregister the provider', async () => {
      await setNextTime(300);
      await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(100), 'test');
      await setNextTime(301 + YEAR);
      await providerRegistry.connect(PROVIDER).providerDeregister(PROVIDER);

      expect((await providerRegistry.getProvider(PROVIDER)).isDeleted).to.equal(true);
      expect(await providerRegistry.getIsProviderActive(PROVIDER)).to.eq(false);
      expect(await token.balanceOf(providerRegistry)).to.eq(0);
      expect(await token.balanceOf(PROVIDER)).to.eq(wei(1000));

      expect(await providerRegistry.getActiveProviders(0, 10)).to.deep.eq([[], 0n]);
    });
    it('should deregister the provider from the delegatee address', async () => {
      await setNextTime(300);
      await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(100), 'test');

      await delegateRegistry
        .connect(PROVIDER)
        .delegateContract(OWNER, providerRegistry, await providerRegistry.DELEGATION_RULES_PROVIDER(), true);
      await setNextTime(301 + YEAR);
      await providerRegistry.connect(OWNER).providerDeregister(PROVIDER);

      expect((await providerRegistry.getProvider(PROVIDER)).isDeleted).to.equal(true);
      expect(await providerRegistry.getIsProviderActive(PROVIDER)).to.eq(false);
      expect(await token.balanceOf(providerRegistry)).to.eq(0);
      expect(await token.balanceOf(PROVIDER)).to.eq(wei(1000));

      expect(await providerRegistry.getActiveProviders(0, 10)).to.deep.eq([[], 0n]);
    });
    it('should deregister the provider without transfer', async () => {
      await providerRegistry.providerSetMinStake(0);
      await setNextTime(300);
      await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(0), 'test');
      await setNextTime(301 + YEAR);
      await providerRegistry.connect(PROVIDER).providerDeregister(PROVIDER);

      expect((await providerRegistry.getProvider(PROVIDER)).isDeleted).to.equal(true);
      expect(await providerRegistry.getIsProviderActive(PROVIDER)).to.eq(false);
      expect(await token.balanceOf(providerRegistry)).to.eq(0);
      expect(await token.balanceOf(PROVIDER)).to.eq(wei(1000));
    });
    it('should throw error when provider is not found', async () => {
      await expect(providerRegistry.connect(OWNER).providerDeregister(OWNER)).to.be.revertedWithCustomError(
        providerRegistry,
        'ProviderNotFound',
      );
    });
    it('should throw error when provider has active bids', async () => {
      await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(100), 'test');
      await modelRegistry
        .connect(PROVIDER)
        .modelRegister(PROVIDER, baseModelId, ipfsCID, 0, wei(100), 'name', ['tag_1']);
      await marketplace.connect(PROVIDER).postModelBid(PROVIDER, modelId, wei(10));
      await expect(providerRegistry.connect(PROVIDER).providerDeregister(PROVIDER)).to.be.revertedWithCustomError(
        providerRegistry,
        'ProviderHasActiveBids',
      );
    });
    it('should throw error when delete provider few times', async () => {
      await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(100), 'test');
      await providerRegistry.connect(PROVIDER).providerDeregister(PROVIDER);
      await expect(providerRegistry.connect(PROVIDER).providerDeregister(PROVIDER)).to.be.revertedWithCustomError(
        providerRegistry,
        'ProviderHasAlreadyDeregistered',
      );
    });
  });
});

// npm run generate-types && npx hardhat test "test/diamond/facets/ProviderRegistry.test.ts"
// npx hardhat coverage --solcoverjs ./.solcover.ts --testfiles "test/diamond/facets/ProviderRegistry.test.ts"
