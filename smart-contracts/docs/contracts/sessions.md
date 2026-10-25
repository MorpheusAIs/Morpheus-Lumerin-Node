# Sessions

Facet: `SessionRouter` (regular sessions). Shared logic: `SessionSettlement`. Storage:
`SessionStorage`, `StatsStorage`. Pool sessions use the same settlement code through
`DelegateStaking`; see [delegate-staking.md](delegate-staking.md).

## Lifecycle

```
provider signs approval ──► openSession ──► (session runs until endsAt or an earlier close)
                                               │
              closeSession(receipt) ◄──────────┘   user or session delegate
                 │  user: unused stake back now, today's used part day-locked
                 │  provider: reward paid now, held (dispute), or recorded as debt
                 ▼
     withdrawUserStakes (user, after the lock)      claimForProvider (anyone, any time later)
```

## Signed messages

Both messages are ABI-encoded and signed as an Ethereum signed message
(`ECDSA.toEthSignedMessageHash(keccak256(message))`, i.e. `personal_sign` over the hash).

| Message | Encoding |
|---|---|
| Provider approval (open) | `abi.encode(bytes32 bidId, uint256 chainId, address user, uint128 timestamp)` |
| Provider receipt (close) | `abi.encode(bytes32 sessionId, uint256 chainId, uint128 timestamp, uint32 tpsScaled1000, uint32 ttftMs)` |

Checks:

- `chainId` must equal `block.chainid`.
- `block.timestamp <= timestamp + SIGNATURE_TTL` (10 minutes). A timestamp in the future
  is not rejected; it simply stays valid until `timestamp + 10 minutes`.
- A valid signer is the bid's provider, a delegate of the provider with
  `delegation.rules.provider`, or, for a contract provider, the address its `owner()`
  returns. The `owner()` call is a static call capped at 100,000 gas that must return exactly
  32 bytes holding a clean address; anything else counts as "not the owner".
- An approval can be used once: `isProviderApprovalUsed[approvalBytes]` is set at open and
  is shared by regular and pool sessions.
- Regular `openSession` does not compare the approval's `user` with the session user
  (legacy behaviour). `openSessionFromPool` requires them to be equal.
- **Known consequences (open question, unchanged from `dev`):**
  - *Griefing.* Anyone who sees an approval in the mempool can spend it first in their own
    `openSession` (with their own stake, returned on close). The intended `openSession` or
    `openSessionFromPool` then reverts with `SessionDuplicateApproval`. No funds are lost.
  - *Delegate spending.* A session delegate (or a full delegate) can open a direct-pay
    session for the user to a bid of a provider it controls; the stake, pulled from the
    user's wallet up to its allowance to the diamond, becomes that provider's reward. See
    [delegation.md](delegation.md).
  - A possible fix is to require the approval's user to equal `user_` in `openSession`. It
    stops the griefing (not the delegate case, since such a provider can sign an approval
    naming the victim) and changes behaviour for integrations that open with an approval
    issued for another address, so it waits for a decision.

## Stipend and session length

A stake buys a daily *stipend*: the stake's share of the non-compute MOR supply, applied to
the compute balance, scaled to one day (1%):

```
stipend(amount, t)      = amount * computeBalance(t) / (totalMORSupply(t) * 100)
stipendToStake(s, t)    = s * totalMORSupply(t) * 100 / computeBalance(t)

duration = min(stipend(amount, openedAt) / pricePerSecond, maxSessionDuration)
endsAt   = openedAt + duration
```

- `computeBalance(t)` is the compute pool's (index 3) emission up to the start of `t`'s UTC
  day minus everything already paid or owed to providers
  (`providersTotalClaimed + pendingProviderClaims`), floored at 0.
- `totalMORSupply(t)` is the emission of all other pools up to the same day plus that same
  provider debit.
- Pools mirror the L1 Distribution contract (`setPoolConfig`, owner). The formulas use
  `LinearDistributionIntervalDecrease.getPeriodReward`.
