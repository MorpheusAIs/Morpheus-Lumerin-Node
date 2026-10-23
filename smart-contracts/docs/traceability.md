# Stage 1 traceability matrix

RFP §4.3 item 1: every Stage 1 acceptance criterion → named tests. Paths are relative to
`smart-contracts/test/`. Abbreviations: **DS** = `diamond/facets/DelegateStaking.test.ts`,
**DSF** = `diamond/facets/DelegateStakingFunding.test.ts`, **PP** =
`diamond/facets/ProviderPayouts.test.ts`, **SR** = `diamond/facets/SessionRouter.test.ts`,
**UP** = `diamond/DiamondUpgrade.test.ts`, **PR** = `diamond/ProtocolProperties.test.ts`.

## §3.4.1 Consumer cold/hot staking pool

| AC | Tests |
|---|---|
| AC-COLDC-1 no grant → no pool session | DS `should reject a pool with neither an external grant nor self-escrow` |
| AC-COLDC-2 hot opens without cold key, receives no MOR | DS `should open from a funded grant without moving MOR through the hot wallet`; DS `should return unused stake, create the exact day lock, release attribution, and keep hot balance unchanged`; PR `should keep pool, funder, debt, emission, and diamond balance invariants (seed 1)`, `should keep pool, funder, debt, emission, and diamond balance invariants (seed 7)`, `should keep pool, funder, debt, emission, and diamond balance invariants (seed 42)` (hot balances asserted constant after every step) |
| AC-COLDC-3 Many:1 FIFO debit and recycle attribution | DS `should draw FIFO with partial draws, skip revoked entries, and draw self-escrow last`; DS `should track per-funder exposure while open and return it on close even after revoke`; DSF `should put a revoked and re-granted funder at the FIFO tail`; DSF `should put an expired funder removed by a draw at the FIFO tail after re-granting`; DSF `should put a fully withdrawn and re-funded funder at the FIFO tail`; DSF `should put an expired funder renewed before any draw at the FIFO tail`; DSF `should keep the FIFO position when a live grant is updated` |
| AC-COLDC-4 withdraw while in use | DSF `should pay a withdrawal during pool use from free liquidity without locking other funders (AC-COLDC-4)`; DSF `should pay free funds, queue the rest, coalesce requests, reserve FIFO, and not lock other funders pro-rata` |
| AC-COLDC-5 cap enforced, worst case fits a block | DSF `should enforce the admission limit, purge expired grants on admission, and expose cursor behavior`; DS `should draw from 128 funders plus self-escrow and return every draw`; PR `should draw, close, and settle 129 draws within the gas budget`; PR `should purge a full list of expired grants when admitting a new funder within the gas budget`; PR `should serve the maximum withdrawal batch within the gas budget` (gas asserted below 10M; the randomized runs also assert that a pool never holds more than one unreleased lock bucket, so the bucket-release loop never has a larger backlog) |
| AC-COLDC-6 revoked/expired grants do not inflate gas, funds withdrawable | DS `should skip a revoked funder approval immediately`; DS `should skip and unlink an expired funder during draw traversal`; DSF `should let a delegate withdraw to the funder and allow withdrawal after revoke or expiry`; PR `should purge a full list of expired grants when admitting a new funder within the gas budget` |
| AC-COLDC-7 dead hot: settle with empty funding, provider claims | DS `should settle a dead node into provider debt without stats, ratings, or hot-wallet transfers`; PR `should recover funder capital and pay the provider without hot, provider, or owner keys` |
| AC-COLDC-8 day-locks self-release | DSF `should release a lock left untouched for weeks on the first withdrawal (AC-COLDC-8)`; DSF `should expose matured-but-unreleased liquidity and lazily release an old bucket on fund`; DS `should lazily release matured buckets and service the withdrawal queue while closing` |
| AC-COLDC-9 solvency invariant across mixed operations | PR `should keep pool, funder, debt, emission, and diamond balance invariants (seed 1)`, `should keep pool, funder, debt, emission, and diamond balance invariants (seed 7)`, `should keep pool, funder, debt, emission, and diamond balance invariants (seed 42)`; DSF `should conserve every funder balance and never pay the hot wallet across a mixed scenario` |
| AC-COLDC-10 funder economics documented | `docs/delegate-staking-funder-economics.md` |

