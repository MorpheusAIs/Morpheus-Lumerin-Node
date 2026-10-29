# Stage 1 RFP v2 — diamond upgrade runbook

Forward-only cut from the pre-pool SessionRouter to the Stage 1 pool-aware facets.
Implements RFP §4.2 / COLDC-R20. Scope: preflight, cut review and verification.
Execution, multisig proposals/signing, facet deployment and production deployment/calldata
scripts are a separate step. The test helper is the cut specification, not an operator script.

## Step 0 — preconditions and snapshot (manual)

- [ ] Pin the reviewed Stage 1 release commit and build artifacts. The supported starting
  implementation is the initialized **pre-Stage-1, pre-pool SessionRouter** represented by
  `contracts/mock/legacy/SessionRouterLegacy.sol` / `ISessionRouterLegacy.sol`, with
  `IStatsStorage` routed to it. This is a code baseline, not a documented release tag;
  establish the actual live version from deployment records. Do not deploy the test mock.
- [ ] At a recorded block number/hash/time, read `owner()` and the loupe selector map.
  Every Replace selector below must already exist on the expected legacy facet; every
  Add selector must be absent. Record the other facet routes too. If the live diamond
  differs, stop and review a cut for that state; do not replay this first-install cut.
- [ ] Read deployed bytecode for all live facets and linked libraries; reconcile it with
  the approved pre-upgrade artifacts/deployment records (including library links and
  compiler metadata). Verify the candidate facets and library against the release build.
  Each new facet must fit EIP-170 (24,576 runtime bytes). Library links: SessionRouter,
  ProviderPayouts and DelegateStaking use `LinearDistributionIntervalDecrease`;
  DelegateStakingFunding has no such link.
- [ ] Establish that Stage 1 provider-payout storage is uninitialized and no delegated
  pool state exists. Existing emission `getPools()` / `getPool()` configuration is legacy
  state and must be preserved; it is distinct from delegated staking pools.
- [ ] Save a block-pinned state export and selector/bytecode manifest. Include sessions
  and their user/provider/model lists; provider records and limiter fields;
  `getProvidersTotalClaimed`, `getUserStakesOnHold`, provider/model stats; funding account,
  max duration, emission pools, total session counts and used approval samples. Save
  emission/math view results at fixed explicit timestamps. Include registry/model/bid
  records, fees, price bounds, token, delegation registry, owner and MOR balances of the
  diamond and affected users/providers/funding account. These mirror `readState`,
  `readViews`, `readStorage` and `getBalances` in `DiamondUpgrade.test.ts`.

No production snapshot/export or live-bytecode reconciliation script is supplied by the
cited upgrade helper/tests. These are manual release gates; a read export is evidence,
not a way to roll back on-chain state. The tests' `evm_snapshot` / `evm_revert` are local
comparison tools only. Rehearse the checks below on a fork of the recorded state before
execution; fork setup and receipt comparison are manual work.

## Step 1 — review the exact cut

One `diamondCut((address,uint8,bytes4[])[],address,bytes)` with these five groups.
Names below identify selectors; derive canonical signatures/bytes4 values from the
release interfaces, including inherited functions, as the helper does. No Remove group.

1. **Replace → SessionRouter (20)** — all `ISessionRouter` + `IStatsStorage` functions:
   - `__SessionRouter_init`, `setPoolConfig`, `setMaxSessionDuration`, `openSession`,
     `closeSession`, `claimForProvider`, `withdrawUserStakes`.
   - Inherited `ISessionStorage`: `getSession`, `getUserSessions`, `getProviderSessions`,
     `getModelSessions`, `getPools`, `getPool`, `getFundingAccount`, `getTotalSessions`,
     `getProvidersTotalClaimed`, `getIsProviderApprovalUsed`, `getMaxSessionDuration`.
   - `IStatsStorage`: `getProviderModelStats`, `getModelStats`.
2. **Replace → ProviderPayouts (9)** — intersection of `IProviderPayouts` and
   `ISessionRouterLegacy`: `getSessionId`, `getSessionEnd`, `stakeToStipend`,
   `stipendToStake`, `getUserStakesOnHold`, `getTodaysBudget`, `getComputeBalance`,
   `totalMORSupply`, `startOfTheDay`.
3. **Add → ProviderPayouts (7)** — remaining `IProviderPayouts` functions:
   `DELEGATION_RULES_PAYOUT`, `__ProviderPayouts_init`, `setPayoutTarget`,
   `getPayoutTarget`, `getProviderOwed`, `getPendingProviderClaims`,
   `getProviderLifetimeClaimed`.
4. **Add → DelegateStaking (3)** — all `IDelegateStaking` functions:
   `openSessionFromPool`, `closePoolSession`, `settleExpiredSession`.
5. **Add → DelegateStakingFunding (27)** — all `IDelegateStakingFunding` functions:
   `approvePoolFunder`, `revokePoolFunderApproval`, `grantStakingAllowance`,
   `revokeStakingAllowance`, `fundStakingAllowance`, `requestStakingWithdrawal`,
   `cancelStakingWithdrawal`, `servePoolWithdrawals`, `releasePoolLocks`,
   `purgeInactiveFunders`, `setMaxActiveFunders`, `setMinPrincipal`,
   `setSettlementGrace`, `setMaxAutoService`, `setMaxAutoReleaseDays`,
   `getStakingAllowance`, `isPoolFunderApproved`, `getAvailablePoolStake`,
   `getPoolFunders`, `getStakingWithdrawal`, `getPoolWithdrawalQueue`,
   `getPoolBalances`, `getPoolLockBuckets`, `getPoolSession`,
   `getDelegateStakingParameters`, `getDelegateStakingLimits`,
   `getDelegateStakingDelegationRight`.

