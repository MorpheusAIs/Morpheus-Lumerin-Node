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
import { Addressable, Fragment } from 'ethers';
import { ethers } from 'hardhat';

import { MAX_UINT8 } from '@/scripts/utils/constants';
import { getHex, randomBytes32, wei } from '@/scripts/utils/utils';
import { FacetAction } from '@/test/helpers/enums';
import { getDefaultPools } from '@/test/helpers/pool-helper';
import { Reverter } from '@/test/helpers/reverter';
import { getCurrentBlockTime, setTime } from '@/utils/block-helper';
import { getProviderApproval, getReport } from '@/utils/provider-helper';
import { DAY, HOUR, YEAR } from '@/utils/time';

describe('Session closeout', () => {
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
    [
      IBidStorage.BidStruct & {
        id: string;
        modelId: string;
      },
      {
        durationSeconds: bigint;
        totalCost: bigint;
        pricePerSecond: bigint;
        user: SignerWithAddress;
        provider: SignerWithAddress;
        modelAgentId: any;
        bidID: string;
        stake: bigint;
      },
    ]
  > {
    let bid = {
      id: '',
      modelId: model.modelId,
      pricePerSecond: wei(0.0001),
      nonce: 0,
      createdAt: 0n,
      deletedAt: 0,
      provider: PROVIDER,
      modelAgentId: model.modelId,
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
      modelAgentId: bid.modelId,
      bidID: bid.id,
      stake: (totalCost * totalSupply) / todaysBudget,
    };

    await MOR.transfer(SECOND, expectedSession.stake);
    await MOR.connect(SECOND).approve(modelRegistry, expectedSession.stake);

    return [bid, expectedSession];
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

  describe('Actions', () => {
    let provider: IProviderStorage.ProviderStruct;
    let model: IModelStorage.ModelStruct & {
      modelId: string;
    };
    let bid: IBidStorage.BidStruct & {
      id: string;
      modelId: string;
    };
    let session: {
      durationSeconds: bigint;
      totalCost: bigint;
      pricePerSecond: bigint;
      user: SignerWithAddress;
      provider: SignerWithAddress;
      modelAgentId: any;
      bidID: string;
      stake: bigint;
    };

    beforeEach(async () => {
      provider = await deployProvider();
      model = await deployModel();
      [bid, session] = await deployBid(model);
    });

    it('should open short (<1D) session and close after expiration', async () => {
      // open session
      const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);

      const sessionId = await sessionRouter.connect(SECOND).openSession.staticCall(session.stake, msg, signature);
      await sessionRouter.connect(SECOND).openSession(session.stake, msg, signature);

      await setTime(Number((await getCurrentBlockTime()) + session.durationSeconds * 2n));

      const userBalanceBefore = await MOR.balanceOf(SECOND);
      const providerBalanceBefore = await MOR.balanceOf(PROVIDER);

      // close session
      const report = await getReport(PROVIDER, sessionId, 10, 1000);
      await sessionRouter.connect(SECOND).closeSession(report.msg, report.sig);

      // verify session is closed without dispute
      const sessionData = await sessionRouter.getSession(sessionId);
      expect(sessionData.closeoutType).to.equal(0n);

      // verify balances
      const userBalanceAfter = await MOR.balanceOf(SECOND);
      const providerBalanceAfter = await MOR.balanceOf(PROVIDER);

      const userStakeReturned = userBalanceAfter - userBalanceBefore;
      const providerEarned = providerBalanceAfter - providerBalanceBefore;

      const totalPrice = (sessionData.endsAt - sessionData.openedAt) * session.pricePerSecond;

      expect(userStakeReturned).to.closeTo(0, Number(session.pricePerSecond) * 5);
      expect(providerEarned).to.closeTo(totalPrice, 1);
    });

    it('should open short (<1D) session and close early', async () => {
      // open session
      const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
      const sessionId = await sessionRouter.connect(SECOND).openSession.staticCall(session.stake, msg, signature);
      await sessionRouter.connect(SECOND).openSession(session.stake, msg, signature);

      await setTime(Number((await getCurrentBlockTime()) + session.durationSeconds / 2n - 1n));

      const userBalanceBefore = await MOR.balanceOf(SECOND);
      const providerBalanceBefore = await MOR.balanceOf(PROVIDER);

      // close session
      const report = await getReport(PROVIDER, sessionId, 10, 1000);
      await sessionRouter.connect(SECOND).closeSession(report.msg, report.sig);

      // verify session is closed without dispute
      const sessionData = await sessionRouter.getSession(sessionId);
      expect(sessionData.closeoutType).to.equal(0n);

      // verify balances
      const userBalanceAfter = await MOR.balanceOf(SECOND);
      const providerBalanceAfter = await MOR.balanceOf(PROVIDER);

      const userStakeReturned = userBalanceAfter - userBalanceBefore;
      const providerEarned = providerBalanceAfter - providerBalanceBefore;

      expect(userStakeReturned).to.closeTo(session.stake / 2n, Number(session.pricePerSecond) * 5);
      expect(providerEarned).to.closeTo(session.totalCost / 2n, 1);
    });

    it('should open and close early with user report - dispute', async () => {
      // open session
      const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
      const sessionId = await sessionRouter.connect(SECOND).openSession.staticCall(session.stake, msg, signature);
      await sessionRouter.connect(SECOND).openSession(session.stake, msg, signature);

      // wait half of the session
      await setTime(Number((await getCurrentBlockTime()) + session.durationSeconds / 2n - 1n));

      const userBalanceBefore = await MOR.balanceOf(SECOND);
      const providerBalanceBefore = await MOR.balanceOf(PROVIDER);

      // close session with user signature
      const report = await getReport(SECOND, sessionId, 10, 1000);
      await sessionRouter.connect(SECOND).closeSession(report.msg, report.sig);

      // verify session is closed with dispute
      const sessionData = await sessionRouter.getSession(sessionId);
      const totalCost = sessionData.pricePerSecond * (sessionData.closedAt - sessionData.openedAt);

      // verify balances
      const userBalanceAfter = await MOR.balanceOf(SECOND);
      const providerBalanceAfter = await MOR.balanceOf(PROVIDER);

      const claimableProvider = await sessionRouter.getProviderClaimableBalance(sessionData.id);

      const [userAvail, userHold] = await sessionRouter.withdrawableUserStake(sessionData.user, MAX_UINT8);

      expect(sessionData.closeoutType).to.equal(1n);
      expect(providerBalanceAfter - providerBalanceBefore).to.equal(0n);
      expect(claimableProvider).to.equal(0n);
      expect(session.stake / 2n).to.closeTo(userBalanceAfter - userBalanceBefore, 1);
      expect(userAvail).to.equal(0n);
      expect(userHold).to.closeTo(userBalanceAfter - userBalanceBefore, 1);

      // verify provider balance after dispute is released
      await setTime(Number((await getCurrentBlockTime()) + DAY));
      const claimableProvider2 = await sessionRouter.getProviderClaimableBalance(sessionId);
      expect(claimableProvider2).to.equal(totalCost);

      // claim provider balance
      await sessionRouter.claimProviderBalance(sessionId, claimableProvider2);

      // verify provider balance after claim
      const providerBalanceAfterClaim = await MOR.balanceOf(PROVIDER);
      const providerClaimed = providerBalanceAfterClaim - providerBalanceAfter;
      expect(providerClaimed).to.equal(totalCost);
    });

    it('should error when not a user trying to close', async () => {
      // open session
      const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
      const sessionId = await sessionRouter.connect(SECOND).openSession.staticCall(session.stake, msg, signature);
      await sessionRouter.connect(SECOND).openSession(session.stake, msg, signature);

      // wait half of the session
      await setTime(Number((await getCurrentBlockTime()) + session.durationSeconds / 2n - 1n));

      // close session with user signature
      const report = await getReport(SECOND, sessionId, 10, 10);

      await expect(sessionRouter.connect(THIRD).closeSession(report.msg, report.sig)).to.be.revertedWithCustomError(
        sessionRouter,
        'NotOwnerOrUser',
      );
    });

    it('should limit reward by stake amount', async () => {
      // expected bid
      const expectedBid = {
        id: '',
        providerAddr: await PROVIDER.getAddress(),
        modelId: model.modelId,
        pricePerSecond: wei('0.1'),
        nonce: 0n,
        createdAt: 0n,
        deletedAt: 0n,
      };

      // add single bid
      const postBidId = await marketplace
        .connect(PROVIDER)
        .postModelBid.staticCall(expectedBid.providerAddr, expectedBid.modelId, expectedBid.pricePerSecond);
      await marketplace
        .connect(PROVIDER)
        .postModelBid(expectedBid.providerAddr, expectedBid.modelId, expectedBid.pricePerSecond);

      expectedBid.id = postBidId;
      expectedBid.createdAt = await getCurrentBlockTime();

      // calculate data for session opening
      const totalCost = BigInt(provider.stake) * 2n;
      const durationSeconds = totalCost / expectedBid.pricePerSecond;
      const totalSupply = await sessionRouter.totalMORSupply(await getCurrentBlockTime());
      const todaysBudget = await sessionRouter.getTodaysBudget(await getCurrentBlockTime());

      const expectedSession = {
        durationSeconds,
        totalCost,
        pricePerSecond: expectedBid.pricePerSecond,
        user: await SECOND.getAddress(),
        provider: expectedBid.providerAddr,
        modelAgentId: expectedBid.modelId,
        bidID: expectedBid.id,
        stake: (totalCost * totalSupply) / todaysBudget,
      };

      // set user balance and approve funds
      await MOR.transfer(SECOND, expectedSession.stake);
      await MOR.connect(SECOND).approve(modelRegistry, expectedSession.stake);

      // open session
      const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), expectedSession.bidID);
      const sessionId = await sessionRouter
        .connect(SECOND)
        .openSession.staticCall(expectedSession.stake, msg, signature);
      await sessionRouter.connect(SECOND).openSession(expectedSession.stake, msg, signature);

      // wait till session ends
      await setTime(Number((await getCurrentBlockTime()) + expectedSession.durationSeconds));

      const providerBalanceBefore = await MOR.balanceOf(PROVIDER);
      // close session without dispute
      const report = await getReport(PROVIDER, sessionId, 10, 1000);
      await sessionRouter.connect(SECOND).closeSession(report.msg, report.sig);

      const providerBalanceAfter = await MOR.balanceOf(PROVIDER);

      const providerEarned = providerBalanceAfter - providerBalanceBefore;

      expect(providerEarned).to.equal(provider.stake);

      // check provider record if earning was updated
      const providerRecord = await providerRegistry.getProvider(PROVIDER);
      expect(providerRecord.limitPeriodEarned).to.equal(provider.stake);
    });

    it('should reset provider limitPeriodEarned after period', async () => {});

    it('should error with WithdrawableBalanceLimitByStakeReached() if claiming more that stake for a period', async () => {});
  });
});
