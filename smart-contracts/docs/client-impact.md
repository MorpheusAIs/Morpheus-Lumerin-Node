# Stage 1 client-impact note

What changes for proxy-router, wallets and indexers after the Stage 1 cut (RFP §3.4.1,
§3.3.3 part A, §3.4.2, §3.6.2). Existing selectors, event signatures and storage slots are
unchanged; everything else is additive unless noted.

## Pool sessions (new, `DelegateStaking` / `DelegateStakingFunding`)

- Cold wallets fund a pool for a hot wallet: `approvePoolFunder` (hot), then
  `grantStakingAllowance` + `fundStakingAllowance` (cold). See
  [delegate-staking-funder-economics.md](delegate-staking-funder-economics.md).
- The hot wallet (or its session delegate) opens with `openSessionFromPool(hot, amount,
  approval, signature)` and closes with `closePoolSession(receipt, signature)`. The
  provider approval must name the hot wallet as the user.
- `SessionRouter.closeSession` **rejects** pool sessions (`SessionPoolSessionCannotBeClosed`);
  `closePoolSession` and `settleExpiredSession` reject regular sessions.
- Pool sessions are stake-pool only and use the same session ID scheme and the shared
  `sessionNonce`. They emit the usual `SessionOpened` / `SessionClosed`.
- After `endsAt + settlementGrace` anyone can call `settleExpiredSession(sessionId)`. It
  returns the stake to the pool and records the provider reward as debt; it does not change
  stats or ratings.
- After an early close, the day-locked part of the stake stays on the funders that financed
  the session (`lockedExposure` in `getStakingAllowance`) until the next UTC midnight, so
  `getAvailablePoolStake` drops by that amount until then.
- A repeated withdrawal request of a funder moves its queue entry to the tail.
- Delegation: the `delegation.rules.delegate.staking` right (or a full delegation) lets a
  delegate of the hot wallet approve/revoke funders and a delegate of a funder request or
  cancel that funder's withdrawals. Funds still go only to the funder. A cold funder should
  not give the hot wallet a full delegation or session rights: those let it spend the
  funder's MOR allowance in regular sessions ([delegation.md](contracts/delegation.md)).

## Close no longer reverts on a short funding account (`SessionRouter`, `DelegateStaking`)

- Before: a stake-pool close reverted if the funding account lacked balance or allowance,
  blocking the user's refund. Now the user side always completes and the provider reward is
  recorded as debt (`providerOwed[provider]`, total `pendingProviderClaims`).
- Debt counts against the compute emission at accrual: `getComputeBalance` and
  `totalMORSupply` use `providersTotalClaimed + pendingProviderClaims`. Paying the debt later
  moves tokens only. `getProvidersTotalClaimed` still means "actually paid".
- When the emission is exhausted, the compute balance is 0 instead of reverting.
- **New rewards and old debts.** A new reward is paid at accrual whenever the funding account's balance
  and allowance cover it, as before; outstanding debts are not reserved. Debts are paid
  best-effort on claim: `claimForProvider` pays the claimed provider's debt up to
  `min(balance, allowance, owed)`, first claim first paid. There is no FIFO or age priority
  between providers, and new rewards may be paid before older debts; keepers should claim
  for every provider with debt after each top-up.
- New events: `ProviderRewardAccrued(sessionId, provider, accrued, paid, owed)`,
  `ProviderDebtPaid(provider, payoutTarget, amount)`.
- Direct-pay sessions are paid from the session escrow as before and never create debt.

## `claimForProvider` is permissionless

- Anyone can call it; the money always goes to the provider's payout target. It accrues
  what the session still owes and then pays the provider's debt.
- It now reverts with `SessionNotEndedOrNotExist` for an unknown session ID. When the
  dispute hold covers the remaining reward it accrues nothing (instead of reverting) and
  still pays any outstanding debt of that provider.

## Provider payout target (`ProviderPayouts`)

- `setPayoutTarget(provider, target)` by the provider itself or by a delegate with the new
  explicit `delegation.rules.payout` right. Existing provider/session delegations and full
  delegations do **not** grant it. `address(0)` or the provider address resets to default;
  the diamond address is rejected.
- All provider payouts (immediate at close, `claimForProvider`, debt payments) go to the
  target set at payment time.
- **Debt is indexed by the provider address**, not by the target (COLDP-R5/R6): query
  `getProviderOwed(provider)` and `getPayoutTarget(provider)`. A target change never strands
  debt; it is paid to the current target.
- `ProvidersDelegate` providers cannot set a target (the contract has no forwarding call).
  This is intended: its rewards must reach the delegate contract for staker accounting.
- A delegate's call checks the provider's outgoing delegations one by one; a provider with a
  very long delegation list should call `setPayoutTarget` directly.
- `getProviderLifetimeClaimed(provider)` returns payouts since the upgrade and the start
  timestamp; history before the upgrade is not available per provider.

## Rating events (§3.6.2)

- Every receipt close emits `SessionRating(sessionId, modelId, provider, closeoutType,
  tpsScaled1000, ttftMs)`; `closeoutType = 1` marks a disputed close. Settlement is unchanged.

## Views moved: diamond selectors unchanged, facet ABI changed

`getSessionId`, `getSessionEnd`, `stakeToStipend`, `stipendToStake`, `getUserStakesOnHold`,
`getTodaysBudget`, `getComputeBalance`, `totalMORSupply`, `startOfTheDay` are served by the
`ProviderPayouts` facet with the same selectors and return values (plus the debt term in the
emission views, which is 0 at upgrade time). Calls to the diamond keep working.

They are no longer in the `SessionRouter` / `ISessionRouter` ABI. Clients that build from
per-facet ABIs must take them from `ProviderPayouts`: after `npm run bindings-go` and
`npm run copy-bindings`, the `sessionrouter` Go binding no longer has these methods, and
proxy-router calls such as `GetUserStakesOnHold`, `GetTodaysBudget` and `TotalMORSupply`
must move to the `providerpayouts` binding (or to a binding built from the merged diamond
ABI) before it builds against the new bindings. Solidity code that calls them through
`ISessionRouter` must use `IProviderPayouts`.

## Intentional tightening

- For contract providers, the `owner()` lookup used to validate approvals and receipts is
  capped at 100,000 gas. Before, a provider contract could burn the caller's gas in
  `owner()` and make the user's close revert. A provider whose `owner()` needs more gas is
  treated as a non-matching signer (approval rejected, receipt counted as a dispute).
  `ProvidersDelegate.owner()` uses a few thousand gas.

## Assumptions

- MOR is a standard ERC-20: no transfer fees, rebasing, hooks or blacklist. Pool and payout
  accounting credit the requested amounts.

## Known legacy behaviour (open question)

- Regular `openSession` does not check that the provider approval names `user_`, and the
  "approval used" flag is shared with `openSessionFromPool`. A third party can spend an
  approval seen in the mempool and make the intended open revert, and a session delegate
  can open a direct-pay session to a provider it controls with the user's MOR. Unchanged
  from `dev`; details and the possible fix in [sessions.md](contracts/sessions.md#signed-messages)
  and [delegation.md](contracts/delegation.md).

## Upgrade policy

- Forward-only (COLDC-R20): never cut back a SessionRouter without the pool checks, or remove
  `DelegateStaking`, while pool state is live. A fix ships as a new facet.
- Sessions opened before the cut are regular sessions (no pool record) and close, claim and
  withdraw as before; on a short funding account they now produce debt instead of reverting.
