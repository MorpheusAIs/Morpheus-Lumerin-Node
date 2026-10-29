import {
  DelegateRegistry,
  DelegateStaking,
  DelegateStakingFunding,
  LumerinDiamond,
  Marketplace,
  ModelRegistry,
  MorpheusToken,
  ProviderPayouts,
  ProviderRegistry,
  SessionRouter,
} from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
import { ContractTransactionResponse, Wallet } from 'ethers';
import { expect } from 'chai';
import { ethers } from 'hardhat';

import { getHex, wei } from '@/scripts/utils/utils';
import {
  deployDelegateRegistry,
  deployFacetDelegateStaking,
  deployFacetDelegateStakingFunding,
  deployFacetDelegation,
  deployFacetMarketplace,
  deployFacetModelRegistry,
  deployFacetProviderPayouts,
  deployFacetProviderRegistry,
  deployFacetSessionRouter,
  deployLumerinDiamond,
  deployMORToken,
} from '@/test/helpers/deployers';
import { payoutStart } from '@/test/helpers/pool-helper';
import { Reverter } from '@/test/helpers/reverter';
import { getCurrentBlockTime, setTime } from '@/utils/block-helper';
import { getProviderApproval, getReceipt } from '@/utils/provider-helper';
import { DAY } from '@/utils/time';

// Arbitrum caps a single transaction at 32M gas; Base blocks are larger. Worst cases must stay well below.
const GAS_BUDGET = 10_000_000n;

const EXPECTED_REVERTS = [
  'DelegateStakingInsufficientLiquidity',
  'DelegateStakingInsufficientCapacity',
  'DelegateStakingFundingCapExceeded',
  'DelegateStakingPrincipalTooLow',
  'DelegateStakingAllowanceNotActive',
  'DelegateStakingWithdrawalNotFound',
  'DelegateStakingInsufficientPrincipal',
  'DelegateStakingSettlementNotReady',
  'SessionTooShort',
  'SessionUserAmountToWithdrawIsZero',
];

// Small deterministic PRNG (mulberry32) so failures reproduce from the seed.
function prng(seed: number) {
  let state = seed >>> 0;
  return (max: number) => {
    state = (state + 0x6d2b79f5) >>> 0;
    let t = state;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return (((t ^ (t >>> 14)) >>> 0) % max) as number;
  };
}

