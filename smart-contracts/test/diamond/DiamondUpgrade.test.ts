import {
  DelegateRegistry,
  LumerinDiamond,
  Marketplace,
  ModelRegistry,
  MorpheusToken,
  ProviderRegistry,
  SessionRouterLegacy,
} from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { expect } from 'chai';
import hre, { ethers, network } from 'hardhat';

import { getHex, wei } from '@/scripts/utils/utils';
import {
  IDelegateStakingFunding__factory,
  IDelegateStaking__factory,
  IDelegation__factory,
  IMarketplace__factory,
  IBidStorage__factory,
  IModelRegistry__factory,
  IProviderPayouts__factory,
  IProviderRegistry__factory,
  ISessionRouterLegacy__factory,
  IStatsStorage__factory,
} from '@/generated-types/ethers';
import {
  deployDelegateRegistry,
  deployFacetDelegation,
  deployFacetMarketplace,
  deployFacetModelRegistry,
  deployFacetProviderRegistry,
  deployFacetSessionRouterLegacy,
  deployLumerinDiamond,
  deployMORToken,
  getSelectors,
  upgradeDiamond,
} from '@/test/helpers/deployers';
import { payoutStart } from '@/test/helpers/pool-helper';
import { Reverter } from '@/test/helpers/reverter';
import { getCurrentBlockTime, setTime } from '@/utils/block-helper';
import { getProviderApproval, getReceipt } from '@/utils/provider-helper';
import { DAY } from '@/utils/time';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const toPlain = (value: any) =>
  JSON.parse(JSON.stringify(value, (_, v) => (typeof v === 'bigint' ? v.toString() : v)));

