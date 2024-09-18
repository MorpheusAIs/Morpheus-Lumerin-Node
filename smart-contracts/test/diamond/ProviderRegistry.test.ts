import {
  IBidStorage,
  IMarketplace__factory,
  IModelRegistry__factory,
  IModelStorage,
  IProviderRegistry__factory,
  IProviderStorage,
  ISessionRouter__factory,
  LumerinDiamond,
  Marketplace,
  ModelRegistry,
  MorpheusToken,
  ProviderRegistry,
  SessionRouter,
} from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { expect } from 'chai';
import { Addressable, Fragment, resolveAddress } from 'ethers';
import { ethers } from 'hardhat';

import { FacetAction } from '../helpers/enums';
import { getDefaultPools } from '../helpers/pool-helper';
import { Reverter } from '../helpers/reverter';

import { getHex, randomBytes32, wei } from '@/scripts/utils/utils';
import { getCurrentBlockTime } from '@/utils/block-helper';
import { HOUR, YEAR } from '@/utils/time';

describe('Provider registry', () => {
  const reverter = new Reverter();

  let OWNER: SignerWithAddress;
  let SECOND: SignerWithAddress;
  let THIRD: SignerWithAddress;
  let PROVIDER: SignerWithAddress;

  let diamond: LumerinDiamond;
  let marketplace: Marketplace;
  let modelRegistry: ModelRegistry;
  let providerRegistry: ProviderRegistry;
  let sessionRouter: SessionRouter;

  let MOR: MorpheusToken;

  async function deployProvider(): Promise<
    IProviderStorage.ProviderStruct & {
      address: Addressable;
    }
  > {
    const provider = {
      endpoint: 'localhost:3334',
      stake: wei(100),
      createdAt: 0n,
      limitPeriodEnd: 0n,
      limitPeriodEarned: 0n,
      isDeleted: false,
      address: PROVIDER,
    };

    await MOR.transfer(PROVIDER, provider.stake * 100n);
    await MOR.connect(PROVIDER).approve(sessionRouter, provider.stake);

    await providerRegistry.connect(PROVIDER).providerRegister(provider.address, provider.stake, provider.endpoint);
    provider.createdAt = await getCurrentBlockTime();
    provider.limitPeriodEnd = provider.createdAt + YEAR;

    return provider;
  }

  async function deployModel(): Promise<
    IModelStorage.ModelStruct & {
      modelId: string;
    }
  > {
    const model = {
      modelId: randomBytes32(),
      ipfsCID: getHex(Buffer.from('ipfs://ipfsaddress')),
      fee: 100,
      stake: 100,
      owner: OWNER,
      name: 'Llama 2.0',
      tags: ['llama', 'animal', 'cute'],
      createdAt: 0n,
      isDeleted: false,
    };

    await MOR.approve(modelRegistry, model.stake);

    await modelRegistry.modelRegister(
      model.modelId,
      model.ipfsCID,
      model.fee,
      model.stake,
      model.owner,
      model.name,
      model.tags,
    );
    model.createdAt = await getCurrentBlockTime();

    return model;
  }

  async function deployBid(model: any): Promise<
    IBidStorage.BidStruct & {
      id: string;
      modelId: string;
    }
  > {
    let bid = {
      id: '',
      modelId: model.modelId,
      pricePerSecond: wei(0.0001),
      nonce: 0,
      createdAt: 0n,
      deletedAt: 0,
      provider: PROVIDER,
    };

    await MOR.approve(modelRegistry, 10000n * 10n ** 18n);

    bid.id = await marketplace.connect(PROVIDER).postModelBid.staticCall(bid.provider, bid.modelId, bid.pricePerSecond);
    await marketplace.connect(PROVIDER).postModelBid(bid.provider, bid.modelId, bid.pricePerSecond);

    bid.createdAt = await getCurrentBlockTime();

    // generating data for sample session
    const durationSeconds = HOUR;
    const totalCost = bid.pricePerSecond * durationSeconds;
    const totalSupply = await sessionRouter.totalMORSupply(await getCurrentBlockTime());
    const todaysBudget = await sessionRouter.getTodaysBudget(await getCurrentBlockTime());

    const expectedSession = {
      durationSeconds,
      totalCost,
      pricePerSecond: bid.pricePerSecond,
      user: SECOND,
      provider: bid.provider,
      modelId: bid.modelId,
      bidId: bid.id,
      stake: (totalCost * totalSupply) / todaysBudget,
    };

    await MOR.transfer(SECOND, expectedSession.stake);
    await MOR.connect(SECOND).approve(modelRegistry, expectedSession.stake);

    return bid;
  }

  before('setup', async () => {
    [OWNER, SECOND, THIRD, PROVIDER] = await ethers.getSigners();

    const LinearDistributionIntervalDecrease = await ethers.getContractFactory('LinearDistributionIntervalDecrease');
    const linearDistributionIntervalDecrease = await LinearDistributionIntervalDecrease.deploy();

    const [LumerinDiamond, Marketplace, ModelRegistry, ProviderRegistry, SessionRouter, MorpheusToken] =
      await Promise.all([
        ethers.getContractFactory('LumerinDiamond'),
        ethers.getContractFactory('Marketplace'),
        ethers.getContractFactory('ModelRegistry'),
        ethers.getContractFactory('ProviderRegistry'),
        ethers.getContractFactory('SessionRouter', {
          libraries: {
            LinearDistributionIntervalDecrease: linearDistributionIntervalDecrease,
          },
        }),
        ethers.getContractFactory('MorpheusToken'),
      ]);

    [diamond, marketplace, modelRegistry, providerRegistry, sessionRouter, MOR] = await Promise.all([
      LumerinDiamond.deploy(),
      Marketplace.deploy(),
      ModelRegistry.deploy(),
      ProviderRegistry.deploy(),
      SessionRouter.deploy(),
      MorpheusToken.deploy(),
    ]);

    await diamond.__LumerinDiamond_init();

    await diamond['diamondCut((address,uint8,bytes4[])[])']([
      {
        facetAddress: marketplace,
        action: FacetAction.Add,
        functionSelectors: IMarketplace__factory.createInterface()
          .fragments.filter(Fragment.isFunction)
          .map((f) => f.selector),
      },
      {
        facetAddress: providerRegistry,
        action: FacetAction.Add,
        functionSelectors: IProviderRegistry__factory.createInterface()
          .fragments.filter(Fragment.isFunction)
          .map((f) => f.selector),
      },
      {
        facetAddress: sessionRouter,
        action: FacetAction.Add,
        functionSelectors: ISessionRouter__factory.createInterface()
          .fragments.filter(Fragment.isFunction)
          .map((f) => f.selector),
      },
      {
        facetAddress: modelRegistry,
        action: FacetAction.Add,
        functionSelectors: IModelRegistry__factory.createInterface()
          .fragments.filter(Fragment.isFunction)
          .map((f) => f.selector),
      },
    ]);

    marketplace = marketplace.attach(diamond.target) as Marketplace;
    providerRegistry = providerRegistry.attach(diamond.target) as ProviderRegistry;
    modelRegistry = modelRegistry.attach(diamond.target) as ModelRegistry;
    sessionRouter = sessionRouter.attach(diamond.target) as SessionRouter;

    await marketplace.__Marketplace_init(MOR);
    await modelRegistry.__ModelRegistry_init();
    await providerRegistry.__ProviderRegistry_init();

    await sessionRouter.__SessionRouter_init(OWNER, getDefaultPools());

    await reverter.snapshot();
  });

  afterEach(reverter.revert);

  describe('Diamond functionality', () => {
    describe('#__ProviderRegistry_init', () => {
      it('should revert if try to call init function twice', async () => {
        const reason = 'Initializable: contract is already initialized';

        await expect(providerRegistry.__ProviderRegistry_init()).to.be.rejectedWith(reason);
      });
    });
  });

  describe('Actions', () => {
    let provider: IProviderStorage.ProviderStruct;
    let model: IModelStorage.ModelStruct & {
      modelId: string;
    };
    let bid: IBidStorage.BidStruct & {
      id: string;
      modelId: string;
    };

    beforeEach(async () => {
      provider = await deployProvider();
      model = await deployModel();
      bid = await deployBid(model);
    });

    it('Should register', async () => {
      await providerRegistry.connect(SECOND).providerRegister(SECOND, provider.stake, provider.endpoint);

      const data = await providerRegistry.getProvider(SECOND);

      expect(data).deep.equal([
        provider.endpoint,
        provider.stake,
        await getCurrentBlockTime(),
        (await getCurrentBlockTime()) + YEAR,
        provider.limitPeriodEarned,
        false,
      ]);
    });

    it('Should error when registering with insufficient stake', async () => {
      const minStake = 100;
      await providerRegistry.providerSetMinStake(minStake);

      await expect(providerRegistry.providerRegister(SECOND, minStake - 1, 'endpoint')).to.be.revertedWithCustomError(
        providerRegistry,
        'StakeTooLow',
      );
    });

    it('Should error when registering with insufficient allowance', async () => {
      // await catchError(MOR.abi, "ERC20InsufficientAllowance", async () => {
      //   await providerRegistry.providerRegister([PROVIDER, 100n, "endpoint"]);
      // });
      await expect(providerRegistry.connect(THIRD).providerRegister(THIRD, 100n, 'endpoint')).to.be.revertedWith(
        'ERC20: insufficient allowance',
      );
    });

    it('Should error when register account doesnt match sender account', async () => {
      await expect(
        providerRegistry.connect(PROVIDER).providerRegister(THIRD, 100n, 'endpoint'),
      ).to.be.revertedWithCustomError(providerRegistry, 'NotOwnerOrProvider');
    });

    describe('Deregister', () => {
      it('Should deregister by provider', async () => {
        await marketplace.connect(PROVIDER).deleteModelBid(bid.id);

        await expect(providerRegistry.connect(PROVIDER).providerDeregister(PROVIDER))
          .to.emit(providerRegistry, 'ProviderDeregistered')
          .withArgs(PROVIDER);

        expect((await providerRegistry.getProvider(PROVIDER)).isDeleted).to.equal(true);
      });

      it('Should deregister by admin', async () => {
        await marketplace.connect(PROVIDER).deleteModelBid(bid.id);

        await expect(providerRegistry.providerDeregister(PROVIDER))
          .to.emit(providerRegistry, 'ProviderDeregistered')
          .withArgs(PROVIDER);
      });

      it('Should return stake on deregister', async () => {
        await marketplace.connect(PROVIDER).deleteModelBid(bid.id);

        const balanceBefore = await MOR.balanceOf(PROVIDER);
        await providerRegistry.connect(PROVIDER).providerDeregister(PROVIDER);
        const balanceAfter = await MOR.balanceOf(PROVIDER);
        expect(balanceAfter - balanceBefore).eq(provider.stake);
      });

      it('should error when deregistering a model that has bids', async () => {
        // try deregistering model
        await expect(providerRegistry.connect(PROVIDER).providerDeregister(PROVIDER)).to.be.revertedWithCustomError(
          providerRegistry,
          'ProviderHasActiveBids',
        );

        // remove bid
        await marketplace.connect(PROVIDER).deleteModelBid(bid.id);
        // deregister model
        await providerRegistry.connect(PROVIDER).providerDeregister(PROVIDER);
      });

      it('Should correctly reregister provider', async () => {
        await marketplace.connect(PROVIDER).deleteModelBid(bid.id);

        // deregister
        await providerRegistry.connect(PROVIDER).providerDeregister(PROVIDER);
        // check indexes
        const provider2 = {
          endpoint: 'new-endpoint-2',
          stake: 123n,
          limitPeriodEarned: provider.limitPeriodEarned,
          limitPeriodEnd: provider.limitPeriodEnd,
          createdAt: provider.createdAt,
        };
        // register again
        await MOR.transfer(PROVIDER, provider2.stake);
        await MOR.connect(PROVIDER).approve(providerRegistry, provider2.stake);
        await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, provider2.stake, provider2.endpoint);
        // check record
        expect(await providerRegistry.getProvider(PROVIDER)).deep.equal([
          provider2.endpoint,
          provider2.stake,
          provider2.createdAt,
          provider2.limitPeriodEnd,
          provider2.limitPeriodEarned,
          false,
        ]);
      });

      it('should error if caller is not an owner or provider', async () => {
        await marketplace.connect(PROVIDER).deleteModelBid(bid.id);

        await expect(providerRegistry.connect(SECOND).providerDeregister(PROVIDER)).to.revertedWithCustomError(
          providerRegistry,
          'NotOwnerOrProvider',
        );
      });

      it('should error if provider is not exists', async () => {
        await marketplace.connect(PROVIDER).deleteModelBid(bid.id);

        await expect(providerRegistry.connect(SECOND).providerDeregister(SECOND)).to.revertedWithCustomError(
          providerRegistry,
          'ProviderNotFound',
        );
      });
    });

    it('Should update stake and url', async () => {
      const updates = {
        addStake: BigInt(provider.stake) * 2n,
        endpoint: 'new-endpoint',
      };
      await MOR.connect(PROVIDER).approve(providerRegistry, updates.addStake);

      await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, updates.addStake, updates.endpoint);

      const providerData = await providerRegistry.getProvider(PROVIDER);
      expect(providerData).deep.equal([
        updates.endpoint,
        BigInt(provider.stake) + updates.addStake,
        provider.createdAt,
        provider.limitPeriodEnd,
        provider.limitPeriodEarned,
        provider.isDeleted,
      ]);
    });

    it('Should emit event on update', async () => {
      const updates = {
        addStake: BigInt(provider.stake) * 2n,
        endpoint: 'new-endpoint',
      };
      await MOR.connect(PROVIDER).approve(providerRegistry, updates.addStake);
      await expect(providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, updates.addStake, updates.endpoint))
        .to.emit(providerRegistry, 'ProviderRegisteredUpdated')
        .withArgs(PROVIDER);
    });

    describe('Getters', () => {
      it('Should get by address', async () => {
        const providerData = await providerRegistry.getProvider(PROVIDER);
        expect(providerData).deep.equal([
          provider.endpoint,
          provider.stake,
          provider.createdAt,
          provider.limitPeriodEnd,
          provider.limitPeriodEarned,
          provider.isDeleted,
        ]);
      });
    });

    describe('Min stake', () => {
      it('Should set min stake', async () => {
        const minStake = 100n;
        await expect(providerRegistry.providerSetMinStake(minStake))
          .to.emit(providerRegistry, 'ProviderMinStakeUpdated')
          .withArgs(minStake);

        expect(await providerRegistry.providerMinimumStake()).eq(minStake);
      });

      it('Should error when not owner is setting min stake', async () => {
        await expect(providerRegistry.connect(SECOND).providerSetMinStake(100)).to.be.revertedWith(
          'OwnableDiamondStorage: not an owner',
        );
      });
    });
  });
});
