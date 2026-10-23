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
import { Wallet, ZeroAddress, ZeroHash } from 'ethers';
import { expect } from 'chai';
import { ethers, network } from 'hardhat';

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
import { getCurrentBlockTime, setNextTime, setTime } from '@/utils/block-helper';
import { getProviderApproval, getReceipt } from '@/utils/provider-helper';
import { DAY } from '@/utils/time';

describe('DelegateStaking', () => {
  const reverter = new Reverter();
  const bidPricePerSecond = wei(0.0001);
  const openTime = payoutStart + 10 * DAY;
  const baseModelId = getHex(Buffer.from('delegate-staking-model'));

  let OWNER: SignerWithAddress;
  let HOT: SignerWithAddress;
  let DELEGATE: SignerWithAddress;
  let STRANGER: SignerWithAddress;
  let FUNDER: SignerWithAddress;
  let FUNDER_2: SignerWithAddress;
  let FUNDER_3: SignerWithAddress;
  let FUNDING: SignerWithAddress;
  let PROVIDER: SignerWithAddress;

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
  let modelId: string;
  let bidId: string;

  before(async () => {
    [OWNER, HOT, DELEGATE, STRANGER, FUNDER, FUNDER_2, FUNDER_3, FUNDING, PROVIDER] = await ethers.getSigners();

    diamond = await deployLumerinDiamond();
    token = await deployMORToken();
    delegateRegistry = await deployDelegateRegistry();
    providerRegistry = await deployFacetProviderRegistry(diamond);
    modelRegistry = await deployFacetModelRegistry(diamond);
    sessionRouter = await deployFacetSessionRouter(diamond, FUNDING);
    marketplace = await deployFacetMarketplace(diamond, token, wei(0.0001), wei(900));
    await deployFacetDelegation(diamond, delegateRegistry);
    providerPayouts = await deployFacetProviderPayouts(diamond);
    delegateStaking = await deployFacetDelegateStaking(diamond);
    delegateStakingFunding = await deployFacetDelegateStakingFunding(diamond);

    for (const account of [HOT, FUNDER, FUNDER_2, FUNDER_3, FUNDING, PROVIDER]) {
      await token.transfer(account, wei(10000));
      await token.connect(account).approve(diamond, wei(10000));
    }

    await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(0.2), 'provider');
    await modelRegistry
      .connect(PROVIDER)
      .modelRegister(PROVIDER, baseModelId, getHex(Buffer.from('ipfs://delegate-staking')), 0, wei(100), 'model', [
        'delegate-staking',
      ]);
    modelId = await modelRegistry.getModelId(PROVIDER, baseModelId);
    await marketplace.connect(PROVIDER).postModelBid(PROVIDER, modelId, bidPricePerSecond);
    bidId = await marketplace.getBidId(PROVIDER, modelId, 0);

    await reverter.snapshot();
  });

  afterEach(reverter.revert);

  async function fundPool(
    hot: SignerWithAddress,
    funder: SignerWithAddress,
    amount: bigint,
    expiry = BigInt(openTime + 10 * DAY),
  ) {
    if (hot.address !== funder.address) {
      await delegateStakingFunding.connect(hot).approvePoolFunder(hot, funder);
    }
    await delegateStakingFunding.connect(funder).grantStakingAllowance(hot, amount, expiry);
    await delegateStakingFunding.connect(funder).fundStakingAllowance(hot, amount);
  }

  async function openPoolSession(
    amount = wei(50),
    hot = HOT,
    caller = hot,
    nonce = 0,
    approvalUser = hot,
  ) {
    const { msg, signature } = await getProviderApproval(PROVIDER, approvalUser, bidId);
    await delegateStaking.connect(caller).openSessionFromPool(hot, amount, msg, signature);
    return {
      approval: msg,
      sessionId: await providerPayouts.getSessionId(hot, PROVIDER, bidId, nonce),
    };
  }

  async function closePoolSession(sessionId: string, caller = HOT, reporter = PROVIDER, tps = 7, ttftMs = 11) {
    const { msg, signature } = await getReceipt(reporter, sessionId, tps, ttftMs);
    await delegateStaking.connect(caller).closePoolSession(msg, signature);
    return msg;
  }

  describe('#openSessionFromPool', () => {
    it('should reject a pool with neither an external grant nor self-escrow', async () => {
      await setTime(openTime);
      const { msg, signature } = await getProviderApproval(PROVIDER, HOT, bidId);

      await expect(delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), msg, signature)).to.be
        .revertedWithCustomError(delegateStaking, 'DelegateStakingInsufficientLiquidity');
    });

    it('should open from a funded grant without moving MOR through the hot wallet', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(100));
      const hotBalance = await token.balanceOf(HOT);
      const { msg, signature } = await getProviderApproval(PROVIDER, HOT, bidId);
      const sessionId = await providerPayouts.getSessionId(HOT, PROVIDER, bidId, 0);

      await expect(delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), msg, signature))
        .to.emit(delegateStaking, 'DelegateStakingPoolDrawn')
        .withArgs(sessionId, HOT, FUNDER, wei(50))
        .and.to.emit(delegateStaking, 'SessionOpened')
        .withArgs(HOT, sessionId, PROVIDER);

      const session = await sessionRouter.getSession(sessionId);
      expect(session.user).to.eq(HOT);
      expect(session.bidId).to.eq(bidId);
      expect(session.stake).to.eq(wei(50));
      expect(session.endsAt).to.eq(await providerPayouts.getSessionEnd(wei(50), bidPricePerSecond, session.openedAt));
      expect(session.isDirectPaymentFromUser).to.eq(false);
      expect(session.isActive).to.eq(true);
      expect(await token.balanceOf(HOT)).to.eq(hotBalance);
      expect(await delegateStakingFunding.getPoolBalances(HOT)).to.deep.eq([wei(50), wei(50), 0n, 0n]);
      expect(await delegateStakingFunding.getPoolSession(sessionId)).to.deep.eq([
        HOT.address,
        [[FUNDER.address, wei(50)]],
        true,
        false,
        false,
      ]);
      expect(await sessionRouter.getIsProviderApprovalUsed(msg)).to.eq(true);
      expect(await sessionRouter.getUserSessions(HOT, 0, 10)).to.deep.eq([[sessionId], 1n]);
      expect(await sessionRouter.getProviderSessions(PROVIDER, 0, 10)).to.deep.eq([[sessionId], 1n]);
      expect(await sessionRouter.getModelSessions(modelId, 0, 10)).to.deep.eq([[sessionId], 1n]);
    });

    it('should use the shared nonce across regular and pool sessions', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(100));
      const a1 = await getProviderApproval(PROVIDER, HOT, bidId);
      await sessionRouter.connect(HOT).openSession(HOT, wei(50), false, a1.msg, a1.signature);
      const a2 = await getProviderApproval(PROVIDER, HOT, bidId);
      await delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), a2.msg, a2.signature);
      const a3 = await getProviderApproval(PROVIDER, HOT, bidId);
      await sessionRouter.connect(HOT).openSession(HOT, wei(50), false, a3.msg, a3.signature);

      const ids = await Promise.all(
        [0, 1, 2].map((nonce) => providerPayouts.getSessionId(HOT, PROVIDER, bidId, nonce)),
      );
      expect(new Set(ids).size).to.eq(3);
      expect((await delegateStakingFunding.getPoolSession(ids[0])).exists).to.eq(false);
      expect((await delegateStakingFunding.getPoolSession(ids[1])).exists).to.eq(true);
      expect((await delegateStakingFunding.getPoolSession(ids[2])).exists).to.eq(false);
      expect(await sessionRouter.getUserSessions(HOT, 0, 10)).to.deep.eq([ids, 3n]);
    });

    it('should allow the hot wallet and its session delegate but reject a stranger', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(100));
      await delegateRegistry
        .connect(HOT)
        .delegateContract(DELEGATE, providerRegistry, await providerRegistry.DELEGATION_RULES_SESSION(), true);

      const delegatedApproval = await getProviderApproval(PROVIDER, HOT, bidId);
      await expect(
        delegateStaking.connect(DELEGATE).openSessionFromPool(HOT, wei(50), delegatedApproval.msg, delegatedApproval.signature),
      ).not.to.be.reverted;

      const strangerApproval = await getProviderApproval(PROVIDER, HOT, bidId);
      await expect(
        delegateStaking.connect(STRANGER).openSessionFromPool(HOT, wei(50), strangerApproval.msg, strangerApproval.signature),
      ).to.be.revertedWithCustomError(delegateStaking, 'InsufficientRightsForOperation');
    });

    it('should reject an approval for another user', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(50));
      const { msg, signature } = await getProviderApproval(PROVIDER, STRANGER, bidId);

      await expect(delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), msg, signature)).to.be
        .revertedWithCustomError(delegateStaking, 'SessionApprovedForAnotherUser');
    });

    it('should not let a delegate draw one user pool for another user', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(50));
      await delegateRegistry
        .connect(HOT)
        .delegateContract(DELEGATE, providerRegistry, await providerRegistry.DELEGATION_RULES_SESSION(), true);
      await delegateRegistry
        .connect(STRANGER)
        .delegateContract(DELEGATE, providerRegistry, await providerRegistry.DELEGATION_RULES_SESSION(), true);
      await fundPool(STRANGER, FUNDER_2, wei(50));

      const forStranger = await getProviderApproval(PROVIDER, STRANGER, bidId);
      await expect(
        delegateStaking.connect(DELEGATE).openSessionFromPool(HOT, wei(50), forStranger.msg, forStranger.signature),
      ).to.be.revertedWithCustomError(delegateStaking, 'SessionApprovedForAnotherUser');

      const forHot = await getProviderApproval(PROVIDER, HOT, bidId);
      await expect(
        delegateStaking.connect(DELEGATE).openSessionFromPool(STRANGER, wei(50), forHot.msg, forHot.signature),
      ).to.be.revertedWithCustomError(delegateStaking, 'SessionApprovedForAnotherUser');

      expect(await delegateStakingFunding.getAvailablePoolStake(HOT)).to.eq(wei(50));
      expect(await delegateStakingFunding.getAvailablePoolStake(STRANGER)).to.eq(wei(50));
    });

    it('should apply regular-session approval, bid, and duration validation', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(300));

      const wrongChain = await getProviderApproval(PROVIDER, HOT, bidId, 1n);
      await expect(
        delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), wrongChain.msg, wrongChain.signature),
      ).to.be.revertedWithCustomError(delegateStaking, 'SessionApprovedForAnotherChainId');

      const mismatch = await getProviderApproval(STRANGER, HOT, bidId);
      await expect(
        delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), mismatch.msg, mismatch.signature),
      ).to.be.revertedWithCustomError(delegateStaking, 'SessionProviderSignatureMismatch');

      const missingBid = ZeroHash;
      const missing = await getProviderApproval(PROVIDER, HOT, missingBid);
      await expect(
        delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), missing.msg, missing.signature),
      ).to.be.revertedWithCustomError(delegateStaking, 'SessionBidNotFound');

      const duplicated = await getProviderApproval(PROVIDER, HOT, bidId);
      await delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), duplicated.msg, duplicated.signature);
      await expect(
        delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), duplicated.msg, duplicated.signature),
      ).to.be.revertedWithCustomError(delegateStaking, 'SessionDuplicateApproval');
    });

    it('should reject an expired approval, a deleted bid, and a too-short session', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(150), BigInt(openTime + 10 * DAY));
      const expired = await getProviderApproval(PROVIDER, HOT, bidId);
      await setTime(Number(await getCurrentBlockTime()) + 601);
      await expect(
        delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), expired.msg, expired.signature),
      ).to.be.revertedWithCustomError(delegateStaking, 'SessionApproveExpired');

      const deleted = await getProviderApproval(PROVIDER, HOT, bidId);
      await marketplace.connect(PROVIDER).deleteModelBid(bidId);
      await expect(
        delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), deleted.msg, deleted.signature),
      ).to.be.revertedWithCustomError(delegateStaking, 'SessionBidNotFound');

      await reverter.revert();
      await setTime(payoutStart);
      await fundPool(HOT, FUNDER, wei(50), BigInt(payoutStart + DAY));
      const tooShort = await getProviderApproval(PROVIDER, HOT, bidId);
      await expect(
        delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), tooShort.msg, tooShort.signature),
      ).to.be.revertedWithCustomError(delegateStaking, 'SessionTooShort');
    });

    it('should distinguish liquidity reserved for withdrawals from grant capacity', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(100));
      await openPoolSession(wei(80));
      await delegateStakingFunding.connect(FUNDER).requestStakingWithdrawal(HOT, FUNDER, wei(100));
      const liquidityApproval = await getProviderApproval(PROVIDER, HOT, bidId);
      await expect(
        delegateStaking
          .connect(HOT)
          .openSessionFromPool(HOT, wei(20), liquidityApproval.msg, liquidityApproval.signature),
      ).to.be.revertedWithCustomError(delegateStaking, 'DelegateStakingInsufficientLiquidity');
      expect((await delegateStakingFunding.getPoolBalances(HOT)).pendingTotal).to.eq(wei(80));

      await reverter.revert();
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(50), BigInt(openTime + 10));
      await setTime(openTime + 11);
      const capacityApproval = await getProviderApproval(PROVIDER, HOT, bidId);
      await expect(
        delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), capacityApproval.msg, capacityApproval.signature),
      ).to.be.revertedWithCustomError(delegateStaking, 'DelegateStakingInsufficientCapacity');
      expect((await delegateStakingFunding.getPoolBalances(HOT)).free).to.eq(wei(50));
    });

    it('should draw FIFO with partial draws, skip revoked entries, and draw self-escrow last', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(40));
      await fundPool(HOT, FUNDER_2, wei(40));
      await fundPool(HOT, FUNDER_3, wei(40));
      await fundPool(HOT, HOT, wei(40));
      await delegateStakingFunding.connect(FUNDER_2).revokeStakingAllowance(HOT);

      const { sessionId } = await openPoolSession(wei(100));
      const poolSession = await delegateStakingFunding.getPoolSession(sessionId);
      expect(poolSession.draws).to.deep.eq([
        [FUNDER.address, wei(40)],
        [FUNDER_3.address, wei(40)],
        [HOT.address, wei(20)],
      ]);
    });

    it('should skip a revoked funder approval immediately', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(50));
      await fundPool(HOT, FUNDER_2, wei(50));
      await delegateStakingFunding.connect(HOT).revokePoolFunderApproval(HOT, FUNDER);

      const { sessionId } = await openPoolSession(wei(50));
      expect((await delegateStakingFunding.getPoolSession(sessionId)).draws).to.deep.eq([
        [FUNDER_2.address, wei(50)],
      ]);
    });

    it('should skip and unlink an expired funder during draw traversal', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(50), BigInt(openTime + 10));
      await fundPool(HOT, FUNDER_2, wei(50), BigInt(openTime + DAY));
      await setTime(openTime + 11);

      const { sessionId } = await openPoolSession(wei(50));
      expect((await delegateStakingFunding.getPoolSession(sessionId)).draws).to.deep.eq([
        [FUNDER_2.address, wei(50)],
      ]);
      const [funders] = await delegateStakingFunding.getPoolFunders(HOT, 0, 10);
      expect(funders[0].inActiveList).to.eq(false);
      expect(funders[1].inActiveList).to.eq(true);
    });
  });

  describe('#closePoolSession', () => {
    it('should reject unknown and regular sessions', async () => {
      await setTime(openTime);
      const unknownReceipt = await getReceipt(PROVIDER, ZeroHash, 0, 0);
      await expect(delegateStaking.closePoolSession(unknownReceipt.msg, unknownReceipt.signature)).to.be
        .revertedWithCustomError(delegateStaking, 'DelegateStakingPoolSessionNotFound');

      const approval = await getProviderApproval(PROVIDER, HOT, bidId);
      await sessionRouter.connect(HOT).openSession(HOT, wei(50), false, approval.msg, approval.signature);
      const regularId = await providerPayouts.getSessionId(HOT, PROVIDER, bidId, 0);
      const regularReceipt = await getReceipt(PROVIDER, regularId, 0, 0);
      await expect(delegateStaking.closePoolSession(regularReceipt.msg, regularReceipt.signature)).to.be
        .revertedWithCustomError(delegateStaking, 'DelegateStakingPoolSessionNotFound');
    });

    it('should allow the hot wallet or session delegate, reject a stranger, and reject repeat close', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(100));
      const { sessionId } = await openPoolSession();
      const receipt = await getReceipt(PROVIDER, sessionId, 0, 0);
      await expect(delegateStaking.connect(STRANGER).closePoolSession(receipt.msg, receipt.signature)).to.be
        .revertedWithCustomError(delegateStaking, 'InsufficientRightsForOperation');

      await delegateRegistry
        .connect(HOT)
        .delegateContract(DELEGATE, providerRegistry, await providerRegistry.DELEGATION_RULES_SESSION(), true);
      await delegateStaking.connect(DELEGATE).closePoolSession(receipt.msg, receipt.signature);
      await expect(delegateStaking.connect(HOT).closePoolSession(receipt.msg, receipt.signature)).to.be
        .revertedWithCustomError(delegateStaking, 'SessionAlreadyClosed');
    });

    it('should validate receipt chain and expiry', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(50));
      const { sessionId } = await openPoolSession();
      const wrongChain = await getReceipt(PROVIDER, sessionId, 0, 0, 1n);
      await expect(delegateStaking.connect(HOT).closePoolSession(wrongChain.msg, wrongChain.signature)).to.be
        .revertedWithCustomError(delegateStaking, 'SessionReceiptForAnotherChainId');

      const expired = await getReceipt(PROVIDER, sessionId, 0, 0);
      await setTime(Number(await getCurrentBlockTime()) + 601);
      await expect(delegateStaking.connect(HOT).closePoolSession(expired.msg, expired.signature)).to.be
        .revertedWithCustomError(delegateStaking, 'SessionReceiptExpired');
    });

    it('should return unused stake, create the exact day lock, release attribution, and keep hot balance unchanged', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(30));
      await fundPool(HOT, FUNDER_2, wei(30));
      const hotBalance = await token.balanceOf(HOT);
      const { sessionId } = await openPoolSession(wei(50));
      const session = await sessionRouter.getSession(sessionId);
      await setTime(Number(session.openedAt) + 200);
      const closeAt = BigInt(Number(session.openedAt) + 201);
      const startOfDay = await providerPayouts.startOfTheDay(closeAt);
      const initialLock = (closeAt - session.openedAt) * bidPricePerSecond;
      const expectedLock = await providerPayouts.stipendToStake(initialLock, startOfDay);
      const receipt = await getReceipt(PROVIDER, sessionId, 7, 11);

      await expect(delegateStaking.connect(HOT).closePoolSession(receipt.msg, receipt.signature))
        .to.emit(delegateStaking, 'DelegateStakingPoolLockCreated')
        .withArgs(HOT, startOfDay + BigInt(DAY), expectedLock)
        .and.to.emit(delegateStaking, 'DelegateStakingPoolReturned')
        .withArgs(sessionId, HOT, FUNDER, wei(30))
        .and.to.emit(delegateStaking, 'DelegateStakingPoolReturned')
        .withArgs(sessionId, HOT, FUNDER_2, wei(20))
        .and.to.emit(delegateStaking, 'SessionClosed')
        .withArgs(HOT, sessionId, PROVIDER)
        .and.to.emit(delegateStaking, 'SessionRating')
        .withArgs(sessionId, modelId, PROVIDER, 0, 7000, 11);

      expect(await delegateStakingFunding.getPoolBalances(HOT)).to.deep.eq([
        wei(60) - expectedLock,
        0n,
        expectedLock,
        0n,
      ]);
      expect(await delegateStakingFunding.getPoolLockBuckets(HOT, 0, 10)).to.deep.eq([
        [[startOfDay + BigInt(DAY), expectedLock]],
        1n,
      ]);
      const poolSession = await delegateStakingFunding.getPoolSession(sessionId);
      expect(poolSession.draws).to.deep.eq([]);
      expect(poolSession.closed).to.eq(true);
      expect(poolSession.settled).to.eq(false);
      expect(await token.balanceOf(HOT)).to.eq(hotBalance);
    });

    for (const [name, offset] of [
      ['an early same-day close', 200],
      ['a late close after endsAt on the same day', 0],
    ] as const) {
      it(`should lock and pay exactly like a regular session for ${name}`, async () => {
        // open around noon so that endsAt and the close stay on the same day
        const t0 = openTime + 3600;
        await setTime(openTime);
        await fundPool(HOT, FUNDER, wei(100));

        const run = async (isPool: boolean) => {
          const snapshot = await network.provider.send('evm_snapshot', []);
          await setTime(t0);
          const { msg, signature } = await getProviderApproval(PROVIDER, HOT, bidId);
          if (isPool) {
            await delegateStaking.connect(HOT).openSessionFromPool(HOT, wei(50), msg, signature);
          } else {
            await sessionRouter.connect(HOT).openSession(HOT, wei(50), false, msg, signature);
          }
          const sessionId = await providerPayouts.getSessionId(HOT, PROVIDER, bidId, 0);
          const opened = await sessionRouter.getSession(sessionId);
          await setTime(offset === 0 ? Number(opened.endsAt) + 100 : t0 + offset);
          const providerBefore = await token.balanceOf(PROVIDER);
          const { msg: receipt, signature: receiptSig } = await getReceipt(PROVIDER, sessionId, 0, 0);
          if (isPool) {
            await delegateStaking.connect(HOT).closePoolSession(receipt, receiptSig);
          } else {
            await sessionRouter.connect(HOT).closeSession(receipt, receiptSig);
          }

          const result = {
            session: (await sessionRouter.getSession(sessionId)).toArray().slice(2),
            providerPaid: (await token.balanceOf(PROVIDER)) - providerBefore,
            lock: isPool
              ? (await delegateStakingFunding.getPoolLockBuckets(HOT, 0, 1))[0].map((b) => [b.releaseAt, b.amount])
              : [[BigInt(DAY) + (await providerPayouts.startOfTheDay(BigInt(t0))), (await providerPayouts.getUserStakesOnHold(HOT, 10))[1]]],
          };
          await network.provider.send('evm_revert', [snapshot]);
          return result;
        };

        const pool = await run(true);
        const regular = await run(false);

        expect(pool.lock[0][1]).to.be.greaterThan(0);
        expect(pool).to.deep.eq(regular);
      });
    }

    it('should create no lock when closed after the settlement day', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(50));
      const hotBalance = await token.balanceOf(HOT);
      const { sessionId } = await openPoolSession();
      const session = await sessionRouter.getSession(sessionId);
      await setTime(Number(session.endsAt) + DAY);
      await closePoolSession(sessionId);

      expect(await delegateStakingFunding.getPoolBalances(HOT)).to.deep.eq([wei(50), 0n, 0n, 0n]);
      expect(await token.balanceOf(HOT)).to.eq(hotBalance);
    });

    it('should pay immediately, apply dispute hold, or accrue debt exactly like regular close', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(150));
      const first = await openPoolSession(wei(50), HOT, HOT, 0);
      const firstData = await sessionRouter.getSession(first.sessionId);
      await setTime(Number(firstData.openedAt) + 200);
      const providerBefore = await token.balanceOf(PROVIDER);
      await closePoolSession(first.sessionId);
      const firstClosed = await sessionRouter.getSession(first.sessionId);
      const firstReward = (firstClosed.closedAt - firstClosed.openedAt) * bidPricePerSecond;
      expect((await token.balanceOf(PROVIDER)) - providerBefore).to.eq(firstReward);

      const second = await openPoolSession(wei(50), HOT, HOT, 1);
      const secondData = await sessionRouter.getSession(second.sessionId);
      await setTime(Number(secondData.openedAt) + 100);
      const disputedBefore = await token.balanceOf(PROVIDER);
      await closePoolSession(second.sessionId, HOT, STRANGER);
      expect((await token.balanceOf(PROVIDER)) - disputedBefore).to.eq(0);
      expect((await sessionRouter.getSession(second.sessionId)).closeoutType).to.eq(1);

      const third = await openPoolSession(wei(50), HOT, HOT, 2);
      const thirdData = await sessionRouter.getSession(third.sessionId);
      await token.connect(FUNDING).transfer(OWNER, await token.balanceOf(FUNDING));
      await setTime(Number(thirdData.openedAt) + 100);
      await closePoolSession(third.sessionId);
      const thirdClosed = await sessionRouter.getSession(third.sessionId);
      const thirdReward = (thirdClosed.closedAt - thirdClosed.openedAt) * bidPricePerSecond;
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(thirdReward);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(thirdReward);
    });

    it('should lazily release matured buckets and service the withdrawal queue while closing', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(150));
      const first = await openPoolSession(wei(50), HOT, HOT, 0);
      const second = await openPoolSession(wei(50), HOT, HOT, 1);
      const firstData = await sessionRouter.getSession(first.sessionId);
      await setTime(Number(firstData.openedAt) + 100);
      await closePoolSession(first.sessionId);
      const bucket = (await delegateStakingFunding.getPoolLockBuckets(HOT, 0, 1))[0][0];
      await delegateStakingFunding.connect(FUNDER).requestStakingWithdrawal(HOT, FUNDER, wei(150));
      expect((await delegateStakingFunding.getPoolBalances(HOT)).pendingTotal).to.be.greaterThan(0);
      const funderBalance = await token.balanceOf(FUNDER);

      await setTime(Number(bucket.releaseAt));
      const receipt = await getReceipt(PROVIDER, second.sessionId, 0, 0);
      await expect(delegateStaking.connect(HOT).closePoolSession(receipt.msg, receipt.signature))
        .to.emit(delegateStakingFunding, 'DelegateStakingPoolLockReleased')
        .withArgs(HOT, bucket.releaseAt, bucket.amount);
      expect((await delegateStakingFunding.getPoolBalances(HOT)).pendingTotal).to.eq(0);
      expect((await token.balanceOf(FUNDER)) - funderBalance).to.be.greaterThan(0);
      expect(await token.balanceOf(HOT)).to.eq(wei(10000));
    });
  });

  describe('#closeSession', () => {
    it('should reject pool sessions for the hot wallet, its delegate, and a stranger', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(50));
      const { sessionId } = await openPoolSession();
      await delegateRegistry
        .connect(HOT)
        .delegateContract(DELEGATE, providerRegistry, await providerRegistry.DELEGATION_RULES_SESSION(), true);
      const receipt = await getReceipt(PROVIDER, sessionId, 0, 0);

      for (const caller of [HOT, DELEGATE, STRANGER]) {
        await expect(sessionRouter.connect(caller).closeSession(receipt.msg, receipt.signature)).to.be
          .revertedWithCustomError(sessionRouter, 'SessionPoolSessionCannotBeClosed');
      }
    });

    it('should leave regular sessions outside the pool registry and close them only through SessionRouter', async () => {
      await setTime(openTime);
      const approval = await getProviderApproval(PROVIDER, HOT, bidId);
      await sessionRouter.connect(HOT).openSession(HOT, wei(50), false, approval.msg, approval.signature);
      const sessionId = await providerPayouts.getSessionId(HOT, PROVIDER, bidId, 0);
      expect(await delegateStakingFunding.getPoolSession(sessionId)).to.deep.eq([
        ZeroAddress,
        [],
        false,
        false,
        false,
      ]);
      const receipt = await getReceipt(PROVIDER, sessionId, 0, 0);
      await expect(sessionRouter.connect(HOT).closeSession(receipt.msg, receipt.signature)).not.to.be.reverted;
    });

    it('should let claimForProvider accrue rewards for a pool session', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(50));
      const { sessionId } = await openPoolSession();
      const session = await sessionRouter.getSession(sessionId);
      await setTime(Number(session.endsAt));
      const providerBalance = await token.balanceOf(PROVIDER);

      await sessionRouter.connect(STRANGER).claimForProvider(sessionId);
      expect((await token.balanceOf(PROVIDER)) - providerBalance).to.eq(
        (session.endsAt - session.openedAt) * bidPricePerSecond,
      );
      expect((await delegateStakingFunding.getPoolSession(sessionId)).closed).to.eq(false);
    });
  });

  describe('#settleExpiredSession', () => {
    it('should reject unknown and regular sessions', async () => {
      await expect(delegateStaking.settleExpiredSession(ZeroHash)).to.be.revertedWithCustomError(
        delegateStaking,
        'DelegateStakingPoolSessionNotFound',
      );
      await setTime(openTime);
      const approval = await getProviderApproval(PROVIDER, HOT, bidId);
      await sessionRouter.connect(HOT).openSession(HOT, wei(50), false, approval.msg, approval.signature);
      const regularId = await providerPayouts.getSessionId(HOT, PROVIDER, bidId, 0);
      await expect(delegateStaking.settleExpiredSession(regularId)).to.be.revertedWithCustomError(
        delegateStaking,
        'DelegateStakingPoolSessionNotFound',
      );
    });

    it('should enforce the exact endsAt plus grace boundary and allow anyone to settle once', async () => {
      await setTime(openTime);
      await delegateStakingFunding.setSettlementGrace(1);
      await fundPool(HOT, FUNDER, wei(50));
      const { sessionId } = await openPoolSession();
      const session = await sessionRouter.getSession(sessionId);
      await setNextTime(Number(session.endsAt));
      await expect(delegateStaking.connect(STRANGER).settleExpiredSession(sessionId)).to.be.revertedWithCustomError(
        delegateStaking,
        'DelegateStakingSettlementNotReady',
      );

      await setNextTime(Number(session.endsAt) + 1);
      await expect(delegateStaking.connect(STRANGER).settleExpiredSession(sessionId))
        .to.emit(delegateStaking, 'SessionClosed')
        .withArgs(HOT, sessionId, PROVIDER)
        .and.to.emit(delegateStaking, 'DelegateStakingPoolSessionSettled')
        .withArgs(sessionId, HOT, PROVIDER, STRANGER);
      await expect(delegateStaking.settleExpiredSession(sessionId)).to.be.revertedWithCustomError(
        delegateStaking,
        'SessionAlreadyClosed',
      );
      await expect(closePoolSession(sessionId)).to.be.rejectedWith('SessionAlreadyClosed');
    });

    it('should reject settlement after a receipt close', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(50));
      const { sessionId } = await openPoolSession();
      await closePoolSession(sessionId);
      await expect(delegateStaking.settleExpiredSession(sessionId)).to.be.revertedWithCustomError(
        delegateStaking,
        'SessionAlreadyClosed',
      );
    });

    it('should settle a dead node into provider debt without stats, ratings, or hot-wallet transfers', async () => {
      await setTime(openTime);
      await delegateStakingFunding.setSettlementGrace(1);
      await fundPool(HOT, FUNDER, wei(50));
      const hotBalance = await token.balanceOf(HOT);
      const providerBalance = await token.balanceOf(PROVIDER);
      const { sessionId } = await openPoolSession();
      const session = await sessionRouter.getSession(sessionId);
      const providerStatsBefore = await sessionRouter.getProviderModelStats(modelId, PROVIDER);
      const modelStatsBefore = await sessionRouter.getModelStats(modelId);
      const computeBefore = await providerPayouts.getComputeBalance(session.endsAt);
      const totalSupplyBefore = await providerPayouts.totalMORSupply(session.endsAt);
      await token.connect(FUNDING).transfer(OWNER, await token.balanceOf(FUNDING));
      await setTime(Number(session.endsAt) + 1);
      const reward = (session.endsAt - session.openedAt) * bidPricePerSecond;

      await expect(delegateStaking.connect(STRANGER).settleExpiredSession(sessionId))
        .to.emit(delegateStaking, 'ProviderRewardAccrued')
        .withArgs(sessionId, PROVIDER, reward, 0, reward)
        .and.not.to.emit(delegateStaking, 'SessionRating');

      const closed = await sessionRouter.getSession(sessionId);
      expect(closed.closeoutType).to.eq(0);
      expect(closed.providerWithdrawnAmount).to.eq(reward);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(reward);
      expect(await token.balanceOf(PROVIDER)).to.eq(providerBalance);
      expect(await token.balanceOf(HOT)).to.eq(hotBalance);
      expect(await sessionRouter.getProviderModelStats(modelId, PROVIDER)).to.deep.eq(providerStatsBefore);
      expect(await sessionRouter.getModelStats(modelId)).to.deep.eq(modelStatsBefore);
      expect(await providerPayouts.getComputeBalance(session.endsAt)).to.eq(computeBefore - reward);
      expect(await providerPayouts.totalMORSupply(session.endsAt)).to.eq(totalSupplyBefore + reward);
      const poolSession = await delegateStakingFunding.getPoolSession(sessionId);
      expect(poolSession.closed).to.eq(true);
      expect(poolSession.settled).to.eq(true);
      expect((await delegateStakingFunding.getPoolBalances(HOT)).locked).to.be.greaterThan(0);

      await token.transfer(FUNDING, reward);
      await sessionRouter.connect(STRANGER).claimForProvider(sessionId);
      expect((await token.balanceOf(PROVIDER)) - providerBalance).to.eq(reward);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(0);
    });
  });

  describe('#settleExpiredSession edge cases', () => {
    it('should use the default one-day grace boundary', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(50));
      const { sessionId } = await openPoolSession();
      const session = await sessionRouter.getSession(sessionId);

      await setNextTime(Number(session.endsAt) + DAY - 1);
      await expect(delegateStaking.settleExpiredSession(sessionId)).to.be.revertedWithCustomError(
        delegateStaking,
        'DelegateStakingSettlementNotReady',
      );
      await setNextTime(Number(session.endsAt) + DAY);
      await expect(delegateStaking.settleExpiredSession(sessionId)).to.emit(
        delegateStaking,
        'DelegateStakingPoolSessionSettled',
      );
      // default grace: settle happens on a later day, so no lock and everything is back in the pool
      expect(await delegateStakingFunding.getPoolBalances(HOT)).to.deep.eq([wei(50), 0n, 0n, 0n]);
    });

    it('should move the whole reward into debt even when funding could pay', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(50));
      const { sessionId } = await openPoolSession();
      const session = await sessionRouter.getSession(sessionId);
      const reward = (session.endsAt - session.openedAt) * bidPricePerSecond;
      await setTime(Number(session.endsAt) + DAY);

      await expect(delegateStaking.connect(STRANGER).settleExpiredSession(sessionId)).to.changeTokenBalances(
        token,
        [FUNDING, PROVIDER, HOT, STRANGER],
        [0n, 0n, 0n, 0n],
      );
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward);
      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(0);

      await expect(sessionRouter.connect(STRANGER).claimForProvider(sessionId)).to.changeTokenBalances(
        token,
        [FUNDING, PROVIDER],
        [-reward, reward],
      );
    });

    it('should accrue only the remainder after a claim made while the session was open', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(50));
      const { sessionId } = await openPoolSession();
      const session = await sessionRouter.getSession(sessionId);
      const reward = (session.endsAt - session.openedAt) * bidPricePerSecond;

      await setTime(Number(session.endsAt) + 10);
      await sessionRouter.connect(STRANGER).claimForProvider(sessionId);
      expect((await sessionRouter.getSession(sessionId)).providerWithdrawnAmount).to.eq(reward);

      await setTime(Number(session.endsAt) + DAY);
      await expect(delegateStaking.settleExpiredSession(sessionId)).to.not.emit(
        delegateStaking,
        'ProviderRewardAccrued',
      );
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(0);
      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(reward);
    });

    it('should cap settle debt by the limiter and leave the rest claimable after the period resets', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(10000));
      const { sessionId } = await openPoolSession(wei(10000));
      const session = await sessionRouter.getSession(sessionId);
      const reward = (session.endsAt - session.openedAt) * bidPricePerSecond;
      expect(reward).to.be.greaterThan(wei(0.2));

      await setTime(Number(session.endsAt) + DAY);
      await delegateStaking.settleExpiredSession(sessionId);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(wei(0.2));

      const limitEnd = (await providerRegistry.getProvider(PROVIDER)).limitPeriodEnd;
      await setTime(Number(limitEnd) + 1);
      await sessionRouter.connect(STRANGER).claimForProvider(sessionId);
      expect((await sessionRouter.getSession(sessionId)).providerWithdrawnAmount).to.eq(wei(0.4));
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(0);
      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(wei(0.4));
    });

    it('should settle with an exact lock when grace ends on the same day as the session', async () => {
      await setTime(openTime);
      await delegateStakingFunding.setSettlementGrace(1);
      await fundPool(HOT, FUNDER, wei(50));
      const { sessionId } = await openPoolSession();
      const session = await sessionRouter.getSession(sessionId);
      const dayStart = await providerPayouts.startOfTheDay(session.endsAt);
      const expectedLock = await providerPayouts.stipendToStake(
        (session.endsAt - (session.openedAt > dayStart ? session.openedAt : dayStart)) * bidPricePerSecond,
        dayStart,
      );

      await setTime(Number(session.endsAt) + 1);
      await expect(delegateStaking.settleExpiredSession(sessionId))
        .to.emit(delegateStaking, 'DelegateStakingPoolLockCreated')
        .withArgs(HOT, dayStart + BigInt(DAY), expectedLock);
      expect(await delegateStakingFunding.getPoolBalances(HOT)).to.deep.eq([
        wei(50) - expectedLock,
        0n,
        expectedLock,
        0n,
      ]);
    });

    it('should close and settle already open pool sessions when the emission is exhausted', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(100));
      const first = await openPoolSession(wei(50), HOT, HOT, 0);
      const second = await openPoolSession(wei(50), HOT, HOT, 1);
      const day = await providerPayouts.startOfTheDay(BigInt(openTime));
      await sessionRouter.setPoolConfig(3, { payoutStart: day, decreaseInterval: DAY, initialReward: 1, rewardDecrease: 0 });
      await token.connect(FUNDING).transfer(OWNER, await token.balanceOf(FUNDING));

      const firstData = await sessionRouter.getSession(first.sessionId);
      await setTime(Number(firstData.openedAt) + 100);
      await closePoolSession(first.sessionId);
      expect(await providerPayouts.getComputeBalance(BigInt(openTime))).to.eq(0);

      const secondData = await sessionRouter.getSession(second.sessionId);
      await setTime(Number(secondData.endsAt) + DAY);
      await delegateStaking.settleExpiredSession(second.sessionId);

      expect(await delegateStakingFunding.getPoolBalances(HOT)).to.deep.eq([wei(100), 0n, 0n, 0n]);
    });
  });

  describe('payout target with pool sessions', () => {
    it('should pay pool-session rewards and settle debt to the current payout target only', async () => {
      const [TARGET, NEW_TARGET] = [DELEGATE, OWNER];
      await setTime(openTime);
      await providerPayouts.connect(PROVIDER).setPayoutTarget(PROVIDER, TARGET);
      await fundPool(HOT, FUNDER, wei(100));

      // immediate payment at closePoolSession goes to the target
      const first = await openPoolSession(wei(50), HOT, HOT, 0);
      const firstData = await sessionRouter.getSession(first.sessionId);
      await setTime(Number(firstData.openedAt) + 200);
      const reward = 201n * bidPricePerSecond;
      const { msg, signature } = await getReceipt(PROVIDER, first.sessionId, 0, 0);
      await expect(delegateStaking.connect(HOT).closePoolSession(msg, signature)).to.changeTokenBalances(
        token,
        [TARGET, PROVIDER, HOT, FUNDING],
        [reward, 0n, 0n, -reward],
      );

      // dead node with empty funding: debt, then the target changes before the claim
      const second = await openPoolSession(wei(50), HOT, HOT, 1);
      const secondData = await sessionRouter.getSession(second.sessionId);
      const debt = (secondData.endsAt - secondData.openedAt) * bidPricePerSecond;
      const fundingBalance = await token.balanceOf(FUNDING);
      await token.connect(FUNDING).transfer(STRANGER, fundingBalance);
      await setTime(Number(secondData.endsAt) + DAY);
      await delegateStaking.connect(STRANGER).settleExpiredSession(second.sessionId);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(debt);

      await providerPayouts.connect(PROVIDER).setPayoutTarget(PROVIDER, NEW_TARGET);
      await token.connect(STRANGER).transfer(FUNDING, fundingBalance);
      await expect(sessionRouter.connect(STRANGER).claimForProvider(second.sessionId))
        .to.emit(providerPayouts, 'ProviderDebtPaid')
        .withArgs(PROVIDER, NEW_TARGET, debt);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(0);
    });
  });

  describe('exposure attribution', () => {
    it('should track per-funder exposure while open and return it on close even after revoke', async () => {
      await setTime(openTime);
      await fundPool(HOT, FUNDER, wei(30));
      await fundPool(HOT, FUNDER_2, wei(30));
      const { sessionId } = await openPoolSession(wei(50));

      expect((await delegateStakingFunding.getStakingAllowance(HOT, FUNDER)).exposure).to.eq(wei(30));
      expect((await delegateStakingFunding.getStakingAllowance(HOT, FUNDER_2)).exposure).to.eq(wei(20));

      await delegateStakingFunding.connect(FUNDER).revokeStakingAllowance(HOT);
      await delegateStakingFunding.connect(HOT).revokePoolFunderApproval(HOT, FUNDER_2);

      const session = await sessionRouter.getSession(sessionId);
      await setTime(Number(session.endsAt) + DAY);
      await closePoolSession(sessionId);

      expect((await delegateStakingFunding.getStakingAllowance(HOT, FUNDER)).exposure).to.eq(0);
      expect((await delegateStakingFunding.getStakingAllowance(HOT, FUNDER_2)).exposure).to.eq(0);

      await expect(
        delegateStakingFunding.connect(FUNDER).requestStakingWithdrawal(HOT, FUNDER, wei(30)),
      ).to.changeTokenBalances(token, [FUNDER, HOT], [wei(30), 0n]);
      await expect(
        delegateStakingFunding.connect(FUNDER_2).requestStakingWithdrawal(HOT, FUNDER_2, wei(30)),
      ).to.changeTokenBalances(token, [FUNDER_2, HOT], [wei(30), 0n]);
      expect(await delegateStakingFunding.getPoolBalances(HOT)).to.deep.eq([0n, 0n, 0n, 0n]);
    });
  });

  describe('pool session draw cap', () => {
    it('should draw from 128 funders plus self-escrow and return every draw', async function () {
      this.timeout(180_000);
      await setTime(openTime);
      await delegateStakingFunding.setMaxActiveFunders(128);
      await delegateStakingFunding.setMinPrincipal(1);

      const wallets = Array.from({ length: 128 }, () => Wallet.createRandom().connect(ethers.provider));
      for (const wallet of wallets) {
        await ethers.provider.send('hardhat_setBalance', [wallet.address, '0x56BC75E2D63100000']);
        await token.transfer(wallet.address, wei(1));
        await token.connect(wallet).approve(diamond, wei(1));
        await delegateStakingFunding.connect(HOT).approvePoolFunder(HOT, wallet.address);
        await delegateStakingFunding.connect(wallet).grantStakingAllowance(HOT, wei(1), openTime + DAY);
        await delegateStakingFunding.connect(wallet).fundStakingAllowance(HOT, wei(1));
      }
      await fundPool(HOT, HOT, wei(1));
      const hotBalance = await token.balanceOf(HOT);
      const { sessionId } = await openPoolSession(wei(129));
      expect((await delegateStakingFunding.getPoolSession(sessionId)).draws.length).to.eq(129);

      const session = await sessionRouter.getSession(sessionId);
      await setTime(Number(session.endsAt) + DAY);
      await closePoolSession(sessionId);
      expect((await delegateStakingFunding.getPoolSession(sessionId)).draws).to.deep.eq([]);
      expect(await delegateStakingFunding.getPoolBalances(HOT)).to.deep.eq([wei(129), 0n, 0n, 0n]);
      expect(await token.balanceOf(HOT)).to.eq(hotBalance);
    });
  });
});