describe('DiamondUpgrade', () => {
  const reverter = new Reverter();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let initialSnapshot: any;

  let OWNER: SignerWithAddress;
  let USER: SignerWithAddress;
  let FUNDING: SignerWithAddress;
  let PROVIDER: SignerWithAddress;
  let HOT: SignerWithAddress;
  let FUNDER: SignerWithAddress;
  let STRANGER: SignerWithAddress;

  let diamond: LumerinDiamond;
  let marketplace: Marketplace;
  let modelRegistry: ModelRegistry;
  let providerRegistry: ProviderRegistry;
  let router: SessionRouterLegacy;
  let legacyFacet: string;

  let token: MorpheusToken;
  let delegateRegistry: DelegateRegistry;

  const baseModelId = getHex(Buffer.from('1'));
  const bidPricePerSecond = wei(0.0001);
  let modelId = '';
  let bidId = '';

  // Day 10 after payoutStart, 00:00 UTC.
  const D = payoutStart - (payoutStart % DAY) + 10 * DAY;
  // First post-upgrade action; the upgrade itself happens in between.
  const T = D + DAY + 1000;

  let nonce = 0;
  const sessionIds: string[] = [];
  // S0: stake-pool, open at upgrade; S1: closed early, on-hold released at upgrade; S2: direct-pay, open;
  // S3: disputed same-day close, provider remainder unclaimed; S4: claimed by provider while open;
  // S5: disputed next-day close, provider partially paid, user on-hold still locked at upgrade.
  let S0: string, S1: string, S2: string, S3: string, S4: string, S5: string;

  const openAt = async (time: number, isDirectPaymentFromUser = false) => {
    await setTime(time);
    const { msg, signature } = await getProviderApproval(PROVIDER, USER, bidId);
    await router.connect(USER).openSession(USER, wei(50), isDirectPaymentFromUser, msg, signature);
    const sessionId = await router.getSessionId(USER, PROVIDER, bidId, nonce++);
    sessionIds.push(sessionId);

    return sessionId;
  };

  const closeAt = async (time: number, sessionId: string, isDispute = false) => {
    await setTime(time);
    const { msg } = await getReceipt(PROVIDER, sessionId, 10, 100);
    const { signature } = await getReceipt(isDispute ? OWNER : PROVIDER, sessionId, 10, 100);
    await router.connect(USER).closeSession(msg, signature);
  };

  const getBalances = async () =>
    Promise.all([USER, PROVIDER, FUNDING, STRANGER, diamond].map((a) => token.balanceOf(a)));

  const readState = async () => {
    const now = await getCurrentBlockTime();

    return toPlain({
      sessions: await Promise.all(sessionIds.map((id) => router.getSession(id))),
      provider: await router.getProvider(PROVIDER),
      providersTotalClaimed: await router.getProvidersTotalClaimed(),
      onHold: await router.getUserStakesOnHold(USER, 20),
      providerModelStats: await router.getProviderModelStats(modelId, PROVIDER),
      modelStats: await router.getModelStats(modelId),
      computeBalance: await router.getComputeBalance(now),
      totalMORSupply: await router.totalMORSupply(now),
      userSessions: await router.getUserSessions(USER, 0, 100),
      providerSessions: await router.getProviderSessions(PROVIDER, 0, 100),
      modelSessions: await router.getModelSessions(modelId, 0, 100),
    });
  };

  // Same calls, same timestamps: run once on the legacy code and once after the cut.
  const runPostUpgradeFlow = async () => {
    const records = [];
    const step = async (name: string, action: () => Promise<unknown>) => {
      const before = await getBalances();
      await action();
      const after = await getBalances();
      records.push({
        name,
        deltas: toPlain(after.map((b, i) => b - before[i])),
        state: await readState(),
      });
    };

    await step('close S0, late', () => closeAt(T, S0));
    await step('close S2, direct-pay, late', () => closeAt(T + 100, S2));
    await step('claim S3 remainder after disputed close', async () => {
      await setTime(T + 200);
      await router.connect(PROVIDER).claimForProvider(S3);
    });
    await step('claim S5 while the provider part is on hold', async () => {
      await setTime(T + 300);
      await router.connect(PROVIDER).claimForProvider(S5);
    });
    await step('close S4 claimed before the upgrade', () => closeAt(T + 400, S4));
    await step('withdraw released on-hold entries', async () => {
      await setTime(T + 500);
      await router.connect(USER).withdrawUserStakes(USER, 20);
    });
    await step('claim S5 remainder after the hold', async () => {
      await setTime(D + 2 * DAY + 100);
      await router.connect(PROVIDER).claimForProvider(S5);
    });
    await step('withdraw the last on-hold entry', async () => {
      await setTime(D + 2 * DAY + 200);
      await router.connect(USER).withdrawUserStakes(USER, 20);
    });
    await step('open and close a new session on the old bid', async () => {
      const sessionId = await openAt(D + 2 * DAY + 300);
      await closeAt(D + 2 * DAY + 600, sessionId);
    });

    return records;
  };

  const withSnapshot = async <R>(fn: () => Promise<R>): Promise<R> => {
    const snapshotId = await network.provider.send('evm_snapshot', []);
    const sessionsLength = sessionIds.length;
    const nonceBefore = nonce;
    const result = await fn();
    await network.provider.send('evm_revert', [snapshotId]);
    sessionIds.length = sessionsLength;
    nonce = nonceBefore;

    return result;
  };

  before(async () => {
    // this suite moves the chain time in before(); restore it for the suites that run next
    initialSnapshot = await network.provider.send('evm_snapshot', []);

    [OWNER, USER, FUNDING, PROVIDER, HOT, FUNDER, STRANGER] = await ethers.getSigners();

    [diamond, token, delegateRegistry] = await Promise.all([
      deployLumerinDiamond(),
      deployMORToken(),
      deployDelegateRegistry(),
    ]);

    [providerRegistry, modelRegistry, router, marketplace] = await Promise.all([
      deployFacetProviderRegistry(diamond),
      deployFacetModelRegistry(diamond),
      deployFacetSessionRouterLegacy(diamond, FUNDING),
      deployFacetMarketplace(diamond, token, wei(0.0001), wei(900)),
      deployFacetDelegation(diamond, delegateRegistry),
    ]);
    legacyFacet = await diamond.facetAddress(ISessionRouterLegacy__factory.createInterface().getFunction('closeSession').selector);

    for (const account of [USER, PROVIDER, FUNDING, HOT, FUNDER]) {
      await token.transfer(account, wei(10000));
      await token.connect(account).approve(diamond, wei(10000));
    }

    await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(1), 'test');
    await modelRegistry
      .connect(PROVIDER)
      .modelRegister(PROVIDER, baseModelId, getHex(Buffer.from('ipfs://ipfsaddress')), 0, wei(100), 'name', ['tag_1']);
    modelId = await modelRegistry.getModelId(PROVIDER, baseModelId);
    await marketplace.connect(PROVIDER).postModelBid(PROVIDER, modelId, bidPricePerSecond);
    bidId = await marketplace.getBidId(PROVIDER, modelId, 0);

    S0 = await openAt(D + 3600);
    S1 = await openAt(D + 3700);
    S2 = await openAt(D + 3800, true);
    S3 = await openAt(D + 3900);
    S4 = await openAt(D + 4000);
    await closeAt(D + 4300, S1);
    await closeAt(D + 4500, S3, true);
    await setTime(D + 6000);
    await router.connect(PROVIDER).claimForProvider(S4);
    S5 = await openAt(D + DAY - 600);
    await closeAt(D + DAY + 100, S5, true);

    await reverter.snapshot();
  });

  after(async () => {
    await network.provider.send('evm_revert', [initialSnapshot]);
  });

  afterEach(async () => {
    await reverter.revert();
    // JS-side mirrors of the chain state created in before()
    nonce = 6;
    sessionIds.length = 6;
  });

  describe('pre-upgrade fixture', () => {
    it('should hold sessions of both payment modes in every pre-upgrade state', async () => {
      const [s0, s1, s2, s3, s4, s5] = await Promise.all([S0, S1, S2, S3, S4, S5].map((id) => router.getSession(id)));

      expect(s0.isActive).to.eq(true);
      expect(s2.isActive).to.eq(true);
      expect(s2.isDirectPaymentFromUser).to.eq(true);
      expect(s1.closedAt).to.eq(D + 4301);
      expect(s3.closeoutType).to.eq(1);
      expect(s3.providerWithdrawnAmount).to.eq(0);
      expect(s4.isActive).to.eq(true);
      expect(s4.providerWithdrawnAmount).to.eq((s4.endsAt - s4.openedAt) * bidPricePerSecond);
      expect(s5.closeoutType).to.eq(1);
      expect(s5.providerWithdrawnAmount).to.eq((BigInt(D + DAY) - s5.openedAt) * bidPricePerSecond);

      const onHold = await router.getUserStakesOnHold(USER, 20);
      expect(onHold.available_).to.be.greaterThan(0);
      expect(onHold.hold_).to.be.greaterThan(0);
      expect((await router.getProvider(PROVIDER)).limitPeriodEarned).to.be.greaterThan(0);
    });
  });

  describe('#diamondCut', () => {
    it('should move every legacy selector off the legacy facet', async () => {
      const facets = await upgradeDiamond(diamond);

      const legacySelectors = [
        ...getSelectors(ISessionRouterLegacy__factory.createInterface()),
        ...getSelectors(IStatsStorage__factory.createInterface()),
      ];
      const movedViews = getSelectors(IProviderPayouts__factory.createInterface());

      for (const selector of legacySelectors) {
        const facet = await diamond.facetAddress(selector);
        expect(facet).to.eq(
          movedViews.includes(selector) ? facets.facetAddresses.providerPayouts : facets.facetAddresses.sessionRouter,
        );
      }
      expect(await diamond.facetAddresses()).to.not.include(legacyFacet);
    });
    it('should add the new selectors and keep the other facets', async () => {
      const otherSelectorsBefore = await Promise.all(
        [
          ...getSelectors(IProviderRegistry__factory.createInterface()),
          ...getSelectors(IModelRegistry__factory.createInterface()),
          ...getSelectors(IMarketplace__factory.createInterface()),
          ...getSelectors(IBidStorage__factory.createInterface()),
          ...getSelectors(IDelegation__factory.createInterface()),
        ].map((s) => diamond.facetAddress(s)),
      );

      const facets = await upgradeDiamond(diamond);

      for (const selector of getSelectors(IDelegateStaking__factory.createInterface())) {
        expect(await diamond.facetAddress(selector)).to.eq(facets.facetAddresses.delegateStaking);
      }
      for (const selector of getSelectors(IDelegateStakingFunding__factory.createInterface())) {
        expect(await diamond.facetAddress(selector)).to.eq(facets.facetAddresses.delegateStakingFunding);
      }
      for (const selector of getSelectors(IProviderPayouts__factory.createInterface())) {
        expect(await diamond.facetAddress(selector)).to.eq(facets.facetAddresses.providerPayouts);
      }

      const otherSelectorsAfter = await Promise.all(
        [
          ...getSelectors(IProviderRegistry__factory.createInterface()),
          ...getSelectors(IModelRegistry__factory.createInterface()),
          ...getSelectors(IMarketplace__factory.createInterface()),
          ...getSelectors(IBidStorage__factory.createInterface()),
          ...getSelectors(IDelegation__factory.createInterface()),
        ].map((s) => diamond.facetAddress(s)),
      );
      expect(otherSelectorsAfter).to.deep.eq(otherSelectorsBefore);
    });
    it('should initialize ProviderPayouts in the cut transaction only once', async () => {
      const { providerPayouts } = await upgradeDiamond(diamond);
      const cutTime = await getCurrentBlockTime();

      expect(await providerPayouts.getProviderLifetimeClaimed(PROVIDER)).to.deep.eq([0n, cutTime]);
      await expect(providerPayouts.__ProviderPayouts_init()).to.be.rejectedWith(
        'Initializable: contract is already initialized',
      );
      await expect(router.__SessionRouter_init(FUNDING, DAY, [])).to.be.rejectedWith(
        'Initializable: contract is already initialized',
      );
    });
  });

  describe('facet bytecode size', () => {
    it('should keep every new facet within the EIP-170 limit', async function () {
      // instrumented coverage builds are larger by design
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      if ((hre as any).__SOLIDITY_COVERAGE_RUNNING) this.skip();

      const { facetAddresses } = await upgradeDiamond(diamond);
      for (const address of Object.values(facetAddresses)) {
        const code = await ethers.provider.getCode(address);
        expect((code.length - 2) / 2).to.be.lessThanOrEqual(24_576);
      }
    });
  });

  describe('views and storage at the upgrade boundary', () => {
    it('should return the same values from all legacy views', async () => {
      const ts = [BigInt(D + DAY + 200), BigInt(D + 2 * DAY)];
      const readViews = async () =>
        toPlain({
          state: await readState(),
          perTimestamp: await Promise.all(
            ts.map(async (t) => [
              await router.getComputeBalance(t),
              await router.totalMORSupply(t),
              await router.getTodaysBudget(t),
              await router.stakeToStipend(wei(50), t),
              await router.stipendToStake(wei(1), t),
              await router.getSessionEnd(wei(50), bidPricePerSecond, t),
              await router.startOfTheDay(t),
            ]),
          ),
          sessionId: await router.getSessionId(USER, PROVIDER, bidId, 42),
          pools: await router.getPools(),
          pool: await router.getPool(3),
          fundingAccount: await router.getFundingAccount(),
          maxSessionDuration: await router.getMaxSessionDuration(),
          totalSessions: await router.getTotalSessions(PROVIDER),
          approvalUsed: await router.getIsProviderApprovalUsed('0x'),
        });

      const before = await readViews();
      const balancesBefore = await getBalances();
      await upgradeDiamond(diamond);

      expect(await readViews()).to.deep.eq(before);
      expect(await getBalances()).to.deep.eq(balancesBefore);
    });
    it('should keep the storage of existing facets intact', async () => {
      const readStorage = async () =>
        toPlain({
          provider: await providerRegistry.getProvider(PROVIDER),
          providerMinimumStake: await providerRegistry.getProviderMinimumStake(),
          activeProviders: await providerRegistry.getActiveProviders(0, 10),
          model: await modelRegistry.getModel(modelId),
          modelIds: await modelRegistry.getModelIds(0, 10),
          activeModelIds: await modelRegistry.getActiveModelIds(0, 10),
          bid: await marketplace.getBid(bidId),
          providerBids: await marketplace.getProviderBids(PROVIDER, 0, 10),
          modelActiveBids: await marketplace.getModelActiveBids(modelId, 0, 10),
          bidFee: await marketplace.getBidFee(),
          feeBalance: await marketplace.getFeeBalance(),
          minMaxPrice: await marketplace.getMinMaxBidPricePerSecond(),
          token: await marketplace.getToken(),
          registry: await marketplace.getRegistry(),
          owner: await diamond.owner(),
        });

      const before = await readStorage();
      await upgradeDiamond(diamond);

      expect(await readStorage()).to.deep.eq(before);
    });
  });

  describe('sessions opened before the upgrade', () => {
    it('should close, claim and withdraw with the same amounts as the legacy code', async () => {
      const legacy = await withSnapshot(runPostUpgradeFlow);

      await upgradeDiamond(diamond);
      const upgraded = await runPostUpgradeFlow();

      expect(upgraded).to.deep.eq(legacy);

      // The flow is not trivial: provider, user and on-hold paths all move money.
      const providerGains = legacy.map((r) => BigInt(r.deltas[1]));
      expect(providerGains.filter((x) => x > 0n).length).to.be.greaterThanOrEqual(4);
      expect(BigInt(legacy[5].deltas[0])).to.be.greaterThan(0);
      expect(BigInt(legacy[7].deltas[0])).to.be.greaterThan(0);
    });
    it('should record provider debt instead of reverting when funding is empty', async () => {
      const closeS0 = () => closeAt(T, S0);

      const funded = await withSnapshot(async () => {
        const before = await getBalances();
        await closeS0();
        const after = await getBalances();
        return { deltas: after.map((b, i) => b - before[i]), state: await readState() };
      });
      const providerPaid = funded.deltas[1];
      expect(providerPaid).to.be.greaterThan(0);

      await token.connect(FUNDING).approve(diamond, 0);
      await withSnapshot(async () => {
        await setTime(T);
        const { msg, signature } = await getReceipt(PROVIDER, S0, 10, 100);
        await expect(router.connect(USER).closeSession(msg, signature)).to.be.reverted;
      });

      const { providerPayouts, sessionRouter } = await upgradeDiamond(diamond);
      const before = await getBalances();
      await closeS0();
      const after = await getBalances();
      const deltas = after.map((b, i) => b - before[i]);

      // user side completes exactly as with funding; provider gets nothing now
      expect(deltas[0]).to.eq(funded.deltas[0]);
      expect(deltas[1]).to.eq(0);
      expect(deltas[2]).to.eq(0);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(providerPaid);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(providerPaid);
      expect(await router.getProvidersTotalClaimed()).to.eq(BigInt(funded.state.providersTotalClaimed) - providerPaid);

      // emission view is the same as if the provider had been paid
      const state = await readState();
      expect(state.computeBalance).to.eq(funded.state.computeBalance);
      expect(state.totalMORSupply).to.eq(funded.state.totalMORSupply);
      expect(state.sessions).to.deep.eq(funded.state.sessions);
      expect(state.provider).to.deep.eq(funded.state.provider);

      await token.connect(FUNDING).approve(diamond, wei(10000));
      await expect(sessionRouter.connect(STRANGER).claimForProvider(S0)).to.changeTokenBalances(
        token,
        [PROVIDER, FUNDING, STRANGER],
        [providerPaid, -providerPaid, 0],
      );
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(0);
      expect(await router.getProvidersTotalClaimed()).to.eq(funded.state.providersTotalClaimed);
    });
    it('should not treat legacy sessions as pool sessions', async () => {
      const { delegateStaking, delegateStakingFunding } = await upgradeDiamond(diamond);

      for (const sessionId of [S0, S1, S2, S3, S4, S5]) {
        expect((await delegateStakingFunding.getPoolSession(sessionId)).exists).to.eq(false);
      }

      await setTime(T);
      const { msg, signature } = await getReceipt(PROVIDER, S0, 10, 100);
      await expect(delegateStaking.connect(USER).closePoolSession(msg, signature)).to.be.revertedWithCustomError(
        delegateStaking,
        'DelegateStakingPoolSessionNotFound',
      );
      await expect(delegateStaking.connect(STRANGER).settleExpiredSession(S0)).to.be.revertedWithCustomError(
        delegateStaking,
        'DelegateStakingPoolSessionNotFound',
      );
      await expect(router.connect(USER).closeSession(msg, signature)).to.emit(router, 'SessionClosed');
    });
  });

  describe('sessions opened after the upgrade', () => {
    it('should continue sessionNonce and open regular and pool sessions on old bids and models', async () => {
      const { delegateStaking, delegateStakingFunding, providerPayouts } = await upgradeDiamond(diamond);

      const regular = await openAt(T);
      expect(regular).to.eq(await providerPayouts.getSessionId(USER, PROVIDER, bidId, 6));

      await delegateStakingFunding.connect(HOT).approvePoolFunder(HOT, FUNDER);
      await delegateStakingFunding.connect(FUNDER).grantStakingAllowance(HOT, wei(100), T + 30 * DAY);
      await delegateStakingFunding.connect(FUNDER).fundStakingAllowance(HOT, wei(100));

      await setTime(T + 100);
      const { msg, signature } = await getProviderApproval(PROVIDER, HOT, bidId);
      await delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), msg, signature);
      const pool = await providerPayouts.getSessionId(HOT, PROVIDER, bidId, 7);
      expect((await delegateStakingFunding.getPoolSession(pool)).exists).to.eq(true);

      const ids = [S0, S1, S2, S3, S4, S5, regular, pool];
      expect(new Set(ids).size).to.eq(ids.length);

      const poolSession = await router.getSession(pool);
      expect(poolSession.user).to.eq(HOT);
      expect(poolSession.bidId).to.eq(bidId);
      expect(poolSession.stake).to.eq(wei(50));
      expect(poolSession.isDirectPaymentFromUser).to.eq(false);

      await setTime(T + 400);
      const { msg: receipt, signature: receiptSig } = await getReceipt(PROVIDER, pool, 10, 100);
      await expect(delegateStaking.connect(HOT).closePoolSession(receipt, receiptSig)).to.changeTokenBalance(
        token,
        HOT,
        0,
      );

      await closeAt(T + 500, regular);
      expect((await router.getSession(regular)).isActive).to.eq(false);
      expect((await router.getSession(S0)).isActive).to.eq(true);
    });
  });
});

// npm run generate-types && npx hardhat test "test/diamond/DiamondUpgrade.test.ts"
