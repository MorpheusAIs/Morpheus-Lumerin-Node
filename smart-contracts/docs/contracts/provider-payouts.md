# Provider payouts

Facet: `ProviderPayouts`. Accrual and debt logic: `SessionSettlement`. Storage:
`ProviderPayoutStorage` (slot `diamond.standard.provider.payout.storage`).

Provider rewards for stake sessions (regular or pool) come from the **funding account**: an
external account that holds the compute share of the emission and approves the diamond to
pull MOR. The diamond pulls with `transferFrom(fundingAccount, payoutTarget, amount)`; the
tokens never sit in the diamond. Direct-pay sessions are paid from the session stake held by
the diamond.

## Reward accrual

Every path that owes a provider money (regular close, pool close, `claimForProvider`,
`settleExpiredSession`) calls the same accrual step:

1. **Limiter rollover.** If `block.timestamp > limitPeriodEnd`, a new 365-day period starts
   and `limitPeriodEarned = 0`.
2. **Cap.** `amount = min(amount, stake - limitPeriodEarned)`. If the result is 0, nothing
   happens (no event).
3. **Record.** `session.providerWithdrawnAmount += amount`, `limitPeriodEarned += amount`.
4. **Pay or owe.**

   | Session | Paid now when | Otherwise |
   |---|---|---|
   | Direct pay | Always, from the diamond's session escrow | — |
   | Stake, close or claim | The funding account's balance **and** allowance each cover `pendingProviderClaims + amount` | Whole amount becomes debt |
   | Stake, `settleExpiredSession` | Never (settlement makes no transfer) | Whole amount becomes debt |

5. **Book.** Paid: `providersTotalClaimed += paid`, `lifetimeClaimed[provider] += paid`.
   Owed: `providerOwed[provider] += owed`, `pendingProviderClaims += owed`.
6. **Event.** `ProviderRewardAccrued(sessionId, provider, accrued, paid, owed)` with
   `accrued == paid + owed`.

A short funding account never makes a close revert: the user side completes and the reward
becomes debt.

## Reward limiter

A provider can earn at most its registered `stake` per 365-day period. The part of a
session's reward above the limit is not accrued; it stays in the session (`reward -
providerWithdrawnAmount`) and can be claimed with `claimForProvider` once a new period has
started. A provider with zero stake accrues nothing.

## Provider debt

Debt is kept per **provider address** (`providerOwed[provider]`), with the global total in
`pendingProviderClaims`. It survives deregistration and payout-target changes.

**Paying debt.** `claimForProvider(sessionId)` (anyone) pays the debt of that session's
provider after accruing what the session itself still owes. Any session of the provider
works, including a fully settled one. The payment is best-effort:

```
amount = min(providerOwed[provider], balanceOf(fundingAccount), allowance(fundingAccount, diamond))
```

**Priority between new rewards and old debts.** A new reward is paid at once only if the
funding account can cover *all* outstanding debts plus the new reward. Otherwise it joins the
debt. A small top-up therefore cannot be spent on new sessions while older debts wait.

**Between providers there is no queue.** Debts are paid in the order claims arrive. There is
no reservation for other providers' debts, because reservation can strand money: with debts
A = 100 and B = 100 and 150 MOR in the funding account, reserving would leave 50 MOR that
nobody can take. After a top-up, keepers can call `claimForProvider` for every provider with
debt.

Paying debt changes only `providerOwed`, `pendingProviderClaims`, `providersTotalClaimed`
and `lifetimeClaimed`; the emission views do not move (see below).

Event: `ProviderDebtPaid(provider, payoutTarget, amount)`.

## Payout target

```solidity
setPayoutTarget(address provider_, address target_)
getPayoutTarget(address provider_) returns (address)   // effective target
```

- Caller: the provider itself, or a delegate with the explicit
  `delegation.rules.payout` right (`DELEGATION_RULES_PAYOUT()`), given either as an `ALL`
  delegation or as a `CONTRACT` delegation for the diamond. A full (empty-rights)
  delegation and the provider/session rights do **not** grant it. The check walks the
  provider's outgoing delegations, so a provider with a very long list should call directly.
- `target_ = address(0)` or `target_ = provider_` resets to the default (the provider).
- The diamond address is rejected (`ProviderPayoutTargetInvalid`). Other addresses are not
  checked; a target that cannot move tokens loses the payouts.
- Every payout (immediate, claim, debt) goes to the target set **at payment time**. Old debt
  is paid to the current target.
- `ProvidersDelegate` has no call that sets a target, so its rewards always reach the proxy
  and its staker accounting.

Event: `ProviderPayoutTargetUpdated(provider, effectiveTarget)`.

## Emission views

The compute emission is debited when a reward is **accrued**, not when it is paid:

```
emissionDebit     = providersTotalClaimed + pendingProviderClaims
getComputeBalance(t) = max(computePoolReward(startOfDay(t)) - emissionDebit, 0)
getTodaysBudget(t)   = getComputeBalance(t) / 100
totalMORSupply(t)    = otherPoolsReward(startOfDay(t)) + emissionDebit
```

Turning debt into payment moves the same value from `pendingProviderClaims` to
`providersTotalClaimed`, so the views stay the same. When the compute emission is used up,
the compute balance is 0 instead of reverting; sessions can still close and claim.
`getProvidersTotalClaimed()` keeps meaning "actually paid".

## Views

| View | Returns |
|---|---|
| `getPayoutTarget(provider)` | Effective target |
| `getProviderOwed(provider)` | Unpaid debt |
| `getPendingProviderClaims()` | Total unpaid debt |
| `getProviderLifetimeClaimed(provider)` | Amount paid since the facet was initialised, and that timestamp |
| `getComputeBalance`, `getTodaysBudget`, `totalMORSupply`, `stakeToStipend`, `stipendToStake`, `getSessionEnd`, `getSessionId`, `getUserStakesOnHold`, `startOfTheDay` | Session helpers, see [sessions.md](sessions.md) |

## Reconciliation

For indexers:

- `pendingProviderClaims == Σ providerOwed[p]`.
- For each provider: `Σ owed (ProviderRewardAccrued) − Σ ProviderDebtPaid == providerOwed`.
- Paid to targets: `Σ paid (ProviderRewardAccrued) + Σ ProviderDebtPaid`.
- `lifetimeClaimed` counts only payments after `__ProviderPayouts_init`.