- If the supply is 0 the stipend is 0; if the compute balance is 0 the reverse conversion is
  0 (no revert). Details of the debit: [provider-payouts.md](provider-payouts.md#emission-views).

## Opening

```solidity
openSession(address user_, uint256 amount_, bool isDirectPaymentFromUser_,
            bytes approvalEncoded_, bytes signature_) returns (bytes32 sessionId)
```

- Caller: `user_` or its delegate with `delegation.rules.session`. MOR is pulled from
  `user_`, so the user must have approved the diamond.
- Session ID: `keccak256(abi.encodePacked(user, provider, bidId, sessionNonce))`; the
  nonce is global and shared with pool sessions (`ProviderPayouts.getSessionId` computes it).
- Reverts if the bid is missing or deleted, the signature does not match, the approval was
  used, the duration is below 5 minutes, or (direct pay) the stake cannot cover
  `duration * pricePerSecond`.
- The session is added to the user, provider and model session lists.

**Direct pay vs stake.** With `isDirectPaymentFromUser = true` the provider is paid from the
session stake held by the diamond; otherwise from the funding account. Duration and lock
rules are the same in both modes.

## Closing

```solidity
closeSession(bytes receiptEncoded_, bytes signature_)
```

- Caller: the session user or its session delegate. The session ID comes from the receipt.
- A pool session is rejected (`SessionPoolSessionCannotBeClosed`); use `closePoolSession`.
- A well-formed signature from a signer that is not authorised for the provider does not
  revert: the close counts as a **dispute** (`closeoutType = 1`). A malformed signature
  (wrong length, non-canonical `s`) reverts inside `ECDSA.recover`, and wrong chain ID or an
  expired receipt revert as well.

### User settlement

```
sessionEnd = min(closedAt, endsAt)
releaseAt  = startOfDay(sessionEnd) + 1 day
userStake  = stake - (direct pay ? full provider reward : 0)

if now < releaseAt:
    usedToday   = sessionEnd - max(openedAt, startOfDay(sessionEnd))
    amountToLock = min(userStake, stipendToStake(usedToday * pricePerSecond, startOfDay(sessionEnd)))
amountToReturn = userStake - amountToLock
```

The stake that paid for the stipend used on the session's last day stays locked until the
next UTC midnight, so the same stake cannot buy stipend twice on one day. The lock is
anchored to `sessionEnd`, so a late close (after `endsAt`) on the same day also locks.

`amountToReturn` is transferred at once; `amountToLock` becomes an `OnHold(amount,
releaseAt)` entry in `userStakesOnHold[user]`.

### Provider side

```
reward = (sessionEnd - openedAt) * pricePerSecond - providerWithdrawnAmount
if dispute and closedAt < endsAt:
    reward -= (closedAt - max(startOfDay(closedAt), openedAt)) * pricePerSecond   // held until next UTC day
```

The amount goes through reward accrual: limiter cap, then immediate payment or debt. See
[provider-payouts.md](provider-payouts.md). The held part is claimed later with
`claimForProvider`.

### Stats and rating

On every receipt close: `totalCount++` for the provider/model pair. Without a dispute:

- pair (`getProviderModelStats`): `successCount++`, running means of `tpsScaled1000` and
  `ttftMs`, and `totalDuration` as a cumulative sum of `closedAt − openedAt` (`uint32`;
  includes time after `endsAt` for late closes);
- model (`getModelStats`): `tpsScaled1000` and `ttftMs` are means over the providers'
  means (each provider counts once); `totalDuration` is fed the pair's cumulative duration
  on every close without removing the previous value, so it is not a duration mean.

With a dispute only `closeoutType = 1` is stored. Then:

```
SessionRating(sessionId, modelId, provider, closeoutType, tpsScaled1000, ttftMs)
SessionClosed(user, sessionId, provider)
```

## Claiming for a provider

```solidity
claimForProvider(bytes32 sessionId_)   // anyone
```

- Reverts with `SessionNotEndedOrNotExist` for an unknown session.
- Accrues what the session still owes: `reward - providerWithdrawnAmount`, minus the
  same-day hold while `block.timestamp` is still on `closedAt`'s UTC day. Payment goes to
  the provider's payout target; the caller gets nothing.
- Then pays the provider's outstanding debt as far as the funding account allows.
- Typical uses: the held part of a disputed close (next day), a remainder capped by the
  limiter (after the period rolls over), an expired session nobody closed (`closedAt = 0`,
  after `endsAt`), paying debt after the funding account is topped up.
- If nothing is due the call only tries to pay debt; it does not revert.
- It does not close the session or return the user's stake. A regular session has no
  permissionless settlement: only the user or its session delegate can close it (pool
  sessions have `settleExpiredSession`).

## Withdrawing locked stake

```solidity
withdrawUserStakes(address user_, uint8 iterations_)   // user or session delegate
```

Walks `userStakesOnHold[user]` from the end and removes up to `iterations_` released
entries, paying their sum to the user. Reverts with `SessionUserAmountToWithdrawIsZero` if
nothing is released. `getUserStakesOnHold(user, iterations)` (served by `ProviderPayouts`)
returns `(available, hold)`; it ignores `iterations` and scans every entry, so its cost
grows with the array.

## Session record

| Field | Meaning |
|---|---|
| `user` | Session owner (the hot wallet for pool sessions) |
| `bidId` | Bid the session was opened on; price is read from it |
| `stake` | Staked amount |
| `closeoutReceipt` | Receipt bytes stored at close |
| `closeoutType` | 0 normal, 1 disputed |
| `providerWithdrawnAmount` | Reward already accrued for the provider (paid or owed) |
| `openedAt`, `endsAt`, `closedAt` | Timestamps; `closedAt = 0` while open |
| `isActive` | True until closed or settled |
| `isDirectPaymentFromUser` | Direct-pay flag (always false for pool sessions) |

## Owner settings

| Function | Rule |
|---|---|
| `setMaxSessionDuration(uint128)` | At least `MIN_SESSION_DURATION` (5 minutes); deploy uses 7 days |
| `setPoolConfig(index, Pool)` | Index must exist; copy values from the L1 Distribution contract |

`fundingAccount` is set once in `__SessionRouter_init`; it must approve the diamond to pull
MOR.

## Views

`SessionRouter`: `getSession`, `getUserSessions`, `getProviderSessions`, `getModelSessions`,
`getTotalSessions`, `getPools`, `getPool`, `getFundingAccount`, `getProvidersTotalClaimed`,
`getIsProviderApprovalUsed`, `getMaxSessionDuration`, `getProviderModelStats`,
`getModelStats`. The constants `MIN_SESSION_DURATION`, `SIGNATURE_TTL` and
`COMPUTE_POOL_INDEX` are public in `SessionStorage` but not routed by the diamond (they are
not in the interfaces the cut uses).

`ProviderPayouts` (same selectors as before the upgrade): `getSessionId`, `getSessionEnd`,
`stakeToStipend`, `stipendToStake`, `getUserStakesOnHold`, `getTodaysBudget`,
`getComputeBalance`, `totalMORSupply`, `startOfTheDay`.
