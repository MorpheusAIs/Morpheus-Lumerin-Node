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
import { Addressable, Fragment, randomBytes, resolveAddress } from 'ethers';
import { ethers } from 'hardhat';

import { getHex, randomBytes32, startOfTheDay, wei } from '@/scripts/utils/utils';
import { FacetAction } from '@/test/helpers/enums';
import { getDefaultPools } from '@/test/helpers/pool-helper';
import { Reverter } from '@/test/helpers/reverter';
import { getCurrentBlockTime, setTime } from '@/utils/block-helper';
import { getProviderApproval, getReport } from '@/utils/provider-helper';
import { DAY, HOUR, YEAR } from '@/utils/time';

describe('session actions', () => {
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
        modelId: any;
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

    return [bid, expectedSession];
  }

  async function openSession(session: any) {
    // open session
    const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
    const sessionId = await sessionRouter.connect(SECOND).openSession.staticCall(session.stake, msg, signature);
    await sessionRouter.connect(SECOND).openSession(session.stake, msg, signature);

    return sessionId;
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
      modelId: any;
      bidID: string;
      stake: bigint;
    };

    beforeEach(async () => {
      provider = await deployProvider();
      model = await deployModel();
      [bid, session] = await deployBid(model);
    });

    describe('positive cases', () => {
      it('should open session without error', async () => {
        const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        const sessionId = await sessionRouter.connect(SECOND).openSession.staticCall(session.stake, msg, signature);
        await sessionRouter.connect(SECOND).openSession(session.stake, msg, signature);

        expect(sessionId).to.be.a('string');
      });

      it('should emit SessionOpened event', async () => {
        const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);

        const sessionId = await sessionRouter.connect(SECOND).openSession.staticCall(session.stake, msg, signature);
        await expect(sessionRouter.connect(SECOND).openSession(session.stake, msg, signature))
          .to.emit(sessionRouter, 'SessionOpened')
          .withArgs(session.user, sessionId, session.provider);
      });

      it('should verify session fields after opening', async () => {
        const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        const sessionId = await sessionRouter.connect(SECOND).openSession.staticCall(session.stake, msg, signature);
        await sessionRouter.connect(SECOND).openSession(session.stake, msg, signature);

        const sessionData = await sessionRouter.sessions(sessionId);
        const createdAt = await getCurrentBlockTime();

        expect(sessionData).to.deep.equal([
          sessionId,
          await resolveAddress(session.user),
          await resolveAddress(session.provider),
          session.modelId,
          session.bidID,
          session.stake,
          session.pricePerSecond,
          getHex(Buffer.from(''), 0),
          0n,
          0n,
          createdAt,
          sessionData.endsAt, // skipped in this test
          0n,
        ]);
      });

      it('should verify balances after opening', async () => {
        const srBefore = await MOR.balanceOf(sessionRouter);
        const userBefore = await MOR.balanceOf(SECOND);

        const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        await sessionRouter.connect(SECOND).openSession(session.stake, msg, signature);

        const srAfter = await MOR.balanceOf(sessionRouter);
        const userAfter = await MOR.balanceOf(SECOND);

        expect(srAfter - srBefore).to.equal(session.stake);
        expect(userBefore - userAfter).to.equal(session.stake);
      });

      it('should allow opening two sessions in the same block', async () => {
        await MOR.transfer(SECOND, session.stake * 2n);
        await MOR.connect(SECOND).approve(sessionRouter, session.stake * 2n);

        const apprv1 = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        await setTime(Number(await getCurrentBlockTime()) + 1);
        const apprv2 = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);

        await ethers.provider.send('evm_setAutomine', [false]);

        await sessionRouter.connect(SECOND).openSession(session.stake, apprv1.msg, apprv1.signature);
        await sessionRouter.connect(SECOND).openSession(session.stake, apprv2.msg, apprv2.signature);

        await ethers.provider.send('evm_setAutomine', [true]);
        await ethers.provider.send('evm_mine', []);

        const sessionId1 = await sessionRouter.getSessionId(SECOND, PROVIDER, session.stake, 0);
        const sessionId2 = await sessionRouter.getSessionId(SECOND, PROVIDER, session.stake, 1);

        expect(sessionId1).not.to.equal(sessionId2);

        const session1 = await sessionRouter.sessions(sessionId1);
        const session2 = await sessionRouter.sessions(sessionId2);

        expect(session1.stake).to.equal(session.stake);
        expect(session2.stake).to.equal(session.stake);
      });

      it('should partially use remaining staked tokens for the opening session', async () => {
        const sessionId = await openSession(session);
        await setTime(Number((await getCurrentBlockTime()) + session.durationSeconds / 2n));

        // close session
        const report = await getReport(PROVIDER, sessionId, 10, 10);
        await sessionRouter.connect(SECOND).closeSession(report.msg, report.sig);

        await setTime(Number(startOfTheDay(await getCurrentBlockTime()) + DAY));

        const [avail] = await sessionRouter.withdrawableUserStake(SECOND, 255);
        expect(avail > 0).to.be.true;

        // reset allowance
        await MOR.connect(SECOND).approve(sessionRouter, 0n);

        const stake = avail / 2n;

        const approval = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        await sessionRouter.connect(SECOND).openSession(stake, approval.msg, approval.signature);

        const [avail2] = await sessionRouter.withdrawableUserStake(SECOND, 255);
        expect(avail2).to.be.equal(stake);
      });

      it('should use all remaining staked tokens for the opening session', async () => {
        const sessionId = await openSession(session);
        await setTime(Number((await getCurrentBlockTime()) + session.durationSeconds / 2n));

        // close session
        const report = await getReport(PROVIDER, sessionId, 10, 10);
        await sessionRouter.connect(SECOND).closeSession(report.msg, report.sig);

        await setTime(Number(startOfTheDay(await getCurrentBlockTime()) + DAY));

        const [avail] = await sessionRouter.withdrawableUserStake(SECOND, 255);
        expect(avail > 0).to.be.true;

        // reset allowance
        await MOR.connect(SECOND).approve(sessionRouter, 0n);

        const approval = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        await sessionRouter.connect(SECOND).openSession(avail, approval.msg, approval.signature);

        const [avail2] = await sessionRouter.withdrawableUserStake(SECOND, 255);
        expect(avail2).to.be.equal(0n);
      });

      it('should use remaining staked tokens and allowance for opening session', async () => {
        const sessionId = await openSession(session);
        await setTime(Number((await getCurrentBlockTime()) + session.durationSeconds / 2n));

        // close session
        const report = await getReport(PROVIDER, sessionId, 10, 10);
        await sessionRouter.connect(SECOND).closeSession(report.msg, report.sig);

        await setTime(Number(startOfTheDay(await getCurrentBlockTime()) + DAY));

        const [avail] = await sessionRouter.withdrawableUserStake(SECOND, 255);
        expect(avail > 0).to.be.true;

        const allowancePart = 1000n;
        const balanceBefore = await MOR.balanceOf(SECOND);

        // reset allowance
        await MOR.connect(SECOND).approve(sessionRouter, allowancePart);

        const approval = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        await sessionRouter.connect(SECOND).openSession(avail + allowancePart, approval.msg, approval.signature);

        // check all onHold used
        const [avail2] = await sessionRouter.withdrawableUserStake(SECOND, 255);
        expect(avail2).to.be.equal(0n);

        // check allowance used
        const balanceAfter = await MOR.balanceOf(SECOND);
        expect(balanceBefore - balanceAfter).to.be.equal(allowancePart);
      });
    });

    describe('negative cases', () => {
      it('should error when approval generated for a different user', async () => {
        const { msg, signature } = await getProviderApproval(PROVIDER, await THIRD.getAddress(), session.bidID);

        await expect(
          sessionRouter.connect(SECOND).openSession(session.stake, msg, signature),
        ).to.be.revertedWithCustomError(sessionRouter, 'ApprovedForAnotherUser');
      });

      it('should error when approval expired', async () => {
        const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        const ttl = await sessionRouter.SIGNATURE_TTL();
        await setTime(Number((await getCurrentBlockTime()) + ttl) + 1);

        await expect(
          sessionRouter.connect(SECOND).openSession(session.stake, msg, signature),
        ).to.be.revertedWithCustomError(sessionRouter, 'SignatureExpired');
        sessionRouter.openSession(session.stake, msg, signature);
      });

      it('should error when bid not exist', async () => {
        const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), randomBytes32());
        await expect(
          sessionRouter.connect(SECOND).openSession(session.stake, msg, signature),
        ).to.be.revertedWithCustomError(sessionRouter, 'BidNotFound');
      });

      it('should error when bid is deleted', async () => {
        await marketplace.connect(PROVIDER).deleteModelBid(session.bidID);

        const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        await expect(
          sessionRouter.connect(SECOND).openSession(session.stake, msg, signature),
        ).to.be.revertedWithCustomError(sessionRouter, 'BidNotFound');
      });

      it('should error when signature has invalid length', async () => {
        const { msg } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);

        await expect(sessionRouter.connect(SECOND).openSession(session.stake, msg, '0x00')).to.be.revertedWith(
          'ECDSA: invalid signature length',
        );
      });

      it('should error when signature is invalid', async () => {
        const { msg } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        const sig = randomBytes(65);

        await expect(sessionRouter.connect(SECOND).openSession(session.stake, msg, sig)).to.be.reverted;
      });

      it('should error when opening two bids with same signature', async () => {
        const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        await sessionRouter.connect(SECOND).openSession(session.stake, msg, signature);

        await approveUserFunds(session.stake);

        await expect(
          sessionRouter.connect(SECOND).openSession(session.stake, msg, signature),
        ).to.be.revertedWithCustomError(sessionRouter, 'DuplicateApproval');
      });

      it('should not error when opening two bids same time', async () => {
        const appr1 = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        await sessionRouter.connect(SECOND).openSession(session.stake, appr1.msg, appr1.signature);

        await approveUserFunds(session.stake);
        const appr2 = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        await sessionRouter.connect(SECOND).openSession(session.stake, appr2.msg, appr2.signature);
      });

      it('should error with insufficient allowance', async () => {
        const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        await expect(sessionRouter.connect(SECOND).openSession(session.stake * 2n, msg, signature)).to.be.revertedWith(
          'ERC20: insufficient allowance',
        );
      });

      it('should error with insufficient allowance', async () => {
        const stake = (await MOR.balanceOf(SECOND)) + 1n;
        await MOR.connect(SECOND).approve(sessionRouter, stake);

        const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        await expect(sessionRouter.connect(SECOND).openSession(stake, msg, signature)).to.be.revertedWith(
          'ERC20: transfer amount exceeds balance',
        );
      });
    });

    describe('verify session end time', () => {
      it("session that doesn't span across midnight (1h)", async () => {
        const durationSeconds = HOUR;
        const stake = await getStake(durationSeconds, session.pricePerSecond);

        const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        const sessionId = await sessionRouter.connect(SECOND).openSession.staticCall(stake, msg, signature);
        await sessionRouter.connect(SECOND).openSession(stake, msg, signature);

        const sessionData = await sessionRouter.sessions(sessionId);

        expect(sessionData.endsAt).to.equal((await getCurrentBlockTime()) + durationSeconds);
      });

      it('session that spans across midnight (6h) should last 6h', async () => {
        const tomorrow9pm = startOfTheDay(await getCurrentBlockTime()) + DAY + 21n * HOUR;
        await setTime(Number(tomorrow9pm));

        // the stake is enough to cover the first day (3h till midnight) and the next day (< 6h)
        const durationSeconds = 6n * HOUR;
        const stake = await getStake(durationSeconds, session.pricePerSecond);
        await approveUserFunds(stake);

        const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        const sessionId = await sessionRouter.connect(SECOND).openSession.staticCall(stake, msg, signature);
        await sessionRouter.connect(SECOND).openSession(stake, msg, signature);

        const expEndsAt = (await getCurrentBlockTime()) + durationSeconds;
        const sessionData = await sessionRouter.sessions(sessionId);

        expect(sessionData.endsAt).closeTo(expEndsAt, 10);
      });

      it('session that lasts multiple days', async () => {
        const midnight = startOfTheDay(await getCurrentBlockTime()) + DAY;
        await setTime(Number(midnight));

        // the stake is enough to cover the whole day + extra 1h
        const durationSeconds = 25n * HOUR;
        const stake = await sessionRouter.stipendToStake(
          durationSeconds * session.pricePerSecond,
          await getCurrentBlockTime(),
        );

        await approveUserFunds(stake);

        const { msg, signature } = await getProviderApproval(PROVIDER, await SECOND.getAddress(), session.bidID);
        const sessionId = await sessionRouter.connect(SECOND).openSession.staticCall(stake, msg, signature);
        await sessionRouter.connect(SECOND).openSession(stake, msg, signature);

        const sessionData = await sessionRouter.sessions(sessionId);
        const durSeconds = Number(sessionData.endsAt - sessionData.openedAt);

        expect(durSeconds).to.equal(DAY);
      });
    });
  });

  async function approveUserFunds(amount: bigint) {
    await MOR.transfer(SECOND, amount);
    await MOR.connect(SECOND).approve(sessionRouter, amount);
  }

  async function getStake(durationSeconds: bigint, pricePerSecond: bigint): Promise<bigint> {
    const totalCost = pricePerSecond * durationSeconds;
    const totalSupply = await sessionRouter.totalMORSupply(await getCurrentBlockTime());
    const todaysBudget = await sessionRouter.getTodaysBudget(await getCurrentBlockTime());
    return (totalCost * totalSupply) / todaysBudget;
  }
});
