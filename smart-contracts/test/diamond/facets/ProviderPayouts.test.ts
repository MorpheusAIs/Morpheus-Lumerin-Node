import {
  DelegateRegistry,
  LumerinDiamond,
  Marketplace,
  ModelRegistry,
  MorpheusToken,
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

describe('ProviderPayouts', () => {
  const reverter = new Reverter();
  const baseModelId = getHex(Buffer.from('provider-payouts-model'));
  const price = wei(0.0001);
  const sessionStake = wei(50);
  const closeDuration = 400n;

  let OWNER: SignerWithAddress;
  let USER: SignerWithAddress;
  let FUNDING: SignerWithAddress;
  let PROVIDER: SignerWithAddress;
  let TARGET: SignerWithAddress;
  let DELEGATE: SignerWithAddress;
  let STRANGER: SignerWithAddress;

  let diamond: LumerinDiamond;
  let token: MorpheusToken;
  let delegateRegistry: DelegateRegistry;
  let providerRegistry: ProviderRegistry;
  let modelRegistry: ModelRegistry;
  let marketplace: Marketplace;
  let sessionRouter: SessionRouter;
  let providerPayouts: ProviderPayouts;
  let bidId: string;
  let modelId: string;
  let nonce: bigint;
  let initializedAt: bigint;

  before(async () => {
    [OWNER, USER, FUNDING, PROVIDER, TARGET, DELEGATE, STRANGER] = await ethers.getSigners();

    [diamond, token, delegateRegistry] = await Promise.all([
      deployLumerinDiamond(),
      deployMORToken(),
      deployDelegateRegistry(),
    ]);

    providerRegistry = await deployFacetProviderRegistry(diamond);
    modelRegistry = await deployFacetModelRegistry(diamond);
    sessionRouter = await deployFacetSessionRouter(diamond, FUNDING);
    marketplace = await deployFacetMarketplace(diamond, token, price, wei(900));
    await deployFacetDelegation(diamond, delegateRegistry);
    providerPayouts = await deployFacetProviderPayouts(diamond);
    initializedAt = await getCurrentBlockTime();
    await deployFacetDelegateStaking(diamond);
    await deployFacetDelegateStakingFunding(diamond);

    await token.transfer(USER, wei(20000));
    await token.transfer(PROVIDER, wei(10000));
    await token.transfer(FUNDING, wei(10000));
    await token.connect(USER).approve(diamond, ethers.MaxUint256);
    await token.connect(PROVIDER).approve(diamond, ethers.MaxUint256);
    await token.connect(FUNDING).approve(diamond, ethers.MaxUint256);

    await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(0.2), 'provider');
    await modelRegistry
      .connect(PROVIDER)
      .modelRegister(PROVIDER, baseModelId, getHex(Buffer.from('ipfs://provider-payouts')), 0, wei(100), 'name', []);
    modelId = await modelRegistry.getModelId(PROVIDER, baseModelId);
    await marketplace.connect(PROVIDER).postModelBid(PROVIDER, modelId, price);
    bidId = await marketplace.getBidId(PROVIDER, modelId, 0);

    await reverter.snapshot();
  });

  beforeEach(() => {
    nonce = 0n;
  });

  afterEach(reverter.revert);

  async function openSession(direct = false, stake = sessionStake) {
    const currentTime = await getCurrentBlockTime();
    const openTime = Math.max(payoutStart + 10 * DAY + Number(nonce) * 10, Number(currentTime) + 1);
    await setTime(openTime);
    const { msg, signature } = await getProviderApproval(PROVIDER, USER, bidId);
    await sessionRouter.connect(USER).openSession(USER, stake, direct, msg, signature);
    const sessionId = await providerPayouts.getSessionId(USER, PROVIDER, bidId, nonce++);
    const session = await sessionRouter.getSession(sessionId);
    return { sessionId, session };
  }

  async function closeSession(sessionId: string, openedAt: bigint, duration = closeDuration) {
    await setTime(Number(openedAt + duration - 1n));
    const { msg, signature } = await getReceipt(PROVIDER, sessionId, 10, 20);
    return sessionRouter.connect(USER).closeSession(msg, signature);
  }

  async function drainFunding() {
    const balance = await token.balanceOf(FUNDING);
    if (balance !== 0n) await token.connect(FUNDING).transfer(OWNER, balance);
  }

  describe('#__ProviderPayouts_init', () => {
    it('should initialize lifetime accounting at the deployment block and expose the payout delegation rule', async () => {
      expect(await providerPayouts.getProviderLifetimeClaimed(PROVIDER)).to.deep.eq([0n, initializedAt]);
      expect(await providerPayouts.DELEGATION_RULES_PAYOUT()).to.eq(
        ethers.keccak256(ethers.toUtf8Bytes('delegation.rules.payout')),
      );
    });

    it('should revert if called twice', async () => {
      await expect(providerPayouts.__ProviderPayouts_init()).to.be.rejectedWith(
        'Initializable: contract is already initialized',
      );
    });
  });

  describe('#setPayoutTarget', () => {
    it('should let the provider set, replace, and reset its payout target', async () => {
      await expect(providerPayouts.connect(PROVIDER).setPayoutTarget(PROVIDER, TARGET))
        .to.emit(providerPayouts, 'ProviderPayoutTargetUpdated')
        .withArgs(PROVIDER, TARGET);
      expect(await providerPayouts.getPayoutTarget(PROVIDER)).to.eq(TARGET);

      await expect(providerPayouts.connect(PROVIDER).setPayoutTarget(PROVIDER, PROVIDER))
        .to.emit(providerPayouts, 'ProviderPayoutTargetUpdated')
        .withArgs(PROVIDER, PROVIDER);
      expect(await providerPayouts.getPayoutTarget(PROVIDER)).to.eq(PROVIDER);

      await providerPayouts.connect(PROVIDER).setPayoutTarget(PROVIDER, TARGET);
      await expect(providerPayouts.connect(PROVIDER).setPayoutTarget(PROVIDER, ethers.ZeroAddress))
        .to.emit(providerPayouts, 'ProviderPayoutTargetUpdated')
        .withArgs(PROVIDER, PROVIDER);
      expect(await providerPayouts.getPayoutTarget(PROVIDER)).to.eq(PROVIDER);
    });

    it('should let contract and global delegates with the explicit payout right set the target', async () => {
      const right = await providerPayouts.DELEGATION_RULES_PAYOUT();
      await delegateRegistry.connect(PROVIDER).delegateContract(DELEGATE, diamond, right, true);
      await providerPayouts.connect(DELEGATE).setPayoutTarget(PROVIDER, TARGET);
      expect(await providerPayouts.getPayoutTarget(PROVIDER)).to.eq(TARGET);

      await delegateRegistry.connect(PROVIDER).delegateContract(DELEGATE, diamond, right, false);
      await delegateRegistry.connect(PROVIDER).delegateAll(DELEGATE, right, true);
      await providerPayouts.connect(DELEGATE).setPayoutTarget(PROVIDER, STRANGER);
      expect(await providerPayouts.getPayoutTarget(PROVIDER)).to.eq(STRANGER);
    });

    it('should deny stale provider/session rights, empty full delegations, and strangers', async () => {
      const cases: Array<() => Promise<unknown>> = [
        async () =>
          delegateRegistry
            .connect(PROVIDER)
            .delegateContract(DELEGATE, diamond, await providerRegistry.DELEGATION_RULES_PROVIDER(), true),
        async () =>
          delegateRegistry
            .connect(PROVIDER)
            .delegateContract(DELEGATE, diamond, await providerRegistry.DELEGATION_RULES_SESSION(), true),
        async () => delegateRegistry.connect(PROVIDER).delegateContract(DELEGATE, diamond, ethers.ZeroHash, true),
        async () => delegateRegistry.connect(PROVIDER).delegateAll(DELEGATE, ethers.ZeroHash, true),
      ];

      for (const grant of cases) {
        await grant();
        await expect(providerPayouts.connect(DELEGATE).setPayoutTarget(PROVIDER, TARGET))
          .to.be.revertedWithCustomError(providerPayouts, 'InsufficientRightsForOperation')
          .withArgs(PROVIDER, DELEGATE);
      }
      await expect(providerPayouts.connect(STRANGER).setPayoutTarget(PROVIDER, TARGET))
        .to.be.revertedWithCustomError(providerPayouts, 'InsufficientRightsForOperation')
        .withArgs(PROVIDER, STRANGER);
    });

    it('should reject the diamond itself as a payout target', async () => {
      await expect(providerPayouts.connect(PROVIDER).setPayoutTarget(PROVIDER, diamond)).to.be.revertedWithCustomError(
        providerPayouts,
        'ProviderPayoutTargetInvalid',
      );
    });

    it('should keep an explicit payout right effective when a full delegation also exists', async () => {
      const right = await providerPayouts.DELEGATION_RULES_PAYOUT();
      await delegateRegistry.connect(PROVIDER).delegateContract(DELEGATE, diamond, right, true);
      await delegateRegistry.connect(PROVIDER).delegateAll(DELEGATE, ethers.ZeroHash, true);

      await providerPayouts.connect(DELEGATE).setPayoutTarget(PROVIDER, TARGET);
      expect(await providerPayouts.getPayoutTarget(PROVIDER)).to.eq(TARGET);
    });

    it('should ignore payout rights given for another contract, to another delegate, or revoked', async () => {
      const right = await providerPayouts.DELEGATION_RULES_PAYOUT();
      await delegateRegistry.connect(PROVIDER).delegateContract(DELEGATE, token, right, true);
      await delegateRegistry.connect(PROVIDER).delegateContract(STRANGER, diamond, right, true);
      await delegateRegistry.connect(PROVIDER).delegateContract(DELEGATE, diamond, right, true);
      await delegateRegistry.connect(PROVIDER).delegateContract(DELEGATE, diamond, right, false);

      await expect(providerPayouts.connect(DELEGATE).setPayoutTarget(PROVIDER, TARGET))
        .to.be.revertedWithCustomError(providerPayouts, 'InsufficientRightsForOperation')
        .withArgs(PROVIDER, DELEGATE);
    });
  });

  describe('payout routing', () => {
    it('should route immediate close payment to the configured target and default to the provider when unset', async () => {
      await providerPayouts.connect(PROVIDER).setPayoutTarget(PROVIDER, TARGET);
      const first = await openSession();
      const reward = price * closeDuration;

      await expect(closeSession(first.sessionId, first.session.openedAt)).to.changeTokenBalances(
        token,
        [FUNDING, TARGET, PROVIDER],
        [-reward, reward, 0n],
      );

      await providerPayouts.connect(PROVIDER).setPayoutTarget(PROVIDER, ethers.ZeroAddress);
      const second = await openSession();
      await expect(closeSession(second.sessionId, second.session.openedAt)).to.changeTokenBalances(
        token,
        [FUNDING, PROVIDER, TARGET],
        [-reward, reward, 0n],
      );
    });

    it('should route a direct-pay session from the diamond to the target', async () => {
      await providerPayouts.connect(PROVIDER).setPayoutTarget(PROVIDER, TARGET);
      const { sessionId, session } = await openSession(true);
      const reward = price * closeDuration;
      const startOfDay = await providerPayouts.startOfTheDay(session.openedAt + closeDuration);
      const amountToLock = await providerPayouts.stipendToStake(reward, startOfDay);

      await expect(closeSession(sessionId, session.openedAt)).to.changeTokenBalances(
        token,
        [diamond, TARGET, FUNDING],
        [-(sessionStake - amountToLock), reward, 0n],
      );
      expect(await providerPayouts.getProviderLifetimeClaimed(PROVIDER)).to.deep.eq([reward, initializedAt]);
    });

    it('should allow a stranger to claim and route payment to the target', async () => {
      await providerPayouts.connect(PROVIDER).setPayoutTarget(PROVIDER, TARGET);
      const { sessionId, session } = await openSession();
      await setTime(Number(session.endsAt));
      const reward = price * (session.endsAt - session.openedAt);

      await expect(sessionRouter.connect(STRANGER).claimForProvider(sessionId)).to.changeTokenBalances(
        token,
        [FUNDING, TARGET, STRANGER],
        [-reward, reward, 0n],
      );
    });

    it('should pay old debt to the current target and pay the provider after clearing the target', async () => {
      await drainFunding();
      const first = await openSession();
      await closeSession(first.sessionId, first.session.openedAt);
      const reward = price * closeDuration;

      await providerPayouts.connect(PROVIDER).setPayoutTarget(PROVIDER, TARGET);
      await token.transfer(FUNDING, reward);
      await expect(sessionRouter.connect(STRANGER).claimForProvider(first.sessionId)).to.changeTokenBalances(
        token,
        [FUNDING, TARGET, PROVIDER],
        [-reward, reward, 0n],
      );

      const second = await openSession();
      await closeSession(second.sessionId, second.session.openedAt);
      await providerPayouts.connect(PROVIDER).setPayoutTarget(PROVIDER, ethers.ZeroAddress);
      await token.transfer(FUNDING, reward);
      await expect(sessionRouter.connect(STRANGER).claimForProvider(second.sessionId)).to.changeTokenBalances(
        token,
        [FUNDING, PROVIDER, TARGET],
        [-reward, reward, 0n],
      );
    });
  });

  describe('provider debt', () => {
    it('should close on a funding balance shortage without changing exact user settlement', async () => {
      const { sessionId, session } = await openSession();
      const startOfDay = await providerPayouts.startOfTheDay(session.openedAt + closeDuration);
      const amountToLock = await providerPayouts.stipendToStake(price * closeDuration, startOfDay);
      const userBefore = await token.balanceOf(USER);
      await drainFunding();
      const claimedBefore = await sessionRouter.getProvidersTotalClaimed();

      await expect(closeSession(sessionId, session.openedAt))
        .to.emit(providerPayouts, 'ProviderRewardAccrued')
        .withArgs(sessionId, PROVIDER, price * closeDuration, 0, price * closeDuration);

      expect((await token.balanceOf(USER)) - userBefore).to.eq(sessionStake - amountToLock);
      expect(await providerPayouts.getUserStakesOnHold(USER, 0)).to.deep.eq([0n, amountToLock]);
      expect(await providerPayouts.getUserStakesOnHold(USER, 255)).to.deep.eq([0n, amountToLock]);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(price * closeDuration);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(price * closeDuration);
      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(claimedBefore);
      expect(await providerPayouts.getProviderLifetimeClaimed(PROVIDER)).to.deep.eq([0n, initializedAt]);
      expect((await providerRegistry.getProvider(PROVIDER)).limitPeriodEarned).to.eq(price * closeDuration);
      expect((await sessionRouter.getSession(sessionId)).providerWithdrawnAmount).to.eq(price * closeDuration);

      const releaseAt = startOfDay + BigInt(DAY);
      await setTime(Number(releaseAt - 1n));
      expect(await providerPayouts.getUserStakesOnHold(USER, 1)).to.deep.eq([0n, amountToLock]);
      await setTime(Number(releaseAt));
      expect(await providerPayouts.getUserStakesOnHold(USER, 1)).to.deep.eq([amountToLock, 0n]);
    });

    it('should record full debt on an allowance shortage and pay when balance and allowance are exactly sufficient', async () => {
      const first = await openSession();
      const reward = price * closeDuration;
      await token.connect(FUNDING).approve(diamond, reward - 1n);
      await closeSession(first.sessionId, first.session.openedAt);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward);

      await token.connect(FUNDING).approve(diamond, reward);
      await expect(sessionRouter.claimForProvider(first.sessionId))
        .to.emit(providerPayouts, 'ProviderDebtPaid')
        .withArgs(PROVIDER, PROVIDER, reward);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(0);

      const second = await openSession();
      await drainFunding();
      await token.transfer(FUNDING, reward);
      await token.connect(FUNDING).approve(diamond, reward);
      await expect(closeSession(second.sessionId, second.session.openedAt))
        .to.emit(providerPayouts, 'ProviderRewardAccrued')
        .withArgs(second.sessionId, PROVIDER, reward, reward, 0);
    });

    it('should not accrue twice and should convert the full debt into paid accounting after refunding', async () => {
      await drainFunding();
      const { sessionId, session } = await openSession();
      const reward = price * closeDuration;
      await closeSession(sessionId, session.openedAt);
      const withdrawn = (await sessionRouter.getSession(sessionId)).providerWithdrawnAmount;
      const earned = (await providerRegistry.getProvider(PROVIDER)).limitPeriodEarned;

      await sessionRouter.connect(STRANGER).claimForProvider(sessionId);
      expect((await sessionRouter.getSession(sessionId)).providerWithdrawnAmount).to.eq(withdrawn);
      expect((await providerRegistry.getProvider(PROVIDER)).limitPeriodEarned).to.eq(earned);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward);

      await token.transfer(FUNDING, reward);
      await expect(sessionRouter.connect(STRANGER).claimForProvider(sessionId))
        .to.emit(providerPayouts, 'ProviderDebtPaid')
        .withArgs(PROVIDER, PROVIDER, reward);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(0);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(0);
      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(reward);
      expect(await providerPayouts.getProviderLifetimeClaimed(PROVIDER)).to.deep.eq([reward, initializedAt]);
      await expect(sessionRouter.connect(STRANGER).claimForProvider(sessionId)).to.changeTokenBalances(
        token,
        [FUNDING, PROVIDER],
        [0n, 0n],
      );
    });

    it('should make balance-limited and allowance-limited partial debt payments and no-op at zero limits', async () => {
      await drainFunding();
      const { sessionId, session } = await openSession();
      const reward = price * closeDuration;
      await closeSession(sessionId, session.openedAt);

      await sessionRouter.claimForProvider(sessionId);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward);

      const balancePart = reward / 4n;
      await token.transfer(FUNDING, balancePart);
      await expect(sessionRouter.claimForProvider(sessionId))
        .to.emit(providerPayouts, 'ProviderDebtPaid')
        .withArgs(PROVIDER, PROVIDER, balancePart);

      const allowancePart = reward / 5n;
      await token.transfer(FUNDING, reward);
      await token.connect(FUNDING).approve(diamond, 0);
      await sessionRouter.claimForProvider(sessionId);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward - balancePart);
      await token.connect(FUNDING).approve(diamond, allowancePart);
      await expect(sessionRouter.claimForProvider(sessionId))
        .to.emit(providerPayouts, 'ProviderDebtPaid')
        .withArgs(PROVIDER, PROVIDER, allowancePart);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward - balancePart - allowancePart);
    });

    it('should pay all provider debt from two sessions when claiming either session', async () => {
      await drainFunding();
      const first = await openSession();
      await closeSession(first.sessionId, first.session.openedAt);
      const second = await openSession();
      await closeSession(second.sessionId, second.session.openedAt);
      const total = price * closeDuration * 2n;
      await token.transfer(FUNDING, total);

      await expect(sessionRouter.connect(STRANGER).claimForProvider(first.sessionId))
        .to.emit(providerPayouts, 'ProviderDebtPaid')
        .withArgs(PROVIDER, PROVIDER, total);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(0);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(0);
    });

    it('should not pay one provider debt when another provider session is claimed', async () => {
      await drainFunding();
      const first = await openSession();
      await closeSession(first.sessionId, first.session.openedAt);
      const owed = price * closeDuration;

      await token.transfer(TARGET, wei(1000));
      await token.connect(TARGET).approve(diamond, ethers.MaxUint256);
      await providerRegistry.connect(TARGET).providerRegister(TARGET, wei(0.2), 'second-provider');
      const secondBaseModelId = getHex(Buffer.from('provider-payouts-second-model'));
      await modelRegistry
        .connect(TARGET)
        .modelRegister(TARGET, secondBaseModelId, getHex(Buffer.from('ipfs://second')), 0, wei(100), 'second', []);
      const secondModelId = await modelRegistry.getModelId(TARGET, secondBaseModelId);
      await marketplace.connect(TARGET).postModelBid(TARGET, secondModelId, price);
      const secondBidId = await marketplace.getBidId(TARGET, secondModelId, 0);

      await setTime(Number(await getCurrentBlockTime()) + 1);
      const { msg, signature } = await getProviderApproval(TARGET, USER, secondBidId);
      await sessionRouter.connect(USER).openSession(USER, sessionStake, false, msg, signature);
      const secondSessionId = await providerPayouts.getSessionId(USER, TARGET, secondBidId, 1);
      await sessionRouter.connect(STRANGER).claimForProvider(secondSessionId);

      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(owed);
      expect(await providerPayouts.getProviderOwed(TARGET)).to.eq(0);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(owed);
    });

    it('should reconcile payout events, provider gain, and outstanding debt', async () => {
      await providerPayouts.connect(PROVIDER).setPayoutTarget(PROVIDER, TARGET);
      const targetBefore = await token.balanceOf(TARGET);
      const immediate = await openSession();
      const reward = price * closeDuration;
      await expect(closeSession(immediate.sessionId, immediate.session.openedAt))
        .to.emit(providerPayouts, 'ProviderRewardAccrued')
        .withArgs(immediate.sessionId, PROVIDER, reward, reward, 0);

      await drainFunding();
      const deferred = await openSession();
      await expect(closeSession(deferred.sessionId, deferred.session.openedAt))
        .to.emit(providerPayouts, 'ProviderRewardAccrued')
        .withArgs(deferred.sessionId, PROVIDER, reward, 0, reward);
      const paidDebt = reward / 2n;
      await token.transfer(FUNDING, paidDebt);
      await expect(sessionRouter.claimForProvider(deferred.sessionId))
        .to.emit(providerPayouts, 'ProviderDebtPaid')
        .withArgs(PROVIDER, TARGET, paidDebt);

      expect((await token.balanceOf(TARGET)) - targetBefore).to.eq(reward + paidDebt);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward - paidDebt);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(reward - paidDebt);
    });
  });

  describe('debt priority', () => {
    const reward = price * closeDuration;

    async function setupSecondProvider() {
      await token.transfer(TARGET, wei(1000));
      await token.connect(TARGET).approve(diamond, ethers.MaxUint256);
      await providerRegistry.connect(TARGET).providerRegister(TARGET, wei(0.2), 'second-provider');
      const secondBaseModelId = getHex(Buffer.from('provider-payouts-priority-model'));
      await modelRegistry
        .connect(TARGET)
        .modelRegister(TARGET, secondBaseModelId, getHex(Buffer.from('ipfs://priority')), 0, wei(100), 'second', []);
      const secondModelId = await modelRegistry.getModelId(TARGET, secondBaseModelId);
      await marketplace.connect(TARGET).postModelBid(TARGET, secondModelId, price);
      return marketplace.getBidId(TARGET, secondModelId, 0);
    }

    async function openSecondProviderSession(secondBidId: string) {
      await setTime(Number(await getCurrentBlockTime()) + 1);
      const { msg, signature } = await getProviderApproval(TARGET, USER, secondBidId);
      await sessionRouter.connect(USER).openSession(USER, sessionStake, false, msg, signature);
      const sessionId = await providerPayouts.getSessionId(USER, TARGET, secondBidId, nonce++);
      return { sessionId, session: await sessionRouter.getSession(sessionId) };
    }

    async function closeSecondProviderSession(sessionId: string, openedAt: bigint) {
      await setTime(Number(openedAt + closeDuration - 1n));
      const { msg, signature } = await getReceipt(TARGET, sessionId, 10, 20);
      return sessionRouter.connect(USER).closeSession(msg, signature);
    }

    it('should turn a new reward into debt while funding only covers outstanding debts of another provider', async () => {
      const secondBidId = await setupSecondProvider();
      await drainFunding();
      const first = await openSession();
      await closeSession(first.sessionId, first.session.openedAt);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward);

      await token.transfer(FUNDING, reward);
      const second = await openSecondProviderSession(secondBidId);
      await expect(closeSecondProviderSession(second.sessionId, second.session.openedAt))
        .to.emit(providerPayouts, 'ProviderRewardAccrued')
        .withArgs(second.sessionId, TARGET, reward, 0, reward);
      expect(await token.balanceOf(FUNDING)).to.eq(reward);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(reward * 2n);

      await expect(sessionRouter.connect(STRANGER).claimForProvider(first.sessionId)).to.changeTokenBalances(
        token,
        [FUNDING, PROVIDER],
        [-reward, reward],
      );
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(0);
      expect(await providerPayouts.getProviderOwed(TARGET)).to.eq(reward);
    });

    it('should pay a new reward immediately once funding covers outstanding debts plus the reward', async () => {
      await drainFunding();
      const first = await openSession();
      await closeSession(first.sessionId, first.session.openedAt);

      await token.transfer(FUNDING, reward * 2n);
      const second = await openSession();
      await expect(closeSession(second.sessionId, second.session.openedAt))
        .to.emit(providerPayouts, 'ProviderRewardAccrued')
        .withArgs(second.sessionId, PROVIDER, reward, reward, 0);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward);
      expect(await token.balanceOf(FUNDING)).to.eq(reward);
    });

    it('should reserve allowance for outstanding debts as well as balance', async () => {
      await drainFunding();
      const first = await openSession();
      await closeSession(first.sessionId, first.session.openedAt);

      await token.transfer(FUNDING, reward * 10n);
      await token.connect(FUNDING).approve(diamond, reward * 2n - 1n);
      const second = await openSession();
      await expect(closeSession(second.sessionId, second.session.openedAt))
        .to.emit(providerPayouts, 'ProviderRewardAccrued')
        .withArgs(second.sessionId, PROVIDER, reward, 0, reward);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward * 2n);
    });

    it('should pay debts best-effort in claim order when funding cannot cover all of them', async () => {
      const secondBidId = await setupSecondProvider();
      await drainFunding();
      const first = await openSession();
      await closeSession(first.sessionId, first.session.openedAt);
      const second = await openSecondProviderSession(secondBidId);
      await closeSecondProviderSession(second.sessionId, second.session.openedAt);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(reward * 2n);

      await token.transfer(FUNDING, reward + reward / 2n);
      await expect(sessionRouter.connect(STRANGER).claimForProvider(second.sessionId))
        .to.emit(providerPayouts, 'ProviderDebtPaid')
        .withArgs(TARGET, TARGET, reward);
      await expect(sessionRouter.connect(STRANGER).claimForProvider(first.sessionId))
        .to.emit(providerPayouts, 'ProviderDebtPaid')
        .withArgs(PROVIDER, PROVIDER, reward / 2n);

      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward / 2n);
      expect(await providerPayouts.getProviderOwed(TARGET)).to.eq(0);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(reward / 2n);
      expect(await token.balanceOf(FUNDING)).to.eq(0);
    });
  });

  describe('#claimForProvider for an unknown session', () => {
    it('should revert and leave limiter, debt, and payout accounting untouched', async () => {
      const unknownId = getHex(Buffer.from('unknown-session'), 32);
      const zeroProviderBefore = await providerRegistry.getProvider(ethers.ZeroAddress);

      await expect(sessionRouter.connect(STRANGER).claimForProvider(unknownId)).to.be.revertedWithCustomError(
        sessionRouter,
        'SessionNotEndedOrNotExist',
      );

      expect(await providerRegistry.getProvider(ethers.ZeroAddress)).to.deep.eq(zeroProviderBefore);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(0);
      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(0);
    });
  });

  describe('accrual edge cases', () => {
    it('should count an immediate stake-pool payment in lifetimeClaimed', async () => {
      const { sessionId, session } = await openSession();
      const reward = price * closeDuration;

      await closeSession(sessionId, session.openedAt);

      expect(await providerPayouts.getProviderLifetimeClaimed(PROVIDER)).to.deep.eq([reward, initializedAt]);
      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(reward);
    });

    it('should emit nothing and move nothing when there is nothing to accrue or pay', async () => {
      const { sessionId, session } = await openSession();
      await closeSession(sessionId, session.openedAt);

      const tx = sessionRouter.connect(STRANGER).claimForProvider(sessionId);
      await expect(tx).to.not.emit(providerPayouts, 'ProviderRewardAccrued');
      await expect(tx).to.not.emit(providerPayouts, 'ProviderDebtPaid');
      await expect(tx).to.changeTokenBalances(token, [FUNDING, PROVIDER, STRANGER], [0n, 0n, 0n]);
    });

    it('should not emit or record debt when the reward limiter is exhausted', async () => {
      await drainFunding();
      const first = await openSession(false, wei(10000));
      await closeSession(first.sessionId, first.session.openedAt, 3000n);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(wei(0.2));

      const second = await openSession();
      await setTime(Number(second.session.openedAt + closeDuration - 1n));
      const { msg, signature } = await getReceipt(PROVIDER, second.sessionId, 10, 20);
      await expect(sessionRouter.connect(USER).closeSession(msg, signature)).to.not.emit(
        providerPayouts,
        'ProviderRewardAccrued',
      );

      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(wei(0.2));
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(wei(0.2));
      expect((await sessionRouter.getSession(second.sessionId)).providerWithdrawnAmount).to.eq(0);
    });

    it('should combine debt with the dispute hold and pay everything once the hold ends and funding is refilled', async () => {
      // open shortly before midnight, close with a dispute shortly after it
      const dayStart = BigInt(payoutStart - (payoutStart % DAY) + 11 * DAY);
      await setTime(Number(dayStart) - 300);
      const { msg, signature } = await getProviderApproval(PROVIDER, USER, bidId);
      await sessionRouter.connect(USER).openSession(USER, sessionStake, false, msg, signature);
      const sessionId = await providerPayouts.getSessionId(USER, PROVIDER, bidId, 0);
      const session = await sessionRouter.getSession(sessionId);

      await drainFunding();
      await setTime(Number(dayStart) + 99);
      const { msg: receipt } = await getReceipt(PROVIDER, sessionId, 10, 20);
      const { signature: disputedSig } = await getReceipt(STRANGER, sessionId, 10, 20);
      await sessionRouter.connect(USER).closeSession(receipt, disputedSig);

      const closedAt = dayStart + 100n;
      const reward = (closedAt - session.openedAt) * price;
      const hold = (closedAt - dayStart) * price;
      expect((await sessionRouter.getSession(sessionId)).closeoutType).to.eq(1);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward - hold);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(reward - hold);
      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(0);

      // hold still active: nothing new accrues, nothing is paid without funding
      await sessionRouter.connect(STRANGER).claimForProvider(sessionId);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward - hold);

      await setTime(Number(dayStart) + DAY);
      await sessionRouter.connect(STRANGER).claimForProvider(sessionId);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(reward);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(reward);
      expect((await sessionRouter.getSession(sessionId)).providerWithdrawnAmount).to.eq(reward);

      await token.transfer(FUNDING, reward);
      await expect(sessionRouter.connect(STRANGER).claimForProvider(sessionId)).to.changeTokenBalances(
        token,
        [FUNDING, PROVIDER, STRANGER],
        [-reward, reward, 0n],
      );
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(0);
      expect(await providerPayouts.getPendingProviderClaims()).to.eq(0);
      expect(await sessionRouter.getProvidersTotalClaimed()).to.eq(reward);
    });
  });

  describe('emission accounting', () => {
    it('should debit claimed plus pending values and keep all emission views unchanged when debt is paid', async () => {
      const timestamp = BigInt(payoutStart + 10 * DAY);
      const computeBefore = await providerPayouts.getComputeBalance(timestamp);
      const supplyBefore = await providerPayouts.totalMORSupply(timestamp);
      const stake = wei(10);
      const stipendBefore = await providerPayouts.stakeToStipend(stake, timestamp);
      await drainFunding();
      const { sessionId, session } = await openSession();
      const reward = price * closeDuration;
      await closeSession(sessionId, session.openedAt);

      expect(await providerPayouts.getComputeBalance(timestamp)).to.eq(computeBefore - reward);
      expect(await providerPayouts.totalMORSupply(timestamp)).to.eq(supplyBefore + reward);
      expect(await providerPayouts.getTodaysBudget(timestamp)).to.eq((computeBefore - reward) / 100n);
      expect(await providerPayouts.stakeToStipend(stake, timestamp)).to.eq(
        (stake * (computeBefore - reward)) / ((supplyBefore + reward) * 100n),
      );
      const stipend = wei(1);
      expect(await providerPayouts.stipendToStake(stipend, timestamp)).to.eq(
        (stipend * (supplyBefore + reward) * 100n) / (computeBefore - reward),
      );
      expect(stipendBefore).to.eq((stake * computeBefore) / (supplyBefore * 100n));

      await token.transfer(FUNDING, reward);
      await sessionRouter.claimForProvider(sessionId);
      expect(await providerPayouts.getComputeBalance(timestamp)).to.eq(computeBefore - reward);
      expect(await providerPayouts.totalMORSupply(timestamp)).to.eq(supplyBefore + reward);
      expect(await providerPayouts.stakeToStipend(stake, timestamp)).to.eq(
        (stake * (computeBefore - reward)) / ((supplyBefore + reward) * 100n),
      );
    });

    it('should saturate a depleted compute pool and still close and claim an already-open session', async () => {
      const { sessionId, session } = await openSession();
      const day = await providerPayouts.startOfTheDay(session.openedAt);
      await sessionRouter.setPoolConfig(3, {
        payoutStart: day,
        decreaseInterval: DAY,
        initialReward: 1,
        rewardDecrease: 0,
      });
      await drainFunding();
      await closeSession(sessionId, session.openedAt);

      expect(await providerPayouts.getComputeBalance(session.openedAt)).to.eq(0);
      expect(await providerPayouts.getTodaysBudget(session.openedAt)).to.eq(0);
      expect(await providerPayouts.stipendToStake(wei(1), session.openedAt)).to.eq(0);
      await expect(sessionRouter.connect(STRANGER).claimForProvider(sessionId)).not.to.be.reverted;

      await setTime(Number(session.openedAt + closeDuration + 10n));
      const { msg, signature } = await getProviderApproval(PROVIDER, USER, bidId);
      await expect(sessionRouter.connect(USER).openSession(USER, sessionStake, false, msg, signature))
        .to.be.revertedWithCustomError(sessionRouter, 'SessionTooShort');
    });
  });

  describe('reward limiter', () => {
    it('should cap debt by provider stake and leave the remainder claimable after the limiter period resets', async () => {
      await drainFunding();
      const { sessionId, session } = await openSession(false, wei(10000));
      const longDuration = 3000n;
      await closeSession(sessionId, session.openedAt, longDuration);

      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(wei(0.2));
      expect((await providerRegistry.getProvider(PROVIDER)).limitPeriodEarned).to.eq(wei(0.2));
      expect((await sessionRouter.getSession(sessionId)).providerWithdrawnAmount).to.eq(wei(0.2));

      const limitEnd = (await providerRegistry.getProvider(PROVIDER)).limitPeriodEnd;
      await setTime(Number(limitEnd + 1n));
      await sessionRouter.connect(STRANGER).claimForProvider(sessionId);
      expect(await providerPayouts.getProviderOwed(PROVIDER)).to.eq(price * longDuration);
      expect((await providerRegistry.getProvider(PROVIDER)).limitPeriodEarned).to.eq(wei(0.1));
      expect((await sessionRouter.getSession(sessionId)).providerWithdrawnAmount).to.eq(price * longDuration);
    });
  });

  describe('views', () => {
    it('should return deterministic IDs, capped ends, day starts, and ignore hold-view iterations', async () => {
      const timestamp = BigInt(payoutStart + 10 * DAY + 12345);
      expect(await providerPayouts.startOfTheDay(timestamp)).to.eq(timestamp - (timestamp % BigInt(DAY)));
      expect(await providerPayouts.getSessionId(USER, PROVIDER, bidId, 7)).to.eq(
        ethers.solidityPackedKeccak256(
          ['address', 'address', 'bytes32', 'uint256'],
          [USER.address, PROVIDER.address, bidId, 7],
        ),
      );
      expect(await providerPayouts.getSessionEnd(wei(10000), price, timestamp)).to.eq(timestamp + BigInt(DAY));

      await drainFunding();
      const { sessionId, session } = await openSession();
      await closeSession(sessionId, session.openedAt);
      expect(await providerPayouts.getUserStakesOnHold(USER, 0)).to.deep.eq(
        await providerPayouts.getUserStakesOnHold(USER, 255),
      );
      const releaseAt = (await providerPayouts.startOfTheDay(session.openedAt + closeDuration)) + BigInt(DAY);
      await setTime(Number(releaseAt));
      expect(await providerPayouts.getUserStakesOnHold(USER, 0)).to.deep.eq(
        await providerPayouts.getUserStakesOnHold(USER, 255),
      );
      expect((await providerPayouts.getUserStakesOnHold(USER, 0))[1]).to.eq(0);
    });
  });
});
