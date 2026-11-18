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
} from '@ethers-v6';
import { SignerWithAddress } from '@nomicfoundation/hardhat-ethers/signers';
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
import { getCurrentBlockTime, setTime } from '@/utils/block-helper';
import { getProviderApproval, getReceipt } from '@/utils/provider-helper';
import { DAY } from '@/utils/time';

describe('DelegateStakingFunding', () => {
  const reverter = new Reverter();
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let initialSnapshot: any;
  const price = wei(0.0001);
  const baseModelId = getHex(Buffer.from('delegate-staking-funding'));

  let OWNER: SignerWithAddress;
  let HOT: SignerWithAddress;
  let HOT_TWO: SignerWithAddress;
  let DELEGATE: SignerWithAddress;
  let STRANGER: SignerWithAddress;
  let PROVIDER: SignerWithAddress;
  let FUNDING: SignerWithAddress;
  let funders: SignerWithAddress[];

  let diamond: LumerinDiamond;
  let token: MorpheusToken;
  let delegateRegistry: DelegateRegistry;
  let providerRegistry: ProviderRegistry;
  let modelRegistry: ModelRegistry;
  let marketplace: Marketplace;
  let providerPayouts: ProviderPayouts;
  let delegateStaking: DelegateStaking;
  let stakingFunding: DelegateStakingFunding;
  let bidId: string;

  before(async () => {
    // this suite moves the chain time in before(); restore it for the suites that run next
    initialSnapshot = await network.provider.send('evm_snapshot', []);

    [OWNER, HOT, HOT_TWO, DELEGATE, STRANGER, PROVIDER, FUNDING, ...funders] = await ethers.getSigners();
    [diamond, token, delegateRegistry] = await Promise.all([
      deployLumerinDiamond(),
      deployMORToken(),
      deployDelegateRegistry(),
    ]);

    providerRegistry = await deployFacetProviderRegistry(diamond);
    modelRegistry = await deployFacetModelRegistry(diamond);
    await deployFacetSessionRouter(diamond, FUNDING);
    marketplace = await deployFacetMarketplace(diamond, token, price, wei(900));
    await deployFacetDelegation(diamond, delegateRegistry);
    providerPayouts = await deployFacetProviderPayouts(diamond);
    delegateStaking = await deployFacetDelegateStaking(diamond);
    stakingFunding = await deployFacetDelegateStakingFunding(diamond);

    for (const account of [HOT, HOT_TWO, PROVIDER, FUNDING, ...funders]) {
      await token.transfer(account, wei(1000000));
      await token.connect(account).approve(diamond, ethers.MaxUint256);
    }
    await token.connect(PROVIDER).approve(providerRegistry, ethers.MaxUint256);
    await token.connect(PROVIDER).approve(modelRegistry, ethers.MaxUint256);
    await token.connect(PROVIDER).approve(marketplace, ethers.MaxUint256);

    await providerRegistry.connect(PROVIDER).providerRegister(PROVIDER, wei(0.2), 'provider');
    await modelRegistry
      .connect(PROVIDER)
      .modelRegister(PROVIDER, baseModelId, getHex(Buffer.from('ipfs://delegate-staking')), 0, wei(100), 'model', []);
    const modelId = await modelRegistry.getModelId(PROVIDER, baseModelId);
    await marketplace.connect(PROVIDER).postModelBid(PROVIDER, modelId, price);
    bidId = await marketplace.getBidId(PROVIDER, modelId, 0);

    // compute emission is zero before payoutStart, sessions need a positive stipend
    await setTime(payoutStart + 10 * DAY);

    await reverter.snapshot();
  });

  after(async () => {
    await network.provider.send('evm_revert', [initialSnapshot]);
  });

  afterEach(reverter.revert);

  async function expiry(days = 30): Promise<bigint> {
    return (await getCurrentBlockTime()) + BigInt(days * DAY);
  }

  async function approveGrantFund(
    funder: SignerWithAddress,
    amount: bigint,
    hot = HOT,
    cap = amount,
    expires?: bigint,
  ) {
    if (funder.address !== hot.address) {
      await stakingFunding.connect(hot).approvePoolFunder(hot, funder);
    }
    await stakingFunding.connect(funder).grantStakingAllowance(hot, cap, expires ?? (await expiry()));
    await stakingFunding.connect(funder).fundStakingAllowance(hot, amount);
  }

  async function openPoolSession(amount: bigint, hot = HOT) {
    const { msg, signature } = await getProviderApproval(PROVIDER, hot, bidId);
    let nonce = 0;
    let sessionId = await providerPayouts.getSessionId(hot, PROVIDER, bidId, nonce);
    while ((await stakingFunding.getPoolSession(sessionId)).exists) {
      sessionId = await providerPayouts.getSessionId(hot, PROVIDER, bidId, ++nonce);
    }
    await delegateStaking.connect(hot).openSessionFromPool(hot, amount, msg, signature);
    return sessionId;
  }

  async function closePoolSession(sessionId: string, elapsed = 100) {
    const opened = await providerPayouts.getSession(sessionId);
    const now = Number(await getCurrentBlockTime());
    await setTime(Math.max(Number(opened.openedAt) + elapsed, now + 1));
    const { msg, signature } = await getReceipt(PROVIDER, sessionId, 10, 20);
    await delegateStaking.connect(HOT).closePoolSession(msg, signature);
  }

  async function expectAccountingInvariant(hot: SignerWithAddress, expectedDiamondDelta?: bigint, before?: bigint) {
    const [page, total] = await stakingFunding.getPoolFunders(hot, 0, 128);
    expect(BigInt(page.length)).to.eq(total);
    let principals = 0n;
    let pending = 0n;
    let exposure = 0n;
    let lockedExposure = 0n;
    for (const entry of page) {
      principals += entry.allowance.currentPrincipal;
      pending += entry.allowance.pending;
      exposure += entry.allowance.exposure;
      lockedExposure += entry.allowance.lockedExposure;
    }
    const balances = await stakingFunding.getPoolBalances(hot);
    expect(principals + pending).to.eq(balances.free + balances.inSession + balances.locked);
    expect(exposure).to.eq(balances.inSession);
    expect(lockedExposure).to.be.lessThanOrEqual(balances.locked);
    expect(pending).to.eq(balances.pendingTotal);
    if (expectedDiamondDelta !== undefined && before !== undefined) {
      expect((await token.balanceOf(diamond)) - before).to.eq(expectedDiamondDelta);
    }
  }

  describe('#approvePoolFunder / #revokePoolFunderApproval', () => {
    it('should allow the hot wallet and its explicitly-authorized delegate, but reject a stranger and zero funder', async () => {
      const right = await stakingFunding.getDelegateStakingDelegationRight();
      await expect(stakingFunding.connect(HOT).approvePoolFunder(HOT, funders[0]))
        .to.emit(stakingFunding, 'DelegateStakingPoolFunderApproved')
        .withArgs(HOT, funders[0]);
      await delegateRegistry.connect(HOT).delegateContract(DELEGATE, diamond, right, true);
      await expect(stakingFunding.connect(DELEGATE).approvePoolFunder(HOT, funders[1]))
        .to.emit(stakingFunding, 'DelegateStakingPoolFunderApproved')
        .withArgs(HOT, funders[1]);
      await expect(stakingFunding.connect(STRANGER).approvePoolFunder(HOT, funders[2])).to.be.revertedWithCustomError(
        stakingFunding,
        'InsufficientRightsForOperation',
      );
      await expect(stakingFunding.connect(HOT).approvePoolFunder(HOT, ethers.ZeroAddress)).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingInvalidAddress',
      );
      await expect(stakingFunding.connect(HOT).revokePoolFunderApproval(HOT, ethers.ZeroAddress)).to.be
        .revertedWithCustomError(stakingFunding, 'DelegateStakingInvalidAddress');
    });

    it('should bind approval to one hot, remove/re-admit at FIFO tail, preserve withdrawal, and keep self out of active list', async () => {
      const first = funders[0];
      const second = funders[1];
      await stakingFunding.connect(HOT).approvePoolFunder(HOT, first);
      await stakingFunding.connect(first).grantStakingAllowance(HOT, wei(200), await expiry());
      await expect(stakingFunding.connect(first).fundStakingAllowance(HOT_TWO, wei(1))).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingPoolFunderNotApproved',
      );
      await stakingFunding.connect(first).fundStakingAllowance(HOT, wei(100));
      await approveGrantFund(second, wei(100));
      await expect(stakingFunding.connect(HOT).revokePoolFunderApproval(HOT, first))
        .to.emit(stakingFunding, 'DelegateStakingPoolFunderApprovalRevoked')
        .withArgs(HOT, first);
      expect((await stakingFunding.getPoolFunders(HOT, 0, 10))[0][0].inActiveList).to.eq(false);
      await stakingFunding.connect(first).requestStakingWithdrawal(HOT, first, wei(10));
      await stakingFunding.connect(HOT).approvePoolFunder(HOT, first);
      const sessionId = await openPoolSession(wei(110));
      expect((await stakingFunding.getPoolSession(sessionId)).draws.map((d) => d.funder)).to.deep.eq([
        second.address,
        first.address,
      ]);

      await approveGrantFund(HOT, wei(50));
      const self = (await stakingFunding.getPoolFunders(HOT, 0, 10))[0].find((f) => f.funder === HOT.address)!;
      expect(self.inActiveList).to.eq(false);
      await expectAccountingInvariant(HOT);
    });
  });

  describe('#grantStakingAllowance / #revokeStakingAllowance', () => {
    it('should validate hot, expiry boundaries and lifetime cap', async () => {
      const now = await getCurrentBlockTime();
      await expect(
        stakingFunding.connect(funders[0]).grantStakingAllowance(ethers.ZeroAddress, wei(1), now + 1n),
      ).to.be.revertedWithCustomError(stakingFunding, 'DelegateStakingInvalidAddress');
      await expect(stakingFunding.connect(funders[0]).grantStakingAllowance(HOT, wei(1), now)).to.be
        .revertedWithCustomError(stakingFunding, 'DelegateStakingInvalidExpiry');
      await expect(stakingFunding.connect(funders[0]).grantStakingAllowance(HOT, wei(1), now - 1n)).to.be
        .revertedWithCustomError(stakingFunding, 'DelegateStakingInvalidExpiry');
      await approveGrantFund(funders[0], wei(2), HOT, wei(3));
      await expect(
        stakingFunding.connect(funders[0]).grantStakingAllowance(HOT, wei(2) - 1n, await expiry()),
      ).to.be.revertedWithCustomError(stakingFunding, 'DelegateStakingCapBelowLifetimeFunding');
    });

    it('should support many funders, update a revoked grant, and emit exact values', async () => {
      const exp = await expiry();
      await stakingFunding.connect(HOT).approvePoolFunder(HOT, funders[0]);
      await expect(stakingFunding.connect(funders[0]).grantStakingAllowance(HOT, wei(4), exp))
        .to.emit(stakingFunding, 'DelegateStakingAllowanceGranted')
        .withArgs(HOT, funders[0], wei(4), exp);
      await stakingFunding.connect(funders[0]).revokeStakingAllowance(HOT);
      const newExp = exp + BigInt(DAY);
      await stakingFunding.connect(funders[0]).grantStakingAllowance(HOT, wei(8), newExp);
      await approveGrantFund(funders[1], wei(2));
      const allowance = await stakingFunding.getStakingAllowance(HOT, funders[0]);
      expect(allowance.cumulativeFundingCap).to.eq(wei(8));
      expect(allowance.expiry).to.eq(newExp);
      expect(allowance.revoked).to.eq(false);
      expect((await stakingFunding.getPoolFunders(HOT, 0, 10))[1]).to.eq(1);
    });

    it('should reject revoking never-granted, already-revoked, and expired grants and remove a live grant from active draw', async () => {
      await expect(stakingFunding.connect(funders[0]).revokeStakingAllowance(HOT)).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingAllowanceNotActive',
      );
      await approveGrantFund(funders[0], wei(2));
      await expect(stakingFunding.connect(funders[0]).revokeStakingAllowance(HOT))
        .to.emit(stakingFunding, 'DelegateStakingAllowanceRevoked')
        .withArgs(HOT, funders[0]);
      expect((await stakingFunding.getPoolFunders(HOT, 0, 10))[0][0].inActiveList).to.eq(false);
      await expect(stakingFunding.connect(funders[0]).revokeStakingAllowance(HOT)).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingAllowanceNotActive',
      );
      await stakingFunding.connect(funders[1]).grantStakingAllowance(HOT, wei(2), (await getCurrentBlockTime()) + 2n);
      await setTime(Number(await getCurrentBlockTime()) + 3);
      await expect(stakingFunding.connect(funders[1]).revokeStakingAllowance(HOT)).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingAllowanceNotActive',
      );
    });
  });

  describe('#fundStakingAllowance', () => {
    it('should validate amount, approval, live grant, cap, and minimum principal boundaries', async () => {
      const f = funders[0];
      await expect(stakingFunding.connect(f).fundStakingAllowance(HOT, 0)).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingInvalidAmount',
      );
      await stakingFunding.connect(f).grantStakingAllowance(HOT, wei(10), await expiry());
      await expect(stakingFunding.connect(f).fundStakingAllowance(HOT, wei(1))).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingPoolFunderNotApproved',
      );
      await stakingFunding.connect(HOT).approvePoolFunder(HOT, f);
      await stakingFunding.connect(f).revokeStakingAllowance(HOT);
      await expect(stakingFunding.connect(f).fundStakingAllowance(HOT, wei(1))).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingAllowanceNotActive',
      );
      await stakingFunding.connect(f).grantStakingAllowance(HOT, wei(2), await expiry());
      await expect(stakingFunding.connect(f).fundStakingAllowance(HOT, wei(1) - 1n)).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingPrincipalTooLow',
      );
      await stakingFunding.connect(f).fundStakingAllowance(HOT, wei(1));
      await stakingFunding.connect(f).fundStakingAllowance(HOT, wei(1));
      await expect(stakingFunding.connect(f).fundStakingAllowance(HOT, 1)).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingFundingCapExceeded',
      );
    });

    it('should register once, account exactly, transfer tokens, emit, and allow unapproved self-escrow', async () => {
      const beforeDiamond = await token.balanceOf(diamond);
      const beforeFunder = await token.balanceOf(funders[0]);
      await stakingFunding.connect(HOT).approvePoolFunder(HOT, funders[0]);
      await stakingFunding.connect(funders[0]).grantStakingAllowance(HOT, wei(10), await expiry());
      await expect(stakingFunding.connect(funders[0]).fundStakingAllowance(HOT, wei(3)))
        .to.emit(stakingFunding, 'DelegateStakingAllowanceFunded')
        .withArgs(HOT, funders[0], wei(3));
      await stakingFunding.connect(funders[0]).fundStakingAllowance(HOT, wei(2));
      await approveGrantFund(HOT, wei(2));
      const [page, total] = await stakingFunding.getPoolFunders(HOT, 0, 10);
      expect(total).to.eq(2);
      expect(page[0].allowance.lifetimeFunded).to.eq(wei(5));
      expect(page[0].allowance.currentPrincipal).to.eq(wei(5));
      expect(page[0].allowance.active).to.eq(true);
      expect(page[0].allowance.approved).to.eq(true);
      expect(await token.balanceOf(funders[0])).to.eq(beforeFunder - wei(5));
      await expectAccountingInvariant(HOT, wei(7), beforeDiamond);
    });
  });

  describe('#active funder list and #purgeInactiveFunders', () => {
    it('should enforce the admission limit, purge expired grants on admission, and expose cursor behavior', async () => {
      await stakingFunding.setMaxActiveFunders(2);
      await approveGrantFund(funders[0], wei(2));
      await approveGrantFund(funders[1], wei(2), HOT, wei(2), (await getCurrentBlockTime()) + 100n);
      await stakingFunding.connect(HOT).approvePoolFunder(HOT, funders[2]);
      await stakingFunding.connect(funders[2]).grantStakingAllowance(HOT, wei(2), await expiry());
      await expect(stakingFunding.connect(funders[2]).fundStakingAllowance(HOT, wei(2))).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingActiveFundersLimitExceeded',
      );
      await setTime(Number(await getCurrentBlockTime()) + 101);
      await stakingFunding.connect(funders[2]).fundStakingAllowance(HOT, wei(2));
      expect(await stakingFunding.purgeInactiveFunders.staticCall(HOT, 1)).to.eq(0);
      await stakingFunding.purgeInactiveFunders(HOT, 1);
      await stakingFunding.connect(funders[0]).revokeStakingAllowance(HOT);
      expect(await stakingFunding.purgeInactiveFunders.staticCall(HOT, 1)).to.eq(0);
      await stakingFunding.purgeInactiveFunders(HOT, 1);
      await expect(stakingFunding.purgeInactiveFunders(HOT, 129)).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingLimitExceeded',
      );
    });

    it('should put a revoked and re-granted funder at the FIFO tail', async () => {
      await approveGrantFund(funders[0], wei(40), HOT, wei(80));
      await approveGrantFund(funders[1], wei(40));
      await stakingFunding.connect(funders[0]).revokeStakingAllowance(HOT);
      await stakingFunding.connect(funders[0]).grantStakingAllowance(HOT, wei(80), await expiry());

      const sessionId = await openPoolSession(wei(60));
      expect((await stakingFunding.getPoolSession(sessionId)).draws.map((d) => [d.funder, d.amount])).to.deep.eq([
        [funders[1].address, wei(40)],
        [funders[0].address, wei(20)],
      ]);
      await expectAccountingInvariant(HOT);
    });

    it('should put an expired funder removed by a draw at the FIFO tail after re-granting', async () => {
      await approveGrantFund(funders[0], wei(40), HOT, wei(40), (await getCurrentBlockTime()) + 100n);
      await approveGrantFund(funders[1], wei(40));
      await setTime(Number(await getCurrentBlockTime()) + 101);

      const first = await openPoolSession(wei(10));
      expect((await stakingFunding.getPoolSession(first)).draws.map((d) => d.funder)).to.deep.eq([funders[1].address]);
      expect((await stakingFunding.getPoolFunders(HOT, 0, 10))[0][0].inActiveList).to.eq(false);

      await stakingFunding.connect(funders[0]).grantStakingAllowance(HOT, wei(40), await expiry());
      const second = await openPoolSession(wei(60));
      expect((await stakingFunding.getPoolSession(second)).draws.map((d) => [d.funder, d.amount])).to.deep.eq([
        [funders[1].address, wei(30)],
        [funders[0].address, wei(30)],
      ]);
      await expectAccountingInvariant(HOT);
    });

    it('should put a fully withdrawn and re-funded funder at the FIFO tail', async () => {
      await approveGrantFund(funders[0], wei(40), HOT, wei(80));
      await approveGrantFund(funders[1], wei(40));
      await stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(40));
      expect((await stakingFunding.getPoolFunders(HOT, 0, 10))[0][0].inActiveList).to.eq(false);
      await stakingFunding.connect(funders[0]).fundStakingAllowance(HOT, wei(40));

      const sessionId = await openPoolSession(wei(60));
      expect((await stakingFunding.getPoolSession(sessionId)).draws.map((d) => [d.funder, d.amount])).to.deep.eq([
        [funders[1].address, wei(40)],
        [funders[0].address, wei(20)],
      ]);
      await expectAccountingInvariant(HOT);
    });

    it('should put an expired funder renewed before any draw at the FIFO tail', async () => {
      await approveGrantFund(funders[0], wei(40), HOT, wei(40), (await getCurrentBlockTime()) + 100n);
      await approveGrantFund(funders[1], wei(40));
      await setTime(Number(await getCurrentBlockTime()) + 101);
      // still linked: nothing has traversed the list since the expiry
      expect((await stakingFunding.getPoolFunders(HOT, 0, 10))[0][0].inActiveList).to.eq(true);

      await stakingFunding.connect(funders[0]).grantStakingAllowance(HOT, wei(40), await expiry());

      const sessionId = await openPoolSession(wei(60));
      expect((await stakingFunding.getPoolSession(sessionId)).draws.map((d) => [d.funder, d.amount])).to.deep.eq([
        [funders[1].address, wei(40)],
        [funders[0].address, wei(20)],
      ]);
      await expectAccountingInvariant(HOT);
    });

    it('should keep the FIFO position when a live grant is updated', async () => {
      await approveGrantFund(funders[0], wei(40), HOT, wei(80));
      await approveGrantFund(funders[1], wei(40));
      await stakingFunding.connect(funders[0]).grantStakingAllowance(HOT, wei(100), await expiry(60));

      const sessionId = await openPoolSession(wei(60));
      expect((await stakingFunding.getPoolSession(sessionId)).draws.map((d) => d.funder)).to.deep.eq([
        funders[0].address,
        funders[1].address,
      ]);
    });

    it('should revert a grant expiry that does not fit uint64', async () => {
      await expect(stakingFunding.connect(funders[0]).grantStakingAllowance(HOT, wei(1), 2n ** 64n)).to.be.revertedWith(
        "SafeCast: value doesn't fit in 64 bits",
      );
      await stakingFunding.connect(funders[0]).grantStakingAllowance(HOT, wei(1), 2n ** 64n - 1n);
      expect((await stakingFunding.getStakingAllowance(HOT, funders[0])).expiry).to.eq(2n ** 64n - 1n);
    });

    it('should merge locks released on the same day into one bucket', async () => {
      await approveGrantFund(funders[0], wei(100));
      const first = await openPoolSession(wei(50));
      const second = await openPoolSession(wei(50));
      await closePoolSession(first, 100);
      const [[bucket]] = await stakingFunding.getPoolLockBuckets(HOT, 0, 10);
      await closePoolSession(second, 200);

      const [buckets, total] = await stakingFunding.getPoolLockBuckets(HOT, 0, 10);
      expect(total).to.eq(1);
      expect(buckets[0].releaseAt).to.eq(bucket.releaseAt);
      expect(buckets[0].amount).to.be.greaterThan(bucket.amount);
      expect((await stakingFunding.getPoolBalances(HOT)).locked).to.eq(buckets[0].amount);

      await setTime(Number(bucket.releaseAt));
      await expect(stakingFunding.releasePoolLocks(HOT, 32))
        .to.emit(stakingFunding, 'DelegateStakingPoolLockReleased')
        .withArgs(HOT, bucket.releaseAt, buckets[0].amount);
      expect(await stakingFunding.getPoolBalances(HOT)).to.deep.eq([wei(100), 0n, 0n, 0n]);
    });

    it('should purge expired funders still linked and walk the list with the cursor', async () => {
      for (let i = 0; i < 4; i++) {
        const expires = i % 2 === 0 ? (await getCurrentBlockTime()) + 100n : undefined;
        await approveGrantFund(funders[i], wei(2), HOT, wei(2), expires);
      }
      await setTime(Number(await getCurrentBlockTime()) + 101);
      const linked = async () => (await stakingFunding.getPoolFunders(HOT, 0, 10))[0].map((f) => f.inActiveList);
      expect(await linked()).to.deep.eq([true, true, true, true]);

      // cursor walk, one entry per call: f0 (expired) removed, f1 kept, f2 (expired) removed, f3 kept, wraps to head
      expect(await stakingFunding.purgeInactiveFunders.staticCall(HOT, 1)).to.eq(1);
      await stakingFunding.purgeInactiveFunders(HOT, 1);
      expect(await linked()).to.deep.eq([false, true, true, true]);
      await stakingFunding.purgeInactiveFunders(HOT, 1);
      expect(await stakingFunding.purgeInactiveFunders.staticCall(HOT, 1)).to.eq(1);
      await stakingFunding.purgeInactiveFunders(HOT, 1);
      expect(await linked()).to.deep.eq([false, true, false, true]);
      await stakingFunding.purgeInactiveFunders(HOT, 1);
      expect(await stakingFunding.purgeInactiveFunders.staticCall(HOT, 128)).to.eq(0);
      expect(await stakingFunding.purgeInactiveFunders.staticCall(HOT, 0)).to.eq(0);
    });

    it('should purge everything inactive in one call and only on admission when the list is full', async () => {
      await approveGrantFund(funders[0], wei(2), HOT, wei(2), (await getCurrentBlockTime()) + 100n);
      await approveGrantFund(funders[1], wei(2), HOT, wei(2), (await getCurrentBlockTime()) + 100n);
      await setTime(Number(await getCurrentBlockTime()) + 101);

      // free slots: admitting a new funder does not purge
      await approveGrantFund(funders[2], wei(2));
      let page = (await stakingFunding.getPoolFunders(HOT, 0, 10))[0];
      expect(page.map((f) => f.inActiveList)).to.deep.eq([true, true, true]);

      // full list: admission purges the expired ones first
      await stakingFunding.setMaxActiveFunders(3);
      await approveGrantFund(funders[3], wei(2));
      page = (await stakingFunding.getPoolFunders(HOT, 0, 10))[0];
      expect(page.map((f) => f.inActiveList)).to.deep.eq([false, false, true, true]);
    });

    it('should return the number of removed funders from a full purge', async () => {
      await approveGrantFund(funders[0], wei(2), HOT, wei(2), (await getCurrentBlockTime()) + 100n);
      await approveGrantFund(funders[1], wei(2), HOT, wei(2), (await getCurrentBlockTime()) + 100n);
      await approveGrantFund(funders[2], wei(2));
      await setTime(Number(await getCurrentBlockTime()) + 101);

      expect(await stakingFunding.purgeInactiveFunders.staticCall(HOT, 128)).to.eq(2);
      await stakingFunding.purgeInactiveFunders(HOT, 128);
      expect((await stakingFunding.getPoolFunders(HOT, 0, 10))[0].map((f) => f.inActiveList)).to.deep.eq([
        false,
        false,
        true,
      ]);
    });

    it('should not let a lowered limit block existing funder draw, close, or withdrawal', async () => {
      await approveGrantFund(funders[0], wei(100));
      await approveGrantFund(funders[1], wei(100));
      await stakingFunding.setMaxActiveFunders(1);
      const sessionId = await openPoolSession(wei(150));
      expect((await stakingFunding.getPoolSession(sessionId)).draws.length).to.eq(2);
      await closePoolSession(sessionId, 1);
      await stakingFunding.connect(funders[1]).requestStakingWithdrawal(HOT, funders[1], wei(10));
      await expectAccountingInvariant(HOT);
    });
  });

  describe('#requestStakingWithdrawal / #cancelStakingWithdrawal', () => {
    it('should enforce rights, amount, principal, and minimum remainder and pay an exact full withdrawal', async () => {
      await approveGrantFund(funders[0], wei(3));
      await expect(stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], 0)).to.be
        .revertedWithCustomError(stakingFunding, 'DelegateStakingInvalidAmount');
      await expect(
        stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(3) + 1n),
      ).to.be.revertedWithCustomError(stakingFunding, 'DelegateStakingInsufficientPrincipal');
      await expect(
        stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(2) + 1n),
      ).to.be.revertedWithCustomError(stakingFunding, 'DelegateStakingPrincipalTooLow');
      await expect(stakingFunding.connect(STRANGER).requestStakingWithdrawal(HOT, funders[0], wei(1))).to.be
        .revertedWithCustomError(stakingFunding, 'InsufficientRightsForOperation');
      await expect(stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(3)))
        .to.emit(stakingFunding, 'DelegateStakingWithdrawalRequested')
        .withArgs(HOT, funders[0], wei(3), wei(3), 0)
        .and.to.emit(stakingFunding, 'DelegateStakingWithdrawalPaid')
        .withArgs(HOT, funders[0], wei(3));
      await expectAccountingInvariant(HOT);
    });

    it('should let a delegate withdraw to the funder and allow withdrawal after revoke or expiry', async () => {
      await approveGrantFund(funders[0], wei(5));
      await delegateRegistry
        .connect(funders[0])
        .delegateContract(DELEGATE, diamond, await stakingFunding.getDelegateStakingDelegationRight(), true);
      const funderBefore = await token.balanceOf(funders[0]);
      const delegateBefore = await token.balanceOf(DELEGATE);
      await stakingFunding.connect(DELEGATE).requestStakingWithdrawal(HOT, funders[0], wei(1));
      expect(await token.balanceOf(funders[0])).to.eq(funderBefore + wei(1));
      expect(await token.balanceOf(DELEGATE)).to.eq(delegateBefore);
      await stakingFunding.connect(funders[0]).revokeStakingAllowance(HOT);
      await stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(1));

      await approveGrantFund(funders[1], wei(2), HOT, wei(2), (await getCurrentBlockTime()) + 10n);
      await setTime(Number(await getCurrentBlockTime()) + 11);
      await stakingFunding.connect(funders[1]).requestStakingWithdrawal(HOT, funders[1], wei(1));
    });

    it('should pay free funds, queue the rest, coalesce requests, reserve FIFO, and not lock other funders pro-rata', async () => {
      await approveGrantFund(funders[0], wei(100));
      await approveGrantFund(funders[1], wei(100));
      const sessionId = await openPoolSession(wei(150));
      await expect(stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(80)))
        .to.emit(stakingFunding, 'DelegateStakingWithdrawalRequested')
        .withArgs(HOT, funders[0], wei(80), wei(50), wei(30));
      await stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(10));
      await stakingFunding.connect(funders[1]).requestStakingWithdrawal(HOT, funders[1], wei(20));
      const [queue, next, total] = await stakingFunding.getPoolWithdrawalQueue(HOT, ethers.ZeroAddress, 10);
      expect(queue.map((q) => [q.funder, q.amount])).to.deep.eq([
        [funders[0].address, wei(40)],
        [funders[1].address, wei(20)],
      ]);
      expect(next).to.eq(ethers.ZeroAddress);
      expect(total).to.eq(2);
      expect((await stakingFunding.getStakingAllowance(HOT, funders[1])).exposure).to.eq(wei(50));
      expect(await stakingFunding.getStakingWithdrawal(HOT, funders[0])).to.eq(wei(40));
      await closePoolSession(sessionId, 1);
      await expectAccountingInvariant(HOT);
    });

    it('should move a repeated request behind earlier requests of other funders', async () => {
      await approveGrantFund(funders[0], wei(100));
      await approveGrantFund(funders[1], wei(100));
      const sessionId = await openPoolSession(wei(200));
      await stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(10));
      await stakingFunding.connect(funders[1]).requestStakingWithdrawal(HOT, funders[1], wei(20));
      await stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(30));

      const [queue, , total] = await stakingFunding.getPoolWithdrawalQueue(HOT, ethers.ZeroAddress, 10);
      expect(queue.map((q) => [q.funder, q.amount])).to.deep.eq([
        [funders[1].address, wei(20)],
        [funders[0].address, wei(40)],
      ]);
      expect(total).to.eq(2);
      await closePoolSession(sessionId, 1);
      await expectAccountingInvariant(HOT);
    });

    it('should validate cancellation and restore/re-admit at the tail', async () => {
      await expect(stakingFunding.connect(funders[0]).cancelStakingWithdrawal(HOT, funders[0])).to.be
        .revertedWithCustomError(stakingFunding, 'DelegateStakingWithdrawalNotFound');
      await approveGrantFund(funders[0], wei(200));
      await approveGrantFund(funders[1], wei(200));
      const sessionId = await openPoolSession(wei(300));
      await stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(200));
      await expect(stakingFunding.connect(STRANGER).cancelStakingWithdrawal(HOT, funders[0])).to.be
        .revertedWithCustomError(stakingFunding, 'InsufficientRightsForOperation');
      await expect(stakingFunding.connect(funders[0]).cancelStakingWithdrawal(HOT, funders[0]))
        .to.emit(stakingFunding, 'DelegateStakingWithdrawalCancelled')
        .withArgs(HOT, funders[0], wei(100));
      expect(await stakingFunding.getStakingWithdrawal(HOT, funders[0])).to.eq(0);
      expect((await stakingFunding.getStakingAllowance(HOT, funders[0])).currentPrincipal).to.eq(wei(100));
      expect((await stakingFunding.getPoolBalances(HOT)).pendingTotal).to.eq(0);
      expect((await stakingFunding.getPoolWithdrawalQueue(HOT, ethers.ZeroAddress, 10))[2]).to.eq(0);

      // funders[0] was unlinked by the full withdrawal and re-admitted at the tail by the cancel
      expect((await stakingFunding.getPoolFunders(HOT, 0, 10))[0][0].inActiveList).to.eq(true);
      await closePoolSession(sessionId, DAY);
      const next = await openPoolSession(wei(250));
      expect((await stakingFunding.getPoolSession(next)).draws.map((d) => [d.funder, d.amount])).to.deep.eq([
        [funders[1].address, wei(200)],
        [funders[0].address, wei(50)],
      ]);
      await expectAccountingInvariant(HOT);
    });

    it('should enforce minimum restored principal only for live approved grants and enforce admission limit', async () => {
      await approveGrantFund(HOT, wei(20));
      await stakingFunding.setMinPrincipal(wei(30));
      await stakingFunding.connect(HOT).approvePoolFunder(HOT, funders[0]);
      await stakingFunding.connect(funders[0]).grantStakingAllowance(HOT, wei(30), await expiry());
      await stakingFunding.connect(funders[0]).fundStakingAllowance(HOT, wei(30));
      await openPoolSession(wei(50));
      await stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(30));
      await stakingFunding.setMinPrincipal(wei(40));
      await expect(stakingFunding.connect(funders[0]).cancelStakingWithdrawal(HOT, funders[0])).to.be
        .revertedWithCustomError(stakingFunding, 'DelegateStakingPrincipalTooLow');
      await stakingFunding.connect(funders[0]).revokeStakingAllowance(HOT);
      await stakingFunding.connect(funders[0]).cancelStakingWithdrawal(HOT, funders[0]);

      await stakingFunding.setMinPrincipal(wei(1));
      await stakingFunding.setMaxActiveFunders(1);
      await approveGrantFund(funders[1], wei(2));
      await expect(stakingFunding.connect(funders[0]).grantStakingAllowance(HOT, wei(40), await expiry())).to.be
        .revertedWithCustomError(stakingFunding, 'DelegateStakingActiveFundersLimitExceeded');
    });
  });

  describe('#servePoolWithdrawals', () => {
    it('should be permissionless, return serviced count, pay FIFO to funders, partially service, and stop at zero free', async () => {
      await stakingFunding.setMaxAutoService(1);
      await approveGrantFund(funders[0], wei(50));
      await approveGrantFund(funders[1], wei(50));
      await approveGrantFund(funders[2], wei(50));
      const sessionId = await openPoolSession(wei(140));
      await stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(40));
      await stakingFunding.connect(funders[1]).requestStakingWithdrawal(HOT, funders[1], wei(40));
      await stakingFunding.connect(funders[2]).requestStakingWithdrawal(HOT, funders[2], wei(40));
      await closePoolSession(sessionId, 1);
      expect(await stakingFunding.connect(STRANGER).servePoolWithdrawals.staticCall(HOT, 1)).to.eq(1);
      const firstBefore = await token.balanceOf(funders[1]);
      await stakingFunding.connect(STRANGER).servePoolWithdrawals(HOT, 1);
      expect(await token.balanceOf(funders[1])).to.eq(firstBefore + wei(40));
      expect(await token.balanceOf(STRANGER)).to.eq(0);
      expect(await stakingFunding.servePoolWithdrawals.staticCall(HOT, 10)).to.eq(1);
      await expect(stakingFunding.servePoolWithdrawals(HOT, 33)).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingLimitExceeded',
      );
    });
  });

  describe('#releasePoolLocks', () => {
    it('should release only matured buckets, emit, return count, move the head, and service queued withdrawals', async () => {
      await approveGrantFund(funders[0], wei(50));
      const sessionId = await openPoolSession(wei(50));
      await closePoolSession(sessionId, 100);
      const [buckets] = await stakingFunding.getPoolLockBuckets(HOT, 0, 10);
      expect(buckets.length).to.eq(1);
      await stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(50));
      expect(await stakingFunding.releasePoolLocks.staticCall(HOT, 10)).to.eq(0);
      await setTime(Number(buckets[0].releaseAt));
      expect(await stakingFunding.releasePoolLocks.staticCall(HOT, 1)).to.eq(1);
      await expect(stakingFunding.releasePoolLocks(HOT, 1))
        .to.emit(stakingFunding, 'DelegateStakingPoolLockReleased')
        .withArgs(HOT, buckets[0].releaseAt, buckets[0].amount)
        .and.to.emit(stakingFunding, 'DelegateStakingWithdrawalPaid');
      expect((await stakingFunding.getPoolLockBuckets(HOT, 0, 10))[0]).to.deep.eq([]);
      await expect(stakingFunding.releasePoolLocks(HOT, 33)).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingLimitExceeded',
      );
      await expectAccountingInvariant(HOT);
    });

    it('should expose matured-but-unreleased liquidity and lazily release an old bucket on fund', async () => {
      await stakingFunding.setMaxAutoReleaseDays(1);
      await approveGrantFund(funders[0], wei(50), HOT, wei(60));
      const sessionId = await openPoolSession(wei(50));
      await closePoolSession(sessionId, 100);
      const [buckets] = await stakingFunding.getPoolLockBuckets(HOT, 0, 10);
      await setTime(Number(buckets[0].releaseAt) + 20 * DAY);
      expect(await stakingFunding.getAvailablePoolStake(HOT)).to.eq(wei(50));
      await stakingFunding.connect(funders[0]).fundStakingAllowance(HOT, wei(1));
      expect((await stakingFunding.getPoolLockBuckets(HOT, 0, 10))[1]).to.eq(0);
      await expectAccountingInvariant(HOT);
    });
  });

  describe('withdrawal queue and lock release scenarios', () => {
    it('should revert a cancel that needs a slot when the active list is full', async () => {
      await approveGrantFund(funders[0], wei(50));
      await approveGrantFund(funders[1], wei(50));
      const sessionId = await openPoolSession(wei(100));
      await stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(50));
      expect((await stakingFunding.getPoolFunders(HOT, 0, 10))[0][0].inActiveList).to.eq(false);

      // only funders[1] is active now; a limit of one leaves no slot for funders[0]
      await stakingFunding.setMaxActiveFunders(1);
      await expect(stakingFunding.connect(funders[0]).cancelStakingWithdrawal(HOT, funders[0])).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingActiveFundersLimitExceeded',
      );

      // the withdrawal stays queued and is paid once the pool has liquidity again
      await closePoolSession(sessionId, DAY);
      expect(await stakingFunding.getStakingWithdrawal(HOT, funders[0])).to.eq(0);
      await expectAccountingInvariant(HOT);
    });

    it('should service the queue partially, keep the head, and stop when free is zero', async () => {
      await stakingFunding.setMaxAutoService(1);
      for (const f of funders.slice(0, 3)) await approveGrantFund(f, wei(50));
      const sessionId = await openPoolSession(wei(150));
      for (const f of funders.slice(0, 3)) {
        await stakingFunding.connect(f).requestStakingWithdrawal(HOT, f, wei(50));
      }
      expect(await stakingFunding.servePoolWithdrawals.staticCall(HOT, 32)).to.eq(0);

      // early close locks part of the stake; close itself serves one entry (funders[0])
      await closePoolSession(sessionId, 100);
      const [[bucket]] = await stakingFunding.getPoolLockBuckets(HOT, 0, 10);
      expect(await stakingFunding.getStakingWithdrawal(HOT, funders[0])).to.eq(0);
      expect((await stakingFunding.getPoolBalances(HOT)).free).to.eq(wei(100) - bucket.amount);

      // funders[1] is paid in full, funders[2] partially and stays at the head
      expect(await stakingFunding.servePoolWithdrawals.staticCall(HOT, 32)).to.eq(2);
      await expect(stakingFunding.connect(STRANGER).servePoolWithdrawals(HOT, 32)).to.changeTokenBalances(
        token,
        [funders[1], funders[2], STRANGER, HOT],
        [wei(50), wei(50) - bucket.amount, 0n, 0n],
      );
      expect(await stakingFunding.getStakingWithdrawal(HOT, funders[2])).to.eq(bucket.amount);
      expect((await stakingFunding.getPoolWithdrawalQueue(HOT, ethers.ZeroAddress, 10))[0]).to.deep.eq([
        [funders[2].address, bucket.amount],
      ]);
      expect((await stakingFunding.getPoolBalances(HOT)).free).to.eq(0);
      expect(await stakingFunding.servePoolWithdrawals.staticCall(HOT, 32)).to.eq(0);

      // the remainder is paid when the lock is released
      await setTime(Number(bucket.releaseAt));
      await expect(stakingFunding.releasePoolLocks(HOT, 32)).to.changeTokenBalance(token, funders[2], bucket.amount);
      expect(await stakingFunding.getPoolBalances(HOT)).to.deep.eq([0n, 0n, 0n, 0n]);
      await expectAccountingInvariant(HOT);
    });

    it('should limit automatic queue service to maxAutoService entries per operation', async () => {
      await stakingFunding.setMaxAutoService(1);
      for (const f of funders.slice(0, 3)) await approveGrantFund(f, wei(50));
      const sessionId = await openPoolSession(wei(150));
      for (const f of funders.slice(0, 3)) {
        await stakingFunding.connect(f).requestStakingWithdrawal(HOT, f, wei(10));
      }
      await closePoolSession(sessionId, DAY);
      // close served exactly one entry
      expect((await stakingFunding.getPoolWithdrawalQueue(HOT, ethers.ZeroAddress, 10))[2]).to.eq(2);

      await approveGrantFund(funders[3], wei(5));
      expect((await stakingFunding.getPoolWithdrawalQueue(HOT, ethers.ZeroAddress, 10))[2]).to.eq(1);

      expect(await stakingFunding.servePoolWithdrawals.staticCall(HOT, 32)).to.eq(1);
    });

    it('should release a lock left untouched for weeks on the first withdrawal (AC-COLDC-8)', async () => {
      await approveGrantFund(funders[0], wei(50));
      const sessionId = await openPoolSession(wei(50));
      await closePoolSession(sessionId, 100);
      const [[bucket]] = await stakingFunding.getPoolLockBuckets(HOT, 0, 10);
      expect(bucket.amount).to.be.greaterThan(0);

      await setTime(Number(bucket.releaseAt) + 30 * DAY);
      expect(await stakingFunding.releasePoolLocks.staticCall(HOT, 0)).to.eq(0);
      const tx = stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(50));
      await expect(tx)
        .to.emit(stakingFunding, 'DelegateStakingPoolLockReleased')
        .withArgs(HOT, bucket.releaseAt, bucket.amount);
      await expect(tx).to.changeTokenBalances(token, [funders[0], HOT], [wei(50), 0n]);
      expect(await stakingFunding.getPoolBalances(HOT)).to.deep.eq([0n, 0n, 0n, 0n]);
    });

    it('should pay a withdrawal during pool use from free liquidity without locking other funders (AC-COLDC-4)', async () => {
      await approveGrantFund(funders[0], wei(60));
      await approveGrantFund(funders[1], wei(60));
      await approveGrantFund(funders[2], wei(60));
      const sessionId = await openPoolSession(wei(100));
      // draws: f0 60, f1 40; free 80

      const tx = stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(60));
      await expect(tx)
        .to.emit(stakingFunding, 'DelegateStakingWithdrawalRequested')
        .withArgs(HOT, funders[0], wei(60), wei(60), 0);
      await expect(tx).to.changeTokenBalance(token, funders[0], wei(60));

      // f2 is not locked pro-rata: its full principal is still withdrawable up to the remaining free 20
      await expect(stakingFunding.connect(funders[2]).requestStakingWithdrawal(HOT, funders[2], wei(60)))
        .to.emit(stakingFunding, 'DelegateStakingWithdrawalRequested')
        .withArgs(HOT, funders[2], wei(60), wei(20), wei(40));
      // the queue is non-empty, so f1's request waits entirely (last-out waits)
      await expect(stakingFunding.connect(funders[1]).requestStakingWithdrawal(HOT, funders[1], wei(10)))
        .to.emit(stakingFunding, 'DelegateStakingWithdrawalRequested')
        .withArgs(HOT, funders[1], wei(10), 0, wei(10));
      // coalesced into the same node
      await expect(stakingFunding.connect(funders[1]).requestStakingWithdrawal(HOT, funders[1], wei(10)))
        .to.emit(stakingFunding, 'DelegateStakingWithdrawalRequested')
        .withArgs(HOT, funders[1], wei(10), 0, wei(20));
      expect((await stakingFunding.getPoolWithdrawalQueue(HOT, ethers.ZeroAddress, 10))[2]).to.eq(2);

      await closePoolSession(sessionId, DAY);
      expect(await stakingFunding.getStakingWithdrawal(HOT, funders[2])).to.eq(0);
      expect(await stakingFunding.getStakingWithdrawal(HOT, funders[1])).to.eq(0);
      await expectAccountingInvariant(HOT);
    });

    it('should conserve every funder balance and never pay the hot wallet across a mixed scenario', async () => {
      const actors = funders.slice(0, 3);
      const start = await Promise.all(actors.map((f) => token.balanceOf(f)));
      const hotStart = await token.balanceOf(HOT);

      for (const f of actors) await approveGrantFund(f, wei(100), HOT, wei(200));
      const first = await openPoolSession(wei(150));
      await stakingFunding.connect(actors[0]).requestStakingWithdrawal(HOT, actors[0], wei(70));
      await stakingFunding.connect(actors[1]).revokeStakingAllowance(HOT);
      await closePoolSession(first, 300);
      const second = await openPoolSession(wei(50));
      await stakingFunding.connect(actors[2]).fundStakingAllowance(HOT, wei(50));
      await stakingFunding.connect(actors[1]).requestStakingWithdrawal(HOT, actors[1], wei(100));
      await closePoolSession(second, DAY);
      await stakingFunding.releasePoolLocks(HOT, 32);
      await stakingFunding.servePoolWithdrawals(HOT, 32);

      for (let i = 0; i < actors.length; i++) {
        const allowance = await stakingFunding.getStakingAllowance(HOT, actors[i]);
        const paidOut = (await token.balanceOf(actors[i])) - start[i] + allowance.lifetimeFunded;
        expect(allowance.lifetimeFunded - paidOut).to.eq(allowance.currentPrincipal + allowance.pending);
      }
      expect(await token.balanceOf(HOT)).to.eq(hotStart);
      await expectAccountingInvariant(HOT);
    });

    it('should pay the hot wallet only its own self-escrow', async () => {
      await approveGrantFund(funders[0], wei(50));
      await approveGrantFund(HOT, wei(30));
      await expect(stakingFunding.connect(HOT).requestStakingWithdrawal(HOT, HOT, wei(30))).to.changeTokenBalances(
        token,
        [HOT, funders[0]],
        [wei(30), 0n],
      );
      await expect(stakingFunding.connect(HOT).requestStakingWithdrawal(HOT, funders[0], wei(1))).to.be.revertedWithCustomError(
        stakingFunding,
        'InsufficientRightsForOperation',
      );
      await expectAccountingInvariant(HOT);
    });
  });

  describe('self-escrow and cursor edge cases', () => {
    it('should let the hot wallet revoke its own approval and self-grant without touching the active list', async () => {
      await approveGrantFund(funders[0], wei(10));
      await approveGrantFund(HOT, wei(10));

      await expect(stakingFunding.connect(HOT).revokePoolFunderApproval(HOT, HOT))
        .to.emit(stakingFunding, 'DelegateStakingPoolFunderApprovalRevoked')
        .withArgs(HOT, HOT);
      // self-escrow never needs an approval
      expect(await stakingFunding.isPoolFunderApproved(HOT, HOT)).to.eq(true);

      await expect(stakingFunding.connect(HOT).revokeStakingAllowance(HOT))
        .to.emit(stakingFunding, 'DelegateStakingAllowanceRevoked')
        .withArgs(HOT, HOT);
      expect((await stakingFunding.getStakingAllowance(HOT, HOT)).active).to.eq(false);
      expect(await stakingFunding.getAvailablePoolStake(HOT)).to.eq(wei(10));

      const page = (await stakingFunding.getPoolFunders(HOT, 0, 10))[0];
      expect(page.map((f) => [f.funder, f.inActiveList])).to.deep.eq([
        [funders[0].address, true],
        [HOT.address, false],
      ]);
      await expect(stakingFunding.connect(HOT).requestStakingWithdrawal(HOT, HOT, wei(10))).to.changeTokenBalance(
        token,
        HOT,
        wei(10),
      );
    });

    it('should count self-escrow capacity and stop at an immature bucket in getAvailablePoolStake', async () => {
      await approveGrantFund(funders[0], wei(20));
      await approveGrantFund(HOT, wei(30));
      expect(await stakingFunding.getAvailablePoolStake(HOT)).to.eq(wei(50));

      const sessionId = await openPoolSession(wei(50));
      await closePoolSession(sessionId, 100);
      const [[bucket]] = await stakingFunding.getPoolLockBuckets(HOT, 0, 10);

      // the bucket is not mature yet: it does not count as liquidity
      expect(await stakingFunding.getAvailablePoolStake(HOT)).to.eq(wei(50) - bucket.amount);
      await setTime(Number(bucket.releaseAt));
      expect(await stakingFunding.getAvailablePoolStake(HOT)).to.eq(wei(50));
    });

    it('should move the purge cursor when the funder under it is removed', async () => {
      for (let i = 0; i < 3; i++) await approveGrantFund(funders[i], wei(2));

      // cursor stops at funders[1]; revoking funders[1] moves it to funders[2]
      await stakingFunding.purgeInactiveFunders(HOT, 1);
      await stakingFunding.connect(funders[1]).revokeStakingAllowance(HOT);
      await stakingFunding.connect(funders[0]).revokeStakingAllowance(HOT);
      expect(await stakingFunding.purgeInactiveFunders.staticCall(HOT, 1)).to.eq(0);

      // cursor at the tail (funders[2]); removing the tail wraps the cursor to the head
      await stakingFunding.purgeInactiveFunders(HOT, 0);
      await approveGrantFund(funders[3], wei(2));
      await stakingFunding.purgeInactiveFunders(HOT, 1);
      await stakingFunding.connect(funders[3]).revokeStakingAllowance(HOT);
      expect((await stakingFunding.getPoolFunders(HOT, 0, 10))[0].map((f) => f.inActiveList)).to.deep.eq([
        false,
        false,
        true,
        false,
      ]);
      expect(await stakingFunding.purgeInactiveFunders.staticCall(HOT, 128)).to.eq(0);
    });
  });

  describe('#parameter setters', () => {
    const setters = [
      ['setMaxActiveFunders', 128n, 'DelegateStakingMaxActiveFundersUpdated', 64n],
      ['setMinPrincipal', wei(1000000000), 'DelegateStakingMinPrincipalUpdated', wei(1)],
      ['setSettlementGrace', BigInt(7 * DAY), 'DelegateStakingSettlementGraceUpdated', BigInt(DAY)],
      ['setMaxAutoService', 32n, 'DelegateStakingMaxAutoServiceUpdated', 5n],
      ['setMaxAutoReleaseDays', 32n, 'DelegateStakingMaxAutoReleaseDaysUpdated', 8n],
    ] as const;

    for (const [method, limit, event, defaultValue] of setters) {
      it(`should enforce ownership and boundaries for ${method}`, async () => {
        await expect((stakingFunding.connect(STRANGER) as any)[method](limit)).to.be.revertedWithCustomError(
          stakingFunding,
          'OwnableUnauthorizedAccount',
        );
        await expect((stakingFunding as any)[method](limit)).to.emit(stakingFunding, event).withArgs(limit);
        await expect((stakingFunding as any)[method](limit + 1n)).to.be.revertedWithCustomError(
          stakingFunding,
          'DelegateStakingLimitExceeded',
        );
        await expect((stakingFunding as any)[method](0)).to.emit(stakingFunding, event).withArgs(0);
        const parameters = await stakingFunding.getDelegateStakingParameters();
        const index = setters.findIndex((row) => row[0] === method);
        expect(parameters[index]).to.eq(defaultValue);
      });
    }

    it('should return all hard limits and the delegation right', async () => {
      expect(await stakingFunding.getDelegateStakingLimits()).to.deep.eq([
        128n,
        32n,
        32n,
        BigInt(7 * DAY),
        wei(1000000000),
      ]);
      expect(await stakingFunding.getDelegateStakingDelegationRight()).to.eq(
        ethers.keccak256(ethers.toUtf8Bytes('delegation.rules.delegate.staking')),
      );
    });
  });

  describe('#views', () => {
    it('should report approval rules, complete allowance fields, funder pagination, and limits', async () => {
      expect(await stakingFunding.isPoolFunderApproved(HOT, ethers.ZeroAddress)).to.eq(false);
      expect(await stakingFunding.isPoolFunderApproved(HOT, HOT)).to.eq(true);
      const exp = await expiry();
      await approveGrantFund(funders[0], wei(3), HOT, wei(5), exp);
      const allowance = await stakingFunding.getStakingAllowance(HOT, funders[0]);
      expect(allowance).to.deep.eq([wei(5), exp, false, wei(3), wei(3), 0n, 0n, 0n, true, true]);
      expect((await stakingFunding.getPoolFunders(HOT, 1, 10))).to.deep.eq([[], 1n]);
      expect((await stakingFunding.getPoolFunders(HOT, 0, 1))[0][0].inActiveList).to.eq(true);
      await expect(stakingFunding.getPoolFunders(HOT, 0, 129)).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingLimitExceeded',
      );
    });

    it('should return min(liquidity, capacity) and zero when pending exceeds free', async () => {
      await approveGrantFund(funders[0], wei(100));
      expect(await stakingFunding.getAvailablePoolStake(HOT)).to.eq(wei(100));
      await openPoolSession(wei(90));
      expect(await stakingFunding.getAvailablePoolStake(HOT)).to.eq(wei(10));
      await stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(50));
      expect(await stakingFunding.getPoolBalances(HOT)).to.deep.eq([0n, wei(90), 0n, wei(40)]);
      expect(await stakingFunding.getAvailablePoolStake(HOT)).to.eq(0);
    });

    it('should paginate and validate the withdrawal queue cursor and limit', async () => {
      await approveGrantFund(funders[0], wei(40));
      await approveGrantFund(funders[1], wei(40));
      await openPoolSession(wei(80));
      await stakingFunding.connect(funders[0]).requestStakingWithdrawal(HOT, funders[0], wei(20));
      await stakingFunding.connect(funders[1]).requestStakingWithdrawal(HOT, funders[1], wei(30));
      const [first, next, total] = await stakingFunding.getPoolWithdrawalQueue(HOT, ethers.ZeroAddress, 1);
      expect(first).to.deep.eq([[funders[0].address, wei(20)]]);
      expect(next).to.eq(funders[1]);
      expect(total).to.eq(2);
      const [second, last] = await stakingFunding.getPoolWithdrawalQueue(HOT, next, 1);
      expect(second).to.deep.eq([[funders[1].address, wei(30)]]);
      expect(last).to.eq(ethers.ZeroAddress);
      expect((await stakingFunding.getPoolWithdrawalQueue(HOT, ethers.ZeroAddress, 10))[0]).to.deep.eq([
        [funders[0].address, wei(20)],
        [funders[1].address, wei(30)],
      ]);
      await expect(stakingFunding.getPoolWithdrawalQueue(HOT, STRANGER, 1)).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingInvalidWithdrawalCursor',
      );
      await expect(stakingFunding.getPoolWithdrawalQueue(HOT, ethers.ZeroAddress, 129)).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingLimitExceeded',
      );
    });

    it('should paginate lock buckets after the head advances and validate the limit', async () => {
      await approveGrantFund(funders[0], wei(100));
      let sessionId = await openPoolSession(wei(20));
      await closePoolSession(sessionId, 100);
      const first = (await stakingFunding.getPoolLockBuckets(HOT, 0, 10))[0][0];
      await setTime(Number(first.releaseAt) + 1);
      await stakingFunding.releasePoolLocks(HOT, 1);
      const [page, total] = await stakingFunding.getPoolLockBuckets(HOT, 0, 1);
      expect(page.length).to.eq(0);
      expect(total).to.eq(0);
      await expect(stakingFunding.getPoolLockBuckets(HOT, 0, 33)).to.be.revertedWithCustomError(
        stakingFunding,
        'DelegateStakingLimitExceeded',
      );
      expect((await stakingFunding.getPoolSession(sessionId)).closed).to.eq(true);
    });
  });
});