The same cut uses the **new ProviderPayouts facet as init target**, with calldata for
`__ProviderPayouts_init()` (no arguments). Routing and initialization must be atomic;
do not split them into transactions. The init sets the lifetime-claimed start timestamp.
The `__SessionRouter_init` selector is replaced, but **must not be called again**.
There is no DelegateStaking initializer call in this cut.

Source: `test/helpers/deployers/diamond/facets/diamond-upgrade.ts:75–113`.

## Step 2 — verify after the cut

- [ ] **Routing:** all 66 selectors above resolve to their intended facets; no legacy
  SessionRouter remains in `facetAddresses()`. ProviderRegistry, ModelRegistry,
  Marketplace, IBidStorage and Delegation routes stay unchanged. Check moved views via
  the diamond using the ProviderPayouts ABI; call `getPoolSession` and `getPoolBalances`
  using DelegateStakingFunding, not DelegateStaking.
- [ ] **Boundary state:** compare the saved legacy views/storage and MOR balances at
  the boundary. Use identical explicit timestamps for math views and account for elapsed
  time/concurrent transactions when comparing live reads. `getProviderLifetimeClaimed`
  starts at `(0, cut block timestamp)` before post-cut payouts. On the fork, repeated
  ProviderPayouts init and SessionRouter init must reject as already initialized.
- [ ] **Legacy sessions:** rehearse late close in both payment modes, disputed remainder
  claims before/after the hold, close after a pre-cut provider claim, and released-stake
  withdrawals. Compare token deltas and state against the legacy flow at the same
  timestamps. Old sessions have `getPoolSession(id).exists == false`; pool close/settle
  must reject them. New regular sessions still work on old bids/models and continue the
  shared nonce without ID collisions.
- [ ] **Funding shortage (fork):** remove funding allowance; old-session close now
  completes user settlement and records provider debt instead of reverting. Check
  `getProviderOwed`, `getPendingProviderClaims`, `getProvidersTotalClaimed` and emission
  views; restore funding allowance and claim as an unrelated caller. Provider receives
  the debt, caller receives nothing, and provider debt clears.
- [ ] **Pool receipt path:** approve funder → grant → fund via DelegateStakingFunding;
  `openSessionFromPool` on an old bid → `closePoolSession` via DelegateStaking. Check
  the pool record, regular session record, unique ID, unchanged hot-wallet MOR balance,
  returned attribution, pool balances and exact day lock/bucket.
- [ ] **Pool expiry path (separate session):** after `endsAt + settlementGrace`, an
  unrelated caller can `settleExpiredSession` once; before the boundary it rejects.
  Repeated settle or settle after receipt close rejects. In the no-funding fork case,
  verify provider debt, unchanged hot-wallet balance and unchanged stats.
- [ ] **Events at the diamond:** regular receipt flow emits `SessionOpened`,
  `ProviderRewardAccrued`, `SessionRating`, `SessionClosed`; released stake withdrawal
  emits `UserWithdrawn`. Pool open emits `DelegateStakingPoolDrawn` + `SessionOpened`;
  receipt close emits `DelegateStakingPoolLockCreated`, `DelegateStakingPoolReturned`,
  `SessionClosed`, `SessionRating` for the tested early-close case. Expiry settlement
  emits `SessionClosed` + `DelegateStakingPoolSessionSettled`, with no `SessionRating`;
  deferred reward emits `ProviderRewardAccrued`, later debt payment `ProviderDebtPaid`.
  Decode and compare event arguments with the chosen scenario, not just event names.

Evidence: `test/diamond/DiamondUpgrade.test.ts:109–169,244–329,332–403,406–528`.
Pool expiry and detailed pool events are additional checks from
`test/diamond/facets/DelegateStaking.test.ts:145–162,417–451,641–707`, not a successful
settle scenario in the upgrade suite. Regular events/withdrawal:
`test/diamond/facets/SessionRouter.test.ts:434–456,1053–1065`; debt events:
`test/diamond/facets/ProviderPayouts.test.ts:303–344`. Live data collection and smoke-test
orchestration remain manual; destructive funding/boundary experiments belong on the fork.

## If a check fails — forward only

Before execution, stop on any mismatch. If the cut transaction reverts, verify that the
old routing/state remain intact before preparing another attempt.

After a successful cut, preserve the transaction receipt and state evidence; diagnose
against the fork and ship a reviewed replacement facet/cut **forward**. Preserve existing
storage, pool accounting, settlement and claim/withdraw access. Recheck both regular and
pool paths when shared settlement code changes (SessionRouter and DelegateStaking).

**No inverse cut restoring a pre-pool SessionRouter over live pool state. No removal of
DelegateStaking while pool state exists. No replay of the old day-lock Remove/Add cut,
no initializer reset and no snapshot-based on-chain rollback.** A code fix is another
pool-aware implementation, never a return to the pre-pool router.

Policy sources: RFP COLDC-R20 / §4.2; `docs/client-impact.md:119–124`;
`docs/contracts/README.md:140–147`. The earlier
[day-lock runbook](session-day-lock-upgrade-runbook.md) is historical; its cut recipe
is not the Stage 1 cut. See [facet routing](contracts/reference.md#facet-routing).
