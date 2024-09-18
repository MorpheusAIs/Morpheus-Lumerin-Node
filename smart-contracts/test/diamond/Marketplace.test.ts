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

describe('Marketplace', () => {
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
    const expectedModel = {
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

    await MOR.approve(modelRegistry, expectedModel.stake);

    await modelRegistry.modelRegister(
      expectedModel.modelId,
      expectedModel.ipfsCID,
      expectedModel.fee,
      expectedModel.stake,
      expectedModel.owner,
      expectedModel.name,
      expectedModel.tags,
    );
    expectedModel.createdAt = await getCurrentBlockTime();

    return expectedModel;
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
      bidID: bid.id,
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

    marketplace = marketplace.attach(diamond) as Marketplace;
    providerRegistry = providerRegistry.attach(diamond) as ProviderRegistry;
    modelRegistry = modelRegistry.attach(diamond) as ModelRegistry;
    sessionRouter = sessionRouter.attach(diamond) as SessionRouter;

    await marketplace.__Marketplace_init(MOR);
    await modelRegistry.__ModelRegistry_init();
    await providerRegistry.__ProviderRegistry_init();

    await sessionRouter.__SessionRouter_init(OWNER, getDefaultPools());

    await reverter.snapshot();
  });

  afterEach(reverter.revert);

  describe('bid actions', () => {
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

    it('Should create a bid and query by id', async () => {
      const data = await marketplace.bids(bid.id);

      expect(data).to.be.deep.equal([
        await resolveAddress(bid.provider),
        bid.modelId,
        bid.pricePerSecond,
        bid.nonce,
        bid.createdAt,
        bid.deletedAt,
      ]);
    });

    it("Should error if provider doesn't exist", async () => {
      await expect(
        marketplace.connect(SECOND).postModelBid(SECOND, bid.modelId, bid.pricePerSecond),
      ).to.be.revertedWithCustomError(marketplace, 'ProviderNotFound');
    });

    it("Should error if model doesn't exist", async () => {
      const unknownModel = randomBytes32();

      await expect(
        marketplace.connect(PROVIDER).postModelBid(bid.provider, unknownModel, bid.pricePerSecond),
      ).to.be.revertedWithCustomError(marketplace, 'ModelNotFound');
    });

    it('Should create second bid', async () => {
      // create new bid with same provider and modelId
      await marketplace.connect(PROVIDER).postModelBid(bid.provider, bid.modelId, bid.pricePerSecond);
      const timestamp = await getCurrentBlockTime();

      // check indexes are updated
      const newBids1 = await marketplace.providerActiveBids(bid.provider, 0, 10);
      const newBids2 = await marketplace.modelActiveBids(bid.modelId, 0, 10);

      expect(newBids1).to.be.deep.equal(newBids2);
      expect(await marketplace.bids(newBids1[0])).to.be.deep.equal([
        await resolveAddress(bid.provider),
        bid.modelId,
        bid.pricePerSecond,
        BigInt(bid.nonce) + 1n,
        timestamp,
        bid.deletedAt,
      ]);

      // check old bid is deleted
      const oldBid = await marketplace.bids(bid.id);
      expect(oldBid).to.be.deep.equal([
        await resolveAddress(bid.provider),
        bid.modelId,
        bid.pricePerSecond,
        bid.nonce,
        bid.createdAt,
        timestamp,
      ]);

      // check old bid is still queried
      const oldBids1 = await marketplace.providerBids(bid.provider, 0, 100);
      const oldBids2 = await marketplace.modelBids(bid.modelId, 0, 100);
      expect(oldBids1).to.be.deep.equal(oldBids2);
      expect(oldBids1.length).to.be.equal(2);
      expect(await marketplace.bids(oldBids1[0])).to.be.deep.equal([
        await resolveAddress(bid.provider),
        bid.modelId,
        bid.pricePerSecond,
        bid.nonce,
        bid.createdAt,
        timestamp,
      ]);
    });

    it('Should query by provider', async () => {
      const activeBidIds = await marketplace.providerActiveBids(bid.provider, 0, 10);

      expect(activeBidIds.length).to.equal(1);
      expect(activeBidIds[0]).to.equal(bid.id);
      expect(await marketplace.bids(activeBidIds[0])).to.deep.equal([
        await resolveAddress(bid.provider),
        bid.modelId,
        bid.pricePerSecond,
        bid.nonce,
        bid.createdAt,
        bid.deletedAt,
      ]);
    });

    describe('delete bid', () => {
      it('Should delete a bid', async () => {
        // delete bid
        await marketplace.connect(PROVIDER).deleteModelBid(bid.id);

        // check indexes are updated
        const activeBidIds1 = await marketplace.providerActiveBids(bid.provider, 0, 10);
        const activeBidIds2 = await marketplace.modelActiveBids(bid.modelId, 0, 10);

        expect(activeBidIds1.length).to.be.equal(0);
        expect(activeBidIds2.length).to.be.equal(0);

        // check bid is deleted
        const data = await marketplace.bids(bid.id);
        expect(data).to.be.deep.equal([
          await resolveAddress(bid.provider),
          bid.modelId,
          bid.pricePerSecond,
          bid.nonce,
          bid.createdAt,
          await getCurrentBlockTime(),
        ]);
      });

      it("Should error if bid doesn't exist", async () => {
        const unknownBid = randomBytes32();

        await expect(marketplace.connect(PROVIDER).deleteModelBid(unknownBid)).to.be.revertedWithCustomError(
          marketplace,
          'ActiveBidNotFound',
        );
      });

      it('Should error if not owner', async () => {
        await expect(marketplace.connect(THIRD).deleteModelBid(bid.id)).to.be.revertedWithCustomError(
          marketplace,
          'NotOwnerOrProvider',
        );
      });

      it('Should allow bid owner to delete bid', async () => {
        // delete bid
        await marketplace.connect(PROVIDER).deleteModelBid(bid.id);
      });

      it('Should allow contract owner to delete bid', async () => {
        // delete bid
        await marketplace.deleteModelBid(bid.id);
      });

      it('Should allow to create bid after it was deleted [H-1]', async () => {
        // delete bid
        await marketplace.connect(PROVIDER).deleteModelBid(bid.id);

        // create new bid with same provider and modelId
        await marketplace.connect(PROVIDER).postModelBid(bid.provider, bid.modelId, bid.pricePerSecond);
      });
    });

    describe('bid fee', () => {
      it('should set bid fee', async () => {
        const newFee = 100;
        await marketplace.setBidFee(newFee);

        const modelBidFee = await marketplace.getBidFee();
        expect(modelBidFee).to.be.equal(newFee);
      });

      it('should collect bid fee', async () => {
        const newFee = 100;
        await marketplace.setBidFee(newFee);
        await MOR.transfer(bid.provider, 100);

        // check balance before
        const balanceBefore = await MOR.balanceOf(marketplace);
        // add bid
        await MOR.connect(PROVIDER).approve(marketplace, Number(bid.pricePerSecond) + newFee);
        await marketplace.connect(PROVIDER).postModelBid(bid.provider, bid.modelId, bid.pricePerSecond);
        // check balance after
        const balanceAfter = await MOR.balanceOf(marketplace);
        expect(balanceAfter - balanceBefore).to.be.equal(newFee);
      });

      it('should allow withdrawal by owner', async () => {
        const newFee = 100;
        await marketplace.setBidFee(newFee);
        await MOR.transfer(bid.provider, 100);
        // add bid
        await MOR.connect(PROVIDER).approve(marketplace, Number(bid.pricePerSecond) + newFee);
        await marketplace.connect(PROVIDER).postModelBid(bid.provider, bid.modelId, bid.pricePerSecond);
        // check balance after
        const balanceBefore = await MOR.balanceOf(OWNER);
        await marketplace.withdraw(OWNER, newFee);
        const balanceAfter = await MOR.balanceOf(OWNER);
        expect(balanceAfter - balanceBefore).to.be.equal(newFee);
      });

      it('should not allow withdrawal by any other account except owner', async () => {
        const newFee = 100;
        await marketplace.setBidFee(newFee);
        await MOR.transfer(bid.provider, 100);
        // add bid
        await MOR.connect(PROVIDER).approve(marketplace, Number(bid.pricePerSecond) + newFee);
        await marketplace.connect(PROVIDER).postModelBid(bid.provider, bid.modelId, bid.pricePerSecond);
        // check balance after
        await expect(marketplace.connect(PROVIDER).withdraw(bid.provider, newFee)).to.be.revertedWith(
          'OwnableDiamondStorage: not an owner',
        );
      });

      it('should not allow withdrawal if not enough balance', async () => {
        await expect(marketplace.withdraw(OWNER, 100000000)).to.be.revertedWithCustomError(
          marketplace,
          'NotEnoughBalance',
        );
      });
    });
  });
});
