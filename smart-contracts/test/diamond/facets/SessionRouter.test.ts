import {
  DelegateRegistry,
  LumerinDiamond,
  Marketplace,
  ModelRegistry,
  MorpheusToken,
  ProviderOwnerMock,
  ProviderPayouts,
  ProviderRegistry,
  SessionRouter,
} from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
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
  deployFacetProviderRegistry,
  deployFacetProviderPayouts,
  deployFacetSessionRouter,
  deployLumerinDiamond,
  deployMORToken,
} from '@/test/helpers/deployers';
import { payoutStart } from '@/test/helpers/pool-helper';
import { Reverter } from '@/test/helpers/reverter';
import { getCurrentBlockTime, setTime } from '@/utils/block-helper';
import { getProviderApproval, getReceipt } from '@/utils/provider-helper';
import { DAY } from '@/utils/time';

describe('SessionRouter', () => {
  const reverter = new Reverter();

  let OWNER: SignerWithAddress;
  let SECOND: SignerWithAddress;
  let FUNDING: SignerWithAddress;
  let PROVIDER: SignerWithAddress;

  let diamond: LumerinDiamond;
  let marketplace: Marketplace;
  let modelRegistry: ModelRegistry;
  let providerRegistry: ProviderRegistry;
  let providerPayouts: ProviderPayouts;
  let sessionRouter: SessionRouter;

  let token: MorpheusToken;
  let delegateRegistry: DelegateRegistry;

  let bidId = '';
  const baseModelId = getHex(Buffer.from('1'));
  let modelId = getHex(Buffer.from(''));
  const bidPricePerSecond = wei(0.0001);

  before(async () => {
    [OWNER, SECOND, FUNDING, PROVIDER] = await ethers.getSigners();

    [diamond, token, delegateRegistry] = await Promise.all([
      deployLumerinDiamond(),
      deployMORToken(),
      deployDelegateRegistry(),
    ]);

    [providerRegistry, modelRegistry, sessionRouter, marketplace, , providerPayouts] = await Promise.all([
      deployFacetProviderRegistry(diamond),
      deployFacetModelRegistry(diamond),
      deployFacetSessionRouter(diamond, FUNDING),
      deployFacetMarketplace(diamond, token, wei(0.0001), wei(900)),
      deployFacetDelegation(diamond, delegateRegistry),
      deployFacetProviderPayouts(diamond),
      deployFacetDelegateStaking(diamond),
      deployFacetDelegateStakingFunding(diamond),
    ]);

    await token.transfer(SECOND, wei(10000));
    await token.transfer(PROVIDER, wei(10000));
    await token.transfer(FUNDING, wei(10000));
    await token.connect(PROVIDER).approve(providerRegistry, wei(10000));
    await token.connect(PROVIDER).approve(modelRegistry, wei(10000));
    await token.connect(PROVIDER).approve(marketplace, wei(10000));
    await token.connect(SECOND).approve(sessionRouter, wei(10000));
    await token.connect(FUNDING).approve(sessionRouter, wei(10000));

    const ipfsCID = getHex(Buffer.from('ipfs://ipfsaddress'));
    await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(0.2), 'test');
    await modelRegistry.connect(PROVIDER).modelRegister(PROVIDER, baseModelId, ipfsCID, 0, wei(100), 'name', ['tag_1']);

    modelId = await modelRegistry.getModelId(PROVIDER, baseModelId);

    await marketplace.connect(PROVIDER).postModelBid(PROVIDER, modelId, bidPricePerSecond);
    bidId = await marketplace.getBidId(PROVIDER, modelId, 0);

    await reverter.snapshot();
  });

  afterEach(reverter.revert);

  describe('#__SessionRouter_init', () => {
    it('should set correct data after creation', async () => {
      expect(await sessionRouter.getFundingAccount()).to.eq(FUNDING);
      expect((await sessionRouter.getPools()).length).to.eq(5);
    });
    it('should revert if try to call init function twice', async () => {
      await expect(sessionRouter.__SessionRouter_init(FUNDING, DAY, [])).to.be.rejectedWith(
        'Initializable: contract is already initialized',
      );
    });
  });

  describe('#getTodaysBudget', () => {
    it('should return not zero amount', async () => {
      expect(await providerPayouts.getTodaysBudget(payoutStart)).to.eq(0);
      expect(await providerPayouts.getTodaysBudget(payoutStart + 10 * DAY)).to.greaterThan(0);
    });
  });

  describe('#setPoolConfig', () => {
    it('should reset pool', async () => {
      await sessionRouter.setPoolConfig(0, {
        payoutStart: 0,
        decreaseInterval: DAY,
        initialReward: wei(1000),
        rewardDecrease: wei(10),
      });

      const pool = await sessionRouter.getPool(0);
      expect(pool.payoutStart).to.eq(0);
      expect(pool.decreaseInterval).to.eq(DAY);
      expect(pool.initialReward).to.eq(wei(1000));
      expect(pool.rewardDecrease).to.eq(wei(10));
    });
    it('should throw error when the pool index is invalid', async () => {
      await expect(
        sessionRouter.setPoolConfig(100, {
          payoutStart: 0,
          decreaseInterval: DAY,
          initialReward: wei(1000),
          rewardDecrease: wei(10),
        }),
      ).to.be.revertedWithCustomError(sessionRouter, 'SessionPoolIndexOutOfBounds');
    });
    it('should throw error when the caller is invalid', async () => {
      await expect(
        sessionRouter.connect(SECOND).setPoolConfig(0, {
          payoutStart: 0,
          decreaseInterval: DAY,
          initialReward: wei(1000),
          rewardDecrease: wei(10),
        }),
      ).to.be.revertedWithCustomError(sessionRouter, 'OwnableUnauthorizedAccount');
    });
  });

  describe('#setMaxSessionDuration', () => {
    it('should set max session duration', async () => {
      await sessionRouter.setMaxSessionDuration(7 * DAY);
      expect(await sessionRouter.getMaxSessionDuration()).to.eq(7 * DAY);

      await sessionRouter.setMaxSessionDuration(8 * DAY);
      expect(await sessionRouter.getMaxSessionDuration()).to.eq(8 * DAY);
    });
    it('should throw error when max session duration too low', async () => {
      await expect(sessionRouter.setMaxSessionDuration(1)).to.be.revertedWithCustomError(
        sessionRouter,
        'SessionMaxDurationTooShort',
      );
    });
    it('should throw error when the caller is invalid', async () => {
      await expect(sessionRouter.connect(SECOND).setMaxSessionDuration(1)).to.be.revertedWithCustomError(
        sessionRouter,
        'OwnableUnauthorizedAccount',
      );
    });
  });

  describe('#openSession', () => {
    let tokenBalBefore = 0n;
    let secondBalBefore = 0n;

    beforeEach(async () => {
      tokenBalBefore = await token.balanceOf(sessionRouter);
      secondBalBefore = await token.balanceOf(SECOND);
    });
    it('should open a session', async () => {
      await setTime(payoutStart + 10 * DAY);
      const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature);

      const sessionId = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 0);
      const data = await sessionRouter.getSession(sessionId);
      expect(data.user).to.eq(SECOND);
      expect(data.bidId).to.eq(bidId);
      expect(data.stake).to.eq(wei(50));
      expect(data.closeoutReceipt).to.eq('0x');
      expect(data.closeoutType).to.eq(0);
      expect(data.providerWithdrawnAmount).to.eq(0);
      expect(data.openedAt).to.eq(payoutStart + 10 * DAY + 1);
      expect(data.endsAt).to.greaterThan(data.openedAt);
      expect(data.closedAt).to.eq(0);
      expect(data.isActive).to.eq(true);
      expect(data.isDirectPaymentFromUser).to.eq(false);

      const tokenBalAfter = await token.balanceOf(sessionRouter);
      expect(tokenBalAfter - tokenBalBefore).to.eq(wei(50));
      const secondBalAfter = await token.balanceOf(SECOND);
      expect(secondBalBefore - secondBalAfter).to.eq(wei(50));

      expect(await sessionRouter.getIsProviderApprovalUsed(msg)).to.eq(true);
      expect(await sessionRouter.getUserSessions(SECOND, 0, 10)).to.deep.eq([[sessionId], 1n]);
      expect(await sessionRouter.getProviderSessions(PROVIDER, 0, 10)).to.deep.eq([[sessionId], 1n]);
      expect(await sessionRouter.getModelSessions(modelId, 0, 10)).to.deep.eq([[sessionId], 1n]);
    });
    it('should open a session from the delegatee address', async () => {
      await delegateRegistry
        .connect(SECOND)
        .delegateContract(OWNER, providerRegistry, await providerRegistry.DELEGATION_RULES_SESSION(), true);

      await setTime(payoutStart + 10 * DAY);
      const { msg, signature } = await getProviderApproval(PROVIDER, OWNER, bidId);
      await sessionRouter.connect(OWNER).openSession(SECOND, wei(50), false, msg, signature);

      const sessionId = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 0);
      const data = await sessionRouter.getSession(sessionId);
      expect(data.user).to.eq(SECOND);
      expect(data.bidId).to.eq(bidId);
      expect(data.stake).to.eq(wei(50));
      expect(data.closeoutReceipt).to.eq('0x');
      expect(data.closeoutType).to.eq(0);
      expect(data.providerWithdrawnAmount).to.eq(0);
      expect(data.openedAt).to.eq(payoutStart + 10 * DAY + 1);
      expect(data.endsAt).to.greaterThan(data.openedAt);
      expect(data.closedAt).to.eq(0);
      expect(data.isActive).to.eq(true);
      expect(data.isDirectPaymentFromUser).to.eq(false);

      const tokenBalAfter = await token.balanceOf(sessionRouter);
      expect(tokenBalAfter - tokenBalBefore).to.eq(wei(50));
      const secondBalAfter = await token.balanceOf(SECOND);
      expect(secondBalBefore - secondBalAfter).to.eq(wei(50));

      expect(await sessionRouter.getIsProviderApprovalUsed(msg)).to.eq(true);
      expect(await sessionRouter.getUserSessions(SECOND, 0, 10)).to.deep.eq([[sessionId], 1n]);
      expect(await sessionRouter.getProviderSessions(PROVIDER, 0, 10)).to.deep.eq([[sessionId], 1n]);
      expect(await sessionRouter.getModelSessions(modelId, 0, 10)).to.deep.eq([[sessionId], 1n]);
    });
    it('should open a session from the delegatee address with approval from delegatee to delegatee', async () => {
      await delegateRegistry
        .connect(SECOND)
        .delegateContract(OWNER, providerRegistry, await providerRegistry.DELEGATION_RULES_SESSION(), true);
      await delegateRegistry
        .connect(PROVIDER)
        .delegateContract(OWNER, providerRegistry, await providerRegistry.DELEGATION_RULES_PROVIDER(), true);

      await setTime(payoutStart + 10 * DAY);
      const { msg, signature } = await getProviderApproval(OWNER, OWNER, bidId);
      await sessionRouter.connect(OWNER).openSession(SECOND, wei(50), false, msg, signature);

      const sessionId = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 0);
      const data = await sessionRouter.getSession(sessionId);
      expect(data.user).to.eq(SECOND);
      expect(data.bidId).to.eq(bidId);
      expect(data.stake).to.eq(wei(50));
      expect(data.closeoutReceipt).to.eq('0x');
      expect(data.closeoutType).to.eq(0);
      expect(data.providerWithdrawnAmount).to.eq(0);
      expect(data.openedAt).to.eq(payoutStart + 10 * DAY + 1);
      expect(data.endsAt).to.greaterThan(data.openedAt);
      expect(data.closedAt).to.eq(0);
      expect(data.isActive).to.eq(true);
      expect(data.isDirectPaymentFromUser).to.eq(false);

      const tokenBalAfter = await token.balanceOf(sessionRouter);
      expect(tokenBalAfter - tokenBalBefore).to.eq(wei(50));
      const secondBalAfter = await token.balanceOf(SECOND);
      expect(secondBalBefore - secondBalAfter).to.eq(wei(50));

      expect(await sessionRouter.getIsProviderApprovalUsed(msg)).to.eq(true);
      expect(await sessionRouter.getUserSessions(SECOND, 0, 10)).to.deep.eq([[sessionId], 1n]);
      expect(await sessionRouter.getProviderSessions(PROVIDER, 0, 10)).to.deep.eq([[sessionId], 1n]);
      expect(await sessionRouter.getModelSessions(modelId, 0, 10)).to.deep.eq([[sessionId], 1n]);
    });
    it('should open two different sessions wit the same input params', async () => {
      await setTime(payoutStart + 10 * DAY);
      const { msg: msg1, signature: signature1 } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await setTime(payoutStart + 10 * DAY + 1);
      const { msg: msg2, signature: signature2 } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg1, signature1);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg2, signature2);

      const sessionId1 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 0);
      const sessionId2 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 1);

      const tokenBalAfter = await token.balanceOf(sessionRouter);
      expect(tokenBalAfter - tokenBalBefore).to.eq(wei(100));
      const secondBalAfter = await token.balanceOf(SECOND);
      expect(secondBalBefore - secondBalAfter).to.eq(wei(100));

      expect(await sessionRouter.getIsProviderApprovalUsed(msg1)).to.eq(true);
      expect(await sessionRouter.getIsProviderApprovalUsed(msg2)).to.eq(true);
      expect(await sessionRouter.getUserSessions(SECOND, 0, 10)).to.deep.eq([[sessionId1, sessionId2], 2n]);
      expect(await sessionRouter.getProviderSessions(PROVIDER, 0, 10)).to.deep.eq([[sessionId1, sessionId2], 2n]);
      expect(await sessionRouter.getModelSessions(modelId, 0, 10)).to.deep.eq([[sessionId1, sessionId2], 2n]);

      expect(await sessionRouter.getTotalSessions(PROVIDER)).to.eq(2);
    });
    it('should open a session with max duration', async () => {
      await setTime(payoutStart + 10 * DAY);
      const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(10000), false, msg, signature);

      const sessionId = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 0);
      const data = await sessionRouter.getSession(sessionId);
      expect(data.endsAt).to.eq(Number(data.openedAt.toString()) + DAY);
    });
    it('should open a session with valid amount for direct user payment', async () => {
      await setTime(payoutStart + 10 * DAY);
      const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), true, msg, signature);

      const sessionId = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 0);
      const data = await sessionRouter.getSession(sessionId);
      expect(data.user).to.eq(SECOND);
      expect(data.bidId).to.eq(bidId);
      expect(data.stake).to.eq(wei(50));
      expect(data.closeoutReceipt).to.eq('0x');
      expect(data.closeoutType).to.eq(0);
      expect(data.providerWithdrawnAmount).to.eq(0);
      expect(data.openedAt).to.eq(payoutStart + 10 * DAY + 1);
      expect(data.endsAt).to.greaterThan(data.openedAt);
      expect(data.closedAt).to.eq(0);
      expect(data.isActive).to.eq(true);
      expect(data.isDirectPaymentFromUser).to.eq(true);

      const tokenBalAfter = await token.balanceOf(sessionRouter);
      expect(tokenBalAfter - tokenBalBefore).to.eq(wei(50));
      const secondBalAfter = await token.balanceOf(SECOND);
      expect(secondBalBefore - secondBalAfter).to.eq(wei(50));

      expect(await sessionRouter.getIsProviderApprovalUsed(msg)).to.eq(true);
      expect(await sessionRouter.getUserSessions(SECOND, 0, 10)).to.deep.eq([[sessionId], 1n]);
      expect(await sessionRouter.getProviderSessions(PROVIDER, 0, 10)).to.deep.eq([[sessionId], 1n]);
      expect(await sessionRouter.getModelSessions(modelId, 0, 10)).to.deep.eq([[sessionId], 1n]);
    });
    it('should throw error when the approval is for an another chain', async () => {
      const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, bidId, 1n);
      await expect(
        sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature),
      ).to.be.revertedWithCustomError(sessionRouter, 'SessionApprovedForAnotherChainId');
    });
    it('should throw error when an approval expired', async () => {
      await setTime(payoutStart);
      const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await setTime(payoutStart + 600);
      await expect(
        sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature),
      ).to.be.revertedWithCustomError(sessionRouter, 'SessionApproveExpired');
    });
    it('should throw error when approval is not from provider and without delegation', async () => {
      const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await expect(
        sessionRouter.connect(OWNER).openSession(PROVIDER, wei(50), false, msg, signature),
      ).to.be.revertedWithCustomError(sessionRouter, 'InsufficientRightsForOperation');
    });
    it('should throw error when the bid is not found', async () => {
      const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, getHex(Buffer.from('1')));
      await expect(
        sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature),
      ).to.be.revertedWithCustomError(sessionRouter, 'SessionBidNotFound');
    });
    it('should throw error when a previously posted bid was deleted', async () => {
      const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await marketplace.connect(PROVIDER).deleteModelBid(bidId);

      await expect(
        sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature),
      ).to.be.revertedWithCustomError(sessionRouter, 'SessionBidNotFound');
    });
    it('should throw error when direct-payment stake is lower than its stipend', async () => {
      await sessionRouter.setPoolConfig(3, {
        payoutStart: 0,
        decreaseInterval: DAY,
        initialReward: wei(1_000_000_000),
        rewardDecrease: 0,
      });
      await setTime(payoutStart + 10 * DAY);
      const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, bidId);

      await expect(
        sessionRouter.connect(SECOND).openSession(SECOND, wei(1), true, msg, signature),
      ).to.be.revertedWithCustomError(sessionRouter, 'SessionStakeTooLow');
    });
    it('should throw error when the signature mismatch', async () => {
      const { msg, signature } = await getProviderApproval(OWNER, SECOND, bidId);
      await expect(
        sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature),
      ).to.be.revertedWithCustomError(sessionRouter, 'SessionProviderSignatureMismatch');
    });
    it('should throw error when an approval duplicated', async () => {
      await setTime(payoutStart + 10 * DAY);
      const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature);
      await expect(
        sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature),
      ).to.be.revertedWithCustomError(sessionRouter, 'SessionDuplicateApproval');
    });
    it('should throw error when session duration too short', async () => {
      const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await expect(
        sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature),
      ).to.be.revertedWithCustomError(sessionRouter, 'SessionTooShort');
    });
    it('should throw error when open a session without delegation or with incorrect rules', async () => {
      await setTime(payoutStart + 10 * DAY);
      const { msg, signature } = await getProviderApproval(PROVIDER, OWNER, bidId);

      await expect(
        sessionRouter.connect(OWNER).openSession(SECOND, wei(50), false, msg, signature),
      ).to.be.revertedWithCustomError(providerRegistry, 'InsufficientRightsForOperation');

      await delegateRegistry
        .connect(SECOND)
        .delegateContract(OWNER, providerRegistry, getHex(Buffer.from('123')), true);
      await expect(
        sessionRouter.connect(OWNER).openSession(SECOND, wei(50), false, msg, signature),
      ).to.be.revertedWithCustomError(providerRegistry, 'InsufficientRightsForOperation');
    });
  });

  describe('#closeSession', () => {
    it('should emit SessionOpened, ProviderRewardAccrued, SessionRating, and SessionClosed on the paid path', async () => {
      await setTime(payoutStart + 10 * DAY);
      const nonce = await sessionRouter.getTotalSessions(PROVIDER);
      const expectedSessionId = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, nonce);
      const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, bidId);

      await expect(sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature))
        .to.emit(sessionRouter, 'SessionOpened')
        .withArgs(SECOND, expectedSessionId, PROVIDER);

      const opened = await sessionRouter.getSession(expectedSessionId);
      await setTime(Number(opened.openedAt) + 200);
      const { msg: receiptMsg, signature: receiptSig } = await getReceipt(PROVIDER, expectedSessionId, 12, 34);
      const reward = bidPricePerSecond * 201n;

      const closeTx = sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);
      await expect(closeTx)
        .to.emit(sessionRouter, 'ProviderRewardAccrued')
        .withArgs(expectedSessionId, PROVIDER, reward, reward, 0);
      await expect(closeTx)
        .to.emit(sessionRouter, 'SessionRating')
        .withArgs(expectedSessionId, modelId, PROVIDER, 0, 12_000, 34);
      await expect(closeTx).to.emit(sessionRouter, 'SessionClosed').withArgs(SECOND, expectedSessionId, PROVIDER);
    });
    it('should emit disputed SessionRating with closeout type 1', async () => {
      const { sessionId, openedAt } = await _createSession();
      await setTime(openedAt + 200);
      const { msg: receiptMsg } = await getReceipt(PROVIDER, sessionId, 7, 89);
      const { signature: receiptSig } = await getReceipt(OWNER, sessionId, 7, 89);

      await expect(sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig))
        .to.emit(sessionRouter, 'SessionRating')
        .withArgs(sessionId, modelId, PROVIDER, 1, 7_000, 89);
    });
    it('should settle exact direct-pay balances for a late close', async () => {
      const created = await _createSession(true);
      const session = await sessionRouter.getSession(created.sessionId);
      await _assertDirectClose(created.sessionId, Number(session.endsAt) + DAY, PROVIDER);
    });
    it('should settle exact direct-pay balances for an undisputed early close', async () => {
      const created = await _createSession(true);
      const session = await sessionRouter.getSession(created.sessionId);
      await _assertDirectClose(created.sessionId, Number(session.openedAt) + 200, PROVIDER);
    });
    it('should settle exact direct-pay balances for a disputed same-day close', async () => {
      const created = await _createSession(true);
      const session = await sessionRouter.getSession(created.sessionId);
      await _assertDirectClose(created.sessionId, Number(session.openedAt) + 200, OWNER);
    });
    it('should settle exact direct-pay balances for a disputed next-day close', async () => {
      const created = await _createSession(true);
      const session = await sessionRouter.getSession(created.sessionId);
      const nextDay = Number(await providerPayouts.startOfTheDay(session.openedAt)) + DAY + 1;
      await _assertDirectClose(created.sessionId, nextDay, OWNER);
    });
    it('should close session and send rewards for the provider, late closure', async () => {
      const { sessionId, openedAt } = await _createSession();

      const providerBalBefore = await token.balanceOf(PROVIDER);
      const fundingBalBefore = await token.balanceOf(FUNDING);

      await setTime(openedAt + 5 * DAY);
      const { msg: receiptMsg } = await getReceipt(PROVIDER, sessionId, 0, 0);
      const { signature: receiptSig } = await getReceipt(OWNER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      const session = await sessionRouter.getSession(sessionId);
      const duration = session.endsAt - session.openedAt;

      expect(session.closedAt).to.eq(openedAt + 5 * DAY + 1);
      expect(session.isActive).to.eq(false);
      expect(session.closeoutReceipt).to.eq(receiptMsg);
      expect(session.providerWithdrawnAmount).to.eq(bidPricePerSecond * duration);

      expect((await sessionRouter.getProvider(PROVIDER)).limitPeriodEarned).to.eq(bidPricePerSecond * duration);

      const providerBalAfter = await token.balanceOf(PROVIDER);
      expect(providerBalAfter - providerBalBefore).to.eq(bidPricePerSecond * duration);
      const fundingBalAfter = await token.balanceOf(FUNDING);
      expect(fundingBalBefore - fundingBalAfter).to.eq(bidPricePerSecond * duration);
    });
    it('should close session and send rewards for the provider from the delegatee address, late closure', async () => {
      const { sessionId, openedAt } = await _createSession();

      const providerBalBefore = await token.balanceOf(PROVIDER);
      const fundingBalBefore = await token.balanceOf(FUNDING);

      await delegateRegistry
        .connect(SECOND)
        .delegateContract(OWNER, providerRegistry, await providerRegistry.DELEGATION_RULES_SESSION(), true);

      await setTime(openedAt + 5 * DAY);
      const { msg: receiptMsg } = await getReceipt(PROVIDER, sessionId, 0, 0);
      const { signature: receiptSig } = await getReceipt(OWNER, sessionId, 0, 0);
      await sessionRouter.connect(OWNER).closeSession(receiptMsg, receiptSig);

      const session = await sessionRouter.getSession(sessionId);
      const duration = session.endsAt - session.openedAt;

      expect(session.closedAt).to.eq(openedAt + 5 * DAY + 1);
      expect(session.isActive).to.eq(false);
      expect(session.closeoutReceipt).to.eq(receiptMsg);
      expect(session.providerWithdrawnAmount).to.eq(bidPricePerSecond * duration);

      expect((await sessionRouter.getProvider(PROVIDER)).limitPeriodEarned).to.eq(bidPricePerSecond * duration);

      const providerBalAfter = await token.balanceOf(PROVIDER);
      expect(providerBalAfter - providerBalBefore).to.eq(bidPricePerSecond * duration);
      const fundingBalAfter = await token.balanceOf(FUNDING);
      expect(fundingBalBefore - fundingBalAfter).to.eq(bidPricePerSecond * duration);
    });
    it('should close session and send rewards for the provider, no dispute, early closure', async () => {
      const { sessionId, openedAt } = await _createSession();
      const providerBalBefore = await token.balanceOf(PROVIDER);

      const fundingBalBefore = await token.balanceOf(FUNDING);
      await setTime(openedAt + 200);
      const { msg: receiptMsg, signature: receiptSig } = await getReceipt(PROVIDER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      const session = await sessionRouter.getSession(sessionId);
      const duration = session.closedAt - session.openedAt;

      expect(session.closedAt).to.eq(openedAt + 201);
      expect(session.isActive).to.eq(false);
      expect(session.closeoutReceipt).to.eq(receiptMsg);
      expect(session.providerWithdrawnAmount).to.eq(bidPricePerSecond * duration);

      expect((await sessionRouter.getProvider(PROVIDER)).limitPeriodEarned).to.eq(bidPricePerSecond * duration);

      const providerBalAfter = await token.balanceOf(PROVIDER);
      expect(providerBalAfter - providerBalBefore).to.eq(bidPricePerSecond * duration);
      const fundingBalAfter = await token.balanceOf(FUNDING);
      expect(fundingBalBefore - fundingBalAfter).to.eq(bidPricePerSecond * duration);
    });
    it('should close session and send rewards for the provider, with dispute, early closure on the next day', async () => {
      const { sessionId, secondsToDayEnd, openedAt } = await _createSession();

      const providerBalBefore = await token.balanceOf(PROVIDER);
      const fundingBalBefore = await token.balanceOf(FUNDING);

      await setTime(openedAt + secondsToDayEnd + 1);
      const { msg: receiptMsg } = await getReceipt(PROVIDER, sessionId, 0, 0);
      const { signature: receiptSig } = await getReceipt(OWNER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      const session = await sessionRouter.getSession(sessionId);
      const duration = BigInt(secondsToDayEnd);

      expect(session.closedAt).to.eq(openedAt + secondsToDayEnd + 2);
      expect(session.isActive).to.eq(false);
      expect(session.closeoutReceipt).to.eq(receiptMsg);
      expect(session.providerWithdrawnAmount).to.eq(bidPricePerSecond * duration);

      expect((await sessionRouter.getProvider(PROVIDER)).limitPeriodEarned).to.eq(bidPricePerSecond * duration);

      const providerBalAfter = await token.balanceOf(PROVIDER);
      expect(providerBalAfter - providerBalBefore).to.eq(bidPricePerSecond * duration);
      const fundingBalAfter = await token.balanceOf(FUNDING);
      expect(fundingBalBefore - fundingBalAfter).to.eq(bidPricePerSecond * duration);
    });
    it('should close session and send rewards for the provider from the delegatee address, with dispute, early closure on the next day', async () => {
      const { sessionId, secondsToDayEnd, openedAt } = await _createSession();

      const providerBalBefore = await token.balanceOf(PROVIDER);
      const fundingBalBefore = await token.balanceOf(FUNDING);

      await delegateRegistry
        .connect(SECOND)
        .delegateContract(OWNER, providerRegistry, await providerRegistry.DELEGATION_RULES_SESSION(), true);

      await setTime(openedAt + secondsToDayEnd + 1);
      const { msg: receiptMsg } = await getReceipt(PROVIDER, sessionId, 0, 0);
      const { signature: receiptSig } = await getReceipt(OWNER, sessionId, 0, 0);
      await sessionRouter.connect(OWNER).closeSession(receiptMsg, receiptSig);

      const session = await sessionRouter.getSession(sessionId);
      const duration = BigInt(secondsToDayEnd);

      expect(session.closedAt).to.eq(openedAt + secondsToDayEnd + 2);
      expect(session.isActive).to.eq(false);
      expect(session.closeoutReceipt).to.eq(receiptMsg);
      expect(session.providerWithdrawnAmount).to.eq(bidPricePerSecond * duration);

      expect((await sessionRouter.getProvider(PROVIDER)).limitPeriodEarned).to.eq(bidPricePerSecond * duration);

      const providerBalAfter = await token.balanceOf(PROVIDER);
      expect(providerBalAfter - providerBalBefore).to.eq(bidPricePerSecond * duration);
      const fundingBalAfter = await token.balanceOf(FUNDING);
      expect(fundingBalBefore - fundingBalAfter).to.eq(bidPricePerSecond * duration);
    });
    it('should close session and do not send rewards for the provider, with dispute, early closure on the same day', async () => {
      const { sessionId, secondsToDayEnd, openedAt } = await _createSession();

      const providerBalBefore = await token.balanceOf(PROVIDER);
      const fundingBalBefore = await token.balanceOf(FUNDING);

      await setTime(openedAt + secondsToDayEnd - 50);
      const { msg: receiptMsg } = await getReceipt(PROVIDER, sessionId, 0, 0);
      const { signature: receiptSig } = await getReceipt(OWNER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      const session = await sessionRouter.getSession(sessionId);

      const duration = 0n;
      expect(session.closedAt).to.eq(openedAt + secondsToDayEnd - 49);
      expect(session.isActive).to.eq(false);
      expect(session.closeoutReceipt).to.eq(receiptMsg);

      expect(session.providerWithdrawnAmount).to.eq(bidPricePerSecond * duration);

      expect((await sessionRouter.getProvider(PROVIDER)).limitPeriodEarned).to.eq(bidPricePerSecond * duration);
      const providerBalAfter = await token.balanceOf(PROVIDER);
      expect(providerBalAfter - providerBalBefore).to.eq(bidPricePerSecond * duration);
      const fundingBalAfter = await token.balanceOf(FUNDING);
      expect(fundingBalBefore - fundingBalAfter).to.eq(bidPricePerSecond * duration);
    });
    it('should close session and send rewards for the provider, early closure', async () => {
      const { sessionId, openedAt, secondsToDayEnd } = await _createSession(true);

      const providerBalBefore = await token.balanceOf(PROVIDER);
      const fundingBalBefore = await token.balanceOf(FUNDING);
      const contractBalBefore = await token.balanceOf(sessionRouter);
      const secondBalBefore = await token.balanceOf(SECOND);

      await setTime(openedAt + secondsToDayEnd + 100);
      const { msg: receiptMsg } = await getReceipt(PROVIDER, sessionId, 0, 0);
      const { signature: receiptSig } = await getReceipt(OWNER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      const session = await sessionRouter.getSession(sessionId);
      const duration = BigInt(secondsToDayEnd);

      expect(session.closedAt).to.eq(openedAt + secondsToDayEnd + 100 + 1);
      expect(session.isActive).to.eq(false);
      expect(session.closeoutReceipt).to.eq(receiptMsg);
      expect(session.providerWithdrawnAmount).to.eq(bidPricePerSecond * duration);

      expect((await sessionRouter.getProvider(PROVIDER)).limitPeriodEarned).to.eq(bidPricePerSecond * duration);

      const providerBalAfter = await token.balanceOf(PROVIDER);
      expect(providerBalAfter - providerBalBefore).to.eq(bidPricePerSecond * duration);
      const fundingBalAfter = await token.balanceOf(FUNDING);
      expect(fundingBalBefore - fundingBalAfter).to.eq(0);
      const contractBalAfter = await token.balanceOf(sessionRouter);
      const secondBalAfter = await token.balanceOf(SECOND);
      expect(contractBalBefore - contractBalAfter).to.eq(
        bidPricePerSecond * duration + secondBalAfter - secondBalBefore,
      );
    });
    it('should close session and send rewards for the user, late closure', async () => {
      const { sessionId, openedAt } = await _createSession();

      const userBalBefore = await token.balanceOf(SECOND);
      const contractBalBefore = await token.balanceOf(sessionRouter);

      await setTime(openedAt + 5 * DAY);
      const { msg: receiptMsg, signature: receiptSig } = await getReceipt(PROVIDER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      const stakesOnHold = await providerPayouts.getUserStakesOnHold(SECOND, 20);
      expect(stakesOnHold[0]).to.eq(0);
      expect(stakesOnHold[1]).to.eq(0);

      const userBalAfter = await token.balanceOf(SECOND);
      expect(userBalAfter - userBalBefore).to.eq(wei(50));
      const contractBalAfter = await token.balanceOf(sessionRouter);
      expect(contractBalBefore - contractBalAfter).to.eq(wei(50));
    });
    it('should day-lock the used stipend when closing late on the same day the session ended (#830)', async () => {
      const { sessionId } = await _createSession();
      const session = await sessionRouter.getSession(sessionId);

      const userBalBefore = await token.balanceOf(SECOND);

      // late close, but still inside the stipend epoch (same calendar day as endsAt)
      await setTime(Number(session.endsAt) + 100);
      const { msg: receiptMsg, signature: receiptSig } = await getReceipt(PROVIDER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      const stakesOnHold = await providerPayouts.getUserStakesOnHold(SECOND, 20);
      expect(stakesOnHold[0]).to.eq(0);
      expect(stakesOnHold[1]).to.greaterThan(0);

      const userBalAfter = await token.balanceOf(SECOND);
      expect(userBalAfter - userBalBefore).to.eq(wei(50) - stakesOnHold[1]);
    });
    it('should close session and send rewards for the user, early closure', async () => {
      const { sessionId, openedAt, secondsToDayEnd } = await _createSession();

      const userBalBefore = await token.balanceOf(SECOND);
      const contractBalBefore = await token.balanceOf(sessionRouter);

      await setTime(openedAt + secondsToDayEnd + 1);
      const { msg: receiptMsg, signature: receiptSig } = await getReceipt(PROVIDER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      const stakesOnHold = await providerPayouts.getUserStakesOnHold(SECOND, 1);
      expect(stakesOnHold[0]).to.eq(0);
      expect(stakesOnHold[1]).to.greaterThan(0);

      const userBalAfter = await token.balanceOf(SECOND);
      expect(userBalAfter - userBalBefore).to.lessThan(wei(50));
      const contractBalAfter = await token.balanceOf(sessionRouter);
      expect(contractBalBefore - contractBalAfter).to.lessThan(wei(50));

      await sessionRouter.getProviderModelStats(modelId, PROVIDER);
      await sessionRouter.getModelStats(modelId);
    });
    it('should claim provider rewards and close session, late closure', async () => {
      const { sessionId, openedAt } = await _createSession(true);

      let userBalBefore = await token.balanceOf(SECOND);
      let providerBalBefore = await token.balanceOf(PROVIDER);
      let contractBalBefore = await token.balanceOf(sessionRouter);

      // Claim for Provider
      await setTime(openedAt + 5 * DAY);
      await sessionRouter.connect(PROVIDER).claimForProvider(sessionId);

      let session = await sessionRouter.getSession(sessionId);
      const duration = session.endsAt - session.openedAt;

      expect(session.providerWithdrawnAmount).to.eq(bidPricePerSecond * duration);
      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(bidPricePerSecond * duration);
      expect((await sessionRouter.getProvider(PROVIDER)).limitPeriodEarned).to.eq(bidPricePerSecond * duration);

      const userBalAfter = await token.balanceOf(SECOND);
      expect(userBalAfter - userBalBefore).to.eq(0);
      const providerBalAfter = await token.balanceOf(PROVIDER);
      expect(providerBalAfter - providerBalBefore).to.eq(bidPricePerSecond * duration);
      const contractBalAfter = await token.balanceOf(sessionRouter);
      expect(contractBalBefore - contractBalAfter).to.eq(bidPricePerSecond * duration);

      // Close session
      userBalBefore = userBalAfter;
      providerBalBefore = providerBalAfter;
      contractBalBefore = contractBalAfter;

      await setTime(openedAt + 6 * DAY);
      const { msg: receiptMsg, signature: receiptSig } = await getReceipt(PROVIDER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      const userBalAfterClose = await token.balanceOf(SECOND);
      expect(userBalAfterClose - userBalBefore).to.eq(wei(50) - bidPricePerSecond * duration);
      const providerBalAfterClose = await token.balanceOf(PROVIDER);
      expect(providerBalAfterClose - providerBalBefore).to.eq(0);
      const contractBalAfterClose = await token.balanceOf(sessionRouter);
      expect(contractBalBefore - contractBalAfterClose).to.eq(wei(50) - bidPricePerSecond * duration);

      session = await sessionRouter.getSession(sessionId);
      expect(session.closedAt).to.eq(openedAt + 6 * DAY + 1);
      expect(session.isActive).to.eq(false);
      expect(session.closeoutReceipt).to.eq(receiptMsg);
      expect(session.providerWithdrawnAmount).to.eq(bidPricePerSecond * duration);
    });
    it('should throw error when the caller is invalid', async () => {
      const { msg: receiptMsg, signature: receiptSig } = await getReceipt(PROVIDER, getHex(Buffer.from('1')), 0, 0);
      await expect(sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig)).to.be.revertedWithCustomError(
        sessionRouter,
        'InsufficientRightsForOperation',
      );
    });
    it('should throw error when the session already closed', async () => {
      const { sessionId, openedAt, secondsToDayEnd } = await _createSession();

      await setTime(openedAt + secondsToDayEnd + 1);
      const { msg: receiptMsg, signature: receiptSig } = await getReceipt(PROVIDER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      await expect(sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig)).to.be.revertedWithCustomError(
        sessionRouter,
        'SessionAlreadyClosed',
      );
    });
    it('should throw error when the provider receipt for another chain', async () => {
      const { sessionId } = await _createSession();

      const { msg: receiptMsg, signature: receiptSig } = await getReceipt(PROVIDER, sessionId, 0, 0, 1n);
      await expect(sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig)).to.be.revertedWithCustomError(
        sessionRouter,
        'SessionReceiptForAnotherChainId',
      );
    });
    it('should throw error when the provider receipt expired', async () => {
      const { sessionId, openedAt } = await _createSession();

      await setTime(openedAt + 100);
      const { msg: receiptMsg, signature: receiptSig } = await getReceipt(PROVIDER, sessionId, 0, 0);

      await setTime(openedAt + 10000);
      await expect(sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig)).to.be.revertedWithCustomError(
        sessionRouter,
        'SessionReceiptExpired',
      );
    });
  });

  describe('#claimForProvider', () => {
    it('should return zero without transfer immediately after an undisputed same-day close', async () => {
      const { sessionId, openedAt } = await _createSession();
      await setTime(openedAt + 200);
      const { msg: receiptMsg, signature: receiptSig } = await getReceipt(PROVIDER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      const providerBalance = await token.balanceOf(PROVIDER);
      const fundingBalance = await token.balanceOf(FUNDING);
      const withdrawn = (await sessionRouter.getSession(sessionId)).providerWithdrawnAmount;

      await sessionRouter.connect(OWNER).claimForProvider(sessionId);

      expect((await sessionRouter.getSession(sessionId)).providerWithdrawnAmount).to.eq(withdrawn);
      expect(await token.balanceOf(PROVIDER)).to.eq(providerBalance);
      expect(await token.balanceOf(FUNDING)).to.eq(fundingBalance);
    });
    it('should revert for an unknown session ID without moving funds', async () => {
      const unknownSessionId = getHex(Buffer.from('unknown-session'));
      const providerBalance = await token.balanceOf(PROVIDER);
      const fundingBalance = await token.balanceOf(FUNDING);

      await expect(sessionRouter.connect(OWNER).claimForProvider(unknownSessionId)).to.be.revertedWithCustomError(
        sessionRouter,
        'SessionNotEndedOrNotExist',
      );

      expect(await token.balanceOf(PROVIDER)).to.eq(providerBalance);
      expect(await token.balanceOf(FUNDING)).to.eq(fundingBalance);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(0);
    });
    it('should claim provider rewards, remainder, with dispute, early closure on the next day', async () => {
      const { sessionId, secondsToDayEnd, openedAt } = await _createSession();

      await setTime(openedAt + secondsToDayEnd + 1);
      const { msg: receiptMsg } = await getReceipt(PROVIDER, sessionId, 0, 0);
      const { signature: receiptSig } = await getReceipt(OWNER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      let session = await sessionRouter.getSession(sessionId);
      const fullDuration = BigInt(secondsToDayEnd + 1);
      const duration = 1n;

      const providerBalBefore = await token.balanceOf(PROVIDER);
      const fundingBalBefore = await token.balanceOf(FUNDING);

      await setTime(openedAt + secondsToDayEnd + 1 + 1 * DAY);
      await sessionRouter.connect(PROVIDER).claimForProvider(sessionId);
      session = await sessionRouter.getSession(sessionId);

      expect(session.providerWithdrawnAmount).to.eq(bidPricePerSecond * fullDuration);
      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(bidPricePerSecond * fullDuration);
      expect((await sessionRouter.getProvider(PROVIDER)).limitPeriodEarned).to.eq(bidPricePerSecond * fullDuration);

      const providerBalAfter = await token.balanceOf(PROVIDER);
      expect(providerBalAfter - providerBalBefore).to.eq(bidPricePerSecond * duration);
      const fundingBalAfter = await token.balanceOf(FUNDING);
      expect(fundingBalBefore - fundingBalAfter).to.eq(bidPricePerSecond * duration);
    });
    it('should claim provider rewards from the delegatee address, remainder, with dispute, early closure on the next day', async () => {
      const { sessionId, secondsToDayEnd, openedAt } = await _createSession();

      await setTime(openedAt + secondsToDayEnd + 1);
      const { msg: receiptMsg } = await getReceipt(PROVIDER, sessionId, 0, 0);
      const { signature: receiptSig } = await getReceipt(OWNER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      let session = await sessionRouter.getSession(sessionId);
      const fullDuration = BigInt(secondsToDayEnd + 1);
      const duration = 1n;

      const providerBalBefore = await token.balanceOf(PROVIDER);
      const fundingBalBefore = await token.balanceOf(FUNDING);

      await delegateRegistry
        .connect(PROVIDER)
        .delegateContract(OWNER, providerRegistry, await providerRegistry.DELEGATION_RULES_SESSION(), true);

      await setTime(openedAt + secondsToDayEnd + 1 + 1 * DAY);
      await sessionRouter.connect(OWNER).claimForProvider(sessionId);
      session = await sessionRouter.getSession(sessionId);

      expect(session.providerWithdrawnAmount).to.eq(bidPricePerSecond * fullDuration);
      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(bidPricePerSecond * fullDuration);
      expect((await sessionRouter.getProvider(PROVIDER)).limitPeriodEarned).to.eq(bidPricePerSecond * fullDuration);

      const providerBalAfter = await token.balanceOf(PROVIDER);
      expect(providerBalAfter - providerBalBefore).to.eq(bidPricePerSecond * duration);
      const fundingBalAfter = await token.balanceOf(FUNDING);
      expect(fundingBalBefore - fundingBalAfter).to.eq(bidPricePerSecond * duration);
    });
    it('should not claim provider rewards, when provider funds on lock', async () => {
      const { sessionId, secondsToDayEnd, openedAt } = await _createSession();

      await setTime(openedAt + secondsToDayEnd + 1);
      const { msg: receiptMsg } = await getReceipt(PROVIDER, sessionId, 0, 0);
      const { signature: receiptSig } = await getReceipt(OWNER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      let session = await sessionRouter.getSession(sessionId);
      const duration = BigInt(secondsToDayEnd);

      const providerBalBefore = await token.balanceOf(PROVIDER);
      const fundingBalBefore = await token.balanceOf(FUNDING);

      await setTime(openedAt + secondsToDayEnd + 100);
      await sessionRouter.connect(PROVIDER).claimForProvider(sessionId);
      session = await sessionRouter.getSession(sessionId);

      expect(session.providerWithdrawnAmount).to.eq(bidPricePerSecond * duration);
      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(bidPricePerSecond * duration);
      expect((await sessionRouter.getProvider(PROVIDER)).limitPeriodEarned).to.eq(bidPricePerSecond * duration);

      const providerBalAfter = await token.balanceOf(PROVIDER);
      expect(providerBalAfter - providerBalBefore).to.eq(0);
      const fundingBalAfter = await token.balanceOf(FUNDING);
      expect(fundingBalBefore - fundingBalAfter).to.eq(0);
    });
    it('should claim provider rewards, full', async () => {
      const { sessionId, openedAt } = await _createSession();

      let session = await sessionRouter.getSession(sessionId);
      const duration = session.endsAt - session.openedAt;

      const providerBalBefore = await token.balanceOf(PROVIDER);
      const fundingBalBefore = await token.balanceOf(FUNDING);

      await setTime(openedAt + 5 * DAY + 1);
      await sessionRouter.connect(PROVIDER).claimForProvider(sessionId);
      session = await sessionRouter.getSession(sessionId);

      expect(session.providerWithdrawnAmount).to.eq(bidPricePerSecond * duration);
      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(bidPricePerSecond * duration);
      expect((await sessionRouter.getProvider(PROVIDER)).limitPeriodEarned).to.eq(bidPricePerSecond * duration);

      const providerBalAfter = await token.balanceOf(PROVIDER);
      expect(providerBalAfter - providerBalBefore).to.eq(bidPricePerSecond * duration);
      const fundingBalAfter = await token.balanceOf(FUNDING);
      expect(fundingBalBefore - fundingBalAfter).to.eq(bidPricePerSecond * duration);
    });
    it('should claim provider rewards with reward limiter amount for the period', async () => {
      const providerBalBefore = await token.balanceOf(PROVIDER);
      const fundingBalBefore = await token.balanceOf(FUNDING);

      await setTime(payoutStart + 10 * DAY);
      const { msg: msg1, signature: sig1 } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg1, sig1);

      const sessionId1 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 0);
      await setTime(payoutStart + 20 * DAY);
      await sessionRouter.connect(PROVIDER).claimForProvider(sessionId1);

      await setTime(payoutStart + 30 * DAY);
      const { msg: msg2, signature: sig2 } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg2, sig2);

      const sessionId2 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 1);
      await setTime(payoutStart + 40 * DAY);
      await sessionRouter.connect(PROVIDER).claimForProvider(sessionId2);

      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(wei(0.2));
      expect((await sessionRouter.getProvider(PROVIDER)).limitPeriodEarned).to.eq(wei(0.2));

      const providerBalAfter = await token.balanceOf(PROVIDER);
      expect(providerBalAfter - providerBalBefore).to.eq(wei(0.2));
      const fundingBalAfter = await token.balanceOf(FUNDING);
      expect(fundingBalBefore - fundingBalAfter).to.eq(wei(0.2));
    });
    it('should claim zero when session is not end', async () => {
      const { sessionId, openedAt } = await _createSession();

      const providerBalBefore = await token.balanceOf(PROVIDER);
      const fundingBalBefore = await token.balanceOf(FUNDING);

      await setTime(openedAt + 10);
      await sessionRouter.connect(PROVIDER).claimForProvider(sessionId);

      const providerBalAfter = await token.balanceOf(PROVIDER);
      expect(providerBalAfter - providerBalBefore).to.eq(0);
      const fundingBalAfter = await token.balanceOf(FUNDING);
      expect(fundingBalBefore - fundingBalAfter).to.eq(0);
    });
    it('should allow anyone to claim while always paying the provider', async () => {
      const { sessionId } = await _createSession();
      const session = await sessionRouter.getSession(sessionId);
      const reward = (session.endsAt - session.openedAt) * bidPricePerSecond;

      await setTime(Number(session.endsAt) + DAY);
      await expect(sessionRouter.connect(SECOND).claimForProvider(sessionId))
        .to.emit(sessionRouter, 'ProviderRewardAccrued')
        .withArgs(sessionId, PROVIDER, reward, reward, 0);
      expect(await token.balanceOf(PROVIDER)).to.eq(wei(10000) - wei(0.2) - wei(100) + reward);

      await expect(sessionRouter.connect(OWNER).claimForProvider(sessionId)).to.changeTokenBalances(
        token,
        [PROVIDER, FUNDING, OWNER],
        [0, 0, 0],
      );
      expect((await sessionRouter.getSession(sessionId)).providerWithdrawnAmount).to.eq(reward);
    });
  });

  describe('#withdrawUserStakes', () => {
    it('should throw InsufficientRightsForOperation for an unauthorized caller', async () => {
      await expect(sessionRouter.connect(OWNER).withdrawUserStakes(SECOND, 1)).to.be.revertedWithCustomError(
        sessionRouter,
        'InsufficientRightsForOperation',
      );
    });
    it('should withdraw the user stake on hold, one entity', async () => {
      const openedAt = payoutStart + (payoutStart % DAY) + 10 * DAY - 201;

      await setTime(openedAt + 1 * DAY);
      const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature);
      const sessionId1 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 0);

      await setTime(openedAt + 1 * DAY + 100);
      const { msg: receiptMsg, signature: receiptSig } = await getReceipt(PROVIDER, sessionId1, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      const userBalBefore = await token.balanceOf(SECOND);
      const contractBalBefore = await token.balanceOf(sessionRouter);

      await setTime(openedAt + 3 * DAY + 2);
      const [available] = await providerPayouts.getUserStakesOnHold(SECOND, 1);
      expect(available).to.greaterThan(0);
      await expect(sessionRouter.connect(SECOND).withdrawUserStakes(SECOND, 1))
        .to.emit(sessionRouter, 'UserWithdrawn')
        .withArgs(SECOND, available);

      const stakesOnHold = await providerPayouts.getUserStakesOnHold(SECOND, 1);
      expect(stakesOnHold[0]).to.eq(0);
      expect(stakesOnHold[1]).to.eq(0);

      const userBalAfter = await token.balanceOf(SECOND);
      expect(userBalAfter - userBalBefore).to.eq(available);
      const contractBalAfter = await token.balanceOf(sessionRouter);
      expect(contractBalBefore - contractBalAfter).to.eq(available);
    });
    it('should withdraw the user stake on hold from the delegatee address, one entity', async () => {
      const openedAt = payoutStart + (payoutStart % DAY) + 10 * DAY - 201;

      await setTime(openedAt + 1 * DAY);
      const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature);
      const sessionId1 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 0);

      await setTime(openedAt + 1 * DAY + 100);
      const { msg: receiptMsg, signature: receiptSig } = await getReceipt(PROVIDER, sessionId1, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig);

      const userBalBefore = await token.balanceOf(SECOND);
      const contractBalBefore = await token.balanceOf(sessionRouter);

      await delegateRegistry
        .connect(SECOND)
        .delegateContract(OWNER, providerRegistry, await providerRegistry.DELEGATION_RULES_SESSION(), true);

      await setTime(openedAt + 3 * DAY + 2);
      await sessionRouter.connect(OWNER).withdrawUserStakes(SECOND, 1);

      const stakesOnHold = await providerPayouts.getUserStakesOnHold(SECOND, 1);
      expect(stakesOnHold[0]).to.eq(0);
      expect(stakesOnHold[1]).to.eq(0);

      const userBalAfter = await token.balanceOf(SECOND);
      expect(userBalAfter - userBalBefore).to.greaterThan(0);
      const contractBalAfter = await token.balanceOf(sessionRouter);
      expect(contractBalBefore - contractBalAfter).to.greaterThan(0);
    });
    it('should withdraw the user stake on hold, few entities, with on hold', async () => {
      const openedAt = payoutStart + (payoutStart % DAY) + 10 * DAY - 201;

      await setTime(openedAt + 1 * DAY);
      const { msg: msg1, signature: sig1 } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg1, sig1);
      const sessionId1 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 0);

      await setTime(openedAt + 1 * DAY + 500);
      const { msg: receiptMsg1, signature: receiptSig1 } = await getReceipt(PROVIDER, sessionId1, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg1, receiptSig1);

      await setTime(openedAt + 3 * DAY);
      const { msg: msg2, signature: sig2 } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg2, sig2);
      const sessionId2 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 1);

      await setTime(openedAt + 3 * DAY + 500);
      const { msg: receiptMsg2, signature: receiptSig2 } = await getReceipt(PROVIDER, sessionId2, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg2, receiptSig2);

      const stakesOnHold = await providerPayouts.getUserStakesOnHold(SECOND, 20);
      expect(stakesOnHold[0]).to.greaterThan(0);
      expect(stakesOnHold[1]).to.greaterThan(0);

      const userBalBefore = await token.balanceOf(SECOND);
      const contractBalBefore = await token.balanceOf(sessionRouter);

      await setTime(openedAt + 4 * DAY);
      await sessionRouter.connect(SECOND).withdrawUserStakes(SECOND, 20);

      const userBalAfter = await token.balanceOf(SECOND);
      expect(userBalAfter - userBalBefore).to.eq(stakesOnHold[0]);
      const contractBalAfter = await token.balanceOf(sessionRouter);
      expect(contractBalBefore - contractBalAfter).to.eq(stakesOnHold[0]);
    });
    it('should withdraw the user stake on hold, few entities, with on hold, partial withdraw', async () => {
      const openedAt = payoutStart + (payoutStart % DAY) + 10 * DAY - 201;

      // Open and close session #1
      await setTime(openedAt + 1 * DAY);
      const { msg: msg1, signature: sig1 } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg1, sig1);
      const sessionId1 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 0);

      await setTime(openedAt + 1 * DAY + 500);
      const { msg: receiptMsg1, signature: receiptSig1 } = await getReceipt(PROVIDER, sessionId1, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg1, receiptSig1);

      let stakesOnHold = await providerPayouts.getUserStakesOnHold(SECOND, 10);
      const onHoldAfterSession1 = stakesOnHold[1];

      // Open and close session #2
      await setTime(openedAt + 3 * DAY);
      const { msg: msg2, signature: sig2 } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg2, sig2);
      const sessionId2 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 1);

      await setTime(openedAt + 3 * DAY + 500);
      const { msg: receiptMsg2, signature: receiptSig2 } = await getReceipt(PROVIDER, sessionId2, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg2, receiptSig2);

      stakesOnHold = await providerPayouts.getUserStakesOnHold(SECOND, 10);
      const onHoldAfterSession2 = stakesOnHold[1];

      // Open and close session #3
      await setTime(openedAt + 5 * DAY);
      const { msg: msg3, signature: sig3 } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg3, sig3);
      const sessionId3 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 2);

      await setTime(openedAt + 5 * DAY + 500);
      const { msg: receiptMsg3, signature: receiptSig3 } = await getReceipt(PROVIDER, sessionId3, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg3, receiptSig3);

      stakesOnHold = await providerPayouts.getUserStakesOnHold(SECOND, 10);
      const onHoldAfterSession3 = stakesOnHold[1];

      // First withdraw
      await setTime(openedAt + 99 * DAY);
      stakesOnHold = await providerPayouts.getUserStakesOnHold(SECOND, 10);
      expect(stakesOnHold[0]).to.greaterThan(0);
      expect(stakesOnHold[1]).to.eq(0);

      let userBalBefore = await token.balanceOf(SECOND);
      let contractBalBefore = await token.balanceOf(sessionRouter);

      await sessionRouter.connect(SECOND).withdrawUserStakes(SECOND, 1);

      let userBalAfter = await token.balanceOf(SECOND);
      expect(userBalAfter - userBalBefore).to.eq(onHoldAfterSession3);
      let contractBalAfter = await token.balanceOf(sessionRouter);
      expect(contractBalBefore - contractBalAfter).to.eq(onHoldAfterSession3);

      // Second withdraw
      stakesOnHold = await providerPayouts.getUserStakesOnHold(SECOND, 1);
      expect(stakesOnHold[0]).to.greaterThan(0);
      expect(stakesOnHold[1]).to.eq(0);

      userBalBefore = await token.balanceOf(SECOND);
      contractBalBefore = await token.balanceOf(sessionRouter);

      await sessionRouter.connect(SECOND).withdrawUserStakes(SECOND, 1);

      userBalAfter = await token.balanceOf(SECOND);
      expect(userBalAfter - userBalBefore).to.eq(onHoldAfterSession2);
      contractBalAfter = await token.balanceOf(sessionRouter);
      expect(contractBalBefore - contractBalAfter).to.eq(onHoldAfterSession2);

      // Third withdraw
      stakesOnHold = await providerPayouts.getUserStakesOnHold(SECOND, 1);
      expect(stakesOnHold[0]).to.greaterThan(0);
      expect(stakesOnHold[1]).to.eq(0);

      userBalBefore = await token.balanceOf(SECOND);
      contractBalBefore = await token.balanceOf(sessionRouter);

      await sessionRouter.connect(SECOND).withdrawUserStakes(SECOND, 1);

      userBalAfter = await token.balanceOf(SECOND);
      expect(userBalAfter - userBalBefore).to.eq(onHoldAfterSession1);
      contractBalAfter = await token.balanceOf(sessionRouter);
      expect(contractBalBefore - contractBalAfter).to.eq(onHoldAfterSession1);
    });
    it('should withdraw the user stake on hold, few entities, without on hold', async () => {
      const openedAt = payoutStart + (payoutStart % DAY) + 10 * DAY - 201;

      await setTime(openedAt + 1 * DAY);
      const { msg: msg1, signature: sig1 } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg1, sig1);
      const sessionId1 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 0);

      await setTime(openedAt + 1 * DAY + 500);
      const { msg: receiptMsg1, signature: receiptSig1 } = await getReceipt(PROVIDER, sessionId1, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg1, receiptSig1);

      await setTime(openedAt + 3 * DAY);
      const { msg: msg2, signature: sig2 } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg2, sig2);
      const sessionId2 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 1);

      await setTime(openedAt + 3 * DAY + 500);
      const { msg: receiptMsg2, signature: receiptSig2 } = await getReceipt(PROVIDER, sessionId2, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg2, receiptSig2);

      await setTime(openedAt + 10 * DAY);
      const stakesOnHold = await providerPayouts.getUserStakesOnHold(SECOND, 20);
      expect(stakesOnHold[0]).to.greaterThan(0);
      expect(stakesOnHold[1]).to.eq(0);

      const userBalBefore = await token.balanceOf(SECOND);
      const contractBalBefore = await token.balanceOf(sessionRouter);

      await sessionRouter.connect(SECOND).withdrawUserStakes(SECOND, 20);

      const userBalAfter = await token.balanceOf(SECOND);
      expect(userBalAfter - userBalBefore).to.eq(stakesOnHold[0]);
      const contractBalAfter = await token.balanceOf(sessionRouter);
      expect(contractBalBefore - contractBalAfter).to.eq(stakesOnHold[0]);
    });
    it('should throw error when withdraw amount is zero', async () => {
      const openedAt = payoutStart + (payoutStart % DAY) + 10 * DAY - 201;

      await setTime(openedAt + 1 * DAY);
      const { msg: msg1, signature: sig1 } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg1, sig1);
      const sessionId1 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 0);

      const { msg: msg2, signature: sig2 } = await getProviderApproval(PROVIDER, SECOND, bidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg2, sig2);
      const sessionId2 = await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 1);

      await setTime(openedAt + 1 * DAY + 500);
      const { msg: receiptMsg1, signature: receiptSig1 } = await getReceipt(PROVIDER, sessionId1, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg1, receiptSig1);

      await setTime(openedAt + 1 * DAY + 550);
      const { msg: receiptMsg2, signature: receiptSig2 } = await getReceipt(PROVIDER, sessionId2, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(receiptMsg2, receiptSig2);

      await expect(sessionRouter.connect(SECOND).withdrawUserStakes(SECOND, 20)).to.be.revertedWithCustomError(
        sessionRouter,
        'SessionUserAmountToWithdrawIsZero',
      );
    });
    it('should throw error when amount of itterations are zero', async () => {
      await expect(sessionRouter.connect(SECOND).withdrawUserStakes(SECOND, 0)).to.be.revertedWithCustomError(
        sessionRouter,
        'SessionUserAmountToWithdrawIsZero',
      );
    });
  });

  describe('#withdrawUserStakes iterations', () => {
    it('should throw error when iterations are zero while entries exist', async () => {
      const { sessionId, openedAt } = await _createSession();
      await setTime(openedAt + 200);
      const { msg, signature } = await getReceipt(PROVIDER, sessionId, 0, 0);
      await sessionRouter.connect(SECOND).closeSession(msg, signature);
      expect((await providerPayouts.getUserStakesOnHold(SECOND, 10))[1]).to.be.greaterThan(0);

      await expect(sessionRouter.connect(SECOND).withdrawUserStakes(SECOND, 0)).to.be.revertedWithCustomError(
        sessionRouter,
        'SessionUserAmountToWithdrawIsZero',
      );
    });
  });

  describe('#model stats', () => {
    it('should keep model averages consistent with two providers closing repeatedly', async () => {
      const [, , , , PROVIDER_2] = await ethers.getSigners();
      await token.transfer(PROVIDER_2, wei(1000));
      await token.connect(PROVIDER_2).approve(diamond, wei(1000));
      await providerRegistry.connect(PROVIDER_2).providerRegister(PROVIDER_2, wei(0.2), 'second');
      await marketplace.connect(PROVIDER_2).postModelBid(PROVIDER_2, modelId, bidPricePerSecond);
      const secondBidId = await marketplace.getBidId(PROVIDER_2, modelId, 0);

      let nonce = 0;
      const openAndClose = async (provider: SignerWithAddress, bid: string, tps: number, ttft: number) => {
        await setTime(Number(await getCurrentBlockTime()) + 10);
        const { msg, signature } = await getProviderApproval(provider, SECOND, bid);
        await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature);
        const sessionId = await providerPayouts.getSessionId(SECOND, provider, bid, nonce++);
        await setTime(Number(await getCurrentBlockTime()) + 100);
        const receipt = await getReceipt(provider, sessionId, tps, ttft);
        await sessionRouter.connect(SECOND).closeSession(receipt.msg, receipt.signature);
      };

      await setTime(payoutStart + 10 * DAY);
      await openAndClose(PROVIDER, bidId, 10, 100);
      await openAndClose(PROVIDER_2, secondBidId, 20, 200);
      await openAndClose(PROVIDER, bidId, 30, 300);

      const first = await sessionRouter.getProviderModelStats(modelId, PROVIDER);
      const second = await sessionRouter.getProviderModelStats(modelId, PROVIDER_2);
      const model = await sessionRouter.getModelStats(modelId);
      expect(first.successCount).to.eq(2);
      expect(first.tpsScaled1000.mean).to.eq(20_000);
      expect(first.ttftMs.mean).to.eq(200);
      expect(second.tpsScaled1000.mean).to.eq(20_000);
      expect(model.count).to.eq(2);
      expect(model.tpsScaled1000.mean).to.eq(20_000);
      expect(model.ttftMs.mean).to.eq(200);
    });
  });

  describe('#stipendToStake', () => {
    it('should return zero if compute balance is zero', async () => {
      expect(await providerPayouts.connect(SECOND).stipendToStake(0, 0)).to.eq(0);
    });
  });

  describe('#contract provider receipts', () => {
    it('should accept an approval signed by the owner of a contract provider', async () => {
      const { provider, contractBidId } = await _createContractProvider();
      const { msg, signature } = await getProviderApproval(OWNER, SECOND, contractBidId);

      await expect(sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature))
        .to.emit(sessionRouter, 'SessionOpened')
        .withArgs(
          SECOND,
          await providerPayouts.getSessionId(SECOND, provider, contractBidId, await sessionRouter.getTotalSessions(provider)),
          provider,
        );
    });

    for (const [name, mode] of [
      ['reverting', 1],
      ['short', 2],
      ['dirty', 3],
    ] as const) {
      it(`should reject an owner-signed approval when owner() has a ${name} return`, async () => {
        const { mock, contractBidId } = await _createContractProvider();
        await mock.setMode(mode, 0);
        const { msg, signature } = await getProviderApproval(OWNER, SECOND, contractBidId);

        await expect(
          sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature),
        ).to.be.revertedWithCustomError(sessionRouter, 'SessionProviderSignatureMismatch');
      });
    }

    it('should accept owner() that burns 50,000 gas', async () => {
      const { mock, contractBidId } = await _createContractProvider();
      await mock.setMode(4, 50_000);
      const { msg, signature } = await getProviderApproval(OWNER, SECOND, contractBidId);

      await expect(sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature)).not.to.be.reverted;
    });

    it('should reject owner() that burns 150,000 gas without exhausting the transaction gas', async () => {
      const { mock, contractBidId } = await _createContractProvider();
      await mock.setMode(4, 150_000);
      const { msg, signature } = await getProviderApproval(OWNER, SECOND, contractBidId);

      await expect(
        sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature, { gasLimit: 2_000_000 }),
      ).to.be.revertedWithCustomError(sessionRouter, 'SessionProviderSignatureMismatch');
    });

    it('should dispute rather than revert when owner() breaks after opening', async () => {
      const { mock, provider, contractBidId } = await _createContractProvider();
      const { msg, signature } = await getProviderApproval(OWNER, SECOND, contractBidId);
      await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), false, msg, signature);
      const sessionId = await providerPayouts.getSessionId(SECOND, provider, contractBidId, 0);

      await mock.setMode(1, 0);
      const session = await sessionRouter.getSession(sessionId);
      await setTime(Number(session.openedAt) + 200);
      const { msg: receiptMsg, signature: receiptSig } = await getReceipt(OWNER, sessionId, 2, 3);

      await expect(sessionRouter.connect(SECOND).closeSession(receiptMsg, receiptSig))
        .to.emit(sessionRouter, 'SessionRating')
        .withArgs(sessionId, modelId, provider, 1, 2_000, 3);
      expect((await sessionRouter.getSession(sessionId)).closeoutType).to.eq(1);
    });
  });

  const _createSession = async (isDirectPaymentFromUser = false) => {
    const secondsToDayEnd = 600n;
    const openedAt = payoutStart + (payoutStart % DAY) + 10 * DAY - Number(secondsToDayEnd) - 1;

    await setTime(openedAt);
    const { msg, signature } = await getProviderApproval(PROVIDER, SECOND, bidId);
    await sessionRouter.connect(SECOND).openSession(SECOND, wei(50), isDirectPaymentFromUser, msg, signature);

    return {
      sessionId: await providerPayouts.getSessionId(SECOND, PROVIDER, bidId, 0),
      secondsToDayEnd: Number(secondsToDayEnd),
      openedAt,
    };
  };

  const _assertDirectClose = async (
    sessionId: string,
    closeAt: number,
    reporter: SignerWithAddress,
  ) => {
    const before = await sessionRouter.getSession(sessionId);
    await setTime(closeAt);
    const { msg, signature } = await getReceipt(reporter, sessionId, 0, 0);
    const providerBalance = await token.balanceOf(PROVIDER);
    const userBalance = await token.balanceOf(SECOND);
    const fundingBalance = await token.balanceOf(FUNDING);
    const diamondBalance = await token.balanceOf(diamond);
    const preCloseBlock = await ethers.provider.getBlockNumber();

    await sessionRouter.connect(SECOND).closeSession(msg, signature);
    const after = await sessionRouter.getSession(sessionId);
    const sessionEnd = after.closedAt < before.endsAt ? after.closedAt : before.endsAt;
    const grossReward = (sessionEnd - before.openedAt) * bidPricePerSecond;
    const startOfEndDay = await providerPayouts.startOfTheDay(sessionEnd);
    const isDisputed = reporter.address !== PROVIDER.address;
    const onHoldReward =
      isDisputed && after.closedAt < before.endsAt && after.closedAt < startOfEndDay + BigInt(DAY)
        ? (after.closedAt - (before.openedAt > startOfEndDay ? before.openedAt : startOfEndDay)) * bidPricePerSecond
        : 0n;
    const providerReward = grossReward - onHoldReward;
    const userStake = before.stake - grossReward;
    const userDuration = sessionEnd - (before.openedAt > startOfEndDay ? before.openedAt : startOfEndDay);
    const lock =
      after.closedAt < startOfEndDay + BigInt(DAY)
        ? await providerPayouts.stipendToStake(userDuration * bidPricePerSecond, startOfEndDay, {
            blockTag: preCloseBlock,
          })
        : 0n;
    const amountToLock = lock < userStake ? lock : userStake;
    const userReturn = userStake - amountToLock;

    expect(after.providerWithdrawnAmount).to.eq(providerReward);
    expect(await token.balanceOf(PROVIDER)).to.eq(providerBalance + providerReward);
    expect(await token.balanceOf(SECOND)).to.eq(userBalance + userReturn);
    expect(await token.balanceOf(FUNDING)).to.eq(fundingBalance);
    expect(await token.balanceOf(diamond)).to.eq(diamondBalance - providerReward - userReturn);
  };

  const _createContractProvider = async () => {
    const factory = await ethers.getContractFactory('ProviderOwnerMock');
    const mock = (await factory.deploy(OWNER)) as ProviderOwnerMock;
    const provider = await mock.getAddress();
    await token.transfer(provider, wei(1));
    await mock.execute(token, token.interface.encodeFunctionData('approve', [await diamond.getAddress(), wei(1)]));
    await mock.execute(
      diamond,
      providerRegistry.interface.encodeFunctionData('providerRegister', [provider, wei(0.2), 'mock']),
    );
    await mock.execute(
      diamond,
      marketplace.interface.encodeFunctionData('postModelBid', [provider, modelId, bidPricePerSecond]),
    );
    const contractBidId = await marketplace.getBidId(provider, modelId, 0);
    await setTime(payoutStart + 10 * DAY);

    return { mock, provider, contractBidId };
  };
});

// npm run generate-types && npx hardhat test "test/diamond/facets/SessionRouter.test.ts"
// npx hardhat coverage --solcoverjs ./.solcover.ts --testfiles "test/diamond/facets/SessionRouter.test.ts"
