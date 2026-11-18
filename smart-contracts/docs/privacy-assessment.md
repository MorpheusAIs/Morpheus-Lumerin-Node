# Stage 1 privacy assessment — on-chain linkability

Covers RFP §3.4.3 (AC-PRIV-1, AC-PRIV-2). Stage 1 adds no privacy infrastructure and no
external dependencies; this note records what the new features make public.

## Consumer side (delegate staking pool)

Everything below is readable from storage views and events by anyone.

| Link | Where it is visible |
|---|---|
| cold funder ↔ hot wallet | `DelegateStakingAllowanceGranted`, `DelegateStakingAllowanceFunded`, `DelegateStakingPoolFunderApproved`; views `getStakingAllowance`, `getPoolFunders` |
| amounts per funder | grant cap, lifetime funded, principal, exposure, locked exposure, pending (`getStakingAllowance`) |
| funder ↔ individual session | `DelegateStakingPoolDrawn` / `DelegateStakingPoolReturned` per funder and session; `getPoolSession` while open |
| hot wallet ↔ provider, model, timing | unchanged from regular sessions: `SessionOpened`, `SessionClosed`, session getters |
| funder withdrawals | `DelegateStakingWithdrawalRequested` / `Paid` / `Cancelled` |

Consequence: a cold wallet that funds a pool is publicly tied to the hot node wallet and,
through the per-session draw events, to that node's provider choices and usage pattern.
Using a fresh cold address per pool limits the link to that pool only; reusing one cold
wallet across pools links those pools together.

## Provider side (payout target)

| Link | Where it is visible |
|---|---|
| provider ↔ payout target | `ProviderPayoutTargetUpdated`, `getPayoutTarget` |
| payouts to the target | `ProviderRewardAccrued`, `ProviderDebtPaid`, and the MOR `Transfer` events |
| provider debt | `getProviderOwed(provider)` per provider address; `getPendingProviderClaims()` as a global total |

Consequence: the cold payout address is publicly tied to the provider address and to
every reward it receives. A target change is recorded on-chain; old targets stay linked.

## What Stage 1 does not do

- No mixing, stealth addresses, relayers or other privacy tooling (AC-PRIV-2).
- The hot wallet stays the visible session actor; the cold wallet never signs session
  operations.