describe('Protocol properties', () => {
  const reverter = new Reverter();
  const price = wei(0.0001);
  const openTime = payoutStart + 10 * DAY;

  let OWNER: SignerWithAddress;
  let HOT_A: SignerWithAddress;
  let HOT_B: SignerWithAddress;
  let FUNDER_1: SignerWithAddress;
  let FUNDER_2: SignerWithAddress;
  let FUNDER_3: SignerWithAddress;
  let FUNDING: SignerWithAddress;
  let PROVIDER_1: SignerWithAddress;
  let PROVIDER_2: SignerWithAddress;
  let USER: SignerWithAddress;
  let KEEPER: SignerWithAddress;

  let diamond: LumerinDiamond;
  let token: MorpheusToken;
  let delegateRegistry: DelegateRegistry;
  let providerRegistry: ProviderRegistry;
  let modelRegistry: ModelRegistry;
  let marketplace: Marketplace;
  let sessionRouter: SessionRouter;
  let providerPayouts: ProviderPayouts;
  let delegateStaking: DelegateStaking;
  let delegateStakingFunding: DelegateStakingFunding;
  let bids: Map<string, string>;

  before(async () => {
    [OWNER, HOT_A, HOT_B, FUNDER_1, FUNDER_2, FUNDER_3, FUNDING, PROVIDER_1, PROVIDER_2, USER, KEEPER] =
      await ethers.getSigners();

    diamond = await deployLumerinDiamond();
    token = await deployMORToken();
    delegateRegistry = await deployDelegateRegistry();
    providerRegistry = await deployFacetProviderRegistry(diamond);
    modelRegistry = await deployFacetModelRegistry(diamond);
    sessionRouter = await deployFacetSessionRouter(diamond, FUNDING);
    marketplace = await deployFacetMarketplace(diamond, token, price, wei(900));
    await deployFacetDelegation(diamond, delegateRegistry);
    providerPayouts = await deployFacetProviderPayouts(diamond);
    delegateStaking = await deployFacetDelegateStaking(diamond);
    delegateStakingFunding = await deployFacetDelegateStakingFunding(diamond);

    for (const account of [FUNDER_1, FUNDER_2, FUNDER_3, FUNDING, PROVIDER_1, PROVIDER_2, USER]) {
      await token.transfer(account, wei(100000));
      await token.connect(account).approve(diamond, ethers.MaxUint256);
    }

    bids = new Map();
    for (const [index, provider] of [PROVIDER_1, PROVIDER_2].entries()) {
      // the second provider has a small stake so its reward limiter engages in the randomized runs
      await providerRegistry.connect(provider).providerRegister(provider, index === 0 ? wei(1000) : wei(0.2), 'provider');
      const baseModelId = getHex(Buffer.from(`properties-${index}`));
      await modelRegistry
        .connect(provider)
        .modelRegister(provider, baseModelId, getHex(Buffer.from('ipfs://properties')), 0, wei(100), 'model', []);
      const modelId = await modelRegistry.getModelId(provider, baseModelId);
      await marketplace.connect(provider).postModelBid(provider, modelId, price);
      bids.set(provider.address, await marketplace.getBidId(provider, modelId, 0));
    }

    await reverter.snapshot();
  });

  afterEach(reverter.revert);

  async function gasOf(tx: Promise<ContractTransactionResponse>) {
    const receipt = await (await tx).wait();
    return receipt!.gasUsed;
  }

  describe('randomized mixed operations', () => {
    interface OpenSession {
      id: string;
      provider: SignerWithAddress;
      hot?: SignerWithAddress;
      direct: boolean;
    }

    let limiterHits = 0;

    async function run(seed: number, steps: number) {
      const rand = prng(seed);
      const pick = <T>(items: T[]): T => items[rand(items.length)];
      const hots = [HOT_A, HOT_B];
      const funders = [FUNDER_1, FUNDER_2, FUNDER_3];
      const providers = [PROVIDER_1, PROVIDER_2];
      const paid = new Map<string, bigint>();
      const key = (hot: SignerWithAddress, funder: SignerWithAddress) => `${hot.address}:${funder.address}`;
      const openPool: OpenSession[] = [];
      const openRegular: OpenSession[] = [];
      const closedDirect: string[] = [];
      const counts = new Map<string, number>();
      const count = (name: string) => counts.set(name, (counts.get(name) ?? 0) + 1);
      let nonce = 0n;
      let afterRollover = false;

      await setTime(openTime);
      const emissionAt = BigInt(openTime);
      const emissionTotal =
        (await providerPayouts.getComputeBalance(emissionAt)) +
        (await sessionRouter.getProvidersTotalClaimed()) +
        (await providerPayouts.getPendingProviderClaims());
      const baseDiamondBalance = await token.balanceOf(diamond);
      const hotBalances = await Promise.all(hots.map((hot) => token.balanceOf(hot)));

      const send = async (tx: () => Promise<ContractTransactionResponse>) => {
        try {
          const receipt = await (await tx()).wait();
          for (const log of receipt!.logs) {
            const parsed = delegateStakingFunding.interface.parseLog(log);
            if (parsed?.name === 'DelegateStakingWithdrawalPaid') {
              const k = `${parsed.args.hot}:${parsed.args.funder}`;
              paid.set(k, (paid.get(k) ?? 0n) + parsed.args.amount);
              count('withdrawalPaid');
            }
            if (parsed?.name === 'DelegateStakingPoolLockReleased') count('lockReleased');
            const payout = providerPayouts.interface.parseLog(log);
            if (payout?.name === 'ProviderRewardAccrued' && payout.args.owed !== 0n) count('debtAccrued');
            if (afterRollover && payout?.name === 'ProviderRewardAccrued' && payout.args.provider === PROVIDER_2.address) {
              count('accruedAfterRollover');
            }
            if (payout?.name === 'ProviderDebtPaid') count('debtPaid');
          }
          return receipt;
        } catch (error) {
          const message = (error as Error).message;
          if (!EXPECTED_REVERTS.some((name) => message.includes(name))) throw error;
          count('expectedRevert');
          return null;
        }
      };

      const openPoolSession = async () => {
        const hot = pick(hots);
        const provider = pick(providers);
        const { msg, signature } = await getProviderApproval(provider, hot, bids.get(provider.address)!);
        const receipt = await send(() =>
          delegateStaking.connect(hot).openSessionFromPool(hot, wei(10 + rand(50)), msg, signature),
        );
        if (receipt) {
          const id = await providerPayouts.getSessionId(hot, provider, bids.get(provider.address)!, nonce++);
          openPool.push({ id, provider, hot, direct: false });
          count('poolOpen');
        }
      };

      const openRegularSession = async () => {
        const provider = pick(providers);
        const direct = rand(2) === 0;
        const { msg, signature } = await getProviderApproval(provider, USER, bids.get(provider.address)!);
        const receipt = await send(() =>
          sessionRouter.connect(USER).openSession(USER, wei(10 + rand(50)), direct, msg, signature),
        );
        if (receipt) {
          const id = await providerPayouts.getSessionId(USER, provider, bids.get(provider.address)!, nonce++);
          openRegular.push({ id, provider, direct });
          count(direct ? 'directOpen' : 'regularOpen');
        }
      };

      const closeSession = async (sessions: OpenSession[], pool: boolean) => {
        if (sessions.length === 0) return;
        const index = rand(sessions.length);
        const session = sessions[index];
        const reporter = rand(4) === 0 ? KEEPER : session.provider;
        const { msg, signature } = await getReceipt(reporter, session.id, 1 + rand(50), 1 + rand(500));
        const receipt = pool
          ? await send(() => delegateStaking.connect(session.hot!).closePoolSession(msg, signature))
          : await send(() => sessionRouter.connect(USER).closeSession(msg, signature));
        expect(receipt).to.not.eq(null);
        for (const log of receipt!.logs) {
          const parsed = providerPayouts.interface.parseLog(log);
          if (session.direct && parsed?.name === 'ProviderRewardAccrued') {
            expect(parsed.args.owed).to.eq(0n);
          }
        }
        if (session.direct) closedDirect.push(session.id);
        count(pool ? 'poolClose' : session.direct ? 'directClose' : 'regularClose');
        sessions.splice(index, 1);
      };

      const settleSession = async () => {
        if (openPool.length === 0) return;
        const index = rand(openPool.length);
        const receipt = await send(() => delegateStaking.connect(KEEPER).settleExpiredSession(openPool[index].id));
        if (receipt) {
          openPool.splice(index, 1);
          count('settle');
        }
      };

      const fund = async () => {
        const hot = pick(hots);
        const funder = pick(funders);
        if (!(await delegateStakingFunding.isPoolFunderApproved(hot, funder))) {
          await delegateStakingFunding.connect(hot).approvePoolFunder(hot, funder);
        }
        const allowance = await delegateStakingFunding.getStakingAllowance(hot, funder);
        if (!allowance.active || rand(3) === 0) {
          const expiry = (await getCurrentBlockTime()) + BigInt(DAY * (1 + rand(5)));
          await send(() =>
            delegateStakingFunding
              .connect(funder)
              .grantStakingAllowance(hot, allowance.lifetimeFunded + wei(100 + rand(200)), expiry),
          );
        }
        await send(() => delegateStakingFunding.connect(funder).fundStakingAllowance(hot, wei(1 + rand(80))));
      };

      const withdraw = async () => {
        const hot = pick(hots);
        const funder = pick(funders);
        const { currentPrincipal } = await delegateStakingFunding.getStakingAllowance(hot, funder);
        if (currentPrincipal === 0n) return;
        const amount = rand(2) === 0 ? currentPrincipal : (currentPrincipal * BigInt(1 + rand(99))) / 100n;
        await send(() => delegateStakingFunding.connect(funder).requestStakingWithdrawal(hot, funder, amount));
      };

      const ops: Array<[number, () => Promise<void>]> = [
        [6, fund],
        [6, openPoolSession],
        [4, openRegularSession],
        [5, () => closeSession(openPool, true)],
        [3, () => closeSession(openRegular, false)],
        [2, settleSession],
        [4, withdraw],
        [
          3,
          async () => {
            const queued: Array<[SignerWithAddress, SignerWithAddress]> = [];
            for (const hot of hots) {
              for (const funder of funders) {
                if ((await delegateStakingFunding.getStakingWithdrawal(hot, funder)) !== 0n) queued.push([hot, funder]);
              }
            }
            if (queued.length === 0) return;
            const [hot, funder] = pick(queued);
            if (await send(() => delegateStakingFunding.connect(funder).cancelStakingWithdrawal(hot, funder))) {
              count('cancel');
            }
          },
        ],
        [
          1,
          async () => {
            const hot = pick(hots);
            const funder = pick(funders);
            const allowance = await delegateStakingFunding.getStakingAllowance(hot, funder);
            if (allowance.active) await delegateStakingFunding.connect(funder).revokeStakingAllowance(hot);
          },
        ],
        [
          1,
          async () => {
            const hot = pick(hots);
            const funder = pick(funders);
            if (rand(2) === 0) await delegateStakingFunding.connect(hot).revokePoolFunderApproval(hot, funder);
            else await delegateStakingFunding.connect(hot).approvePoolFunder(hot, funder);
          },
        ],
        [
          2,
          async () => {
            const hot = pick(hots);
            await send(() => delegateStakingFunding.connect(KEEPER).releasePoolLocks(hot, 1 + rand(32)));
            await send(() => delegateStakingFunding.connect(KEEPER).servePoolWithdrawals(hot, 1 + rand(32)));
          },
        ],
        [
          2,
          async () => {
            const sessions = [...openPool, ...openRegular];
            if (sessions.length !== 0) {
              await send(() => sessionRouter.connect(KEEPER).claimForProvider(pick(sessions).id));
            }
          },
        ],
        [1, () => send(() => sessionRouter.connect(USER).withdrawUserStakes(USER, 255)).then(() => undefined)],
        [
          2,
          async () => {
            const balance = await token.balanceOf(FUNDING);
            if (rand(2) === 0 && balance !== 0n) await token.connect(FUNDING).transfer(OWNER, balance);
            else await token.transfer(FUNDING, wei(rand(20)));
          },
        ],
        [
          5,
          async () => {
            const now = Number(await getCurrentBlockTime());
            await setTime(now + (rand(6) === 0 ? DAY + rand(DAY) : 60 + rand(3 * 3600)));
          },
        ],
      ];
      const totalWeight = ops.reduce((sum, [weight]) => sum + weight, 0);

      const checkInvariants = async (step: number) => {
        const context = `seed ${seed}, step ${step}`;
        let poolsTotal = 0n;

        for (const hot of hots) {
          const balances = await delegateStakingFunding.getPoolBalances(hot);
          let principal = 0n;
          let pending = 0n;
          let exposure = 0n;
          let lockedExposure = 0n;
          for (const funder of funders) {
            const allowance = await delegateStakingFunding.getStakingAllowance(hot, funder);
            principal += allowance.currentPrincipal;
            pending += allowance.pending;
            exposure += allowance.exposure;
            lockedExposure += allowance.lockedExposure;
            expect(allowance.lifetimeFunded - (paid.get(key(hot, funder)) ?? 0n), context).to.eq(
              allowance.currentPrincipal + allowance.pending,
            );
          }
          expect(principal + balances.pendingTotal, context).to.eq(
            balances.free + balances.inSession + balances.locked,
          );
          expect(pending, context).to.eq(balances.pendingTotal);
          expect(exposure, context).to.eq(balances.inSession);
          expect(lockedExposure, context).to.be.lessThanOrEqual(balances.locked);

          // A lock is always dated the next day and every lock-creating operation releases matured buckets
          // first, so a pool never holds more than one unreleased bucket.
          const [buckets, bucketCount] = await delegateStakingFunding.getPoolLockBuckets(hot, 0, 32);
          expect(bucketCount, context).to.be.lessThanOrEqual(1n);
          expect(
            buckets.reduce((sum, bucket) => sum + bucket.amount, 0n),
            context,
          ).to.eq(balances.locked);

          // funders keep their shares of the unreleased lock exactly until its date; the second right
          // before the date is skipped because a view may run at the next block time
          // (more than one bucket already fails above)
          const now = await getCurrentBlockTime();
          if (buckets.length === 1 && buckets[0].releaseAt > now + 1n) {
            expect(lockedExposure, context).to.eq(buckets[0].amount);
          } else if (buckets.length === 0 || buckets[0].releaseAt <= now) {
            expect(lockedExposure, context).to.eq(0n);
          }
          poolsTotal += balances.free + balances.inSession + balances.locked;
        }

        let regularTotal = 0n;
        for (const { id, direct } of openRegular) {
          const session = await sessionRouter.getSession(id);
          regularTotal += direct ? session.stake - session.providerWithdrawnAmount : session.stake;
        }
        // a closed direct-pay session keeps the provider's unpaid part (dispute hold, limiter) in the diamond
        for (const id of closedDirect) {
          const session = await sessionRouter.getSession(id);
          const end = session.closedAt < session.endsAt ? session.closedAt : session.endsAt;
          regularTotal += (end - session.openedAt) * price - session.providerWithdrawnAmount;
        }
        const [available, hold] = await providerPayouts.getUserStakesOnHold(USER, 0);
        expect(await token.balanceOf(diamond), context).to.eq(
          baseDiamondBalance + poolsTotal + regularTotal + available + hold,
        );

        const owed =
          (await providerPayouts.getProviderOwed(PROVIDER_1)) + (await providerPayouts.getProviderOwed(PROVIDER_2));
        const pendingClaims = await providerPayouts.getPendingProviderClaims();
        expect(pendingClaims, context).to.eq(owed);
        expect(
          (await providerPayouts.getComputeBalance(emissionAt)) +
            (await sessionRouter.getProvidersTotalClaimed()) +
            pendingClaims,
          context,
        ).to.eq(emissionTotal);

        for (let i = 0; i < hots.length; i++) {
          expect(await token.balanceOf(hots[i]), context).to.eq(hotBalances[i]);
        }

        const limited = await providerRegistry.getProvider(PROVIDER_2);
        if (limited.limitPeriodEarned === limited.stake) count('limiterHit');
      };

      for (let step = 0; step < steps; step++) {
        let roll = rand(totalWeight);
        for (const [weight, op] of ops) {
          if (roll < weight) {
            await op();
            break;
          }
          roll -= weight;
        }
        await checkInvariants(step);
      }

      for (let attempt = 0; openPool.length === 0 && attempt < 5; attempt++) {
        await fund();
        await openPoolSession();
      }

      // wind down: every session ends, the limiter period rolls over, funding is refilled, and anyone settles and claims
      const now = Number(await getCurrentBlockTime());
      await setTime(now + 366 * DAY);
      afterRollover = true;
      for (const session of [...openPool]) {
        expect(await send(() => delegateStaking.connect(KEEPER).settleExpiredSession(session.id))).to.not.eq(null);
        count('settle');
      }
      openPool.length = 0;
      while (openRegular.length !== 0) await closeSession(openRegular, false);
      await token.transfer(FUNDING, wei(1000));
      for (const id of [...closedDirect]) await send(() => sessionRouter.connect(KEEPER).claimForProvider(id));
      for (const provider of providers) {
        const [sessionIds] = await sessionRouter.getProviderSessions(provider, 0, 1000);
        for (const id of sessionIds) await send(() => sessionRouter.connect(KEEPER).claimForProvider(id));
      }
      await checkInvariants(steps);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(0);

      limiterHits += counts.get('limiterHit') ?? 0;
      const coverage = Object.fromEntries(counts);
      for (const name of ['poolOpen', 'poolClose', 'regularClose', 'directClose', 'settle', 'withdrawalPaid', 'debtAccrued', 'debtPaid',
        'cancel', 'lockReleased', 'expectedRevert', 'accruedAfterRollover']) {
        expect(counts.get(name) ?? 0, `seed ${seed}: no ${name} in ${JSON.stringify(coverage)}`).to.be.greaterThan(0);
      }
    }

    for (const seed of [1, 7, 42]) {
      it(`should keep pool, funder, debt, emission, and diamond balance invariants (seed ${seed})`, async function () {
        this.timeout(600_000);
        await run(seed, 150);
      });
    }

    it('should have hit the reward limiter in the randomized runs', () => {
      expect(limiterHits).to.be.greaterThan(0);
    });
  });

  describe('gas at hard caps', () => {
    async function createFunders(hot: SignerWithAddress, count: number, amount: bigint, expiry: number) {
      const wallets = Array.from({ length: count }, () => Wallet.createRandom().connect(ethers.provider));
      for (const wallet of wallets) {
        await ethers.provider.send('hardhat_setBalance', [wallet.address, '0x56BC75E2D63100000']);
        await token.transfer(wallet.address, amount);
        await token.connect(wallet).approve(diamond, amount);
        await delegateStakingFunding.connect(hot).approvePoolFunder(hot, wallet.address);
        await delegateStakingFunding.connect(wallet).grantStakingAllowance(hot, amount, expiry);
        await delegateStakingFunding.connect(wallet).fundStakingAllowance(hot, amount);
      }
      return wallets;
    }

    async function openFromPool(hot: SignerWithAddress, amount: bigint, nonce: number) {
      const bidId = bids.get(PROVIDER_1.address)!;
      const { msg, signature } = await getProviderApproval(PROVIDER_1, hot, bidId);
      const gas = await gasOf(delegateStaking.connect(hot).openSessionFromPool(hot, amount, msg, signature));
      return { gas, sessionId: await providerPayouts.getSessionId(hot, PROVIDER_1, bidId, nonce) };
    }

    async function closeFromPool(hot: SignerWithAddress, sessionId: string) {
      const { msg, signature } = await getReceipt(PROVIDER_1, sessionId, 10, 10);
      return gasOf(delegateStaking.connect(hot).closePoolSession(msg, signature));
    }

    it('should draw, close, and settle 129 draws within the gas budget', async function () {
      this.timeout(300_000);
      await setTime(openTime);
      await delegateStakingFunding.setMaxActiveFunders(128);
      await delegateStakingFunding.setMinPrincipal(1);
      await delegateStakingFunding.setSettlementGrace(1);
      await createFunders(HOT_A, 128, wei(1), openTime + 5 * DAY);
      await token.transfer(HOT_A, wei(1));
      await token.connect(HOT_A).approve(diamond, wei(1));
      await delegateStakingFunding.connect(HOT_A).grantStakingAllowance(HOT_A, wei(1), openTime + 5 * DAY);
      await delegateStakingFunding.connect(HOT_A).fundStakingAllowance(HOT_A, wei(1));

      const first = await openFromPool(HOT_A, wei(129), 0);
      expect((await delegateStakingFunding.getPoolSession(first.sessionId)).draws.length).to.eq(129);
      // close on a later day so no lock stays behind and the second session can draw everything again
      await setTime(Number((await sessionRouter.getSession(first.sessionId)).endsAt) + DAY);
      const closeGas = await closeFromPool(HOT_A, first.sessionId);

      const second = await openFromPool(HOT_A, wei(129), 1);
      expect((await delegateStakingFunding.getPoolSession(second.sessionId)).draws.length).to.eq(129);
      const session = await sessionRouter.getSession(second.sessionId);
      await setTime(Number(session.endsAt) + 1);
      const settleGas = await gasOf(delegateStaking.connect(KEEPER).settleExpiredSession(second.sessionId));

      console.log(`      gas: draw 129 = ${first.gas}, close 129 = ${closeGas}, settle 129 = ${settleGas}`);
      for (const gas of [first.gas, closeGas, settleGas]) expect(gas).to.be.lessThan(GAS_BUDGET);
    });

    it('should close 129 draws with a day lock split across every funder within the gas budget', async function () {
      this.timeout(300_000);
      await setTime(openTime);
      await delegateStakingFunding.setMaxActiveFunders(128);
      await delegateStakingFunding.setMinPrincipal(1);
      const wallets = await createFunders(HOT_A, 128, wei(1), openTime + 5 * DAY);
      await token.transfer(HOT_A, wei(1));
      await token.connect(HOT_A).approve(diamond, wei(1));
      await delegateStakingFunding.connect(HOT_A).grantStakingAllowance(HOT_A, wei(1), openTime + 5 * DAY);
      await delegateStakingFunding.connect(HOT_A).fundStakingAllowance(HOT_A, wei(1));

      const { sessionId } = await openFromPool(HOT_A, wei(129), 0);
      const session = await sessionRouter.getSession(sessionId);
      // an early close on the same day leaves a lock on every draw
      await setTime(Number(session.openedAt) + 3600);
      const closeGas = await closeFromPool(HOT_A, sessionId);

      const [[bucket]] = await delegateStakingFunding.getPoolLockBuckets(HOT_A, 0, 1);
      let lockedExposure = 0n;
      for (const funder of [...wallets.map((wallet) => wallet.address), HOT_A.address]) {
        const share = (await delegateStakingFunding.getStakingAllowance(HOT_A, funder)).lockedExposure;
        expect(share).to.be.greaterThan(0);
        lockedExposure += share;
      }
      expect(bucket.amount).to.be.greaterThan(0);
      expect(lockedExposure).to.eq(bucket.amount);

      console.log(`      gas: close 129 with lock = ${closeGas}`);
      expect(closeGas).to.be.lessThan(GAS_BUDGET);
    });

    it('should purge a full list of expired grants when admitting a new funder within the gas budget', async function () {
      this.timeout(300_000);
      await setTime(openTime);
      await delegateStakingFunding.setMaxActiveFunders(128);
      await delegateStakingFunding.setMinPrincipal(1);
      await createFunders(HOT_A, 128, wei(1), openTime + DAY);
      await setTime(openTime + 2 * DAY);

      await delegateStakingFunding.connect(HOT_A).approvePoolFunder(HOT_A, FUNDER_1);
      await delegateStakingFunding
        .connect(FUNDER_1)
        .grantStakingAllowance(HOT_A, wei(100), openTime + 5 * DAY);
      const fundGas = await gasOf(delegateStakingFunding.connect(FUNDER_1).fundStakingAllowance(HOT_A, wei(100)));
      const [funders] = await delegateStakingFunding.getPoolFunders(HOT_A, 0, 128);
      expect(funders.filter((funder) => funder.inActiveList).length).to.eq(0);
      expect((await delegateStakingFunding.getPoolFunders(HOT_A, 128, 1))[0][0].inActiveList).to.eq(true);

      const open = await openFromPool(HOT_A, wei(50), 0);
      console.log(`      gas: admit with 128-entry purge = ${fundGas}, draw after purge = ${open.gas}`);
      expect(fundGas).to.be.lessThan(GAS_BUDGET);
      expect(open.gas).to.be.lessThan(GAS_BUDGET);
    });

    // A lock is always created for the next day and every pool operation releases matured buckets first,
    // so the matured backlog stays tiny in practice; the bucket loop is bounded by maxAutoReleaseDays anyway.
    it('should serve the maximum withdrawal batch within the gas budget', async function () {
      this.timeout(300_000);
      await setTime(openTime);
      await delegateStakingFunding.setMinPrincipal(1);
      const wallets = await createFunders(HOT_A, 33, wei(5), openTime + 60 * DAY);

      // put everything in a session and queue 33 withdrawals; the close serves one, the measured call 32
      const { sessionId } = await openFromPool(HOT_A, wei(165), 0);
      for (const wallet of wallets) {
        await delegateStakingFunding.connect(wallet).requestStakingWithdrawal(HOT_A, wallet.address, wei(5));
      }
      expect((await delegateStakingFunding.getPoolWithdrawalQueue(HOT_A, ethers.ZeroAddress, 1))[2]).to.eq(33);
      await delegateStakingFunding.setMaxAutoService(1);
      const session = await sessionRouter.getSession(sessionId);
      await setTime(Number(session.endsAt) + DAY);
      await closeFromPool(HOT_A, sessionId);
      expect((await delegateStakingFunding.getPoolWithdrawalQueue(HOT_A, ethers.ZeroAddress, 1))[2]).to.eq(32);
      const serveGas = await gasOf(delegateStakingFunding.connect(KEEPER).servePoolWithdrawals(HOT_A, 32));
      expect((await delegateStakingFunding.getPoolBalances(HOT_A)).pendingTotal).to.eq(0);

      console.log(`      gas: serve 32 withdrawals = ${serveGas}`);
      expect(serveGas).to.be.lessThan(GAS_BUDGET);
    });

    it('should release a day lock and serve the maximum withdrawal batch within the gas budget', async function () {
      this.timeout(300_000);
      await setTime(openTime);
      await delegateStakingFunding.setMinPrincipal(1);
      const wallets = await createFunders(HOT_A, 33, wei(5), openTime + 60 * DAY);

      // an early close on the same day leaves a lock; the close serves one queued withdrawal, the release 32
      const { sessionId } = await openFromPool(HOT_A, wei(165), 0);
      for (const wallet of wallets) {
        await delegateStakingFunding.connect(wallet).requestStakingWithdrawal(HOT_A, wallet.address, wei(5));
      }
      await delegateStakingFunding.setMaxAutoService(1);
      const session = await sessionRouter.getSession(sessionId);
      await setTime(Number(session.openedAt) + 3600);
      await closeFromPool(HOT_A, sessionId);
      const [[bucket]] = await delegateStakingFunding.getPoolLockBuckets(HOT_A, 0, 1);
      expect(bucket.amount).to.be.greaterThan(0);
      expect((await delegateStakingFunding.getPoolWithdrawalQueue(HOT_A, ethers.ZeroAddress, 1))[2]).to.eq(32);

      await delegateStakingFunding.setMaxAutoService(32);
      await setTime(Number(bucket.releaseAt));
      const releaseGas = await gasOf(delegateStakingFunding.connect(KEEPER).releasePoolLocks(HOT_A, 32));
      const balances = await delegateStakingFunding.getPoolBalances(HOT_A);
      expect(balances.locked).to.eq(0);
      expect(balances.pendingTotal).to.eq(0);

      console.log(`      gas: release lock + serve 32 withdrawals = ${releaseGas}`);
      expect(releaseGas).to.be.lessThan(GAS_BUDGET);
    });
  });

  describe('interleaved regular and pool sessions', () => {
    it('should close each interleaved session only through its own path', async () => {
      await setTime(openTime);
      const bidId = bids.get(PROVIDER_1.address)!;
      await delegateStakingFunding.connect(HOT_A).approvePoolFunder(HOT_A, FUNDER_1);
      await delegateStakingFunding.connect(FUNDER_1).grantStakingAllowance(HOT_A, wei(100), openTime + 10 * DAY);
      await delegateStakingFunding.connect(FUNDER_1).fundStakingAllowance(HOT_A, wei(100));
      await token.transfer(HOT_A, wei(100));
      await token.connect(HOT_A).approve(diamond, wei(100));

      const kinds = ['regular', 'pool', 'regular', 'pool'];
      const ids: string[] = [];
      for (const [nonce, kind] of kinds.entries()) {
        const { msg, signature } = await getProviderApproval(PROVIDER_1, HOT_A, bidId);
        if (kind === 'pool') await delegateStaking.connect(HOT_A).openSessionFromPool(HOT_A, wei(50), msg, signature);
        else await sessionRouter.connect(HOT_A).openSession(HOT_A, wei(50), false, msg, signature);
        ids.push(await providerPayouts.getSessionId(HOT_A, PROVIDER_1, bidId, nonce));
      }
      expect(new Set(ids).size).to.eq(kinds.length);
      expect(await delegateStakingFunding.getPoolBalances(HOT_A)).to.deep.eq([0n, wei(100), 0n, 0n]);

      // close on a later day than every endsAt, so no day-lock stays behind
      await setTime(openTime + 2 * DAY);
      const hotBalance = await token.balanceOf(HOT_A);
      for (const [index, kind] of kinds.entries()) {
        const { msg, signature } = await getReceipt(PROVIDER_1, ids[index], 10, 10);
        if (kind === 'pool') {
          await expect(sessionRouter.connect(HOT_A).closeSession(msg, signature)).to.be.revertedWithCustomError(
            sessionRouter,
            'SessionPoolSessionCannotBeClosed',
          );
          await delegateStaking.connect(HOT_A).closePoolSession(msg, signature);
        } else {
          await expect(delegateStaking.connect(HOT_A).closePoolSession(msg, signature)).to.be.revertedWithCustomError(
            delegateStaking,
            'DelegateStakingPoolSessionNotFound',
          );
          await expect(delegateStaking.connect(KEEPER).settleExpiredSession(ids[index])).to.be.revertedWithCustomError(
            delegateStaking,
            'DelegateStakingPoolSessionNotFound',
          );
          await sessionRouter.connect(HOT_A).closeSession(msg, signature);
        }
        expect((await sessionRouter.getSession(ids[index])).closedAt).to.not.eq(0n);
      }

      // regular stakes went back to the hot wallet, pool stakes went back to the pool only
      expect((await token.balanceOf(HOT_A)) - hotBalance).to.eq(wei(100));
      expect(await delegateStakingFunding.getPoolBalances(HOT_A)).to.deep.eq([wei(100), 0n, 0n, 0n]);
      expect((await delegateStakingFunding.getStakingAllowance(HOT_A, FUNDER_1)).exposure).to.eq(0n);
    });
  });

  describe('bus-factor drill', () => {
    it('should recover funder capital and pay the provider without hot, provider, or owner keys', async () => {
      await setTime(openTime);
      const bidId = bids.get(PROVIDER_1.address)!;
      const target = Wallet.createRandom().address;
      await providerPayouts.connect(PROVIDER_1).setPayoutTarget(PROVIDER_1, target);
      await delegateStakingFunding.connect(HOT_A).approvePoolFunder(HOT_A, FUNDER_1);
      await delegateStakingFunding.connect(FUNDER_1).grantStakingAllowance(HOT_A, wei(100), openTime + 10 * DAY);
      await delegateStakingFunding.connect(FUNDER_1).fundStakingAllowance(HOT_A, wei(100));

      const { msg, signature } = await getProviderApproval(PROVIDER_1, HOT_A, bidId);
      await delegateStaking.connect(HOT_A).openSessionFromPool(HOT_A, wei(100), msg, signature);
      const sessionId = await providerPayouts.getSessionId(HOT_A, PROVIDER_1, bidId, 0);
      const session = await sessionRouter.getSession(sessionId);
      const reward = (session.endsAt - session.openedAt) * price;

      // from here on: hot, provider, and owner keys are lost; only the funder, the funding wallet, and a keeper act
      await delegateStakingFunding.connect(FUNDER_1).requestStakingWithdrawal(HOT_A, FUNDER_1, wei(100));
      await token.connect(FUNDING).transfer(USER, await token.balanceOf(FUNDING));
      await setTime(Number(session.endsAt) + DAY);
      const funderBefore = await token.balanceOf(FUNDER_1);
      await delegateStaking.connect(KEEPER).settleExpiredSession(sessionId);
      expect(await providerPayouts.getProviderOwed(PROVIDER_1)).to.eq(reward);

      // the settle already serviced the queued withdrawal; a keeper call is a harmless no-op
      await delegateStakingFunding.connect(KEEPER).servePoolWithdrawals(HOT_A, 32);
      expect((await token.balanceOf(FUNDER_1)) - funderBefore).to.eq(wei(100));
      expect(await delegateStakingFunding.getPoolBalances(HOT_A)).to.deep.eq([0n, 0n, 0n, 0n]);

      await token.connect(USER).transfer(FUNDING, reward);
      await expect(sessionRouter.connect(KEEPER).claimForProvider(sessionId)).to.changeTokenBalances(
        token,
        [FUNDING, target, PROVIDER_1, KEEPER],
        [-reward, reward, 0n, 0n],
      );
      expect(await providerPayouts.getProviderOwed(PROVIDER_1)).to.eq(0);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(0);
    });
  });
});
