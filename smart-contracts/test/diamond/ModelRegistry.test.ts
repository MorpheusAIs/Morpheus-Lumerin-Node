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

describe('Model registry', () => {
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
    const expectedProvider = {
      endpoint: 'localhost:3334',
      stake: wei(100),
      createdAt: 0n,
      limitPeriodEnd: 0n,
      limitPeriodEarned: 0n,
      isDeleted: false,
      address: PROVIDER,
    };

    await MOR.transfer(PROVIDER, expectedProvider.stake * 100n);
    await MOR.connect(PROVIDER).approve(sessionRouter, expectedProvider.stake);

    await providerRegistry
      .connect(PROVIDER)
      .providerRegister(expectedProvider.address, expectedProvider.stake, expectedProvider.endpoint);
    expectedProvider.createdAt = await getCurrentBlockTime();
    expectedProvider.limitPeriodEnd = expectedProvider.createdAt + YEAR;

    return expectedProvider;
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
    describe('#__ModelRegistry_init', () => {
      it('should revert if try to call init function twice', async () => {
        const reason = 'Initializable: contract is already initialized';

        await expect(modelRegistry.__ModelRegistry_init()).to.be.rejectedWith(reason);
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
      const data = await modelRegistry.getModel(model.modelId);

      expect(await modelRegistry.models(0)).eq(model.modelId);
      expect(data).deep.equal([
        model.ipfsCID,
        model.fee,
        model.stake,
        await resolveAddress(model.owner),
        model.name,
        model.tags,
        model.createdAt,
        model.isDeleted,
      ]);
    });

    it('Should error when registering with insufficient stake', async () => {
      const minStake = 100n;
      await modelRegistry.setModelMinimumStake(minStake);

      await expect(
        modelRegistry.modelRegister(randomBytes32(), randomBytes32(), 0n, 0n, OWNER, 'a', []),
      ).revertedWithCustomError(modelRegistry, 'StakeTooLow');
    });

    it('Should error when registering with insufficient allowance', async () => {
      await expect(
        modelRegistry.connect(THIRD).modelRegister(randomBytes32(), randomBytes32(), 0n, 100n, THIRD, 'a', []),
      ).to.rejectedWith('ERC20: insufficient allowance');
    });

    it('Should error when register account doesnt match sender account', async () => {
      await MOR.approve(modelRegistry, 100n);

      await expect(
        modelRegistry.connect(THIRD).modelRegister(randomBytes32(), randomBytes32(), 0n, 100n, SECOND, 'a', []),
      ).to.revertedWithCustomError(modelRegistry, 'NotOwnerOrModelOwner');
    });

    it('Should deregister by owner', async () => {
      await marketplace.connect(PROVIDER).deleteModelBid(bid.id);

      await modelRegistry.modelDeregister(model.modelId);

      expect((await modelRegistry.getModel(model.modelId)).isDeleted).to.equal(true);
      expect(await modelRegistry.models(0n)).equals(model.modelId);
    });

    it('Should error if model not known by admin', async () => {
      await expect(modelRegistry.modelDeregister(randomBytes32())).to.revertedWithCustomError(
        modelRegistry,
        'ModelNotFound',
      );
    });

    it('Should error if caller is not owner or model owner', async () => {
      await expect(modelRegistry.connect(SECOND).modelDeregister(model.modelId)).to.revertedWithCustomError(
        modelRegistry,
        'NotOwnerOrModelOwner',
      );
    });

    it('Should return stake on deregister', async () => {
      await marketplace.connect(PROVIDER).deleteModelBid(bid.id);

      const balanceBefore = await MOR.balanceOf(model.owner);
      await modelRegistry.modelDeregister(model.modelId);
      const balanceAfter = await MOR.balanceOf(model.owner);

      expect(balanceAfter - balanceBefore).eq(model.stake);
    });

    it('should error when deregistering a model that has bids', async () => {
      // try deregistering model
      await expect(modelRegistry.modelDeregister(model.modelId)).to.revertedWithCustomError(
        modelRegistry,
        'ModelHasActiveBids',
      );

      // remove bid
      await marketplace.connect(PROVIDER).deleteModelBid(bid.id);

      // deregister model
      await modelRegistry.modelDeregister(model.modelId);
    });

    it('Should update existing model', async () => {
      const updates = {
        ipfsCID: getHex(Buffer.from('ipfs://new-ipfsaddress')),
        fee: BigInt(model.fee) * 2n,
        addStake: BigInt(model.stake) * 2n,
        owner: PROVIDER,
        name: 'Llama 3.0',
        tags: ['llama', 'smart', 'angry'],
      };
      await MOR.approve(modelRegistry, updates.addStake);

      await modelRegistry.modelRegister(
        model.modelId,
        updates.ipfsCID,
        updates.fee,
        updates.addStake,
        updates.owner,
        updates.name,
        updates.tags,
      );
      const providerData = await modelRegistry.getModel(model.modelId);

      expect(providerData).deep.equal([
        updates.ipfsCID,
        updates.fee,
        BigInt(model.stake) + updates.addStake,
        await resolveAddress(updates.owner),
        updates.name,
        updates.tags,
        model.createdAt,
        model.isDeleted,
      ]);
    });

    it('Should emit event on update', async () => {
      const updates = {
        ipfsCID: getHex(Buffer.from('ipfs://new-ipfsaddress')),
        fee: BigInt(model.fee) * 2n,
        addStake: BigInt(model.stake) * 2n,
        owner: PROVIDER,
        name: 'Llama 3.0',
        tags: ['llama', 'smart', 'angry'],
      };

      await MOR.approve(modelRegistry, updates.addStake);

      await expect(
        modelRegistry.modelRegister(
          model.modelId,
          updates.ipfsCID,
          updates.fee,
          updates.addStake,
          updates.owner,
          updates.name,
          updates.tags,
        ),
      ).to.emit(modelRegistry, 'ModelRegisteredUpdated');
    });

    it('should reregister model', async () => {
      await marketplace.connect(PROVIDER).deleteModelBid(bid.id);

      // check indexes
      expect(await modelRegistry.models(0)).eq(model.modelId);

      // deregister
      await modelRegistry.modelDeregister(model.modelId);

      // check indexes
      expect(await modelRegistry.models(0)).eq(model.modelId);

      // reregister
      const modelId = model.modelId;
      const model2 = {
        ipfsCID: randomBytes32(),
        fee: 100n,
        stake: 100n,
        owner: await resolveAddress(OWNER),
        name: 'model2',
        tags: ['model', '2'],
        createdAt: model.createdAt,
      };
      await MOR.transfer(OWNER, model2.stake);
      await MOR.approve(modelRegistry, model2.stake);
      await modelRegistry.modelRegister(
        modelId,
        model2.ipfsCID,
        model2.fee,
        model2.stake,
        model2.owner,
        model2.name,
        model2.tags,
      );
      // check indexes
      expect(await modelRegistry.models(0)).eq(modelId);
      expect(await modelRegistry.getModel(modelId)).deep.equal([
        model2.ipfsCID,
        model2.fee,
        model2.stake,
        model2.owner,
        model2.name,
        model2.tags,
        model2.createdAt,
        false,
      ]);
    });

    it('Should error if reregister model by caller is not owner or model owner', async () => {
      await expect(
        modelRegistry
          .connect(SECOND)
          .modelRegister(model.modelId, model.ipfsCID, model.fee, model.stake, model.owner, model.name, model.tags),
      ).to.revertedWithCustomError(modelRegistry, 'NotOwnerOrModelOwner');
    });

    describe('Getters', () => {
      it('Should get by address', async () => {
        const providerData = await modelRegistry.getModel(model.modelId);
        expect(providerData).deep.equal([
          model.ipfsCID,
          model.fee,
          model.stake,
          await resolveAddress(model.owner),
          model.name,
          model.tags,
          model.createdAt,
          model.isDeleted,
        ]);
      });
    });

    describe('Min stake', () => {
      it('Should set min stake', async () => {
        const minStake = 100n;
        await expect(modelRegistry.setModelMinimumStake(minStake))
          .to.emit(modelRegistry, 'ModelMinimumStakeSet')
          .withArgs(minStake);

        expect(await modelRegistry.modelMinimumStake()).eq(minStake);
      });
      it('Should error when not owner is setting min stake', async () => {
        await expect(modelRegistry.connect(THIRD).setModelMinimumStake(0)).to.revertedWithCustomError(
          diamond,
          'OwnableUnauthorizedAccount',
        );
      });
      // it("Should get model stats", async () => {
      //   const stats = await modelRegistry.modelStats([model.modelId]);

      //   expect(stats).deep.equal({
      //     count: 0,
      //     totalDuration: {
      //       mean: 0n,
      //       sqSum: 0n,
      //     },
      //     tpsScaled1000: {
      //       mean: 0n,
      //       sqSum: 0n,
      //     },
      //     ttftMs: {
      //       mean: 0n,
      //       sqSum: 0n,
      //     },
      //   });
      // });
    });
  });
});