Path separation and approval binding (COLDC-R6, R9, R18): DS `#closeSession` tests, DS
`should reject an approval for another user`, DS `should not let a delegate draw one user pool
for another user`, DS `should use the shared nonce across regular and pool sessions`, PR `should close each interleaved session only through its own path`, UP
`should not treat legacy sessions as pool sessions`.

## §3.3.3 part A: close without revert, provider debt

| AC | Tests |
|---|---|
| AC-CLOSE-1 unused stake returns, on-hold formula unchanged | SR `should close session and send rewards for the provider, no dispute, early closure`; SR `should day-lock the used stipend when closing late on the same day the session ended (#830)`; UP `should close, claim and withdraw with the same amounts as the legacy code`; DS `should lock and pay exactly like a regular session for an early same-day close`; DS `should lock and pay exactly like a regular session for a late close after endsAt on the same day` |
| AC-CLOSE-3 disputed close unchanged | SR `should close session and send rewards for the provider, with dispute, early closure on the next day`; SR `should close session and do not send rewards for the provider, with dispute, early closure on the same day`; DS `should pay immediately, apply dispute hold, or accrue debt exactly like regular close` |
| AC-CLOSE-4 under-funded close completes, debt, later claim | PP `should close on a funding balance shortage without changing exact user settlement`; PP `should record full debt on an allowance shortage and pay when balance and allowance are exactly sufficient`; UP `should record provider debt instead of reverting when funding is empty` |
| AC-CLOSE-4b emission counters and conservation at accrual | PP `should debit claimed plus pending values and keep all emission views unchanged when debt is paid`; PP `should saturate a depleted compute pool and still close and claim an already-open session`; PR `should keep pool, funder, debt, emission, and diamond balance invariants (seed 1)`, `should keep pool, funder, debt, emission, and diamond balance invariants (seed 7)`, `should keep pool, funder, debt, emission, and diamond balance invariants (seed 42)` (emission debit and exact diamond balance asserted after every step) |
| AC-CLAIM-3 permissionless claim cannot redirect funds | PP `should allow a stranger to claim and route payment to the target`; SR `should allow anyone to claim while always paying the provider`; PP `#claimForProvider for an unknown session` |
| CLAIM-R4 reconcile to close events | PP `should reconcile payout events, provider gain, and outstanding debt`; PP `debt priority` tests |

AC-CLOSE-2, AC-CLOSE-5, AC-CLAIM-1, AC-CLAIM-2 belong to §3.3.3 part B (Stage 2).

## §3.4.2 Provider cold payout target

| AC | Tests |
|---|---|
| AC-COLDP-1 rewards land at target | PP `should route immediate close payment to the configured target and default to the provider when unset`; PP `should route a direct-pay session from the diamond to the target`; DS `should pay pool-session rewards and settle debt to the current payout target only` |
| AC-COLDP-2 unset → provider | PP `should route immediate close payment to the configured target and default to the provider when unset` |
| AC-COLDP-3 auth, event, view | PP `#setPayoutTarget` tests |
| AC-COLDP-5 (Stage 1 part) debt paid to target | PP `should pay old debt to the current target and pay the provider after clearing the target`; documented in `docs/client-impact.md` |

AC-COLDP-4 (fee before routing) is Stage 2 (§3.3.4).

## §3.6.2 Rating events

| AC | Tests |
|---|---|
| AC-RATE-1 stable schema, settlement untouched | SR `should emit SessionOpened, ProviderRewardAccrued, SessionRating, and SessionClosed on the paid path`; SR `should emit disputed SessionRating with closeout type 1`; DS `should settle a dead node into provider debt without stats, ratings, or hot-wallet transfers` |

## §3.4.3 Privacy

| AC | Evidence |
|---|---|
| AC-PRIV-1, AC-PRIV-2 | `docs/privacy-assessment.md` |

## Upgrade (F1, COLDC-R20)

UP `#diamondCut`, `views and storage at the upgrade boundary`, `sessions opened before the
upgrade`, `sessions opened after the upgrade`, `facet bytecode size`.
